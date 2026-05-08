// Package push 提供 WebSocket 推送客户端，支持实时行情和账户推送。
package push

// SubjectType 订阅主题类型
type SubjectType string

const (
	SubjectQuote       SubjectType = "quote"
	SubjectTick        SubjectType = "tick"
	SubjectDepth       SubjectType = "depth"
	SubjectOption      SubjectType = "option"
	SubjectFuture      SubjectType = "future"
	SubjectKline       SubjectType = "kline"
	SubjectStockTop    SubjectType = "stock_top"
	SubjectOptionTop   SubjectType = "option_top"
	SubjectFullTick    SubjectType = "full_tick"
	SubjectQuoteBBO    SubjectType = "quote_bbo"
	SubjectCc          SubjectType = "cc"
	SubjectMarket      SubjectType = "market"
	SubjectAsset       SubjectType = "asset"
	SubjectPosition    SubjectType = "position"
	SubjectOrder       SubjectType = "order"
	SubjectTransaction SubjectType = "transaction"
)
