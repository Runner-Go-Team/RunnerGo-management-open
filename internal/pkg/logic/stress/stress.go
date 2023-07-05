package stress

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/errno"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/log"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/record"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/uuid"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/conf"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/runner"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/packer"
	"github.com/go-omnibus/omnibus"
	"github.com/go-omnibus/proof"
	"github.com/go-resty/resty/v2"
	"go.mongodb.org/mongo-driver/bson"
	"gorm.io/gen"
	"gorm.io/gorm"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/mao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/query"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/run_plan"
	"github.com/shirou/gopsutil/load"
)

type Baton struct {
	Ctx                   context.Context
	PlanID                string
	TeamID                string
	UserID                string
	RunType               int
	SceneIDs              []string
	plan                  *model.StressPlan
	scenes                []*model.Target
	task                  map[string]*run_plan.Task // sceneID 对应任务配置
	globalVariables       run_plan.GlobalVariable
	flows                 []*mao.Flow
	sceneVariables        map[string]run_plan.GlobalVariable
	importVariables       []*model.VariableImport
	reports               []model.StressPlanReport
	ReportIDs             []string
	balance               *DispatchMachineBalance
	stress                []run_plan.Stress
	stressRun             []run_plan.Stress
	MachineList           []*HeartBeat
	MachineMaxConcurrence int64
}

type UsableMachineMap struct {
	IP     string // IP地址(包含端口号)
	Region string // 机器所属区域
	Weight int64  // 权重
	//UsableGoroutines int64  // 可用协程数
	MaxGoroutines int64 // 可用并发数
}

// 压力机心跳上报数据
type HeartBeat struct {
	Name              string        `json:"name"`               // 机器名称
	CpuUsage          float64       `json:"cpu_usage"`          // CPU使用率
	CpuLoad           *load.AvgStat `json:"cpu_load"`           // CPU负载信息
	MemInfo           []MemInfo     `json:"mem_info"`           // 内存使用情况
	Networks          []Network     `json:"networks"`           // 网络连接情况
	DiskInfos         []DiskInfo    `json:"disk_infos"`         // 磁盘IO情况
	MaxGoroutines     int64         `json:"max_goroutines"`     // 当前机器支持最大协程数
	CurrentGoroutines int64         `json:"current_goroutines"` // 当前已用协程数
	ServerType        int64         `json:"server_type"`        // 压力机类型：0-主力机器，1-备用机器
	CreateTime        int64         `json:"create_time"`        // 数据上报时间（时间戳）
	FmtCreateTime     time.Time     `json:"fmt_create_time"`    // 格式化时间
}

