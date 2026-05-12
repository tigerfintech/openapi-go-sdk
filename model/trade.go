package model

// Asset 账户资产条目(来自 /assets)。
type Asset struct {
	Account        string         `json:"account,omitempty"`
	Capability     string         `json:"capability,omitempty"`
	Currency       string         `json:"currency,omitempty"`
	BuyingPower    float64        `json:"buyingPower,omitempty"`
	CashValue      float64        `json:"cashValue,omitempty"`
	NetLiquidation float64        `json:"netLiquidation,omitempty"`
	RealizedPnL    float64        `json:"realizedPnL,omitempty"`
	UnrealizedPnL  float64        `json:"unrealizedPnL,omitempty"`
	Segments       []AssetSegment `json:"segments,omitempty"`
}

// AssetSegment 资产分段(securities / commodities 等)。
type AssetSegment struct {
	Account            string  `json:"account,omitempty"`
	Category           string  `json:"category,omitempty"`
	Title              string  `json:"title,omitempty"`
	NetLiquidation     float64 `json:"netLiquidation,omitempty"`
	CashValue          float64 `json:"cashValue,omitempty"`
	AvailableFunds     float64 `json:"availableFunds,omitempty"`
	EquityWithLoan     float64 `json:"equityWithLoan,omitempty"`
	ExcessLiquidity    float64 `json:"excessLiquidity,omitempty"`
	AccruedCash        float64 `json:"accruedCash,omitempty"`
	AccruedDividend    float64 `json:"accruedDividend,omitempty"`
	InitMarginReq      float64 `json:"initMarginReq,omitempty"`
	MaintMarginReq     float64 `json:"maintMarginReq,omitempty"`
	GrossPositionValue float64 `json:"grossPositionValue,omitempty"`
	Leverage           float64 `json:"leverage,omitempty"`
}

// PrimeAsset 综合账户资产(来自 /prime_assets)。
type PrimeAsset struct {
	AccountID       string              `json:"accountId,omitempty"`
	UpdateTimestamp int64               `json:"updateTimestamp,omitempty"`
	Segments        []PrimeAssetSegment `json:"segments,omitempty"`
}

// PrimeAssetSegment 综合账户分段。
type PrimeAssetSegment struct {
	Capability                string          `json:"capability,omitempty"`
	Category                  string          `json:"category,omitempty"`
	Currency                  string          `json:"currency,omitempty"`
	CashBalance               float64         `json:"cashBalance,omitempty"`
	CashAvailableForTrade     float64         `json:"cashAvailableForTrade,omitempty"`
	GrossPositionValue        float64         `json:"grossPositionValue,omitempty"`
	EquityWithLoan            float64         `json:"equityWithLoan,omitempty"`
	NetLiquidation            float64         `json:"netLiquidation,omitempty"`
	InitMargin                float64         `json:"initMargin,omitempty"`
	MaintainMargin            float64         `json:"maintainMargin,omitempty"`
	OvernightMargin           float64         `json:"overnightMargin,omitempty"`
	UnrealizedPL              float64         `json:"unrealizedPL,omitempty"`
	UnrealizedPLByCostOfCarry float64         `json:"unrealizedPLByCostOfCarry,omitempty"`
	RealizedPL                float64         `json:"realizedPL,omitempty"`
	TotalTodayPL              float64         `json:"totalTodayPL,omitempty"`
	ExcessLiquidation         float64         `json:"excessLiquidation,omitempty"`
	OvernightLiquidation      float64         `json:"overnightLiquidation,omitempty"`
	BuyingPower               float64         `json:"buyingPower,omitempty"`
	LockedFunds               float64         `json:"lockedFunds,omitempty"`
	Leverage                  float64         `json:"leverage,omitempty"`
	Uncollected               float64         `json:"uncollected,omitempty"`
	CurrencyAssets            []CurrencyAsset `json:"currencyAssets,omitempty"`
	ConsolidatedSegTypes      []string        `json:"consolidatedSegTypes,omitempty"`
}

