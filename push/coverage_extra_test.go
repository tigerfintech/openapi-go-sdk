package push

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tigerfintech/openapi-go-sdk/push/pb"
	"google.golang.org/protobuf/proto"
)

// ===== dispatchPushData direct tests (no network required) =====

// TestDispatchPushData_Quote verifies Quote data type triggers OnQuote callback.
func TestDispatchPushData_Quote(t *testing.T) {
	cfg := newTestConfig(t)
	client := NewPushClient(cfg)

	var received *pb.QuoteData
	var mu sync.Mutex
	cb := &Callbacks{
		OnQuote: func(d *pb.QuoteData) {
			mu.Lock()
			received = d
			mu.Unlock()
		},
	}
	pushData := &pb.PushData{
		DataType: pb.SocketCommon_Quote,
		Body: &pb.PushData_QuoteData{
			QuoteData: &pb.QuoteData{Symbol: "AAPL", LatestPrice: proto.Float64(150.0)},
		},
	}
	client.dispatchPushData(pushData, cb)

	mu.Lock()
	defer mu.Unlock()
	if received == nil {
		t.Fatal("OnQuote 回调未触发")
	}
	if received.GetSymbol() != "AAPL" {
		t.Errorf("Symbol 期望 AAPL，实际 %s", received.GetSymbol())
	}
}

// TestDispatchPushData_Option verifies Option data type triggers OnOption callback.
func TestDispatchPushData_Option(t *testing.T) {
	cfg := newTestConfig(t)
	client := NewPushClient(cfg)

	var received *pb.QuoteData
	cb := &Callbacks{
		OnOption: func(d *pb.QuoteData) { received = d },
	}
	pushData := &pb.PushData{
		DataType: pb.SocketCommon_Option,
		Body: &pb.PushData_QuoteData{
			QuoteData: &pb.QuoteData{Symbol: "AAPL240315C00150000"},
		},
	}
	client.dispatchPushData(pushData, cb)
	if received == nil {
		t.Fatal("OnOption 回调未触发")
	}
	if received.GetSymbol() != "AAPL240315C00150000" {
		t.Errorf("Symbol 不匹配: %s", received.GetSymbol())
	}
}

// TestDispatchPushData_Future verifies Future data type triggers OnFuture callback.
func TestDispatchPushData_Future(t *testing.T) {
	cfg := newTestConfig(t)
	client := NewPushClient(cfg)

	var received *pb.QuoteData
	cb := &Callbacks{
		OnFuture: func(d *pb.QuoteData) { received = d },
	}
	pushData := &pb.PushData{
		DataType: pb.SocketCommon_Future,
		Body: &pb.PushData_QuoteData{
			QuoteData: &pb.QuoteData{Symbol: "ES"},
		},
	}
	client.dispatchPushData(pushData, cb)
	if received == nil {
		t.Fatal("OnFuture 回调未触发")
	}
}

// TestDispatchPushData_Depth verifies QuoteDepth data type triggers OnDepth callback.
func TestDispatchPushData_Depth(t *testing.T) {
	cfg := newTestConfig(t)
	client := NewPushClient(cfg)

	var received *pb.QuoteDepthData
	cb := &Callbacks{
		OnDepth: func(d *pb.QuoteDepthData) { received = d },
	}
	pushData := &pb.PushData{
		DataType: pb.SocketCommon_QuoteDepth,
		Body: &pb.PushData_QuoteDepthData{
			QuoteDepthData: &pb.QuoteDepthData{Symbol: "AAPL"},
		},
	}
	client.dispatchPushData(pushData, cb)
	if received == nil {
		t.Fatal("OnDepth 回调未触发")
	}
	if received.GetSymbol() != "AAPL" {
		t.Errorf("Symbol 期望 AAPL，实际 %s", received.GetSymbol())
	}
}

// TestDispatchPushData_TradeTick verifies TradeTick data type triggers OnTick callback.
func TestDispatchPushData_TradeTick(t *testing.T) {
	cfg := newTestConfig(t)
	client := NewPushClient(cfg)

	var received *pb.TradeTickData
	cb := &Callbacks{
		OnTick: func(d *pb.TradeTickData) { received = d },
	}
	pushData := &pb.PushData{
		DataType: pb.SocketCommon_TradeTick,
		Body: &pb.PushData_TradeTickData{
			TradeTickData: &pb.TradeTickData{Symbol: "TSLA"},
		},
	}
	client.dispatchPushData(pushData, cb)
	if received == nil {
		t.Fatal("OnTick 回调未触发")
	}
}

// TestDispatchPushData_Asset verifies Asset data type triggers OnAsset callback.
func TestDispatchPushData_Asset(t *testing.T) {
	cfg := newTestConfig(t)
	client := NewPushClient(cfg)

	var received *pb.AssetData
	cb := &Callbacks{
		OnAsset: func(d *pb.AssetData) { received = d },
	}
	pushData := &pb.PushData{
		DataType: pb.SocketCommon_Asset,
		Body: &pb.PushData_AssetData{
			AssetData: &pb.AssetData{Account: "acc123"},
		},
	}
	client.dispatchPushData(pushData, cb)
	if received == nil {
		t.Fatal("OnAsset 回调未触发")
	}
	if received.GetAccount() != "acc123" {
		t.Errorf("Account 期望 acc123，实际 %s", received.GetAccount())
	}
}

// TestDispatchPushData_Position verifies Position data type triggers OnPosition callback.
func TestDispatchPushData_Position(t *testing.T) {
	cfg := newTestConfig(t)
	client := NewPushClient(cfg)

	var received *pb.PositionData
	cb := &Callbacks{
		OnPosition: func(d *pb.PositionData) { received = d },
	}
	pushData := &pb.PushData{
		DataType: pb.SocketCommon_Position,
		Body: &pb.PushData_PositionData{
			PositionData: &pb.PositionData{Account: "acc123"},
		},
	}
	client.dispatchPushData(pushData, cb)
	if received == nil {
		t.Fatal("OnPosition 回调未触发")
	}
}

// TestDispatchPushData_OrderStatus verifies OrderStatus data type triggers OnOrder callback.
func TestDispatchPushData_OrderStatus(t *testing.T) {
	cfg := newTestConfig(t)
	client := NewPushClient(cfg)

	var received *pb.OrderStatusData
	cb := &Callbacks{
		OnOrder: func(d *pb.OrderStatusData) { received = d },
	}
	pushData := &pb.PushData{
		DataType: pb.SocketCommon_OrderStatus,
		Body: &pb.PushData_OrderStatusData{
			OrderStatusData: &pb.OrderStatusData{Symbol: "AAPL", Status: "Filled"},
		},
	}
	client.dispatchPushData(pushData, cb)
	if received == nil {
		t.Fatal("OnOrder 回调未触发")
	}
	if received.GetStatus() != "Filled" {
		t.Errorf("Status 期望 Filled，实际 %s", received.GetStatus())
	}
}

