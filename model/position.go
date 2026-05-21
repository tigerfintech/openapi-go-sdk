package model

// Position 持仓模型。
type Position struct {
	Account                           string   `json:"account,omitempty"`
	Symbol                            string   `json:"symbol,omitempty"`
	SecType                           string   `json:"secType,omitempty"`
	Market                            string   `json:"market,omitempty"`
	Currency                          string   `json:"currency,omitempty"`
	Position                          int64    `json:"position,omitempty"`
	PositionScale                     int      `json:"positionScale,omitempty"`
	PositionQty                       float64  `json:"positionQty,omitempty"`
	SalableQty                        float64  `json:"salableQty,omitempty"`
	AverageCost                       float64  `json:"averageCost,omitempty"`
	AverageCostByAverage              float64  `json:"averageCostByAverage,omitempty"`
	AverageCostOfCarry                float64  `json:"averageCostOfCarry,omitempty"`
	MarketValue                       float64  `json:"marketValue,omitempty"`
	RealizedPnl                       float64  `json:"realizedPnl,omitempty"`
	RealizedPnlByAverage              float64  `json:"realizedPnlByAverage,omitempty"`
	UnrealizedPnl                     float64  `json:"unrealizedPnl,omitempty"`
	UnrealizedPnlByAverage            float64  `json:"unrealizedPnlByAverage,omitempty"`
	UnrealizedPnlByCostOfCarry        float64  `json:"unrealizedPnlByCostOfCarry,omitempty"`
	UnrealizedPnlPercent              float64  `json:"unrealizedPnlPercent,omitempty"`
	UnrealizedPnlPercentByAverage     float64  `json:"unrealizedPnlPercentByAverage,omitempty"`
	UnrealizedPnlPercentByCostOfCarry float64  `json:"unrealizedPnlPercentByCostOfCarry,omitempty"`
	ContractId                        int64    `json:"contractId,omitempty"`
	Identifier                        string   `json:"identifier,omitempty"`
	Name                              string   `json:"name,omitempty"`
	LatestPrice                       float64  `json:"latestPrice,omitempty"`
	LastClosePrice                    float64  `json:"lastClosePrice,omitempty"`
	Multiplier                        float64  `json:"multiplier,omitempty"`
	Status                            int      `json:"status,omitempty"`
	UpdateTimestamp                   int64    `json:"updateTimestamp,omitempty"`
	MmPercent                         float64  `json:"mmPercent,omitempty"`
	MmValue                           float64  `json:"mmValue,omitempty"`
	TodayPnl                          float64  `json:"todayPnl,omitempty"`
	TodayPnlPercent                   float64  `json:"todayPnlPercent,omitempty"`
	ComboTypes                        []string `json:"comboTypes,omitempty"`
	Categories                        []string `json:"categories,omitempty"`
	IsLevel0Price                     bool     `json:"isLevel0Price,omitempty"`
	YesterdayPnl                      float64  `json:"yesterdayPnl,omitempty"`
	UnderlyingContractName            string   `json:"underlyingContractName,omitempty"`
}
