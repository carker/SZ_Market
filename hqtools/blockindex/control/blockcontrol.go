package control

import (
	"ProtocolBuffer/projects/hqpublish/go/protocol"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"

	"haina.com/market/hqtools/blockindex/finchina"
	"haina.com/share/kityxu/utils"
	"haina.com/share/lib"
	"haina.com/share/logging"

	//"github.com/garyburd/redigo/redis"
	"github.com/golang/protobuf/proto"
	"haina.com/share/models"
	"haina.com/share/store/redis"
)

var (
	RedisStore *redis.RedisPool
	index      int
	kData      map[int32]protocol.KInfoTable
	Circu      map[int32]int64
)

const (
	baseSid = 100000001 //上证指数
	baseday = 20100101  //板块指数基准日
	workday = 251       // 一年工作日（只多不少）
)

const (
	REDIS_BLOCK_ID_LIST         = "hq:init:bk:1100"    //全板块ID列表
	REDIS_BLOCK_ELEMENT_ID_LSIT = "hq:init:bk:1100:%d" //板块下成份股list
	REDIS_SECURITY_SID_LIST     = "hq:st:nsid"         //证券代码表sids
	REDIS_SECURITY_HDAY         = "hq:st:hday:%s"      //历史日K线

	REDISKEY_SECURITY_STATIC = "hq:st:static:%d" ///<证券静态数据(参数：sid) (hq-init写入)
)

func getBytes(reply interface{}, err error) ([]byte, error) {
	if err != nil {
		return nil, err
	}
	switch reply := reply.(type) {
	case []byte:
		return reply, nil
	case string:
		return []byte(reply), nil
	case nil:
		return nil, nil
	}
	return nil, fmt.Errorf("redigo: unexpected type for Bytes, got type %T", reply)
}

func initRedis() {
	//	c, err := redis.Dial("tcp", "47.94.16.69:61380") //开发
	//	c.Send("AUTH", "8dc40c2c4598ae5a")
	//	if err != nil {
	//		return err
	//	}
	//	return nil
	RedisStore = redis.NewRedisPool("127.0.0.1:61380", "0", "8dc40c2c4598ae5a", 3)
}

func initMysql() error {
	//初始化 MySQL 配置
	err := models.Init("mysql", "finchina:finchina@tcp(114.55.105.11:3306)/finchina?charset=utf8")
	if err != nil {
		logging.Fatal(err)
		return err
	}
	return nil
}

func Operation() {
	initRedis()
	if err := initMysql(); err != nil {
		return
	}
	if err := initSecurityHisDKlines(); err != nil { //初始化所有证券历史日K线入Map(前复权后)
		logging.Error("%v", err.Error())
		return
	}

	bids, err := RedisStore.GetBytes(REDIS_BLOCK_ID_LIST)
	if err != nil {
		logging.Error("%v", err.Error())
		return
	}

	block := &protocol.BlockList{}
	if err = proto.Unmarshal(bids, block); err != nil {
		logging.Error("%v", err.Error())
		return
	}

	for _, b := range block.List { // 板块id
		indexTable, err := getBlockIndexByBlockID(b.SetID)
		if err != nil {
			return
		}
		if err = writeFile(b.SetID, indexTable); err != nil {
			logging.Error("write file error:%v", err.Error())
			return
		}
	}
}

