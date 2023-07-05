package rao

type SavePreinstallReq struct {
	ID                      int32                   `json:"id"`
	TeamID                  string                  `json:"team_id" binding:"required"`
	ConfName                string                  `json:"conf_name"`
	TaskType                int32                   `json:"task_type"`
	TaskMode                int32                   `json:"task_mode"`
	ControlMode             int32                   `json:"control_mode"`
	DebugMode               string                  `json:"debug_mode"`
	ModeConf                *ModeConf               `json:"mode_conf"`
	TimedTaskConf           *TimedTaskConf          `json:"timed_task_conf"`
	IsOpenDistributed       int32                   `json:"is_open_distributed"`        // 是否开启分布式调度：0-关闭，1-开启
	MachineDispatchModeConf MachineDispatchModeConf `json:"machine_dispatch_mode_conf"` // 分布式压力机配置
}

type GetPreinstallDetailReq struct {
	ID int32 `json:"id"`
}

type PreinstallDetailResponse struct {
	ID                      int32                   `json:"id"`
	TeamID                  string                  `json:"team_id" binding:"required"`
	ConfName                string                  `json:"conf_name" binding:"required"`
	UserName                string                  `json:"user_name" binding:"required"`
	TaskType                int32                   `json:"task_type" binding:"required"`
	TaskMode                int32                   `json:"task_mode" binding:"required"`
	ControlMode             int32                   `json:"control_mode"`
	DebugMode               string                  `json:"debug_mode"`
	ModeConf                ModeConf                `json:"mode_conf" binding:"required"`
	TimedTaskConf           TimedTaskConf           `json:"timed_task_conf"`
	IsOpenDistributed       int32                   `json:"is_open_distributed"`        // 是否开启分布式调度：0-关闭，1-开启
	MachineDispatchModeConf MachineDispatchModeConf `json:"machine_dispatch_mode_conf"` // 分布式压力机配置
}

type GetPreinstallListReq struct {
	TeamID   string `json:"team_id" binding:"required"`
	ConfName string `json:"conf_name"`
	TaskType int32  `json:"task_type"`
	Page     int    `json:"page" form:"page,default=1"`
	Size     int    `json:"size" form:"size,default=10"`
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

type GetAvailableMachineListReq struct {
	TeamID string `json:"team_id"`
}

type GetAvailableMachineListResp struct {
	ID     int64  `json:"id"`
	Region string `json:"region"`
	IP     string `json:"ip"`
	Port   int32  `json:"port"`
	Name   string `json:"name"`
}
