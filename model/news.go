package model

// NewsItem 单条资讯，news_stock_list 与 news_highlight_list 共用。
// 服务端出参为 camelCase，字段名与 JSON key 相同。
type NewsItem struct {
	// ID 文章 ID（同时是 highlight 游标翻页 last_id 的取值来源）
	ID string `json:"id"`
	// Title 标题
	Title string `json:"title"`
	// Summary 摘要
	Summary string `json:"summary"`
	// PubTimestamp 发布时间戳（秒级）
	PubTimestamp int64 `json:"pubTimestamp"`
	// URL 详情页 URL
	URL string `json:"url"`
}

// NewsPage 资讯列表分页信封，news_stock_list 与 news_highlight_list 共用。
type NewsPage struct {
	// Items 当前页资讯列表
	Items []NewsItem `json:"items"`
	// PageCount 当前页码（highlight 接口恒为 1，游标翻页）
	PageCount int `json:"pageCount"`
	// PageSize 每页条数
	PageSize int `json:"pageSize"`
	// TotalPage 总页数
	TotalPage int `json:"totalPage"`
	// TotalSize 总条数
	TotalSize int `json:"totalSize"`
}
