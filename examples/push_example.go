// Push example
//
// Demonstrates how to use PushClient to receive real-time push data.
// Registers ALL available callback types and subscribes to multiple data feeds
// including quotes, depth, tick, kline, options, futures, and account events.
//
// Config is auto-discovered from ./tiger_openapi_config.properties or
// ~/.tigeropen/tiger_openapi_config.properties.
//
// Run: go run examples/push_example.go
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/tigerfintech/openapi-go-sdk/config"
	"github.com/tigerfintech/openapi-go-sdk/push"
	"github.com/tigerfintech/openapi-go-sdk/push/pb"
)

func main() {
	cfg, err := config.NewClientConfig()
	if err != nil {
		log.Fatal("failed to create config:", err)
	}

	pc := push.NewPushClient(cfg)

	// Register ALL available callback types
	pc.SetCallbacks(push.Callbacks{
		// --- Quote push callbacks ---

		// OnQuote: stock quote updates (price, volume, etc.)
		OnQuote: func(data *pb.QuoteData) {
			fmt.Printf("[Quote] %s: price=%.4f volume=%d high=%.4f low=%.4f\n",
				data.GetSymbol(), data.GetLatestPrice(), data.GetVolume(),
				data.GetHigh(), data.GetLow())
		},

		// OnTick: trade-by-trade tick data
		OnTick: func(data *pb.TradeTickData) {
			prices := data.GetPrice()
			volumes := data.GetVolume()
			count := len(prices)
			fmt.Printf("[Tick] %s: %d ticks, sn=%d\n",
				data.GetSymbol(), count, data.GetSn())
			if count > 0 {
				fmt.Printf("  latest: price=%d volume=%d\n", prices[count-1], volumes[count-1])
			}
		},

		// OnDepth: order book depth (bid/ask levels)
		OnDepth: func(data *pb.QuoteDepthData) {
			ask := data.GetAsk()
			bid := data.GetBid()
			fmt.Printf("[Depth] %s: timestamp=%d\n", data.GetSymbol(), data.GetTimestamp())
			if ask != nil {
				fmt.Printf("  ask levels: %d\n", len(ask.GetPrice()))
			}
			if bid != nil {
				fmt.Printf("  bid levels: %d\n", len(bid.GetPrice()))
			}
		},

		// OnKline: real-time kline (candlestick) updates
		OnKline: func(data *pb.KlineData) {
			fmt.Printf("[Kline] %s: O=%.2f H=%.2f L=%.2f C=%.2f vol=%d\n",
				data.GetSymbol(), data.GetOpen(), data.GetHigh(),
				data.GetLow(), data.GetClose(), data.GetVolume())
		},

		// OnOption: option quote updates
		OnOption: func(data *pb.QuoteData) {
			fmt.Printf("[Option] %s: price=%.4f bid=%.4f ask=%.4f openInt=%d\n",
				data.GetSymbol(), data.GetLatestPrice(),
				data.GetBidPrice(), data.GetAskPrice(), data.GetOpenInt())
		},

		// OnFuture: futures quote updates
		OnFuture: func(data *pb.QuoteData) {
			fmt.Printf("[Future] %s: price=%.4f volume=%d openInt=%d\n",
				data.GetSymbol(), data.GetLatestPrice(),
				data.GetVolume(), data.GetOpenInt())
		},

		// OnStockTop: top stock movers data
		OnStockTop: func(data *pb.StockTopData) {
			fmt.Printf("[StockTop] market=%s timestamp=%d entries=%d\n",
				data.GetMarket(), data.GetTimestamp(), len(data.GetTopData()))
		},

		// OnOptionTop: top option activity data
		OnOptionTop: func(data *pb.OptionTopData) {
			fmt.Printf("[OptionTop] market=%s timestamp=%d entries=%d\n",
				data.GetMarket(), data.GetTimestamp(), len(data.GetTopData()))
		},

		// OnFullTick: full tick data with detailed trade info
		OnFullTick: func(data *pb.TickData) {
			fmt.Printf("[FullTick] %s: ticks=%d timestamp=%d source=%s\n",
				data.GetSymbol(), len(data.GetTicks()),
				data.GetTimestamp(), data.GetSource())
		},

		// OnQuoteBBO: best bid/offer quote updates
		OnQuoteBBO: func(data *pb.QuoteData) {
			fmt.Printf("[QuoteBBO] %s: bid=%.4f x %d  ask=%.4f x %d\n",
				data.GetSymbol(),
				data.GetBidPrice(), data.GetBidSize(),
				data.GetAskPrice(), data.GetAskSize())
		},

		// --- Account push callbacks ---

		// OnOrder: order status changes
		OnOrder: func(data *pb.OrderStatusData) {
			fmt.Printf("[Order] id=%d %s %s: status=%s\n",
				data.GetId(), data.GetAction(), data.GetSymbol(), data.GetStatus())
		},

		// OnAsset: account asset changes
		OnAsset: func(data *pb.AssetData) {
			fmt.Printf("[Asset] account=%s netLiquidation=%.2f buyingPower=%.2f\n",
				data.GetAccount(), data.GetNetLiquidation(), data.GetBuyingPower())
		},

		// OnPosition: position changes
		OnPosition: func(data *pb.PositionData) {
			fmt.Printf("[Position] %s: qty=%d avgCost=%.4f marketValue=%.2f\n",
				data.GetSymbol(), data.GetPosition(),
				data.GetAverageCost(), data.GetMarketValue())
		},

		// OnTransaction: order fill/transaction events
		OnTransaction: func(data *pb.OrderTransactionData) {
			fmt.Printf("[Transaction] orderId=%d %s %s: filled %.4f x %d\n",
				data.GetOrderId(), data.GetAction(), data.GetSymbol(),
				data.GetFilledPrice(), data.GetFilledQuantity())
		},

		// --- Connection status callbacks ---

		OnConnect:    func() { fmt.Println("[Status] connected to push server") },
		OnDisconnect: func() { fmt.Println("[Status] disconnected from push server") },
		OnError:      func(err error) { fmt.Printf("[Error] %v\n", err) },
		OnKickout:    func(msg string) { fmt.Printf("[Kickout] %s\n", msg) },
	})

	// Connect to the push server
	if err := pc.Connect(); err != nil {
		log.Fatal("connect failed:", err)
	}
	defer pc.Disconnect()

	// --- Subscribe to market data ---
	fmt.Println("=== Subscribing to market data ===")

	// Real-time quotes for multiple symbols
	if err := pc.SubscribeQuote([]string{"AAPL", "TSLA", "GOOG"}); err != nil {
		log.Printf("subscribe quote failed: %v", err)
	}
	fmt.Println("subscribed to quotes: AAPL, TSLA, GOOG")

	// Order book depth for AAPL
	if err := pc.SubscribeDepth([]string{"AAPL"}); err != nil {
		log.Printf("subscribe depth failed: %v", err)
	}
	fmt.Println("subscribed to depth: AAPL")

	// Trade-by-trade tick data for AAPL
	if err := pc.SubscribeTick([]string{"AAPL"}); err != nil {
		log.Printf("subscribe tick failed: %v", err)
	}
	fmt.Println("subscribed to tick: AAPL")

	// --- Subscribe to account events ---
	fmt.Println("\n=== Subscribing to account events ===")

	// Order status updates (pass "" to use the account from config)
	if err := pc.SubscribeOrder(""); err != nil {
		log.Printf("subscribe order failed: %v", err)
	}
	fmt.Println("subscribed to order updates")

	// Position changes
	if err := pc.SubscribePosition(""); err != nil {
		log.Printf("subscribe position failed: %v", err)
	}
	fmt.Println("subscribed to position updates")

	// Asset changes
	if err := pc.SubscribeAsset(""); err != nil {
		log.Printf("subscribe asset failed: %v", err)
	}
	fmt.Println("subscribed to asset updates")

	// Transaction (fill) events
	if err := pc.SubscribeTransaction(""); err != nil {
		log.Printf("subscribe transaction failed: %v", err)
	}
	fmt.Println("subscribed to transaction updates")

	fmt.Println("\nwaiting for push data... press Ctrl+C to exit")

	// Wait for interrupt signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	fmt.Println("\nshutting down...")
}