//获取某板块指数
func getBlockIndexByBlockID(bid int32) (*protocol.KInfoTable, error) {
	ekey := fmt.Sprintf(REDIS_BLOCK_ELEMENT_ID_LSIT, bid)

	ele, err := RedisStore.GetBytes(ekey)
	if err != nil {
		logging.Error("%v", err.Error())
		return nil, err
	}

	element := &protocol.ElementList{}

	if err = proto.Unmarshal(ele, element); err != nil {
		logging.Error("%v", err.Error())
		return nil, err
	}

	binfos := make(map[int32][]*protocol.KInfo) //该板块该交易日下的所有成分股
	var tradDay []int32                         //上证指数的交易日作为板块交易日
	for _, bv := range kData[baseSid].List {    //交易日
		tradDay = append(tradDay, bv.NTime)

		for _, e := range element.List { //成份股
			//此处以上证指数的交易日作为板块交易日
			//成份股都是按日期从大到小排的序
			//初始日期为20100101及之后的第一个交易日(20100104)
			logging.Debug("len-element:%v", len(element.List))

			klen := len(kData[e.NSid].List)
			for i, kinfo := range kData[e.NSid].List {
				if kinfo.NTime == bv.NTime {
					binfos[bv.NTime] = append(binfos[bv.NTime], kinfo)
					break
				} else if kinfo.NTime < bv.NTime {
					if i == klen-1 {
						logging.Debug("股票:%v %v之后再无交易", kinfo.NSID, kinfo.NTime)
						skinfo := &protocol.KInfo{
							NSID:  kinfo.NSID,
							NTime: bv.NTime,
						}
						binfos[bv.NTime] = append(binfos[bv.NTime], skinfo)
					}
					continue
				} else {
					//此处停盘
					logging.Debug("%v停盘中-----------NTime:%v", kinfo.NSID, bv.NTime)
					skinfo := &protocol.KInfo{
						NSID:  kinfo.NSID,
						NTime: bv.NTime,
					}
					binfos[bv.NTime] = append(binfos[bv.NTime], skinfo)
					break
				}
			}
			//logging.Debug("len-binfos[bv.NTime]:%v", len(binfos[bv.NTime]))
		}
	}

	indexTable := blockIndexCoreAlgorithm(bid, binfos, tradDay) //板块指数
	if len(indexTable.List) == 0 {
		return nil, fmt.Errorf("Get block stocks failed...")
	}
	return indexTable, nil
}

//板块指数的核心算法
//通过板块ID(bid)、该板块成份股以交易日为分割的日K线(bmap)、所有交易日期(tradDay)
func blockIndexCoreAlgorithm(bid int32, bmap map[int32][]*protocol.KInfo, tradDay []int32) *protocol.KInfoTable {
	//tradDay 中时间是有序的， 而map无序
	if len(tradDay) == 0 {
		logging.Error("获取板块交易日失败")
		return nil
	}

	var tmp *protocol.KInfo
	var intable = &protocol.KInfoTable{}
	for i, key := range tradDay {
		//logging.Debug("板块ID:%v---交易日：%v", bid, key)
		var (
			//yesValue int64   = 0 //昨天总市值
			K float64 //系数

			preCPx int64       // 昨收zhishu
			openPx int64       // 开盘
			highPx int32       // 最高
			lowPx  int32  = -1 // 最低
			lastPx int64       // 最新
			volume int64       //成交量 * 10000
			value  int64       //成交额 * 10000
			avgPx  uint32      //平均价

			count uint32 //成份股有效个数
		)
		for _, kinfo := range bmap[key] {
			//logging.Debug("kinfo.NPreCPx:%v   NOpenPx:%v   NLastPx:%v-----%v", kinfo.NPreCPx, kinfo.NOpenPx, kinfo.NLastPx, circu)

			circu := Circu[kinfo.NSID] //某板块 某日 某成份股的流通盘

			preCPx += circu * int64(kinfo.NPreCPx) //昨收价 累计（昨天总市值）
			openPx += circu * int64(kinfo.NOpenPx) //开盘价
			lastPx += circu * int64(kinfo.NLastPx) //收盘价
			if highPx < kinfo.NHighPx {
				highPx = kinfo.NHighPx //最高价
			}
			if kinfo.NOpenPx != 0 { //该成份股在市交易
				if lowPx == -1 {
					lowPx = kinfo.NLowPx
				} else {
					if lowPx > kinfo.NLowPx {
						lowPx = kinfo.NLowPx
					}
				}
				count++
			}

			avgPx += kinfo.NAvgPx
			volume += kinfo.LlVolume
			value += kinfo.LlValue
		}
		opRate := (float64(openPx-preCPx)/float64(preCPx) + float64(1)) //C
		coRate := (float64(lastPx-preCPx)/float64(preCPx) + float64(1))

		//	logging.Debug("opRate:%v  coRate:%v ", opRate, coRate)
		var pinjun uint32
		var precpx int32
		if i != 0 {
			K = float64(tmp.NLastPx)
			precpx = int32(K)
		} else {
			K = float64(1000) * 10000
			precpx = 1000 * 10000
		}

		if count != 0 {
			pinjun = avgPx / count
		}
		//logging.Debug("BID:%v  time:%v K:%v preCPx=%v openPx=%v  highPx=%v  lowPx=%v  lastPx=%v  volume=%v  value=%v  avgPx=%v",
		//	bid, key, K, preCPx, openPx, highPx, lowPx, lastPx, volume, value, avgPx)

		indexInfo := &protocol.KInfo{
			NSID:     bid,
			NTime:    key,
			NPreCPx:  precpx,
			NOpenPx:  int32(opRate * K),
			NHighPx:  highPx,
			NLowPx:   lowPx,
			NLastPx:  int32(coRate * K),
			NAvgPx:   pinjun,
			LlVolume: volume,
			LlValue:  value,
		}
		tmp = indexInfo
		//logging.Debug("BID:%v  time:%v  indexInfo:%v", bid, key, indexInfo)
		intable.List = append(intable.List, indexInfo)
	}

	return intable
}