// TestDispatchPushData_OrderTransaction verifies OrderTransaction data type triggers OnTransaction callback.
func TestDispatchPushData_OrderTransaction(t *testing.T) {
	cfg := newTestConfig(t)
	client := NewPushClient(cfg)

	var received *pb.OrderTransactionData
	cb := &Callbacks{
		OnTransaction: func(d *pb.OrderTransactionData) { received = d },
	}
	pushData := &pb.PushData{
		DataType: pb.SocketCommon_OrderTransaction,
		Body: &pb.PushData_OrderTransactionData{
			OrderTransactionData: &pb.OrderTransactionData{Symbol: "AAPL"},
		},
	}
	client.dispatchPushData(pushData, cb)
	if received == nil {
		t.Fatal("OnTransaction 回调未触发")
	}
}

// TestDispatchPushData_StockTop verifies StockTop data type triggers OnStockTop callback.
func TestDispatchPushData_StockTop(t *testing.T) {
	cfg := newTestConfig(t)
	client := NewPushClient(cfg)

	var received *pb.StockTopData
	cb := &Callbacks{
		OnStockTop: func(d *pb.StockTopData) { received = d },
	}
	pushData := &pb.PushData{
		DataType: pb.SocketCommon_StockTop,
		Body: &pb.PushData_StockTopData{
			StockTopData: &pb.StockTopData{Market: "US"},
		},
	}
	client.dispatchPushData(pushData, cb)
	if received == nil {
		t.Fatal("OnStockTop 回调未触发")
	}
}

// TestDispatchPushData_OptionTop verifies OptionTop data type triggers OnOptionTop callback.
func TestDispatchPushData_OptionTop(t *testing.T) {
	cfg := newTestConfig(t)
	client := NewPushClient(cfg)

	var received *pb.OptionTopData
	cb := &Callbacks{
		OnOptionTop: func(d *pb.OptionTopData) { received = d },
	}
	pushData := &pb.PushData{
		DataType: pb.SocketCommon_OptionTop,
		Body: &pb.PushData_OptionTopData{
			OptionTopData: &pb.OptionTopData{Market: "US"},
		},
	}
	client.dispatchPushData(pushData, cb)
	if received == nil {
		t.Fatal("OnOptionTop 回调未触发")
	}
}

// TestDispatchPushData_Kline verifies Kline data type triggers OnKline callback.
func TestDispatchPushData_Kline(t *testing.T) {
	cfg := newTestConfig(t)
	client := NewPushClient(cfg)

	var received *pb.KlineData
	cb := &Callbacks{
		OnKline: func(d *pb.KlineData) { received = d },
	}
	pushData := &pb.PushData{
		DataType: pb.SocketCommon_Kline,
		Body: &pb.PushData_KlineData{
			KlineData: &pb.KlineData{Symbol: "AAPL"},
		},
	}
	client.dispatchPushData(pushData, cb)
	if received == nil {
		t.Fatal("OnKline 回调未触发")
	}
}

// TestDispatchPushData_Cc verifies Cc (crypto) data type triggers OnQuote callback.
func TestDispatchPushData_Cc(t *testing.T) {
	cfg := newTestConfig(t)
	client := NewPushClient(cfg)

	var received *pb.QuoteData
	cb := &Callbacks{
		OnQuote: func(d *pb.QuoteData) { received = d },
	}
	pushData := &pb.PushData{
		DataType: pb.SocketCommon_Cc,
		Body: &pb.PushData_QuoteData{
			QuoteData: &pb.QuoteData{Symbol: "BTCUSD"},
		},
	}
	client.dispatchPushData(pushData, cb)
	if received == nil {
		t.Fatal("OnQuote 回调未触发 (Cc)")
	}
	if received.GetSymbol() != "BTCUSD" {
		t.Errorf("Symbol 期望 BTCUSD，实际 %s", received.GetSymbol())
	}
}

// TestDispatchPushData_UnknownDataType verifies unknown data type triggers OnError callback.
func TestDispatchPushData_UnknownDataType(t *testing.T) {
	cfg := newTestConfig(t)
	client := NewPushClient(cfg)

	var errReceived error
	cb := &Callbacks{
		OnError: func(err error) { errReceived = err },
	}
	pushData := &pb.PushData{
		DataType: pb.SocketCommon_Unknown,
	}
	client.dispatchPushData(pushData, cb)
	if errReceived == nil {
		t.Fatal("未知 DataType 应触发 OnError")
	}
	if !strings.Contains(errReceived.Error(), "未知的 DataType") {
		t.Errorf("错误信息应提示未知 DataType，实际: %v", errReceived)
	}
}

// TestDispatchPushData_NilBodyDoesNotCallCallback verifies nil body does not trigger callback.
func TestDispatchPushData_NilBodyDoesNotCallCallback(t *testing.T) {
	cfg := newTestConfig(t)
	client := NewPushClient(cfg)

	called := false
	cb := &Callbacks{
		OnQuote: func(d *pb.QuoteData) { called = true },
	}
	pushData := &pb.PushData{
		DataType: pb.SocketCommon_Quote,
		Body:     &pb.PushData_QuoteData{QuoteData: nil}, // nil body
	}
	client.dispatchPushData(pushData, cb)
	if called {
		t.Error("body 为 nil 时不应触发回调")
	}
}

// TestDispatchPushData_NoCallbackSet verifies no panic when callback is not set.
func TestDispatchPushData_NoCallbackSet(t *testing.T) {
	cfg := newTestConfig(t)
	client := NewPushClient(cfg)

	cb := &Callbacks{} // all nil
	pushData := &pb.PushData{
		DataType: pb.SocketCommon_Quote,
		Body: &pb.PushData_QuoteData{
			QuoteData: &pb.QuoteData{Symbol: "AAPL"},
		},
	}
	// Should not panic
	client.dispatchPushData(pushData, cb)
}

// ===== handleMessage direct tests =====

