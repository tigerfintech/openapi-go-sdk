# Changelog

All notable changes to the Tiger Brokers OpenAPI Go SDK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.4.4] - 2026-07-07

### Breaking Changes

- **多 symbol 支持（行情接口签名变更）**：下列接口参数由单 symbol 改为 slice，调用方需更新：
  - `GetKline(symbol, period string)` → `GetKline(symbols []string, period string)`
  - `GetOptionExpiration(symbol string)` → `GetOptionExpiration(symbols []string)`
  - `GetOptionChain(symbol, expiry string)` → `GetOptionChain(items [][2]string)`（每项为 `[symbol, "YYYY-MM-DD"]` 对）
  - `GetOptionKline(identifier, period string)` → `GetOptionKline(identifiers []string, period string)`
- **删除重复方法 `GetOptionBars`**：该方法与 `GetOptionKline` 均对应 `option_kline` API，已删除，请统一使用 `GetOptionKline`。
- **`NewTradeClientFromConfig` 签名变更**：原 `NewTradeClientFromConfig(httpClient, cfg)` 改为 `NewTradeClientFromConfig(cfg)`，内部自动创建 `HttpClient`，无需调用方手动传入。
- **`GetKline` 签名变更（Request 结构体）**：`GetKline(symbols []string, period string)` → `GetKline(req model.KlineRequest)`；删除同名 `GetBars` 方法。
- **`GetKlineByPage` 重命名**：`GetBarsByPage(req BarsByPageRequest)` → `GetKlineByPage(req KlineByPageRequest)`。
- **`GetFutureBars` 删除**：改用 `GetFutureKline(req FutureKlineRequest)`，`FutureKlineRequest` 合并了原 `FutureBarsRequest` 的全部字段（均带 `omitempty`）；删除 `FutureBarsRequest` 类型。
- **`GetFutureBarsByPage` 重命名**：`GetFutureBarsByPage(req FutureBarsByPageRequest)` → `GetFutureKlineByPage(req FutureKlineByPageRequest)`。
- **请求类型重命名**：`BarsRequest` → `KlineRequest`，`BarsByPageRequest` → `KlineByPageRequest`，`FutureBarsRequest`（已删除，字段合并入 `FutureKlineRequest`），`FutureBarsByPageRequest` → `FutureKlineByPageRequest`。

### Added

- **`NewQuoteClientFromConfig(cfg)`**：直接从 `ClientConfig` 创建 `QuoteClient`，内部自动使用 `NewQuoteHttpClient(cfg)`，无需调用方手动构造 `HttpClient`。



### Added

- **请求日志**：`Execute` / `ExecuteRaw` 在公共层统一输出日志；成功 DEBUG，业务错误 WARN（含 code/msg），重试 WARN，失败 ERROR；下单遇 EOF 额外提示"可能已提交，请查询确认"
- **`WithLogger(l)` ClientOption**：注入自定义 Logger；传 `&logger.NopLogger{}` 可静默所有 SDK 日志
- **`IsStaleConnectionError(err)`**：判断 EOF / connection-reset / broken-pipe 类错误的辅助函数

### Fixed

- **HTTP 非 2xx 响应**：之前状态码 500 且 body 为空时报 "unexpected end of JSON"；现在非 2xx 先尝试解析 `TigerError`，失败时将 HTTP 状态码附加到错误信息
- **`AddonEntitlement.UserLevel` 类型不匹配**：服务端偶发返回数字而非字符串，新增 `model.FlexString` 类型兼容两种格式，不再报 `cannot unmarshal number into string`
- **`FundDetails.ID` 类型不匹配**：服务端返回字符串而非数字，新增 `model.FlexInt64` 类型兼容两种格式，不再报 `cannot unmarshal string into int64`

### Changed

- **连接池**：显式配置 `http.Transport`：`IdleConnTimeout` 90s→60s、`MaxIdleConnsPerHost` 2→10、新增 `DialContext`（TCP 超时 10s）和 `TLSHandshakeTimeout: 10s`

