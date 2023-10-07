package consts

const (
	MachineStatusAvailable = 1 // 使用中
	MachineStatusUnload    = 2 // 已卸载

	MachineListRedisKey         = "RunnerMachineList"
	MachineUseStatePrefix       = "MachineUseState:"
	UiEngineMachineListRedisKey = "UiEngineMachineList"

	MachineMonitorPrefix   = "MachineMonitor:"     // 压力机监控数据
	TotalKafkaPartitionKey = "TotalKafkaPartition" // 总的可用分区数据
	RunKafkaPartitionKey   = "RunKafkaPartition"   // 发送广播消息
)
