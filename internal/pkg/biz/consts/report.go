package consts

const (
	ReportStatusNormal = 1
	ReportStatusFinish = 2

	// kafka全局的报告分区key名
	KafkaReportPartition = "kafka:report:partition"

	RedisPlanRunUUIDRelateReports = "PlanRunUUIDRelateReports:" // 执行计划和报告的关系
	RedisReportPlanRunUUID        = "ReportPlanRunUUID:"        // 报告对应的执行唯一值
)
