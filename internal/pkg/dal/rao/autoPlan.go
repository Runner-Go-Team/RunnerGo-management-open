package rao

type RunAutoPlanReq struct {
	PlanID  string   `json:"plan_id"`
	TeamID  string   `json:"team_id"`
	SceneID []string `json:"scene_id"`
}

type RunAutoPlanResp struct {
	TaskType int32 `json:"task_type"`
}

type SaveAutoPlanReq struct {
	PlanName string `json:"plan_name" binding:"required"`
	Remark   string `json:"remark"`
	TeamID   string `json:"team_id" binding:"required"`
}

type SaveAutoPlanResp struct {
	PlanID string `json:"plan_id"`
}

type GetAutoPlanListReq struct {
	PlanName     string `json:"plan_name"`
	TeamId       string `json:"team_id"`
	TaskType     int32  `json:"task_type"`
	Status       int32  `json:"status"`
	StartTimeSec int64  `json:"start_time_sec"`
	EndTimeSec   int64  `json:"end_time_sec"`

	Page int   `json:"page" form:"page,default=1"`
	Size int   `json:"size" form:"size,default=10"`
	Sort int32 `json:"sort" form:"sort"`
}

type AutoPlanListResp struct {
	AutoPlanList []*AutoPlanDetailResp `json:"auto_plan_list"`
	Total        int64                 `json:"total"`
}

type AutoPlanDetailResp struct {
	RankID    int64  `json:"rank_id"`
	PlanID    string `json:"plan_id"`
	TeamID    string `json:"team_id"`
	PlanName  string `json:"plan_name"`
	TaskType  int32  `json:"task_type"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
	Status    int32  `json:"status"`
	UserName  string `json:"user_name"`
	Remark    string `json:"remark"`
}

type DeleteAutoPlanReq struct {
	PlanID string `json:"plan_id"`
	TeamID string `json:"team_id"`
}

type GetAutoPlanDetailReq struct {
	PlanID string `json:"plan_id"`
	TeamID string `json:"team_id"`
}

type GetAutoPlanDetailResp struct {
	PlanID    string `json:"plan_id"`
	TeamID    string `json:"team_id"`
	PlanName  string `json:"plan_name"`
	TaskType  int32  `json:"task_type"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
	Status    int32  `json:"status"`
	UserName  string `json:"user_name"`
	Remark    string `json:"remark"`
	Avatar    string `json:"avatar"`
}

type CopyAutoPlanReq struct {
	PlanID string `json:"plan_id" binding:"required"`
	TeamID string `json:"team_id" binding:"required"`
}

type UpdateAutoPlanReq struct {
	PlanID   string `json:"plan_id" binding:"required"`
	TeamID   string `json:"team_id" binding:"required"`
	PlanName string `json:"plan_name" binding:"required"`
	Remark   string `json:"remark"`
}

type AddEmailReq struct {
	PlanID string   `json:"plan_id" binding:"required"`
	TeamID string   `json:"team_id" binding:"required"`
	Emails []string `json:"emails" binding:"required"`
}

type GetEmailListReq struct {
	PlanID       string `json:"plan_id" binding:"required"`
	TeamID       string `json:"team_id" binding:"required"`
	PlanCategory int32  `json:"plan_category,default=1"` // 默认自动化测试
}

type GetEmailListResp struct {
	Emails []*AutoPlanEmail `json:"emails"`
}

type AutoPlanEmail struct {
	ID           int64  `json:"id"`
	PlanID       string `json:"plan_id"`
	TeamID       string `json:"team_id"`
	PlanCategory int32  `json:"plan_category"` // 默认自动化测试
	Email        string `json:"email"`
}

type DeleteEmailReq struct {
	PlanID string `json:"plan_id" binding:"required"`
	TeamID string `json:"team_id" binding:"required"`
	ID     int64  `json:"id" binding:"required"`
}

type BatchDeleteAutoPlanReq struct {
	PlanIDs []string `json:"plan_ids" binding:"required"`
	TeamID  string   `json:"team_id" binding:"required"`
}

