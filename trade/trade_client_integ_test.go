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
	"strings"
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

// ============================================================================
// Additional trade integration tests — covering all remaining public methods.
// Write operations (PlaceOrder/ModifyOrder/CancelOrder/TransferSegmentFund/
// CancelSegmentFund/PlaceForexOrder/OptionExerciseSubmit/OptionExerciseCancel/
// TransferPosition) are skipped; only PreviewOrder and read-only methods are tested.
// ============================================================================

func TestIntegContract3(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)

	data, err := tc.Contract3("AAPL", "STK")
	if err != nil {
		t.Fatalf("Contract3: %v", err)
	}
	if data == nil {
		t.Fatal("Contract3 returned nil")
	}
	if data.Symbol != "AAPL" {
		t.Errorf("Contract.Symbol = %q, want AAPL", data.Symbol)
	}
	if data.Currency == "" {
		t.Errorf("Contract.Currency is empty, want non-empty")
	}
}

func TestIntegContracts(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)

	data, err := tc.Contracts([]string{"AAPL"}, "STK")
	if err != nil {
		t.Fatalf("Contracts: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("Contracts returned empty for AAPL")
	}
	if data[0].Symbol != "AAPL" {
		t.Errorf("Contract.Symbol = %q, want AAPL", data[0].Symbol)
	}
}

func TestIntegQuoteContract(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)

	// QuoteContract requires derivative sec types (OPT/WAR/IOPT) and an expiry.
	// Fetch a valid option expiry from the quote client. Use a mid-range date
	// to avoid the nearest expiry becoming stale on expiry day.
	qc := integtest.NewQuoteClient(t)
	exps, err := qc.GetOptionExpiration([]string{"AAPL"})
	if err != nil || len(exps) == 0 || len(exps[0].Dates) == 0 {
		t.Skipf("no option expiration for AAPL: err=%v", err)
	}
	// Convert "2024-01-19" to "20240119" for the trade API.
	idx := len(exps[0].Dates) / 2
	expiry := strings.ReplaceAll(exps[0].Dates[idx], "-", "")

	data, err := tc.QuoteContract("AAPL", "OPT", expiry)
	if err != nil {
		t.Fatalf("QuoteContract: %v", err)
	}
	// May return empty if no matching contracts.
	_ = data
}

func TestIntegPreviewOrder(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)

	// PreviewOrder sends a preview request (no actual order placed).
	result, err := tc.PreviewOrder(model.OrderRequest{
		Symbol:        "AAPL",
		SecType:       "STK",
		Action:        "BUY",
		OrderType:     "MKT",
		TotalQuantity: 1,
		Currency:      "USD",
		Market:        "US",
	})
	if err != nil {
		t.Fatalf("PreviewOrder: %v", err)
	}
	if result == nil {
		t.Fatal("PreviewOrder returned nil")
	}
	// isPass may be false for accounts without sufficient buying power,
	// but the call itself should succeed.
}

func TestIntegGetOrder(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)

	// Get an order ID from the orders list first.
	now := time.Now()
	orders, err := tc.Orders(model.OrdersRequest{
		StartDate: now.AddDate(0, 0, -30).UnixMilli(),
		EndDate:   now.UnixMilli(),
		Limit:     1,
	})
	if err != nil {
		t.Fatalf("Orders prerequisite: %v", err)
	}
	if len(orders) == 0 {
		t.Skip("no orders in last 30 days to test GetOrder")
	}

	order, err := tc.GetOrder(model.GetOrderRequest{Id: orders[0].ID})
	if err != nil {
		t.Fatalf("GetOrder: %v", err)
	}
	if order == nil {
		t.Fatal("GetOrder returned nil for existing order ID")
	}
	if order.ID != orders[0].ID {
		t.Errorf("Order.ID = %v, want %v", order.ID, orders[0].ID)
	}
}

func TestIntegOrderTransactions(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)

	now := time.Now()
	data, err := tc.OrderTransactions(model.OrderTransactionsRequest{
		Symbol:    "AAPL",
		StartDate: now.AddDate(0, 0, -30).UnixMilli(),
		EndDate:   now.UnixMilli(),
		Limit:     10,
	})
	if err != nil {
		t.Fatalf("OrderTransactions: %v", err)
	}
	// Transactions may be empty if no trades in last 30 days.
	for i, tr := range data {
		if tr.OrderID == 0 {
			t.Errorf("Transaction[%d].OrderID = 0, want non-zero", i)
		}
		if tr.Symbol == "" {
			t.Errorf("Transaction[%d].Symbol is empty, want non-empty", i)
		}
	}
}

