package news

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tigerfintech/openapi-go-sdk/client"
	"github.com/tigerfintech/openapi-go-sdk/config"
	"github.com/tigerfintech/openapi-go-sdk/model"
	"github.com/tigerfintech/openapi-go-sdk/quote"
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

// stockNewsWireData 个股新闻接口的真实 wire 形状（camelCase 出参），
// 取自 Java SDK NewsApiLogicTest 同源数据，保证跨 SDK 行为一致。
func stockNewsWireData() map[string]interface{} {
	return map[string]interface{}{
		"items": []map[string]interface{}{
			{"id": "123456", "title": "Apple Q2", "summary": "Earnings beat",
				"pubTimestamp": 1729413600, "url": "https://example.com/news/123456"},
		},
		"pageCount": 1, "pageSize": 20, "totalPage": 10, "totalSize": 200,
	}
}

// highlightNewsWireData 要闻接口的真实 wire 形状（camelCase 出参）。
// highlight 走游标翻页，响应信封的 pageCount 恒为 1。
func highlightNewsWireData() map[string]interface{} {
	return map[string]interface{}{
		"items": []map[string]interface{}{
			{"id": "789", "title": "Fed cuts rates", "summary": "FOMC decision",
				"pubTimestamp": 1761429600, "url": "https://example.com/highlight/detail?id=789"},
		},
		"pageCount": 1, "pageSize": 20, "totalPage": 5, "totalSize": 100,
	}
}

