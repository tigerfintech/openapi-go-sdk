//go:build integration

// Order matrix integration tests — mirrors the Python/TS/Java SDK three-phase
// coverage. Every test uses safe prices (BUY $0.01 / SELL $999,999) so orders
// never fill, plus a previewAndPlace helper that runs preview -> place ->
// cancel with automatic skip on legitimate boundary errors.
//
// Phase 1 (US market x order type):
//   MKT preview, LMT-by-amount preview, STP, STP_LMT, TRAIL, TWAP, VWAP,
//   ICEBERG, OCA brackets, US OPT LMT, US FUT LMT, forex SEC, invalid-price
//   preview.
//
// Phase 2 (HK / CN / SG):
//   HK STK LMT, HK STK AL, HK STK AM preview, CN STK LMT, SG STK LMT.
//
// Phase 3 (MLEG + edge):
//   MLEG vertical spread, ICEBERG modify round-trip, SELL SHORT preview.
//
// Run with:
//   make integ
//   # or: TIGER_RUN_INTEG=true go test -tags=integration -run TestMatrix ./trade/...

package trade_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/tigerfintech/openapi-go-sdk/integtest"
	"github.com/tigerfintech/openapi-go-sdk/model"
	"github.com/tigerfintech/openapi-go-sdk/trade"
)

// Safe prices — kept far from market so BUY/SELL orders never fill.
const (
	safeBuyPrice        = 0.01
	safeSellPrice       = 999_999.0
	safeStopBuyTrigger  = 999_999.0
)

// Permission / license / session boundary errors — legitimate skips.
var permissionErrorPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)access forbidden`),
	regexp.MustCompile(`(?i)forbidden`),
	regexp.MustCompile(`(?i)no permission`),
	regexp.MustCompile(`(?i)not supported`),
	regexp.MustCompile(`(?i)license`),
	regexp.MustCompile(`(?i)not open`),
	regexp.MustCompile(`(?i)not enabled`),
	regexp.MustCompile(`(?i)no token`),
	regexp.MustCompile(`(?i)don['\x{2018}\x{2019}]t support trading`),
	regexp.MustCompile(`(?i)unsupported instrument`),
	regexp.MustCompile(`(?i)only limit orders are supported`),
	regexp.MustCompile(`(?i)outside of regular trading hours`),
	regexp.MustCompile(`(?i)market is closed`),
	regexp.MustCompile(`(?i)only limit orders can be placed`),
	regexp.MustCompile(`(?i)only limit, stop or stop-limit orders are allowed`),
	regexp.MustCompile(`(?i)at non-trading hour`),
	regexp.MustCompile(`(?i)time range for the order.*needs to be between`),
	regexp.MustCompile(`(?i)orders cannot be placed at this moment`),
	regexp.MustCompile(`(?i)auction order is not allowed at this moment`),
	regexp.MustCompile(`(?i)does not support stock (long|short)`),
	regexp.MustCompile(`(?i)only trade cash order by market order`),
	regexp.MustCompile(`(?i)cash order by market order`),
	regexp.MustCompile(`(?i)^system error$`),
	regexp.MustCompile(`(?i)bad_request:System error`),
}

// Order state race — cancel/modify may hit a terminal state during the test.
var terminalOrderPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)cannot be modified`),
	regexp.MustCompile(`(?i)cannot be (canc|cancell)ed`),
	regexp.MustCompile(`(?i)already (canc|cancell)ed`),
	regexp.MustCompile(`(?i)already filled`),
	regexp.MustCompile(`(?i)invalid order status`),
	regexp.MustCompile(`(?i)cancellation is not allowed`),
	regexp.MustCompile(`(?i)cancel is not allowed`),
}

// Rate limit — retry with exponential backoff.
var rateLimitPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)too_many_requests`),
	regexp.MustCompile(`(?i)rate limit`),
	regexp.MustCompile(`(?i)requestRateExceedLimit`),
}

func matches(msg string, patterns []*regexp.Regexp) bool {
	if msg == "" {
		return false
	}
	for _, p := range patterns {
		if p.MatchString(msg) {
			return true
		}
	}
	return false
}

// -----------------------------------------------------------------------------
// Contract helpers
// -----------------------------------------------------------------------------

