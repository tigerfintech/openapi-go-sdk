# Tiger OpenAPI Go SDK

Go SDK for Tiger Brokers OpenAPI, providing market data queries, order placement, account management, and real-time push notifications.

[![Go Reference](https://pkg.go.dev/badge/github.com/tigerfintech/openapi-go-sdk.svg)](https://pkg.go.dev/github.com/tigerfintech/openapi-go-sdk)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

> Current version: **0.2.0** — see [CHANGELOG.md](./CHANGELOG.md).

## Features

- Market data queries (quotes, K-line, depth, options, futures)
- Order placement and management (limit/market orders, modify, cancel)
- Account management (assets, positions, order history)
- Real-time push via TCP + TLS + Protobuf (quotes, trades, account updates)
- Auto-reconnect with exponential backoff
- Response signature verification using Tiger public key (`tigerPublicKey`)
- Config auto-discovery from properties files

## Installation

```bash
go get github.com/tigerfintech/openapi-go-sdk
```

Requires Go 1.20 or later.

## Quick Start

Place a `tiger_openapi_config.properties` file in the current directory or `~/.tigeropen/tiger_openapi_config.properties`, then use auto-discovery:

```go
package main

import (
	"fmt"
	"log"

	"github.com/tigerfintech/openapi-go-sdk/client"
	"github.com/tigerfintech/openapi-go-sdk/config"
	"github.com/tigerfintech/openapi-go-sdk/quote"
)

func main() {
	// Auto-discovers config from ./tiger_openapi_config.properties
	// or ~/.tigeropen/tiger_openapi_config.properties
	cfg, err := config.NewClientConfig()
	if err != nil {
		log.Fatal(err)
	}

	httpClient := client.NewHttpClient(cfg)
	qc := quote.NewQuoteClient(httpClient)

	states, err := qc.MarketState("US")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("US market state:", string(states))
}
```

## Configuration

The SDK supports four configuration methods. Priority (highest → lowest): **environment variables > code options > auto-discovered properties file > defaults**.

### Method 1: Properties File (Explicit Path)

```go
cfg, err := config.NewClientConfig(
	config.WithPropertiesFile("path/to/tiger_openapi_config.properties"),
)
```

Properties file format:

```properties
tiger_id=your_developer_id
private_key=your_RSA_private_key
account=your_trading_account
license=TBNZ
language=en_US
```

### Method 2: Auto-Discovery

Call `NewClientConfig()` with no arguments. The SDK searches for a properties file in order:

1. `./tiger_openapi_config.properties` (current working directory)
2. `~/.tigeropen/tiger_openapi_config.properties` (home directory)

```go
cfg, err := config.NewClientConfig()
```

The first file found is loaded. Fields already set by explicit options are not overwritten.

### Method 3: Code Options

```go
cfg, err := config.NewClientConfig(
	config.WithTigerID("your_tiger_id"),
	config.WithPrivateKey("your_RSA_private_key"),
	config.WithAccount("your_trading_account"),
	config.WithLanguage("en_US"),
)
```

You can combine code options with a properties file. Explicit code options take precedence over values in the file:

```go
cfg, err := config.NewClientConfig(
	config.WithPropertiesFile("config.properties"),
	config.WithAccount("override_account"),
)
```

### Method 4: Environment Variables

These have the highest priority and override all other methods:

```bash
export TIGEROPEN_TIGER_ID=your_developer_id
export TIGEROPEN_PRIVATE_KEY=your_RSA_private_key
export TIGEROPEN_ACCOUNT=your_trading_account
```

### Configuration Reference

| Field | Env Variable | Description | Required | Default |
|-------|-------------|-------------|----------|---------|
| `tiger_id` | `TIGEROPEN_TIGER_ID` | Developer ID | Yes | — |
| `private_key` | `TIGEROPEN_PRIVATE_KEY` | RSA private key | Yes | — |
| `account` | `TIGEROPEN_ACCOUNT` | Trading account | No | — |
| `tigerPublicKey` | — | Tiger RSA public key for response signature verification | No | Auto-set |
| `language` | — | Language (`zh_CN` / `zh_TW` / `en_US`) | No | `zh_CN` |
| `timeout` | — | HTTP request timeout | No | `15s` |

## Market Data

```go
httpClient := client.NewHttpClient(cfg)
qc := quote.NewQuoteClient(httpClient)

// Market state
states, err := qc.MarketState("US")

// Real-time quotes
quotes, err := qc.QuoteRealTime([]string{"AAPL", "TSLA"})

// K-line data
klines, err := qc.Kline("AAPL", "day")

// Timeline data
timeline, err := qc.Timeline([]string{"AAPL"})

// Depth quotes
depth, err := qc.QuoteDepth("AAPL")

// Option expiration dates
expiry, err := qc.OptionExpiration("AAPL")

// Option chain
chain, err := qc.OptionChain("AAPL", "2024-01-19")

// Futures exchange list
exchanges, err := qc.FutureExchange()
```

## Trading

```go
tc := trade.NewTradeClient(httpClient, cfg.Account)

// Place a limit order
order := model.Order{
	Symbol:        "AAPL",
	SecType:       "STK",
	Action:        "BUY",
	OrderType:     "LMT",
	TotalQuantity: 100,
	LimitPrice:    150.0,
	TimeInForce:   "DAY",
}
result, err := tc.PlaceOrder(order)

// Preview order (no actual execution)
preview, err := tc.PreviewOrder(order)

// Modify order
order.LimitPrice = 155.0
result, err = tc.ModifyOrder(orderId, order)

// Cancel order
result, err = tc.CancelOrder(orderId)

// Query orders
orders, err := tc.Orders()

// Query positions
positions, err := tc.Positions()

// Query assets
assets, err := tc.Assets()
```

## Raw API Access

When the SDK has not yet wrapped a specific API, use `ExecuteRaw` to call it directly:

```go
httpClient := client.NewHttpClient(cfg)

resp, err := httpClient.ExecuteRaw("market_state", `{"market":"US"}`)
if err != nil {
	log.Fatal(err)
}
fmt.Println("Raw response:", resp)
```

## Real-Time Push

The push client connects to the Tiger push server over **TCP + TLS** and uses a varint32-framed **Protobuf** binary protocol. It supports auto-reconnect with exponential backoff and heartbeat keep-alive.

Callback data types are Protobuf generated structs from the `push/pb` package (e.g. `*pb.QuoteData`, `*pb.OrderStatusData`, `*pb.AssetData`, `*pb.PositionData`).

```go
import (
	"fmt"
	"log"

	"github.com/tigerfintech/openapi-go-sdk/config"
	"github.com/tigerfintech/openapi-go-sdk/push"
	"github.com/tigerfintech/openapi-go-sdk/push/pb"
)

cfg, err := config.NewClientConfig()
if err != nil {
	log.Fatal(err)
}

pc := push.NewPushClient(cfg)

pc.SetCallbacks(push.Callbacks{
	OnQuote: func(data *pb.QuoteData) {
		fmt.Printf("Quote: %s price=%.2f volume=%d\n",
			data.GetSymbol(), data.GetLatestPrice(), data.GetVolume())
	},
	OnTick: func(data *pb.TradeTickData) {
		fmt.Println("Trade tick received")
	},
	OnDepth: func(data *pb.QuoteDepthData) {
		fmt.Println("Depth update received")
	},
	OnKline: func(data *pb.KlineData) {
		fmt.Println("Kline update received")
	},
	OnOrder: func(data *pb.OrderStatusData) {
		fmt.Printf("Order: %s status=%s\n", data.GetSymbol(), data.GetStatus())
	},
	OnAsset: func(data *pb.AssetData) {
		fmt.Printf("Asset: account=%s\n", data.GetAccount())
	},
	OnPosition: func(data *pb.PositionData) {
		fmt.Printf("Position: %s\n", data.GetSymbol())
	},
	OnTransaction: func(data *pb.OrderTransactionData) {
		fmt.Println("Transaction update received")
	},
	OnConnect:    func() { fmt.Println("Push connected") },
	OnDisconnect: func() { fmt.Println("Push disconnected") },
	OnError:      func(err error) { fmt.Println("Push error:", err) },
	OnKickout:    func(msg string) { fmt.Println("Kicked out:", msg) },
})

if err := pc.Connect(); err != nil {
	log.Fatal(err)
}
defer pc.Disconnect()

// Subscribe to quote updates
pc.SubscribeQuote([]string{"AAPL", "TSLA"})

// Subscribe to account updates (empty string uses the account from config)
pc.SubscribeAsset("")
pc.SubscribeOrder("")
pc.SubscribePosition("")
```

### Available Subscriptions

| Method | Callback | Data Type |
|--------|----------|-----------|
| `SubscribeQuote` | `OnQuote` | `*pb.QuoteData` |
| `SubscribeTick` | `OnTick` | `*pb.TradeTickData` |
| `SubscribeDepth` | `OnDepth` | `*pb.QuoteDepthData` |
| `SubscribeOption` | `OnOption` | `*pb.QuoteData` |
| `SubscribeFuture` | `OnFuture` | `*pb.QuoteData` |
| `SubscribeKline` | `OnKline` | `*pb.KlineData` |
| `SubscribeAsset` | `OnAsset` | `*pb.AssetData` |
| `SubscribePosition` | `OnPosition` | `*pb.PositionData` |
| `SubscribeOrder` | `OnOrder` | `*pb.OrderStatusData` |
| `SubscribeTransaction` | `OnTransaction` | `*pb.OrderTransactionData` |

## Project Structure

```
openapi-go-sdk/
├── config/    # Configuration management (ClientConfig, properties parser, dynamic domains)
├── signer/    # RSA signing and verification
├── client/    # HTTP client (request/response, retry, ExecuteRaw)
├── model/     # Data models (Order, Contract, Position, enums)
├── quote/     # Market data query client
├── trade/     # Trading client
├── push/      # TCP+TLS push client (Protobuf)
│   └── pb/    # Generated Protobuf types
├── logger/    # Logging module
└── examples/  # Example code
```

## API Reference

- [Tiger OpenAPI Documentation](https://quant.itigerup.com/openapi/zh/python/overview/introduction.html)
- [Go SDK pkg.go.dev](https://pkg.go.dev/github.com/tigerfintech/openapi-go-sdk)

## License

[MIT License](LICENSE)

---

# Tiger OpenAPI Go SDK（中文）

Tiger 证券 OpenAPI Go SDK，提供行情查询、订单交易、账户管理和实时推送功能。

[![Go Reference](https://pkg.go.dev/badge/github.com/tigerfintech/openapi-go-sdk.svg)](https://pkg.go.dev/github.com/tigerfintech/openapi-go-sdk)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

> 当前版本：**0.2.0**，详见 [CHANGELOG.md](./CHANGELOG.md)。

## 功能特性

- 行情查询（报价、K 线、深度、期权、期货）
- 订单交易（限价/市价单、修改、撤单）
- 账户管理（资产、持仓、订单历史）
- 实时推送：TCP + TLS + Protobuf（行情、成交、账户变动）
- 自动重连（指数退避）
- 响应签名验证（使用 Tiger 公钥 `tigerPublicKey`）
- 配置文件自动发现

## 安装

```bash
go get github.com/tigerfintech/openapi-go-sdk
```

需要 Go 1.20 或更高版本。

## 快速开始

将 `tiger_openapi_config.properties` 文件放在当前目录或 `~/.tigeropen/tiger_openapi_config.properties`，然后使用自动发现：

```go
package main

import (
	"fmt"
	"log"

	"github.com/tigerfintech/openapi-go-sdk/client"
	"github.com/tigerfintech/openapi-go-sdk/config"
	"github.com/tigerfintech/openapi-go-sdk/quote"
)

func main() {
	// 自动从 ./tiger_openapi_config.properties
	// 或 ~/.tigeropen/tiger_openapi_config.properties 加载配置
	cfg, err := config.NewClientConfig()
	if err != nil {
		log.Fatal(err)
	}

	httpClient := client.NewHttpClient(cfg)
	qc := quote.NewQuoteClient(httpClient)

	states, err := qc.MarketState("US")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("美股市场状态:", string(states))
}
```

## 配置

SDK 支持四种配置方式。优先级（从高到低）：**环境变量 > 代码选项 > 自动发现的配置文件 > 默认值**。

### 方式一：配置文件（指定路径）

```go
cfg, err := config.NewClientConfig(
	config.WithPropertiesFile("path/to/tiger_openapi_config.properties"),
)
```

配置文件格式：

```properties
tiger_id=你的开发者ID
private_key=你的RSA私钥
account=你的交易账户
license=TBNZ
language=zh_CN
```

### 方式二：自动发现

不传参数调用 `NewClientConfig()`，SDK 按以下顺序搜索配置文件：

1. `./tiger_openapi_config.properties`（当前工作目录）
2. `~/.tigeropen/tiger_openapi_config.properties`（用户主目录）

```go
cfg, err := config.NewClientConfig()
```

找到的第一个文件会被加载。已通过代码选项设置的字段不会被覆盖。

### 方式三：代码选项

```go
cfg, err := config.NewClientConfig(
	config.WithTigerID("your_tiger_id"),
	config.WithPrivateKey("your_RSA_private_key"),
	config.WithAccount("your_trading_account"),
	config.WithLanguage("zh_CN"),
)
```

可以将代码选项与配置文件组合使用，代码选项优先级高于配置文件中的值：

```go
cfg, err := config.NewClientConfig(
	config.WithPropertiesFile("config.properties"),
	config.WithAccount("override_account"),
)
```

### 方式四：环境变量

环境变量具有最高优先级，会覆盖所有其他配置方式：

```bash
export TIGEROPEN_TIGER_ID=你的开发者ID
export TIGEROPEN_PRIVATE_KEY=你的RSA私钥
export TIGEROPEN_ACCOUNT=你的交易账户
```

### 配置参考

| 字段 | 环境变量 | 说明 | 必填 | 默认值 |
|------|---------|------|------|--------|
| `tiger_id` | `TIGEROPEN_TIGER_ID` | 开发者 ID | 是 | — |
| `private_key` | `TIGEROPEN_PRIVATE_KEY` | RSA 私钥 | 是 | — |
| `account` | `TIGEROPEN_ACCOUNT` | 交易账户 | 否 | — |
| `tigerPublicKey` | — | Tiger RSA 公钥，用于响应签名验证 | 否 | 自动设置 |
| `language` | — | 语言（`zh_CN` / `zh_TW` / `en_US`） | 否 | `zh_CN` |
| `timeout` | — | HTTP 请求超时 | 否 | `15s` |

## 行情查询

```go
httpClient := client.NewHttpClient(cfg)
qc := quote.NewQuoteClient(httpClient)

// 市场状态
states, err := qc.MarketState("US")

// 实时报价
quotes, err := qc.QuoteRealTime([]string{"AAPL", "TSLA"})

// K 线数据
klines, err := qc.Kline("AAPL", "day")

// 分时数据
timeline, err := qc.Timeline([]string{"AAPL"})

// 深度行情
depth, err := qc.QuoteDepth("AAPL")

// 期权到期日
expiry, err := qc.OptionExpiration("AAPL")

// 期权链
chain, err := qc.OptionChain("AAPL", "2024-01-19")

// 期货交易所列表
exchanges, err := qc.FutureExchange()
```

## 交易

```go
tc := trade.NewTradeClient(httpClient, cfg.Account)

// 下限价单
order := model.Order{
	Symbol:        "AAPL",
	SecType:       "STK",
	Action:        "BUY",
	OrderType:     "LMT",
	TotalQuantity: 100,
	LimitPrice:    150.0,
	TimeInForce:   "DAY",
}
result, err := tc.PlaceOrder(order)

// 预览订单（不实际执行）
preview, err := tc.PreviewOrder(order)

// 修改订单
order.LimitPrice = 155.0
result, err = tc.ModifyOrder(orderId, order)

// 撤单
result, err = tc.CancelOrder(orderId)

// 查询订单
orders, err := tc.Orders()

// 查询持仓
positions, err := tc.Positions()

// 查询资产
assets, err := tc.Assets()
```

## 原始 API 调用

当 SDK 尚未封装某个 API 时，可以使用 `ExecuteRaw` 直接调用：

```go
httpClient := client.NewHttpClient(cfg)

resp, err := httpClient.ExecuteRaw("market_state", `{"market":"US"}`)
if err != nil {
	log.Fatal(err)
}
fmt.Println("原始响应:", resp)
```

## 实时推送

推送客户端通过 **TCP + TLS** 连接 Tiger 推送服务器，使用 varint32 帧 + **Protobuf** 二进制协议。支持自动重连（指数退避）和心跳保活。

回调数据类型为 `push/pb` 包中的 Protobuf 生成结构体（如 `*pb.QuoteData`、`*pb.OrderStatusData`、`*pb.AssetData`、`*pb.PositionData`）。

```go
import (
	"fmt"
	"log"

	"github.com/tigerfintech/openapi-go-sdk/config"
	"github.com/tigerfintech/openapi-go-sdk/push"
	"github.com/tigerfintech/openapi-go-sdk/push/pb"
)

cfg, err := config.NewClientConfig()
if err != nil {
	log.Fatal(err)
}

pc := push.NewPushClient(cfg)

pc.SetCallbacks(push.Callbacks{
	OnQuote: func(data *pb.QuoteData) {
		fmt.Printf("行情: %s 价格=%.2f 成交量=%d\n",
			data.GetSymbol(), data.GetLatestPrice(), data.GetVolume())
	},
	OnTick: func(data *pb.TradeTickData) {
		fmt.Println("收到逐笔成交")
	},
	OnDepth: func(data *pb.QuoteDepthData) {
		fmt.Println("收到深度行情更新")
	},
	OnKline: func(data *pb.KlineData) {
		fmt.Println("收到 K 线更新")
	},
	OnOrder: func(data *pb.OrderStatusData) {
		fmt.Printf("订单: %s 状态=%s\n", data.GetSymbol(), data.GetStatus())
	},
	OnAsset: func(data *pb.AssetData) {
		fmt.Printf("资产变动: 账户=%s\n", data.GetAccount())
	},
	OnPosition: func(data *pb.PositionData) {
		fmt.Printf("持仓变动: %s\n", data.GetSymbol())
	},
	OnTransaction: func(data *pb.OrderTransactionData) {
		fmt.Println("收到成交明细更新")
	},
	OnConnect:    func() { fmt.Println("推送已连接") },
	OnDisconnect: func() { fmt.Println("推送已断开") },
	OnError:      func(err error) { fmt.Println("推送错误:", err) },
	OnKickout:    func(msg string) { fmt.Println("被踢出:", msg) },
})

if err := pc.Connect(); err != nil {
	log.Fatal(err)
}
defer pc.Disconnect()

// 订阅行情
pc.SubscribeQuote([]string{"AAPL", "TSLA"})

// 订阅账户推送（空字符串使用配置中的账户）
pc.SubscribeAsset("")
pc.SubscribeOrder("")
pc.SubscribePosition("")
```

### 可用订阅

| 方法 | 回调 | 数据类型 |
|------|------|----------|
| `SubscribeQuote` | `OnQuote` | `*pb.QuoteData` |
| `SubscribeTick` | `OnTick` | `*pb.TradeTickData` |
| `SubscribeDepth` | `OnDepth` | `*pb.QuoteDepthData` |
| `SubscribeOption` | `OnOption` | `*pb.QuoteData` |
| `SubscribeFuture` | `OnFuture` | `*pb.QuoteData` |
| `SubscribeKline` | `OnKline` | `*pb.KlineData` |
| `SubscribeAsset` | `OnAsset` | `*pb.AssetData` |
| `SubscribePosition` | `OnPosition` | `*pb.PositionData` |
| `SubscribeOrder` | `OnOrder` | `*pb.OrderStatusData` |
| `SubscribeTransaction` | `OnTransaction` | `*pb.OrderTransactionData` |

## 项目结构

```
openapi-go-sdk/
├── config/    # 配置管理（ClientConfig、properties 解析、动态域名）
├── signer/    # RSA 签名与验证
├── client/    # HTTP 客户端（请求/响应、重试、ExecuteRaw）
├── model/     # 数据模型（Order、Contract、Position、枚举）
├── quote/     # 行情查询客户端
├── trade/     # 交易客户端
├── push/      # TCP+TLS 推送客户端（Protobuf）
│   └── pb/    # 生成的 Protobuf 类型
├── logger/    # 日志模块
└── examples/  # 示例代码
```

## API 参考

- [Tiger OpenAPI 文档](https://quant.itigerup.com/openapi/zh/python/overview/introduction.html)
- [Go SDK pkg.go.dev](https://pkg.go.dev/github.com/tigerfintech/openapi-go-sdk)

## 许可证

[MIT License](LICENSE)