// CurrencyAsset 分币种资产明细。
type CurrencyAsset struct {
	Currency              string  `json:"currency,omitempty"`
	CashBalance           float64 `json:"cashBalance,omitempty"`
	CashAvailableForTrade float64 `json:"cashAvailableForTrade,omitempty"`
	ForexRate             float64 `json:"forexRate,omitempty"`
}

// PreviewResult 订单预览结果。
type PreviewResult struct {
	Account              string  `json:"account,omitempty"`
	IsPass               bool    `json:"isPass"`
	Commission           float64 `json:"commission,omitempty"`
	CommissionCurrency   string  `json:"commissionCurrency,omitempty"`
	MarginCurrency       string  `json:"marginCurrency,omitempty"`
	InitMargin           float64 `json:"initMargin,omitempty"`
	InitMarginBefore     float64 `json:"initMarginBefore,omitempty"`
	MaintMargin          float64 `json:"maintMargin,omitempty"`
	MaintMarginBefore    float64 `json:"maintMarginBefore,omitempty"`
	EquityWithLoan       float64 `json:"equityWithLoan,omitempty"`
	EquityWithLoanBefore float64 `json:"equityWithLoanBefore,omitempty"`
	AvailableEE          float64 `json:"availableEE,omitempty"`
	ExcessLiquidity      float64 `json:"excessLiquidity,omitempty"`
	OvernightLiquidation float64 `json:"overnightLiquidation,omitempty"`
	Gst                  float64 `json:"gst,omitempty"`
	Message              string  `json:"message,omitempty"`
}

// PlaceOrderResult 下单返回结果。
type PlaceOrderResult struct {
	ID      int64   `json:"id"`
	OrderID int64   `json:"order_id,omitempty"`
	SubIDs  []int64 `json:"subIds,omitempty"`
	Orders  []Order `json:"orders,omitempty"`
}

// OrderIDResult 仅含订单 ID 的响应(ModifyOrder/CancelOrder)。
type OrderIDResult struct {
	ID int64 `json:"id"`
}

// Transaction 成交记录。
type Transaction struct {
	ID                  int64   `json:"id,omitempty"`
	OrderID             int64   `json:"orderId,omitempty"`
	AccountId           int64   `json:"accountId,omitempty"`
	Account             string  `json:"account,omitempty"`
	Symbol              string  `json:"symbol,omitempty"`
	SecType             string  `json:"secType,omitempty"`
	Market              string  `json:"market,omitempty"`
	Currency            string  `json:"currency,omitempty"`
	Identifier          string  `json:"identifier,omitempty"`
	Action              string  `json:"action,omitempty"`
	// Price 委托价
	Price               float64 `json:"price,omitempty"`
	// FilledPrice 成交价（服务端字段名 filledPrice）
	FilledPrice         float64 `json:"filledPrice,omitempty"`
	Quantity            int64   `json:"quantity,omitempty"`
	FilledQuantity      int64   `json:"filledQuantity,omitempty"`
	FilledQuantityScale int     `json:"filledQuantityScale,omitempty"`
	// Amount 委托金额
	Amount              float64 `json:"amount,omitempty"`
	// FilledAmount 成交金额（服务端字段名 filledAmount）
	FilledAmount        float64 `json:"filledAmount,omitempty"`
	Commission          float64 `json:"commission,omitempty"`
	// TransactedAt 成交时间字符串，格式 "YYYY-MM-DD HH:MM:SS"（服务端返回字符串非时间戳）
	TransactedAt        string  `json:"transactedAt,omitempty"`
	// TransactionTime 成交时间毫秒时间戳
	TransactionTime     int64   `json:"transactionTime,omitempty"`
	// Time 兼容旧字段
	Time                int64   `json:"time,omitempty"`
}

