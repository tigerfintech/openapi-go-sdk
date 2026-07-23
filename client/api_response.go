package client

import (
	"encoding/json"
	"fmt"
)

// ApiResponse API 响应结构体
type ApiResponse struct {
	// Code 状态码（0=成功）
	Code int `json:"code"`
	// Message 状态描述
	Message string `json:"message"`
	// Data 业务数据（原始 JSON）
	Data json.RawMessage `json:"data"`
	// Timestamp 服务器时间戳
	Timestamp int64 `json:"timestamp"`
	// Sign response signature from server
	Sign string `json:"sign,omitempty"`
}

// ParseApiResponse 解析 API 响应 JSON 字符串
// 当 code 为 0 时返回 ApiResponse；当 code 不为 0 时返回 TigerError
func ParseApiResponse(body []byte) (*ApiResponse, error) {
	var resp ApiResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("解析响应 JSON 失败: %w", err)
	}
	if resp.Code != 0 {
		return &resp, NewTigerError(resp.Code, resp.Message)
	}
	return &resp, nil
}

// UnmarshalData 把 ApiResponse.Data 解码到 out。
// 兼容服务端偶发的双重编码(data 是 JSON 字符串包裹的 JSON)场景。
func UnmarshalData(data json.RawMessage, out interface{}) error {
	if len(data) == 0 || string(data) == "null" {
		return nil
	}
	// 尝试直接解码
	firstErr := json.Unmarshal(data, out)
	if firstErr == nil {
		return nil
	}
	// 服务端偶尔把 data 编码成 JSON 字符串(如 trade 的部分接口)。先解一层字符串。
	var asStr string
	if err := json.Unmarshal(data, &asStr); err == nil {
		return json.Unmarshal([]byte(asStr), out)
	}
	// 双层解码均失败，返回第一次的原始错误（语义最清晰）
	return firstErr
}
