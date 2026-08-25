package trade

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tigerfintech/openapi-go-sdk/model"
)

// 本文件补充 trade_client.go 中尚未覆盖的方法测试，目标覆盖率 >= 80%。
// 复用 trade_client_test.go 中的 mockServer / mockErrorServer / newTestTradeClient 等辅助函数。

// captureServer 捕获请求 biz_content 用于验证 wire payload（与 quote 测试中同名的辅助函数模式一致）。
func captureServer(t *testing.T, captured *map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
			if bizStr, ok := body["biz_content"].(string); ok {
				var biz map[string]interface{}
				if err := json.Unmarshal([]byte(bizStr), &biz); err == nil {
					*captured = biz
				}
			}
		}
		resp := map[string]interface{}{"code": 0, "message": "success", "data": json.RawMessage("[]"), "timestamp": 1700000000}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
}

// === SetSecretKey 测试 ===

func TestSetSecretKey(t *testing.T) {
	tc := newTestTradeClient("http://unused.local")
	tc.SetSecretKey("my-institutional-secret")
	if tc.secretKey != "my-institutional-secret" {
		t.Fatalf("SetSecretKey 未生效，实际: %q", tc.secretKey)
	}
}

// === 15.2 账户管理测试 ===

func TestManagedAccounts(t *testing.T) {
	server := mockServer(t, map[string]interface{}{
		"items": []map[string]interface{}{
			{"accountId": "U123", "name": "Sub1"},
			{"accountId": "U456", "name": "Sub2"},
		},
	})
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	data, err := tc.ManagedAccounts(model.ManagedAccountsRequest{})
	if err != nil {
		t.Fatalf("ManagedAccounts 失败: %v", err)
	}
	if len(data) != 2 {
		t.Fatalf("ManagedAccounts 期望 2 条，实际 %d", len(data))
	}
}

// === 衍生品合约批量查询测试 ===

func TestDerivativeContracts(t *testing.T) {
	server := mockServer(t, map[string]interface{}{
		"items": []map[string]interface{}{
			{"symbol": "AAPL", "secType": "OPT"},
		},
	})
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	data, err := tc.DerivativeContracts(model.DerivativeContractsRequest{
		Symbols: []string{"AAPL"},
		SecType: "OPT",
	})
	if err != nil {
		t.Fatalf("DerivativeContracts 失败: %v", err)
	}
	if len(data) != 1 {
		t.Fatalf("DerivativeContracts 期望 1 条，实际 %d", len(data))
	}
}

// === 15.4 资产分析测试 ===

func TestAnalyticsAsset(t *testing.T) {
	server := mockServer(t, map[string]interface{}{
		"items": []map[string]interface{}{
			{"netLiquidation": 100000.0, "date": "2026-08-04"},
		},
	})
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	data, err := tc.AnalyticsAsset(model.AnalyticsAssetRequest{
		StartDate: "2026-08-01",
		EndDate:   "2026-08-04",
	})
	if err != nil {
		t.Fatalf("AnalyticsAsset 失败: %v", err)
	}
	if len(data) != 1 {
		t.Fatalf("AnalyticsAsset 期望 1 条，实际 %d", len(data))
	}
}

func TestAggregateAssets(t *testing.T) {
	server := mockServer(t, map[string]interface{}{
		"baseCurrency":   "USD",
		"netLiquidation": 200000.0,
	})
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	data, err := tc.AggregateAssets(model.AggregateAssetsRequest{
		BaseCurrency: "USD",
	})
	if err != nil {
		t.Fatalf("AggregateAssets 失败: %v", err)
	}
	if data == nil {
		t.Fatal("AggregateAssets 返回 nil")
	}
}

// === 估算/外汇测试 ===

func TestEstimateTradableQuantity(t *testing.T) {
	server := mockServer(t, map[string]interface{}{
		"quantity": 100.0,
		"currency": "USD",
	})
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	data, err := tc.EstimateTradableQuantity(model.EstimateTradableQuantityRequest{
		Symbol:    "AAPL",
		SecType:   "STK",
		Action:    "BUY",
		OrderType: "LMT",
	})
	if err != nil {
		t.Fatalf("EstimateTradableQuantity 失败: %v", err)
	}
	if data == nil {
		t.Fatal("EstimateTradableQuantity 返回 nil")
	}
}

