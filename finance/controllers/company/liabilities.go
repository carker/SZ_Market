// 资产负债表
package company

import (
	"github.com/gin-gonic/gin"
	"haina.com/market/finance/models/company"
	"haina.com/share/lib"
	"haina.com/share/logging"
)

type LiabilitiesInfo struct {
}

func NewLiabilitiesInfo() *LiabilitiesInfo {
	return &LiabilitiesInfo{}
}

func (this *LiabilitiesInfo) getJson(c *gin.Context) (*company.ResponseInfo, error) {
	return company.NewLiabilities().GetJson(c)
}

func (this *LiabilitiesInfo) GET(c *gin.Context) {
	data, err := this.getJson(c)
	if err != nil {
		logging.Debug("%v", err)
		lib.WriteString(c, 40002, nil)
		return
	}

	lib.WriteString(c, 200, data)
}
