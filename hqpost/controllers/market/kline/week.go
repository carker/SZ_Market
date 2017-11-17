package kline

import (
	"ProtocolBuffer/projects/hqpost/go/protocol"
	"bytes"
	"encoding/binary"
	"io"
	"io/ioutil"

	"haina.com/market/hqpost/config"
	"haina.com/market/hqpost/models/filestore"
	"haina.com/market/hqpost/models/kline"
	"haina.com/share/lib"
	"haina.com/share/logging"
)

type BaseLine struct {
	sid      int32                     //股票SID
	date     []int32                   //单个股票的历史日期
	sigStock map[int32]*protocol.KInfo //单个股票的历史数据
	kindDays *[][]int32
}

// 生成周线
//this.WeekDay
func HisWeekKline(sids *[]int32) {
	for _, sid := range *sids {
		base := new(BaseLine)
		// 获取当天快照
		today, err := GetIntradayKInfo(sid)
		if err != nil {
			logging.Error("严重错误！！！，程序被迫停止执行")
			return
		}

		weekName := filePath(cfg, cfg.File.Week, sid)
		if !lib.IsFileExist(weekName) { // 不存在或其他做第一从生成 TODO
			if err = base.CreateWeekLine(sid, weekName); err != nil { // 在此之前day已更新
				continue
			}
		} else { // 追加周线
			if err = filestore.UpdateWeekLineToFile(sid, weekName, today); err != nil {
				logging.Error("UpdateWeekLineToFile: %v", err)
			}
		}
	}
}

// 读日线文件数据(用来生成其他线)
func (this *BaseLine) ReadHGSDayLines(sid int32) error {
	this.sid = sid
	sigle := make(map[int32]*protocol.KInfo)

	dayPath, is := kline.IsExistFileInHGSFileStore(cfg, cfg.File.Day, sid) // 文件是否存在
	if !is {
		logging.Error("%v:%d", NOTFOUND_DAYLINE_IN_HGSFILE, sid)
		return NOTFOUND_DAYLINE_IN_HGSFILE
	}
	bs, err := ioutil.ReadFile(dayPath)
	if err != nil || len(bs) == 0 {
		logging.Error("Create WeekLine: Read dayLine null|%v", err)
		return err
	}

	buff := &protocol.KInfo{}
	size := binary.Size(buff)
	for i := 0; i < len(bs); i += size {
		buffer := &protocol.KInfo{}
		if err = binary.Read(bytes.NewBuffer(bs[i:size+i]), binary.LittleEndian, buffer); err != nil && err != io.EOF {
			logging.Error("Create WeekLine: binary read dayline error|%v", err)
			return err
		}
		this.date = append(this.date, buffer.NTime)
		sigle[buffer.NTime] = buffer
	}
	this.sigStock = sigle
	return nil
}

// 读day文件生成week
func (this *BaseLine) CreateWeekLine(sid int32, weekFile string) error {
	err := this.ReadHGSDayLines(sid)
	if err != nil {
		return err
	}
	this.getSecurityWeekDay()
	wTable := this.ProduceWeekprotocol()
	if err := filestore.WiteHainaFileStore(weekFile, wTable); err != nil {
		logging.Error("CreateWeekLine: WiteHainaFileStore error | %v", err)
	}
	return nil
}

// 生成周线日期
func (this *BaseLine) getSecurityWeekDay() {
	if len(this.date) < 1 {
		logging.Error("Create WeekLine:%v---No historical data...", this.sid)
		return
	}
	var wday [][]int32
	sat, _ := filestore.DateAdd(this.date[0]) //该股票第一个交易日所在周的周日（周六可能会有交易）

	var dates []int32
	for j, date := range this.date {
		if filestore.IntToTime(int(date)).Before(sat) {
			dates = append(dates, date)
			if j == int(len(this.date)-1) { //执行到最后一个
				wday = append(wday, dates)
			}
		} else {
			wday = append(wday, dates)

			sat, _ = filestore.DateAdd(date)
			dates = nil
			dates = append(dates, date)
		}
	}
	this.kindDays = &wday
}

// 生成周K线
func (this *BaseLine) ProduceWeekprotocol() *protocol.KInfoTable {
	var tmps []protocol.KInfo
	var klist protocol.KInfoTable

	for _, week := range *(this.kindDays) { //每一周
		tmp := protocol.KInfo{}

		var (
			i          int
			day        int32
			AvgPxTotal uint32
		)
		for i, day = range week { //每一天
			stockday := this.sigStock[day]
			if tmp.NHighPx < stockday.NHighPx || tmp.NHighPx == 0 { //最高价
				tmp.NHighPx = stockday.NHighPx
			}
			if tmp.NLowPx > stockday.NLowPx || tmp.NLowPx == 0 { //最低价
				tmp.NLowPx = stockday.NLowPx
			}
			tmp.LlVolume += stockday.LlVolume //成交量
			tmp.LlValue += stockday.LlValue   //成交额
			AvgPxTotal += stockday.NAvgPx
		}

		tmp.NSID = this.sid
		tmp.NTime = this.sigStock[week[i]].NTime     //时间（取每周最后一天）
		tmp.NOpenPx = this.sigStock[week[0]].NOpenPx //开盘价（每周第一天的开盘价）
		tmp.NPreCPx = this.sigStock[week[0]].NPreCPx //本周一的昨收
		tmp.NLastPx = this.sigStock[week[i]].NLastPx //最新价
		tmp.NAvgPx = AvgPxTotal / uint32(i+1)        //平均价
		tmps = append(tmps, tmp)
		//logging.Debug("周线是:%v", tmps)
		klist.List = append(klist.List, &tmp)
	}
	return &klist
}

func filePath(cfg *config.AppConfig, kind string, sid int32) string {
	file, _ := kline.IsExistFileInHGSFileStore(cfg, kind, sid)
	return kline.HGSFilepath(file)
}
