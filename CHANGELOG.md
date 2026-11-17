# Changelog

## 0.5.7 (2026-11-18)

### New

- Contract responses expose supported combinations of order types, time in force, and trading sessions through order capabilities.

## 0.5.5 (2026-11-09)

### New

- Asset and position queries support `AssetQuoteType` values `ETH`, `RTH`, and `OVERNIGHT`.
- `SubscribeAssetWithQuoteType` and `SubscribePositionWithQuoteType` preserve the selected quote basis after reconnecting.
- `BuildSubscribeMessageWithAssetQuoteType` supports quote-basis selection.

### Changed

- Existing subscription methods retain their signatures and behavior when no quote basis is specified.

## 0.5.4 (2026-09-02)

### Fixed

- Option quote and K-line queries accept identifiers with integer strike prices.
- Option chains return Greeks and implied volatility by default.

## 0.5.3 (2026-08-27)

### New

- Real-time stock and cryptocurrency quotes provide turnover through `Brief.Amount`.

## 0.5.2 (2026-08-26)

### Fixed

- The published module contains the intended 0.5.1 changes and reports the matching SDK version.

## 0.5.1 (2026-08-19)

### New

- Option quotes provide mark prices and midpoint prices.
- Trade-tick queries support overnight sessions, and push ticks provide readable trade-condition descriptions.
- `WithUseFullTick` enables complete tick messages through `OnFullTick`.
- K-line and timeline responses provide optional fractional volume.
- `GetTimelineByReq` accepts `SecType` for cryptocurrency queries using API 3.0; existing symbol-list calls retain API 2.0 behavior.
- Paginated K-line queries preserve `SecType`.

### Fixed

- Overnight quote fields match the service response.

### Breaking

- `Callbacks.OnTick` receives `PushTradeTick` instead of `*pb.TradeTickData`. Update callback signatures and access decoded fields instead of parsing raw Protobuf.

## 0.5.0 (2026-07-24)

### New

- Order helpers support monetary market and limit orders, trailing-stop prices, multi-leg orders, combination orders, and OCA orders.
- `OrderRequest.ComboType` selects combination-order types.
- `WithTigerPublicKey` configures a custom service public key.

### Changed

- `IsQuantityByAmount` is deprecated; set `Amount` directly for monetary orders.

### Fixed

- Non-positive token-refresh intervals use the default interval instead of causing a panic.

## 0.4.9 (2026-07-23)

### New

- `GetCorporateSymbolChange` queries stock symbol changes.
- `GetCorporateDelisting` queries delisting events.
- `GetCorporateIPO` queries IPO events.
- `CorporateActionType` includes `SymbolChange`, `Delisting`, and `IPO`.

## 0.4.8 (2026-07-22)

### New

- `WithSharedTokenFrom` shares token updates between clients.

### Fixed

- Concurrent logger access and token refresh no longer cause panics, missed updates, or leaked goroutines.
- Quote clients use refreshed tokens.
- Failed response decoding no longer continues with partially decoded data.
- Properties-file continuations no longer include comment lines in values.

## 0.4.7 (2026-07-13)

### New

- Option-chain queries support Greeks, moneyness, implied-volatility, and open-interest filters.
- `RangeFloat64`, `RangeInt`, and `OptionChainFilterGreeks` provide range filters.
- `GetOptionChainByReq` accepts a complete request while existing calls remain available.
- `GetOptionKlineWithOpts` supports limits and sort direction.
- Order requests support time in force, after-hours prices, batch allocation, virtual orders, brackets, and multi-leg orders.
- `ContractLegRequest` represents multi-leg option legs.

### Breaking

- `OptionAnalysisRequest.Symbols` is now `[]OptionAnalysisSymbol`; move the removed top-level `Period` and `RequireVolatilityList` values into each symbol item. For example, replace `OptionAnalysisRequest{Symbols: []string{"AAPL"}, Period: "26week"}` with `OptionAnalysisRequest{Symbols: []OptionAnalysisSymbol{{Symbol: "AAPL", Period: "26week"}}}`.

## 0.4.6 (2026-07-09)

### New

- Option-expiration queries accept an optional market.
- Option-chain, option-quote, and option-K-line queries accept a timezone, defaulting to New York for US options and Hong Kong for HK options.

### Fixed

- Option-expiration timestamps no longer shift by four to eight hours and return empty results.

### Breaking

- `GetOptionKline` requires begin and end timestamps in milliseconds. Replace `GetOptionKline(identifiers, period)` with `GetOptionKline(identifiers, period, -1, -1)` to retain server defaults.
- `model.Brief.Expiry` is now an `int64` millisecond timestamp instead of a string.

## 0.4.5 (2026-07-07)

### Changed

- `GetBrief`, `GetOptionBrief`, `GetStockDelayBriefs`, and `GetWarrantBriefs` are deprecated. Use `GetRealTimeQuote`, `GetOptionQuote`, `GetDelayedQuote`, and `GetWarrantQuote`, respectively.

## 0.4.4 (2026-07-07)

### New

- `NewQuoteClientFromConfig` constructs a quote client directly from configuration.
- Request logging reports success, business errors, retries, and failures, including uncertain order-submission outcomes.
- `WithLogger` accepts a custom logger; `logger.NopLogger` disables SDK logs.
- `IsStaleConnectionError` identifies stale-connection failures.

### Changed

- `GetBrief`, `GetOptionBrief`, `GetStockDelayBriefs`, and `GetWarrantBriefs` are deprecated in favor of `GetRealTimeQuote`, `GetOptionQuote`, `GetDelayedQuote`, and `GetWarrantQuote`.
- HTTP connections use a 60-second idle timeout, up to 10 idle connections per host, and 10-second connection and TLS-handshake timeouts.

