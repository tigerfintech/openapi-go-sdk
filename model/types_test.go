package model

import (
	"encoding/json"
	"testing"
)

// === FlexString 测试 ===

// 测试 FlexString 从字符串 JSON 反序列化
func TestFlexString_UnmarshalJSON_string(t *testing.T) {
	var f FlexString
	if err := json.Unmarshal([]byte(`"abc123"`), &f); err != nil {
		t.Fatalf("反序列化字符串失败: %v", err)
	}
	if f != "abc123" {
		t.Errorf("FlexString = %q, want abc123", f)
	}
}

// 测试 FlexString 从整数 JSON 反序列化（服务端不一致返回数字）
func TestFlexString_UnmarshalJSON_int(t *testing.T) {
	var f FlexString
	if err := json.Unmarshal([]byte(`42`), &f); err != nil {
		t.Fatalf("反序列化整数失败: %v", err)
	}
	if f != "42" {
		t.Errorf("FlexString = %q, want 42", f)
	}
}

// 测试 FlexString 从浮点数 JSON 反序列化
func TestFlexString_UnmarshalJSON_float(t *testing.T) {
	var f FlexString
	if err := json.Unmarshal([]byte(`3.14`), &f); err != nil {
		t.Fatalf("反序列化浮点数失败: %v", err)
	}
	if f != "3.14" {
		t.Errorf("FlexString = %q, want 3.14", f)
	}
}

// 测试 FlexString 从负数 JSON 反序列化
func TestFlexString_UnmarshalJSON_negativeNumber(t *testing.T) {
	var f FlexString
	if err := json.Unmarshal([]byte(`-1`), &f); err != nil {
		t.Fatalf("反序列化负数失败: %v", err)
	}
	if f != "-1" {
		t.Errorf("FlexString = %q, want -1", f)
	}
}

// 测试 FlexString 无效 JSON 返回错误
func TestFlexString_UnmarshalJSON_invalid_returnsError(t *testing.T) {
	var f FlexString
	err := json.Unmarshal([]byte(`true`), &f)
	if err == nil {
		t.Fatal("无效 JSON 应返回错误，got nil")
	}
}

// 测试 FlexString 从空字符串 JSON 反序列化
func TestFlexString_UnmarshalJSON_emptyString(t *testing.T) {
	var f FlexString
	if err := json.Unmarshal([]byte(`""`), &f); err != nil {
		t.Fatalf("反序列化空字符串失败: %v", err)
	}
	if f != "" {
		t.Errorf("FlexString = %q, want empty", f)
	}
}

// 测试 FlexString.String() 返回字符串值
func TestFlexString_String(t *testing.T) {
	f := FlexString("hello")
	if f.String() != "hello" {
		t.Errorf("String() = %q, want hello", f.String())
	}
}

// 测试 FlexString.String() 空值
func TestFlexString_String_empty(t *testing.T) {
	f := FlexString("")
	if f.String() != "" {
		t.Errorf("String() = %q, want empty", f.String())
	}
}

// 测试 FlexString 在 struct 中的 round-trip（数字场景）
func TestFlexString_structRoundTrip(t *testing.T) {
	type sample struct {
		Level FlexString `json:"level"`
	}
	// 服务端返回数字
	in := `{"level": 5}`
	var s sample
	if err := json.Unmarshal([]byte(in), &s); err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}
	if s.Level.String() != "5" {
		t.Errorf("Level = %q, want 5", s.Level.String())
	}
}

// === FlexInt64 测试 ===

// 测试 FlexInt64 从数字 JSON 反序列化
func TestFlexInt64_UnmarshalJSON_number(t *testing.T) {
	var f FlexInt64
	if err := json.Unmarshal([]byte(`12345`), &f); err != nil {
		t.Fatalf("反序列化数字失败: %v", err)
	}
	if f.Int64() != 12345 {
		t.Errorf("Int64() = %d, want 12345", f.Int64())
	}
}

// 测试 FlexInt64 从引号字符串 JSON 反序列化（服务端不一致返回字符串）
func TestFlexInt64_UnmarshalJSON_quotedString(t *testing.T) {
	var f FlexInt64
	if err := json.Unmarshal([]byte(`"67890"`), &f); err != nil {
		t.Fatalf("反序列化字符串失败: %v", err)
	}
	if f.Int64() != 67890 {
		t.Errorf("Int64() = %d, want 67890", f.Int64())
	}
}

