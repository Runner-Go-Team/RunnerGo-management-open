package rao

type GetSendSqlResultResp struct {
	RetID        string                   `json:"ret_id"`
	RequestTime  int64                    `json:"request_time"`
	Status       string                   `json:"status"`
	Assert       []map[string]interface{} `json:"assert"`
	Regex        []map[string]interface{} `json:"regex"`
	ResponseBody interface{}              `json:"response_body"`
}
