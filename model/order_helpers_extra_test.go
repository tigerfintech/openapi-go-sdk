package model

import (
	"testing"
)

// 测试 MarketOrderByAmount 构造函数（按金额市价单）
func TestMarketOrderByAmount(t *testing.T) {
	o := MarketOrderByAmount("U123456", "AAPL", "STK", "BUY", 10000.0)
	if o.Account != "U123456" {
		t.Errorf("Account = %q, want U123456", o.Account)
	}
	if o.Symbol != "AAPL" {
		t.Errorf("Symbol = %q, want AAPL", o.Symbol)
	}
	if o.SecType != "STK" {
		t.Errorf("SecType = %q, want STK", o.SecType)
	}
	if o.Action != "BUY" {
		t.Errorf("Action = %q, want BUY", o.Action)
	}
	if o.OrderType != string(OrderTypeMKT) {
		t.Errorf("OrderType = %q, want MKT", o.OrderType)
	}
	if o.TotalQuantity != 0 {
		t.Errorf("TotalQuantity = %d, want 0 (按金额下单不设数量)", o.TotalQuantity)
	}
	if o.Amount != 10000.0 {
		t.Errorf("Amount = %f, want 10000.0", o.Amount)
	}
	if o.TimeInForce != string(TimeInForceDAY) {
		t.Errorf("TimeInForce = %q, want DAY", o.TimeInForce)
	}
}

// 测试 MarketOrderByAmount 零金额边界
func TestMarketOrderByAmount_zeroAmount(t *testing.T) {
	o := MarketOrderByAmount("U1", "AAPL", "STK", "BUY", 0)
	if o.Amount != 0 {
		t.Errorf("Amount = %f, want 0", o.Amount)
	}
	if o.OrderType != string(OrderTypeMKT) {
		t.Errorf("OrderType = %q, want MKT", o.OrderType)
	}
}

// 测试 LimitOrderByAmount 构造函数（按金额限价单）
func TestLimitOrderByAmount(t *testing.T) {
	o := LimitOrderByAmount("U123456", "AAPL", "STK", "BUY", 10000.0, 150.50)
	if o.OrderType != string(OrderTypeLMT) {
		t.Errorf("OrderType = %q, want LMT", o.OrderType)
	}
	if o.TotalQuantity != 0 {
		t.Errorf("TotalQuantity = %d, want 0 (按金额下单不设数量)", o.TotalQuantity)
	}
	if o.Amount != 10000.0 {
		t.Errorf("Amount = %f, want 10000.0", o.Amount)
	}
	if o.LimitPrice != 150.50 {
		t.Errorf("LimitPrice = %f, want 150.50", o.LimitPrice)
	}
	if o.TimeInForce != string(TimeInForceDAY) {
		t.Errorf("TimeInForce = %q, want DAY", o.TimeInForce)
	}
}

// 测试 TrailOrderByPrice 构造函数（按固定价差跟踪止损单）
func TestTrailOrderByPrice(t *testing.T) {
	o := TrailOrderByPrice("U123456", "AAPL", "STK", "SELL", 100, 2.50)
	if o.OrderType != string(OrderTypeTRAIL) {
		t.Errorf("OrderType = %q, want TRAIL", o.OrderType)
	}
	if o.TotalQuantity != 100 {
		t.Errorf("TotalQuantity = %d, want 100", o.TotalQuantity)
	}
	if o.AuxPrice != 2.50 {
		t.Errorf("AuxPrice = %f, want 2.50", o.AuxPrice)
	}
	if o.TrailingPercent != 0 {
		t.Errorf("TrailingPercent = %f, want 0 (使用 AuxPrice 而非百分比)", o.TrailingPercent)
	}
	if o.TimeInForce != string(TimeInForceDAY) {
		t.Errorf("TimeInForce = %q, want DAY", o.TimeInForce)
	}
}

