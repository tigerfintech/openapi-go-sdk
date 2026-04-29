// Quote example
//
// Demonstrates how to use QuoteClient to query a wide range of market data:
// market state, real-time quotes, kline, timeline, depth, trade ticks,
// option expirations, futures exchanges, capital flow, and quote permissions.
//
// Config is auto-discovered from ./tiger_openapi_config.properties or
// ~/.tigeropen/tiger_openapi_config.properties.
//
// Run: go run examples/quote_example.go
package main

import (
	"fmt"
	"log"

	"github.com/tigerfintech/openapi-go-sdk/client"
	"github.com/tigerfintech/openapi-go-sdk/config"
	"github.com/tigerfintech/openapi-go-sdk/quote"
)

func main() {
	cfg, err := config.NewClientConfig()
	if err != nil {
		log.Fatal("failed to create config:", err)
	}

	httpClient := client.NewHttpClient(cfg)
	qc := quote.NewQuoteClient(httpClient)

	// --- Basic market data ---

	fmt.Println("=== Market State ===")
	states, err := qc.MarketState("US")
	if err != nil {
		log.Printf("failed to query market state: %v", err)
	} else {
		fmt.Printf("market state: %s\n", states)
	}

	fmt.Println("\n=== Real-time Quote ===")
	quotes, err := qc.QuoteRealTime([]string{"AAPL", "TSLA", "GOOG"})
	if err != nil {
		log.Printf("failed to query real-time quote: %v", err)
	} else {
		fmt.Printf("quote data: %s\n", quotes)
	}

	fmt.Println("\n=== Kline (daily) ===")
	klines, err := qc.Kline("AAPL", "day")
	if err != nil {
		log.Printf("failed to query kline: %v", err)
	} else {
		fmt.Printf("kline data: %s\n", klines)
	}

	// --- Timeline and tick data ---

	fmt.Println("\n=== Timeline ===")
	timeline, err := qc.Timeline([]string{"AAPL"})
	if err != nil {
		log.Printf("failed to query timeline: %v", err)
	} else {
		fmt.Printf("timeline data: %s\n", timeline)
	}

	fmt.Println("\n=== Quote Depth ===")
	depth, err := qc.QuoteDepth("AAPL")
	if err != nil {
		log.Printf("failed to query quote depth: %v", err)
	} else {
		fmt.Printf("depth data: %s\n", depth)
	}

	fmt.Println("\n=== Trade Tick ===")
	ticks, err := qc.TradeTick([]string{"AAPL"})
	if err != nil {
		log.Printf("failed to query trade tick: %v", err)
	} else {
		fmt.Printf("trade tick data: %s\n", ticks)
	}

	// --- Options data ---

	fmt.Println("\n=== Option Expiration (AAPL) ===")
	expirations, err := qc.OptionExpiration("AAPL")
	if err != nil {
		log.Printf("failed to query option expiration: %v", err)
	} else {
		fmt.Printf("option expirations: %s\n", expirations)
	}

	// --- Futures data ---

	fmt.Println("\n=== Future Exchange ===")
	exchanges, err := qc.FutureExchange()
	if err != nil {
		log.Printf("failed to query future exchange: %v", err)
	} else {
		fmt.Printf("future exchanges: %s\n", exchanges)
	}

	// --- Capital flow ---

	fmt.Println("\n=== Capital Flow (AAPL) ===")
	flow, err := qc.CapitalFlow("AAPL")
	if err != nil {
		log.Printf("failed to query capital flow: %v", err)
	} else {
		fmt.Printf("capital flow: %s\n", flow)
	}

	// --- Quote permissions ---

	fmt.Println("\n=== Quote Permission ===")
	perm, err := qc.GrabQuotePermission()
	if err != nil {
		log.Printf("failed to grab quote permission: %v", err)
	} else {
		fmt.Printf("quote permission: %s\n", perm)
	}
}
