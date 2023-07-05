package rao

type ListUnderwayReportReq struct {
	TeamID string `form:"team_id" binding:"required,gt=0"`
	Page   int    `form:"page,default=1"`
	Size   int    `form:"size,default=10"`
}

type ListUnderwayReportResp struct {
	Reports []*StressPlanReport `json:"reports"`
	Total   int64               `json:"total"`
}

type ListReportsReq struct {
	TeamID string `form:"team_id" binding:"required,gt=0"`
	Page   int    `form:"page,default=1"`
	Size   int    `form:"size,default=10"`

	Keyword      string `form:"keyword"`
	TaskType     int32  `form:"task_type"`
	TaskMode     int32  `form:"task_mode"`
	Status       int32  `form:"status"`
	StartTimeSec int64  `form:"start_time_sec"`
	EndTimeSec   int64  `form:"end_time_sec"`
	Sort         int32  `form:"sort"`
}

type ListReportsResp struct {
	Reports []*StressPlanReport `json:"reports"`
	Total   int64               `json:"total"`
}

type StressPlanReport struct {
	ReportID    string `json:"report_id"`
	ReportName  string `json:"report_name"`
	RankID      int64  `json:"rank_id"`
	TeamID      string `json:"team_id"`
	TaskMode    int32  `json:"task_mode"`
	TaskType    int32  `json:"task_type"`
	Status      int32  `json:"status"`
	RunTimeSec  int64  `json:"run_time_sec"`
	LastTimeSec int64  `json:"last_time_sec"`
	RunUserID   string `json:"run_user_id"`
	RunUserName string `json:"run_user_name"`
	PlanID      string `json:"plan_id"`
	PlanName    string `json:"plan_name"`
	SceneID     string `json:"scene_id"`
	SceneName   string `json:"scene_name"`
}

type DeleteReportReq struct {
	TeamID   string `json:"team_id"`
	ReportID string `json:"report_id"`
}

type DeleteReportResp struct {
}

type StopReportReq struct {
	TeamID    string   `json:"team_id"`
	PlanID    string   `json:"plan_id"`
	ReportIDs []string `json:"report_ids"`
}

type StopReportResp struct {
}

type ListMachineReq struct {
	ReportID string `form:"report_id" binding:"required,gt=0" json:"report_id"`
	TeamID   string `form:"team_id" json:"team_id"`
	PlanID   string `form:"plan_id" json:"plan_id"`
}

type ListMachineResp struct {
	StartTimeSec int64    `json:"start_time_sec"`
	EndTimeSec   int64    `json:"end_time_sec"`
	ReportStatus int32    `json:"report_status"`
	Metrics      []Metric `json:"metrics"`
}

type Metric struct {
	IP          string          `json:"ip"`
	MachineName string          `json:"machine_name"`
	CPU         [][]interface{} `json:"cpu"`
	Mem         [][]interface{} `json:"mem"`
	NetIO       [][]interface{} `json:"net_io"`
	DiskIO      [][]interface{} `json:"disk_io"`
}

type GetReportReq struct {
	TeamID   string `form:"team_id" json:"team_id"`
	PlanID   string `form:"plan_id" json:"plan_id"`
	ReportID string `form:"report_id" json:"report_id"`
}

type GetReportResp struct {
	Report *ReportTask `json:"report"`
}

type GetReportTaskDetailReq struct {
	TeamID   string `form:"team_id" json:"team_id"`
	ReportID string `form:"report_id" json:"report_id"`
}

type GetReportTaskDetailResp struct {
	Report *ReportTask `json:"report"`
}

type ReportTask struct {
	UserID            string           `json:"user_id"`
	UserName          string           `json:"user_name"`
	UserAvatar        string           `json:"user_avatar"`
	PlanID            string           `json:"plan_id"`
	PlanName          string           `json:"plan_name"`
	SceneID           string           `json:"scene_id"`
	SceneName         string           `json:"scene_name"`
	ReportID          string           `json:"report_id"`
	ReportName        string           `json:"report_name"`
	CreatedTimeSec    int64            `json:"created_time_sec"`
	TaskType          int32            `json:"task_type"`
	TaskMode          int32            `json:"task_mode"`
	ControlMode       int32            `json:"control_mode"` // 控制模式
	DebugMode         string           `json:"debug_mode"`
	TaskStatus        int32            `json:"task_status"`
	ModeConf          ModeConf         `json:"mode_conf"`
	ChangeTakeConf    []ChangeTakeConf `json:"change_take_conf"`
	IsOpenDistributed int32            `json:"is_open_distributed"` // 是否开启分布式调度：0-关闭，1-开启
	MachineAllotType  int32            `json:"machine_allot_type"`  // 机器分配方式：0-权重，1-自定义
}

type DebugSettingReq struct {
	ReportID string `json:"report_id"`
	PlanID   string `json:"plan_id"`
	TeamID   string `json:"team_id"`
	Setting  string `json:"setting"`
}

type ReportEmailReq struct {
	TeamID   string   `json:"team_id"`
	ReportID string   `json:"report_id"`
	Emails   []string `json:"emails"`
}

type ReportEmailResp struct {
}

type ChangeTaskConfReq struct {
	ReportID                string                  `json:"report_id"`
	PlanID                  string                  `json:"plan_id"`
	TeamID                  string                  `json:"team_id"`
	ModeConf                *ModeConf               `json:"mode_conf"`
	IsOpenDistributed       int32                   `json:"is_open_distributed"`        // 是否开启分布式调度：0-关闭，1-开启
	MachineDispatchModeConf MachineDispatchModeConf `json:"machine_dispatch_mode_conf"` // 分布式压力机配置
}

type CompareReportReq struct {
	TeamID    string   `json:"team_id"`
	ReportIDs []string `json:"report_ids"`
}

type UpdateDescriptionReq struct {
	ReportID    string `json:"report_id"`
	TeamID      string `json:"team_id"`
	Description string `json:"description"`
}

type BatchDeleteReportReq struct {
	ReportIDs []string `json:"report_ids"`
	TeamID    string   `json:"team_id"`
}

type UpdateReportNameReq struct {
	ReportID   string `json:"report_id"`
	ReportName string `json:"report_name"`
}
