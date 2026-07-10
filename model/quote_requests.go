package model

// FinancialDailyRequest 日级财务数据请求
type FinancialDailyRequest struct {
	Symbols   []string `json:"symbols"`
	Market    string   `json:"market"`
	Fields    []string `json:"fields"`
	BeginDate string   `json:"begin_date"`
	EndDate   string   `json:"end_date"`
}

// FinancialReportRequest 财报数据请求
type FinancialReportRequest struct {
	Symbols    []string `json:"symbols"`
	Market     string   `json:"market"`
	Fields     []string `json:"fields"`
	PeriodType string   `json:"period_type"`
	BeginDate  string   `json:"begin_date,omitempty"`
	EndDate    string   `json:"end_date,omitempty"`
}

// CorporateActionRequest 公司行动请求
type CorporateActionRequest struct {
	Symbols    []string `json:"symbols"`
	Market     string   `json:"market"`
	ActionType string   `json:"action_type"`
	BeginDate  string   `json:"begin_date,omitempty"`
	EndDate    string   `json:"end_date,omitempty"`
}

// FutureKlineRequest 期货 K 线请求
type FutureKlineRequest struct {
	ContractCodes []string `json:"contract_codes,omitempty"`
	ContractCode  string   `json:"contract_code,omitempty"`
	Period        string   `json:"period,omitempty"`
	BeginTime     int64    `json:"begin_time,omitempty"`
	EndTime       int64    `json:"end_time,omitempty"`
	BeginIndex    int      `json:"begin_index,omitempty"`
	EndIndex      int      `json:"end_index,omitempty"`
	Limit         int      `json:"limit,omitempty"`
	PageToken     string   `json:"page_token,omitempty"`
	Lang          string   `json:"lang,omitempty"`
}

// MarketScannerRequest 选股扫描请求
type MarketScannerRequest struct {
	Market               string                   `json:"market"`
	Page                 int                      `json:"page,omitempty"`
	PageSize             int                      `json:"page_size,omitempty"`
	CursorID             string                   `json:"cursor_id,omitempty"`
	BaseFilterList       []map[string]interface{} `json:"base_filter_list,omitempty"`
	AccumulateFilterList []map[string]interface{} `json:"accumulate_filter_list,omitempty"`
	FinancialFilterList  []map[string]interface{} `json:"financial_filter_list,omitempty"`
	MultiTagsFilterList  []map[string]interface{} `json:"multi_tags_filter_list,omitempty"`
	SortFieldData        map[string]interface{}   `json:"sort_field_data,omitempty"`
	MultiTagsFields      []string                 `json:"multi_tags_fields,omitempty"`
}

// ============================================================================
// Batch 3: 股票基础查询 + 时间序列
// ============================================================================

// SymbolsRequest — 查询全量合约代码/名称。
// wire: all_symbols / all_symbol_names
type SymbolsRequest struct {
	Market     string `json:"market,omitempty"`
	SecType    string `json:"sec_type,omitempty"`
	IncludeOtc bool   `json:"include_otc,omitempty"`
	Lang       string `json:"lang,omitempty"`
}

// TradeMetasRequest — 交易元数据。wire: quote_stock_trade
type TradeMetasRequest struct {
	Symbols []string `json:"symbols,omitempty"`
	Lang    string   `json:"lang,omitempty"`
}

// StockDetailsRequest — 股票详情。wire: stock_detail
type StockDetailsRequest struct {
	Symbols []string `json:"symbols,omitempty"`
	SecType string   `json:"sec_type,omitempty"`
	Lang    string   `json:"lang,omitempty"`
}

// StockDelayBriefsRequest — 延时行情。wire: quote_delay
type StockDelayBriefsRequest struct {
	Symbols []string `json:"symbols,omitempty"`
	SecType string   `json:"sec_type,omitempty"`
	Lang    string   `json:"lang,omitempty"`
}

// BriefRequest — 股票实时报价。wire: quote_real_time
// 对应 Python get_stock_briefs。服务端返回 {items: [...]} 对象。
type BriefRequest struct {
	Symbols            []string `json:"symbols,omitempty"`
	IncludeHourTrading *bool    `json:"include_hour_trading,omitempty"`
	SecType            string   `json:"sec_type,omitempty"`
	Lang               string   `json:"lang,omitempty"`
}

// DepthQuoteRequest — 盘口深度。wire: quote_depth
type DepthQuoteRequest struct {
	Symbols      []string `json:"symbols,omitempty"`
	Market       string   `json:"market,omitempty"`
	TradeSession string   `json:"trade_session,omitempty"`
	Lang         string   `json:"lang,omitempty"`
}

