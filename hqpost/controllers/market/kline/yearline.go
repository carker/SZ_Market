package kline

import (
	"ProtocolBuffer/format/kline"

	"haina.com/market/hqpost/models"
	"haina.com/market/hqpost/models/filestore"
	"haina.com/market/hqpost/models/redistore"
	"haina.com/share/logging"
)

func (this *Security) YearLine() {
	this.GetYearDay()
	securitys := *this.list.Securitys

	rstore := redistore.NewHKLine(REDISKEY_SECURITY_HYEAR)
	for _, single := range securitys { //每支股票
		var (
			klist *kline.KInfoTable
			err   error
		)
		filepath, ok := filestore.CheckFileSoteDir(single.Sid, cfg.File.Path, cfg.File.Year)
		if !ok { //不存在，做第一次生成
			klist = produceYearline(&single)

			//1.入文件
			filestore.WiteHainaFileStore(filepath, klist)
			//2.redis做第一次生成
			for _, v := range klist.List {
				if err := rstore.LPushHisKLine(single.Sid, v); err != nil {
					logging.Error("%v", err.Error())
					return
				}
			}
		} else {
			if single.today != nil {
				if err = filestore.UpdateYearLineToFile(filepath, single.today); err != nil {
					logging.Error("%v", err.Error())
				}

				var ss []kline.KInfo
				if err = rstore.LRangeHisKLine(single.Sid, 1, &ss); err != nil {
					if err != models.ERROR_REDIS_LIST_NULL {
						logging.Error("%v", err.Error())
						return
					} else {
						continue
					}
				}
				latest := redistore.CompareKInfo(&ss[0], single.today)

				if err := rstore.LSetHisKLine(single.Sid, latest); err != nil {
					logging.Error("%v", err.Error())
					return
				}
			}
		}
	}

}

func produceYearline(single *SingleSecurity) *kline.KInfoTable {
	//PB
	var klist kline.KInfoTable

	for _, year := range *single.YearDays { //每年
		var (
			i          int
			day        int32
			AvgPxTotal uint32
			tmp        kline.KInfo //pb类型
		)

		for i, day = range year { //每一天
			stockday := single.SigStock[day]
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
		tmp.NSID = single.Sid
		tmp.NTime = single.SigStock[year[0]].NTime     //时间（取每周第一天）
		tmp.NOpenPx = single.SigStock[year[0]].NOpenPx //开盘价（每周第一天的开盘价）
		if len(klist.List) > 0 {
			tmp.NPreCPx = klist.List[len(klist.List)-1].NLastPx //昨收价(上周的最新价)
		} else {
			tmp.NPreCPx = 0
		}
		tmp.NLastPx = single.SigStock[year[i]].NLastPx //最新价
		tmp.NAvgPx = AvgPxTotal / uint32(i+1)          //平均价

		klist.List = append(klist.List, &tmp)
		//logging.Debug("year线是:%v", klist.List)
	}
	return &klist
}

func (this *Security) GetYearDay() {
	securitys := *this.list.Securitys

	for i, v := range securitys { // v: 单个股票
		var lastyear int32 = 0
		var dates [][]int32
		var years []int32

		if len(v.Date) < 1 {
			logging.Error("SID:%v---No historical data...", v.Sid)
			continue
		}
		for j, day := range v.Date { // v.Date: 单个股票的所有时间
			if lastyear == 0 {
				years = append(years, day)
				lastyear = day / 10000
				continue
			}
			if lastyear == day/10000 {
				years = append(years, day)
				if j == int(len(v.Date)-1) { //执行到最后一个
					dates = append(dates, years)
				}
			} else {
				dates = append(dates, years)
				years = nil
				years = append(years, day)
			}
			lastyear = day / 10000

		}
		securitys[i].YearDays = &dates
	}
}