func TestIntegPrimeAssets(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)

	data, err := tc.PrimeAssets(model.AssetsRequest{})
	if err != nil {
		t.Fatalf("PrimeAssets: %v", err)
	}
	if data == nil {
		t.Skip("PrimeAssets returned nil (may not be a prime account)")
	}
	if data.AccountID == "" {
		t.Errorf("PrimeAsset.AccountID is empty, want non-empty")
	}
}

func TestIntegInactiveOrders(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)

	data, err := tc.InactiveOrders(model.OrdersRequest{})
	if err != nil {
		t.Fatalf("InactiveOrders: %v", err)
	}
	for i, o := range data {
		if o.ID == 0 {
			t.Errorf("InactiveOrder[%d].ID = 0, want non-zero", i)
		}
		if o.Symbol == "" {
			t.Errorf("InactiveOrder[%d].Symbol is empty, want non-empty", i)
		}
	}
}

func TestIntegFilledOrders(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)

	now := time.Now()
	data, err := tc.FilledOrders(model.OrdersRequest{
		StartDate: now.AddDate(0, 0, -30).UnixMilli(),
		EndDate:   now.UnixMilli(),
		Limit:     10,
	})
	if err != nil {
		t.Fatalf("FilledOrders: %v", err)
	}
	for i, o := range data {
		if o.ID == 0 {
			t.Errorf("FilledOrder[%d].ID = 0, want non-zero", i)
		}
		if o.Status == "" {
			t.Errorf("FilledOrder[%d].Status is empty, want non-empty", i)
		}
	}
}

func TestIntegManagedAccounts(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)

	data, err := tc.ManagedAccounts(model.ManagedAccountsRequest{})
	if err != nil {
		t.Fatalf("ManagedAccounts: %v", err)
	}
	// May be empty for individual accounts.
	for i, a := range data {
		if a.Account == "" {
			t.Errorf("ManagedAccount[%d].Account is empty, want non-empty", i)
		}
	}
}

func TestIntegDerivativeContracts(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)

	// DerivativeContracts requires an expiry for option sec types.
	// Fetch a valid expiry from the quote client.
	qc := integtest.NewQuoteClient(t)
	exps, err := qc.GetOptionExpiration([]string{"AAPL"})
	if err != nil || len(exps) == 0 || len(exps[0].Dates) == 0 {
		t.Skipf("no option expiration for AAPL: err=%v", err)
	}
	// Convert "2024-01-19" to "20240119" for the trade API.
	expiry := strings.ReplaceAll(exps[0].Dates[0], "-", "")

	data, err := tc.DerivativeContracts(model.DerivativeContractsRequest{
		Symbols: []string{"AAPL"},
		SecType: "OPT",
		Expiry:  expiry,
	})
	if err != nil {
		t.Fatalf("DerivativeContracts: %v", err)
	}
	// May be empty if no option contracts available.
	_ = data
}

func TestIntegAnalyticsAsset(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)

	now := time.Now()
	data, err := tc.AnalyticsAsset(model.AnalyticsAssetRequest{
		StartDate: now.AddDate(0, 0, -7).Format("2006-01-02"),
		EndDate:   now.Format("2006-01-02"),
	})
	if err != nil {
		t.Fatalf("AnalyticsAsset: %v", err)
	}
	// May be empty for new accounts.
	for i, a := range data {
		if a.Date == "" {
			t.Errorf("AnalyticsAsset[%d].Date is empty, want non-empty", i)
		}
	}
}

func TestIntegAggregateAssets(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)

	data, err := tc.AggregateAssets(model.AggregateAssetsRequest{})
	if err != nil {
		if strings.Contains(err.Error(), "institution account") {
			t.Skipf("AggregateAssets requires institution account: %v", err)
		}
		t.Fatalf("AggregateAssets: %v", err)
	}
	if data == nil {
		t.Skip("AggregateAssets returned nil")
	}
	if data.AccountID == "" {
		t.Errorf("AggregateAssets.AccountID is empty, want non-empty")
	}
}

func TestIntegEstimateTradableQuantity(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)

	data, err := tc.EstimateTradableQuantity(model.EstimateTradableQuantityRequest{
		Symbol:    "AAPL",
		SecType:   "STK",
		Action:    "BUY",
		OrderType: "MKT",
		SegType:   "SEC",
	})
	if err != nil {
		t.Fatalf("EstimateTradableQuantity: %v", err)
	}
	if data == nil {
		t.Fatal("EstimateTradableQuantity returned nil")
	}
	if data.Currency == "" {
		t.Skip("EstimateTradableQuantity.Currency is empty (may not be returned for this account)")
	}
}

