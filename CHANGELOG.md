# Changelog

All notable changes to the Tiger Brokers OpenAPI Go SDK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- `Brief.Amount` 实时行情字段，支持股票和数字货币成交额。

## [0.5.2] - 2026-08-26

### Fixed
- Re-publish the 0.5.1 release content under a new tag because the public `v0.5.1` tag pointed to an older code snapshot.
- Synchronize SDK version constants with the released Go module version.

## [0.5.1] - 2026-08-19

### Breaking
- `Callbacks.OnTick` 回调参数类型由原始 `*pb.TradeTickData` 改为解码后的 `PushTradeTick`，每笔成交以结构化字段返回。迁移：将 `OnTick` 回调签名中的 `*pb.TradeTickData` 改为 `PushTradeTick`，通过字段访问替代原始 protobuf 解析。

### Added
- `OptionLeg` 补充标记价格和中间价相关字段
- `TradeTickRequest.TradeSession`，支持通过 `GetTradeTick` 查询美股夜盘逐笔成交
- `WithUseFullTick`，支持启用并分发完整逐笔推送到 `OnFullTick`
- `PushTick.Cond` 字段，原始单字符代码已转换为可读字符串（如 `US_REGULAR_SALE`、`HK_AUTOMATCH_NORMAL`）
- `KlineItem` 和 `TimelineItem` 新增可空的 `VolumeDecimal` 字段，支持数字货币小数成交量
- `GetTimelineByReq` 接受包含 `SecType` 的 `TimelineRequest`（数字货币使用 `"CC"`）并调用 3.0 版本 API；现有 `GetTimeline(symbols []string)` 保持 v2 行为
- `KlineByPageRequest` 新增 `SecType` 字段，并在所有分页 K 线请求中传递该字段

### Fixed
- 修正 `QuoteOvernight` 字段，使其与隔夜行情服务响应一致

## [0.5.0] - 2026-07-24

### Added
- 订单工具函数：新增 `MarketOrderByAmount`、`LimitOrderByAmount`、`TrailOrderByPrice`、`LimitOrderWithLegs`、`ComboOrder`、`OcaOrder`、`NewContractLeg`
- `OrderRequest` 新增 `ComboType` 字段
- `WithTigerPublicKey` 配置选项，支持自定义服务端公钥

### Fixed
- `TokenManager.refreshLoopWith`：interval ≤ 0 时回退到默认值，防止 panic

### Changed
- `IsQuantityByAmount` 标记为废弃（by-amount 订单直接设置 `Amount`）

## [0.4.9] - 2026-07-23

### Added
- `GetCorporateSymbolChange` — 股票代码变更查询
- `GetCorporateDelisting` — 退市事件查询
- `GetCorporateIPO` — 新股上市查询
- `CorporateActionType`：新增 `SymbolChange`、`Delisting`、`IPO`

