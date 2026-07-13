// Package quote provides a quote query client wrapping all market data APIs.
package quote

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/tigerfintech/openapi-go-sdk/client"
	"github.com/tigerfintech/openapi-go-sdk/config"
	"github.com/tigerfintech/openapi-go-sdk/model"
)

// QuoteClient wraps all market data API calls.
type QuoteClient struct {
	httpClient *client.HttpClient
}

// NewQuoteClient creates a quote query client.
func NewQuoteClient(httpClient *client.HttpClient) *QuoteClient {
	return &QuoteClient{httpClient: httpClient}
}

// NewQuoteClientFromConfig creates a QuoteClient directly from a ClientConfig —
// no need to construct HttpClient manually.
func NewQuoteClientFromConfig(cfg *config.ClientConfig) *QuoteClient {
	return &QuoteClient{httpClient: client.NewQuoteHttpClient(cfg)}
}

// callInto sends a request and unmarshals data into out.
func (c *QuoteClient) callInto(method string, bizParams interface{}, out interface{}) error {
	req, err := client.NewApiRequest(method, bizParams)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Execute(req)
	if err != nil {
		return err
	}
	return client.UnmarshalData(resp.Data, out)
}

// callIntoVersioned is like callInto but with a specific API version.
func (c *QuoteClient) callIntoVersioned(method string, bizParams interface{}, version string, out interface{}) error {
	req, err := client.NewVersionedApiRequest(method, bizParams, version)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Execute(req)
	if err != nil {
		return err
	}
	return client.UnmarshalData(resp.Data, out)
}

// callIntoItems unwraps server {"items":[...]} envelope and unmarshals into items.
func (c *QuoteClient) callIntoItems(method string, bizParams interface{}, items interface{}) error {
	var wrap struct {
		Items json.RawMessage `json:"items"`
	}
	if err := c.callInto(method, bizParams, &wrap); err != nil {
		return err
	}
	if len(wrap.Items) == 0 || string(wrap.Items) == "null" {
		return nil
	}
	return json.Unmarshal(wrap.Items, items)
}

// callIntoListOrObject tries to unmarshal as a list first; if server returns a
// single object, wraps it into a single-element list. Used by endpoints like
// future_contract_by_contract_code / future_continuous_contracts that have
// inconsistent server response shapes.
func (c *QuoteClient) callIntoListOrObject(method string, bizParams interface{}, out interface{}) error {
	req, err := client.NewApiRequest(method, bizParams)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Execute(req)
	if err != nil {
		return err
	}
	// Try list first
	if err := client.UnmarshalData(resp.Data, out); err == nil {
		return nil
	}
	// Fall back to single object, wrap into [raw]
	var single json.RawMessage
	if err := client.UnmarshalData(resp.Data, &single); err != nil {
		return err
	}
	wrapped := append(append([]byte("["), []byte(single)...), ']')
	return json.Unmarshal(wrapped, out)
}

// === Basic market data methods ===

// GetMarketState returns market status.
func (c *QuoteClient) GetMarketState(market string) ([]model.MarketState, error) {
	var out []model.MarketState
	err := c.callInto("market_state", map[string]interface{}{"market": market}, &out)
	return out, err
}

// GetRealTimeQuote returns real-time stock quotes with full filter options.
// Wire method: quote_real_time. Accepts BriefRequest for include_hour_trading / sec_type / lang.
func (c *QuoteClient) GetRealTimeQuote(req model.BriefRequest) ([]model.Brief, error) {
	var out []model.Brief
	err := c.callInto("quote_real_time", req, &out)
	return out, err
}

// Deprecated: Use GetRealTimeQuote instead.
func (c *QuoteClient) GetBrief(req model.BriefRequest) ([]model.Brief, error) {
	return c.GetRealTimeQuote(req)
}

// GetKline returns K-line data. Supports time range (BeginTime/EndTime) or index paging (BeginIndex/EndIndex).
func (c *QuoteClient) GetKline(req model.KlineRequest) ([]model.Kline, error) {
	var out []model.Kline
	err := c.callInto("kline", req, &out)
	return out, err
}

// Deprecated: Use GetKline with KlineRequest instead.
func (c *QuoteClient) GetBars(req model.KlineRequest) ([]model.Kline, error) {
	return c.GetKline(req)
}

// GetTimeline returns intraday timeline data.
func (c *QuoteClient) GetTimeline(symbols []string) ([]model.Timeline, error) {
	var out []model.Timeline
	err := c.callInto("timeline", map[string]interface{}{"symbols": symbols}, &out)
	return out, err
}

