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
// req 所有字段可选；Account 留空时自动填默认账户。
func (c *TradeClient) Orders(req model.OrdersRequest) ([]model.Order, error) {
	if req.Account == "" {
		req.Account = c.account
	}
	var out []model.Order
	err := c.callIntoItems("orders", req, &out)
	return out, err
}

// ActiveOrders 查询待成交订单。支持按 ParentId 过滤附加订单。
func (c *TradeClient) ActiveOrders(req model.OrdersRequest) ([]model.Order, error) {
	if req.Account == "" {
		req.Account = c.account
	}
	var out []model.Order
	err := c.callIntoItems("active_orders", req, &out)
	return out, err
}

// InactiveOrders 查询已撤销订单。
func (c *TradeClient) InactiveOrders(req model.OrdersRequest) ([]model.Order, error) {
	if req.Account == "" {
		req.Account = c.account
	}
	var out []model.Order
	err := c.callIntoItems("inactive_orders", req, &out)
	return out, err
}

// FilledOrders 查询已成交订单。
func (c *TradeClient) FilledOrders(req model.OrdersRequest) ([]model.Order, error) {
	if req.Account == "" {
		req.Account = c.account
	}
	var out []model.Order
	err := c.callIntoItems("filled_orders", req, &out)
	return out, err
}

// GetOrder 按订单 ID 查询单个订单详情。
// wire method: orders（传 id 或 order_id 参数，服务端直接返回单个 Order 对象，不是 {items:[]} 包装）。
// 必须提供 Id 或 OrderId 之一。没有匹配时返回 nil。
func (c *TradeClient) GetOrder(req model.GetOrderRequest) (*model.Order, error) {
	if req.Account == "" {
		req.Account = c.account
	}
	var out model.Order
	if err := c.callInto("orders", req, &out); err != nil {
		return nil, err
	}
	if out.ID == 0 && out.OrderId == 0 {
		return nil, nil
	}
	return &out, nil
}

// OrderTransactions 查询订单成交明细。
// 所有过滤字段（Symbol / SecType / OrderId / 时间范围等）均可选。
func (c *TradeClient) OrderTransactions(req model.OrderTransactionsRequest) ([]model.Transaction, error) {
	if req.Account == "" {
		req.Account = c.account
	}
	var out []model.Transaction
	err := c.callIntoItems("order_transactions", req, &out)
	return out, err
}

// === 持仓与资产 ===

// Positions 查询持仓。支持按 Symbol / SecType / Currency / Market 等过滤。
func (c *TradeClient) Positions(req model.PositionsRequest) ([]model.Position, error) {
	if req.Account == "" {
		req.Account = c.account
	}
	var out []model.Position
	err := c.callIntoItems("positions", req, &out)
	return out, err
}

// Assets 查询资产。支持子账户列表、按市场/币种聚合等选项。
func (c *TradeClient) Assets(req model.AssetsRequest) ([]model.Asset, error) {
	if req.Account == "" {
		req.Account = c.account
	}
	var out []model.Asset
	err := c.callIntoItems("assets", req, &out)
	return out, err
}

