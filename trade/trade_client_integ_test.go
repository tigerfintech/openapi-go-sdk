//go:build integration

// Integration tests against the real OpenAPI gateway. Run with:
//
//	make integ
//	# or: TIGER_RUN_INTEG=true go test -tags=integration ./trade/...
//
// These tests only exercise read-only trade endpoints (contract/asset/position/
// order queries) — no order placement, modification, or cancellation.
// Credentials come from TIGEROPEN_TIGER_ID / PRIVATE_KEY / ACCOUNT env vars
// (or TIGEROPEN_PROPS_PATH). See integtest.NewConfig.

package trade_test

import (
	"testing"
	"time"

	"github.com/tigerfintech/openapi-go-sdk/integtest"
	"github.com/tigerfintech/openapi-go-sdk/model"
)

func TestIntegGetContract(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)

	data, err := tc.Contract("AAPL", "STK")
	if err != nil {
		t.Fatalf("Contract: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("Contract returned empty for AAPL")
	}
	if data[0].Symbol != "AAPL" {
		t.Errorf("symbol = %q, want AAPL", data[0].Symbol)
	}
}

func TestIntegGetAssets(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)

	data, err := tc.Assets(model.AssetsRequest{})
	if err != nil {
		t.Fatalf("Assets: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("Assets returned empty")
	}
}

func TestIntegGetPositions(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)

	// Positions may legitimately be empty for a fresh paper account; only
	// assert the call succeeds.
	_, err := tc.Positions(model.PositionsRequest{})
	if err != nil {
		t.Fatalf("Positions: %v", err)
	}
}

func TestIntegGetOrders(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)

	now := time.Now()
	_, err := tc.Orders(model.OrdersRequest{
		StartDate: now.AddDate(0, 0, -30).UnixMilli(),
		EndDate:   now.UnixMilli(),
		Limit:     10,
	})
	if err != nil {
		t.Fatalf("Orders: %v", err)
	}
}

func TestIntegGetActiveOrders(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)

	_, err := tc.ActiveOrders(model.OrdersRequest{})
	if err != nil {
		t.Fatalf("ActiveOrders: %v", err)
	}
}
