// Package quote provides a quote query client wrapping all market data APIs.
package quote

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/tigerfintech/openapi-go-sdk/client"
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

// === Basic market data methods ===

// GetMarketState returns market status.
func (c *QuoteClient) GetMarketState(market string) ([]model.MarketState, error) {
	var out []model.MarketState
	err := c.callInto("market_state", map[string]interface{}{"market": market}, &out)
	return out, err
}

// GetBrief returns real-time quotes.
func (c *QuoteClient) GetBrief(symbols []string) ([]model.Brief, error) {
	var out []model.Brief
	err := c.callInto("quote_real_time", map[string]interface{}{"symbols": symbols}, &out)
	return out, err
}

// GetKline returns K-line (candlestick) data.
func (c *QuoteClient) GetKline(symbol, period string) ([]model.Kline, error) {
	var out []model.Kline
	err := c.callInto("kline", map[string]interface{}{
		"symbols": []string{symbol},
		"period":  period,
	}, &out)
	return out, err
}

// GetTimeline returns intraday timeline data.
func (c *QuoteClient) GetTimeline(symbols []string) ([]model.Timeline, error) {
	var out []model.Timeline
	err := c.callInto("timeline", map[string]interface{}{"symbols": symbols}, &out)
	return out, err
}

// GetTradeTick returns tick-by-tick trade data.
func (c *QuoteClient) GetTradeTick(symbols []string) ([]model.TradeTick, error) {
	var out []model.TradeTick
	err := c.callInto("trade_tick", map[string]interface{}{"symbols": symbols}, &out)
	return out, err
}

// GetQuoteDepth returns order book depth data.
// market: required, e.g. "US"/"HK".
func (c *QuoteClient) GetQuoteDepth(symbol, market string) ([]model.Depth, error) {
	var out []model.Depth
	err := c.callInto("quote_depth", map[string]interface{}{
		"symbols": []string{symbol},
		"market":  market,
	}, &out)
	return out, err
}

// === Option market data methods ===

// GetOptionExpiration returns option expiration dates.
func (c *QuoteClient) GetOptionExpiration(symbol string) ([]model.OptionExpiration, error) {
	var out []model.OptionExpiration
	err := c.callInto("option_expiration", map[string]interface{}{"symbols": []string{symbol}}, &out)
	return out, err
}

// GetOptionChain returns the option chain for a symbol and expiry date.
// expiry: "YYYY-MM-DD" string.
func (c *QuoteClient) GetOptionChain(symbol, expiry string) ([]model.OptionChain, error) {
	expiryTs, err := parseOptionExpiry(expiry)
	if err != nil {
		return nil, fmt.Errorf("invalid expiry date: %w", err)
	}
	params := map[string]interface{}{
		"option_basic": []map[string]interface{}{
			{
				"symbol": symbol,
				"expiry": expiryTs,
			},
		},
	}
	var out []model.OptionChain
	err = c.callIntoVersioned("option_chain", params, "3.0", &out)
	return out, err
}

// GetOptionBrief returns option quotes for the given identifiers.
func (c *QuoteClient) GetOptionBrief(identifiers []string) ([]model.Brief, error) {
	optionBasics := make([]map[string]interface{}, 0, len(identifiers))
	for _, id := range identifiers {
		contract, err := optionContractFromIdentifier(id)
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

// GetOptionKline returns option K-line data.
func (c *QuoteClient) GetOptionKline(identifier, period string) ([]model.Kline, error) {
	contract, err := optionContractFromIdentifier(identifier)
	if err != nil {
		return nil, fmt.Errorf("invalid option identifier %q: %w", identifier, err)
	}
	params := map[string]interface{}{
		"option_query": []map[string]interface{}{
			{
				"symbol": contract.Symbol,
				"expiry": contract.Expiry,
				"right":  contract.Right,
				"strike": contract.Strike,
				"period": period,
			},
		},
	}
	var out []model.Kline
	err = c.callIntoVersioned("option_kline", params, "2.0", &out)
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
// contractCodes: future contract codes, e.g. ["CL2609"].
func (c *QuoteClient) GetFutureRealTimeQuote(contractCodes []string) ([]model.FutureQuote, error) {
	var out []model.FutureQuote
	err := c.callInto("future_real_time_quote", map[string]interface{}{"contract_codes": contractCodes}, &out)
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

// === Option helper functions ===

// optionContract holds parsed fields from an OCC option identifier.
type optionContract struct {
	Symbol string
	Expiry int64  // millisecond timestamp
	Right  string // "CALL" or "PUT"
	Strike float64
}

// parseOptionExpiry parses a date string like "2024-01-19" into a millisecond timestamp (UTC).
func parseOptionExpiry(expiry string) (int64, error) {
	t, err := time.Parse("2006-01-02", expiry)
	if err != nil {
		return 0, fmt.Errorf("expected format YYYY-MM-DD, got %q: %w", expiry, err)
	}
	return t.UnixMilli(), nil
}

// optionContractFromIdentifier parses an OCC-style option identifier like "AAPL 240119C00150000"
// into its component parts: symbol, expiry (ms timestamp), right (CALL/PUT), strike.
func optionContractFromIdentifier(identifier string) (optionContract, error) {
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
	t, err := time.Parse("060102", dateStr)
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
