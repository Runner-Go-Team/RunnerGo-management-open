package plan

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/errno"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/jwt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/log"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/uuid"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/mao"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"gorm.io/gen"
	"gorm.io/gen/field"

	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/record"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/query"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/packer"
)

func ListByStatus(ctx context.Context, teamID string) (int, error) {
	runPlanNum := 0
	tx := query.Use(dal.DB()).StressPlan
	stressPlanList, err := tx.WithContext(ctx).Where(tx.TeamID.Eq(teamID), tx.Status.Eq(consts.PlanStatusUnderway)).Find()
	if err != nil {
		return 0, err
	}
	tx2 := dal.GetQuery().AutoPlan
	autoPlanList, err := tx2.WithContext(ctx).Where(tx2.TeamID.Eq(teamID), tx2.Status.Eq(consts.PlanStatusUnderway)).Find()
	if err != nil {
		return 0, err
	}
	runPlanNum = len(stressPlanList) + len(autoPlanList)
	return runPlanNum, nil
}

func ListByTeamID(ctx context.Context, req *rao.ListPlansReq) ([]*rao.StressPlan, int64, error) {
	tx := query.Use(dal.DB()).StressPlan
	conditions := make([]gen.Condition, 0)
	conditions = append(conditions, tx.TeamID.Eq(req.TeamID))
	if req.Keyword != "" {
		planIDs := make([]string, 0, req.Size)

		planIDs1, err := KeywordFindPlanForPlanName(ctx, req.TeamID, req.Keyword)
		if err == nil {
			planIDs = append(planIDs, planIDs1...)
		}

		planIDs2, err := KeywordFindPlanForUserName(ctx, req.TeamID, req.Keyword)
		if err == nil {
			planIDs = append(planIDs, planIDs2...)
		}

		if len(planIDs) > 0 {
			conditions = append(conditions, tx.PlanID.In(planIDs...))
		} else {
			conditions = append(conditions, tx.PlanID.In(""))
		}
	}

	if req.StartTimeSec != 0 && req.EndTimeSec != 0 {
		startTime := time.Unix(req.StartTimeSec, 0)
		endTime := time.Unix(req.EndTimeSec, 0)
		conditions = append(conditions, tx.CreatedAt.Between(startTime, endTime))
	}

	if req.TaskType > 0 {
		conditions = append(conditions, tx.TaskType.Eq(req.TaskType))
	}

	if req.TaskMode > 0 {
		conditions = append(conditions, tx.TaskMode.Eq(req.TaskMode))
	}

	if req.Status > 0 {
		conditions = append(conditions, tx.Status.Eq(req.Status))
	}

	sort := make([]field.Expr, 0)
	if req.Sort == 0 { // 默认排序
		sort = append(sort, tx.RankID.Desc())
	}
	if req.Sort == 1 { // 创建时间倒序
		sort = append(sort, tx.CreatedAt.Desc())
	}
	if req.Sort == 2 { // 创建时间正序
		sort = append(sort, tx.CreatedAt)
	}
	if req.Sort == 3 { // 修改时间倒序
		sort = append(sort, tx.UpdatedAt.Desc())
	}
	if req.Sort == 4 { // 修改时间正序
		sort = append(sort, tx.UpdatedAt)
	}

	offset := (req.Page - 1) * req.Size
	limit := req.Size
	ret, cnt, err := tx.WithContext(ctx).Where(conditions...).Order(sort...).FindByPage(offset, limit)

	if err != nil {
		return nil, 0, err
	}

	var userIDs []string
	for _, r := range ret {
		userIDs = append(userIDs, r.CreateUserID)
	}

	u := query.Use(dal.DB()).User
	users, err := u.WithContext(ctx).Where(u.UserID.In(userIDs...)).Find()
	if err != nil {
		return nil, 0, err
	}

	return packer.TransPlansToRaoPlanList(ret, users), cnt, nil
}

func KeywordFindPlanForPlanName(ctx context.Context, teamID string, keyword string) ([]string, error) {
	planIDs := make([]string, 0, 100)

	p := dal.GetQuery().StressPlan
	err := p.WithContext(ctx).Where(p.TeamID.Eq(teamID), p.PlanName.Like(fmt.Sprintf("%%%s%%", keyword))).Pluck(p.PlanID, &planIDs)
	if err != nil {
		return planIDs, err
	}

	return planIDs, nil
}

func KeywordFindPlanForUserName(ctx context.Context, teamID string, keyword string) ([]string, error) {
	userIDs := make([]string, 0, 100)

	u := query.Use(dal.DB()).User
	err := u.WithContext(ctx).Where(u.Nickname.Like(fmt.Sprintf("%%%s%%", keyword))).Pluck(u.UserID, &userIDs)
	if err != nil {
		return nil, err
	}

	if len(userIDs) == 0 {
		return nil, fmt.Errorf("没有查到任何用户信息")
	}

	planIDs := make([]string, 0, 100)
	sp := dal.GetQuery().StressPlan
	err = sp.WithContext(ctx).Where(sp.TeamID.Eq(teamID), sp.RunUserID.In(userIDs...)).Pluck(sp.PlanID, &planIDs)
	if err != nil {
		return nil, err
	}

	return planIDs, nil
}

func Save(ctx *gin.Context, req *rao.SavePlanReq) (string, int, error) {
	if req.PlanName == "" {
		return req.PlanID, errno.ErrPlanNameNotEmpty, fmt.Errorf("计划名称不能为空")
	}

	// 用户信息
	userID := jwt.GetUserIDByCtx(ctx)
	planID := req.PlanID
	var rankID int64 = 1
	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		// 查询当前计划名称是否存在
		_, err := tx.StressPlan.WithContext(ctx).Where(tx.StressPlan.TeamID.Eq(req.TeamID),
			tx.StressPlan.PlanName.Eq(req.PlanName), tx.StressPlan.PlanID.Neq(req.PlanID)).First()
		if err != nil && err != gorm.ErrRecordNotFound {
			log.Logger.Info("新建性能计划失败，err:", err)
			return err
		}

		if err == nil { // 查到了
			return fmt.Errorf("名称已存在")
		}

		// 判断是否传了plan_id
		if req.PlanID == "" { // 新建计划
			// 查询当前团队下最大的plan_id数
			StressPlanInfo, err := tx.StressPlan.WithContext(ctx).Where(tx.StressPlan.TeamID.Eq(req.TeamID)).Order(tx.StressPlan.RankID.Desc()).Limit(1).First()
			if err != nil && err != gorm.ErrRecordNotFound {
				return err
			}
			if err == nil {
				rankID = StressPlanInfo.RankID + 1
			}

			planID = uuid.GetUUID()

			// 不存在，则创建数据
			insertData := &model.StressPlan{
				PlanName:     req.PlanName,
				PlanID:       planID,
				RankID:       rankID,
				TeamID:       req.TeamID,
				CreateUserID: userID,
				RunUserID:    userID,
				Status:       consts.PlanStatusNormal,
				Remark:       req.Remark,
				TaskType:     req.TaskType,
			}

			err = tx.StressPlan.WithContext(ctx).Create(insertData)
			if err != nil {
				return err
			}
			if err := record.InsertCreate(ctx, req.TeamID, userID, record.OperationOperateCreatePlan, req.PlanName); err != nil {
				return err
			}
		} else { // 修改计划
			_, err := tx.StressPlan.WithContext(ctx).Where(tx.StressPlan.TeamID.Eq(req.TeamID),
				tx.StressPlan.PlanID.Eq(req.PlanID)).UpdateSimple(tx.StressPlan.PlanName.Value(req.PlanName), tx.StressPlan.Remark.Value(req.Remark))
			if err != nil {
				return err
			}
			if err := record.InsertUpdate(ctx, req.TeamID, userID, record.OperationOperateUpdatePlan, req.PlanName); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		if err.Error() == "名称已存在" {
			return "", errno.ErrPlanNameAlreadyExist, err
		}
		return "", errno.ErrMysqlFailed, err
	}

	return planID, errno.Ok, nil
}