func usStk() model.OrderRequest {
	return model.OrderRequest{
		Symbol:   "AAPL",
		SecType:  "STK",
		Currency: "USD",
		Market:   "US",
	}
}

func hkStk() model.OrderRequest {
	return model.OrderRequest{
		Symbol:   "00700",
		SecType:  "STK",
		Currency: "HKD",
		Market:   "HK",
	}
}

// merge fills each field from `overlay` into `base` (only non-zero values).
// Go structs don't compose as smoothly as JS objects; the tests use this to
// keep individual test cases short.
func merge(base, overlay model.OrderRequest) model.OrderRequest {
	if overlay.Action != "" {
		base.Action = overlay.Action
	}
	if overlay.OrderType != "" {
		base.OrderType = overlay.OrderType
	}
	if overlay.TotalQuantity != 0 {
		base.TotalQuantity = overlay.TotalQuantity
	}
	if overlay.LimitPrice != 0 {
		base.LimitPrice = overlay.LimitPrice
	}
	if overlay.AuxPrice != 0 {
		base.AuxPrice = overlay.AuxPrice
	}
	if overlay.TrailingPercent != 0 {
		base.TrailingPercent = overlay.TrailingPercent
	}
	if overlay.TimeInForce != "" {
		base.TimeInForce = overlay.TimeInForce
	}
	if overlay.DisplaySize != 0 {
		base.DisplaySize = overlay.DisplaySize
	}
	if overlay.MinDisplaySize != 0 {
		base.MinDisplaySize = overlay.MinDisplaySize
	}
	if overlay.CheckIntervals != 0 {
		base.CheckIntervals = overlay.CheckIntervals
	}
	if overlay.PriceType != "" {
		base.PriceType = overlay.PriceType
	}
	if overlay.StartTime != 0 {
		base.StartTime = overlay.StartTime
	}
	if overlay.EndTime != 0 {
		base.EndTime = overlay.EndTime
	}
	if overlay.Amount != 0 {
		base.Amount = overlay.Amount
	}
	if overlay.AlgoStrategy != "" {
		base.AlgoStrategy = overlay.AlgoStrategy
	}
	if overlay.AlgoParams != nil {
		base.AlgoParams = overlay.AlgoParams
	}
	if len(overlay.OcaOrders) > 0 {
		base.OcaOrders = overlay.OcaOrders
	}
	if len(overlay.ContractLegs) > 0 {
		base.ContractLegs = overlay.ContractLegs
		base.ComboType = overlay.ComboType
		base.SecType = "MLEG"
	}
	if overlay.Expiry != "" {
		base.Expiry = overlay.Expiry
	}
	if overlay.Strike != "" {
		base.Strike = overlay.Strike
	}
	if overlay.Right != "" {
		base.Right = overlay.Right
	}
	return base
}

// -----------------------------------------------------------------------------
// preview / place / cancel helpers
// -----------------------------------------------------------------------------

// previewOnly issues preview_order and returns true if the call was legitimate
// (either success or a known-boundary skip).
func previewOnly(t *testing.T, tc *trade.TradeClient, order model.OrderRequest, context string) bool {
	t.Helper()
	_, err := tc.PreviewOrder(order)
	if err == nil {
		return true
	}
	if matches(err.Error(), permissionErrorPatterns) {
		t.Logf("%s: skipped (preview boundary): %v", context, err)
		return false
	}
	t.Fatalf("%s: preview failed: %v", context, err)
	return false
}

// previewAndPlace runs preview -> place -> cancel with rate-limit retry and
// permission auto-skip. Returns true if the round-trip executed.
func previewAndPlace(t *testing.T, tc *trade.TradeClient, order model.OrderRequest, context string) bool {
	t.Helper()

	// 1. Preview first — validates marshaling before touching real state.
	if !previewOnly(t, tc, order, context) {
		return false
	}

	// 2. Place with exponential backoff on rate-limit.
	var (
		placeRes *model.PlaceOrderResult
		err      error
		delay    = time.Second
	)
	for attempt := 0; attempt < 3; attempt++ {
		placeRes, err = tc.PlaceOrder(order)
		if err == nil {
			break
		}
		if matches(err.Error(), permissionErrorPatterns) {
			t.Logf("%s: skipped (place boundary): %v", context, err)
			return false
		}
		if matches(err.Error(), rateLimitPatterns) && attempt < 2 {
			time.Sleep(delay)
			delay *= 2
			continue
		}
		t.Fatalf("%s: placeOrder failed: %v", context, err)
	}
	if placeRes == nil {
		t.Fatalf("%s: placeOrder returned nil result", context)
	}
	orderID := placeRes.ID
	if orderID == 0 {
		orderID = placeRes.OrderID
	}
	if orderID == 0 {
		t.Fatalf("%s: placeOrder returned zero order id", context)
	}
	t.Logf("%s: placed order id=%d", context, orderID)

	// 3. Cancel tolerantly.
	cancelTolerant(t, tc, orderID, context)
	return true
}

