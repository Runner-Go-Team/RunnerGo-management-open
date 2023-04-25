package mao

type ReportTask struct {
	TeamID      string    `bson:"team_id"`
	PlanID      string    `bson:"plan_id"`
	PlanName    string    `bson:"plan_name"`
	ReportID    string    `bson:"report_id"`
	TaskType    int32     `bson:"task_type"`
	TaskMode    int32     `bson:"task_mode"`
	ControlMode int32     `bson:"control_mode"` // 控制模式：0-集中模式，1-单独模式
	DebugMode   string    `bson:"debug_mode"`   // debug模式：stop-关闭，all-开启全部日志，only_success-开启仅成功日志，only_error-开启仅错误日志
	ModeConf    *ModeConf `bson:"mode_conf"`
}