## [0.4.2] - 2026-07-01

### Added

- **`config.WithServerURL` 选项**：新增 `config.WithServerURL(url)`，用于覆盖默认生产 gateway（此前仅有 `WithQuoteServerURL`）。
- **从 properties 加载 `server_url` / `quote_server_url`**：`NewClientConfig` 现在读取 properties 文件中的 `server_url` 和 `quote_server_url`（字段名与 Python SDK 对齐）；此前这两个字段被忽略，导致配置文件中的非生产地址被默认生产 gateway 覆盖。
- **`HttpClient.SecretKey()`**：暴露底层 config 的机构 Secret Key（未设置时返回空字符串）。

### Fixed

- **`NewTradeClient` 未注入 `SecretKey`**：`NewTradeClient(httpClient, account)` 之前不注入 `secretKey`，机构账户若用该构造函数（且未额外调用 `SetSecretKey`）会在 `CancelOrder` 等接口缺失 `secret_key`；现改为自动从 `httpClient` 的 config 注入，与 `NewTradeClientFromConfig` 行为一致。
- **凭据 JSON 序列化泄露**：`ClientConfig.PrivateKey` / `SecretKey` 的 json tag 改为 `"-"`，避免 `json.Marshal(cfg)`（日志、调试 dump、错误上报）时明文输出私钥与密钥。
- **`Version` 常量未同步**：修正 `tigeropen.go` 中的 `Version` 常量（此前停留在 `0.3.7`，未随发布更新）。
## [0.4.0] - 2026-06-24

### Added

- **冰山单辅助函数**：新增 `IcebergOrder(account, symbol, secType, action, quantity, price, displaySize)` 及完整参数版本 `IcebergOrderFull`，支持 `MinDisplaySize`、`CheckIntervals`、`PriceType`（`IcebergPriceTypeLimit`/`IcebergPriceTypeOpponent`）、`StartTime`/`EndTime`（epoch ms）
- **`IcebergPriceType` 枚举**：新增 `IcebergPriceTypeLimit`（固定限价）和 `IcebergPriceTypeOpponent`（对手价）
- **`Order` 结构体新增冰山单字段**：`DisplaySize`、`MinDisplaySize`、`CheckIntervals`、`PriceType`、`StartTime`、`EndTime`
- **单元测试**：`TestIcebergOrder`、`TestIcebergOrderFull`、`TestIcebergOrderFullZeroTimes`，覆盖基础构造、完整参数、零值省略行为

### Fixed

- **`CancelOrder` 未注入 `SecretKey`**：`TradeClient.CancelOrder` 之前无条件将空字符串 `secret_key` 发送至服务端，现改为仅在 `c.secretKey != ""` 时才附带该字段，与 `PlaceOrder`/`ModifyOrder` 行为对齐。

## [0.3.9] - 2026-06-09

### Added

- **期权行权 5 个接口**：新增 `OptionExerciseCheck`、`OptionExercisePositions`、`OptionExerciseSubmit`、`OptionExerciseRecords`、`OptionExerciseCancel`，对应 wire method `option_exercise_check / option_exercise_position / option_exercise_submit / option_exercise_record / option_exercise_cancel`。全部接口自动注入 `Account` 与 `SecretKey`。
- **请求/响应模型**：新增 `OptionExerciseCheckRequest`、`OptionExercisePositionRequest`、`OptionExerciseSubmitRequest`、`OptionExercisePageRequest`、`OptionExerciseCancelRequest` 及对应结果类型。

## [0.3.8] - 2026-06-08

### Added

- **`QuoteClient.GetAddonEntitlement`**：新增附加套餐权益查询接口（wire 方法 `addon_entitlements`），对齐 Java SDK `AddonEntitlementRequest`，无入参，返回 `*model.AddonEntitlement`。
- **`model.AddonEntitlement` 响应模型**：包含 `UserLevel`、当前生效套餐 `ActivePlan`、附加套餐列表 `Addons`（`AddonInfo`）及生效权益额度明细 `EffectiveEntitlement`（`AddonEntitlementDetail`，含历史行情/订阅/深度/频率等各项 limit 与 remaining 字段）。

