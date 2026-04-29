package push

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tigerfintech/openapi-go-sdk/config"
	"github.com/tigerfintech/openapi-go-sdk/push/pb"
	"google.golang.org/protobuf/proto"
)

// ===== 测试辅助工具 =====

// generateTestPrivateKey 生成测试用 RSA PKCS#1 私钥 PEM
func generateTestPrivateKey(t *testing.T) string {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("生成 RSA 密钥对失败: %v", err)
	}
	privDER := x509.MarshalPKCS1PrivateKey(privateKey)
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privDER,
	})
	return string(privPEM)
}

// newTestConfig 创建测试用 ClientConfig（跳过校验）
func newTestConfig(t *testing.T) *config.ClientConfig {
	t.Helper()
	return &config.ClientConfig{
		TigerID:    "test_tiger_id",
		PrivateKey: generateTestPrivateKey(t),
		Account:    "test_account",
		Language:   "zh_CN",
		Timeout:    15 * time.Second,
		ServerURL:  "https://openapi.tigerfintech.com/gateway",
	}
}

// pipeDialer 使用 net.Pipe() 创建 mock TCP 连接的拨号器
type pipeDialer struct {
	// serverConn 是服务端的连接端，测试代码通过它读写数据
	serverConn net.Conn
	// clientConn 是客户端的连接端，PushClient 通过它读写数据
	clientConn net.Conn
}

// newPipeDialer 创建一对 net.Pipe 连接
func newPipeDialer() *pipeDialer {
	client, server := net.Pipe()
	return &pipeDialer{
		serverConn: server,
		clientConn: client,
	}
}

func (d *pipeDialer) Dial(address string, timeout time.Duration) (net.Conn, error) {
	return d.clientConn, nil
}

// readProtobufRequest 从 TCP 连接读取一个 varint32+protobuf 帧并解析为 Request
func readProtobufRequest(t *testing.T, conn net.Conn) *pb.Request {
	t.Helper()
	data := readFrame(t, conn)
	var req pb.Request
	if err := proto.Unmarshal(data, &req); err != nil {
		t.Fatalf("反序列化 Request 失败: %v", err)
	}
	return &req
}

// readFrame 从 TCP 连接读取一个 varint32 帧
func readFrame(t *testing.T, conn net.Conn) []byte {
	t.Helper()
	var buf []byte
	tmp := make([]byte, 1024)
	for {
		n, err := conn.Read(tmp)
		if err != nil {
			t.Fatalf("读取数据失败: %v", err)
		}
		buf = append(buf, tmp[:n]...)
		msg, _, ok := DecodeVarint32(buf)
		if ok {
			return msg
		}
	}
}

// sendProtobufResponse 构建并发送一个 varint32+protobuf 帧的 Response 到 TCP 连接
func sendProtobufResponse(t *testing.T, conn net.Conn, resp *pb.Response) {
	t.Helper()
	data, err := proto.Marshal(resp)
	if err != nil {
		t.Fatalf("序列化 Response 失败: %v", err)
	}
	framed := EncodeVarint32(data)
	_, err = conn.Write(framed)
	if err != nil {
		t.Fatalf("发送 Response 失败: %v", err)
	}
}

// drainConn 持续读取连接直到关闭（用于模拟服务端保持连接）
func drainConn(conn net.Conn) {
	buf := make([]byte, 1024)
	for {
		_, err := conn.Read(buf)
		if err != nil {
			return
		}
	}
}


// ===== 连接和认证测试 =====

func TestPushClient_NewPushClient(t *testing.T) {
	cfg := newTestConfig(t)
	client := NewPushClient(cfg)

	if client == nil {
		t.Fatal("NewPushClient 不应返回 nil")
	}
	if client.State() != StateDisconnected {
		t.Errorf("初始状态应为 StateDisconnected，实际为 %v", client.State())
	}
	if client.pushURL != defaultPushURL {
		t.Errorf("默认推送地址应为 %s，实际为 %s", defaultPushURL, client.pushURL)
	}
	if !client.autoReconnect {
		t.Error("默认应启用自动重连")
	}
}

