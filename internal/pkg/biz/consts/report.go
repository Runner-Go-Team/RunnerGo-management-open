package consts

const (
	ReportStatusNormal = 1
	ReportStatusFinish = 2

	// report使用的kafka分区数量
	KafkaReportPartitionNum = 10
	// kafka全局的报告分区key名
	KafkaReportPartition = "kafka:report:partition"
)
