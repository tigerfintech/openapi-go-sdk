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
}

func TestIntegGetTimeline(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	data, err := qc.GetTimeline([]string{"AAPL"})
	if err != nil {
		t.Fatalf("GetTimeline: %v", err)
	}
	// Timeline may legitimately be empty outside trading hours; only assert no error.
	_ = data
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
	if len(data[0].Dates) == 0 {
		t.Fatal("no expiry dates for AAPL")
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
	if data == nil {
		t.Fatal("GetCorporateAction returned nil")
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
}

func TestIntegGrabQuotePermission(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	qc := integtest.NewQuoteClient(t)

	if _, err := qc.GrabQuotePermission(); err != nil {
		t.Fatalf("GrabQuotePermission: %v", err)
	}
}