func TestPushClient_Options(t *testing.T) {
	cfg := newTestConfig(t)
	client := NewPushClient(cfg,
		WithPushURL("custom.example.com:9883"),
		WithHeartbeatInterval(20*time.Second),
		WithReconnectInterval(10*time.Second),
		WithAutoReconnect(false),
		WithConnectTimeout(60*time.Second),
	)

	if client.pushURL != "custom.example.com:9883" {
		t.Errorf("自定义推送地址未生效")
	}
	if client.heartbeatInterval != 20*time.Second {
		t.Errorf("自定义心跳间隔未生效")
	}
	if client.reconnectInterval != 10*time.Second {
		t.Errorf("自定义重连间隔未生效")
	}
	if client.autoReconnect {
		t.Error("禁用自动重连未生效")
	}
	if client.connectTimeout != 60*time.Second {
		t.Errorf("自定义连接超时未生效")
	}
}

// TestPushClient_ConnectAndAuthenticate 测试连接和认证流程（Protobuf 协议 over TCP）
func TestPushClient_ConnectAndAuthenticate(t *testing.T) {
	var receivedReq *pb.Request
	var reqMu sync.Mutex

	dialer := newPipeDialer()

	// 模拟服务端：读取 CONNECT 请求，发送 CONNECTED 响应
	go func() {
		req := readProtobufRequest(t, dialer.serverConn)
		reqMu.Lock()
		receivedReq = req
		reqMu.Unlock()

		sendProtobufResponse(t, dialer.serverConn, &pb.Response{
			Command: pb.SocketCommon_CONNECTED,
		})

		// 保持连接直到客户端断开
		drainConn(dialer.serverConn)
	}()

	cfg := newTestConfig(t)
	client := NewPushClient(cfg, WithAutoReconnect(false))
	client.dialer = dialer

	err := client.Connect()
	if err != nil {
		t.Fatalf("连接失败: %v", err)
	}
	defer client.Disconnect()

	reqMu.Lock()
	defer reqMu.Unlock()

	// 验证认证消息是 Protobuf CONNECT 命令
	if receivedReq.Command != pb.SocketCommon_CONNECT {
		t.Errorf("认证消息 command 应为 CONNECT，实际为 %v", receivedReq.Command)
	}
	if receivedReq.Connect == nil {
		t.Fatal("认证消息应包含 Connect 子消息")
	}
	if receivedReq.Connect.TigerId != "test_tiger_id" {
		t.Errorf("TigerId 应为 test_tiger_id，实际为 %s", receivedReq.Connect.TigerId)
	}
	if receivedReq.Connect.Sign == "" {
		t.Error("签名不应为空")
	}
	if receivedReq.Id == 0 {
		t.Error("请求 ID 不应为 0")
	}

	// 验证连接状态
	if client.State() != StateConnected {
		t.Errorf("连接后状态应为 StateConnected，实际为 %v", client.State())
	}
}

// TestPushClient_Disconnect 测试断开连接（发送 DISCONNECT 消息）
func TestPushClient_Disconnect(t *testing.T) {
	var disconnectReceived int32

	dialer := newPipeDialer()

	go func() {
		// 读取 CONNECT 请求
		readProtobufRequest(t, dialer.serverConn)
		// 发送 CONNECTED 响应
		sendProtobufResponse(t, dialer.serverConn, &pb.Response{
			Command: pb.SocketCommon_CONNECTED,
		})

		// 读取后续消息（心跳或 DISCONNECT）
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
				if proto.Unmarshal(msg, &req) == nil && req.Command == pb.SocketCommon_DISCONNECT {
					atomic.StoreInt32(&disconnectReceived, 1)
				}
				buf = remaining
			}
		}
	}()

	cfg := newTestConfig(t)
	client := NewPushClient(cfg, WithAutoReconnect(false))
	client.dialer = dialer

	client.Connect()
	err := client.Disconnect()
	if err != nil {
		t.Fatalf("断开连接失败: %v", err)
	}

	if client.State() != StateDisconnected {
		t.Errorf("断开后状态应为 StateDisconnected，实际为 %v", client.State())
	}

	// 验证发送了 DISCONNECT 消息
	time.Sleep(100 * time.Millisecond)
	if atomic.LoadInt32(&disconnectReceived) != 1 {
		t.Error("应发送 DISCONNECT 消息")
	}
}

