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
	Expiry       int64   `json:"expiry,omitempty"`
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
	Symbol     string          `json:"symbol"`
	Period     string          `json:"period"`
	PreClose   float64         `json:"preClose"`
	Intraday   *TimelineBucket `json:"intraday,omitempty"`
	PreHours   *TimelineBucket `json:"preHours,omitempty"`
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
	Symbol             string           `json:"symbol"`
	Market             string           `json:"market"`
	BaseDataList       []ScannerDataRow `json:"baseDataList,omitempty"`
	AccumulateDataList []ScannerDataRow `json:"accumulateDataList,omitempty"`
	FinancialDataList  []ScannerDataRow `json:"financialDataList,omitempty"`
	MultiTagDataList   []ScannerDataRow `json:"multiTagDataList,omitempty"`
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

// AddonEntitlement 附加套餐权益（wire 方法 addon_entitlements 返回）。
type AddonEntitlement struct {
	// UserLevel 用户套餐等级，服务端偶发返回数字而非字符串，用 FlexString 兼容两种格式。
	UserLevel            FlexString              `json:"userLevel,omitempty"`
	ActivePlan           *AddonActivePlan        `json:"activePlan,omitempty"`
	Addons               []AddonInfo             `json:"addons,omitempty"`
	EffectiveEntitlement *AddonEntitlementDetail `json:"effectiveEntitlement,omitempty"`
}

// AddonActivePlan 当前生效的套餐。
type AddonActivePlan struct {
	PlanType   string `json:"planType,omitempty"`
	ExpireTime int64  `json:"expireTime,omitempty"`
}

// AddonInfo 单个附加套餐信息。
type AddonInfo struct {
	PlanType   string `json:"planType,omitempty"`
	Active     bool   `json:"active,omitempty"`
	StartTime  int64  `json:"startTime,omitempty"`
	ExpireTime int64  `json:"expireTime,omitempty"`
}

// AddonEntitlementDetail 生效后的权益额度明细。
type AddonEntitlementDetail struct {
	HistoryStockLimit       int `json:"historyStockLimit,omitempty"`
	HistoryStockRemaining   int `json:"historyStockRemaining,omitempty"`
	HistoryFutureLimit      int `json:"historyFutureLimit,omitempty"`
	HistoryFutureRemaining  int `json:"historyFutureRemaining,omitempty"`
	HistoryOptionLimit      int `json:"historyOptionLimit,omitempty"`
	HistoryOptionRemaining  int `json:"historyOptionRemaining,omitempty"`
	SubscribeLimit          int `json:"subscribeLimit,omitempty"`
	SubscribeRemaining      int `json:"subscribeRemaining,omitempty"`
	SubscribeDepthLimit     int `json:"subscribeDepthLimit,omitempty"`
	SubscribeDepthRemaining int `json:"subscribeDepthRemaining,omitempty"`
	HighFreqLimit           int `json:"highFreqLimit,omitempty"`
	MidFreqLimit            int `json:"midFreqLimit,omitempty"`
	LowFreqLimit            int `json:"lowFreqLimit,omitempty"`
	RateMultiple            int `json:"rateMultiple,omitempty"`
}

// ============================================================================
// Batch 3-5: 扩展响应模型
// ============================================================================

// SymbolItem 合约代码条目（all_symbols 返回）。
type SymbolItem struct {
	Symbol string `json:"symbol,omitempty"`
	Market string `json:"market,omitempty"`
}

// SymbolName 合约代码 + 名称（all_symbol_names 返回）。
type SymbolName struct {
	Symbol string `json:"symbol,omitempty"`
	Name   string `json:"name,omitempty"`
	Market string `json:"market,omitempty"`
}

// TradeMeta 交易元数据（quote_stock_trade）。
type TradeMeta struct {
	Symbol         string  `json:"symbol,omitempty"`
	LotSize        int     `json:"lotSize,omitempty"`
	MinTick        float64 `json:"minTick,omitempty"`
	SpreadScale    float64 `json:"spreadScale,omitempty"`
	ShortableFlag  string  `json:"shortableFlag,omitempty"`
	MarginableFlag string  `json:"marginableFlag,omitempty"`
}

