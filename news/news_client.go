// Package news 提供资讯查询客户端，封装个股新闻列表（news_stock_list）
// 与市场要闻列表（news_highlight_list）两个接口。
//
// 路由说明：资讯接口必须走主网关（config.ServerURL，与交易域名一致），
// 不走行情集群域名（QuoteServerURL）。本 SDK 的行情/主网关路由由
// HttpClient 构造方式决定（client.NewQuoteHttpClient 会把 ServerURL
// 切换为 QuoteServerURL），因此资讯客户端必须基于 client.NewHttpClient
// 构造的主网关客户端。NewNewsClientFromConfig 内部即如此构造；
// 切勿把 client.NewQuoteHttpClient 构造的行情客户端传给 NewNewsClient，
// 否则资讯请求会被路由到行情集群。
package news

import (
	"github.com/tigerfintech/openapi-go-sdk/client"
	"github.com/tigerfintech/openapi-go-sdk/config"
	"github.com/tigerfintech/openapi-go-sdk/model"
)

// NewsClient 资讯客户端。
type NewsClient struct {
	httpClient *client.HttpClient
}

// NewNewsClient 创建资讯客户端。
// httpClient 必须是基于主网关构造的客户端（client.NewHttpClient）；
// 传入 client.NewQuoteHttpClient 构造的行情客户端会把资讯请求
// 错误路由到行情域名。
func NewNewsClient(httpClient *client.HttpClient) *NewsClient {
	return &NewsClient{httpClient: httpClient}
}

// NewNewsClientFromConfig 从 ClientConfig 直接创建资讯客户端，无需手动构造 HttpClient。
// 内部使用 client.NewHttpClient（主网关域名），保证资讯请求不走行情集群。
func NewNewsClientFromConfig(cfg *config.ClientConfig) *NewsClient {
	return &NewsClient{httpClient: client.NewHttpClient(cfg)}
}

// callInto 内部通用：构造请求、发送、把 data 解码到 out。
func (c *NewsClient) callInto(method string, bizParams interface{}, out interface{}) error {
	req, err := client.NewApiRequest(method, bizParams)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Execute(req)
	if err != nil {
		return err
	}
	return client.UnmarshalData(resp.Data, out)
}

// GetStockNews 查询个股新闻列表（method=news_stock_list）。
// 返回分页信封 *model.NewsPage：Items 每条含 ID / Title / Summary /
// PubTimestamp（秒级时间戳）/ URL，另有 PageCount / PageSize /
// TotalPage / TotalSize 分页字段。
//
// 翻页：调整 req.PageCount / req.PageSize（缺省时服务端默认 1 / 20，
// PageCount >= 1、PageSize 1-50）；Lang 可选（zh_CN/zh_TW/en_US）。
// 详细规则见 model.NewsStockListRequest。
func (c *NewsClient) GetStockNews(req model.NewsStockListRequest) (*model.NewsPage, error) {
	var out model.NewsPage
	if err := c.callInto("news_stock_list", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetMarketNews 查询市场要闻列表（method=news_highlight_list）。
// 返回分页信封 *model.NewsPage，字段同 GetStockNews；
// 本接口的 PageCount 恒为 1（服务端固定，游标翻页）。
//
// 翻页只走 LastID 游标：取上一页 items 最后一条的 ID 作为
// req.LastID 传入即取下一页；留空从最新开始。Lang 语义同 GetStockNews。
func (c *NewsClient) GetMarketNews(req model.NewsHighlightListRequest) (*model.NewsPage, error) {
	var out model.NewsPage
	if err := c.callInto("news_highlight_list", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
