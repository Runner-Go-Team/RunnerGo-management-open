package rao

type UIPlanReport struct {
	ReportID      string `json:"report_id"`
	ReportName    string `json:"report_name"`
	RankID        int64  `json:"rank_id"`
	TeamID        string `json:"team_id"`
	TaskType      int32  `json:"task_type"`
	Status        int32  `json:"status"`
	SceneRunOrder int32  `json:"scene_run_order"`
	RunTimeSec    int64  `json:"run_time_sec"`
	LastTimeSec   int64  `json:"last_time_sec"`
	RunUserID     string `json:"run_user_id"`
	RunUserName   string `json:"run_user_name"`
	PlanID        string `json:"plan_id"`
	PlanName      string `json:"plan_name"`
}

type UIReportDetail struct {
	RunUserID          string           `json:"run_user_id"`
	RunUserName        string           `json:"run_user_name"`
	UserAvatar         string           `json:"user_avatar"`
	PlanID             string           `json:"plan_id"`
	PlanName           string           `json:"plan_name"`
	ReportID           string           `json:"report_id"`
	ReportName         string           `json:"report_name"`
	CreatedTimeSec     int64            `json:"created_time_sec"`
	EndTimeSec         int64            `json:"end_time_sec"`
	TaskType           int32            `json:"task_type"`
	SceneRunOrder      int32            `json:"scene_run_order"`
	Status             int32            `json:"status"`
	RunDurationTime    int64            `json:"run_duration_time"` // 报告运行时长
	RunTimeSec         int64            `json:"run_time_sec"`
	LastTimeSec        int64            `json:"last_time_sec"`
	SceneTotalNum      int              `json:"scene_total_num"`      // 场景总数
	SceneSuccessNum    int              `json:"scene_success_num"`    // 场景成功数
	SceneErrorNum      int              `json:"scene_error_num"`      // 场景失败数
	SceneUnExecNum     int              `json:"scene_un_exec_num"`    // 场景未执行数
	OperatorTotalNum   int              `json:"operator_total_num"`   // 场景操作总数
	OperatorSuccessNum int              `json:"operator_success_num"` // 场景操作成功数
	OperatorErrorNum   int              `json:"operator_error_num"`   // 场景操作失败数
	OperatorUnExecNum  int              `json:"operator_un_exec_num"` // 场景操作未执行数
	AssertTotalNum     int              `json:"assert_total_num"`     // 断言总数
	AssertSuccessNum   int              `json:"assert_success_num"`   // 断言成功数
	AssertErrorNum     int              `json:"assert_error_num"`     // 断言失败数
	AssertUnExecNum    int              `json:"assert_un_exec_num"`
	Scenes             []*UIReportScene `json:"scenes"`
}

type UIReportScene struct {
	SceneID            string                   `json:"scene_id" bson:"scene_id"`
	Name               string                   `json:"name" bson:"name"`
	RunStatus          int                      `json:"run_status"` // 1:未测  2:成功   3:失败
	OperatorTotalNum   int                      `json:"operator_total_num" bson:"operator_total_num"`
	OperatorSuccessNum int                      `json:"operator_success_num" bson:"operator_success_num"`
	OperatorErrorNum   int                      `json:"operator_error_num" bson:"operator_error_num"`
	OperatorUnExecNum  int                      `json:"operator_un_exec_num" bson:"operator_un_exec_num"`
	AssertTotalNum     int                      `json:"assert_total_num" bson:"assert_total_num"`
	AssertSuccessNum   int                      `json:"assert_success_num" bson:"assert_success_num"`
	AssertErrorNum     int                      `json:"assert_error_num" bson:"assert_error_num"`
	AssertUnExecNum    int                      `json:"assert_un_exec_num" bson:"assert_un_exec_num"`
	Operators          []*UIReportSceneOperator `json:"operators"`
}