func SaveTask(ctx *gin.Context, req *rao.SavePlanConfReq) (int, error) {
	userID := jwt.GetUserIDByCtx(ctx)

	// 判断任务配置类型
	var err error

	if req.TaskType == consts.PlanTaskTypeNormal { // 普通任务
		err = query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
			// 1、先去把定时任务数据删掉
			_, err := tx.StressPlanTimedTaskConf.WithContext(ctx).
				Where(tx.StressPlanTimedTaskConf.TeamID.Eq(req.TeamID)).
				Where(tx.StressPlanTimedTaskConf.PlanID.Eq(req.PlanID)).
				Where(tx.StressPlanTimedTaskConf.SceneID.Eq(req.SceneID)).Delete()
			if err != nil {
				log.Logger.Info("保存配置--不存在定时任务或删除mysql失败,err:", err)
			}

			// 查询是否存在配置
			_, err = tx.StressPlanTaskConf.WithContext(ctx).
				Where(tx.StressPlanTaskConf.TeamID.Eq(req.TeamID),
					tx.StressPlanTaskConf.PlanID.Eq(req.PlanID),
					tx.StressPlanTaskConf.SceneID.Eq(req.SceneID)).First()
			if err != nil && err != gorm.ErrRecordNotFound {
				return err
			}

			// 压缩任务配置详情
			modeConfString, err2 := json.Marshal(*req.ModeConf)
			if err2 != nil {
				log.Logger.Info("保存任务配置--任务配置数据压缩失败")
				return err2
			}

			// 压缩分布式压力机列表数据
			machineDispatchModeConfString, err3 := json.Marshal(req.MachineDispatchModeConf)
			if err3 != nil {
				log.Logger.Info("保存任务配置--分布式压力机数据压缩失败")
				return err3
			}

			if err == nil { // 已存在 则修改
				_, err = tx.StressPlanTaskConf.WithContext(ctx).
					Where(tx.StressPlanTaskConf.TeamID.Eq(req.TeamID),
						tx.StressPlanTaskConf.PlanID.Eq(req.PlanID),
						tx.StressPlanTaskConf.SceneID.Eq(req.SceneID)).UpdateSimple(
					tx.StressPlanTaskConf.TaskType.Value(req.TaskType),
					tx.StressPlanTaskConf.TaskMode.Value(req.Mode),
					tx.StressPlanTaskConf.ControlMode.Value(req.ControlMode),
					tx.StressPlanTaskConf.DebugMode.Value(req.DebugMode),
					tx.StressPlanTaskConf.ModeConf.Value(string(modeConfString)),
					tx.StressPlanTaskConf.IsOpenDistributed.Value(req.IsOpenDistributed),
					tx.StressPlanTaskConf.MachineDispatchModeConf.Value(string(machineDispatchModeConfString)),
					tx.StressPlanTaskConf.RunUserID.Value(userID),
				)
				if err != nil {
					return err
				}
			} else { // 不存在，则新增
				insertData := &model.StressPlanTaskConf{
					PlanID:                  req.PlanID,
					TeamID:                  req.TeamID,
					SceneID:                 req.SceneID,
					TaskType:                req.TaskType,
					TaskMode:                req.Mode,
					ControlMode:             req.ControlMode,
					DebugMode:               req.DebugMode,
					ModeConf:                string(modeConfString),
					IsOpenDistributed:       req.IsOpenDistributed,
					MachineDispatchModeConf: string(machineDispatchModeConfString),
					RunUserID:               userID,
				}
				err = tx.StressPlanTaskConf.WithContext(ctx).Create(insertData)
				if err != nil {
					return err
				}
			}
			return err
		})
	} else { // 定时任务
		err = query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
			// 1、先把对应的普通任务删掉
			_, err := tx.StressPlanTaskConf.WithContext(ctx).
				Where(tx.StressPlanTaskConf.TeamID.Eq(req.TeamID),
					tx.StressPlanTaskConf.PlanID.Eq(req.PlanID),
					tx.StressPlanTaskConf.SceneID.Eq(req.SceneID)).Delete()
			if err != nil {
				return err
			}

			// 把定时任务保存到数据库中
			// 查询当前定时任务是否存在
			_, err = tx.StressPlanTimedTaskConf.WithContext(ctx).
				Where(tx.StressPlanTimedTaskConf.TeamID.Eq(req.TeamID)).
				Where(tx.StressPlanTimedTaskConf.PlanID.Eq(req.PlanID)).
				Where(tx.StressPlanTimedTaskConf.SceneID.Eq(req.SceneID)).First()
			if err != nil && err != gorm.ErrRecordNotFound { // 查询出错
				log.Logger.Info("保存配置--查询定时任务数据失败，err:", req)
				return err
			} else if err == gorm.ErrRecordNotFound { // 数据不存在
				// 新增配置
				timingTaskConfig, err := packer.TransSaveTimingTaskConfigReqToModelData(req, userID)
				if err != nil {
					log.Logger.Info("保存配置--压缩mode_conf为字符串时失败", err)
					return err
				}
				err = tx.StressPlanTimedTaskConf.WithContext(ctx).Create(timingTaskConfig)
				if err != nil {
					log.Logger.Info("保存配置--定时任务配置项保存失败，err：", err)
					return err
				}
			} else {
				// 把mode_conf压缩成字符串
				modeConfString, err := json.Marshal(req.ModeConf)
				if err != nil {
					log.Logger.Info("保存配置--压缩mode_conf为字符串时失败", err)
					return err
				}

				// 把mode_conf压缩成字符串
				machineDispatchModeConfString, err := json.Marshal(req.MachineDispatchModeConf)
				if err != nil {
					return err
				}

				// 修改配置
				_, err = tx.StressPlanTimedTaskConf.WithContext(ctx).Where(tx.StressPlanTimedTaskConf.TeamID.Eq(req.TeamID)).
					Where(tx.StressPlanTimedTaskConf.PlanID.Eq(req.PlanID)).
					Where(tx.StressPlanTimedTaskConf.SceneID.Eq(req.SceneID)).UpdateSimple(
					tx.StressPlanTimedTaskConf.UserID.Value(userID),
					tx.StressPlanTimedTaskConf.Frequency.Value(req.TimedTaskConf.Frequency),
					tx.StressPlanTimedTaskConf.TaskExecTime.Value(req.TimedTaskConf.TaskExecTime),
					tx.StressPlanTimedTaskConf.TaskCloseTime.Value(req.TimedTaskConf.TaskCloseTime),
					tx.StressPlanTimedTaskConf.TaskMode.Value(req.Mode),
					tx.StressPlanTimedTaskConf.ControlMode.Value(req.ControlMode),
					tx.StressPlanTimedTaskConf.DebugMode.Value(req.DebugMode),
					tx.StressPlanTimedTaskConf.ModeConf.Value(string(modeConfString)),
					tx.StressPlanTimedTaskConf.IsOpenDistributed.Value(req.IsOpenDistributed),
					tx.StressPlanTimedTaskConf.MachineDispatchModeConf.Value(string(machineDispatchModeConfString)),
					tx.StressPlanTimedTaskConf.Status.Value(consts.TimedTaskWaitEnable),
					tx.StressPlanTimedTaskConf.RunUserID.Value(userID),
				)
				if err != nil {
					log.Logger.Info("保存配置--更新定时任务配置失败，err:", err)
					return err
				}
			}
			// 事务的返回
			return nil
		})
	}
	if err != nil {
		log.Logger.Info("保存配置--保存普通任务配置失败，err:", err)
		return errno.ErrMysqlFailed, err
	}

	// 修改计划压测模式
	tx := dal.GetQuery()
	var planMode int32 = 0
	if req.TaskType == consts.PlanTaskTypeNormal { // 普通任务
		tasks, err := tx.StressPlanTaskConf.WithContext(ctx).
			Where(tx.StressPlanTaskConf.TeamID.Eq(req.TeamID),
				tx.StressPlanTaskConf.PlanID.Eq(req.PlanID)).Find()
		if err != nil {
			log.Logger.Info("保存配置--查询当前计划下是否有定时任务时出错，err:", err)
			return errno.ErrMysqlFailed, err
		}
		if len(tasks) > 0 {
			// 模式
			planMode = tasks[0].TaskMode
			for i, t := range tasks {
				if i > 0 && t.TaskMode != planMode && planMode != 0 {
					planMode = consts.PlanModeMix
					break
				}
			}
		}
	} else {
		// 查询当前计划下是否有定时任务
		timedTaskInfo, err := tx.StressPlanTimedTaskConf.WithContext(ctx).Where(tx.StressPlanTimedTaskConf.TeamID.Eq(req.TeamID),
			tx.StressPlanTimedTaskConf.PlanID.Eq(req.PlanID)).First()
		if err != nil {
			log.Logger.Info("保存配置--查询当前计划下是否有定时任务时出错，err:", err)
			return errno.ErrMysqlFailed, err
		}

		planMode = timedTaskInfo.TaskMode
	}

	_, err = tx.StressPlan.WithContext(ctx).Where(tx.StressPlan.TeamID.Eq(req.TeamID), tx.StressPlan.PlanID.Eq(req.PlanID)).UpdateSimple(tx.StressPlan.TaskMode.Value(planMode))
	if err != nil {
		log.Logger.Info("保存配置--修改计划的任务类型和也测模式失败，err:", err)
		return errno.ErrMysqlFailed, err
	}
	// 最后的返回
	return errno.Ok, nil
}

