package push

import (
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/tigerfintech/openapi-go-sdk/config"
	"github.com/tigerfintech/openapi-go-sdk/push/pb"
	"github.com/tigerfintech/openapi-go-sdk/signer"
	"google.golang.org/protobuf/proto"
)

const (
	// 默认推送服务器地址（raw TCP + TLS）
	defaultPushURL = "openapi.tigerfintech.com:9883"
	// 默认心跳间隔
	defaultHeartbeatInterval = 10 * time.Second
	// 默认重连间隔
	defaultReconnectInterval = 5 * time.Second
	// 最大重连间隔
	maxReconnectInterval = 60 * time.Second
	// 默认连接超时
	defaultConnectTimeout = 30 * time.Second
	// SDK 版本
	sdkVersion = "go/1.0.0"
	// 协议版本
	acceptVersion = "2"
	// 默认心跳发送间隔（毫秒）
	defaultSendInterval = 10000
	// 默认心跳接收间隔（毫秒）
	defaultReceiveInterval = 10000
)

// ConnectionState 连接状态
type ConnectionState int

const (
	StateDisconnected ConnectionState = iota
	StateConnecting
	StateConnected
)

// PushClientOption PushClient 配置选项
type PushClientOption func(*PushClient)

// WithPushURL 设置推送服务器地址
func WithPushURL(url string) PushClientOption {
	return func(c *PushClient) { c.pushURL = url }
}

// WithHeartbeatInterval 设置心跳间隔
func WithHeartbeatInterval(d time.Duration) PushClientOption {
	return func(c *PushClient) { c.heartbeatInterval = d }
}

// WithReconnectInterval 设置初始重连间隔
func WithReconnectInterval(d time.Duration) PushClientOption {
	return func(c *PushClient) { c.reconnectInterval = d }
}

// WithAutoReconnect 设置是否自动重连
func WithAutoReconnect(auto bool) PushClientOption {
	return func(c *PushClient) { c.autoReconnect = auto }
}

// WithConnectTimeout 设置连接超时
func WithConnectTimeout(d time.Duration) PushClientOption {
	return func(c *PushClient) { c.connectTimeout = d }
}

// PushClient TCP+TLS 推送客户端
type PushClient struct {
	config            *config.ClientConfig
	pushURL           string
	heartbeatInterval time.Duration
	reconnectInterval time.Duration
	connectTimeout    time.Duration
	autoReconnect     bool

	// TCP 连接
	conn  net.Conn
	state ConnectionState

	// 回调
	callbacks Callbacks

	// 订阅状态管理
	subscriptions map[SubjectType]map[string]bool // subject -> symbols set
	accountSubs   map[SubjectType]bool            // 账户级别订阅

	// 并发控制
	mu      sync.RWMutex
	stopCh  chan struct{}
	doneCh  chan struct{}
	writeMu sync.Mutex

	// 认证完成信号
	connectedCh chan struct{}

	// 用于测试的 dialer（可注入）
	dialer TCPDialer
}

// TCPDialer TCP 拨号器接口，方便测试注入
type TCPDialer interface {
	Dial(address string, timeout time.Duration) (net.Conn, error)
}

// defaultDialer 默认的 TCP+TLS 拨号器
type defaultDialer struct{}

func (d *defaultDialer) Dial(address string, timeout time.Duration) (net.Conn, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		MinVersion:         tls.VersionTLS12,
	}
	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: timeout}, "tcp", address, tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("TCP+TLS 连接失败: %w", err)
	}
	return conn, nil
}

// NewPushClient 创建推送客户端
func NewPushClient(cfg *config.ClientConfig, opts ...PushClientOption) *PushClient {
	c := &PushClient{
		config:            cfg,
		pushURL:           defaultPushURL,
		heartbeatInterval: defaultHeartbeatInterval,
		reconnectInterval: defaultReconnectInterval,
		connectTimeout:    defaultConnectTimeout,
		autoReconnect:     true,
		state:             StateDisconnected,
		subscriptions:     make(map[SubjectType]map[string]bool),
		accountSubs:       make(map[SubjectType]bool),
		dialer:            &defaultDialer{},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// SetCallbacks 设置回调函数集合
func (c *PushClient) SetCallbacks(cb Callbacks) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.callbacks = cb
}

// State 获取当前连接状态
func (c *PushClient) State() ConnectionState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state
}