type SaveTaskConfReq struct {
	PlanID           string `json:"plan_id" binding:"required"`
	TeamID           string `json:"team_id" binding:"required"`
	TaskType         int32  `json:"task_type" binding:"required"`
	TaskMode         int32  `json:"task_mode" binding:"required"`
	SceneRunOrder    int32  `json:"scene_run_order" binding:"required"`
	TestCaseRunOrder int32  `json:"test_case_run_order" binding:"required"`
	Frequency        int32  `json:"frequency"`       // 频次: 0-一次，1-每天，2-每周，3-每月
	TaskExecTime     int64  `json:"task_exec_time"`  // 任执行时间
	TaskCloseTime    int64  `json:"task_close_time"` // 结束时间
}

type GetTaskConfReq struct {
	PlanID string `json:"plan_id" binding:"required"`
	TeamID string `json:"team_id" binding:"required"`
}

type GetTaskConfResp struct {
	PlanID           string `json:"plan_id"`
	TeamID           string `json:"team_id"`
	TaskType         int32  `json:"task_type"`
	TaskMode         int32  `json:"task_mode"`
	SceneRunOrder    int32  `json:"scene_run_order"`
	TestCaseRunOrder int32  `json:"test_case_run_order"`
	Frequency        int32  `json:"frequency"`       // 频次: 0-一次，1-每天，2-每周，3-每月
	TaskExecTime     int64  `json:"task_exec_time"`  // 任执行时间
	TaskCloseTime    int64  `json:"task_close_time"` // 结束时间
	Status           int32  `json:"status"`
}

type GetAutoPlanReportListReq struct {
	PlanName     string `json:"plan_name"`
	TeamId       string `json:"team_id"`
	TaskType     int32  `json:"task_type"`
	Status       int32  `json:"status"`
	StartTimeSec int64  `json:"start_time_sec"`
	EndTimeSec   int64  `json:"end_time_sec"`

	Page int   `json:"page" form:"page,default=1"`
	Size int   `json:"size" form:"size,default=10"`
	Sort int32 `form:"sort"`
}

type GetAutoPlanReportList struct {
	RankID           int64  `json:"rank_id"`
	ReportID         string `json:"report_id"`
	PlanID           string `json:"plan_id"`
	TeamID           string `json:"team_id"`
	PlanName         string `json:"plan_name"`
	RunUserName      string `json:"run_user_name"`
	TaskType         int32  `json:"task_type"`
	TaskMode         int32  `json:"task_mode"`           // 运行模式：1-按测试用例运行
	SceneRunOrder    int32  `json:"scene_run_order"`     // 场景运行次序：1-顺序执行，2-同时执行
	TestCaseRunOrder int32  `json:"test_case_run_order"` // 测试用例运行次序：1-顺序执行，2-同时执行
	Status           int32  `json:"status"`
	StartTimeSec     int64  `json:"start_time_sec"`
	EndTimeSec       int64  `json:"end_time_sec"`
	Remark           string `json:"remark"`
}

type GetAutoPlanReportListResp struct {
	AutoPlanReportList []*GetAutoPlanReportList `json:"auto_plan_report_list"`
	Total              int64                    `json:"total"`
}

type BatchDeleteAutoPlanReportReq struct {
	ReportIDs []string `json:"report_ids"`
	TeamID    string   `json:"team_id"`
}

type CloneAutoPlanSceneReq struct {
	TeamID  string `json:"team_id" binding:"required"`
	SceneID string `json:"scene_id" binding:"required"`
	PlanID  string `json:"plan_id"`
	Source  int32  `json:"source"`
}

type NotifyRunFinishReq struct {
	TeamID          string `json:"team_id" binding:"required"`
	PlanID          string `json:"plan_id" binding:"required"`
	ReportID        string `json:"report_id" binding:"required"`
	RunDurationTime int64  `json:"run_duration_time"`
}

type StopAutoPlanReq struct {
	TeamID string `json:"team_id" binding:"required"`
	PlanID string `json:"plan_id" binding:"required"`
}

type GetAutoPlanReportDetailReq struct {
	ReportID        string `json:"report_id"`
	TeamID          string `json:"team_id"`
	RunDurationTime int64  `json:"run_duration_time"`
	UpdatedAt       int64  `json:"updated_at"`
}

type ReportEmailNotifyReq struct {
	TeamID   string   `json:"team_id"`
	ReportID string   `json:"report_id"`
	Emails   []string `json:"emails"`
}