// ManagedAccount 机构子账户信息（来自 /accounts）。
type ManagedAccount struct {
	Account     string `json:"account,omitempty"`
	AccountType string `json:"accountType,omitempty"`
	Capability  string `json:"capability,omitempty"`
	Status      string `json:"status,omitempty"`
}

// AnalyticsAsset 资产分析（按日）条目。
type AnalyticsAsset struct {
	Date          string  `json:"date,omitempty"`
	HoldingValue  float64 `json:"holdingValue,omitempty"`
	CashBalance   float64 `json:"cashBalance,omitempty"`
	Pnl           float64 `json:"pnl,omitempty"`
	PnlRate       float64 `json:"pnlRate,omitempty"`
	NetValueIndex float64 `json:"netValueIndex,omitempty"`
	Currency      string  `json:"currency,omitempty"`
	SegType       string  `json:"segType,omitempty"`
}

// AggregateAssets 综合账户总览。
type AggregateAssets struct {
	AccountID          string          `json:"accountId,omitempty"`
	NetLiquidation     float64         `json:"netLiquidation,omitempty"`
	GrossPositionValue float64         `json:"grossPositionValue,omitempty"`
	CashBalance        float64         `json:"cashBalance,omitempty"`
	BaseCurrency       string          `json:"baseCurrency,omitempty"`
	CurrencyAssets     []CurrencyAsset `json:"currencyAssets,omitempty"`
}

// SegmentFundAvailableItem 可调拨资金条目（segment_fund_available 响应）。
type SegmentFundAvailableItem struct {
	FromSegment string  `json:"fromSegment,omitempty"`
	Currency    string  `json:"currency,omitempty"`
	Amount      float64 `json:"amount,omitempty"`
}

// SegmentFund 子账户资金调拨响应（transfer/cancel 共用）。
type SegmentFund struct {
	ID          interface{} `json:"id,omitempty"`
	FromSegment string      `json:"fromSegment,omitempty"`
	ToSegment   string      `json:"toSegment,omitempty"`
	Currency    string      `json:"currency,omitempty"`
	Amount      float64     `json:"amount,omitempty"`
	Status      string      `json:"status,omitempty"`
	StatusDesc  string      `json:"statusDesc,omitempty"`
	Message     string      `json:"message,omitempty"`
	SettledAt   int64       `json:"settledAt,omitempty"`
	CreatedAt   int64       `json:"createdAt,omitempty"`
	UpdatedAt   int64       `json:"updatedAt,omitempty"`
}

// SegmentFundHistoryItem 资金调拨历史条目。
type SegmentFundHistoryItem struct {
	ID          int64   `json:"id,omitempty"`
	FromSegment string  `json:"fromSegment,omitempty"`
	ToSegment   string  `json:"toSegment,omitempty"`
	Currency    string  `json:"currency,omitempty"`
	Amount      float64 `json:"amount,omitempty"`
	Status      string  `json:"status,omitempty"`
	StatusDesc  string  `json:"statusDesc,omitempty"`
	SettledAt   int64   `json:"settledAt,omitempty"`
	CreatedAt   int64   `json:"createdAt,omitempty"`
	UpdatedAt   int64   `json:"updatedAt,omitempty"`
}

// FundDetails 资金明细条目（/fund_details）。
type FundDetails struct {
	ID         int64   `json:"id,omitempty"`
	Account    string  `json:"account,omitempty"`
	SegType    string  `json:"segType,omitempty"`
	FundType   string  `json:"fundType,omitempty"`
	Currency   string  `json:"currency,omitempty"`
	Amount     float64 `json:"amount,omitempty"`
	Balance    float64 `json:"balance,omitempty"`
	OccurTime  int64   `json:"occurTime,omitempty"`
	Remark     string  `json:"remark,omitempty"`
	ExternalID string  `json:"externalId,omitempty"`
}

