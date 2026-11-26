package quote

import (
	"encoding/json"
	"github.com/tigerfintech/openapi-go-sdk/model"
	"testing"
)

func TestGthKlineFlag(t *testing.T) {
	for _, flag := range []*bool{nil, boolPtr(true), boolPtr(false)} {
		var captured string
		server := mockServerCapture(t, "option_kline", []model.Kline{}, &captured)
		qc := newTestQuoteClient(server.URL)
		_, err := qc.GetOptionKlineWithHourTrading([]string{"SPXW  260723P07380000", "SPXW  260723C07380000"},
			"5min", -1, -1, 20, "desc", flag)
		server.Close()
		if err != nil {
			t.Fatal(err)
		}
		var body struct{ Queries []map[string]interface{} }
		var raw map[string]json.RawMessage
		if err := json.Unmarshal([]byte(captured), &raw); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(raw["option_query"], &body.Queries); err != nil {
			t.Fatal(err)
		}
		if len(body.Queries) != 2 {
			t.Fatalf("queries: %s", captured)
		}
		for _, query := range body.Queries {
			value, present := query["with_hour_trading"]
			if flag == nil && present {
				t.Fatal("omitted flag was sent")
			}
			if flag != nil && (!present || value != *flag) {
				t.Fatalf("wrong flag: %s", captured)
			}
		}
	}
}
func boolPtr(value bool) *bool { return &value }

func TestGthHTTPQuote(t *testing.T) {
	var brief model.Brief
	err := json.Unmarshal([]byte("{\"symbol\":\"SPXW\",\"cumVolume\":0,\"cumAmount\":0,\"hourTrading\":{\"session\":\"GTH\",\"latestTimestamp\":1784796380425,\"volume\":0}}"), &brief)
	if err != nil {
		t.Fatal(err)
	}
	if brief.CumVolume == nil || *brief.CumVolume != 0 || brief.HourTrading == nil ||
		brief.HourTrading.LatestTimestamp == nil || *brief.HourTrading.LatestTimestamp != 1784796380425 {
		t.Fatalf("extended quote was lost: %+v", brief)
	}
	var leg model.OptionLeg
	if err := json.Unmarshal([]byte("{\"hourTrading\":null}"), &leg); err != nil {
		t.Fatal(err)
	}
	if leg.HourTrading != nil {
		t.Fatal("null hourTrading should stay nil")
	}
}
