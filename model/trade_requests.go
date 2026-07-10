// Package model — trade 请求参数结构。
//
// 字段命名规则：
//   - JSON tag 用服务端接收的 wire 真名（snake_case）
//   - Go 字段名由 wire 名转成 PascalCase，不做客户端别名
//   - 所有字段都是可选，都加 omitempty
package model

// OrdersRequest — 查询订单列表的请求参数。
// 与 Python OrdersParams 对齐。
// 对应 wire methods: orders / active_orders / inactive_orders / filled_orders
type OrdersRequest struct {
	Account     string   `json:"account,omitempty"`
	SecretKey   string   `json:"secret_key,omitempty"`
	Market      string   `json:"market,omitempty"`
	SecType     string   `json:"sec_type,omitempty"`
	SegType     string   `json:"seg_type,omitempty"`
	Symbol      string   `json:"symbol,omitempty"`
	IsBrief     bool     `json:"is_brief,omitempty"`
	StartDate   int64    `json:"start_date,omitempty"` // ms 时间戳
	EndDate     int64    `json:"end_date,omitempty"`
	Limit       int      `json:"limit,omitempty"`
	States      []string `json:"states,omitempty"` // OrderStatus 值列表
	ParentId    int64    `json:"parent_id,omitempty"`
	SortBy      string   `json:"sort_by,omitempty"`
	ShowCharges *bool    `json:"show_charges,omitempty"` // Python 用 is not None 判断
	PageToken   string   `json:"page_token,omitempty"`
	Lang        string   `json:"lang,omitempty"`
}

// GetOrderRequest — 按订单 ID 查询单个订单。
// 对应 wire method: orders
type GetOrderRequest struct {
	Account     string `json:"account,omitempty"`
	SecretKey   string `json:"secret_key,omitempty"`
	Id          int64  `json:"id,omitempty"`       // 全局订单 ID
	OrderId     int64  `json:"order_id,omitempty"` // 账户维度订单 ID
	IsBrief     bool   `json:"is_brief,omitempty"`
	ShowCharges *bool  `json:"show_charges,omitempty"`
	Lang        string `json:"lang,omitempty"`
}

// OrderTransactionsRequest — 查询订单成交明细。
// 与 Python TransactionsParams 对齐。
// 对应 wire method: order_transactions
type OrderTransactionsRequest struct {
	Account   string  `json:"account,omitempty"`
	SecretKey string  `json:"secret_key,omitempty"`
	OrderId   int64   `json:"order_id,omitempty"`
	Symbol    string  `json:"symbol,omitempty"`
	SecType   string  `json:"sec_type,omitempty"`
	StartDate int64   `json:"start_date,omitempty"` // ms 时间戳
	EndDate   int64   `json:"end_date,omitempty"`
	SinceDate string  `json:"since_date,omitempty"` // yyyyMMdd
	ToDate    string  `json:"to_date,omitempty"`
	Limit     int     `json:"limit,omitempty"`
	Expiry    string  `json:"expiry,omitempty"`
	Strike    float64 `json:"strike,omitempty"`
	Right     string  `json:"right,omitempty"`
	PageToken string  `json:"page_token,omitempty"`
	Lang      string  `json:"lang,omitempty"`
}

// PositionsRequest — 查询持仓。
// 与 Python PositionParams 对齐。
// 对应 wire method: positions
type PositionsRequest struct {
	Account        string   `json:"account,omitempty"`
	SecretKey      string   `json:"secret_key,omitempty"`
	Symbol         string   `json:"symbol,omitempty"`
	SecType        string   `json:"sec_type,omitempty"`
	Currency       string   `json:"currency,omitempty"`
	Market         string   `json:"market,omitempty"`
	SubAccounts    []string `json:"sub_accounts,omitempty"`
	Expiry         string   `json:"expiry,omitempty"`
	Strike         string   `json:"strike,omitempty"`
	Right          string   `json:"right,omitempty"`
	AssetQuoteType string   `json:"asset_quote_type,omitempty"`
	Lang           string   `json:"lang,omitempty"`
}

// AssetsRequest — 查询资产。
// 与 Python AssetParams 对齐。
// 对应 wire method: assets / prime_assets
type AssetsRequest struct {
	Account      string   `json:"account,omitempty"`
	SecretKey    string   `json:"secret_key,omitempty"`
	Segment      bool     `json:"segment,omitempty"`
	MarketValue  bool     `json:"market_value,omitempty"`
	SubAccounts  []string `json:"sub_accounts,omitempty"`
	BaseCurrency string   `json:"base_currency,omitempty"`
	Consolidated *bool    `json:"consolidated,omitempty"` // Python 用 is not None 判断
	Lang         string   `json:"lang,omitempty"`
}

// ManagedAccountsRequest — 查询机构子账户列表。
// 对应 wire method: accounts
type ManagedAccountsRequest struct {
	Account   string `json:"account,omitempty"`
	SecretKey string `json:"secret_key,omitempty"`
	Lang      string `json:"lang,omitempty"`
}

