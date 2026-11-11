package push

import (
	"encoding/hex"
	"testing"

	"github.com/tigerfintech/openapi-go-sdk/push/pb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Fixtures generated from stock-openapi feat_gth commit 33df7142f.
func TestGthServerWire(t *testing.T) {
	for _, tc := range []struct {
		name, wire string
		price      float64
		time       uint64
		volume     int64
		amount     float64
	}{
		{"trade", "0a185350585720323032363037323320373338302e3020505554100118a4fb9df1f83331000000000000f83f3889fa9df1f833499a9999999999f13f500e5900000000000029407a095072654d61726b6574e9010000000000001c40f001d00ff901000000008087c340820203475448", 1.5, 1784796380425, 7, 12.5},
		{"zero", "0a185350585720323032363037323320373338302e3020505554100118a4fb9df1f8333100000000000000003800499a9999999999f13f50005900000000000000007a095072654d61726b6574e9010000000000000000f001d00ff901000000008087c340820203475448", 0, 0, 0, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			wire, err := hex.DecodeString(tc.wire)
			if err != nil {
				t.Fatal(err)
			}
			var q pb.QuoteData
			if err := proto.Unmarshal(wire, &q); err != nil {
				t.Fatal(err)
			}
			var basic pb.QuoteBasicData
			if err := proto.Unmarshal(wire, &basic); err != nil {
				t.Fatal(err)
			}
			for _, msg := range []proto.Message{&q, &basic} {
				m := msg.ProtoReflect()
				for name, value := range map[protoreflect.Name]any{
					"latestPrice": tc.price, "latestPriceTimestamp": tc.time,
					"preClose": 1.1, "volume": tc.volume, "volumeDecimal": float64(tc.volume),
					"amount": tc.amount, "cumVolume": int64(1000), "cumAmount": float64(9999),
					"session": "GTH", "hourTradingTag": "PreMarket",
				} {
					field := m.Descriptor().Fields().ByName(name)
					if !m.Has(field) || m.Get(field).Interface() != value {
						t.Fatalf("%s: expected present %v, got %v", name, value, m.Get(field))
					}
				}
				if m.Descriptor().Fields().ByName("hourTrading") != nil ||
					m.Descriptor().Messages().ByName("HourTrading") != nil {
					t.Fatal("obsolete nested push contract remains")
				}
			}
			// Exercise the same dispatch as a received Option push frame.
			push := &pb.PushData{DataType: pb.SocketCommon_Option, Body: &pb.PushData_QuoteData{QuoteData: &q}}
			encoded, err := proto.Marshal(push)
			if err != nil {
				t.Fatal(err)
			}
			var received pb.PushData
			if err := proto.Unmarshal(encoded, &received); err != nil {
				t.Fatal(err)
			}
			called := false
			client := &PushClient{}
			client.dispatchPushData(&received, &Callbacks{OnOption: func(got *pb.QuoteData) {
				called = true
				if !proto.Equal(got, &q) {
					t.Fatal("callback changed quote data")
				}
			}})
			if !called {
				t.Fatal("option callback was not invoked")
			}
		})
	}
}

func TestGthMissingFieldsRemainAbsent(t *testing.T) {
	for _, msg := range []proto.Message{&pb.QuoteData{}, &pb.QuoteBasicData{}} {
		if err := proto.Unmarshal([]byte{0x10, 0x01}, msg); err != nil {
			t.Fatal(err)
		}
		m := msg.ProtoReflect()
		for _, name := range []protoreflect.Name{"latestPrice", "latestPriceTimestamp", "volume",
			"volumeDecimal", "amount", "cumVolume", "cumAmount", "session", "hourTradingTag"} {
			if m.Has(m.Descriptor().Fields().ByName(name)) {
				t.Fatalf("%s should stay absent", name)
			}
		}
	}
}