// KlineRequest — K 线查询。wire: kline
// 支持时间范围 (BeginTime/EndTime 毫秒) 或分页 (BeginIndex/EndIndex)，两者只能二选一。
type KlineRequest struct {
	Symbols         []string `json:"symbols,omitempty"`
	Period          string   `json:"period,omitempty"`
	Right           string   `json:"right,omitempty"`
	BeginTime       int64    `json:"begin_time,omitempty"`
	EndTime         int64    `json:"end_time,omitempty"`
	Limit           int      `json:"limit,omitempty"`
	BeginIndex      int      `json:"begin_index,omitempty"`
	EndIndex        int      `json:"end_index,omitempty"`
	PageToken       string   `json:"page_token,omitempty"`
	TradeSession    string   `json:"trade_session,omitempty"`
	Date            string   `json:"date,omitempty"`
	WithFundamental *bool    `json:"with_fundamental,omitempty"`
	SecType         string   `json:"sec_type,omitempty"`
	Lang            string   `json:"lang,omitempty"`
}

// KlineByPageRequest — K 线分页查询（客户端分页包装）。
type KlineByPageRequest struct {
	Symbol       string `json:"symbol,omitempty"`
	Period       string `json:"period,omitempty"`
	BeginTime    int64  `json:"begin_time,omitempty"`
	EndTime      int64  `json:"end_time,omitempty"`
	TotalSize    int    `json:"total_size,omitempty"` // 想获取的总条数
	PageSize     int    `json:"page_size,omitempty"`  // 每页条数
	Right        string `json:"right,omitempty"`
	Lang         string `json:"lang,omitempty"`
	TradeSession string `json:"trade_session,omitempty"`
}

// TimelineHistoryRequest — 历史分时。wire: history_timeline
type TimelineHistoryRequest struct {
	Symbols      []string `json:"symbols,omitempty"`
	Date         string   `json:"date,omitempty"` // yyyy-MM-dd
	Right        string   `json:"right,omitempty"`
	TradeSession string   `json:"trade_session,omitempty"`
	Lang         string   `json:"lang,omitempty"`
}

// TradeTickRequest — 逐笔成交。wire: trade_tick
type TradeTickRequest struct {
	Symbols    []string `json:"symbols,omitempty"`
	BeginIndex int      `json:"begin_index,omitempty"`
	EndIndex   int      `json:"end_index,omitempty"`
	Limit      int      `json:"limit,omitempty"`
	Lang       string   `json:"lang,omitempty"`
}

// TradeRankRequest — 成交榜单。wire: trade_rank
type TradeRankRequest struct {
	Market string `json:"market,omitempty"`
	Lang   string `json:"lang,omitempty"`
}

// ShortInterestRequest — 做空数据。wire: short_interest
type ShortInterestRequest struct {
	Symbols []string `json:"symbols,omitempty"`
	Lang    string   `json:"lang,omitempty"`
}

// StockBrokerRequest — 经纪商持仓。wire: stock_broker
type StockBrokerRequest struct {
	Symbol  string `json:"symbol,omitempty"`
	Limit   int    `json:"limit,omitempty"`
	SecType string `json:"sec_type,omitempty"`
	Lang    string `json:"lang,omitempty"`
}

// StockFundamentalRequest — 股票基本面。wire: stock_fundamental
type StockFundamentalRequest struct {
	Symbols []string `json:"symbols,omitempty"`
	Market  string   `json:"market,omitempty"`
	SecType string   `json:"sec_type,omitempty"`
	Lang    string   `json:"lang,omitempty"`
}

// StockIndustryRequest — 股票所属行业。wire: industry_stock
type StockIndustryRequest struct {
	Symbol  string `json:"symbol,omitempty"`
	Market  string `json:"market,omitempty"`
	SecType string `json:"sec_type,omitempty"`
	Lang    string `json:"lang,omitempty"`
}

// KlineQuotaRequest — K 线配额查询。wire: kline_quota
type KlineQuotaRequest struct {
	WithDetails bool   `json:"with_details,omitempty"`
	Lang        string `json:"lang,omitempty"`
}