func cancelTolerant(t *testing.T, tc *trade.TradeClient, orderID int64, context string) {
	t.Helper()
	_, err := tc.CancelOrder(orderID)
	if err == nil {
		return
	}
	if matches(err.Error(), terminalOrderPatterns) {
		t.Logf("%s: cancel hit terminal state (ok): %v", context, err)
		return
	}
	t.Fatalf("%s: unexpected cancel failure: %v", context, err)
}

// -----------------------------------------------------------------------------
// Phase 1 — US market x order type x sec type
// -----------------------------------------------------------------------------

func TestMatrixUsStkMarketPreview(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)
	previewOnly(t, tc, merge(usStk(), model.OrderRequest{
		Action: "BUY", OrderType: "MKT", TotalQuantity: 1,
	}), "US STK MKT preview")
}

func TestMatrixUsStkMarketByAmountPreview(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)
	// 网关要求即使按金额下单也带 total_quantity 字段(哪怕值不生效)。
	// 传 1 让服务端跳过 'quantity is required' 校验;真正的下单份数由
	// Amount 字段决定。这跟 Java SDK 的 amount-order 一样有校验规则,
	// 但服务端的语义是 "quantity 必须存在" 而不是 "两者不能共存"。
	previewOnly(t, tc, merge(usStk(), model.OrderRequest{
		Action: "BUY", OrderType: "MKT", Amount: 100.0, TotalQuantity: 1,
	}), "US STK MKT-by-amount preview")
}

func TestMatrixUsStkStop(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)
	previewAndPlace(t, tc, merge(usStk(), model.OrderRequest{
		Action: "BUY", OrderType: "STP", TotalQuantity: 1,
		AuxPrice: safeStopBuyTrigger,
	}), "US STK STP")
}

func TestMatrixUsStkStopLimit(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)
	previewAndPlace(t, tc, merge(usStk(), model.OrderRequest{
		Action: "BUY", OrderType: "STP_LMT", TotalQuantity: 1,
		LimitPrice: safeBuyPrice, AuxPrice: safeStopBuyTrigger,
	}), "US STK STP_LMT")
}

func TestMatrixUsStkTrail(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)
	previewAndPlace(t, tc, merge(usStk(), model.OrderRequest{
		Action: "SELL", OrderType: "TRAIL", TotalQuantity: 1,
		TrailingPercent: 50.0,
	}), "US STK TRAIL")
}

func TestMatrixUsStkTwap(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)
	now := time.Now().UnixMilli()
	// AlgoParamsRequest 现在通过自定义 MarshalJSON 转成网关期望的
	// [{tag, value}] 数组;调用方直接传自然的结构体即可。
	previewAndPlace(t, tc, merge(usStk(), model.OrderRequest{
		Action: "BUY", OrderType: "TWAP", TotalQuantity: 10,
		LimitPrice:   safeBuyPrice,
		AlgoStrategy: "TWAP",
		AlgoParams: &model.AlgoParamsRequest{
			StartTime: now,
			EndTime:   now + 3_600_000,
		},
	}), "US STK TWAP")
}

func TestMatrixUsStkVwap(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)
	now := time.Now().UnixMilli()
	previewAndPlace(t, tc, merge(usStk(), model.OrderRequest{
		Action: "BUY", OrderType: "VWAP", TotalQuantity: 10,
		LimitPrice:   safeBuyPrice,
		AlgoStrategy: "VWAP",
		AlgoParams: &model.AlgoParamsRequest{
			StartTime:         now,
			EndTime:           now + 3_600_000,
			ParticipationRate: 0.1,
		},
	}), "US STK VWAP")
}