func TestPlaceForexOrder(t *testing.T) {
	server := mockServer(t, map[string]interface{}{
		"orderId": "FX-001",
		"status":  "Submitted",
	})
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	data, err := tc.PlaceForexOrder(model.ForexOrderRequest{
		SourceCurrency: "USD",
		SourceAmount:   1000.0,
		TargetCurrency: "HKD",
	})
	if err != nil {
		t.Fatalf("PlaceForexOrder 失败: %v", err)
	}
	if data == nil {
		t.Fatal("PlaceForexOrder 返回 nil")
	}
}

// === 子账户资金调拨测试 ===

func TestSegmentFundAvailable(t *testing.T) {
	server := mockServer(t, []map[string]interface{}{
		{"segmentType": "SEC", "available": 5000.0},
		{"segmentType": "CASH", "available": 2000.0},
	})
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	data, err := tc.SegmentFundAvailable(model.SegmentFundRequest{
		Currency: "USD",
	})
	if err != nil {
		t.Fatalf("SegmentFundAvailable 失败: %v", err)
	}
	if len(data) != 2 {
		t.Fatalf("SegmentFundAvailable 期望 2 条，实际 %d", len(data))
	}
}

func TestSegmentFundHistory(t *testing.T) {
	server := mockServer(t, []map[string]interface{}{
		{"id": 1, "amount": 1000.0, "currency": "USD"},
	})
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	data, err := tc.SegmentFundHistory(model.SegmentFundRequest{
		Currency: "USD",
	})
	if err != nil {
		t.Fatalf("SegmentFundHistory 失败: %v", err)
	}
	if len(data) != 1 {
		t.Fatalf("SegmentFundHistory 期望 1 条，实际 %d", len(data))
	}
}

func TestTransferSegmentFund(t *testing.T) {
	server := mockServer(t, map[string]interface{}{
		"id":     "SF-002",
		"amount": 500.0,
		"status": "Processing",
	})
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	data, err := tc.TransferSegmentFund(model.SegmentFundRequest{
		FromSegment: "SEC",
		ToSegment:   "CASH",
		Currency:    "USD",
		Amount:      500.0,
	})
	if err != nil {
		t.Fatalf("TransferSegmentFund 失败: %v", err)
	}
	if data == nil {
		t.Fatal("TransferSegmentFund 返回 nil")
	}
}

func TestCancelSegmentFund(t *testing.T) {
	server := mockServer(t, map[string]interface{}{
		"id":     "SF-002",
		"status": "Cancelled",
	})
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	data, err := tc.CancelSegmentFund(model.SegmentFundRequest{
		ID: "SF-002",
	})
	if err != nil {
		t.Fatalf("CancelSegmentFund 失败: %v", err)
	}
	if data == nil {
		t.Fatal("CancelSegmentFund 返回 nil")
	}
}

// === 资金明细/历史测试 ===

func TestFundDetails(t *testing.T) {
	server := mockServer(t, map[string]interface{}{
		"items": []map[string]interface{}{
			{"id": 1, "amount": 100.0, "currency": "USD"},
		},
	})
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	data, err := tc.FundDetails(model.FundDetailsRequest{
		Currency: "USD",
	})
	if err != nil {
		t.Fatalf("FundDetails 失败: %v", err)
	}
	if len(data) != 1 {
		t.Fatalf("FundDetails 期望 1 条，实际 %d", len(data))
	}
}

func TestFundingHistory(t *testing.T) {
	server := mockServer(t, []map[string]interface{}{
		{"id": 1, "amount": 200.0, "currency": "USD"},
	})
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	data, err := tc.FundingHistory(model.FundingHistoryRequest{
		SegType: "SEC",
	})
	if err != nil {
		t.Fatalf("FundingHistory 失败: %v", err)
	}
	if len(data) != 1 {
		t.Fatalf("FundingHistory 期望 1 条，实际 %d", len(data))
	}
}

// === 内部/外部转股测试 ===