## [0.3.7] - 2026-05-26

### Added

- **`WithSecretKey` 配置选项**：新增 `config.WithSecretKey(key)` 选项，支持从代码、properties 文件（`secret_key`）或环境变量（`TIGEROPEN_SECRET_KEY`）注入机构账户鉴权 Secret Key。
- **`NewTradeClientFromConfig` 构造函数**：从 `ClientConfig` 自动注入 `Account` 和 `SecretKey` 到 `TradeClient`，无需手动设置。
- **`TradeClient.SetSecretKey`**：运行时动态更新 `SecretKey` 的 setter 方法。
- **SecretKey 自动注入**：所有交易接口（下单、改单、预览、订单查询、持仓、资产、账户管理、资金明细、资产分析、外汇、资金调拨、转股等）在调用时自动将 `TradeClient` 中的 `secretKey` 注入请求参数，与 Python SDK 行为对齐。
- **`OrderRequest.SecretKey`**：`model.OrderRequest` 新增 `SecretKey` 字段（`json:"secret_key,omitempty"`），用于下单/改单/预览订单接口。

## [0.3.6] - 2026-05-25

### Added

- **`HttpClient.Close()`**：新增关闭方法，停止 `NewHttpClient` 自动启动的后台 token 刷新 goroutine，避免长期运行服务中的 goroutine 泄漏。
- **`WithTokenLoader` / `WithTokenWriter`**：新增自定义 token 加载和写入回调选项，支持将 token 存储在数据库、KV 等自定义来源；`WithTokenLoader` 在 `NewClientConfig` 初始化时自动调用，用于填充初始 token。
- **`TokenManager` 按需文件持久化**：只有显式调用 `WithTokenFilePath` 后，`SetToken` 才写文件；未配置路径时 `SetToken` 仅更新内存。

## [0.3.5] - 2026-05-19

### Added

- **`Contract3` 方法**：新增 `TradeClient.Contract3(symbol, secType)` 方法，使用 API version 3.0，服务端直接返回单个合约对象（`*model.Contract`），无需 `items` 数组解包。
- **`callIntoVersioned` (TradeClient)**：TradeClient 新增带版本号的通用请求方法，支持指定 API version。

## [0.3.4] - 2026-05-12

### Fixed

- **`FundingHistoryItem` 响应模型修正**：对照 `FundDepositWithdrawDTO` 实际字段重写。字段变更：
  - `ID` 类型 `string` → `int64`
  - `SubmitTime` (`submitTime`) → `CreatedAt` (`createdAt`)
  - `UpdateTime` (`updateTime`) → `UpdatedAt` (`updatedAt`)
  - 移除不存在的 `SegType` 字段
  - 新增 `RefID` (`refId`)、`Type` (`type`)、`TypeDesc` (`typeDesc`)、`BusinessDate` (`businessDate`)、`StatusDesc` (`statusDesc`)、`CompletedStatus` (`completedStatus`)

## [0.3.3] - 2026-05-11

### Fixed

- **`GetFutureTradeTicks` 响应解包修正**：服务端实际返回 `{"contractCode":"...","items":[...]}` 包装对象，而非裸数组；修正为先解包 `items`，再回填 `contractCode` 到每个 tick。
- **`FutureTradeTicksRequest` 字段 omitempty 移除**：`begin_index` 和 `end_index` 去掉 `omitempty`，确保零值（0）能发送到服务端；同时为 `end_index` 默认填 30（与 Python SDK 一致）。
- **`FundingHistory` 反序列化修正**：服务端直接返回裸 list，从错误的 `callIntoItems`（期望 `{items:[...]}`）改为 `callInto`。

## [0.3.2] - 2026-05-11

### Fixed