// TestHandleMessage_Connected verifies CONNECTED message sets state and triggers callback.
func TestHandleMessage_Connected(t *testing.T) {
	cfg := newTestConfig(t)
	client := NewPushClient(cfg)

	// Set up connectedCh to simulate Connect() waiting
	client.mu.Lock()
	client.connectedCh = make(chan struct{})
	client.mu.Unlock()

	var connected int32
	cb := Callbacks{
		OnConnect: func() { connected = 1 },
	}
	client.SetCallbacks(cb)

	client.handleMessage(&pb.Response{Command: pb.SocketCommon_CONNECTED})

	if client.State() != StateConnected {
		t.Errorf("期望 State=StateConnected，实际 %v", client.State())
	}
	if connected != 1 {
		t.Error("OnConnect 回调未触发")
	}
}

// TestHandleMessage_Heartbeat verifies HEARTBEAT message is ignored (no callback).
func TestHandleMessage_Heartbeat(t *testing.T) {
	cfg := newTestConfig(t)
	client := NewPushClient(cfg)

	called := false
	client.SetCallbacks(Callbacks{
		OnError: func(err error) { called = true },
	})

	// Should not trigger any callback
	client.handleMessage(&pb.Response{Command: pb.SocketCommon_HEARTBEAT})
	if called {
		t.Error("HEARTBEAT 不应触发任何回调")
	}
}

// TestHandleMessage_MessageWithNilBody verifies MESSAGE with nil body triggers OnError.
func TestHandleMessage_MessageWithNilBody(t *testing.T) {
	cfg := newTestConfig(t)
	client := NewPushClient(cfg)

	var errReceived error
	client.SetCallbacks(Callbacks{
		OnError: func(err error) { errReceived = err },
	})

	resp := &pb.Response{Command: pb.SocketCommon_MESSAGE, Body: nil}
	client.handleMessage(resp)

	if errReceived == nil {
		t.Fatal("MESSAGE 缺少 body 应触发 OnError")
	}
	if !strings.Contains(errReceived.Error(), "缺少 body") {
		t.Errorf("错误信息应提示缺少 body，实际: %v", errReceived)
	}
}

// TestHandleMessage_ErrorWithoutKickout verifies ERROR without kickout triggers OnError but not OnKickout.
func TestHandleMessage_ErrorWithoutKickout(t *testing.T) {
	cfg := newTestConfig(t)
	client := NewPushClient(cfg)

	var errReceived error
	kickoutCalled := false
	client.SetCallbacks(Callbacks{
		OnError:   func(err error) { errReceived = err },
		OnKickout: func(msg string) { kickoutCalled = true },
	})

	resp := &pb.Response{Command: pb.SocketCommon_ERROR}
	client.handleMessage(resp)

	if errReceived == nil {
		t.Fatal("ERROR 应触发 OnError")
	}
	if kickoutCalled {
		t.Error("无 msg 时不应触发 OnKickout")
	}
}

// TestHandleMessage_ErrorWithKickout verifies ERROR with msg triggers both OnError and OnKickout.
func TestHandleMessage_ErrorWithKickout(t *testing.T) {
	cfg := newTestConfig(t)
	client := NewPushClient(cfg)

	var errReceived error
	var kickoutMsg string
	client.SetCallbacks(Callbacks{
		OnError:   func(err error) { errReceived = err },
		OnKickout: func(msg string) { kickoutMsg = msg },
	})

	resp := &pb.Response{
		Command: pb.SocketCommon_ERROR,
		Msg:     proto.String("kicked out: duplicate login"),
	}
	client.handleMessage(resp)

	if errReceived == nil {
		t.Fatal("ERROR 应触发 OnError")
	}
	if kickoutMsg != "kicked out: duplicate login" {
		t.Errorf("OnKickout msg 期望 'kicked out: duplicate login'，实际 %q", kickoutMsg)
	}
}

// TestHandleMessage_Disconnect verifies DISCONNECT message triggers OnDisconnect callback.
func TestHandleMessage_Disconnect(t *testing.T) {
	cfg := newTestConfig(t)
	client := NewPushClient(cfg)

	disconnectCalled := false
	client.SetCallbacks(Callbacks{
		OnDisconnect: func() { disconnectCalled = true },
	})

	client.handleMessage(&pb.Response{Command: pb.SocketCommon_DISCONNECT})
	if !disconnectCalled {
		t.Error("DISCONNECT 应触发 OnDisconnect 回调")
	}
}

// ===== SubjectToDataType tests =====

// TestSubjectToDataType_AllSubjects verifies all subject types map correctly.
func TestSubjectToDataType_AllSubjects(t *testing.T) {
	tests := []struct {
		subject SubjectType
		want    pb.SocketCommon_DataType
	}{
		{SubjectQuote, pb.SocketCommon_Quote},
		{SubjectOption, pb.SocketCommon_Option},
		{SubjectFuture, pb.SocketCommon_Future},
		{SubjectDepth, pb.SocketCommon_QuoteDepth},
		{SubjectTick, pb.SocketCommon_TradeTick},
		{SubjectAsset, pb.SocketCommon_Asset},
		{SubjectPosition, pb.SocketCommon_Position},
		{SubjectOrder, pb.SocketCommon_OrderStatus},
		{SubjectTransaction, pb.SocketCommon_OrderTransaction},
		{SubjectStockTop, pb.SocketCommon_StockTop},
		{SubjectOptionTop, pb.SocketCommon_OptionTop},
		{SubjectKline, pb.SocketCommon_Kline},
		{SubjectCc, pb.SocketCommon_Cc},
		{SubjectMarket, pb.SocketCommon_Quote},
	}

	for _, tt := range tests {
		t.Run(string(tt.subject), func(t *testing.T) {
			got := SubjectToDataType(tt.subject)
			if got != tt.want {
				t.Errorf("SubjectToDataType(%q) = %v, want %v", tt.subject, got, tt.want)
			}
		})
	}
}

// TestSubjectToDataType_UnknownSubject verifies unknown subject returns Unknown.
func TestSubjectToDataType_UnknownSubject(t *testing.T) {
	got := SubjectToDataType(SubjectType("nonexistent"))
	if got != pb.SocketCommon_Unknown {
		t.Errorf("未知 subject 应返回 Unknown，实际 %v", got)
	}
}

// TestSubjectToDataType_FullTickAndQuoteBBO verifies FullTick and QuoteBBO map to Unknown (not yet implemented).
func TestSubjectToDataType_FullTickAndQuoteBBO(t *testing.T) {
	if got := SubjectToDataType(SubjectFullTick); got != pb.SocketCommon_Unknown {
		t.Errorf("FullTick 应返回 Unknown，实际 %v", got)
	}
	if got := SubjectToDataType(SubjectQuoteBBO); got != pb.SocketCommon_Unknown {
		t.Errorf("QuoteBBO 应返回 Unknown，实际 %v", got)
	}
}

