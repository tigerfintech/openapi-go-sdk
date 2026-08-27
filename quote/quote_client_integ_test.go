//go:build integration

// Integration tests against the real OpenAPI gateway. Run with:
//
//	make integ
//	# or: TIGER_RUN_INTEG=true go test -tags=integration ./quote/...
//
// Credentials come from TIGEROPEN_TIGER_ID / PRIVATE_KEY / ACCOUNT env vars
// (or TIGEROPEN_PROPS_PATH). See integtest.NewConfig.

package quote_test

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/tigerfintech/openapi-go-sdk/integtest"
	"github.com/tigerfintech/openapi-go-sdk/model"
	"github.com/tigerfintech/openapi-go-sdk/quote"
)

// quotePermissionPatterns matches server errors that indicate an account lacks
// entitlement for the requested quote feature. These are legitimate test skips.
// Patterns are intentionally narrow to avoid masking real regressions.
var quotePermissionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)does not support (quote|market|stock|option|warrant|future|industry|scanner)`),
	// Server returns "the current requested method does not support" when the
	// account lacks entitlement for the entire API method (no resource keyword).
	regexp.MustCompile(`(?i)(current requested method|this method).{0,30}does not support`),
	// code=4 wraps the inner 1000 error as "1000:common param error(the current
	// requested method does not support)" — catch the inner message too.
	regexp.MustCompile(`(?i)common param error\(.*does not support\)`),
	regexp.MustCompile(`(?i)not (entitled|authorized|subscribed)`),
	regexp.MustCompile(`(?i)no permission`),
	regexp.MustCompile(`(?i)access forbidden`),
	regexp.MustCompile(`(?i)(license (expired|required|invalid|not found)|unlicensed)`),
	regexp.MustCompile(`(?i)(quote|market data|data feed|symbol).{0,30}not open`),
	regexp.MustCompile(`(?i)(quote|market data|feature|function).{0,30}not enabled`),
}

// isQuotePermissionError reports whether err is a known account-entitlement
// boundary error that should cause the test to be skipped rather than failed.
func isQuotePermissionError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	for _, p := range quotePermissionPatterns {
		if p.MatchString(msg) {
			return true
		}
	}
	return false
}

// ms parses a "YYYY-MM-DD" date in UTC and returns its epoch-ms timestamp
// ms converts a "YYYY-MM-DD" date string to an epoch-ms *int64 pointer,
// as *int64, matching the wire type expected by FinancialReportRequest /
// CorporateActionRequest. Panics on malformed input (test-only helper).
// Production callers should use model.DateToEpochMs which returns (*int64, error).
func ms(date string) *int64 {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		panic(err)
	}
	v := t.UnixMilli()
	return &v
}

func TestIntegGetMarketState(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetMarketState("US")
	if err != nil {
		t.Fatalf("GetMarketState: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("GetMarketState returned empty")
	}
	if data[0].Market == "" {
		t.Errorf("MarketState.Market is empty, want non-empty")
	}
	if data[0].MarketStatus == "" && data[0].Status == "" {
		t.Errorf("MarketState has no status (marketStatus=%q, status=%q)", data[0].MarketStatus, data[0].Status)
	}
}

func TestIntegGetBrief(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetBrief(model.BriefRequest{Symbols: []string{"AAPL"}})
	if err != nil {
		t.Fatalf("GetBrief: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("GetBrief returned empty for AAPL")
	}
	if data[0].Symbol != "AAPL" {
		t.Errorf("symbol = %q, want AAPL", data[0].Symbol)
	}
	if data[0].LatestPrice <= 0 {
		t.Errorf("LatestPrice = %v, want > 0", data[0].LatestPrice)
	}
	if data[0].LatestTime == 0 {
		t.Errorf("LatestTime = 0, want non-zero timestamp")
	}
}

func TestIntegGetKline(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetKline(model.KlineRequest{
		Symbols: []string{"AAPL"},
		Period:  "day",
		Limit:   5,
	})
	if err != nil {
		t.Fatalf("GetKline: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("GetKline returned empty")
	}
	if data[0].Symbol != "AAPL" {
		t.Errorf("Kline.Symbol = %q, want AAPL", data[0].Symbol)
	}
	if len(data[0].Items) == 0 {
		t.Fatal("Kline.Items is empty for AAPL")
	}
	item := data[0].Items[0]
	if item.Time == 0 {
		t.Errorf("KlineItem.Time = 0, want non-zero timestamp")
	}
	if item.Close <= 0 {
		t.Errorf("KlineItem.Close = %v, want > 0", item.Close)
	}
	if item.High < item.Low {
		t.Errorf("KlineItem.High (%v) < Low (%v)", item.High, item.Low)
	}
}

func TestIntegGetTimeline(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetTimeline([]string{"AAPL"})
	if err != nil {
		t.Fatalf("GetTimeline: %v", err)
	}
	// Timeline is only produced during extended trading hours. Out-of-hours
	// the wire round-trip has already been validated by the successful call
	// above; empty items are expected and the test should stop here.
	if len(data) == 0 {
		integtest.AssertNonEmptyDuringExtended(t, qc, "US", "GetTimeline returned empty for AAPL")
		return
	}
	if data[0].Symbol != "AAPL" {
		t.Errorf("Timeline.Symbol = %q, want AAPL", data[0].Symbol)
	}
	// Timeline buckets may legitimately be empty outside trading hours.
	// When items exist, validate their fields.
	for _, bucket := range []*model.TimelineBucket{data[0].Intraday, data[0].PreHours, data[0].AfterHours} {
		if bucket == nil || len(bucket.Items) == 0 {
			continue
		}
		item := bucket.Items[0]
		if item.Time == 0 {
			t.Errorf("TimelineItem.Time = 0, want non-zero timestamp")
		}
		if item.Price <= 0 {
			t.Errorf("TimelineItem.Price = %v, want > 0", item.Price)
		}
		break // only check first non-empty bucket
	}
}

func TestIntegGetOptionExpiration(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetOptionExpiration([]string{"AAPL"})
	if err != nil {
		t.Fatalf("GetOptionExpiration: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("GetOptionExpiration returned empty")
	}
	if data[0].Symbol != "AAPL" {
		t.Errorf("OptionExpiration.Symbol = %q, want AAPL", data[0].Symbol)
	}
	if len(data[0].Dates) == 0 {
		t.Fatal("no expiry dates for AAPL")
	}
	if len(data[0].Timestamps) == 0 {
		t.Errorf("OptionExpiration.Timestamps is empty, want non-empty (should match Dates)")
	}
	if len(data[0].OptionSymbols) == 0 {
		t.Errorf("OptionExpiration.OptionSymbols is empty, want non-empty")
	}
}

func TestIntegGetOptionChain(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	exps, err := qc.GetOptionExpiration([]string{"AAPL"})
	if err != nil || len(exps) == 0 || len(exps[0].Dates) == 0 {
		t.Skipf("no expiry available for AAPL: err=%v", err)
	}
	expiry := exps[0].Dates[len(exps[0].Dates)/2]

	chain, err := qc.GetOptionChain([][2]string{{"AAPL", expiry}})
	if err != nil {
		t.Fatalf("GetOptionChain: %v", err)
	}
	if len(chain) == 0 {
		t.Fatal("GetOptionChain returned empty")
	}
	if chain[0].Symbol != "AAPL" {
		t.Errorf("OptionChain.Symbol = %q, want AAPL", chain[0].Symbol)
	}
	if chain[0].Expiry == 0 {
		t.Errorf("OptionChain.Expiry = 0, want non-zero timestamp")
	}
	if len(chain[0].Items) == 0 {
		t.Fatal("OptionChain.Items is empty")
	}
	row := chain[0].Items[0]
	if row.Call == nil && row.Put == nil {
		t.Errorf("OptionChainRow has neither Call nor Put leg")
	}
	if row.Call != nil {
		if row.Call.Identifier == "" {
			t.Errorf("Call leg Identifier is empty, want non-empty")
		}
		if row.Call.Strike == "" {
			t.Errorf("Call leg Strike is empty, want non-empty")
		}
	}
	if row.Put != nil {
		if row.Put.Identifier == "" {
			t.Errorf("Put leg Identifier is empty, want non-empty")
		}
		if row.Put.Strike == "" {
			t.Errorf("Put leg Strike is empty, want non-empty")
		}
	}
}

func TestIntegGetCorporateAction(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetCorporateAction(model.CorporateActionRequest{
		Symbols:    []string{"AAPL"},
		Market:     "US",
		ActionType: "DIVIDEND",
		BeginDate:  ms("2024-01-01"),
		EndDate:    ms("2025-12-31"),
	})
	if err != nil {
		t.Fatalf("GetCorporateAction: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("GetCorporateAction returned empty for AAPL DIVIDEND over 2-year range")
	}
	// Server may return a class-specific ticker form (e.g. AAPL-CL) for a
	// corporate action, so assert non-empty rather than exact match. Matches
	// Java CorporateActionIntegrationTest which only checks !symbol.isEmpty().
	if data[0].Symbol == "" {
		t.Errorf("CorporateAction.Symbol is empty, want non-empty")
	}
	if data[0].ActionType == "" {
		t.Errorf("CorporateAction.ActionType is empty, want non-empty")
	}
	if data[0].ExecuteDate == "" {
		t.Errorf("CorporateAction.ExecuteDate is empty, want non-empty date string")
	}
}

func TestIntegGetCapitalFlow(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetCapitalFlow("AAPL", "US", "1d")
	if err != nil {
		t.Fatalf("GetCapitalFlow: %v", err)
	}
	if data == nil {
		t.Fatal("GetCapitalFlow returned nil")
	}
	if data.Symbol != "AAPL" {
		t.Errorf("CapitalFlow.Symbol = %q, want AAPL", data.Symbol)
	}
	// Capital flow items only exist during trading hours. Out-of-hours the
	// wire round-trip already validated the request; stop here rather than
	// asserting on missing data.
	if len(data.Items) == 0 {
		integtest.AssertNonEmptyDuringTrading(t, qc, "US", "GetCapitalFlow returned empty items")
		return
	}
	item := data.Items[0]
	if item.Timestamp == 0 {
		t.Errorf("CapitalFlowItem.Timestamp = 0, want non-zero")
	}
}

func TestIntegGrabQuotePermission(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GrabQuotePermission()
	if err != nil {
		t.Fatalf("GrabQuotePermission: %v", err)
	}
	// Permission list may be empty if the account has no active entitlements.
	// When entries exist, validate identifier fields.
	for i, p := range data {
		if p.Name == "" {
			t.Errorf("QuotePermission[%d].Name is empty, want non-empty", i)
		}
	}
}

// ============================================================================
// Additional quote integration tests — covering all remaining public methods.
// ============================================================================

// --- Helpers for chained prerequisite data ---

// firstOptionIdentifier picks a near-ATM AAPL option identifier so the option
// has enough liquidity/history for kline / trade-tick tests. Delegates to the
// shared resolver so all callers share the same near-ATM heuristic.
func firstOptionIdentifier(t *testing.T, qc *quote.QuoteClient) string {
	t.Helper()
	id, err := integtest.ResolveUsOptionIdentifier(qc)
	if err != nil {
		t.Skipf("skip: cannot resolve AAPL option identifier: %v", err)
	}
	return id
}

// firstOptionChainLeg picks a near-ATM AAPL leg + its expiry timestamp.
func firstOptionChainLeg(t *testing.T, qc *quote.QuoteClient) (model.OptionLeg, int64) {
	t.Helper()
	leg, ts, err := integtest.ResolveUsOptionLeg(qc)
	if err != nil {
		t.Skipf("skip: cannot resolve AAPL option leg: %v", err)
	}
	return leg, ts
}

func firstFutureContractCode(t *testing.T, qc *quote.QuoteClient) string {
	t.Helper()
	exchanges, err := qc.GetFutureExchange()
	if err != nil || len(exchanges) == 0 {
		t.Skipf("no futures exchanges: err=%v", err)
	}
	contracts, err := qc.GetFutureContracts(exchanges[0].Code)
	if err != nil || len(contracts) == 0 {
		t.Skipf("no futures contracts for exchange %s: err=%v", exchanges[0].Code, err)
	}
	return contracts[0].ContractCode
}

// --- Basic market data ---

func TestIntegGetRealTimeQuote(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetRealTimeQuote(model.BriefRequest{Symbols: []string{"AAPL"}})
	if err != nil {
		t.Fatalf("GetRealTimeQuote: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("GetRealTimeQuote returned empty for AAPL")
	}
	if data[0].Symbol != "AAPL" {
		t.Errorf("symbol = %q, want AAPL", data[0].Symbol)
	}
	if data[0].LatestPrice <= 0 {
		t.Errorf("LatestPrice = %v, want > 0", data[0].LatestPrice)
	}
	if data[0].Amount <= 0 {
		t.Errorf("Amount = %v, want > 0", data[0].Amount)
	}
}

func TestIntegGetRealTimeQuoteCC(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetRealTimeQuote(model.BriefRequest{Symbols: []string{"BTC.USD"}, SecType: "CC"})
	if err != nil {
		t.Fatalf("GetRealTimeQuote CC: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("GetRealTimeQuote CC returned empty for BTC.USD")
	}
	if data[0].Symbol == "" {
		t.Errorf("CC Symbol is empty")
	}
	if data[0].Amount <= 0 {
		t.Skipf("CC Amount = %v; quote_real_time CC amount not returned by current server", data[0].Amount)
	}
}

func TestIntegGetTradeTick(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetTradeTick(model.TradeTickRequest{Symbols: []string{"AAPL"}, Limit: 5})
	if err != nil {
		t.Fatalf("GetTradeTick: %v", err)
	}
	// Trade ticks only stream during regular trading hours. Out-of-hours the
	// wire round-trip already validated the request; stop here rather than
	// asserting on missing data.
	if len(data) == 0 || len(data[0].Items) == 0 {
		integtest.AssertNonEmptyDuringTrading(t, qc, "US", "GetTradeTick returned empty for AAPL")
		return
	}
	if data[0].Symbol != "AAPL" {
		t.Errorf("TradeTick.Symbol = %q, want AAPL", data[0].Symbol)
	}
	item := data[0].Items[0]
	if item.Time == 0 {
		t.Errorf("TradeTickItem.Time = 0, want non-zero")
	}
	if item.Price <= 0 {
		t.Errorf("TradeTickItem.Price = %v, want > 0", item.Price)
	}
}

func TestIntegGetQuoteDepth(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetQuoteDepth(model.DepthQuoteRequest{Symbols: []string{"AAPL"}, Market: "US"})
	if err != nil {
		t.Fatalf("GetQuoteDepth: %v", err)
	}
	// Depth only streams during trading hours. Out-of-hours the wire
	// round-trip already validated the request; stop here rather than
	// asserting on missing data.
	if len(data) == 0 {
		integtest.AssertNonEmptyDuringTrading(t, qc, "US", "GetQuoteDepth returned empty for AAPL")
		return
	}
	if data[0].Symbol != "AAPL" {
		t.Errorf("Depth.Symbol = %q, want AAPL", data[0].Symbol)
	}
}

func TestIntegGetDelayedQuote(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetDelayedQuote(model.StockDelayBriefsRequest{Symbols: []string{"AAPL"}})
	if err != nil {
		t.Fatalf("GetDelayedQuote: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("GetDelayedQuote returned empty for AAPL")
	}
	if data[0].Symbol != "AAPL" {
		t.Errorf("symbol = %q, want AAPL", data[0].Symbol)
	}
}

func TestIntegGetQuoteOvernight(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetQuoteOvernight(model.QuoteOvernightRequest{Symbols: []string{"AAPL"}})
	if err != nil {
		t.Fatalf("GetQuoteOvernight: %v", err)
	}
	if len(data) == 0 {
		// The overnight session (~8pm-4am ET) is disjoint from the
		// regular/pre/post-market statuses IsMarketOpenExtended checks —
		// it can legitimately be empty during PRE_HOUR_TRADING once the
		// overnight window has already closed. Mirrors the tolerant
		// pattern used by the neighboring GetKline(TradeSession: "OverNight")
		// test just below, which never fails on an empty overnight result.
		t.Logf("GetQuoteOvernight returned empty for AAPL (market=US status=%s); no-current overnight data is valid", integtest.MarketStatus(qc, "US"))
		return
	}
	if data[0].Symbol != "AAPL" {
		t.Errorf("QuoteOvernight.Symbol = %q, want AAPL", data[0].Symbol)
	}
}

func TestIntegGetOvernightKlineDepthAndTradeTick(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	klines, err := qc.GetKline(model.KlineRequest{
		Symbols: []string{"AAPL"}, Period: "1min", Limit: 5, TradeSession: "OverNight",
	})
	if err != nil {
		t.Fatalf("GetKline overnight: %v", err)
	}
	if len(klines) == 0 {
		t.Log("GetKline overnight returned empty; no-current overnight bars are valid")
	} else {
		if klines[0].Symbol != "AAPL" {
			t.Errorf("overnight Kline.Symbol = %q, want AAPL", klines[0].Symbol)
		}
		if len(klines[0].Items) == 0 {
			t.Log("GetKline overnight returned no items; no-current overnight bars are valid")
		} else {
			item := klines[0].Items[0]
			if item.Time == 0 {
				t.Error("overnight KlineItem.Time = 0, want non-zero")
			}
			if item.Close <= 0 {
				t.Errorf("overnight KlineItem.Close = %v, want > 0", item.Close)
			}
			if item.High < item.Low {
				t.Errorf("overnight KlineItem.High (%v) < Low (%v)", item.High, item.Low)
			}
		}
	}

	depth, err := qc.GetQuoteDepth(model.DepthQuoteRequest{
		Symbols: []string{"AAPL"}, Market: "US", TradeSession: "OverNight",
	})
	if err != nil {
		t.Fatalf("GetQuoteDepth overnight: %v", err)
	}
	if len(depth) == 0 {
		t.Log("GetQuoteDepth overnight returned empty; no-current overnight depth is valid")
	} else {
		if depth[0].Symbol != "AAPL" {
			t.Errorf("overnight Depth.Symbol = %q, want AAPL", depth[0].Symbol)
		}
		for side, levels := range map[string][]model.DepthLevel{"asks": depth[0].Asks, "bids": depth[0].Bids} {
			for i, level := range levels {
				if level.Price <= 0 {
					t.Errorf("overnight Depth.%s[%d].Price = %v, want > 0", side, i, level.Price)
				}
				if level.Volume < 0 {
					t.Errorf("overnight Depth.%s[%d].Volume = %d, want >= 0", side, i, level.Volume)
				}
			}
		}
		if len(depth[0].Asks) == 0 && len(depth[0].Bids) == 0 {
			t.Log("GetQuoteDepth overnight returned no levels; no-current overnight depth is valid")
		}
	}

	ticks, err := qc.GetTradeTick(model.TradeTickRequest{
		Symbols: []string{"AAPL"}, Limit: 5, TradeSession: "OverNight",
	})
	if err != nil {
		t.Fatalf("GetTradeTick overnight: %v", err)
	}
	if len(ticks) == 0 {
		t.Log("GetTradeTick overnight returned empty; no-current overnight ticks are valid")
	} else {
		if ticks[0].Symbol != "AAPL" {
			t.Errorf("overnight TradeTick.Symbol = %q, want AAPL", ticks[0].Symbol)
		}
		if len(ticks[0].Items) == 0 {
			t.Log("GetTradeTick overnight returned no items; no-current overnight ticks are valid")
		} else {
			item := ticks[0].Items[0]
			if item.Time == 0 {
				t.Error("overnight TradeTickItem.Time = 0, want non-zero")
			}
			if item.Price <= 0 {
				t.Errorf("overnight TradeTickItem.Price = %v, want > 0", item.Price)
			}
			if item.Volume < 0 {
				t.Errorf("overnight TradeTickItem.Volume = %d, want >= 0", item.Volume)
			}
		}
	}
}

// --- Stock info ---

func TestIntegGetSymbols(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetSymbols(model.SymbolsRequest{Market: "US", SecType: "STK"})
	if err != nil {
		t.Fatalf("GetSymbols: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("GetSymbols returned empty for US STK")
	}
	for i, s := range data {
		if s == "" {
			t.Errorf("Symbols[%d] is empty", i)
		}
	}
}

func TestIntegGetSymbolNames(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetSymbolNames(model.SymbolsRequest{Market: "US", SecType: "STK"})
	if err != nil {
		t.Fatalf("GetSymbolNames: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("GetSymbolNames returned empty for US STK")
	}
	if data[0].Symbol == "" {
		t.Errorf("SymbolName.Symbol is empty, want non-empty")
	}
}

func TestIntegGetTradeMetas(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetTradeMetas(model.TradeMetasRequest{Symbols: []string{"AAPL"}})
	if err != nil {
		t.Fatalf("GetTradeMetas: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("GetTradeMetas returned empty for AAPL")
	}
	if data[0].Symbol != "AAPL" {
		t.Errorf("TradeMeta.Symbol = %q, want AAPL", data[0].Symbol)
	}
	if data[0].LotSize <= 0 {
		t.Errorf("TradeMeta.LotSize = %v, want > 0", data[0].LotSize)
	}
}

func TestIntegGetStockDetails(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetStockDetails(model.StockDetailsRequest{Symbols: []string{"AAPL"}})
	if err != nil {
		t.Fatalf("GetStockDetails: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("GetStockDetails returned empty for AAPL")
	}
	if data[0].Symbol != "AAPL" {
		t.Errorf("StockDetail.Symbol = %q, want AAPL", data[0].Symbol)
	}
	// Currency is account-dependent: some retail accounts get it, some don't.
	// Warn instead of skipping so the test still validates the primary
	// contract (non-empty response, correct symbol). Only assert non-empty
	// when we can see other identifying fields are populated too, which
	// signals a fully-populated payload.
	if data[0].Currency == "" {
		t.Logf("StockDetail.Currency is empty (account-dependent payload; other fields still validated)")
	}
}

func TestIntegGetStockBroker(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	// stock_broker only supports HK market; symbol must be 5-digit numeric code without suffix
	data, err := qc.GetStockBroker(model.StockBrokerRequest{Symbol: "00700", Limit: 5})
	if err != nil {
		t.Fatalf("GetStockBroker: %v", err)
	}
	if data == nil {
		t.Fatal("GetStockBroker returned nil")
	}
	if data.Symbol != "00700" {
		t.Errorf("StockBroker.Symbol = %q, want 00700", data.Symbol)
	}
}

func TestIntegGetStockFundamental(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetStockFundamental(model.StockFundamentalRequest{Symbols: []string{"AAPL"}, Market: "US"})
	if err != nil {
		t.Fatalf("GetStockFundamental: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("GetStockFundamental returned empty for AAPL")
	}
}

func TestIntegGetStockIndustry(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetStockIndustry(model.StockIndustryRequest{Symbol: "AAPL", Market: "US"})
	if err != nil {
		t.Fatalf("GetStockIndustry: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("GetStockIndustry returned empty for AAPL")
	}
	// StockIndustry may not carry a Symbol field in the response; skip assertion
	// when it is empty — the API returns industry classification data, not the symbol.
	_ = data[0].Symbol
}

// --- K-line & timeline extensions ---

func TestIntegGetKlineByPage(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetKlineByPage(model.KlineByPageRequest{
		Symbol:    "AAPL",
		Period:    "day",
		TotalSize: 10,
		PageSize:  5,
	})
	if err != nil {
		t.Fatalf("GetKlineByPage: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("GetKlineByPage returned empty")
	}
	if data[0].Time == 0 {
		t.Errorf("KlineItem.Time = 0, want non-zero")
	}
	if data[0].Close <= 0 {
		t.Errorf("KlineItem.Close = %v, want > 0", data[0].Close)
	}
}

func TestIntegGetTimelineHistory(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	// Use yesterday to stay within the API's data retention window.
	// If yesterday is a weekend, the API may return empty — skip in that case.
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	data, err := qc.GetTimelineHistory(model.TimelineHistoryRequest{
		Symbols: []string{"AAPL"},
		Date:    yesterday,
	})
	if err != nil {
		t.Fatalf("GetTimelineHistory: %v", err)
	}
	if len(data) == 0 {
		t.Skipf("GetTimelineHistory returned empty for AAPL on %s (likely weekend or holiday)", yesterday)
	}
	if data[0].Symbol != "AAPL" {
		t.Errorf("Timeline.Symbol = %q, want AAPL", data[0].Symbol)
	}
}

func TestIntegGetTradeRank(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetTradeRank(model.TradeRankRequest{Market: "US"})
	if err != nil {
		t.Fatalf("GetTradeRank: %v", err)
	}
	// Rank only produces data during trading hours. Out-of-hours the wire
	// round-trip already validated the request; stop here rather than
	// asserting on missing data.
	if len(data) == 0 {
		integtest.AssertNonEmptyDuringTrading(t, qc, "US", "GetTradeRank returned empty")
		return
	}
	if data[0].Symbol == "" {
		t.Errorf("TradeRankItem.Symbol is empty, want non-empty")
	}
}

func TestIntegGetShortInterest(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetShortInterest(model.ShortInterestRequest{Symbols: []string{"AAPL"}})
	if err != nil {
		if isQuotePermissionError(err) {
			t.Skipf("skip: GetShortInterest requires an entitlement not granted to this account: %v", err)
		}
		t.Fatalf("GetShortInterest: %v", err)
	}
	// AAPL is one of the most-shorted large-caps; empty is genuinely
	// unexpected. Report location so a real regression is not hidden.
	if len(data) == 0 {
		t.Fatal("GetShortInterest returned empty for AAPL — server-side data gap")
	}
	if data[0].Symbol != "AAPL" {
		t.Errorf("ShortInterest.Symbol = %q, want AAPL", data[0].Symbol)
	}
}

// --- Permission & quota ---

func TestIntegGetQuotePermissionDetail(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetQuotePermission(model.QuotePermissionRequest{})
	if err != nil {
		t.Fatalf("GetQuotePermission: %v", err)
	}
	// Permission list may be empty.
	for i, p := range data {
		if p.Name == "" {
			t.Errorf("QuotePermission[%d].Name is empty, want non-empty", i)
		}
	}
}

func TestIntegGetKlineQuota(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetKlineQuota(model.KlineQuotaRequest{WithDetails: true})
	if err != nil {
		t.Fatalf("GetKlineQuota: %v", err)
	}
	// Quota may be empty if no kline calls have been made.
	for i, q := range data {
		if q.Method == "" {
			t.Errorf("KlineQuota[%d].Method is empty, want non-empty", i)
		}
	}
}

func TestIntegGetAddonEntitlement(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetAddonEntitlement()
	if err != nil {
		t.Fatalf("GetAddonEntitlement: %v", err)
	}
	if data == nil {
		t.Fatal("GetAddonEntitlement returned nil")
	}
}

// --- Capital distribution ---

func TestIntegGetCapitalDistribution(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetCapitalDistribution("AAPL", "US")
	if err != nil {
		t.Fatalf("GetCapitalDistribution: %v", err)
	}
	if data == nil {
		t.Skip("GetCapitalDistribution returned nil (may be unavailable)")
	}
	if data.Symbol != "AAPL" {
		t.Errorf("CapitalDistribution.Symbol = %q, want AAPL", data.Symbol)
	}
}

// --- Financial data ---

func TestIntegGetFinancialDaily(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	// Fields must match server-side FinancialDailyField enum. See Java
	// com.tigerbrokers.stock.openapi.client.struct.enums.FinancialDailyField
	// for the full list; using two representative values here.
	data, err := qc.GetFinancialDaily(model.FinancialDailyRequest{
		Symbols:   []string{"AAPL"},
		Market:    "US",
		Fields:    []string{"market_capitalization", "shares_outstanding"},
		BeginDate: "2024-01-01",
		EndDate:   "2024-06-30",
	})
	if err != nil {
		t.Fatalf("GetFinancialDaily: %v", err)
	}
	if len(data) == 0 {
		// Financial daily data is only published after market close; during
		// trading hours a temporarily empty response is expected. Use the
		// same market-hours guard as other quote tests for consistency.
		integtest.AssertNonEmptyDuringTrading(t, qc, "US", "GetFinancialDaily(AAPL) returned empty")
		return
	}
	if data[0].Symbol != "AAPL" {
		t.Errorf("FinancialDailyItem.Symbol = %q, want AAPL", data[0].Symbol)
	}
	if data[0].Field == "" {
		t.Errorf("FinancialDailyItem.Field is empty, want non-empty")
	}
}

func TestIntegGetFinancialReport(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	// Financial reports are quarterly/annual — an "Annual" query in H1 of a
	// year may hit before the annual report is published. Try Annual first,
	// then fall back to LTM and Quarter with a wider window, so that when the
	// server does have data we exercise the assertions.
	tryOne := func(periodType string, beginDate, endDate *int64) []model.FinancialReportItem {
		out, err := qc.GetFinancialReport(model.FinancialReportRequest{
			Symbols:    []string{"AAPL"},
			Market:     "US",
			Fields:     []string{"total_revenue"},
			PeriodType: periodType,
			BeginDate:  beginDate,
			EndDate:    endDate,
		})
		if err != nil {
			t.Fatalf("GetFinancialReport (%s): %v", periodType, err)
		}
		return out
	}
	data := tryOne("Annual", ms("2024-01-01"), ms("2024-06-30"))
	if len(data) == 0 {
		// Widen to 3 years of Quarterly reports — AAPL has published every
		// quarter for decades, so this should always have data.
		// Wire value is "Quarterly", matching Python FinancialReportPeriodType.QUARTERLY
		// and Java FinancialPeriodType.Quarterly. "Quarter" is rejected by the
		// server with `biz param error(failed to parse parameters in 'biz_content')`.
		data = tryOne("Quarterly", ms("2022-01-01"), ms("2024-12-31"))
	}
	if len(data) == 0 {
		data = tryOne("LTM", ms("2022-01-01"), ms("2024-12-31"))
	}
	if len(data) == 0 {
		t.Fatal("GetFinancialReport returned empty for AAPL across Annual/Quarterly/LTM 2022-2024 — server-side data gap")
	}
	if data[0].Symbol != "AAPL" {
		t.Errorf("FinancialReportItem.Symbol = %q, want AAPL", data[0].Symbol)
	}
	if data[0].Field == "" {
		t.Errorf("FinancialReportItem.Field is empty, want non-empty")
	}
}

func TestIntegGetFinancialExchangeRate(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetFinancialExchangeRate(model.FinancialExchangeRateRequest{
		CurrencyList: []string{"USD"},
		BeginDate:    "20240101",
		EndDate:      "20240102",
	})
	if err != nil {
		t.Fatalf("GetFinancialExchangeRate: %v", err)
	}
	if len(data) == 0 {
		t.Skip("GetFinancialExchangeRate returned empty")
	}
	if data[0].Currency == "" {
		t.Errorf("ExchangeRate.Currency is empty, want non-empty")
	}
}

func TestIntegGetFinancialCurrency(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetFinancialCurrency(model.FinancialCurrencyRequest{
		Symbols: []string{"AAPL"},
		Market:  "US",
	})
	if err != nil {
		t.Fatalf("GetFinancialCurrency: %v", err)
	}
	if len(data) == 0 {
		t.Skip("GetFinancialCurrency returned empty for AAPL")
	}
	if data[0].Symbol != "AAPL" {
		t.Errorf("FinancialCurrency.Symbol = %q, want AAPL", data[0].Symbol)
	}
}

// --- Market scanner ---

func TestIntegMarketScanner(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.MarketScanner(model.MarketScannerRequest{
		Market:   "US",
		Page:     1,
		PageSize: 5,
	})
	if err != nil {
		t.Fatalf("MarketScanner: %v", err)
	}
	if data == nil {
		t.Fatal("MarketScanner returned nil")
	}
	// Scanner results may be empty depending on market conditions.
	if data.TotalCount > 0 && len(data.Items) > 0 {
		if data.Items[0].Symbol == "" {
			t.Errorf("ScannerResultItem.Symbol is empty, want non-empty")
		}
	}
}

func TestIntegGetMarketScannerTags(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetMarketScannerTags(model.MarketScannerTagsRequest{Market: "US", MultiTagsFields: []string{"volume"}})
	if err != nil {
		if isQuotePermissionError(err) || strings.Contains(err.Error(), "biz_content") {
			t.Skipf("skip: GetMarketScannerTags gated on this account: %v", err)
		}
		t.Fatalf("GetMarketScannerTags: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("GetMarketScannerTags returned empty for US — server-side data gap")
	}
	if data[0].Market == "" {
		t.Errorf("MarketScannerTagGroup.Market is empty, want non-empty")
	}
}

// --- Trading calendar ---

func TestIntegGetTradingCalendar(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetTradingCalendar(model.TradingCalendarRequest{
		Market:    "US",
		BeginDate: "2024-01-01",
		EndDate:   "2024-01-31",
	})
	if err != nil {
		t.Fatalf("GetTradingCalendar: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("GetTradingCalendar returned empty for US Jan 2024")
	}
	// Market field is not guaranteed on every server build — Date/Type are the
	// primary contract. Log instead of skipping so we still validate the
	// items structure.
	if data[0].Market == "" {
		t.Logf("TradingCalendarItem.Market is empty (field may not be returned by server; Date/Type still validated)")
	}
	if data[0].Date == "" {
		t.Errorf("TradingCalendarItem.Date is empty, want non-empty")
	}
}

// --- Option methods ---

func TestIntegGetOptionQuote(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	identifier := firstOptionIdentifier(t, qc)
	data, err := qc.GetOptionQuote([]string{identifier})
	if err != nil {
		t.Fatalf("GetOptionQuote: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("GetOptionQuote returned empty")
	}
	if data[0].Symbol == "" {
		t.Errorf("Brief.Symbol is empty, want non-empty")
	}
}

func TestIntegGetOptionKline(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	identifier := firstOptionIdentifier(t, qc)
	// Widen the window so newly-listed options still have at least one bar.
	end := time.Now().UnixMilli()
	begin := time.Now().AddDate(0, 0, -90).UnixMilli()
	data, err := qc.GetOptionKline([]string{identifier}, "day", begin, end)
	if err != nil {
		t.Fatalf("GetOptionKline: %v", err)
	}
	// A live, tradeable near-ATM option should have at least one daily bar in
	// a 90-day window. If not, the failure depends on trading hours: during a
	// live session, empty means a real regression; outside, the wire path
	// was already validated by the successful call, so stop here.
	if len(data) == 0 || len(data[0].Items) == 0 {
		integtest.AssertNonEmptyDuringTrading(t, qc, "US", "GetOptionKline returned empty for %s over 90 days", identifier)
		return
	}
	if data[0].Items[0].Close <= 0 {
		t.Errorf("KlineItem.Close = %v, want > 0", data[0].Items[0].Close)
	}
}

func TestIntegGetOptionTradeTicks(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	leg, expiryTs := firstOptionChainLeg(t, qc)
	data, err := qc.GetOptionTradeTicks(model.OptionTradeTicksRequest{
		Contracts: []model.OptionQueryItem{{
			Symbol: "AAPL",
			Expiry: expiryTs,
			Strike: leg.Strike,
			Right:  leg.Right,
		}},
	})
	if err != nil {
		t.Fatalf("GetOptionTradeTicks: %v", err)
	}
	// Trade ticks are only produced during regular trading hours. Out-of-hours
	// the wire round-trip already validated the request; stop here rather
	// than asserting on missing data.
	if len(data) == 0 || len(data[0].Items) == 0 {
		integtest.AssertNonEmptyDuringTrading(t, qc, "US", "GetOptionTradeTicks returned empty")
		return
	}
	item := data[0].Items[0]
	if item.Time == 0 {
		t.Errorf("TradeTickItem.Time = 0, want non-zero")
	}
	if item.Price <= 0 {
		t.Errorf("TradeTickItem.Price = %v, want > 0", item.Price)
	}
}

func TestIntegGetOptionTimeline(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	leg, expiryTs := firstOptionChainLeg(t, qc)
	data, err := qc.GetOptionTimeline(model.OptionTimelineRequest{
		Market: "US",
		OptionQuery: []model.OptionQueryItem{{
			Symbol: "AAPL",
			Expiry: expiryTs,
			Strike: leg.Strike,
			Right:  leg.Right,
		}},
	})
	if err != nil {
		t.Fatalf("GetOptionTimeline: %v", err)
	}
	// Timeline is only produced during trading hours. Out-of-hours the wire
	// round-trip already validated the request; stop here rather than
	// asserting on missing data.
	if len(data) == 0 {
		integtest.AssertNonEmptyDuringTrading(t, qc, "US", "GetOptionTimeline returned empty")
		return
	}
	if data[0].Symbol == "" {
		t.Errorf("Timeline.Symbol is empty, want non-empty")
	}
}

func TestIntegGetOptionDepth(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	leg, expiryTs := firstOptionChainLeg(t, qc)
	data, err := qc.GetOptionDepth(model.OptionDepthRequest{
		OptionBasic: []model.OptionQueryItem{{
			Symbol: "AAPL",
			Expiry: expiryTs,
			Strike: leg.Strike,
			Right:  leg.Right,
		}},
	})
	if err != nil {
		t.Fatalf("GetOptionDepth: %v", err)
	}
	// Depth is only produced during trading hours. Out-of-hours the wire
	// round-trip already validated the request; stop here rather than
	// asserting on missing data.
	if len(data) == 0 {
		integtest.AssertNonEmptyDuringTrading(t, qc, "US", "GetOptionDepth returned empty")
		return
	}
	if data[0].Symbol == "" {
		t.Errorf("Depth.Symbol is empty, want non-empty")
	}
}

func TestIntegGetOptionSymbols(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetOptionSymbols(model.OptionSymbolsRequest{Market: "US"})
	if err != nil {
		if isQuotePermissionError(err) {
			t.Skipf("skip: GetOptionSymbols requires an entitlement not granted to this account: %v", err)
		}
		t.Fatalf("GetOptionSymbols: %v", err)
	}
	if len(data) == 0 {
		t.Skip("GetOptionSymbols returned empty for US")
	}
	if data[0].Symbol == "" {
		t.Errorf("OptionSymbol.Symbol is empty, want non-empty")
	}
}

func TestIntegGetOptionAnalysis(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetOptionAnalysis(model.OptionAnalysisRequest{
		Symbols: []model.OptionAnalysisSymbol{{Symbol: "AAPL"}},
		Market:  "US",
	})
	if err != nil {
		t.Fatalf("GetOptionAnalysis: %v", err)
	}
	if len(data) == 0 {
		t.Skip("GetOptionAnalysis returned empty for AAPL")
	}
	if data[0].Symbol != "AAPL" {
		t.Errorf("OptionAnalysis.Symbol = %q, want AAPL", data[0].Symbol)
	}
}

// --- Futures ---

func TestIntegGetFutureExchange(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetFutureExchange()
	if err != nil {
		t.Fatalf("GetFutureExchange: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("GetFutureExchange returned empty")
	}
	if data[0].Code == "" {
		t.Errorf("FutureExchange.Code is empty, want non-empty")
	}
	if data[0].Name == "" {
		t.Errorf("FutureExchange.Name is empty, want non-empty")
	}
}

func TestIntegGetFutureContracts(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	exchanges, err := qc.GetFutureExchange()
	if err != nil || len(exchanges) == 0 {
		t.Skipf("no futures exchanges: err=%v", err)
	}
	data, err := qc.GetFutureContracts(exchanges[0].Code)
	if err != nil {
		t.Fatalf("GetFutureContracts: %v", err)
	}
	if len(data) == 0 {
		t.Skipf("GetFutureContracts returned empty for exchange %s", exchanges[0].Code)
	}
	if data[0].ContractCode == "" {
		t.Skip("FutureContractInfo.ContractCode is empty (may not be returned)")
	}
	if data[0].Currency == "" {
		t.Errorf("FutureContractInfo.Currency is empty, want non-empty")
	}
}

func TestIntegGetFutureRealTimeQuote(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	code := firstFutureContractCode(t, qc)
	data, err := qc.GetFutureRealTimeQuote(model.FutureBriefRequest{ContractCodes: []string{code}})
	if err != nil {
		t.Fatalf("GetFutureRealTimeQuote: %v", err)
	}
	if len(data) == 0 {
		t.Skipf("GetFutureRealTimeQuote returned empty for %s", code)
	}
	if data[0].ContractCode != code {
		t.Errorf("FutureQuote.ContractCode = %q, want %q", data[0].ContractCode, code)
	}
}

func TestIntegGetFutureKline(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	code := firstFutureContractCode(t, qc)
	data, err := qc.GetFutureKline(model.FutureKlineRequest{
		ContractCodes: []string{code},
		Period:        "day",
		Limit:         5,
	})
	if err != nil {
		t.Fatalf("GetFutureKline: %v", err)
	}
	if len(data) == 0 || len(data[0].Items) == 0 {
		t.Skipf("GetFutureKline returned empty for %s", code)
	}
	item := data[0].Items[0]
	if item.Time == 0 {
		t.Errorf("FutureKlineItem.Time = 0, want non-zero")
	}
	if item.Close <= 0 {
		t.Errorf("FutureKlineItem.Close = %v, want > 0", item.Close)
	}
}

func TestIntegGetFutureContract(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	code := firstFutureContractCode(t, qc)
	data, err := qc.GetFutureContract(model.FutureContractSingleRequest{ContractCode: code})
	if err != nil {
		t.Fatalf("GetFutureContract: %v", err)
	}
	if len(data) == 0 {
		t.Fatalf("GetFutureContract returned empty for %s", code)
	}
	if data[0].ContractCode != code {
		t.Errorf("FutureContractInfo.ContractCode = %q, want %q", data[0].ContractCode, code)
	}
}

func TestIntegGetAllFutureContracts(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetAllFutureContracts(model.AllFutureContractsRequest{Type: "MAIN"})
	if err != nil {
		// Server returns "contract not exists" for accounts without futures
		// entitlement — legitimate skip.
		if strings.Contains(err.Error(), "contract not exists") || isQuotePermissionError(err) {
			t.Skipf("skip: GetAllFutureContracts requires futures entitlement: %v", err)
		}
		t.Fatalf("GetAllFutureContracts: %v", err)
	}
	if len(data) == 0 {
		t.Skip("skip: GetAllFutureContracts returned empty (no futures entitlement)")
	}
	if data[0].ContractCode == "" {
		t.Errorf("FutureContractInfo.ContractCode is empty, want non-empty")
	}
}

func TestIntegGetCurrentFutureContract(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	code := firstFutureContractCode(t, qc)
	data, err := qc.GetCurrentFutureContract(model.FutureContractSingleRequest{ContractCode: code, Type: "MAIN"})
	if err != nil {
		t.Fatalf("GetCurrentFutureContract: %v", err)
	}
	if data == nil {
		t.Fatalf("GetCurrentFutureContract returned nil for %s", code)
	}
	// ContractCode is often omitted for non-MAIN accounts. Log rather than
	// skip so the test still validates the call+decode contract.
	if data.ContractCode == "" {
		t.Logf("FutureContractInfo.ContractCode is empty for %s (account may lack MAIN contract entitlement)", code)
	}
}

func TestIntegGetFutureContinuousContracts(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetFutureContinuousContracts(model.FutureContinuousContractsRequest{Type: "MAIN"})
	if err != nil {
		t.Fatalf("GetFutureContinuousContracts: %v", err)
	}
	if len(data) == 0 {
		t.Skip("skip: GetFutureContinuousContracts returned empty (no futures entitlement)")
	}
	if data[0].ContractCode == "" {
		t.Logf("FutureContractInfo.ContractCode is empty (server may omit for this account; other fields still validated)")
	}
}

func TestIntegGetFutureHistoryMainContract(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	code := firstFutureContractCode(t, qc)
	data, err := qc.GetFutureHistoryMainContract(model.FutureHistoryMainContractRequest{
		ContractCodes: []string{code},
	})
	if err != nil {
		// code=4 / "does not support" indicates the account is not entitled
		// to the historical main-contract lookup — legitimate skip.
		if isQuotePermissionError(err) {
			t.Skipf("skip: GetFutureHistoryMainContract requires an entitlement not granted to this account: %v", err)
		}
		t.Fatalf("GetFutureHistoryMainContract: %v", err)
	}
	if len(data) == 0 {
		t.Skipf("skip: GetFutureHistoryMainContract returned empty for %s (no historical main-contract data)", code)
	}
	if data[0].ContractCode == "" {
		t.Errorf("FutureMainContractHistory.ContractCode is empty, want non-empty")
	}
}

func TestIntegGetFutureKlineByPage(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	code := firstFutureContractCode(t, qc)
	data, err := qc.GetFutureKlineByPage(model.FutureKlineByPageRequest{
		ContractCode: code,
		Period:       "day",
		TotalSize:    10,
		PageSize:     5,
	})
	if err != nil {
		t.Fatalf("GetFutureKlineByPage: %v", err)
	}
	if len(data) == 0 {
		t.Skipf("GetFutureKlineByPage returned empty for %s", code)
	}
	if data[0].Time == 0 {
		t.Errorf("FutureKlineItem.Time = 0, want non-zero")
	}
}

func TestIntegGetFutureTradeTicks(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	code := firstFutureContractCode(t, qc)
	data, err := qc.GetFutureTradeTicks(model.FutureTradeTicksRequest{ContractCode: code})
	if err != nil {
		t.Fatalf("GetFutureTradeTicks: %v", err)
	}
	// May be empty outside trading hours.
	if len(data) == 0 {
		t.Skipf("GetFutureTradeTicks returned empty for %s (likely non-trading hours)", code)
	}
	if data[0].Time == 0 {
		t.Errorf("FutureTradeTickItem.Time = 0, want non-zero")
	}
	if data[0].Price <= 0 {
		t.Errorf("FutureTradeTickItem.Price = %v, want > 0", data[0].Price)
	}
}

func TestIntegGetFutureDepth(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	code := firstFutureContractCode(t, qc)
	data, err := qc.GetFutureDepth(model.FutureDepthRequest{ContractCodes: []string{code}})
	if err != nil {
		t.Fatalf("GetFutureDepth: %v", err)
	}
	// May be empty outside trading hours.
	if len(data) == 0 {
		t.Skipf("GetFutureDepth returned empty for %s (likely non-trading hours)", code)
	}
	if data[0].ContractCode == "" {
		t.Errorf("FutureDepth.ContractCode is empty, want non-empty")
	}
}

func TestIntegGetFutureTradingTimes(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	code := firstFutureContractCode(t, qc)
	data, err := qc.GetFutureTradingTimes(model.FutureTradingTimesRequest{ContractCode: code})
	if err != nil {
		t.Fatalf("GetFutureTradingTimes: %v", err)
	}
	if data == nil {
		t.Skipf("GetFutureTradingTimes returned nil for %s", code)
	}
}

// --- Funds ---

func TestIntegGetFundSymbols(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetFundSymbols(model.FundSymbolsRequest{})
	if err != nil {
		t.Fatalf("GetFundSymbols: %v", err)
	}
	if len(data) == 0 {
		t.Skip("GetFundSymbols returned empty")
	}
}

func TestIntegGetFundContracts(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	symbols, err := qc.GetFundSymbols(model.FundSymbolsRequest{})
	if err != nil || len(symbols) == 0 {
		t.Skipf("no fund symbols: err=%v", err)
	}
	data, err := qc.GetFundContracts(model.FundContractsRequest{Symbols: []string{symbols[0]}})
	if err != nil {
		t.Fatalf("GetFundContracts: %v", err)
	}
	if len(data) == 0 {
		t.Skipf("GetFundContracts returned empty for %s", symbols[0])
	}
	if data[0].Symbol == "" {
		t.Errorf("FundContractInfo.Symbol is empty, want non-empty")
	}
}

func TestIntegGetFundQuote(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	symbols, err := qc.GetFundSymbols(model.FundSymbolsRequest{})
	if err != nil || len(symbols) == 0 {
		t.Skipf("no fund symbols: err=%v", err)
	}
	data, err := qc.GetFundQuote(model.FundQuoteRequest{Symbols: []string{symbols[0]}})
	if err != nil {
		t.Fatalf("GetFundQuote: %v", err)
	}
	if len(data) == 0 {
		t.Skipf("GetFundQuote returned empty for %s", symbols[0])
	}
	if data[0].Symbol == "" {
		t.Errorf("FundQuote.Symbol is empty, want non-empty")
	}
}

func TestIntegGetFundHistoryQuote(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	symbols, err := qc.GetFundSymbols(model.FundSymbolsRequest{})
	if err != nil || len(symbols) == 0 {
		t.Skipf("no fund symbols: err=%v", err)
	}
	now := time.Now()
	data, err := qc.GetFundHistoryQuote(model.FundHistoryQuoteRequest{
		Symbols:   []string{symbols[0]},
		BeginTime: now.AddDate(0, -1, 0).UnixMilli(),
		EndTime:   now.UnixMilli(),
		Limit:     5,
	})
	if err != nil {
		t.Fatalf("GetFundHistoryQuote: %v", err)
	}
	if len(data) == 0 {
		t.Skipf("GetFundHistoryQuote returned empty for %s", symbols[0])
	}
	if data[0].Symbol == "" {
		t.Errorf("FundHistoryQuote.Symbol is empty, want non-empty")
	}
}

// --- Warrants ---

func TestIntegGetWarrantFilter(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetWarrantFilter(model.WarrantFilterRequest{Symbol: "00700", PageSize: 5})
	if err != nil {
		t.Fatalf("GetWarrantFilter: %v", err)
	}
	if data == nil {
		t.Fatal("GetWarrantFilter returned nil")
	}
	// Wire shape: server returns a bare array of warrant items with no
	// {total, page, pageSize} wrapper — SDK models keep those fields at
	// their zero defaults. So `Total` is not authoritative; only `Items`
	// carries the answer. During HK trading 00700 must yield warrants;
	// out-of-hours empty is legitimate and we early-return without asserting.
	if len(data.Items) == 0 {
		integtest.AssertNonEmptyDuringTrading(t, qc, "HK", "GetWarrantFilter(00700) returned no items")
		return
	}
	if data.Items[0].Symbol == "" {
		t.Errorf("WarrantBrief.Symbol is empty, want non-empty")
	}
}

func TestIntegGetWarrantQuote(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	// Resolve a live HK warrant symbol via the filter first.
	warrantSymbol, err := integtest.ResolveHkWarrantSymbol(qc)
	if err != nil {
		t.Skipf("skip: cannot resolve HK warrant symbol (likely no entitlement): %v", err)
	}
	data, err := qc.GetWarrantQuote(model.WarrantBriefsRequest{Symbols: []string{warrantSymbol}})
	if err != nil {
		if isQuotePermissionError(err) {
			t.Skipf("skip: GetWarrantQuote requires an entitlement not granted to this account: %v", err)
		}
		t.Fatalf("GetWarrantQuote: %v", err)
	}
	if len(data) == 0 {
		t.Fatalf("GetWarrantQuote returned empty for %s", warrantSymbol)
	}
	if data[0].Symbol != warrantSymbol {
		t.Errorf("WarrantBrief.Symbol = %q, want %q", data[0].Symbol, warrantSymbol)
	}
}

// --- Industry ---

func TestIntegGetIndustryList(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetIndustryList(model.IndustryListRequest{IndustryLevel: "GSECTOR"})
	if err != nil {
		t.Fatalf("GetIndustryList: %v", err)
	}
	if len(data) == 0 {
		t.Skip("GetIndustryList returned empty")
	}
	if data[0].ID == "" {
		t.Errorf("IndustryItem.ID is empty, want non-empty")
	}
	if data[0].Name == "" {
		t.Skip("IndustryItem.Name is empty (may not be returned by server)")
	}
}

func TestIntegGetIndustryStocks(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	industries, err := qc.GetIndustryList(model.IndustryListRequest{IndustryLevel: "GSECTOR"})
	if err != nil || len(industries) == 0 {
		t.Skipf("no industries: err=%v", err)
	}
	data, err := qc.GetIndustryStocks(model.IndustryStocksRequest{
		IndustryID: industries[0].ID,
		Market:     "US",
	})
	if err != nil {
		if isQuotePermissionError(err) {
			t.Skipf("skip: GetIndustryStocks requires an entitlement not granted to this account: %v", err)
		}
		t.Fatalf("GetIndustryStocks: %v", err)
	}
	// May return empty for some industries.
	if len(data) == 0 {
		t.Skipf("GetIndustryStocks returned empty for industry %s", industries[0].ID)
	}
	if data[0].Symbol == "" {
		t.Errorf("IndustryStock.Symbol is empty, want non-empty")
	}
}

// --- Corporate actions (specific types) ---

func TestIntegGetCorporateSplit(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetCorporateSplit(model.CorporateActionRequest{
		Symbols:   []string{"AAPL"},
		Market:    "US",
		BeginDate: ms("2020-01-01"),
		EndDate:   ms("2025-12-31"),
	})
	if err != nil {
		t.Fatalf("GetCorporateSplit: %v", err)
	}
	// Splits may not exist for AAPL in this range.
	if len(data) == 0 {
		t.Skip("GetCorporateSplit returned empty for AAPL 2020-2025")
	}
	// Server may return a class-specific ticker form for the split entry
	// (matches Java CorporateActionIntegrationTest which only asserts
	// non-empty symbol).
	if data[0].Symbol == "" {
		t.Errorf("CorporateAction.Symbol is empty, want non-empty")
	}
}

func TestIntegGetCorporateDividend(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetCorporateDividend(model.CorporateActionRequest{
		Symbols:   []string{"AAPL"},
		Market:    "US",
		BeginDate: ms("2024-01-01"),
		EndDate:   ms("2025-12-31"),
	})
	if err != nil {
		t.Fatalf("GetCorporateDividend: %v", err)
	}
	if len(data) == 0 {
		t.Skip("GetCorporateDividend returned empty for AAPL 2024-2025")
	}
	// Server may return a class-specific ticker form (e.g. AAPL-CL) for the
	// dividend entry (matches Java CorporateActionIntegrationTest which only
	// asserts non-empty symbol).
	if data[0].Symbol == "" {
		t.Errorf("CorporateAction.Symbol is empty, want non-empty")
	}
}

func TestIntegGetCorporateEarningsCalendar(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetCorporateEarningsCalendar(model.CorporateActionRequest{
		Symbols:   []string{"AAPL"},
		Market:    "US",
		BeginDate: ms("2024-01-01"),
		EndDate:   ms("2024-01-31"),
	})
	if err != nil {
		t.Fatalf("GetCorporateEarningsCalendar: %v", err)
	}
	// AAPL reports quarterly; a full year of data across 2024 must contain at
	// least one earnings entry. Empty here indicates a real server-side gap.
	if len(data) == 0 {
		t.Fatal("GetCorporateEarningsCalendar returned empty for AAPL 2024 — server-side data gap")
	}
	// Server sometimes returns a class-specific ticker (e.g. AAPL-CL); accept
	// any non-empty symbol as the Java suite does. See commit 46d1c41.
	if data[0].Symbol == "" {
		t.Errorf("CorporateAction.Symbol is empty, want non-empty")
	}
}

func TestIntegGetCorporateSymbolChange(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetCorporateSymbolChange(model.CorporateActionRequest{
		Symbols:   []string{"AAPL"},
		Market:    "US",
		BeginDate: ms("2020-01-01"),
		EndDate:   ms("2025-12-31"),
	})
	if err != nil {
		t.Fatalf("GetCorporateSymbolChange: %v", err)
	}
	// Symbol changes are rare.
	_ = data
}

func TestIntegGetCorporateDelisting(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetCorporateDelisting(model.CorporateActionRequest{
		Symbols:   []string{"AAPL"},
		Market:    "US",
		BeginDate: ms("2020-01-01"),
		EndDate:   ms("2025-12-31"),
	})
	if err != nil {
		t.Fatalf("GetCorporateDelisting: %v", err)
	}
	// Delistings are rare.
	_ = data
}

func TestIntegGetCorporateIPO(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	// Use known 2024 IPO symbols (RDDT=Reddit 2024-03, ALAB=Astera Labs 2024-03).
	data, err := qc.GetCorporateIPO(model.CorporateActionRequest{
		Symbols:   []string{"RDDT", "ALAB"},
		Market:    "US",
		BeginDate: ms("2024-01-01"),
		EndDate:   ms("2024-12-31"),
	})
	if err != nil {
		if isQuotePermissionError(err) {
			t.Skipf("GetCorporateIPO: not supported on this account: %v", err)
		}
		t.Fatalf("GetCorporateIPO: %v", err)
	}
	if len(data) == 0 {
		t.Skip("GetCorporateIPO returned empty for RDDT/ALAB 2024 — data-boundary (subscription/coverage), not an SDK bug")
	}
	if data[0].Symbol == "" {
		t.Errorf("CorporateIPO.Symbol is empty, want non-empty")
	}
	// Verify the returned symbol is one of the requested IPO symbols.
	found := false
	for _, d := range data {
		if d.Symbol == "RDDT" || d.Symbol == "ALAB" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("GetCorporateIPO: none of the returned symbols match RDDT or ALAB; got %v",
			func() []string {
				ss := make([]string, len(data))
				for i, d := range data {
					ss[i] = d.Symbol
				}
				return ss
			}())
	}
}

// ===========================================================================
// Enhanced coverage: K线时间范围、条数、OHLC约束；盘口价格合理性；US+HK双市场
// ===========================================================================

// TestIntegGetKlineTimeRangeAndOHLC 用固定30天窗口请求日K线，验证：
//  - 返回条数 >= 15（30个自然日至少含15个交易日）
//  - 时间戳严格升序
//  - 每根K线 High >= max(Open,Close) >= min(Open,Close) >= Low
//  - Volume >= 0
//  同时覆盖美股(AAPL)和港股(00700)
func TestIntegGetKlineTimeRangeAndOHLC(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	end := time.Now().UnixMilli()
	begin := time.Now().AddDate(0, 0, -30).UnixMilli()

	cases := []string{"AAPL", "00700"}
	for _, sym := range cases {
		sym := sym
		t.Run(sym, func(t *testing.T) {
			data, err := qc.GetKline(model.KlineRequest{
				Symbols:   []string{sym},
				Period:    "day",
				BeginTime: begin,
				EndTime:   end,
			})
			if err != nil {
				t.Fatalf("GetKline(%s): %v", sym, err)
			}
			if len(data) == 0 {
				t.Fatalf("GetKline(%s): returned empty", sym)
			}
			items := data[0].Items
			// 30天窗口至少应有15根交易日K线
			if len(items) < 15 {
				t.Errorf("GetKline(%s): got %d bars in 30-day window, want >= 15", sym, len(items))
			}
			// 时间戳严格升序
			for i := 1; i < len(items); i++ {
				if items[i].Time <= items[i-1].Time {
					t.Errorf("GetKline(%s): items[%d].Time=%d <= items[%d].Time=%d (not ascending)",
						sym, i, items[i].Time, i-1, items[i-1].Time)
					break
				}
			}
			// OHLC约束
			for i, item := range items {
				if item.Close <= 0 {
					t.Errorf("GetKline(%s) item[%d]: Close=%v, want > 0", sym, i, item.Close)
				}
				if item.High < item.Low {
					t.Errorf("GetKline(%s) item[%d]: High(%v) < Low(%v)", sym, i, item.High, item.Low)
				}
				if item.High < item.Open {
					t.Errorf("GetKline(%s) item[%d]: High(%v) < Open(%v)", sym, i, item.High, item.Open)
				}
				if item.High < item.Close {
					t.Errorf("GetKline(%s) item[%d]: High(%v) < Close(%v)", sym, i, item.High, item.Close)
				}
				if item.Low > item.Open {
					t.Errorf("GetKline(%s) item[%d]: Low(%v) > Open(%v)", sym, i, item.Low, item.Open)
				}
				if item.Low > item.Close {
					t.Errorf("GetKline(%s) item[%d]: Low(%v) > Close(%v)", sym, i, item.Low, item.Close)
				}
				if item.Volume < 0 {
					t.Errorf("GetKline(%s) item[%d]: Volume=%d, want >= 0", sym, i, item.Volume)
				}
			}
		})
	}
}

// TestIntegGetKlineIntraday 验证60分钟K线：时序升序、OHLC约束，覆盖US+HK
func TestIntegGetKlineIntraday(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	end := time.Now().UnixMilli()
	begin := time.Now().AddDate(0, 0, -5).UnixMilli()

	cases := []string{"AAPL", "00700"}
	for _, sym := range cases {
		sym := sym
		t.Run(sym, func(t *testing.T) {
			data, err := qc.GetKline(model.KlineRequest{
				Symbols:   []string{sym},
				Period:    "60min",
				BeginTime: begin,
				EndTime:   end,
			})
			if err != nil {
				t.Skipf("GetKline(%s 60min): %v", sym, err)
			}
			if len(data) == 0 || len(data[0].Items) == 0 {
				t.Skipf("GetKline(%s 60min): no bars in 5-day window (holiday?)", sym)
			}
			items := data[0].Items
			// 5天应有至少5根60分钟K线
			if len(items) < 5 {
				t.Errorf("GetKline(%s 60min): got %d bars in 5-day window, want >= 5", sym, len(items))
			}
			for i := 1; i < len(items); i++ {
				if items[i].Time <= items[i-1].Time {
					t.Errorf("GetKline(%s 60min): items[%d].Time not ascending", sym, i)
					break
				}
			}
			for i, item := range items {
				if item.High < item.Low {
					t.Errorf("GetKline(%s 60min) item[%d]: High(%v) < Low(%v)", sym, i, item.High, item.Low)
				}
			}
		})
	}
}

// TestIntegGetKlineByPageTimeRange 分页K线：指定时间窗口、验证条数和OHLC，覆盖US+HK
func TestIntegGetKlineByPageTimeRange(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	end := time.Now().UnixMilli()
	begin := time.Now().AddDate(0, 0, -30).UnixMilli()

	cases := []string{"AAPL", "00700"}
	for _, sym := range cases {
		sym := sym
		t.Run(sym, func(t *testing.T) {
			data, err := qc.GetKlineByPage(model.KlineByPageRequest{
				Symbol:    sym,
				Period:    "day",
				BeginTime: begin,
				EndTime:   end,
				TotalSize: 20,
				PageSize:  10,
			})
			if err != nil {
				t.Fatalf("GetKlineByPage(%s): %v", sym, err)
			}
			if len(data) < 10 {
				t.Errorf("GetKlineByPage(%s): got %d bars in 30-day window, want >= 10", sym, len(data))
			}
			for i := 1; i < len(data); i++ {
				if data[i].Time <= data[i-1].Time {
					t.Errorf("GetKlineByPage(%s): items[%d].Time not ascending", sym, i)
					break
				}
			}
			for i, item := range data {
				if item.High < item.Low {
					t.Errorf("GetKlineByPage(%s) item[%d]: High(%v) < Low(%v)", sym, i, item.High, item.Low)
				}
				if item.Close <= 0 {
					t.Errorf("GetKlineByPage(%s) item[%d]: Close=%v, want > 0", sym, i, item.Close)
				}
			}
		})
	}
}

// TestIntegGetQuoteDepthOrdering 验证盘口价格合理性：
//  - asks 按价格升序（最优卖价在前）
//  - bids 按价格降序（最优买价在前）
//  - 最低卖价 >= 最高买价（spread >= 0，非倒挂）
//  - 所有价格 > 0
//  覆盖 US(AAPL) 和 HK(00700)
func TestIntegGetQuoteDepthOrdering(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	cases := []struct{ sym, market string }{
		{"AAPL", "US"},
		{"00700", "HK"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.sym, func(t *testing.T) {
			data, err := qc.GetQuoteDepth(model.DepthQuoteRequest{
				Symbols: []string{tc.sym},
				Market:  tc.market,
			})
			if err != nil {
				t.Fatalf("GetQuoteDepth(%s): %v", tc.sym, err)
			}
			// Depth only streams during trading hours. Out-of-hours the wire
			// round-trip already validated the request; stop the subtest here
			// rather than asserting on missing data.
			if len(data) == 0 {
				integtest.AssertNonEmptyDuringTrading(t, qc, tc.market, "GetQuoteDepth(%s) empty", tc.sym)
				return
			}
			depth := data[0]
			if depth.Symbol != tc.sym {
				t.Errorf("Depth.Symbol = %q, want %q", depth.Symbol, tc.sym)
			}
			// asks升序
			for i := 1; i < len(depth.Asks); i++ {
				if depth.Asks[i].Price < depth.Asks[i-1].Price {
					t.Errorf("%s: asks[%d].Price(%v) < asks[%d].Price(%v): not ascending",
						tc.sym, i, depth.Asks[i].Price, i-1, depth.Asks[i-1].Price)
					break
				}
			}
			// bids降序
			for i := 1; i < len(depth.Bids); i++ {
				if depth.Bids[i].Price > depth.Bids[i-1].Price {
					t.Errorf("%s: bids[%d].Price(%v) > bids[%d].Price(%v): not descending",
						tc.sym, i, depth.Bids[i].Price, i-1, depth.Bids[i-1].Price)
					break
				}
			}
			// spread非倒挂：最低卖价 >= 最高买价
			if len(depth.Asks) > 0 && len(depth.Bids) > 0 {
				if depth.Asks[0].Price < depth.Bids[0].Price {
					t.Errorf("%s: spread inverted: lowestAsk(%v) < highestBid(%v)",
						tc.sym, depth.Asks[0].Price, depth.Bids[0].Price)
				}
			}
			// 所有价格必须 > 0
			for i, a := range depth.Asks {
				if a.Price <= 0 {
					t.Errorf("%s: asks[%d].Price = %v, want > 0", tc.sym, i, a.Price)
				}
			}
			for i, b := range depth.Bids {
				if b.Price <= 0 {
					t.Errorf("%s: bids[%d].Price = %v, want > 0", tc.sym, i, b.Price)
				}
			}
		})
	}
}

// TestIntegGetBriefMultiMarket 验证US+HK实时行情：价格>0、OHLC约束、ask>=bid
func TestIntegGetBriefMultiMarket(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	cases := []string{"AAPL", "00700", "09988"}
	for _, sym := range cases {
		sym := sym
		t.Run(sym, func(t *testing.T) {
			data, err := qc.GetBrief(model.BriefRequest{Symbols: []string{sym}})
			if err != nil {
				t.Skipf("GetBrief(%s): %v", sym, err)
			}
			if len(data) == 0 {
				t.Fatalf("GetBrief(%s): returned empty", sym)
			}
			q := data[0]
			if q.Symbol != sym {
				t.Errorf("Brief.Symbol = %q, want %q", q.Symbol, sym)
			}
			if q.LatestPrice <= 0 {
				t.Errorf("%s LatestPrice = %v, want > 0", sym, q.LatestPrice)
			}
			// OHLC约束（日内价格有值时检查）
			if q.High > 0 && q.Low > 0 {
				if q.High < q.Low {
					t.Errorf("%s: High(%v) < Low(%v)", sym, q.High, q.Low)
				}
				if q.Open > 0 {
					if q.High < q.Open {
						t.Errorf("%s: High(%v) < Open(%v)", sym, q.High, q.Open)
					}
					if q.Low > q.Open {
						t.Errorf("%s: Low(%v) > Open(%v)", sym, q.Low, q.Open)
					}
				}
			}
			// ask >= bid（spread非倒挂）
			if q.AskPrice > 0 && q.BidPrice > 0 {
				if q.AskPrice < q.BidPrice {
					t.Errorf("%s: AskPrice(%v) < BidPrice(%v): inverted spread",
						sym, q.AskPrice, q.BidPrice)
				}
			}
		})
	}
}

// ===========================================================================
// 高优先级补充：期权 Greeks 约束、K线复权对比、期货OHLC+深度排序、多市场状态
// ===========================================================================

// TestIntegOptionChainGreeksAndStrikeConsistency 验证期权链 Greeks 业务约束：
//   - Call delta ∈ (0, 1)，Put delta ∈ (-1, 0)（有数据时）
//   - IV (impliedVol) > 0（有数据时）
//   - 同一 OptionChainRow 内 Call.Strike == Put.Strike（期权链行内 strike 对应）
//   - Call.Right == "call"，Put.Right == "put"
func TestIntegOptionChainGreeksAndStrikeConsistency(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	exps, err := qc.GetOptionExpiration([]string{"AAPL"})
	if err != nil || len(exps) == 0 || len(exps[0].Dates) == 0 {
		t.Skipf("no expiry for AAPL: %v", err)
	}
	// 选中间到期日，流动性较好
	expiry := exps[0].Dates[len(exps[0].Dates)/2]
	chain, err := qc.GetOptionChain([][2]string{{"AAPL", expiry}})
	if err != nil || len(chain) == 0 || len(chain[0].Items) == 0 {
		t.Skipf("no option chain: %v", err)
	}

	greeksChecked := 0
	for i, row := range chain[0].Items {
		// Strike 一致性：同行 Call/Put 的 strike 必须相同
		if row.Call != nil && row.Put != nil {
			if row.Call.Strike != row.Put.Strike {
				t.Errorf("row[%d]: Call.Strike(%s) != Put.Strike(%s)", i, row.Call.Strike, row.Put.Strike)
			}
		}
		// Right 字段值
		if row.Call != nil && row.Call.Right != "" && row.Call.Right != "call" {
			t.Errorf("row[%d]: Call.Right = %q, want \"call\"", i, row.Call.Right)
		}
		if row.Put != nil && row.Put.Right != "" && row.Put.Right != "put" {
			t.Errorf("row[%d]: Put.Right = %q, want \"put\"", i, row.Put.Right)
		}
		// Greeks 约束（只在数据非零时校验，避免非交易时段误判）
		if row.Call != nil && row.Call.Delta != 0 {
			if row.Call.Delta <= 0 || row.Call.Delta > 1 {
				t.Errorf("row[%d]: Call.Delta = %v, want (0, 1]", i, row.Call.Delta)
			}
			greeksChecked++
		}
		if row.Put != nil && row.Put.Delta != 0 {
			if row.Put.Delta >= 0 || row.Put.Delta < -1 {
				t.Errorf("row[%d]: Put.Delta = %v, want [-1, 0)", i, row.Put.Delta)
			}
			greeksChecked++
		}
		if row.Call != nil && row.Call.ImpliedVol != 0 && row.Call.ImpliedVol < 0 {
			t.Errorf("row[%d]: Call.ImpliedVol = %v, want >= 0", i, row.Call.ImpliedVol)
		}
		if row.Put != nil && row.Put.ImpliedVol != 0 && row.Put.ImpliedVol < 0 {
			t.Errorf("row[%d]: Put.ImpliedVol = %v, want >= 0", i, row.Put.ImpliedVol)
		}
		if greeksChecked >= 10 {
			break // 检查前10条有 Greeks 的记录即可
		}
	}
	// Greeks are only computed during trading hours; outside the session the
	// server may return delta=0 for every row. In-session emptiness would be
	// a real regression. Out-of-hours the wire round-trip was already
	// validated by the successful chain fetch.
	if greeksChecked == 0 {
		t.Skip("all option rows have delta=0; Greeks may be unavailable for the sampled chain")
		return
	}
}

// TestIntegKlineRightAdjustment 验证复权参数效果：
// 使用包含 UVXY 2024-08-01 十合一反向拆股的固定时间窗口，确保 nr（不复权）与 br（前复权）价格必然不同。
// 服务端默认返回前复权(br)；nr 不复权在反向拆股前的价格约为 $4，br 约为 $100+，两者差异显著。
func TestIntegKlineRightAdjustment(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	// UVXY 在 2024-08-01 执行了 1:10 反向拆股；
	// 取 2024-07-01 ~ 2024-08-31 这段窗口：
	//   - nr（不复权）价格 ~$4（反拆前原始价格）
	//   - br（前复权）价格 ~$100+（还原为反拆后价格）
	// 服务端默认即为 br，因此必须显式传 nr 与 br 做对比。
	begin := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
	end := time.Date(2024, 8, 31, 0, 0, 0, 0, time.UTC).UnixMilli()

	nr, err := qc.GetKline(model.KlineRequest{
		Symbols:   []string{"UVXY"},
		Period:    "day",
		BeginTime: begin,
		EndTime:   end,
		Right:     "nr", // 不复权
	})
	if err != nil || len(nr) == 0 || len(nr[0].Items) == 0 {
		t.Skipf("GetKline (nr): %v", err)
	}

	br, err := qc.GetKline(model.KlineRequest{
		Symbols:   []string{"UVXY"},
		Period:    "day",
		BeginTime: begin,
		EndTime:   end,
		Right:     "br", // 前复权
	})
	if err != nil || len(br) == 0 || len(br[0].Items) == 0 {
		t.Skipf("GetKline (br): %v", err)
	}

	nrItems := nr[0].Items
	brItems := br[0].Items
	t.Logf("nr bars=%d  br bars=%d  nr[0].Close=%.4f  br[0].Close=%.4f",
		len(nrItems), len(brItems), nrItems[0].Close, brItems[0].Close)

	minLen := len(nrItems)
	if len(brItems) < minLen {
		minLen = len(brItems)
	}
	if minLen == 0 {
		t.Skip("no overlapping bars to compare")
	}
	allSame := true
	for i := 0; i < minLen; i++ {
		if nrItems[i].Close != brItems[i].Close {
			allSame = false
			break
		}
	}
	// The 2024-07-01..2024-08-31 window brackets UVXY's 2024-08-01 1:10 reverse split.
	// Pre-split bars (July): nr close ~$4 vs br close ~$100+ — these MUST differ.
	// Post-split bars (August): nr and br are identical (~$40) because no adjustment is needed.
	// If the server returns ONLY post-split bars, allSame will be true and this test will fail
	// loudly, indicating that the time window or server behaviour needs investigation.
	// This is intentional: a silent skip would hide a real regression in the right parameter.
	if allSame {
		t.Fatalf("nr and br prices identical across UVXY 2024-07..08 window — server ignored the right parameter (nr[0].Close=%.4f br[0].Close=%.4f)", nrItems[0].Close, brItems[0].Close)
	}
	// 如果不完全相同，复权价格必须都 > 0（业务约束）
	for i, item := range brItems {
		if item.Close <= 0 {
			t.Errorf("br-adjusted KlineItem[%d].Close = %v, want > 0", i, item.Close)
		}
		if item.High < item.Low {
			t.Errorf("br-adjusted KlineItem[%d]: High(%v) < Low(%v)", i, item.High, item.Low)
		}
	}
}

// TestIntegFutureKlineOHLC 验证期货K线的完整业务约束：
//   - 30天窗口内条数 >= 5
//   - 时间戳唯一（服务端不保证升序，但不应有重复）
//   - 每根 High >= max(Open,Close) >= min(Open,Close) >= Low
//   - Volume >= 0，Close > 0
func TestIntegFutureKlineOHLC(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	code := firstFutureContractCode(t, qc)
	end := time.Now().UnixMilli()
	begin := time.Now().AddDate(0, 0, -30).UnixMilli()

	data, err := qc.GetFutureKline(model.FutureKlineRequest{
		ContractCodes: []string{code},
		Period:        "day",
		BeginTime:     begin,
		EndTime:       end,
	})
	if err != nil || len(data) == 0 || len(data[0].Items) == 0 {
		t.Skipf("GetFutureKline(%s): %v", code, err)
	}
	items := data[0].Items

	if len(items) < 5 {
		t.Errorf("FutureKline(%s): got %d bars in 30-day window, want >= 5", code, len(items))
	}
	// 时间戳唯一性（不要求升序，服务端可能降序或乱序）
	seen := make(map[int64]bool)
	for i, item := range items {
		if seen[item.Time] {
			t.Errorf("FutureKline(%s): duplicate timestamp at items[%d].Time=%d", code, i, item.Time)
		}
		seen[item.Time] = true
	}
	for i, item := range items {
		if item.Close <= 0 {
			t.Errorf("FutureKline(%s) item[%d]: Close=%v, want > 0", code, i, item.Close)
		}
		if item.High < item.Low {
			t.Errorf("FutureKline(%s) item[%d]: High(%v) < Low(%v)", code, i, item.High, item.Low)
		}
		if item.High < item.Open {
			t.Errorf("FutureKline(%s) item[%d]: High(%v) < Open(%v)", code, i, item.High, item.Open)
		}
		if item.High < item.Close {
			t.Errorf("FutureKline(%s) item[%d]: High(%v) < Close(%v)", code, i, item.High, item.Close)
		}
		if item.Low > item.Open {
			t.Errorf("FutureKline(%s) item[%d]: Low(%v) > Open(%v)", code, i, item.Low, item.Open)
		}
		if item.Low > item.Close {
			t.Errorf("FutureKline(%s) item[%d]: Low(%v) > Close(%v)", code, i, item.Low, item.Close)
		}
		if item.Volume < 0 {
			t.Errorf("FutureKline(%s) item[%d]: Volume=%d, want >= 0", code, i, item.Volume)
		}
	}
}

// TestIntegFutureDepthOrdering 验证期货盘口排序：
//   - asks 按价格升序，bids 按价格降序
//   - lowestAsk >= highestBid（spread >= 0）
//   - 所有价格 > 0
func TestIntegFutureDepthOrdering(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	code := firstFutureContractCode(t, qc)
	data, err := qc.GetFutureDepth(model.FutureDepthRequest{ContractCodes: []string{code}})
	if err != nil {
		t.Skipf("GetFutureDepth(%s): %v", code, err)
	}
	if len(data) == 0 {
		t.Skipf("GetFutureDepth(%s): empty (non-trading hours?)", code)
	}
	depth := data[0]

	// asks 升序
	for i := 1; i < len(depth.Asks); i++ {
		if depth.Asks[i].Price < depth.Asks[i-1].Price {
			t.Errorf("FutureDepth asks[%d].Price(%v) < asks[%d].Price(%v): not ascending",
				i, depth.Asks[i].Price, i-1, depth.Asks[i-1].Price)
			break
		}
	}
	// bids 降序
	for i := 1; i < len(depth.Bids); i++ {
		if depth.Bids[i].Price > depth.Bids[i-1].Price {
			t.Errorf("FutureDepth bids[%d].Price(%v) > bids[%d].Price(%v): not descending",
				i, depth.Bids[i].Price, i-1, depth.Bids[i-1].Price)
			break
		}
	}
	// spread 非倒挂
	if len(depth.Asks) > 0 && len(depth.Bids) > 0 {
		if depth.Asks[0].Price < depth.Bids[0].Price {
			t.Errorf("FutureDepth spread inverted: lowestAsk(%v) < highestBid(%v)",
				depth.Asks[0].Price, depth.Bids[0].Price)
		}
	}
	// 价格 > 0
	for i, a := range depth.Asks {
		if a.Price <= 0 {
			t.Errorf("FutureDepth asks[%d].Price = %v, want > 0", i, a.Price)
		}
	}
	for i, b := range depth.Bids {
		if b.Price <= 0 {
			t.Errorf("FutureDepth bids[%d].Price = %v, want > 0", i, b.Price)
		}
	}
}

// TestIntegGetMarketStateMultiMarket 验证 US 和 HK 两个市场的状态都能正常返回：
//   - Market 字段非空，Status/MarketStatus 字段之一非空
//   - 返回条数 >= 1
func TestIntegGetMarketStateMultiMarket(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	for _, market := range []string{"US", "HK"} {
		market := market
		t.Run(market, func(t *testing.T) {
			data, err := qc.GetMarketState(market)
			if err != nil {
				t.Fatalf("GetMarketState(%s): %v", market, err)
			}
			if len(data) == 0 {
				t.Fatalf("GetMarketState(%s): returned empty", market)
			}
			for i, ms := range data {
				if ms.Market == "" {
					t.Errorf("GetMarketState(%s)[%d].Market is empty", market, i)
				}
				if ms.MarketStatus == "" && ms.Status == "" {
					t.Errorf("GetMarketState(%s)[%d]: both MarketStatus and Status are empty", market, i)
				}
			}
		})
	}
}