- **`SegmentFundHistoryItem` 响应模型修正**：服务端实际字段名为 `createdAt` / `updatedAt` / `settledAt`，原代码错误映射为 `submitTime` / `updateTime`；补充 `StatusDesc string`（`statusDesc`）和 `SettledAt int64`（`settledAt`）字段。
- **`SegmentFundAvailable` 返回类型修正**：由错误的 `[]SegmentFund` 改为专用的 `[]SegmentFundAvailableItem`（仅含 `fromSegment`、`currency`、`amount` 三个字段，与服务端实际响应对齐）。
- **`SegmentFund`（transfer/cancel 响应）模型更新**：字段从旧的 `submitTime`/`availableAmount` 映射修正为 `createdAt`/`updatedAt`/`settledAt`；补充 `StatusDesc`、`Message` 字段，`ID` 类型改为 `interface{}` 以兼容服务端可能返回的数字或字符串。

## [0.3.1] - 2026-05-11

### Fixed

- **Push TLS 证书验证**：服务端证书已更新为有效的 `*.tigerfintech.com`，Go SDK push 客户端现默认开启 TLS 证书验证。验证失败时自动降级并打印 warning 日志，不影响连接（向前兼容）。
- **`Transaction` 响应模型修正**：
  - `TransactedAt` 字段类型 `int64` → `string`（服务端返回 `"YYYY-MM-DD HH:MM:SS"` 格式字符串，不是时间戳）
  - 新增 `AccountId int64`、`FilledPrice float64`、`FilledAmount float64`、`FilledQuantityScale int`、`TransactionTime int64`（毫秒时间戳）等服务端实际返回的字段

## [0.3.0] - 2026-05-08

本次发布达到与 Python SDK **100% API 覆盖**。新增 71 个方法，重构 11 个方法签名。包含多处 breaking change。

### Added

**Trade（17 个新方法）**

- `ManagedAccounts(req)` — 查询机构子账户列表（`accounts`）
- `DerivativeContracts(req)` — 衍生品合约列表（`derivative_contracts`）
- `GetOrder(req)` — 按 ID 查询单个订单详情（`orders`）
- `AnalyticsAsset(req)` — 按日资产分析（`analytics_asset`）
- `AggregateAssets(req)` — 综合账户资产汇总（`aggregate_assets`）
- `EstimateTradableQuantity(req)` — 可交易数量估算（`estimate_tradable_quantity`）
- `PlaceForexOrder(req)` — 外汇下单（`place_forex_order`）
- `SegmentFundAvailable(req)` / `SegmentFundHistory(req)` / `TransferSegmentFund(req)` / `CancelSegmentFund(req)` — 子账户资金调拨
- `FundDetails(req)` — 资金流水明细（`fund_details`）
- `FundingHistory(req)` — 资金调拨记录（`transfer_fund`）
- `TransferPosition(req)` — 内部转股（`position_transfer`）
- `PositionTransferRecords(req)` / `PositionTransferDetail(req)` / `PositionTransferExternalRecords(req)` — 转股记录查询

**Quote（45 个新方法）**

- 股票基础：`GetSymbols` / `GetSymbolNames` / `GetTradeMetas` / `GetStockDetails` / `GetStockDelayBriefs`
- K 线/分时/逐笔：`GetBars` / `GetBarsByPage`（客户端分页）/ `GetTimelineHistory` / `GetTradeRank` / `GetShortInterest`
- 股票其他：`GetStockBroker` / `GetStockFundamental` / `GetStockIndustry` / `GetQuotePermission` / `GetKlineQuota`
- 期权扩展：`GetOptionBars` / `GetOptionTradeTicks` / `GetOptionTimeline` / `GetOptionDepth` / `GetOptionSymbols` / `GetOptionAnalysis`
- 期货扩展：`GetFutureContract` / `GetAllFutureContracts` / `GetCurrentFutureContract` / `GetFutureContinuousContracts` / `GetFutureHistoryMainContract` / `GetFutureBars` / `GetFutureBarsByPage` / `GetFutureTradeTicks` / `GetFutureDepth` / `GetFutureTradingTimes`
- 基金：`GetFundSymbols` / `GetFundContracts` / `GetFundQuote` / `GetFundHistoryQuote`
- 窝轮：`GetWarrantBriefs` / `GetWarrantFilter`
- 行业：`GetIndustryList` / `GetIndustryStocks`
- 公司行动：`GetCorporateSplit` / `GetCorporateDividend` / `GetCorporateEarningsCalendar`
- 财务/日历：`GetFinancialExchangeRate` / `GetFinancialCurrency` / `GetTradingCalendar`
- 其他：`GetMarketScannerTags` / `GetQuoteOvernight`

