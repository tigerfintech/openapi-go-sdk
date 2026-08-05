package quote

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tigerfintech/openapi-go-sdk/model"
)

// pagingMockServer 创建一个按顺序返回多页数据的 mock 服务器。
// 第 n 次请求返回 pages[n]，n 超出范围时返回空数组。
// 使用 channel 做计数，避免引入新的 sync 依赖。
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
		{"id": "1", "name": "Technology"},
	})
	defer server.Close()

	qc := newTestQuoteClient(server.URL)
	data, err := qc.GetIndustryList(model.IndustryListRequest{IndustryLevel: "level1"})
	if err != nil {
		t.Fatalf("GetIndustryList 失败: %v", err)
	}
	if data == nil {
		t.Fatal("GetIndustryList 返回 nil")
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
	data, err := qc.GetCorporateSplit(model.CorporateActionRequest{
		Symbols:   []string{"AAPL"},
		Market:    "US",
		BeginDate: "2024-01-01",
		EndDate:   "2024-12-31",
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
	data, err := qc.GetCorporateDividend(model.CorporateActionRequest{
		Symbols:   []string{"AAPL"},
		Market:    "US",
		BeginDate: "2024-01-01",
		EndDate:   "2024-12-31",
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
	data, err := qc.GetCorporateEarningsCalendar(model.CorporateActionRequest{
		Symbols:   []string{"AAPL"},
		Market:    "US",
		BeginDate: "2024-01-01",
		EndDate:   "2024-12-31",
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
		{"symbol": "AAPL"},
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
