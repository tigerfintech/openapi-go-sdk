package model

import (
	"encoding/json"
	"testing"
)

// 测试 MarketOrder 构造函数
func TestMarketOrder(t *testing.T) {
	o := MarketOrder("U123456", "AAPL", "STK", "BUY", 100)
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
	if o.TotalQuantity != 100 {
		t.Errorf("TotalQuantity = %d, want 100", o.TotalQuantity)
	}
	if o.TimeInForce != string(TimeInForceDAY) {
		t.Errorf("TimeInForce = %q, want DAY", o.TimeInForce)
	}
}

// 测试 LimitOrder 构造函数
func TestLimitOrder(t *testing.T) {
	o := LimitOrder("U123456", "AAPL", "STK", "BUY", 100, 150.50)
	if o.OrderType != string(OrderTypeLMT) {
		t.Errorf("OrderType = %q, want LMT", o.OrderType)
	}
	if o.LimitPrice != 150.50 {
		t.Errorf("LimitPrice = %f, want 150.50", o.LimitPrice)
	}
	if o.TimeInForce != string(TimeInForceDAY) {
		t.Errorf("TimeInForce = %q, want DAY", o.TimeInForce)
	}
}

// 测试 StopOrder 构造函数
func TestStopOrder(t *testing.T) {
	o := StopOrder("U123456", "AAPL", "STK", "SELL", 100, 140.00)
	if o.OrderType != string(OrderTypeSTP) {
		t.Errorf("OrderType = %q, want STP", o.OrderType)
	}
	if o.AuxPrice != 140.00 {
		t.Errorf("AuxPrice = %f, want 140.00", o.AuxPrice)
	}
}

// 测试 StopLimitOrder 构造函数
func TestStopLimitOrder(t *testing.T) {
	o := StopLimitOrder("U123456", "AAPL", "STK", "SELL", 100, 145.00, 140.00)
	if o.OrderType != string(OrderTypeSTPLMT) {
		t.Errorf("OrderType = %q, want STP_LMT", o.OrderType)
	}
	if o.LimitPrice != 145.00 {
		t.Errorf("LimitPrice = %f, want 145.00", o.LimitPrice)
	}
	if o.AuxPrice != 140.00 {
		t.Errorf("AuxPrice = %f, want 140.00", o.AuxPrice)
	}
}

// 测试 TrailOrder 构造函数
func TestTrailOrder(t *testing.T) {
	o := TrailOrder("U123456", "AAPL", "STK", "SELL", 100, 5.0)
	if o.OrderType != string(OrderTypeTRAIL) {
		t.Errorf("OrderType = %q, want TRAIL", o.OrderType)
	}
	if o.TrailingPercent != 5.0 {
		t.Errorf("TrailingPercent = %f, want 5.0", o.TrailingPercent)
	}
}

// 测试 AuctionLimitOrder 构造函数
func TestAuctionLimitOrder(t *testing.T) {
	o := AuctionLimitOrder("U123456", "AAPL", "STK", "BUY", 100, 150.00)
	if o.OrderType != string(OrderTypeAL) {
		t.Errorf("OrderType = %q, want AL", o.OrderType)
	}
	if o.LimitPrice != 150.00 {
		t.Errorf("LimitPrice = %f, want 150.00", o.LimitPrice)
	}
}

// 测试 AuctionMarketOrder 构造函数
func TestAuctionMarketOrder(t *testing.T) {
	o := AuctionMarketOrder("U123456", "AAPL", "STK", "BUY", 100)
	if o.OrderType != string(OrderTypeAM) {
		t.Errorf("OrderType = %q, want AM", o.OrderType)
	}
}

// AlgoParamsRequest 必须序列化成 [{tag, value}, ...] 数组
// (网关期望的形状,与 Python SDK 的 AlgoParams.to_dict 一致)。
// 直接发 object 会触发 "failed to parse parameters in 'biz_content'"。
func TestAlgoParamsRequestMarshalsAsTagValueArray(t *testing.T) {
	yes := true
	p := AlgoParamsRequest{
		StartTime:         1700000000000,
		EndTime:           1700003600000,
		ParticipationRate: 0.1,
		AllowPastEndTime:  &yes,
	}
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var arr []map[string]interface{}
	if err := json.Unmarshal(data, &arr); err != nil {
		t.Fatalf("output is not a JSON array of {tag,value}: %v, raw=%s", err, string(data))
	}
	tags := map[string]interface{}{}
	for _, entry := range arr {
		if tag, ok := entry["tag"].(string); ok {
			tags[tag] = entry["value"]
		}
	}
	if v, ok := tags["start_time"]; !ok || v.(float64) != 1.7e12 {
		t.Errorf("start_time missing or wrong: got %v (raw=%s)", v, string(data))
	}
	if v, ok := tags["end_time"]; !ok || v.(float64) != 1700003600000 {
		t.Errorf("end_time missing or wrong: got %v", v)
	}
	if v, ok := tags["participation_rate"]; !ok || v.(float64) != 0.1 {
		t.Errorf("participation_rate missing or wrong: got %v", v)
	}
	if v, ok := tags["allow_past_end_time"]; !ok || v != true {
		t.Errorf("allow_past_end_time missing or wrong: got %v", v)
	}
}

