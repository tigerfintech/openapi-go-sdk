// Trade example - 覆盖 TradeClient 全部公开接口,全部强类型响应
//
// 配置从 ./tiger_openapi_config.properties 或 ~/.tigeropen/tiger_openapi_config.properties 自动读取。
// 会真实下单:极低价限价单 BUY 1 股 AAPL @ $1.00(绝不会成交),立即 ModifyOrder 改价,再 CancelOrder 撤销。
// 单个接口失败不中断后续,最后统计 PASS/FAIL。
//
// Run: go run ./examples/trade
package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/tigerfintech/openapi-go-sdk/client"
	"github.com/tigerfintech/openapi-go-sdk/config"
	"github.com/tigerfintech/openapi-go-sdk/model"
	"github.com/tigerfintech/openapi-go-sdk/trade"
)

type result struct {
	name string
	ok   bool
	err  error
}

var results []result

func ok(name, note string) {
	fmt.Printf("[ OK ] %-36s %s\n", name, note)
	results = append(results, result{name: name, ok: true})
}

func fail(name string, err error) {
	fmt.Printf("[FAIL] %-36s %v\n", name, err)
	results = append(results, result{name: name, err: err})
}

func skip(name, reason string) {
	fmt.Printf("[SKIP] %-36s %s\n", name, reason)
	results = append(results, result{name: name, err: fmt.Errorf("skipped: %s", reason)})
}

