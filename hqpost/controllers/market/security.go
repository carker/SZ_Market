package market

import (
	"time"

	"haina.com/market/hqpost/controllers/market/sidcode"

	"haina.com/market/hqpost/controllers/market/kline"

	"haina.com/market/hqpost/config"
	"haina.com/share/logging"
)

func Update(cfg *config.AppConfig) {
	/*********************开始时间************************/
	start := time.Now()

	//股票代码表
	codes, err := sidcode.GetSecurityTable()
	if err != nil {
		logging.Error("%v", err.Error())
		return
	}

	security := new(kline.Security)
	//--------------------------------------------------/
	//日K线
	security.DayLine(cfg, codes)

	//周K线
	security.WeekLine()

	//月线
	security.MonthLine()

	//年线
	security.YearLine()
	//--------------------------------------------------/
	end := time.Now()
	logging.Info("Update Kline historical data successed, and running time:%v", end.Sub(start))
	/*********************结束时间***********************/

}
