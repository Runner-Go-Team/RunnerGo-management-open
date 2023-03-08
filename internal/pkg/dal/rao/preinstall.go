package rao

type SavePreinstallReq struct {
	ID            int32          `json:"id"`
	TeamID        string         `json:"team_id" binding:"required"`
	ConfName      string         `json:"conf_name"`
	TaskType      int32          `json:"task_type"`
	TaskMode      int32          `json:"task_mode"`
	ControlMode   int32          `json:"control_mode"`
	ModeConf      *ModeConf      `json:"mode_conf"`
	TimedTaskConf *TimedTaskConf `json:"timed_task_conf"`
}

type GetPreinstallDetailReq struct {
	ID int32 `json:"id"`
}

type PreinstallDetailResponse struct {
	ID            int32          `json:"id"`
	TeamID        string         `json:"team_id" binding:"required"`
	ConfName      string         `json:"conf_name" binding:"required"`
	UserName      string         `json:"user_name" binding:"required"`
	TaskType      int32          `json:"task_type" binding:"required"`
	TaskMode      int32          `json:"task_mode" binding:"required"`
	ControlMode   int32          `json:"control_mode"`
	ModeConf      *ModeConf      `json:"mode_conf" binding:"required"`
	TimedTaskConf *TimedTaskConf `json:"timed_task_conf"`
}

type GetPreinstallListReq struct {
	TeamID   string `json:"team_id" binding:"required"`
	ConfName string `json:"conf_name"`
	TaskType int32  `json:"task_type"`

	Page int `json:"page" form:"page,default=1"`
	Size int `json:"size" form:"size,default=10"`
}

type GetPreinstallResponse struct {
	PreinstallList []*PreinstallDetailResponse `json:"preinstall_list"`
	Total          int64                       `json:"total"`
}

type DeletePreinstallReq struct {
	ID       int32  `json:"id" binding:"required"`
	TeamID   string `json:"team_id" binding:"required"`
	ConfName string `json:"conf_name" binding:"required"`
}

type CopyPreinstallReq struct {
	ID     int32  `json:"id" binding:"required"`
	TeamID string `json:"team_id" binding:"required"`
}