// Connect 连接到推送服务器并进行认证
func (c *PushClient) Connect() error {
	c.mu.Lock()
	if c.state != StateDisconnected {
		c.mu.Unlock()
		return fmt.Errorf("客户端已连接或正在连接中")
	}
	c.state = StateConnecting
	c.stopCh = make(chan struct{})
	c.doneCh = make(chan struct{})
	c.connectedCh = make(chan struct{})
	c.mu.Unlock()

	// 建立 TCP+TLS 连接
	conn, err := c.dialer.Dial(c.pushURL, c.connectTimeout)
	if err != nil {
		c.mu.Lock()
		c.state = StateDisconnected
		c.mu.Unlock()
		return err
	}

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()

	// 启动消息读取协程（需要在发送 CONNECT 之前启动，以接收 CONNECTED 响应）
	go c.readLoop()

	// 发送认证消息
	if err := c.authenticate(); err != nil {
		c.mu.Lock()
		c.state = StateDisconnected
		if c.stopCh != nil {
			close(c.stopCh)
		}
		c.conn = nil
		c.mu.Unlock()
		conn.Close()
		return fmt.Errorf("认证失败: %w", err)
	}

	// 等待 CONNECTED 响应
	select {
	case <-c.connectedCh:
		// 连接成功
	case <-time.After(c.connectTimeout):
		c.mu.Lock()
		c.state = StateDisconnected
		if c.stopCh != nil {
			close(c.stopCh)
		}
		c.conn = nil
		c.mu.Unlock()
		conn.Close()
		return fmt.Errorf("等待 CONNECTED 响应超时")
	}

	// 启动心跳协程
	go c.heartbeatLoop()

	return nil
}

// authenticate 发送认证消息
func (c *PushClient) authenticate() error {
	signContent := c.config.TigerID
	sign, err := signer.SignWithRSA(c.config.PrivateKey, signContent)
	if err != nil {
		return fmt.Errorf("签名失败: %w", err)
	}

	req := BuildConnectMessage(
		c.config.TigerID,
		sign,
		sdkVersion,
		acceptVersion,
		defaultSendInterval,
		defaultReceiveInterval,
		false,
	)

	return c.sendMessage(req)
}

// Disconnect 断开连接
func (c *PushClient) Disconnect() error {
	c.mu.Lock()
	if c.state == StateDisconnected {
		c.mu.Unlock()
		return nil
	}
	c.state = StateDisconnected
	conn := c.conn
	c.mu.Unlock()

	// 发送 DISCONNECT 消息（在关闭连接之前）
	if conn != nil {
		disconnectReq := BuildDisconnectMessage()
		// 忽略发送错误，因为连接可能已经断开
		_ = c.sendMessage(disconnectReq)
	}

	c.mu.Lock()
	if c.stopCh != nil {
		close(c.stopCh)
	}
	c.conn = nil
	c.mu.Unlock()

	var err error
	if conn != nil {
		err = conn.Close()
	}

	// 等待协程退出
	if c.doneCh != nil {
		select {
		case <-c.doneCh:
		case <-time.After(5 * time.Second):
		}
	}

	// 触发断开回调
	c.mu.RLock()
	cb := c.callbacks.OnDisconnect
	c.mu.RUnlock()
	if cb != nil {
		cb()
	}

	return err
}

