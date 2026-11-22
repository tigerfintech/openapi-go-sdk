package trade

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tigerfintech/openapi-go-sdk/model"
)

func TestTradingSessionTypePlaceAndPreview(t *testing.T) {
	for _, account := range []string{"600021175417", "20220101000000001"} {
		for _, session := range []string{"PRE_RTH", "RTH_POST", ""} {
			t.Run(account+"/"+session, func(t *testing.T) {
				var captured map[string]interface{}
				var method string
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					var body struct {
						Method  string `json:"method"`
						Content string `json:"biz_content"`
					}
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Error(err)
					}
					if err := json.Unmarshal([]byte(body.Content), &captured); err != nil {
						t.Error(err)
					}
					method = body.Method
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "message": "success", "data": map[string]interface{}{}})
				}))
				defer server.Close()
				tc := newTestTradeClient(server.URL)
				tc.account = account
				order := model.OrderRequest{Symbol: "AAPL", SecType: "OPT", Action: "BUY", OrderType: "LMT",
					TotalQuantity: 1, LimitPrice: 100, OutsideRth: true, TradingSessionType: session}
				for _, wantMethod := range []string{"place_order", "preview_order"} {
					var err error
					if wantMethod == "place_order" {
						_, err = tc.PlaceOrder(order)
					} else {
						_, err = tc.PreviewOrder(order)
					}
					if err != nil {
						t.Fatal(err)
					}
					if method != wantMethod {
						t.Fatalf("method = %s, want %s", method, wantMethod)
					}
					got, present := captured["trading_session_type"]
					if session == "" && present {
						t.Fatal("unset session must be omitted")
					}
					if session != "" && got != session {
						t.Fatalf("session = %v, want %s", got, session)
					}
					if _, exists := captured["tradingSessionType"]; exists {
						t.Fatal("camelCase request field must not be sent")
					}
				}
			})
		}
	}
}
