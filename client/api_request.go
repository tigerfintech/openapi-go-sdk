package client

import (
	"encoding/json"
	"fmt"
)

// ApiRequest API request structure
type ApiRequest struct {
	// Method API method name (e.g. "market_state", "place_order")
	Method string `json:"method"`
	// BizContent business parameters as JSON string
	BizContent string `json:"biz_content"`
	// Version API version (e.g. "2.0", "3.0"). Empty means use default.
	Version string `json:"version,omitempty"`
}

// NewApiRequest creates an API request, serializing business params to JSON as biz_content.
func NewApiRequest(method string, bizParams interface{}) (*ApiRequest, error) {
	var bizContent string
	switch v := bizParams.(type) {
	case string:
		bizContent = v
	case nil:
		bizContent = "{}"
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize biz_content: %w", err)
		}
		bizContent = string(data)
	}
	return &ApiRequest{
		Method:     method,
		BizContent: bizContent,
	}, nil
}

// NewVersionedApiRequest creates an API request with a specific API version.
func NewVersionedApiRequest(method string, bizParams interface{}, version string) (*ApiRequest, error) {
	req, err := NewApiRequest(method, bizParams)
	if err != nil {
		return nil, err
	}
	req.Version = version
	return req, nil
}
