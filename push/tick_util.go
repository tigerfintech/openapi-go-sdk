package push

import (
	"math"
	"regexp"

	"github.com/tigerfintech/openapi-go-sdk/push/pb"
)

var partCodeNameMap = map[string]string{
	"a": "NYSE American, LLC (NYSE American)",
	"b": "NASDAQ OMX BX, Inc. (NASDAQ OMX BX)",
	"c": "NYSE National, Inc. (NYSE National)",
	"d": "FINRA Alternative Display Facility (ADF)",
	"h": "MIAX Pearl Exchange, LLC (MIAX)",
	"i": "International Securities Exchange, LLC (ISE)",
	"j": "Cboe EDGA Exchange, Inc. (Cboe EDGA)",
	"k": "Cboe EDGX Exchange, Inc. (Cboe EDGX)",
	"l": "Long-Term Stock Exchange, Inc. (LTSE)",
	"m": "NYSE Chicago, Inc. (NYSE Chicago)",
	"n": "New York Stock Exchange, LLC (NYSE)",
	"p": "NYSE Arca, Inc. (NYSE Arca)",
	"s": "Consolidated Tape System (CTS)",
	"t": "NASDAQ Stock Market, LLC (NASDAQ)",
	"u": "Members Exchange, LLC (MEMX)",
	"v": "Investors' Exchange, LLC. (IEX)",
	"w": "CBOE Stock Exchange, Inc. (CBSX)",
	"x": "NASDAQ OMX PSX, Inc. (NASDAQ OMX PSX)",
	"y": "Cboe BYX Exchange, Inc. (Cboe BYX)",
	"z": "Cboe BZX Exchange, Inc. (Cboe BZX)",
}

var partCodeShortNameMap = map[string]string{
	"a": "AMEX",
	"b": "BX",
	"c": "NSX",
	"d": "ADF",
	"h": "MIAX",
	"i": "ISE",
	"j": "EDGA",
	"k": "EDGX",
	"l": "LTSE",
	"m": "CHO",
	"n": "NYSE",
	"p": "ARCA",
	"s": "CTS",
	"t": "NSDQ",
	"u": "MEMX",
	"v": "IEX",
	"w": "CBSX",
	"x": "PSX",
	"y": "BYX",
	"z": "BZX",
}

var usTradeCondMap = map[rune]string{
	' ': "US_REGULAR_SALE",
	'B': "US_BUNCHED_TRADE",
	'C': "US_CASH_TRADE",
	'F': "US_INTERMARKET_SWEEP",
	'G': "US_BUNCHED_SOLD_TRADE",
	'H': "US_PRICE_VARIATION_TRADE",
	'I': "US_ODD_LOT_TRADE",
	'K': "US_RULE_127_OR_155_TRADE",
	'L': "US_SOLD_LAST",
	'M': "US_MARKET_CENTER_CLOSE_PRICE",
	'N': "US_NEXT_DAY_TRADE",
	'O': "US_MARKET_CENTER_OPENING_TRADE",
	'P': "US_PRIOR_REFERENCE_PRICE",
	'Q': "US_MARKET_CENTER_OPEN_PRICE",
	'R': "US_SELLER",
	'T': "US_FORM_T",
	'U': "US_EXTENDED_TRADING_HOURS",
	'V': "US_CONTINGENT_TRADE",
	'W': "US_AVERAGE_PRICE_TRADE",
	'X': "US_CROSS_TRADE",
	'Z': "US_SOLD_OUT_OF_SEQUENCE",
	'0': "US_ODD_LOST_CROSS_TRADE",
	'4': "US_DERIVATIVELY_PRICED",
	'5': "US_MARKET_CENTER_RE_OPENING_TRADE",
	'6': "US_MARKET_CENTER_CLOSING_TRADE",
	'7': "US_QUALIFIED_CONTINGENT_TRADE",
	'9': "US_CONSOLIDATED_LAST_PRICE_PER_LISTING_PACKET",
}

var hkTradeCondMap = map[rune]string{
	' ': "HK_AUTOMATCH_NORMAL",
	'D': "HK_ODD_LOT_TRADE",
	'U': "HK_AUCTION_TRADE",
	'*': "HK_OVERSEAS_TRADE",
	'P': "HK_LATE_TRADE_OFF_EXCHG",
	'M': "HK_NON_DIRECT_OFF_EXCHG_TRADE",
	'X': "HK_DIRECT_OFF_EXCHG_TRADE",
	'Y': "HK_AUTOMATIC_INTERNALIZED",
}

// PushTick is a single decoded tick from a push message.
type PushTick struct {
	Sn       int64
	Time     int64
	Price    float64
	Volume   int64
	TickType string
	Cond     string
	PartCode string
	PartName string
}

