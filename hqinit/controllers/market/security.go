package market

import (
	"time"

	"haina.com/market/hqinit/config"

	"haina.com/market/hqinit/controllers/market/security"
	"haina.com/share/logging"
)

func Update(cfg *config.AppConfig) {
	/*********************开始时间************************/
	start := time.Now()

	//股票代码表
	security.UpdateSecurityCodeTable()

	//市场代码表及证券基本数据
	security.UpdateSecurityTable(cfg)

	//指数基本数据
	security.UpdateIndexTable(cfg)

	// 板块信息
	security.UpdateStockBlockSet(cfg)

	// 指数成分股
	security.UpdateIndexComponent(cfg)

	//证券静态数据
	security.UpdateSecurityStaticInfo(cfg)

	end := time.Now()
	logging.Info("Update Kline historical data successed, and running time:%v", end.Sub(start))

	/*********************结束时间***********************/

}