## [0.4.8] - 2026-07-22
### Fixed
- 修复全局 logger 并发读写可能 panic 的问题。
- 修复后台 token 自动刷新在并发场景下的竞态问题，避免极端情况下 token 更新不可见或 goroutine 泄漏。
- 修复 quote client 在 token 刷新后仍使用旧 token 发起请求的问题；新增 `WithSharedTokenFrom(primary)` 选项，支持与其他 client 共享 token。
- 修复响应 data 反序列化失败时可能对已部分写入的目标对象继续操作的问题。
- 修复 properties 配置文件续行（`\`）解析错误：注释行不再被误拼入值。

## [0.4.7] - 2026-07-13

### Breaking Changes

- **`OptionAnalysisRequest.Symbols` 类型变更**：从 `[]string` 改为 `[]OptionAnalysisSymbol`，顶层 `Period` 和 `RequireVolatilityList` 字段已删除。迁移：`OptionAnalysisRequest{Symbols: []string{"AAPL"}, Period: "26week"}` → `OptionAnalysisRequest{Symbols: []OptionAnalysisSymbol{{Symbol: "AAPL", Period: "26week"}}}`。

### Added

- **`OptionChainRequest`**：新增 `ReturnGreekValue` 和 `OptionFilter` 字段，支持按价内外、隐含波动率、持仓量、Greeks 范围过滤期权链。
- **`RangeFloat64` / `RangeInt`**：新增范围类型，序列化为 `{"min":...,"max":...}` 嵌套对象（对应服务端 Range 格式）。
- **`OptionChainFilterGreeks`**：delta / gamma / vega / theta / rho 各字段支持范围过滤。
- **`GetOptionChainByReq`**：新增接受完整 `OptionChainRequest` 的方法；原 `GetOptionChain` 保持不变（向前兼容）。
- **`GetOptionKlineWithOpts`**：新增 `limit` 和 `sort_dir` 参数；`GetOptionKline` 委托至此方法（向前兼容）。
- **`OrderRequest`**：补齐下单请求缺失字段，支持有效期、盘后价格、批量分配、虚拟订单、止盈止损单及多腿组合单等场景。
- **`ContractLegRequest`**：新增多腿期权子腿结构体（MLEG）。

## [0.4.6] - 2026-07-09

### Breaking Changes

- **`GetOptionKline` 签名变更**：新增必填参数 `beginTime, endTime int64`（毫秒时间戳，传 `-1` 为服务端默认值）。旧调用 `GetOptionKline(identifiers, period)` 需改为 `GetOptionKline(identifiers, period, -1, -1)`。
- **`model.Brief.Expiry` 类型变更**：`string → int64`（服务端返回毫秒时间戳数字，原 string 类型会导致 JSON 反序列化失败）。

### Added

- **`GetOptionExpiration` 新增可选 `market` 参数**：`GetOptionExpiration(symbols []string, market ...string)`，查询 HK 期权时传 `"HK"`。
- **`GetOptionChain`/`GetOptionQuote`/`GetOptionKline` 新增可选 `timezone` 参数**：`timezone ...string`，向后兼容；US 期权默认 `America/New_York`，HK 期权（`.HK` 后缀）默认 `Asia/Hong_Kong`。

### Fixed

- **期权到期日时区错误**：修复 US/HK 期权到期日时间戳偏差 4–8 小时导致服务端可能返回空数据的问题。

## [0.4.5] - 2026-07-07

### Deprecated

- **`GetBrief` → `GetRealTimeQuote`**：方法名与 wire method `quote_real_time` 不一致，现以 `GetRealTimeQuote` 为主，旧名保留并标记为 Deprecated。
- **`GetOptionBrief` → `GetOptionQuote`**：wire method 为 `option_brief`，更名为 `GetOptionQuote`。
- **`GetStockDelayBriefs` → `GetDelayedQuote`**：wire method 为 `quote_delay`，更名与之对齐。
- **`GetWarrantBriefs` → `GetWarrantQuote`**：wire method 为 `warrant_briefs`，更名以保持接口层风格一致。

## [0.4.4] - 2026-07-07

### Breaking Changes

- **K 线相关接口统一重构**：行情查询接口改为支持多 symbol 传入，并统一采用 Request 结构体参数；`Bars` 系列方法与类型统一重命名为 `Kline`（如 `GetBars` → `GetKline`、`BarsRequest` → `KlineRequest`），删除与 `GetOptionKline` 重复的 `GetOptionBars`。迁移：调用方需将单 symbol 参数改为 slice，将位置参数改为对应的 `*Request` 结构体，并将方法/类型名中的 `Bars` 替换为 `Kline`。`NewTradeClientFromConfig` 不再需要手动传入 `HttpClient`。

### Added

- **`NewQuoteClientFromConfig(cfg)`**：直接从 `ClientConfig` 创建 `QuoteClient`，内部自动使用 `NewQuoteHttpClient(cfg)`，无需调用方手动构造 `HttpClient`。

### Deprecated

- **`GetBrief` → `GetRealTimeQuote`**：方法名与 wire method `quote_real_time` 不一致，现以 `GetRealTimeQuote` 为主，旧名保留并标记为 Deprecated。
- **`GetOptionBrief` → `GetOptionQuote`**：wire method 为 `option_brief`，语义上表示期权实时行情，更名为 `GetOptionQuote`。
- **`GetStockDelayBriefs` → `GetDelayedQuote`**：wire method 为 `quote_delay`，更名与之对齐。
- **`GetWarrantBriefs` → `GetWarrantQuote`**：wire method 为 `warrant_briefs`，更名以保持接口层风格一致。



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

- **`NewTradeClient` 未注入 `SecretKey`**：修复机构账户使用该构造函数时 `CancelOrder` 等接口缺失 `secret_key` 的问题，现自动注入，与 `NewTradeClientFromConfig` 行为一致。
- **凭据 JSON 序列化泄露**：修复日志、调试 dump、错误上报等场景下 `ClientConfig` 可能明文输出私钥与机构密钥的问题。
- **`Version` 常量未同步**：修正版本常量未随发布更新的问题。

## [0.4.0] - 2026-06-24

### Added

- **冰山单辅助函数**：新增 `IcebergOrder(account, symbol, secType, action, quantity, price, displaySize)` 及完整参数版本 `IcebergOrderFull`，支持 `MinDisplaySize`、`CheckIntervals`、`PriceType`（`IcebergPriceTypeLimit`/`IcebergPriceTypeOpponent`）、`StartTime`/`EndTime`（epoch ms）
- **`IcebergPriceType` 枚举**：新增 `IcebergPriceTypeLimit`（固定限价）和 `IcebergPriceTypeOpponent`（对手价）
- **`Order` 结构体新增冰山单字段**：`DisplaySize`、`MinDisplaySize`、`CheckIntervals`、`PriceType`、`StartTime`、`EndTime`

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

- **`FundingHistoryItem` 响应模型修正**：修复资金调拨记录字段与服务端实际返回不一致导致解析错误的问题，补充退款单号、类型说明、业务日期等字段。

## [0.3.3] - 2026-05-11

### Fixed

- **`GetFutureTradeTicks` 响应解包修正**：修复响应解包错误导致取不到数据的问题。
- **`FutureTradeTicksRequest` 字段默认值修正**：修复 `begin_index` 为 0 时未正确发送到服务端的问题。
- **`FundingHistory` 反序列化修正**：修复响应解析失败的问题。

## [0.3.2] - 2026-05-11

### Fixed

- **`SegmentFundHistoryItem` 响应模型修正**：修复字段与服务端实际返回不一致导致解析错误的问题。
- **`SegmentFundAvailable` 返回类型修正**：修复返回类型与服务端实际响应不匹配的问题。
- **`SegmentFund`（transfer/cancel 响应）模型更新**：修复字段映射错误导致的解析问题。

## [0.3.1] - 2026-05-11

### Fixed

- **Push TLS 证书验证**：Go SDK push 客户端现默认开启 TLS 证书验证，验证失败时自动降级并打印 warning 日志，不影响连接。
- **`Transaction` 响应模型修正**：修复字段类型与服务端实际返回不一致导致的解析错误，并补充成交相关字段。

## [0.3.0] - 2026-05-08

本次发布达到与 Python SDK **100% API 覆盖**。新增 71 个方法，重构 11 个方法签名。包含多处 breaking change。

### Added

- **交易接口扩展**：新增机构子账户查询、衍生品合约列表、按 ID 查询订单详情、资产分析与汇总、可交易数量估算、外汇下单、子账户资金调拨、资金流水明细、内部转股及转股记录查询等接口。
- **行情接口扩展**：新增股票基础信息、K 线分页、分时/逐笔历史、期权扩展行情（逐笔、分时、深度、代码列表、分析）、期货扩展行情（合约、连续合约、历史主力合约、K 线分页、深度、交易时间）、基金行情、窝轮行情、行业分类、公司行动（拆股、分红、财报日历）、财务汇率与货币、交易日历等接口。
- **推送能力扩展**：新增股票榜单、期权榜单、数字货币、市场状态的订阅与取消订阅。
- **新增枚举与请求/响应模型**：为上述新增接口配套新增枚举类型及请求、响应结构体。
- **订单推送消息补充订单信息更新时间和状态更新时间字段**。

### Changed (BREAKING)

所有订单/持仓/资产/行情查询接口统一重构为接受 Request struct 参数，旧的位置参数签名被替换：

- **Trade**：`Orders` / `ActiveOrders` / `InactiveOrders` / `FilledOrders` / `OrderTransactions` / `Positions` / `Assets` / `PrimeAssets` 改为接受对应 Request struct，并支持更多可选过滤条件（如按父订单过滤、可选 symbol/secType）。
- **Quote**：`GetBrief` / `GetTradeTick` / `GetQuoteDepth` / `GetFutureRealTimeQuote` 改为接受对应 Request struct，并支持更多可选参数（如盘前盘后、买卖盘、多 symbol）。

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

### Fixed

- `GetOrder(req)` 修复按 ID 查询订单详情时误用错误 wire method 导致返回下一个可用订单号而非订单详情的问题。
- `Order.Status` 反序列化时自动把服务端可能返回的整数转为可读字符串状态。
- `OrderStatus` 枚举重新定义，与 Java SDK 完全对齐；新增 `OrderStatus.Code()` 方法返回服务端数字码。
- 修复推送 dispatcher 缺少数字货币路由导致订阅成功后收到数据报错的问题。

## [0.2.0] - 2026-05-06

This release aligns the SDK with the server's wire format (`snake_case` requests, `camelCase` responses)
and wraps every endpoint with strongly-typed request/response structs. Contains breaking changes.

### Added

- **Strongly-typed request/response models for every endpoint.** All methods now return concrete
  `model.*` types instead of `json.RawMessage`, covering quote data (market state, brief, kline,
  timeline, trade ticks, depth, options, futures, financials, corporate actions, scanner results)
  and trade data (assets, orders, transactions).
- **`client.UnmarshalData`** helper that transparently handles the server's occasional
  double-encoded JSON payloads (where `data` is a JSON string containing JSON).
- **Reorganized examples** into per-client subdirectories (`examples/quote/`, `examples/trade/`,
  `examples/push/`) covering all `QuoteClient` and `TradeClient` methods, including a real
  place → modify → cancel order flow.

### Changed

- **Request payloads are now `snake_case`** end-to-end, matching the server contract, while
  response-side models keep `camelCase` tags.
- **Order construction helpers** (`model.MarketOrder` / `LimitOrder` / `StopOrder` /
  `StopLimitOrder` / `TrailOrder` / `AuctionLimitOrder` / `AuctionMarketOrder` / `AlgoOrder` /
  `NewOrderLeg`) now produce request structs matching the new signatures below.
- **`TradeClient` method signatures** (breaking): order placement, modification, cancellation,
  and order/position/asset queries now take request structs; several queries that previously
  accepted no filters now require a time range or symbol/secType.
- **`QuoteClient` method signatures** (breaking): several endpoints (`GetQuoteDepth`,
  `GetFutureContracts`, `GetFutureRealTimeQuote`, `GetFutureKline`, `GetFinancialDaily`,
  `GetFinancialReport`, `GetCorporateAction`, `MarketScanner`, `GetCapitalFlow`,
  `GetCapitalDistribution`) now require previously-missing server-mandated fields such as
  `market` or a request struct.
- **Position / Contract / Order structs** extended with previously missing response fields.

### Fixed

- Several quote endpoints that previously sent camelCase field names now send `snake_case`,
  so the server no longer rejects requests with `biz param error`.
- `place_order` / `modify_order` / `cancel_order` responses that sometimes arrive as
  JSON-encoded strings (double-encoded) are now parsed correctly.

### Removed

- Flat example files consolidated into per-client subdirectories (see Added above).

## [0.1.0] - 2026-03

### Added

- Initial Go SDK release: `QuoteClient`, `TradeClient`, `PushClient`,
  RSA-signed HTTP transport, retry policy, domain discovery, and Protobuf/TCP push.
