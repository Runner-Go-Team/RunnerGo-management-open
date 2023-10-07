package mao

type DebugMsg struct {
	Msg                  string    `json:"msg" bson:"msg"`
	UUID                 string    `json:"uuid" bson:"uuid"`
	Type                 string    `json:"type" bson:"type"`
	ApiId                string    `json:"api_id" bson:"api_id"`
	Regex                RegexObj  `json:"regex" bson:"regex"`
	PlanId               string    `json:"plan_id" bson:"plan_id"`
	Method               string    `json:"method" bson:"method"`
	CaseId               string    `json:"case_id" bson:"case_id"`
	TeamId               string    `json:"team_id" bson:"team_id"`
	Status               string    `json:"status" bson:"status"`
	IsStop               bool      `json:"is_stop" bson:"is_stop"`
	Assert               AssertObj `json:"assert" bson:"assert"`
	SceneId              string    `json:"scene_id" bson:"scene_id"`
	ApiName              string    `json:"api_name" bson:"api_name"`
	EventId              string    `json:"event_id" bson:"event_id"`
	ParentId             string    `json:"parent_id" bson:"parent_id"`
	ReportId             string    `json:"report_id" bson:"report_id"`
	NextList             []string  `json:"next_list" bson:"next_list"`
	AssertNum            int       `json:"assert_num" bson:"assert_num"`
	RequestUrl           string    `json:"request_url" bson:"request_url"`
	RequestType          string    `json:"request_type" bson:"request_type"`
	RequestBody          string    `json:"request_body" bson:"request_body"`
	RequestTime          uint64    `json:"request_time" bson:"request_time"`
	RequestCode          int       `json:"request_code" bson:"request_code"`
	ResponseTime         string    `json:"response_time" bson:"response_time"`
	ResponseBody         string    `json:"response_body" bson:"response_body"`
	RequestHeader        string    `json:"request_header" bson:"request_header"`
	ResponseBytes        float64   `json:"response_bytes" bson:"response_bytes"`
	ResponseHeader       string    `json:"response_header" bson:"response_header"`
	AssertFailedNum      int       `json:"assert_failed_num" bson:"assert_failed_num"`
	RequestParameterType string    `json:"request_parameter_type" bson:"request_parameter_type"`
	ResponseMessageType  int       `json:"response_message_type" bson:"response_message_type"`
}

type AssertObj struct {
	AssertionMsgs []AssertionMsg `json:"assertion_msgs" bson:"assertion_msgs"`
}

type AssertionMsg struct {
	Type      string `json:"type" bson:"type"`
	Code      int64  `json:"code" bson:"code"`
	IsSucceed bool   `json:"is_succeed" bson:"is_succeed"`
	Msg       string `json:"msg" bson:"msg"`
}

type RegexObj struct {
	Regs []Reg `json:"regs" bson:"regs"`
}

type Reg struct {
	Key   string      `json:"key" bson:"key"`
	Value interface{} `json:"value" bson:"value"`
}

// debug日志状态
type DebugStatus struct {
	Debug    string `json:"debug" bson:"debug"`
	TeamID   string `json:"team_id" bson:"team_id"`
	ReportID string `json:"report_id" bson:"report_id"`
	PlanID   string `json:"plan_id" bson:"plan_id"`
}