// sendMessage 发送 Protobuf 消息到 TCP 连接
func (c *PushClient) sendMessage(req *pb.Request) error {
	data, err := proto.Marshal(req)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	// 添加 varint32 长度前缀
	framed := EncodeVarint32(data)

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return fmt.Errorf("TCP 连接未建立")
	}

	// 确保完整写入所有数据
	written := 0
	for written < len(framed) {
		n, err := conn.Write(framed[written:])
		if err != nil {
			return fmt.Errorf("写入 TCP 连接失败: %w", err)
		}
		written += n
	}
	return nil
}

// readLoop TCP 流消息读取循环，处理 varint32 帧的分包/粘包
func (c *PushClient) readLoop() {
	defer func() {
		select {
		case <-c.doneCh:
		default:
			close(c.doneCh)
		}
	}()

	buf := make([]byte, 4096)
	var buffer []byte

	for {
		select {
		case <-c.stopCh:
			return
		default:
		}

		c.mu.RLock()
		conn := c.conn
		c.mu.RUnlock()

		if conn == nil {
			return
		}

		n, err := conn.Read(buf)
		if err != nil {
			// 检查是否是主动关闭
			select {
			case <-c.stopCh:
				return
			default:
			}

			// 触发错误回调
			c.mu.RLock()
			errCb := c.callbacks.OnError
			c.mu.RUnlock()
			if errCb != nil {
				errCb(err)
			}

			// 尝试自动重连
			c.mu.RLock()
			autoReconnect := c.autoReconnect
			c.mu.RUnlock()
			if autoReconnect {
				go c.reconnect()
			}
			return
		}

		buffer = append(buffer, buf[:n]...)

		// 循环解析 varint32 帧（处理粘包：一次 Read 可能包含多个完整帧）
		for {
			msg, remaining, ok := DecodeVarint32(buffer)
			if !ok {
				// 帧不完整，等待更多数据（处理分包）
				break
			}

			// 反序列化 Protobuf Response
			var response pb.Response
			if err := proto.Unmarshal(msg, &response); err != nil {
				c.mu.RLock()
				errCb := c.callbacks.OnError
				c.mu.RUnlock()
				if errCb != nil {
					errCb(fmt.Errorf("反序列化 Response 失败: %w", err))
				}
				buffer = remaining
				continue
			}

			c.handleMessage(&response)
			buffer = remaining
		}
	}
}

// heartbeatLoop 心跳保活循环
func (c *PushClient) heartbeatLoop() {
	ticker := time.NewTicker(c.heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopCh:
			return
		case <-ticker.C:
			req := BuildHeartBeatMessage()
			if err := c.sendMessage(req); err != nil {
				// 心跳发送失败，可能连接已断开
				return
			}
		}
	}
}

// reconnect 自动重连
func (c *PushClient) reconnect() {
	c.mu.Lock()
	c.state = StateDisconnected
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	c.mu.Unlock()

	interval := c.reconnectInterval
	for {
		select {
		case <-c.stopCh:
			return
		default:
		}

		time.Sleep(interval)

		if err := c.Connect(); err != nil {
			// 指数退避
			interval = interval * 2
			if interval > maxReconnectInterval {
				interval = maxReconnectInterval
			}
			continue
		}

		// 重连成功，恢复订阅
		c.resubscribe()
		return
	}
}

// resubscribe 重连后恢复之前的订阅
func (c *PushClient) resubscribe() {
	c.mu.RLock()
	subs := make(map[SubjectType][]string)
	for subject, symbols := range c.subscriptions {
		list := make([]string, 0, len(symbols))
		for s := range symbols {
			list = append(list, s)
		}
		subs[subject] = list
	}
	acctSubs := make(map[SubjectType]bool)
	for k, v := range c.accountSubs {
		acctSubs[k] = v
	}
	c.mu.RUnlock()

	// 恢复行情订阅
	for subject, symbols := range subs {
		c.subscribe(subject, symbols, "", "")
	}

	// 恢复账户订阅
	for subject := range acctSubs {
		c.subscribe(subject, nil, c.config.Account, "")
	}
}