func TestTransferPosition(t *testing.T) {
	server := mockServer(t, map[string]interface{}{
		"id":     "PT-001",
		"status": "Submitted",
	})
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	data, err := tc.TransferPosition(model.PositionTransferRequest{
		FromAccount: "U123",
		ToAccount:   "U456",
		Market:      "US",
	})
	if err != nil {
		t.Fatalf("TransferPosition 失败: %v", err)
	}
	if data == nil {
		t.Fatal("TransferPosition 返回 nil")
	}
}

func TestPositionTransferRecords(t *testing.T) {
	server := mockServer(t, []map[string]interface{}{
		{"id": "PT-001", "status": "Submitted"},
	})
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	data, err := tc.PositionTransferRecords(model.PositionTransferRecordsRequest{
		SinceDate: "2026-01-01",
		ToDate:    "2026-08-04",
	})
	if err != nil {
		t.Fatalf("PositionTransferRecords 失败: %v", err)
	}
	if len(data) != 1 {
		t.Fatalf("PositionTransferRecords 期望 1 条，实际 %d", len(data))
	}
}

func TestPositionTransferDetail(t *testing.T) {
	server := mockServer(t, map[string]interface{}{
		"id":     "PT-001",
		"status": "Submitted",
	})
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	data, err := tc.PositionTransferDetail(model.PositionTransferDetailRequest{
		ID: "PT-001",
	})
	if err != nil {
		t.Fatalf("PositionTransferDetail 失败: %v", err)
	}
	if data == nil {
		t.Fatal("PositionTransferDetail 返回 nil")
	}
}

func TestPositionTransferExternalRecords(t *testing.T) {
	server := mockServer(t, []map[string]interface{}{
		{"id": "PTX-001", "status": "Completed"},
	})
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	data, err := tc.PositionTransferExternalRecords(model.PositionTransferExternalRecordsRequest{
		SinceDate: "2026-01-01",
	})
	if err != nil {
		t.Fatalf("PositionTransferExternalRecords 失败: %v", err)
	}
	if len(data) != 1 {
		t.Fatalf("PositionTransferExternalRecords 期望 1 条，实际 %d", len(data))
	}
}

// === 期权行权测试 ===

func TestOptionExerciseCheck(t *testing.T) {
	server := mockServer(t, map[string]interface{}{
		"contractId": 12345,
		"result":     "ok",
	})
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	data, err := tc.OptionExerciseCheck(model.OptionExerciseCheckRequest{
		ContractId: 12345,
		Type:       "Exercise",
		Quantity:   1.0,
	})
	if err != nil {
		t.Fatalf("OptionExerciseCheck 失败: %v", err)
	}
	if data == nil {
		t.Fatal("OptionExerciseCheck 返回 nil")
	}
}

func TestOptionExercisePositions(t *testing.T) {
	server := mockServer(t, map[string]interface{}{
		"page":      1,
		"totalPage": 2,
		"items": []map[string]interface{}{
			{"contractId": 12345, "symbol": "AAPL"},
		},
	})
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	data, err := tc.OptionExercisePositions(model.OptionExercisePositionRequest{
		Type: "Exercise",
	})
	if err != nil {
		t.Fatalf("OptionExercisePositions 失败: %v", err)
	}
	if data == nil {
		t.Fatal("OptionExercisePositions 返回 nil")
	}
}

func TestOptionExerciseSubmit(t *testing.T) {
	server := mockServer(t, true)
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	ok, err := tc.OptionExerciseSubmit(model.OptionExerciseSubmitRequest{
		ContractId:    12345,
		Type:          "Exercise",
		Quantity:      1.0,
		ExecutingDate: "2026-08-19",
	})
	if err != nil {
		t.Fatalf("OptionExerciseSubmit 失败: %v", err)
	}
	if !ok {
		t.Fatal("OptionExerciseSubmit 期望返回 true")
	}
}

func TestOptionExerciseRecords(t *testing.T) {
	server := mockServer(t, map[string]interface{}{
		"page":       1,
		"totalPage":  1,
		"totalCount": 2,
		"items": []map[string]interface{}{
			{"id": 1, "status": "Success"},
			{"id": 2, "status": "Cancel"},
		},
	})
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	data, err := tc.OptionExerciseRecords(model.OptionExercisePageRequest{
		Page: 1,
		Size: 20,
	})
	if err != nil {
		t.Fatalf("OptionExerciseRecords 失败: %v", err)
	}
	if data == nil {
		t.Fatal("OptionExerciseRecords 返回 nil")
	}
}

