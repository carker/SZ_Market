// 个股详情
package publish

import (
	. "haina.com/market/hqpublish/controllers"
	"haina.com/market/hqpublish/models"
	"haina.com/market/hqpublish/models/publish"

	"ProtocolBuffer/projects/hqpublish/go/protocol"

	"haina.com/share/logging"

	"github.com/gin-gonic/gin"
)

type PerSDetail struct{}

func NewPerSDetail() *PerSDetail {
	return &PerSDetail{}
}

func (this *PerSDetail) POST(c *gin.Context) {
	replayfmt := c.Query(models.CONTEXT_FORMAT)
	if len(replayfmt) == 0 {
		replayfmt = "pb" // 默认
	}

	switch replayfmt {
	case "json":
		this.PostJson(c)
	case "pb":
		this.PostPB(c)
	default:
		return
	}
}

func (this *PerSDetail) PostJson(c *gin.Context) {
	var req protocol.RequestPerSDetail

	code, err := RecvAndUnmarshalJson(c, 1024, &req)
	if err != nil {
		logging.Error("post json %v", err)
		WriteJson(c, code, nil)
		return
	}
	logging.Info("Request %+v", req)
	if req.SID == 0 {
		WriteJson(c, 40004, nil)
		return
	}

	data, err := publish.NewPerSDetailM().GetPerSDtail(&req)

	if err != nil {
		logging.Error("%v", err)
		WriteJson(c, 40002, nil)
		return
	}
	WriteJson(c, 200, data)
}

func (this *PerSDetail) PostPB(c *gin.Context) {
	var req protocol.RequestPerSDetail
	code, err := RecvAndUnmarshalPB(c, 1024, &req)
	if err != nil {
		logging.Error("post pb %v", err)
		WriteDataErrCode(c, code)
		return
	}
	logging.Info("Request %+v", req)
	if req.SID == 0 {
		WriteDataErrCode(c, 40004)
		return
	}

	data, err := publish.NewPerSDetailM().GetPerSDtail(&req)
	if err != nil {
		logging.Error("%v", err)
		WriteDataErrCode(c, 40002)
		return
	}
	WriteDataPB(c, protocol.HAINA_PUBLISH_CMD_ACK_PERSDETAIL, data)
}
