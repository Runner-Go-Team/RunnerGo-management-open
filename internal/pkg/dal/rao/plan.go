package rao

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
)

type StressPlan struct {
	ID                int64     `json:"id"`
	PlanID            string    `json:"plan_id"`
	RankID            int64     `json:"rank_id"`
	TeamID            string    `json:"team_id"`
	PlanName          string    `json:"plan_name"`
	TaskType          int32     `json:"task_type"`
	TaskMode          int32     `json:"task_mode"`
	Status            int32     `json:"status"`
	CreatedUserID     string    `json:"created_user_id"`
	CreatedUserName   string    `json:"created_user_name"`
	CreatedUserAvatar string    `json:"created_user_avatar"`
	Remark            string    `json:"remark"`
	CreatedTimeSec    int64     `json:"created_time_sec"`
	UpdatedTimeSec    int64     `json:"updated_time_sec"`
	ModeConf          *ModeConf `json:"mode_conf"`
}

type PlanTask struct {
	PlanID        string         `json:"plan_id"`
	SceneID       string         `json:"scene_id"`
	TaskType      int32          `json:"task_type"`
	Mode          int32          `json:"mode"`
	ModeConf      *ModeConf      `json:"mode_conf"`
	TimedTaskConf *TimedTaskConf `json:"timed_task_conf"`
}

type ModeConf struct {
	RoundNum         int64 `json:"round_num"`         // 轮次
	Concurrency      int64 `json:"concurrency"`       // 并发数
	ThresholdValue   int64 `json:"threshold_value"`   // 阈值
	StartConcurrency int64 `json:"start_concurrency"` // 起始并发数
	Step             int64 `json:"step"`              // 步长
	StepRunTime      int64 `json:"step_run_time"`     // 步长执行时长
	MaxConcurrency   int64 `json:"max_concurrency"`   // 最大并发数
	Duration         int64 `json:"duration"`          // 稳定持续时长，持续时长
	CreatedTimeSec   int64 `json:"created_time_sec"`  // 创建时间
}

type SendEditModeConf struct {
	RoundNum         int64 `json:"round_num"`         // 轮次
	Concurrency      int64 `json:"concurrency"`       // 并发数
	ThresholdValue   int64 `json:"threshold_value"`   // 阈值
	StartConcurrency int64 `json:"start_concurrency"` // 起始并发数
	Step             int64 `json:"step"`              // 步长
	StepRunTime      int64 `json:"step_run_time"`     // 步长执行时长
	MaxConcurrency   int64 `json:"max_concurrency"`   // 最大并发数
	Duration         int64 `json:"duration"`          // 稳定持续时长，持续时长
	CreatedTimeSec   int64 `json:"created_time_sec"`  // 创建时间
}

type ChangeTakeConf struct {
	RoundNum          int64               `json:"round_num"`           // 轮次
	Concurrency       int64               `json:"concurrency"`         // 并发数
	ThresholdValue    int64               `json:"threshold_value"`     // 阈值
	StartConcurrency  int64               `json:"start_concurrency"`   // 起始并发数
	Step              int64               `json:"step"`                // 步长
	StepRunTime       int64               `json:"step_run_time"`       // 步长执行时长
	MaxConcurrency    int64               `json:"max_concurrency"`     // 最大并发数
	Duration          int64               `json:"duration"`            // 稳定持续时长，持续时长
	CreatedTimeSec    int64               `json:"created_time_sec"`    // 创建时间
	UsableMachineList []UsableMachineInfo `json:"usable_machine_list"` // 可选机器列表
}

type RunPlanReq struct {
	PlanID  string   `json:"plan_id"`
	TeamID  string   `json:"team_id"`
	SceneID []string `json:"scene_id"`
}

type RunPlanResp struct {
}

type StopPlanReq struct {
	TeamID  string   `json:"team_id"`
	PlanIDs []string `json:"plan_ids"`
}

type StopPlanResp struct {
}

type ListUnderwayPlanReq struct {
	TeamID string `form:"team_id" binding:"required,gt=0"`
	Page   int    `form:"page,default=1"`
	Size   int    `form:"size,default=10"`
}

type ListUnderwayPlanResp struct {
	Plans      []*StressPlan `json:"plans"`
	Total      int64         `json:"total"`
	RunPlanNum int           `json:"run_plan_num"`
}

type ClonePlanReq struct {
	TeamID string `json:"team_id"`
	PlanID string `json:"plan_id"`
}

type ClonePlanResp struct {
}

