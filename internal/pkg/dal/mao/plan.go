package mao

type Task struct {
	PlanID   int64     `bson:"plan_id"`
	SceneID  string    `bson:"scene_id"`
	TaskType int32     `bson:"task_type"`
	TaskMode int32     `bson:"task_mode"`
	ModeConf *ModeConf `bson:"mode_conf"`
}

type ModeConf struct {
	ReheatTime       int64 `bson:"reheat_time" json:"reheat_time"`             // 预热时长
	RoundNum         int64 `bson:"round_num" json:"round_num"`                 // 轮次
	Concurrency      int64 `bson:"concurrency" json:"concurrency"`             // 并发数
	ThresholdValue   int64 `bson:"threshold_value" json:"threshold_value"`     // 阈值
	StartConcurrency int64 `bson:"start_concurrency" json:"start_concurrency"` // 起始并发数
	Step             int64 `bson:"step" json:"step"`                           // 步长
	StepRunTime      int64 `bson:"step_run_time" json:"step_run_time"`         // 步长执行时长
	MaxConcurrency   int64 `bson:"max_concurrency" json:"max_concurrency"`     // 最大并发数
	Duration         int64 `bson:"duration" json:"duration"`                   // 稳定持续时长，持续时长
	CreatedTimeSec   int64 `bson:"created_time_sec" json:"created_time_sec"`   // 创建时间
}

type Preinstall struct {
	TeamID   string    `bson:"team_id"`
	PlanID   int64     `bson:"plan_id"`
	TaskType int32     `bson:"task_type"`
	CronExpr string    `bson:"cron_expr"`
	Mode     int32     `bson:"mode"`
	ModeConf *ModeConf `bson:"mode_conf"`
}

type ChangeTaskConf struct {
	ReportID                string                  `bson:"report_id"`
	TeamID                  string                  `bson:"team_id"`
	PlanID                  string                  `bson:"plan_id"`
	ModeConf                *ModeConf               `bson:"mode_conf"`
	IsOpenDistributed       int32                   `bson:"is_open_distributed"`        // 是否开启分布式调度：0-关闭，1-开启
	MachineDispatchModeConf MachineDispatchModeConf `bson:"machine_dispatch_mode_conf"` // 分布式压力机配置
}

//type ChangeTaskModeConf struct {
//	ReheatTime       int64 `bson:"reheat_time" json:"reheat_time"`             // 预热时长
//	RoundNum         int64 `bson:"round_num" json:"round_num"`                 // 轮次
//	Concurrency      int64 `bson:"concurrency" json:"concurrency"`             // 并发数
//	ThresholdValue   int64 `bson:"threshold_value" json:"threshold_value"`     // 阈值
//	StartConcurrency int64 `bson:"start_concurrency" json:"start_concurrency"` // 起始并发数
//	Step             int64 `bson:"step" json:"step"`                           // 步长
//	StepRunTime      int64 `bson:"step_run_time" json:"step_run_time"`         // 步长执行时长
//	MaxConcurrency   int64 `bson:"max_concurrency" json:"max_concurrency"`     // 最大并发数
//	Duration         int64 `bson:"duration" json:"duration"`                   // 稳定持续时长，持续时长
//	CreatedTimeSec   int64 `json:"created_time_sec"`                           // 创建时间
//}
