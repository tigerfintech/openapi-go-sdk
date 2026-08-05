package model

import (
	"encoding/json"
	"testing"
)

// 测试 NewRangeFloat64 构造函数
func TestNewRangeFloat64(t *testing.T) {
	r := NewRangeFloat64(1.5, 3.5)
	if r.Min == nil {
		t.Fatal("Min 不应为 nil")
	}
	if r.Max == nil {
		t.Fatal("Max 不应为 nil")
	}
	if *r.Min != 1.5 {
		t.Errorf("Min = %f, want 1.5", *r.Min)
	}
	if *r.Max != 3.5 {
		t.Errorf("Max = %f, want 3.5", *r.Max)
	}
}

// 测试 NewRangeFloat64 序列化为 {"min":...,"max":...}
func TestNewRangeFloat64_json(t *testing.T) {
	r := NewRangeFloat64(0.5, 0.95)
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if v, ok := m["min"].(float64); !ok || v != 0.5 {
		t.Errorf("min = %v, want 0.5", m["min"])
	}
	if v, ok := m["max"].(float64); !ok || v != 0.95 {
		t.Errorf("max = %v, want 0.95", m["max"])
	}
}

// 测试 NewRangeFloat64 零值
func TestNewRangeFloat64_zero(t *testing.T) {
	r := NewRangeFloat64(0, 0)
	if r.Min == nil || *r.Min != 0 {
		t.Errorf("Min = %v, want 0", r.Min)
	}
	if r.Max == nil || *r.Max != 0 {
		t.Errorf("Max = %v, want 0", r.Max)
	}
}

// 测试 NewRangeFloat64 负值
func TestNewRangeFloat64_negative(t *testing.T) {
	r := NewRangeFloat64(-1.0, -0.1)
	if *r.Min != -1.0 {
		t.Errorf("Min = %f, want -1.0", *r.Min)
	}
	if *r.Max != -0.1 {
		t.Errorf("Max = %f, want -0.1", *r.Max)
	}
}

// 测试 NewRangeInt 构造函数
func TestNewRangeInt(t *testing.T) {
	r := NewRangeInt(10, 100)
	if r.Min == nil {
		t.Fatal("Min 不应为 nil")
	}
	if r.Max == nil {
		t.Fatal("Max 不应为 nil")
	}
	if *r.Min != 10 {
		t.Errorf("Min = %d, want 10", *r.Min)
	}
	if *r.Max != 100 {
		t.Errorf("Max = %d, want 100", *r.Max)
	}
}

// 测试 NewRangeInt 序列化为 {"min":...,"max":...}
func TestNewRangeInt_json(t *testing.T) {
	r := NewRangeInt(5, 50)
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if v, ok := m["min"].(float64); !ok || v != 5 {
		t.Errorf("min = %v, want 5", m["min"])
	}
	if v, ok := m["max"].(float64); !ok || v != 50 {
		t.Errorf("max = %v, want 50", m["max"])
	}
}

// 测试 NewRangeInt 零值
func TestNewRangeInt_zero(t *testing.T) {
	r := NewRangeInt(0, 0)
	if r.Min == nil || *r.Min != 0 {
		t.Errorf("Min = %v, want 0", r.Min)
	}
	if r.Max == nil || *r.Max != 0 {
		t.Errorf("Max = %v, want 0", r.Max)
	}
}

// 测试 NewRangeInt 负值
func TestNewRangeInt_negative(t *testing.T) {
	r := NewRangeInt(-5, -1)
	if *r.Min != -5 {
		t.Errorf("Min = %d, want -5", *r.Min)
	}
	if *r.Max != -1 {
		t.Errorf("Max = %d, want -1", *r.Max)
	}
}
