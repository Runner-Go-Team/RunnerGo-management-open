package packer

import (
	"kp-management/internal/pkg/dal/mao"
)

func TransMachineMonitorToMao(machineIP string, monitorData mao.HeartBeat, createdAt int64) *mao.MachineMonitorData {
	return &mao.MachineMonitorData{
		MachineIP:   machineIP,
		CreatedAt:   createdAt,
		MonitorData: monitorData,
	}
}