**Push（8 个新方法）**

- `SubscribeStockTop` / `UnsubscribeStockTop` — 股票榜单
- `SubscribeOptionTop` / `UnsubscribeOptionTop` — 期权榜单
- `SubscribeCc` / `UnsubscribeCc` — 数字货币
- `SubscribeMarket` / `UnsubscribeMarket` — 市场状态

**新增枚举（`model/enums.go`）**

- `OrderSortBy` / `SegmentType` / `CorporateActionType` / `IndustryLevel` / `SortDirection` / `OptionAnalysisPeriod` / `FinancialReportPeriod`

**新增 Request struct（~40 个）**

- `OrdersRequest` / `GetOrderRequest` / `OrderTransactionsRequest` / `PositionsRequest` / `AssetsRequest`
- `ManagedAccountsRequest` / `DerivativeContractsRequest` / `AnalyticsAssetRequest` / `AggregateAssetsRequest` / `SegmentFundRequest` / `ForexOrderRequest` / `EstimateTradableQuantityRequest` / `FundingHistoryRequest` / `FundDetailsRequest` / `PositionTransferRequest` / `PositionTransferRecordsRequest` / `PositionTransferDetailRequest`
- `SymbolsRequest` / `TradeMetasRequest` / `StockDetailsRequest` / `StockDelayBriefsRequest` / `BriefRequest` / `DepthQuoteRequest` / `BarsRequest` / `BarsByPageRequest` / `TimelineHistoryRequest` / `TradeTickRequest` / `TradeRankRequest` / `ShortInterestRequest` / `StockBrokerRequest` / `StockFundamentalRequest` / `StockIndustryRequest` / `KlineQuotaRequest` / `QuotePermissionRequest`
- `OptionBarsRequest` / `OptionTradeTicksRequest` / `OptionTimelineRequest` / `OptionDepthRequest` / `OptionSymbolsRequest` / `OptionAnalysisRequest` / `OptionQueryItem`
- `FutureContractSingleRequest` / `AllFutureContractsRequest` / `FutureContinuousContractsRequest` / `FutureHistoryMainContractRequest` / `FutureBarsRequest` / `FutureBarsByPageRequest` / `FutureTradeTicksRequest` / `FutureDepthRequest` / `FutureTradingTimesRequest` / `FutureBriefRequest`
- `FundSymbolsRequest` / `FundContractsRequest` / `FundQuoteRequest` / `FundHistoryQuoteRequest` / `WarrantBriefsRequest` / `WarrantFilterRequest` / `IndustryListRequest` / `IndustryStocksRequest` / `TradingCalendarRequest` / `FinancialExchangeRateRequest` / `FinancialCurrencyRequest` / `MarketScannerTagsRequest` / `QuoteOvernightRequest`

**新增 Response struct（~40 个）**

