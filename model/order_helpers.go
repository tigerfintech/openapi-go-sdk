package model

import "fmt"

// MarketOrder 构造市价单
func MarketOrder(account, symbol, secType, action string, quantity int64) OrderRequest {
	return OrderRequest{
		Account:       account,
		Symbol:        symbol,
		SecType:       secType,
		Action:        action,
		OrderType:     string(OrderTypeMKT),
		TotalQuantity: quantity,
		TimeInForce:   string(TimeInForceDAY),
	}
}

// LimitOrder 构造限价单
func LimitOrder(account, symbol, secType, action string, quantity int64, limitPrice float64) OrderRequest {
	return OrderRequest{
		Account:       account,
		Symbol:        symbol,
		SecType:       secType,
		Action:        action,
		OrderType:     string(OrderTypeLMT),
		TotalQuantity: quantity,
		LimitPrice:    limitPrice,
		TimeInForce:   string(TimeInForceDAY),
	}
}

// StopOrder 构造止损单
func StopOrder(account, symbol, secType, action string, quantity int64, auxPrice float64) OrderRequest {
	return OrderRequest{
		Account:       account,
		Symbol:        symbol,
		SecType:       secType,
		Action:        action,
		OrderType:     string(OrderTypeSTP),
		TotalQuantity: quantity,
		AuxPrice:      auxPrice,
		TimeInForce:   string(TimeInForceDAY),
	}
}

// StopLimitOrder 构造止损限价单
func StopLimitOrder(account, symbol, secType, action string, quantity int64, limitPrice, auxPrice float64) OrderRequest {
	return OrderRequest{
		Account:       account,
		Symbol:        symbol,
		SecType:       secType,
		Action:        action,
		OrderType:     string(OrderTypeSTPLMT),
		TotalQuantity: quantity,
		LimitPrice:    limitPrice,
		AuxPrice:      auxPrice,
		TimeInForce:   string(TimeInForceDAY),
	}
}

// TrailOrder 构造跟踪止损单
func TrailOrder(account, symbol, secType, action string, quantity int64, trailingPercent float64) OrderRequest {
	return OrderRequest{
		Account:         account,
		Symbol:          symbol,
		SecType:         secType,
		Action:          action,
		OrderType:       string(OrderTypeTRAIL),
		TotalQuantity:   quantity,
		TrailingPercent: trailingPercent,
		TimeInForce:     string(TimeInForceDAY),
	}
}

// AuctionLimitOrder 构造竞价限价单
func AuctionLimitOrder(account, symbol, secType, action string, quantity int64, limitPrice float64) OrderRequest {
	return OrderRequest{
		Account:       account,
		Symbol:        symbol,
		SecType:       secType,
		Action:        action,
		OrderType:     string(OrderTypeAL),
		TotalQuantity: quantity,
		LimitPrice:    limitPrice,
		TimeInForce:   string(TimeInForceDAY),
	}
}

// AuctionMarketOrder 构造竞价市价单
func AuctionMarketOrder(account, symbol, secType, action string, quantity int64) OrderRequest {
	return OrderRequest{
		Account:       account,
		Symbol:        symbol,
		SecType:       secType,
		Action:        action,
		OrderType:     string(OrderTypeAM),
		TotalQuantity: quantity,
		TimeInForce:   string(TimeInForceDAY),
	}
}

// AlgoOrder 构造算法订单（TWAP/VWAP）
func AlgoOrder(account, symbol, secType, action string, quantity int64, limitPrice float64, algoType string, params AlgoParamsRequest) OrderRequest {
	return OrderRequest{
		Account:       account,
		Symbol:        symbol,
		SecType:       secType,
		Action:        action,
		OrderType:     algoType,
		TotalQuantity: quantity,
		LimitPrice:    limitPrice,
		AlgoParams:    &params,
		TimeInForce:   string(TimeInForceDAY),
	}
}

// NewOrderLeg 构造附加订单（止盈/止损）
func NewOrderLeg(legType string, price float64, timeInForce string) OrderLegRequest {
	return OrderLegRequest{
		LegType:     legType,
		Price:       price,
		TimeInForce: timeInForce,
	}
}

// IcebergOrder 构造冰山单（最简参数）。
// displaySize 为每次展示数量；其余冰山参数使用服务端默认值。
// 需要设置更多参数（minDisplaySize/checkIntervals/priceType/startTime/endTime）时，
// 直接在返回的 OrderRequest struct 上赋值即可。
func IcebergOrder(account, symbol, secType, action string, quantity int64, limitPrice float64, displaySize int64) OrderRequest {
	return OrderRequest{
		Account:       account,
		Symbol:        symbol,
		SecType:       secType,
		Action:        action,
		OrderType:     string(OrderTypeICEBERG),
		TotalQuantity: quantity,
		LimitPrice:    limitPrice,
		TimeInForce:   string(TimeInForceDAY),
		DisplaySize:   displaySize,
		PriceType:     string(PriceTypeLimitPrice),
	}
}

