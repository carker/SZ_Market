package f10

import (
	"strconv"

	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	. "haina.com/market/hqpublish/models"
	"haina.com/market/hqpublish/models/finchina"
	"haina.com/market/hqpublish/models/publish"
	hsgrr "haina.com/share/garyburd/redigo/redis"
	"haina.com/share/logging"
)

type F10MobileTerminal struct {
	CompInfo F10_Compinfo           `json:"compinfo"`
	Equity   F10_Equity_Shareholder `json:"equity"`
	Dividend F10_Dividend_Ro        `json:"dividend"`
	Finance  F10_Finance            `json:"finance"`
}

//1.公司资料
type F10_Compinfo struct {
	Name       string              `json:"name"`       // 公司名称
	ListTime   int32               `json:"listTime"`   // 上市日期
	Indus      string              `json:"industry"`   // 公司所属证监会行业（聚源）
	Prov       string              `json:"area"`       // 省份
	PTime      string              `json:"pTime"`      // 主营收入构成 日期
	Constitute []*BusiinfoKeyValue `json:"constitute"` // 主营收入构成

}

//2.股本股东
type F10_Equity_Shareholder struct {
	Totalshare float64 `json:"totalShare"` //总股本(万股)          ///TQ_SK_SHARESTRUCHG
	Circskamt  float64 `json:"circskamt"`  //流通股本
	Totalshamt float64 `json:"totalshamt"` //股东总户数            ///TQ_SK_SHAREHOLDERNUM
	//Totalshrto       float64 `json:"Totalshrto"`       //股东总户数较上期增减
	Top1sha          string  `json:"no1share"`         //第一大股东           ///TQ_SK_SHAREHOLDER
	Top10Rate        float64 `json:"top10rate"`        //前十大股东占比
	LegalPersonsRate float64 `json:"legalPersonsRate"` //法人所占比例         ///TQ_SK_SHAREHOLDERNUM
}

//3.分红配股
type F10_Dividend_Ro struct {
	List []DividendRo `json:"list"`
}

type DividendRo struct {
	Date      string  `json:"date"`      //年度
	Dividend  float64 `json:"dividend"`  //分红（元，税前）
	PRO       float64 `json:"pro"`       //送股（股）
	TranAddRT float64 `json:"tranAddRT"` //转增（股）
	BonusRT   float64 `json:"bonusRT"`   //赠股（股）
	RegDate   string  `json:"regDate"`   //股权登记日
}

//4.财务数据
type F10_Finance struct {
	EndDate    string  `json:"lDate"`     // 日期
	LPE        float32 `json:"lpe"`       // 市盈率
	EPS        float32 `json:"lEPS"`      // 每股收益
	MainIncome float64 `json:"income"`    // 主营业务收入        ///TQ_FIN_PROINCSTATEMENTNEW
	NetProfit  float64 `json:"netProfit"` // 净利润

	LPB         float32 `json:"lpb"`      // 市净率
	LBVPS       float32 `json:"lBVPS"`    // 每股净资产
	MainBizRate float32 `json:"inRate"`   // 主营收入同比增长率
	NProRate    float64 `json:"nProRate"` // 净利润增长率

}

// F10 首页主营收入构成
type BusiinfoKeyValue struct {
	KeyName string  `json:"keyName"`
	Value   float64 `json:"value"`
	Ratio   float64 `json:"ratio"`
}

/// 证券快照
const (
	REDISKEY_SECURITY_SNAP = "hq:st:snap:%d" ///<证券快照数据(参数：sid) (calc写入)
)

