// Trading-hours + dynamic identifier resolvers for integration tests.
//
// Rationale: tests that depend on live market data must not silently skip
// during trading hours — that would mask real regressions. This file provides
// helpers to check market status via the `market_state` API and to resolve
// fresh identifiers (option contracts, HK warrants) at test time.
//
// Behavior contract for callers:
//   - In-hours + empty response → t.Fatal (real bug).
//   - Out-of-hours + empty response → t.Logf (wire path was exercised; test
//     should early-return without asserting on missing data). Tests must NOT
//     skip based on trading hours — the wire round-trip is a real assertion.
//   - Resolver failure (network/API) always propagates as skip or fatal per
//     caller policy; helpers themselves do not call t.Skip or t.Fatal to
//     keep decisions at the call site.

package integtest

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/tigerfintech/openapi-go-sdk/model"
	"github.com/tigerfintech/openapi-go-sdk/quote"
)

// Diagnostic renders compact request/response context for integration-test
// failures. It intentionally truncates large live payloads so CI logs stay
// readable while still showing enough data to reproduce the call.
func Diagnostic(request, response interface{}) string {
	return fmt.Sprintf("request=%s response=%s", compactJSON(request), compactJSON(response))
}

func compactJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%+v", v)
	}
	s := string(b)
	if len(s) > 3000 {
		return s[:3000] + "...<truncated>"
	}
	return s
}

// marketStateCache memoizes the market_state result within a single test
// process so we don't hammer the gateway when many tests each guard on
// trading hours.
var marketStateCache sync.Map // map[string]marketStateEntry

type marketStateEntry struct {
	status string
	err    error
}

// getMarketStatus returns the raw status string for a market, e.g. "TRADING",
// "PRE_HOUR_TRADING", "MARKET_CLOSED". Cached per-run.
func getMarketStatus(qc *quote.QuoteClient, market string) (string, error) {
	if v, ok := marketStateCache.Load(market); ok {
		e := v.(marketStateEntry)
		return e.status, e.err
	}
	entry := marketStateEntry{}
	states, err := qc.GetMarketState(market)
	if err != nil {
		entry.err = err
		// Do NOT cache error state — a transient gateway hiccup should not
		// permanently poison the cache for the lifetime of the test process.
		return entry.status, entry.err
	} else if len(states) == 0 {
		entry.err = fmt.Errorf("market_state returned empty for %s", market)
		// Do NOT cache empty-response errors either — same transient reasoning.
		return entry.status, entry.err
	} else {
		// Prefer `status` (e.g. TRADING); fall back to marketStatus which is
		// the same value on newer gateway builds.
		entry.status = states[0].Status
		if entry.status == "" {
			entry.status = states[0].MarketStatus
		}
	}
	marketStateCache.Store(market, entry)
	return entry.status, entry.err
}

// IsMarketTrading reports whether the given market is currently in the
// regular trading session. Returns false on error (fail closed — treat
// unreachable status as non-trading so out-of-hours skips don't accidentally
// promote to hard failures during a gateway hiccup).
func IsMarketTrading(qc *quote.QuoteClient, market string) bool {
	status, err := getMarketStatus(qc, market)
	if err != nil {
		return false
	}
	return status == "TRADING"
}

// IsMarketOpenExtended reports whether the market is open in either regular
// or extended (pre/post) hours. Used for overnight/extended-hour APIs.
func IsMarketOpenExtended(qc *quote.QuoteClient, market string) bool {
	status, err := getMarketStatus(qc, market)
	if err != nil {
		return false
	}
	switch status {
	case "TRADING", "PRE_HOUR_TRADING", "POST_HOUR_TRADING", "MIDDLE_CLOSE":
		return true
	}
	return false
}

// MarketStatus returns the raw status string, mainly for use in Skipf
// messages so the reader can see why a test skipped.
func MarketStatus(qc *quote.QuoteClient, market string) string {
	status, err := getMarketStatus(qc, market)
	if err != nil {
		return fmt.Sprintf("unknown (%v)", err)
	}
	if status == "" {
		return "unknown"
	}
	return status
}

// usSpotPrice returns AAPL's latest price, used to anchor ATM strike selection.
func usSpotPrice(qc *quote.QuoteClient) (float64, error) {
	briefs, err := qc.GetRealTimeQuote(model.BriefRequest{Symbols: []string{"AAPL"}})
	if err != nil {
		return 0, fmt.Errorf("GetRealTimeQuote AAPL: %w", err)
	}
	if len(briefs) == 0 || briefs[0].LatestPrice <= 0 {
		return 0, fmt.Errorf("no usable AAPL spot price")
	}
	return briefs[0].LatestPrice, nil
}

