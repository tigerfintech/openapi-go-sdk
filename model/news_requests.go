package model

// NewsStockListRequest — 个股新闻列表请求。wire: news_stock_list
//
// 契约见 stock-openapi docs/news_api_design.md §3.1：
//   - Symbol 必填、单标的（下游资讯服务实际不支持批量查询，传多个仅单个标的生效），
//     空值由服务端校验报参数错误；
//   - PageCount / PageSize 可选：缺省（零值，omitempty 不发送）= 服务端默认 1 / 20；
//     显式传值须 PageCount >= 1、PageSize 1-50（闭区间），越界由服务端校验报参数错误；
//   - Lang 可选（zh_CN/zh_TW/en_US）。
type NewsStockListRequest struct {
	// Symbol 股票代码，必填、仅支持单只。
	Symbol string `json:"symbol"`
	// Lang 语言（zh_CN/zh_TW/en_US）。
	Lang string `json:"lang,omitempty"`
	// PageCount 页码，缺省 = 服务端默认 1。
	PageCount int `json:"page_count,omitempty"`
	// PageSize 每页条数，缺省 = 服务端默认 20。
	PageSize int `json:"page_size,omitempty"`
}

// NewsHighlightListRequest — 市场要闻列表请求。wire: news_highlight_list
//
// 契约见 stock-openapi docs/news_api_design.md §3.2：
//   - 翻页只走 LastID 游标（服务端已撤除 page_count/page_size），
//     取上一页 items 最后一条的 id 作为 LastID；空 = 不发送 = 从最新开始；
//   - Lang 可选，语义同 NewsStockListRequest.Lang。
type NewsHighlightListRequest struct {
	// Lang 语言（zh_CN/zh_TW/en_US）。
	Lang string `json:"lang,omitempty"`
	// LastID 游标翻页参数，取上一页 items 最后一条的 id；空 = 从最新开始。
	LastID string `json:"last_id,omitempty"`
}
