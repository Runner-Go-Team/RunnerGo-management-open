package packer

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/mao"
)

func TransMachineMonitorToMao(machineIP string, monitorData mao.HeartBeat, createdAt int64) *mao.MachineMonitorData {
	return &mao.MachineMonitorData{
		MachineIP:   machineIP,
		CreatedAt:   createdAt,
		MonitorData: monitorData,
	}
}