func F10Mobile(scode string) (*F10MobileTerminal, *string, error) {

	//

	sc := finchina.NewTQ_OA_STCODE()
	if err := sc.GetCompcode(scode); err != nil {
		return nil, nil, err
	}
	/*-------------------------------------------------------------------*/
	/*----------------------------公司信息--------------------------------*/
	comp := finchina.NewCompInfo()
	cinfo, err := comp.GetCompInfo(sc.COMPCODE.String)     // 查询公司资料
	industry, err := comp.GetCompTrade(sc.COMPCODE.String) // 查询行业
	if err != nil {
		logging.Debug("%v", err.Error())
		return nil, nil, err
	}
	// 查询上市日期 总股本
	sercui := finchina.NewSecurityInfo()
	securdate, err := sercui.GetSecurityBasicInfo(sc.COMPCODE.String)
	if err != nil {
		logging.Debug("%v", err.Error())
		return nil, nil, err
	}
	listdate, err := strconv.Atoi(securdate.LISTDATE.String) // 上市日期转int
	if err != nil {
		logging.Debug("%v", err.Error())
		return nil, nil, err
	}
	// 主营收入构成
	busiinfo := finchina.NewTQ_SK_BUSIINFO()
	busilist, err := busiinfo.GetBusiInfo(sc.COMPCODE.String)
	if err != nil {
		logging.Debug("%v", err.Error())
		return nil, nil, err
	}
	fbdata := ""
	var busil []*BusiinfoKeyValue
	for i, v := range busilist {
		if i == 0 {
			fbdata = v.ENTRYDATE.String
		}
		var kv BusiinfoKeyValue
		kv.KeyName = v.CLASSNAME.String
		kv.Value = v.TCOREBIZINCOME.Float64
		kv.Ratio = v.COREBIZINCRTO.Float64
		busil = append(busil, &kv)
	}

	t1 := F10_Compinfo{
		Name:       cinfo.COMPNAME.String,
		ListTime:   int32(listdate),
		Indus:      industry,
		Prov:       getProvince(cinfo.REGION.String),
		PTime:      fbdata,
		Constitute: busil,
	}

	/*-------------------------------------------------------------------*/
	/*--------------------------分红配股----------------------------------*/
	divs, err := finchina.NewDividendRO().GetDividendRO(sc.COMPCODE.String)
	if err != nil {
		logging.Debug("%v", err.Error())
		return nil, nil, err
	}

	t3 := F10_Dividend_Ro{}
	for _, v := range *divs {
		div := DividendRo{
			Date:      v.DIVIYEAR.String,
			Dividend:  v.PRETAXCASHMAXDVCNY.Float64, //分红
			PRO:       v.PROBONUSRT.Float64,         //送股（股）
			TranAddRT: v.TRANADDRT.Float64,          //转增（股）
			BonusRT:   v.BONUSRT.Float64,            //赠股（股）
			RegDate:   v.EQURECORDDATE.String,
		}
		t3.List = append(t3.List, div)
	}

	/*-------------------------------------------------------------------*/
	/*--------------------------财务数据----------------------------------*/
	// 调用快照接口获取最新价
	snapdate, err := getStockSnapshot(scode)
	if err != nil {
		logging.Debug("%v", err.Error())
		return nil, nil, err
	}
	// 查询TQ_FIN_PROTTMINDIC       财务数据_TTM指标（产品表）
	prottmdate, err := finchina.NewTQ_FIN_PROTTMINDIC().GetBaseByComCode(sc.COMPCODE.String)
	if err != nil {
		logging.Debug("%v", err.Error())
		return nil, nil, err
	}
	// 计算市盈率
	// 市盈率（动）=收盘价/EPSDILUTEDNEWP
	var lep float64
	if prottmdate.EPSDILUTEDNEWP.Float64 > 0 {
		lep = (float64(snapdate.NLastPx) / 10000) / prottmdate.EPSDILUTEDNEWP.Float64
	}
	// 查询资产负债表
	proba, err := finchina.NewTQ_FIN_PROBALSHEETNEW().GetBaseInfo(sc.COMPCODE.String)
	if err != nil {
		logging.Debug("%v", err.Error())
		return nil, nil, err
	}
	// 计算每股净资产
	// 每股净资产=RIGHAGGR/总股本
	var jzc float64
	if securdate.TOTALSHARE.Float64 > 0 {
		jzc = proba.RIGHAGGR.Float64 / (securdate.TOTALSHARE.Float64 * 10000)
	}
	logging.Error("================RIGHAGGR:%f=============TOTALSHARE:%f", proba.RIGHAGGR, securdate.TOTALSHARE)
	// 计算市净率
	// 市净率（动）=收盘价/每股净资产；
	var lpb float64
	if jzc > 0 {
		lpb = (float64(snapdate.NLastPx) / 10000) / jzc
	}

	f1, err := finchina.NewF10_MB_PROINCSTATEMENTNEW().GetF10_MB_PROINCSTATEMENTNEW(sc.COMPCODE.String)
	if err != nil || len(f1) == 0 {
		logging.Debug("%v", err.Error())
		return nil, nil, err
	}
	var bizRate float64 //营业收入同比增长 （本季度的营业收入-上一年的该季度的营业收入）/上一年的该季度的营业收入；
	if len(f1) < 5 || f1[4].BIZINCO.Float64 == float64(0.0) {
		bizRate = 1.0
	} else {
		bizRate = (f1[0].BIZINCO.Float64 - f1[4].BIZINCO.Float64) / f1[4].BIZINCO.Float64
	}

	f2, err := finchina.NewF10_MB_PROINDICDATA().GetF10_MB_PROINDICDATA(sc.COMPCODE.String)
	if err != nil {
		logging.Debug("%v", err.Error())
		return nil, nil, err
	}
	// 计算净利润增长率
	// =（本季度的净利润-上一年的该季度的净利润）/上一年的该季度的净利润；
	var nprate float64
	if f1[4].PARENETP.Float64 > 0 {
		nprate = (f1[0].PARENETP.Float64 - f1[4].PARENETP.Float64) / f1[4].PARENETP.Float64
	}

	t4 := F10_Finance{
		EndDate:    f2.ENDDATE.String,               // 日期
		LPE:        float32(lep),                    // 市盈率
		EPS:        float32(f1[0].BASICEPS.Float64), // 每股收益
		MainIncome: f1[0].BIZINCO.Float64,           // 营业收入
		NetProfit:  f1[0].PARENETP.Float64,          // 净利润

		LPB:         float32(lpb),     // 市净率
		LBVPS:       float32(jzc),     // 每股净资产
		MainBizRate: float32(bizRate), // 营业收入同比增长
		NProRate:    nprate,           // 净利润同比增长

	}

	/*-------------------------------------------------------------------*/
	/*-------------------------股本股东-----------------------------------*/
	equity, err := finchina.NewEquity().GetEquity(sc.COMPCODE.String)
	if err != nil {
		logging.Debug("%v", err.Error())
		return nil, nil, err
	}

	shnum, err := finchina.NewShareHolders().GetShareHolders(sc.COMPCODE.String)
	if err != nil {
		logging.Debug("%v", err.Error())
		return nil, nil, err
	}

	top10, err := finchina.NewShareHoldersTop10().GetShareHoldersTop10(sc.COMPCODE.String, t4.EndDate)
	if err != nil {
		logging.Debug("%v", err.Error())
		return nil, nil, err
	}

	var top10rate float64 = 0.0
	var nametop1 string

	for _, v := range *top10 {
		top10rate += v.HOLDERRTO.Float64
		if v.RANK.Int64 == int64(1) {
			nametop1 = v.SHHOLDERNAME.String
		}
	}

	num := finchina.NewTQ_SK_IINVHOLDCHG().GetInstitutionStockNum(sc.COMPCODE.String, t4.EndDate)

	t2 := F10_Equity_Shareholder{
		Totalshare: equity.TOTALSHARE.Float64,
		Circskamt:  equity.CIRCSKAMT.Float64,
		Totalshamt: shnum.TOTALSHAMT.Float64,
		//Totalshrto:       shrto,
		Top1sha:          nametop1,
		Top10Rate:        top10rate,
		LegalPersonsRate: num / (equity.ASK.Float64 * 10000),
	}

	f10 := &F10MobileTerminal{
		CompInfo: t1,
		Equity:   t2,
		Dividend: t3,
		Finance:  t4,
	}
	return f10, &cinfo.COMPNAME.String, nil
}

// 获取个股快照
func getStockSnapshot(scode string) (*publish.REDIS_BIN_STOCK_SNAPSHOT, error) {

	i, err := strconv.Atoi(scode)
	if err != nil {
		logging.Error("String to int Error | %v", err)
	}
	key := fmt.Sprintf(REDISKEY_SECURITY_SNAP, i)

	bin, err := RedisStore.GetBytes(key)
	if err != nil {
		if err == hsgrr.ErrNil {
			logging.Warning("redis not found key: %v", key)
			return nil, err
		}
		return nil, err
	}

	data := publish.REDIS_BIN_STOCK_SNAPSHOT{}
	buffer := bytes.NewBuffer(bin)
	if err := binary.Read(buffer, binary.LittleEndian, &data); err != nil && err != io.EOF {
		logging.Error("binary decode error: %v", err)
		return nil, err
	}
	return &data, err
}