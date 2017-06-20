package fcmysql

import (
	"haina.com/share/gocraft/dbr"
	"haina.com/share/logging"
	. "haina.com/share/models"
)

const (
	TABLE_TQ_SK_ANNOUNCEMT = "TQ_SK_ANNOUNCEMT" // 公告信息表
)

type TQ_SK_ANNOUNCEMT struct {
	Model       `db:"-" `
	DECLAREDATE int32          // 公告日期
	ANNTITLE    dbr.NullString // 公告标题
	ANNTEXT     dbr.NullString // 公告内容
	ANNTYPE     dbr.NullString // 全文类型
	LEVEL1      dbr.NullString // 一级分类

}

func NewTQ_SK_ANNOUNCEMT() *TQ_SK_ANNOUNCEMT {
	return &TQ_SK_ANNOUNCEMT{
		Model: Model{
			TableName: TABLE_TQ_SK_ANNOUNCEMT,
			Db:        MyCat,
		},
	}
}

func (this *TQ_SK_ANNOUNCEMT) GetNoticeInfo(ccode string) ([]*TQ_SK_ANNOUNCEMT, error) {
	var tsa []*TQ_SK_ANNOUNCEMT

	bulid := this.Db.Select("ANNTYPE,DECLAREDATE,ANNTITLE,ANNTEXT,LEVEL1").
		From(this.TableName).
		Where("COMPCODE='" + ccode + "'").
		Where("ISVALID=1").
		OrderBy("DECLAREDATE desc")

	_, err := this.SelectWhere(bulid, nil).LoadStructs(&tsa)

	if err != nil {
		logging.Debug("%v", err)
		return tsa, err
	}
	return tsa, err
}