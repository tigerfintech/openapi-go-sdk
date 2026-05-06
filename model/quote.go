// Package model 行情响应模型。所有 json tag 与服务端返回的 camelCase 字段对齐。
package model

// MarketState 市场状态
type MarketState struct {
	Market       string `json:"market"`
	MarketStatus string `json:"marketStatus"`
	Status       string `json:"status"`
	OpenTime     string `json:"openTime,omitempty"`
}

// Brief 实时快照(quote_real_time)
type Brief struct {
	Symbol       string  `json:"symbol"`
	Open         float64 `json:"open,omitempty"`
	High         float64 `json:"high,omitempty"`
	Low          float64 `json:"low,omitempty"`
	Close        float64 `json:"close,omitempty"`
	PreClose     float64 `json:"preClose,omitempty"`
	LatestPrice  float64 `json:"latestPrice,omitempty"`
	LatestTime   int64   `json:"latestTime,omitempty"`
	AskPrice     float64 `json:"askPrice,omitempty"`
	AskSize      int64   `json:"askSize,omitempty"`
	BidPrice     float64 `json:"bidPrice,omitempty"`
	BidSize      int64   `json:"bidSize,omitempty"`
	Volume       int64   `json:"volume,omitempty"`
	Status       string  `json:"status,omitempty"`
	AdjPreClose  float64 `json:"adjPreClose,omitempty"`
	Change       float64 `json:"change,omitempty"`
	ChangeRate   float64 `json:"changeRate,omitempty"`
	Amplitude    float64 `json:"amplitude,omitempty"`
	Expiry       string  `json:"expiry,omitempty"`
	Strike       string  `json:"strike,omitempty"`
	Right        string  `json:"right,omitempty"`
	Multiplier   int     `json:"multiplier,omitempty"`
	OpenInterest int64   `json:"openInterest,omitempty"`
}

// Kline K 线(一个标的一组)
type Kline struct {
	Symbol        string      `json:"symbol"`
	Period        string      `json:"period,omitempty"`
	NextPageToken string      `json:"nextPageToken,omitempty"`
	Items         []KlineItem `json:"items"`
}