// MarketOrderByAmount 构造按金额市价单（以现金金额代替数量下单）
func MarketOrderByAmount(account, symbol, secType, action string, amount float64) OrderRequest {
	return OrderRequest{
		Account:       account,
		Symbol:        symbol,
		SecType:       secType,
		Action:        action,
		OrderType:     string(OrderTypeMKT),
		TotalQuantity: 0,
		Amount:        amount,
		TimeInForce:   string(TimeInForceDAY),
	}
}

// LimitOrderByAmount 构造按金额限价单（以现金金额代替数量下单）
func LimitOrderByAmount(account, symbol, secType, action string, amount, limitPrice float64) OrderRequest {
	return OrderRequest{
		Account:       account,
		Symbol:        symbol,
		SecType:       secType,
		Action:        action,
		OrderType:     string(OrderTypeLMT),
		TotalQuantity: 0,
		Amount:        amount,
		LimitPrice:    limitPrice,
		TimeInForce:   string(TimeInForceDAY),
	}
}

// TrailOrderByPrice 构造按固定价差跟踪止损单（使用 AuxPrice 而非百分比）
func TrailOrderByPrice(account, symbol, secType, action string, quantity int64, auxPrice float64) OrderRequest {
	return OrderRequest{
		Account:       account,
		Symbol:        symbol,
		SecType:       secType,
		Action:        action,
		OrderType:     string(OrderTypeTRAIL),
		TotalQuantity: quantity,
		AuxPrice:      auxPrice,
		TimeInForce:   string(TimeInForceDAY),
	}
}

// LimitOrderWithLegs 构造限价单 + 附加止盈/止损腿（bracket 单，最多 2 腿）。
// 返回 error 当 orderLegs 为空或超过 2 条。
func LimitOrderWithLegs(account, symbol, secType, action string, quantity int64, limitPrice float64, orderLegs []OrderLegRequest) (OrderRequest, error) {
	if len(orderLegs) == 0 {
		return OrderRequest{}, fmt.Errorf("LimitOrderWithLegs: at least 1 order leg is required")
	}
	if len(orderLegs) > 2 {
		return OrderRequest{}, fmt.Errorf("LimitOrderWithLegs: at most 2 order legs are supported (got %d)", len(orderLegs))
	}
	return OrderRequest{
		Account:       account,
		Symbol:        symbol,
		SecType:       secType,
		Action:        action,
		OrderType:     string(OrderTypeLMT),
		TotalQuantity: quantity,
		LimitPrice:    limitPrice,
		OrderLegs:     orderLegs,
		TimeInForce:   string(TimeInForceDAY),
	}, nil
}

// ComboOrder 构造多腿组合单（MLEG）
// comboType 通常为 "MLEG"；limitPrice/auxPrice/trailingPercent 按需传入（零值省略）。
func ComboOrder(account, action, orderType string, quantity int64, contractLegs []ContractLegRequest, comboType string, limitPrice, auxPrice, trailingPercent float64) OrderRequest {
	req := OrderRequest{
		Account:       account,
		SecType:       "MLEG",
		Action:        action,
		OrderType:     orderType,
		TotalQuantity: quantity,
		ContractLegs:  contractLegs,
		ComboType:     comboType,
		TimeInForce:   string(TimeInForceDAY),
	}
	if limitPrice != 0 {
		req.LimitPrice = limitPrice
	}
	if auxPrice != 0 {
		req.AuxPrice = auxPrice
	}
	if trailingPercent != 0 {
		req.TrailingPercent = trailingPercent
	}
	return req
}

// OcaOrder 构造 OCA（One-Cancels-All）单
func OcaOrder(account, symbol, secType, action string, quantity int64, ocaOrders []*OrderRequest) OrderRequest {
	return OrderRequest{
		Account:       account,
		Symbol:        symbol,
		SecType:       secType,
		Action:        action,
		OrderType:     string(OrderTypeOCA),
		TotalQuantity: quantity,
		OcaOrders:     ocaOrders,
		TimeInForce:   string(TimeInForceDAY),
	}
}

// NewContractLeg 构造多腿组合单的子腿
func NewContractLeg(symbol, secType, action string, ratio int, expiry, strike, right string) ContractLegRequest {
	return ContractLegRequest{
		Symbol:  symbol,
		SecType: secType,
		Action:  action,
		Ratio:   &ratio,
		Expiry:  expiry,
		Strike:  strike,
		Right:   right,
	}
}
