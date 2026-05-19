package trade

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"github.com/tigerfintech/openapi-go-sdk/client"
	"github.com/tigerfintech/openapi-go-sdk/config"
	"github.com/tigerfintech/openapi-go-sdk/model"
	"time"
)

// mustGenerateKeyPEM 生成测试用 PKCS#1 PEM 格式私钥
func mustGenerateKeyPEM() string {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic("生成 RSA 密钥对失败: " + err.Error())
	}
	privDER := x509.MarshalPKCS1PrivateKey(privateKey)
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privDER,
	})
	return string(privPEM)
}

var cachedTestPrivateKey = mustGenerateKeyPEM()

// mockServer 创建一个 mock HTTP 服务器，返回指定的 data
func mockServer(t *testing.T, data interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dataBytes, _ := json.Marshal(data)
		resp := map[string]interface{}{
			"code":      0,
			"message":   "success",
			"data":      json.RawMessage(dataBytes),
			"timestamp": 1700000000,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
}

// mockErrorServer 创建一个返回错误的 mock HTTP 服务器
func mockErrorServer(t *testing.T, code int, message string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"code":      code,
			"message":   message,
			"timestamp": 1700000000,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
}

func newTestTradeClient(serverURL string) *TradeClient {
	cfg := &config.ClientConfig{
		TigerID:    "test_tiger_id",
		PrivateKey: cachedTestPrivateKey,
		Account:    "test_account",
		Language:   "zh_CN",
		Timeout:    5 * time.Second,
		ServerURL:  serverURL,
	}
	httpClient := client.NewHttpClient(cfg)
	return NewTradeClient(httpClient, "test_account")
}

// === 15.1 合约查询测试 ===

func TestGetContract(t *testing.T) {
	server := mockServer(t, map[string]interface{}{
		"items": []map[string]interface{}{
			{"symbol": "AAPL", "secType": "STK"},
		},
	})
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	data, err := tc.Contract("AAPL", "STK")
	if err != nil {
		t.Fatalf("GetContract 失败: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("GetContract 返回空")
	}
}

func TestGetContract3(t *testing.T) {
	server := mockServer(t, map[string]interface{}{
		"symbol": "AAPL", "secType": "STK",
	})
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	data, err := tc.Contract3("AAPL", "STK")
	if err != nil {
		t.Fatalf("GetContract3 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetContract3 返回 nil")
	}
	if data.Symbol != "AAPL" {
		t.Fatalf("GetContract3 symbol 期望 AAPL, 实际 %s", data.Symbol)
	}
}

func TestGetContracts(t *testing.T) {
	server := mockServer(t, map[string]interface{}{
		"items": []map[string]interface{}{
			{"symbol": "AAPL", "secType": "STK"},
			{"symbol": "GOOG", "secType": "STK"},
		},
	})
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	data, err := tc.Contracts([]string{"AAPL", "GOOG"}, "STK")
	if err != nil {
		t.Fatalf("GetContracts 失败: %v", err)
	}
	if len(data) != 2 {
		t.Fatalf("GetContracts 期望返回 2 条,实际 %d", len(data))
	}
}

func TestGetQuoteContract(t *testing.T) {
	server := mockServer(t, map[string]interface{}{
		"symbol":  "AAPL",
		"secType": "OPT",
		"items":   []map[string]interface{}{{"symbol": "AAPL", "secType": "OPT"}},
	})
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	data, err := tc.QuoteContract("AAPL", "OPT", "20260619")
	if err != nil {
		t.Fatalf("GetQuoteContract 失败: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("GetQuoteContract 返回空")
	}
}

// === 15.3 订单操作测试 ===

func TestPlaceOrder(t *testing.T) {
	server := mockServer(t, map[string]interface{}{
		"id": 12345, "orderId": 1,
	})
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	order := model.OrderRequest{
		Symbol:        "AAPL",
		SecType:       "STK",
		Action:        "BUY",
		OrderType:     "LMT",
		TotalQuantity: 100,
		LimitPrice:    150.0,
		TimeInForce:   "DAY",
	}
	data, err := tc.PlaceOrder(order)
	if err != nil {
		t.Fatalf("PlaceOrder 失败: %v", err)
	}
	if data == nil {
		t.Fatal("PlaceOrder 返回 nil data")
	}
}

func TestPreviewOrder(t *testing.T) {
	server := mockServer(t, map[string]interface{}{
		"estimatedCommission": 1.5,
	})
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	order := model.OrderRequest{
		Symbol:        "AAPL",
		SecType:       "STK",
		Action:        "BUY",
		OrderType:     "MKT",
		TotalQuantity: 100,
	}
	data, err := tc.PreviewOrder(order)
	if err != nil {
		t.Fatalf("PreviewOrder 失败: %v", err)
	}
	if data == nil {
		t.Fatal("PreviewOrder 返回 nil data")
	}
}

func TestModifyOrder(t *testing.T) {
	server := mockServer(t, map[string]interface{}{
		"id": 12345,
	})
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	order := model.OrderRequest{
		OrderType:     "LMT",
		TotalQuantity: 200,
		LimitPrice:    155.0,
	}
	data, err := tc.ModifyOrder(12345, order)
	if err != nil {
		t.Fatalf("ModifyOrder 失败: %v", err)
	}
	if data == nil {
		t.Fatal("ModifyOrder 返回 nil data")
	}
}

func TestCancelOrder(t *testing.T) {
	server := mockServer(t, map[string]interface{}{
		"id": 12345,
	})
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	data, err := tc.CancelOrder(12345)
	if err != nil {
		t.Fatalf("CancelOrder 失败: %v", err)
	}
	if data == nil {
		t.Fatal("CancelOrder 返回 nil data")
	}
}

// === 15.5 订单查询测试 ===

func TestGetOrders(t *testing.T) {
	server := mockServer(t, map[string]interface{}{
		"items": []map[string]interface{}{
			{"id": 1, "symbol": "AAPL"},
		},
	})
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	data, err := tc.Orders(model.OrdersRequest{})
	if err != nil {
		t.Fatalf("GetOrders 失败: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("GetOrders 返回空")
	}
}

func TestGetActiveOrders(t *testing.T) {
	server := mockServer(t, map[string]interface{}{
		"items": []map[string]interface{}{
			{"id": 1, "status": "Submitted"},
		},
	})
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	data, err := tc.ActiveOrders(model.OrdersRequest{})
	if err != nil {
		t.Fatalf("GetActiveOrders 失败: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("GetActiveOrders 返回空")
	}
}

func TestGetInactiveOrders(t *testing.T) {
	server := mockServer(t, map[string]interface{}{
		"items": []map[string]interface{}{
			{"id": 1, "status": "Cancelled"},
		},
	})
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	data, err := tc.InactiveOrders(model.OrdersRequest{})
	if err != nil {
		t.Fatalf("GetInactiveOrders 失败: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("GetInactiveOrders 返回空")
	}
}

func TestGetFilledOrders(t *testing.T) {
	server := mockServer(t, map[string]interface{}{
		"items": []map[string]interface{}{
			{"id": 1, "status": "Filled"},
		},
	})
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	data, err := tc.FilledOrders(model.OrdersRequest{StartDate: 0, EndDate: 0})
	if err != nil {
		t.Fatalf("GetFilledOrders 失败: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("GetFilledOrders 返回空")
	}
}

// === 15.7 持仓和资产查询测试 ===

func TestGetPositions(t *testing.T) {
	server := mockServer(t, map[string]interface{}{
		"items": []map[string]interface{}{
			{"symbol": "AAPL", "position": 100},
		},
	})
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	data, err := tc.Positions(model.PositionsRequest{})
	if err != nil {
		t.Fatalf("GetPositions 失败: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("GetPositions 返回空")
	}
}

func TestGetAssets(t *testing.T) {
	server := mockServer(t, map[string]interface{}{
		"items": []map[string]interface{}{
			{"netLiquidation": 100000.0},
		},
	})
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	data, err := tc.Assets(model.AssetsRequest{})
	if err != nil {
		t.Fatalf("GetAssets 失败: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("GetAssets 返回空")
	}
}

func TestGetPrimeAssets(t *testing.T) {
	server := mockServer(t, map[string]interface{}{
		"accountId":      "U123456",
		"netLiquidation": 200000.0,
	})
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	data, err := tc.PrimeAssets(model.AssetsRequest{})
	if err != nil {
		t.Fatalf("GetPrimeAssets 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetPrimeAssets 返回 nil")
	}
}

func TestGetOrderTransactions(t *testing.T) {
	server := mockServer(t, map[string]interface{}{
		"items": []map[string]interface{}{
			{"id": 12345, "filledQuantity": 50},
		},
	})
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	data, err := tc.OrderTransactions(model.OrderTransactionsRequest{OrderId: 12345, Symbol: "AAPL", SecType: "STK"})
	if err != nil {
		t.Fatalf("GetOrderTransactions 失败: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("GetOrderTransactions 返回空")
	}
}

func TestGetOrder(t *testing.T) {
	server := mockServer(t, map[string]interface{}{
		"id":     12345,
		"symbol": "AAPL",
		"status": "Filled",
	})
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	data, err := tc.GetOrder(model.GetOrderRequest{Id: 12345})
	if err != nil {
		t.Fatalf("GetOrder 失败: %v", err)
	}
	if data == nil || data.ID != 12345 {
		t.Fatalf("GetOrder 返回异常: %+v", data)
	}
}

// === 错误处理测试 ===

func TestTradeClientApiError(t *testing.T) {
	server := mockErrorServer(t, 1100, "交易操作失败")
	defer server.Close()

	tc := newTestTradeClient(server.URL)
	_, err := tc.Orders(model.OrdersRequest{})
	if err == nil {
		t.Fatal("期望返回错误，但返回了 nil")
	}
}