//初始化所有证券历史日K线入Map(前复权后)
func initSecurityHisDKlines() error {
	skey := REDIS_SECURITY_SID_LIST
	index = getLrangeIndex()

	kData = make(map[int32]protocol.KInfoTable) //日线
	Circu = make(map[int32]int64)               //流通盘

	sids, err := RedisStore.LRange(skey, 0, -1) // 取股票代码表
	if err != nil {
		logging.Error("%v", err.Error())
		return err
	}

	for _, sid := range sids {
		hkey := fmt.Sprintf(REDIS_SECURITY_HDAY, sid)
		hday, err := RedisStore.LRange(hkey, 0, index)
		if err != nil {
			logging.Error("%v", err)
			continue
		}
		if len(hday) == 0 {
			continue
		}

		ktable := make([]*protocol.KInfo, 0, 1024) // 历史K线集合
		for _, day := range hday {                 //单个股票的历史K线
			kinfo := &protocol.KInfo{}
			if err = proto.Unmarshal([]byte(day), kinfo); err != nil {
				logging.Error("%v", err.Error())
				return err
			}
			if kinfo.NTime < baseday {
				break
			}
			ktable = append(ktable, kinfo)
		}

		if len(ktable) == 0 {
			logging.Debug("the sid:%v history day Kline is null", sid)
			continue
		}

		// 对历史K线进行复权除权操作
		sd, _ := strconv.Atoi(sid)
		nsid := int32(sd)

		facs, err := finchina.GetReferFactors(nsid)
		if err != nil {
			logging.Error("%v", err.Error())
			return err
		}

		var kinfotable = protocol.KInfoTable{}
		if len(facs) != 0 {
			fgs, err := FactorGroupOp(facs, ktable)
			if err != nil {
				logging.Error("%v", err.Error())
				return err
			}

			for _, v := range fgs {
				kinfotable.List = append(kinfotable.List, v.Ls[:]...)
			}
		} else {
			ReverseKList(ktable) //此处翻转 但不做复权处理
			kinfotable.List = append(kinfotable.List, ktable...)
		}

		kData[nsid] = kinfotable //股票历史日K线
		circu, err := getCirculationFromSecurityStaitc(nsid)
		if err == nil {
			Circu[nsid] = circu //股票流通盘
		}
	}
	return nil
}

//从证券静态数据获取流通盘
func getCirculationFromSecurityStaitc(sid int32) (int64, error) {
	key := fmt.Sprintf(REDISKEY_SECURITY_STATIC, sid)

	static := &protocol.StockStatic{}

	bs, err := RedisStore.GetBytes(key)
	if err != nil {
		return -1, err
	}

	if err = proto.Unmarshal(bs, static); err != nil {
		logging.Error("-----getSecurityStaticFromeStore--error..%v", err.Error())
		return -1, err
	}
	return static.LlCircuShare, nil
}

/* 参数: bid- 板块id； table- 板块历史K线数组
 *
 * 返回: error nil:写文件成功  其他：失败
 */
func writeFile(bid int32, table *protocol.KInfoTable) error {
	dir := "/opt/develop/hgs/filestore/blockindex"
	filepath := fmt.Sprintf("%s/%d", dir, bid)

	// golang 先创建目录 再创建文件；如果目录不存在，直接全路劲创建文件是会失败的
	lib.CheckDir(dir)

	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	buffer := new(bytes.Buffer)
	for _, v := range table.List {
		if err := binary.Write(buffer, binary.LittleEndian, v); err != nil {
			return err
		}
	}

	if _, err = file.Write(buffer.Bytes()); err != nil {
		return err
	}
	return nil
}

func getLrangeIndex() int {
	today := utils.Today()
	return (today/10000 - baseday/10000 + 1) * workday
}