func TestOptionExerciseCancel(t *testing.T) {
	server := mockServer(t, true)
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	ok, err := tc.OptionExerciseCancel(model.OptionExerciseCancelRequest{
		Id: 100,
	})
	if err != nil {
		t.Fatalf("OptionExerciseCancel 失败: %v", err)
	}
	if !ok {
		t.Fatal("OptionExerciseCancel 期望返回 true")
	}
}

// === 错误路径补充测试 ===

func TestManagedAccountsError(t *testing.T) {
	server := mockErrorServer(t, 1100, "查询机构子账户失败")
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	_, err := tc.ManagedAccounts(model.ManagedAccountsRequest{})
	if err == nil {
		t.Fatal("期望返回错误，但返回了 nil")
	}
}

func TestAggregateAssetsError(t *testing.T) {
	server := mockErrorServer(t, 1100, "综合资产查询失败")
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	_, err := tc.AggregateAssets(model.AggregateAssetsRequest{})
	if err == nil {
		t.Fatal("期望返回错误，但返回了 nil")
	}
}

func TestTransferSegmentFundError(t *testing.T) {
	server := mockErrorServer(t, 1100, "子账户调拨失败")
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	_, err := tc.TransferSegmentFund(model.SegmentFundRequest{})
	if err == nil {
		t.Fatal("期望返回错误，但返回了 nil")
	}
}

func TestOptionExerciseSubmitError(t *testing.T) {
	server := mockErrorServer(t, 1100, "行权提交失败")
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	_, err := tc.OptionExerciseSubmit(model.OptionExerciseSubmitRequest{})
	if err == nil {
		t.Fatal("期望返回错误，但返回了 nil")
	}
}

func TestPositionTransferDetailError(t *testing.T) {
	server := mockErrorServer(t, 1100, "转股详情查询失败")
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	_, err := tc.PositionTransferDetail(model.PositionTransferDetailRequest{})
	if err == nil {
		t.Fatal("期望返回错误，但返回了 nil")
	}
}

// === 默认账户与 SecretKey 注入测试 ===

// TestSecretKeyAutoInjection 验证机构 SecretKey 被自动注入到请求参数中。
func TestSecretKeyAutoInjection(t *testing.T) {
	var captured map[string]interface{}
	server := captureServer(t, &captured)
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	tc.SetSecretKey("inst_secret_abc")

	_, _ = tc.AggregateAssets(model.AggregateAssetsRequest{})
	if v, ok := captured["secret_key"].(string); !ok || v != "inst_secret_abc" {
		t.Errorf("期望 secret_key=inst_secret_abc，得到: %v", captured["secret_key"])
	}
	if v, ok := captured["account"].(string); !ok || v != "test_account" {
		t.Errorf("期望 account=test_account，得到: %v", captured["account"])
	}
}

// TestExplicitAccountOverridesDefault 验证显式传入 Account 时不被默认账户覆盖，
// 且 SecretKey 不被自动注入（因为已显式提供 Account 分支不进入 SecretKey 注入）。
func TestExplicitAccountOverridesDefault(t *testing.T) {
	var captured map[string]interface{}
	server := captureServer(t, &captured)
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	tc.SetSecretKey("explicit_sk")
	req := model.AssetsRequest{Account: "UExplicit", SecretKey: "user_sk"}
	_, _ = tc.Assets(req)
	if v, ok := captured["account"].(string); !ok || v != "UExplicit" {
		t.Errorf("期望 account=UExplicit，得到: %v", captured["account"])
	}
	if v, ok := captured["secret_key"].(string); !ok || v != "user_sk" {
		t.Errorf("期望 secret_key=user_sk，得到: %v", captured["secret_key"])
	}
}