// 未设置的字段应被丢弃(不出现在 tag/value 数组里)。
func TestAlgoParamsRequestOmitsZeroFields(t *testing.T) {
	// 只设一个字段,其他保持 zero value
	p := AlgoParamsRequest{StartTime: 100}
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var arr []map[string]interface{}
	if err := json.Unmarshal(data, &arr); err != nil {
		t.Fatalf("output not array: %v raw=%s", err, string(data))
	}
	if len(arr) != 1 {
		t.Fatalf("expected exactly 1 tag entry (start_time), got %d: %+v", len(arr), arr)
	}
	if arr[0]["tag"] != "start_time" {
		t.Errorf("expected tag=start_time, got %v", arr[0]["tag"])
	}
}

// 嵌入到 OrderRequest 后 algo_params 在最终 JSON 里也应是数组。
func TestOrderRequestNestsAlgoParamsAsArray(t *testing.T) {
	req := OrderRequest{
		Symbol:        "AAPL",
		SecType:       "STK",
		Action:        "BUY",
		OrderType:     "TWAP",
		TotalQuantity: 100,
		AlgoStrategy:  "TWAP",
		AlgoParams: &AlgoParamsRequest{
			StartTime: 1700000000000,
			EndTime:   1700003600000,
		},
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal failed: %v raw=%s", err, string(data))
	}
	if _, ok := m["algo_params"].([]interface{}); !ok {
		t.Errorf("expected algo_params to be array, got %T: %+v", m["algo_params"], m["algo_params"])
	}
	if m["algo_strategy"] != "TWAP" {
		t.Errorf("expected algo_strategy=TWAP at top level, got %v", m["algo_strategy"])
	}
}

// 测试 AlgoOrder 构造函数
func TestAlgoOrder(t *testing.T) {
	// AlgoStrategy 现在是 OrderRequest 顶层字段(不在 algo_params 里);
	// StartTime / EndTime 用 epoch-ms(int64) 而不是字符串。
	params := AlgoParamsRequest{
		StartTime:         1700000000000,
		EndTime:           1700003600000,
		ParticipationRate: 0.1,
	}
	o := AlgoOrder("U123456", "AAPL", "STK", "BUY", 1000, 150.00, "TWAP", params)
	if o.OrderType != string(OrderTypeTWAP) {
		t.Errorf("OrderType = %q, want TWAP", o.OrderType)
	}
	if o.AlgoStrategy != "TWAP" {
		t.Errorf("AlgoStrategy = %q, want TWAP", o.AlgoStrategy)
	}
	if o.LimitPrice != 150.00 {
		t.Errorf("LimitPrice = %f, want 150.00", o.LimitPrice)
	}
	if o.AlgoParams == nil {
		t.Fatal("AlgoParams 不应为 nil")
	}
	if o.AlgoParams.ParticipationRate != 0.1 {
		t.Errorf("ParticipationRate = %f, want 0.1", o.AlgoParams.ParticipationRate)
	}
}

// 测试 NewOrderLeg 构造函数
func TestNewOrderLeg(t *testing.T) {
	leg := NewOrderLeg("PROFIT", 160.00, "GTC")
	if leg.LegType != "PROFIT" {
		t.Errorf("LegType = %q, want PROFIT", leg.LegType)
	}
	if leg.Price != 160.00 {
		t.Errorf("Price = %f, want 160.00", leg.Price)
	}
	if leg.TimeInForce != "GTC" {
		t.Errorf("TimeInForce = %q, want GTC", leg.TimeInForce)
	}
}

// 测试 IcebergOrder 构造函数（最简参数）
func TestIcebergOrder(t *testing.T) {
	o := IcebergOrder("U123456", "AAPL", "STK", "BUY", 1000, 180.0, 100)
	if o.OrderType != string(OrderTypeICEBERG) {
		t.Errorf("OrderType = %q, want ICEBERG", o.OrderType)
	}
	if o.TotalQuantity != 1000 {
		t.Errorf("TotalQuantity = %d, want 1000", o.TotalQuantity)
	}
	if o.LimitPrice != 180.0 {
		t.Errorf("LimitPrice = %f, want 180.0", o.LimitPrice)
	}
	if o.DisplaySize != 100 {
		t.Errorf("DisplaySize = %d, want 100", o.DisplaySize)
	}
	if o.TimeInForce != string(TimeInForceDAY) {
		t.Errorf("TimeInForce = %q, want DAY", o.TimeInForce)
	}
	// 未填的可选字段应为零值
	if o.MinDisplaySize != 0 {
		t.Errorf("MinDisplaySize = %d, want 0 (not set)", o.MinDisplaySize)
	}
	if o.StartTime != 0 {
		t.Errorf("StartTime = %d, want 0 (not set)", o.StartTime)
	}
}
