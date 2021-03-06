package fcmysql

import (
	"time"

	_ "github.com/go-sql-driver/mysql"
	"haina.com/share/gocraft/dbr"
	. "haina.com/share/models"
)

// 数据对象名称：TQ_SK_BASICINFO    中文名称：股票基本信息表

type TQ_SK_BASICINFO struct {
	Model      `db:"-"`
	LISTDATE   dbr.NullString `db:"LISTDATE"`   // 上市日期
	DELISTDATE dbr.NullString `db:"DELISTDATE"` // 退市日期
	SYMBOL     dbr.NullString `db:"SYMBOL"`     // 证券代码
}

func NewTQ_SK_BASICINFO() *TQ_SK_BASICINFO {
	return &TQ_SK_BASICINFO{
		Model: Model{
			TableName: TABLE_TQ_SK_BASICINFO,
			Db:        MyCat,
		},
	}
}

// 查询证券信息
func (this *TQ_SK_BASICINFO) GetBasicinfoList(symb string) (TQ_SK_BASICINFO, error) {
	var tsb TQ_SK_BASICINFO
	err := this.Db.Select("LISTDATE,DELISTDATE").From("TQ_SK_BASICINFO").
		Where("SYMBOL='" + symb + "' and  ISVALID=1").
		Limit(1).
		LoadStruct(&tsb)
	return tsb, err
}

// 查询每天新股
func (this *TQ_SK_BASICINFO) GetNewBasicinfo() ([]TQ_SK_BASICINFO, error) {
	var nbsi []TQ_SK_BASICINFO
	// 获取当前日期
	timed := time.Now().Format("20060102")
	err := this.Db.Select("LISTDATE,SYMBOL").
		From(this.TableName).
		Where("LISTDATE='" + timed + "'").
		Where("SETYPE='101'").
		Where("ISVALID=1").
		LoadStruct(&nbsi)
	return nbsi, err
}
