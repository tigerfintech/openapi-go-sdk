package client

import (
	"net"
	"strings"
	"testing"
	"time"
)

// TestExecute_StaleConnectionEOF 验证服务端在建立连接后立即关闭（模拟 stale keep-alive 连接）时：
//  1. 行情接口（非交易）会重试
//  2. 下单接口（交易）不重试，日志提示"订单可能已提交"
func TestExecute_StaleConnectionEOF(t *testing.T) {
	// 启动一个 TCP 服务，接受连接后立即关闭，模拟服务端已关闭的 keep-alive 连接
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			conn.Close() // 立即关闭，触发客户端 EOF
		}
	}()

	addr := "http://" + ln.Addr().String()
	cfg := newTestConfig(t, addr)

	cl := &capturingLogger{}
	hc := NewHttpClient(cfg, WithLogger(cl))
	// 缩短重试策略：最多 2 次，退避 10ms，加快测试速度
	hc.retryPolicy = &RetryPolicy{
		MaxRetries: 2,
		BaseDelay:  10 * time.Millisecond,
		MaxDelay:   20 * time.Millisecond,
	}

	t.Run("非交易接口_触发重试", func(t *testing.T) {
		cl.entries = nil
		req := &ApiRequest{Method: "market_state", BizContent: `{}`}
		_, err := hc.Execute(req)
		if err == nil {
			t.Fatal("期望 EOF 错误，但没有")
		}

		// 非交易接口应有 WARN retry 日志，最终 ERROR exhausted
		var warnCount, errCount int
		for _, e := range cl.entries {
			if e.level == "WARN" && strings.Contains(e.msg, "retry") {
				warnCount++
			}
			if e.level == "ERROR" && strings.Contains(e.msg, "exhausted") {
				errCount++
			}
		}
		if warnCount == 0 {
			t.Errorf("期望至少 1 条 WARN retry 日志，实际: %v", cl.entries)
		}
		if errCount != 1 {
			t.Errorf("期望 1 条 ERROR exhausted 日志，实际: %v", cl.entries)
		}
		t.Logf("非交易 EOF 日志（共 %d 条）: %v", len(cl.entries), cl.entries)
	})

	t.Run("下单接口_不重试_日志含订单确认提示", func(t *testing.T) {
		cl.entries = nil
		req := &ApiRequest{Method: "place_order", BizContent: `{}`}
		_, err := hc.Execute(req)
		if err == nil {
			t.Fatal("期望 EOF 错误，但没有")
		}

		// 下单不重试：无 WARN retry，有 ERROR 含 "query order status"
		for _, e := range cl.entries {
			if e.level == "WARN" && strings.Contains(e.msg, "retry") {
				t.Errorf("下单不应有 retry 日志，但出现了: %s", e.msg)
			}
		}
		var found bool
		for _, e := range cl.entries {
			if e.level == "ERROR" && strings.Contains(e.msg, "query order status") {
				found = true
			}
		}
		if !found {
			t.Errorf("下单 EOF 时应提示 'query order status'，实际日志: %v", cl.entries)
		}
		t.Logf("下单 EOF 日志（共 %d 条）: %v", len(cl.entries), cl.entries)
	})
}

// TestExecute_HTTP500EmptyBody 验证服务端返回 HTTP 500 空 body 时错误信息包含状态码
func TestExecute_HTTP500EmptyBody(t *testing.T) {
	server := newHTTP500Server(t, 500, "")
	defer server.Close()

	cfg := newTestConfig(t, server.URL)
	cl := &capturingLogger{}
	hc := NewHttpClient(cfg, WithLogger(cl))

	req := &ApiRequest{Method: "market_state", BizContent: `{}`}
	_, err := hc.Execute(req)
	if err == nil {
		t.Fatal("期望错误，但没有")
	}

	// 错误信息应包含 HTTP 500
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("错误信息应包含 HTTP 500，实际: %v", err)
	}
	t.Logf("HTTP 500 空 body 错误: %v", err)
	t.Logf("日志: %v", cl.entries)
}

// TestExecute_HTTP500WithJSON 验证服务端返回 HTTP 500 + 合法 JSON body 时能解析出业务错误码
func TestExecute_HTTP500WithJSON(t *testing.T) {
	server := newHTTP500Server(t, 500, `{"code":500,"message":"internal server error","data":null,"timestamp":0}`)
	defer server.Close()

	cfg := newTestConfig(t, server.URL)
	cl := &capturingLogger{}
	hc := NewHttpClient(cfg, WithLogger(cl))

	req := &ApiRequest{Method: "market_state", BizContent: `{}`}
	_, err := hc.Execute(req)
	if err == nil {
		t.Fatal("期望错误，但没有")
	}

	// 应该解析出 TigerError code=500，或者包含 HTTP 500 上下文
	errMsg := err.Error()
	if !strings.Contains(errMsg, "500") {
		t.Errorf("错误应包含 500，实际: %v", err)
	}
	t.Logf("HTTP 500 JSON body 错误: %v", err)
	t.Logf("日志: %v", cl.entries)
}