// QuotePermissionRequest — 行情权限查询。wire: quote_permission
type QuotePermissionRequest struct {
	BeginDate string `json:"begin_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
	Lang      string `json:"lang,omitempty"`
}

// ============================================================================
// Batch 4: 期权/期货扩展
// ============================================================================

// OptionQueryItem 期权查询嵌套条目（option_query / option_basic 用）。
type OptionQueryItem struct {
	Symbol     string `json:"symbol,omitempty"`
	Expiry     int64  `json:"expiry,omitempty"`
	Strike     string `json:"strike,omitempty"`
	Right      string `json:"right,omitempty"`
	Period     string `json:"period,omitempty"`
	BeginTime  int64  `json:"begin_time,omitempty"`
	EndTime    int64  `json:"end_time,omitempty"`
	Limit      int    `json:"limit,omitempty"`
	BeginIndex int    `json:"begin_index,omitempty"`
	EndIndex   int    `json:"end_index,omitempty"`
	PageToken  string `json:"page_token,omitempty"`
}

// OptionTradeTicksRequest 期权逐笔。wire: option_trade_tick
type OptionTradeTicksRequest struct {
	Contracts []OptionQueryItem `json:"contracts,omitempty"`
	Lang      string            `json:"lang,omitempty"`
}

// OptionTimelineRequest 期权分时。wire: option_timeline
type OptionTimelineRequest struct {
	OptionQuery []OptionQueryItem `json:"option_query,omitempty"`
	Market      string            `json:"market,omitempty"`
	Lang        string            `json:"lang,omitempty"`
}

// OptionDepthRequest 期权盘口。wire: option_depth
type OptionDepthRequest struct {
	OptionBasic []OptionQueryItem `json:"option_basic,omitempty"`
	Market      string            `json:"market,omitempty"`
	Lang        string            `json:"lang,omitempty"`
}

// OptionSymbolsRequest 期权代码列表。wire: option_symbols
type OptionSymbolsRequest struct {
	Market string `json:"market,omitempty"`
	Lang   string `json:"lang,omitempty"`
}

// OptionAnalysisRequest 期权分析。wire: option_analysis
type OptionAnalysisRequest struct {
	Symbols               []string `json:"symbols,omitempty"`
	Market                string   `json:"market,omitempty"`
	Period                string   `json:"period,omitempty"` // OptionAnalysisPeriod
	RequireVolatilityList bool     `json:"require_volatility_list,omitempty"`
	Lang                  string   `json:"lang,omitempty"`
}

// FutureContractSingleRequest 按 contract_code/identifier 查询期货合约。wire: future_contract_by_contract_code / future_current_contract 等
type FutureContractSingleRequest struct {
	ContractCode string `json:"contract_code,omitempty"`
	Type         string `json:"type,omitempty"`
	Lang         string `json:"lang,omitempty"`
}

// AllFutureContractsRequest 查询所有期货合约。wire: future_contracts
type AllFutureContractsRequest struct {
	Type     string `json:"type,omitempty"`
	Exchange string `json:"exchange,omitempty"`
	Lang     string `json:"lang,omitempty"`
}

// FutureContinuousContractsRequest 连续主力合约。wire: future_continuous_contracts
type FutureContinuousContractsRequest struct {
	Type string `json:"type,omitempty"`
	Lang string `json:"lang,omitempty"`
}

// FutureHistoryMainContractRequest 主力合约历史。wire: future_main_contract
type FutureHistoryMainContractRequest struct {
	ContractCodes []string `json:"contract_codes,omitempty"`
	BeginTime     int64    `json:"begin_time,omitempty"`
	EndTime       int64    `json:"end_time,omitempty"`
	Lang          string   `json:"lang,omitempty"`
}

// FutureKlineByPageRequest 期货 K 线（客户端分页包装）。
type FutureKlineByPageRequest struct {
	ContractCode string `json:"contract_code,omitempty"`
	Period       string `json:"period,omitempty"`
	BeginTime    int64  `json:"begin_time,omitempty"`
	EndTime      int64  `json:"end_time,omitempty"`
	TotalSize    int    `json:"total_size,omitempty"`
	PageSize     int    `json:"page_size,omitempty"`
	Lang         string `json:"lang,omitempty"`
}

// FutureTradeTicksRequest 期货逐笔。wire: future_tick
// BeginIndex/EndIndex 默认值 0/30；不加 omitempty，确保零值也能发送到服务端。
type FutureTradeTicksRequest struct {
	ContractCode string `json:"contract_code,omitempty"`
	BeginIndex   int    `json:"begin_index"`
	EndIndex     int    `json:"end_index"`
	Limit        int    `json:"limit,omitempty"`
	Lang         string `json:"lang,omitempty"`
}

// FutureDepthRequest 期货盘口。wire: future_depth
type FutureDepthRequest struct {
	ContractCodes []string `json:"contract_codes,omitempty"`
	Lang          string   `json:"lang,omitempty"`
}

// FutureTradingTimesRequest 期货交易时段。wire: future_trading_date
type FutureTradingTimesRequest struct {
	ContractCode string `json:"contract_code,omitempty"`
	TradingDate  string `json:"trading_date,omitempty"`
	Lang         string `json:"lang,omitempty"`
}

// FutureBriefRequest 期货实时行情（contract_codes）。wire: future_real_time_quote
type FutureBriefRequest struct {
	ContractCodes []string `json:"contract_codes,omitempty"`
	Lang          string   `json:"lang,omitempty"`
}

// ============================================================================
// Batch 5: 基金/窝轮/行业/公司行动/财务/日历
// ============================================================================

// FundSymbolsRequest 基金代码列表。wire: fund_symbols
type FundSymbolsRequest struct {
	Lang string `json:"lang,omitempty"`
}

// FundContractsRequest 基金合约。wire: fund_contracts
type FundContractsRequest struct {
	Symbols []string `json:"symbols,omitempty"`
	Lang    string   `json:"lang,omitempty"`
}

// FundQuoteRequest 基金实时净值。wire: fund_quote
type FundQuoteRequest struct {
	Symbols []string `json:"symbols,omitempty"`
	Lang    string   `json:"lang,omitempty"`
}

// FundHistoryQuoteRequest 基金历史净值。wire: fund_history_quote
type FundHistoryQuoteRequest struct {
	Symbols   []string `json:"symbols,omitempty"`
	BeginTime int64    `json:"begin_time,omitempty"`
	EndTime   int64    `json:"end_time,omitempty"`
	Limit     int      `json:"limit,omitempty"`
	Lang      string   `json:"lang,omitempty"`
}

// WarrantBriefsRequest 窝轮简要报价。wire: warrant_briefs
type WarrantBriefsRequest struct {
	Symbols []string `json:"symbols,omitempty"`
	Lang    string   `json:"lang,omitempty"`
}

// WarrantFilterRequest 窝轮筛选。wire: warrant_filter
type WarrantFilterRequest struct {
	Symbol        string `json:"symbol,omitempty"`
	Page          int    `json:"page,omitempty"`
	PageSize      int    `json:"page_size,omitempty"`
	SortFieldName string `json:"sort_field_name,omitempty"`
	SortDir       string `json:"sort_dir,omitempty"`
	IssuerName    string `json:"issuer_name,omitempty"`
	ExpireYm      string `json:"expire_ym,omitempty"`
	Lang          string `json:"lang,omitempty"`
}

// IndustryListRequest 行业列表。wire: industry_list
type IndustryListRequest struct {
	IndustryLevel string `json:"industry_level,omitempty"` // IndustryLevel 枚举值
	Lang          string `json:"lang,omitempty"`
}

// IndustryStocksRequest 行业下股票列表。wire: industry_stock_list
type IndustryStocksRequest struct {
	IndustryID string `json:"industry_id,omitempty"`
	Market     string `json:"market,omitempty"`
	Lang       string `json:"lang,omitempty"`
}

// TradingCalendarRequest 交易日历。wire: trading_calendar
type TradingCalendarRequest struct {
	Market    string `json:"market,omitempty"`
	BeginDate string `json:"begin_date,omitempty"` // yyyy-MM-dd
	EndDate   string `json:"end_date,omitempty"`
	Lang      string `json:"lang,omitempty"`
}

// FinancialExchangeRateRequest 汇率。wire: financial_exchange_rate
type FinancialExchangeRateRequest struct {
	CurrencyList []string `json:"currency_list,omitempty"`
	BeginDate    string   `json:"begin_date,omitempty"` // yyyyMMdd
	EndDate      string   `json:"end_date,omitempty"`
	Timezone     string   `json:"timezone,omitempty"`
	Lang         string   `json:"lang,omitempty"`
}

// FinancialCurrencyRequest 财报币种。wire: financial_currency
type FinancialCurrencyRequest struct {
	Symbols []string `json:"symbols,omitempty"`
	Market  string   `json:"market,omitempty"`
	Lang    string   `json:"lang,omitempty"`
}

// MarketScannerTagsRequest 扫描器标签。wire: market_scanner_tags
type MarketScannerTagsRequest struct {
	Market          string   `json:"market,omitempty"`
	MultiTagsFields []string `json:"multi_tag_field_list,omitempty"`
	Lang            string   `json:"lang,omitempty"`
}

// QuoteOvernightRequest 隔夜行情。wire: quote_overnight
type QuoteOvernightRequest struct {
	Symbols []string `json:"symbols,omitempty"`
	Lang    string   `json:"lang,omitempty"`
}

// Deprecated: Use KlineRequest instead.
type BarsRequest = KlineRequest

// Deprecated: Use KlineByPageRequest instead.
type BarsByPageRequest = KlineByPageRequest