// StockDetail 股票详情（stock_detail）。
type StockDetail struct {
	Symbol         string  `json:"symbol,omitempty"`
	NameCN         string  `json:"nameCN,omitempty"`
	NameEN         string  `json:"nameEN,omitempty"`
	Exchange       string  `json:"exchange,omitempty"`
	Market         string  `json:"market,omitempty"`
	Currency       string  `json:"currency,omitempty"`
	SecType        string  `json:"secType,omitempty"`
	Sector         string  `json:"sector,omitempty"`
	Industry       string  `json:"industry,omitempty"`
	ListingDate    int64   `json:"listingDate,omitempty"` // ms timestamp
	MarketCap      float64 `json:"marketCap,omitempty"`
	CirculationCap float64 `json:"circulationCap,omitempty"`
	TotalShares    float64 `json:"totalShares,omitempty"`
	EpsTtm         float64 `json:"epsTtm,omitempty"`
	PeRatioTtm     float64 `json:"peRatioTtm,omitempty"`
}

// ShortInterest 做空数据。
type ShortInterest struct {
	Symbol                string  `json:"symbol,omitempty"`
	SettlementDate        string  `json:"settlementDate,omitempty"`
	ShortInterest         float64 `json:"shortInterest,omitempty"`
	AvgDailyVolume        float64 `json:"avgDailyVolume,omitempty"`
	DaysToCover           float64 `json:"daysToCover,omitempty"`
	PercentOfFloat        float64 `json:"percentOfFloat,omitempty"`
	ShortInterestPrevious float64 `json:"shortInterestPrevious,omitempty"`
	PercentChange         float64 `json:"percentChange,omitempty"`
}

// StockBroker 经纪商持仓。
type StockBroker struct {
	Symbol       string            `json:"symbol,omitempty"`
	LevelAskList []StockBrokerItem `json:"levelAskList,omitempty"`
	LevelBidList []StockBrokerItem `json:"levelBidList,omitempty"`
}

// StockBrokerItem 经纪商档位条目。
type StockBrokerItem struct {
	Level   int            `json:"level,omitempty"`
	Price   float64        `json:"price,omitempty"`
	Brokers []BrokerDetail `json:"brokers,omitempty"`
}

// BrokerDetail 经纪商明细。
type BrokerDetail struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// StockFundamental 股票基本面数据（服务端每个 symbol 下是 items 列表）。
type StockFundamental struct {
	Symbol string                   `json:"symbol,omitempty"`
	Items  []map[string]interface{} `json:"items,omitempty"`
}

// StockIndustry 股票行业归属。
type StockIndustry struct {
	Symbol  string `json:"symbol,omitempty"`
	GSector string `json:"gsector,omitempty"`
	GGroup  string `json:"ggroup,omitempty"`
	GInd    string `json:"gind,omitempty"`
	GSubInd string `json:"gsubind,omitempty"`
	Level   string `json:"level,omitempty"`
}

// TradeRankItem 成交榜单条目。
type TradeRankItem struct {
	Symbol     string  `json:"symbol,omitempty"`
	Name       string  `json:"name,omitempty"`
	LatestPr   float64 `json:"latestPrice,omitempty"`
	Change     float64 `json:"change,omitempty"`
	ChangeRate float64 `json:"changeRate,omitempty"`
	Volume     int64   `json:"volume,omitempty"`
	Amount     float64 `json:"amount,omitempty"`
}

// KlineQuota K 线配额。
type KlineQuota struct {
	Method string             `json:"method,omitempty"`
	Used   int                `json:"used,omitempty"`
	Quota  int                `json:"quota,omitempty"`
	Detail []KlineQuotaDetail `json:"detail,omitempty"`
}

// KlineQuotaDetail K 线配额明细。
type KlineQuotaDetail struct {
	Symbol     string `json:"symbol,omitempty"`
	Market     string `json:"market,omitempty"`
	UsedBars   int    `json:"usedBars,omitempty"`
	QuotaBars  int    `json:"quotaBars,omitempty"`
	LastAccess int64  `json:"lastAccess,omitempty"`
}

