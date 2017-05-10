package company

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"haina.com/market/finance/models"
	"haina.com/market/finance/models/company"
	"haina.com/share/lib"
	"haina.com/share/logging"
)

type EquityInfo struct {
}

func NewEquityInfo() *EquityInfo {
	return &EquityInfo{}
}

/**
获取股东人数信息
*/
func (this *EquityInfo) GetShareholderJson(c *gin.Context) {

	enddate := c.Query(models.CONTEXT_END_DATE)
	count := c.Query(models.CONTEXT_COUNT)
	scode := strings.Split(c.Query(models.CONTEXT_SCODE), ".")[0]
	scodePrefix, market, err := ParseSCode(c.Query(models.CONTEXT_SCODE))

	if err != nil {
		lib.WriteString(c, 40004, "")
		return
	}
	var value_int = 0
	var err error
	if count != "" {
		value_int, err = strconv.Atoi(count)
		logging.Info("类型转换%v", err)
	} else {
		value_int = 10
	}

	//根据条件查询股东信息
	var strDate = ""
	if enddate != "" {
		strDate = " and ENDDATE<'" + enddate + "'"
	}
	data, err := company.GetShareholderGroup(scode, value_int, strDate, market)
	lib.WriteString(c, 200, data)
}

/**
获取十大流通股东信息
*/
func (this *EquityInfo) GetTop10Json(c *gin.Context) {

	enddate := c.Query(models.CONTEXT_END_DATE)
	count := c.Query(models.CONTEXT_COUNT)
	scode := strings.Split(c.Query(models.CONTEXT_SCODE), ".")[0]
	scodePrefix, market, err := ParseSCode(c.Query(models.CONTEXT_SCODE))

	if err != nil {
		lib.WriteString(c, 40004, "")
		return
	}
	var value_int = 0
	var err error
	if count == "" {
		value_int = 10
	} else {
		value_int, err = strconv.Atoi(count)
	}

	if err != nil {
		logging.Debug("%v", err)
		lib.WriteString(c, 88888, nil)
	}

	data, err := company.GetTop10Group(enddate, scode, value_int, market)

	lib.WriteString(c, 200, data)
}

/**
获取机构持股信息
*/
func (this *EquityInfo) GetOrganizationJson(c *gin.Context) {
	scode := strings.Split(c.Query(models.CONTEXT_SCODE), ".")[0]
	scodePrefix, market, err := ParseSCode(c.Query(models.CONTEXT_SCODE))

	if err != nil {
		lib.WriteString(c, 40004, "")
		return
	}
	data, err := company.GetCompGroup(scode, market)
	if err != nil {
		lib.WriteString(c, 300, err.Error())
		return
	}
	lib.WriteString(c, 200, data)
}
