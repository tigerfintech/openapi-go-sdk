package push

import (
	"github.com/tigerfintech/openapi-go-sdk/push/pb"
)

// Callbacks 所有回调函数的集合
type Callbacks struct {
	// 行情推送回调
	OnQuote     func(data *pb.QuoteData)
	OnTick      func(data *pb.TradeTickData)
	OnDepth     func(data *pb.QuoteDepthData)
	OnOption    func(data *pb.QuoteData)
	OnFuture    func(data *pb.QuoteData)
	OnKline     func(data *pb.KlineData)
	OnStockTop  func(data *pb.StockTopData)
	OnOptionTop func(data *pb.OptionTopData)
	OnFullTick  func(data *pb.TickData)
	OnQuoteBBO  func(data *pb.QuoteData)

	// 账户推送回调
	OnAsset       func(data *pb.AssetData)
	OnPosition    func(data *pb.PositionData)
	OnOrder       func(data *pb.OrderStatusData)
	OnTransaction func(data *pb.OrderTransactionData)

	// 连接状态回调
	OnConnect    func()
	OnDisconnect func()
	OnError      func(err error)
	OnKickout    func(message string)
}