// 测试 LimitOrderWithLegs 正常路径（1 条腿）
func TestLimitOrderWithLegs_oneLeg_succeeds(t *testing.T) {
	legs := []OrderLegRequest{
		{LegType: "PROFIT", Price: 160.0, TimeInForce: "GTC"},
	}
	o, err := LimitOrderWithLegs("U123456", "AAPL", "STK", "BUY", 100, 150.0, legs)
	if err != nil {
		t.Fatalf("不应返回错误: %v", err)
	}
	if o.OrderType != string(OrderTypeLMT) {
		t.Errorf("OrderType = %q, want LMT", o.OrderType)
	}
	if o.LimitPrice != 150.0 {
		t.Errorf("LimitPrice = %f, want 150.0", o.LimitPrice)
	}
	if len(o.OrderLegs) != 1 {
		t.Fatalf("OrderLegs len = %d, want 1", len(o.OrderLegs))
	}
	if o.OrderLegs[0].LegType != "PROFIT" {
		t.Errorf("OrderLegs[0].LegType = %q, want PROFIT", o.OrderLegs[0].LegType)
	}
	if o.TimeInForce != string(TimeInForceDAY) {
		t.Errorf("TimeInForce = %q, want DAY", o.TimeInForce)
	}
}

// 测试 LimitOrderWithLegs 正常路径（2 条腿）
func TestLimitOrderWithLegs_twoLegs_succeeds(t *testing.T) {
	legs := []OrderLegRequest{
		{LegType: "PROFIT", Price: 160.0, TimeInForce: "GTC"},
		{LegType: "LOSS", Price: 140.0, TimeInForce: "GTC"},
	}
	o, err := LimitOrderWithLegs("U123456", "AAPL", "STK", "BUY", 100, 150.0, legs)
	if err != nil {
		t.Fatalf("不应返回错误: %v", err)
	}
	if len(o.OrderLegs) != 2 {
		t.Errorf("OrderLegs len = %d, want 2", len(o.OrderLegs))
	}
	if o.OrderLegs[1].LegType != "LOSS" {
		t.Errorf("OrderLegs[1].LegType = %q, want LOSS", o.OrderLegs[1].LegType)
	}
}

// 测试 LimitOrderWithLegs 空腿返回错误
func TestLimitOrderWithLegs_emptyLegs_returnsError(t *testing.T) {
	o, err := LimitOrderWithLegs("U1", "AAPL", "STK", "BUY", 100, 150.0, nil)
	if err == nil {
		t.Fatal("空腿应返回错误，got nil")
	}
	if o.OrderType != "" {
		t.Errorf("错误时返回的 OrderRequest 应为零值，OrderType = %q", o.OrderType)
	}
}

// 测试 LimitOrderWithLegs 超过 2 条腿返回错误
func TestLimitOrderWithLegs_tooManyLegs_returnsError(t *testing.T) {
	legs := []OrderLegRequest{
		{LegType: "PROFIT", Price: 160.0, TimeInForce: "GTC"},
		{LegType: "LOSS", Price: 140.0, TimeInForce: "GTC"},
		{LegType: "EXTRA", Price: 130.0, TimeInForce: "GTC"},
	}
	o, err := LimitOrderWithLegs("U1", "AAPL", "STK", "BUY", 100, 150.0, legs)
	if err == nil {
		t.Fatal("超过 2 条腿应返回错误，got nil")
	}
	if o.OrderType != "" {
		t.Errorf("错误时返回的 OrderRequest 应为零值，OrderType = %q", o.OrderType)
	}
}