func TestPushClient_DisconnectWhenNotConnected(t *testing.T) {
	cfg := newTestConfig(t)
	client := NewPushClient(cfg)

	err := client.Disconnect()
	if err != nil {
		t.Fatalf("未连接时断开不应报错: %v", err)
	}
}

func TestPushClient_ConnectWhenAlreadyConnected(t *testing.T) {
	dialer := newPipeDialer()

	go func() {
		readProtobufRequest(t, dialer.serverConn)
		sendProtobufResponse(t, dialer.serverConn, &pb.Response{
			Command: pb.SocketCommon_CONNECTED,
		})
		drainConn(dialer.serverConn)
	}()

	cfg := newTestConfig(t)
	client := NewPushClient(cfg, WithAutoReconnect(false))
	client.dialer = dialer

	client.Connect()
	defer client.Disconnect()

	err := client.Connect()
	if err == nil {
		t.Fatal("重复连接应返回错误")
	}
}


// ===== 心跳测试 =====

func TestPushClient_Heartbeat(t *testing.T) {
	var heartbeatCount int32

	dialer := newPipeDialer()

	go func() {
		readProtobufRequest(t, dialer.serverConn)
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
				if proto.Unmarshal(msg, &req) == nil && req.Command == pb.SocketCommon_HEARTBEAT {
					atomic.AddInt32(&heartbeatCount, 1)
				}
				buf = remaining
			}
		}
	}()

	cfg := newTestConfig(t)
	client := NewPushClient(cfg,
		WithAutoReconnect(false),
		WithHeartbeatInterval(100*time.Millisecond),
	)
	client.dialer = dialer

	client.Connect()
	defer client.Disconnect()

	// 等待几个心跳周期
	time.Sleep(350 * time.Millisecond)

	count := atomic.LoadInt32(&heartbeatCount)
	if count < 2 {
		t.Errorf("应至少收到 2 个心跳，实际收到 %d 个", count)
	}
}

// ===== 订阅/退订测试 =====