// TestPlaceOrderSecretKeyInjection 验证 PlaceOrder 自动注入 SecretKey（机构场景）。
func TestPlaceOrderSecretKeyInjection(t *testing.T) {
	var captured map[string]interface{}
	server := captureServer(t, &captured)
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	tc.SetSecretKey("inst_place_sk")
	_, _ = tc.PlaceOrder(model.OrderRequest{
		Symbol:        "AAPL",
		SecType:       "STK",
		Action:        "BUY",
		OrderType:     "LMT",
		TotalQuantity: 10,
	})
	if v, ok := captured["secret_key"].(string); !ok || v != "inst_place_sk" {
		t.Errorf("期望 secret_key=inst_place_sk，得到: %v", captured["secret_key"])
	}
	if v, ok := captured["account"].(string); !ok || v != "test_account" {
		t.Errorf("期望 account=test_account，得到: %v", captured["account"])
	}
}

// TestCancelOrderWithSecretKey 验证 CancelOrder 在有 secretKey 时注入 secret_key 字段。
func TestCancelOrderWithSecretKey(t *testing.T) {
	var captured map[string]interface{}
	server := captureServer(t, &captured)
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	tc.SetSecretKey("cancel_sk")
	_, _ = tc.CancelOrder(999)
	if v, ok := captured["secret_key"].(string); !ok || v != "cancel_sk" {
		t.Errorf("期望 secret_key=cancel_sk，得到: %v", captured["secret_key"])
	}
	if v, ok := captured["id"].(float64); !ok || int64(v) != 999 {
		t.Errorf("期望 id=999，得到: %v", captured["id"])
	}
}

// TestGetOrderEmptyReturnsNil 验证 GetOrder 在服务端返回空对象（无 id/orderId）时返回 nil。
func TestGetOrderEmptyReturnsNil(t *testing.T) {
	server := mockServer(t, map[string]interface{}{})
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	data, err := tc.GetOrder(model.GetOrderRequest{Id: 12345})
	if err != nil {
		t.Fatalf("GetOrder 失败: %v", err)
	}
	if data != nil {
		t.Fatalf("GetOrder 空响应期望 nil，得到 %+v", data)
	}
}

// TestGetOrderError 验证 GetOrder 在服务端报错时返回错误。
func TestGetOrderError(t *testing.T) {
	server := mockErrorServer(t, 1100, "订单查询失败")
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	_, err := tc.GetOrder(model.GetOrderRequest{Id: 12345})
	if err == nil {
		t.Fatal("期望返回错误，但返回了 nil")
	}
}

// TestContract3Error 验证 Contract3 在服务端报错时返回错误。
func TestContract3Error(t *testing.T) {
	server := mockErrorServer(t, 1100, "合约查询失败")
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	_, err := tc.Contract3("AAPL", "STK")
	if err == nil {
		t.Fatal("期望返回错误，但返回了 nil")
	}
}

// === callIntoItems 边界路径测试 ===

// TestContractsWithNullItems 验证当服务端返回 items=null 时 callIntoItems 返回 nil 切片且无错误。
func TestContractsWithNullItems(t *testing.T) {
	server := mockServer(t, map[string]interface{}{
		"items": nil,
	})
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	data, err := tc.Contracts([]string{"AAPL"}, "STK")
	if err != nil {
		t.Fatalf("Contracts 不期望错误: %v", err)
	}
	if data != nil && len(data) != 0 {
		t.Fatalf("期望 nil/空切片，得到: %+v", data)
	}
}

// TestOrdersWithoutItemsField 验证当服务端返回对象不含 items 字段时 callIntoItems 不报错。
func TestOrdersWithoutItemsField(t *testing.T) {
	server := mockServer(t, map[string]interface{}{
		"symbol": "AAPL",
	})
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	data, err := tc.Orders(model.OrdersRequest{})
	if err != nil {
		t.Fatalf("Orders 不期望错误: %v", err)
	}
	if len(data) != 0 {
		t.Fatalf("期望空切片，得到: %+v", data)
	}
}

// === SecretKey 自动注入分支补充 ===

// TestPreviewOrderSecretKeyInjection 验证 PreviewOrder 自动注入 SecretKey。
func TestPreviewOrderSecretKeyInjection(t *testing.T) {
	var captured map[string]interface{}
	server := captureServer(t, &captured)
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	tc.SetSecretKey("preview_sk")
	_, _ = tc.PreviewOrder(model.OrderRequest{
		Symbol:        "AAPL",
		SecType:       "STK",
		Action:        "BUY",
		OrderType:     "MKT",
		TotalQuantity: 10,
	})
	if v, ok := captured["secret_key"].(string); !ok || v != "preview_sk" {
		t.Errorf("期望 secret_key=preview_sk，得到: %v", captured["secret_key"])
	}
}