// mockNewsServer 创建 mock HTTP 服务器，返回指定的 data，
// 并把收到的 biz_content 写入 *capturedBiz（可为 nil）。
func mockNewsServer(t *testing.T, data interface{}, capturedBiz *string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if capturedBiz != nil {
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if biz, ok := body["biz_content"].(string); ok {
				*capturedBiz = biz
			}
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

// mockNewsErrorServer 创建返回业务错误的 mock HTTP 服务器。
func mockNewsErrorServer(t *testing.T, code int, message string) *httptest.Server {
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

// newTestConfig 构造测试配置。quoteServerURL 可为空（不区分主/行情域名时）。
func newTestConfig(serverURL, quoteServerURL string) *config.ClientConfig {
	return &config.ClientConfig{
		TigerID:        "test_tiger_id",
		PrivateKey:     cachedTestPrivateKey,
		Account:        "test_account",
		Language:       "zh_CN",
		Timeout:        5 * time.Second,
		ServerURL:      serverURL,
		QuoteServerURL: quoteServerURL,
	}
}

// newTestNewsClient 创建指向 serverURL（主网关语义）的资讯客户端。
func newTestNewsClient(serverURL string) *NewsClient {
	return NewNewsClient(client.NewHttpClient(newTestConfig(serverURL, "")))
}

// decodeBiz 把捕获的 biz_content JSON 解码为 map，便于按 key 断言。
func decodeBiz(t *testing.T, raw string) map[string]interface{} {
	t.Helper()
	var biz map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &biz); err != nil {
		t.Fatalf("biz_content 解码失败: %v (raw=%s)", err, raw)
	}
	return biz
}

// assertBizKeys 断言 biz_content 的 key 集合与 want 完全一致
// （不发送未设置的 key），且每个 key 的值相等。
// 数值统一按 float64 比较（encoding/json 默认解码类型）。
func assertBizKeys(t *testing.T, raw string, want map[string]interface{}) {
	t.Helper()
	biz := decodeBiz(t, raw)
	if len(biz) != len(want) {
		t.Fatalf("biz_content key 集合不符: got=%v want=%v (raw=%s)", biz, want, raw)
	}
	for k, wantV := range want {
		gotV, ok := biz[k]
		if !ok {
			t.Fatalf("biz_content 缺少 key %q (raw=%s)", k, raw)
		}
		if gotV != wantV {
			t.Fatalf("biz_content[%q] = %v, want %v (raw=%s)", k, gotV, wantV, raw)
		}
	}
}

// === 响应解析（camelCase 出参） ===

func TestGetStockNews(t *testing.T) {
	server := mockNewsServer(t, stockNewsWireData(), nil)
	defer server.Close()

	nc := newTestNewsClient(server.URL)
	page, err := nc.GetStockNews(model.NewsStockListRequest{Symbol: "AAPL"})
	if err != nil {
		t.Fatalf("GetStockNews 失败: %v", err)
	}
	if page == nil {
		t.Fatal("GetStockNews 返回 nil")
	}
	if page.PageCount != 1 || page.PageSize != 20 || page.TotalPage != 10 || page.TotalSize != 200 {
		t.Fatalf("分页信封字段解码错误: %+v", page)
	}
	if len(page.Items) != 1 {
		t.Fatalf("items 数量 = %d, want 1", len(page.Items))
	}
	item := page.Items[0]
	if item.ID != "123456" || item.Title != "Apple Q2" || item.Summary != "Earnings beat" {
		t.Fatalf("item 文本字段解码错误: %+v", item)
	}
	if item.PubTimestamp != 1729413600 {
		t.Fatalf("item pubTimestamp = %d, want 1729413600", item.PubTimestamp)
	}
	if item.URL != "https://example.com/news/123456" {
		t.Fatalf("item url = %q", item.URL)
	}
}

func TestGetMarketNews(t *testing.T) {
	server := mockNewsServer(t, highlightNewsWireData(), nil)
	defer server.Close()

	nc := newTestNewsClient(server.URL)
	page, err := nc.GetMarketNews(model.NewsHighlightListRequest{})
	if err != nil {
		t.Fatalf("GetMarketNews 失败: %v", err)
	}
	if page == nil {
		t.Fatal("GetMarketNews 返回 nil")
	}
	if page.PageCount != 1 {
		t.Fatalf("highlight 接口 pageCount 恒为 1, got %d", page.PageCount)
	}
	if len(page.Items) != 1 {
		t.Fatalf("items 数量 = %d, want 1", len(page.Items))
	}
	item := page.Items[0]
	if item.ID != "789" || item.Title != "Fed cuts rates" {
		t.Fatalf("item 字段解码错误: %+v", item)
	}
	if item.PubTimestamp != 1761429600 {
		t.Fatalf("item pubTimestamp = %d, want 1761429600", item.PubTimestamp)
	}
}

// === 请求 wire 格式（snake_case 入参、可选字段缺省不发送） ===

func TestGetStockNewsWireFormat(t *testing.T) {
	cases := []struct {
		name string
		req  model.NewsStockListRequest
		want map[string]interface{}
	}{
		{
			name: "仅必填 symbol，可选字段缺省不发送",
			req:  model.NewsStockListRequest{Symbol: "AAPL"},
			want: map[string]interface{}{"symbol": "AAPL"},
		},
		{
			name: "全量参数",
			req:  model.NewsStockListRequest{Symbol: "AAPL", Lang: "zh_CN", PageCount: 2, PageSize: 50},
			want: map[string]interface{}{"symbol": "AAPL", "lang": "zh_CN", "page_count": float64(2), "page_size": float64(50)},
		},
		{
			name: "lang 合法值域",
			req:  model.NewsStockListRequest{Symbol: "TSLA", Lang: "en_US"},
			want: map[string]interface{}{"symbol": "TSLA", "lang": "en_US"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var captured string
			server := mockNewsServer(t, stockNewsWireData(), &captured)
			defer server.Close()

			nc := newTestNewsClient(server.URL)
			if _, err := nc.GetStockNews(tc.req); err != nil {
				t.Fatalf("GetStockNews 失败: %v", err)
			}
			assertBizKeys(t, captured, tc.want)
		})
	}
}

func TestGetMarketNewsWireFormat(t *testing.T) {
	cases := []struct {
		name string
		req  model.NewsHighlightListRequest
		want map[string]interface{}
	}{
		{
			name: "无参数，从最新开始",
			req:  model.NewsHighlightListRequest{},
			want: map[string]interface{}{},
		},
		{
			name: "仅 lang",
			req:  model.NewsHighlightListRequest{Lang: "zh_TW"},
			want: map[string]interface{}{"lang": "zh_TW"},
		},
		{
			name: "lang + last_id 游标",
			req:  model.NewsHighlightListRequest{Lang: "en_US", LastID: "789"},
			want: map[string]interface{}{"lang": "en_US", "last_id": "789"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var captured string
			server := mockNewsServer(t, highlightNewsWireData(), &captured)
			defer server.Close()

			nc := newTestNewsClient(server.URL)
			if _, err := nc.GetMarketNews(tc.req); err != nil {
				t.Fatalf("GetMarketNews 失败: %v", err)
			}
			assertBizKeys(t, captured, tc.want)
		})
	}
}

// === 游标翻页语义 ===

// TestGetMarketNewsCursorPagination 验证 highlight 游标翻页链路：
// 首次调用不传 last_id（biz_content 无该 key，从最新开始），
// 取本页 items 最后一条的 id 作为下一次的 last_id，
// 第二次调用的 biz_content 应携带该游标。
func TestGetMarketNewsCursorPagination(t *testing.T) {
	var captured []string
	pageData := map[string]interface{}{
		"items": []map[string]interface{}{
			{"id": "100", "title": "first", "summary": "", "pubTimestamp": 1, "url": "https://example.com/100"},
			{"id": "200", "title": "last", "summary": "", "pubTimestamp": 2, "url": "https://example.com/200"},
		},
		"pageCount": 1, "pageSize": 20, "totalPage": 5, "totalSize": 100,
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		captured = append(captured, body["biz_content"].(string))
		resp := map[string]interface{}{
			"code":      0,
			"message":   "success",
			"data":      pageData,
			"timestamp": 1700000000,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	nc := newTestNewsClient(server.URL)
	page, err := nc.GetMarketNews(model.NewsHighlightListRequest{})
	if err != nil {
		t.Fatalf("首次 GetMarketNews 失败: %v", err)
	}
	if len(captured) != 1 {
		t.Fatalf("请求数 = %d, want 1", len(captured))
	}
	if biz := decodeBiz(t, captured[0]); len(biz) != 0 {
		t.Fatalf("首次调用不应发送任何 key, got %v", biz)
	}
	if len(page.Items) != 2 {
		t.Fatalf("items 数量 = %d, want 2", len(page.Items))
	}

	// 游标取上一页最后一条的 id，第二次请求应携带该游标
	lastID := page.Items[len(page.Items)-1].ID
	if lastID != "200" {
		t.Fatalf("游标应取本页最后一条 id, got %q", lastID)
	}
	if _, err := nc.GetMarketNews(model.NewsHighlightListRequest{LastID: lastID}); err != nil {
		t.Fatalf("第二次 GetMarketNews 失败: %v", err)
	}
	if len(captured) != 2 {
		t.Fatalf("请求数 = %d, want 2", len(captured))
	}
	assertBizKeys(t, captured[1], map[string]interface{}{"last_id": lastID})
}

// === 路由回归：资讯必须走主网关，不得走行情域名 ===

// TestNewsRoutesToMainGatewayNotQuote 资讯请求（news_stock_list /
// news_highlight_list）必须发往主网关（cfg.ServerURL），不得走行情域名
// （cfg.QuoteServerURL）——行情域名按 license 配置了 <license>-QUOTE 分流，
// 标错路由会把资讯流量打进行情集群。对照断言：行情请求仍走行情域名，
// 防止路由逻辑被误改（与 Java SDK testNewsRoutesToMainGatewayNotQuote 对齐）。
func TestNewsRoutesToMainGatewayNotQuote(t *testing.T) {
	var mainHits, quoteHits int

	mainServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mainHits++
		resp := map[string]interface{}{
			"code":      0,
			"message":   "success",
			"data":      stockNewsWireData(),
			"timestamp": 1700000000,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer mainServer.Close()

	quoteServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		quoteHits++
		resp := map[string]interface{}{
			"code":      0,
			"message":   "success",
			"data":      []map[string]interface{}{{"market": "US", "status": "Trading"}},
			"timestamp": 1700000000,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer quoteServer.Close()

	cfg := newTestConfig(mainServer.URL, quoteServer.URL)

	// NewsClient 必须基于主网关构造（NewNewsClientFromConfig 内部使用 client.NewHttpClient）
	nc := NewNewsClientFromConfig(cfg)
	if _, err := nc.GetStockNews(model.NewsStockListRequest{Symbol: "AAPL"}); err != nil {
		t.Fatalf("GetStockNews 失败: %v", err)
	}
	if _, err := nc.GetMarketNews(model.NewsHighlightListRequest{}); err != nil {
		t.Fatalf("GetMarketNews 失败: %v", err)
	}
	if mainHits != 2 {
		t.Fatalf("两个 news 请求都应命中主网关, mainHits = %d, want 2", mainHits)
	}
	if quoteHits != 0 {
		t.Fatalf("news 请求不得路由到行情域名, quoteHits = %d, want 0", quoteHits)
	}

	// 显式传入主网关 HttpClient 的构造方式同样只命中主网关
	ncExplicit := NewNewsClient(client.NewHttpClient(cfg))
	if _, err := ncExplicit.GetMarketNews(model.NewsHighlightListRequest{}); err != nil {
		t.Fatalf("GetMarketNews 失败: %v", err)
	}
	if mainHits != 3 || quoteHits != 0 {
		t.Fatalf("显式主网关构造路由错误: mainHits=%d want 3, quoteHits=%d want 0", mainHits, quoteHits)
	}

	// 对照组：行情请求仍走行情域名（证明两个 URL 生效、路由分流未被误改）
	qc := quote.NewQuoteClient(client.NewQuoteHttpClient(cfg))
	if _, err := qc.GetMarketState("US"); err != nil {
		t.Fatalf("对照组 GetMarketState 失败: %v", err)
	}
	if quoteHits != 1 {
		t.Fatalf("QUOTE 类型应走行情域名, quoteHits = %d, want 1", quoteHits)
	}
	if mainHits != 3 {
		t.Fatalf("行情请求不应额外命中主网关, mainHits = %d, want 3", mainHits)
	}
}

// === 错误信封 ===

func TestNewsServerError(t *testing.T) {
	// 1010 = 服务端 BIZ_PARAM_ERROR（参数校验失败，如 symbol 为空）
	server := mockNewsErrorServer(t, 1010, "biz param error")
	defer server.Close()

	nc := newTestNewsClient(server.URL)
	_, err := nc.GetStockNews(model.NewsStockListRequest{Symbol: "AAPL"})
	if err == nil {
		t.Fatal("服务端业务错误应返回 error")
	}
	te, ok := err.(*client.TigerError)
	if !ok {
		t.Fatalf("错误类型应为 *client.TigerError, got %T", err)
	}
	if te.Code != 1010 {
		t.Fatalf("错误码 = %d, want 1010", te.Code)
	}
	if te.Category != client.CategoryBizParam {
		t.Fatalf("错误分类 = %s, want %s", te.Category, client.CategoryBizParam)
	}
}
