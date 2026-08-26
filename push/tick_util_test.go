package push

import (
	"testing"

	"github.com/tigerfintech/openapi-go-sdk/push/pb"
)

func TestGetPartShortNameByCode(t *testing.T) {
	cases := []struct{ code, want string }{
		{"n", "NYSE"},
		{"t", "NSDQ"},
		{"p", "ARCA"},
		{"z", "BZX"},
		{"a", "AMEX"},
		{"", ""},
		{"?", "?"},
	}
	for _, c := range cases {
		if got := GetPartShortNameByCode(c.code); got != c.want {
			t.Errorf("GetPartShortNameByCode(%q) = %q, want %q", c.code, got, c.want)
		}
	}
}

func TestGetPartNameByCode(t *testing.T) {
	cases := []struct{ code, want string }{
		{"n", "New York Stock Exchange, LLC (NYSE)"},
		{"t", "NASDAQ Stock Market, LLC (NASDAQ)"},
		{"", ""},
		{"?", "?"},
	}
	for _, c := range cases {
		if got := GetPartNameByCode(c.code); got != c.want {
			t.Errorf("GetPartNameByCode(%q) = %q, want %q", c.code, got, c.want)
		}
	}
}

func TestConvertTradeTick_Stock(t *testing.T) {
	src := &pb.TradeTickData{
		Symbol:      "AAPL",
		SecType:     "STK",
		QuoteLevel:  "L1",
		Timestamp:   1700000000,
		Sn:          100,
		PriceBase:   1500000,
		PriceOffset: 4,
		Time:        []int64{1000, 500, 300},
		Price:       []int64{0, 100, -50},
		Volume:      []int64{200, 100, 150},
		PartCode:    []string{"n", "t", "p"},
		Cond:        "B T",
		Type:        "+-*",
	}

	out := ConvertTradeTick(src)

	if out.Symbol != "AAPL" {
		t.Errorf("Symbol = %q", out.Symbol)
	}
	if out.SecType != "STK" {
		t.Errorf("SecType = %q", out.SecType)
	}
	if len(out.Ticks) != 3 {
		t.Fatalf("len(Ticks) = %d, want 3", len(out.Ticks))
	}

	// time is cumulative
	if out.Ticks[0].Time != 1000 {
		t.Errorf("Ticks[0].Time = %d, want 1000", out.Ticks[0].Time)
	}
	if out.Ticks[1].Time != 1500 {
		t.Errorf("Ticks[1].Time = %d, want 1500", out.Ticks[1].Time)
	}
	if out.Ticks[2].Time != 1800 {
		t.Errorf("Ticks[2].Time = %d, want 1800", out.Ticks[2].Time)
	}

	// price: (priceBase + price[i]) / 10^priceOffset
	wantPrice0 := float64(1500000+0) / 10000.0
	if out.Ticks[0].Price != wantPrice0 {
		t.Errorf("Ticks[0].Price = %v, want %v", out.Ticks[0].Price, wantPrice0)
	}

	// partCode / partName
	if out.Ticks[0].PartCode != "NYSE" {
		t.Errorf("Ticks[0].PartCode = %q, want NYSE", out.Ticks[0].PartCode)
	}
	if out.Ticks[0].PartName != "New York Stock Exchange, LLC (NYSE)" {
		t.Errorf("Ticks[0].PartName = %q", out.Ticks[0].PartName)
	}
	if out.Ticks[1].PartCode != "NSDQ" {
		t.Errorf("Ticks[1].PartCode = %q, want NSDQ", out.Ticks[1].PartCode)
	}
	if out.Ticks[2].PartCode != "ARCA" {
		t.Errorf("Ticks[2].PartCode = %q, want ARCA", out.Ticks[2].PartCode)
	}

	// cond: US stock 'B' -> US_BUNCHED_TRADE
	if out.Ticks[0].Cond != "US_BUNCHED_TRADE" {
		t.Errorf("Ticks[0].Cond = %q, want US_BUNCHED_TRADE", out.Ticks[0].Cond)
	}
	// cond: US stock ' ' -> US_REGULAR_SALE
	if out.Ticks[1].Cond != "US_REGULAR_SALE" {
		t.Errorf("Ticks[1].Cond = %q, want US_REGULAR_SALE", out.Ticks[1].Cond)
	}

	// sn
	if out.Ticks[0].Sn != 100 {
		t.Errorf("Ticks[0].Sn = %d, want 100", out.Ticks[0].Sn)
	}
	if out.Ticks[2].Sn != 102 {
		t.Errorf("Ticks[2].Sn = %d, want 102", out.Ticks[2].Sn)
	}
}

func TestConvertTradeTick_Future(t *testing.T) {
	src := &pb.TradeTickData{
		Symbol:      "ES2312",
		SecType:     "FUT",
		Timestamp:   1700000000,
		Sn:          10,
		PriceBase:   45000000,
		PriceOffset: 4,
		Time:        []int64{1000, 500},
		Price:       []int64{0, 100},
		Volume:      []int64{}, // not used for FUT
		MergedVols: []*pb.TradeTickData_MergedVol{
			{MergeTimes: 2, Vol: []int64{100, 200}},
			{MergeTimes: 1, Vol: []int64{300}},
		},
	}

	out := ConvertTradeTick(src)

	if out.SecType != "FUT" {
		t.Errorf("SecType = %q", out.SecType)
	}
	if len(out.Ticks) != 3 {
		t.Fatalf("len(Ticks) = %d, want 3", len(out.Ticks))
	}

	// first merged group: sn=10*10+0, 10*10+1
	if out.Ticks[0].Sn != 100 {
		t.Errorf("Ticks[0].Sn = %d, want 100", out.Ticks[0].Sn)
	}
	if out.Ticks[1].Sn != 101 {
		t.Errorf("Ticks[1].Sn = %d, want 101", out.Ticks[1].Sn)
	}
	// second merged group: sn=11*10+0
	if out.Ticks[2].Sn != 110 {
		t.Errorf("Ticks[2].Sn = %d, want 110", out.Ticks[2].Sn)
	}

	if out.Ticks[0].Volume != 100 {
		t.Errorf("Ticks[0].Volume = %d, want 100", out.Ticks[0].Volume)
	}
	if out.Ticks[2].Volume != 300 {
		t.Errorf("Ticks[2].Volume = %d, want 300", out.Ticks[2].Volume)
	}

	// future ticks have no partCode
	if out.Ticks[0].PartCode != "" {
		t.Errorf("FUT Ticks[0].PartCode should be empty")
	}
}

func TestConvertTradeTick_HKCond(t *testing.T) {
	src := &pb.TradeTickData{
		Symbol:      "00700.HK",
		SecType:     "STK",
		Sn:          1,
		PriceBase:   3000000,
		PriceOffset: 4,
		Time:        []int64{1000},
		Price:       []int64{0},
		Volume:      []int64{100},
		PartCode:    []string{},
		Cond:        "U",
	}

	out := ConvertTradeTick(src)
	if out.Ticks[0].Cond != "HK_AUCTION_TRADE" {
		t.Errorf("HK cond 'U' = %q, want HK_AUCTION_TRADE", out.Ticks[0].Cond)
	}
}