// TestModifyOrderSecretKeyInjection 验证 ModifyOrder 自动注入 SecretKey。
func TestModifyOrderSecretKeyInjection(t *testing.T) {
	var captured map[string]interface{}
	server := captureServer(t, &captured)
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	tc.SetSecretKey("modify_sk")
	_, _ = tc.ModifyOrder(555, model.OrderRequest{
		OrderType:     "LMT",
		TotalQuantity: 20,
	})
	if v, ok := captured["secret_key"].(string); !ok || v != "modify_sk" {
		t.Errorf("期望 secret_key=modify_sk，得到: %v", captured["secret_key"])
	}
	if v, ok := captured["id"].(float64); !ok || int64(v) != 555 {
		t.Errorf("期望 id=555，得到: %v", captured["id"])
	}
}

// TestPrimeAssetsSecretKeyInjection 验证 PrimeAssets 自动注入 SecretKey。
func TestPrimeAssetsSecretKeyInjection(t *testing.T) {
	var captured map[string]interface{}
	server := captureServer(t, &captured)
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	tc.SetSecretKey("prime_sk")
	_, _ = tc.PrimeAssets(model.AssetsRequest{})
	if v, ok := captured["secret_key"].(string); !ok || v != "prime_sk" {
		t.Errorf("期望 secret_key=prime_sk，得到: %v", captured["secret_key"])
	}
}

// TestCancelSegmentFundSecretKeyInjection 验证 CancelSegmentFund 自动注入 SecretKey。
func TestCancelSegmentFundSecretKeyInjection(t *testing.T) {
	var captured map[string]interface{}
	server := captureServer(t, &captured)
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	tc.SetSecretKey("cancel_sf_sk")
	_, _ = tc.CancelSegmentFund(model.SegmentFundRequest{})
	if v, ok := captured["secret_key"].(string); !ok || v != "cancel_sf_sk" {
		t.Errorf("期望 secret_key=cancel_sf_sk，得到: %v", captured["secret_key"])
	}
}

// TestOptionExerciseCheckSecretKeyInjection 验证 OptionExerciseCheck 自动注入 SecretKey。
func TestOptionExerciseCheckSecretKeyInjection(t *testing.T) {
	var captured map[string]interface{}
	server := captureServer(t, &captured)
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	tc.SetSecretKey("check_sk")
	_, _ = tc.OptionExerciseCheck(model.OptionExerciseCheckRequest{ContractId: 1, Type: "Exercise"})
	if v, ok := captured["secret_key"].(string); !ok || v != "check_sk" {
		t.Errorf("期望 secret_key=check_sk，得到: %v", captured["secret_key"])
	}
}

// TestOptionExerciseRecordsSecretKeyInjection 验证 OptionExerciseRecords 自动注入 SecretKey。
func TestOptionExerciseRecordsSecretKeyInjection(t *testing.T) {
	var captured map[string]interface{}
	server := captureServer(t, &captured)
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	tc.SetSecretKey("records_sk")
	_, _ = tc.OptionExerciseRecords(model.OptionExercisePageRequest{})
	if v, ok := captured["secret_key"].(string); !ok || v != "records_sk" {
		t.Errorf("期望 secret_key=records_sk，得到: %v", captured["secret_key"])
	}
}

// TestTransferPositionAutoInjectFromAccount 验证 TransferPosition 在 FromAccount 为空时自动填入默认账户与 SecretKey。
func TestTransferPositionAutoInjectFromAccount(t *testing.T) {
	var captured map[string]interface{}
	server := captureServer(t, &captured)
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	tc.SetSecretKey("tp_sk")
	// FromAccount 留空，触发自动注入分支
	_, _ = tc.TransferPosition(model.PositionTransferRequest{
		ToAccount: "U456",
		Market:    "US",
	})
	if v, ok := captured["from_account"].(string); !ok || v != "test_account" {
		t.Errorf("期望 from_account=test_account，得到: %v", captured["from_account"])
	}
	if v, ok := captured["secret_key"].(string); !ok || v != "tp_sk" {
		t.Errorf("期望 secret_key=tp_sk，得到: %v", captured["secret_key"])
	}
}

