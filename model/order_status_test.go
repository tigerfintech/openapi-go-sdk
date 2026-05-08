package model

import (
	"encoding/json"
	"testing"
)

func TestOrderStatus_IntToString(t *testing.T) {
	cases := []struct {
		raw  string
		want string
	}{
		{`{"status": -2}`, "Invalid"},
		{`{"status": -1}`, "Initial"},
		{`{"status": 3}`, "PendingCancel"},
		{`{"status": 4}`, "Cancelled"},
		{`{"status": 5}`, "Submitted"},
		{`{"status": 6}`, "Filled"},
		{`{"status": 7}`, "Inactive"},
		{`{"status": 8}`, "PendingSubmit"},
	}
	for _, c := range cases {
		var o Order
		if err := json.Unmarshal([]byte(c.raw), &o); err != nil {
			t.Fatalf("%s: %v", c.raw, err)
		}
		if o.Status != c.want {
			t.Errorf("%s: got status=%q want %q", c.raw, o.Status, c.want)
		}
	}
}

func TestOrderStatus_StringPassthrough(t *testing.T) {
	cases := []string{"Filled", "Cancelled", "Inactive", "Invalid", "Submitted", "PendingSubmit", "PendingCancel", "Initial"}
	for _, want := range cases {
		raw := `{"status":"` + want + `"}`
		var o Order
		if err := json.Unmarshal([]byte(raw), &o); err != nil {
			t.Fatalf("%s: %v", raw, err)
		}
		if o.Status != want {
			t.Errorf("status passthrough: got %q want %q", o.Status, want)
		}
	}
}

func TestOrderStatus_Code(t *testing.T) {
	cases := []struct {
		s    OrderStatus
		code int
	}{
		{OrderStatusInvalid, -2},
		{OrderStatusInitial, -1},
		{OrderStatusPendingCancel, 3},
		{OrderStatusCancelled, 4},
		{OrderStatusSubmitted, 5},
		{OrderStatusFilled, 6},
		{OrderStatusInactive, 7},
		{OrderStatusPendingSubmit, 8},
	}
	for _, c := range cases {
		if got := c.s.Code(); got != c.code {
			t.Errorf("%s.Code() = %d, want %d", c.s, got, c.code)
		}
	}
}

func TestOrderStatus_PreservesOtherFields(t *testing.T) {
	raw := `{"id": 123, "orderId": 456, "symbol": "AAPL", "status": 6, "filledQuantity": 100}`
	var o Order
	if err := json.Unmarshal([]byte(raw), &o); err != nil {
		t.Fatal(err)
	}
	if o.ID != 123 || o.OrderId != 456 || o.Symbol != "AAPL" || o.FilledQuantity != 100 {
		t.Errorf("other fields lost: %+v", o)
	}
	if o.Status != "Filled" {
		t.Errorf("status: got %q want Filled", o.Status)
	}
}
