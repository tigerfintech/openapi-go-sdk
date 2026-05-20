package model

// Contract 合约模型，字段与 Java SDK ContractItem 对齐。
// Go struct 字段用 PascalCase，json tag 用 API 原始 camelCase。
type Contract struct {
	// 合约 ID
	ContractId int64 `json:"contractId,omitempty"`
	// 合约标识符（期权 OCC 格式，如 AAPL  240119C00150000）
	Identifier string `json:"identifier,omitempty"`
	// 标的代码（如 AAPL）
	Symbol string `json:"symbol"`
	// 证券类型（STK/OPT/FUT/WAR/CASH/FUND 等）
	SecType string `json:"secType"`
	// 到期日（期权/期货）
	Expiry string `json:"expiry,omitempty"`
	// 合约月份（期货）
	ContractMonth string `json:"contractMonth,omitempty"`
	// 行权价（期权）
	Strike float64 `json:"strike,omitempty"`
	// 看涨/看跌（PUT/CALL）
	Right string `json:"right,omitempty"`
	// 合约乘数
	Multiplier float64 `json:"multiplier,omitempty"`
	// 每手数量
	LotSize float64 `json:"lotSize,omitempty"`
	// 交易所
	Exchange string `json:"exchange,omitempty"`
	// 市场
	Market string `json:"market,omitempty"`
	// 主交易所
	PrimaryExchange string `json:"primaryExchange,omitempty"`
	// 货币
	Currency string `json:"currency,omitempty"`
	// 本地代码
	LocalSymbol string `json:"localSymbol,omitempty"`
	// 交易类别
	TradingClass string `json:"tradingClass,omitempty"`
	// 合约名称
	Name string `json:"name,omitempty"`
	// 是否可交易
	Tradeable bool `json:"tradeable,omitempty"`
	// 是否仅可平仓
	CloseOnly bool `json:"closeOnly,omitempty"`
	// 最小报价变动
	MinTick float64 `json:"minTick,omitempty"`
	// 是否可融资
	Marginable bool `json:"marginable,omitempty"`
	// 做空初始保证金比例
	ShortInitialMargin float64 `json:"shortInitialMargin,omitempty"`
	// 做空维持保证金比例
	ShortMaintenanceMargin float64 `json:"shortMaintenanceMargin,omitempty"`
	// 做空费率
	ShortFeeRate float64 `json:"shortFeeRate,omitempty"`
	// 是否可做空
	Shortable bool `json:"shortable,omitempty"`
	// 可做空数量
	ShortableCount int64 `json:"shortableCount,omitempty"`
	// 做多初始保证金
	LongInitialMargin float64 `json:"longInitialMargin,omitempty"`
	// 做多维持保证金
	LongMaintenanceMargin float64 `json:"longMaintenanceMargin,omitempty"`
	// 最后交易日（期货）
	LastTradingDate string `json:"lastTradingDate,omitempty"`
	// 第一通知日（期货）
	FirstNoticeDate string `json:"firstNoticeDate,omitempty"`
	// 最后竞价截止时间（期货，毫秒时间戳）
	LastBiddingCloseTime int64 `json:"lastBiddingCloseTime,omitempty"`
	// 是否连续合约（期货）
	Continuous bool `json:"continuous,omitempty"`
	// 期货类型
	Type string `json:"type,omitempty"`
	// IB 代码（期货）
	IbCode string `json:"ibCode,omitempty"`
	// 最小报价单位价格区间
	TickSizes []TickSize `json:"tickSizes,omitempty"`
	// 是否 ETF
	IsEtf bool `json:"isEtf,omitempty"`
	// ETF 杠杆倍数
	EtfLeverage int `json:"etfLeverage,omitempty"`
	// 日内初始保证金折扣
	DiscountedDayInitialMargin float64 `json:"discountedDayInitialMargin,omitempty"`
	// 日内维持保证金折扣
	DiscountedDayMaintenanceMargin float64 `json:"discountedDayMaintenanceMargin,omitempty"`
	// 日内保证金折扣时区
	DiscountedTimeZoneCode string `json:"discountedTimeZoneCode,omitempty"`
	// 日内保证金折扣开始时间
	DiscountedStartAt string `json:"discountedStartAt,omitempty"`
	// 日内保证金折扣结束时间
	DiscountedEndAt string `json:"discountedEndAt,omitempty"`
	// 是否支持隔夜交易
	SupportOvernightTrading bool `json:"supportOvernightTrading,omitempty"`
	// 是否支持碎股交易
	SupportFractionalShare bool `json:"supportFractionalShare,omitempty"`
}

// TickSize 最小报价单位价格区间
type TickSize struct {
	// 起始价格（字符串，如 "0"、"1"）
	Begin string `json:"begin,omitempty"`
	// 结束价格（字符串，如 "1"、"Infinity"）
	End string `json:"end,omitempty"`
	// 最小报价单位
	TickSize float64 `json:"tickSize,omitempty"`
	// 类型（CLOSED/OPEN）
	Type string `json:"type,omitempty"`
}
