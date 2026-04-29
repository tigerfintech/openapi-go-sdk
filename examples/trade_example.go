// Trade example
//
// Demonstrates how to use TradeClient to build orders, query orders (all, active,
// filled), check positions, and view account assets.
//
// Config is auto-discovered from ./tiger_openapi_config.properties or
// ~/.tigeropen/tiger_openapi_config.properties.
//
// Run: go run examples/trade_example.go
package main

import (
	"fmt"
	"log"

	"github.com/tigerfintech/openapi-go-sdk/client"
	"github.com/tigerfintech/openapi-go-sdk/config"
	"github.com/tigerfintech/openapi-go-sdk/model"
	"github.com/tigerfintech/openapi-go-sdk/trade"
)

func main() {
	cfg, err := config.NewClientConfig()
	if err != nil {
		log.Fatal("failed to create config:", err)
	}

	httpClient := client.NewHttpClient(cfg)
	tc := trade.NewTradeClient(httpClient, cfg.Account)

	// --- Build a complete limit order (not submitted) ---
	fmt.Println("=== Order Struct ===")
	order := model.Order{
		Symbol:        "AAPL",
		SecType:       "STK",
		Market:        "US",
		Currency:      "USD",
		Action:        "BUY",
		OrderType:     "LMT",
		TotalQuantity: 100,
		LimitPrice:    150.00,
		TimeInForce:   "DAY",
		OutsideRth:    false,
		// Attach a profit-taking leg
		OrderLegs: []model.OrderLeg{
			{
				LegType:     "PROFIT",
				Price:       180.00,
				TimeInForce: "GTC",
			},
		},
	}
	fmt.Printf("order to place:\n")
	fmt.Printf("  symbol:     %s\n", order.Symbol)
	fmt.Printf("  secType:    %s\n", order.SecType)
	fmt.Printf("  market:     %s\n", order.Market)
	fmt.Printf("  currency:   %s\n", order.Currency)
	fmt.Printf("  action:     %s\n", order.Action)
	fmt.Printf("  orderType:  %s\n", order.OrderType)
	fmt.Printf("  quantity:   %d\n", order.TotalQuantity)
	fmt.Printf("  limitPrice: %.2f\n", order.LimitPrice)
	fmt.Printf("  tif:        %s\n", order.TimeInForce)
	fmt.Printf("  outsideRth: %v\n", order.OutsideRth)
	if len(order.OrderLegs) > 0 {
		fmt.Printf("  legs:\n")
		for i, leg := range order.OrderLegs {
			fmt.Printf("    [%d] type=%s price=%.2f tif=%s\n",
				i, leg.LegType, leg.Price, leg.TimeInForce)
		}
	}

	// Uncomment to actually place the order (requires a funded account):
	// result, err := tc.PlaceOrder(order)
	// if err != nil {
	//     log.Printf("place order failed: %v", err)
	// } else {
	//     fmt.Printf("place order result: %s\n", result)
	// }

	// --- Query account assets ---
	fmt.Println("\n=== Assets ===")
	assets, err := tc.Assets()
	if err != nil {
		log.Printf("failed to query assets: %v", err)
	} else {
		fmt.Printf("assets: %s\n", assets)
	}

	// --- Query all orders ---
	fmt.Println("\n=== All Orders ===")
	orders, err := tc.Orders()
	if err != nil {
		log.Printf("failed to query orders: %v", err)
	} else {
		fmt.Printf("orders: %s\n", orders)
	}

	// --- Query active (pending) orders ---
	fmt.Println("\n=== Active Orders ===")
	activeOrders, err := tc.ActiveOrders()
	if err != nil {
		log.Printf("failed to query active orders: %v", err)
	} else {
		fmt.Printf("active orders: %s\n", activeOrders)
	}

	// --- Query filled orders ---
	fmt.Println("\n=== Filled Orders ===")
	filledOrders, err := tc.FilledOrders()
	if err != nil {
		log.Printf("failed to query filled orders: %v", err)
	} else {
		fmt.Printf("filled orders: %s\n", filledOrders)
	}

	// --- Query positions ---
	fmt.Println("\n=== Positions ===")
	positions, err := tc.Positions()
	if err != nil {
		log.Printf("failed to query positions: %v", err)
	} else {
		fmt.Printf("positions: %s\n", positions)
	}
}