// handleMessage 处理收到的 Protobuf Response 消息
func (c *PushClient) handleMessage(response *pb.Response) {
	c.mu.RLock()
	cb := c.callbacks
	c.mu.RUnlock()

	switch response.Command {
	case pb.SocketCommon_CONNECTED:
		// 标记连接成功
		c.mu.Lock()
		c.state = StateConnected
		connectedCh := c.connectedCh
		c.mu.Unlock()

		// 通知 Connect() 方法认证完成
		if connectedCh != nil {
			select {
			case <-connectedCh:
			default:
				close(connectedCh)
			}
		}

		// 触发连接成功回调
		if cb.OnConnect != nil {
			cb.OnConnect()
		}

	case pb.SocketCommon_HEARTBEAT:
		// 心跳响应，忽略

	case pb.SocketCommon_MESSAGE:
		// 提取 PushData 并分发
		pushData := response.GetBody()
		if pushData == nil {
			if cb.OnError != nil {
				cb.OnError(fmt.Errorf("MESSAGE 响应缺少 body"))
			}
			return
		}
		c.dispatchPushData(pushData, &cb)

	case pb.SocketCommon_ERROR:
		if cb.OnError != nil {
			msg := response.GetMsg()
			cb.OnError(fmt.Errorf("服务端错误: %s", msg))
		}
		// 检查是否是 kickout
		if cb.OnKickout != nil && response.GetMsg() != "" {
			cb.OnKickout(response.GetMsg())
		}

	case pb.SocketCommon_DISCONNECT:
		if cb.OnDisconnect != nil {
			cb.OnDisconnect()
		}
	}
}

// dispatchPushData 根据 PushData 的 dataType 分发到对应回调
func (c *PushClient) dispatchPushData(pushData *pb.PushData, cb *Callbacks) {
	switch pushData.DataType {
	case pb.SocketCommon_Quote:
		if cb.OnQuote != nil {
			if d := pushData.GetQuoteData(); d != nil {
				cb.OnQuote(d)
			}
		}
	case pb.SocketCommon_Option:
		if cb.OnOption != nil {
			if d := pushData.GetQuoteData(); d != nil {
				cb.OnOption(d)
			}
		}
	case pb.SocketCommon_Future:
		if cb.OnFuture != nil {
			if d := pushData.GetQuoteData(); d != nil {
				cb.OnFuture(d)
			}
		}
	case pb.SocketCommon_QuoteDepth:
		if cb.OnDepth != nil {
			if d := pushData.GetQuoteDepthData(); d != nil {
				cb.OnDepth(d)
			}
		}
	case pb.SocketCommon_TradeTick:
		if cb.OnTick != nil {
			if d := pushData.GetTradeTickData(); d != nil {
				cb.OnTick(d)
			}
		}
	case pb.SocketCommon_Asset:
		if cb.OnAsset != nil {
			if d := pushData.GetAssetData(); d != nil {
				cb.OnAsset(d)
			}
		}
	case pb.SocketCommon_Position:
		if cb.OnPosition != nil {
			if d := pushData.GetPositionData(); d != nil {
				cb.OnPosition(d)
			}
		}
	case pb.SocketCommon_OrderStatus:
		if cb.OnOrder != nil {
			if d := pushData.GetOrderStatusData(); d != nil {
				cb.OnOrder(d)
			}
		}
	case pb.SocketCommon_OrderTransaction:
		if cb.OnTransaction != nil {
			if d := pushData.GetOrderTransactionData(); d != nil {
				cb.OnTransaction(d)
			}
		}
	case pb.SocketCommon_StockTop:
		if cb.OnStockTop != nil {
			if d := pushData.GetStockTopData(); d != nil {
				cb.OnStockTop(d)
			}
		}
	case pb.SocketCommon_OptionTop:
		if cb.OnOptionTop != nil {
			if d := pushData.GetOptionTopData(); d != nil {
				cb.OnOptionTop(d)
			}
		}
	case pb.SocketCommon_Kline:
		if cb.OnKline != nil {
			if d := pushData.GetKlineData(); d != nil {
				cb.OnKline(d)
			}
		}
	case pb.SocketCommon_Cc:
		// Cc(加密货币)推送复用 QuoteData body,路由到 OnQuote 回调。
		if cb.OnQuote != nil {
			if d := pushData.GetQuoteData(); d != nil {
				cb.OnQuote(d)
			}
		}
	default:
		if cb.OnError != nil {
			cb.OnError(fmt.Errorf("未知的 DataType: %v", pushData.DataType))
		}
	}
}

