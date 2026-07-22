package model

import (
	"bytes"
	"encoding/json"
)

// NormalizeOrderStatus 将服务端可能返回的整数状态码归一化为字符串名称。
// 映射与 Java SDK OrderStatus 枚举一致:
//
//	-2: Invalid    -1: Initial    3: PendingCancel    4: Cancelled
//	 5: Submitted   6: Filled      7: Inactive        8: PendingSubmit
//
// 推送 / 查询接口混用的 status 字段,有时是 int,有时已经是字符串。
// 这里只做数字→字符串映射,不做跨别名合并(Submitted↔Held、Inactive↔REJECTED 等)。
func NormalizeOrderStatus(v interface{}) string {
	switch x := v.(type) {
	case string:
		return x
	case float64:
		return statusIntToString(int(x))
	case int:
		return statusIntToString(x)
	case int64:
		return statusIntToString(int(x))
	case json.Number:
		if i, err := x.Int64(); err == nil {
			return statusIntToString(int(i))
		}
		return x.String()
	}
	return ""
}

func statusIntToString(n int) string {
	switch n {
	case -2:
		return string(OrderStatusInvalid)
	case -1:
		return string(OrderStatusInitial)
	case 3:
		return string(OrderStatusPendingCancel)
	case 4:
		return string(OrderStatusCancelled)
	case 5:
		return string(OrderStatusSubmitted)
	case 6:
		return string(OrderStatusFilled)
	case 7:
		return string(OrderStatusInactive)
	case 8:
		return string(OrderStatusPendingSubmit)
	}
	return ""
}

// UnmarshalJSON 定制反序列化：把可能为整数的 status 字段统一转成字符串。
// 其他字段保持默认行为。
func (o *Order) UnmarshalJSON(data []byte) error {
	type orderAlias Order
	var aux struct {
		*orderAlias
		Status json.RawMessage `json:"status,omitempty"`
	}
	aux.orderAlias = (*orderAlias)(o)
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if len(aux.Status) == 0 || bytes.Equal(aux.Status, []byte("null")) {
		return nil
	}
	var raw interface{}
	dec := json.NewDecoder(bytes.NewReader(aux.Status))
	dec.UseNumber()
	if err := dec.Decode(&raw); err != nil {
		return err
	}
	o.Status = NormalizeOrderStatus(raw)
	return nil
}
