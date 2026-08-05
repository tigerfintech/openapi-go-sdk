package model

import (
	"encoding/json"
	"testing"
)

// === NormalizeOrderStatus 直接测试（覆盖各类型分支）===

// 测试 NormalizeOrderStatus 接收 string 直接返回
func TestNormalizeOrderStatus_stringPassesThrough(t *testing.T) {
	if got := NormalizeOrderStatus("Filled"); got != "Filled" {
		t.Errorf("string passthrough: got %q, want Filled", got)
	}
}

// 测试 NormalizeOrderStatus 接收 int
func TestNormalizeOrderStatus_int(t *testing.T) {
	cases := []struct {
		in   int
		want string
	}{
		{-2, "Invalid"},
		{-1, "Initial"},
		{3, "PendingCancel"},
		{4, "Cancelled"},
		{5, "Submitted"},
		{6, "Filled"},
		{7, "Inactive"},
		{8, "PendingSubmit"},
	}
	for _, c := range cases {
		if got := NormalizeOrderStatus(c.in); got != c.want {
			t.Errorf("int %d: got %q, want %q", c.in, got, c.want)
		}
	}
}

// 测试 NormalizeOrderStatus 接收 float64
func TestNormalizeOrderStatus_float64(t *testing.T) {
	if got := NormalizeOrderStatus(float64(6)); got != "Filled" {
		t.Errorf("float64 6: got %q, want Filled", got)
	}
	if got := NormalizeOrderStatus(float64(-2)); got != "Invalid" {
		t.Errorf("float64 -2: got %q, want Invalid", got)
	}
}

// 测试 NormalizeOrderStatus 接收 int64
func TestNormalizeOrderStatus_int64(t *testing.T) {
	if got := NormalizeOrderStatus(int64(5)); got != "Submitted" {
		t.Errorf("int64 5: got %q, want Submitted", got)
	}
	if got := NormalizeOrderStatus(int64(-1)); got != "Initial" {
		t.Errorf("int64 -1: got %q, want Initial", got)
	}
}

// 测试 NormalizeOrderStatus 接收 json.Number
func TestNormalizeOrderStatus_jsonNumber(t *testing.T) {
	n := json.Number("6")
	if got := NormalizeOrderStatus(n); got != "Filled" {
		t.Errorf("json.Number 6: got %q, want Filled", got)
	}
}

// 测试 NormalizeOrderStatus 接收无法解析的 json.Number
func TestNormalizeOrderStatus_jsonNumberUnparseable_returnsString(t *testing.T) {
	n := json.Number("not-a-number")
	got := NormalizeOrderStatus(n)
	if got != "not-a-number" {
		t.Errorf("json.Number unparseable: got %q, want not-a-number", got)
	}
}

// 测试 NormalizeOrderStatus 接收未知整数返回空字符串
func TestNormalizeOrderStatus_unknownInt_returnsEmpty(t *testing.T) {
	if got := NormalizeOrderStatus(999); got != "" {
		t.Errorf("unknown int 999: got %q, want empty", got)
	}
}

// 测试 NormalizeOrderStatus 接收未知类型返回空字符串
func TestNormalizeOrderStatus_unknownType_returnsEmpty(t *testing.T) {
	if got := NormalizeOrderStatus([]int{1, 2}); got != "" {
		t.Errorf("unknown type: got %q, want empty", got)
	}
	if got := NormalizeOrderStatus(nil); got != "" {
		t.Errorf("nil: got %q, want empty", got)
	}
	if got := NormalizeOrderStatus(true); got != "" {
		t.Errorf("bool: got %q, want empty", got)
	}
}

// 测试 NormalizeOrderStatus 接收未知字符串返回原值
func TestNormalizeOrderStatus_unknownString_returnsAsIs(t *testing.T) {
	if got := NormalizeOrderStatus("PendingNew"); got != "PendingNew" {
		t.Errorf("unknown string: got %q, want PendingNew", got)
	}
}

// === OrderStatus.Code 默认分支测试 ===

// 测试 OrderStatus.Code 未知状态返回 0
func TestOrderStatus_Code_unknownReturnsZero(t *testing.T) {
	if got := OrderStatus("PendingNew").Code(); got != 0 {
		t.Errorf("unknown status Code = %d, want 0", got)
	}
	if got := OrderStatus("").Code(); got != 0 {
		t.Errorf("empty status Code = %d, want 0", got)
	}
	if got := OrderStatus("Unknown").Code(); got != 0 {
		t.Errorf("Unknown status Code = %d, want 0", got)
	}
}

// === Order.UnmarshalJSON 错误路径测试 ===

// 测试 Order.UnmarshalJSON 外层解码失败（字段类型不匹配）
func TestOrderUnmarshalJSON_invalidJSON_returnsError(t *testing.T) {
	// 合法 JSON，但 id 字段类型不匹配 int64，使内层 json.Unmarshal 失败
	raw := `{"id": "not-a-number", "status": 6}`
	var o Order
	err := json.Unmarshal([]byte(raw), &o)
	if err == nil {
		t.Fatal("类型不匹配 JSON 应返回错误，got nil")
	}
}

// 测试 Order.UnmarshalJSON status 字段为 null 时保持空状态
func TestOrderUnmarshalJSON_statusNull_returnsEmpty(t *testing.T) {
	raw := `{"id": 1, "status": null}`
	var o Order
	if err := json.Unmarshal([]byte(raw), &o); err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}
	if o.ID != 1 {
		t.Errorf("ID = %d, want 1", o.ID)
	}
	if o.Status != "" {
		t.Errorf("Status = %q, want empty (null status)", o.Status)
	}
}

// 测试 Order.UnmarshalJSON 缺少 status 字段时保持空状态
func TestOrderUnmarshalJSON_missingStatus_returnsEmpty(t *testing.T) {
	raw := `{"id": 42, "symbol": "AAPL"}`
	var o Order
	if err := json.Unmarshal([]byte(raw), &o); err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}
	if o.ID != 42 {
		t.Errorf("ID = %d, want 42", o.ID)
	}
	if o.Status != "" {
		t.Errorf("Status = %q, want empty (no status field)", o.Status)
	}
}