// PushTradeTick is the decoded push trade tick message, equivalent to Java TradeTick.
type PushTradeTick struct {
	Symbol     string
	SecType    string
	QuoteLevel string
	Timestamp  uint64
	Ticks      []PushTick
}

// GetPartNameByCode returns the full exchange name for a partCode letter.
func GetPartNameByCode(code string) string {
	if code == "" {
		return code
	}
	if name, ok := partCodeNameMap[code]; ok {
		return name
	}
	return code
}

// GetPartShortNameByCode returns the short exchange name for a partCode letter.
func GetPartShortNameByCode(code string) string {
	if code == "" {
		return code
	}
	if name, ok := partCodeShortNameMap[code]; ok {
		return name
	}
	return code
}

func getTradeCondByCode(isUS bool, ch rune) string {
	m := hkTradeCondMap
	if isUS {
		m = usTradeCondMap
	}
	if s, ok := m[ch]; ok {
		return s
	}
	return string(ch)
}

var usSymbolRe = regexp.MustCompile(`^[A-Z]+(\.[A-Z0-9]+)?$`)

func isUsStockSymbol(symbol string) bool {
	return usSymbolRe.MatchString(symbol)
}

// GetTradeCondByCode resolves a raw cond character to a readable string.
// Set isUS=true for US stocks, false for HK stocks.
// Falls back to the raw character if the code is not in the map.
func GetTradeCondByCode(isUS bool, ch rune) string {
	return getTradeCondByCode(isUS, ch)
}

// IsUsStockSymbol reports whether symbol is a US stock (no dot separator).
func IsUsStockSymbol(symbol string) bool {
	return isUsStockSymbol(symbol)
}

// ConvertTradeTick decodes a pb.TradeTickData into a PushTradeTick with resolved partCode/partName.
func ConvertTradeTick(src *pb.TradeTickData) PushTradeTick {
	if src.GetSecType() == "FUT" {
		return convertFutureTick(src)
	}
	return convertStockTick(src)
}

func convertStockTick(src *pb.TradeTickData) PushTradeTick {
	out := PushTradeTick{
		Symbol:     src.GetSymbol(),
		SecType:    src.GetSecType(),
		QuoteLevel: src.GetQuoteLevel(),
		Timestamp:  src.GetTimestamp(),
	}

	timeList := src.GetTime()
	priceList := src.GetPrice()
	volumeList := src.GetVolume()
	partCodeList := src.GetPartCode()
	cond := src.GetCond()
	tickType := src.GetType()
	startSn := src.GetSn()
	isUS := isUsStockSymbol(src.GetSymbol())
	denominator := math.Pow10(int(src.GetPriceOffset()))
	priceBase := src.GetPriceBase()

	var currentTime int64
	ticks := make([]PushTick, 0, len(timeList))
	for i, dt := range timeList {
		currentTime += dt
		tick := PushTick{
			Sn:     startSn + int64(i),
			Time:   currentTime,
			Price:  float64(priceBase+priceList[i]) / denominator,
			Volume: volumeList[i],
		}
		if i < len(partCodeList) {
			tick.PartCode = GetPartShortNameByCode(partCodeList[i])
			tick.PartName = GetPartNameByCode(partCodeList[i])
		}
		condCh := ' '
		if i < len(cond) {
			condCh = rune(cond[i])
		}
		tick.Cond = getTradeCondByCode(isUS, condCh)
		if i < len(tickType) {
			tick.TickType = string(tickType[i])
		}
		ticks = append(ticks, tick)
	}
	out.Ticks = ticks
	return out
}

func convertFutureTick(src *pb.TradeTickData) PushTradeTick {
	out := PushTradeTick{
		Symbol:    src.GetSymbol(),
		SecType:   src.GetSecType(),
		Timestamp: src.GetTimestamp(),
	}

	timeList := src.GetTime()
	priceList := src.GetPrice()
	mergedVols := src.GetMergedVols()
	startSn := src.GetSn()
	denominator := math.Pow10(int(src.GetPriceOffset()))
	priceBase := src.GetPriceBase()

	var currentTime int64
	var ticks []PushTick
	for i, dt := range timeList {
		currentTime += dt
		curPrice := float64(priceBase+priceList[i]) / denominator
		mv := mergedVols[i]
		for j, vol := range mv.GetVol() {
			ticks = append(ticks, PushTick{
				Sn:     startSn*10 + int64(j),
				Time:   currentTime,
				Price:  curPrice,
				Volume: vol,
			})
		}
		startSn++
	}
	out.Ticks = ticks
	return out
}
