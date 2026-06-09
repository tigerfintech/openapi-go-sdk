package model

// Order 订单响应模型。
// 服务端响应字段为 camelCase，此结构体仅用于 unmarshal 查询类接口的返回。
// 下单/改单请使用 OrderRequest。
type Order struct {
	// 交易账户
	Account string `json:"account,omitempty"`
	// 全局订单 ID
	ID int64 `json:"id,omitempty"`
	// 账户自增订单号
	OrderId int64 `json:"orderId,omitempty"`
	// 买卖方向（BUY/SELL）
	Action string `json:"action,omitempty"`
	// 订单类型（MKT/LMT/STP/STP_LMT/TRAIL 等）
	OrderType string `json:"orderType,omitempty"`
	// 总数量（API 返回字段名为 totalQuantity）
	TotalQuantity int64 `json:"totalQuantity,omitempty"`
	// 限价
	LimitPrice float64 `json:"limitPrice,omitempty"`
	// 辅助价格（止损价）
	AuxPrice float64 `json:"auxPrice,omitempty"`
	// 跟踪止损百分比
	TrailingPercent float64 `json:"trailingPercent,omitempty"`
	// 订单状态
	Status string `json:"status,omitempty"`
	// 已成交数量（API 返回字段名为 filledQuantity）
	FilledQuantity int64 `json:"filledQuantity,omitempty"`
	// 平均成交价
	AvgFillPrice float64 `json:"avgFillPrice,omitempty"`
	// 有效期（DAY/GTC/OPG）
	TimeInForce string `json:"timeInForce,omitempty"`
	// 是否允许盘前盘后
	OutsideRth bool `json:"outsideRth,omitempty"`
	// 附加订单（止盈/止损）
	OrderLegs []OrderLeg `json:"orderLegs,omitempty"`
	// 算法参数
	AlgoParams *AlgoParams `json:"algoParams,omitempty"`
	// 股票代码
	Symbol string `json:"symbol,omitempty"`
	// 合约类型
	SecType string `json:"secType,omitempty"`
	// 市场
	Market string `json:"market,omitempty"`
	// 货币
	Currency string `json:"currency,omitempty"`
	// 到期日（期权/期货）
	Expiry string `json:"expiry,omitempty"`
	// 行权价（期权），API 返回为字符串
	Strike string `json:"strike,omitempty"`
	// 看涨/看跌（PUT/CALL），保持 API 原始名 right
	Right string `json:"right,omitempty"`
	// 合约标识符
	Identifier string `json:"identifier,omitempty"`
	// 合约名称
	Name string `json:"name,omitempty"`
	// 佣金
	Commission float64 `json:"commission,omitempty"`
	// 已实现盈亏
	RealizedPnl float64 `json:"realizedPnl,omitempty"`
	// 开仓时间（毫秒时间戳）
	OpenTime int64 `json:"openTime,omitempty"`
	// 更新时间（毫秒时间戳）
	UpdateTime int64 `json:"updateTime,omitempty"`
	// 最新时间（毫秒时间戳）
	LatestTime int64 `json:"latestTime,omitempty"`
	// 备注
	Remark string `json:"remark,omitempty"`
	// 订单来源
	Source string `json:"source,omitempty"`
	// 用户标记
	UserMark string `json:"userMark,omitempty"`
	// 服务端额外返回字段
	ExternalId          string   `json:"externalId,omitempty"`
	TotalQuantityScale  int      `json:"totalQuantityScale,omitempty"`
	FilledQuantityScale int      `json:"filledQuantityScale,omitempty"`
	FilledCashAmount    float64  `json:"filledCashAmount,omitempty"`
	Gst                 float64  `json:"gst,omitempty"`
	Liquidation         bool     `json:"liquidation,omitempty"`
	AttrDesc            string   `json:"attrDesc,omitempty"`
	AttrList            []string `json:"attrList,omitempty"`
	AlgoStrategy        string   `json:"algoStrategy,omitempty"`
	Discount            float64  `json:"discount,omitempty"`
	ReplaceStatus       string   `json:"replaceStatus,omitempty"`
	CancelStatus        string   `json:"cancelStatus,omitempty"`
	CanModify           bool     `json:"canModify,omitempty"`
	CanCancel           bool     `json:"canCancel,omitempty"`
	IsOpen              bool     `json:"isOpen,omitempty"`
	OrderDiscount       float64  `json:"orderDiscount,omitempty"`
	TradingSessionType  string   `json:"tradingSessionType,omitempty"`
	LatestPrice         float64  `json:"latestPrice,omitempty"`
}