// OptionAnalysis 期权分析（波动率等）。
type OptionAnalysis struct {
	Symbol           string                  `json:"symbol,omitempty"`
	ImpliedVol30Days float64                 `json:"impliedVol30Days,omitempty"`
	HisVolatility    float64                 `json:"hisVolatility,omitempty"`
	IvHisVRatio      float64                 `json:"ivHisVRatio,omitempty"`
	CallPutRatio     float64                 `json:"callPutRatio,omitempty"`
	ImpliedVolMetric *ImpliedVolMetric       `json:"impliedVolMetric,omitempty"`
	VolatilityList   []OptionVolatilityPoint `json:"volatilityList,omitempty"`
}

// ImpliedVolMetric 隐含波动率指标（IV 百分位/排名）。
type ImpliedVolMetric struct {
	Period     string  `json:"period,omitempty"`
	Percentile float64 `json:"percentile,omitempty"`
	Rank       float64 `json:"rank,omitempty"`
}

// OptionVolatilityPoint 期权波动率时序点。
type OptionVolatilityPoint struct {
	ImpliedVol    float64 `json:"impliedVol,omitempty"`
	Percentile    float64 `json:"percentile,omitempty"`
	Rank          float64 `json:"rank,omitempty"`
	HisVolatility float64 `json:"hisVolatility,omitempty"`
	Timestamp     int64   `json:"timestamp,omitempty"`
}

// OptionSymbol 期权代码（get_option_symbols 返回）。
type OptionSymbol struct {
	Symbol string `json:"symbol,omitempty"`
	Market string `json:"market,omitempty"`
	NameCN string `json:"nameCN,omitempty"`
	NameEN string `json:"nameEN,omitempty"`
}

// (FutureContract 使用已有的 FutureContractInfo 类型，见文件上方)

// FutureMainContractHistory 期货主力合约历史。
type FutureMainContractHistory struct {
	ContractCode string `json:"contractCode,omitempty"`
	Symbol       string `json:"symbol,omitempty"`
	BeginDate    string `json:"beginDate,omitempty"`
	EndDate      string `json:"endDate,omitempty"`
}

// FutureTradingTime 期货交易时段。
type FutureTradingTime struct {
	ContractCode string                 `json:"contractCode,omitempty"`
	BizDate      string                 `json:"bizDate,omitempty"`
	Zone         string                 `json:"zone,omitempty"`
	TradingTimes []FutureTradingSegment `json:"tradingTimes,omitempty"`
}

// FutureTradingSegment 单个交易时段。
type FutureTradingSegment struct {
	Start int64  `json:"start,omitempty"` // ms timestamp
	End   int64  `json:"end,omitempty"`
	Type  string `json:"type,omitempty"` // RTH / Night / Break
}

// FutureTradeTickItem 期货逐笔。
type FutureTradeTickItem struct {
	ContractCode string  `json:"contractCode,omitempty"`
	Index        int64   `json:"index,omitempty"`
	Time         int64   `json:"time,omitempty"`
	Price        float64 `json:"price,omitempty"`
	Volume       int64   `json:"volume,omitempty"`
	Direction    string  `json:"direction,omitempty"`
}

// FutureDepth 期货盘口。
type FutureDepth struct {
	ContractCode string       `json:"contractCode,omitempty"`
	Timestamp    int64        `json:"timestamp,omitempty"`
	Asks         []DepthLevel `json:"asks,omitempty"`
	Bids         []DepthLevel `json:"bids,omitempty"`
}

// WarrantBrief 窝轮简要信息。
type WarrantBrief struct {
	Symbol      string  `json:"symbol,omitempty"`
	Name        string  `json:"name,omitempty"`
	LatestPrice float64 `json:"latestPrice,omitempty"`
	Change      float64 `json:"change,omitempty"`
	ChangeRate  float64 `json:"changeRate,omitempty"`
	Volume      int64   `json:"volume,omitempty"`
	Amount      float64 `json:"amount,omitempty"`
	Underlying  string  `json:"underlying,omitempty"`
	Issuer      string  `json:"issuer,omitempty"`
	ExpiryDate  string  `json:"expiryDate,omitempty"`
	StrikePrice float64 `json:"strikePrice,omitempty"`
	WarrantType string  `json:"warrantType,omitempty"`
}