func TestPushClient_SubscribeQuote(t *testing.T) {
	var receivedReqs []*pb.Request
	var mu sync.Mutex

	dialer := newPipeDialer()

	go func() {
		readProtobufRequest(t, dialer.serverConn)
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
					receivedReqs = append(receivedReqs, proto.Clone(&req).(*pb.Request))
					mu.Unlock()
				}
				buf = remaining
			}
		}
	}()

	cfg := newTestConfig(t)
	client := NewPushClient(cfg, WithAutoReconnect(false), WithHeartbeatInterval(10*time.Second))
	client.dialer = dialer
	client.Connect()
	defer client.Disconnect()

	err := client.SubscribeQuote([]string{"AAPL", "TSLA"})
	if err != nil {
		t.Fatalf("订阅行情失败: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// 验证订阅状态
	subs := client.GetSubscriptions()
	if symbols, ok := subs[SubjectQuote]; !ok {
		t.Error("应有 quote 订阅记录")
	} else if len(symbols) != 2 {
		t.Errorf("应订阅 2 个标的，实际 %d 个", len(symbols))
	}

	// 验证发送了 SUBSCRIBE 请求
	mu.Lock()
	defer mu.Unlock()
	found := false
	for _, req := range receivedReqs {
		if req.Command == pb.SocketCommon_SUBSCRIBE {
			found = true
			if req.Subscribe == nil {
				t.Error("SUBSCRIBE 请求应包含 Subscribe 子消息")
			} else {
				if req.Subscribe.DataType != pb.SocketCommon_Quote {
					t.Errorf("DataType 应为 Quote，实际为 %v", req.Subscribe.DataType)
				}
				if req.Subscribe.GetSymbols() != "AAPL,TSLA" {
					t.Errorf("Symbols 应为 AAPL,TSLA，实际为 %s", req.Subscribe.GetSymbols())
				}
			}
		}
	}
	if !found {
		t.Error("应收到 SUBSCRIBE 请求")
	}
}

func TestPushClient_UnsubscribeQuote(t *testing.T) {
	dialer := newPipeDialer()

	go func() {
		readProtobufRequest(t, dialer.serverConn)
		sendProtobufResponse(t, dialer.serverConn, &pb.Response{
			Command: pb.SocketCommon_CONNECTED,
		})
		drainConn(dialer.serverConn)
	}()

	cfg := newTestConfig(t)
	client := NewPushClient(cfg, WithAutoReconnect(false), WithHeartbeatInterval(10*time.Second))
	client.dialer = dialer
	client.Connect()
	defer client.Disconnect()

	client.SubscribeQuote([]string{"AAPL", "TSLA", "GOOG"})
	client.UnsubscribeQuote([]string{"TSLA"})

	time.Sleep(100 * time.Millisecond)

	subs := client.GetSubscriptions()
	if symbols, ok := subs[SubjectQuote]; !ok {
		t.Error("应有 quote 订阅记录")
	} else if len(symbols) != 2 {
		t.Errorf("退订后应剩 2 个标的，实际 %d 个", len(symbols))
	}
}

func TestPushClient_SubscribeMultipleSubjects(t *testing.T) {
	dialer := newPipeDialer()

	go func() {
		readProtobufRequest(t, dialer.serverConn)
		sendProtobufResponse(t, dialer.serverConn, &pb.Response{
			Command: pb.SocketCommon_CONNECTED,
		})
		drainConn(dialer.serverConn)
	}()

	cfg := newTestConfig(t)
	client := NewPushClient(cfg, WithAutoReconnect(false), WithHeartbeatInterval(10*time.Second))
	client.dialer = dialer
	client.Connect()
	defer client.Disconnect()

	client.SubscribeQuote([]string{"AAPL"})
	client.SubscribeTick([]string{"TSLA"})
	client.SubscribeDepth([]string{"GOOG"})
	client.SubscribeOption([]string{"AAPL"})
	client.SubscribeFuture([]string{"ES"})
	client.SubscribeKline([]string{"AAPL"})

	subs := client.GetSubscriptions()
	if len(subs) != 6 {
		t.Errorf("应有 6 种订阅，实际 %d 种", len(subs))
	}
}

func TestPushClient_UnsubscribeAll(t *testing.T) {
	dialer := newPipeDialer()

	go func() {
		readProtobufRequest(t, dialer.serverConn)
		sendProtobufResponse(t, dialer.serverConn, &pb.Response{
			Command: pb.SocketCommon_CONNECTED,
		})
		drainConn(dialer.serverConn)
	}()

	cfg := newTestConfig(t)
	client := NewPushClient(cfg, WithAutoReconnect(false), WithHeartbeatInterval(10*time.Second))
	client.dialer = dialer
	client.Connect()
	defer client.Disconnect()

	client.SubscribeQuote([]string{"AAPL", "TSLA"})
	client.UnsubscribeQuote(nil)

	subs := client.GetSubscriptions()
	if _, ok := subs[SubjectQuote]; ok {
		t.Error("退订全部后不应有 quote 订阅记录")
	}
}


// ===== 账户推送测试 =====

func TestPushClient_SubscribeAsset(t *testing.T) {
	dialer := newPipeDialer()

	go func() {
		readProtobufRequest(t, dialer.serverConn)
		sendProtobufResponse(t, dialer.serverConn, &pb.Response{
			Command: pb.SocketCommon_CONNECTED,
		})
		drainConn(dialer.serverConn)
	}()

	cfg := newTestConfig(t)
	client := NewPushClient(cfg, WithAutoReconnect(false), WithHeartbeatInterval(10*time.Second))
	client.dialer = dialer
	client.Connect()
	defer client.Disconnect()

	err := client.SubscribeAsset("")
	if err != nil {
		t.Fatalf("订阅资产失败: %v", err)
	}

	acctSubs := client.GetAccountSubscriptions()
	found := false
	for _, s := range acctSubs {
		if s == SubjectAsset {
			found = true
		}
	}
	if !found {
		t.Error("应有 asset 账户订阅记录")
	}
}

func TestPushClient_SubscribeAllAccountPush(t *testing.T) {
	dialer := newPipeDialer()

	go func() {
		readProtobufRequest(t, dialer.serverConn)
		sendProtobufResponse(t, dialer.serverConn, &pb.Response{
			Command: pb.SocketCommon_CONNECTED,
		})
		drainConn(dialer.serverConn)
	}()

	cfg := newTestConfig(t)
	client := NewPushClient(cfg, WithAutoReconnect(false), WithHeartbeatInterval(10*time.Second))
	client.dialer = dialer
	client.Connect()
	defer client.Disconnect()

	client.SubscribeAsset("")
	client.SubscribePosition("")
	client.SubscribeOrder("")
	client.SubscribeTransaction("")

	acctSubs := client.GetAccountSubscriptions()
	if len(acctSubs) != 4 {
		t.Errorf("应有 4 种账户订阅，实际 %d 种", len(acctSubs))
	}
}

func TestPushClient_UnsubscribeAccountPush(t *testing.T) {
	dialer := newPipeDialer()

	go func() {
		readProtobufRequest(t, dialer.serverConn)
		sendProtobufResponse(t, dialer.serverConn, &pb.Response{
			Command: pb.SocketCommon_CONNECTED,
		})
		drainConn(dialer.serverConn)
	}()

	cfg := newTestConfig(t)
	client := NewPushClient(cfg, WithAutoReconnect(false), WithHeartbeatInterval(10*time.Second))
	client.dialer = dialer
	client.Connect()
	defer client.Disconnect()

	client.SubscribeAsset("")
	client.SubscribePosition("")
	client.UnsubscribeAsset()
	client.UnsubscribePosition()

	acctSubs := client.GetAccountSubscriptions()
	if len(acctSubs) != 0 {
		t.Errorf("退订后不应有账户订阅，实际 %d 种", len(acctSubs))
	}
}

// ===== 回调测试 =====

func TestPushClient_ConnectCallback(t *testing.T) {
	dialer := newPipeDialer()

	go func() {
		readProtobufRequest(t, dialer.serverConn)
		sendProtobufResponse(t, dialer.serverConn, &pb.Response{
			Command: pb.SocketCommon_CONNECTED,
		})
		drainConn(dialer.serverConn)
	}()

	cfg := newTestConfig(t)
	client := NewPushClient(cfg, WithAutoReconnect(false))
	client.dialer = dialer

	var connected int32
	client.SetCallbacks(Callbacks{
		OnConnect: func() {
			atomic.StoreInt32(&connected, 1)
		},
	})

	client.Connect()
	defer client.Disconnect()

	time.Sleep(100 * time.Millisecond)
	if atomic.LoadInt32(&connected) != 1 {
		t.Error("连接成功回调未触发")
	}
}

func TestPushClient_DisconnectCallback(t *testing.T) {
	dialer := newPipeDialer()

	go func() {
		readProtobufRequest(t, dialer.serverConn)
		sendProtobufResponse(t, dialer.serverConn, &pb.Response{
			Command: pb.SocketCommon_CONNECTED,
		})
		drainConn(dialer.serverConn)
	}()

	cfg := newTestConfig(t)
	client := NewPushClient(cfg, WithAutoReconnect(false))
	client.dialer = dialer

	var disconnected int32
	client.SetCallbacks(Callbacks{
		OnDisconnect: func() {
			atomic.StoreInt32(&disconnected, 1)
		},
	})

	client.Connect()
	client.Disconnect()

	time.Sleep(100 * time.Millisecond)
	if atomic.LoadInt32(&disconnected) != 1 {
		t.Error("断开连接回调未触发")
	}
}

// TestPushClient_QuoteCallback 测试行情推送回调（Protobuf over TCP）
func TestPushClient_QuoteCallback(t *testing.T) {
	dialer := newPipeDialer()

	go func() {
		readProtobufRequest(t, dialer.serverConn)
		sendProtobufResponse(t, dialer.serverConn, &pb.Response{
			Command: pb.SocketCommon_CONNECTED,
		})

		// 发送行情推送
		sendProtobufResponse(t, dialer.serverConn, &pb.Response{
			Command: pb.SocketCommon_MESSAGE,
			Body: &pb.PushData{
				DataType: pb.SocketCommon_Quote,
				Body: &pb.PushData_QuoteData{
					QuoteData: &pb.QuoteData{
						Symbol:      "AAPL",
						LatestPrice: proto.Float64(155.0),
					},
				},
			},
		})

		drainConn(dialer.serverConn)
	}()

	cfg := newTestConfig(t)
	client := NewPushClient(cfg, WithAutoReconnect(false))
	client.dialer = dialer

	var receivedQuote *pb.QuoteData
	var quoteMu sync.Mutex
	client.SetCallbacks(Callbacks{
		OnQuote: func(data *pb.QuoteData) {
			quoteMu.Lock()
			receivedQuote = data
			quoteMu.Unlock()
		},
	})

	client.Connect()
	defer client.Disconnect()

	time.Sleep(300 * time.Millisecond)

	quoteMu.Lock()
	defer quoteMu.Unlock()
	if receivedQuote == nil {
		t.Fatal("行情回调未触发")
	}
	if receivedQuote.GetSymbol() != "AAPL" {
		t.Errorf("Symbol 应为 AAPL，实际为 %s", receivedQuote.GetSymbol())
	}
	if receivedQuote.GetLatestPrice() != 155.0 {
		t.Errorf("LatestPrice 应为 155.0，实际为 %f", receivedQuote.GetLatestPrice())
	}
}

// TestPushClient_OrderCallback 测试订单推送回调（Protobuf over TCP）
func TestPushClient_OrderCallback(t *testing.T) {
	dialer := newPipeDialer()

	go func() {
		readProtobufRequest(t, dialer.serverConn)
		sendProtobufResponse(t, dialer.serverConn, &pb.Response{
			Command: pb.SocketCommon_CONNECTED,
		})

		sendProtobufResponse(t, dialer.serverConn, &pb.Response{
			Command: pb.SocketCommon_MESSAGE,
			Body: &pb.PushData{
				DataType: pb.SocketCommon_OrderStatus,
				Body: &pb.PushData_OrderStatusData{
					OrderStatusData: &pb.OrderStatusData{
						Account: "acc123",
						Symbol:  "AAPL",
						Status:  "Filled",
					},
				},
			},
		})

		drainConn(dialer.serverConn)
	}()

	cfg := newTestConfig(t)
	client := NewPushClient(cfg, WithAutoReconnect(false))
	client.dialer = dialer

	var received *pb.OrderStatusData
	var mu sync.Mutex
	client.SetCallbacks(Callbacks{
		OnOrder: func(data *pb.OrderStatusData) {
			mu.Lock()
			received = data
			mu.Unlock()
		},
	})

	client.Connect()
	defer client.Disconnect()
	time.Sleep(300 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if received == nil {
		t.Fatal("订单回调未触发")
	}
	if received.GetStatus() != "Filled" {
		t.Errorf("Status 应为 Filled，实际为 %s", received.GetStatus())
	}
}

// TestPushClient_ErrorCallback 测试错误回调（Protobuf ERROR 响应）
func TestPushClient_ErrorCallback(t *testing.T) {
	dialer := newPipeDialer()

	go func() {
		readProtobufRequest(t, dialer.serverConn)
		sendProtobufResponse(t, dialer.serverConn, &pb.Response{
			Command: pb.SocketCommon_CONNECTED,
		})

		sendProtobufResponse(t, dialer.serverConn, &pb.Response{
			Command: pb.SocketCommon_ERROR,
			Msg:     proto.String("服务端内部错误"),
		})

		drainConn(dialer.serverConn)
	}()

	cfg := newTestConfig(t)
	client := NewPushClient(cfg, WithAutoReconnect(false))
	client.dialer = dialer

	var errReceived error
	var mu sync.Mutex
	client.SetCallbacks(Callbacks{
		OnError: func(err error) {
			mu.Lock()
			errReceived = err
			mu.Unlock()
		},
	})

	client.Connect()
	defer client.Disconnect()
	time.Sleep(300 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if errReceived == nil {
		t.Fatal("错误回调未触发")
	}
}

// TestPushClient_MultipleCallbacks 测试多种回调同时注册
func TestPushClient_MultipleCallbacks(t *testing.T) {
	dialer := newPipeDialer()

	go func() {
		readProtobufRequest(t, dialer.serverConn)
		sendProtobufResponse(t, dialer.serverConn, &pb.Response{
			Command: pb.SocketCommon_CONNECTED,
		})

		// 发送行情推送
		sendProtobufResponse(t, dialer.serverConn, &pb.Response{
			Command: pb.SocketCommon_MESSAGE,
			Body: &pb.PushData{
				DataType: pb.SocketCommon_Quote,
				Body: &pb.PushData_QuoteData{
					QuoteData: &pb.QuoteData{Symbol: "AAPL"},
				},
			},
		})

		// 发送逐笔推送
		sendProtobufResponse(t, dialer.serverConn, &pb.Response{
			Command: pb.SocketCommon_MESSAGE,
			Body: &pb.PushData{
				DataType: pb.SocketCommon_TradeTick,
				Body: &pb.PushData_TradeTickData{
					TradeTickData: &pb.TradeTickData{Symbol: "TSLA"},
				},
			},
		})

		drainConn(dialer.serverConn)
	}()

	cfg := newTestConfig(t)
	client := NewPushClient(cfg, WithAutoReconnect(false))
	client.dialer = dialer

	var quoteCount, tickCount int32
	client.SetCallbacks(Callbacks{
		OnQuote: func(data *pb.QuoteData) { atomic.AddInt32(&quoteCount, 1) },
		OnTick:  func(data *pb.TradeTickData) { atomic.AddInt32(&tickCount, 1) },
	})

	client.Connect()
	defer client.Disconnect()
	time.Sleep(300 * time.Millisecond)

	if atomic.LoadInt32(&quoteCount) != 1 {
		t.Errorf("行情回调应触发 1 次，实际 %d 次", quoteCount)
	}
	if atomic.LoadInt32(&tickCount) != 1 {
		t.Errorf("逐笔回调应触发 1 次，实际 %d 次", tickCount)
	}
}

// TestPushClient_SubscriptionStateManagement 测试订阅状态管理的纯逻辑
func TestPushClient_SubscriptionStateManagement(t *testing.T) {
	cfg := newTestConfig(t)
	client := NewPushClient(cfg)

	client.addSubscription(SubjectQuote, []string{"AAPL", "TSLA"})
	client.addSubscription(SubjectTick, []string{"GOOG"})

	subs := client.GetSubscriptions()
	if len(subs) != 2 {
		t.Errorf("应有 2 种订阅，实际 %d 种", len(subs))
	}

	client.addSubscription(SubjectQuote, []string{"GOOG"})
	subs = client.GetSubscriptions()
	if len(subs[SubjectQuote]) != 3 {
		t.Errorf("quote 应有 3 个标的，实际 %d 个", len(subs[SubjectQuote]))
	}

	client.removeSubscription(SubjectQuote, []string{"TSLA"})
	subs = client.GetSubscriptions()
	if len(subs[SubjectQuote]) != 2 {
		t.Errorf("退订后 quote 应有 2 个标的，实际 %d 个", len(subs[SubjectQuote]))
	}

	client.removeSubscription(SubjectQuote, nil)
	subs = client.GetSubscriptions()
	if _, ok := subs[SubjectQuote]; ok {
		t.Error("全部退订后不应有 quote 记录")
	}
}

// Ensure io import is used (for drainConn pattern)
var _ = io.EOF