type UIReportSceneOperator struct {
	SceneID        string                   `json:"scene_id" bson:"scene_id"`
	SceneName      string                   `json:"scene_name" bson:"scene_name"`
	OperatorID     string                   `json:"operator_id" bson:"operator_id"`
	ParentID       string                   `json:"parent_id" bson:"parent_id"`
	Name           string                   `json:"name" bson:"name"`
	Sort           int32                    `json:"sort" bson:"sort"`
	Type           string                   `json:"type" bson:"type"`
	Action         string                   `json:"action" bson:"action"`
	RunStatus      int32                    `json:"run_status" bson:"run_status"` // 1:未测 2:成功  3:失败
	ExecTime       float64                  `json:"exec_time" bson:"exec_time"`   //  运行时长	RunEndTimes    int64                    `json:"run_end_times" bson:"run_end_times"` // 运行结束时间
	Status         string                   `json:"status" bson:"status"`         // 状态
	Msg            string                   `json:"msg" bson:"msg"`
	Screenshot     string                   `json:"screenshot" bson:"screenshot"`
	ScreenshotUrl  string                   `json:"screenshot_url" bson:"screenshot_url"`
	End            bool                     `json:"end" bson:"end"`
	AssertTotalNum int                      `json:"assert_total_num" bson:"assert_total_num"`
	IsMulti        bool                     `json:"is_multi" bson:"is_multi"` // 是否展示多条
	MultiResult    []*UIEngineResultDataMsg `json:"multi_result"`             // 多条数据结果
	Assertions     []*UIEngineAssertion     `json:"assertions"`
	Withdraws      []*UIEngineDataWithdraw  `json:"withdraws"`
}

type ListUIReportsReq struct {
	TeamID string `json:"team_id" binding:"required,gt=0"`
	Page   int    `json:"page,default=1"`
	Size   int    `json:"size,default=10"`

	ReportName    string `json:"report_name"`
	PlanName      string `json:"plan_name"`
	TaskType      int32  `json:"task_type"`
	Status        int32  `json:"status"`
	SceneRunOrder int32  `json:"scene_run_order"`
	RunUserID     string `json:"run_user_id"`
	StartTime     string `json:"start_time"`
	EndTime       string `json:"end_time"`
	Sort          int32  `json:"sort"`
}

type ListUIReportsResp struct {
	Reports []*UIPlanReport `json:"reports"`
	Total   int64           `json:"total"`
}

type UIReportDetailReq struct {
	TeamID   string `form:"team_id" json:"team_id"`
	ReportID string `form:"report_id" json:"report_id"`
}

type UIReportDetailResp struct {
	Report *UIReportDetail `json:"report"`
}

type UIReportDeleteReq struct {
	ReportIDs []string `json:"report_ids"`
	TeamID    string   `json:"team_id"`
}

type StopUIReportReq struct {
	TeamID   string `json:"team_id" binding:"required,gt=0"`
	ReportID string `json:"report_id" binding:"required,gt=0"`
}

type UIReportUpdateReq struct {
	ReportID   string `json:"report_id"`
	TeamID     string `json:"team_id"`
	ReportName string `json:"report_name"`
}

type UIEngineResultDataMsg struct {
	TopicID       string                   `json:"topic"`
	UserID        string                   `json:"user_id"`
	OperatorID    string                   `json:"operator_id"`                  // 操作ID
	SceneID       string                   `json:"scene_id"`                     // 场景
	Sort          int32                    `json:"sort"`                         // 步骤
	ExecTime      float64                  `json:"exec_time"`                    // 执行时间
	Status        string                   `json:"status"`                       // 状态
	RunStatus     int32                    `json:"run_status" bson:"run_status"` // 1:未测 2:成功  3:失败
	Msg           string                   `json:"msg"`
	Screenshot    string                   `json:"screenshot"`
	ScreenshotUrl string                   `json:"screenshot_url"`
	End           bool                     `json:"end"`
	Assertions    []*UIEngineAssertion     `json:"assertions"`
	Withdraws     []*UIEngineDataWithdraw  `json:"withdraws"`
	IsMulti       bool                     `json:"is_multi"`     // 是否展示多条
	IsReport      bool                     `json:"is_report"`    // 是否是报告
	MultiResult   []*UIEngineResultDataMsg `json:"multi_result"` // 多条数据结果
}

type RunUIReportReq struct {
	ReportID string `json:"report_id"`
	TeamID   string `json:"team_id"`
}