type MemInfo struct {
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"usedPercent"`
}

type DiskInfo struct {
	Total       uint64  `json:"total"`
	Free        uint64  `json:"free"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"usedPercent"`
}

type Network struct {
	Name        string `json:"name"`
	BytesSent   uint64 `json:"bytesSent"`
	BytesRecv   uint64 `json:"bytesRecv"`
	PacketsSent uint64 `json:"packetsSent"`
	PacketsRecv uint64 `json:"packetsRecv"`
}

type Stress interface {
	Execute(baton *Baton) (int, error)
	SetNext(Stress)
}

type CheckStressPlanTaskType struct {
	next Stress
}

func (s *CheckStressPlanTaskType) Execute(baton *Baton) (int, error) {
	log.Logger.Info("运行计划--检查运行计划方式", baton.TeamID, baton.PlanID)
	if baton.RunType != 2 { // 非定时任务的执行
		tx := dal.GetQuery().StressPlan
		planInfo, err := tx.WithContext(baton.Ctx).Where(tx.TeamID.Eq(baton.TeamID),
			tx.PlanID.Eq(baton.PlanID)).First()
		if err != nil {
			return errno.ErrMysqlFailed, fmt.Errorf("计划没查到")
		}
		if planInfo.TaskType == consts.PlanTaskTypeCronjob { // 定时任务
			stressPlanTimedTaskConfTable := dal.GetQuery().StressPlanTimedTaskConf
			ttcInfo, err := stressPlanTimedTaskConfTable.WithContext(baton.Ctx).Where(stressPlanTimedTaskConfTable.PlanID.Eq(baton.PlanID)).First()
			if err != nil {
				return errno.ErrMustTaskInit, err
			}

			// 检查定时任务时间是否过期
			nowTime := time.Now().Unix()
			var taskCloseTime int64 = 0
			if ttcInfo.Frequency == 0 {
				taskCloseTime = ttcInfo.TaskExecTime
			} else {
				taskCloseTime = ttcInfo.TaskCloseTime
			}

			if taskCloseTime <= nowTime {
				return errno.ErrTimedTaskOverdue, fmt.Errorf("开始或结束时间不能早于当前时间")
			}

			_, err = stressPlanTimedTaskConfTable.WithContext(baton.Ctx).Where(stressPlanTimedTaskConfTable.TeamID.Eq(baton.TeamID),
				stressPlanTimedTaskConfTable.PlanID.Eq(baton.PlanID)).UpdateSimple(stressPlanTimedTaskConfTable.Status.Value(consts.TimedTaskInExec),
				stressPlanTimedTaskConfTable.RunUserID.Value(baton.UserID))
			if err != nil {
				return errno.ErrMysqlFailed, fmt.Errorf("定时任务状态修改失败")
			}

			// 修改计划的状态
			ap := dal.GetQuery().StressPlan
			_, err = ap.WithContext(baton.Ctx).Where(ap.TeamID.Eq(baton.TeamID), ap.PlanID.Eq(baton.PlanID)).UpdateSimple(ap.Status.Value(consts.PlanStatusUnderway))
			if err != nil {
				return errno.ErrMysqlFailed, fmt.Errorf("计划状态修改失败")
			}
			_ = record.InsertExecute(baton.Ctx, baton.TeamID, baton.UserID, record.OperationOperateExecPlan, planInfo.PlanName)
			return errno.Ok, fmt.Errorf("定时任务已经开启")
		}
	}
	return s.next.Execute(baton)
}

func (s *CheckStressPlanTaskType) SetNext(stress Stress) {
	s.next = stress
}

type CheckIdleMachine struct {
	next Stress
}

func (s *CheckIdleMachine) Execute(baton *Baton) (int, error) {
	log.Logger.Info("运行计划--检查压力机", baton.TeamID, baton.PlanID)
	// 从Redis获取压力机列表
	machineListRes := dal.GetRDB().HGetAll(baton.Ctx, consts.MachineListRedisKey)
	if len(machineListRes.Val()) == 0 || machineListRes.Err() != nil {
		// todo 后面可能增加兜底策略
		log.Logger.Info("资源不足--没有上报上来的空闲压力机可用")
		return errno.ErrResourceNotEnough, fmt.Errorf("资源不足--没有上报上来的空闲压力机可用")
	}

	baton.balance = &DispatchMachineBalance{}

	usableMachineMap := UsableMachineMap{}                                       // 单个压力机基本数据
	usableMachineSlice := make([]UsableMachineMap, 0, len(machineListRes.Val())) // 所有上报过来的压力机切片
	var minWeight int64                                                          // 所有可用压力机里面最小的权重的值
	var inUseMachineNum int                                                      // 所有有任务在运行的压力机数量

	tx := dal.GetQuery().Machine
	// 查到了机器列表，然后判断可用性
	runnerMachineInfo := HeartBeat{}
	log.Logger.Info("从redis获取到的当前所有可用机器列表数据：", machineListRes.Val())
	for machineAddr, machineDetail := range machineListRes.Val() {
		breakFor := false
		// 把机器详情信息解析成格式化数据
		err := json.Unmarshal([]byte(machineDetail), &runnerMachineInfo)
		if err != nil {
			log.Logger.Info("runner_machine_detail 数据解析失败 err：", err)
			continue
		}

		// 压力机数据上报时间超过3秒，则认为服务不可用，不参与本次压力测试
		nowTime := time.Now().Unix()
		fmtNowTime := time.Now()
		if nowTime-runnerMachineInfo.CreateTime > int64(conf.Conf.MachineConfig.MachineAliveTime) {
			log.Logger.Info("当前压力机上报心跳数据超时，暂不可用,当前时间为：", fmtNowTime, " 机器上报时间为：", runnerMachineInfo.FmtCreateTime)
			continue
		}

		// 判断当前压力机性能是否爆满,如果某个指标爆满，则不参与本次压力测试
		if runnerMachineInfo.CpuUsage >= float64(conf.Conf.MachineConfig.CpuTopLimit) { // CPU使用判断
			log.Logger.Info("CPU超过使用阈值，阈值为：", conf.Conf.MachineConfig.CpuTopLimit, "当前cpu使用率为：", runnerMachineInfo.CpuUsage, "机器信息为：", machineAddr)
			continue
		}
		for _, memInfo := range runnerMachineInfo.MemInfo { // 内存使用判断
			if memInfo.UsedPercent >= float64(conf.Conf.MachineConfig.MemoryTopLimit) {
				log.Logger.Info("内存超过使用阈值，阈值为：", conf.Conf.MachineConfig.MemoryTopLimit, "当前内存使用率为：", memInfo.UsedPercent, "机器信息为：", machineAddr)
				breakFor = true
				break
			}
		}
		for _, diskInfo := range runnerMachineInfo.DiskInfos { // 磁盘使用判断
			if diskInfo.UsedPercent >= float64(conf.Conf.MachineConfig.DiskTopLimit) {
				log.Logger.Info("磁盘超过使用阈值，阈值为：", conf.Conf.MachineConfig.DiskTopLimit, "当前磁盘使用率为：", diskInfo.UsedPercent, "机器信息为：", machineAddr)
				breakFor = true
				break
			}
		}

		// 最后判断是否结束当前循环
		if breakFor {
			continue
		}

		machineAddrSlice := strings.Split(machineAddr, "_")
		if len(machineAddrSlice) != 3 {
			log.Logger.Info("机器信息解析失败,数据为：", machineAddr)
			continue
		}

		// 判断当前压力机是否被停用，如果停用，则不参与压测
		machineInfo, err := tx.WithContext(baton.Ctx).Where(tx.IP.Eq(machineAddrSlice[0])).First()
		if err != nil {
			log.Logger.Info("运行计划--没有查到当前压力机数据，err:", err)
			continue
		}
		if machineInfo.Status == 2 { // 已停用
			log.Logger.Info("运行计划--压力机已停用，机器IP:", machineInfo.IP)
			continue
		}

		// 组装可用机器结构化数据
		usableMachineMap.IP = machineAddrSlice[0] + ":" + machineAddrSlice[1]
		usableMachineMap.Weight = runnerMachineInfo.MaxGoroutines
		usableMachineMap.MaxGoroutines = runnerMachineInfo.MaxGoroutines
		log.Logger.Info("插入可用列表的机器：", usableMachineMap)
		usableMachineSlice = append(usableMachineSlice, usableMachineMap)

		// 获取当前压力机当中最小的权重值
		if minWeight == 0 || minWeight > runnerMachineInfo.MaxGoroutines {
			minWeight = runnerMachineInfo.MaxGoroutines
		}

		// 获取当前机器是否使用当中
		machineUseStateKey := consts.MachineUseStatePrefix + machineAddrSlice[0] + ":" + machineAddrSlice[1]
		useStateVal, _ := dal.GetRDB().Get(baton.Ctx, machineUseStateKey).Result()
		if useStateVal != "" {
			inUseMachineNum++
		}

		// 统计所有可用的机器信息
		baton.MachineList = append(baton.MachineList, &runnerMachineInfo)
	}

	for k, machineInfo := range usableMachineSlice {
		if inUseMachineNum < len(usableMachineSlice) {
			// 获取当前机器是否使用当中
			machineUseStateKey := consts.MachineUseStatePrefix + machineInfo.IP
			useStateVal, _ := dal.GetRDB().Get(baton.Ctx, machineUseStateKey).Result()
			if useStateVal != "" {
				usableMachineSlice[k].Weight = int64(minWeight) - 1
				if machineInfo.Weight <= 0 {
					usableMachineSlice[k].Weight = 1
				}
			}
		}
		log.Logger.Info("当前机器信息：", machineInfo.IP)
	}

	log.Logger.Info("当前所有可用机器列表：", usableMachineSlice)

	sort.Slice(usableMachineSlice, func(i, j int) bool {
		return usableMachineSlice[i].Weight > usableMachineSlice[j].Weight
	})
	log.Logger.Info("当前所有可用机器列表,排序后：", usableMachineSlice)

	// 按当前顺序把机器放到备用列表
	for k, machineInfo := range usableMachineSlice {
		log.Logger.Info("序号：", k, " 可用压力机IP:", machineInfo.IP, " 最大可用并发数为：", machineInfo.MaxGoroutines)
		addErr := baton.balance.AddMachine(machineInfo.IP, machineInfo.MaxGoroutines)
		if addErr != nil {
			continue
		}
	}

	if len(baton.balance.rss) == 0 {
		log.Logger.Info("资源不足--当前没有空闲压力机可用")
		return errno.ErrResourceNotEnough, fmt.Errorf("资源不足--当前没有空闲压力机可用")
	}

	return s.next.Execute(baton)
}

func (s *CheckIdleMachine) SetNext(stress Stress) {
	s.next = stress
}

type AssemblePlan struct {
	next Stress
}

func (s *AssemblePlan) Execute(baton *Baton) (int, error) {
	log.Logger.Info("运行计划--组装计划", baton.TeamID, baton.PlanID)
	tx := dal.GetQuery().StressPlan
	p, err := tx.WithContext(baton.Ctx).Where(tx.TeamID.Eq(baton.TeamID), tx.PlanID.Eq(baton.PlanID)).First()
	if err != nil {
		return errno.ErrMysqlFailed, fmt.Errorf("没有查到性能计划相关信息")
	}
	baton.plan = p
	return s.next.Execute(baton)
}

func (s *AssemblePlan) SetNext(stress Stress) {
	s.next = stress
}

type AssembleScenes struct {
	next Stress
}

func (s *AssembleScenes) Execute(baton *Baton) (int, error) {
	log.Logger.Info("运行计划--组装场景", baton.TeamID, baton.PlanID)
	tx := query.Use(dal.DB()).Target
	conditions := make([]gen.Condition, 0)
	conditions = append(conditions, tx.TeamID.Eq(baton.TeamID))
	conditions = append(conditions, tx.PlanID.Eq(baton.PlanID))
	conditions = append(conditions, tx.TargetType.Eq(consts.TargetTypeScene))
	conditions = append(conditions, tx.Source.Eq(consts.TargetSourcePlan))
	conditions = append(conditions, tx.Status.Eq(consts.TargetStatusNormal))
	conditions = append(conditions, tx.IsDisabled.Eq(consts.TargetIsDisabledNo))
	if len(baton.SceneIDs) > 0 {
		conditions = append(conditions, tx.TargetID.In(baton.SceneIDs...))
	}

	scenes, err := tx.WithContext(baton.Ctx).Where(conditions...).Find()
	if err != nil {
		return errno.ErrMysqlFailed, err
	}
	if len(scenes) == 0 { // 场景为空，直接返回
		return errno.ErrEmptyScene, fmt.Errorf("场景不能为空")
	}
	baton.scenes = scenes
	return s.next.Execute(baton)
}

func (s *AssembleScenes) SetNext(stress Stress) {
	s.next = stress
}

type AssembleTask struct {
	next Stress
}

func (s *AssembleTask) Execute(baton *Baton) (int, error) {
	log.Logger.Info("运行计划--组装任务", baton.TeamID, baton.PlanID)
	memo := make(map[string]*run_plan.Task)
	tx := dal.GetQuery().StressPlanTimedTaskConf
	// 判断参数是否包含scene_ids
	if baton.RunType == 2 { // 说明当前任务时定时任务自动过来执行的
		timedTaskConfInfo, err := tx.WithContext(baton.Ctx).Where(tx.SceneID.Eq(baton.SceneIDs[0])).First()
		if err != nil {
			return errno.ErrMysqlFailed, err
		}

		var modeConf run_plan.ModeConf
		err = json.Unmarshal([]byte(timedTaskConfInfo.ModeConf), &modeConf)
		if err != nil {
			log.Logger.Info("运行定时任务--解析任务配置文件失败")
			return errno.ErrUnMarshalFailed, err
		}
		machineDispatchModeConf := run_plan.MachineDispatchModeConf{}
		if timedTaskConfInfo.IsOpenDistributed == 1 && timedTaskConfInfo.MachineDispatchModeConf != "" {
			err := json.Unmarshal([]byte(timedTaskConfInfo.MachineDispatchModeConf), &machineDispatchModeConf)
			if err != nil {
				log.Logger.Info("运行定时任务--解析任务分布式配置文件失败")
				return errno.ErrUnMarshalFailed, err
			}
		}

		memo[baton.SceneIDs[0]] = &run_plan.Task{
			PlanID:                  timedTaskConfInfo.PlanID,
			SceneID:                 timedTaskConfInfo.SceneID,
			TaskType:                baton.plan.TaskType,
			TaskMode:                timedTaskConfInfo.TaskMode,
			ControlMode:             timedTaskConfInfo.ControlMode,
			DebugMode:               timedTaskConfInfo.DebugMode,
			ModeConf:                modeConf,
			IsOpenDistributed:       timedTaskConfInfo.IsOpenDistributed,
			MachineDispatchModeConf: machineDispatchModeConf,
		}
	} else { // 普通任务
		// 查询普通任务的配置
		taskConfTable := dal.GetQuery().StressPlanTaskConf
		taskConfList, err := taskConfTable.WithContext(baton.Ctx).Where(taskConfTable.TeamID.Eq(baton.TeamID),
			taskConfTable.PlanID.Eq(baton.PlanID)).Find()
		if err != nil || len(taskConfList) == 0 {
			return errno.ErrMustTaskInit, fmt.Errorf("请填写任务配置并保存")
		}
		for _, taskConfInfo := range taskConfList {
			var mc run_plan.ModeConf
			err := json.Unmarshal([]byte(taskConfInfo.ModeConf), &mc)
			if err != nil {
				return errno.ErrUnMarshalFailed, err
			}

			// 解析分布式配置信息
			machineDispatchModeConf := run_plan.MachineDispatchModeConf{}
			if taskConfInfo.IsOpenDistributed == 1 && taskConfInfo.MachineDispatchModeConf != "" {
				err := json.Unmarshal([]byte(taskConfInfo.MachineDispatchModeConf), &machineDispatchModeConf)
				if err != nil {
					log.Logger.Info("运行定时任务--解析任务分布式配置文件失败")
					return errno.ErrUnMarshalFailed, err
				}
			}

			temp := &run_plan.Task{
				PlanID:                  taskConfInfo.PlanID,
				SceneID:                 taskConfInfo.SceneID,
				TaskType:                baton.plan.TaskType,
				TaskMode:                taskConfInfo.TaskMode,
				ControlMode:             taskConfInfo.ControlMode,
				DebugMode:               taskConfInfo.DebugMode,
				ModeConf:                mc,
				IsOpenDistributed:       taskConfInfo.IsOpenDistributed,
				MachineDispatchModeConf: machineDispatchModeConf,
			}
			memo[temp.SceneID] = temp
		}
	}
	baton.task = memo
	return s.next.Execute(baton)
}

func (s *AssembleTask) SetNext(stress Stress) {
	s.next = stress
}

type AssembleGlobalVariables struct {
	next Stress
}

func (s *AssembleGlobalVariables) Execute(baton *Baton) (int, error) {
	log.Logger.Info("运行计划--组装全局变量", baton.TeamID, baton.PlanID)
	globalVariable := run_plan.GlobalVariable{}
	// 查询全局变量
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectGlobalParam)
	cur, err := collection.Find(baton.Ctx, bson.D{{"team_id", baton.TeamID}})
	var globalParamDataArr []*mao.GlobalParamData
	if err == nil {
		if err := cur.All(baton.Ctx, &globalParamDataArr); err != nil {
			return errno.ErrMongoFailed, fmt.Errorf("全局参数数据获取失败")
		}
	}

	cookieParam := make([]rao.CookieParam, 0, 100)
	headerParam := make([]rao.HeaderParam, 0, 100)
	variableParam := make([]rao.VariableParam, 0, 100)
	assertParam := make([]rao.AssertParam, 0, 100)
	for _, globalParamInfo := range globalParamDataArr {
		if globalParamInfo.ParamType == 1 {
			err = json.Unmarshal([]byte(globalParamInfo.DataDetail), &cookieParam)
			if err != nil {
				return errno.ErrUnMarshalFailed, err
			}
			parameter := make([]*run_plan.Parameter, 0, len(cookieParam))
			for _, v := range cookieParam {
				temp := &run_plan.Parameter{
					IsChecked: v.IsChecked,
					Key:       v.Key,
					Value:     v.Value,
				}
				parameter = append(parameter, temp)
			}
			globalVariable.Cookie.Parameter = parameter
		}
		if globalParamInfo.ParamType == 2 {
			err = json.Unmarshal([]byte(globalParamInfo.DataDetail), &headerParam)
			if err != nil {
				return errno.ErrUnMarshalFailed, err
			}

			parameter := make([]*run_plan.Parameter, 0, len(headerParam))
			for _, v := range headerParam {
				temp := &run_plan.Parameter{
					IsChecked: v.IsChecked,
					Key:       v.Key,
					Value:     v.Value,
				}
				parameter = append(parameter, temp)
			}
			globalVariable.Header.Parameter = parameter

		}
		if globalParamInfo.ParamType == 3 {
			err = json.Unmarshal([]byte(globalParamInfo.DataDetail), &variableParam)
			if err != nil {
				return errno.ErrUnMarshalFailed, err
			}

			parameter := make([]run_plan.VarForm, 0, len(variableParam))
			for _, v := range variableParam {
				temp := run_plan.VarForm{
					IsChecked: int64(v.IsChecked),
					Key:       v.Key,
					Value:     v.Value,
				}
				parameter = append(parameter, temp)
			}
			globalVariable.Variable = parameter

		}
		if globalParamInfo.ParamType == 4 {
			err = json.Unmarshal([]byte(globalParamInfo.DataDetail), &assertParam)
			if err != nil {
				return errno.ErrUnMarshalFailed, err
			}

			parameter := make([]run_plan.AssertionText, 0, len(assertParam))
			for _, v := range assertParam {
				temp := run_plan.AssertionText{
					IsChecked:    int(v.IsChecked),
					ResponseType: int8(v.ResponseType),
					Compare:      v.Compare,
					Var:          v.Var,
					Val:          v.Val,
				}
				parameter = append(parameter, temp)
			}
			globalVariable.Assert = parameter
		}
	}
	baton.globalVariables = globalVariable
	return s.next.Execute(baton)
}

func (s *AssembleGlobalVariables) SetNext(stress Stress) {
	s.next = stress
}

type AssembleFlows struct {
	next Stress
}

func (s *AssembleFlows) Execute(baton *Baton) (int, error) {
	log.Logger.Info("运行计划--组装flow", baton.TeamID, baton.PlanID)
	var sceneIDs []string
	for _, scene := range baton.scenes {
		sceneIDs = append(sceneIDs, scene.TargetID)
	}

	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectFlow)
	cur, err := collection.Find(baton.Ctx, bson.D{{"scene_id", bson.D{{"$in", sceneIDs}}}})
	if err != nil {
		return errno.ErrMongoFailed, err
	}

	var flows []*mao.Flow
	if err := cur.All(baton.Ctx, &flows); err != nil {
		return errno.ErrMongoFailed, err
	}

	// 判断场景flow是否为空
	if len(flows) > 0 {
		for _, flow := range flows {
			var sceneFlowNodeTemp mao.Node
			err := bson.Unmarshal(flow.Nodes, &sceneFlowNodeTemp)
			if err != nil {
				return errno.ErrMongoFailed, fmt.Errorf("场景flow解析失败")
			}
			if len(sceneFlowNodeTemp.Nodes) == 0 {
				return errno.ErrEmptySceneFlow, fmt.Errorf("场景flow为空")
			}
		}
	}

	if len(flows) != len(sceneIDs) {
		log.Logger.Info("场景flow不能为空")
		return errno.ErrEmptySceneFlow, errors.New("场景flow不能为空")
	}

	baton.flows = flows
	return s.next.Execute(baton)
}

func (s *AssembleFlows) SetNext(stress Stress) {
	s.next = stress
}

type AssembleSceneVariables struct {
	next Stress
}

func (s *AssembleSceneVariables) Execute(baton *Baton) (int, error) {
	log.Logger.Info("运行计划--组装场景变量", baton.TeamID, baton.PlanID)
	sceneVariableMap := make(map[string]run_plan.GlobalVariable, len(baton.scenes))
	// 查询全局变量
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectSceneParam)
	for _, sceneInfo := range baton.scenes {
		sceneVariable := run_plan.GlobalVariable{}
		cur, err := collection.Find(baton.Ctx, bson.D{{"team_id", baton.TeamID}, {"scene_id", sceneInfo.TargetID}})
		var sceneParamDataArr []*mao.SceneParamData
		if err == nil {
			if err := cur.All(baton.Ctx, &sceneParamDataArr); err != nil {
				return errno.ErrMongoFailed, fmt.Errorf("场景参数数据获取失败")
			}
		}

		cookieParam := make([]rao.CookieParam, 0, 100)
		headerParam := make([]rao.HeaderParam, 0, 100)
		variableParam := make([]rao.VariableParam, 0, 100)
		assertParam := make([]rao.AssertParam, 0, 100)
		for _, sceneParamInfo := range sceneParamDataArr {
			if sceneParamInfo.ParamType == 1 {
				err = json.Unmarshal([]byte(sceneParamInfo.DataDetail), &cookieParam)
				if err != nil {
					return errno.ErrUnMarshalFailed, err
				}
				parameter := make([]*run_plan.Parameter, 0, len(cookieParam))
				for _, v := range cookieParam {
					temp := &run_plan.Parameter{
						IsChecked: v.IsChecked,
						Key:       v.Key,
						Value:     v.Value,
					}
					parameter = append(parameter, temp)
				}
				sceneVariable.Cookie.Parameter = parameter
			}
			if sceneParamInfo.ParamType == 2 {
				err = json.Unmarshal([]byte(sceneParamInfo.DataDetail), &headerParam)
				if err != nil {
					return errno.ErrUnMarshalFailed, err
				}

				parameter := make([]*run_plan.Parameter, 0, len(headerParam))
				for _, v := range headerParam {
					temp := &run_plan.Parameter{
						IsChecked: v.IsChecked,
						Key:       v.Key,
						Value:     v.Value,
					}
					parameter = append(parameter, temp)
				}
				sceneVariable.Header.Parameter = parameter

			}
			if sceneParamInfo.ParamType == 3 {
				err = json.Unmarshal([]byte(sceneParamInfo.DataDetail), &variableParam)
				if err != nil {
					return errno.ErrUnMarshalFailed, err
				}

				parameter := make([]run_plan.VarForm, 0, len(variableParam))
				for _, v := range variableParam {
					temp := run_plan.VarForm{
						IsChecked: int64(v.IsChecked),
						Key:       v.Key,
						Value:     v.Value,
					}
					parameter = append(parameter, temp)
				}
				sceneVariable.Variable = parameter
			}
			if sceneParamInfo.ParamType == 4 {
				err = json.Unmarshal([]byte(sceneParamInfo.DataDetail), &assertParam)
				if err != nil {
					return errno.ErrUnMarshalFailed, err
				}

				parameter := make([]run_plan.AssertionText, 0, len(assertParam))
				for _, v := range assertParam {
					temp := run_plan.AssertionText{
						IsChecked:    int(v.IsChecked),
						ResponseType: int8(v.ResponseType),
						Compare:      v.Compare,
						Var:          v.Var,
						Val:          v.Val,
					}
					parameter = append(parameter, temp)
				}
				sceneVariable.Assert = parameter
			}
		}
		sceneVariableMap[sceneInfo.TargetID] = sceneVariable
	}
	baton.sceneVariables = sceneVariableMap
	return s.next.Execute(baton)
}

func (s *AssembleSceneVariables) SetNext(stress Stress) {
	s.next = stress
}

type AssembleImportVariables struct {
	next Stress
}

func (s *AssembleImportVariables) Execute(baton *Baton) (int, error) {
	log.Logger.Info("运行计划--组装导入变量", baton.TeamID, baton.PlanID)
	var sceneIDs []string
	for _, scene := range baton.scenes {
		sceneIDs = append(sceneIDs, scene.TargetID)
	}

	tx := query.Use(dal.DB()).VariableImport
	vis, err := tx.WithContext(baton.Ctx).Where(tx.SceneID.In(sceneIDs...), tx.Status.Eq(consts.VariableStatusOpen)).Find()
	if err != nil {
		return errno.ErrMysqlFailed, err
	}

	baton.importVariables = vis
	return s.next.Execute(baton)
}

func (s *AssembleImportVariables) SetNext(stress Stress) {
	s.next = stress
}

type MakeReport struct {
	next Stress
}

func (s *MakeReport) Execute(baton *Baton) (int, error) {
	log.Logger.Info("运行计划--创建报告", baton.TeamID, baton.PlanID)
	tx := dal.GetQuery().StressPlanReport
	reports := make([]model.StressPlanReport, 0, len(baton.scenes))
	for _, scene := range baton.scenes {
		if _, ok := baton.task[scene.TargetID]; !ok {
			return errno.ErrMustTaskInit, errors.New("当前场景没有配置任务类型或任务模式，场景id：" + scene.TargetID)
		}

		reportInfo, err := tx.WithContext(baton.Ctx).Where(tx.TeamID.Eq(baton.TeamID)).Order(tx.RankID.Desc()).First()
		if err != nil && err != gorm.ErrRecordNotFound {
			return errno.ErrMysqlFailed, err
		}

		var rankID int64 = 1
		if err == nil {
			rankID = reportInfo.RankID + 1
		}
		tempReportID := uuid.GetUUID()
		baton.ReportIDs = append(baton.ReportIDs, tempReportID)
		log.Logger.Info("报告的用户ID1:", baton.UserID)
		reportData := model.StressPlanReport{
			ReportID:    tempReportID,
			ReportName:  baton.plan.PlanName + "-" + scene.Name,
			RankID:      rankID,
			TeamID:      scene.TeamID,
			PlanID:      baton.plan.PlanID,
			PlanName:    baton.plan.PlanName,
			SceneID:     scene.TargetID,
			SceneName:   scene.Name,
			TaskType:    baton.plan.TaskType,
			TaskMode:    baton.task[scene.TargetID].TaskMode,
			ControlMode: baton.task[scene.TargetID].ControlMode,
			DebugMode:   baton.task[scene.TargetID].DebugMode,
			Status:      consts.ReportStatusNormal,
			CreatedAt:   time.Now(),
			RunUserID:   baton.UserID,
		}
		log.Logger.Info("报告的用户ID2:", reportData.RunUserID)
		if err := tx.WithContext(baton.Ctx).Create(&reportData); err != nil {
			return errno.ErrMysqlFailed, err
		}
		reports = append(reports, reportData)
	}

	for _, report := range reports {
		usableMachineInfo := make([]mao.UsableMachineInfo, 0, len(baton.task[report.SceneID].MachineDispatchModeConf.UsableMachineList))
		for _, v := range baton.task[report.SceneID].MachineDispatchModeConf.UsableMachineList {
			usableMachineInfoTemp := mao.UsableMachineInfo{
				MachineStatus:    v.MachineStatus,
				MachineName:      v.MachineName,
				Region:           v.Region,
				Ip:               v.Ip,
				Weight:           v.Weight,
				RoundNum:         v.RoundNum,
				Concurrency:      v.Concurrency,
				ThresholdValue:   v.ThresholdValue,
				StartConcurrency: v.StartConcurrency,
				Step:             v.Step,
				StepRunTime:      v.StepRunTime,
				MaxConcurrency:   v.MaxConcurrency,
				Duration:         v.Duration,
				CreatedTimeSec:   v.CreatedTimeSec,
			}
			usableMachineInfo = append(usableMachineInfo, usableMachineInfoTemp)
		}

		collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectReportTask)
		temp := mao.ReportTask{
			ReportID:    report.ReportID,
			TaskType:    report.TaskType,
			TaskMode:    report.TaskMode,
			ControlMode: baton.task[report.SceneID].ControlMode,
			DebugMode:   baton.task[report.SceneID].DebugMode,
			TeamID:      baton.TeamID,
			PlanID:      baton.plan.PlanID,
			PlanName:    baton.plan.PlanName,
			ModeConf: mao.ModeConf{
				ReheatTime:       baton.task[report.SceneID].ModeConf.ReheatTime,
				RoundNum:         baton.task[report.SceneID].ModeConf.RoundNum,
				Concurrency:      baton.task[report.SceneID].ModeConf.Concurrency,
				ThresholdValue:   baton.task[report.SceneID].ModeConf.ThresholdValue,
				StartConcurrency: baton.task[report.SceneID].ModeConf.StartConcurrency,
				Step:             baton.task[report.SceneID].ModeConf.Step,
				StepRunTime:      baton.task[report.SceneID].ModeConf.StepRunTime,
				MaxConcurrency:   baton.task[report.SceneID].ModeConf.MaxConcurrency,
				Duration:         baton.task[report.SceneID].ModeConf.Duration,
			},
			IsOpenDistributed: baton.task[report.SceneID].IsOpenDistributed,
			MachineDispatchModeConf: mao.MachineDispatchModeConf{
				MachineAllotType:  baton.task[report.SceneID].MachineDispatchModeConf.MachineAllotType,
				UsableMachineList: usableMachineInfo,
			},
		}

		_, err := collection.InsertOne(baton.Ctx, temp)
		if err != nil {
			return errno.ErrMongoFailed, err
		}

		// 把报告的初始化debug模式写入mg数据库
		collection = dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectDebugStatus)
		debug := bson.D{{"report_id", report.ReportID}, {"team_id", baton.TeamID}, {"plan_id", baton.PlanID}, {"debug", baton.task[report.SceneID].DebugMode}}
		_, err = collection.InsertOne(baton.Ctx, debug)
		if err != nil {
			return errno.ErrMongoFailed, err
		}
	}

	baton.reports = reports
	return s.next.Execute(baton)
}

func (s *MakeReport) SetNext(stress Stress) {
	s.next = stress
}

type MakeStress struct {
	next Stress
}

func (s *MakeStress) Execute(baton *Baton) (int, error) {
	log.Logger.Info("运行计划--组装压测参数", baton.TeamID, baton.PlanID)
	var allSceneTotalConcurrency int64 // 所有任务的总并发

	pathArr := make([]run_plan.FileList, 0, len(baton.importVariables))
	for _, importVariableInfo := range baton.importVariables {
		temp := run_plan.FileList{
			IsChecked: int64(importVariableInfo.Status),
			Path:      importVariableInfo.URL,
		}
		pathArr = append(pathArr, temp)
	}

	importVariables := make([]run_plan.FileList, 0)
	for _, v := range baton.importVariables {
		temp := run_plan.FileList{
			IsChecked: int64(v.Status),
			Path:      v.URL,
		}
		importVariables = append(importVariables, temp)
	}

	// 优化代码
	reportMap := make(map[string]model.StressPlanReport, len(baton.reports))
	for _, report := range baton.reports {
		reportMap[report.SceneID] = report
	}

	flowMap := make(map[string]*mao.Flow, len(baton.flows))
	for _, flowInfo := range baton.flows {
		flowMap[flowInfo.SceneID] = flowInfo
	}

	for _, scene := range baton.scenes {
		nodes := mao.Node{}
		if err := bson.Unmarshal(flowMap[scene.TargetID].Nodes, &nodes); err != nil {
			log.Logger.Errorf("node bson unmarshal err:%v", err)
			continue
		}

		prepositions := mao.Preposition{}
		if err := bson.Unmarshal(flowMap[scene.TargetID].Prepositions, &prepositions); err != nil {
			log.Logger.Errorf("prepositions bson unmarshal err:%v", err)
			continue
		}
		prepositionsArr := make([]rao.Preposition, 0, len(prepositions.Prepositions))
		for _, v := range prepositions.Prepositions {
			temp := rao.Preposition{
				Type:  v.Type,
				Event: v,
			}
			prepositionsArr = append(prepositionsArr, temp)
		}

		edges := mao.Edge{}
		if err := bson.Unmarshal(flowMap[scene.TargetID].Edges, &edges); err != nil {
			log.Logger.Errorf("edges bson unmarshal err:%v", err)
		}

		if _, ok := baton.task[scene.TargetID]; !ok {
			return errno.ErrMustTaskInit, errors.New("请填写任务配置并保存")
		}

		nodesRound := packer.GetNodesByLevel(nodes.Nodes, edges.Edges)

		// 把flow里面的pre_url替换成环境变量的
		for k1, v1 := range nodesRound {
			for k2, v2 := range v1 {
				if flowMap[scene.TargetID].EnvID != 0 {
					nodesRound[k1][k2].API.Request.PreUrl = v2.API.EnvInfo.PreUrl
				} else {
					nodesRound[k1][k2].API.Request.PreUrl = ""
				}
			}
		}

		req := run_plan.Stress{
			PlanID:     baton.plan.PlanID,
			PlanName:   baton.plan.PlanName,
			ReportID:   reportMap[scene.TargetID].ReportID,
			TeamID:     baton.TeamID,
			ReportName: baton.plan.PlanName,
			ConfigTask: run_plan.ConfigTask{
				TaskType:    baton.plan.TaskType,
				Mode:        baton.task[scene.TargetID].TaskMode,
				ControlMode: baton.task[scene.TargetID].ControlMode,
				DebugMode:   baton.task[scene.TargetID].DebugMode,
				Remark:      baton.plan.Remark,
				ModeConf: run_plan.ModeConf{
					ReheatTime:       baton.task[scene.TargetID].ModeConf.ReheatTime,
					RoundNum:         baton.task[scene.TargetID].ModeConf.RoundNum,
					Concurrency:      baton.task[scene.TargetID].ModeConf.Concurrency,
					ThresholdValue:   baton.task[scene.TargetID].ModeConf.ThresholdValue,
					StartConcurrency: baton.task[scene.TargetID].ModeConf.StartConcurrency,
					Step:             baton.task[scene.TargetID].ModeConf.Step,
					StepRunTime:      baton.task[scene.TargetID].ModeConf.StepRunTime,
					MaxConcurrency:   baton.task[scene.TargetID].ModeConf.MaxConcurrency,
					Duration:         baton.task[scene.TargetID].ModeConf.Duration,
				},
				IsOpenDistributed:       baton.task[scene.TargetID].IsOpenDistributed,
				MachineDispatchModeConf: baton.task[scene.TargetID].MachineDispatchModeConf,
			},
			GlobalVariable: baton.globalVariables,
			Scene: run_plan.Scene{
				SceneID:                 scene.TargetID,
				EnablePlanConfiguration: false,
				SceneName:               scene.Name,
				TeamID:                  baton.TeamID,
				Configuration: run_plan.SceneConfiguration{
					ParameterizedFile: run_plan.SceneVariablePath{
						Paths: importVariables,
					},
					SceneVariable: baton.sceneVariables[scene.TargetID],
				},
				NodesRound:   nodesRound,
				Prepositions: prepositionsArr,
			},
		}
		baton.stress = append(baton.stress, req)

		// 统计总的并发
		// 接口总数
		var apiCnt int64
		for _, nodeArr := range req.Scene.NodesRound {
			for _, node := range nodeArr {
				if node.Type == "api" {
					apiCnt++
				}
			}
		}

		var oneSceneTotalConcurrency int64                         // 当前任务的总并发
		if req.ConfigTask.TaskType == consts.PlanModeConcurrence { // 并发模式
			oneSceneTotalConcurrency = req.ConfigTask.ModeConf.Concurrency
		} else { // 其他模式
			oneSceneTotalConcurrency = req.ConfigTask.ModeConf.MaxConcurrency
		}
		allSceneTotalConcurrency = allSceneTotalConcurrency + oneSceneTotalConcurrency
	}

	// 统计压力机最大的并发能力
	var machineTotalConcurrency int64 = 0
	for _, machineInfo := range baton.balance.rss {
		log.Logger.Info("当前压力机IP：", machineInfo.addr, " 可用并发数为：", machineInfo.maxGoroutines)
		machineTotalConcurrency = machineTotalConcurrency + machineInfo.maxGoroutines
	}
	log.Logger.Info("当前全部压力机总的并发数和当前计划总并发数分别为：", machineTotalConcurrency, " ", allSceneTotalConcurrency)
	if allSceneTotalConcurrency > machineTotalConcurrency {
		log.Logger.Info("资源不足--当前计划的总并发大于压力机可用并发数")
		_ = DeletePlanReport(baton)
		return errno.ErrResourceNotEnough, fmt.Errorf("资源不足--当前计划的总并发大于压力机可用并发数")
	}
	// 当前压力机总的可用并发数
	baton.MachineMaxConcurrence = machineTotalConcurrency
	return s.next.Execute(baton)
}

func (s *MakeStress) SetNext(stress Stress) {
	s.next = stress
}

type SplitStress struct {
	next Stress
}

func (s *SplitStress) Execute(baton *Baton) (int, error) {
	log.Logger.Info("运行计划--开始拆分任务方法", baton.TeamID, baton.PlanID)
	stressRun := make([]run_plan.Stress, 0, len(baton.stress))

	// 机器剩余可用协程map
	machineUsableGoroutines := make(map[string]int64)
	// 获取当前机器对应可用协程数
	for _, machineInfo := range baton.balance.rss {
		machineUsableGoroutines[machineInfo.addr] = machineInfo.maxGoroutines
	}
	curIndex := 0 // 当前使用的压力机数组下标
	usableMachineNum := len(baton.balance.rss)
	reportMachineNumMap := make(map[string]int32) // 报告对应的机器数量
	for k, stress := range baton.stress {
		log.Logger.Info("当前计划执行的任务序号：", k, " 报告id:", stress.ReportID)

		tx := dal.GetQuery()
		// 判断是否是分布式任务
		if stress.ConfigTask.IsOpenDistributed == 1 { // 分布式任务
			if stress.ConfigTask.MachineDispatchModeConf.MachineAllotType == 0 { // 权重
				if stress.ConfigTask.Mode == consts.PlanModeConcurrence || stress.ConfigTask.Mode == consts.PlanModeRound { // 并发模式 或者 轮次模式
					for _, v := range stress.ConfigTask.MachineDispatchModeConf.UsableMachineList {
						addrInfo, err := tx.Machine.WithContext(baton.Ctx).Where(tx.Machine.IP.Eq(v.Ip)).First()
						if err != nil {
							log.Logger.Info("没有查到配置的压力机信息：", " 机器ip为：", v.Ip)
							continue
						} else { // 查到了
							concurrencyNum := math.Ceil(float64(v.Weight) * float64(stress.ConfigTask.ModeConf.Concurrency) / 100)
							stress.IsRun = 1
							stress.ConfigTask.ModeConf.Concurrency = int64(concurrencyNum)
							addr := addrInfo.IP + fmt.Sprintf(":%d", addrInfo.Port)
							stress.Addr = addr
							stress.ExecConcurrency = int64(concurrencyNum)
							stressRun = append(stressRun, stress)
							reportMachineNumMap[stress.ReportID]++
						}
					}
				} else {
					for _, v := range stress.ConfigTask.MachineDispatchModeConf.UsableMachineList {
						addrInfo, err := tx.Machine.WithContext(baton.Ctx).Where(tx.Machine.IP.Eq(v.Ip)).First()
						if err != nil {
							log.Logger.Info("没有查到配置的压力机信息：", " 机器ip为：", v.Ip)
							continue
						} else { // 查到了
							stress.IsRun = 1
							stress.ConfigTask.ModeConf.StartConcurrency = int64(math.Ceil(float64(v.StartConcurrency) * float64(v.Weight) / 100))
							stress.ConfigTask.ModeConf.Step = int64(math.Ceil(float64(v.Step) * float64(v.Weight) / 100))
							stress.ConfigTask.ModeConf.StepRunTime = v.StepRunTime
							stress.ConfigTask.ModeConf.MaxConcurrency = int64(math.Ceil(float64(v.MaxConcurrency) * float64(v.Weight) / 100))
							stress.ConfigTask.ModeConf.Duration = v.Duration
							addr := addrInfo.IP + fmt.Sprintf(":%d", addrInfo.Port)
							stress.Addr = addr
							stress.ExecConcurrency = int64(math.Ceil(float64(v.MaxConcurrency) * float64(v.Weight) / 100))
							stressRun = append(stressRun, stress)
							reportMachineNumMap[stress.ReportID]++
						}
					}
				}
			} else { // 自定义
				if stress.ConfigTask.Mode == consts.PlanModeConcurrence || stress.ConfigTask.Mode == consts.PlanModeRound { // 并发模式或者轮次模式
					for _, v := range stress.ConfigTask.MachineDispatchModeConf.UsableMachineList {
						addrInfo, err := tx.Machine.WithContext(baton.Ctx).Where(tx.Machine.IP.Eq(v.Ip)).First()
						if err != nil {
							log.Logger.Info("没有查到配置的压力机信息：", " 机器ip为：", v.Ip)
							continue
						} else { // 查到了
							stress.IsRun = 1
							stress.ConfigTask.ModeConf.Concurrency = v.Concurrency
							stress.ConfigTask.ModeConf.Duration = v.Duration
							stress.ConfigTask.ModeConf.RoundNum = v.RoundNum
							addr := addrInfo.IP + fmt.Sprintf(":%d", addrInfo.Port)
							stress.Addr = addr
							stress.ExecConcurrency = v.Concurrency
							stressRun = append(stressRun, stress)
							reportMachineNumMap[stress.ReportID]++
						}
					}
				} else { // 非并发模式
					for _, v := range stress.ConfigTask.MachineDispatchModeConf.UsableMachineList {
						addrInfo, err := tx.Machine.WithContext(baton.Ctx).Where(tx.Machine.IP.Eq(v.Ip)).First()
						if err != nil {
							log.Logger.Info("没有查到配置的压力机信息：", " 机器ip为：", v.Ip)
							continue
						} else { // 查到了
							stress.IsRun = 1
							stress.ConfigTask.ModeConf.StartConcurrency = v.StartConcurrency
							stress.ConfigTask.ModeConf.Step = v.Step
							stress.ConfigTask.ModeConf.StepRunTime = v.StepRunTime
							stress.ConfigTask.ModeConf.MaxConcurrency = v.MaxConcurrency
							stress.ConfigTask.ModeConf.Duration = v.Duration
							addr := addrInfo.IP + fmt.Sprintf(":%d", addrInfo.Port)
							stress.Addr = addr
							stress.ExecConcurrency = v.MaxConcurrency
							stressRun = append(stressRun, stress)
							reportMachineNumMap[stress.ReportID]++
						}
					}
				}
			}
		} else {
			if curIndex == usableMachineNum {
				curIndex = 0
			}

			reportMachineNumMap[stress.ReportID]++

			// 获取当前任务的总并发数
			//oneSceneTotalConcurrency := GetOneTaskTotalGoroutines(stress, 0)

			var oneSceneTotalConcurrency int64 = 0
			if stress.ConfigTask.Mode == consts.PlanModeConcurrence {
				oneSceneTotalConcurrency = stress.ConfigTask.ModeConf.Concurrency
			} else {
				oneSceneTotalConcurrency = stress.ConfigTask.ModeConf.MaxConcurrency
			}

			if oneSceneTotalConcurrency > baton.MachineMaxConcurrence { // 当前任务的总并发数大于机器可用的总并发数
				log.Logger.Info("当前任务超出资源能力，不予执行，report_id为：", stress.ReportID, " 当前任务所需并发数：", oneSceneTotalConcurrency, " 当前所有机器可用并发数：", baton.MachineMaxConcurrence)
				stress.IsRun = 2
				stress.ExecConcurrency = oneSceneTotalConcurrency
				stressRun = append(stressRun, stress)
				continue
			}

			// 如果小于5000  直接分配、
			// 判断当前任务单个接口并发数是否超5000
			if stress.ConfigTask.Mode == consts.PlanModeConcurrence { // 并发模式
				if stress.ConfigTask.ModeConf.Concurrency <= int64(conf.Conf.OneMachineCanConcurrenceNum) {
					stress.IsRun = 1
					addr := baton.balance.GetMachine(curIndex)
					stress.Addr = addr
					stress.ExecConcurrency = oneSceneTotalConcurrency
					stressRun = append(stressRun, stress)
					curIndex++
					machineUsableGoroutines[addr] = machineUsableGoroutines[addr] - oneSceneTotalConcurrency
					baton.MachineMaxConcurrence = baton.MachineMaxConcurrence - oneSceneTotalConcurrency
					continue
				}
			} else { // 非并发模式
				if stress.ConfigTask.ModeConf.MaxConcurrency <= int64(conf.Conf.OneMachineCanConcurrenceNum) {
					stress.IsRun = 1
					addr := baton.balance.GetMachine(curIndex)
					stress.Addr = addr
					stress.ExecConcurrency = oneSceneTotalConcurrency
					stressRun = append(stressRun, stress)
					curIndex++
					machineUsableGoroutines[addr] = machineUsableGoroutines[addr] - oneSceneTotalConcurrency
					baton.MachineMaxConcurrence = baton.MachineMaxConcurrence - oneSceneTotalConcurrency
					continue
				}
			}

			// 任务并发数大于5000，判断当前机器是否满足
			addr := baton.balance.GetMachine(curIndex)
			if machineUsableGoroutines[addr] >= oneSceneTotalConcurrency {
				stress.IsRun = 1
				stress.Addr = addr
				stress.ExecConcurrency = oneSceneTotalConcurrency
				stressRun = append(stressRun, stress)
				curIndex++
				machineUsableGoroutines[addr] = machineUsableGoroutines[addr] - oneSceneTotalConcurrency
				baton.MachineMaxConcurrence = baton.MachineMaxConcurrence - oneSceneTotalConcurrency

				log.Logger.Info("运行计划--当前任务可以被当前机器运行：", addr, machineUsableGoroutines[addr], oneSceneTotalConcurrency)
				continue
			}

			// 当前机器满足不了当前任务的并发
			// 并发大于5000
			reportMachineNumMap[stress.ReportID]++
			var mNum int64 = 0
			var mNumReal int64 = 0
			var yuNum int64 = 0
			if stress.ConfigTask.Mode == consts.PlanModeConcurrence { // 并发模式
				mNum = stress.ConfigTask.ModeConf.Concurrency / int64(conf.Conf.OneMachineCanConcurrenceNum)
				mNumReal = mNum // 当前任务拆分后需要多少台机器
				yuNum = oneSceneTotalConcurrency % int64(conf.Conf.OneMachineCanConcurrenceNum)
			} else {
				mNum = stress.ConfigTask.ModeConf.MaxConcurrency / int64(conf.Conf.OneMachineCanConcurrenceNum)
				mNumReal = mNum // 当前任务拆分后需要多少台机器
				yuNum = oneSceneTotalConcurrency % int64(conf.Conf.OneMachineCanConcurrenceNum)
			}

			if yuNum > 0 {
				mNumReal = mNumReal + 1
			}

			// 判断当前任务需要的机器数是否超过总的机器数量
			if int(mNumReal) > len(baton.balance.rss) {
				return errno.ErrResourceNotEnough, errors.New("资源不足--当前计划中，某单个任务所需压力机数量，超过当前可用压力机数量")
			}

			oneSceneTotalConcurrencyTemp := GetOneTaskTotalGoroutines(stress, int64(conf.Conf.OneMachineCanConcurrenceNum))
			if stress.ConfigTask.Mode == consts.PlanModeConcurrence { // 并发模式
				for j := 0; j < int(mNum); j++ {
					stress.IsRun = 1
					stress.ConfigTask.ModeConf.Concurrency = int64(conf.Conf.OneMachineCanConcurrenceNum)
					addr := baton.balance.GetMachine(curIndex)
					stress.Addr = addr
					stress.ExecConcurrency = oneSceneTotalConcurrency
					stressRun = append(stressRun, stress)
					machineUsableGoroutines[addr] = machineUsableGoroutines[addr] - oneSceneTotalConcurrencyTemp
					baton.MachineMaxConcurrence = baton.MachineMaxConcurrence - oneSceneTotalConcurrencyTemp
					reportMachineNumMap[stress.ReportID]++
					curIndex++
				}
				if yuNum > 0 {
					stress.IsRun = 1
					stress.ConfigTask.ModeConf.Concurrency = yuNum
					addr := baton.balance.GetMachine(curIndex)
					stress.Addr = addr
					stress.ExecConcurrency = oneSceneTotalConcurrency
					stressRun = append(stressRun, stress)
					oneSceneTotalConcurrencyTemp = GetOneTaskTotalGoroutines(stress, yuNum)
					machineUsableGoroutines[addr] = machineUsableGoroutines[addr] - oneSceneTotalConcurrencyTemp
					baton.MachineMaxConcurrence = baton.MachineMaxConcurrence - oneSceneTotalConcurrencyTemp
					reportMachineNumMap[stress.ReportID]++
					curIndex++
				}

			} else {
				// 判断起始并发数是否大于所需机器数量
				if stress.ConfigTask.ModeConf.StartConcurrency < mNumReal {
					stress.ConfigTask.ModeConf.StartConcurrency = mNumReal
				}
				// 判断步长数
				if stress.ConfigTask.ModeConf.Step < mNumReal {
					stress.ConfigTask.ModeConf.Step = stress.ConfigTask.ModeConf.Step / mNumReal
				}

				// 起始并发计算
				var nowStartConcurrency int64 = 1
				var yuStartConcurrency int64 = 1
				nowStartConcurrency = stress.ConfigTask.ModeConf.StartConcurrency / mNumReal
				yuStartConcurrency = int64(math.Ceil(float64(stress.ConfigTask.ModeConf.StartConcurrency % mNumReal)))

				// 步长分配
				var nowStep int64 = 1
				var yuStep int64 = 1
				nowStep = stress.ConfigTask.ModeConf.Step / mNumReal
				yuStep = int64(math.Ceil(float64(stress.ConfigTask.ModeConf.Step % mNumReal)))

				for j := 0; j < int(mNum); j++ {
					stress.IsRun = 1
					stress.ConfigTask.ModeConf.StartConcurrency = nowStartConcurrency
					stress.ConfigTask.ModeConf.Step = nowStep
					stress.ConfigTask.ModeConf.MaxConcurrency = int64(conf.Conf.OneMachineCanConcurrenceNum)
					addr := baton.balance.GetMachine(curIndex)
					stress.Addr = addr
					stress.ExecConcurrency = oneSceneTotalConcurrency
					stressRun = append(stressRun, stress)
					machineUsableGoroutines[addr] = machineUsableGoroutines[addr] - oneSceneTotalConcurrencyTemp
					baton.MachineMaxConcurrence = baton.MachineMaxConcurrence - oneSceneTotalConcurrencyTemp
					reportMachineNumMap[stress.ReportID]++
					curIndex++
				}

				if yuNum > 0 {
					stress.IsRun = 1
					stress.ConfigTask.ModeConf.StartConcurrency = yuStartConcurrency
					stress.ConfigTask.ModeConf.Step = yuStep
					stress.ConfigTask.ModeConf.MaxConcurrency = yuNum
					addr := baton.balance.GetMachine(curIndex)
					stress.Addr = addr
					stress.ExecConcurrency = oneSceneTotalConcurrency
					stressRun = append(stressRun, stress)
					oneSceneTotalConcurrencyTemp = GetOneTaskTotalGoroutines(stress, yuNum)
					machineUsableGoroutines[addr] = machineUsableGoroutines[addr] - oneSceneTotalConcurrencyTemp
					baton.MachineMaxConcurrence = baton.MachineMaxConcurrence - oneSceneTotalConcurrencyTemp
					reportMachineNumMap[stress.ReportID]++
					curIndex++
				}
			}
		}
	}

	for k, stress := range stressRun {
		stressRun[k].MachineNum = reportMachineNumMap[stress.ReportID]
	}

	log.Logger.Info("当前计划总共被拆了几个任务：", len(stressRun))
	baton.stressRun = stressRun
	return s.next.Execute(baton)
}

func (s *SplitStress) SetNext(stress Stress) {
	s.next = stress
}

// GetOneTaskTotalGoroutines 获取单个任务总并发
func GetOneTaskTotalGoroutines(stress run_plan.Stress, concurrencyNum int64) int64 {
	// 接口总数
	var apiCnt int64 = 0
	for _, nodeArr := range stress.Scene.NodesRound {
		apiCnt = apiCnt + int64(len(nodeArr))
	}

	var oneSceneTotalConcurrency int64                        // 当前任务的总并发
	if stress.ConfigTask.Mode == consts.PlanModeConcurrence { // 并发模式
		if concurrencyNum == 0 {
			oneSceneTotalConcurrency = apiCnt * stress.ConfigTask.ModeConf.Concurrency
		} else {
			oneSceneTotalConcurrency = apiCnt * concurrencyNum
		}

	} else { // 其他模式
		if concurrencyNum == 0 {
			oneSceneTotalConcurrency = apiCnt * stress.ConfigTask.ModeConf.MaxConcurrency
		} else {
			oneSceneTotalConcurrency = apiCnt * concurrencyNum
		}

	}
	return oneSceneTotalConcurrency
}

type RunMachineStress struct {
	next Stress
}

func (s *RunMachineStress) Execute(baton *Baton) (int, error) {
	log.Logger.Info("运行计划--开始执行计划", baton.TeamID, baton.PlanID)
	// 当前可用压力机数量
	//machinesNum := len(baton.balance.rss)
	//curIndex := 0 // 当前使用的压力机数组下标

	// 报告ID与kafka分区映射
	partitionMap := make(map[string]int32, len(baton.stress))

	// 资源不足标志
	sourceNotEnough := 0

	for k, stress := range baton.stressRun {
		log.Logger.Info("当前计划执行的任务序号：", k, " 报告id:", stress.ReportID)
		// 获取分区
		if partitionNum, ok := partitionMap[stress.ReportID]; ok {
			stress.Partition = partitionNum
		} else {
			// 增加分区字段判断
			partition := GetPartition(baton, stress)
			if partition == -1 {
				log.Logger.Info("资源不足--当前没有可用的kafka分区")
				_ = DeletePlanReport(baton)
				_ = UpdateStressPlanStatus(baton, int32(consts.PlanStatusNormal))
				return errno.ErrResourceNotEnough, fmt.Errorf("资源不足--当前没有可用的kafka分区")
			}
			partitionMap[stress.ReportID] = partition
			stress.Partition = partition
		}

		if stress.IsRun == 2 { // 判断是否包含资源不足的任务
			log.Logger.Info("当前任务超出资源能力，不予执行，report_id为：", stress.ReportID)
			sourceNotEnough++
			continue
		}

		addr := stress.Addr
		machinesState := GetRunnerMachineState(baton, addr) // 获取当前压力机可用状态

		if machinesState { // 如果当前机器可用
			// 把当前机器信息写入到数据表当中
			tx := query.Use(dal.DB()).ReportMachine
			insertData := &model.ReportMachine{
				ReportID:    stress.ReportID,
				TeamID:      stress.TeamID,
				PlanID:      stress.PlanID,
				IP:          omnibus.Explode(":", addr)[0],
				Concurrency: stress.ExecConcurrency,
			}
			err := tx.WithContext(baton.Ctx).Create(insertData)
			if err != nil {
				_ = DeletePlanReport(baton)
				_ = UpdateStressPlanStatus(baton, int32(consts.PlanStatusNormal))
				log.Logger.Info("把报告和对应机器写入到数据库失败，err：", err)
				return errno.ErrMysqlFailed, err
			}
			runResponse, err := resty.New().R().SetBody(stress).Post(fmt.Sprintf("http://%s"+runner.EngineRunPlan, addr))
			log.Logger.Info("请求压力机运行情况，report_id:", stress.ReportID, " 压测机器IP为：", addr, " 运行参数为：", proof.Render("req", stress), " err:", err)
			log.Logger.Info("请求压力机返回结果，report_id:", stress.ReportID, " 压测机器IP为：", addr, " body为：", proof.Render("body", string(runResponse.Body())))
			if err != nil {
				_ = DeletePlanReport(baton)
				_ = UpdateStressPlanStatus(baton, int32(consts.PlanStatusNormal))
				log.Logger.Info("请求压力机进行压测失败，err：", err)
				return errno.ErrHttpFailed, err
			}

			// 把当前压力机使用状态设置到redis当中
			machineUseStateKey := consts.MachineUseStatePrefix + addr
			dal.GetRDB().SetNX(baton.Ctx, machineUseStateKey, 1, 3600*24)
			err = UpdateStressPlanStatus(baton, consts.PlanStatusUnderway)
			if err != nil {
				log.Logger.Info("修改计划状态失败，err：", err)
				return errno.ErrMysqlFailed, err
			}
		} else {
			sourceNotEnough++
		}
	}

	if sourceNotEnough > 0 {
		_ = DeletePlanReport(baton)
		err := UpdateStressPlanStatus(baton, int32(consts.PlanStatusNormal))
		if err != nil {
			log.Logger.Info("修改计划状态失败，err：", err)
		}
		return errno.ErrResourceNotEnough, errors.New("资源不足--压力机无法支持当前任务执行")
	}
	return errno.Ok, nil
}

func (s *RunMachineStress) SetNext(stress Stress) {
	s.next = stress
}

// GetPartition 获取可用分区
func GetPartition(baton *Baton, stress run_plan.Stress) int32 {
	//默认分区为0
	var partition int32 = -1 //默认为-1 表示不可用分区锁

	// 从redis查询当前
	StressBelongPartitionInfo := dal.GetRDB().HGetAll(baton.Ctx, consts.StressBelongPartitionKey)
	if StressBelongPartitionInfo.Err() != nil || len(StressBelongPartitionInfo.Val()) == 0 { // 没有查到消费者对应的分区数据
		return partition
	}

	// 已经使用的分数切片
	usedPartitionMap := make(map[int]int)

	for ip, partitionInfo := range StressBelongPartitionInfo.Val() {
		var tempData []int
		err := json.Unmarshal([]byte(partitionInfo), &tempData)
		if err != nil {
			log.Logger.Info("获取分区--解析消费者服务对应的分区数据失败，消费者IP:", ip, " 对应的分数数据：", partitionInfo)
			continue
		}
		if len(tempData) > 0 {
			for _, partitionNum := range tempData {
				usedPartitionMap[partitionNum] = 1
			}
		}
	}

	// kafka全局的报告分区key名
	partitionLock := consts.KafkaReportPartition

	for pNum := range usedPartitionMap {
		// 获取当前时间戳
		nowTime := time.Now().Unix()

		partitionValue := fmt.Sprintf("%s_%s_%s_%d", stress.TeamID, stress.PlanID, stress.ReportID, nowTime)

		// 把分区转换成字符串
		partitionNumString := strconv.Itoa(pNum)
		// 尝试获取当前分区锁
		res, _ := dal.GetRDB().HSetNX(baton.Ctx, partitionLock, partitionNumString, partitionValue).Result()
		if res == false { // 获取失败或者当前分区锁已被占用
			continue
		} else {
			partition = int32(pNum)
			break
		}
	}
	return partition
}

// GetRunnerMachineState 获取当前压力机是否可用
func GetRunnerMachineState(baton *Baton, addr string) bool {
	// 从Redis获取压力机列表
	machineListRes := dal.GetRDB().HGetAll(baton.Ctx, consts.MachineListRedisKey)
	if len(machineListRes.Val()) == 0 || machineListRes.Err() != nil {
		log.Logger.Info("当前没有上报上来的空闲压力机可用")
		return false
	}

	//// 查到了机器列表，然后判断可用性
	var runnerMachineInfo *HeartBeat
	// 初始化机器状态map
	machineStateMap := make(map[string]bool, len(machineListRes.Val()))
	for machineAddr, machineDetail := range machineListRes.Val() {
		// 退出循环的标识
		breakFor := false
		// 解析hash的field字段
		machineAddrSlice := strings.Split(machineAddr, "_")
		if len(machineAddrSlice) != 3 {
			continue
		}

		// 组装可用机器map的key
		addrKey := machineAddrSlice[0] + ":" + machineAddrSlice[1]
		machineStateMap[addrKey] = false

		// 把机器详情信息解析成格式化数据
		err := json.Unmarshal([]byte(machineDetail), &runnerMachineInfo)
		if err != nil {
			log.Logger.Info("压力机数据解析失败，err:", err)
			continue
		}

		// 压力机数据上报时间超过3秒，则认为服务不可用，不参与本次压力测试
		nowTime := time.Now().Unix()
		fmtNowTime := time.Now()
		if runnerMachineInfo.CreateTime < nowTime-int64(conf.Conf.MachineConfig.MachineAliveTime) {
			log.Logger.Info("资源不足--运行前最后验证机器状态，上报数据超时，当前时间为：", fmtNowTime, " 上报时间为：", runnerMachineInfo.FmtCreateTime)
			continue
		}

		// 判断当前压力机性能是否爆满,如果某个指标爆满，则不参与本次压力测试
		if runnerMachineInfo.CpuUsage >= float64(conf.Conf.MachineConfig.CpuTopLimit) { // CPU使用判断
			log.Logger.Info("资源不足--CPU超过指标,指标为：", runnerMachineInfo.CpuUsage, "机器信息为：", machineAddr)
			continue
		}
		for _, memInfo := range runnerMachineInfo.MemInfo { // 内存使用判断
			if memInfo.UsedPercent >= float64(conf.Conf.MachineConfig.MemoryTopLimit) {
				log.Logger.Info("资源不足--内存超过指标,指标为：", memInfo.UsedPercent, "机器信息为：", machineAddr)
				breakFor = true
				break
			}
		}
		for _, diskInfo := range runnerMachineInfo.DiskInfos { // 磁盘使用判断
			if diskInfo.UsedPercent >= float64(conf.Conf.MachineConfig.DiskTopLimit) {
				log.Logger.Info("资源不足--磁盘超过指标,指标为：", diskInfo.UsedPercent, "机器信息为：", machineAddr)
				breakFor = true
				break
			}
		}

		// 最后判断是否结束当前循环
		if breakFor {
			continue
		}

		// 当前机器可用协程数
		usableGoroutines := runnerMachineInfo.MaxGoroutines - runnerMachineInfo.CurrentGoroutines
		if usableGoroutines <= 0 {
			log.Logger.Info("资源不足--可用协程数过低,指标为：", usableGoroutines, "机器信息为：", machineAddr)
			continue
		}
		machineStateMap[addr] = true
	}

	if _, ok := machineStateMap[addr]; !ok {
		return false
	} else {
		return true
	}
}

// DeletePlanReport 删除执行失败的计划下的所有报告
func DeletePlanReport(baton *Baton) error {
	for _, reportInfo := range baton.reports {
		// 如果调用施压接口失败，则删除掉当前的这个报告id
		reportTable := dal.GetQuery().StressPlanReport
		_, err := reportTable.WithContext(baton.Ctx).Where(reportTable.TeamID.Eq(baton.TeamID), reportTable.ReportID.Eq(reportInfo.ReportID)).Delete()
		if err != nil {
			log.Logger.Info("运行计划--删除报告失败，报告id为：", reportInfo.ReportID)
		}

		// 同时删除掉该报告对应的机器信息
		rmTable := dal.GetQuery().ReportMachine
		_, err = rmTable.WithContext(baton.Ctx).Where(rmTable.TeamID.Eq(baton.TeamID), rmTable.PlanID.Eq(reportInfo.PlanID), rmTable.ReportID.Eq(reportInfo.ReportID)).Delete()
		if err != nil {
			log.Logger.Info("运行计划--删除报告对应机器表失败，团队id,计划id,报告id分别为：", reportInfo.TeamID, reportInfo.PlanID, reportInfo.ReportID)
		}

		// 清空新的报告id集合
		baton.ReportIDs = []string{}

	}
	return nil
}

func UpdateStressPlanStatus(baton *Baton, status int32) error {
	allErr := dal.GetQuery().Transaction(func(tx *query.Query) error {
		if status == consts.PlanStatusNormal { // 要改成未开始状态
			_, err := tx.StressPlan.WithContext(baton.Ctx).Where(tx.StressPlan.PlanID.Eq(baton.PlanID)).UpdateSimple(tx.StressPlan.Status.Value(status))
			if err != nil {
				log.Logger.Info("修改计划状态失败，err：", err)
				return err
			}
		} else {
			_, err := tx.StressPlan.WithContext(baton.Ctx).Where(tx.StressPlan.PlanID.Eq(baton.PlanID)).UpdateSimple(tx.StressPlan.Status.Value(status),
				tx.StressPlan.RunCount.Value(baton.plan.RunCount+1))
			if err != nil {
				log.Logger.Info("修改计划状态失败，err：", err)
				return err
			}
		}
		return nil
	})
	return allErr
}
