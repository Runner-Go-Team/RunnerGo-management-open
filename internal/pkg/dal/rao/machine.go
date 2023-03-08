package rao

import (
	"time"
)

type GetMachineListParam struct {
	ServerType int32  `json:"server_type"`
	Name       string `json:"name"`
	SortTag    int    `json:"sort_tag"`

	Page int `json:"page" form:"page,default=1"`
	Size int `json:"size" form:"size,default=10"`
}

type GetMachineListResponse struct {
	MachineList []*MachineList `json:"machine_list"`
	Total       int64          `json:"total"`
}

type MachineList struct {
	ID                int64     `json:"id"`
	Name              string    `json:"name"`
	CPUUsage          float32   `json:"cpu_usage"`          // CPU使用率
	CPULoadOne        float32   `json:"cpu_Load_one"`       // CPU-1分钟内平均负载
	CPULoadFive       float32   `json:"cpu_load_five"`      // CPU-5分钟内平均负载
	CPULoadFifteen    float32   `json:"cpu_load_fifteen"`   // CPU-15分钟内平均负载
	MemUsage          float32   `json:"mem_usage"`          // 内存使用率
	DiskUsage         float32   `json:"disk_usage"`         // 磁盘使用率
	MaxGoroutines     int64     `json:"max_goroutines"`     // 最大协程数
	CurrentGoroutines int64     `json:"current_goroutines"` // 已用协程数
	ServerType        int32     `json:"server_type"`
	Status            int32     `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type ChangeMachineOnOff struct {
	ID     int64 `json:"id"`
	Status int32 `json:"status"`
}
