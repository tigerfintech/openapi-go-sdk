# Changelog

All notable changes to the Tiger Brokers OpenAPI Go SDK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
