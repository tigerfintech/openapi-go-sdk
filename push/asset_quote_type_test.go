package push

import (
	"bytes"
	"sync"
	"testing"
	"time"

	"github.com/tigerfintech/openapi-go-sdk/push/pb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// 本文件覆盖 Subscribe.assetQuoteType（proto 字段 5）的构造、wire 编码与订阅恢复。

// newRecordingPushClient 建立 pipe 连接的 PushClient，并记录客户端发出的所有 Request。
// 返回的取数函数会先等待一小段时间，确保异步写入的帧已被服务端读到。
func newRecordingPushClient(t *testing.T) (*PushClient, func() []*pb.Request) {
	t.Helper()
	dialer := newPipeDialer()

	var mu sync.Mutex
	var reqs []*pb.Request

	go func() {
		readProtobufRequest(t, dialer.serverConn) // CONNECT
		sendProtobufResponse(t, dialer.serverConn, &pb.Response{
			Command: pb.SocketCommon_CONNECTED,
		})

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
					mu.Lock()
					reqs = append(reqs, proto.Clone(&req).(*pb.Request))
					mu.Unlock()
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

	return client, func() []*pb.Request {
		time.Sleep(100 * time.Millisecond)
		mu.Lock()
		defer mu.Unlock()
		out := make([]*pb.Request, len(reqs))
		copy(out, reqs)
		return out
	}
}

// findSubscribe 返回第一个指定 dataType 的 SUBSCRIBE 请求。
func findSubscribe(reqs []*pb.Request, dataType pb.SocketCommon_DataType) *pb.Request_Subscribe {
	for _, req := range reqs {
		if req.Command == pb.SocketCommon_SUBSCRIBE && req.Subscribe != nil &&
			req.Subscribe.DataType == dataType {
			return req.Subscribe
		}
	}
	return nil
}

// TestSubscribeAssetQuoteTypeFieldDescriptor 校验 proto 契约：字段号 5、字段名 assetQuoteType、optional string。
func TestSubscribeAssetQuoteTypeFieldDescriptor(t *testing.T) {
	fields := (&pb.Request_Subscribe{}).ProtoReflect().Descriptor().Fields()
	fd := fields.ByNumber(protoreflect.FieldNumber(5))
	if fd == nil {
		t.Fatal("Subscribe 应有字段号为 5 的字段")
	}
	if got := string(fd.Name()); got != "assetQuoteType" {
		t.Errorf("字段 5 名称期望 assetQuoteType，实际 %s", got)
	}
	if fd.Kind() != protoreflect.StringKind {
		t.Errorf("字段 5 类型期望 string，实际 %v", fd.Kind())
	}
	if !fd.HasOptionalKeyword() {
		t.Error("字段 5 应为 optional")
	}
}

// TestBuildSubscribeMessageWithAssetQuoteType_WireBytes 断言 assetQuoteType 编码到 wire 上的
// tag 为字段 5 + wire type 2（0x2a），值为枚举名。
func TestBuildSubscribeMessageWithAssetQuoteType_WireBytes(t *testing.T) {
	req := BuildSubscribeMessageWithAssetQuoteType(pb.SocketCommon_Asset, "", "acc123", "", "ETH")
	if req.Subscribe.GetAssetQuoteType() != "ETH" {
		t.Fatalf("AssetQuoteType 期望 ETH，实际 %s", req.Subscribe.GetAssetQuoteType())
	}

	data, err := proto.Marshal(req.Subscribe)
	if err != nil {
		t.Fatalf("序列化 Subscribe 失败: %v", err)
	}
	want := []byte{0x2a, 0x03, 'E', 'T', 'H'} // field 5, LEN, len=3, "ETH"
	if !bytes.Contains(data, want) {
		t.Errorf("wire 字节应包含 %v（字段 5 + \"ETH\"），实际 %v", want, data)
	}
}

// TestBuildSubscribeMessage_NoAssetQuoteType 验证老签名不会带上 assetQuoteType 字段。
func TestBuildSubscribeMessage_NoAssetQuoteType(t *testing.T) {
	req := BuildSubscribeMessage(pb.SocketCommon_Asset, "", "acc123", "")
	if req.Subscribe.AssetQuoteType != nil {
		t.Errorf("未指定时 AssetQuoteType 应为 nil，实际 %s", req.Subscribe.GetAssetQuoteType())
	}
}

// TestBuildSubscribeMessageWithAssetQuoteType_Empty 验证空串不设置该字段。
func TestBuildSubscribeMessageWithAssetQuoteType_Empty(t *testing.T) {
	req := BuildSubscribeMessageWithAssetQuoteType(pb.SocketCommon_Position, "", "acc123", "", "")
	if req.Subscribe.AssetQuoteType != nil {
		t.Errorf("空串时 AssetQuoteType 应为 nil，实际 %s", req.Subscribe.GetAssetQuoteType())
	}
}

// TestSubscribeAssetWithQuoteType 验证资产订阅把 assetQuoteType 发到服务端。
func TestSubscribeAssetWithQuoteType(t *testing.T) {
	client, reqs := newRecordingPushClient(t)
	defer client.Disconnect()

	if err := client.SubscribeAssetWithQuoteType("acc123", "ETH"); err != nil {
		t.Fatalf("订阅资产失败: %v", err)
	}

	sub := findSubscribe(reqs(), pb.SocketCommon_Asset)
	if sub == nil {
		t.Fatal("应收到 Asset 的 SUBSCRIBE 请求")
	}
	if sub.GetAccount() != "acc123" {
		t.Errorf("Account 期望 acc123，实际 %s", sub.GetAccount())
	}
	if sub.GetAssetQuoteType() != "ETH" {
		t.Errorf("AssetQuoteType 期望 ETH，实际 %s", sub.GetAssetQuoteType())
	}
}

// TestSubscribePositionWithQuoteType 验证持仓订阅把 assetQuoteType 发到服务端。
func TestSubscribePositionWithQuoteType(t *testing.T) {
	client, reqs := newRecordingPushClient(t)
	defer client.Disconnect()

	if err := client.SubscribePositionWithQuoteType("", "OVERNIGHT"); err != nil {
		t.Fatalf("订阅持仓失败: %v", err)
	}

	sub := findSubscribe(reqs(), pb.SocketCommon_Position)
	if sub == nil {
		t.Fatal("应收到 Position 的 SUBSCRIBE 请求")
	}
	if sub.GetAccount() != "test_account" {
		t.Errorf("Account 应回退到配置账户 test_account，实际 %s", sub.GetAccount())
	}
	if sub.GetAssetQuoteType() != "OVERNIGHT" {
		t.Errorf("AssetQuoteType 期望 OVERNIGHT，实际 %s", sub.GetAssetQuoteType())
	}
}

// TestSubscribeAssetLegacyNoQuoteType 验证老方法 SubscribeAsset 不带 assetQuoteType（老用户零改动）。
func TestSubscribeAssetLegacyNoQuoteType(t *testing.T) {
	client, reqs := newRecordingPushClient(t)
	defer client.Disconnect()

	if err := client.SubscribeAsset(""); err != nil {
		t.Fatalf("订阅资产失败: %v", err)
	}

	sub := findSubscribe(reqs(), pb.SocketCommon_Asset)
	if sub == nil {
		t.Fatal("应收到 Asset 的 SUBSCRIBE 请求")
	}
	if sub.AssetQuoteType != nil {
		t.Errorf("老方法不应带 AssetQuoteType，实际 %s", sub.GetAssetQuoteType())
	}
}

// TestResubscribeKeepsAssetQuoteType 验证重连恢复订阅时仍带原来的 assetQuoteType。
func TestResubscribeKeepsAssetQuoteType(t *testing.T) {
	client, reqs := newRecordingPushClient(t)
	defer client.Disconnect()

	if err := client.SubscribeAssetWithQuoteType("", "RTH"); err != nil {
		t.Fatalf("订阅资产失败: %v", err)
	}
	if err := client.SubscribeOrder(""); err != nil {
		t.Fatalf("订阅订单失败: %v", err)
	}

	client.resubscribe()

	all := reqs()
	var assetSubs []*pb.Request_Subscribe
	for _, req := range all {
		if req.Command == pb.SocketCommon_SUBSCRIBE && req.Subscribe != nil &&
			req.Subscribe.DataType == pb.SocketCommon_Asset {
			assetSubs = append(assetSubs, req.Subscribe)
		}
	}
	if len(assetSubs) < 2 {
		t.Fatalf("恢复订阅后应有 2 条 Asset SUBSCRIBE，实际 %d 条", len(assetSubs))
	}
	if got := assetSubs[len(assetSubs)-1].GetAssetQuoteType(); got != "RTH" {
		t.Errorf("恢复订阅的 AssetQuoteType 期望 RTH，实际 %s", got)
	}

	// 订单订阅没有 assetQuoteType，恢复时也不应凭空带上
	if sub := findSubscribe(all, pb.SocketCommon_OrderStatus); sub == nil {
		t.Error("应收到 OrderStatus 的 SUBSCRIBE 请求")
	} else if sub.AssetQuoteType != nil {
		t.Errorf("订单订阅不应带 AssetQuoteType，实际 %s", sub.GetAssetQuoteType())
	}
}

// TestUnsubscribeClearsAssetQuoteType 验证退订后清除记录的 assetQuoteType。
func TestUnsubscribeClearsAssetQuoteType(t *testing.T) {
	client, reqs := newRecordingPushClient(t)
	defer client.Disconnect()

	if err := client.SubscribeAssetWithQuoteType("", "ETH"); err != nil {
		t.Fatalf("订阅资产失败: %v", err)
	}
	if err := client.SubscribePositionWithQuoteType("", "ETH"); err != nil {
		t.Fatalf("订阅持仓失败: %v", err)
	}

	client.mu.RLock()
	assetQuoteType := client.accountQuoteTypes[SubjectAsset]
	client.mu.RUnlock()
	if assetQuoteType != "ETH" {
		t.Errorf("订阅后应记录 assetQuoteType=ETH，实际 %q", assetQuoteType)
	}

	if err := client.UnsubscribeAsset(); err != nil {
		t.Fatalf("退订资产失败: %v", err)
	}
	if err := client.UnsubscribePosition(); err != nil {
		t.Fatalf("退订持仓失败: %v", err)
	}
	_ = reqs()

	client.mu.RLock()
	defer client.mu.RUnlock()
	if len(client.accountQuoteTypes) != 0 {
		t.Errorf("退订后应清空 accountQuoteTypes，实际 %v", client.accountQuoteTypes)
	}
}

// TestSubscribeAssetWithEmptyQuoteTypeClearsRecord 验证再次以空串订阅会清掉之前记录的时段。
func TestSubscribeAssetWithEmptyQuoteTypeClearsRecord(t *testing.T) {
	client, reqs := newRecordingPushClient(t)
	defer client.Disconnect()

	if err := client.SubscribeAssetWithQuoteType("", "ETH"); err != nil {
		t.Fatalf("订阅资产失败: %v", err)
	}
	if err := client.SubscribeAsset(""); err != nil {
		t.Fatalf("重新订阅资产失败: %v", err)
	}
	_ = reqs()

	client.mu.RLock()
	defer client.mu.RUnlock()
	if v, ok := client.accountQuoteTypes[SubjectAsset]; ok {
		t.Errorf("以空 assetQuoteType 重订后不应保留记录，实际 %q", v)
	}
}