// subscribe 内部订阅方法
func (c *PushClient) subscribe(subject SubjectType, symbols []string, account string, market string) error {
	symbolsStr := strings.Join(symbols, ",")
	req := BuildSubscribeMessage(SubjectToDataType(subject), symbolsStr, account, market)
	return c.sendMessage(req)
}

// unsubscribe 内部退订方法
func (c *PushClient) unsubscribe(subject SubjectType, symbols []string) error {
	symbolsStr := strings.Join(symbols, ",")
	req := BuildUnSubscribeMessage(SubjectToDataType(subject), symbolsStr, "", "")
	return c.sendMessage(req)
}

// SubscribeQuote 订阅行情
func (c *PushClient) SubscribeQuote(symbols []string) error {
	if err := c.subscribe(SubjectQuote, symbols, "", ""); err != nil {
		return err
	}
	c.addSubscription(SubjectQuote, symbols)
	return nil
}

// UnsubscribeQuote 退订行情
func (c *PushClient) UnsubscribeQuote(symbols []string) error {
	if err := c.unsubscribe(SubjectQuote, symbols); err != nil {
		return err
	}
	c.removeSubscription(SubjectQuote, symbols)
	return nil
}

// SubscribeTick 订阅逐笔成交
func (c *PushClient) SubscribeTick(symbols []string) error {
	if err := c.subscribe(SubjectTick, symbols, "", ""); err != nil {
		return err
	}
	c.addSubscription(SubjectTick, symbols)
	return nil
}

// UnsubscribeTick 退订逐笔成交
func (c *PushClient) UnsubscribeTick(symbols []string) error {
	if err := c.unsubscribe(SubjectTick, symbols); err != nil {
		return err
	}
	c.removeSubscription(SubjectTick, symbols)
	return nil
}

// SubscribeDepth 订阅深度行情
func (c *PushClient) SubscribeDepth(symbols []string) error {
	if err := c.subscribe(SubjectDepth, symbols, "", ""); err != nil {
		return err
	}
	c.addSubscription(SubjectDepth, symbols)
	return nil
}

// UnsubscribeDepth 退订深度行情
func (c *PushClient) UnsubscribeDepth(symbols []string) error {
	if err := c.unsubscribe(SubjectDepth, symbols); err != nil {
		return err
	}
	c.removeSubscription(SubjectDepth, symbols)
	return nil
}

// SubscribeOption 订阅期权行情
func (c *PushClient) SubscribeOption(symbols []string) error {
	if err := c.subscribe(SubjectOption, symbols, "", ""); err != nil {
		return err
	}
	c.addSubscription(SubjectOption, symbols)
	return nil
}

// UnsubscribeOption 退订期权行情
func (c *PushClient) UnsubscribeOption(symbols []string) error {
	if err := c.unsubscribe(SubjectOption, symbols); err != nil {
		return err
	}
	c.removeSubscription(SubjectOption, symbols)
	return nil
}

// SubscribeFuture 订阅期货行情
func (c *PushClient) SubscribeFuture(symbols []string) error {
	if err := c.subscribe(SubjectFuture, symbols, "", ""); err != nil {
		return err
	}
	c.addSubscription(SubjectFuture, symbols)
	return nil
}

// UnsubscribeFuture 退订期货行情
func (c *PushClient) UnsubscribeFuture(symbols []string) error {
	if err := c.unsubscribe(SubjectFuture, symbols); err != nil {
		return err
	}
	c.removeSubscription(SubjectFuture, symbols)
	return nil
}

