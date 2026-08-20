package quote

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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
func mockServer(t *testing.T, expectedMethod string, data interface{}) *httptest.Server {
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

// mockServerCapture is like mockServer but writes the received biz_content
// field into *capturedBiz so tests can assert wire-shape / default values.
func mockServerCapture(t *testing.T, expectedMethod string, data interface{}, capturedBiz *string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if biz, ok := body["biz_content"].(string); ok {
			*capturedBiz = biz
		}
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

func newTestQuoteClient(serverURL string) *QuoteClient {
	cfg := &config.ClientConfig{
		TigerID:    "test_tiger_id",
		PrivateKey: cachedTestPrivateKey,
		Account:    "test_account",
		Language:   "zh_CN",
		Timeout:    5 * time.Second,
		ServerURL:  serverURL,
	}
	httpClient := client.NewHttpClient(cfg)
	return NewQuoteClient(httpClient)
}

// === 14.1 基础行情测试 ===

func TestGetMarketState(t *testing.T) {
	server := mockServer(t, "market_state", []map[string]interface{}{
		{"market": "US", "status": "Trading"},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetMarketState("US")
	if err != nil {
		t.Fatalf("GetMarketState 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetMarketState 返回 nil data")
	}
}

func TestGetRealTimeQuote(t *testing.T) {
	server := mockServer(t, "quote_real_time", []map[string]interface{}{
		{"symbol": "AAPL", "latestPrice": 150.0},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetRealTimeQuote(model.BriefRequest{Symbols: []string{"AAPL"}})
	if err != nil {
		t.Fatalf("GetRealTimeQuote 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetRealTimeQuote 返回 nil data")
	}
}

func TestGetKline(t *testing.T) {
	server := mockServer(t, "kline", []map[string]interface{}{
		{"symbol": "BTCUSD", "period": "day", "items": []map[string]interface{}{
			{"time": 1700000000000, "volume": 0, "volumeDecimal": 1.25, "open": 1, "close": 1, "high": 1, "low": 1},
			{"time": 1700000001000, "volume": 0, "open": 1, "close": 1, "high": 1, "low": 1},
		}},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetKline(model.KlineRequest{Symbols: []string{"BTCUSD"}, SecType: "CC", Period: "day"})
	if err != nil {
		t.Fatalf("GetKline 失败: %v", err)
	}
	if len(data) != 1 || len(data[0].Items) != 2 {
		t.Fatalf("GetKline 返回数据错误: %+v", data)
	}
	if data[0].Items[0].VolumeDecimal == nil || *data[0].Items[0].VolumeDecimal != 1.25 {
		t.Fatalf("volumeDecimal 解码错误: %+v", data[0].Items[0].VolumeDecimal)
	}
	if data[0].Items[1].VolumeDecimal != nil {
		t.Fatalf("缺失的 volumeDecimal 应解码为 nil: %v", data[0].Items[1].VolumeDecimal)
	}
}

func TestGetTimeline(t *testing.T) {
	var captured map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Errorf("请求解码失败: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"message":"success","data":[{"symbol":"AAPL"}]}`))
	}))
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetTimeline([]string{"AAPL"})
	if err != nil {
		t.Fatalf("GetTimeline 失败: %v", err)
	}
	if captured["version"] != "2.0" {
		t.Fatalf("timeline version = %v, want 2.0", captured["version"])
	}
	if len(data) != 1 || data[0].Symbol != "AAPL" {
		t.Fatalf("GetTimeline 返回数据错误: %+v", data)
	}
}

func TestGetTimelineByReq(t *testing.T) {
	var captured map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Errorf("请求解码失败: %v", err)
		}
		resp := `{"code":0,"message":"success","data":[{"symbol":"BTCUSD","intraday":{"items":[{"time":1700000000000,"volume":0,"volumeDecimal":0,"price":1,"avgPrice":1},{"time":1700000001000,"volume":0,"price":1,"avgPrice":1}]}}]}`
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	}))
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetTimelineByReq(model.TimelineRequest{Symbols: []string{"BTCUSD"}, SecType: "CC"})
	if err != nil {
		t.Fatalf("GetTimelineByReq 失败: %v", err)
	}
	if captured["version"] != "3.0" {
		t.Fatalf("timeline version = %v, want 3.0", captured["version"])
	}
	var biz map[string]interface{}
	if err := json.Unmarshal([]byte(captured["biz_content"].(string)), &biz); err != nil {
		t.Fatalf("biz_content 解码失败: %v", err)
	}
	if biz["sec_type"] != "CC" {
		t.Fatalf("timeline sec_type = %v, want CC", biz["sec_type"])
	}
	if len(data) != 1 || data[0].Intraday == nil || len(data[0].Intraday.Items) != 2 {
		t.Fatalf("GetTimelineByReq 返回数据错误: %+v", data)
	}
	items := data[0].Intraday.Items
	if items[0].VolumeDecimal == nil || *items[0].VolumeDecimal != 0 {
		t.Fatalf("显式零 volumeDecimal 应解码为非 nil 指针: %+v", items[0].VolumeDecimal)
	}
	if items[1].VolumeDecimal != nil {
		t.Fatalf("缺失的 volumeDecimal 应解码为 nil: %v", items[1].VolumeDecimal)
	}
}

func TestGetKlineByPagePropagatesSecType(t *testing.T) {
	var captured map[string]interface{}
	server := captureServer(t, &captured)
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	_, err := qc.GetKlineByPage(model.KlineByPageRequest{
		Symbol: "BTCUSD", SecType: "CC", Period: "day", TotalSize: 1, PageSize: 1,
	})
	if err != nil {
		t.Fatalf("GetKlineByPage 失败: %v", err)
	}
	if captured["sec_type"] != "CC" {
		t.Fatalf("paginated kline sec_type = %v, want CC", captured["sec_type"])
	}
}

func TestGetTradeTick(t *testing.T) {
	var captured string
	server := mockServerCapture(t, "trade_tick", []map[string]interface{}{
		{"symbol": "AAPL"},
	}, &captured)
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetTradeTick(model.TradeTickRequest{Symbols: []string{"AAPL"}, TradeSession: "OverNight"})
	if err != nil {
		t.Fatalf("GetTradeTick 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetTradeTick 返回 nil data")
	}
	if !strings.Contains(captured, `"trade_session":"OverNight"`) {
		t.Errorf("biz_content = %s, want overnight trade_session", captured)
	}
}

func TestGetTradeTick_PartCodeAndCondDecoding(t *testing.T) {
	// Mock response contains partCode and cond fields that should be decoded
	// into human-readable exchange names and condition strings.
	server := mockServer(t, "trade_tick", []map[string]interface{}{
		{
			"symbol":     "AAPL",
			"beginIndex": 0,
			"endIndex":   2,
			"items": []map[string]interface{}{
				{"time": 1700000000000, "volume": 100, "price": 150.25, "type": "+", "partCode": "NYSE", "partName": "New York Stock Exchange, LLC (NYSE)"},
				{"time": 1700000001000, "volume": 200, "price": 150.50, "type": "-", "partCode": "NSDQ", "partName": "NASDAQ Stock Market, LLC (NASDAQ)"},
			},
		},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetTradeTick(model.TradeTickRequest{Symbols: []string{"AAPL"}})
	if err != nil {
		t.Fatalf("GetTradeTick 失败: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("GetTradeTick 返回空")
	}
	items := data[0].Items
	if len(items) != 2 {
		t.Fatalf("items 长度 = %d, want 2", len(items))
	}
	if items[0].PartCode != "NYSE" {
		t.Errorf("items[0].PartCode = %q, want NYSE", items[0].PartCode)
	}
	if items[0].PartName != "New York Stock Exchange, LLC (NYSE)" {
		t.Errorf("items[0].PartName = %q", items[0].PartName)
	}
	if items[1].PartCode != "NSDQ" {
		t.Errorf("items[1].PartCode = %q, want NSDQ", items[1].PartCode)
	}
	if items[1].PartName != "NASDAQ Stock Market, LLC (NASDAQ)" {
		t.Errorf("items[1].PartName = %q", items[1].PartName)
	}
}

func TestGetQuoteDepth(t *testing.T) {
	server := mockServer(t, "quote_depth", []map[string]interface{}{
		{"symbol": "AAPL"},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetQuoteDepth(model.DepthQuoteRequest{Symbols: []string{"AAPL"}, Market: "US"})
	if err != nil {
		t.Fatalf("GetQuoteDepth 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetQuoteDepth 返回 nil data")
	}
}

// === 14.3 期权行情测试 ===

func TestGetOptionExpiration(t *testing.T) {
	server := mockServer(t, "option_expiration", []map[string]interface{}{
		{"symbol": "AAPL", "dates": []string{"2024-01-19"}},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetOptionExpiration([]string{"AAPL"})
	if err != nil {
		t.Fatalf("GetOptionExpiration 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetOptionExpiration 返回 nil data")
	}
}

func TestGetOptionChain(t *testing.T) {
	server := mockServer(t, "option_chain", []map[string]interface{}{
		{"symbol": "AAPL", "expiry": 1705622400000},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetOptionChain([][2]string{{"AAPL", "2024-01-19"}})
	if err != nil {
		t.Fatalf("GetOptionChain 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetOptionChain 返回 nil data")
	}
}

func TestGetOptionQuote(t *testing.T) {
	server := mockServer(t, "option_brief", []map[string]interface{}{
		{"identifier": "AAPL 240119C00150000"},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetOptionQuote([]string{"AAPL 240119C00150000"})
	if err != nil {
		t.Fatalf("GetOptionQuote 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetOptionQuote 返回 nil data")
	}
}

func TestGetOptionKline(t *testing.T) {
	server := mockServer(t, "option_kline", []map[string]interface{}{
		{"identifier": "AAPL 240119C00150000", "period": "day"},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetOptionKline([]string{"AAPL 240119C00150000"}, "day", -1, -1)
	if err != nil {
		t.Fatalf("GetOptionKline 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetOptionKline 返回 nil data")
	}
}

func TestGetDelayedQuote(t *testing.T) {
	server := mockServer(t, "quote_delay", []map[string]interface{}{
		{"symbol": "AAPL", "latestPrice": 148.0},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetDelayedQuote(model.StockDelayBriefsRequest{Symbols: []string{"AAPL"}})
	if err != nil {
		t.Fatalf("GetDelayedQuote 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetDelayedQuote 返回 nil data")
	}
}

func TestGetWarrantQuote(t *testing.T) {
	server := mockServer(t, "warrant_briefs", []map[string]interface{}{
		{"symbol": "28888.HK"},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetWarrantQuote(model.WarrantBriefsRequest{Symbols: []string{"28888.HK"}})
	if err != nil {
		t.Fatalf("GetWarrantQuote 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetWarrantQuote 返回 nil data")
	}
}

// === 14.5 期货行情测试 ===

func TestGetFutureExchange(t *testing.T) {
	server := mockServer(t, "future_exchange", []map[string]interface{}{
		{"code": "CME", "name": "CME", "zoneId": "America/Chicago"},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetFutureExchange()
	if err != nil {
		t.Fatalf("GetFutureExchange 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetFutureExchange 返回 nil data")
	}
}

func TestGetFutureContracts(t *testing.T) {
	server := mockServer(t, "future_contracts", []map[string]interface{}{
		{"symbol": "ES", "exchange": "CME"},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetFutureContracts("CME")
	if err != nil {
		t.Fatalf("GetFutureContracts 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetFutureContracts 返回 nil data")
	}
}

func TestGetFutureRealTimeQuote(t *testing.T) {
	server := mockServer(t, "future_real_time_quote", []map[string]interface{}{
		{"symbol": "ES2312"},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetFutureRealTimeQuote(model.FutureBriefRequest{ContractCodes: []string{"ES2312"}})
	if err != nil {
		t.Fatalf("GetFutureRealTimeQuote 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetFutureRealTimeQuote 返回 nil data")
	}
}

func TestGetFutureKline(t *testing.T) {
	server := mockServer(t, "future_kline", []map[string]interface{}{
		{"symbol": "ES2312", "period": "day"},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetFutureKline(model.FutureKlineRequest{
		ContractCodes: []string{"ES2312"},
		Period:        "day",
		BeginTime:     -1,
		EndTime:       -1,
	})
	if err != nil {
		t.Fatalf("GetFutureKline 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetFutureKline 返回 nil data")
	}
}

// === 14.7 基本面和资金流向测试 ===

func TestGetFinancialDaily(t *testing.T) {
	server := mockServer(t, "financial_daily", []map[string]interface{}{
		{"symbol": "AAPL"},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetFinancialDaily(model.FinancialDailyRequest{
		Symbols:   []string{"AAPL"},
		Market:    "US",
		Fields:    []string{"shares_outstanding"},
		BeginDate: "2024-01-01",
		EndDate:   "2024-01-31",
	})
	if err != nil {
		t.Fatalf("GetFinancialDaily 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetFinancialDaily 返回 nil data")
	}
}

func TestGetFinancialReport(t *testing.T) {
	server := mockServer(t, "financial_report", []map[string]interface{}{
		{"symbol": "AAPL"},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetFinancialReport(model.FinancialReportRequest{
		Symbols:    []string{"AAPL"},
		Market:     "US",
		Fields:     []string{"total_revenue"},
		PeriodType: "Annual",
	})
	if err != nil {
		t.Fatalf("GetFinancialReport 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetFinancialReport 返回 nil data")
	}
}

func TestGetCorporateAction(t *testing.T) {
	server := mockServer(t, "corporate_action", map[string]interface{}{
		"AAPL": []map[string]interface{}{
			{"symbol": "AAPL", "actionType": "DIVIDEND"},
		},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	begin := int64(1704067200000) // 2024-01-01 UTC
	end := int64(1735603200000)   // 2024-12-31 UTC
	data, err := qc.GetCorporateAction(model.CorporateActionRequest{
		Symbols:    []string{"AAPL"},
		Market:     "US",
		ActionType: "DIVIDEND",
		BeginDate:  &begin,
		EndDate:    &end,
	})
	if err != nil {
		t.Fatalf("GetCorporateAction 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetCorporateAction 返回 nil data")
	}
}

func TestGetCorporateSymbolChange(t *testing.T) {
	server := mockServer(t, "corporate_action", map[string]interface{}{
		"META": []map[string]interface{}{
			{"symbol": "META", "actionType": "symbol_change", "oldSymbol": "FB", "newSymbol": "META", "executeDate": "2022-06-09", "market": "US"},
		},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	begin := int64(1640995200000) // 2022-01-01 UTC
	end := int64(1672531200000)   // 2023-01-01 UTC
	data, err := qc.GetCorporateSymbolChange(model.CorporateActionRequest{
		Symbols:   []string{"META"},
		Market:    "US",
		BeginDate: &begin,
		EndDate:   &end,
	})
	if err != nil {
		t.Fatalf("GetCorporateSymbolChange 失败: %v", err)
	}
	if len(data) != 1 {
		t.Fatalf("期望 1 条，实际 %d 条", len(data))
	}
	if data[0].OldSymbol != "FB" || data[0].NewSymbol != "META" {
		t.Errorf("symbol change 字段不匹配: %+v", data[0])
	}
}

func TestGetCorporateDelisting(t *testing.T) {
	server := mockServer(t, "corporate_action", map[string]interface{}{
		"TWTR": []map[string]interface{}{
			{"symbol": "TWTR", "actionType": "delisting", "announcedDate": "2022-10-27", "reason": "acquired", "executeDate": "2022-10-28", "market": "US"},
		},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	begin := int64(1640995200000) // 2022-01-01 UTC
	end := int64(1672531200000)   // 2023-01-01 UTC
	data, err := qc.GetCorporateDelisting(model.CorporateActionRequest{
		Symbols:   []string{"TWTR"},
		Market:    "US",
		BeginDate: &begin,
		EndDate:   &end,
	})
	if err != nil {
		t.Fatalf("GetCorporateDelisting 失败: %v", err)
	}
	if len(data) != 1 {
		t.Fatalf("期望 1 条，实际 %d 条", len(data))
	}
	if data[0].Reason != "acquired" {
		t.Errorf("delisting reason 不匹配: %+v", data[0])
	}
}

func TestGetCorporateIPO(t *testing.T) {
	server := mockServer(t, "corporate_action", map[string]interface{}{
		"RIVN": []map[string]interface{}{
			{
				"symbol": "RIVN", "actionType": "ipo", "ipoName": "Rivian Automotive",
				"listingDate": "2021-11-10", "listingPrice": 78.0,
				"sharesOutstanding": 864000000, "sharesFloat": 153000000,
				"offerAmount": 11932000000.0, "priceRange": "72-74",
				"currency": "USD", "minPurchaseQuantity": 1,
				"leverageRatio": 1.0, "executeDate": "2021-11-10", "market": "US",
			},
		},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	begin := int64(1609459200000) // 2021-01-01 UTC
	end := int64(1640995200000)   // 2022-01-01 UTC
	data, err := qc.GetCorporateIPO(model.CorporateActionRequest{
		Symbols:   []string{"RIVN"},
		Market:    "US",
		BeginDate: &begin,
		EndDate:   &end,
	})
	if err != nil {
		t.Fatalf("GetCorporateIPO 失败: %v", err)
	}
	if len(data) != 1 {
		t.Fatalf("期望 1 条，实际 %d 条", len(data))
	}
	if data[0].IpoName != "Rivian Automotive" || data[0].ListingPrice != 78.0 {
		t.Errorf("IPO 字段不匹配: %+v", data[0])
	}
}

func TestGetCapitalFlow(t *testing.T) {
	server := mockServer(t, "capital_flow", map[string]interface{}{
		"symbol": "AAPL",
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetCapitalFlow("AAPL", "US", "day")
	if err != nil {
		t.Fatalf("GetCapitalFlow 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetCapitalFlow 返回 nil data")
	}
}

func TestGetCapitalDistribution(t *testing.T) {
	server := mockServer(t, "capital_distribution", map[string]interface{}{
		"symbol": "AAPL",
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetCapitalDistribution("AAPL", "US")
	if err != nil {
		t.Fatalf("GetCapitalDistribution 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetCapitalDistribution 返回 nil data")
	}
}

// === 14.9 选股器和行情权限测试 ===

func TestMarketScanner(t *testing.T) {
	server := mockServer(t, "market_scanner", map[string]interface{}{
		"page":       0,
		"totalPage":  1,
		"totalCount": 1,
		"pageSize":   5,
		"items":      []map[string]interface{}{{"symbol": "AAPL", "market": "US"}},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.MarketScanner(model.MarketScannerRequest{Market: "US"})
	if err != nil {
		t.Fatalf("MarketScanner 失败: %v", err)
	}
	if data == nil {
		t.Fatal("MarketScanner 返回 nil data")
	}
}

func TestGrabQuotePermission(t *testing.T) {
	server := mockServer(t, "grab_quote_permission", []map[string]interface{}{
		{"name": "usStockQuote", "expireAt": 1700000000},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GrabQuotePermission()
	if err != nil {
		t.Fatalf("GrabQuotePermission 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GrabQuotePermission 返回 nil data")
	}
}

func TestGetAddonEntitlement(t *testing.T) {
	server := mockServer(t, "addon_entitlements", map[string]interface{}{
		"userLevel": "PRO",
		"activePlan": map[string]interface{}{
			"planType":   "QUOTE_PRO",
			"expireTime": 1700000000,
		},
		"addons": []map[string]interface{}{
			{"planType": "DEPTH", "active": true, "startTime": 1690000000, "expireTime": 1700000000},
		},
		"effectiveEntitlement": map[string]interface{}{
			"subscribeLimit":     100,
			"subscribeRemaining": 80,
			"rateMultiple":       2,
		},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetAddonEntitlement()
	if err != nil {
		t.Fatalf("GetAddonEntitlement 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetAddonEntitlement 返回 nil data")
	}
	if data.UserLevel != "PRO" {
		t.Errorf("UserLevel = %q, 期望 PRO", data.UserLevel)
	}
	if data.ActivePlan == nil || data.ActivePlan.PlanType != "QUOTE_PRO" {
		t.Errorf("ActivePlan 解析错误: %+v", data.ActivePlan)
	}
	if len(data.Addons) != 1 || !data.Addons[0].Active {
		t.Errorf("Addons 解析错误: %+v", data.Addons)
	}
	if data.EffectiveEntitlement == nil || data.EffectiveEntitlement.SubscribeLimit != 100 {
		t.Errorf("EffectiveEntitlement 解析错误: %+v", data.EffectiveEntitlement)
	}
}

// === 错误处理测试 ===

func TestQuoteClientApiError(t *testing.T) {
	server := mockErrorServer(t, 2100, "行情查询失败")
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	_, err := qc.GetMarketState("US")
	if err == nil {
		t.Fatal("期望返回错误，但返回了 nil")
	}
}

// === 多 symbol 请求 payload 验证 ===

// captureServer 捕获请求 body 用于验证 wire payload
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

func TestGetKlineMultiSymbol(t *testing.T) {
	var captured map[string]interface{}
	server := captureServer(t, &captured)
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	qc.GetKline(model.KlineRequest{Symbols: []string{"AAPL", "TSLA"}, Period: "day"})

	syms, ok := captured["symbols"].([]interface{})
	if !ok || len(syms) != 2 {
		t.Fatalf("期望 symbols=[AAPL,TSLA]，得到: %v", captured["symbols"])
	}
}

func TestGetOptionExpirationMultiSymbol(t *testing.T) {
	var captured map[string]interface{}
	server := captureServer(t, &captured)
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	qc.GetOptionExpiration([]string{"AAPL", "TSLA"})

	syms, ok := captured["symbols"].([]interface{})
	if !ok || len(syms) != 2 {
		t.Fatalf("期望 symbols=[AAPL,TSLA]，得到: %v", captured["symbols"])
	}
}

func TestGetOptionChainMultiItem(t *testing.T) {
	var captured map[string]interface{}
	server := captureServer(t, &captured)
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	_, err := qc.GetOptionChain([][2]string{{"AAPL", "2024-01-19"}, {"TSLA", "2024-02-16"}})
	if err != nil {
		t.Fatalf("GetOptionChain 失败: %v", err)
	}

	basics, ok := captured["option_basic"].([]interface{})
	if !ok || len(basics) != 2 {
		t.Fatalf("期望 option_basic 有 2 项，得到: %v", captured["option_basic"])
	}
}

func TestGetOptionKlineMultiIdentifier(t *testing.T) {
	var captured map[string]interface{}
	server := captureServer(t, &captured)
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	_, err := qc.GetOptionKline([]string{"AAPL 240119C00150000", "AAPL 240119P00140000"}, "day", -1, -1)
	if err != nil {
		t.Fatalf("GetOptionKline 失败: %v", err)
	}

	queries, ok := captured["option_query"].([]interface{})
	if !ok || len(queries) != 2 {
		t.Fatalf("期望 option_query 有 2 项，得到: %v", captured["option_query"])
	}
}

// === Timezone unit tests ===

// TestParseOptionExpiryTimezone verifies that parseOptionExpiry uses the correct timezone.
// 2024-01-19 00:00:00 UTC        = 1705622400000 ms
// 2024-01-19 00:00:00 America/NY = 1705640400000 ms  (UTC-5 in Jan)
// 2024-01-19 00:00:00 Asia/HK    = 1705593600000 ms  (UTC+8)
func TestParseOptionExpiryTimezone(t *testing.T) {
	tests := []struct {
		name   string
		expiry string
		tz     string
		wantMs int64
	}{
		{"us_ny", "2024-01-19", "America/New_York", 1705640400000},
		{"hk", "2024-01-19", "Asia/Hong_Kong", 1705593600000},
		{"utc", "2024-01-19", "UTC", 1705622400000},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseOptionExpiry(tc.expiry, tc.tz)
			if err != nil {
				t.Fatalf("parseOptionExpiry(%q, %q) error: %v", tc.expiry, tc.tz, err)
			}
			if got != tc.wantMs {
				t.Errorf("parseOptionExpiry(%q, %q) = %d, want %d", tc.expiry, tc.tz, got, tc.wantMs)
			}
		})
	}
}

// TestResolveOptionTimezone verifies timezone inference from symbol.
func TestResolveOptionTimezone(t *testing.T) {
	if got := resolveOptionTimezone("", "AAPL"); got != "America/New_York" {
		t.Errorf("resolveOptionTimezone(\"\", \"AAPL\") = %q, want America/New_York", got)
	}
	if got := resolveOptionTimezone("", "00700.HK"); got != "Asia/Hong_Kong" {
		t.Errorf("resolveOptionTimezone(\"\", \"00700.HK\") = %q, want Asia/Hong_Kong", got)
	}
	if got := resolveOptionTimezone("UTC", "AAPL"); got != "UTC" {
		t.Errorf("resolveOptionTimezone(\"UTC\", \"AAPL\") = %q, want UTC (explicit override)", got)
	}
}

// TestGetOptionChainExpiryTimezone verifies the correct expiry timestamp is sent to the server
// for a US option (AAPL). America/New_York midnight of 2024-01-19 = 1705640400000 ms.
func TestGetOptionChainExpiryTimezone(t *testing.T) {
	var captured map[string]interface{}
	server := captureServer(t, &captured)
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	qc.GetOptionChain([][2]string{{"AAPL", "2024-01-19"}})

	basics, ok := captured["option_basic"].([]interface{})
	if !ok || len(basics) != 1 {
		t.Fatalf("期望 option_basic 有 1 项，得到: %v", captured["option_basic"])
	}
	item, ok := basics[0].(map[string]interface{})
	if !ok {
		t.Fatalf("option_basic[0] 不是 map: %T", basics[0])
	}
	// JSON numbers decode as float64
	expiryFloat, ok := item["expiry"].(float64)
	if !ok {
		t.Fatalf("option_basic[0].expiry 不是 float64: %T = %v", item["expiry"], item["expiry"])
	}
	const wantExpiry = float64(1705640400000) // America/New_York midnight
	if expiryFloat != wantExpiry {
		t.Errorf("option_basic[0].expiry = %.0f, want %.0f (America/New_York midnight for 2024-01-19)", expiryFloat, wantExpiry)
	}
}

// TestGetOptionKlineExpiryTimezone verifies the correct expiry timestamp is sent to the server
// for a US option (AAPL). America/New_York midnight of 2024-01-19 = 1705640400000 ms.
func TestGetOptionKlineExpiryTimezone(t *testing.T) {
	var captured map[string]interface{}
	server := captureServer(t, &captured)
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	qc.GetOptionKline([]string{"AAPL 240119C00150000"}, "day", -1, -1)

	queries, ok := captured["option_query"].([]interface{})
	if !ok || len(queries) != 1 {
		t.Fatalf("期望 option_query 有 1 项，得到: %v", captured["option_query"])
	}
	item, ok := queries[0].(map[string]interface{})
	if !ok {
		t.Fatalf("option_query[0] 不是 map: %T", queries[0])
	}
	expiryFloat, ok := item["expiry"].(float64)
	if !ok {
		t.Fatalf("option_query[0].expiry 不是 float64: %T = %v", item["expiry"], item["expiry"])
	}
	const wantExpiry = float64(1705640400000) // America/New_York midnight
	if expiryFloat != wantExpiry {
		t.Errorf("option_query[0].expiry = %.0f, want %.0f (America/New_York midnight for 2024-01-19)", expiryFloat, wantExpiry)
	}
}

func TestNewQuoteClientFromConfig(t *testing.T) {
	server := mockServer(t, "market_state", []map[string]interface{}{{"market": "US"}})
	defer server.Close()

	cfg := &config.ClientConfig{
		TigerID:        "test_tiger_id",
		PrivateKey:     cachedTestPrivateKey,
		Account:        "test_account",
		Language:       "zh_CN",
		Timeout:        5 * time.Second,
		QuoteServerURL: server.URL,
	}
	qc := NewQuoteClientFromConfig(cfg)
	if qc == nil {
		t.Fatal("NewQuoteClientFromConfig 返回 nil")
	}
	// 验证能正常发请求（走 QuoteServerURL）
	data, err := qc.GetMarketState("US")
	if err != nil {
		t.Fatalf("NewQuoteClientFromConfig 创建的客户端调用失败: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("期望有数据，但返回空")
	}
}

func TestGetOptionChainWithFilter(t *testing.T) {
	var captured map[string]interface{}
	server := captureServer(t, &captured)
	defer server.Close()

	trueVal := true
	qc := newTestQuoteClient(server.URL)
	_, err := qc.GetOptionChainByReq(model.OptionChainRequest{
		OptionBasic:      []model.OptionQueryItem{{Symbol: "AAPL", Expiry: int64(1705640400000)}},
		ReturnGreekValue: &trueVal,
		OptionFilter: &model.OptionChainFilter{
			InTheMoney:        &trueVal,
			ImpliedVolatility: model.NewRangeFloat64(0.1, 1.0),
			Greeks: &model.OptionChainFilterGreeks{
				Delta: model.NewRangeFloat64(0.0, 0.6),
			},
		},
	})
	if err != nil {
		t.Fatalf("GetOptionChainByReq 失败: %v", err)
	}

	if captured["return_greek_value"] != true {
		t.Errorf("期望 return_greek_value=true，得到: %v", captured["return_greek_value"])
	}
	filter, ok := captured["option_filter"].(map[string]interface{})
	if !ok {
		t.Fatalf("期望 option_filter 为 map，得到: %T", captured["option_filter"])
	}
	if filter["in_the_money"] != true {
		t.Errorf("期望 in_the_money=true，得到: %v", filter["in_the_money"])
	}
	iv, ok := filter["implied_volatility"].(map[string]interface{})
	if !ok {
		t.Fatalf("期望 implied_volatility 为 {min,max} 对象，得到: %T", filter["implied_volatility"])
	}
	if iv["min"] == nil || iv["max"] == nil {
		t.Errorf("implied_volatility 缺少 min/max，得到: %v", iv)
	}
}

func TestOptionChainFilterRangeJSON(t *testing.T) {
	f := model.OptionChainFilter{
		ImpliedVolatility: model.NewRangeFloat64(0.2, 0.8),
		OpenInterest:      model.NewRangeInt(100, 5000),
	}
	b, _ := json.Marshal(f)
	s := string(b)
	if !strings.Contains(s, `"implied_volatility":{"min":0.2,"max":0.8}`) {
		t.Errorf("implied_volatility 序列化错误: %s", s)
	}
	if !strings.Contains(s, `"open_interest":{"min":100,"max":5000}`) {
		t.Errorf("open_interest 序列化错误: %s", s)
	}
}

func TestOptionAnalysisPerSymbol(t *testing.T) {
	var captured map[string]interface{}
	server := captureServer(t, &captured)
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	_, err := qc.GetOptionAnalysis(model.OptionAnalysisRequest{
		Symbols: []model.OptionAnalysisSymbol{
			{Symbol: "AAPL", Period: "26week", RequireVolatilityList: func() *bool { v := true; return &v }()},
			{Symbol: "TSLA", Period: "month"},
		},
		Market: "US",
	})
	if err != nil {
		t.Fatalf("GetOptionAnalysis 失败: %v", err)
	}

	syms, ok := captured["symbols"].([]interface{})
	if !ok || len(syms) != 2 {
		t.Fatalf("期望 symbols 有 2 项，得到: %v", captured["symbols"])
	}
	first, ok := syms[0].(map[string]interface{})
	if !ok {
		t.Fatalf("期望每个 symbol 是 {symbol, period, ...} 对象，得到: %T", syms[0])
	}
	if first["symbol"] != "AAPL" || first["period"] != "26week" {
		t.Errorf("第一个 symbol 对象不符合预期: %v", first)
	}
}

func TestGetOptionKlineWithLimit(t *testing.T) {
	var captured map[string]interface{}
	server := captureServer(t, &captured)
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	qc.GetOptionKlineWithOpts([]string{"AAPL 240119C00150000"}, "day", -1, -1, 50, "asc")

	queries, ok := captured["option_query"].([]interface{})
	if !ok || len(queries) != 1 {
		t.Fatalf("期望 option_query 有 1 项，得到: %v", captured["option_query"])
	}
	q := queries[0].(map[string]interface{})
	if q["limit"] != float64(50) {
		t.Errorf("期望 limit=50，得到: %v", q["limit"])
	}
	if q["sort_dir"] != "asc" {
		t.Errorf("期望 sort_dir=asc，得到: %v", q["sort_dir"])
	}
}

func TestOrderRequestNewFields(t *testing.T) {
	req := model.OrderRequest{
		Symbol:             "AAPL",
		AllocAccounts:      []string{"ACC1", "ACC2"},
		AllocShares:        []float64{50.0, 50.0},
		ProfitTakerOrderId: 12345,
		StopLossOrderId:    67890,
		ContractLegs: []model.ContractLegRequest{
			{Symbol: "AAPL", Right: "CALL", Action: "BUY", Ratio: func() *int { v := 1; return &v }()},
		},
	}
	b, _ := json.Marshal(req)
	s := string(b)
	if !strings.Contains(s, `"alloc_accounts":["ACC1","ACC2"]`) {
		t.Errorf("alloc_accounts 序列化错误: %s", s)
	}
	if !strings.Contains(s, `"alloc_shares":[50,50]`) {
		t.Errorf("alloc_shares 序列化错误: %s", s)
	}
	if !strings.Contains(s, `"profit_taker_orderId":12345`) {
		t.Errorf("profit_taker_orderId wire name 错误: %s", s)
	}
	if !strings.Contains(s, `"stop_loss_orderId":67890`) {
		t.Errorf("stop_loss_orderId wire name 错误: %s", s)
	}
	if !strings.Contains(s, `"contract_legs"`) {
		t.Errorf("contract_legs 缺失: %s", s)
	}
}

func pagingMockServer(t *testing.T, pages []interface{}) *httptest.Server {
	t.Helper()
	idx := make(chan int, 1)
	idx <- 0
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := <-idx
		var data interface{}
		if n < len(pages) {
			data = pages[n]
		} else {
			data = []interface{}{}
		}
		idx <- n + 1
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

// ============================================================
// Batch 3: 股票基础查询 + 时间序列
// ============================================================

func TestGetSymbols(t *testing.T) {
	server := mockServer(t, "all_symbols", []string{"AAPL", "TSLA"})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetSymbols(model.SymbolsRequest{Market: "US"})
	if err != nil {
		t.Fatalf("GetSymbols 失败: %v", err)
	}
	if len(data) != 2 {
		t.Fatalf("期望 2 个 symbol，实际 %d", len(data))
	}
	if data[0] != "AAPL" {
		t.Errorf("data[0] = %q, 期望 AAPL", data[0])
	}
}

func TestGetSymbolNames(t *testing.T) {
	server := mockServer(t, "all_symbol_names", []map[string]interface{}{
		{"symbol": "AAPL", "name": "Apple Inc."},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetSymbolNames(model.SymbolsRequest{Market: "US"})
	if err != nil {
		t.Fatalf("GetSymbolNames 失败: %v", err)
	}
	if len(data) != 1 {
		t.Fatalf("期望 1 条，实际 %d", len(data))
	}
	if data[0].Symbol != "AAPL" {
		t.Errorf("Symbol = %q, 期望 AAPL", data[0].Symbol)
	}
}

func TestGetTradeMetas(t *testing.T) {
	server := mockServer(t, "quote_stock_trade", []map[string]interface{}{
		{"symbol": "AAPL"},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetTradeMetas(model.TradeMetasRequest{Symbols: []string{"AAPL"}})
	if err != nil {
		t.Fatalf("GetTradeMetas 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetTradeMetas 返回 nil")
	}
}

func TestGetStockDetails(t *testing.T) {
	server := mockServer(t, "stock_detail", map[string]interface{}{
		"items": []map[string]interface{}{
			{"symbol": "AAPL"},
		},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetStockDetails(model.StockDetailsRequest{Symbols: []string{"AAPL"}})
	if err != nil {
		t.Fatalf("GetStockDetails 失败: %v", err)
	}
	if len(data) != 1 {
		t.Fatalf("期望 1 条，实际 %d", len(data))
	}
}

// TestGetStockDetails_emptyItems 验证 callIntoItems 在 items 字段缺失时返回空切片无错误
func TestGetStockDetails_emptyItems(t *testing.T) {
	server := mockServer(t, "stock_detail", map[string]interface{}{})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetStockDetails(model.StockDetailsRequest{Symbols: []string{"AAPL"}})
	if err != nil {
		t.Fatalf("GetStockDetails 空响应失败: %v", err)
	}
	if data != nil {
		t.Fatalf("期望 nil 切片，实际 %v", data)
	}
}

// TestGetStockDetails_nullItems 验证 callIntoItems 在 items=null 时返回空切片无错误
func TestGetStockDetails_nullItems(t *testing.T) {
	server := mockServer(t, "stock_detail", map[string]interface{}{"items": nil})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetStockDetails(model.StockDetailsRequest{Symbols: []string{"AAPL"}})
	if err != nil {
		t.Fatalf("GetStockDetails null items 失败: %v", err)
	}
	if data != nil {
		t.Fatalf("期望 nil 切片，实际 %v", data)
	}
}

func TestGetStockDelayBriefs(t *testing.T) {
	server := mockServer(t, "quote_delay", []map[string]interface{}{
		{"symbol": "AAPL"},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetStockDelayBriefs(model.StockDelayBriefsRequest{Symbols: []string{"AAPL"}})
	if err != nil {
		t.Fatalf("GetStockDelayBriefs 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetStockDelayBriefs 返回 nil")
	}
}

// TestGetKlineByPage_pagination 验证客户端分页循环：第一页满页 → 继续翻页；第二页空 → 退出
func TestGetKlineByPage_pagination(t *testing.T) {
	// 第一页返回 PageSize 条（满足 len >= PageSize 继续循环），第二页返回空 items 触发 break
	page1 := []map[string]interface{}{
		{"symbol": "AAPL", "items": []map[string]interface{}{
			{"time": 100, "open": 1.0, "close": 2.0, "high": 3.0, "low": 0.5, "volume": 1000},
			{"time": 90, "open": 1.0, "close": 2.0, "high": 3.0, "low": 0.5, "volume": 1000},
		}},
	}
	page2 := []map[string]interface{}{
		{"symbol": "AAPL", "items": []map[string]interface{}{}},
	}
	server := pagingMockServer(t, []interface{}{page1, page2})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetKlineByPage(model.KlineByPageRequest{
		Symbol:    "AAPL",
		Period:    "day",
		PageSize:  2,
		TotalSize: 10,
	})
	if err != nil {
		t.Fatalf("GetKlineByPage 失败: %v", err)
	}
	if len(data) != 2 {
		t.Fatalf("期望累积 2 条，实际 %d", len(data))
	}
}

// TestGetKlineByPage_smallPage 验证单页不足 PageSize 时直接退出循环
func TestGetKlineByPage_smallPage(t *testing.T) {
	page1 := []map[string]interface{}{
		{"symbol": "AAPL", "items": []map[string]interface{}{
			{"time": 100, "open": 1.0, "close": 2.0, "high": 3.0, "low": 0.5, "volume": 1000},
		}},
	}
	server := pagingMockServer(t, []interface{}{page1})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetKlineByPage(model.KlineByPageRequest{
		Symbol:    "AAPL",
		Period:    "day",
		PageSize:  200,
		TotalSize: 1000,
	})
	if err != nil {
		t.Fatalf("GetKlineByPage 失败: %v", err)
	}
	if len(data) != 1 {
		t.Fatalf("期望 1 条，实际 %d", len(data))
	}
}

func TestGetBarsByPage(t *testing.T) {
	page1 := []map[string]interface{}{
		{"symbol": "AAPL", "items": []map[string]interface{}{
			{"time": 100, "open": 1.0, "close": 2.0, "high": 3.0, "low": 0.5, "volume": 1000},
		}},
	}
	server := pagingMockServer(t, []interface{}{page1})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetBarsByPage(model.KlineByPageRequest{
		Symbol:    "AAPL",
		Period:    "day",
		PageSize:  200,
		TotalSize: 1000,
	})
	if err != nil {
		t.Fatalf("GetBarsByPage 失败: %v", err)
	}
	if len(data) != 1 {
		t.Fatalf("期望 1 条，实际 %d", len(data))
	}
}

func TestGetTimelineHistory(t *testing.T) {
	server := mockServer(t, "history_timeline", []map[string]interface{}{
		{"symbol": "AAPL"},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetTimelineHistory(model.TimelineHistoryRequest{
		Symbols: []string{"AAPL"},
		Date:    "2024-01-19",
	})
	if err != nil {
		t.Fatalf("GetTimelineHistory 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetTimelineHistory 返回 nil")
	}
}

func TestGetTradeRank(t *testing.T) {
	server := mockServer(t, "trade_rank", []map[string]interface{}{
		{"symbol": "AAPL"},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetTradeRank(model.TradeRankRequest{Market: "US"})
	if err != nil {
		t.Fatalf("GetTradeRank 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetTradeRank 返回 nil")
	}
}

func TestGetShortInterest(t *testing.T) {
	server := mockServer(t, "quote_shortable_stocks", []map[string]interface{}{
		{"symbol": "AAPL"},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetShortInterest(model.ShortInterestRequest{Symbols: []string{"AAPL"}})
	if err != nil {
		t.Fatalf("GetShortInterest 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetShortInterest 返回 nil")
	}
}

func TestGetStockBroker(t *testing.T) {
	server := mockServer(t, "stock_broker", map[string]interface{}{
		"symbol": "AAPL",
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetStockBroker(model.StockBrokerRequest{Symbol: "AAPL"})
	if err != nil {
		t.Fatalf("GetStockBroker 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetStockBroker 返回 nil")
	}
}

func TestGetStockFundamental(t *testing.T) {
	server := mockServer(t, "stock_fundamental", map[string]interface{}{
		"AAPL": map[string]interface{}{"pe": 25.0},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetStockFundamental(model.StockFundamentalRequest{
		Symbols: []string{"AAPL"},
		Market:  "US",
	})
	if err != nil {
		t.Fatalf("GetStockFundamental 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetStockFundamental 返回 nil")
	}
}

func TestGetStockIndustry(t *testing.T) {
	server := mockServer(t, "stock_industry", []map[string]interface{}{
		{"symbol": "AAPL"},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetStockIndustry(model.StockIndustryRequest{Symbol: "AAPL", Market: "US"})
	if err != nil {
		t.Fatalf("GetStockIndustry 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetStockIndustry 返回 nil")
	}
}

func TestGetQuotePermission(t *testing.T) {
	server := mockServer(t, "get_quote_permission", []map[string]interface{}{
		{"name": "usStockQuote"},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetQuotePermission(model.QuotePermissionRequest{
		BeginDate: "2024-01-01",
		EndDate:   "2024-12-31",
	})
	if err != nil {
		t.Fatalf("GetQuotePermission 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetQuotePermission 返回 nil")
	}
}

func TestGetKlineQuota(t *testing.T) {
	server := mockServer(t, "kline_quota", []map[string]interface{}{
		{"name": "usKline"},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetKlineQuota(model.KlineQuotaRequest{WithDetails: true})
	if err != nil {
		t.Fatalf("GetKlineQuota 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetKlineQuota 返回 nil")
	}
}

// ============================================================
// Batch 4: 期权/期货扩展
// ============================================================

func TestGetOptionTradeTicks(t *testing.T) {
	server := mockServer(t, "option_trade_tick", []map[string]interface{}{
		{"symbol": "AAPL"},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetOptionTradeTicks(model.OptionTradeTicksRequest{
		Contracts: []model.OptionQueryItem{{Symbol: "AAPL", Expiry: 1705640400000, Right: "CALL", Strike: "150"}},
	})
	if err != nil {
		t.Fatalf("GetOptionTradeTicks 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetOptionTradeTicks 返回 nil")
	}
}

// TestOptionTradeTicksRequest_WireShape locks in the wire contract: the
// gateway requires the biz_content to be a top-level JSON array, not an
// object with a "contracts" key. See Java BatchApiModel and Python
// MultipleContractParams for reference.
func TestOptionTradeTicksRequest_WireShape(t *testing.T) {
	req := model.OptionTradeTicksRequest{
		Contracts: []model.OptionQueryItem{
			{Symbol: "AAPL", Expiry: 1705640400000, Right: "CALL", Strike: "150"},
		},
		Lang: "en_US", // must be ignored
	}
	raw, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got := string(raw)
	if !strings.HasPrefix(got, "[") || !strings.HasSuffix(got, "]") {
		t.Fatalf("expected top-level JSON array, got: %s", got)
	}
	if strings.Contains(got, `"contracts"`) {
		t.Errorf("wire must not contain \"contracts\" key, got: %s", got)
	}
	if strings.Contains(got, `"lang"`) {
		t.Errorf("wire must not contain \"lang\" key, got: %s", got)
	}
	if !strings.Contains(got, `"symbol":"AAPL"`) {
		t.Errorf("expected symbol AAPL in wire, got: %s", got)
	}
}

// TestOptionTradeTicksRequest_Empty ensures nil contracts serialize as [].
func TestOptionTradeTicksRequest_Empty(t *testing.T) {
	raw, err := json.Marshal(model.OptionTradeTicksRequest{})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(raw) != "[]" {
		t.Errorf("expected [], got %s", string(raw))
	}
}

func TestGetOptionTimeline(t *testing.T) {
	server := mockServer(t, "option_timeline", []map[string]interface{}{
		{"symbol": "AAPL"},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetOptionTimeline(model.OptionTimelineRequest{
		OptionQuery: []model.OptionQueryItem{{Symbol: "AAPL", Expiry: 1705640400000, Right: "CALL", Strike: "150"}},
		Market:      "US",
	})
	if err != nil {
		t.Fatalf("GetOptionTimeline 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetOptionTimeline 返回 nil")
	}
}

func TestGetOptionDepth(t *testing.T) {
	server := mockServer(t, "option_depth", []map[string]interface{}{
		{"symbol": "AAPL"},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetOptionDepth(model.OptionDepthRequest{
		OptionBasic: []model.OptionQueryItem{{Symbol: "AAPL", Expiry: 1705640400000, Right: "CALL", Strike: "150"}},
		Market:      "US",
	})
	if err != nil {
		t.Fatalf("GetOptionDepth 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetOptionDepth 返回 nil")
	}
}

func TestGetOptionSymbols(t *testing.T) {
	server := mockServer(t, "option_symbol", []map[string]interface{}{
		{"symbol": "AAPL240119C00150000"},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetOptionSymbols(model.OptionSymbolsRequest{Market: "US"})
	if err != nil {
		t.Fatalf("GetOptionSymbols 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetOptionSymbols 返回 nil")
	}
}

func TestGetOptionAnalysis_mock(t *testing.T) {
	server := mockServer(t, "option_analysis", []map[string]interface{}{
		{"symbol": "AAPL"},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetOptionAnalysis(model.OptionAnalysisRequest{
		Symbols: []model.OptionAnalysisSymbol{{Symbol: "AAPL", Period: "26week"}},
		Market:  "US",
	})
	if err != nil {
		t.Fatalf("GetOptionAnalysis 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetOptionAnalysis 返回 nil")
	}
}

// TestGetFutureContract_list 验证 callIntoListOrObject 在服务端返回列表时正常解析
func TestGetFutureContract_list(t *testing.T) {
	server := mockServer(t, "future_contract_by_contract_code", []map[string]interface{}{
		{"symbol": "ES240321"},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetFutureContract(model.FutureContractSingleRequest{ContractCode: "ES240321"})
	if err != nil {
		t.Fatalf("GetFutureContract 列表响应失败: %v", err)
	}
	if len(data) != 1 {
		t.Fatalf("期望 1 条，实际 %d", len(data))
	}
}

// TestGetFutureContract_single 验证 callIntoListOrObject 在服务端返回单对象时回退包装
func TestGetFutureContract_single(t *testing.T) {
	server := mockServer(t, "future_contract_by_contract_code", map[string]interface{}{
		"symbol": "ES240321",
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetFutureContract(model.FutureContractSingleRequest{ContractCode: "ES240321"})
	if err != nil {
		t.Fatalf("GetFutureContract 单对象响应失败: %v", err)
	}
	if len(data) != 1 {
		t.Fatalf("期望回退包装后 1 条，实际 %d", len(data))
	}
}

func TestGetAllFutureContracts(t *testing.T) {
	server := mockServer(t, "future_contracts", []map[string]interface{}{
		{"symbol": "ES240321"},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetAllFutureContracts(model.AllFutureContractsRequest{Exchange: "CME"})
	if err != nil {
		t.Fatalf("GetAllFutureContracts 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetAllFutureContracts 返回 nil")
	}
}

func TestGetCurrentFutureContract(t *testing.T) {
	server := mockServer(t, "future_current_contract", map[string]interface{}{
		"symbol": "ES240321",
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetCurrentFutureContract(model.FutureContractSingleRequest{ContractCode: "ES"})
	if err != nil {
		t.Fatalf("GetCurrentFutureContract 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetCurrentFutureContract 返回 nil")
	}
}

// TestGetFutureContinuousContracts_list 验证 callIntoListOrObject 列表响应
func TestGetFutureContinuousContracts_list(t *testing.T) {
	server := mockServer(t, "future_continuous_contracts", []map[string]interface{}{
		{"symbol": "ES240321"},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetFutureContinuousContracts(model.FutureContinuousContractsRequest{Type: "main"})
	if err != nil {
		t.Fatalf("GetFutureContinuousContracts 列表响应失败: %v", err)
	}
	if len(data) != 1 {
		t.Fatalf("期望 1 条，实际 %d", len(data))
	}
}

// TestGetFutureContinuousContracts_single 验证 callIntoListOrObject 单对象回退
func TestGetFutureContinuousContracts_single(t *testing.T) {
	server := mockServer(t, "future_continuous_contracts", map[string]interface{}{
		"symbol": "ES240321",
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetFutureContinuousContracts(model.FutureContinuousContractsRequest{Type: "main"})
	if err != nil {
		t.Fatalf("GetFutureContinuousContracts 单对象响应失败: %v", err)
	}
	if len(data) != 1 {
		t.Fatalf("期望回退包装后 1 条，实际 %d", len(data))
	}
}

func TestGetFutureHistoryMainContract(t *testing.T) {
	server := mockServer(t, "future_main_contract", []map[string]interface{}{
		{"contractCode": "ES240321"},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetFutureHistoryMainContract(model.FutureHistoryMainContractRequest{
		ContractCodes: []string{"ES"},
	})
	if err != nil {
		t.Fatalf("GetFutureHistoryMainContract 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetFutureHistoryMainContract 返回 nil")
	}
}

// TestGetFutureKlineByPage_pagination 验证期货 K 线客户端分页循环
func TestGetFutureKlineByPage_pagination(t *testing.T) {
	page1 := []map[string]interface{}{
		{"items": []map[string]interface{}{
			{"time": 100, "open": 1.0, "close": 2.0, "high": 3.0, "low": 0.5, "volume": 1000},
			{"time": 90, "open": 1.0, "close": 2.0, "high": 3.0, "low": 0.5, "volume": 1000},
		}},
	}
	page2 := []map[string]interface{}{
		{"items": []map[string]interface{}{}},
	}
	server := pagingMockServer(t, []interface{}{page1, page2})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetFutureKlineByPage(model.FutureKlineByPageRequest{
		ContractCode: "ES240321",
		Period:       "day",
		PageSize:     2,
		TotalSize:    10,
	})
	if err != nil {
		t.Fatalf("GetFutureKlineByPage 失败: %v", err)
	}
	if len(data) != 2 {
		t.Fatalf("期望累积 2 条，实际 %d", len(data))
	}
}

func TestGetFutureDepth(t *testing.T) {
	server := mockServer(t, "future_depth", []map[string]interface{}{
		{"contractCode": "ES240321"},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetFutureDepth(model.FutureDepthRequest{ContractCodes: []string{"ES240321"}})
	if err != nil {
		t.Fatalf("GetFutureDepth 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetFutureDepth 返回 nil")
	}
}

func TestGetFutureTradingTimes(t *testing.T) {
	server := mockServer(t, "future_trading_date", map[string]interface{}{
		"timeSection": "09:30-16:00",
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetFutureTradingTimes(model.FutureTradingTimesRequest{
		ContractCode: "ES240321",
		TradingDate:  "2024-01-19",
	})
	if err != nil {
		t.Fatalf("GetFutureTradingTimes 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetFutureTradingTimes 返回 nil")
	}
}

// TestGetFutureTradeTicks_wrap 验证 future_tick 解包 items 并回填 contractCode
func TestGetFutureTradeTicks_wrap(t *testing.T) {
	server := mockServer(t, "future_tick", map[string]interface{}{
		"contractCode": "ES240321",
		"items": []map[string]interface{}{
			{"index": 0, "time": 100, "price": 100.0, "volume": 10, "direction": "buy"},
		},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetFutureTradeTicks(model.FutureTradeTicksRequest{ContractCode: "ES240321"})
	if err != nil {
		t.Fatalf("GetFutureTradeTicks 失败: %v", err)
	}
	if len(data) != 1 {
		t.Fatalf("期望 1 条，实际 %d", len(data))
	}
	// items 原本没有 contractCode，应被回填为 wrap.ContractCode
	if data[0].ContractCode != "ES240321" {
		t.Errorf("ContractCode = %q, 期望 ES240321 (回填)", data[0].ContractCode)
	}
}

// TestGetFutureTradeTicks_defaultEndIndex 验证 EndIndex=0 时默认置为 30
func TestGetFutureTradeTicks_defaultEndIndex(t *testing.T) {
	var captured map[string]interface{}
	server := captureServer(t, &captured)
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	qc.GetFutureTradeTicks(model.FutureTradeTicksRequest{ContractCode: "ES240321"})

	if captured["end_index"] != float64(30) {
		t.Errorf("期望 end_index=30，得到: %v", captured["end_index"])
	}
}

// ============================================================
// Batch 5: 基金/窝轮/行业/公司行动/财务/日历
// ============================================================

func TestGetFundSymbols(t *testing.T) {
	server := mockServer(t, "fund_all_symbols", []string{"000001", "000002"})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetFundSymbols(model.FundSymbolsRequest{})
	if err != nil {
		t.Fatalf("GetFundSymbols 失败: %v", err)
	}
	if len(data) != 2 {
		t.Fatalf("期望 2 个 symbol，实际 %d", len(data))
	}
}

func TestGetFundContracts(t *testing.T) {
	server := mockServer(t, "fund_contracts", []map[string]interface{}{
		{"symbol": "000001"},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetFundContracts(model.FundContractsRequest{Symbols: []string{"000001"}})
	if err != nil {
		t.Fatalf("GetFundContracts 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetFundContracts 返回 nil")
	}
}

func TestGetFundQuote(t *testing.T) {
	server := mockServer(t, "fund_quote", []map[string]interface{}{
		{"symbol": "000001"},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetFundQuote(model.FundQuoteRequest{Symbols: []string{"000001"}})
	if err != nil {
		t.Fatalf("GetFundQuote 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetFundQuote 返回 nil")
	}
}

func TestGetFundHistoryQuote(t *testing.T) {
	server := mockServer(t, "fund_history_quote", []map[string]interface{}{
		{"symbol": "000001"},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetFundHistoryQuote(model.FundHistoryQuoteRequest{
		Symbols:   []string{"000001"},
		BeginTime: 1700000000000,
		EndTime:   1700086400000,
	})
	if err != nil {
		t.Fatalf("GetFundHistoryQuote 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetFundHistoryQuote 返回 nil")
	}
}

func TestGetWarrantBriefs(t *testing.T) {
	server := mockServer(t, "warrant_briefs", []map[string]interface{}{
		{"symbol": "28888.HK"},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetWarrantBriefs(model.WarrantBriefsRequest{Symbols: []string{"28888.HK"}})
	if err != nil {
		t.Fatalf("GetWarrantBriefs 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetWarrantBriefs 返回 nil")
	}
}

func TestGetWarrantFilter(t *testing.T) {
	server := mockServer(t, "warrant_filter", map[string]interface{}{
		"symbol": "AAPL",
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetWarrantFilter(model.WarrantFilterRequest{Symbol: "AAPL"})
	if err != nil {
		t.Fatalf("GetWarrantFilter 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetWarrantFilter 返回 nil")
	}
}

func TestGetIndustryList(t *testing.T) {
	server := mockServer(t, "industry_list", []map[string]interface{}{
		{"id": "1", "nameEN": "Technology", "nameCN": "科技", "industryLevel": "GSECTOR"},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetIndustryList(model.IndustryListRequest{IndustryLevel: "GSECTOR"})
	if err != nil {
		t.Fatalf("GetIndustryList 失败: %v", err)
	}
	if len(data) != 1 {
		t.Fatalf("期望 1 条，实际 %d", len(data))
	}
	if data[0].ID != "1" {
		t.Errorf("IndustryItem.ID = %q, want 1", data[0].ID)
	}
	if data[0].NameEN != "Technology" {
		t.Errorf("IndustryItem.NameEN = %q, want Technology", data[0].NameEN)
	}
	if data[0].NameCN != "科技" {
		t.Errorf("IndustryItem.NameCN = %q, want 科技", data[0].NameCN)
	}
	if data[0].Level != "GSECTOR" {
		t.Errorf("IndustryItem.Level = %q, want GSECTOR", data[0].Level)
	}
	// Name compat field prefers NameEN.
	if data[0].Name != "Technology" {
		t.Errorf("IndustryItem.Name = %q, want Technology (hydrated from NameEN)", data[0].Name)
	}
}

// TestGetIndustryList_DefaultLevel verifies GetIndustryList defaults
// IndustryLevel to GGROUP when caller leaves it empty (matches Python SDK).
func TestGetIndustryList_DefaultLevel(t *testing.T) {
	var gotBiz string
	server := mockServerCapture(t, "industry_list", []map[string]interface{}{
		{"id": "1", "nameEN": "Semiconductors", "industryLevel": "GGROUP"},
	}, &gotBiz)
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	_, err := qc.GetIndustryList(model.IndustryListRequest{}) // no level provided
	if err != nil {
		t.Fatalf("GetIndustryList 失败: %v", err)
	}
	if !strings.Contains(gotBiz, `"industry_level":"GGROUP"`) {
		t.Errorf("expected industry_level defaulted to GGROUP in biz_content, got %s", gotBiz)
	}
}

// TestGetIndustryList_NameFallbackToCN verifies Name falls back to NameCN
// when NameEN is empty.
func TestGetIndustryList_NameFallbackToCN(t *testing.T) {
	server := mockServer(t, "industry_list", []map[string]interface{}{
		{"id": "2", "nameCN": "金融", "industryLevel": "GSECTOR"},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetIndustryList(model.IndustryListRequest{IndustryLevel: "GSECTOR"})
	if err != nil {
		t.Fatalf("GetIndustryList 失败: %v", err)
	}
	if len(data) != 1 || data[0].Name != "金融" {
		t.Errorf("expected Name to fall back to NameCN, got %+v", data)
	}
}

func TestGetIndustryStocks(t *testing.T) {
	server := mockServer(t, "industry_stock_list", []map[string]interface{}{
		{"symbol": "AAPL"},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetIndustryStocks(model.IndustryStocksRequest{IndustryID: "1", Market: "US"})
	if err != nil {
		t.Fatalf("GetIndustryStocks 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetIndustryStocks 返回 nil")
	}
}

func TestGetCorporateSplit(t *testing.T) {
	server := mockServer(t, "corporate_action", map[string]interface{}{
		"AAPL": []map[string]interface{}{
			{"symbol": "AAPL", "actionType": "split"},
		},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	begin := int64(1704067200000) // 2024-01-01 UTC
	end := int64(1735603200000)   // 2024-12-31 UTC
	data, err := qc.GetCorporateSplit(model.CorporateActionRequest{
		Symbols:   []string{"AAPL"},
		Market:    "US",
		BeginDate: &begin,
		EndDate:   &end,
	})
	if err != nil {
		t.Fatalf("GetCorporateSplit 失败: %v", err)
	}
	if len(data) != 1 {
		t.Fatalf("期望 1 条，实际 %d", len(data))
	}
}

func TestGetCorporateDividend(t *testing.T) {
	server := mockServer(t, "corporate_action", map[string]interface{}{
		"AAPL": []map[string]interface{}{
			{"symbol": "AAPL", "actionType": "dividend"},
		},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	begin := int64(1704067200000) // 2024-01-01 UTC
	end := int64(1735603200000)   // 2024-12-31 UTC
	data, err := qc.GetCorporateDividend(model.CorporateActionRequest{
		Symbols:   []string{"AAPL"},
		Market:    "US",
		BeginDate: &begin,
		EndDate:   &end,
	})
	if err != nil {
		t.Fatalf("GetCorporateDividend 失败: %v", err)
	}
	if len(data) != 1 {
		t.Fatalf("期望 1 条，实际 %d", len(data))
	}
}

func TestGetCorporateEarningsCalendar(t *testing.T) {
	server := mockServer(t, "corporate_action", map[string]interface{}{
		"AAPL": []map[string]interface{}{
			{"symbol": "AAPL", "actionType": "earning"},
		},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	begin := int64(1704067200000) // 2024-01-01 UTC
	end := int64(1735603200000)   // 2024-12-31 UTC
	data, err := qc.GetCorporateEarningsCalendar(model.CorporateActionRequest{
		Symbols:   []string{"AAPL"},
		Market:    "US",
		BeginDate: &begin,
		EndDate:   &end,
	})
	if err != nil {
		t.Fatalf("GetCorporateEarningsCalendar 失败: %v", err)
	}
	if len(data) != 1 {
		t.Fatalf("期望 1 条，实际 %d", len(data))
	}
}

func TestGetFinancialExchangeRate(t *testing.T) {
	server := mockServer(t, "financial_exchange_rate", []map[string]interface{}{
		{"currency": "USD/CNY"},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetFinancialExchangeRate(model.FinancialExchangeRateRequest{
		CurrencyList: []string{"USD/CNY"},
		BeginDate:    "20240101",
		EndDate:      "20240131",
	})
	if err != nil {
		t.Fatalf("GetFinancialExchangeRate 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetFinancialExchangeRate 返回 nil")
	}
}

func TestGetFinancialCurrency(t *testing.T) {
	server := mockServer(t, "financial_currency", []map[string]interface{}{
		{"symbol": "AAPL"},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetFinancialCurrency(model.FinancialCurrencyRequest{
		Symbols: []string{"AAPL"},
		Market:  "US",
	})
	if err != nil {
		t.Fatalf("GetFinancialCurrency 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetFinancialCurrency 返回 nil")
	}
}

func TestGetTradingCalendar(t *testing.T) {
	server := mockServer(t, "trading_calendar", []map[string]interface{}{
		{"date": "2024-01-19", "isTrading": true},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetTradingCalendar(model.TradingCalendarRequest{
		Market:    "US",
		BeginDate: "2024-01-01",
		EndDate:   "2024-01-31",
	})
	if err != nil {
		t.Fatalf("GetTradingCalendar 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetTradingCalendar 返回 nil")
	}
}

func TestGetMarketScannerTags(t *testing.T) {
	server := mockServer(t, "market_scanner_tags", []map[string]interface{}{
		{"market": "US"},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetMarketScannerTags(model.MarketScannerTagsRequest{Market: "US"})
	if err != nil {
		t.Fatalf("GetMarketScannerTags 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetMarketScannerTags 返回 nil")
	}
}

func TestGetQuoteOvernight(t *testing.T) {
	server := mockServer(t, "quote_overnight", []map[string]interface{}{
		{
			"symbol": "AAPL", "latestPrice": 224.5, "askPrice": 224.6, "askSize": 12,
			"bidPrice": 224.4, "bidSize": 8, "preClose": 220.0, "volume": 1000,
			"amount": 224500.0, "timestamp": 1720000000000, "tradingStatus": 2,
			"change": 4.5, "changeRate": 0.02045, "amplitude": 0.03,
		},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetQuoteOvernight(model.QuoteOvernightRequest{Symbols: []string{"AAPL"}})
	if err != nil {
		t.Fatalf("GetQuoteOvernight 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetQuoteOvernight 返回 nil")
	}
	if len(data) != 1 {
		t.Fatalf("len(data) = %d, want 1", len(data))
	}
	got := data[0]
	if got.Symbol != "AAPL" || got.LatestPrice != 224.5 || got.AskPrice != 224.6 || got.AskSize != 12 ||
		got.BidPrice != 224.4 || got.BidSize != 8 || got.PreClose != 220 || got.Volume != 1000 ||
		got.Amount != 224500 || got.Timestamp != 1720000000000 || got.TradingStatus != 2 ||
		got.Change != 4.5 || got.ChangeRate != 0.02045 || got.Amplitude != 0.03 {
		t.Errorf("decoded QuoteOvernight = %+v", got)
	}
}

// ============================================================
// 辅助函数错误路径测试
// ============================================================

// TestOptionContractFromIdentifier_errors 覆盖 optionContractFromIdentifier 各错误分支
func TestOptionContractFromIdentifier_errors(t *testing.T) {
	tz := "America/New_York"
	tests := []struct {
		name       string
		identifier string
		errSubstr  string
	}{
		{"invalid_format_no_space", "AAPL", "expected format"},
		{"too_short", "AAPL 24011", "too short"},
		{"invalid_right_char", "AAPL 240119X00150000", "invalid right character"},
		{"invalid_strike_digits", "AAPL 240119C0015A000", "invalid strike digits"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := optionContractFromIdentifier(tc.identifier, tz)
			if err == nil {
				t.Fatalf("期望错误，但返回 nil")
			}
			if !strings.Contains(err.Error(), tc.errSubstr) {
				t.Errorf("错误信息 %q 不包含 %q", err.Error(), tc.errSubstr)
			}
		})
	}
}

// TestOptionContractFromIdentifier_happy 验证正常解析的 strike 和 right
func TestOptionContractFromIdentifier_happy(t *testing.T) {
	c, err := optionContractFromIdentifier("AAPL 240119C00150000", "America/New_York")
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if c.Symbol != "AAPL" {
		t.Errorf("Symbol = %q", c.Symbol)
	}
	if c.Right != "CALL" {
		t.Errorf("Right = %q, 期望 CALL", c.Right)
	}
	if c.Strike != 150.0 {
		t.Errorf("Strike = %v, 期望 150.0", c.Strike)
	}
	c2, err := optionContractFromIdentifier("AAPL 240119P00140000", "America/New_York")
	if err != nil {
		t.Fatalf("解析 PUT 失败: %v", err)
	}
	if c2.Right != "PUT" {
		t.Errorf("Right = %q, 期望 PUT", c2.Right)
	}
	if c2.Strike != 140.0 {
		t.Errorf("Strike = %v, 期望 140.0", c2.Strike)
	}
}

// TestParseOptionExpiry_errors 覆盖 parseOptionExpiry 错误分支
func TestParseOptionExpiry_errors(t *testing.T) {
	_, err := parseOptionExpiry("2024-01-19", "Invalid/Zone")
	if err == nil {
		t.Fatal("期望无效时区错误，但返回 nil")
	}
	if !strings.Contains(err.Error(), "invalid timezone") {
		t.Errorf("错误信息 %q 不包含 'invalid timezone'", err.Error())
	}

	_, err = parseOptionExpiry("2024/01/19", "America/New_York")
	if err == nil {
		t.Fatal("期望无效日期格式错误，但返回 nil")
	}
	if !strings.Contains(err.Error(), "expected format") {
		t.Errorf("错误信息 %q 不包含 'expected format'", err.Error())
	}
}

// TestGetOptionKlineWithOpts_zeroTime 验证 beginTime=0/endTime=0 被转换为 -1
func TestGetOptionKlineWithOpts_zeroTime(t *testing.T) {
	var captured map[string]interface{}
	server := captureServer(t, &captured)
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	_, err := qc.GetOptionKlineWithOpts([]string{"AAPL 240119C00150000"}, "day", 0, 0, 0, "")
	if err != nil {
		t.Fatalf("GetOptionKlineWithOpts 失败: %v", err)
	}
	queries, ok := captured["option_query"].([]interface{})
	if !ok || len(queries) != 1 {
		t.Fatalf("期望 option_query 有 1 项，得到: %v", captured["option_query"])
	}
	q := queries[0].(map[string]interface{})
	if q["begin_time"] != float64(-1) {
		t.Errorf("期望 begin_time=-1，得到: %v", q["begin_time"])
	}
	if q["end_time"] != float64(-1) {
		t.Errorf("期望 end_time=-1，得到: %v", q["end_time"])
	}
}

// TestGetOptionQuote_invalidIdentifier 验证 GetOptionQuote 在非法 identifier 时返回错误
func TestGetOptionQuote_invalidIdentifier(t *testing.T) {
	server := mockServer(t, "option_brief", []map[string]interface{}{})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	_, err := qc.GetOptionQuote([]string{"INVALID"})
	if err == nil {
		t.Fatal("期望错误，但返回 nil")
	}
}

// TestGetOptionChain_invalidExpiry 验证 GetOptionChain 在非法 expiry 时返回错误
func TestGetOptionChain_invalidExpiry(t *testing.T) {
	server := mockServer(t, "option_chain", []map[string]interface{}{})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	_, err := qc.GetOptionChain([][2]string{{"AAPL", "2024/01/19"}})
	if err == nil {
		t.Fatal("期望错误，但返回 nil")
	}
}

// TestGetBrief_deprecated 验证 deprecated GetBrief 委托到 GetRealTimeQuote
func TestGetBrief_deprecated(t *testing.T) {
	server := mockServer(t, "quote_real_time", []map[string]interface{}{
		{"symbol": "AAPL"},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetBrief(model.BriefRequest{Symbols: []string{"AAPL"}})
	if err != nil {
		t.Fatalf("GetBrief 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetBrief 返回 nil")
	}
}

// TestGetBars_deprecated 验证 deprecated GetBars 委托到 GetKline
func TestGetBars_deprecated(t *testing.T) {
	server := mockServer(t, "kline", []map[string]interface{}{
		{"symbol": "AAPL"},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetBars(model.KlineRequest{Symbols: []string{"AAPL"}, Period: "day"})
	if err != nil {
		t.Fatalf("GetBars 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetBars 返回 nil")
	}
}

// TestGetOptionBrief_deprecated 验证 deprecated GetOptionBrief 委托到 GetOptionQuote
func TestGetOptionBrief_deprecated(t *testing.T) {
	server := mockServer(t, "option_brief", []map[string]interface{}{
		{"identifier": "AAPL 240119C00150000"},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetOptionBrief([]string{"AAPL 240119C00150000"})
	if err != nil {
		t.Fatalf("GetOptionBrief 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetOptionBrief 返回 nil")
	}
}