func GetPlanTask(ctx context.Context, req *rao.GetPlanTaskReq) (*rao.PlanTaskResp, error) {
	// 初始化返回值
	planTaskConf := &rao.PlanTaskResp{
		PlanID:            req.PlanID,
		SceneID:           req.SceneID,
		TaskType:          req.TaskType,
		Mode:              consts.PlanModeConcurrence,
		ModeConf:          rao.ModeConf{},
		TimedTaskConf:     rao.TimedTaskConf{},
		IsOpenDistributed: 0,
	}

	tx := dal.GetQuery()
	// 获取当前所有可用机器列表
	allMachineList, err := tx.Machine.WithContext(ctx).Where(tx.Machine.Status.Eq(consts.MachineStatusAvailable)).Find()
	if err != nil {
		return nil, err
	}

	defaultUsableMachineList := make([]rao.UsableMachineInfo, 0, len(allMachineList))
	allMachineMap := make(map[string]rao.UsableMachineInfo, len(allMachineList))
	for k, v := range allMachineList {
		weight := 0
		if k == 0 {
			weight = 100
		}
		temp := rao.UsableMachineInfo{
			MachineStatus:  v.Status,
			MachineName:    v.Name,
			Region:         v.Region,
			Ip:             v.IP,
			Weight:         weight,
			CreatedTimeSec: v.CreatedAt.Unix(),
		}
		allMachineMap[v.IP] = temp
		defaultUsableMachineList = append(defaultUsableMachineList, temp)
	}
	planTaskConf.MachineDispatchModeConf.UsableMachineList = defaultUsableMachineList

	if req.TaskType == consts.PlanTaskTypeNormal { // 普通任务
		taskConfInfo, err := tx.StressPlanTaskConf.WithContext(ctx).
			Where(tx.StressPlanTaskConf.TeamID.Eq(req.TeamID), tx.StressPlanTaskConf.PlanID.Eq(req.PlanID),
				tx.StressPlanTaskConf.SceneID.Eq(req.SceneID)).First()
		if err == nil { // 查到了，普通任务
			// 解析配置详情
			taskConfDetail := rao.ModeConf{}
			err := json.Unmarshal([]byte(taskConfInfo.ModeConf), &taskConfDetail)
			if err != nil {
				log.Logger.Info("获取配置详情--解析数据失败")
				return nil, err
			}

			// 解析分布式配置详情
			machineDispatchModeConfDetail := rao.MachineDispatchModeConf{}
			if taskConfInfo.MachineDispatchModeConf != "" {
				err = json.Unmarshal([]byte(taskConfInfo.MachineDispatchModeConf), &machineDispatchModeConfDetail)
				if err != nil {
					log.Logger.Info("获取配置详情--解析分布式配置数据失败")
					return nil, err
				}
			}

			usableMachineList := make([]rao.UsableMachineInfo, 0, len(machineDispatchModeConfDetail.UsableMachineList))
			if len(machineDispatchModeConfDetail.UsableMachineList) == 0 {
				usableMachineList = defaultUsableMachineList
			} else {
				for _, v := range machineDispatchModeConfDetail.UsableMachineList {
					temp := rao.UsableMachineInfo{
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

					// 判断配置过的压力机是否在全部压力机列表里面
					if machineInfo, ok := allMachineMap[v.Ip]; ok {
						temp.MachineStatus = machineInfo.MachineStatus
						allMachineMap[v.Ip] = temp
					} else {
						temp.MachineStatus = 2 // 机器不可用
						allMachineMap[v.Ip] = temp
					}
				}
				for _, v := range allMachineMap {
					usableMachineList = append(usableMachineList, v)
				}
			}

			// 组装接口返回值
			planTaskConf = &rao.PlanTaskResp{
				PlanID:      req.PlanID,
				SceneID:     req.SceneID,
				TaskType:    req.TaskType,
				Mode:        taskConfInfo.TaskMode,
				ControlMode: taskConfInfo.ControlMode,
				DebugMode:   taskConfInfo.DebugMode,
				ModeConf: rao.ModeConf{
					RoundNum:         taskConfDetail.RoundNum,
					Concurrency:      taskConfDetail.Concurrency,
					ThresholdValue:   taskConfDetail.ThresholdValue,
					StartConcurrency: taskConfDetail.StartConcurrency,
					Step:             taskConfDetail.Step,
					StepRunTime:      taskConfDetail.StepRunTime,
					MaxConcurrency:   taskConfDetail.MaxConcurrency,
					Duration:         taskConfDetail.Duration,
				},
				IsOpenDistributed: taskConfInfo.IsOpenDistributed,
				MachineDispatchModeConf: rao.MachineDispatchModeConf{
					MachineAllotType:  machineDispatchModeConfDetail.MachineAllotType,
					UsableMachineList: usableMachineList,
				},
			}
		}
	} else { // 定时任务
		// 查询定时任务信息
		timingTaskConfigInfo, err := tx.StressPlanTimedTaskConf.WithContext(ctx).Where(tx.StressPlanTimedTaskConf.TeamID.Eq(req.TeamID),
			tx.StressPlanTimedTaskConf.PlanID.Eq(req.PlanID),
			tx.StressPlanTimedTaskConf.SceneID.Eq(req.SceneID)).First()
		if err == nil {
			var modeConf rao.ModeConf
			err := json.Unmarshal([]byte(timingTaskConfigInfo.ModeConf), &modeConf)
			if err != nil {
				log.Logger.Info("获取任务配置详情--解析定时任务详细配置失败，err:", err)
				return nil, err
			}

			// 解析分布式配置详情
			machineDispatchModeConfDetail := rao.MachineDispatchModeConf{}
			if timingTaskConfigInfo.MachineDispatchModeConf != "" {
				err = json.Unmarshal([]byte(timingTaskConfigInfo.MachineDispatchModeConf), &machineDispatchModeConfDetail)
				if err != nil {
					log.Logger.Info("获取配置详情--解析分布式配置数据失败")
					return nil, err
				}
			}
			usableMachineList := make([]rao.UsableMachineInfo, 0, len(machineDispatchModeConfDetail.UsableMachineList))
			if len(machineDispatchModeConfDetail.UsableMachineList) == 0 {
				usableMachineList = defaultUsableMachineList
			} else {
				for _, v := range machineDispatchModeConfDetail.UsableMachineList {
					temp := rao.UsableMachineInfo{
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

					// 判断配置过的压力机是否在全部压力机列表里面
					if machineInfo, ok := allMachineMap[v.Ip]; ok {
						temp.MachineStatus = machineInfo.MachineStatus
						allMachineMap[v.Ip] = temp
					} else {
						temp.MachineStatus = 2 // 机器不可用
						allMachineMap[v.Ip] = temp
					}
				}
				for _, v := range allMachineMap {
					usableMachineList = append(usableMachineList, v)
				}
			}

			planTaskConf = &rao.PlanTaskResp{
				PlanID:      req.PlanID,
				SceneID:     req.SceneID,
				TaskType:    req.TaskType,
				Mode:        timingTaskConfigInfo.TaskMode,
				ControlMode: timingTaskConfigInfo.ControlMode,
				DebugMode:   timingTaskConfigInfo.DebugMode,
				ModeConf:    modeConf,
				TimedTaskConf: rao.TimedTaskConf{
					Frequency:     timingTaskConfigInfo.Frequency,
					TaskExecTime:  timingTaskConfigInfo.TaskExecTime,
					TaskCloseTime: timingTaskConfigInfo.TaskCloseTime,
				},
				IsOpenDistributed: timingTaskConfigInfo.IsOpenDistributed,
				MachineDispatchModeConf: rao.MachineDispatchModeConf{
					MachineAllotType:  machineDispatchModeConfDetail.MachineAllotType,
					UsableMachineList: usableMachineList,
				},
			}

			if timingTaskConfigInfo.Frequency == 0 { // 频次一次
				planTaskConf.TimedTaskConf.TaskCloseTime = 0
			}
		}
	}

	return planTaskConf, nil
}

func GetByPlanID(ctx context.Context, teamID string, planID string) (*rao.StressPlan, error) {
	tx := dal.GetQuery().StressPlan
	planInfo, err := tx.WithContext(ctx).Where(tx.TeamID.Eq(teamID), tx.PlanID.Eq(planID)).First()
	if err != nil {
		return nil, err
	}

	// 查询用户信息
	u := query.Use(dal.DB()).User
	user, err := u.WithContext(ctx).Where(u.UserID.Eq(planInfo.CreateUserID)).First()
	if err != nil {
		return nil, err
	}

	// 查询配置信息
	taskConfTable := dal.GetQuery().StressPlanTaskConf
	taskConfInfo, err := taskConfTable.WithContext(ctx).Where(taskConfTable.TeamID.Eq(teamID), taskConfTable.PlanID.Eq(planID)).Order(taskConfTable.SceneID).First()
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	var taskConf rao.ModeConf
	if err == nil {
		err := json.Unmarshal([]byte(taskConfInfo.ModeConf), &taskConf)
		if err != nil {
			log.Logger.Info("性能计划--任务配置数据解析失败，配置为：", taskConfInfo.ModeConf)
		}
	}

	return packer.TransTaskToRaoPlan(planInfo, taskConf, user), nil
}

func DeleteByPlanID(ctx context.Context, teamID string, planID string, userID string) error {
	return dal.GetQuery().Transaction(func(tx *query.Query) error {
		planInfo, err := tx.StressPlan.WithContext(ctx).Where(tx.StressPlan.TeamID.Eq(teamID), tx.StressPlan.PlanID.Eq(planID)).First()
		if err != nil {
			return err
		}

		if planInfo.Status == consts.PlanStatusUnderway {
			return fmt.Errorf("不能删除正在运行的计划")
		}

		// 删除计划信息
		if _, err := tx.StressPlan.WithContext(ctx).Where(tx.StressPlan.TeamID.Eq(teamID),
			tx.StressPlan.PlanID.Eq(planID)).Delete(); err != nil {
			return err
		}

		// 获取所有计划下的场景
		sceneList, err := tx.Target.WithContext(ctx).Where(tx.Target.TeamID.Eq(teamID), tx.Target.PlanID.Eq(planID),
			tx.Target.Source.Eq(consts.TargetSourcePlan)).Find()
		if err != nil {
			return err
		}
		//删除所有场景内的接口详情
		if len(sceneList) > 0 {
			sceneIDs := make([]string, 0, len(sceneList))
			for _, sceneInfo := range sceneList {
				sceneIDs = append(sceneIDs, sceneInfo.TargetID)
			}

			// 删除场景下的flow
			collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectFlow)
			_, err = collection.DeleteMany(ctx, bson.D{{"team_id", teamID}, {"scene_id", bson.D{{"$in", sceneIDs}}}})
			if err != nil {
				return err
			}

			// 删除场景下的变量
			_, err = tx.Variable.WithContext(ctx).Where(tx.Variable.SceneID.In(sceneIDs...)).Delete()
			if err != nil {
				return err
			}

			// 删除场景下的导入变量
			_, err = tx.VariableImport.WithContext(ctx).Where(tx.VariableImport.SceneID.In(sceneIDs...)).Delete()
			if err != nil {
				return err
			}
		}

		// 删除计划下场景
		if _, err = tx.Target.WithContext(ctx).Where(tx.Target.TeamID.Eq(teamID), tx.Target.PlanID.Eq(planID),
			tx.Target.Source.Eq(consts.TargetSourcePlan)).Delete(); err != nil {
			return err
		}

		if planInfo.TaskType == consts.PlanTaskTypeNormal {
			// 删除计划下所有普通任务配置
			if _, err = tx.StressPlanTaskConf.WithContext(ctx).Where(tx.StressPlanTaskConf.TeamID.Eq(teamID),
				tx.StressPlanTaskConf.PlanID.Eq(planID)).Delete(); err != nil {
				return err
			}
		} else {
			// 删除计划下所有定时任务配置
			if _, err = tx.StressPlanTimedTaskConf.WithContext(ctx).Where(tx.StressPlanTimedTaskConf.TeamID.Eq(teamID),
				tx.StressPlanTimedTaskConf.PlanID.Eq(planID)).Delete(); err != nil {
				return err
			}
		}
		return record.InsertDelete(ctx, teamID, userID, record.OperationOperateDeletePlan, planInfo.PlanName)
	})
}

func ClonePlan(ctx context.Context, req *rao.ClonePlanReq, userID string) error {
	return dal.GetQuery().Transaction(func(tx *query.Query) error {
		//克隆计划
		oldPlanInfo, err := tx.StressPlan.WithContext(ctx).Where(tx.StressPlan.TeamID.Eq(req.TeamID), tx.StressPlan.PlanID.Eq(req.PlanID)).First()
		if err != nil {
			return err
		}

		oldPlanName := oldPlanInfo.PlanName
		newPlanName := oldPlanName + "_1"

		// 查询老配置相关的
		list, err := tx.StressPlan.WithContext(ctx).Where(tx.StressPlan.TeamID.Eq(req.TeamID)).Where(tx.StressPlan.PlanName.Like(fmt.Sprintf("%s%%", oldPlanName+"_"))).Find()
		if err == nil {
			// 有复制过得配置
			maxNum := 0
			for _, autoPlanInfo := range list {
				nameTmp := autoPlanInfo.PlanName
				postfixSlice := strings.Split(nameTmp, "_")
				if len(postfixSlice) < 2 {
					continue
				}
				currentNum, err := strconv.Atoi(postfixSlice[len(postfixSlice)-1])
				if err != nil {
					log.Logger.Info("复制性能计划--类型转换失败，err:", err)
					continue
				}
				if currentNum > maxNum {
					maxNum = currentNum
				}
			}
			newPlanName = oldPlanName + fmt.Sprintf("_%d", maxNum+1)
		}

		// 查询当前团队内的计划最大
		newPlanID := uuid.GetUUID()
		var rankID int64 = 1
		stressPlanInfo, err := tx.StressPlan.WithContext(ctx).Where(tx.StressPlan.TeamID.Eq(req.TeamID)).Order(tx.StressPlan.RankID.Desc()).Limit(1).First()
		if err == nil { // 查到了
			rankID = stressPlanInfo.RankID + 1
		}

		oldPlanInfo.ID = 0
		oldPlanInfo.PlanID = newPlanID
		oldPlanInfo.RankID = rankID
		oldPlanInfo.PlanName = newPlanName
		oldPlanInfo.CreatedAt = time.Now()
		oldPlanInfo.UpdatedAt = time.Now()
		oldPlanInfo.Status = consts.PlanStatusNormal
		oldPlanInfo.CreateUserID = userID
		oldPlanInfo.RunUserID = userID
		if err := tx.StressPlan.WithContext(ctx).Create(oldPlanInfo); err != nil {
			return err
		}
		// 克隆场景，分组
		targets, err := tx.Target.WithContext(ctx).Where(tx.Target.TeamID.Eq(req.TeamID),
			tx.Target.PlanID.Eq(req.PlanID), tx.Target.Source.Eq(consts.TargetSourcePlan),
			tx.Target.Status.Eq(consts.TargetStatusNormal)).Order(tx.Target.ParentID).Find()
		if err != nil {
			return err
		}

		var sceneIDs []string
		targetMemo := make(map[string]string)
		for _, target := range targets {
			if target.TargetType == consts.TargetTypeScene {
				sceneIDs = append(sceneIDs, target.TargetID)
			}

			oldTargetID := target.TargetID

			newSceneID := uuid.GetUUID()
			target.ID = 0
			target.TargetID = newSceneID
			target.ParentID = targetMemo[target.ParentID]
			target.PlanID = newPlanID
			target.CreatedAt = time.Now()
			target.UpdatedAt = time.Now()
			if err := tx.Target.WithContext(ctx).Create(target); err != nil {
				return err
			}

			targetMemo[oldTargetID] = newSceneID
		}

		if len(sceneIDs) > 0 {
			// 克隆场景变量
			for _, sceneID := range sceneIDs {
				collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectSceneParam)
				cur, err := collection.Find(ctx, bson.D{{"team_id", req.TeamID}, {"scene_id", sceneID}})
				var sceneParamDataArr []*mao.SceneParamData
				if err == nil {
					if err := cur.All(ctx, &sceneParamDataArr); err != nil {
						return fmt.Errorf("场景参数数据获取失败")
					}
					for _, sv := range sceneParamDataArr {
						sv.SceneID = targetMemo[sceneID]
						if _, err := collection.InsertOne(ctx, sv); err != nil {
							return err
						}
					}
				}
			}

			// 克隆导入变量
			vi, err := tx.VariableImport.WithContext(ctx).Where(tx.VariableImport.SceneID.In(sceneIDs...)).Find()
			if err != nil {
				return err
			}

			for _, variableImport := range vi {
				variableImport.ID = 0
				variableImport.SceneID = targetMemo[variableImport.SceneID]
				variableImport.CreatedAt = time.Now()
				variableImport.UpdatedAt = time.Now()
				if err := tx.VariableImport.WithContext(ctx).Create(variableImport); err != nil {
					return err
				}
			}

			// 克隆流程
			var flows []*mao.Flow
			c1 := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectFlow)
			cur, err := c1.Find(ctx, bson.D{{"scene_id", bson.D{{"$in", sceneIDs}}}})
			if err != nil {
				return err
			}
			if err := cur.All(ctx, &flows); err != nil {
				return err
			}
			for _, flow := range flows {
				flow.SceneID = targetMemo[flow.SceneID]
				// 更新api的uuid
				err := packer.ChangeSceneNodeUUID(flow)
				if err != nil {
					log.Logger.Info("复制性能计划--替换event_id失败")
					return err
				}
				if _, err := c1.InsertOne(ctx, flow); err != nil {
					return err
				}
			}
		}

		// 克隆任务配置
		if oldPlanInfo.TaskType == consts.PlanTaskTypeNormal {
			// 查询老的场景对应配置
			oldTaskConfList, err := tx.StressPlanTaskConf.WithContext(ctx).Where(tx.StressPlanTaskConf.TeamID.Eq(req.TeamID),
				tx.StressPlanTaskConf.PlanID.Eq(req.PlanID)).Find()
			if err == nil && len(oldTaskConfList) > 0 {
				insertData := make([]*model.StressPlanTaskConf, 0, len(oldTaskConfList))
				for _, taskInfo := range oldTaskConfList {
					taskInfo.ID = 0
					taskInfo.PlanID = newPlanID
					taskInfo.SceneID = targetMemo[taskInfo.SceneID]
					taskInfo.RunUserID = userID
					taskInfo.CreatedAt = time.Now()
					taskInfo.UpdatedAt = time.Now()
					insertData = append(insertData, taskInfo)
				}
				err := tx.StressPlanTaskConf.WithContext(ctx).CreateInBatches(insertData, 10)
				if err != nil {
					return err
				}
			}
		} else {
			// 克隆定时任务
			timedTaskList, err := tx.StressPlanTimedTaskConf.WithContext(ctx).Where(tx.StressPlanTimedTaskConf.TeamID.Eq(req.TeamID),
				tx.StressPlanTimedTaskConf.PlanID.Eq(req.PlanID)).Find()
			if err != nil {
				return err
			}
			for _, timedTaskInfo := range timedTaskList {
				sceneID := timedTaskInfo.SceneID
				timedTaskInfo.ID = 0
				timedTaskInfo.PlanID = newPlanID
				timedTaskInfo.SceneID = targetMemo[sceneID]
				timedTaskInfo.TeamID = req.TeamID
				timedTaskInfo.Status = consts.TimedTaskWaitEnable
				timedTaskInfo.RunUserID = userID
				timedTaskInfo.CreatedAt = time.Now()
				timedTaskInfo.UpdatedAt = time.Now()
				if err := tx.StressPlanTimedTaskConf.WithContext(ctx).Create(timedTaskInfo); err != nil {
					return err
				}
			}
		}
		return record.InsertCreate(ctx, req.TeamID, userID, record.OperationOperateClonePlan, newPlanName)
	})
}