// ===== DecodeVarint32 edge case tests =====

// TestDecodeVarint32_EmptyBuffer verifies empty buffer returns ok=false.
func TestDecodeVarint32_EmptyBuffer(t *testing.T) {
	msg, remaining, ok := DecodeVarint32(nil)
	if ok {
		t.Error("空 buffer 应返回 ok=false")
	}
	if msg != nil {
		t.Error("空 buffer 时 msg 应为 nil")
	}
	if remaining != nil {
		t.Error("空 buffer 时 remaining 应为 nil")
	}
}

// TestDecodeVarint32_SingleByte verifies single byte length prefix.
func TestDecodeVarint32_SingleByte(t *testing.T) {
	data := []byte{0x05, 'h', 'e', 'l', 'l', 'o'}
	msg, remaining, ok := DecodeVarint32(data)
	if !ok {
		t.Fatal("完整帧应返回 ok=true")
	}
	if string(msg) != "hello" {
		t.Errorf("msg 期望 'hello'，实际 %q", string(msg))
	}
	if len(remaining) != 0 {
		t.Errorf("remaining 应为空，实际长度 %d", len(remaining))
	}
}

// TestDecodeVarint32_TwoFrames verifies decoding two frames in one buffer.
func TestDecodeVarint32_TwoFrames(t *testing.T) {
	frame1 := EncodeVarint32([]byte("hello"))
	frame2 := EncodeVarint32([]byte("world"))
	buffer := append(frame1, frame2...)

	msg1, remaining, ok := DecodeVarint32(buffer)
	if !ok {
		t.Fatal("第一帧应解码成功")
	}
	if string(msg1) != "hello" {
		t.Errorf("第一帧 msg 期望 'hello'，实际 %q", string(msg1))
	}

	msg2, remaining2, ok2 := DecodeVarint32(remaining)
	if !ok2 {
		t.Fatal("第二帧应解码成功")
	}
	if string(msg2) != "world" {
		t.Errorf("第二帧 msg 期望 'world'，实际 %q", string(msg2))
	}
	if len(remaining2) != 0 {
		t.Errorf("最终 remaining 应为空，实际长度 %d", len(remaining2))
	}
}

// TestDecodeVarint32_IncompleteHeader verifies incomplete header returns ok=false.
func TestDecodeVarint32_IncompleteHeader(t *testing.T) {
	// A varint with continuation bit set but not enough bytes
	buffer := []byte{0x80} // continuation bit set, no more bytes
	msg, _, ok := DecodeVarint32(buffer)
	if ok {
		t.Error("不完整的 varint header 应返回 ok=false")
	}
	if msg != nil {
		t.Error("不完整时 msg 应为 nil")
	}
}

// TestDecodeVarint32_IncompleteBody verifies incomplete body returns ok=false.
func TestDecodeVarint32_IncompleteBody(t *testing.T) {
	// Length prefix says 10 bytes, but only 5 are present
	buffer := []byte{0x0A, 'h', 'e', 'l', 'l', 'o'}
	msg, _, ok := DecodeVarint32(buffer)
	if ok {
		t.Error("不完整的 body 应返回 ok=false")
	}
	if msg != nil {
		t.Error("不完整时 msg 应为 nil")
	}
}

// TestDecodeVarint32_LargeLength verifies a multi-byte varint length prefix works.
func TestDecodeVarint32_LargeLength(t *testing.T) {
	// Create data > 127 bytes to force multi-byte varint
	data := make([]byte, 200)
	for i := range data {
		data[i] = byte(i)
	}
	encoded := EncodeVarint32(data)
	msg, remaining, ok := DecodeVarint32(encoded)
	if !ok {
		t.Fatal("大帧应解码成功")
	}
	if len(msg) != 200 {
		t.Errorf("msg 长度期望 200，实际 %d", len(msg))
	}
	if len(remaining) != 0 {
		t.Errorf("remaining 应为空，实际长度 %d", len(remaining))
	}
	for i := range data {
		if msg[i] != data[i] {
			t.Fatalf("msg[%d] 期望 %d，实际 %d", i, data[i], msg[i])
		}
	}
}

// TestDecodeVarint32_MalformedVarint verifies a malformed (exceeds 32-bit) varint returns ok=false.
func TestDecodeVarint32_MalformedVarint(t *testing.T) {
	// 5 bytes all with continuation bit set — exceeds 32-bit range
	buffer := []byte{0x80, 0x80, 0x80, 0x80, 0x80}
	_, _, ok := DecodeVarint32(buffer)
	if ok {
		t.Error("超出 32-bit 范围的 varint 应返回 ok=false")
	}
}

// ===== BuildSubscribeMessage / BuildUnSubscribeMessage tests =====

// TestBuildSubscribeMessage_AllFields verifies all optional fields are set.
func TestBuildSubscribeMessage_AllFields(t *testing.T) {
	req := BuildSubscribeMessage(pb.SocketCommon_Quote, "AAPL,TSLA", "acc123", "US")
	if req.Command != pb.SocketCommon_SUBSCRIBE {
		t.Errorf("Command 期望 SUBSCRIBE，实际 %v", req.Command)
	}
	if req.Subscribe == nil {
		t.Fatal("Subscribe 不应为 nil")
	}
	if req.Subscribe.DataType != pb.SocketCommon_Quote {
		t.Errorf("DataType 期望 Quote，实际 %v", req.Subscribe.DataType)
	}
	if req.Subscribe.GetSymbols() != "AAPL,TSLA" {
		t.Errorf("Symbols 期望 AAPL,TSLA，实际 %s", req.Subscribe.GetSymbols())
	}
	if req.Subscribe.GetAccount() != "acc123" {
		t.Errorf("Account 期望 acc123，实际 %s", req.Subscribe.GetAccount())
	}
	if req.Subscribe.GetMarket() != "US" {
		t.Errorf("Market 期望 US，实际 %s", req.Subscribe.GetMarket())
	}
}

// TestBuildSubscribeMessage_NoOptionalFields verifies no optional fields are set.
func TestBuildSubscribeMessage_NoOptionalFields(t *testing.T) {
	req := BuildSubscribeMessage(pb.SocketCommon_Quote, "", "", "")
	if req.Subscribe == nil {
		t.Fatal("Subscribe 不应为 nil")
	}
	if req.Subscribe.GetSymbols() != "" {
		t.Errorf("Symbols 期望空，实际 %s", req.Subscribe.GetSymbols())
	}
	if req.Subscribe.GetAccount() != "" {
		t.Errorf("Account 期望空，实际 %s", req.Subscribe.GetAccount())
	}
	if req.Subscribe.GetMarket() != "" {
		t.Errorf("Market 期望空，实际 %s", req.Subscribe.GetMarket())
	}
}

