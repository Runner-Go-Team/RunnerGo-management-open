package consts

const (
	// 报告状态
	ReportStatusNormal = 1 // 进行中
	ReportStatusFinish = 2 // 已完成

	// kafka全局的报告分区key名
	KafkaReportPartition = "kafka:report:partition"

	RedisPlanRunUUIDRelateReports = "PlanRunUUIDRelateReports:" // 执行计划和报告的关系
	RedisReportPlanRunUUID        = "ReportPlanRunUUID:"        // 报告对应的执行唯一值

	// api调试接口结果状态
	APiDebugStatusSuccess = "success"
	APiDebugStatusFail    = "failed"
)
