package model

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
	}
}

// IcebergOrderFull 构造冰山单（完整参数）。
// priceType 传空字符串时使用服务端默认值（LIMIT_PRICE）。
// startTime / endTime 传 0 时不发送（服务端不限制生效时间范围）。
func IcebergOrderFull(account, symbol, secType, action string, quantity int64, limitPrice float64,
	displaySize, minDisplaySize, checkIntervals int64, priceType IcebergPriceType,
	startTime, endTime int64) OrderRequest {
	req := OrderRequest{
		Account:        account,
		Symbol:         symbol,
		SecType:        secType,
		Action:         action,
		OrderType:      string(OrderTypeICEBERG),
		TotalQuantity:  quantity,
		LimitPrice:     limitPrice,
		TimeInForce:    string(TimeInForceDAY),
		DisplaySize:    displaySize,
		MinDisplaySize: minDisplaySize,
		CheckIntervals: checkIntervals,
	}
	if priceType != "" {
		req.PriceType = string(priceType)
	}
	if startTime > 0 {
		req.StartTime = startTime
	}
	if endTime > 0 {
		req.EndTime = endTime
	}
	return req
}
