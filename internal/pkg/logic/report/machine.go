package report

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"kp-management/internal/pkg/dal/mao"
	"strconv"
	"time"

	"kp-management/internal/pkg/biz/consts"
	"kp-management/internal/pkg/dal"
	"kp-management/internal/pkg/dal/query"
	"kp-management/internal/pkg/dal/rao"
)

func ListMachines(ctx context.Context, req *rao.ListMachineReq) (*rao.ListMachineResp, error) {
	r := query.Use(dal.DB()).StressPlanReport
	report, err := r.WithContext(ctx).Where(r.TeamID.Eq(req.TeamID), r.ReportID.Eq(req.ReportID)).First()
	if err != nil {
		return nil, err
	}

	//// 产过半个月以上的报告，不让查询压力机监控数据
	//deleteTime := time.Now().Unix() - (15 * 24 * 3600)
	//if report.CreatedAt.Unix() <= deleteTime {
	//	return nil, fmt.Errorf("只能查询15天以内的压力机监控数据")
	//}

	startTimeSec := report.CreatedAt.Unix() - 60
	var endTimeSec int64
	// 判断报告是否完成
	if report.Status == consts.ReportStatusNormal { // 进行中
		endTimeSec = time.Now().Unix()
	} else { // 已完成
		endTimeSec = report.UpdatedAt.Unix() + 60
	}

	resp := rao.ListMachineResp{
		StartTimeSec: startTimeSec,
		EndTimeSec:   endTimeSec,
		ReportStatus: report.Status,
		Metrics:      make([]*rao.Metric, 0),
	}

	rm := dal.GetQuery().ReportMachine
	rms, err := rm.WithContext(ctx).Where(rm.TeamID.Eq(req.TeamID), rm.ReportID.Eq(req.ReportID)).Find()
	if err != nil {
		return nil, err
	}

	// 排重字典
	machineMap := make(map[string]int, len(rms))

	machineTable := dal.GetQuery().Machine

	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectMachineMonitorData)
	for _, machine := range rms {
		if _, ok := machineMap[machine.IP]; ok {
			continue
		} else {
			machineMap[machine.IP] = 1
		}

		// 查询机器信息
		machineInfo, err := machineTable.WithContext(ctx).Where(machineTable.IP.Eq(machine.IP)).First()
		if err != nil {
			return nil, err
		}

		// 从mg里面查出来压力机监控数据
		mmd, err := collection.Find(ctx, bson.D{{"machine_ip", machine.IP}, {"created_at", bson.D{{"$gte", startTimeSec}}}, {"created_at", bson.D{{"$lte", endTimeSec}}}})
		if err != nil {
			return nil, err
		}
		var machineMonitorSlice []*mao.MachineMonitorData
		if err = mmd.All(ctx, &machineMonitorSlice); err != nil {
			return nil, err
		}

		cpu := make([][]interface{}, 0, len(machineMonitorSlice))
		mem := make([][]interface{}, 0, len(machineMonitorSlice))
		net := make([][]interface{}, 0, len(machineMonitorSlice))
		disk := make([][]interface{}, 0, len(machineMonitorSlice))
		for _, machineMonitorInfo := range machineMonitorSlice {
			cpuTmp := make([]interface{}, 0, 2)
			cpuUsage, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", machineMonitorInfo.MonitorData.CpuUsage), 64)
			cpuTmp = append(cpuTmp, machineMonitorInfo.MonitorData.CreateTime, cpuUsage)
			cpu = append(cpu, cpuTmp)

			memTmp := make([]interface{}, 0, 5)
			memUsedPercent, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", machineMonitorInfo.MonitorData.MemInfo[0].UsedPercent), 64)
			memTmp = append(memTmp, machineMonitorInfo.MonitorData.CreateTime, memUsedPercent, machineMonitorInfo.MonitorData.MemInfo[0].Total, machineMonitorInfo.MonitorData.MemInfo[0].Used, machineMonitorInfo.MonitorData.MemInfo[0].Free)
			mem = append(mem, memTmp)

			//统计网络IO
			var bytesSent uint64 = 0
			var bytesRecv uint64 = 0
			for _, netInfo := range machineMonitorInfo.MonitorData.Networks {
				if netInfo.Name == "eth0" {
					bytesSent = netInfo.BytesSent
					bytesRecv = netInfo.BytesRecv
					break
				}
			}

			totalIOBytes := bytesSent + bytesRecv
			ioBytes := totalIOBytes / (1024 * 1024)
			netTmp := make([]interface{}, 0, 5)
			netTmp = append(netTmp, machineMonitorInfo.MonitorData.CreateTime, ioBytes, bytesSent, bytesRecv)
			net = append(net, netTmp)

			diskTmp := make([]interface{}, 0, 5)
			diskUsedPercent, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", machineMonitorInfo.MonitorData.DiskInfos[0].UsedPercent), 64)
			diskTmp = append(diskTmp, machineMonitorInfo.MonitorData.CreateTime, diskUsedPercent, machineMonitorInfo.MonitorData.DiskInfos[0].Total, machineMonitorInfo.MonitorData.DiskInfos[0].Used, machineMonitorInfo.MonitorData.DiskInfos[0].Free)
			disk = append(disk, diskTmp)

		}
		resp.Metrics = append(resp.Metrics, &rao.Metric{
			MachineName: machineInfo.Name,
			CPU:         cpu,
			Mem:         mem,
			NetIO:       net,
			DiskIO:      disk,
		})
	}

	return &resp, nil
}