func BatchDeletePlan(ctx *gin.Context, req *rao.BatchDeletePlanReq) error {
	userID := jwt.GetUserIDByCtx(ctx)
	return dal.GetQuery().Transaction(func(tx *query.Query) error {
		planList, err := tx.StressPlan.WithContext(ctx).Where(tx.StressPlan.TeamID.Eq(req.TeamID), tx.StressPlan.PlanID.In(req.PlanIDs...)).Find()
		if err != nil {
			return err
		}

		for _, planInfo := range planList {
			if planInfo.Status == consts.PlanStatusUnderway {
				return fmt.Errorf("存在运行中的计划，无法删除")
			}
		}

		// 删除计划信息
		if _, err := tx.StressPlan.WithContext(ctx).Where(tx.StressPlan.TeamID.Eq(req.TeamID),
			tx.StressPlan.PlanID.In(req.PlanIDs...)).Delete(); err != nil {
			return err
		}

		// 获取所有计划下的场景
		sceneList, err := tx.Target.WithContext(ctx).Where(tx.Target.TeamID.Eq(req.TeamID), tx.Target.PlanID.In(req.PlanIDs...),
			tx.Target.Source.Eq(consts.TargetSourcePlan)).Find()
		if err != nil {
			return err
		}
		//删除所有场景内的接口详情
		if len(sceneList) > 0 {
			sceneIDs := make([]string, 0, len(sceneList))
			for _, sceneInfo := range sceneList {
				sceneIDs = append(sceneIDs, sceneInfo.TargetID)
			}

			collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectFlow)
			_, err = collection.DeleteMany(ctx, bson.D{{"team_id", req.TeamID}, {"scene_id", bson.D{{"$in", sceneIDs}}}})
			if err != nil {
				return err
			}

			// 删除场景下的变量
			_, err = tx.Variable.WithContext(ctx).Where(tx.Variable.SceneID.In(sceneIDs...)).Delete()
			if err != nil {
				return err
			}

			// 删除场景下的导入变量
			_, err = tx.VariableImport.WithContext(ctx).Where(tx.VariableImport.SceneID.In(sceneIDs...)).Delete()
			if err != nil {
				return err
			}
		}

		// 删除计划下场景
		if _, err = tx.Target.WithContext(ctx).Where(tx.Target.TeamID.Eq(req.TeamID), tx.Target.PlanID.In(req.PlanIDs...),
			tx.Target.Source.Eq(consts.TargetSourcePlan)).Delete(); err != nil {
			return err
		}

		// 删除计划下所有普通任务配置
		if _, err = tx.StressPlanTaskConf.WithContext(ctx).Where(tx.StressPlanTaskConf.TeamID.Eq(req.TeamID),
			tx.StressPlanTaskConf.PlanID.In(req.PlanIDs...)).Delete(); err != nil {
			return err
		}

		// 删除计划下所有定时任务配置
		if _, err = tx.StressPlanTimedTaskConf.WithContext(ctx).Where(tx.StressPlanTimedTaskConf.TeamID.Eq(req.TeamID),
			tx.StressPlanTimedTaskConf.PlanID.In(req.PlanIDs...)).Delete(); err != nil {
			return err
		}
		for _, planInfo := range planList {
			_ = record.InsertDelete(ctx, req.TeamID, userID, record.OperationOperateDeletePlan, planInfo.PlanName)
		}
		return nil
	})
}