// SubscribeKline 订阅 K 线
func (c *PushClient) SubscribeKline(symbols []string) error {
	if err := c.subscribe(SubjectKline, symbols, "", ""); err != nil {
		return err
	}
	c.addSubscription(SubjectKline, symbols)
	return nil
}

// UnsubscribeKline 退订 K 线
func (c *PushClient) UnsubscribeKline(symbols []string) error {
	if err := c.unsubscribe(SubjectKline, symbols); err != nil {
		return err
	}
	c.removeSubscription(SubjectKline, symbols)
	return nil
}

// SubscribeAsset 订阅资产变动
func (c *PushClient) SubscribeAsset(account string) error {
	if account == "" {
		account = c.config.Account
	}
	if err := c.subscribe(SubjectAsset, nil, account, ""); err != nil {
		return err
	}
	c.mu.Lock()
	c.accountSubs[SubjectAsset] = true
	c.mu.Unlock()
	return nil
}

// UnsubscribeAsset 退订资产变动
func (c *PushClient) UnsubscribeAsset() error {
	if err := c.unsubscribe(SubjectAsset, nil); err != nil {
		return err
	}
	c.mu.Lock()
	delete(c.accountSubs, SubjectAsset)
	c.mu.Unlock()
	return nil
}

// SubscribePosition 订阅持仓变动
func (c *PushClient) SubscribePosition(account string) error {
	if account == "" {
		account = c.config.Account
	}
	if err := c.subscribe(SubjectPosition, nil, account, ""); err != nil {
		return err
	}
	c.mu.Lock()
	c.accountSubs[SubjectPosition] = true
	c.mu.Unlock()
	return nil
}

// UnsubscribePosition 退订持仓变动
func (c *PushClient) UnsubscribePosition() error {
	if err := c.unsubscribe(SubjectPosition, nil); err != nil {
		return err
	}
	c.mu.Lock()
	delete(c.accountSubs, SubjectPosition)
	c.mu.Unlock()
	return nil
}

// SubscribeOrder 订阅订单状态
func (c *PushClient) SubscribeOrder(account string) error {
	if account == "" {
		account = c.config.Account
	}
	if err := c.subscribe(SubjectOrder, nil, account, ""); err != nil {
		return err
	}
	c.mu.Lock()
	c.accountSubs[SubjectOrder] = true
	c.mu.Unlock()
	return nil
}

// UnsubscribeOrder 退订订单状态
func (c *PushClient) UnsubscribeOrder() error {
	if err := c.unsubscribe(SubjectOrder, nil); err != nil {
		return err
	}
	c.mu.Lock()
	delete(c.accountSubs, SubjectOrder)
	c.mu.Unlock()
	return nil
}

// SubscribeTransaction 订阅成交明细
func (c *PushClient) SubscribeTransaction(account string) error {
	if account == "" {
		account = c.config.Account
	}
	if err := c.subscribe(SubjectTransaction, nil, account, ""); err != nil {
		return err
	}
	c.mu.Lock()
	c.accountSubs[SubjectTransaction] = true
	c.mu.Unlock()
	return nil
}

// UnsubscribeTransaction 退订成交明细
func (c *PushClient) UnsubscribeTransaction() error {
	if err := c.unsubscribe(SubjectTransaction, nil); err != nil {
		return err
	}
	c.mu.Lock()
	delete(c.accountSubs, SubjectTransaction)
	c.mu.Unlock()
	return nil
}

// GetSubscriptions 获取当前订阅状态
func (c *PushClient) GetSubscriptions() map[SubjectType][]string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[SubjectType][]string)
	for subject, symbols := range c.subscriptions {
		list := make([]string, 0, len(symbols))
		for s := range symbols {
			list = append(list, s)
		}
		result[subject] = list
	}
	return result
}

// GetAccountSubscriptions 获取账户级别订阅状态
func (c *PushClient) GetAccountSubscriptions() []SubjectType {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]SubjectType, 0, len(c.accountSubs))
	for subject := range c.accountSubs {
		result = append(result, subject)
	}
	return result
}

