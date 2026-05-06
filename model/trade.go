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
	ID             int64   `json:"id,omitempty"`
	OrderID        int64   `json:"orderId,omitempty"`
	Account        string  `json:"account,omitempty"`
	Symbol         string  `json:"symbol,omitempty"`
	SecType        string  `json:"secType,omitempty"`
	Market         string  `json:"market,omitempty"`
	Currency       string  `json:"currency,omitempty"`
	Identifier     string  `json:"identifier,omitempty"`
	Action         string  `json:"action,omitempty"`
	Price          float64 `json:"price,omitempty"`
	Quantity       int64   `json:"quantity,omitempty"`
	FilledQuantity int64   `json:"filledQuantity,omitempty"`
	Amount         float64 `json:"amount,omitempty"`
	Commission     float64 `json:"commission,omitempty"`
	TransactedAt   int64   `json:"transactedAt,omitempty"`
	Time           int64   `json:"time,omitempty"`
}