### Fixed

- Non-success HTTP responses preserve business errors or include the HTTP status when the body cannot be decoded.
- Add-on entitlement levels accept numeric or string values.
- Fund-detail IDs accept numeric or string values.

### Breaking

- K-line methods accept symbol slices and request structs. Replace positional arguments with the corresponding request types and rename `Bars` methods and types to `Kline`.
- `GetOptionBars` is removed; use `GetOptionKline`.
- `NewTradeClientFromConfig` no longer takes a manually created HTTP client.

## 0.4.2 (2026-07-01)

### New

- `WithServerURL` and properties-file URL settings override default service addresses.
- `HttpClient.SecretKey` exposes the configured institutional secret key.

### Fixed

- `NewTradeClient` automatically includes the configured secret key in institutional requests.
- JSON serialization of configuration no longer exposes private keys or institutional secret keys.
- SDK version constants match the release.

## 0.4.0 (2026-06-24)

### New

- `IcebergOrder` and `IcebergOrderFull` create iceberg orders with display size, minimum display size, check intervals, price type, and optional execution times.
- `IcebergPriceType` supports fixed-limit and opponent prices.
- Order models include iceberg-order fields.

### Fixed

- Order cancellation includes a configured secret key and omits an empty one.

## 0.3.9 (2026-06-09)

### New

- Option-exercise APIs support checks, eligible-position queries, submission, records, and cancellation with typed requests and results.
- Exercise requests inherit the account and institutional secret key.

## 0.3.8 (2026-06-08)

### New

- `GetAddonEntitlement` queries active plans, add-ons, and effective quotas for history, subscriptions, depth, and request frequency.

## 0.3.7 (2026-05-26)

### New

- `WithSecretKey` configures institutional authentication from code, properties, or `TIGEROPEN_SECRET_KEY`.
- `NewTradeClientFromConfig` inherits the configured account and secret key.
- `TradeClient.SetSecretKey` updates the secret key at runtime.
- Trading requests automatically include the client's secret key, and order requests expose a secret-key field.

## 0.3.6 (2026-05-25)

### New

- `HttpClient.Close` stops background token refresh.
- Custom token loader and writer callbacks support external storage.
- Token persistence writes files only when a token-file path is configured.

## 0.3.5 (2026-05-19)

### New

- `TradeClient.Contract3` queries a single contract through API version 3.0.
- Trading requests can select an API version.

## 0.3.4 (2026-05-12)

### Fixed

- Funding-history responses correctly expose refund references, type descriptions, business dates, and related fields.

## 0.3.3 (2026-05-11)

### Fixed

- Futures trade ticks decode correctly and accept an explicit zero begin index.
- Funding-history responses decode correctly.

## 0.3.2 (2026-05-11)

### Fixed

- Segment-fund history, available funds, transfer, and cancellation responses decode the correct fields and result types.

## 0.3.1 (2026-05-11)

### Changed

- Push connections verify TLS certificates by default, with a warning and fallback when verification fails.

### Fixed

- Transactions decode the correct field types and include additional execution information.

## 0.3.0 (2026-05-08)

### New

- Trading APIs add managed accounts, derivative contracts, individual orders, asset analytics and aggregation, tradable quantities, forex orders, funding, and position transfers.
- Quote APIs add stock data, paginated bars, historical timelines and ticks, extended option and futures data, funds, warrants, industries, corporate actions, currencies, exchange rates, and calendars.
- Push subscriptions support stock rankings, option rankings, cryptocurrency, and market status.
- Added corresponding request and response models and enums.
- Order push messages provide order-information and status-update timestamps.

### Fixed

- `GetOrder` returns order details instead of allocating an order number.
- Numeric order statuses convert to readable names, and `OrderStatus.Code` exposes the server code.
- Cryptocurrency push messages reach the correct callbacks.

### Breaking

- Order, position, asset, and quote queries require request structs instead of positional arguments. Update calls as shown below.

```go
// Before
tc.Orders()
tc.FilledOrders(startMs, endMs)
tc.OrderTransactions(orderId, "AAPL", "STK")
tc.Positions()
qc.GetBrief([]string{"AAPL"})
qc.GetQuoteDepth("AAPL", "US")

// After
tc.Orders(model.OrdersRequest{})
tc.FilledOrders(model.OrdersRequest{StartDate: startMs, EndDate: endMs})
tc.OrderTransactions(model.OrderTransactionsRequest{OrderId: orderId})
tc.Positions(model.PositionsRequest{})
qc.GetBrief(model.BriefRequest{Symbols: []string{"AAPL"}})
qc.GetQuoteDepth(model.DepthQuoteRequest{Symbols: []string{"AAPL"}, Market: "US"})
```

## 0.2.0 (2026-05-06)

### New

- All endpoints provide typed request and response models.
- `client.UnmarshalData` supports double-encoded JSON responses.
- Per-client examples cover quotes, trading, push, and the order placement, modification, and cancellation flow.
- Position, contract, and order models expose additional response fields.

### Changed

- Order construction helpers create the new request types.

### Fixed

- Quote requests use the field names required by the service.
- Order placement, modification, and cancellation decode JSON-encoded response strings.

### Breaking

- Methods return concrete model types instead of `json.RawMessage`.
- Trading methods require request structs; some queries also require a time range or symbol and security type.
- Several quote methods require request structs or previously missing mandatory fields such as market.
- Requests use snake_case field names, while response models retain camelCase JSON names.

## 0.1.0 (2026-03)

### New

- Initial release with quote, trading, and push clients, RSA-signed requests, retries, domain discovery, and Protobuf push connections.