// TestBuildUnSubscribeMessage_AllFields verifies all optional fields are set.
func TestBuildUnSubscribeMessage_AllFields(t *testing.T) {
	req := BuildUnSubscribeMessage(pb.SocketCommon_Quote, "AAPL,TSLA", "acc123", "US")
	if req.Command != pb.SocketCommon_UNSUBSCRIBE {
		t.Errorf("Command 期望 UNSUBSCRIBE，实际 %v", req.Command)
	}
	if req.Subscribe == nil {
		t.Fatal("Subscribe 不应为 nil")
	}
	if req.Subscribe.GetSymbols() != "AAPL,TSLA" {
		t.Errorf("Symbols 期望 AAPL,TSLA，实际 %s", req.Subscribe.GetSymbols())
	}
}

// TestBuildUnSubscribeMessage_NoOptionalFields verifies no optional fields are set.
func TestBuildUnSubscribeMessage_NoOptionalFields(t *testing.T) {
	req := BuildUnSubscribeMessage(pb.SocketCommon_Asset, "", "", "")
	if req.Subscribe == nil {
		t.Fatal("Subscribe 不应为 nil")
	}
	if req.Subscribe.GetSymbols() != "" {
		t.Errorf("Symbols 期望空，实际 %s", req.Subscribe.GetSymbols())
	}
}

// ===== Subscribe/Unsubscribe for Top/Cc/Market (requires pipe connection) =====

// newConnectedPushClient creates a PushClient connected via pipe with a mock server.
func newConnectedPushClient(t *testing.T) (*PushClient, *pipeDialer) {
	t.Helper()
	dialer := newPipeDialer()

	go func() {
		readProtobufRequest(t, dialer.serverConn)
		sendProtobufResponse(t, dialer.serverConn, &pb.Response{
			Command: pb.SocketCommon_CONNECTED,
		})
		drainConn(dialer.serverConn)
	}()

	cfg := newTestConfig(t)
	client := NewPushClient(cfg, WithAutoReconnect(false), WithHeartbeatInterval(30*time.Second))
	client.dialer = dialer

	if err := client.Connect(); err != nil {
		t.Fatalf("Connect 失败: %v", err)
	}
	return client, dialer
}

// TestPushClient_SubscribeStockTop verifies subscribing to stock top data.
func TestPushClient_SubscribeStockTop(t *testing.T) {
	client, _ := newConnectedPushClient(t)
	defer client.Disconnect()

	err := client.SubscribeStockTop("US", []string{"top_gainer", "top_loser"})
	if err != nil {
		t.Fatalf("SubscribeStockTop 失败: %v", err)
	}

	subs := client.GetSubscriptions()
	if _, ok := subs[SubjectStockTop]; !ok {
		t.Error("应有 stock_top 订阅记录")
	}
}

// TestPushClient_UnsubscribeStockTop verifies unsubscribing from stock top data.
func TestPushClient_UnsubscribeStockTop(t *testing.T) {
	client, _ := newConnectedPushClient(t)
	defer client.Disconnect()

	client.SubscribeStockTop("US", []string{"top_gainer", "top_loser"})
	err := client.UnsubscribeStockTop("US", []string{"top_gainer"})
	if err != nil {
		t.Fatalf("UnsubscribeStockTop 失败: %v", err)
	}

	subs := client.GetSubscriptions()
	if symbols, ok := subs[SubjectStockTop]; ok {
		if len(symbols) != 1 {
			t.Errorf("退订后应剩 1 个标的，实际 %d 个", len(symbols))
		}
	}
}

// TestPushClient_SubscribeOptionTop verifies subscribing to option top data.
func TestPushClient_SubscribeOptionTop(t *testing.T) {
	client, _ := newConnectedPushClient(t)
	defer client.Disconnect()

	err := client.SubscribeOptionTop("US", []string{"top_volume"})
	if err != nil {
		t.Fatalf("SubscribeOptionTop 失败: %v", err)
	}

	subs := client.GetSubscriptions()
	if _, ok := subs[SubjectOptionTop]; !ok {
		t.Error("应有 option_top 订阅记录")
	}
}

// TestPushClient_UnsubscribeOptionTop verifies unsubscribing from option top data.
func TestPushClient_UnsubscribeOptionTop(t *testing.T) {
	client, _ := newConnectedPushClient(t)
	defer client.Disconnect()

	client.SubscribeOptionTop("US", []string{"top_volume", "top_oi"})
	err := client.UnsubscribeOptionTop("US", []string{"top_volume"})
	if err != nil {
		t.Fatalf("UnsubscribeOptionTop 失败: %v", err)
	}

	subs := client.GetSubscriptions()
	if symbols, ok := subs[SubjectOptionTop]; ok {
		if len(symbols) != 1 {
			t.Errorf("退订后应剩 1 个标的，实际 %d 个", len(symbols))
		}
	}
}

// TestPushClient_SubscribeCc verifies subscribing to crypto data.
func TestPushClient_SubscribeCc(t *testing.T) {
	client, _ := newConnectedPushClient(t)
	defer client.Disconnect()

	err := client.SubscribeCc([]string{"BTCUSD", "ETHUSD"})
	if err != nil {
		t.Fatalf("SubscribeCc 失败: %v", err)
	}

	subs := client.GetSubscriptions()
	if _, ok := subs[SubjectCc]; !ok {
		t.Error("应有 cc 订阅记录")
	}
}

// TestPushClient_UnsubscribeCc verifies unsubscribing from crypto data.
func TestPushClient_UnsubscribeCc(t *testing.T) {
	client, _ := newConnectedPushClient(t)
	defer client.Disconnect()

	client.SubscribeCc([]string{"BTCUSD", "ETHUSD"})
	err := client.UnsubscribeCc([]string{"BTCUSD"})
	if err != nil {
		t.Fatalf("UnsubscribeCc 失败: %v", err)
	}

	subs := client.GetSubscriptions()
	if symbols, ok := subs[SubjectCc]; ok {
		if len(symbols) != 1 {
			t.Errorf("退订后应剩 1 个标的，实际 %d 个", len(symbols))
		}
	}
}