func InsertReportData(ctx *gin.Context, req *rao.NotifyStopStressReq) error {
	resultData := rao.ResultData{}
	resultData.ReportRunTime = req.DurationTime

	// 查询报告基本信息
	tx := dal.GetQuery().StressPlanReport
	reportInfo, err := tx.WithContext(ctx).Where(tx.ReportID.Eq(req.ReportID)).First()
	if err != nil {
		log.Logger.Info("NotifyStopStress--查询报告结果失败，err:", err)
		return err
	}

	// 查询报告详情数据
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectReportData)
	filter := bson.D{{"team_id", reportInfo.TeamID}, {"report_id", req.ReportID}}
	dataMap := make(map[string]interface{}, 1)
	err = collection.FindOne(ctx, filter).Decode(&dataMap)
	_, ok := dataMap["data"]
	log.Logger.Info("NotifyStopStress--从MongoDB库查询报告详情结果为，err:", err, " ok:", ok)
	if err != nil || !ok {
		log.Logger.Info("mango数据为空，开始查询redis")
		rdb := dal.GetRDBForReport()
		key := fmt.Sprintf("reportData:%s", req.ReportID)
		dataList := rdb.LRange(ctx, key, 0, -1).Val()
		log.Logger.Info("查询redis报告数据，报告数据的Key:", key, "，数组长度为：", len(dataList), dataList)
		if len(dataList) < 1 {
			log.Logger.Info("redis里面没有查到报告详情数据")
			err = nil
			return err
		}

		// 初始化几个报告线的数据
		concurrencyListMap := make(map[string][]rao.TimeValue, 200)
		rpsListMap := make(map[string][]rao.TimeValue, 200)
		tpsListMap := make(map[string][]rao.TimeValue, 200)
		avgListMap := make(map[string][]rao.TimeValue, 200)
		fiftyListMap := make(map[string][]rao.TimeValue, 200)
		ninetyListMap := make(map[string][]rao.TimeValue, 200)
		ninetyFiveListMap := make(map[string][]rao.TimeValue, 200)
		ninetyNineListMap := make(map[string][]rao.TimeValue, 200)

		for i := len(dataList) - 1; i >= 0; i-- {
			resultMsgString := dataList[i]

			resultMsg := rao.SceneTestResultDataMsg{}
			err = json.Unmarshal([]byte(resultMsgString), &resultMsg)
			if err != nil {
				log.Logger.Info("json转换格式错误：", err)
				continue
			}

			resultData.ReportId = resultMsg.ReportId
			resultData.End = resultMsg.End
			resultData.ReportName = resultMsg.ReportName
			resultData.PlanId = resultMsg.PlanId
			resultData.PlanName = resultMsg.PlanName
			resultData.SceneId = resultMsg.SceneId
			resultData.SceneName = resultMsg.SceneName
			resultData.TimeStamp = resultMsg.TimeStamp
			resultsTemp := make(map[string]rao.ResultDataMsg, len(resultMsg.Results))
			if resultMsg.Results != nil && len(resultMsg.Results) > 0 {
				for k, apiResult := range resultMsg.Results {
					var errorRate float64 = 0
					totalRequestTime, _ := decimal.NewFromFloat(float64(apiResult.TotalRequestTime) / float64(time.Second)).Round(2).Float64()
					if apiResult.TotalRequestNum != 0 {
						errRate := float64(apiResult.ErrorNum) / float64(apiResult.TotalRequestNum)
						errorRate, _ = strconv.ParseFloat(fmt.Sprintf("%0.2f", errRate), 64)
					}
					avgRequestTime, _ := decimal.NewFromFloat(apiResult.AvgRequestTime / float64(time.Millisecond)).Round(1).Float64()
					maxRequestTime, _ := decimal.NewFromFloat(apiResult.MaxRequestTime / float64(time.Millisecond)).Round(1).Float64()
					minRequestTime, _ := decimal.NewFromFloat(apiResult.MinRequestTime / float64(time.Millisecond)).Round(1).Float64()
					customRequestTimeLineValue, _ := decimal.NewFromFloat(apiResult.CustomRequestTimeLineValue / float64(time.Millisecond)).Round(1).Float64()
					fiftyRequestTimelineValue, _ := decimal.NewFromFloat(apiResult.FiftyRequestTimelineValue / float64(time.Millisecond)).Round(1).Float64()
					ninetyRequestTimeLineValue, _ := decimal.NewFromFloat(apiResult.NinetyRequestTimeLineValue / float64(time.Millisecond)).Round(1).Float64()
					ninetyFiveRequestTimeLineValue, _ := decimal.NewFromFloat(apiResult.NinetyFiveRequestTimeLineValue / float64(time.Millisecond)).Round(1).Float64()
					ninetyNineRequestTimeLineValue, _ := decimal.NewFromFloat(apiResult.NinetyNineRequestTimeLineValue / float64(time.Millisecond)).Round(1).Float64()
					sendBytes, _ := decimal.NewFromFloat(apiResult.SendBytes).Round(1).Float64()
					receivedBytes, _ := decimal.NewFromFloat(apiResult.ReceivedBytes).Round(1).Float64()

					// rps列表
					rpsListMap[k] = append(rpsListMap[k], rao.TimeValue{
						TimeStamp: resultMsg.TimeStamp,
						Value:     apiResult.Rps,
					})

					// tps列表
					tpsListMap[k] = append(tpsListMap[k], rao.TimeValue{
						TimeStamp: resultMsg.TimeStamp,
						Value:     apiResult.Tps,
					})

					// 并发列表
					concurrencyListMap[k] = append(concurrencyListMap[k], rao.TimeValue{
						TimeStamp: resultMsg.TimeStamp,
						Value:     apiResult.Concurrency,
					})

					// 平均时长列表
					avgListMap[k] = append(avgListMap[k], rao.TimeValue{
						TimeStamp: resultMsg.TimeStamp,
						Value:     avgRequestTime,
					})

					// 50线
					fiftyListMap[k] = append(fiftyListMap[k], rao.TimeValue{
						TimeStamp: resultMsg.TimeStamp,
						Value:     fiftyRequestTimelineValue,
					})

					// 90线
					ninetyListMap[k] = append(ninetyListMap[k], rao.TimeValue{
						TimeStamp: resultMsg.TimeStamp,
						Value:     ninetyRequestTimeLineValue,
					})

					// 95线
					ninetyFiveListMap[k] = append(ninetyFiveListMap[k], rao.TimeValue{
						TimeStamp: resultMsg.TimeStamp,
						Value:     ninetyFiveRequestTimeLineValue,
					})

					// 99线
					ninetyNineListMap[k] = append(ninetyNineListMap[k], rao.TimeValue{
						TimeStamp: resultMsg.TimeStamp,
						Value:     ninetyNineRequestTimeLineValue,
					})

					resultsTemp[k] = rao.ResultDataMsg{
						ApiName:                        apiResult.Name,
						Concurrency:                    apiResult.Concurrency,
						TotalRequestNum:                apiResult.TotalRequestNum,
						TotalRequestTime:               totalRequestTime,
						SuccessNum:                     apiResult.SuccessNum,
						ErrorRate:                      errorRate,
						ErrorNum:                       apiResult.ErrorNum,
						AvgRequestTime:                 avgRequestTime,
						MaxRequestTime:                 maxRequestTime,
						MinRequestTime:                 minRequestTime,
						PercentAge:                     apiResult.PercentAge,
						ErrorThreshold:                 apiResult.ErrorThreshold,
						RequestThreshold:               apiResult.RequestThreshold,
						ResponseThreshold:              apiResult.ResponseThreshold,
						CustomRequestTimeLine:          apiResult.CustomRequestTimeLine,
						FiftyRequestTimeline:           apiResult.FiftyRequestTimeline,
						NinetyRequestTimeLine:          apiResult.NinetyRequestTimeLine,
						NinetyFiveRequestTimeLine:      apiResult.NinetyFiveRequestTimeLine,
						NinetyNineRequestTimeLine:      apiResult.NinetyNineRequestTimeLine,
						CustomRequestTimeLineValue:     customRequestTimeLineValue,
						FiftyRequestTimelineValue:      fiftyRequestTimelineValue,
						NinetyRequestTimeLineValue:     ninetyRequestTimeLineValue,
						NinetyFiveRequestTimeLineValue: ninetyFiveRequestTimeLineValue,
						NinetyNineRequestTimeLineValue: ninetyNineRequestTimeLineValue,
						SendBytes:                      sendBytes,
						ReceivedBytes:                  receivedBytes,
						Rps:                            apiResult.Rps,
						SRps:                           apiResult.SRps,
						Tps:                            apiResult.Tps,
						STps:                           apiResult.STps,
						ConcurrencyList:                concurrencyListMap[k],
						RpsList:                        rpsListMap[k],
						TpsList:                        tpsListMap[k],
						AvgList:                        avgListMap[k],
						FiftyList:                      fiftyListMap[k],
						NinetyList:                     ninetyListMap[k],
						NinetyFiveList:                 ninetyFiveListMap[k],
						NinetyNineList:                 ninetyNineListMap[k],
					}
				}
				resultData.Results = resultsTemp
			}
			if resultMsg.End {
				resultDataByte := make([]byte, 0, len(resultData.Results))
				resultDataByte, err = json.Marshal(resultData)
				if err != nil {
					log.Logger.Info("resultData转字节失败：：    ", err)
					return err
				}

				analysisByte := make([]byte, 0, len(resultData.Results))
				apiResultTotalMsg := make(map[string]string, len(resultData.Results))
				for _, value := range resultData.Results {
					apiResultTotalMsg[value.ApiName] = fmt.Sprintf("平均响应时间为%0.1fms； 百分之五十响应时间线的值为%0.1fms; 百分之九十响应时间线的值为%0.1fms; 百分之九十五响应时间线的值为%0.1fms; 百分之九十九响应时间线的值为%0.1fms; RPS为%0.1f; SRPS为%0.1f; TPS为%0.1f; STPS为%0.1f",
						value.AvgRequestTime, value.FiftyRequestTimelineValue, value.NinetyRequestTimeLineValue, value.NinetyFiveRequestTimeLineValue, value.NinetyNineRequestTimeLineValue, value.Rps, value.SRps, value.Tps, value.STps)
				}
				dataMap["report_id"] = resultData.ReportId
				dataMap["team_id"] = reportInfo.TeamID
				dataMap["plan_id"] = reportInfo.PlanID
				dataMap["data"] = string(resultDataByte)
				analysisByte, _ = json.Marshal(apiResultTotalMsg)
				dataMap["analysis"] = string(analysisByte)
				dataMap["report_run_time"] = req.DurationTime
				dataMap["description"] = ""
				_, err = collection.InsertOne(ctx, dataMap)
				if err != nil {
					log.Logger.Info("测试数据写入mongo失败:", err)
					return err
				}
				err = rdb.Del(ctx, key).Err()
				if err != nil {
					log.Logger.Info(fmt.Sprintf("删除redis的key：%s:", key), err)
					return err
				}
			}
		}
	}

	return nil
}

