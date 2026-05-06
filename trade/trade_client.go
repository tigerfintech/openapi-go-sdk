// Package trade 提供交易客户端,封装所有交易相关 API。
package trade

import (
	"encoding/json"

	"github.com/tigerfintech/openapi-go-sdk/client"
	"github.com/tigerfintech/openapi-go-sdk/model"
)

// TradeClient 交易客户端,封装所有交易相关 API。
type TradeClient struct {
	httpClient *client.HttpClient
	account    string
}

// NewTradeClient 创建交易客户端
func NewTradeClient(httpClient *client.HttpClient, account string) *TradeClient {
	return &TradeClient{httpClient: httpClient, account: account}
}

// callInto 内部通用:构造请求、发送、把 data 解码到 out。
func (c *TradeClient) callInto(method string, bizParams interface{}, out interface{}) error {
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

// callIntoItems 剥掉服务端 {"items":[...]} 的外包装。
func (c *TradeClient) callIntoItems(method string, bizParams interface{}, items interface{}) error {
	var wrap struct {
		Items json.RawMessage `json:"items"`
	}
	if err := c.callInto(method, bizParams, &wrap); err != nil {
		return err
	}
	if len(wrap.Items) == 0 || string(wrap.Items) == "null" {
		return nil
	}
	return json.Unmarshal(wrap.Items, items)
}

// === 合约查询 ===

// Contract 查询单个合约。
func (c *TradeClient) Contract(symbol, secType string) ([]model.Contract, error) {
	var out []model.Contract
	err := c.callIntoItems("contract", map[string]interface{}{
		"account":  c.account,
		"symbol":   symbol,
		"sec_type": secType,
	}, &out)
	return out, err
}

// Contracts 批量查询合约。
func (c *TradeClient) Contracts(symbols []string, secType string) ([]model.Contract, error) {
	var out []model.Contract
	err := c.callIntoItems("contracts", map[string]interface{}{
		"account":  c.account,
		"symbols":  symbols,
		"sec_type": secType,
	}, &out)
	return out, err
}

// QuoteContract 查询衍生品合约(期权/认股/牛熊)。
// secType 必须是 OPT/WAR/IOPT; symbol 是标的代码; expiry 是到期日(如 "20260619")。
// 服务端返回 {"symbol":..,"secType":..,"items":[...]}。这里只返回 items。
func (c *TradeClient) QuoteContract(symbol, secType, expiry string) ([]model.Contract, error) {
	var out []model.Contract
	err := c.callIntoItems("quote_contract", map[string]interface{}{
		"account":  c.account,
		"symbols":  []string{symbol},
		"sec_type": secType,
		"expiry":   expiry,
	}, &out)
	return out, err
}

// === 订单操作 ===

// PlaceOrder 下单。
func (c *TradeClient) PlaceOrder(order model.OrderRequest) (*model.PlaceOrderResult, error) {
	order.Account = c.account
	var out model.PlaceOrderResult
	err := c.callInto("place_order", order, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// PreviewOrder 预览订单。
func (c *TradeClient) PreviewOrder(order model.OrderRequest) (*model.PreviewResult, error) {
	order.Account = c.account
	var out model.PreviewResult
	err := c.callInto("preview_order", order, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ModifyOrder 修改订单。
func (c *TradeClient) ModifyOrder(id int64, order model.OrderRequest) (*model.OrderIDResult, error) {
	order.Account = c.account
	order.ID = id
	var out model.OrderIDResult
	err := c.callInto("modify_order", order, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// CancelOrder 取消订单。
func (c *TradeClient) CancelOrder(id int64) (*model.OrderIDResult, error) {
	var out model.OrderIDResult
	err := c.callInto("cancel_order", map[string]interface{}{
		"account": c.account,
		"id":      id,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// === 订单查询 ===

// Orders 查询全部订单。
func (c *TradeClient) Orders() ([]model.Order, error) {
	var out []model.Order
	err := c.callIntoItems("orders", map[string]interface{}{"account": c.account}, &out)
	return out, err
}

// ActiveOrders 查询待成交订单。
func (c *TradeClient) ActiveOrders() ([]model.Order, error) {
	var out []model.Order
	err := c.callIntoItems("active_orders", map[string]interface{}{"account": c.account}, &out)
	return out, err
}

// InactiveOrders 查询已撤销订单。
func (c *TradeClient) InactiveOrders() ([]model.Order, error) {
	var out []model.Order
	err := c.callIntoItems("inactive_orders", map[string]interface{}{"account": c.account}, &out)
	return out, err
}

// FilledOrders 查询已成交订单。
// startDateMs / endDateMs 是 13 位毫秒时间戳,0 可以但服务端要求字段存在。
func (c *TradeClient) FilledOrders(startDateMs, endDateMs int64) ([]model.Order, error) {
	var out []model.Order
	err := c.callIntoItems("filled_orders", map[string]interface{}{
		"account":    c.account,
		"start_date": startDateMs,
		"end_date":   endDateMs,
	}, &out)
	return out, err
}

// OrderTransactions 查询订单成交明细。
// id: 全局订单 ID; symbol/secType 必填。
func (c *TradeClient) OrderTransactions(id int64, symbol, secType string) ([]model.Transaction, error) {
	var out []model.Transaction
	err := c.callIntoItems("order_transactions", map[string]interface{}{
		"account":  c.account,
		"order_id": id,
		"symbol":   symbol,
		"sec_type": secType,
	}, &out)
	return out, err
}

// === 持仓与资产 ===

// Positions 查询持仓。
func (c *TradeClient) Positions() ([]model.Position, error) {
	var out []model.Position
	err := c.callIntoItems("positions", map[string]interface{}{"account": c.account}, &out)
	return out, err
}

// Assets 查询资产。
func (c *TradeClient) Assets() ([]model.Asset, error) {
	var out []model.Asset
	err := c.callIntoItems("assets", map[string]interface{}{"account": c.account}, &out)
	return out, err
}

// PrimeAssets 查询综合账户资产。
func (c *TradeClient) PrimeAssets() (*model.PrimeAsset, error) {
	var out model.PrimeAsset
	err := c.callInto("prime_assets", map[string]interface{}{"account": c.account}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