// PrimeAssets 查询综合账户资产。
func (c *TradeClient) PrimeAssets(req model.AssetsRequest) (*model.PrimeAsset, error) {
	if req.Account == "" {
		req.Account = c.account
	}
	var out model.PrimeAsset
	if err := c.callInto("prime_assets", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// === 账户管理 ===

// ManagedAccounts 查询机构子账户列表。
func (c *TradeClient) ManagedAccounts(req model.ManagedAccountsRequest) ([]model.ManagedAccount, error) {
	if req.Account == "" {
		req.Account = c.account
	}
	var out []model.ManagedAccount
	err := c.callIntoItems("accounts", req, &out)
	return out, err
}

// DerivativeContracts 查询衍生品合约列表。wire: quote_contract
// 注意：Python get_derivative_contracts 与 get_contract(secType=OPT/WAR/IOPT) 用同一 wire 方法。
func (c *TradeClient) DerivativeContracts(req model.DerivativeContractsRequest) ([]model.Contract, error) {
	if req.Account == "" {
		req.Account = c.account
	}
	var out []model.Contract
	err := c.callIntoItems("quote_contract", req, &out)
	return out, err
}

// === 资产分析/综合 ===

// AnalyticsAsset 按日资产分析（P&L / 净值曲线 / 持仓价值）。
func (c *TradeClient) AnalyticsAsset(req model.AnalyticsAssetRequest) ([]model.AnalyticsAsset, error) {
	if req.Account == "" {
		req.Account = c.account
	}
	var out []model.AnalyticsAsset
	err := c.callIntoItems("analytics_asset", req, &out)
	return out, err
}

// AggregateAssets 综合账户 base_currency 维度资产汇总。
func (c *TradeClient) AggregateAssets(req model.AggregateAssetsRequest) (*model.AggregateAssets, error) {
	if req.Account == "" {
		req.Account = c.account
	}
	var out model.AggregateAssets
	if err := c.callInto("aggregate_assets", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// === 估算/外汇 ===

// EstimateTradableQuantity 估算可交易数量。
func (c *TradeClient) EstimateTradableQuantity(req model.EstimateTradableQuantityRequest) (*model.EstimateTradableQuantity, error) {
	if req.Account == "" {
		req.Account = c.account
	}
	var out model.EstimateTradableQuantity
	if err := c.callInto("estimate_tradable_quantity", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// PlaceForexOrder 外汇下单（子账户 base_currency 互换）。
func (c *TradeClient) PlaceForexOrder(req model.ForexOrderRequest) (*model.ForexOrderResult, error) {
	if req.Account == "" {
		req.Account = c.account
	}
	var out model.ForexOrderResult
	if err := c.callInto("place_forex_order", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// === 子账户资金调拨 ===

// SegmentFundAvailable 查询可用于调拨的金额。
// 服务端返回数组（按 SegmentType 分项）。
func (c *TradeClient) SegmentFundAvailable(req model.SegmentFundRequest) ([]model.SegmentFundAvailableItem, error) {
	if req.Account == "" {
		req.Account = c.account
	}
	var out []model.SegmentFundAvailableItem
	err := c.callInto("segment_fund_available", req, &out)
	return out, err
}

// SegmentFundHistory 查询子账户调拨历史。
// 服务端直接返回数组（不带 items 包装）。
func (c *TradeClient) SegmentFundHistory(req model.SegmentFundRequest) ([]model.SegmentFundHistoryItem, error) {
	if req.Account == "" {
		req.Account = c.account
	}
	var out []model.SegmentFundHistoryItem
	err := c.callInto("segment_fund_history", req, &out)
	return out, err
}

// TransferSegmentFund 子账户间资金调拨。
func (c *TradeClient) TransferSegmentFund(req model.SegmentFundRequest) (*model.SegmentFund, error) {
	if req.Account == "" {
		req.Account = c.account
	}
	var out model.SegmentFund
	if err := c.callInto("transfer_segment_fund", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CancelSegmentFund 撤销一笔子账户调拨申请。
func (c *TradeClient) CancelSegmentFund(req model.SegmentFundRequest) (*model.SegmentFund, error) {
	if req.Account == "" {
		req.Account = c.account
	}
	var out model.SegmentFund
	if err := c.callInto("cancel_segment_fund", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// === 资金明细/历史 ===

// FundDetails 资金流水明细（按子账户/币种/业务类型过滤）。
func (c *TradeClient) FundDetails(req model.FundDetailsRequest) ([]model.FundDetails, error) {
	if req.Account == "" {
		req.Account = c.account
	}
	var out []model.FundDetails
	err := c.callIntoItems("fund_details", req, &out)
	return out, err
}

// FundingHistory 资金调拨记录（wire method: transfer_fund）。
func (c *TradeClient) FundingHistory(req model.FundingHistoryRequest) ([]model.FundingHistoryItem, error) {
	if req.Account == "" {
		req.Account = c.account
	}
	var out []model.FundingHistoryItem
	err := c.callIntoItems("transfer_fund", req, &out)
	return out, err
}

// === 内部/外部转股 ===

// TransferPosition 内部转股（跨子账户换仓）。
func (c *TradeClient) TransferPosition(req model.PositionTransferRequest) (*model.PositionTransferRecord, error) {
	if req.FromAccount == "" {
		req.FromAccount = c.account
	}
	var out model.PositionTransferRecord
	if err := c.callInto("position_transfer", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// PositionTransferRecords 查询内部转股记录列表。
// 注意：账户参数使用 AccountID（对应 wire account_id）。
func (c *TradeClient) PositionTransferRecords(req model.PositionTransferRecordsRequest) ([]model.PositionTransferRecord, error) {
	if req.AccountID == "" {
		req.AccountID = c.account
	}
	var out []model.PositionTransferRecord
	err := c.callIntoItems("position_transfer_records", req, &out)
	return out, err
}

// PositionTransferDetail 按 ID 查询内部转股详情。
func (c *TradeClient) PositionTransferDetail(req model.PositionTransferDetailRequest) (*model.PositionTransferDetail, error) {
	if req.AccountID == "" {
		req.AccountID = c.account
	}
	var out model.PositionTransferDetail
	if err := c.callInto("position_transfer_detail", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// PositionTransferExternalRecords 查询外部转股记录列表。
func (c *TradeClient) PositionTransferExternalRecords(req model.PositionTransferExternalRecordsRequest) ([]model.PositionTransferExternalRecord, error) {
	if req.AccountID == "" {
		req.AccountID = c.account
	}
	var out []model.PositionTransferExternalRecord
	err := c.callIntoItems("position_transfer_external_records", req, &out)
	return out, err
}
