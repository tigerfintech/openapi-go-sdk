// Package quote provides a quote query client wrapping all market data APIs.
package quote

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/tigerfintech/openapi-go-sdk/client"
)

// QuoteClient wraps all market data API calls.
type QuoteClient struct {
	httpClient *client.HttpClient
}

// NewQuoteClient creates a quote query client.
func NewQuoteClient(httpClient *client.HttpClient) *QuoteClient {
	return &QuoteClient{httpClient: httpClient}
}

// execute is the internal helper: build request, send, return data field.
func (c *QuoteClient) execute(method string, bizParams interface{}) (json.RawMessage, error) {
	req, err := client.NewApiRequest(method, bizParams)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Execute(req)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// executeVersioned is like execute but with a specific API version.
func (c *QuoteClient) executeVersioned(method string, bizParams interface{}, version string) (json.RawMessage, error) {
	req, err := client.NewVersionedApiRequest(method, bizParams, version)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Execute(req)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// === Basic market data methods ===

// GetMarketState returns market status.
func (c *QuoteClient) GetMarketState(market string) (json.RawMessage, error) {
	params := map[string]interface{}{"market": market}
	return c.execute("market_state", params)
}

// GetBrief returns real-time quotes.
func (c *QuoteClient) GetBrief(symbols []string) (json.RawMessage, error) {
	params := map[string]interface{}{"symbols": symbols}
	return c.execute("quote_real_time", params)
}

// GetKline returns K-line (candlestick) data.
func (c *QuoteClient) GetKline(symbol, period string) (json.RawMessage, error) {
	params := map[string]interface{}{
		"symbols": []string{symbol},
		"period":  period,
	}
	return c.execute("kline", params)
}

// GetTimeline returns intraday timeline data.
func (c *QuoteClient) GetTimeline(symbols []string) (json.RawMessage, error) {
	params := map[string]interface{}{"symbols": symbols}
	return c.execute("timeline", params)
}

// GetTradeTick returns tick-by-tick trade data.
func (c *QuoteClient) GetTradeTick(symbols []string) (json.RawMessage, error) {
	params := map[string]interface{}{"symbols": symbols}
	return c.execute("trade_tick", params)
}

// GetQuoteDepth returns order book depth data.
func (c *QuoteClient) GetQuoteDepth(symbol string) (json.RawMessage, error) {
	params := map[string]interface{}{"symbol": symbol}
	return c.execute("quote_depth", params)
}

// === Option market data methods ===

// GetOptionExpiration returns option expiration dates.
func (c *QuoteClient) GetOptionExpiration(symbol string) (json.RawMessage, error) {
	params := map[string]interface{}{"symbols": []string{symbol}}
	return c.execute("option_expiration", params)
}

// GetOptionChain returns the option chain for a symbol and expiry date.
// Uses API v3 with option_basic array format and parsed expiry timestamp.
func (c *QuoteClient) GetOptionChain(symbol, expiry string) (json.RawMessage, error) {
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
	return c.executeVersioned("option_chain", params, "3.0")
}

// GetOptionBrief returns option quotes for the given identifiers.
// Uses API v2 with parsed identifier fields.
func (c *QuoteClient) GetOptionBrief(identifiers []string) (json.RawMessage, error) {
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
	params := map[string]interface{}{
		"option_basic": optionBasics,
	}
	return c.executeVersioned("option_brief", params, "2.0")
}

// GetOptionKline returns option K-line data.
// Uses API v2 with option_query array format.
func (c *QuoteClient) GetOptionKline(identifier, period string) (json.RawMessage, error) {
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
	return c.executeVersioned("option_kline", params, "2.0")
}

// === Futures market data methods ===

// GetFutureExchange returns the list of futures exchanges.
func (c *QuoteClient) GetFutureExchange() (json.RawMessage, error) {
	params := map[string]interface{}{"sec_type": "FUT"}
	return c.execute("future_exchange", params)
}

// GetFutureContracts returns futures contracts for an exchange.
func (c *QuoteClient) GetFutureContracts(exchange string) (json.RawMessage, error) {
	params := map[string]interface{}{"exchange": exchange}
	return c.execute("future_contracts", params)
}

// GetFutureRealTimeQuote returns real-time futures quotes.
func (c *QuoteClient) GetFutureRealTimeQuote(symbols []string) (json.RawMessage, error) {
	params := map[string]interface{}{"symbols": symbols}
	return c.execute("future_real_time_quote", params)
}

// GetFutureKline returns futures K-line data.
func (c *QuoteClient) GetFutureKline(symbol, period string) (json.RawMessage, error) {
	params := map[string]interface{}{
		"symbol": symbol,
		"period": period,
	}
	return c.execute("future_kline", params)
}

// === Fundamentals and capital flow methods ===

// GetFinancialDaily returns daily financial data.
func (c *QuoteClient) GetFinancialDaily(symbol string) (json.RawMessage, error) {
	params := map[string]interface{}{"symbol": symbol}
	return c.execute("financial_daily", params)
}

// GetFinancialReport returns financial reports.
func (c *QuoteClient) GetFinancialReport(symbol string) (json.RawMessage, error) {
	params := map[string]interface{}{"symbol": symbol}
	return c.execute("financial_report", params)
}

// GetCorporateAction returns corporate actions.
func (c *QuoteClient) GetCorporateAction(symbol string) (json.RawMessage, error) {
	params := map[string]interface{}{"symbol": symbol}
	return c.execute("corporate_action", params)
}

// GetCapitalFlow returns capital flow data.
func (c *QuoteClient) GetCapitalFlow(symbol string) (json.RawMessage, error) {
	params := map[string]interface{}{"symbol": symbol}
	return c.execute("capital_flow", params)
}

// GetCapitalDistribution returns capital distribution data.
func (c *QuoteClient) GetCapitalDistribution(symbol string) (json.RawMessage, error) {
	params := map[string]interface{}{"symbol": symbol}
	return c.execute("capital_distribution", params)
}

// === Scanner and quote permission methods ===

// MarketScanner runs the stock screener.
func (c *QuoteClient) MarketScanner(params map[string]interface{}) (json.RawMessage, error) {
	return c.execute("market_scanner", params)
}

// GrabQuotePermission requests quote permissions.
func (c *QuoteClient) GrabQuotePermission() (json.RawMessage, error) {
	return c.execute("grab_quote_permission", nil)
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
	// Format: "SYMBOL YYMMDDX00000000" where X is C or P
	// The second part is always 15 characters: 6 date + 1 right + 8 strike
	parts := strings.SplitN(strings.TrimSpace(identifier), " ", 2)
	if len(parts) != 2 {
		return optionContract{}, fmt.Errorf("expected format 'SYMBOL YYMMDDX00000000', got %q", identifier)
	}

	symbol := parts[0]
	rest := parts[1]
	if len(rest) < 15 {
		return optionContract{}, fmt.Errorf("option code too short: %q", rest)
	}

	// Parse date: YYMMDD
	dateStr := rest[:6]
	t, err := time.Parse("060102", dateStr)
	if err != nil {
		return optionContract{}, fmt.Errorf("invalid date in identifier: %q", dateStr)
	}

	// Parse right: C or P
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

	// Parse strike: 8 digits, divide by 1000 to get actual price
	strikeStr := rest[7:]
	var strikeInt int64
	for _, c := range strikeStr {
		if c < '0' || c > '9' {
			return optionContract{}, fmt.Errorf("invalid strike digits: %q", strikeStr)
		}
		strikeInt = strikeInt*10 + int64(c-'0')
	}
	strike := float64(strikeInt) / 1000.0

	return optionContract{
		Symbol: symbol,
		Expiry: t.UnixMilli(),
		Right:  right,
		Strike: strike,
	}, nil
}
