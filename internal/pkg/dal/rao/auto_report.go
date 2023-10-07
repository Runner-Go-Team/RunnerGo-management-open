package rao

type GetReportDetailResp struct {
	PlanID               string                      `json:"plan_id" bson:"plan_id"`
	PlanName             string                      `json:"plan_name" bson:"plan_name"`
	ReportName           string                      `json:"report_name" bson:"report_name"`
	Avatar               string                      `json:"avatar" bson:"avatar"`
	Nickname             string                      `json:"nickname" bson:"nickname"`
	Remark               string                      `json:"remark" bson:"remark"`
	TaskMode             int32                       `json:"task_mode" bson:"task_mode"`
	SceneRunOrder        int32                       `json:"scene_run_order" bson:"scene_run_order"`
	TestCaseRunOrder     int32                       `json:"test_case_run_order" bson:"test_case_run_order"`
	ReportStatus         int32                       `json:"report_status" bson:"report_status"`
	ReportStartTime      int64                       `json:"report_start_time" bson:"report_start_time"`
	ReportEndTime        int64                       `json:"report_end_time" bson:"report_end_time"`
	ReportRunTime        int64                       `json:"report_run_time" bson:"report_run_time"`
	SceneBaseInfo        SceneBaseInfo               `json:"scene_base_info" bson:"scene_base_info"`
	CaseBaseInfo         CaseBaseInfo                `json:"case_base_info" bson:"case_base_info"`
	ApiBaseInfo          ApiBaseInfo                 `json:"api_base_info" bson:"api_base_info"`
	AssertionBaseInfo    AssertionBaseInfo           `json:"assertion_base_info" bson:"assertion_base_info"`
	SceneResult          []SceneResult               `json:"scene_result" bson:"scene_result"`
	SceneIDCaseResultMap map[string][]TestCaseResult `json:"scene_id_case_result_map" bson:"scene_id_case_result_map"`
}

type SceneBaseInfo struct {
	SceneTotalNum int64 `json:"scene_total_num" bson:"scene_total_num"`
}

type CaseBaseInfo struct {
	CaseTotalNum int64 `json:"case_total_num" bson:"case_total_num"`
	SucceedNum   int64 `json:"succeed_num" bson:"succeed_num"`
	FailNum      int64 `json:"fail_num" bson:"fail_num"`
}

type ApiBaseInfo struct {
	ApiTotalNum int64 `json:"api_total_num" bson:"api_total_num"`
	SucceedNum  int64 `json:"succeed_num" bson:"succeed_num"`
	FailNum     int64 `json:"fail_num" bson:"fail_num"`
	NotTestNum  int64 `json:"not_test_num" bson:"not_test_num"`
}

type AssertionBaseInfo struct {
	AssertionTotalNum int64 `json:"assertion_total_num" bson:"assertion_total_num"`
	SucceedNum        int64 `json:"succeed_num" bson:"succeed_num"`
	FailNum           int64 `json:"fail_num" bson:"fail_num"`
}

// SceneResult 场景结果
type SceneResult struct {
	SceneID      string `json:"scene_id" bson:"scene_id"`
	SceneName    string `json:"scene_name" bson:"scene_name"`
	CaseFailNum  int    `json:"case_fail_num" bson:"case_fail_num"`
	CaseTotalNum int    `json:"case_total_num" bson:"case_total_num"`
	State        int    `json:"state" bson:"state"` // 1-成功，2-失败
}

type TestCaseResult struct {
	CaseName   string    `json:"case_name" bson:"case_name"`
	CaseID     string    `json:"case_id" bson:"case_id"`
	SucceedNum int64     `json:"succeed_num" bson:"succeed_num"`
	TotalNum   int64     `json:"total_num" bson:"total_num"`
	ApiList    []ApiList `json:"api_list" bson:"api_list"`
}
type ApiList struct {
	EventID        string    `json:"event_id" bson:"event_id"`
	TargetID       string    `json:"target_id" bson:"target_id"`
	CaseID         string    `json:"case_id" bson:"case_id"`
	ApiName        string    `json:"api_name" bson:"api_name"`
	Method         string    `json:"method" bson:"method"`
	Url            string    `json:"url" bson:"url"`
	Status         string    `json:"status" bson:"status"`
	ResponseBytes  float64   `json:"response_bytes" bson:"response_bytes"`
	RequestTime    int64     `json:"request_time" bson:"request_time"`
	RequestCode    int32     `json:"request_code" bson:"request_code"`
	RequestHeader  string    `json:"request_header" bson:"request_header"`
	RequestBody    string    `json:"request_body" bson:"request_body"`
	ResponseHeader string    `json:"response_header" bson:"response_header"`
	ResponseBody   string    `json:"response_body" bson:"response_body"`
	AssertionMsg   AssertObj `json:"assert" bson:"assert"`
	RegexMsg       RegexObj  `json:"regex" bson:"regex"`
}

type AssertionMsg struct {
	Type      string `json:"type" bson:"type"`
	Code      int64  `json:"code" bson:"code"`
	IsSucceed bool   `json:"is_succeed" bson:"is_succeed"`
	Msg       string `json:"msg" bson:"msg"`
}