type ListPlansReq struct {
	TeamID string `form:"team_id" binding:"required,gt=0"`
	Page   int    `form:"page,default=1"`
	Size   int    `form:"size,default=10"`

	Keyword      string `form:"keyword"`
	StartTimeSec int64  `form:"start_time_sec"`
	EndTimeSec   int64  `form:"end_time_sec"`

	TaskType int32 `form:"task_type"`
	TaskMode int32 `form:"task_mode"`
	Status   int32 `form:"status"`
	Sort     int32 `form:"sort"`
}

type ListPlansResp struct {
	Plans []*StressPlan `json:"plans"`
	Total int64         `json:"total"`
}

type SavePlanReq struct {
	ID       int64  `json:"id"`
	PlanID   string `json:"plan_id"`
	TeamID   string `json:"team_id" binding:"required,gt=0"`
	PlanName string `json:"plan_name" binding:"required"`
	Remark   string `json:"remark"`
	TaskType int32  `json:"task_type" binding:"required"`
}

type SavePlanConfReq struct {
	PlanID                  string                  `json:"plan_id" binding:"required,gt=0"`
	SceneID                 string                  `json:"scene_id" binding:"required,gt=0"`
	TeamID                  string                  `json:"team_id" binding:"required,gt=0"`
	Name                    string                  `json:"name" binding:"required"`
	TaskType                int32                   `json:"task_type" binding:"required,gt=0"`
	Mode                    int32                   `json:"mode" binding:"required,gt=0"`
	ControlMode             int32                   `json:"control_mode"`
	DebugMode               string                  `json:"debug_mode"`
	Remark                  string                  `json:"remark"`
	ModeConf                *ModeConf               `json:"mode_conf"`
	TimedTaskConf           *TimedTaskConf          `json:"timed_task_conf"`
	IsOpenDistributed       int32                   `json:"is_open_distributed"`        // 是否开启分布式调度：0-关闭，1-开启
	MachineDispatchModeConf MachineDispatchModeConf `json:"machine_dispatch_mode_conf"` // 分布式压力机配置
}

type MachineDispatchModeConf struct {
	MachineAllotType  int32               `json:"machine_allot_type"`  // 机器分配方式：0-权重，1-自定义
	UsableMachineList []UsableMachineInfo `json:"usable_machine_list"` // 可选机器列表
}

type UsableMachineInfo struct {
	MachineStatus    int32  `json:"machine_status"`    // 是否可用：1-使用中，2-已卸载
	MachineName      string `json:"machine_name"`      // 机器名称
	Region           string `json:"region"`            // 区域
	Ip               string `json:"ip"`                // ip
	Weight           int    `json:"weight"`            // 权重
	RoundNum         int64  `json:"round_num"`         // 轮次
	Concurrency      int64  `json:"concurrency"`       // 并发数
	ThresholdValue   int64  `json:"threshold_value"`   // 阈值
	StartConcurrency int64  `json:"start_concurrency"` // 起始并发数
	Step             int64  `json:"step"`              // 步长
	StepRunTime      int64  `json:"step_run_time"`     // 步长执行时长
	MaxConcurrency   int64  `json:"max_concurrency"`   // 最大并发数
	Duration         int64  `json:"duration"`          // 稳定持续时长，持续时长
	CreatedTimeSec   int64  `json:"created_time_sec"`  // 创建时间
}

type TimedTaskConf struct {
	Frequency     int32 `json:"frequency"`       // 频次: 0-一次，1-每天，2-每周，3-每月
	TaskExecTime  int64 `json:"task_exec_time"`  // 任执行时间
	TaskCloseTime int64 `json:"task_close_time"` // 结束时间
}

type SavePlanResp struct {
	PlanID string `json:"plan_id"`
}

type GetPlanTaskReq struct {
	PlanID   string `form:"plan_id" binding:"required,gt=0"`
	SceneID  string `form:"scene_id" binding:"required,gt=0"`
	TeamID   string `form:"team_id" binding:"required,gt=0"`
	TaskType int32  `form:"task_type"`
}

type PlanTaskResp struct {
	PlanID                  string                  `json:"plan_id"`
	SceneID                 string                  `json:"scene_id"`
	TaskType                int32                   `json:"task_type"`
	Mode                    int32                   `json:"mode"`
	ControlMode             int32                   `json:"control_mode"`
	DebugMode               string                  `json:"debug_mode"`
	ModeConf                ModeConf                `json:"mode_conf"`
	TimedTaskConf           TimedTaskConf           `json:"timed_task_conf"`
	IsOpenDistributed       int32                   `json:"is_open_distributed"`        // 是否开启分布式调度：0-关闭，1-开启
	MachineDispatchModeConf MachineDispatchModeConf `json:"machine_dispatch_mode_conf"` // 分布式压力机配置
}