// OrderLeg 附加订单（止盈/止损）- 响应模型
type OrderLeg struct {
	// 附加订单类型（PROFIT/LOSS）
	LegType string `json:"legType,omitempty"`
	// 价格
	Price float64 `json:"price,omitempty"`
	// 有效期
	TimeInForce string `json:"timeInForce,omitempty"`
	// 数量
	Quantity int64 `json:"quantity,omitempty"`
}

// AlgoParams 算法订单参数 - 响应模型
type AlgoParams struct {
	// 算法策略（TWAP/VWAP）
	AlgoStrategy string `json:"algoStrategy,omitempty"`
	// 开始时间
	StartTime string `json:"startTime,omitempty"`
	// 结束时间
	EndTime string `json:"endTime,omitempty"`
	// 参与率
	ParticipationRate float64 `json:"participationRate,omitempty"`
}

// OrderRequest 订单请求模型。
// 服务端请求体字段为 snake_case，此结构体仅用于 marshal 下单/改单/预览订单等接口。
// 查询返回请使用 Order。
type OrderRequest struct {
	// 交易账户
	Account string `json:"account,omitempty"`
	// 全局订单 ID（修改订单时必填）
	ID int64 `json:"id,omitempty"`
	// 账户自增订单号
	OrderId int64 `json:"order_id,omitempty"`
	// 买卖方向（BUY/SELL）
	Action string `json:"action,omitempty"`
	// 订单类型（MKT/LMT/STP/STP_LMT/TRAIL 等）
	OrderType string `json:"order_type,omitempty"`
	// 总数量
	TotalQuantity int64 `json:"total_quantity,omitempty"`
	// 限价
	LimitPrice float64 `json:"limit_price,omitempty"`
	// 辅助价格（止损价）
	AuxPrice float64 `json:"aux_price,omitempty"`
	// 跟踪止损百分比
	TrailingPercent float64 `json:"trailing_percent,omitempty"`
	// 有效期（DAY/GTC/OPG）
	TimeInForce string `json:"time_in_force,omitempty"`
	// 是否允许盘前盘后
	OutsideRth bool `json:"outside_rth,omitempty"`
	// 附加订单（止盈/止损）
	OrderLegs []OrderLegRequest `json:"order_legs,omitempty"`
	// 算法参数
	AlgoParams *AlgoParamsRequest `json:"algo_params,omitempty"`
	// 股票代码
	Symbol string `json:"symbol,omitempty"`
	// 合约类型
	SecType string `json:"sec_type,omitempty"`
	// 市场
	Market string `json:"market,omitempty"`
	// 货币
	Currency string `json:"currency,omitempty"`
	// 到期日（期权/期货）
	Expiry string `json:"expiry,omitempty"`
	// 行权价（期权）
	Strike string `json:"strike,omitempty"`
	// 看涨/看跌（PUT/CALL）
	Right string `json:"right,omitempty"`
	// 合约标识符
	Identifier string `json:"identifier,omitempty"`
	// 备注
	Remark string `json:"remark,omitempty"`
	// 用户标记
	UserMark string `json:"user_mark,omitempty"`
	// 机构账户鉴权 Secret Key
	SecretKey string `json:"secret_key,omitempty"`
}

// OrderLegRequest 附加订单请求模型（止盈/止损）
type OrderLegRequest struct {
	// 附加订单类型（PROFIT/LOSS）
	LegType string `json:"leg_type,omitempty"`
	// 价格
	Price float64 `json:"price,omitempty"`
	// 有效期
	TimeInForce string `json:"time_in_force,omitempty"`
	// 数量
	Quantity int64 `json:"quantity,omitempty"`
}

// AlgoParamsRequest 算法订单参数请求模型
type AlgoParamsRequest struct {
	// 算法策略（TWAP/VWAP）
	AlgoStrategy string `json:"algo_strategy,omitempty"`
	// 开始时间
	StartTime string `json:"start_time,omitempty"`
	// 结束时间
	EndTime string `json:"end_time,omitempty"`
	// 参与率
	ParticipationRate float64 `json:"participation_rate,omitempty"`
}
