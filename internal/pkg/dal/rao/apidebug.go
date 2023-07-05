package rao

type APIDebug struct {
	ApiID                 string                   `json:"api_id"`
	APIName               string                   `json:"api_name"`
	Assert                []*DebugAssert           `json:"assert"`
	EventID               string                   `json:"event_id"`
	Regex                 []map[string]interface{} `json:"regex"`
	RequestBody           string                   `json:"request_body"`
	RequestCode           int64                    `json:"request_code"`
	RequestHeader         string                   `json:"request_header"`
	RequestTime           int64                    `json:"request_time"`
	ResponseBody          string                   `json:"response_body"`
	ResponseBytes         float64                  `json:"response_bytes"`
	ResponseHeader        string                   `json:"response_header"`
	ResponseTime          string                   `json:"response_time"`
	ResponseLen           int32                    `json:"response_len"`
	ResponseStatusMessage string                   `json:"response_status_message"`
	UUID                  string                   `json:"uuid"`
	Status                string                   `json:"status"`
}

type DebugAssert struct {
	Code      int    `json:"code"`
	IsSucceed bool   `json:"is_succeed"`
	Msg       string `json:"msg"`
}

type DebugRegex struct {
	Token string `json:"token,omitempty"`
	Code  string `json:"code,omitempty"`
}
