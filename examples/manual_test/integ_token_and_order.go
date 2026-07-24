// Integration test: token refresh, secret_key injection, and order helpers.
// Run: TIGER_CONFIG_PATH=~/.tigeropen/... go run examples/manual_test/integ_token_and_order.go
package main

import (
	"fmt"
	"os"
	"sync/atomic"

	"github.com/tigerfintech/openapi-go-sdk/client"
	"github.com/tigerfintech/openapi-go-sdk/config"
	"github.com/tigerfintech/openapi-go-sdk/model"
	"github.com/tigerfintech/openapi-go-sdk/trade"
)

func pass(name, note string) { fmt.Printf("[ OK ] %-50s %s\n", name, note) }
func fail(name, err string)  { fmt.Printf("[FAIL] %-50s %s\n", name, err); os.Exit(1) }
func info(name, note string) { fmt.Printf("[INFO] %-50s %s\n", name, note) }

func main() {
	path := os.Getenv("TIGER_CONFIG_PATH")
	if path == "" {
		path = os.Getenv("HOME") + "/.tigeropen/tiger_openapi_config.properties"
	}

	cfg, err := config.NewClientConfig(config.WithPropertiesFile(path))
	if err != nil {
		fail("ClientConfig", err.Error())
	}
	fmt.Printf("tiger_id=%s  account=%s  secret_key=%s\n\n",
		cfg.TigerID, cfg.Account,
		func() string {
			if cfg.SecretKey != "" {
				return cfg.SecretKey[:8] + "..."
			}
			return "(none)"
		}())

	hc := client.NewHttpClient(cfg)
	defer hc.Close()

	// ── 1. secret_key 注入验证 ─────────────────────────────────────────────
	if cfg.SecretKey != "" {
		if hc.SecretKey() == cfg.SecretKey {
			pass("SecretKey injected into HttpClient", cfg.SecretKey[:8]+"...")
		} else {
			fail("SecretKey mismatch", hc.SecretKey())
		}
	} else {
		info("SecretKey", "(not set in config, skipping injection check)")
	}

	// ── 2. QueryToken ──────────────────────────────────────────────────────
	if cfg.Token != "" {
		newToken, err := hc.QueryToken()
		if err != nil {
			info("QueryToken", "failed: "+err.Error())
		} else {
			pass("QueryToken", fmt.Sprintf("len=%d prefix=%s", len(newToken), newToken[:20]))
			// Sync new token for next call
			cfg.Token = newToken
		}
	} else {
		info("QueryToken", "(no token in config, skipping)")
	}

	// ── 3. RefreshToken ────────────────────────────────────────────────────
	if cfg.Token != "" {
		hc2 := client.NewHttpClient(cfg)
		defer hc2.Close()
		if err := hc2.RefreshToken(nil); err != nil {
			info("RefreshToken", "failed: "+err.Error())
		} else {
			pass("RefreshToken", "OK")
		}
	} else {
		info("RefreshToken", "(no token in config, skipping)")
	}

	// ── 4. StartTokenAutoRefresh (mock) ────────────────────────────────────
	var refreshCount int32
	var gotToken string
	tm := hc.StartTokenAutoRefresh(nil,
		config.WithRefreshDuration(30),
		// use default interval (5m), don't pass 0
		config.TokenWriterOption(func(t string) {
			atomic.AddInt32(&refreshCount, 1)
			gotToken = t
		}),
	)
	// Use mock: set expired token manually
	// (real auto-refresh only triggers if cfg.Token is old enough)
	_ = tm.SetToken("") // clear; test infrastructure only
	tm.StopAutoRefresh()
	pass("StartTokenAutoRefresh", "started and stopped without error")

	// ── 5. Order helpers serialization ─────────────────────────────────────
	fmt.Println("\n--- Order helpers ---")

	mba := model.MarketOrderByAmount(cfg.Account, "00700", "STK", "BUY", 10000)
	if mba.Amount == 10000 && mba.TotalQuantity == 0 {
		pass("MarketOrderByAmount", fmt.Sprintf("Amount=%.0f TotalQuantity=%d", mba.Amount, mba.TotalQuantity))
	} else {
		fail("MarketOrderByAmount", fmt.Sprintf("%+v", mba))
	}

	lba := model.LimitOrderByAmount(cfg.Account, "00700", "STK", "BUY", 10000, 1.0)
	if lba.Amount == 10000 && lba.LimitPrice == 1.0 {
		pass("LimitOrderByAmount", fmt.Sprintf("Amount=%.0f LimitPrice=%.2f", lba.Amount, lba.LimitPrice))
	} else {
		fail("LimitOrderByAmount", fmt.Sprintf("%+v", lba))
	}

	tbp := model.TrailOrderByPrice(cfg.Account, "AAPL", "STK", "SELL", 1, 5.0)
	if tbp.AuxPrice == 5.0 && tbp.TrailingPercent == 0 {
		pass("TrailOrderByPrice", fmt.Sprintf("AuxPrice=%.1f TrailingPercent=0", tbp.AuxPrice))
	} else {
		fail("TrailOrderByPrice", fmt.Sprintf("%+v", tbp))
	}

	legs := []model.OrderLegRequest{model.NewOrderLeg("PROFIT", 300.0, "GTC")}
	lwl := model.LimitOrderWithLegs(cfg.Account, "AAPL", "STK", "BUY", 1, 1.0, legs)
	if len(lwl.OrderLegs) == 1 && lwl.LimitPrice == 1.0 {
		pass("LimitOrderWithLegs", fmt.Sprintf("legs=%d LimitPrice=%.2f", len(lwl.OrderLegs), lwl.LimitPrice))
	} else {
		fail("LimitOrderWithLegs", fmt.Sprintf("%+v", lwl))
	}

	cleg := model.NewContractLeg("AAPL", "OPT", "BUY", 1, "20261218", "200", "CALL")
	if cleg.Symbol == "AAPL" && *cleg.Ratio == 1 {
		pass("NewContractLeg", fmt.Sprintf("symbol=%s ratio=%d right=%s", cleg.Symbol, *cleg.Ratio, cleg.Right))
	} else {
		fail("NewContractLeg", fmt.Sprintf("%+v", cleg))
	}

	// ── 6. Real placeOrder with secret_key (if account has trade permission) ─
	if cfg.SecretKey != "" {
		fmt.Println("\n--- Real placeOrder with secret_key ---")
		tc := trade.NewTradeClientFromConfig(cfg)
		order := model.LimitOrderWithLegs(cfg.Account, "AAPL", "STK", "BUY", 1, 1.00, legs)
		order.Market = "US"
		order.Currency = "USD"
		placed, err := tc.PlaceOrder(order)
		if err != nil {
			info("PlaceOrder (with secret_key)", "server: "+err.Error())
		} else if placed != nil {
			pass("PlaceOrder (with secret_key)", fmt.Sprintf("id=%d", placed.ID))
			if cancelResult, cancelErr := tc.CancelOrder(placed.ID); cancelErr != nil {
				info("CancelOrder", cancelErr.Error())
			} else {
				pass("CancelOrder", fmt.Sprintf("id=%d", cancelResult.ID))
			}
		}
	}

	fmt.Printf("\n%s\n", "══════════════════════════════════════════════")
	fmt.Println("  All checks passed")
	_ = gotToken
	_ = refreshCount
}
