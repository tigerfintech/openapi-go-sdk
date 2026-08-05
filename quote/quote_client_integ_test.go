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

	"github.com/tigerfintech/openapi-go-sdk/integtest"
	"github.com/tigerfintech/openapi-go-sdk/model"
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
	if len(data) == 0 {
		t.Fatal("GetTimeline returned empty for AAPL")
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
	if len(data.Items) == 0 {
		t.Errorf("CapitalFlow.Items is empty for AAPL daily")
	} else {
		item := data.Items[0]
		if item.Timestamp == 0 {
			t.Errorf("CapitalFlowItem.Timestamp = 0, want non-zero")
		}
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