- Trade: `ManagedAccount` / `AnalyticsAsset` / `AggregateAssets` / `SegmentFund` / `SegmentFundHistoryItem` / `FundDetails` / `FundingHistoryItem` / `EstimateTradableQuantity` / `ForexOrderResult` / `TransferItem` / `PositionTransferRecord` / `PositionTransferDetail` / `PositionTransferExternalRecord`
- Quote: `SymbolItem` / `SymbolName` / `TradeMeta` / `StockDetail` / `ShortInterest` / `StockBroker` / `StockBrokerItem` / `BrokerDetail` / `StockFundamental` / `StockIndustry` / `TradeRankItem` / `KlineQuota` / `KlineQuotaDetail` / `OptionAnalysis` / `OptionVolatilityPoint` / `OptionSymbol` / `FutureMainContractHistory` / `FutureTradingTime` / `FutureTradingSegment` / `FutureTradeTickItem` / `FutureDepth` / `WarrantBrief` / `WarrantFilterResult` / `IndustryItem` / `IndustryStock` / `TradingCalendarItem` / `ExchangeRate` / `FinancialCurrency` / `QuoteOvernight` / `MarketScannerTags` / `MarketScannerTag` / `FundContractInfo` / `FundQuote` / `FundHistoryQuote`

**Push Subject 枚举**

- 新增 `SubjectCc` / `SubjectMarket`

**Proto**

- `OrderStatusData` push message：新增 `updateTime` (字段 44, 订单信息更新毫秒时间戳) 和 `latestTime` (字段 45, 订单状态更新毫秒时间戳)。已重新生成 `push/pb/OrderStatusData.pb.go`。

### Changed (BREAKING)

所有订单/持仓/资产/行情查询接口统一重构为接受 Request struct 参数。旧的位置参数签名被替换。

**Trade：**
- `Orders()` → `Orders(req OrdersRequest)`
- `ActiveOrders()` → `ActiveOrders(req OrdersRequest)`（支持按 `ParentId` 过滤附加订单）
- `InactiveOrders()` → `InactiveOrders(req OrdersRequest)`
- `FilledOrders(startMs, endMs int64)` → `FilledOrders(req OrdersRequest)`
- `OrderTransactions(id int64, symbol, secType string)` → `OrderTransactions(req OrderTransactionsRequest)`（`symbol` / `secType` 不再必填）
- `Positions()` → `Positions(req PositionsRequest)`
- `Assets()` → `Assets(req AssetsRequest)`
- `PrimeAssets()` → `PrimeAssets(req AssetsRequest)`

**Quote：**
- `GetBrief(symbols []string)` → `GetBrief(req BriefRequest)`（wire method 从 `quote_real_time` 改为 `brief`，支持 `include_hour_trading` / `include_ask_bid` / `right` / `lang`）
- `GetTradeTick(symbols []string)` → `GetTradeTick(req TradeTickRequest)`（支持 `begin_index` / `end_index` / `limit`）
- `GetQuoteDepth(symbol, market string)` → `GetQuoteDepth(req DepthQuoteRequest)`（支持多 symbol、`trade_session`）
- `GetFutureRealTimeQuote(contractCodes []string)` → `GetFutureRealTimeQuote(req FutureBriefRequest)`

### 迁移指引

```go
// v0.2.x
tc.Orders()
tc.FilledOrders(startMs, endMs)
tc.OrderTransactions(orderId, "AAPL", "STK")
tc.Positions()
qc.GetBrief([]string{"AAPL"})
qc.GetQuoteDepth("AAPL", "US")

// v0.3.0
tc.Orders(model.OrdersRequest{})
tc.FilledOrders(model.OrdersRequest{StartDate: startMs, EndDate: endMs})
tc.OrderTransactions(model.OrderTransactionsRequest{OrderId: orderId}) // symbol/secType 可选
tc.Positions(model.PositionsRequest{})
qc.GetBrief(model.BriefRequest{Symbols: []string{"AAPL"}})
qc.GetQuoteDepth(model.DepthQuoteRequest{Symbols: []string{"AAPL"}, Market: "US"})
```

### 设计原则