// 测试 FlexInt64 从负数 JSON 反序列化
func TestFlexInt64_UnmarshalJSON_negativeNumber(t *testing.T) {
	var f FlexInt64
	if err := json.Unmarshal([]byte(`-100`), &f); err != nil {
		t.Fatalf("反序列化负数失败: %v", err)
	}
	if f.Int64() != -100 {
		t.Errorf("Int64() = %d, want -100", f.Int64())
	}
}

// 测试 FlexInt64 从引号负数字符串反序列化
func TestFlexInt64_UnmarshalJSON_negativeQuotedString(t *testing.T) {
	var f FlexInt64
	if err := json.Unmarshal([]byte(`"-200"`), &f); err != nil {
		t.Fatalf("反序列化负数字符串失败: %v", err)
	}
	if f.Int64() != -200 {
		t.Errorf("Int64() = %d, want -200", f.Int64())
	}
}

// 测试 FlexInt64 从零反序列化
func TestFlexInt64_UnmarshalJSON_zero(t *testing.T) {
	var f FlexInt64
	if err := json.Unmarshal([]byte(`0`), &f); err != nil {
		t.Fatalf("反序列化零失败: %v", err)
	}
	if f.Int64() != 0 {
		t.Errorf("Int64() = %d, want 0", f.Int64())
	}
}

// 测试 FlexInt64 非数字字符串返回错误
func TestFlexInt64_UnmarshalJSON_nonNumericString_returnsError(t *testing.T) {
	var f FlexInt64
	err := json.Unmarshal([]byte(`"abc"`), &f)
	if err == nil {
		t.Fatal("非数字字符串应返回错误，got nil")
	}
}

// 测试 FlexInt64 无效 JSON（布尔值）返回错误
func TestFlexInt64_UnmarshalJSON_invalid_returnsError(t *testing.T) {
	var f FlexInt64
	err := json.Unmarshal([]byte(`true`), &f)
	if err == nil {
		t.Fatal("无效 JSON 应返回错误，got nil")
	}
}

// 测试 FlexInt64 JSON 对象命中最终 fallback 分支
func TestFlexInt64_UnmarshalJSON_object_returnsError(t *testing.T) {
	var f FlexInt64
	err := json.Unmarshal([]byte(`{}`), &f)
	if err == nil {
		t.Fatal("对象 JSON 应返回错误，got nil")
	}
}

// 测试 FlexInt64 JSON 数组命中最终 fallback 分支
func TestFlexInt64_UnmarshalJSON_array_returnsError(t *testing.T) {
	var f FlexInt64
	err := json.Unmarshal([]byte(`[]`), &f)
	if err == nil {
		t.Fatal("数组 JSON 应返回错误，got nil")
	}
}

// 测试 FlexInt64 null 反序列化（json.Number 分支返回错误因 Int64 失败）
func TestFlexInt64_UnmarshalJSON_null(t *testing.T) {
	var f FlexInt64
	// null 命中 json.Number 分支但 Int64() 失败，应返回错误
	err := json.Unmarshal([]byte(`null`), &f)
	if err == nil {
		t.Fatal("null 应返回错误，got nil")
	}
}

// 测试 FlexInt64 浮点数（小数）返回错误
func TestFlexInt64_UnmarshalJSON_float_returnsError(t *testing.T) {
	var f FlexInt64
	// UseNumber 解析为 json.Number 后 Int64() 应失败
	err := json.Unmarshal([]byte(`3.14`), &f)
	if err == nil {
		t.Fatal("浮点数应返回错误，got nil")
	}
}

// 测试 FlexInt64.Int64() 直接返回值
func TestFlexInt64_Int64(t *testing.T) {
	f := FlexInt64(99999)
	if f.Int64() != 99999 {
		t.Errorf("Int64() = %d, want 99999", f.Int64())
	}
}

// 测试 FlexInt64.Int64() 负值
func TestFlexInt64_Int64_negative(t *testing.T) {
	f := FlexInt64(-42)
	if f.Int64() != -42 {
		t.Errorf("Int64() = %d, want -42", f.Int64())
	}
}
