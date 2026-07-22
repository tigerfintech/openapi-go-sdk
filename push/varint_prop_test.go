package push

import (
	"testing"

	"github.com/tigerfintech/openapi-go-sdk/push/pb"
	"google.golang.org/protobuf/proto"
	"pgregory.net/rapid"
)

// Property 1: Varint32 编码/解码往返一致性
// For any byte slice data, EncodeVarint32(data) then DecodeVarint32 should return the original data.
// **Validates: Requirements 2.1, 2.2**
func TestPropVarint32RoundTrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		data := rapid.SliceOf(rapid.Byte()).Draw(t, "data")

		encoded := EncodeVarint32(data)
		msg, remaining, ok := DecodeVarint32(encoded)

		if !ok {
			t.Fatal("DecodeVarint32 returned ok=false for a complete encoded frame")
		}
		if len(remaining) != 0 {
			t.Fatalf("expected no remaining bytes, got %d", len(remaining))
		}
		if len(data) == 0 && len(msg) == 0 {
			// Both nil/empty slices are equivalent for empty input
			return
		}
		if len(msg) != len(data) {
			t.Fatalf("decoded length %d != original length %d", len(msg), len(data))
		}
		for i := range data {
			if msg[i] != data[i] {
				t.Fatalf("decoded byte at index %d differs: got %d, want %d", i, msg[i], data[i])
			}
		}
	})
}

// Property 2: Varint32 分块解码正确性
// For any byte slice data, encode it, split the encoded bytes at a random position,
// feed chunks to decoder — first chunk may return ok=false, but after feeding all data,
// should decode correctly.
// **Validates: Requirements 2.4**
func TestPropVarint32ChunkedDecode(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		data := rapid.SliceOf(rapid.Byte()).Draw(t, "data")

		encoded := EncodeVarint32(data)
		if len(encoded) == 0 {
			t.Skip("empty encoded frame")
		}

		// Pick a random split point in [0, len(encoded)]
		splitPoint := rapid.IntRange(0, len(encoded)).Draw(t, "splitPoint")

		chunk1 := encoded[:splitPoint]
		chunk2 := encoded[splitPoint:]

		// Feed first chunk — may or may not decode
		var buffer []byte
		buffer = append(buffer, chunk1...)
		msg, remaining, ok := DecodeVarint32(buffer)

		if ok {
			// If first chunk was enough, verify correctness
			if len(data) == 0 && len(msg) == 0 {
				return
			}
			if len(msg) != len(data) {
				t.Fatalf("decoded length %d != original length %d (first chunk)", len(msg), len(data))
			}
			for i := range data {
				if msg[i] != data[i] {
					t.Fatalf("decoded byte at index %d differs (first chunk)", i)
				}
			}
			// remaining + chunk2 should be empty
			finalBuf := append(remaining, chunk2...)
			if len(finalBuf) != 0 {
				t.Fatalf("unexpected extra bytes after decode: %d", len(finalBuf))
			}
			return
		}

		// First chunk wasn't enough, feed the rest
		buffer = append(buffer, chunk2...)
		msg, remaining, ok = DecodeVarint32(buffer)

		if !ok {
			t.Fatal("DecodeVarint32 returned ok=false after feeding all encoded data")
		}
		if len(remaining) != 0 {
			t.Fatalf("expected no remaining bytes, got %d", len(remaining))
		}
		if len(data) == 0 && len(msg) == 0 {
			return
		}
		if len(msg) != len(data) {
			t.Fatalf("decoded length %d != original length %d", len(msg), len(data))
		}
		for i := range data {
			if msg[i] != data[i] {
				t.Fatalf("decoded byte at index %d differs: got %d, want %d", i, msg[i], data[i])
			}
		}
	})
}