- **Request struct 字段名 = 服务端 wire 真名**，不学 Python 客户端的参数别名。例如 Trade 的时间字段统一用 `StartDate` / `EndDate`（wire `start_date` / `end_date`），Quote bars 用 `BeginTime` / `EndTime`（wire `begin_time` / `end_time`），Fundamental 用 `BeginDate` / `EndDate`（wire `begin_date` / `end_date`）。
- 所有 Request 字段都可选，添加 `omitempty`；Account 留空时自动填充 client 初始化时的默认账户。
- 枚举一律用字符串常量，与 Python `common.consts` 同步。

### Fixed

- `GetOrder(req)` 的 wire method 修正为 `orders`（传 `id` / `order_id`，服务端返回单个 Order 对象，非 `{items:[]}` 包装）。之前错用的 `order_no` 只返回下一个可用 orderId，不是订单详情。
- `Order.Status` 反序列化时自动把服务端可能返回的整数转为字符串：
  - `-2 → Invalid`, `-1 → Initial`, `3 → PendingCancel`, `4 → Cancelled`,
    `5 → Submitted`, `6 → Filled`, `7 → Inactive`, `8 → PendingSubmit`
  - 与 Java SDK `OrderStatus` 枚举对齐（不再引入 Python SDK 的 `PendingNew` / `PartiallyFilled` 等客户端派生状态）。
- `OrderStatus` 枚举重定义,与 Java SDK 完全对齐:`Invalid/Initial/PendingCancel/Cancelled/Submitted/Filled/Inactive/PendingSubmit`(8 个)。移除了旧版的 `PendingNew` / `PartiallyFilled`。新增 `OrderStatus.Code()` 方法返回服务端数字码。
- 推送 dispatcher 补充 `SocketCommon_Cc` case,路由到 `OnQuote` 回调(与 Python SDK 一致)。之前订阅成功但收到数据时会抛 "未知的 DataType" 错误。

## [0.2.0] - 2026-05-06

This release aligns the SDK with the server's wire format (`snake_case` requests, `camelCase` responses)
and wraps every endpoint with strongly-typed request/response structs. Contains breaking changes.

### Added

- **Strongly-typed responses for every endpoint.** All methods now return concrete `model.*` types
  (e.g. `[]model.Brief`, `*model.PrimeAsset`, `*model.PlaceOrderResult`) instead of `json.RawMessage`.
- **Request structs for complex endpoints.** `model.OrderRequest`, `FinancialDailyRequest`,
  `FinancialReportRequest`, `CorporateActionRequest`, `FutureKlineRequest`, `MarketScannerRequest`.
- **New response models** in `model/`:
  - Quote: `MarketState`, `Brief`, `Kline` / `KlineItem`, `Timeline` / `TimelineItem`,
    `TradeTick` / `TradeTickItem`, `Depth` / `DepthLevel`, `OptionExpiration`, `OptionChain`
    / `OptionChainRow` / `OptionLeg`, `FutureExchange`, `FutureContractInfo`, `FutureQuote`,
    `FutureKline` / `FutureKlineItem`, `FinancialDailyItem`, `FinancialReportItem`,
    `CorporateAction`, `CapitalFlow` / `CapitalFlowItem`, `CapitalDistribution`,
    `ScannerResult` / `ScannerResultItem`, `QuotePermission`.
  - Trade: `Asset` / `AssetSegment`, `PrimeAsset` / `PrimeAssetSegment` / `CurrencyAsset`,
    `PreviewResult`, `PlaceOrderResult`, `OrderIDResult`, `Transaction`.
- **`client.UnmarshalData`** helper that transparently handles the server's occasional
  double-encoded JSON payloads (where `data` is a JSON string containing JSON).
- **Per-example subdirectories** under `examples/`:
  - `examples/quote/` — covers all 22 `QuoteClient` methods.
  - `examples/trade/` — covers all 15 `TradeClient` methods, including a real
    place → modify → cancel flow with a deep out-of-the-money limit order.
  - `examples/push/` — relocated from top-level `push_example.go`.

### Changed

- **Request payloads are now `snake_case`** end-to-end, matching the server contract.
  `model.OrderRequest` / `OrderLegRequest` / `AlgoParamsRequest` use `snake_case` JSON tags,
  while the response-side `model.Order` / `OrderLeg` / `AlgoParams` keep `camelCase` tags.