func TestIntegSegmentFundAvailable(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)

	data, err := tc.SegmentFundAvailable(model.SegmentFundRequest{FromSegment: "SEC"})
	if err != nil {
		t.Fatalf("SegmentFundAvailable: %v", err)
	}
	// May be empty for non-segment accounts.
	_ = data
}

func TestIntegSegmentFundHistory(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)

	data, err := tc.SegmentFundHistory(model.SegmentFundRequest{Limit: 5})
	if err != nil {
		t.Fatalf("SegmentFundHistory: %v", err)
	}
	// May be empty if no transfers.
	_ = data
}

func TestIntegFundDetails(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)

	now := time.Now()
	data, err := tc.FundDetails(model.FundDetailsRequest{
		SegTypes:  []string{"SEC"},
		StartDate: now.AddDate(0, 0, -30).Format("2006-01-02"),
		EndDate:   now.Format("2006-01-02"),
		Limit:     5,
	})
	if err != nil {
		t.Fatalf("FundDetails: %v", err)
	}
	// May be empty.
	for i, f := range data {
		if f.Account == "" {
			t.Skipf("FundDetails[%d].Account is empty (may not be returned for this account)", i)
		}
	}
}

func TestIntegFundingHistory(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)

	data, err := tc.FundingHistory(model.FundingHistoryRequest{})
	if err != nil {
		t.Fatalf("FundingHistory: %v", err)
	}
	// May be empty.
	_ = data
}

func TestIntegPositionTransferRecords(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)

	now := time.Now()
	data, err := tc.PositionTransferRecords(model.PositionTransferRecordsRequest{
		SinceDate: now.AddDate(0, 0, -7).Format("2006-01-02"),
		ToDate:    now.Format("2006-01-02"),
	})
	if err != nil {
		t.Fatalf("PositionTransferRecords: %v", err)
	}
	// May be empty if no transfers.
	_ = data
}

func TestIntegPositionTransferExternalRecords(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)

	now := time.Now()
	data, err := tc.PositionTransferExternalRecords(model.PositionTransferExternalRecordsRequest{
		SinceDate: now.AddDate(0, 0, -7).Format("2006-01-02"),
		ToDate:    now.Format("2006-01-02"),
	})
	if err != nil {
		t.Fatalf("PositionTransferExternalRecords: %v", err)
	}
	// May be empty.
	_ = data
}

func TestIntegPositionTransferDetail(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)

	// PositionTransferDetail requires a valid transfer record ID.
	// Fetch records first; skip if none exist.
	now := time.Now()
	records, err := tc.PositionTransferRecords(model.PositionTransferRecordsRequest{
		SinceDate: now.AddDate(0, 0, -7).Format("2006-01-02"),
		ToDate:    now.Format("2006-01-02"),
	})
	if err != nil {
		t.Fatalf("PositionTransferRecords prerequisite: %v", err)
	}
	if len(records) == 0 {
		t.Skip("no position transfer records to test PositionTransferDetail")
	}
	detail, err := tc.PositionTransferDetail(model.PositionTransferDetailRequest{
		ID:        records[0].ID,
		AccountID: records[0].FromAccount,
	})
	if err != nil {
		t.Fatalf("PositionTransferDetail: %v", err)
	}
	if detail == nil {
		t.Fatal("PositionTransferDetail returned nil for existing record ID")
	}
	if detail.ID == "" {
		t.Errorf("PositionTransferDetail.ID is empty, want non-empty")
	}
}

func TestIntegOptionExerciseCheck(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)

	// OptionExerciseCheck requires a valid option contract ID.
	// Fetch exercisable positions first; skip if none exist.
	positions, err := tc.OptionExercisePositions(model.OptionExercisePositionRequest{Type: "Exercise"})
	if err != nil || positions == nil || positions.ItemCount == 0 || len(positions.Items) == 0 {
		t.Skipf("no exercisable positions: err=%v", err)
	}
	contractId := positions.Items[0].ContractId
	result, err := tc.OptionExerciseCheck(model.OptionExerciseCheckRequest{
		ContractId: contractId,
		Type:       "Exercise",
		Quantity:   1,
	})
	if err != nil {
		t.Fatalf("OptionExerciseCheck: %v", err)
	}
	if result == nil {
		t.Fatal("OptionExerciseCheck returned nil")
	}
	if result.Symbol == "" {
		t.Errorf("OptionExerciseCheckResult.Symbol is empty, want non-empty")
	}
}