// addSubscription 添加订阅记录
func (c *PushClient) addSubscription(subject SubjectType, symbols []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.subscriptions[subject] == nil {
		c.subscriptions[subject] = make(map[string]bool)
	}
	for _, s := range symbols {
		c.subscriptions[subject][s] = true
	}
}

// removeSubscription 移除订阅记录
func (c *PushClient) removeSubscription(subject SubjectType, symbols []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if symbols == nil {
		// 退订全部
		delete(c.subscriptions, subject)
		return
	}
	if m, ok := c.subscriptions[subject]; ok {
		for _, s := range symbols {
			delete(m, s)
		}
		if len(m) == 0 {
			delete(c.subscriptions, subject)
		}
	}
}

// ============================================================================
// Batch 6: stock_top / option_top / cc / market 订阅
// ============================================================================

// SubscribeStockTop 订阅股票榜单行情。indicators 是 indicator 字符串列表（作为 symbols 字段传输）。
func (c *PushClient) SubscribeStockTop(market string, indicators []string) error {
	if err := c.subscribe(SubjectStockTop, indicators, "", market); err != nil {
		return err
	}
	c.addSubscription(SubjectStockTop, indicators)
	return nil
}

// UnsubscribeStockTop 退订股票榜单行情。
func (c *PushClient) UnsubscribeStockTop(market string, indicators []string) error {
	symbolsStr := strings.Join(indicators, ",")
	req := BuildUnSubscribeMessage(SubjectToDataType(SubjectStockTop), symbolsStr, "", market)
	if err := c.sendMessage(req); err != nil {
		return err
	}
	c.removeSubscription(SubjectStockTop, indicators)
	return nil
}

// SubscribeOptionTop 订阅期权榜单行情。
func (c *PushClient) SubscribeOptionTop(market string, indicators []string) error {
	if err := c.subscribe(SubjectOptionTop, indicators, "", market); err != nil {
		return err
	}
	c.addSubscription(SubjectOptionTop, indicators)
	return nil
}

// UnsubscribeOptionTop 退订期权榜单行情。
func (c *PushClient) UnsubscribeOptionTop(market string, indicators []string) error {
	symbolsStr := strings.Join(indicators, ",")
	req := BuildUnSubscribeMessage(SubjectToDataType(SubjectOptionTop), symbolsStr, "", market)
	if err := c.sendMessage(req); err != nil {
		return err
	}
	c.removeSubscription(SubjectOptionTop, indicators)
	return nil
}

// SubscribeCc 订阅数字货币行情。
func (c *PushClient) SubscribeCc(symbols []string) error {
	if err := c.subscribe(SubjectCc, symbols, "", ""); err != nil {
		return err
	}
	c.addSubscription(SubjectCc, symbols)
	return nil
}

// UnsubscribeCc 退订数字货币行情。symbols 为空则退订全部。
func (c *PushClient) UnsubscribeCc(symbols []string) error {
	if err := c.unsubscribe(SubjectCc, symbols); err != nil {
		return err
	}
	c.removeSubscription(SubjectCc, symbols)
	return nil
}

// SubscribeMarket 订阅市场状态推送（dataType=Quote + market）。
func (c *PushClient) SubscribeMarket(market string) error {
	// 服务端约定：订阅 market 时 symbols 为空、只带 market 字段。
	if err := c.subscribe(SubjectMarket, nil, "", market); err != nil {
		return err
	}
	c.addSubscription(SubjectMarket, []string{market})
	return nil
}

// UnsubscribeMarket 退订市场状态推送。
func (c *PushClient) UnsubscribeMarket(market string) error {
	req := BuildUnSubscribeMessage(SubjectToDataType(SubjectMarket), "", "", market)
	if err := c.sendMessage(req); err != nil {
		return err
	}
	c.removeSubscription(SubjectMarket, []string{market})
	return nil
}
