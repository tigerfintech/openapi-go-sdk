package client

import (
	"encoding/json"
	"testing"
)

// TestNewApiRequest_WithString 测试使用字符串创建 API 请求
func TestNewApiRequest_WithString(t *testing.T) {
	req, err := NewApiRequest("market_state", `{"market":"US"}`)
	if err != nil {
		t.Fatalf("创建请求失败: %v", err)
	}
	if req.Method != "market_state" {
		t.Errorf("期望 method=market_state，实际为 %s", req.Method)
	}
	if req.BizContent != `{"market":"US"}` {
		t.Errorf("期望 biz_content={\"market\":\"US\"}，实际为 %s", req.BizContent)
	}
}

// TestNewApiRequest_WithStruct 测试使用结构体创建 API 请求
func TestNewApiRequest_WithStruct(t *testing.T) {
	params := map[string]string{"market": "US", "symbol": "AAPL"}
	req, err := NewApiRequest("brief", params)
	if err != nil {
		t.Fatalf("创建请求失败: %v", err)
	}
	if req.Method != "brief" {
		t.Errorf("期望 method=brief，实际为 %s", req.Method)
	}
	// 验证 biz_content 是有效 JSON
	if !json.Valid([]byte(req.BizContent)) {
		t.Errorf("biz_content 不是有效 JSON: %s", req.BizContent)
	}
	// 反序列化验证内容
	var parsed map[string]string
	json.Unmarshal([]byte(req.BizContent), &parsed)
	if parsed["market"] != "US" {
		t.Errorf("market 不匹配: %s", parsed["market"])
	}
	if parsed["symbol"] != "AAPL" {
		t.Errorf("symbol 不匹配: %s", parsed["symbol"])
	}
}

// TestNewApiRequest_WithNil 测试使用 nil 创建 API 请求
func TestNewApiRequest_WithNil(t *testing.T) {
	req, err := NewApiRequest("market_state", nil)
	if err != nil {
		t.Fatalf("创建请求失败: %v", err)
	}
	if req.BizContent != "{}" {
		t.Errorf("nil 参数应生成空 JSON 对象，实际为 %s", req.BizContent)
	}
}

// TestUnmarshalData_NoThirdRetry 验证 UnmarshalData 失败时返回第一次错误，不会对 out 第三次调用 Unmarshal
func TestUnmarshalData_NoThirdRetry(t *testing.T) {
	// out 类型与 data 完全不兼容（data 是 JSON 数字，out 是 struct）
	type Foo struct{ Name string }
	var out Foo
	data := json.RawMessage(`12345`) // 数字，无法反序列化为 struct 也无法作为字符串层
	err := UnmarshalData(data, &out)
	if err == nil {
		t.Fatal("期望返回错误，但得到 nil")
	}
	// out 不应被污染（Foo{} 的零值）
	if out.Name != "" {
		t.Errorf("out 不应被修改，Name=%q", out.Name)
	}
}

// TestUnmarshalData_DoubleEncoded 验证双重编码（data 是 JSON string 包裹 JSON object）正确解码
func TestUnmarshalData_DoubleEncoded(t *testing.T) {
	type Inner struct{ OrderID int64 }
	inner := Inner{OrderID: 12345678}
	innerJSON, _ := json.Marshal(inner)
	// 把 inner JSON 再编码为 JSON string
	outerEncoded, _ := json.Marshal(string(innerJSON))
	data := json.RawMessage(outerEncoded)

	var out Inner
	if err := UnmarshalData(data, &out); err != nil {
		t.Fatalf("双重编码解码失败: %v", err)
	}
	if out.OrderID != 12345678 {
		t.Errorf("OrderID 期望 12345678，实际 %d", out.OrderID)
	}
}

// TestUnmarshalData_NullData 验证 null data 返回 nil
func TestUnmarshalData_NullData(t *testing.T) {
	var out map[string]interface{}
	if err := UnmarshalData(json.RawMessage(`null`), &out); err != nil {
		t.Fatalf("null data 应返回 nil 错误，实际 %v", err)
	}
}