// GetTradeTick returns tick-by-tick trade data with pagination support.
// Wire method: trade_tick. Accepts TradeTickRequest for begin_index/end_index/limit/lang.
func (c *QuoteClient) GetTradeTick(req model.TradeTickRequest) ([]model.TradeTick, error) {
	var out []model.TradeTick
	err := c.callInto("trade_tick", req, &out)
	return out, err
}

// GetQuoteDepth returns order book depth data.
// Wire method: quote_depth. Accepts DepthQuoteRequest for symbols/market/trade_session/lang.
func (c *QuoteClient) GetQuoteDepth(req model.DepthQuoteRequest) ([]model.Depth, error) {
	var out []model.Depth
	err := c.callInto("quote_depth", req, &out)
	return out, err
}

// === Option market data methods ===

// GetOptionExpiration returns option expiration dates for multiple symbols.
// market is optional; pass "HK" or "US" to filter by market. Defaults to US if omitted.
func (c *QuoteClient) GetOptionExpiration(symbols []string, market ...string) ([]model.OptionExpiration, error) {
	var out []model.OptionExpiration
	params := map[string]interface{}{"symbols": symbols}
	if len(market) > 0 && market[0] != "" {
		params["market"] = market[0]
	}
	err := c.callInto("option_expiration", params, &out)
	return out, err
}

// GetOptionChain returns option chains for multiple (symbol, expiry) pairs.
// Each item is [2]string{symbol, "YYYY-MM-DD"}.
// timezone is optional; if provided, the first element overrides the inferred timezone.
// US options default to America/New_York; HK options (.HK suffix) default to Asia/Hong_Kong.
func (c *QuoteClient) GetOptionChain(items [][2]string, timezone ...string) ([]model.OptionChain, error) {
	req := model.OptionChainRequest{}
	tz := ""
	if len(timezone) > 0 {
		tz = timezone[0]
	}
	optionBasic := make([]map[string]interface{}, 0, len(items))
	for _, it := range items {
		resolvedTz := resolveOptionTimezone(tz, it[0])
		expiryTs, err := parseOptionExpiry(it[1], resolvedTz)
		if err != nil {
			return nil, fmt.Errorf("invalid expiry date for %q: %w", it[0], err)
		}
		optionBasic = append(optionBasic, map[string]interface{}{
			"symbol": it[0],
			"expiry": expiryTs,
		})
	}
	req.OptionBasic = optionBasic
	return c.GetOptionChainByReq(req)
}

// GetOptionChainByReq retrieves option chain data with full filter and Greeks support.
func (c *QuoteClient) GetOptionChainByReq(req model.OptionChainRequest) ([]model.OptionChain, error) {
	var out []model.OptionChain
	err := c.callIntoVersioned("option_chain", req, "3.0", &out)
	return out, err
}

// GetOptionQuote returns option quotes for the given identifiers.
// timezone is optional; if provided, the first element overrides the inferred timezone.
// US options default to America/New_York; HK options (.HK suffix) default to Asia/Hong_Kong.
func (c *QuoteClient) GetOptionQuote(identifiers []string, timezone ...string) ([]model.Brief, error) {
	tz := ""
	if len(timezone) > 0 {
		tz = timezone[0]
	}
	optionBasics := make([]map[string]interface{}, 0, len(identifiers))
	for _, id := range identifiers {
		contract, err := optionContractFromIdentifier(id, resolveOptionTimezone(tz, strings.SplitN(strings.TrimSpace(id), " ", 2)[0]))
		if err != nil {
			return nil, fmt.Errorf("invalid option identifier %q: %w", id, err)
		}
		optionBasics = append(optionBasics, map[string]interface{}{
			"symbol": contract.Symbol,
			"expiry": contract.Expiry,
			"right":  contract.Right,
			"strike": contract.Strike,
		})
	}
	var out []model.Brief
	err := c.callIntoVersioned("option_brief", map[string]interface{}{"option_basic": optionBasics}, "2.0", &out)
	return out, err
}

// Deprecated: Use GetOptionQuote instead.
func (c *QuoteClient) GetOptionBrief(identifiers []string) ([]model.Brief, error) {
	return c.GetOptionQuote(identifiers)
}

// GetOptionKline returns option K-line data for multiple identifiers.
// beginTime and endTime are millisecond timestamps; pass -1 for server defaults, 0 is treated as -1.
// timezone is optional; if provided, the first element overrides the inferred timezone.
// US options default to America/New_York; HK options (.HK suffix) default to Asia/Hong_Kong.
func (c *QuoteClient) GetOptionKline(identifiers []string, period string, beginTime, endTime int64, timezone ...string) ([]model.Kline, error) {
	return c.GetOptionKlineWithOpts(identifiers, period, beginTime, endTime, 0, "", timezone...)
}

