package rao

type DashboardDefaultReq struct {
	TeamID string `form:"team_id" binding:"required,gt=0"`
}

type DashboardDefaultResp struct {
	User   *Member `json:"user"`
	Mobile string  `json:"mobile"`
}

type HomePageReq struct {
	TeamID string `json:"team_id" form:"team_id" binding:"required,gt=0"`
}

type HomePageResp struct {
	TeamName         string          `json:"team_name"`
	ApiManageData    ApiManageData   `json:"api_manage_data"`
	SceneManageData  SceneManageData `json:"scene_manage_data"`
	CaseAddSevenData map[string]int  `json:"case_add_seven_data"`
	TeamOverview     []TeamOverview  `json:"team_overview"`
	AutoPlanData     AutoPlanData    `json:"auto_plan_data"`
	StressPlanData   StressPlanData  `json:"stress_plan_data"`
}

type ApiManageData struct {
	ApiCiteCount  int            `json:"api_cite_count"`
	ApiTotalCount int            `json:"api_total_count"`
	ApiDebugCount map[string]int `json:"api_debug"`
	ApiAddCount   map[string]int `json:"api_add_count"`
}
type SceneManageData struct {
	SceneCiteCount  int            `json:"scene_cite_count"`
	SceneTotalCount int            `json:"scene_total_count"`
	SceneDebugCount map[string]int `json:"scene_debug_count"`
	SceneAddCount   map[string]int `json:"scene_add_count"`
}

type TeamOverview struct {
	TeamName           string `json:"team_name"`
	AutoPlanTotalNum   int    `json:"auto_plan_total_num"`
	AutoPlanExecNum    int    `json:"auto_plan_exec_num"`
	StressPlanTotalNum int    `json:"stress_plan_total_num"`
	StressPlanExecNum  int    `json:"stress_plan_exec_num"`
}

type AutoPlanData struct {
	PlanNum                   int                `json:"plan_num"`
	ReportNum                 int64              `json:"report_num"`
	CaseTotalNum              int                `json:"case_total_num"`
	CaseExecNum               int                `json:"case_exec_num"`
	CasePassNum               int                `json:"case_pass_num"`
	CiteApiNum                int                `json:"cite_api_num"`
	TotalApiNum               int                `json:"total_api_num"`
	CiteSceneNum              int                `json:"cite_scene_num"`
	TotalSceneNum             int                `json:"total_scene_num"`
	CasePassPercent           float64            `json:"case_pass_percent"`
	CaseNotTestAndPassPercent float64            `json:"case_not_test_and_pass_percent"`
	LatelyReportList          []LatelyReportList `json:"lately_report_list"`
}

type LatelyReportList struct {
	ReportID    string `json:"report_id"`
	RankID      int64  `json:"rank_id"`
	PlanName    string `json:"plan_name"`
	TaskType    int32  `json:"task_type"`
	TaskMode    int32  `json:"task_mode"`
	RunUserName string `json:"run_user_name"`
	Status      int32  `json:"status"`
}

type StressPlanData struct {
	PlanNum          int                `json:"plan_num"`
	ReportNum        int                `json:"report_num"`
	ApiNum           int                `json:"api_num"`
	SceneNum         int                `json:"scene_num"`
	CiteApiNum       int                `json:"cite_api_num"`
	TotalApiNum      int                `json:"total_api_num"`
	CiteSceneNum     int                `json:"cite_scene_num"`
	TotalSceneNum    int                `json:"total_scene_num"`
	TimedPlanNum     float64            `json:"timed_plan_num"`
	NormalPlanNum    float64            `json:"normal_plan_num"`
	LatelyReportList []LatelyReportList `json:"lately_report_list"`
}