- **Order construction helpers return `OrderRequest`.** `model.MarketOrder` / `LimitOrder`
  / `StopOrder` / `StopLimitOrder` / `TrailOrder` / `AuctionLimitOrder` / `AuctionMarketOrder`
  / `AlgoOrder` / `NewOrderLeg` now produce `OrderRequest` / `OrderLegRequest` / `AlgoParamsRequest`.
- **`TradeClient` method signatures** (breaking):
  - `PlaceOrder(OrderRequest) (*PlaceOrderResult, error)`
  - `PreviewOrder(OrderRequest) (*PreviewResult, error)`
  - `ModifyOrder(id, OrderRequest) (*OrderIDResult, error)`
  - `CancelOrder(id) (*OrderIDResult, error)`
  - `Orders/ActiveOrders/InactiveOrders() ([]Order, error)`
  - `FilledOrders(startMs, endMs int64) ([]Order, error)` — added mandatory time range.
  - `OrderTransactions(id int64, symbol, secType string) ([]Transaction, error)` — added
    mandatory `symbol` + `secType`, switched payload key `id` → `order_id`.
  - `Contract / Contracts() ([]Contract, error)`
  - `QuoteContract(symbol, secType, expiry string) ([]Contract, error)` — added mandatory
    `expiry`, takes underlying symbol rather than an option identifier; supports OPT/WAR/IOPT only.
  - `Positions() ([]Position, error)`, `Assets() ([]Asset, error)`, `PrimeAssets() (*PrimeAsset, error)`
- **`QuoteClient` method signatures** (breaking):
  - `GetQuoteDepth(symbol, market string) ([]Depth, error)` — added mandatory `market`;
    now wraps `symbol` in a `symbols` array as the server requires.
  - `GetFutureContracts(exchange) ([]FutureContractInfo, error)` — changed method to
    `future_contract_by_exchange_code`, field to `exchange_code`.
  - `GetFutureRealTimeQuote(contractCodes) ([]FutureQuote, error)` — field renamed to
    `contract_codes`.
  - `GetFutureKline(FutureKlineRequest) ([]FutureKline, error)` — switched to Request struct;
    `begin_time` / `end_time` are now required (use `-1` for unbounded).
  - `GetFinancialDaily / GetFinancialReport / GetCorporateAction / MarketScanner` take
    their respective `*Request` structs; all now pass the previously-missing required
    server fields (`market`, `fields`, `period_type`, `action_type`, etc.).
  - `GetCapitalFlow(symbol, market, period string) (*CapitalFlow, error)` — added required
    `market` + `period`.
  - `GetCapitalDistribution(symbol, market string) (*CapitalDistribution, error)` — added
    required `market`.
- **Position / Contract / Order structs** extended with previously missing response fields
  (e.g. `PositionQty`, `SalableQty`, `TodayPnl`, `CanModify`, `CanCancel`, `IsOpen`,
  `TradingSessionType`, `PrimaryExchange`, `SupportFractionalShare`).

### Fixed

- `GetBrief` / `GetTimeline` / `GetTradeTick` and several other endpoints that previously
  sent camelCase field names (e.g. `"secType"`) now send `snake_case` (`"sec_type"`) so the
  server no longer rejects them with `biz param error`.
- `place_order` / `modify_order` / `cancel_order` responses sometimes arrive as JSON-encoded
  strings (double-encoded) — the new `client.UnmarshalData` path parses both forms.
- `examples/` directory no longer contains three conflicting `package main` files.

### Removed

- Flat `examples/quote_example.go` / `trade_example.go` / `push_example.go` files (moved into
  subdirectories as described above).

## [0.1.0] - 2026-03

### Added

- Initial Go SDK release: `QuoteClient`, `TradeClient`, `PushClient`,
  RSA-signed HTTP transport, retry policy, domain discovery, and Protobuf/TCP push.