// GetOptionKlineWithOpts retrieves option kline data with optional limit and sort direction.
func (c *QuoteClient) GetOptionKlineWithOpts(identifiers []string, period string, beginTime, endTime int64, limit int, sortDir string, timezone ...string) ([]model.Kline, error) {
	tz := ""
	if len(timezone) > 0 {
		tz = timezone[0]
	}
	if beginTime == 0 {
		beginTime = -1
	}
	if endTime == 0 {
		endTime = -1
	}
	optionQuery := make([]map[string]interface{}, 0, len(identifiers))
	for _, id := range identifiers {
		contract, err := optionContractFromIdentifier(id, resolveOptionTimezone(tz, strings.SplitN(strings.TrimSpace(id), " ", 2)[0]))
		if err != nil {
			return nil, fmt.Errorf("invalid option identifier %q: %w", id, err)
		}
		entry := map[string]interface{}{
			"symbol":     contract.Symbol,
			"expiry":     contract.Expiry,
			"right":      contract.Right,
			"strike":     contract.Strike,
			"period":     period,
			"begin_time": beginTime,
			"end_time":   endTime,
		}
		if limit > 0 {
			entry["limit"] = limit
		}
		if sortDir != "" {
			entry["sort_dir"] = sortDir
		}
		optionQuery = append(optionQuery, entry)
	}
	params := map[string]interface{}{"option_query": optionQuery}
	var out []model.Kline
	err := c.callIntoVersioned("option_kline", params, "2.0", &out)
	return out, err
}

// === Futures market data methods ===

// GetFutureExchange returns the list of futures exchanges.
func (c *QuoteClient) GetFutureExchange() ([]model.FutureExchange, error) {
	var out []model.FutureExchange
	err := c.callInto("future_exchange", map[string]interface{}{"sec_type": "FUT"}, &out)
	return out, err
}

// GetFutureContracts returns futures contracts for an exchange.
func (c *QuoteClient) GetFutureContracts(exchange string) ([]model.FutureContractInfo, error) {
	var out []model.FutureContractInfo
	err := c.callInto("future_contract_by_exchange_code", map[string]interface{}{"exchange_code": exchange}, &out)
	return out, err
}

// GetFutureRealTimeQuote returns real-time futures quotes.
// Wire method: future_real_time_quote. Accepts FutureBriefRequest for contract_codes/lang.
func (c *QuoteClient) GetFutureRealTimeQuote(req model.FutureBriefRequest) ([]model.FutureQuote, error) {
	var out []model.FutureQuote
	err := c.callInto("future_real_time_quote", req, &out)
	return out, err
}

// GetFutureKline returns futures K-line data.
// beginTimeMs / endTimeMs: 13-digit ms timestamps; use -1 to represent unbounded.
func (c *QuoteClient) GetFutureKline(req model.FutureKlineRequest) ([]model.FutureKline, error) {
	if req.BeginTime == 0 {
		req.BeginTime = -1
	}
	if req.EndTime == 0 {
		req.EndTime = -1
	}
	var out []model.FutureKline
	err := c.callInto("future_kline", req, &out)
	return out, err
}

// === Fundamentals and capital flow methods ===

// GetFinancialDaily returns daily financial data.
func (c *QuoteClient) GetFinancialDaily(req model.FinancialDailyRequest) ([]model.FinancialDailyItem, error) {
	var out []model.FinancialDailyItem
	err := c.callInto("financial_daily", req, &out)
	return out, err
}

// GetFinancialReport returns financial reports.
func (c *QuoteClient) GetFinancialReport(req model.FinancialReportRequest) ([]model.FinancialReportItem, error) {
	var out []model.FinancialReportItem
	err := c.callInto("financial_report", req, &out)
	return out, err
}

// GetCorporateAction returns corporate actions keyed by symbol.
// The server returns a map {symbol: [...actions]}; this method flattens to a single slice.
func (c *QuoteClient) GetCorporateAction(req model.CorporateActionRequest) ([]model.CorporateAction, error) {
	var grouped map[string][]model.CorporateAction
	if err := c.callInto("corporate_action", req, &grouped); err != nil {
		return nil, err
	}
	var out []model.CorporateAction
	for _, list := range grouped {
		out = append(out, list...)
	}
	return out, nil
}