// pickNearExpiry returns the nearest AAPL expiry that is not same-day, plus its
// timestamp. Dates come back ascending; index 0 can be an expiring-today contract
// whose quotes are degenerate, so prefer index 1 when available.
//
// Do NOT pick a mid-range expiry: AAPL lists LEAPS a year out, so len(Dates)/2
// lands on a far-dated contract with no daily bars and delta≈0, which made
// TestIntegGetOptionKline / ...GreeksAndStrikeConsistency fail during live hours.
func pickNearExpiry(exps []model.OptionExpiration) (expiry string, expiryTs int64, err error) {
	if len(exps) == 0 || len(exps[0].Dates) == 0 {
		return "", 0, fmt.Errorf("no option expirations returned for AAPL")
	}
	now := time.Now().UnixMilli()
	min := now + int64(14*24*time.Hour/time.Millisecond)
	max := now + int64(42*24*time.Hour/time.Millisecond)
	for i, ts := range exps[0].Timestamps {
		if ts >= min && ts <= max && i < len(exps[0].Dates) {
			return exps[0].Dates[i], ts, nil
		}
	}
	idx := 0
	if len(exps[0].Dates) > 1 {
		idx = 1
	}
	expiry = exps[0].Dates[idx]
	if idx < len(exps[0].Timestamps) {
		expiryTs = exps[0].Timestamps[idx]
	}
	return expiry, expiryTs, nil
}

// ResolveUsOptionExpiry returns a near-term AAPL option expiry, preferring
// contracts expiring in the next 2-6 weeks to avoid same-day and far-dated LEAPS.
func ResolveUsOptionExpiry(qc *quote.QuoteClient) (string, int64, error) {
	exps, err := qc.GetOptionExpiration([]string{"AAPL"})
	if err != nil {
		return "", 0, fmt.Errorf("GetOptionExpiration: %w", err)
	}
	return pickNearExpiry(exps)
}

func legIsActive(leg *model.OptionLeg) bool {
	return leg != nil && (leg.Volume > 0 || leg.LatestPrice > 0 || (leg.BidPrice > 0 && leg.AskPrice > 0))
}

func collectNearAtmLegs(rows []model.OptionChainRow, spot float64) []*model.OptionLeg {
	mid := len(rows) / 2
	if spot > 0 {
		mid = pickAtmRow(rows, spot)
	}
	offsets := []int{0, 1, -1, 2, -2, 3, -3, 4, -4, 5, -5}
	active := make([]*model.OptionLeg, 0, len(offsets)*2)
	fallback := make([]*model.OptionLeg, 0, len(offsets)*2)
	for _, offset := range offsets {
		i := mid + offset
		if i < 0 || i >= len(rows) {
			continue
		}
		for _, leg := range []*model.OptionLeg{rows[i].Call, rows[i].Put} {
			if leg == nil || leg.Identifier == "" {
				continue
			}
			if legIsActive(leg) {
				active = append(active, leg)
			} else {
				fallback = append(fallback, leg)
			}
		}
	}
	return append(active, fallback...)
}

// pickAtmRow returns the index of the chain row whose strike is closest to spot.
// Falls back to the structural middle row when no strike parses.
//
// The chain spans strikes far wider than spot (AAPL lists ~50 to ~400), so the
// median row is deep out-of-the-money and carries delta=0.
func pickAtmRow(rows []model.OptionChainRow, spot float64) int {
	best, bestDiff := -1, 0.0
	for i, r := range rows {
		var s string
		switch {
		case r.Call != nil:
			s = r.Call.Strike
		case r.Put != nil:
			s = r.Put.Strike
		default:
			continue
		}
		v, e := strconv.ParseFloat(s, 64)
		if e != nil {
			continue
		}
		diff := math.Abs(v - spot)
		if best == -1 || diff < bestDiff {
			best, bestDiff = i, diff
		}
	}
	if best == -1 {
		return len(rows) / 2
	}
	return best
}

// ResolveUsOptionIdentifier returns a live US option identifier for AAPL,
// picked as close to at-the-money as the chain permits. Never calls t.Skip
// or t.Fatal — callers decide policy based on whether the market is trading.
func ResolveUsOptionIdentifier(qc *quote.QuoteClient) (string, error) {
	exps, err := qc.GetOptionExpiration([]string{"AAPL"})
	if err != nil {
		return "", fmt.Errorf("GetOptionExpiration: %w", err)
	}
	expiry, _, err := pickNearExpiry(exps)
	if err != nil {
		return "", err
	}

	chain, err := qc.GetOptionChain([][2]string{{"AAPL", expiry}})
	if err != nil {
		return "", fmt.Errorf("GetOptionChain %s: %w", expiry, err)
	}
	if len(chain) == 0 || len(chain[0].Items) == 0 {
		return "", fmt.Errorf("empty option chain for AAPL %s", expiry)
	}

	// Anchor on spot so the pick is genuinely near-the-money and therefore
	// liquid enough to carry daily bars and non-zero Greeks. A spot lookup
	// failure is non-fatal: fall back to the structural middle.
	rows := chain[0].Items
	spot := 0.0
	if s, e := usSpotPrice(qc); e == nil {
		spot = s
	}
	for _, leg := range collectNearAtmLegs(rows, spot) {
		return leg.Identifier, nil
	}
	// Last resort: any identifier in the chain.
	for _, r := range rows {
		if r.Call != nil && r.Call.Identifier != "" {
			return r.Call.Identifier, nil
		}
		if r.Put != nil && r.Put.Identifier != "" {
			return r.Put.Identifier, nil
		}
	}
	return "", fmt.Errorf("no identifiers in AAPL %s chain", expiry)
}

