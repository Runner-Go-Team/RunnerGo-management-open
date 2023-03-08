package consts

const (
	MachineStatusIdle = 1
	MachineStatusBusy = 2

	MachineListRedisKey   = "RunnerMachineList"
	MachineUseStatePrefix = "MachineUseState:"
	MachineAliveTime      = 10

	MachineMonitorPrefix = "MachineMonitor:" // 压力机监控数据

	StressBelongPartitionKey = "StressBelongPartition" // 压力机对应已经使用的分区数据前缀
	TotalKafkaPartitionKey   = "TotalKafkaPartition"

	InitPartitionTotalNum = 2 // 初始化总共可使用的分区上限数量
)