// 测试 ComboOrder 构造多腿组合单（带限价/止损/跟踪百分比）
func TestComboOrder(t *testing.T) {
	legs := []ContractLegRequest{
		NewContractLeg("AAPL", "STK", "BUY", 1, "", "", ""),
		NewContractLeg("AAPL", "STK", "SELL", 1, "", "", ""),
	}
	o := ComboOrder("U123456", "BUY", string(OrderTypeLMT), 2, legs, "MLEG", 150.0, 0, 0)
	if o.Account != "U123456" {
		t.Errorf("Account = %q, want U123456", o.Account)
	}
	if o.SecType != "MLEG" {
		t.Errorf("SecType = %q, want MLEG", o.SecType)
	}
	if o.Action != "BUY" {
		t.Errorf("Action = %q, want BUY", o.Action)
	}
	if o.OrderType != string(OrderTypeLMT) {
		t.Errorf("OrderType = %q, want LMT", o.OrderType)
	}
	if o.TotalQuantity != 2 {
		t.Errorf("TotalQuantity = %d, want 2", o.TotalQuantity)
	}
	if len(o.ContractLegs) != 2 {
		t.Errorf("ContractLegs len = %d, want 2", len(o.ContractLegs))
	}
	if o.ComboType != "MLEG" {
		t.Errorf("ComboType = %q, want MLEG", o.ComboType)
	}
	if o.LimitPrice != 150.0 {
		t.Errorf("LimitPrice = %f, want 150.0", o.LimitPrice)
	}
	if o.AuxPrice != 0 {
		t.Errorf("AuxPrice = %f, want 0", o.AuxPrice)
	}
	if o.TrailingPercent != 0 {
		t.Errorf("TrailingPercent = %f, want 0", o.TrailingPercent)
	}
	if o.TimeInForce != string(TimeInForceDAY) {
		t.Errorf("TimeInForce = %q, want DAY", o.TimeInForce)
	}
}

// 测试 ComboOrder 只设 auxPrice 不设 limitPrice
func TestComboOrder_auxPriceOnly(t *testing.T) {
	o := ComboOrder("U1", "SELL", string(OrderTypeSTP), 1, nil, "MLEG", 0, 1.5, 0)
	if o.LimitPrice != 0 {
		t.Errorf("LimitPrice = %f, want 0", o.LimitPrice)
	}
	if o.AuxPrice != 1.5 {
		t.Errorf("AuxPrice = %f, want 1.5", o.AuxPrice)
	}
}

// 测试 ComboOrder 只设 trailingPercent
func TestComboOrder_trailingPercentOnly(t *testing.T) {
	o := ComboOrder("U1", "SELL", string(OrderTypeTRAIL), 1, nil, "MLEG", 0, 0, 3.5)
	if o.LimitPrice != 0 {
		t.Errorf("LimitPrice = %f, want 0", o.LimitPrice)
	}
	if o.AuxPrice != 0 {
		t.Errorf("AuxPrice = %f, want 0", o.AuxPrice)
	}
	if o.TrailingPercent != 3.5 {
		t.Errorf("TrailingPercent = %f, want 3.5", o.TrailingPercent)
	}
}

// 测试 ComboOrder 全部价格为零（省略）
func TestComboOrder_allPricesZero(t *testing.T) {
	o := ComboOrder("U1", "BUY", string(OrderTypeMKT), 1, nil, "MLEG", 0, 0, 0)
	if o.LimitPrice != 0 || o.AuxPrice != 0 || o.TrailingPercent != 0 {
		t.Errorf("零值价格应省略: LimitPrice=%f AuxPrice=%f TrailingPercent=%f",
			o.LimitPrice, o.AuxPrice, o.TrailingPercent)
	}
}

