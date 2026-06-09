// Push example - 覆盖 PushClient 全部推送类型,用真实配置验证长连接推送
//
// 运行方式:
//
//	go run ./examples/push               # 默认 30 秒后自动退出
//	go run ./examples/push -duration=60  # 自定义运行时长(秒)
//	go run ./examples/push -trade=false  # 不触发下单,仅接收行情推送
//
// 为了产生账户类推送事件(OnOrder / OnAsset / OnPosition / OnTransaction),
// 默认会发一笔深度 OTM 的限价单 AAPL @ $1.00 并立即撤销。
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/tigerfintech/openapi-go-sdk/client"
	"github.com/tigerfintech/openapi-go-sdk/config"
	"github.com/tigerfintech/openapi-go-sdk/model"
	"github.com/tigerfintech/openapi-go-sdk/push"
	"github.com/tigerfintech/openapi-go-sdk/push/pb"
	"github.com/tigerfintech/openapi-go-sdk/trade"
)

// 各类推送事件计数
type counters struct {
	connect     int32
	disconnect  int32
	errorCount  int32
	kickout     int32
	quote       int32
	quoteBBO    int32
	tick        int32
	depth       int32
	option      int32
	future      int32
	kline       int32
	stockTop    int32
	optionTop   int32
	fullTick    int32
	asset       int32
	position    int32
	order       int32
	transaction int32
}

func main() {
	duration := flag.Int("duration", 30, "推送监听时长(秒)")
	doTrade := flag.Bool("trade", true, "是否触发下单/撤单以产生账户类推送事件")
	flag.Parse()

	cfg, err := config.NewClientConfig()
	if err != nil {
		log.Fatal("load config failed: ", err)
	}
	fmt.Printf("tiger_id=%s account=%s\n\n", cfg.TigerID, cfg.Account)

	var cnt counters
	pc := push.NewPushClient(cfg)

	pc.SetCallbacks(push.Callbacks{
		OnConnect: func() {
			atomic.AddInt32(&cnt.connect, 1)
			fmt.Println("[Status] connected")
		},
		OnDisconnect: func() {
			atomic.AddInt32(&cnt.disconnect, 1)
			fmt.Println("[Status] disconnected")
		},
		OnError: func(err error) {
			atomic.AddInt32(&cnt.errorCount, 1)
			fmt.Printf("[Error] %v\n", err)
		},
		OnKickout: func(msg string) {
			atomic.AddInt32(&cnt.kickout, 1)
			fmt.Printf("[Kickout] %s\n", msg)
		},

		OnQuote: func(data *pb.QuoteData) {
			n := atomic.AddInt32(&cnt.quote, 1)
			if n <= 3 {
				fmt.Printf("[Quote] %s price=%.4f vol=%d\n", data.GetSymbol(), data.GetLatestPrice(), data.GetVolume())
			}
		},
		OnQuoteBBO: func(data *pb.QuoteData) {
			atomic.AddInt32(&cnt.quoteBBO, 1)
		},
		OnTick: func(data *pb.TradeTickData) {
			n := atomic.AddInt32(&cnt.tick, 1)
			if n <= 3 {
				fmt.Printf("[Tick] %s ticks=%d sn=%d\n", data.GetSymbol(), len(data.GetPrice()), data.GetSn())
			}
		},
		OnDepth: func(data *pb.QuoteDepthData) {
			n := atomic.AddInt32(&cnt.depth, 1)
			if n <= 3 {
				askLevels := 0
				if data.GetAsk() != nil {
					askLevels = len(data.GetAsk().GetPrice())
				}
				fmt.Printf("[Depth] %s ask_levels=%d\n", data.GetSymbol(), askLevels)
			}
		},
		OnOption: func(data *pb.QuoteData) {
			atomic.AddInt32(&cnt.option, 1)
		},
		OnFuture: func(data *pb.QuoteData) {
			atomic.AddInt32(&cnt.future, 1)
		},
		OnKline: func(data *pb.KlineData) {
			atomic.AddInt32(&cnt.kline, 1)
		},
		OnStockTop: func(data *pb.StockTopData) {
			atomic.AddInt32(&cnt.stockTop, 1)
		},
		OnOptionTop: func(data *pb.OptionTopData) {
			atomic.AddInt32(&cnt.optionTop, 1)
		},
		OnFullTick: func(data *pb.TickData) {
			atomic.AddInt32(&cnt.fullTick, 1)
		},

		OnOrder: func(data *pb.OrderStatusData) {
			atomic.AddInt32(&cnt.order, 1)
			fmt.Printf("[Order] id=%d %s %s status=%s\n", data.GetId(), data.GetAction(), data.GetSymbol(), data.GetStatus())
		},
		OnAsset: func(data *pb.AssetData) {
			n := atomic.AddInt32(&cnt.asset, 1)
			if n <= 3 {
				fmt.Printf("[Asset] netLiq=%.2f buyingPower=%.2f\n", data.GetNetLiquidation(), data.GetBuyingPower())
			}
		},
		OnPosition: func(data *pb.PositionData) {
			n := atomic.AddInt32(&cnt.position, 1)
			if n <= 3 {
				fmt.Printf("[Position] %s qty=%d avgCost=%.4f\n", data.GetSymbol(), data.GetPosition(), data.GetAverageCost())
			}
		},
		OnTransaction: func(data *pb.OrderTransactionData) {
			atomic.AddInt32(&cnt.transaction, 1)
			fmt.Printf("[Transaction] orderId=%d %s filled %.4f x %d\n",
				data.GetOrderId(), data.GetSymbol(), data.GetFilledPrice(), data.GetFilledQuantity())
		},
	})

	if err := pc.Connect(); err != nil {
		log.Fatal("connect failed: ", err)
	}
	defer pc.Disconnect()

	// 订阅行情
	fmt.Println("=== Subscribing ===")
	mustSub("SubscribeQuote", pc.SubscribeQuote([]string{"AAPL", "TSLA", "GOOG"}))
	mustSub("SubscribeTick", pc.SubscribeTick([]string{"AAPL"}))
	mustSub("SubscribeDepth", pc.SubscribeDepth([]string{"AAPL"}))
	mustSub("SubscribeKline", pc.SubscribeKline([]string{"AAPL"}))
	mustSub("SubscribeOption", pc.SubscribeOption([]string{"AAPL  260619C00200000"}))

	mustSub("SubscribeOrder", pc.SubscribeOrder(""))
	mustSub("SubscribePosition", pc.SubscribePosition(""))
	mustSub("SubscribeAsset", pc.SubscribeAsset(""))
	mustSub("SubscribeTransaction", pc.SubscribeTransaction(""))

	// 如启用 -trade,触发一次下单/撤单,让账户类推送有数据可验证
	if *doTrade {
		fmt.Println("\n=== 触发下单/撤单以产生账户类推送事件 ===")
		tc := trade.NewTradeClient(client.NewHttpClient(cfg), cfg.Account)
		time.Sleep(2 * time.Second) // 等订阅完成
		req := model.LimitOrder(cfg.Account, "AAPL", "STK", "BUY", 1, 1.00)
		req.Market = "US"
		req.Currency = "USD"
		req.TimeInForce = "DAY"
		placed, err := tc.PlaceOrder(req)
		if err != nil {
			fmt.Printf("PlaceOrder 失败(不影响推送验证): %v\n", err)
		} else {
			fmt.Printf("已下单 id=%d, 等待 3 秒后撤销...\n", placed.ID)
			time.Sleep(3 * time.Second)
			if _, err := tc.CancelOrder(placed.ID); err != nil {
				fmt.Printf("CancelOrder 失败: %v\n", err)
			} else {
				fmt.Printf("已撤销 id=%d\n", placed.ID)
			}
		}
	}

	// 等待推送数据 & 允许 Ctrl+C 提前退出
	fmt.Printf("\n=== 等待推送数据 %ds,按 Ctrl+C 提前退出 ===\n", *duration)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-time.After(time.Duration(*duration) * time.Second):
	case <-sig:
		fmt.Println("收到退出信号...")
	}

	printSummary(&cnt)
}