// ResolveUsOptionLeg returns a live option chain leg for AAPL near ATM,
// along with the expiry timestamp needed by the trade-tick / timeline /
// depth APIs. Prefers the call leg.
func ResolveUsOptionLeg(qc *quote.QuoteClient) (leg model.OptionLeg, expiryTs int64, err error) {
	exps, err := qc.GetOptionExpiration([]string{"AAPL"})
	if err != nil {
		return model.OptionLeg{}, 0, fmt.Errorf("GetOptionExpiration: %w", err)
	}
	expiry, expiryTs, err := pickNearExpiry(exps)
	if err != nil {
		return model.OptionLeg{}, 0, err
	}

	chain, err := qc.GetOptionChain([][2]string{{"AAPL", expiry}})
	if err != nil {
		return model.OptionLeg{}, 0, fmt.Errorf("GetOptionChain %s: %w", expiry, err)
	}
	if len(chain) == 0 || len(chain[0].Items) == 0 {
		return model.OptionLeg{}, 0, fmt.Errorf("empty option chain for AAPL %s", expiry)
	}

	// Anchor on spot rather than the median strike, and prefer active legs so
	// realtime tick/timeline tests avoid cold contracts.
	rows := chain[0].Items
	spot := 0.0
	if s, e := usSpotPrice(qc); e == nil {
		spot = s
	}
	for _, leg := range collectNearAtmLegs(rows, spot) {
		return *leg, expiryTs, nil
	}
	return model.OptionLeg{}, 0, fmt.Errorf("no leg data in AAPL %s chain", expiry)
}

// ResolveHkWarrantSymbol returns a live HK warrant symbol via WarrantFilter
// on 00700 (Tencent). Underlying is Tencent because it has deep, liquid
// warrant coverage; if 00700 has zero warrants returned, the API itself is
// likely gated for this account, and callers should skip.
func ResolveHkWarrantSymbol(qc *quote.QuoteClient) (string, error) {
	filter, err := qc.GetWarrantFilter(model.WarrantFilterRequest{Symbol: "00700", PageSize: 5})
	if err != nil {
		return "", fmt.Errorf("GetWarrantFilter: %w", err)
	}
	if filter == nil || len(filter.Items) == 0 || filter.Items[0].Symbol == "" {
		return "", fmt.Errorf("no warrants returned for 00700")
	}
	return filter.Items[0].Symbol, nil
}

// AssertNonEmptyDuringTrading enforces the following policy for an empty
// response body from a real-time market-data endpoint:
//   - In-hours (status=TRADING): call t.Fatalf — empty is a real regression.
//   - Out-of-hours: call t.Logf and return false — the wire round-trip already
//     validated the request/response path, so the caller must NOT skip and
//     should early-return from the test without asserting on missing data.
//
// AssertNonEmptyDuringTrading checks that an empty API response is only
// acceptable outside trading hours. Call this when len(items) == 0.
//
//   - In-hours (status=TRADING): calls t.Fatalf — empty is a real regression.
//   - Out-of-hours: logs the empty result and returns — the caller should
//     early-return without asserting on missing data.
//
// The function never returns true; the bool return is retained for backward
// compatibility with callers that use the idiomatic pattern:
//
//	if len(items) == 0 {
//	    integtest.AssertNonEmptyDuringTrading(t, qc, "US", "…")
//	    return
//	}
func AssertNonEmptyDuringTrading(t testing.TB, qc *quote.QuoteClient, market, reason string, args ...interface{}) bool {
	t.Helper()
	msg := fmt.Sprintf(reason, args...)
	if IsMarketTrading(qc, market) {
		t.Fatalf("%s (market=%s status=TRADING — expected data)", msg, market)
	}
	t.Logf("%s (market=%s status=%s — empty is expected outside trading hours; wire path validated)", msg, market, MarketStatus(qc, market))
	return false
}

// AssertNonEmptyDuringExtended is like AssertNonEmptyDuringTrading but treats
// pre/post-market as valid data windows. Use for overnight / extended-hour
// APIs. Same no-skip semantics: out-of-hours logs and returns; the caller
// must early-return without further data assertions.
func AssertNonEmptyDuringExtended(t testing.TB, qc *quote.QuoteClient, market, reason string, args ...interface{}) bool {
	t.Helper()
	msg := fmt.Sprintf(reason, args...)
	if IsMarketOpenExtended(qc, market) {
		t.Fatalf("%s (market=%s status=%s — expected data in extended hours)", msg, market, MarketStatus(qc, market))
	}
	t.Logf("%s (market=%s status=%s — empty is expected outside extended hours; wire path validated)", msg, market, MarketStatus(qc, market))
	return false
}
