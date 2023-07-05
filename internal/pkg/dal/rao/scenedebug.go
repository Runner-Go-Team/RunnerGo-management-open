package rao

type SceneDebug struct {
	ApiID          string                   `json:"api_id"`
	APIName        string                   `json:"api_name"`
	Assert         []*DebugAssert           `json:"assert"`
	EventID        string                   `json:"event_id"`
	NextList       []string                 `json:"next_list"`
	Regex          []map[string]interface{} `json:"regex"`
	RequestBody    string                   `json:"request_body"`
	RequestCode    int64                    `json:"request_code"`
	RequestHeader  string                   `json:"request_header"`
	ResponseBody   interface{}              `json:"response_body"`
	ResponseBytes  float64                  `json:"response_bytes"`
	ResponseHeader string                   `json:"response_header"`
	Status         string                   `json:"status"`
	UUID           string                   `json:"uuid"`
	ResponseTime   string                   `json:"response_time"`
	RequestType    string                   `json:"request_type"`
}
