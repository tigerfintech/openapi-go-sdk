// Package integtest provides shared helpers for integration tests that hit
// the real OpenAPI gateway. Unit tests (the default `go test ./...`) do not
// depend on this package; integration tests opt in via the `integration` build
// tag and read credentials from environment variables.
package integtest

import (
	"os"
	"testing"

	"github.com/tigerfintech/openapi-go-sdk/client"
	"github.com/tigerfintech/openapi-go-sdk/config"
	"github.com/tigerfintech/openapi-go-sdk/quote"
	"github.com/tigerfintech/openapi-go-sdk/trade"
)

// Env var names are aligned with the Java and Python SDKs so the same CI
// variables work across all three.
const (
	envTigerID    = "TIGEROPEN_TIGER_ID"
	envPrivateKey = "TIGEROPEN_PRIVATE_KEY"
	envAccount    = "TIGEROPEN_ACCOUNT"
	envServerURL  = "TIGEROPEN_SERVER_URL"
	envPropsPath  = "TIGEROPEN_PROPS_PATH"
	envRunInteg   = "TIGER_RUN_INTEG"

	defaultServerURL = "https://openapi.tigerfintech.com/gateway"
)

// IsEnabled reports whether integration tests should run. Mirrors the
// TIGER_RUN_INTEG gate used by the Python SDK.
func IsEnabled() bool {
	return os.Getenv(envRunInteg) == "true"
}

// SkipUnlessEnabled skips the current test when integration mode is off.
func SkipUnlessEnabled(t testing.TB) {
	t.Helper()
	if !IsEnabled() {
		t.Skipf("skipped: set %s=true to run integration tests", envRunInteg)
	}
}

// SkipUnlessCreds skips when no credentials are available, regardless of the
// run flag. Useful for local runs where a developer simply exports the
// TIGEROPEN_* vars without setting TIGER_RUN_INTEG.
func SkipUnlessCreds(t testing.TB) {
	t.Helper()
	if !hasCreds() {
		t.Skipf("skipped: no credentials — set %s/%s/%s env vars or %s",
			envTigerID, envPrivateKey, envAccount, envPropsPath)
	}
}

func hasCreds() bool {
	if os.Getenv(envTigerID) != "" && os.Getenv(envPrivateKey) != "" {
		return true
	}
	return os.Getenv(envPropsPath) != ""
}

// NewConfig builds a *config.ClientConfig from env vars or a properties file.
// It delegates to config.NewClientConfig, which already reads the
// TIGEROPEN_TIGER_ID / PRIVATE_KEY / ACCOUNT env vars with the highest
// priority. When TIGEROPEN_PROPS_PATH points at a properties file, it is
// loaded via WithPropertiesFile, which also disables ./ auto-discovery so
// local dev configs cannot leak into CI runs.
//
// Skips only when credentials are absent. A malformed properties file or
// invalid key is a real configuration error and must fail, not skip — otherwise
// the integration job would report success while having run nothing.
func NewConfig(t testing.TB) *config.ClientConfig {
	t.Helper()
	if !hasCreds() {
		t.Skipf("skipped: no credentials — set %s/%s/%s env vars or %s",
			envTigerID, envPrivateKey, envAccount, envPropsPath)
	}

	opts := []config.Option{config.WithServerURL(serverURL())}
	if propsPath := os.Getenv(envPropsPath); propsPath != "" {
		opts = append(opts, config.WithPropertiesFile(propsPath))
	}
	cfg, err := config.NewClientConfig(opts...)
	if err != nil {
		t.Fatalf("build integ config: %v", err)
	}
	return cfg
}

// NewQuoteClient builds a QuoteClient wired to the real gateway.
func NewQuoteClient(t testing.TB) *quote.QuoteClient {
	t.Helper()
	cfg := NewConfig(t)
	httpClient := client.NewQuoteHttpClient(cfg)
	return quote.NewQuoteClient(httpClient)
}

// NewTradeClient builds a TradeClient wired to the real gateway.
func NewTradeClient(t testing.TB) *trade.TradeClient {
	t.Helper()
	cfg := NewConfig(t)
	httpClient := client.NewHttpClient(cfg)
	return trade.NewTradeClient(httpClient, cfg.Account)
}

func serverURL() string {
	if u := os.Getenv(envServerURL); u != "" {
		return u
	}
	return defaultServerURL
}
