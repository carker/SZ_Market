// 融资融券
package publish2

import (
	"fmt"

	. "haina.com/market/hqpublish/models"

	"haina.com/market/hqpublish/models/finchina"
)

type SMTModel struct {
	Balance float64 `json:"balance"` // 融资余额
	Amount  float64 `json:"amount"`  // 融资余量
	Finbal  float64 `json:"finbal"`  // 融资融券余额
	FinDate int32   `json:"findate"` // 日期
}

type resSmt struct {
	Num  int        `json:"num"`
	MId  int32      `json:"marketId"`
	Smts []SMTModel `json:"slist"`
}

func GetSMTbyMarket(mid int32, ntype string) *resSmt {

	res := &resSmt{}
	key := fmt.Sprintf(REDIS_CACHE_CAPITAL_SMT, mid)
	if _, err := GetResFromCache(key, res); err == nil {
		return res
	}

	smts, err := finchina.NewTQ_SK_FINMRGNTRADE().GetSMTFromFC(60, ntype)
	if err != nil {
		return nil
	}

	var Smts []SMTModel
	for _, v := range smts {
		smt := SMTModel{
			Balance: v.FINBALANCE.Float64,
			Amount:  v.MRGNRESQTY.Float64,
			Finbal:  v.FINMRGHBAL.Float64,
			FinDate: int32(v.TRADEDATE.Int64),
		}
		Smts = append(Smts, smt)
	}
	res.MId = mid
	res.Num = len(Smts)
	res.Smts = Smts

	SetResToCache(key, res)
	return res
}
