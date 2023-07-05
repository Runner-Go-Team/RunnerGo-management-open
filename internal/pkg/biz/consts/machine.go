package consts

const (
	MachineStatusAvailable = 1 // 使用中
	MachineStatusUnload    = 2 // 已卸载

	MachineListRedisKey   = "RunnerMachineList"
	MachineUseStatePrefix = "MachineUseState:"
	MachineAliveTime      = 10

	MachineMonitorPrefix = "MachineMonitor:" // 压力机监控数据

	StressBelongPartitionKey = "StressBelongPartition" // 压力机对应已经使用的分区数据前缀
	TotalKafkaPartitionKey   = "TotalKafkaPartition"
)
