// Trading-hours + dynamic identifier resolvers for integration tests.
//
// Rationale: tests that depend on live market data must not silently skip
// during trading hours — that would mask real regressions. This file provides
// helpers to check market status via the `market_state` API and to resolve
// fresh identifiers (option contracts, HK warrants) at test time.
//
// Behavior contract for callers:
//   - In-hours + empty response → t.Fatal (real bug).
//   - Out-of-hours + empty response → t.Skip (data-driven, expected).
//   - Resolver failure (network/API) always propagates as skip or fatal per
//     caller policy; helpers themselves do not call t.Skip or t.Fatal to
//     keep decisions at the call site.

package integtest

import (
	"fmt"
	"sort"
	"strconv"
	"sync"
	"testing"

	"github.com/tigerfintech/openapi-go-sdk/model"
	"github.com/tigerfintech/openapi-go-sdk/quote"
)

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
	} else if len(states) == 0 {
		entry.err = fmt.Errorf("market_state returned empty for %s", market)
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

// ResolveUsOptionIdentifier returns a live US option identifier for AAPL,
// picked as close to at-the-money as the chain permits. Never calls t.Skip
// or t.Fatal — callers decide policy based on whether the market is trading.
func ResolveUsOptionIdentifier(qc *quote.QuoteClient) (string, error) {
	exps, err := qc.GetOptionExpiration([]string{"AAPL"})
	if err != nil {
		return "", fmt.Errorf("GetOptionExpiration: %w", err)
	}
	if len(exps) == 0 || len(exps[0].Dates) == 0 {
		return "", fmt.Errorf("no option expirations returned for AAPL")
	}
	// Pick a mid-range expiry: nearest is often illiquid on expiry day.
	idx := len(exps[0].Dates) / 2
	if idx == 0 && len(exps[0].Dates) > 1 {
		idx = 1
	}
	expiry := exps[0].Dates[idx]

	chain, err := qc.GetOptionChain([][2]string{{"AAPL", expiry}})
	if err != nil {
		return "", fmt.Errorf("GetOptionChain %s: %w", expiry, err)
	}
	if len(chain) == 0 || len(chain[0].Items) == 0 {
		return "", fmt.Errorf("empty option chain for AAPL %s", expiry)
	}

	// Prefer ATM (nearest to spot). We don't have spot here; the chain is
	// typically returned sorted by strike, so pick the middle row and its
	// call leg. Fall back to put if call is missing.
	rows := chain[0].Items
	mid := len(rows) / 2
	for _, offset := range []int{0, 1, -1, 2, -2, 3, -3} {
		i := mid + offset
		if i < 0 || i >= len(rows) {
			continue
		}
		if rows[i].Call != nil && rows[i].Call.Identifier != "" {
			return rows[i].Call.Identifier, nil
		}
		if rows[i].Put != nil && rows[i].Put.Identifier != "" {
			return rows[i].Put.Identifier, nil
		}
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
	if len(exps) == 0 || len(exps[0].Dates) == 0 {
		return model.OptionLeg{}, 0, fmt.Errorf("no option expirations for AAPL")
	}
	idx := len(exps[0].Dates) / 2
	if idx == 0 && len(exps[0].Dates) > 1 {
		idx = 1
	}
	expiry := exps[0].Dates[idx]
	if idx < len(exps[0].Timestamps) {
		expiryTs = exps[0].Timestamps[idx]
	}

	chain, err := qc.GetOptionChain([][2]string{{"AAPL", expiry}})
	if err != nil {
		return model.OptionLeg{}, 0, fmt.Errorf("GetOptionChain %s: %w", expiry, err)
	}
	if len(chain) == 0 || len(chain[0].Items) == 0 {
		return model.OptionLeg{}, 0, fmt.Errorf("empty option chain for AAPL %s", expiry)
	}

	// Try to pick a leg near the median strike. Strikes come as strings; parse
	// as best effort. If parsing fails we fall back to structural mid.
	rows := chain[0].Items
	type legScore struct {
		strike float64
		row    int
	}
	strikes := make([]legScore, 0, len(rows))
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
		if v, e := strconv.ParseFloat(s, 64); e == nil {
			strikes = append(strikes, legScore{v, i})
		}
	}
	pickRow := len(rows) / 2
	if len(strikes) > 0 {
		sort.Slice(strikes, func(i, j int) bool { return strikes[i].strike < strikes[j].strike })
		pickRow = strikes[len(strikes)/2].row
	}
	row := rows[pickRow]
	if row.Call != nil {
		return *row.Call, expiryTs, nil
	}
	if row.Put != nil {
		return *row.Put, expiryTs, nil
	}
	return model.OptionLeg{}, 0, fmt.Errorf("no leg data at row %d", pickRow)
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

// FailOrSkip is a small convenience: if the market is trading, call t.Fatalf
// with reason; otherwise t.Skipf. Used at call sites that want the standard
// in-hours=fail / out-of-hours=skip policy without repeating the branch.
func FailOrSkip(t testing.TB, qc *quote.QuoteClient, market, reason string, args ...interface{}) {
	t.Helper()
	msg := fmt.Sprintf(reason, args...)
	if IsMarketTrading(qc, market) {
		t.Fatalf("%s (market=%s status=TRADING — expected data)", msg, market)
	}
	t.Skipf("%s (market=%s status=%s — expected outside trading hours)", msg, market, MarketStatus(qc, market))
}

// FailOrSkipExtended is like FailOrSkip but treats pre/post-market as valid
// data windows. Use for overnight / extended-hour APIs.
func FailOrSkipExtended(t testing.TB, qc *quote.QuoteClient, market, reason string, args ...interface{}) {
	t.Helper()
	msg := fmt.Sprintf(reason, args...)
	if IsMarketOpenExtended(qc, market) {
		t.Fatalf("%s (market=%s status=%s — expected data in extended hours)", msg, market, MarketStatus(qc, market))
	}
	t.Skipf("%s (market=%s status=%s — expected outside extended hours)", msg, market, MarketStatus(qc, market))
}