// TestPositionTransferRecordsAutoInjectAccount 验证 PositionTransferRecords 在 AccountID 为空时自动填入默认账户。
func TestPositionTransferRecordsAutoInjectAccount(t *testing.T) {
	var captured map[string]interface{}
	server := captureServer(t, &captured)
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	tc.SetSecretKey("ptr_sk")
	_, _ = tc.PositionTransferRecords(model.PositionTransferRecordsRequest{})
	if v, ok := captured["account_id"].(string); !ok || v != "test_account" {
		t.Errorf("期望 account_id=test_account，得到: %v", captured["account_id"])
	}
	if v, ok := captured["secret_key"].(string); !ok || v != "ptr_sk" {
		t.Errorf("期望 secret_key=ptr_sk，得到: %v", captured["secret_key"])
	}
}

// TestEstimateTradableQuantitySecretKeyInjection 验证 EstimateTradableQuantity 自动注入 SecretKey。
func TestEstimateTradableQuantitySecretKeyInjection(t *testing.T) {
	var captured map[string]interface{}
	server := captureServer(t, &captured)
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	tc.SetSecretKey("etq_sk")
	_, _ = tc.EstimateTradableQuantity(model.EstimateTradableQuantityRequest{Symbol: "AAPL"})
	if v, ok := captured["secret_key"].(string); !ok || v != "etq_sk" {
		t.Errorf("期望 secret_key=etq_sk，得到: %v", captured["secret_key"])
	}
}

// TestPlaceForexOrderSecretKeyInjection 验证 PlaceForexOrder 自动注入 SecretKey。
func TestPlaceForexOrderSecretKeyInjection(t *testing.T) {
	var captured map[string]interface{}
	server := captureServer(t, &captured)
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	tc.SetSecretKey("forex_sk")
	_, _ = tc.PlaceForexOrder(model.ForexOrderRequest{SourceCurrency: "USD"})
	if v, ok := captured["secret_key"].(string); !ok || v != "forex_sk" {
		t.Errorf("期望 secret_key=forex_sk，得到: %v", captured["secret_key"])
	}
}

// TestOptionExerciseCancelSecretKeyInjection 验证 OptionExerciseCancel 自动注入 SecretKey。
func TestOptionExerciseCancelSecretKeyInjection(t *testing.T) {
	var captured map[string]interface{}
	server := captureServer(t, &captured)
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	tc.SetSecretKey("oec_sk")
	_, _ = tc.OptionExerciseCancel(model.OptionExerciseCancelRequest{Id: 5})
	if v, ok := captured["secret_key"].(string); !ok || v != "oec_sk" {
		t.Errorf("期望 secret_key=oec_sk，得到: %v", captured["secret_key"])
	}
}

// TestOptionExercisePositionsSecretKeyInjection 验证 OptionExercisePositions 自动注入 SecretKey。
func TestOptionExercisePositionsSecretKeyInjection(t *testing.T) {
	var captured map[string]interface{}
	server := captureServer(t, &captured)
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	tc.SetSecretKey("oep_sk")
	_, _ = tc.OptionExercisePositions(model.OptionExercisePositionRequest{Type: "Exercise"})
	if v, ok := captured["secret_key"].(string); !ok || v != "oep_sk" {
		t.Errorf("期望 secret_key=oep_sk，得到: %v", captured["secret_key"])
	}
}

// TestPositionTransferDetailSecretKeyInjection 验证 PositionTransferDetail 自动注入 SecretKey。
func TestPositionTransferDetailSecretKeyInjection(t *testing.T) {
	var captured map[string]interface{}
	server := captureServer(t, &captured)
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	tc.SetSecretKey("ptd_sk")
	_, _ = tc.PositionTransferDetail(model.PositionTransferDetailRequest{ID: "PT-1"})
	if v, ok := captured["secret_key"].(string); !ok || v != "ptd_sk" {
		t.Errorf("期望 secret_key=ptd_sk，得到: %v", captured["secret_key"])
	}
	if v, ok := captured["account_id"].(string); !ok || v != "test_account" {
		t.Errorf("期望 account_id=test_account，得到: %v", captured["account_id"])
	}
}