func GetPublicFunctionList(ctx *gin.Context) ([]*rao.GetPublicFunctionListResp, error) {
	tx := dal.GetQuery().PublicFunction
	list, err := tx.WithContext(ctx).Find()
	if err != nil {
		return nil, err
	}

	res := make([]*rao.GetPublicFunctionListResp, 0, len(list))
	for _, functionInfo := range list {
		tempData := &rao.GetPublicFunctionListResp{
			Function:     functionInfo.Function,
			FunctionName: functionInfo.FunctionName,
			Remark:       functionInfo.Remark,
		}
		res = append(res, tempData)
	}
	return res, nil
}

func GetNewestStressPlanList(ctx context.Context, req *rao.GetNewestStressPlanListReq) ([]rao.GetNewestStressPlanListResp, error) {
	res := make([]rao.GetNewestStressPlanListResp, 0, req.Size)
	_ = dal.GetQuery().Transaction(func(tx *query.Query) error {
		conditions := make([]gen.Condition, 0)
		conditions = append(conditions, tx.StressPlan.TeamID.Eq(req.TeamID))

		sort := make([]field.Expr, 0)
		sort = append(sort, tx.StressPlan.CreatedAt.Desc())

		planList, _, err := tx.StressPlan.WithContext(ctx).Select(tx.StressPlan.TeamID, tx.StressPlan.PlanID, tx.StressPlan.PlanName,
			tx.StressPlan.CreateUserID, tx.StressPlan.UpdatedAt).Where(conditions...).Order(sort...).FindByPage((req.Page-1)*req.Size, req.Size)
		if err != nil {
			return err
		}

		//统计用户id
		var userIDs []string
		for _, planInfo := range planList {
			userIDs = append(userIDs, planInfo.CreateUserID)
		}

		users, err := tx.User.WithContext(ctx).Select(tx.User.UserID, tx.User.Nickname, tx.User.Avatar).Where(tx.User.UserID.In(userIDs...)).Find()
		if err != nil {
			return err
		}

		res = packer.TransNewestPlansToRaoPlanList(planList, users)
		return nil
	})
	return res, nil
}
