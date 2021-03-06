package finchina

const (
	TABLE_TQ_COMP_INDUSTRY    = "TQ_COMP_INDUSTRY"    //行业分类表
	TABLE_TQ_COMP_INFO        = "TQ_COMP_INFO"        //机构资料表（公司信息）
	TABLE_TQ_COMP_MANAGER     = "TQ_COMP_MANAGER"     //公司高管表
	TABLE_TQ_COMP_SKHOLDERCHG = "TQ_COMP_SKHOLDERCHG" //高管和关联人持股变动情况表
	TABLE_TQ_OA_STCODE        = "TQ_OA_STCODE"        //证券内码表
	TABLE_TQ_SK_DIVIDENTS     = "TQ_SK_DIVIDENTS"     //分红情况表
	TABLE_TQ_SK_PROADDISS     = "TQ_SK_PROADDISS"     //上市公司增发情况表
	TABLE_TQ_SK_PROPLACING    = "TQ_SK_PROPLACING"    //上市公司配股情况表
	TABLE_TQ_SK_SHAREHDCHG    = "TQ_SK_SHAREHDCHG"
	TABLE_TQ_SK_BASICINFO     = "TQ_SK_BASICINFO" //股票基本信息表
)

// market SCHEMA
const (
	TABLE_TQ_SK_SHAREHOLDERNUM = "TQ_SK_SHAREHOLDERNUM" // 股东户数统计表
	TABLE_TQ_SK_OTSHOLDER      = "TQ_SK_OTSHOLDER"      // 流通股东信息表
	TABLE_TQ_SK_SHAREHOLDER    = "TQ_SK_SHAREHOLDER"    // 股东名单信息表
	TABLE_TQ_SK_SHARESTRUCHG   = "TQ_SK_SHARESTRUCHG"   // 股本结构变化
	TABLE_TQ_SK_LCPERSON       = "TQ_SK_LCPERSON"       // 上市公司董事名单
	TABLE_TQ_SK_IINVHOLDCHG    = "TQ_SK_IINVHOLDCHG"    //机构持股明细表
)

//--------------------------------------------------------------------------------
// 财务分析
const ( // 利润表
	TABLE_TQ_FIN_PROINCSTATEMENTNEW  = "TQ_FIN_PROINCSTATEMENTNEW"  //一般企业利润表(新准则产品表)
	TABLE_TQ_FIN_PROBINCSTATEMENTNEW = "TQ_FIN_PROBINCSTATEMENTNEW" //银行利润表(新准则产品表)
	TABLE_TQ_FIN_PROIINCSTATEMENTNEW = "TQ_FIN_PROIINCSTATEMENTNEW" //保险利润表(新准则产品表)
	TABLE_TQ_FIN_PROSINCSTATEMENTNEW = "TQ_FIN_PROSINCSTATEMENTNEW" //证券利润表(新准则产品表)
)

const ( // 现金流量表
	TABLE_TQ_FIN_PROCFSTATEMENTNEW  = "TQ_FIN_PROCFSTATEMENTNEW"  //一般企业现金流量表(新准则产品表)
	TABLE_TQ_FIN_PROBCFSTATEMENTNEW = "TQ_FIN_PROBCFSTATEMENTNEW" //银行现金流量表(新准则产品表)
	TABLE_TQ_FIN_PROICFSTATEMENTNEW = "TQ_FIN_PROICFSTATEMENTNEW" //保险现金流量表(新准则产品表)
	TABLE_TQ_FIN_PROSCFSTATEMENTNEW = "TQ_FIN_PROSCFSTATEMENTNEW" //证券现金流量表(新准则产品表)
)

const ( // 资产负债表
	TABLE_TQ_FIN_PROBALSHEETNEW   = "TQ_FIN_PROBALSHEETNEW"   //一般企业资产负债表(新准则产品表)
	TABLE_TQ_FIN_PROBBALBSHEETNEW = "TQ_FIN_PROBBALBSHEETNEW" //银行资产负债表(新准则产品表)
	TABLE_TQ_FIN_PROIBALSHEETNEW  = "TQ_FIN_PROIBALSHEETNEW"  //保险资产负债表(新准则产品表)
	TABLE_TQ_FIN_PROSBALSHEETNEW  = "TQ_FIN_PROSBALSHEETNEW"  //证券资产负债表(新准则产品表)
)

const ( // 关键指标
	TABLE_TQ_FIN_PROFINMAININDEX  = "TQ_FIN_PROFINMAININDEX"  //主要财务指标（产品表）
	TABLE_TQ_FIN_PROINDICDATA     = "TQ_FIN_PROINDICDATA"     //衍生财务指标（产品表）
	TABLE_TQ_FIN_PROTTMINDIC      = "TQ_FIN_PROTTMINDIC"      //财务数据_TTM指标（产品表）
	TABLE_TQ_FIN_PROCFSTTMSUBJECT = "TQ_FIN_PROCFSTTMSUBJECT" //TTM现金科目产品表
)

//--------------------------------------------------------------------------------
const (
	REDIS_TTL = 60 * 60 * 24
)
const (
	REDIS_SYMBOL_COMPCODE = "finchina:symbol:%s:compcode"
)