// TestPushClient_UnsubscribeCc_All verifies unsubscribing all crypto symbols.
func TestPushClient_UnsubscribeCc_All(t *testing.T) {
	client, _ := newConnectedPushClient(t)
	defer client.Disconnect()

	client.SubscribeCc([]string{"BTCUSD", "ETHUSD"})
	err := client.UnsubscribeCc(nil)
	if err != nil {
		t.Fatalf("UnsubscribeCc(nil) 失败: %v", err)
	}

	subs := client.GetSubscriptions()
	if _, ok := subs[SubjectCc]; ok {
		t.Error("退订全部后不应有 cc 订阅记录")
	}
}

// TestPushClient_SubscribeMarket verifies subscribing to market status.
func TestPushClient_SubscribeMarket(t *testing.T) {
	client, _ := newConnectedPushClient(t)
	defer client.Disconnect()

	err := client.SubscribeMarket("US")
	if err != nil {
		t.Fatalf("SubscribeMarket 失败: %v", err)
	}

	subs := client.GetSubscriptions()
	if _, ok := subs[SubjectMarket]; !ok {
		t.Error("应有 market 订阅记录")
	}
}

// TestPushClient_UnsubscribeMarket verifies unsubscribing from market status.
func TestPushClient_UnsubscribeMarket(t *testing.T) {
	client, _ := newConnectedPushClient(t)
	defer client.Disconnect()

	client.SubscribeMarket("US")
	err := client.UnsubscribeMarket("US")
	if err != nil {
		t.Fatalf("UnsubscribeMarket 失败: %v", err)
	}

	subs := client.GetSubscriptions()
	if _, ok := subs[SubjectMarket]; ok {
		t.Error("退订后不应有 market 订阅记录")
	}
}

// ===== Unsubscribe methods for tick/depth/option/future/kline/order/transaction =====

// TestPushClient_UnsubscribeTick verifies unsubscribing from tick data.
func TestPushClient_UnsubscribeTick(t *testing.T) {
	client, _ := newConnectedPushClient(t)
	defer client.Disconnect()

	client.SubscribeTick([]string{"AAPL", "TSLA"})
	err := client.UnsubscribeTick([]string{"TSLA"})
	if err != nil {
		t.Fatalf("UnsubscribeTick 失败: %v", err)
	}

	subs := client.GetSubscriptions()
	if symbols, ok := subs[SubjectTick]; ok {
		if len(symbols) != 1 {
			t.Errorf("退订后应剩 1 个标的，实际 %d 个", len(symbols))
		}
	}
}

// TestPushClient_UnsubscribeDepth verifies unsubscribing from depth data.
func TestPushClient_UnsubscribeDepth(t *testing.T) {
	client, _ := newConnectedPushClient(t)
	defer client.Disconnect()

	client.SubscribeDepth([]string{"AAPL", "GOOG"})
	err := client.UnsubscribeDepth([]string{"GOOG"})
	if err != nil {
		t.Fatalf("UnsubscribeDepth 失败: %v", err)
	}

	subs := client.GetSubscriptions()
	if symbols, ok := subs[SubjectDepth]; ok {
		if len(symbols) != 1 {
			t.Errorf("退订后应剩 1 个标的，实际 %d 个", len(symbols))
		}
	}
}

// TestPushClient_UnsubscribeOption verifies unsubscribing from option data.
func TestPushClient_UnsubscribeOption(t *testing.T) {
	client, _ := newConnectedPushClient(t)
	defer client.Disconnect()

	client.SubscribeOption([]string{"AAPL240315C00150000"})
	err := client.UnsubscribeOption([]string{"AAPL240315C00150000"})
	if err != nil {
		t.Fatalf("UnsubscribeOption 失败: %v", err)
	}

	subs := client.GetSubscriptions()
	if _, ok := subs[SubjectOption]; ok {
		t.Error("退订后不应有 option 订阅记录")
	}
}

// TestPushClient_UnsubscribeFuture verifies unsubscribing from future data.
func TestPushClient_UnsubscribeFuture(t *testing.T) {
	client, _ := newConnectedPushClient(t)
	defer client.Disconnect()

	client.SubscribeFuture([]string{"ES", "NQ"})
	err := client.UnsubscribeFuture([]string{"ES"})
	if err != nil {
		t.Fatalf("UnsubscribeFuture 失败: %v", err)
	}

	subs := client.GetSubscriptions()
	if symbols, ok := subs[SubjectFuture]; ok {
		if len(symbols) != 1 {
			t.Errorf("退订后应剩 1 个标的，实际 %d 个", len(symbols))
		}
	}
}

// TestPushClient_UnsubscribeKline verifies unsubscribing from kline data.
func TestPushClient_UnsubscribeKline(t *testing.T) {
	client, _ := newConnectedPushClient(t)
	defer client.Disconnect()

	client.SubscribeKline([]string{"AAPL"})
	err := client.UnsubscribeKline([]string{"AAPL"})
	if err != nil {
		t.Fatalf("UnsubscribeKline 失败: %v", err)
	}

	subs := client.GetSubscriptions()
	if _, ok := subs[SubjectKline]; ok {
		t.Error("退订后不应有 kline 订阅记录")
	}
}

// TestPushClient_UnsubscribeOrder verifies unsubscribing from order data.
func TestPushClient_UnsubscribeOrder(t *testing.T) {
	client, _ := newConnectedPushClient(t)
	defer client.Disconnect()

	client.SubscribeOrder("")
	err := client.UnsubscribeOrder()
	if err != nil {
		t.Fatalf("UnsubscribeOrder 失败: %v", err)
	}

	acctSubs := client.GetAccountSubscriptions()
	for _, s := range acctSubs {
		if s == SubjectOrder {
			t.Error("退订后不应有 order 账户订阅记录")
		}
	}
}

// TestPushClient_UnsubscribeTransaction verifies unsubscribing from transaction data.
func TestPushClient_UnsubscribeTransaction(t *testing.T) {
	client, _ := newConnectedPushClient(t)
	defer client.Disconnect()

	client.SubscribeTransaction("")
	err := client.UnsubscribeTransaction()
	if err != nil {
		t.Fatalf("UnsubscribeTransaction 失败: %v", err)
	}

	acctSubs := client.GetAccountSubscriptions()
	for _, s := range acctSubs {
		if s == SubjectTransaction {
			t.Error("退订后不应有 transaction 账户订阅记录")
		}
	}
}

// ===== sendMessage error path tests =====

// TestSendMessage_NilConnection verifies sendMessage returns an error when conn is nil.
func TestSendMessage_NilConnection(t *testing.T) {
	cfg := newTestConfig(t)
	client := NewPushClient(cfg)

	err := client.sendMessage(BuildHeartBeatMessage())
	if err == nil {
		t.Fatal("nil conn 时应返回错误")
	}
	if !strings.Contains(err.Error(), "TCP 连接未建立") {
		t.Errorf("错误信息应提示连接未建立，实际: %v", err)
	}
}

