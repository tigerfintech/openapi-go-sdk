// Quote example - 覆盖 QuoteClient 全部公开接口
//
// 配置从 ./tiger_openapi_config.properties 或 ~/.tigeropen/tiger_openapi_config.properties 自动读取。
// 单个接口失败不中断后续,最后统计 PASS/FAIL。
//
// Run: go run ./examples/quote
package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/tigerfintech/openapi-go-sdk/client"
	"github.com/tigerfintech/openapi-go-sdk/config"
	"github.com/tigerfintech/openapi-go-sdk/model"
	"github.com/tigerfintech/openapi-go-sdk/quote"
)

type result struct {
	name string
	ok   bool
	err  error
	note string
}

var results []result

func ok(name, note string) {
	fmt.Printf("[ OK ] %-32s %s\n", name, truncate(note, 140))
	results = append(results, result{name: name, ok: true, note: note})
}

func fail(name string, err error) {
	fmt.Printf("[FAIL] %-32s %v\n", name, err)
	results = append(results, result{name: name, err: err})
}

func skip(name, reason string) {
	fmt.Printf("[SKIP] %-32s %s\n", name, reason)
	results = append(results, result{name: name, err: fmt.Errorf("skipped: %s", reason)})
}

func main() {
	var cfgOpts []config.Option
	if p := os.Getenv("TIGER_CONFIG_PATH"); p != "" {
		cfgOpts = append(cfgOpts, config.WithPropertiesFile(p))
	}
	cfg, err := config.NewClientConfig(cfgOpts...)
	if err != nil {
		log.Fatal("load config failed: ", err)
	}
	fmt.Printf("tiger_id=%s account=%s\n\n", cfg.TigerID, cfg.Account)

	qc := quote.NewQuoteClient(client.NewQuoteHttpClient(cfg))

	fmt.Println("=== Basic market data ===")
	if states, err := qc.GetMarketState("US"); err != nil {
		fail("GetMarketState(US)", err)
	} else if len(states) > 0 {
		ok("GetMarketState(US)", fmt.Sprintf("%s status=%s openTime=%s", states[0].Market, states[0].MarketStatus, states[0].OpenTime))
	} else {
		ok("GetMarketState(US)", "(empty)")
	}

	if briefs, err := qc.GetRealTimeQuote(model.BriefRequest{Symbols: []string{"AAPL", "TSLA"}}); err != nil {
		fail("GetRealTimeQuote", err)
	} else {
		summary := make([]string, 0, len(briefs))
		for _, b := range briefs {
			summary = append(summary, fmt.Sprintf("%s=%.2f", b.Symbol, b.LatestPrice))
		}
		ok("GetRealTimeQuote", strings.Join(summary, " "))
	}

	if klines, err := qc.GetKline(model.KlineRequest{Symbols: []string{"AAPL"}, Period: "day"}); err != nil {
		fail("GetKline(AAPL day)", err)
	} else if len(klines) > 0 {
		ok("GetKline(AAPL day)", fmt.Sprintf("symbol=%s bars=%d", klines[0].Symbol, len(klines[0].Items)))
	} else {
		ok("GetKline(AAPL day)", "(empty)")
	}

	if tl, err := qc.GetTimeline([]string{"AAPL"}); err != nil {
		fail("GetTimeline", err)
	} else if len(tl) > 0 {
		n := 0
		if tl[0].Intraday != nil {
			n = len(tl[0].Intraday.Items)
		}
		ok("GetTimeline", fmt.Sprintf("intraday_points=%d preClose=%.2f", n, tl[0].PreClose))
	} else {
		ok("GetTimeline", "(empty)")
	}

	if tt, err := qc.GetTradeTick(model.TradeTickRequest{Symbols: []string{"AAPL"}}); err != nil {
		fail("GetTradeTick", err)
	} else if len(tt) > 0 {
		ok("GetTradeTick", fmt.Sprintf("ticks=%d", len(tt[0].Items)))
	} else {
		ok("GetTradeTick", "(empty)")
	}

	if d, err := qc.GetQuoteDepth(model.DepthQuoteRequest{Symbols: []string{"AAPL"}, Market: "US"}); err != nil {
		fail("GetQuoteDepth(AAPL)", err)
	} else if len(d) > 0 {
		ok("GetQuoteDepth(AAPL)", fmt.Sprintf("asks=%d bids=%d", len(d[0].Asks), len(d[0].Bids)))
	} else {
		ok("GetQuoteDepth(AAPL)", "(empty)")
	}

	fmt.Println("\n=== Options ===")
	var expiryDate, optIdentifier string
	if exps, err := qc.GetOptionExpiration([]string{"AAPL"}); err != nil {
		fail("GetOptionExpiration(AAPL)", err)
	} else if len(exps) > 0 && len(exps[0].Dates) > 0 {
		ok("GetOptionExpiration(AAPL)", fmt.Sprintf("dates=%d first=%s", len(exps[0].Dates), exps[0].Dates[0]))
		// 挑一个远期过期日
		dates := exps[0].Dates
		expiryDate = dates[len(dates)/2]
	} else {
		ok("GetOptionExpiration(AAPL)", "(empty)")
	}

	if expiryDate == "" {
		skip("GetOptionChain", "no expiry available")
		skip("GetOptionQuote", "no expiry available")
		skip("GetOptionKline", "no expiry available")
	} else {
		if chain, err := qc.GetOptionChain([][2]string{{"AAPL", expiryDate}}); err != nil {
			fail("GetOptionChain(AAPL)", err)
		} else if len(chain) > 0 && len(chain[0].Items) > 0 {
			ok(fmt.Sprintf("GetOptionChain(%s)", expiryDate), fmt.Sprintf("rows=%d", len(chain[0].Items)))
			mid := chain[0].Items[len(chain[0].Items)/2]
			if mid.Call != nil {
				optIdentifier = mid.Call.Identifier
			} else if mid.Put != nil {
				optIdentifier = mid.Put.Identifier
			}

			// Verify option_filter and return_greek_value — reuse expiry from chain response
			trueVal := true
			if filtered, err := qc.GetOptionChainByReq(model.OptionChainRequest{
				OptionBasic:      []map[string]interface{}{{"symbol": chain[0].Symbol, "expiry": chain[0].Expiry}},
				ReturnGreekValue: &trueVal,
				OptionFilter: &model.OptionChainFilter{
					InTheMoney:        func() *bool { f := false; return &f }(),
					ImpliedVolatility: model.NewRangeFloat64(0.0, 5.0),
					Greeks: &model.OptionChainFilterGreeks{
						Delta: model.NewRangeFloat64(0.0, 0.6),
					},
				},
			}); err != nil {
				fail("GetOptionChain(filter+greek)", err)
			} else {
				rows := 0
				hasGreek := false
				for _, c := range filtered {
					rows += len(c.Items)
					for _, row := range c.Items {
						if row.Call != nil && row.Call.Delta != 0 {
							hasGreek = true
						}
					}
				}
				ok("GetOptionChain(filter+greek)", fmt.Sprintf("rows=%d has_greek=%v", rows, hasGreek))
			}
		} else {
			ok(fmt.Sprintf("GetOptionChain(%s)", expiryDate), "(empty items)")
		}


		if optIdentifier == "" {
			skip("GetOptionQuote", "no identifier from chain")
			skip("GetOptionKline", "no identifier from chain")
		} else {
			if briefs, err := qc.GetOptionQuote([]string{optIdentifier}); err != nil {
				fail("GetOptionQuote", err)
			} else if len(briefs) > 0 {
				ok("GetOptionQuote", fmt.Sprintf("%s latestPrice=%.4f", briefs[0].Symbol, briefs[0].LatestPrice))
			} else {
				ok("GetOptionQuote", "(empty)")
			}
			nowMs := time.Now().UnixMilli()
			beginMs := nowMs - 30*24*60*60*1000 // 30 days ago
			if ks, err := qc.GetOptionKline([]string{optIdentifier}, "day", beginMs, nowMs); err != nil {
				fail("GetOptionKline", err)
			} else if len(ks) > 0 {
				ok("GetOptionKline", fmt.Sprintf("bars=%d", len(ks[0].Items)))
			} else {
				ok("GetOptionKline", "(empty)")
			}
		}
	}

	// HK options smoke test
	fmt.Println("\n=== Options (HK) ===")
	var hkExpiryDate, hkOptIdentifier string
	if exps, err := qc.GetOptionExpiration([]string{"00700.HK"}, "HK"); err != nil {
		fail("GetOptionExpiration(00700.HK)", err)
	} else if len(exps) > 0 && len(exps[0].Dates) > 0 {
		ok("GetOptionExpiration(00700.HK)", fmt.Sprintf("dates=%d first=%s", len(exps[0].Dates), exps[0].Dates[0]))
		hkExpiryDate = exps[0].Dates[0]
	} else {
		ok("GetOptionExpiration(00700.HK)", "(empty)")
	}

	if hkExpiryDate == "" {
		skip("GetOptionChain(HK)", "no expiry available")
		skip("GetOptionQuote(HK)", "no expiry available")
		skip("GetOptionKline(HK)", "no expiry available")
	} else {
		if chain, err := qc.GetOptionChain([][2]string{{"00700.HK", hkExpiryDate}}, "Asia/Hong_Kong"); err != nil {
			fail("GetOptionChain(00700.HK)", err)
		} else if len(chain) > 0 && len(chain[0].Items) > 0 {
			ok(fmt.Sprintf("GetOptionChain(HK %s)", hkExpiryDate), fmt.Sprintf("rows=%d", len(chain[0].Items)))
			mid := chain[0].Items[len(chain[0].Items)/2]
			if mid.Call != nil {
				hkOptIdentifier = mid.Call.Identifier
			} else if mid.Put != nil {
				hkOptIdentifier = mid.Put.Identifier
			}
		} else {
			ok(fmt.Sprintf("GetOptionChain(HK %s)", hkExpiryDate), "(empty items)")
		}

		if hkOptIdentifier == "" {
			skip("GetOptionQuote(HK)", "no identifier from chain")
			skip("GetOptionKline(HK)", "no identifier from chain")
		} else {
			if briefs, err := qc.GetOptionQuote([]string{hkOptIdentifier}, "Asia/Hong_Kong"); err != nil {
				fail("GetOptionQuote(HK)", err)
			} else if len(briefs) > 0 {
				ok("GetOptionQuote(HK)", fmt.Sprintf("%s latestPrice=%.4f", briefs[0].Symbol, briefs[0].LatestPrice))
			} else {
				ok("GetOptionQuote(HK)", "(empty)")
			}
			nowMs := time.Now().UnixMilli()
			beginMs := nowMs - 30*24*60*60*1000
			if ks, err := qc.GetOptionKline([]string{hkOptIdentifier}, "day", beginMs, nowMs, "Asia/Hong_Kong"); err != nil {
				fail("GetOptionKline(HK)", err)
			} else if len(ks) > 0 {
				ok("GetOptionKline(HK)", fmt.Sprintf("bars=%d", len(ks[0].Items)))
			} else {
				ok("GetOptionKline(HK)", "(empty)")
			}
		}
	}

	fmt.Println("\n=== Futures ===")
	var exchangeCode, contractCode string
	if exs, err := qc.GetFutureExchange(); err != nil {
		fail("GetFutureExchange", err)
	} else if len(exs) > 0 {
		ok("GetFutureExchange", fmt.Sprintf("exchanges=%d first=%s", len(exs), exs[0].Code))
		exchangeCode = exs[0].Code
	} else {
		ok("GetFutureExchange", "(empty)")
	}

	if exchangeCode == "" {
		skip("GetFutureContracts", "no exchange")
	} else {
		if cs, err := qc.GetFutureContracts(exchangeCode); err != nil {
			fail(fmt.Sprintf("GetFutureContracts(%s)", exchangeCode), err)
		} else if len(cs) > 0 {
			ok(fmt.Sprintf("GetFutureContracts(%s)", exchangeCode), fmt.Sprintf("contracts=%d first=%s", len(cs), cs[0].ContractCode))
			contractCode = cs[0].ContractCode
		} else {
			ok(fmt.Sprintf("GetFutureContracts(%s)", exchangeCode), "(empty)")
		}
	}

	if contractCode == "" {
		skip("GetFutureRealTimeQuote", "no contract")
		skip("GetFutureKline", "no contract")
	} else {
		if q, err := qc.GetFutureRealTimeQuote(model.FutureBriefRequest{ContractCodes: []string{contractCode}}); err != nil {
			fail("GetFutureRealTimeQuote", err)
		} else if len(q) > 0 {
			ok("GetFutureRealTimeQuote", fmt.Sprintf("%s latestPrice=%.4f", q[0].ContractCode, q[0].LatestPrice))
		} else {
			ok("GetFutureRealTimeQuote", "(empty)")
		}
		if ks, err := qc.GetFutureKline(model.FutureKlineRequest{
			ContractCodes: []string{contractCode}, Period: "day", BeginTime: -1, EndTime: -1,
		}); err != nil {
			fail(fmt.Sprintf("GetFutureKline(%s)", contractCode), err)
		} else if len(ks) > 0 {
			ok(fmt.Sprintf("GetFutureKline(%s)", contractCode), fmt.Sprintf("bars=%d", len(ks[0].Items)))
		} else {
			ok(fmt.Sprintf("GetFutureKline(%s)", contractCode), "(empty)")
		}
	}

	fmt.Println("\n=== Fundamentals & capital flow ===")
	if items, err := qc.GetFinancialDaily(model.FinancialDailyRequest{
		Symbols: []string{"AAPL"}, Market: "US",
		Fields:    []string{"shares_outstanding"},
		BeginDate: "2026-05-05", EndDate: "2026-05-06",
	}); err != nil {
		fail("GetFinancialDaily(AAPL)", err)
	} else {
		ok("GetFinancialDaily(AAPL)", fmt.Sprintf("rows=%d", len(items)))
	}

	if items, err := qc.GetFinancialReport(model.FinancialReportRequest{
		Symbols: []string{"AAPL"}, Market: "US",
		Fields: []string{"total_revenue"}, PeriodType: "Annual",
	}); err != nil {
		fail("GetFinancialReport(AAPL)", err)
	} else if len(items) > 0 {
		ok("GetFinancialReport(AAPL)", fmt.Sprintf("%s %s=%s @%s", items[0].Symbol, items[0].Field, items[0].Value, items[0].FilingDate))
	} else {
		ok("GetFinancialReport(AAPL)", "(empty)")
	}

	if items, err := qc.GetCorporateAction(model.CorporateActionRequest{
		Symbols: []string{"AAPL"}, Market: "US", ActionType: "DIVIDEND",
		BeginDate: "2024-01-01", EndDate: "2024-12-31",
	}); err != nil {
		fail("GetCorporateAction(AAPL)", err)
	} else {
		ok("GetCorporateAction(AAPL)", fmt.Sprintf("rows=%d", len(items)))
	}

	if cf, err := qc.GetCapitalFlow("AAPL", "US", "day"); err != nil {
		fail("GetCapitalFlow(AAPL)", err)
	} else {
		ok("GetCapitalFlow(AAPL)", fmt.Sprintf("%s period=%s rows=%d", cf.Symbol, cf.Period, len(cf.Items)))
	}

	if cd, err := qc.GetCapitalDistribution("AAPL", "US"); err != nil {
		fail("GetCapitalDistribution(AAPL)", err)
	} else {
		ok("GetCapitalDistribution(AAPL)", fmt.Sprintf("%s netInflow=%.2f", cd.Symbol, cd.NetInflow))
	}

	fmt.Println("\n=== Scanner & permission ===")
	if res, err := qc.MarketScanner(model.MarketScannerRequest{
		Market: "US", Page: 0, PageSize: 10,
	}); err != nil {
		fail("MarketScanner", err)
	} else {
		ok("MarketScanner", fmt.Sprintf("page=%d/%d totalCount=%d items=%d", res.Page, res.TotalPage, res.TotalCount, len(res.Items)))
	}

	if tags, err := qc.GetMarketScannerTags(model.MarketScannerTagsRequest{
		Market:          "US",
		MultiTagsFields: []string{"MultiTagField_Industry"},
	}); err != nil {
		fail("GetMarketScannerTags(US)", err)
	} else {
		ok("GetMarketScannerTags(US)", fmt.Sprintf("groups=%d", len(tags)))
	}

	if analyses, err := qc.GetOptionAnalysis(model.OptionAnalysisRequest{
		Symbols: []model.OptionAnalysisSymbol{{Symbol: "AAPL", Period: "26week"}},
		Market:  "US",
	}); err != nil {
		fail("GetOptionAnalysis(AAPL 26week)", err)
	} else if len(analyses) > 0 {
		a := analyses[0]
		ok("GetOptionAnalysis(AAPL 26week)", fmt.Sprintf("symbol=%s impliedVol30Days=%.4f hisVol=%.4f ivHisVRatio=%.4f callPutRatio=%.4f points=%d",
			a.Symbol, a.ImpliedVol30Days, a.HisVolatility, a.IvHisVRatio, a.CallPutRatio, len(a.VolatilityList)))
	} else {
		skip("GetOptionAnalysis(AAPL 26week)", "empty response")
	}

	if perms, err := qc.GrabQuotePermission(); err != nil {
		fail("GrabQuotePermission", err)
	} else {
		ok("GrabQuotePermission", fmt.Sprintf("permissions=%d", len(perms)))
	}

	if ent, err := qc.GetAddonEntitlement(); err != nil {
		fail("GetAddonEntitlement", err)
	} else {
		ok("GetAddonEntitlement", fmt.Sprintf("userLevel=%s addons=%d", ent.UserLevel, len(ent.Addons)))
	}

	// ===== v0.3.0 新增 API =====
	fmt.Println("\n=== v0.3.0: 股票基础扩展 ===")

	if items, err := qc.GetSymbols(model.SymbolsRequest{Market: "US", SecType: "STK"}); err != nil {
		fail("GetSymbols(US)", err)
	} else {
		ok("GetSymbols(US)", fmt.Sprintf("count=%d first=%s", len(items), firstOrEmpty(items)))
	}

	if items, err := qc.GetSymbolNames(model.SymbolsRequest{Market: "US"}); err != nil {
		fail("GetSymbolNames(US)", err)
	} else {
		ok("GetSymbolNames(US)", fmt.Sprintf("count=%d", len(items)))
	}

	if items, err := qc.GetTradeMetas(model.TradeMetasRequest{Symbols: []string{"AAPL"}}); err != nil {
		fail("GetTradeMetas(AAPL)", err)
	} else {
		ok("GetTradeMetas(AAPL)", fmt.Sprintf("count=%d", len(items)))
	}

	if items, err := qc.GetStockDetails(model.StockDetailsRequest{Symbols: []string{"AAPL"}}); err != nil {
		fail("GetStockDetails(AAPL)", err)
	} else {
		ok("GetStockDetails(AAPL)", fmt.Sprintf("count=%d", len(items)))
	}

	if items, err := qc.GetDelayedQuote(model.StockDelayBriefsRequest{Symbols: []string{"AAPL"}}); err != nil {
		fail("GetDelayedQuote(AAPL)", err)
	} else {
		ok("GetDelayedQuote(AAPL)", fmt.Sprintf("count=%d", len(items)))
	}

	if items, err := qc.GetKline(model.KlineRequest{
		Symbols: []string{"AAPL"}, Period: "day", Limit: 10,
	}); err != nil {
		fail("GetKline(AAPL day)", err)
	} else if len(items) > 0 {
		ok("GetKline(AAPL day)", fmt.Sprintf("bars=%d", len(items[0].Items)))
	} else {
		ok("GetBars(AAPL day)", "(empty)")
	}

	if items, err := qc.GetTimelineHistory(model.TimelineHistoryRequest{
		Symbols: []string{"AAPL"}, Date: "2025-05-07",
	}); err != nil {
		fail("GetTimelineHistory(AAPL)", err)
	} else {
		ok("GetTimelineHistory(AAPL)", fmt.Sprintf("count=%d", len(items)))
	}

	if items, err := qc.GetTradeRank(model.TradeRankRequest{Market: "US"}); err != nil {
		fail("GetTradeRank(US)", err)
	} else {
		ok("GetTradeRank(US)", fmt.Sprintf("rows=%d", len(items)))
	}

	if items, err := qc.GetShortInterest(model.ShortInterestRequest{Symbols: []string{"AAPL"}}); err != nil {
		if strings.Contains(err.Error(), "not support") || strings.Contains(err.Error(), "permission") {
			skip("GetShortInterest(AAPL)", err.Error())
		} else {
			fail("GetShortInterest(AAPL)", err)
		}
	} else {
		ok("GetShortInterest(AAPL)", fmt.Sprintf("rows=%d", len(items)))
	}

	if sb, err := qc.GetStockBroker(model.StockBrokerRequest{Symbol: "00700", Limit: 3}); err != nil {
		fail("GetStockBroker(00700)", err)
	} else if sb != nil {
		ok("GetStockBroker(00700)", fmt.Sprintf("askLevels=%d bidLevels=%d", len(sb.LevelAskList), len(sb.LevelBidList)))
	}

	if data, err := qc.GetStockFundamental(model.StockFundamentalRequest{Symbols: []string{"AAPL"}, Market: "US"}); err != nil {
		fail("GetStockFundamental(AAPL)", err)
	} else {
		ok("GetStockFundamental(AAPL)", fmt.Sprintf("keys=%d", len(data)))
	}

	if si, err := qc.GetStockIndustry(model.StockIndustryRequest{Symbol: "AAPL", Market: "US"}); err != nil {
		fail("GetStockIndustry(AAPL)", err)
	} else if len(si) > 0 {
		ok("GetStockIndustry(AAPL)", fmt.Sprintf("levels=%d first=%s", len(si), si[0].Level))
	} else {
		ok("GetStockIndustry(AAPL)", "(empty)")
	}

	if perms, err := qc.GetQuotePermission(model.QuotePermissionRequest{}); err != nil {
		fail("GetQuotePermission", err)
	} else {
		ok("GetQuotePermission", fmt.Sprintf("count=%d", len(perms)))
	}

	if items, err := qc.GetKlineQuota(model.KlineQuotaRequest{WithDetails: true}); err != nil {
		fail("GetKlineQuota", err)
	} else {
		ok("GetKlineQuota", fmt.Sprintf("methods=%d", len(items)))
	}

	fmt.Println("\n=== v0.3.0: 期货扩展 ===")

	// 尝试用已有 contractCode 测期货扩展接口
	futureContract := "MEUR2609"

	if items, err := qc.GetFutureContract(model.FutureContractSingleRequest{ContractCode: futureContract}); err != nil {
		fail("GetFutureContract(MEUR2609)", err)
	} else {
		ok("GetFutureContract(MEUR2609)", fmt.Sprintf("count=%d", len(items)))
	}

	if items, err := qc.GetAllFutureContracts(model.AllFutureContractsRequest{Type: "MEUR"}); err != nil {
		fail("GetAllFutureContracts(MEUR)", err)
	} else {
		ok("GetAllFutureContracts(MEUR)", fmt.Sprintf("count=%d", len(items)))
	}

	if c, err := qc.GetCurrentFutureContract(model.FutureContractSingleRequest{Type: "MEUR"}); err != nil {
		fail("GetCurrentFutureContract(MEUR)", err)
	} else if c != nil {
		ok("GetCurrentFutureContract(MEUR)", fmt.Sprintf("code=%s", c.ContractCode))
	}

	if items, err := qc.GetFutureContinuousContracts(model.FutureContinuousContractsRequest{Type: "MEUR"}); err != nil {
		fail("GetFutureContinuousContracts(MEUR)", err)
	} else {
		ok("GetFutureContinuousContracts(MEUR)", fmt.Sprintf("count=%d", len(items)))
	}

	if items, err := qc.GetFutureKline(model.FutureKlineRequest{
		ContractCodes: []string{futureContract}, Period: "day", Limit: 10,
	}); err != nil {
		fail("GetFutureKline", err)
	} else if len(items) > 0 {
		ok("GetFutureKline", fmt.Sprintf("bars=%d", len(items[0].Items)))
	} else {
		ok("GetFutureBars", "(empty)")
	}

	if items, err := qc.GetFutureTradeTicks(model.FutureTradeTicksRequest{ContractCode: futureContract, Limit: 10}); err != nil {
		fail("GetFutureTradeTicks", err)
	} else {
		ok("GetFutureTradeTicks", fmt.Sprintf("count=%d", len(items)))
	}

	if items, err := qc.GetFutureDepth(model.FutureDepthRequest{ContractCodes: []string{futureContract}}); err != nil {
		fail("GetFutureDepth", err)
	} else {
		ok("GetFutureDepth", fmt.Sprintf("count=%d", len(items)))
	}

	if tt, err := qc.GetFutureTradingTimes(model.FutureTradingTimesRequest{ContractCode: futureContract}); err != nil {
		fail("GetFutureTradingTimes", err)
	} else if tt != nil {
		ok("GetFutureTradingTimes", fmt.Sprintf("zone=%s segments=%d", tt.Zone, len(tt.TradingTimes)))
	}

	fmt.Println("\n=== v0.3.0: 基金/窝轮/行业/日历 ===")

	if items, err := qc.GetFundSymbols(model.FundSymbolsRequest{}); err != nil {
		fail("GetFundSymbols", err)
	} else {
		ok("GetFundSymbols", fmt.Sprintf("count=%d first=%s", len(items), firstOrEmpty(items)))
	}

	if items, err := qc.GetIndustryList(model.IndustryListRequest{IndustryLevel: "GSECTOR"}); err != nil {
		fail("GetIndustryList(GSECTOR)", err)
	} else {
		ok("GetIndustryList(GSECTOR)", fmt.Sprintf("count=%d", len(items)))
	}

	if items, err := qc.GetTradingCalendar(model.TradingCalendarRequest{
		Market: "US", BeginDate: "2025-05-01", EndDate: "2025-05-31",
	}); err != nil {
		fail("GetTradingCalendar(US)", err)
	} else {
		ok("GetTradingCalendar(US)", fmt.Sprintf("days=%d", len(items)))
	}

	if items, err := qc.GetCorporateSplit(model.CorporateActionRequest{
		Symbols: []string{"AAPL"}, Market: "US", BeginDate: "2020-01-01", EndDate: "2024-12-31",
	}); err != nil {
		fail("GetCorporateSplit(AAPL)", err)
	} else {
		ok("GetCorporateSplit(AAPL)", fmt.Sprintf("rows=%d", len(items)))
	}

	if items, err := qc.GetCorporateDividend(model.CorporateActionRequest{
		Symbols: []string{"AAPL"}, Market: "US", BeginDate: "2024-01-01", EndDate: "2024-12-31",
	}); err != nil {
		fail("GetCorporateDividend(AAPL)", err)
	} else {
		ok("GetCorporateDividend(AAPL)", fmt.Sprintf("rows=%d", len(items)))
	}
	if items, err := qc.GetCorporateSymbolChange(model.CorporateActionRequest{
		Symbols: []string{"META"}, Market: "US", BeginDate: "2022-01-01", EndDate: "2023-01-01",
	}); err != nil {
		fail("GetCorporateSymbolChange(META)", err)
	} else {
		ok("GetCorporateSymbolChange(META)", fmt.Sprintf("rows=%d", len(items)))
	}
	if items, err := qc.GetCorporateDelisting(model.CorporateActionRequest{
		Symbols: []string{"TWTR"}, Market: "US", BeginDate: "2022-01-01", EndDate: "2023-01-01",
	}); err != nil {
		fail("GetCorporateDelisting(TWTR)", err)
	} else {
		ok("GetCorporateDelisting(TWTR)", fmt.Sprintf("rows=%d", len(items)))
	}
	if items, err := qc.GetCorporateIPO(model.CorporateActionRequest{
		Symbols: []string{"RIVN"}, Market: "US", BeginDate: "2021-01-01", EndDate: "2022-01-01",
	}); err != nil {
		fail("GetCorporateIPO(RIVN)", err)
	} else {
		ok("GetCorporateIPO(RIVN)", fmt.Sprintf("rows=%d", len(items)))
	}

	if items, err := qc.GetFinancialCurrency(model.FinancialCurrencyRequest{Symbols: []string{"AAPL"}, Market: "US"}); err != nil {
		fail("GetFinancialCurrency(AAPL)", err)
	} else {
		ok("GetFinancialCurrency(AAPL)", fmt.Sprintf("rows=%d", len(items)))
	}

	if items, err := qc.GetFinancialExchangeRate(model.FinancialExchangeRateRequest{
		CurrencyList: []string{"USD"}, BeginDate: "20250501", EndDate: "20250507",
	}); err != nil {
		fail("GetFinancialExchangeRate(USD)", err)
	} else {
		ok("GetFinancialExchangeRate(USD)", fmt.Sprintf("rows=%d", len(items)))
	}

	if items, err := qc.GetQuoteOvernight(model.QuoteOvernightRequest{Symbols: []string{"AAPL"}}); err != nil {
		fail("GetQuoteOvernight(AAPL)", err)
	} else {
		ok("GetQuoteOvernight(AAPL)", fmt.Sprintf("rows=%d", len(items)))
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

func truncate(s string, max int) string {
	if len(s) > max {
		return s[:max] + "..."
	}
	return s
}

func firstOrEmpty(xs []string) string {
	if len(xs) == 0 {
		return ""
	}
	return xs[0]
}
