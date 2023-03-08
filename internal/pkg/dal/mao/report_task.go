package mao

type ReportTask struct {
	TeamID   string    `bson:"team_id"`
	PlanID   string    `bson:"plan_id"`
	PlanName string    `bson:"plan_name"`
	ReportID string    `bson:"report_id"`
	TaskType int32     `bson:"task_type"`
	TaskMode int32     `bson:"task_mode"`
	ModeConf *ModeConf `bson:"mode_conf"`
}