type GetPlanTaskResp struct {
	PlanTask *PlanTaskResp `json:"plan_task"`
}

type GetPlanConfReq struct {
	TeamID string `form:"team_id" binding:"required,gt=0"`
	PlanID string `form:"plan_id" binding:"required,gt=0"`
}

type GetPlanResp struct {
	Plan *StressPlan `json:"plan"`
}

type DeletePlanReq struct {
	PlanID string `json:"plan_id"`
	TeamID string `json:"team_id"`
}

type DeletePlanResp struct {
}

type PlanEmailReq struct {
	PlanID string   `json:"plan_id"`
	TeamID string   `json:"team_id"`
	Emails []string `json:"emails"`
}

type PlanEmailResp struct {
}

type PlanListEmailReq struct {
	TeamID string `form:"team_id"`
	PlanID string `form:"plan_id"`
}

type PlanListEmailResp struct {
	Emails []*PlanEmail `json:"emails"`
}

type PlanEmail struct {
	ID            int64  `json:"id"`
	PlanID        string `json:"plan_id"`
	TeamID        string `json:"team_id"`
	Email         string `json:"email"`
	CreateTimeSec int64  `json:"create_time_sec"`
}

type PlanDeleteEmailReq struct {
	PlanID  string `json:"plan_id"`
	TeamID  string `json:"team_id"`
	EmailID int64  `json:"email_id"`
}

type PlanDeleteEmailResp struct {
}

type ImportSceneReq struct {
	PlanID       string   `json:"plan_id"`
	TeamID       string   `json:"team_id"`
	Source       int32    `json:"source"`
	TargetIDList []string `json:"target_id_list"`
}

type ImportSceneResp struct {
	Scenes []*model.Target `json:"scenes"`
}

type BatchDeletePlanReq struct {
	PlanIDs []string `json:"plan_ids" binding:"required"`
	TeamID  string   `json:"team_id" binding:"required"`
}

type NotifyStopStressReq struct {
	ReportID     string   `json:"report_id"`
	TeamID       string   `json:"team_id"`
	PlanID       string   `json:"plan_id"`
	DurationTime int64    `json:"duration_time"`
	Machines     []string `json:"machines"`
}

type GetEstimateUseVumNumReq struct {
	TeamID string `json:"team_id"`
	PlanID string `json:"plan_id"`
}

type GetEstimateUseVumNumResp struct {
	MaybeUseVumNum   int64 `json:"maybe_use_vum_num"`
	TeamUsableVumNum int64 `json:"team_usable_vum_num"`
}

// SubscriptionStressPlanStatusChange 订阅压测计划状态变更
type SubscriptionStressPlanStatusChange struct {
	Type            int             `json:"type"` // 1: stopPlan; 2: debug; 3.报告变更
	StopPlan        string          `json:"stop_plan"`
	Debug           string          `json:"debug"`
	MachineModeConf MachineModeConf `json:"machine_mode_conf"`
}

type MachineModeConf struct {
	Machine  string         `json:"machine"`
	ModeConf ChangeTakeConf `json:"mode_conf"`
}

type GetPublicFunctionListResp struct {
	Function     string `json:"function"`
	FunctionName string `json:"function_name"`
	Remark       string `json:"remark"`
}

type GetNewestStressPlanListReq struct {
	TeamID string `json:"team_id"`
	Page   int    `json:"page,default=1"`
	Size   int    `json:"size,default=10"`
}

type GetNewestStressPlanListResp struct {
	TeamID     string `json:"team_id"`
	PlanID     string `json:"plan_id"`
	PlanName   string `json:"plan_name"`
	PlanType   string `json:"plan_type"`
	Username   string `json:"username"`
	UserAvatar string `json:"user_avatar"`
	UpdatedAt  int64  `json:"updated_at"`
}

type GetNewestAutoPlanListReq struct {
	TeamID string `json:"team_id"`
	Page   int    `json:"page,default=1"`
	Size   int    `json:"size,default=10"`
}

type GetNewestAutoPlanListResp struct {
	TeamID     string `json:"team_id"`
	PlanID     string `json:"plan_id"`
	PlanName   string `json:"plan_name"`
	PlanType   string `json:"plan_type"`
	Username   string `json:"username"`
	UserAvatar string `json:"user_avatar"`
	UpdatedAt  int64  `json:"updated_at"`
}
