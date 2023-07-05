package handler

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/errno"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/log"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/response"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/conf"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/mao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/machine"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/stress"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/packer"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"golang.org/x/net/context"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"time"
)

// GetMachineList 获取机器列表
func GetMachineList(ctx *gin.Context) {
	var req rao.GetMachineListParam
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	res, total, err := machine.GetMachineList(ctx, req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}
	response.SuccessWithData(ctx, rao.GetMachineListResponse{
		MachineList: res,
		Total:       total,
	})
	return
}

// ChangeMachineOnOff 压力机启用或卸载
func ChangeMachineOnOff(ctx *gin.Context) {
	var req rao.ChangeMachineOnOff
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	err := machine.ChangeMachineOnOff(ctx, req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}
	response.Success(ctx)
	return
}

// MachineDataInsert 把压力机上报的机器数据插入数据库
func MachineDataInsert() {
	for {
		ctx := context.Background()
		// 从Redis获取压力机列表
		machineListRes := dal.GetRDB().HGetAll(ctx, consts.MachineListRedisKey)
		if len(machineListRes.Val()) == 0 || machineListRes.Err() != nil {
			log.Logger.Info("压力机数据入库--没有获取到任何压力机上报数据，err:", machineListRes.Err())
			time.Sleep(5 * time.Second) // 5秒循环一次
			continue
		}

		// 有数据，则入库
		for machineAddr, machineDetail := range machineListRes.Val() {
			// 获取机器IP，端口号，区域
			machineAddrSlice := strings.Split(machineAddr, "_")
			if len(machineAddrSlice) != 3 {
				continue
			}

			// 把机器详情信息解析成格式化数据
			var runnerMachineInfo stress.HeartBeat
			err := json.Unmarshal([]byte(machineDetail), &runnerMachineInfo)
			if err != nil {
				log.Logger.Info("压力机数据入库--压力机详情数据解析失败，err：", err)
				continue
			}

			// 当前时间
			nowTime := time.Now().Unix()
			if runnerMachineInfo.CreateTime < nowTime-60 {
				dal.GetRDB().HDel(ctx, consts.MachineListRedisKey, machineAddr)
				continue
			}

			ip := machineAddrSlice[0]
			port := machineAddrSlice[1]
			portInt, err := strconv.Atoi(port)
			if err != nil {
				log.Logger.Info("压力机数据入库--转换类型失败，err:", err)
				continue
			}
			region := machineAddrSlice[2]

			// 查询当前机器信息是否存在数据库
			tx := dal.GetQuery().Machine

			// 查询数据
			_, err = tx.WithContext(ctx).Where(tx.IP.Eq(ip)).First()
			if err != nil && err != gorm.ErrRecordNotFound {
				log.Logger.Info("压力机数据入库--查询数据出错，err:", err)
				continue
			}

			if err == nil { // 查到了，修改数据
				updateData := model.Machine{
					Port:              int32(portInt),
					Region:            region,
					Name:              runnerMachineInfo.Name,
					CPUUsage:          float32(runnerMachineInfo.CpuUsage),
					CPULoadOne:        float32(runnerMachineInfo.CpuLoad.Load1),
					CPULoadFive:       float32(runnerMachineInfo.CpuLoad.Load5),
					CPULoadFifteen:    float32(runnerMachineInfo.CpuLoad.Load15),
					MemUsage:          float32(runnerMachineInfo.MemInfo[0].UsedPercent),
					DiskUsage:         float32(runnerMachineInfo.DiskInfos[0].UsedPercent),
					MaxGoroutines:     runnerMachineInfo.MaxGoroutines,
					CurrentGoroutines: runnerMachineInfo.CurrentGoroutines,
					ServerType:        int32(runnerMachineInfo.ServerType),
				}
				_, err := tx.WithContext(ctx).Where(tx.IP.Eq(ip)).Updates(&updateData)
				if err != nil {
					log.Logger.Info("压力机数据入库--更新数据失败，err:", err)
					continue
				}
			} else { // 没查到，新增数据
				insertData := model.Machine{
					IP:                ip,
					Port:              int32(portInt),
					Region:            region,
					Name:              runnerMachineInfo.Name,
					CPUUsage:          float32(runnerMachineInfo.CpuUsage),
					CPULoadOne:        float32(runnerMachineInfo.CpuLoad.Load1),
					CPULoadFive:       float32(runnerMachineInfo.CpuLoad.Load5),
					CPULoadFifteen:    float32(runnerMachineInfo.CpuLoad.Load15),
					MemUsage:          float32(runnerMachineInfo.MemInfo[0].UsedPercent),
					DiskUsage:         float32(runnerMachineInfo.DiskInfos[0].UsedPercent),
					MaxGoroutines:     runnerMachineInfo.MaxGoroutines,
					CurrentGoroutines: runnerMachineInfo.CurrentGoroutines,
					ServerType:        int32(runnerMachineInfo.ServerType),
				}
				err := tx.WithContext(ctx).Create(&insertData)
				if err != nil {
					log.Logger.Info("压力机数据入库")
					continue
				}
			}
		}

		time.Sleep(5 * time.Second) // 5秒循环一次
	}

}