// FundingHistoryItem 调拨记录（/transfer_fund）。
type FundingHistoryItem struct {
	ID              int64   `json:"id,omitempty"`
	RefID           string  `json:"refId,omitempty"`
	Type            int     `json:"type,omitempty"`
	TypeDesc        string  `json:"typeDesc,omitempty"`
	Currency        string  `json:"currency,omitempty"`
	Amount          float64 `json:"amount,omitempty"`
	BusinessDate    string  `json:"businessDate,omitempty"`
	Status          string  `json:"status,omitempty"`
	StatusDesc      string  `json:"statusDesc,omitempty"`
	CompletedStatus bool    `json:"completedStatus,omitempty"`
	CreatedAt       int64   `json:"createdAt,omitempty"`
	UpdatedAt       int64   `json:"updatedAt,omitempty"`
}

// EstimateTradableQuantity 可交易数量估算结果。
type EstimateTradableQuantity struct {
	TradableQuantity      float64 `json:"tradableQuantity,omitempty"`
	MaxCashBuyQuantity    float64 `json:"maxCashBuyQuantity,omitempty"`
	MaxMarginBuyQuantity  float64 `json:"maxMarginBuyQuantity,omitempty"`
	MaxShortSellQuantity  float64 `json:"maxShortSellQuantity,omitempty"`
	MaxPositionSellQty    float64 `json:"maxPositionSellQuantity,omitempty"`
	CashBuyingPower       float64 `json:"cashBuyingPower,omitempty"`
	Currency              string  `json:"currency,omitempty"`
}

// ForexOrderResult 外汇下单返回结果。
type ForexOrderResult struct {
	ID             string  `json:"id,omitempty"`
	Status         string  `json:"status,omitempty"`
	SourceCurrency string  `json:"sourceCurrency,omitempty"`
	TargetCurrency string  `json:"targetCurrency,omitempty"`
	SourceAmount   float64 `json:"sourceAmount,omitempty"`
	TargetAmount   float64 `json:"targetAmount,omitempty"`
	Rate           float64 `json:"rate,omitempty"`
	SubmitTime     int64   `json:"submitTime,omitempty"`
}

// TransferItem 内部转股单项。
type TransferItem struct {
	Symbol   string `json:"symbol,omitempty"`
	Quantity int64  `json:"quantity,omitempty"`
	Expiry   string `json:"expiry,omitempty"`
	Strike   string `json:"strike,omitempty"`
	Right    string `json:"right,omitempty"`
	SecType  string `json:"sec_type,omitempty"`
}

// PositionTransferRecord 内部转股记录。
type PositionTransferRecord struct {
	ID          string         `json:"id,omitempty"`
	FromAccount string         `json:"fromAccount,omitempty"`
	ToAccount   string         `json:"toAccount,omitempty"`
	Market      string         `json:"market,omitempty"`
	Status      string         `json:"status,omitempty"`
	SubmitTime  int64          `json:"submitTime,omitempty"`
	Transfers   []TransferItem `json:"transfers,omitempty"`
}

// PositionTransferDetail 内部转股详情。
type PositionTransferDetail struct {
	ID          string         `json:"id,omitempty"`
	FromAccount string         `json:"fromAccount,omitempty"`
	ToAccount   string         `json:"toAccount,omitempty"`
	Market      string         `json:"market,omitempty"`
	Status      string         `json:"status,omitempty"`
	SubmitTime  int64          `json:"submitTime,omitempty"`
	UpdateTime  int64          `json:"updateTime,omitempty"`
	Transfers   []TransferItem `json:"transfers,omitempty"`
	Remark      string         `json:"remark,omitempty"`
}

// PositionTransferExternalRecord 外部转股记录。
type PositionTransferExternalRecord struct {
	ID         string `json:"id,omitempty"`
	Market     string `json:"market,omitempty"`
	Symbol     string `json:"symbol,omitempty"`
	Quantity   int64  `json:"quantity,omitempty"`
	Direction  string `json:"direction,omitempty"`
	Status     string `json:"status,omitempty"`
	SubmitTime int64  `json:"submitTime,omitempty"`
	UpdateTime int64  `json:"updateTime,omitempty"`
}