func TestIntegOptionExercisePositions(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)

	data, err := tc.OptionExercisePositions(model.OptionExercisePositionRequest{Type: "Exercise"})
	if err != nil {
		t.Fatalf("OptionExercisePositions: %v", err)
	}
	if data == nil {
		t.Fatal("OptionExercisePositions returned nil")
	}
	// May have no exercisable positions.
	if data.ItemCount > 0 && len(data.Items) > 0 {
		if data.Items[0].Symbol == "" {
			t.Errorf("OptionExercisePosition.Symbol is empty, want non-empty")
		}
	}
}

func TestIntegOptionExerciseRecords(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)

	data, err := tc.OptionExerciseRecords(model.OptionExercisePageRequest{})
	if err != nil {
		t.Fatalf("OptionExerciseRecords: %v", err)
	}
	if data == nil {
		t.Fatal("OptionExerciseRecords returned nil")
	}
	// May have no exercise records.
	if data.ItemCount > 0 && len(data.Items) > 0 {
		if data.Items[0].Symbol == "" {
			t.Errorf("OptionExerciseRecord.Symbol is empty, want non-empty")
		}
	}
}

// TestIntegAssetsNumericConstraints 验证账户资产的数值业务约束：
//   - BuyingPower >= 0
//   - NetLiquidation >= 0
//   - Segments[].Category 非空
func TestIntegAssetsNumericConstraints(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)

	data, err := tc.Assets(model.AssetsRequest{})
	if err != nil {
		t.Fatalf("Assets: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("Assets returned empty")
	}
	for i, asset := range data {
		if asset.BuyingPower < 0 {
			t.Errorf("Asset[%d].BuyingPower = %v, want >= 0", i, asset.BuyingPower)
		}
		if asset.NetLiquidation < 0 {
			t.Errorf("Asset[%d].NetLiquidation = %v, want >= 0", i, asset.NetLiquidation)
		}
		for j, seg := range asset.Segments {
			if seg.Category == "" {
				t.Errorf("Asset[%d].Segments[%d].Category is empty", i, j)
			}
		}
	}
}

// TestIntegPositionsNumericConstraints 验证持仓数值业务约束（有持仓时）：
//   - Quantity != 0
//   - AverageCost > 0
//   - MarketValue 与 Quantity 同号
func TestIntegPositionsNumericConstraints(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)

	data, err := tc.Positions(model.PositionsRequest{})
	if err != nil {
		t.Fatalf("Positions: %v", err)
	}
	if len(data) == 0 {
		t.Skip("no positions to validate (empty account)")
	}
	for i, p := range data {
		if p.Quantity == 0 {
			t.Errorf("Position[%d](%s): Quantity = 0, want != 0", i, p.Symbol)
		}
		if p.AverageCost <= 0 {
			t.Errorf("Position[%d](%s): AverageCost = %v, want > 0", i, p.Symbol, p.AverageCost)
		}
		if p.Quantity > 0 && p.MarketValue < 0 {
			t.Errorf("Position[%d](%s): Quantity=%v > 0 but MarketValue=%v < 0",
				i, p.Symbol, p.Quantity, p.MarketValue)
		}
		if p.Quantity < 0 && p.MarketValue > 0 {
			t.Errorf("Position[%d](%s): Quantity=%v < 0 but MarketValue=%v > 0",
				i, p.Symbol, p.Quantity, p.MarketValue)
		}
	}
}

// TestIntegPreviewOrderFields 验证 PreviewOrder 返回的费用/保证金字段存在：
//   - Commission >= 0，CommissionCurrency 非空
//   - InitMargin >= 0，MaintMargin >= 0
func TestIntegPreviewOrderFields(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)

	result, err := tc.PreviewOrder(model.OrderRequest{
		Symbol:        "AAPL",
		SecType:       "STK",
		Action:        "BUY",
		OrderType:     "MKT",
		TotalQuantity: 1,
		Currency:      "USD",
		Market:        "US",
	})
	if err != nil {
		t.Fatalf("PreviewOrder: %v", err)
	}
	if result == nil {
		t.Fatal("PreviewOrder returned nil")
	}
	if result.Commission < 0 {
		t.Errorf("PreviewResult.Commission = %v, want >= 0", result.Commission)
	}
	if result.CommissionCurrency == "" {
		t.Errorf("PreviewResult.CommissionCurrency is empty, want non-empty")
	}
	if result.InitMargin < 0 {
		t.Errorf("PreviewResult.InitMargin = %v, want >= 0", result.InitMargin)
	}
	if result.MaintMargin < 0 {
		t.Errorf("PreviewResult.MaintMargin = %v, want >= 0", result.MaintMargin)
	}
}