// KlineItem 单根 K 线
type KlineItem struct {
	Time   int64   `json:"time"`
	Volume int64   `json:"volume"`
	Open   float64 `json:"open"`
	Close  float64 `json:"close"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Amount float64 `json:"amount,omitempty"`
}

// Timeline 分时
type Timeline struct {
	Symbol   string          `json:"symbol"`
	Period   string          `json:"period"`
	PreClose float64         `json:"preClose"`
	Intraday *TimelineBucket `json:"intraday,omitempty"`
	PreHours *TimelineBucket `json:"preHours,omitempty"`
	AfterHours *TimelineBucket `json:"afterHours,omitempty"`
}

// TimelineBucket 分时数据块
type TimelineBucket struct {
	Items []TimelineItem `json:"items"`
}

// TimelineItem 分时点
type TimelineItem struct {
	Time     int64   `json:"time"`
	Volume   int64   `json:"volume"`
	Price    float64 `json:"price"`
	AvgPrice float64 `json:"avgPrice"`
}

// TradeTick 逐笔(一个标的一组)
type TradeTick struct {
	Symbol     string          `json:"symbol"`
	BeginIndex int64           `json:"beginIndex"`
	EndIndex   int64           `json:"endIndex"`
	Items      []TradeTickItem `json:"items"`
}

// TradeTickItem 单笔成交
type TradeTickItem struct {
	Time   int64   `json:"time"`
	Volume int64   `json:"volume"`
	Price  float64 `json:"price"`
	Type   string  `json:"type"`
}

// Depth 深度报价(一个标的一组)
type Depth struct {
	Symbol string       `json:"symbol"`
	Asks   []DepthLevel `json:"asks"`
	Bids   []DepthLevel `json:"bids"`
}

// DepthLevel 深度报价一档
type DepthLevel struct {
	Price  float64 `json:"price"`
	Count  int     `json:"count"`
	Volume int64   `json:"volume"`
}

// OptionExpiration 期权到期日(一个标的一组)
type OptionExpiration struct {
	Symbol        string   `json:"symbol"`
	OptionSymbols []string `json:"optionSymbols"`
	Dates         []string `json:"dates"`
	Timestamps    []int64  `json:"timestamps"`
	Periods       []string `json:"periods,omitempty"`
	Counts        []int    `json:"counts,omitempty"`
}

// OptionChain 期权链(一个标的 + 到期日一组)
type OptionChain struct {
	Symbol string           `json:"symbol"`
	Expiry int64            `json:"expiry"`
	Items  []OptionChainRow `json:"items"`
}

// OptionChainRow Put/Call 配对
type OptionChainRow struct {
	Put  *OptionLeg `json:"put,omitempty"`
	Call *OptionLeg `json:"call,omitempty"`
}

// OptionLeg 单腿期权数据
type OptionLeg struct {
	Identifier    string  `json:"identifier"`
	Strike        string  `json:"strike"`
	Right         string  `json:"right"`
	BidPrice      float64 `json:"bidPrice,omitempty"`
	BidSize       int64   `json:"bidSize,omitempty"`
	AskPrice      float64 `json:"askPrice,omitempty"`
	AskSize       int64   `json:"askSize,omitempty"`
	Volume        int64   `json:"volume,omitempty"`
	LatestPrice   float64 `json:"latestPrice,omitempty"`
	PreClose      float64 `json:"preClose,omitempty"`
	OpenInterest  int64   `json:"openInterest,omitempty"`
	Multiplier    int     `json:"multiplier,omitempty"`
	LastTimestamp int64   `json:"lastTimestamp,omitempty"`
	ImpliedVol    float64 `json:"impliedVol,omitempty"`
	Delta         float64 `json:"delta,omitempty"`
	Gamma         float64 `json:"gamma,omitempty"`
	Theta         float64 `json:"theta,omitempty"`
	Vega          float64 `json:"vega,omitempty"`
	Rho           float64 `json:"rho,omitempty"`
}

// FutureExchange 期货交易所
type FutureExchange struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	ZoneID string `json:"zoneId"`
}

// FutureContractInfo 期货合约详情(来自 future_contract_by_exchange_code 接口)
type FutureContractInfo struct {
	Continuous           bool    `json:"continuous"`
	Trade                bool    `json:"trade"`
	Type                 string  `json:"type"`
	ContractCode         string  `json:"contractCode"`
	IbCode               string  `json:"ibCode,omitempty"`
	Name                 string  `json:"name"`
	ContractMonth        string  `json:"contractMonth,omitempty"`
	LastTradingDate      string  `json:"lastTradingDate,omitempty"`
	FirstNoticeDate      string  `json:"firstNoticeDate,omitempty"`
	LastBiddingCloseTime int64   `json:"lastBiddingCloseTime,omitempty"`
	Currency             string  `json:"currency"`
	ExchangeCode         string  `json:"exchangeCode,omitempty"`
	Multiplier           float64 `json:"multiplier,omitempty"`
	MinTick              float64 `json:"minTick,omitempty"`
	DisplayMultiplier    float64 `json:"displayMultiplier,omitempty"`
	Exchange             string  `json:"exchange"`
	ProductWorth         string  `json:"productWorth,omitempty"`
	DeliveryMode         string  `json:"deliveryMode,omitempty"`
	ProductType          string  `json:"productType,omitempty"`
	ProductScale         string  `json:"productScale,omitempty"`
	LastTradingTimestamp int64   `json:"lastTradingTimestamp,omitempty"`
}

// FutureQuote 期货实时报价
type FutureQuote struct {
	ContractCode       string  `json:"contractCode"`
	LatestPrice        float64 `json:"latestPrice,omitempty"`
	LatestSize         int64   `json:"latestSize,omitempty"`
	LatestTime         int64   `json:"latestTime,omitempty"`
	BidPrice           float64 `json:"bidPrice,omitempty"`
	AskPrice           float64 `json:"askPrice,omitempty"`
	BidSize            int64   `json:"bidSize,omitempty"`
	AskSize            int64   `json:"askSize,omitempty"`
	OpenInterest       int64   `json:"openInterest,omitempty"`
	OpenInterestChange int64   `json:"openInterestChange,omitempty"`
	Volume             int64   `json:"volume,omitempty"`
	Open               float64 `json:"open,omitempty"`
	High               float64 `json:"high,omitempty"`
	Low                float64 `json:"low,omitempty"`
	Settlement         float64 `json:"settlement,omitempty"`
	LimitUp            float64 `json:"limitUp,omitempty"`
	LimitDown          float64 `json:"limitDown,omitempty"`
	AvgPrice           float64 `json:"avgPrice,omitempty"`
}

// FutureKline 期货 K 线(一个合约一组)
type FutureKline struct {
	NextPageToken string            `json:"nextPageToken,omitempty"`
	Items         []FutureKlineItem `json:"items"`
}

// FutureKlineItem 单根期货 K 线
type FutureKlineItem struct {
	Time         int64   `json:"time"`
	Volume       int64   `json:"volume"`
	Open         float64 `json:"open"`
	Close        float64 `json:"close"`
	High         float64 `json:"high"`
	Low          float64 `json:"low"`
	LastTime     int64   `json:"lastTime,omitempty"`
	OpenInterest int64   `json:"openInterest,omitempty"`
	Settlement   float64 `json:"settlement,omitempty"`
}

// FinancialDailyItem 日级财务数据行
type FinancialDailyItem struct {
	Symbol string  `json:"symbol"`
	Field  string  `json:"field"`
	Date   int64   `json:"date,omitempty"`
	Value  float64 `json:"value,omitempty"`
}

// FinancialReportItem 财报数据行
type FinancialReportItem struct {
	Symbol        string `json:"symbol"`
	Currency      string `json:"currency,omitempty"`
	Field         string `json:"field"`
	Value         string `json:"value,omitempty"`
	FilingDate    string `json:"filingDate,omitempty"`
	PeriodEndDate string `json:"periodEndDate,omitempty"`
}

// CorporateAction 公司行动
type CorporateAction struct {
	Symbol        string  `json:"symbol"`
	Market        string  `json:"market,omitempty"`
	Exchange      string  `json:"exchange,omitempty"`
	ExecuteDate   string  `json:"executeDate,omitempty"`
	ActionType    string  `json:"actionType,omitempty"`
	RecordDate    string  `json:"recordDate,omitempty"`
	AnnouncedDate string  `json:"announcedDate,omitempty"`
	PayDate       string  `json:"payDate,omitempty"`
	Amount        float64 `json:"amount,omitempty"`
	Currency      string  `json:"currency,omitempty"`
	// SPLIT 等场景可能出现的字段
	FromFactor float64 `json:"fromFactor,omitempty"`
	ToFactor   float64 `json:"toFactor,omitempty"`
}

// CapitalFlow 资金流向
type CapitalFlow struct {
	Symbol string            `json:"symbol"`
	Period string            `json:"period,omitempty"`
	Items  []CapitalFlowItem `json:"items"`
}

// CapitalFlowItem 资金流向明细
type CapitalFlowItem struct {
	Time      string  `json:"time"`
	Timestamp int64   `json:"timestamp"`
	NetInflow float64 `json:"netInflow"`
}

// CapitalDistribution 资金分布
type CapitalDistribution struct {
	Symbol    string  `json:"symbol"`
	NetInflow float64 `json:"netInflow"`
	InAll     float64 `json:"inAll"`
	InBig     float64 `json:"inBig"`
	InMid     float64 `json:"inMid"`
	InSmall   float64 `json:"inSmall"`
	OutAll    float64 `json:"outAll"`
	OutBig    float64 `json:"outBig"`
	OutMid    float64 `json:"outMid"`
	OutSmall  float64 `json:"outSmall"`
}

// ScannerResult 选股扫描结果(分页)
type ScannerResult struct {
	Page       int                 `json:"page"`
	TotalPage  int                 `json:"totalPage"`
	TotalCount int                 `json:"totalCount"`
	PageSize   int                 `json:"pageSize"`
	CursorID   string              `json:"cursorId,omitempty"`
	Items      []ScannerResultItem `json:"items"`
}

// ScannerResultItem 扫描结果行
type ScannerResultItem struct {
	Symbol              string           `json:"symbol"`
	Market              string           `json:"market"`
	BaseDataList        []ScannerDataRow `json:"baseDataList,omitempty"`
	AccumulateDataList  []ScannerDataRow `json:"accumulateDataList,omitempty"`
	FinancialDataList   []ScannerDataRow `json:"financialDataList,omitempty"`
	MultiTagDataList    []ScannerDataRow `json:"multiTagDataList,omitempty"`
}

// ScannerDataRow 扫描结果字段
type ScannerDataRow struct {
	Index int     `json:"index"`
	Name  string  `json:"name"`
	Value string  `json:"value,omitempty"`
	Data  float64 `json:"data,omitempty"`
}

// QuotePermission 行情权限
type QuotePermission struct {
	Name     string `json:"name"`
	ExpireAt int64  `json:"expireAt"`
}
