package push

import (
	"sync/atomic"

	"github.com/tigerfintech/openapi-go-sdk/push/pb"
	"google.golang.org/protobuf/proto"
)

// requestID is an atomic counter for assigning unique IDs to each request.
var requestID uint32

// nextRequestID returns the next request ID by atomically incrementing the counter.
func nextRequestID() uint32 {
	return atomic.AddUint32(&requestID, 1)
}

// BuildConnectMessage constructs a CONNECT Request message for authentication.
func BuildConnectMessage(tigerId, sign, sdkVersion, acceptVersion string, sendInterval, receiveInterval uint32, useFullTick bool) *pb.Request {
	return &pb.Request{
		Command: pb.SocketCommon_CONNECT,
		Id:      nextRequestID(),
		Connect: &pb.Request_Connect{
			TigerId:         tigerId,
			Sign:            sign,
			SdkVersion:      sdkVersion,
			AcceptVersion:   proto.String(acceptVersion),
			SendInterval:    proto.Uint32(sendInterval),
			ReceiveInterval: proto.Uint32(receiveInterval),
			UseFullTick:     proto.Bool(useFullTick),
		},
	}
}

// BuildHeartBeatMessage constructs a HEARTBEAT Request message.
func BuildHeartBeatMessage() *pb.Request {
	return &pb.Request{
		Command: pb.SocketCommon_HEARTBEAT,
		Id:      nextRequestID(),
	}
}

// BuildSubscribeMessage constructs a SUBSCRIBE Request message.
// symbols is a comma-separated string of symbol codes.
func BuildSubscribeMessage(dataType pb.SocketCommon_DataType, symbols, account, market string) *pb.Request {
	sub := &pb.Request_Subscribe{
		DataType: dataType,
	}
	if symbols != "" {
		sub.Symbols = proto.String(symbols)
	}
	if account != "" {
		sub.Account = proto.String(account)
	}
	if market != "" {
		sub.Market = proto.String(market)
	}
	return &pb.Request{
		Command:   pb.SocketCommon_SUBSCRIBE,
		Id:        nextRequestID(),
		Subscribe: sub,
	}
}

// BuildUnSubscribeMessage constructs an UNSUBSCRIBE Request message.
// symbols is a comma-separated string of symbol codes.
func BuildUnSubscribeMessage(dataType pb.SocketCommon_DataType, symbols, account, market string) *pb.Request {
	sub := &pb.Request_Subscribe{
		DataType: dataType,
	}
	if symbols != "" {
		sub.Symbols = proto.String(symbols)
	}
	if account != "" {
		sub.Account = proto.String(account)
	}
	if market != "" {
		sub.Market = proto.String(market)
	}
	return &pb.Request{
		Command:   pb.SocketCommon_UNSUBSCRIBE,
		Id:        nextRequestID(),
		Subscribe: sub,
	}
}

// BuildDisconnectMessage constructs a DISCONNECT Request message.
func BuildDisconnectMessage() *pb.Request {
	return &pb.Request{
		Command: pb.SocketCommon_DISCONNECT,
		Id:      nextRequestID(),
	}
}

// SubjectToDataType maps a SubjectType to the corresponding SocketCommon_DataType.
func SubjectToDataType(subject SubjectType) pb.SocketCommon_DataType {
	switch subject {
	case SubjectQuote:
		return pb.SocketCommon_Quote
	case SubjectOption:
		return pb.SocketCommon_Option
	case SubjectFuture:
		return pb.SocketCommon_Future
	case SubjectDepth:
		return pb.SocketCommon_QuoteDepth
	case SubjectTick:
		return pb.SocketCommon_TradeTick
	case SubjectAsset:
		return pb.SocketCommon_Asset
	case SubjectPosition:
		return pb.SocketCommon_Position
	case SubjectOrder:
		return pb.SocketCommon_OrderStatus
	case SubjectTransaction:
		return pb.SocketCommon_OrderTransaction
	case SubjectStockTop:
		return pb.SocketCommon_StockTop
	case SubjectOptionTop:
		return pb.SocketCommon_OptionTop
	case SubjectKline:
		return pb.SocketCommon_Kline
	default:
		return pb.SocketCommon_Unknown
	}
}