// Property 3: Request 消息帧格式往返一致性
// For any valid pb.Request (random command, random fields), serialize to protobuf →
// EncodeVarint32 → DecodeVarint32 → proto.Unmarshal should produce an equivalent Request.
// **Validates: Requirements 11.6, 12.4**
func TestPropRequestFrameRoundTrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		req := genRequest(t)

		// Serialize to protobuf
		data, err := proto.Marshal(req)
		if err != nil {
			t.Fatalf("proto.Marshal failed: %v", err)
		}

		// Encode with varint32 frame
		framed := EncodeVarint32(data)

		// Decode varint32 frame
		msg, remaining, ok := DecodeVarint32(framed)
		if !ok {
			t.Fatal("DecodeVarint32 returned ok=false for complete framed data")
		}
		if len(remaining) != 0 {
			t.Fatalf("expected no remaining bytes, got %d", len(remaining))
		}

		// Unmarshal protobuf
		var decoded pb.Request
		if err := proto.Unmarshal(msg, &decoded); err != nil {
			t.Fatalf("proto.Unmarshal failed: %v", err)
		}

		// Verify equivalence
		if !proto.Equal(req, &decoded) {
			t.Fatalf("round-trip mismatch:\noriginal: %v\ndecoded:  %v", req, &decoded)
		}
	})
}

// genRequest generates a random pb.Request using rapid generators.
func genRequest(t *rapid.T) *pb.Request {
	commands := []pb.SocketCommon_Command{
		pb.SocketCommon_CONNECT,
		pb.SocketCommon_HEARTBEAT,
		pb.SocketCommon_SUBSCRIBE,
		pb.SocketCommon_UNSUBSCRIBE,
		pb.SocketCommon_DISCONNECT,
	}
	cmdIdx := rapid.IntRange(0, len(commands)-1).Draw(t, "commandIdx")
	cmd := commands[cmdIdx]
	id := rapid.Uint32().Draw(t, "id")

	req := &pb.Request{
		Command: cmd,
		Id:      id,
	}

	switch cmd {
	case pb.SocketCommon_CONNECT:
		req.Connect = &pb.Request_Connect{
			TigerId:         rapid.String().Draw(t, "tigerId"),
			Sign:            rapid.String().Draw(t, "sign"),
			SdkVersion:      rapid.String().Draw(t, "sdkVersion"),
			AcceptVersion:   proto.String(rapid.String().Draw(t, "acceptVersion")),
			SendInterval:    proto.Uint32(rapid.Uint32().Draw(t, "sendInterval")),
			ReceiveInterval: proto.Uint32(rapid.Uint32().Draw(t, "receiveInterval")),
			UseFullTick:     proto.Bool(rapid.Bool().Draw(t, "useFullTick")),
		}
	case pb.SocketCommon_SUBSCRIBE, pb.SocketCommon_UNSUBSCRIBE:
		dataTypes := []pb.SocketCommon_DataType{
			pb.SocketCommon_Quote, pb.SocketCommon_Option, pb.SocketCommon_Future,
			pb.SocketCommon_QuoteDepth, pb.SocketCommon_TradeTick, pb.SocketCommon_Asset,
			pb.SocketCommon_Position, pb.SocketCommon_OrderStatus, pb.SocketCommon_OrderTransaction,
			pb.SocketCommon_StockTop, pb.SocketCommon_OptionTop, pb.SocketCommon_Kline,
		}
		dtIdx := rapid.IntRange(0, len(dataTypes)-1).Draw(t, "dataTypeIdx")
		sub := &pb.Request_Subscribe{
			DataType: dataTypes[dtIdx],
		}
		if rapid.Bool().Draw(t, "hasSymbols") {
			sub.Symbols = proto.String(rapid.String().Draw(t, "symbols"))
		}
		if rapid.Bool().Draw(t, "hasAccount") {
			sub.Account = proto.String(rapid.String().Draw(t, "account"))
		}
		if rapid.Bool().Draw(t, "hasMarket") {
			sub.Market = proto.String(rapid.String().Draw(t, "market"))
		}
		req.Subscribe = sub
	}
	// HEARTBEAT and DISCONNECT have no sub-messages

	return req
}