// GetCapitalFlow returns capital flow data.
// period: e.g. "day"/"intraday"/"week"/"month".
func (c *QuoteClient) GetCapitalFlow(symbol, market, period string) (*model.CapitalFlow, error) {
	var out model.CapitalFlow
	err := c.callInto("capital_flow", map[string]interface{}{
		"symbol": symbol,
		"market": market,
		"period": period,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// GetCapitalDistribution returns capital distribution data.
func (c *QuoteClient) GetCapitalDistribution(symbol, market string) (*model.CapitalDistribution, error) {
	var out model.CapitalDistribution
	err := c.callInto("capital_distribution", map[string]interface{}{
		"symbol": symbol,
		"market": market,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// === Scanner and quote permission methods ===

// MarketScanner runs the stock screener.
func (c *QuoteClient) MarketScanner(req model.MarketScannerRequest) (*model.ScannerResult, error) {
	var out model.ScannerResult
	err := c.callInto("market_scanner", req, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// GrabQuotePermission requests quote permissions.
func (c *QuoteClient) GrabQuotePermission() ([]model.QuotePermission, error) {
	var out []model.QuotePermission
	err := c.callInto("grab_quote_permission", nil, &out)
	return out, err
}

// GetAddonEntitlement returns the user's addon plan entitlements.
// Wire method: addon_entitlements. 对应 Java AddonEntitlementRequest。
func (c *QuoteClient) GetAddonEntitlement() (*model.AddonEntitlement, error) {
	var out model.AddonEntitlement
	if err := c.callInto("addon_entitlements", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// === Option helper functions ===

// optionContract holds parsed fields from an OCC option identifier.
type optionContract struct {
	Symbol string
	Expiry int64  // millisecond timestamp
	Right  string // "CALL" or "PUT"
	Strike float64
}

// resolveOptionTimezone returns the timezone string to use for expiry conversion.
// tz takes precedence; if empty, infer from symbol (HK suffix → Asia/Hong_Kong, else America/New_York).
func resolveOptionTimezone(tz string, symbol string) string {
	if tz != "" {
		return tz
	}
	if strings.HasSuffix(strings.ToUpper(symbol), ".HK") {
		return "Asia/Hong_Kong"
	}
	return "America/New_York"
}

// parseOptionExpiry parses a date string like "2024-01-19" into a millisecond timestamp
// in the given timezone (e.g. "America/New_York" or "Asia/Hong_Kong").
func parseOptionExpiry(expiry string, tz string) (int64, error) {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return 0, fmt.Errorf("invalid timezone %q: %w", tz, err)
	}
	t, err := time.ParseInLocation("2006-01-02", expiry, loc)
	if err != nil {
		return 0, fmt.Errorf("expected format YYYY-MM-DD, got %q: %w", expiry, err)
	}
	return t.UnixMilli(), nil
}

// optionContractFromIdentifier parses an OCC-style option identifier like "AAPL 240119C00150000"
// into its component parts: symbol, expiry (ms timestamp in tz), right (CALL/PUT), strike.
func optionContractFromIdentifier(identifier string, tz string) (optionContract, error) {
	parts := strings.SplitN(strings.TrimSpace(identifier), " ", 2)
	if len(parts) != 2 {
		return optionContract{}, fmt.Errorf("expected format 'SYMBOL YYMMDDX00000000', got %q", identifier)
	}

	symbol := parts[0]
	rest := strings.TrimSpace(parts[1])
	if len(rest) < 15 {
		return optionContract{}, fmt.Errorf("option code too short: %q", rest)
	}

	dateStr := rest[:6]
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return optionContract{}, fmt.Errorf("invalid timezone %q: %w", tz, err)
	}
	t, err := time.ParseInLocation("060102", dateStr, loc)
	if err != nil {
		return optionContract{}, fmt.Errorf("invalid date in identifier: %q", dateStr)
	}

	rightChar := rest[6]
	var right string
	switch rightChar {
	case 'C':
		right = "CALL"
	case 'P':
		right = "PUT"
	default:
		return optionContract{}, fmt.Errorf("invalid right character: %c", rightChar)
	}

	strikeStr := rest[7:]
	var strikeInt int64
	for _, ch := range strikeStr {
		if ch < '0' || ch > '9' {
			return optionContract{}, fmt.Errorf("invalid strike digits: %q", strikeStr)
		}
		strikeInt = strikeInt*10 + int64(ch-'0')
	}
	strike := float64(strikeInt) / 1000.0

	return optionContract{
		Symbol: symbol,
		Expiry: t.UnixMilli(),
		Right:  right,
		Strike: strike,
	}, nil
}

// unused import guard for json (some build tags might omit all uses)
var _ = json.RawMessage(nil)

// ============================================================================
// Batch 3: 股票基础查询 + 时间序列
// ============================================================================

// GetSymbols 查询全量合约代码。wire: all_symbols
// 服务端返回字符串列表，不是对象列表。
func (c *QuoteClient) GetSymbols(req model.SymbolsRequest) ([]string, error) {
	var out []string
	err := c.callInto("all_symbols", req, &out)
	return out, err
}

// GetSymbolNames 查询全量合约代码+名称。wire: all_symbol_names
func (c *QuoteClient) GetSymbolNames(req model.SymbolsRequest) ([]model.SymbolName, error) {
	var out []model.SymbolName
	err := c.callInto("all_symbol_names", req, &out)
	return out, err
}

// GetTradeMetas 交易元数据。wire: quote_stock_trade
func (c *QuoteClient) GetTradeMetas(req model.TradeMetasRequest) ([]model.TradeMeta, error) {
	var out []model.TradeMeta
	err := c.callInto("quote_stock_trade", req, &out)
	return out, err
}

// GetStockDetails 股票详情。wire: stock_detail。服务端返回 {items:[...]}。
func (c *QuoteClient) GetStockDetails(req model.StockDetailsRequest) ([]model.StockDetail, error) {
	var out []model.StockDetail
	err := c.callIntoItems("stock_detail", req, &out)
	return out, err
}

// GetDelayedQuote returns delayed stock quotes. wire: quote_delay
func (c *QuoteClient) GetDelayedQuote(req model.StockDelayBriefsRequest) ([]model.Brief, error) {
	var out []model.Brief
	err := c.callInto("quote_delay", req, &out)
	return out, err
}

// Deprecated: Use GetDelayedQuote instead.
func (c *QuoteClient) GetStockDelayBriefs(req model.StockDelayBriefsRequest) ([]model.Brief, error) {
	return c.GetDelayedQuote(req)
}

// GetKlineByPage client-side paginated K-line: loops until TotalSize bars are collected.
// Returns merged slice in ascending time order.
func (c *QuoteClient) GetKlineByPage(req model.KlineByPageRequest) ([]model.KlineItem, error) {
	if req.PageSize <= 0 {
		req.PageSize = 200
	}
	if req.TotalSize <= 0 {
		req.TotalSize = 1000
	}
	var acc []model.KlineItem
	beginTime := req.BeginTime
	endTime := req.EndTime
	if endTime == 0 {
		endTime = -1
	}
	if beginTime == 0 {
		beginTime = -1
	}
	for len(acc) < req.TotalSize {
		sub := model.KlineRequest{
			Symbols:      []string{req.Symbol},
			Period:       req.Period,
			Right:        req.Right,
			BeginTime:    beginTime,
			EndTime:      endTime,
			Limit:        req.PageSize,
			TradeSession: req.TradeSession,
			Lang:         req.Lang,
		}
		var pageOut []model.Kline
		if err := c.callInto("kline", sub, &pageOut); err != nil {
			return acc, err
		}
		if len(pageOut) == 0 || len(pageOut[0].Items) == 0 {
			break
		}
		items := pageOut[0].Items
		acc = append(acc, items...)
		if len(items) < req.PageSize {
			break
		}
		oldest := items[0].Time
		for _, it := range items {
			if it.Time < oldest {
				oldest = it.Time
			}
		}
		endTime = oldest - 1
	}
	return acc, nil
}

// Deprecated: Use GetKlineByPage with KlineByPageRequest instead.
func (c *QuoteClient) GetBarsByPage(req model.KlineByPageRequest) ([]model.KlineItem, error) {
	return c.GetKlineByPage(req)
}

// GetTimelineHistory 历史分时。wire: history_timeline
func (c *QuoteClient) GetTimelineHistory(req model.TimelineHistoryRequest) ([]model.Timeline, error) {
	var out []model.Timeline
	err := c.callInto("history_timeline", req, &out)
	return out, err
}

// GetTradeRank 成交榜单（涨跌幅榜）。wire: trade_rank
func (c *QuoteClient) GetTradeRank(req model.TradeRankRequest) ([]model.TradeRankItem, error) {
	var out []model.TradeRankItem
	err := c.callInto("trade_rank", req, &out)
	return out, err
}

// GetShortInterest 做空数据。wire: quote_shortable_stocks
func (c *QuoteClient) GetShortInterest(req model.ShortInterestRequest) ([]model.ShortInterest, error) {
	var out []model.ShortInterest
	err := c.callInto("quote_shortable_stocks", req, &out)
	return out, err
}

// GetStockBroker 经纪商持仓。wire: stock_broker
func (c *QuoteClient) GetStockBroker(req model.StockBrokerRequest) (*model.StockBroker, error) {
	var out model.StockBroker
	if err := c.callInto("stock_broker", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetStockFundamental 股票基本面。wire: stock_fundamental
// 服务端返回按 symbol 分组的结构；此处返回原始 JSON 由调用方进一步解析。
func (c *QuoteClient) GetStockFundamental(req model.StockFundamentalRequest) (map[string]interface{}, error) {
	var out map[string]interface{}
	err := c.callInto("stock_fundamental", req, &out)
	return out, err
}

// GetStockIndustry 股票所属行业。wire: stock_industry
// 服务端返回数组（按 level 返回各级行业）。
func (c *QuoteClient) GetStockIndustry(req model.StockIndustryRequest) ([]model.StockIndustry, error) {
	var out []model.StockIndustry
	err := c.callInto("stock_industry", req, &out)
	return out, err
}

// GetQuotePermission 查询行情权限详情。wire: get_quote_permission
func (c *QuoteClient) GetQuotePermission(req model.QuotePermissionRequest) ([]model.QuotePermission, error) {
	var out []model.QuotePermission
	err := c.callInto("get_quote_permission", req, &out)
	return out, err
}

// GetKlineQuota K 线配额查询。wire: kline_quota
func (c *QuoteClient) GetKlineQuota(req model.KlineQuotaRequest) ([]model.KlineQuota, error) {
	var out []model.KlineQuota
	err := c.callInto("kline_quota", req, &out)
	return out, err
}

// ============================================================================
// Batch 4: 期权/期货扩展
// ============================================================================

// GetOptionTradeTicks 期权逐笔成交。wire: option_trade_tick
func (c *QuoteClient) GetOptionTradeTicks(req model.OptionTradeTicksRequest) ([]model.TradeTick, error) {
	var out []model.TradeTick
	err := c.callInto("option_trade_tick", req, &out)
	return out, err
}

// GetOptionTimeline 期权分时。wire: option_timeline
func (c *QuoteClient) GetOptionTimeline(req model.OptionTimelineRequest) ([]model.Timeline, error) {
	var out []model.Timeline
	err := c.callInto("option_timeline", req, &out)
	return out, err
}

// GetOptionDepth 期权盘口深度。wire: option_depth
func (c *QuoteClient) GetOptionDepth(req model.OptionDepthRequest) ([]model.Depth, error) {
	var out []model.Depth
	err := c.callInto("option_depth", req, &out)
	return out, err
}

// GetOptionSymbols 期权代码列表。wire: option_symbol
func (c *QuoteClient) GetOptionSymbols(req model.OptionSymbolsRequest) ([]model.OptionSymbol, error) {
	var out []model.OptionSymbol
	err := c.callInto("option_symbol", req, &out)
	return out, err
}

// GetOptionAnalysis 期权分析（隐含/历史波动率）。wire: option_analysis
func (c *QuoteClient) GetOptionAnalysis(req model.OptionAnalysisRequest) ([]model.OptionAnalysis, error) {
	var out []model.OptionAnalysis
	err := c.callInto("option_analysis", req, &out)
	return out, err
}

// GetFutureContract 按 contract_code 查单个期货合约。wire: future_contract_by_contract_code
// 服务端可能返回单个对象或列表，此处统一展开为 []FutureContractInfo。
func (c *QuoteClient) GetFutureContract(req model.FutureContractSingleRequest) ([]model.FutureContractInfo, error) {
	var out []model.FutureContractInfo
	err := c.callIntoListOrObject("future_contract_by_contract_code", req, &out)
	return out, err
}

// GetAllFutureContracts 查询所有期货合约。wire: future_contracts
func (c *QuoteClient) GetAllFutureContracts(req model.AllFutureContractsRequest) ([]model.FutureContractInfo, error) {
	var out []model.FutureContractInfo
	err := c.callInto("future_contracts", req, &out)
	return out, err
}

// GetCurrentFutureContract 当前主力合约。wire: future_current_contract
func (c *QuoteClient) GetCurrentFutureContract(req model.FutureContractSingleRequest) (*model.FutureContractInfo, error) {
	var out model.FutureContractInfo
	if err := c.callInto("future_current_contract", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetFutureContinuousContracts 连续主力合约。wire: future_continuous_contracts
func (c *QuoteClient) GetFutureContinuousContracts(req model.FutureContinuousContractsRequest) ([]model.FutureContractInfo, error) {
	var out []model.FutureContractInfo
	err := c.callIntoListOrObject("future_continuous_contracts", req, &out)
	return out, err
}

// GetFutureHistoryMainContract 期货主力合约历史。wire: future_main_contract
func (c *QuoteClient) GetFutureHistoryMainContract(req model.FutureHistoryMainContractRequest) ([]model.FutureMainContractHistory, error) {
	var out []model.FutureMainContractHistory
	err := c.callInto("future_main_contract", req, &out)
	return out, err
}

// GetFutureKlineByPage client-side paginated futures K-line: loops until TotalSize bars are collected.
func (c *QuoteClient) GetFutureKlineByPage(req model.FutureKlineByPageRequest) ([]model.FutureKlineItem, error) {
	if req.PageSize <= 0 {
		req.PageSize = 200
	}
	if req.TotalSize <= 0 {
		req.TotalSize = 1000
	}
	var acc []model.FutureKlineItem
	endTime := req.EndTime
	if endTime == 0 {
		endTime = -1
	}
	beginTime := req.BeginTime
	if beginTime == 0 {
		beginTime = -1
	}
	for len(acc) < req.TotalSize {
		sub := model.FutureKlineRequest{
			ContractCodes: []string{req.ContractCode},
			Period:        req.Period,
			BeginTime:     beginTime,
			EndTime:       endTime,
			Limit:         req.PageSize,
			Lang:          req.Lang,
		}
		var pageOut []model.FutureKline
		if err := c.callInto("future_kline", sub, &pageOut); err != nil {
			return acc, err
		}
		if len(pageOut) == 0 || len(pageOut[0].Items) == 0 {
			break
		}
		items := pageOut[0].Items
		acc = append(acc, items...)
		if len(items) < req.PageSize {
			break
		}
		oldest := items[0].Time
		for _, it := range items {
			if it.Time < oldest {
				oldest = it.Time
			}
		}
		endTime = oldest - 1
	}
	return acc, nil
}

// GetFutureTradeTicks 期货逐笔。wire: future_tick (API version 3.0)
// 服务端返回 {contractCode, items:[...]} 对象，需解包 items 并回填 contractCode。
// end_index 服务端要求必须 >= 0；未设置时默认 30（与 Python SDK 一致）。
func (c *QuoteClient) GetFutureTradeTicks(req model.FutureTradeTicksRequest) ([]model.FutureTradeTickItem, error) {
	if req.EndIndex == 0 {
		req.EndIndex = 30
	}
	var wrap struct {
		ContractCode string                      `json:"contractCode"`
		Items        []model.FutureTradeTickItem `json:"items"`
	}
	if err := c.callIntoVersioned("future_tick", req, "3.0", &wrap); err != nil {
		return nil, err
	}
	for i := range wrap.Items {
		if wrap.Items[i].ContractCode == "" {
			wrap.Items[i].ContractCode = wrap.ContractCode
		}
	}
	return wrap.Items, nil
}

// GetFutureDepth 期货盘口。wire: future_depth
func (c *QuoteClient) GetFutureDepth(req model.FutureDepthRequest) ([]model.FutureDepth, error) {
	var out []model.FutureDepth
	err := c.callInto("future_depth", req, &out)
	return out, err
}

// GetFutureTradingTimes 期货交易时段。wire: future_trading_date
// 服务端返回单个对象（{timeSection/tradingTimes/biddingTimes}）。
func (c *QuoteClient) GetFutureTradingTimes(req model.FutureTradingTimesRequest) (*model.FutureTradingTime, error) {
	var out model.FutureTradingTime
	if err := c.callInto("future_trading_date", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ============================================================================
// Batch 5: 基金/窝轮/行业/公司行动/财务/日历
// ============================================================================

// GetFundSymbols 基金代码列表。wire: fund_all_symbols
func (c *QuoteClient) GetFundSymbols(req model.FundSymbolsRequest) ([]string, error) {
	var out []string
	err := c.callInto("fund_all_symbols", req, &out)
	return out, err
}

// GetFundContracts 基金合约信息。wire: fund_contracts
func (c *QuoteClient) GetFundContracts(req model.FundContractsRequest) ([]model.FundContractInfo, error) {
	var out []model.FundContractInfo
	err := c.callInto("fund_contracts", req, &out)
	return out, err
}

// GetFundQuote 基金实时净值。wire: fund_quote
func (c *QuoteClient) GetFundQuote(req model.FundQuoteRequest) ([]model.FundQuote, error) {
	var out []model.FundQuote
	err := c.callInto("fund_quote", req, &out)
	return out, err
}

// GetFundHistoryQuote 基金历史净值。wire: fund_history_quote
func (c *QuoteClient) GetFundHistoryQuote(req model.FundHistoryQuoteRequest) ([]model.FundHistoryQuote, error) {
	var out []model.FundHistoryQuote
	err := c.callInto("fund_history_quote", req, &out)
	return out, err
}

// GetWarrantQuote returns real-time warrant quotes. wire: warrant_briefs
func (c *QuoteClient) GetWarrantQuote(req model.WarrantBriefsRequest) ([]model.WarrantBrief, error) {
	var out []model.WarrantBrief
	err := c.callInto("warrant_briefs", req, &out)
	return out, err
}

// Deprecated: Use GetWarrantQuote instead.
func (c *QuoteClient) GetWarrantBriefs(req model.WarrantBriefsRequest) ([]model.WarrantBrief, error) {
	return c.GetWarrantQuote(req)
}

// GetWarrantFilter 窝轮筛选。wire: warrant_filter
func (c *QuoteClient) GetWarrantFilter(req model.WarrantFilterRequest) (*model.WarrantFilterResult, error) {
	var out model.WarrantFilterResult
	if err := c.callInto("warrant_filter", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetIndustryList 行业列表。wire: industry_list
func (c *QuoteClient) GetIndustryList(req model.IndustryListRequest) ([]model.IndustryItem, error) {
	var out []model.IndustryItem
	err := c.callInto("industry_list", req, &out)
	return out, err
}

// GetIndustryStocks 行业下股票列表。wire: industry_stock_list
func (c *QuoteClient) GetIndustryStocks(req model.IndustryStocksRequest) ([]model.IndustryStock, error) {
	var out []model.IndustryStock
	err := c.callInto("industry_stock_list", req, &out)
	return out, err
}

// GetCorporateSplit 公司行动 - 拆股。wire: corporate_action (action_type=split)
// 服务端返回 {SYMBOL: [items]} 按 symbol 分组的 map。
func (c *QuoteClient) GetCorporateSplit(req model.CorporateActionRequest) ([]model.CorporateAction, error) {
	req.ActionType = "split"
	var grouped map[string][]model.CorporateAction
	if err := c.callInto("corporate_action", req, &grouped); err != nil {
		return nil, err
	}
	var out []model.CorporateAction
	for _, list := range grouped {
		out = append(out, list...)
	}
	return out, nil
}

// GetCorporateDividend 公司行动 - 分红。wire: corporate_action (action_type=dividend)
func (c *QuoteClient) GetCorporateDividend(req model.CorporateActionRequest) ([]model.CorporateAction, error) {
	req.ActionType = "dividend"
	var grouped map[string][]model.CorporateAction
	if err := c.callInto("corporate_action", req, &grouped); err != nil {
		return nil, err
	}
	var out []model.CorporateAction
	for _, list := range grouped {
		out = append(out, list...)
	}
	return out, nil
}

// GetCorporateEarningsCalendar 公司行动 - 财报日历。wire: corporate_action (action_type=earning)
func (c *QuoteClient) GetCorporateEarningsCalendar(req model.CorporateActionRequest) ([]model.CorporateAction, error) {
	req.ActionType = "earning"
	var grouped map[string][]model.CorporateAction
	if err := c.callInto("corporate_action", req, &grouped); err != nil {
		return nil, err
	}
	var out []model.CorporateAction
	for _, list := range grouped {
		out = append(out, list...)
	}
	return out, nil
}

// GetFinancialExchangeRate 汇率数据。wire: financial_exchange_rate
func (c *QuoteClient) GetFinancialExchangeRate(req model.FinancialExchangeRateRequest) ([]model.ExchangeRate, error) {
	var out []model.ExchangeRate
	err := c.callInto("financial_exchange_rate", req, &out)
	return out, err
}

// GetFinancialCurrency 财报币种。wire: financial_currency
func (c *QuoteClient) GetFinancialCurrency(req model.FinancialCurrencyRequest) ([]model.FinancialCurrency, error) {
	var out []model.FinancialCurrency
	err := c.callInto("financial_currency", req, &out)
	return out, err
}

// GetTradingCalendar 交易日历。wire: trading_calendar
func (c *QuoteClient) GetTradingCalendar(req model.TradingCalendarRequest) ([]model.TradingCalendarItem, error) {
	var out []model.TradingCalendarItem
	err := c.callInto("trading_calendar", req, &out)
	return out, err
}

// GetMarketScannerTags 扫描器可用标签。wire: market_scanner_tags
// Server 返回数组：[{market, multiTagField, tagList:[{field,name,values}]}]
func (c *QuoteClient) GetMarketScannerTags(req model.MarketScannerTagsRequest) ([]model.MarketScannerTagGroup, error) {
	var out []model.MarketScannerTagGroup
	if err := c.callInto("market_scanner_tags", req, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetQuoteOvernight 隔夜行情。wire: quote_overnight
func (c *QuoteClient) GetQuoteOvernight(req model.QuoteOvernightRequest) ([]model.QuoteOvernight, error) {
	var out []model.QuoteOvernight
	err := c.callInto("quote_overnight", req, &out)
	return out, err
}
