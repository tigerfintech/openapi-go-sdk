package model

// FinancialDailyRequest 日级财务数据请求
type FinancialDailyRequest struct {
	Symbols   []string `json:"symbols"`
	Market    string   `json:"market"`
	Fields    []string `json:"fields"`
	BeginDate string   `json:"begin_date"`
	EndDate   string   `json:"end_date"`
}

// FinancialReportRequest 财报数据请求
type FinancialReportRequest struct {
	Symbols    []string `json:"symbols"`
	Market     string   `json:"market"`
	Fields     []string `json:"fields"`
	PeriodType string   `json:"period_type"`
	BeginDate  string   `json:"begin_date,omitempty"`
	EndDate    string   `json:"end_date,omitempty"`
}

// CorporateActionRequest 公司行动请求
type CorporateActionRequest struct {
	Symbols    []string `json:"symbols"`
	Market     string   `json:"market"`
	ActionType string   `json:"action_type"`
	BeginDate  string   `json:"begin_date,omitempty"`
	EndDate    string   `json:"end_date,omitempty"`
}

// FutureKlineRequest 期货 K 线请求
type FutureKlineRequest struct {
	ContractCodes []string `json:"contract_codes"`
	Period        string   `json:"period"`
	BeginTime     int64    `json:"begin_time"`
	EndTime       int64    `json:"end_time"`
	Limit         int      `json:"limit,omitempty"`
	PageToken     string   `json:"page_token,omitempty"`
}

// MarketScannerRequest 选股扫描请求
type MarketScannerRequest struct {
	Market               string                   `json:"market"`
	Page                 int                      `json:"page,omitempty"`
	PageSize             int                      `json:"page_size,omitempty"`
	CursorID             string                   `json:"cursor_id,omitempty"`
	BaseFilterList       []map[string]interface{} `json:"base_filter_list,omitempty"`
	AccumulateFilterList []map[string]interface{} `json:"accumulate_filter_list,omitempty"`
	FinancialFilterList  []map[string]interface{} `json:"financial_filter_list,omitempty"`
	MultiTagsFilterList  []map[string]interface{} `json:"multi_tags_filter_list,omitempty"`
	SortFieldData        map[string]interface{}   `json:"sort_field_data,omitempty"`
	MultiTagsFields      []string                 `json:"multi_tags_fields,omitempty"`
}