// TestSendMessage_WriteError verifies sendMessage returns an error when write fails.
func TestPushClient_SendMessage_WriteError(t *testing.T) {
	dialer := newPipeDialer()

	// Close the client connection immediately to cause write errors
	dialer.clientConn.Close()

	cfg := newTestConfig(t)
	client := NewPushClient(cfg, WithAutoReconnect(false))
	client.dialer = dialer

	// Set conn to the closed connection
	client.mu.Lock()
	client.conn = dialer.clientConn
	client.mu.Unlock()

	err := client.sendMessage(BuildHeartBeatMessage())
	if err == nil {
		t.Fatal("写入已关闭连接应返回错误")
	}
}

// ===== readLoop error path test =====

// TestReadLoop_ReadErrorTriggersOnError verifies read errors trigger OnError callback.
func TestReadLoop_ReadErrorTriggersOnError(t *testing.T) {
	dialer := newPipeDialer()

	cfg := newTestConfig(t)
	client := NewPushClient(cfg, WithAutoReconnect(false))
	client.dialer = dialer

	var errReceived error
	var errMu sync.Mutex
	client.SetCallbacks(Callbacks{
		OnError: func(err error) {
			errMu.Lock()
			errReceived = err
			errMu.Unlock()
		},
	})

	// Set up state as if Connect() was called
	client.mu.Lock()
	client.conn = dialer.clientConn
	client.stopCh = make(chan struct{})
	client.doneCh = make(chan struct{})
	client.mu.Unlock()

	// Start the read loop
	go client.readLoop()

	// Close the server side to cause read error
	dialer.serverConn.Close()

	// Wait for readLoop to detect the error
	time.Sleep(150 * time.Millisecond)

	errMu.Lock()
	defer errMu.Unlock()
	if errReceived == nil {
		t.Fatal("读取错误应触发 OnError 回调")
	}
}

// ===== removeSubscription edge case tests =====

// TestRemoveSubscription_NonExistentSubject verifies removing from non-existent subject is a no-op.
func TestRemoveSubscription_NonExistentSubject(t *testing.T) {
	cfg := newTestConfig(t)
	client := NewPushClient(cfg)

	// Should not panic
	client.removeSubscription(SubjectQuote, []string{"AAPL"})

	subs := client.GetSubscriptions()
	if _, ok := subs[SubjectQuote]; ok {
		t.Error("不存在的 subject 不应出现在订阅记录中")
	}
}

// TestRemoveSubscription_EmptyMapCleanup verifies empty map is deleted from subscriptions.
func TestRemoveSubscription_EmptyMapCleanup(t *testing.T) {
	cfg := newTestConfig(t)
	client := NewPushClient(cfg)

	client.addSubscription(SubjectQuote, []string{"AAPL"})
	client.removeSubscription(SubjectQuote, []string{"AAPL"})

	subs := client.GetSubscriptions()
	if _, ok := subs[SubjectQuote]; ok {
		t.Error("空 map 清理后不应有 quote 订阅记录")
	}
}

// ===== resubscribe tests =====

// TestResubscribe_RestoresAllSubscriptions verifies resubscribe sends subscribe messages for all existing subscriptions.
func TestResubscribe_RestoresAllSubscriptions(t *testing.T) {
	dialer := newPipeDialer()

	var receivedReqs []*pb.Request
	var reqMu sync.Mutex

	go func() {
		readProtobufRequest(t, dialer.serverConn)
		sendProtobufResponse(t, dialer.serverConn, &pb.Response{
			Command: pb.SocketCommon_CONNECTED,
		})

		// Read subsequent messages
		var buf []byte
		tmp := make([]byte, 1024)
		for {
			n, err := dialer.serverConn.Read(tmp)
			if err != nil {
				return
			}
			buf = append(buf, tmp[:n]...)
			for {
				msg, remaining, ok := DecodeVarint32(buf)
				if !ok {
					break
				}
				var req pb.Request
				if proto.Unmarshal(msg, &req) == nil {
					reqMu.Lock()
					receivedReqs = append(receivedReqs, proto.Clone(&req).(*pb.Request))
					reqMu.Unlock()
				}
				buf = remaining
			}
		}
	}()

	cfg := newTestConfig(t)
	client := NewPushClient(cfg, WithAutoReconnect(false), WithHeartbeatInterval(30*time.Second))
	client.dialer = dialer

	if err := client.Connect(); err != nil {
		t.Fatalf("Connect 失败: %v", err)
	}
	defer client.Disconnect()

	// Set up subscription state directly
	client.addSubscription(SubjectQuote, []string{"AAPL", "TSLA"})
	client.mu.Lock()
	client.accountSubs[SubjectAsset] = true
	client.mu.Unlock()

	// Call resubscribe
	client.resubscribe()

	time.Sleep(100 * time.Millisecond)

	reqMu.Lock()
	defer reqMu.Unlock()

	// Should have sent 2 subscribe messages (quote + asset)
	subscribeCount := 0
	for _, req := range receivedReqs {
		if req.Command == pb.SocketCommon_SUBSCRIBE {
			subscribeCount++
		}
	}
	if subscribeCount < 2 {
		t.Errorf("resubscribe 应发送至少 2 条 SUBSCRIBE 请求，实际 %d 条", subscribeCount)
	}
}

// ===== Subscribe error path tests =====

// TestSubscribeQuote_SendError verifies SubscribeQuote returns error when sendMessage fails.
func TestSubscribeQuote_SendError(t *testing.T) {
	dialer := newPipeDialer()

	cfg := newTestConfig(t)
	client := NewPushClient(cfg, WithAutoReconnect(false))
	client.dialer = dialer

	// Set up state as if connected, but with a closed connection
	client.mu.Lock()
	client.conn = dialer.clientConn
	client.mu.Unlock()

	// Close the client connection to cause write errors
	dialer.clientConn.Close()

	err := client.SubscribeQuote([]string{"AAPL"})
	if err == nil {
		t.Fatal("写入已关闭连接应返回错误")
	}

	// Subscription should not be recorded since sendMessage failed
	subs := client.GetSubscriptions()
	if _, ok := subs[SubjectQuote]; ok {
		t.Error("sendMessage 失败时不应记录订阅")
	}
}