// WarrantFilterResult 窝轮筛选结果。
type WarrantFilterResult struct {
	Total    int            `json:"total,omitempty"`
	Items    []WarrantBrief `json:"items,omitempty"`
	PageSize int            `json:"pageSize,omitempty"`
	Page     int            `json:"page,omitempty"`
}

// IndustryItem 行业列表条目。
type IndustryItem struct {
	ID    string `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	Level string `json:"level,omitempty"`
}

// IndustryStock 行业归属股票条目。
type IndustryStock struct {
	Symbol     string  `json:"symbol,omitempty"`
	Name       string  `json:"name,omitempty"`
	IndustryID string  `json:"industryId,omitempty"`
	Change     float64 `json:"change,omitempty"`
	ChangeRate float64 `json:"changeRate,omitempty"`
}

// TradingCalendarItem 交易日历。
type TradingCalendarItem struct {
	Market      string `json:"market,omitempty"`
	Date        string `json:"date,omitempty"`
	IsTrading   bool   `json:"isTrading,omitempty"`
	SessionType string `json:"sessionType,omitempty"`
}

// ExchangeRate 汇率数据。
type ExchangeRate struct {
	Currency     string  `json:"currency,omitempty"`
	Date         string  `json:"date,omitempty"`
	Rate         float64 `json:"rate,omitempty"`
	BaseCurrency string  `json:"baseCurrency,omitempty"`
}

// FinancialCurrency 财报货币。
type FinancialCurrency struct {
	Symbol   string `json:"symbol,omitempty"`
	Market   string `json:"market,omitempty"`
	Currency string `json:"currency,omitempty"`
}

// QuoteOvernight 隔夜行情。
type QuoteOvernight struct {
	Symbol     string  `json:"symbol,omitempty"`
	PreClose   float64 `json:"preClose,omitempty"`
	Open       float64 `json:"open,omitempty"`
	Close      float64 `json:"close,omitempty"`
	High       float64 `json:"high,omitempty"`
	Low        float64 `json:"low,omitempty"`
	Volume     int64   `json:"volume,omitempty"`
	Amount     float64 `json:"amount,omitempty"`
	Change     float64 `json:"change,omitempty"`
	ChangeRate float64 `json:"changeRate,omitempty"`
	BeginTime  int64   `json:"beginTime,omitempty"`
	EndTime    int64   `json:"endTime,omitempty"`
}

// MarketScannerTagGroup 扫描器标签分组（server 返回数组，每个元素一个 market/multiTagField 组合）。
type MarketScannerTagGroup struct {
	Market        string             `json:"market,omitempty"`
	MultiTagField string             `json:"multiTagField,omitempty"`
	TagList       []MarketScannerTag `json:"tagList,omitempty"`
}

// MarketScannerTag 单个标签。
type MarketScannerTag struct {
	Field  string   `json:"field,omitempty"`
	Name   string   `json:"name,omitempty"`
	Values []string `json:"values,omitempty"`
}

// Deprecated: MarketScannerTags has wrong shape; use []MarketScannerTagGroup instead.
type MarketScannerTags = []MarketScannerTagGroup

// FundContractInfo 基金合约。
type FundContractInfo struct {
	Symbol       string  `json:"symbol,omitempty"`
	Name         string  `json:"name,omitempty"`
	Currency     string  `json:"currency,omitempty"`
	FundType     string  `json:"fundType,omitempty"`
	Inception    string  `json:"inception,omitempty"`
	NetAssetVal  float64 `json:"netAssetValue,omitempty"`
	ExpenseRatio float64 `json:"expenseRatio,omitempty"`
}

// FundQuote 基金净值报价。
type FundQuote struct {
	Symbol     string  `json:"symbol,omitempty"`
	LatestNav  float64 `json:"latestNav,omitempty"`
	Change     float64 `json:"change,omitempty"`
	ChangeRate float64 `json:"changeRate,omitempty"`
	Date       string  `json:"date,omitempty"`
}

// FundHistoryQuote 基金历史净值。
type FundHistoryQuote struct {
	Symbol string  `json:"symbol,omitempty"`
	Date   string  `json:"date,omitempty"`
	Nav    float64 `json:"nav,omitempty"`
}
