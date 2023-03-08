package mao

type MachineMonitorData struct {
	MachineIP   string    `bson:"machine_ip" json:"machine_ip"`
	CreatedAt   int64     `bson:"created_at" json:"created_at"`
	MonitorData HeartBeat `bson:"monitor_data" json:"monitor_data"`
}

// 压力机心跳上报数据
type HeartBeat struct {
	Name     string  `json:"name" bson:"name"`           // 机器名称
	CpuUsage float64 `json:"cpu_usage" bson:"cpu_usage"` // CPU使用率
	//CpuLoad           *load.AvgStat `json:"cpu_load" bson:"cpu_load"`                     // CPU负载信息
	MemInfo           []MemInfo  `json:"mem_info" bson:"mem_info"`                     // 内存使用情况
	Networks          []Network  `json:"networks" bson:"networks"`                     // 网络连接情况
	DiskInfos         []DiskInfo `json:"disk_infos" bson:"disk_infos"`                 // 磁盘IO情况
	MaxGoroutines     int64      `json:"max_goroutines" bson:"max_goroutines"`         // 当前机器支持最大协程数
	CurrentGoroutines int64      `json:"current_goroutines" bson:"current_goroutines"` // 当前已用协程数
	ServerType        int64      `json:"server_type" bson:"server_type"`               // 压力机类型：0-主力机器，1-备用机器
	CreateTime        int64      `json:"create_time" bson:"create_time"`               // 数据上报时间（时间戳）
}

type MemInfo struct {
	Total       uint64  `json:"total" bson:"total"`
	Used        uint64  `json:"used" bson:"used"`
	Free        uint64  `json:"free" bson:"free"`
	UsedPercent float64 `json:"usedPercent" bson:"usedPercent"`
}

type DiskInfo struct {
	Total       uint64  `json:"total" bson:"total"`
	Free        uint64  `json:"free" bson:"free"`
	Used        uint64  `json:"used" bson:"used"`
	UsedPercent float64 `json:"usedPercent" bson:"usedPercent"`
}

type Network struct {
	Name        string `json:"name" bson:"name"`
	BytesSent   uint64 `json:"bytesSent" bson:"bytesSent"`
	BytesRecv   uint64 `json:"bytesRecv" bson:"bytesRecv"`
	PacketsSent uint64 `json:"packetsSent" bson:"packetsSent"`
	PacketsRecv uint64 `json:"packetsRecv" bson:"packetsRecv"`
}
