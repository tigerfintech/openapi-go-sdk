// Package model 定义老虎证券 OpenAPI 的数据模型和枚举类型。
package model

// Market 市场枚举
type Market string

const (
	MarketAll Market = "ALL"
	MarketUS  Market = "US"
	MarketHK  Market = "HK"
	MarketCN  Market = "CN"
	MarketSG  Market = "SG"
)

// SecurityType 证券类型枚举
type SecurityType string

const (
	SecTypeAll  SecurityType = "ALL"
	SecTypeSTK  SecurityType = "STK"
	SecTypeOPT  SecurityType = "OPT"
	SecTypeWAR  SecurityType = "WAR"
	SecTypeIOPT SecurityType = "IOPT"
	SecTypeFUT  SecurityType = "FUT"
	SecTypeFOP  SecurityType = "FOP"
	SecTypeCASH SecurityType = "CASH"
	SecTypeMLEG SecurityType = "MLEG"
	SecTypeFUND SecurityType = "FUND"
)

// Currency 货币枚举
type Currency string

const (
	CurrencyAll Currency = "ALL"
	CurrencyUSD Currency = "USD"
	CurrencyHKD Currency = "HKD"
	CurrencyCNH Currency = "CNH"
	CurrencySGD Currency = "SGD"
)

// OrderType 订单类型枚举
type OrderType string

const (
	OrderTypeMKT    OrderType = "MKT"
	OrderTypeLMT    OrderType = "LMT"
	OrderTypeSTP    OrderType = "STP"
	OrderTypeSTPLMT OrderType = "STP_LMT"
	OrderTypeTRAIL  OrderType = "TRAIL"
	OrderTypeAM     OrderType = "AM"
	OrderTypeAL     OrderType = "AL"
	OrderTypeTWAP    OrderType = "TWAP"
	OrderTypeVWAP    OrderType = "VWAP"
	OrderTypeOCA     OrderType = "OCA"
	OrderTypeICEBERG OrderType = "ICEBERG"
)

// PriceType 冰山单价格类型
type PriceType string

const (
	// PriceTypeLimitPrice 限价
	PriceTypeLimitPrice PriceType = "LIMIT_PRICE"
	// PriceTypeAskPrice 卖一价
	PriceTypeAskPrice PriceType = "ASK_PRICE"
	// PriceTypeBidPrice 买一价
	PriceTypeBidPrice PriceType = "BID_PRICE"
	// PriceTypeLatestPrice 最新价
	PriceTypeLatestPrice PriceType = "LATEST_PRICE"
)

// OrderStatus 订单状态枚举。对齐 Java SDK `OrderStatus` 定义。
// Code 值与服务端 wire 数字一致；字符串值与服务端返回的字符串 status 一致。
type OrderStatus string

const (
	OrderStatusInvalid       OrderStatus = "Invalid"       // -2
	OrderStatusInitial       OrderStatus = "Initial"       // -1
	OrderStatusPendingCancel OrderStatus = "PendingCancel" //  3
	OrderStatusCancelled     OrderStatus = "Cancelled"     //  4
	OrderStatusSubmitted     OrderStatus = "Submitted"     //  5
	OrderStatusFilled        OrderStatus = "Filled"        //  6
	OrderStatusInactive      OrderStatus = "Inactive"      //  7
	OrderStatusPendingSubmit OrderStatus = "PendingSubmit" //  8
)

// Code 返回订单状态对应的服务端数字码,未知状态返回 0。
func (s OrderStatus) Code() int {
	switch s {
	case OrderStatusInvalid:
		return -2
	case OrderStatusInitial:
		return -1
	case OrderStatusPendingCancel:
		return 3
	case OrderStatusCancelled:
		return 4
	case OrderStatusSubmitted:
		return 5
	case OrderStatusFilled:
		return 6
	case OrderStatusInactive:
		return 7
	case OrderStatusPendingSubmit:
		return 8
	}
	return 0
}

// BarPeriod K 线周期枚举
type BarPeriod string