func mustSub(name string, err error) {
	if err != nil {
		fmt.Printf("[FAIL] %-25s %v\n", name, err)
	} else {
		fmt.Printf("[ OK ] %-25s subscribed\n", name)
	}
}

func printSummary(c *counters) {
	fmt.Println("\n================ PUSH SUMMARY ================")
	fmt.Println("连接状态事件:")
	fmt.Printf("  OnConnect       = %d\n", atomic.LoadInt32(&c.connect))
	fmt.Printf("  OnDisconnect    = %d\n", atomic.LoadInt32(&c.disconnect))
	fmt.Printf("  OnError         = %d\n", atomic.LoadInt32(&c.errorCount))
	fmt.Printf("  OnKickout       = %d\n", atomic.LoadInt32(&c.kickout))
	fmt.Println("行情推送事件:")
	fmt.Printf("  OnQuote         = %d\n", atomic.LoadInt32(&c.quote))
	fmt.Printf("  OnQuoteBBO      = %d\n", atomic.LoadInt32(&c.quoteBBO))
	fmt.Printf("  OnTick          = %d\n", atomic.LoadInt32(&c.tick))
	fmt.Printf("  OnDepth         = %d\n", atomic.LoadInt32(&c.depth))
	fmt.Printf("  OnOption        = %d\n", atomic.LoadInt32(&c.option))
	fmt.Printf("  OnFuture        = %d\n", atomic.LoadInt32(&c.future))
	fmt.Printf("  OnKline         = %d\n", atomic.LoadInt32(&c.kline))
	fmt.Printf("  OnStockTop      = %d\n", atomic.LoadInt32(&c.stockTop))
	fmt.Printf("  OnOptionTop     = %d\n", atomic.LoadInt32(&c.optionTop))
	fmt.Printf("  OnFullTick      = %d\n", atomic.LoadInt32(&c.fullTick))
	fmt.Println("账户推送事件:")
	fmt.Printf("  OnOrder         = %d\n", atomic.LoadInt32(&c.order))
	fmt.Printf("  OnAsset         = %d\n", atomic.LoadInt32(&c.asset))
	fmt.Printf("  OnPosition      = %d\n", atomic.LoadInt32(&c.position))
	fmt.Printf("  OnTransaction   = %d\n", atomic.LoadInt32(&c.transaction))
	fmt.Println("==============================================")
}