// 测试 OcaOrder 构造 OCA 单
func TestOcaOrder(t *testing.T) {
	child1 := MarketOrder("U123456", "AAPL", "STK", "BUY", 100)
	child2 := LimitOrder("U123456", "AAPL", "STK", "BUY", 100, 150.0)
	o := OcaOrder("U123456", "AAPL", "STK", "BUY", 100, []*OrderRequest{&child1, &child2})
	if o.Account != "U123456" {
		t.Errorf("Account = %q, want U123456", o.Account)
	}
	if o.Symbol != "AAPL" {
		t.Errorf("Symbol = %q, want AAPL", o.Symbol)
	}
	if o.SecType != "STK" {
		t.Errorf("SecType = %q, want STK", o.SecType)
	}
	if o.Action != "BUY" {
		t.Errorf("Action = %q, want BUY", o.Action)
	}
	if o.OrderType != string(OrderTypeOCA) {
		t.Errorf("OrderType = %q, want OCA", o.OrderType)
	}
	if o.TotalQuantity != 100 {
		t.Errorf("TotalQuantity = %d, want 100", o.TotalQuantity)
	}
	if len(o.OcaOrders) != 2 {
		t.Fatalf("OcaOrders len = %d, want 2", len(o.OcaOrders))
	}
	if o.OcaOrders[0] != &child1 {
		t.Errorf("OcaOrders[0] 指针不一致")
	}
	if o.TimeInForce != string(TimeInForceDAY) {
		t.Errorf("TimeInForce = %q, want DAY", o.TimeInForce)
	}
}

// 测试 OcaOrder 空 OCA 子单列表
func TestOcaOrder_emptyChildren(t *testing.T) {
	o := OcaOrder("U1", "AAPL", "STK", "BUY", 100, nil)
	if len(o.OcaOrders) != 0 {
		t.Errorf("OcaOrders len = %d, want 0", len(o.OcaOrders))
	}
	if o.OrderType != string(OrderTypeOCA) {
		t.Errorf("OrderType = %q, want OCA", o.OrderType)
	}
}

// 测试 NewContractLeg 构造多腿组合单的子腿
func TestNewContractLeg(t *testing.T) {
	leg := NewContractLeg("AAPL", "OPT", "BUY", 1, "20250620", "150.0", "CALL")
	if leg.Symbol != "AAPL" {
		t.Errorf("Symbol = %q, want AAPL", leg.Symbol)
	}
	if leg.SecType != "OPT" {
		t.Errorf("SecType = %q, want OPT", leg.SecType)
	}
	if leg.Action != "BUY" {
		t.Errorf("Action = %q, want BUY", leg.Action)
	}
	if leg.Ratio == nil {
		t.Fatal("Ratio 不应为 nil")
	}
	if *leg.Ratio != 1 {
		t.Errorf("Ratio = %d, want 1", *leg.Ratio)
	}
	if leg.Expiry != "20250620" {
		t.Errorf("Expiry = %q, want 20250620", leg.Expiry)
	}
	if leg.Strike != "150.0" {
		t.Errorf("Strike = %q, want 150.0", leg.Strike)
	}
	if leg.Right != "CALL" {
		t.Errorf("Right = %q, want CALL", leg.Right)
	}
}

// 测试 NewContractLeg 股票腿（无 expiry/strike/right）
func TestNewContractLeg_stockLeg(t *testing.T) {
	leg := NewContractLeg("AAPL", "STK", "BUY", 1, "", "", "")
	if leg.Expiry != "" {
		t.Errorf("Expiry = %q, want empty", leg.Expiry)
	}
	if leg.Strike != "" {
		t.Errorf("Strike = %q, want empty", leg.Strike)
	}
	if leg.Right != "" {
		t.Errorf("Right = %q, want empty", leg.Right)
	}
	if leg.Ratio == nil || *leg.Ratio != 1 {
		t.Errorf("Ratio = %v, want 指向 1", leg.Ratio)
	}
}

// 测试 NewContractLeg 零比率
func TestNewContractLeg_zeroRatio(t *testing.T) {
	leg := NewContractLeg("AAPL", "OPT", "SELL", 0, "20250620", "150.0", "PUT")
	if leg.Ratio == nil {
		t.Fatal("Ratio 不应为 nil")
	}
	if *leg.Ratio != 0 {
		t.Errorf("Ratio = %d, want 0", *leg.Ratio)
	}
}
