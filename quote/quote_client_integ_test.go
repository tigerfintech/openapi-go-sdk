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
	"testing"
	"time"

	"github.com/tigerfintech/openapi-go-sdk/integtest"
	"github.com/tigerfintech/openapi-go-sdk/model"
	"github.com/tigerfintech/openapi-go-sdk/quote"
)

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
	// Timeline may return no data outside trading hours.
	if len(data) == 0 {
		t.Skip("GetTimeline returned empty for AAPL (likely outside trading hours)")
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
		BeginDate:  "2024-01-01",
		EndDate:    "2025-12-31",
	})
	if err != nil {
		t.Fatalf("GetCorporateAction: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("GetCorporateAction returned empty for AAPL DIVIDEND over 2-year range")
	}
	if data[0].Symbol != "AAPL" {
		t.Errorf("CorporateAction.Symbol = %q, want AAPL", data[0].Symbol)
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
	// Capital flow items may be empty outside trading hours.
	if len(data.Items) == 0 {
		t.Skip("GetCapitalFlow returned empty items (likely non-trading hours)")
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

func firstOptionIdentifier(t *testing.T, qc *quote.QuoteClient) string {
	t.Helper()
	exps, err := qc.GetOptionExpiration([]string{"AAPL"})
	if err != nil || len(exps) == 0 || len(exps[0].OptionSymbols) == 0 {
		t.Skipf("no option symbols for AAPL: err=%v", err)
	}
	return exps[0].OptionSymbols[0]
}

func firstOptionChainLeg(t *testing.T, qc *quote.QuoteClient) (model.OptionLeg, int64) {
	t.Helper()
	exps, err := qc.GetOptionExpiration([]string{"AAPL"})
	if err != nil || len(exps) == 0 || len(exps[0].Dates) == 0 {
		t.Skipf("no option expiration for AAPL: err=%v", err)
	}
	idx := len(exps[0].Dates) / 2
	expiry := exps[0].Dates[idx]
	var expiryTs int64
	if len(exps[0].Timestamps) > idx {
		expiryTs = exps[0].Timestamps[idx]
	}
	chain, err := qc.GetOptionChain([][2]string{{"AAPL", expiry}})
	if err != nil || len(chain) == 0 || len(chain[0].Items) == 0 {
		t.Skipf("no option chain for AAPL %s: err=%v", expiry, err)
	}
	row := chain[0].Items[0]
	if row.Call != nil {
		return *row.Call, expiryTs
	}
	if row.Put != nil {
		return *row.Put, expiryTs
	}
	t.Skip("no option legs in chain row")
	return model.OptionLeg{}, 0
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
}

func TestIntegGetTradeTick(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetTradeTick(model.TradeTickRequest{Symbols: []string{"AAPL"}, Limit: 5})
	if err != nil {
		t.Fatalf("GetTradeTick: %v", err)
	}
	// Trade ticks may be empty outside trading hours.
	if len(data) == 0 || len(data[0].Items) == 0 {
		t.Skip("GetTradeTick returned empty (likely outside trading hours)")
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
	// Depth may be empty outside trading hours.
	if len(data) == 0 {
		t.Skip("GetQuoteDepth returned empty (likely outside trading hours)")
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
		t.Skip("GetQuoteOvernight returned empty")
	}
	if data[0].Symbol != "AAPL" {
		t.Errorf("QuoteOvernight.Symbol = %q, want AAPL", data[0].Symbol)
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
	if data[0].Currency == "" {
		t.Errorf("StockDetail.Currency is empty, want non-empty")
	}
}

func TestIntegGetStockBroker(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	// stock_broker only supports HK market
	data, err := qc.GetStockBroker(model.StockBrokerRequest{Symbol: "0700.HK", Limit: 5})
	if err != nil {
		t.Fatalf("GetStockBroker: %v", err)
	}
	if data == nil {
		t.Fatal("GetStockBroker returned nil")
	}
	if data.Symbol != "0700.HK" {
		t.Errorf("StockBroker.Symbol = %q, want 0700.HK", data.Symbol)
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
	if data[0].Symbol != "AAPL" {
		t.Errorf("StockIndustry.Symbol = %q, want AAPL", data[0].Symbol)
	}
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
	// Rank may be empty outside trading hours.
	if len(data) == 0 {
		t.Skip("GetTradeRank returned empty (likely outside trading hours)")
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
		t.Fatalf("GetShortInterest: %v", err)
	}
	// Short interest may not be available for all symbols.
	if len(data) == 0 {
		t.Skip("GetShortInterest returned empty for AAPL")
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

	data, err := qc.GetFinancialDaily(model.FinancialDailyRequest{
		Symbols:   []string{"AAPL"},
		Market:    "US",
		Fields:    []string{"total_revenue"},
		BeginDate: "2024-01-01",
		EndDate:   "2024-06-30",
	})
	if err != nil {
		t.Fatalf("GetFinancialDaily: %v", err)
	}
	if len(data) == 0 {
		t.Skip("GetFinancialDaily returned empty for AAPL")
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

	data, err := qc.GetFinancialReport(model.FinancialReportRequest{
		Symbols:    []string{"AAPL"},
		Market:     "US",
		Fields:     []string{"total_revenue"},
		PeriodType: "Annual",
	})
	if err != nil {
		t.Fatalf("GetFinancialReport: %v", err)
	}
	if len(data) == 0 {
		t.Skip("GetFinancialReport returned empty for AAPL")
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

	data, err := qc.GetMarketScannerTags(model.MarketScannerTagsRequest{Market: "US"})
	if err != nil {
		t.Fatalf("GetMarketScannerTags: %v", err)
	}
	// Tags may be empty for some markets.
	if len(data) == 0 {
		t.Skip("GetMarketScannerTags returned empty for US")
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
	if data[0].Market == "" {
		t.Errorf("TradingCalendarItem.Market is empty, want non-empty")
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
	data, err := qc.GetOptionKline([]string{identifier}, "day", -1, -1)
	if err != nil {
		t.Fatalf("GetOptionKline: %v", err)
	}
	// Option kline may return empty if the option has no historical data.
	if len(data) == 0 || len(data[0].Items) == 0 {
		t.Skip("GetOptionKline returned empty")
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
	// Trade ticks may be empty outside trading hours.
	if len(data) == 0 || len(data[0].Items) == 0 {
		t.Skip("GetOptionTradeTicks returned empty (likely non-trading hours)")
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
	// Timeline may be empty outside trading hours.
	if len(data) == 0 {
		t.Skip("GetOptionTimeline returned empty (likely non-trading hours)")
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
	// Depth may be empty outside trading hours.
	if len(data) == 0 {
		t.Skip("GetOptionDepth returned empty (likely non-trading hours)")
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
		t.Errorf("FutureContractInfo.ContractCode is empty, want non-empty")
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
		t.Fatalf("GetAllFutureContracts: %v", err)
	}
	if len(data) == 0 {
		t.Skip("GetAllFutureContracts returned empty")
	}
	if data[0].ContractCode == "" {
		t.Errorf("FutureContractInfo.ContractCode is empty, want non-empty")
	}
}

func TestIntegGetCurrentFutureContract(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	code := firstFutureContractCode(t, qc)
	data, err := qc.GetCurrentFutureContract(model.FutureContractSingleRequest{ContractCode: code})
	if err != nil {
		t.Fatalf("GetCurrentFutureContract: %v", err)
	}
	if data == nil {
		t.Skipf("GetCurrentFutureContract returned nil for %s", code)
	}
	if data.ContractCode == "" {
		t.Errorf("FutureContractInfo.ContractCode is empty, want non-empty")
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
		t.Skip("GetFutureContinuousContracts returned empty")
	}
	if data[0].ContractCode == "" {
		t.Errorf("FutureContractInfo.ContractCode is empty, want non-empty")
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
		t.Fatalf("GetFutureHistoryMainContract: %v", err)
	}
	if len(data) == 0 {
		t.Skipf("GetFutureHistoryMainContract returned empty for %s", code)
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

	data, err := qc.GetWarrantFilter(model.WarrantFilterRequest{Symbol: "0700.HK", PageSize: 5})
	if err != nil {
		t.Fatalf("GetWarrantFilter: %v", err)
	}
	if data == nil {
		t.Fatal("GetWarrantFilter returned nil")
	}
	// Warrants may not be available depending on entitlements.
	if data.Total == 0 {
		t.Skip("GetWarrantFilter returned 0 warrants for 0700.HK")
	}
	if len(data.Items) == 0 {
		t.Fatal("GetWarrantFilter has Total > 0 but no items")
	}
	if data.Items[0].Symbol == "" {
		t.Errorf("WarrantBrief.Symbol is empty, want non-empty")
	}
}

func TestIntegGetWarrantQuote(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	// First get a warrant via filter for a HK underlying.
	filter, err := qc.GetWarrantFilter(model.WarrantFilterRequest{Symbol: "0700.HK", PageSize: 5})
	if err != nil || filter == nil || len(filter.Items) == 0 {
		t.Skipf("no warrants for 0700.HK: err=%v", err)
	}
	warrantSymbol := filter.Items[0].Symbol
	data, err := qc.GetWarrantQuote(model.WarrantBriefsRequest{Symbols: []string{warrantSymbol}})
	if err != nil {
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
		t.Errorf("IndustryItem.Name is empty, want non-empty")
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
		BeginDate: "2020-01-01",
		EndDate:   "2025-12-31",
	})
	if err != nil {
		t.Fatalf("GetCorporateSplit: %v", err)
	}
	// Splits may not exist for AAPL in this range.
	if len(data) == 0 {
		t.Skip("GetCorporateSplit returned empty for AAPL 2020-2025")
	}
	if data[0].Symbol != "AAPL" {
		t.Errorf("CorporateAction.Symbol = %q, want AAPL", data[0].Symbol)
	}
}

func TestIntegGetCorporateDividend(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetCorporateDividend(model.CorporateActionRequest{
		Symbols:   []string{"AAPL"},
		Market:    "US",
		BeginDate: "2024-01-01",
		EndDate:   "2025-12-31",
	})
	if err != nil {
		t.Fatalf("GetCorporateDividend: %v", err)
	}
	if len(data) == 0 {
		t.Skip("GetCorporateDividend returned empty for AAPL 2024-2025")
	}
	if data[0].Symbol != "AAPL" {
		t.Errorf("CorporateAction.Symbol = %q, want AAPL", data[0].Symbol)
	}
}

func TestIntegGetCorporateEarningsCalendar(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetCorporateEarningsCalendar(model.CorporateActionRequest{
		Symbols:   []string{"AAPL"},
		Market:    "US",
		BeginDate: "2024-01-01",
		EndDate:   "2025-12-31",
	})
	if err != nil {
		t.Fatalf("GetCorporateEarningsCalendar: %v", err)
	}
	if len(data) == 0 {
		t.Skip("GetCorporateEarningsCalendar returned empty for AAPL 2024-2025")
	}
	if data[0].Symbol != "AAPL" {
		t.Errorf("CorporateAction.Symbol = %q, want AAPL", data[0].Symbol)
	}
}

func TestIntegGetCorporateSymbolChange(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetCorporateSymbolChange(model.CorporateActionRequest{
		Symbols:   []string{"AAPL"},
		Market:    "US",
		BeginDate: "2020-01-01",
		EndDate:   "2025-12-31",
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
		Market:    "US",
		BeginDate: "2020-01-01",
		EndDate:   "2025-12-31",
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

	data, err := qc.GetCorporateIPO(model.CorporateActionRequest{
		Market:    "US",
		BeginDate: "2024-01-01",
		EndDate:   "2025-12-31",
	})
	if err != nil {
		t.Fatalf("GetCorporateIPO: %v", err)
	}
	if len(data) == 0 {
		t.Skip("GetCorporateIPO returned empty for US 2024-2025")
	}
	if data[0].Symbol == "" {
		t.Errorf("CorporateIPO.Symbol is empty, want non-empty")
	}
}