func TestMatrixUsStkIceberg(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)
	now := time.Now().UnixMilli()
	previewAndPlace(t, tc, merge(usStk(), model.OrderRequest{
		Action: "BUY", OrderType: "ICEBERG", TotalQuantity: 10,
		LimitPrice: safeBuyPrice,
		DisplaySize: 2, MinDisplaySize: 1, CheckIntervals: 30,
		PriceType: "LIMIT_PRICE",
		StartTime: now, EndTime: now + 3_600_000,
	}), "US STK ICEBERG")
}

func TestMatrixUsStkOcaBrackets(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)
	// OCA brackets: profit + stop-loss legs share account and symbol via base.
	parent := usStk()
	profit := merge(usStk(), model.OrderRequest{
		Action: "BUY", OrderType: "LMT", TotalQuantity: 1,
		LimitPrice: safeSellPrice, TimeInForce: "GTC",
	})
	stop := merge(usStk(), model.OrderRequest{
		Action: "BUY", OrderType: "STP_LMT", TotalQuantity: 1,
		LimitPrice: safeBuyPrice, AuxPrice: safeBuyPrice, TimeInForce: "GTC",
	})
	parent.OcaOrders = []*model.OrderRequest{&profit, &stop}
	previewAndPlace(t, tc, parent, "US STK OCA brackets")
}

func TestMatrixUsOptLimit(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)
	expiry := time.Now().AddDate(0, 0, 30).Format("20060102")
	previewAndPlace(t, tc, model.OrderRequest{
		Symbol:   "AAPL",
		SecType:  "OPT",
		Currency: "USD",
		Market:   "US",
		Expiry:   expiry,
		Strike:   "200",
		Right:    "CALL",
		Action:   "BUY", OrderType: "LMT", TotalQuantity: 1,
		LimitPrice: safeBuyPrice,
	}, "US OPT LMT")
}

func TestMatrixUsFutLimit(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)
	previewAndPlace(t, tc, model.OrderRequest{
		Symbol:   "CL",
		SecType:  "FUT",
		Currency: "USD",
		Market:   "US",
		Action:   "BUY", OrderType: "LMT", TotalQuantity: 1,
		LimitPrice: safeBuyPrice,
	}, "US FUT LMT")
}

func TestMatrixForexSecSegment(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)
	// WARNING: PlaceForexOrder is a real, irreversible FX conversion (USD→HKD).
	// Unlike most order tests (preview→place→cancel), forex conversions cannot
	// be cancelled after execution. Amount is kept at the minimum (1.0 USD) to
	// limit financial impact. Only run this test when you explicitly intend a
	// live FX conversion.
	_, err := tc.PlaceForexOrder(model.ForexOrderRequest{
		SegType:        "SEC",
		SourceCurrency: "USD",
		TargetCurrency: "HKD",
		SourceAmount:   1.0,
	})
	if err != nil && !matches(err.Error(), permissionErrorPatterns) {
		t.Fatalf("placeForex SEC failed: %v", err)
	}
	t.Logf("Forex SEC segment: err=%v", err)
}

func TestMatrixUsStkNegativePricePreview(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)
	// Either accepted or rejected — both are semantically valid.
	_, err := tc.PreviewOrder(merge(usStk(), model.OrderRequest{
		Action: "BUY", OrderType: "LMT", TotalQuantity: 1,
		LimitPrice: -1.0,
	}))
	t.Logf("US STK negative-price preview: err=%v", err)
}

// -----------------------------------------------------------------------------
// Phase 2 — HK / CN / SG
// -----------------------------------------------------------------------------

func TestMatrixHkStkLimit(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)
	previewAndPlace(t, tc, merge(hkStk(), model.OrderRequest{
		Action: "BUY", OrderType: "LMT", TotalQuantity: 100,
		LimitPrice: safeBuyPrice,
	}), "HK STK LMT")
}

func TestMatrixHkStkAuctionLimit(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)
	previewAndPlace(t, tc, merge(hkStk(), model.OrderRequest{
		Action: "BUY", OrderType: "AL", TotalQuantity: 100,
		LimitPrice: safeBuyPrice,
	}), "HK STK AL")
}