const (
	BarPeriodDay   BarPeriod = "day"
	BarPeriodWeek  BarPeriod = "week"
	BarPeriodMonth BarPeriod = "month"
	BarPeriodYear  BarPeriod = "year"
	BarPeriod1Min  BarPeriod = "1min"
	BarPeriod5Min  BarPeriod = "5min"
	BarPeriod15Min BarPeriod = "15min"
	BarPeriod30Min BarPeriod = "30min"
	BarPeriod60Min BarPeriod = "60min"
)

// Language 语言枚举
type Language string

const (
	LanguageZhCN Language = "zh_CN"
	LanguageZhTW Language = "zh_TW"
	LanguageEnUS Language = "en_US"
)

// QuoteRight 复权类型枚举
type QuoteRight string

const (
	QuoteRightBr QuoteRight = "br" // 前复权
	QuoteRightNr QuoteRight = "nr" // 不复权
)

// License 牌照类型枚举
type License string

const (
	LicenseTBNZ License = "TBNZ"
	LicenseTBSG License = "TBSG"
	LicenseTBHK License = "TBHK"
	LicenseTBAU License = "TBAU"
	LicenseTBMS License = "TBMS"
	LicenseTBUS License = "TBUS"
)

// TimeInForce 订单有效期枚举
type TimeInForce string

const (
	TimeInForceDAY TimeInForce = "DAY"
	TimeInForceGTC TimeInForce = "GTC"
	TimeInForceGTD TimeInForce = "GTD"
	TimeInForceOPG TimeInForce = "OPG"
)

// OrderSortBy 订单排序字段
type OrderSortBy string

const (
	OrderSortByLatestCreated       OrderSortBy = "LATEST_CREATED"
	OrderSortByLatestStatusUpdated OrderSortBy = "LATEST_STATUS_UPDATED"
)

// SegmentType 账户分部类型
type SegmentType string

const (
	SegmentTypeAll  SegmentType = "ALL"
	SegmentTypeSec  SegmentType = "SEC"
	SegmentTypeFut  SegmentType = "FUT"
	SegmentTypeFund SegmentType = "FUND"
)

// CorporateActionType 公司行动类型
type CorporateActionType string

const (
	CorporateActionTypeSplit        CorporateActionType = "split"
	CorporateActionTypeDividend     CorporateActionType = "dividend"
	CorporateActionTypeEarning      CorporateActionType = "earning"
	CorporateActionTypeSymbolChange CorporateActionType = "symbol_change"
	CorporateActionTypeDelisting    CorporateActionType = "delisting"
	CorporateActionTypeIPO          CorporateActionType = "ipo"
)

// IndustryLevel 行业级别（1~4 级）
type IndustryLevel string

const (
	IndustryLevelGSector IndustryLevel = "GSECTOR"
	IndustryLevelGGroup  IndustryLevel = "GGROUP"
	IndustryLevelGInd    IndustryLevel = "GIND"
	IndustryLevelGSubInd IndustryLevel = "GSUBIND"
)

// SortDirection 排序方向
type SortDirection string

const (
	SortDirectionNo      SortDirection = "SortDir_No"
	SortDirectionAscend  SortDirection = "SortDir_Ascend"
	SortDirectionDescend SortDirection = "SortDir_Descend"
)

// OptionAnalysisPeriod 期权分析周期
type OptionAnalysisPeriod string

const (
	OptionAnalysisPeriodThreeYear     OptionAnalysisPeriod = "3year"
	OptionAnalysisPeriodFiftyTwoWeek  OptionAnalysisPeriod = "52week"
	OptionAnalysisPeriodTwentySixWeek OptionAnalysisPeriod = "26week"
	OptionAnalysisPeriodThirteenWeek  OptionAnalysisPeriod = "13week"
)

// FinancialReportPeriod 财报类型
type FinancialReportPeriod string

const (
	FinancialReportPeriodAnnual    FinancialReportPeriod = "Annual"
	FinancialReportPeriodQuarterly FinancialReportPeriod = "Quarterly"
	FinancialReportPeriodLTM       FinancialReportPeriod = "LTM"
)
