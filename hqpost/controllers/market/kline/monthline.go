package kline

import (
	pbk "ProtocolBuffer/format/kline"
	"fmt"

	"haina.com/share/store/redis"

	"github.com/golang/protobuf/proto"
	"haina.com/share/logging"
)

func (this *Security) MonthLine() {
	this.GetMonthDay()
	securitys := *this.list.Securitys

	for _, single := range securitys { //每支股票

		var tmps []StockSingle
		//PB
		var klist pbk.KInfoTable

		for _, month := range *single.MonthDays { //每个月

			tmp := StockSingle{}
			var mdata pbk.KInfo //pb类型

			var (
				i          int
				day        int32
				AvgPxTotal uint32
			)

			for i, day = range month { //每一天
				stockday := single.SigStock[day]
				if tmp.HighPx < stockday.HighPx || tmp.HighPx == 0 { //最高价
					tmp.HighPx = stockday.HighPx
				}
				if tmp.LowPx > stockday.LowPx || tmp.LowPx == 0 { //最低价
					tmp.LowPx = stockday.LowPx
				}
				tmp.Volume += stockday.Volume //成交量
				tmp.Value += stockday.Value   //成交额
				AvgPxTotal += stockday.AvgPx
			}
			tmp.SID = single.Sid
			tmp.Time = single.SigStock[month[0]].Time     //时间（取每周第一天）
			tmp.OpenPx = single.SigStock[month[0]].OpenPx //开盘价（每周第一天的开盘价）
			if len(tmps) > 0 {
				tmp.PreCPx = tmps[len(tmps)-1].LastPx //昨收价(上周的最新价)
			} else {
				tmp.PreCPx = 0
			}
			tmp.LastPx = single.SigStock[month[i]].LastPx //最新价
			tmp.AvgPx = AvgPxTotal / uint32(i+1)          //平均价

			tmps = append(tmps, tmp)
			//logging.Debug("yue线是:%v", tmps)
			//入PB
			mdata.NSID = tmp.SID
			mdata.NTime = tmp.Time
			mdata.NPreCPx = tmp.PreCPx
			mdata.NOpenPx = tmp.OpenPx
			mdata.NHighPx = tmp.HighPx
			mdata.NLowPx = tmp.LowPx
			mdata.NLastPx = tmp.LastPx
			mdata.LlVolume = tmp.Volume
			mdata.LlValue = tmp.Value
			mdata.NAvgPx = tmp.AvgPx

			klist.List = append(klist.List, &mdata)
		}
		//入PB 入redis
		data, err := proto.Marshal(&klist)
		if err != nil {
			logging.Error("Encode protocbuf of week Line error...%v", err.Error())
			return
		}

		key := fmt.Sprintf(REDISKEY_SECURITY_HMONTH, single.Sid)
		if err := redis.Set(key, data); err != nil {
			logging.Fatal("%v", err)
		}
	}

}

func (this *Security) GetMonthDay() {
	securitys := *this.list.Securitys

	for i, v := range securitys { // v: 单个股票
		var yesterday int32 = 0

		var dates [][]int32
		var month []int32
		for j, day := range v.Date { // v.Date: 单个股票的所有时间

			if j == 0 {
				month = append(month, day)
				yesterday = day / 100
				continue
			}
			if yesterday == day/100 {
				month = append(month, day)
			} else {
				dates = append(dates, month)
				month = nil
				month = append(month, day)
			}
			yesterday = day / 100

		}
		//logging.Debug("------day:%v", v.Date)
		//logging.Debug("-month---dates:%+v", dates)
		securitys[i].MonthDays = &dates
	}
}