// DerivativeContractsRequest — 查询衍生品合约列表。
// 对应 wire method: derivative_contracts
type DerivativeContractsRequest struct {
	Account   string   `json:"account,omitempty"`
	SecretKey string   `json:"secret_key,omitempty"`
	Symbols   []string `json:"symbols,omitempty"`
	SecType   string   `json:"sec_type,omitempty"`
	Expiry    string   `json:"expiry,omitempty"`
	Lang      string   `json:"lang,omitempty"`
}

// AnalyticsAssetRequest — 查询资产分析。
// 对应 wire method: analytics_asset
type AnalyticsAssetRequest struct {
	Account     string   `json:"account,omitempty"`
	SubAccount  string   `json:"sub_account,omitempty"`
	SecretKey   string   `json:"secret_key,omitempty"`
	SegType     string   `json:"seg_type,omitempty"`
	Currency    string   `json:"currency,omitempty"`
	SubAccounts []string `json:"sub_accounts,omitempty"`
	StartDate   string   `json:"start_date,omitempty"` // yyyy-MM-dd
	EndDate     string   `json:"end_date,omitempty"`
	Lang        string   `json:"lang,omitempty"`
}

// AggregateAssetsRequest — 查询综合资产（base_currency 视角下汇总）。
// 对应 wire method: aggregate_assets
type AggregateAssetsRequest struct {
	Account      string `json:"account,omitempty"`
	SecretKey    string `json:"secret_key,omitempty"`
	SegType      string `json:"seg_type,omitempty"`
	BaseCurrency string `json:"base_currency,omitempty"`
	Lang         string `json:"lang,omitempty"`
}

// SegmentFundRequest — 子账户资金调拨（available/history/transfer/cancel 共用）。
// 对应 wire methods: segment_fund_available / segment_fund_history / transfer_segment_fund / cancel_segment_fund
type SegmentFundRequest struct {
	ID          string  `json:"id,omitempty"`
	Account     string  `json:"account,omitempty"`
	SecretKey   string  `json:"secret_key,omitempty"`
	FromSegment string  `json:"from_segment,omitempty"`
	ToSegment   string  `json:"to_segment,omitempty"`
	Currency    string  `json:"currency,omitempty"`
	Amount      float64 `json:"amount,omitempty"`
	Limit       int     `json:"limit,omitempty"`
	Lang        string  `json:"lang,omitempty"`
}

// ForexOrderRequest — 外汇下单。
// 对应 wire method: place_forex_order
type ForexOrderRequest struct {
	Account        string  `json:"account,omitempty"`
	SecretKey      string  `json:"secret_key,omitempty"`
	SegType        string  `json:"seg_type,omitempty"`
	SourceCurrency string  `json:"source_currency,omitempty"`
	SourceAmount   float64 `json:"source_amount,omitempty"`
	TargetCurrency string  `json:"target_currency,omitempty"`
	ExternalID     string  `json:"external_id,omitempty"`
	TimeInForce    string  `json:"time_in_force,omitempty"`
	Lang           string  `json:"lang,omitempty"`
}

// EstimateTradableQuantityRequest — 估算可交易数量。
// 对应 wire method: estimate_tradable_quantity
// 注意：Python 这里把 Contract 拆平为 symbol/expiry/strike/right/sec_type 发送（不是嵌套对象）。
type EstimateTradableQuantityRequest struct {
	Account    string  `json:"account,omitempty"`
	SecretKey  string  `json:"secret_key,omitempty"`
	Symbol     string  `json:"symbol,omitempty"`
	Expiry     string  `json:"expiry,omitempty"`
	Strike     string  `json:"strike,omitempty"`
	Right      string  `json:"right,omitempty"`
	SecType    string  `json:"sec_type,omitempty"`
	SegType    string  `json:"seg_type,omitempty"`
	Action     string  `json:"action,omitempty"`
	OrderType  string  `json:"order_type,omitempty"`
	LimitPrice float64 `json:"limit_price,omitempty"`
	StopPrice  float64 `json:"stop_price,omitempty"`
	Lang       string  `json:"lang,omitempty"`
}

// FundingHistoryRequest — 资金调拨历史。
// 对应 wire method: transfer_fund
type FundingHistoryRequest struct {
	Account   string `json:"account,omitempty"`
	SecretKey string `json:"secret_key,omitempty"`
	SegType   string `json:"seg_type,omitempty"`
	Lang      string `json:"lang,omitempty"`
}

// FundDetailsRequest — 资金明细。
// 对应 wire method: fund_details
type FundDetailsRequest struct {
	Account   string   `json:"account,omitempty"`
	SecretKey string   `json:"secret_key,omitempty"`
	SegTypes  []string `json:"seg_types,omitempty"`
	FundType  string   `json:"fund_type,omitempty"`
	Currency  string   `json:"currency,omitempty"`
	StartDate int64    `json:"start_date,omitempty"`
	EndDate   int64    `json:"end_date,omitempty"`
	Start     int      `json:"start,omitempty"`
	Limit     int      `json:"limit,omitempty"`
	Lang      string   `json:"lang,omitempty"`
}