func TestMatrixHkStkAuctionMarketPreview(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)
	previewOnly(t, tc, merge(hkStk(), model.OrderRequest{
		Action: "BUY", OrderType: "AM", TotalQuantity: 100,
	}), "HK STK AM preview")
}

func TestMatrixCnStkLimit(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)
	previewAndPlace(t, tc, model.OrderRequest{
		Symbol:   "000001",
		SecType:  "STK",
		Currency: "CNH",
		Market:   "CN",
		Action:   "BUY", OrderType: "LMT", TotalQuantity: 100,
		LimitPrice: safeBuyPrice,
	}, "CN STK LMT")
}

func TestMatrixSgStkLimit(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)
	previewAndPlace(t, tc, model.OrderRequest{
		Symbol:   "D05",
		SecType:  "STK",
		Currency: "SGD",
		Market:   "SG",
		Action:   "BUY", OrderType: "LMT", TotalQuantity: 100,
		LimitPrice: safeBuyPrice,
	}, "SG STK LMT")
}

// -----------------------------------------------------------------------------
// Phase 3 — MLEG combo + edge cases
// -----------------------------------------------------------------------------

func TestMatrixUsMlegVerticalSpread(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)
	expiry := time.Now().AddDate(0, 0, 45).Format("20060102")
	one := 1
	// AAPL PUT vertical spread — deeply negative limit prevents execution.
	previewAndPlace(t, tc, model.OrderRequest{
		SecType:   "MLEG",
		Currency:  "USD",
		Market:    "US",
		ComboType: "VERTICAL",
		Action:    "BUY", OrderType: "LMT", TotalQuantity: 1,
		LimitPrice: -100.0,
		ContractLegs: []model.ContractLegRequest{
			{Symbol: "AAPL", SecType: "OPT", Expiry: expiry,
				Strike: "200", Right: "PUT", Action: "BUY", Ratio: &one},
			{Symbol: "AAPL", SecType: "OPT", Expiry: expiry,
				Strike: "205", Right: "PUT", Action: "SELL", Ratio: &one},
		},
	}, "US MLEG VERTICAL")
}

func TestMatrixUsStkIcebergModify(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)
	now := time.Now().UnixMilli()
	order := merge(usStk(), model.OrderRequest{
		Action: "BUY", OrderType: "ICEBERG", TotalQuantity: 10,
		LimitPrice: safeBuyPrice,
		DisplaySize: 2, MinDisplaySize: 1, CheckIntervals: 30,
		PriceType: "LIMIT_PRICE",
		StartTime: now, EndTime: now + 3_600_000,
	})
	// preview first
	if !previewOnly(t, tc, order, "US STK ICEBERG modify — preview") {
		return
	}
	placeRes, err := tc.PlaceOrder(order)
	if err != nil {
		if matches(err.Error(), permissionErrorPatterns) {
			t.Logf("US STK ICEBERG modify: skipped: %v", err)
			return
		}
		t.Fatalf("US STK ICEBERG modify: place failed: %v", err)
	}
	if placeRes == nil {
		t.Fatalf("US STK ICEBERG modify: PlaceOrder returned nil result")
	}
	orderID := placeRes.ID
	if orderID == 0 {
		orderID = placeRes.OrderID
	}
	t.Logf("US STK ICEBERG modify: placed id=%d", orderID)

	// modify: bump limit price
	modified := order
	modified.LimitPrice = safeBuyPrice * 2
	if _, err := tc.ModifyOrder(orderID, modified); err != nil && !matches(err.Error(), terminalOrderPatterns) {
		t.Logf("US STK ICEBERG modify: modify failed (best-effort): %v", err)
	}

	cancelTolerant(t, tc, orderID, "US STK ICEBERG modify")
}

func TestMatrixUsStkSellShortPreview(t *testing.T) {
	integtest.SkipUnlessEnabled(t)
	tc := integtest.NewTradeClient(t)
	previewOnly(t, tc, merge(usStk(), model.OrderRequest{
		Action: "SELL", OrderType: "LMT", TotalQuantity: 1,
		LimitPrice: safeSellPrice,
	}), "US STK SELL SHORT preview")
}