// MachineMonitorInsert 压力机监控数据入库
func MachineMonitorInsert() {
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectMachineMonitorData)
	for {
		ctx := context.Background()
		machineList, _ := dal.GetRDB().Keys(ctx, consts.MachineMonitorPrefix+"*").Result()

		for _, MachineMonitorKey := range machineList {
			machineAddrSlice := strings.Split(MachineMonitorKey, ":")
			if len(machineAddrSlice) != 2 {
				continue
			}
			machineIP := machineAddrSlice[1]
			// 从Redis获取压力机列表
			machineListRes := dal.GetRDB().LRange(ctx, MachineMonitorKey, 0, -1).Val()
			if len(machineListRes) == 0 {
				continue
			}
			for _, monitorData := range machineListRes {
				var runnerMachineInfo mao.HeartBeat
				// 把机器详情信息解析成格式化数据
				err := json.Unmarshal([]byte(monitorData), &runnerMachineInfo)
				if err != nil {
					log.Logger.Info("压力机监控数据入库--数据解析失败 err：", err)
					continue
				}

				machineMonitorInsertData := packer.TransMachineMonitorToMao(machineIP, runnerMachineInfo, runnerMachineInfo.CreateTime)
				_, err = collection.InsertOne(ctx, machineMonitorInsertData)
				if err != nil {
					log.Logger.Info("压力机监控数据入库--插入mg数据失败，err:", err)
					continue
				}
			}
			// 数据入库完毕，把redis列表删掉
			err := dal.GetRDB().Del(ctx, MachineMonitorKey)
			if err.Err() != nil {
				log.Logger.Info("压力机监控数据入库--删除redis列表失败，err:", err.Err())
			}
		}

		time.Sleep(5 * time.Second)
	}
}

// InitTotalKafkaPartition 写入压力机所需总的分区数
func InitTotalKafkaPartition() {
	ctx := context.Background()

	dal.GetRDB().Del(ctx, consts.TotalKafkaPartitionKey)

	// 从redis查询当前
	StressBelongPartitionInfo := dal.GetRDB().HGetAll(ctx, consts.StressBelongPartitionKey)

	// 已经使用的分数切片
	usedPartitionMap := make(map[int]int)

	if StressBelongPartitionInfo.Err() == nil && len(StressBelongPartitionInfo.Val()) > 0 { // 查到数据了
		for _, partitionInfo := range StressBelongPartitionInfo.Val() {
			var tempData []int
			err := json.Unmarshal([]byte(partitionInfo), &tempData)
			if err != nil {
				continue
			}
			if len(tempData) > 0 {
				for _, partitionNum := range tempData {
					usedPartitionMap[partitionNum] = 1
				}
			}
		}
	}

	//组装总共需要初始化的分区数组
	canUsePartitionTotalNum := conf.Conf.CanUsePartitionTotalNum
	totalKafkaPartitionArr := make([]interface{}, 0, conf.Conf.MachineConfig.InitPartitionTotalNum)
	for i := 0; i < canUsePartitionTotalNum; i++ {
		if _, ok := usedPartitionMap[i]; !ok {
			totalKafkaPartitionArr = append(totalKafkaPartitionArr, i)
		}

	}
	// 初始化总的分区数（排除已经使用过的）
	_, err := dal.GetRDB().LPush(ctx, consts.TotalKafkaPartitionKey, totalKafkaPartitionArr...).Result()
	if err != nil {
		log.Logger.Info("初始化压力机分区总数失败")
	}
	return
}

// DeleteLostConnectMachine 删除失去连接的压力机
func DeleteLostConnectMachine() {
	for {
		ctx := context.Background()
		currentTime := time.Now()
		m, _ := time.ParseDuration("-1m")
		oldTime := currentTime.Add(m)
		tx := dal.GetQuery().Machine
		_, err := tx.WithContext(ctx).Where(tx.UpdatedAt.Lt(oldTime)).Unscoped().Delete()
		if err != nil {
			log.Logger.Info("删除失效的压力机--删除压力机数据失败")
		}
		time.Sleep(60 * time.Second)
	}
}
