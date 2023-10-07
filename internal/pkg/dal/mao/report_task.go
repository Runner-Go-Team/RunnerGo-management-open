package mao

type ReportTask struct {
	TeamID                  string                  `bson:"team_id"`
	PlanID                  string                  `bson:"plan_id"`
	PlanName                string                  `bson:"plan_name"`
	ReportID                string                  `bson:"report_id"`
	TaskType                int32                   `bson:"task_type"`
	TaskMode                int32                   `bson:"task_mode"`
	ControlMode             int32                   `bson:"control_mode"` // 控制模式：0-集中模式，1-单独模式
	DebugMode               string                  `bson:"debug_mode"`   // debug模式：stop-关闭，all-开启全部日志，only_success-开启仅成功日志，only_error-开启仅错误日志
	ModeConf                ModeConf                `bson:"mode_conf"`
	IsOpenDistributed       int32                   `bson:"is_open_distributed"`        // 是否开启分布式调度：0-关闭，1-开启
	MachineDispatchModeConf MachineDispatchModeConf `bson:"machine_dispatch_mode_conf"` // 分布式压力机配置
}

type MachineDispatchModeConf struct {
	MachineAllotType  int32               `bson:"machine_allot_type"`  // 机器分配方式：0-权重，1-自定义
	UsableMachineList []UsableMachineInfo `bson:"usable_machine_list"` // 可选机器列表
}

type UsableMachineInfo struct {
	MachineStatus    int32  `bson:"machine_status"`    // 是否可用：1-使用中，2-已卸载
	MachineName      string `bson:"machine_name"`      // 机器名称
	Region           string `bson:"region"`            // 区域
	Ip               string `bson:"ip"`                // ip
	Weight           int    `bson:"weight"`            // 权重
	RoundNum         int64  `bson:"round_num"`         // 轮次
	Concurrency      int64  `bson:"concurrency"`       // 并发数
	ThresholdValue   int64  `bson:"threshold_value"`   // 阈值
	StartConcurrency int64  `bson:"start_concurrency"` // 起始并发数
	Step             int64  `bson:"step"`              // 步长
	StepRunTime      int64  `bson:"step_run_time"`     // 步长执行时长
	MaxConcurrency   int64  `bson:"max_concurrency"`   // 最大并发数
	Duration         int64  `bson:"duration"`          // 稳定持续时长，持续时长
	CreatedTimeSec   int64  `bson:"created_time_sec"`  // 创建时间
}