// PositionTransferRequest — 内部转股（跨账户换仓）。
// 对应 wire method: position_transfer
type PositionTransferRequest struct {
	FromAccount string         `json:"from_account,omitempty"`
	ToAccount   string         `json:"to_account,omitempty"`
	Market      string         `json:"market,omitempty"`
	Transfers   []TransferItem `json:"transfers,omitempty"`
	SecretKey   string         `json:"secret_key,omitempty"`
}

// PositionTransferRecordsRequest — 内部转股记录查询。
// 对应 wire method: position_transfer_records
// 注意：账户字段服务端为 account_id（不是 account）。
type PositionTransferRecordsRequest struct {
	AccountID string `json:"account_id,omitempty"`
	SinceDate string `json:"since_date,omitempty"` // YYYY-MM-DD
	ToDate    string `json:"to_date,omitempty"`
	Status    string `json:"status,omitempty"`
	Market    string `json:"market,omitempty"`
	Symbol    string `json:"symbol,omitempty"`
	SecretKey string `json:"secret_key,omitempty"`
	Lang      string `json:"lang,omitempty"`
}

// PositionTransferDetailRequest — 内部转股详情（按 ID）。
// 对应 wire method: position_transfer_detail
type PositionTransferDetailRequest struct {
	ID        string `json:"id,omitempty"`
	AccountID string `json:"account_id,omitempty"`
	SecretKey string `json:"secret_key,omitempty"`
	Lang      string `json:"lang,omitempty"`
}

// PositionTransferExternalRecordsRequest — 外部转股记录查询。
// 对应 wire method: position_transfer_external_records
// 参数字段与 PositionTransferRecordsRequest 相同。
type PositionTransferExternalRecordsRequest = PositionTransferRecordsRequest

// OptionExerciseCheckRequest — 行权检验请求。
// 对应 wire method: option_exercise_check
type OptionExerciseCheckRequest struct {
	Account       string  `json:"account,omitempty"`
	SecretKey     string  `json:"secret_key,omitempty"`
	ContractId    int64   `json:"contract_id,omitempty"`
	Type          string  `json:"type,omitempty"` // Exercise | Expire
	Quantity      float64 `json:"quantity,omitempty"`
	ExecutingDate string  `json:"executing_date,omitempty"` // yyyy-MM-dd，Exercise 类型建议填
	IsForce       *bool   `json:"is_force,omitempty"`       // Exercise 类型建议填
	ItmRate       *int    `json:"itm_rate,omitempty"`       // 0–10，Expire 类型专用
	Lang          string  `json:"lang,omitempty"`
}

// OptionExercisePositionRequest — 查询可行权持仓请求。
// 对应 wire method: option_exercise_position
type OptionExercisePositionRequest struct {
	Account   string `json:"account,omitempty"`
	SecretKey string `json:"secret_key,omitempty"`
	Type      string `json:"type,omitempty"` // Exercise | Expire
	Lang      string `json:"lang,omitempty"`
}

// OptionExerciseSubmitRequest — 提交行权申请请求。
// 对应 wire method: option_exercise_submit
type OptionExerciseSubmitRequest struct {
	Account       string  `json:"account,omitempty"`
	SecretKey     string  `json:"secret_key,omitempty"`
	ContractId    int64   `json:"contract_id,omitempty"`
	Type          string  `json:"type,omitempty"` // Exercise | Expire
	Quantity      float64 `json:"quantity,omitempty"`
	ExecutingDate string  `json:"executing_date,omitempty"` // Exercise 必填，yyyy-MM-dd
	IsForce       *bool   `json:"is_force,omitempty"`       // Exercise 必填
	ItmRate       *int    `json:"itm_rate,omitempty"`       // 0–10，Expire 专用
	Lang          string  `json:"lang,omitempty"`
}

// OptionExercisePageRequest — 分页查询行权记录请求。
// 对应 wire method: option_exercise_record
type OptionExercisePageRequest struct {
	Account   string `json:"account,omitempty"`
	SecretKey string `json:"secret_key,omitempty"`
	Page      int    `json:"page,omitempty"`   // 从 1 开始，默认 1
	Size      int    `json:"size,omitempty"`   // 1–100，默认 20
	Status    string `json:"status,omitempty"` // New | Cancel | Success | Fail
	Type      string `json:"type,omitempty"`   // Exercise | Expire
	Symbol    string `json:"symbol,omitempty"`
	OrderBy   string `json:"order_by,omitempty"` // symbol | expire_date | strike | is_call
	Lang      string `json:"lang,omitempty"`
}

// OptionExerciseCancelRequest — 撤销行权申请请求。
// 对应 wire method: option_exercise_cancel
type OptionExerciseCancelRequest struct {
	Account   string `json:"account,omitempty"`
	SecretKey string `json:"secret_key,omitempty"`
	Id        int64  `json:"id,omitempty"`
	Lang      string `json:"lang,omitempty"`
}
