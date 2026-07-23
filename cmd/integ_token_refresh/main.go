// Integration test: real API connectivity + token refresh mechanism
//
// Run: go run ./cmd/integ_token_refresh
// Override server: TIGER_SERVER_URL=<url> go run ./cmd/integ_token_refresh
package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"sync/atomic"
	"time"

	"github.com/tigerfintech/openapi-go-sdk/client"
	"github.com/tigerfintech/openapi-go-sdk/config"
)

func pass(name, note string) { fmt.Printf("[ OK ] %-58s %s\n", name, note) }
func failf(name string, err error) {
	fmt.Printf("[FAIL] %-58s %v\n", name, err)
	log.Fatalf("FAIL: %v", err)
}

// makeExpiredToken returns a token whose gen_ts = 1ms, always triggering refresh.
func makeExpiredToken() string {
	payload := "0000000000001,0000000000002extra_payload_padding"
	return base64.StdEncoding.EncodeToString([]byte(payload))
}

func main() {
	log.SetFlags(log.Ltime)

	configPath := os.Getenv("TIGER_CONFIG_PATH")
	var cfgOpts []config.Option
	if configPath != "" {
		cfgOpts = append(cfgOpts, config.WithPropertiesFile(configPath))
	}
	if serverURL := os.Getenv("TIGER_SERVER_URL"); serverURL != "" {
		cfgOpts = append(cfgOpts, config.WithServerURL(serverURL))
	}

	cfg, err := config.NewClientConfig(cfgOpts...)
	if err != nil {
		failf("NewClientConfig", err)
	}
	cfg.TigerPublicKey = ""
	log.Printf("tiger_id=%s server=%s", cfg.TigerID, cfg.ServerURL)

	// ── Test 1: Connectivity ──────────────────────────────────────────────────
	hc1 := client.NewHttpClient(cfg)
	defer hc1.Close()

	req1, _ := client.NewApiRequest("market_state", map[string]string{"market": "US"})
	resp1, err := hc1.Execute(req1)
	if err != nil {
		failf("market_state", err)
	}
	pass("market_state", fmt.Sprintf("code=%d", resp1.Code))

	// ── Test 2: TokenManager auto-refresh ─────────────────────────────────────
	wantToken := fmt.Sprintf("refreshed_token_%d", time.Now().Unix())
	var refreshCount int64
	var lastToken string
	expiredTok := makeExpiredToken()

	tm := config.NewTokenManager(
		config.WithRefreshDuration(1),
		config.WithTokenRefreshInterval(200*time.Millisecond),
		config.TokenWriterOption(func(t string) {
			if t == expiredTok {
				return
			}
			n := atomic.AddInt64(&refreshCount, 1)
			lastToken = t
			log.Printf("    [writer] refresh #%d", n)
		}),
	)
	_ = tm.SetToken(expiredTok)
	tm.StartAutoRefresh(func() (string, error) { return wantToken, nil })
	defer tm.StopAutoRefresh()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if atomic.LoadInt64(&refreshCount) > 0 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if atomic.LoadInt64(&refreshCount) == 0 {
		failf("StartAutoRefresh", fmt.Errorf("not triggered within 2s"))
	}
	if lastToken != wantToken {
		failf("TokenWriter", fmt.Errorf("got %q, want %q", lastToken, wantToken))
	}
	if got := tm.GetToken(); got != wantToken {
		failf("GetToken", fmt.Errorf("got %q", got))
	}
	pass("TokenManager.StartAutoRefresh + TokenWriter", fmt.Sprintf("count=%d", atomic.LoadInt64(&refreshCount)))

	// ── Test 3: WithSharedTokenFrom ───────────────────────────────────────────
	cfg3, err := config.NewClientConfig(cfgOpts...)
	if err != nil {
		failf("NewClientConfig (test3)", err)
	}
	cfg3.TigerPublicKey = ""
	if cfg3.QuoteServerURL == "" {
		cfg3.QuoteServerURL = cfg3.ServerURL
	}

	primary3 := client.NewHttpClient(cfg3)
	defer primary3.Close()
	quoteHC := client.NewQuoteHttpClient(cfg3, client.WithSharedTokenFrom(primary3))

	simulatedToken := fmt.Sprintf("shared_token_%d", time.Now().Unix())
	var sharedCount int64

	sharedTM := primary3.StartTokenAutoRefresh(nil,
		config.WithRefreshDuration(1),
		config.WithTokenRefreshInterval(200*time.Millisecond),
		config.TokenWriterOption(func(t string) {
			if t != makeExpiredToken() {
				atomic.AddInt64(&sharedCount, 1)
			}
		}),
	)
	defer sharedTM.StopAutoRefresh()
	sharedTM.StartAutoRefresh(func() (string, error) { return simulatedToken, nil })
	_ = sharedTM.SetToken(makeExpiredToken())

	deadline3 := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline3) {
		if atomic.LoadInt64(&sharedCount) > 0 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if atomic.LoadInt64(&sharedCount) == 0 {
		failf("primary3 auto-refresh", fmt.Errorf("not triggered within 2s"))
	}
	if got := sharedTM.GetToken(); got != simulatedToken {
		failf("primary3 token after refresh", fmt.Errorf("got %q", got))
	}

	req3, _ := client.NewApiRequest("market_state", map[string]string{"market": "US"})
	resp3, err := quoteHC.Execute(req3)
	if err != nil {
		failf("quote client (WithSharedTokenFrom)", err)
	}
	pass("WithSharedTokenFrom: quote client API call after primary refresh", fmt.Sprintf("code=%d", resp3.Code))

	// ── Test 4: Double StartAutoRefresh — no goroutine leak ───────────────────
	var callCount int64
	tm4 := config.NewTokenManager(
		config.WithRefreshDuration(1),
		config.WithTokenRefreshInterval(100*time.Millisecond),
	)
	_ = tm4.SetToken(makeExpiredToken())
	fn4 := func() (string, error) { atomic.AddInt64(&callCount, 1); return "tok4", nil }
	tm4.StartAutoRefresh(fn4)
	tm4.StartAutoRefresh(fn4)

	time.Sleep(500 * time.Millisecond)
	tm4.StopAutoRefresh()
	countAtStop := atomic.LoadInt64(&callCount)
	time.Sleep(500 * time.Millisecond)

	if countAtStop == 0 {
		failf("double StartAutoRefresh", fmt.Errorf("never called"))
	}
	if got := atomic.LoadInt64(&callCount); got != countAtStop {
		failf("goroutine leak", fmt.Errorf("count grew from %d to %d after stop", countAtStop, got))
	}
	pass("double StartAutoRefresh + StopAutoRefresh (no goroutine leak)", fmt.Sprintf("calls=%d", countAtStop))

	fmt.Println()
	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Println("  PASS: All integration tests passed")
	fmt.Println("════════════════════════════════════════════════════════════════")
}
