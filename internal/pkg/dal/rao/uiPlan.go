package rao

type UIPlan struct {
	PlanID            string     `json:"plan_id"`
	RankID            int64      `json:"rank_id"`
	TeamID            string     `json:"team_id"`
	Name              string     `json:"name"`
	TaskType          int32      `json:"task_type"`
	CreatedUserID     string     `json:"created_user_id"`
	CreatedUserName   string     `json:"created_user_name"`
	CreatedUserAvatar string     `json:"created_user_avatar"`
	HeadUserIDs       []string   `json:"head_user_ids"`
	Description       string     `json:"description"`
	InitStrategy      int32      `json:"init_strategy" binding:"required"` // 初始化策略：1-计划执行前重启浏览器，2-场景执行前重启浏览器，3-无初始化
	CreatedTimeSec    int64      `json:"created_time_sec"`
	UpdatedTimeSec    int64      `json:"updated_time_sec"`
	ModeConf          *ModeConf  `json:"mode_conf,omitempty"`
	Browsers          []*Browser `json:"browsers,omitempty"`
	UIMachineKey      string     `json:"ui_machine_key"`
	TimedStatus       int32      `json:"timed_status"`
}

type UIPlanHeadUser struct {
	HeadUserID     string `json:"head_user_id"`
	HeadUserName   string `json:"head_user_name"`
	HeadUserAvatar string `json:"head_user_avatar"`
}

type UIPlanSaveReq struct {
	PlanID       string     `json:"plan_id"`
	TeamID       string     `json:"team_id" binding:"required,gt=0"`
	Name         string     `json:"name" binding:"required"`
	Description  *string    `json:"description"`
	HeadUserIDs  []string   `json:"head_user_ids"`
	InitStrategy int32      `json:"init_strategy" binding:"required"` // 初始化策略：1-计划执行前重启浏览器，2-场景执行前重启浏览器，3-无初始化
	TaskType     int32      `json:"task_type"`
	RankID       int64      `json:"rank_id"`
	Browsers     []*Browser `json:"browsers,omitempty"`
	UIMachineKey string     `json:"ui_machine_key"`
}

type ListUIPlanReq struct {
	TeamID string `json:"team_id" binding:"required,gt=0"`
	Page   int    `json:"page,default=1"`
	Size   int    `json:"size,default=10"`

	Name            string `json:"name"`
	CreatedUserName string `json:"created_user_name"`
	HeadUserName    string `json:"head_user_name"`

	UpdatedTime []string `json:"updated_time"`
	CreatedTime []string `json:"created_time"`

	TaskType int32 `json:"task_type"`
	Sort     int32 `json:"sort"`
}

type ListUIPlanResp struct {
	Plans []*UIPlan `json:"plans"`
	Total int64     `json:"total"`
}

type UIPlanImportSceneReq struct {
	PlanID   string   `json:"plan_id"`
	TeamID   string   `json:"team_id"`
	SceneIDs []string `json:"scene_ids"`
	SyncMode int32    `json:"sync_mode"`
}

type UIPlanSetSceneSyncModeReq struct {
	PlanID       string `json:"plan_id"`
	TeamID       string `json:"team_id"`
	SceneID      string `json:"scene_id"`
	SyncMode     int32  `json:"sync_mode"`     // 1-实时，2-手动,已场景为准   3-手动,已计划为准
	TargetSource int32  `json:"target_source"` // 1:已场景为准  2:已计划为准
}

type UIPlanSetSceneHandSyncReq struct {
	PlanID  string `json:"plan_id"`
	TeamID  string `json:"team_id"`
	SceneID string `json:"scene_id"`
}

type UIPlanDetailReq struct {
	PlanID string `form:"plan_id"`
	TeamID string `form:"team_id"`
}

type UIPlanDetailResp struct {
	Plan *UIPlan `json:"plan"`
}

type UIPlanDeleteReq struct {
	PlanIDs []string `json:"plan_ids"`
	TeamID  string   `json:"team_id"`
}

type UIPlanCopyReq struct {
	PlanID string `json:"plan_id" binding:"required"`
	TeamID string `json:"team_id" binding:"required"`
}

type UIPlanSaveTaskConfReq struct {
	PlanID                 string `json:"plan_id" binding:"required"`
	TeamID                 string `json:"team_id" binding:"required"`
	TaskType               int32  `json:"task_type" binding:"required"`
	SceneRunOrder          int32  `json:"scene_run_order" binding:"required"`
	Frequency              int32  `json:"frequency"`                 // 频次: 0-一次，1-每天，2-每周，3-每月
	TaskExecTime           int64  `json:"task_exec_time"`            // 任执行时间
	TaskCloseTime          int64  `json:"task_close_time"`           // 结束时间
	FixedIntervalStartTime int64  `json:"fixed_interval_start_time"` // 固定时间间隔开始时间
	FixedIntervalTime      int32  `json:"fixed_interval_time"`       // 固定间隔时间
	FixedRunNum            int32  `json:"fixed_run_num"`             // 运行次数
	FixedIntervalTimeType  int32  `json:"fixed_interval_time_type"`  // 固定间隔时间类型：0-分钟，1-小时
}

type UIPlanGetTaskConfReq struct {
	PlanID string `json:"plan_id" binding:"required"`
	TeamID string `json:"team_id" binding:"required"`
}

type UIPlanGetTaskConfResp struct {
	PlanID                 string `json:"plan_id"`
	TeamID                 string `json:"team_id"`
	TaskType               int32  `json:"task_type"`
	SceneRunOrder          int32  `json:"scene_run_order"`
	Frequency              int32  `json:"frequency"`                 // 频次: 0-一次，1-每天，2-每周，3-每月
	TaskExecTime           int64  `json:"task_exec_time"`            // 任执行时间
	TaskCloseTime          int64  `json:"task_close_time"`           // 结束时间
	FixedIntervalStartTime int64  `json:"fixed_interval_start_time"` // 固定时间间隔开始时间
	FixedIntervalTime      int32  `json:"fixed_interval_time"`       // 固定间隔时间
	FixedRunNum            int32  `json:"fixed_run_num"`             // 运行次数
	FixedIntervalTimeType  int32  `json:"fixed_interval_time_type"`  // 运行次数
	Status                 int32  `json:"status"`
}

type RunUIPlanReq struct {
	PlanID string `json:"plan_id"`
	TeamID string `json:"team_id"`
}

type RunUIPlanResp struct {
	ReportID string `json:"report_id"`
}

type StopUIPlanReq struct {
	PlanID string `json:"plan_id"`
	TeamID string `json:"team_id" binding:"required,gt=0"`
}
