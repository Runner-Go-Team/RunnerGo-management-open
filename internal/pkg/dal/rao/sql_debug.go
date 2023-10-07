package rao

type GetSendSqlResultResp struct {
	RetID        string    `json:"ret_id"`
	RequestTime  int64     `json:"request_time"`
	Status       string    `json:"status"`
	Assert       AssertObj `json:"assert"`
	Regex        RegexObj  `json:"regex"`
	ResponseBody string    `json:"response_body"`
}