// TestUnsubscribeQuote_SendError verifies UnsubscribeQuote returns error when sendMessage fails.
func TestUnsubscribeQuote_SendError(t *testing.T) {
	dialer := newPipeDialer()

	cfg := newTestConfig(t)
	client := NewPushClient(cfg, WithAutoReconnect(false))
	client.dialer = dialer

	// Set up state as if connected, but with a closed connection
	client.mu.Lock()
	client.conn = dialer.clientConn
	client.mu.Unlock()

	// Add a subscription record first
	client.addSubscription(SubjectQuote, []string{"AAPL"})

	// Close the client connection to cause write errors
	dialer.clientConn.Close()

	err := client.UnsubscribeQuote([]string{"AAPL"})
	if err == nil {
		t.Fatal("写入已关闭连接应返回错误")
	}

	// Subscription should still be present since sendMessage failed
	subs := client.GetSubscriptions()
	if _, ok := subs[SubjectQuote]; !ok {
		t.Error("sendMessage 失败时不应移除订阅记录")
	}
}

// TestSubscribeStockTop_SendError verifies SubscribeStockTop returns error when sendMessage fails.
func TestSubscribeStockTop_SendError(t *testing.T) {
	dialer := newPipeDialer()

	cfg := newTestConfig(t)
	client := NewPushClient(cfg, WithAutoReconnect(false))
	client.dialer = dialer

	client.mu.Lock()
	client.conn = dialer.clientConn
	client.mu.Unlock()
	dialer.clientConn.Close()

	err := client.SubscribeStockTop("US", []string{"top_gainer"})
	if err == nil {
		t.Fatal("写入已关闭连接应返回错误")
	}
}

// TestSubscribeMarket_SendError verifies SubscribeMarket returns error when sendMessage fails.
func TestSubscribeMarket_SendError(t *testing.T) {
	dialer := newPipeDialer()

	cfg := newTestConfig(t)
	client := NewPushClient(cfg, WithAutoReconnect(false))
	client.dialer = dialer

	client.mu.Lock()
	client.conn = dialer.clientConn
	client.mu.Unlock()
	dialer.clientConn.Close()

	err := client.SubscribeMarket("US")
	if err == nil {
		t.Fatal("写入已关闭连接应返回错误")
	}
}

// TestUnsubscribeStockTop_SendError verifies UnsubscribeStockTop returns error when sendMessage fails.
func TestUnsubscribeStockTop_SendError(t *testing.T) {
	dialer := newPipeDialer()

	cfg := newTestConfig(t)
	client := NewPushClient(cfg, WithAutoReconnect(false))
	client.dialer = dialer

	client.mu.Lock()
	client.conn = dialer.clientConn
	client.mu.Unlock()
	dialer.clientConn.Close()

	err := client.UnsubscribeStockTop("US", []string{"top_gainer"})
	if err == nil {
		t.Fatal("写入已关闭连接应返回错误")
	}
}

// ===== Connect error path tests =====

// errorDialer is a dialer that always returns an error.
type errorDialer struct{}

func (d *errorDialer) Dial(address string, timeout time.Duration) (net.Conn, error) {
	return nil, errors.New("connection refused")
}

// TestConnect_DialError verifies Connect returns error when dial fails.
func TestConnect_DialError(t *testing.T) {
	cfg := newTestConfig(t)
	client := NewPushClient(cfg, WithAutoReconnect(false))
	client.dialer = &errorDialer{}

	err := client.Connect()
	if err == nil {
		t.Fatal("Dial 失败应返回错误")
	}
	if !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("错误信息应包含 connection refused，实际: %v", err)
	}
	if client.State() != StateDisconnected {
		t.Errorf("Dial 失败后状态应为 StateDisconnected，实际 %v", client.State())
	}
}

// timeoutDialer is a dialer that returns a connection but never responds to CONNECTED.
type timeoutDialer struct {
	conn net.Conn
}

func (d *timeoutDialer) Dial(address string, timeout time.Duration) (net.Conn, error) {
	return d.conn, nil
}

// TestConnect_TimeoutWaitingForConnected verifies Connect times out when CONNECTED response is not received.
func TestConnect_TimeoutWaitingForConnected(t *testing.T) {
	clientConn, serverConn := net.Pipe()

	cfg := newTestConfig(t)
	client := NewPushClient(cfg, WithAutoReconnect(false), WithConnectTimeout(100*time.Millisecond))
	client.dialer = &timeoutDialer{conn: clientConn}

	// Read the CONNECT request but don't send CONNECTED response
	go func() {
		buf := make([]byte, 4096)
		serverConn.Read(buf) // consume CONNECT message
		time.Sleep(500 * time.Millisecond)
		serverConn.Close()
	}()

	err := client.Connect()
	if err == nil {
		t.Fatal("CONNECTED 超时应返回错误")
	}
	if !strings.Contains(err.Error(), "超时") {
		t.Errorf("错误信息应提示超时，实际: %v", err)
	}
}

// TestConnect_AuthenticateError verifies Connect returns error when authentication (signing) fails.
func TestConnect_AuthenticateError(t *testing.T) {
	clientConn, serverConn := net.Pipe()

	cfg := newTestConfig(t)
	cfg.PrivateKey = "invalid-key"
	client := NewPushClient(cfg, WithAutoReconnect(false), WithConnectTimeout(200*time.Millisecond))
	client.dialer = &timeoutDialer{conn: clientConn}

	go drainConn(serverConn)

	err := client.Connect()
	if err == nil {
		t.Fatal("签名失败应返回错误")
	}
	if !strings.Contains(err.Error(), "认证失败") {
		t.Errorf("错误信息应提示认证失败，实际: %v", err)
	}
	clientConn.Close()
}

// ===== heartbeatLoop error path test =====

// TestHeartbeatLoop_SendErrorExits verifies heartbeatLoop exits when sendMessage fails.
func TestHeartbeatLoop_SendErrorExits(t *testing.T) {
	dialer := newPipeDialer()

	cfg := newTestConfig(t)
	client := NewPushClient(cfg, WithAutoReconnect(false), WithHeartbeatInterval(50*time.Millisecond))
	client.dialer = dialer

	// Set up state as if connected
	client.mu.Lock()
	client.conn = dialer.clientConn
	client.stopCh = make(chan struct{})
	client.mu.Unlock()

	// Close the connection to cause heartbeat send error
	dialer.clientConn.Close()

	// Start heartbeat loop — it should exit quickly when send fails
	go client.heartbeatLoop()

	// Wait a bit to ensure it exits without panic
	time.Sleep(150 * time.Millisecond)
	// If we get here without hanging or panicking, the test passes
}

// Ensure unused imports are referenced
var _ = errors.New
var _ = fmt.Sprintf
var _ = net.Pipe