func main() {
	cfg, err := config.NewClientConfig()
	if err != nil {
		log.Fatal("load config failed: ", err)
	}
	fmt.Printf("tiger_id=%s account=%s\n\n", cfg.TigerID, cfg.Account)

	tc := trade.NewTradeClient(client.NewHttpClient(cfg), cfg.Account)

	fmt.Println("=== Contract 查询 ===")
	if cs, err := tc.Contract("AAPL", "STK"); err != nil {
		fail("Contract(AAPL, STK)", err)
	} else if len(cs) > 0 {
		ok("Contract(AAPL, STK)", fmt.Sprintf("%s contractId=%d exchange=%s", cs[0].Symbol, cs[0].ContractId, cs[0].PrimaryExchange))
	} else {
		ok("Contract(AAPL, STK)", "(empty)")
	}

	if c, err := tc.Contract3("AAPL", "STK"); err != nil {
		fail("Contract3(AAPL, STK)", err)
	} else if c != nil {
		ok("Contract3(AAPL, STK)", fmt.Sprintf("%s contractId=%d exchange=%s", c.Symbol, c.ContractId, c.PrimaryExchange))
	} else {
		ok("Contract3(AAPL, STK)", "(empty)")
	}

	if cs, err := tc.Contracts([]string{"AAPL", "TSLA"}, "STK"); err != nil {
		fail("Contracts([AAPL TSLA])", err)
	} else {
		names := make([]string, 0, len(cs))
		for _, c := range cs {
			names = append(names, c.Symbol)
		}
		ok("Contracts([AAPL TSLA])", fmt.Sprintf("count=%d %s", len(cs), strings.Join(names, ",")))
	}

	if cs, err := tc.QuoteContract("AAPL", "OPT", "20260619"); err != nil {
		fail("QuoteContract(AAPL OPT)", err)
	} else {
		ok("QuoteContract(AAPL OPT)", fmt.Sprintf("count=%d", len(cs)))
	}

	fmt.Println("\n=== 账户/持仓 查询 ===")
	if assets, err := tc.Assets(model.AssetsRequest{}); err != nil {
		fail("Assets", err)
	} else if len(assets) > 0 {
		a := assets[0]
		ok("Assets", fmt.Sprintf("account=%s buyingPower=%.2f netLiquidation=%.2f segments=%d",
			a.Account, a.BuyingPower, a.NetLiquidation, len(a.Segments)))
	} else {
		ok("Assets", "(empty)")
	}

	if pa, err := tc.PrimeAssets(model.AssetsRequest{}); err != nil {
		fail("PrimeAssets", err)
	} else {
		var totalBP float64
		for _, s := range pa.Segments {
			totalBP += s.BuyingPower
		}
		ok("PrimeAssets", fmt.Sprintf("account=%s segments=%d totalBuyingPower=%.2f",
			pa.AccountID, len(pa.Segments), totalBP))
	}

	if ps, err := tc.Positions(model.PositionsRequest{}); err != nil {
		fail("Positions", err)
	} else {
		var totalMV float64
		for _, p := range ps {
			totalMV += p.MarketValue
		}
		ok("Positions", fmt.Sprintf("count=%d totalMarketValue=%.2f", len(ps), totalMV))
	}

	fmt.Println("\n=== 订单 查询 ===")
	if os, err := tc.Orders(model.OrdersRequest{}); err != nil {
		fail("Orders", err)
	} else {
		ok("Orders", fmt.Sprintf("count=%d", len(os)))
	}
	if os, err := tc.ActiveOrders(model.OrdersRequest{}); err != nil {
		fail("ActiveOrders", err)
	} else {
		ok("ActiveOrders", fmt.Sprintf("count=%d", len(os)))
	}
	if os, err := tc.InactiveOrders(model.OrdersRequest{}); err != nil {
		fail("InactiveOrders", err)
	} else {
		ok("InactiveOrders", fmt.Sprintf("count=%d", len(os)))
	}
	now := time.Now().UnixMilli()
	if os, err := tc.FilledOrders(model.OrdersRequest{StartDate: now - 30*24*3600*1000, EndDate: now}); err != nil {
		fail("FilledOrders", err)
	} else {
		ok("FilledOrders", fmt.Sprintf("count=%d (last 30d)", len(os)))
	}

	var existingOrderID int64
	if orders, err := tc.Orders(model.OrdersRequest{}); err == nil && len(orders) > 0 {
		existingOrderID = orders[0].ID
	}
	if existingOrderID != 0 {
		// 按 Order ID 查单笔订单详情（演示新增的 GetOrder）
		if o, err := tc.GetOrder(model.GetOrderRequest{Id: existingOrderID}); err != nil {
			fail(fmt.Sprintf("GetOrder(%d)", existingOrderID), err)
		} else {
			ok(fmt.Sprintf("GetOrder(%d)", existingOrderID), fmt.Sprintf("status=%s symbol=%s", o.Status, o.Symbol))
		}
		// 查成交明细（过滤条件改为全可选，按 OrderId 过滤）
		if txs, err := tc.OrderTransactions(model.OrderTransactionsRequest{OrderId: existingOrderID}); err != nil {
			fail(fmt.Sprintf("OrderTransactions(%d)", existingOrderID), err)
		} else {
			ok(fmt.Sprintf("OrderTransactions(%d)", existingOrderID), fmt.Sprintf("count=%d", len(txs)))
		}
	} else {
		skip("GetOrder / OrderTransactions", "no existing order")
	}

	fmt.Println("\n=== 下单/改单/撤单 ===")
	orderReq := model.LimitOrder(cfg.Account, "AAPL", "STK", "BUY", 1, 1.00)
	orderReq.Market = "US"
	orderReq.Currency = "USD"
	orderReq.TimeInForce = "DAY"

	if preview, err := tc.PreviewOrder(orderReq); err != nil {
		fail("PreviewOrder", err)
	} else {
		ok("PreviewOrder", fmt.Sprintf("isPass=%v commission=%.2f initMargin=%.2f",
			preview.IsPass, preview.Commission, preview.InitMargin))
	}

	placed, err := tc.PlaceOrder(orderReq)
	if err != nil {
		fail("PlaceOrder", err)
		skip("ModifyOrder", "PlaceOrder failed")
		skip("CancelOrder", "PlaceOrder failed")
	} else {
		ok("PlaceOrder", fmt.Sprintf("id=%d orderId=%d", placed.ID, placed.OrderID))

		modifyReq := orderReq
		modifyReq.LimitPrice = 1.50
		if mod, err := tc.ModifyOrder(placed.ID, modifyReq); err != nil {
			fail(fmt.Sprintf("ModifyOrder(%d)", placed.ID), err)
		} else {
			ok(fmt.Sprintf("ModifyOrder(%d)", placed.ID), fmt.Sprintf("id=%d", mod.ID))
		}

		if canc, err := tc.CancelOrder(placed.ID); err != nil {
			fail(fmt.Sprintf("CancelOrder(%d)", placed.ID), err)
		} else {
			ok(fmt.Sprintf("CancelOrder(%d)", placed.ID), fmt.Sprintf("id=%d", canc.ID))
		}
	}

	// ===== v0.3.0 新增 trade 接口 =====
	fmt.Println("\n=== v0.3.0: 账户管理 / 衍生品 ===")

	if accts, err := tc.ManagedAccounts(model.ManagedAccountsRequest{}); err != nil {
		fail("ManagedAccounts", err)
	} else {
		ok("ManagedAccounts", fmt.Sprintf("count=%d", len(accts)))
	}

	if dc, err := tc.DerivativeContracts(model.DerivativeContractsRequest{
		Symbols: []string{"AAPL"}, SecType: "OPT", Expiry: "20260619",
	}); err != nil {
		fail("DerivativeContracts(AAPL OPT)", err)
	} else {
		ok("DerivativeContracts(AAPL OPT)", fmt.Sprintf("count=%d", len(dc)))
	}

	fmt.Println("\n=== v0.3.0: 资产分析 ===")

	if aa, err := tc.AnalyticsAsset(model.AnalyticsAssetRequest{
		SegType: "SEC", StartDate: "2025-05-01", EndDate: "2025-05-07",
	}); err != nil {
		fail("AnalyticsAsset", err)
	} else {
		ok("AnalyticsAsset", fmt.Sprintf("rows=%d", len(aa)))
	}

	if ag, err := tc.AggregateAssets(model.AggregateAssetsRequest{BaseCurrency: "USD"}); err != nil {
		fail("AggregateAssets", err)
	} else if ag != nil {
		ok("AggregateAssets", fmt.Sprintf("netLiquidation=%.2f currencies=%d", ag.NetLiquidation, len(ag.CurrencyAssets)))
	}

	fmt.Println("\n=== v0.3.0: 资金明细/历史 ===")

	if fd, err := tc.FundDetails(model.FundDetailsRequest{
		SegTypes: []string{"SEC"}, Limit: 10,
	}); err != nil {
		fail("FundDetails", err)
	} else {
		ok("FundDetails", fmt.Sprintf("rows=%d", len(fd)))
	}

	if fh, err := tc.FundingHistory(model.FundingHistoryRequest{SegType: "SEC"}); err != nil {
		fail("FundingHistory", err)
	} else {
		ok("FundingHistory", fmt.Sprintf("rows=%d", len(fh)))
	}

	fmt.Println("\n=== v0.3.0: 子账户资金调拨（只读） ===")

	if sf, err := tc.SegmentFundAvailable(model.SegmentFundRequest{
		FromSegment: "SEC", ToSegment: "FUT", Currency: "USD",
	}); err != nil {
		fail("SegmentFundAvailable", err)
	} else {
		ok("SegmentFundAvailable", fmt.Sprintf("rows=%d", len(sf)))
	}

	if sh, err := tc.SegmentFundHistory(model.SegmentFundRequest{Limit: 10}); err != nil {
		fail("SegmentFundHistory", err)
	} else {
		ok("SegmentFundHistory", fmt.Sprintf("rows=%d", len(sh)))
	}

	fmt.Println("\n=== v0.3.0: 内部/外部转股（只读） ===")

	if recs, err := tc.PositionTransferRecords(model.PositionTransferRecordsRequest{
		SinceDate: "2025-01-01", ToDate: "2025-05-07",
	}); err != nil {
		fail("PositionTransferRecords", err)
	} else {
		ok("PositionTransferRecords", fmt.Sprintf("rows=%d", len(recs)))
	}

	if ext, err := tc.PositionTransferExternalRecords(model.PositionTransferExternalRecordsRequest{
		SinceDate: "2025-01-01", ToDate: "2025-05-07",
	}); err != nil {
		fail("PositionTransferExternalRecords", err)
	} else {
		ok("PositionTransferExternalRecords", fmt.Sprintf("rows=%d", len(ext)))
	}

	fmt.Println("\n=== v0.3.8: 期权行权（只读） ===")

	if res, err := tc.OptionExercisePositions(model.OptionExercisePositionRequest{
		Type: "Exercise",
	}); err != nil {
		fail("OptionExercisePositions(Exercise)", err)
	} else if res != nil {
		ok("OptionExercisePositions(Exercise)", fmt.Sprintf("rows=%d pageCount=%d", len(res.Items), res.PageCount))
	} else {
		ok("OptionExercisePositions(Exercise)", "(nil result)")
	}

	if res, err := tc.OptionExercisePositions(model.OptionExercisePositionRequest{
		Type: "Expire",
	}); err != nil {
		fail("OptionExercisePositions(Expire)", err)
	} else if res != nil {
		ok("OptionExercisePositions(Expire)", fmt.Sprintf("rows=%d pageCount=%d", len(res.Items), res.PageCount))
	} else {
		ok("OptionExercisePositions(Expire)", "(nil result)")
	}

	if res, err := tc.OptionExerciseRecords(model.OptionExercisePageRequest{
		Page: 1, Size: 10,
	}); err != nil {
		fail("OptionExerciseRecords", err)
	} else if res != nil {
		ok("OptionExerciseRecords", fmt.Sprintf("rows=%d itemCount=%d", len(res.Items), res.ItemCount))
	} else {
		ok("OptionExerciseRecords", "(nil result)")
	}

	// OptionExerciseCheck 需要持仓，仅在有持仓时执行
	if exercisePositions, err := tc.OptionExercisePositions(model.OptionExercisePositionRequest{Type: "Exercise"}); err == nil &&
		exercisePositions != nil && len(exercisePositions.Items) > 0 {
		p := exercisePositions.Items[0]
		if checkRes, err := tc.OptionExerciseCheck(model.OptionExerciseCheckRequest{
			ContractId:    p.ContractId,
			Type:          "Exercise",
			Quantity:      p.AvailableQuantity,
			ExecutingDate: p.ExpireDate,
		}); err != nil {
			fail("OptionExerciseCheck", err)
		} else if checkRes != nil {
			ok("OptionExerciseCheck", fmt.Sprintf("symbol=%s availableQty=%.0f stkBefore=%.0f stkAfter=%.0f", checkRes.Symbol, checkRes.AvailableQuantity, checkRes.StkPositionBefore, checkRes.StkPositionAfter))
		}
	} else {
		skip("OptionExerciseCheck", "no exercisable positions")
	}

	printSummary()
}

func printSummary() {
	fmt.Println("\n================ SUMMARY ================")
	var pass, fa, sk int
	for _, r := range results {
		if r.ok {
			pass++
		} else if r.err != nil && strings.HasPrefix(r.err.Error(), "skipped") {
			sk++
		} else {
			fa++
		}
	}
	fmt.Printf("PASS=%d  FAIL=%d  SKIP=%d  TOTAL=%d\n", pass, fa, sk, len(results))
	if fa > 0 {
		fmt.Println("\nFailures:")
		for _, r := range results {
			if !r.ok && (r.err == nil || !strings.HasPrefix(r.err.Error(), "skipped")) {
				fmt.Printf("  - %s: %v\n", r.name, r.err)
			}
		}
	}
	fmt.Println("=========================================")
}
