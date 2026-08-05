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
	if data[0].SecType != "STK" {
		t.Errorf("SecType = %q, want STK", data[0].SecType)
	}
	if data[0].Currency == "" {
		t.Errorf("Currency is empty, want non-empty (e.g. USD)")
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
	if data[0].Account == "" {
		t.Errorf("Asset.Account is empty, want non-empty")
	}
	if data[0].Currency == "" {
		t.Errorf("Asset.Currency is empty, want non-empty")
	}
	if data[0].NetLiquidation < 0 {
		t.Errorf("Asset.NetLiquidation = %v, want >= 0", data[0].NetLiquidation)
	}
}

func TestIntegGetPositions(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)

	// Positions may legitimately be empty for a fresh paper account; only
	// assert the call succeeds. When positions exist, validate key fields.
	data, err := tc.Positions(model.PositionsRequest{})
	if err != nil {
		t.Fatalf("Positions: %v", err)
	}
	for i, p := range data {
		if p.Symbol == "" {
			t.Errorf("Position[%d].Symbol is empty, want non-empty", i)
		}
		if p.Account == "" {
			t.Errorf("Position[%d].Account is empty, want non-empty", i)
		}
		if p.SecType == "" {
			t.Errorf("Position[%d].SecType is empty, want non-empty", i)
		}
	}
}

func TestIntegGetOrders(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)

	now := time.Now()
	data, err := tc.Orders(model.OrdersRequest{
		StartDate: now.AddDate(0, 0, -30).UnixMilli(),
		EndDate:   now.UnixMilli(),
		Limit:     10,
	})
	if err != nil {
		t.Fatalf("Orders: %v", err)
	}
	// Orders may be empty if no trades in the last 30 days. When orders
	// exist, validate identifier and numeric fields.
	for i, o := range data {
		if o.ID == 0 {
			t.Errorf("Order[%d].ID = 0, want non-zero", i)
		}
		if o.Symbol == "" {
			t.Errorf("Order[%d].Symbol is empty, want non-empty", i)
		}
		if o.Action == "" {
			t.Errorf("Order[%d].Action is empty, want non-empty", i)
		}
		if o.Status == "" {
			t.Errorf("Order[%d].Status is empty, want non-empty", i)
		}
	}
}

func TestIntegGetActiveOrders(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)

	data, err := tc.ActiveOrders(model.OrdersRequest{})
	if err != nil {
		t.Fatalf("ActiveOrders: %v", err)
	}
	// Active orders may be empty if no open orders exist. When orders
	// exist, validate identifier fields.
	for i, o := range data {
		if o.ID == 0 {
			t.Errorf("ActiveOrder[%d].ID = 0, want non-zero", i)
		}
		if o.Symbol == "" {
			t.Errorf("ActiveOrder[%d].Symbol is empty, want non-empty", i)
		}
		if o.Status == "" {
			t.Errorf("ActiveOrder[%d].Status is empty, want non-empty", i)
		}
	}
}
