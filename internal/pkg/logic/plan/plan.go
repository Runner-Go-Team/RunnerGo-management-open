package plan

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-omnibus/proof"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"kp-management/internal/pkg/biz/errno"
	"kp-management/internal/pkg/biz/jwt"
	"kp-management/internal/pkg/biz/log"
	"kp-management/internal/pkg/biz/uuid"
	"kp-management/internal/pkg/dal/mao"
	"kp-management/internal/pkg/logic/report"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"gorm.io/gen"
	"gorm.io/gen/field"

	"kp-management/internal/pkg/biz/consts"
	"kp-management/internal/pkg/biz/record"
	"kp-management/internal/pkg/dal"
	"kp-management/internal/pkg/dal/model"
	"kp-management/internal/pkg/dal/query"
	"kp-management/internal/pkg/dal/rao"
	"kp-management/internal/pkg/packer"
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

func CountByTeamID(ctx context.Context, teamID string) (int64, error) {
	tx := query.Use(dal.DB()).StressPlan

	return tx.WithContext(ctx).Where(tx.TeamID.Eq(teamID)).Count()
}

func ListByTeamID(ctx context.Context, teamID string, limit, offset int, keyword string, startTimeSec, endTimeSec int64, taskType, taskMode, status, sortTag int32) ([]*rao.StressPlan, int64, error) {
	tx := query.Use(dal.DB()).StressPlan
	conditions := make([]gen.Condition, 0)
	conditions = append(conditions, tx.TeamID.Eq(teamID))

	if keyword != "" {
		conditions = append(conditions, tx.PlanName.Like(fmt.Sprintf("%%%s%%", keyword)))

		u := query.Use(dal.DB()).User
		users, err := u.WithContext(ctx).Where(u.Nickname.Like(fmt.Sprintf("%%%s%%", keyword))).Find()
		if err != nil {
			return nil, 0, err
		}

		if len(users) > 0 {
			conditions[1] = tx.RunUserID.Eq(users[0].UserID)
		}
	}

	if startTimeSec > 0 && endTimeSec > 0 {
		startTime := time.Unix(startTimeSec, 0)
		endTime := time.Unix(endTimeSec, 0)
		conditions = append(conditions, tx.CreatedAt.Between(startTime, endTime))
	}

	if taskType > 0 {
		conditions = append(conditions, tx.TaskType.Eq(taskType))
	}

	if taskMode > 0 {
		conditions = append(conditions, tx.TaskMode.Eq(taskMode))
	}

	if status > 0 {
		conditions = append(conditions, tx.Status.Eq(status))
	}

	sort := make([]field.Expr, 0)
	if sortTag == 0 { // 默认排序
		sort = append(sort, tx.CreatedAt.Desc())
	}
	if sortTag == 1 { // 创建时间倒序
		sort = append(sort, tx.CreatedAt.Desc())
	}
	if sortTag == 2 { // 创建时间正序
		sort = append(sort, tx.CreatedAt)
	}
	if sortTag == 3 { // 修改时间倒序
		sort = append(sort, tx.UpdatedAt.Desc())
	}
	if sortTag == 4 { // 修改时间正序
		sort = append(sort, tx.UpdatedAt)
	}

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

func SaveTask(ctx *gin.Context, req *rao.SavePlanConfReq, userID string) (int, error) {
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

			if err == nil { // 已存在 则修改
				updateData := model.StressPlanTaskConf{
					TaskType:    req.TaskType,
					TaskMode:    req.Mode,
					ControlMode: req.ControlMode,
					ModeConf:    string(modeConfString),
					RunUserID:   userID,
				}
				_, err = tx.StressPlanTaskConf.WithContext(ctx).
					Where(tx.StressPlanTaskConf.TeamID.Eq(req.TeamID),
						tx.StressPlanTaskConf.PlanID.Eq(req.PlanID),
						tx.StressPlanTaskConf.SceneID.Eq(req.SceneID)).Updates(updateData)
				if err != nil {
					return err
				}
			} else { // 不存在，则新增
				insertData := &model.StressPlanTaskConf{
					PlanID:      req.PlanID,
					TeamID:      req.TeamID,
					SceneID:     req.SceneID,
					TaskType:    req.TaskType,
					TaskMode:    req.Mode,
					ControlMode: req.ControlMode,
					ModeConf:    string(modeConfString),
					RunUserID:   userID,
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

				// 修改配置
				updateData := make(map[string]interface{}, 3)
				updateData["user_id"] = userID
				updateData["frequency"] = req.TimedTaskConf.Frequency
				updateData["task_exec_time"] = req.TimedTaskConf.TaskExecTime
				updateData["task_close_time"] = req.TimedTaskConf.TaskCloseTime
				updateData["task_mode"] = req.Mode
				updateData["control_mode"] = req.ControlMode
				updateData["mode_conf"] = modeConfString
				updateData["status"] = consts.TimedTaskWaitEnable
				_, err = tx.StressPlanTimedTaskConf.WithContext(ctx).Where(tx.StressPlanTimedTaskConf.TeamID.Eq(req.TeamID)).
					Where(tx.StressPlanTimedTaskConf.PlanID.Eq(req.PlanID)).
					Where(tx.StressPlanTimedTaskConf.SceneID.Eq(req.SceneID)).Updates(updateData)
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
		PlanID:        req.PlanID,
		SceneID:       req.SceneID,
		TaskType:      req.TaskType,
		Mode:          consts.PlanModeConcurrence,
		ModeConf:      rao.ModeConf{},
		TimedTaskConf: rao.TimedTaskConf{},
	}

	tx := dal.GetQuery()
	if req.TaskType == consts.PlanTaskTypeNormal { // 普通任务
		taskConfInfo, err := tx.StressPlanTaskConf.WithContext(ctx).
			Where(tx.StressPlanTaskConf.TeamID.Eq(req.TeamID), tx.StressPlanTaskConf.PlanID.Eq(req.PlanID),
				tx.StressPlanTaskConf.SceneID.Eq(req.SceneID)).First()
		if err == nil { // 查到了，普通任务
			// 解析配置详情
			var taskConfDetail rao.ModeConf
			err := json.Unmarshal([]byte(taskConfInfo.ModeConf), &taskConfDetail)
			if err != nil {
				log.Logger.Info("获取配置详情--解析数据失败")
				return nil, err
			}

			planTaskConf = &rao.PlanTaskResp{
				PlanID:      req.PlanID,
				SceneID:     req.SceneID,
				TaskType:    req.TaskType,
				Mode:        taskConfInfo.TaskMode,
				ControlMode: taskConfInfo.ControlMode,
				ModeConf: rao.ModeConf{
					ReheatTime:       taskConfDetail.ReheatTime,
					RoundNum:         taskConfDetail.RoundNum,
					Concurrency:      taskConfDetail.Concurrency,
					ThresholdValue:   taskConfDetail.ThresholdValue,
					StartConcurrency: taskConfDetail.StartConcurrency,
					Step:             taskConfDetail.Step,
					StepRunTime:      taskConfDetail.StepRunTime,
					MaxConcurrency:   taskConfDetail.MaxConcurrency,
					Duration:         taskConfDetail.Duration,
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
			planTaskConf = &rao.PlanTaskResp{
				PlanID:      req.PlanID,
				SceneID:     req.SceneID,
				TaskType:    req.TaskType,
				Mode:        timingTaskConfigInfo.TaskMode,
				ControlMode: timingTaskConfigInfo.ControlMode,
				ModeConf:    modeConf,
				TimedTaskConf: rao.TimedTaskConf{
					Frequency:     timingTaskConfigInfo.Frequency,
					TaskExecTime:  timingTaskConfigInfo.TaskExecTime,
					TaskCloseTime: timingTaskConfigInfo.TaskCloseTime,
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
			v, err := tx.Variable.WithContext(ctx).Where(tx.Variable.SceneID.In(sceneIDs...)).Find()
			if err != nil {
				return err
			}

			for _, variable := range v {
				variable.ID = 0
				variable.SceneID = targetMemo[variable.SceneID]
				variable.CreatedAt = time.Now()
				variable.UpdatedAt = time.Now()
				if err := tx.Variable.WithContext(ctx).Create(variable); err != nil {
					return err
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
		//return record.InsertCreate(ctx, req.TeamID, userID, record.OperationOperateClonePlan, newPlanName)
		return nil
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
	var resultData report.ResultData

	// 查询报告详情数据
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectReportData)
	filter := bson.D{{"team_id", req.TeamID}, {"report_id", req.ReportID}}
	var resultMsg report.SceneTestResultDataMsg
	var dataMap = make(map[string]interface{})
	err := collection.FindOne(ctx, filter).Decode(&dataMap)
	_, ok := dataMap["data"]
	log.Logger.Info("NotifyStopStress--从MongoDB库查询报告详情结果为，err:", err, " ok:", ok)
	if err != nil || !ok {
		log.Logger.Info("NotifyStopStress--把redis数据写到mg库")
		rdb := dal.GetRDBForReport()
		key := fmt.Sprintf("reportData:%s", req.ReportID)
		dataList := rdb.LRange(ctx, key, 0, -1).Val()
		if len(dataList) < 1 {
			log.Logger.Info("NotifyStopStress--redis里面没有查到报告详情数据，err:", proof.WithError(err))
			return nil
		}
		log.Logger.Info("NotifyStopStress--redis报告队里里面的数据个数为：", len(dataList))
		for i := len(dataList) - 1; i >= 0; i-- {
			resultMsgString := dataList[i]
			err = json.Unmarshal([]byte(resultMsgString), &resultMsg)
			if err != nil {
				log.Logger.Info("NotifyStopStress--json转换格式错误，err:", proof.WithError(err))
			}
			if resultData.Results == nil {
				resultData.Results = make(map[string]*report.ResultDataMsg)
			}
			log.Logger.Info("NotifyStopStress--循环报告数据入库，报告id为：", resultMsg.ReportId)
			resultData.ReportId = resultMsg.ReportId
			resultData.End = resultMsg.End
			resultData.ReportName = resultMsg.ReportName
			resultData.PlanId = resultMsg.PlanId
			resultData.PlanName = resultMsg.PlanName
			resultData.SceneId = resultMsg.SceneId
			resultData.SceneName = resultMsg.SceneName
			resultData.TimeStamp = resultMsg.TimeStamp
			if resultMsg.Results != nil && len(resultMsg.Results) > 0 {
				log.Logger.Info("NotifyStopStress--resultMsg.Results有值，end值为：", resultMsg.End)
				for k, apiResult := range resultMsg.Results {
					//log.Logger.Info("NotifyStopStress--组装添加数据开始")
					if resultData.Results[k] == nil {
						resultData.Results[k] = new(report.ResultDataMsg)
					}
					resultData.Results[k].ApiName = apiResult.Name
					resultData.Results[k].Concurrency = apiResult.Concurrency
					resultData.Results[k].TotalRequestNum = apiResult.TotalRequestNum
					resultData.Results[k].TotalRequestTime, _ = decimal.NewFromFloat(float64(apiResult.TotalRequestTime) / float64(time.Second)).Round(2).Float64()
					resultData.Results[k].SuccessNum = apiResult.SuccessNum
					resultData.Results[k].ErrorNum = apiResult.ErrorNum
					if apiResult.TotalRequestNum != 0 {
						errRate := float64(apiResult.ErrorNum) / float64(apiResult.TotalRequestNum)
						resultData.Results[k].ErrorRate, _ = strconv.ParseFloat(fmt.Sprintf("%0.2f", errRate), 64)
					}
					resultData.Results[k].PercentAge = apiResult.PercentAge
					resultData.Results[k].ErrorThreshold = apiResult.ErrorThreshold
					resultData.Results[k].ResponseThreshold = apiResult.ResponseThreshold
					resultData.Results[k].RequestThreshold = apiResult.RequestThreshold
					resultData.Results[k].AvgRequestTime, _ = decimal.NewFromFloat(apiResult.AvgRequestTime / float64(time.Millisecond)).Round(1).Float64()
					resultData.Results[k].MaxRequestTime, _ = decimal.NewFromFloat(apiResult.MaxRequestTime / float64(time.Millisecond)).Round(1).Float64()
					resultData.Results[k].MinRequestTime, _ = decimal.NewFromFloat(apiResult.MinRequestTime / float64(time.Millisecond)).Round(1).Float64()
					resultData.Results[k].CustomRequestTimeLine = apiResult.CustomRequestTimeLine
					resultData.Results[k].CustomRequestTimeLineValue, _ = decimal.NewFromFloat(apiResult.CustomRequestTimeLineValue / float64(time.Millisecond)).Round(1).Float64()
					resultData.Results[k].FiftyRequestTimelineValue, _ = decimal.NewFromFloat(apiResult.FiftyRequestTimelineValue / float64(time.Millisecond)).Round(1).Float64()
					resultData.Results[k].NinetyRequestTimeLine = apiResult.NinetyRequestTimeLine
					resultData.Results[k].NinetyRequestTimeLineValue, _ = decimal.NewFromFloat(apiResult.NinetyRequestTimeLineValue / float64(time.Millisecond)).Round(1).Float64()
					resultData.Results[k].NinetyFiveRequestTimeLine = apiResult.NinetyFiveRequestTimeLine
					resultData.Results[k].NinetyFiveRequestTimeLineValue, _ = decimal.NewFromFloat(apiResult.NinetyFiveRequestTimeLineValue / float64(time.Millisecond)).Round(1).Float64()
					resultData.Results[k].NinetyNineRequestTimeLine = apiResult.NinetyNineRequestTimeLine
					resultData.Results[k].NinetyNineRequestTimeLineValue, _ = decimal.NewFromFloat(apiResult.NinetyNineRequestTimeLineValue / float64(time.Millisecond)).Round(1).Float64()
					resultData.Results[k].SendBytes, _ = decimal.NewFromFloat(apiResult.SendBytes).Round(1).Float64()
					resultData.Results[k].ReceivedBytes, _ = decimal.NewFromFloat(apiResult.ReceivedBytes).Round(1).Float64()
					resultData.Results[k].Rps = apiResult.Rps
					resultData.Results[k].SRps = apiResult.SRps
					resultData.Results[k].Tps = apiResult.Tps
					resultData.Results[k].STps = apiResult.STps
					if resultData.Results[k].RpsList == nil {
						resultData.Results[k].RpsList = []report.TimeValue{}
					}
					var timeValue = report.TimeValue{}
					timeValue.TimeStamp = resultData.TimeStamp
					// qps列表
					timeValue.Value = resultData.Results[k].Rps
					resultData.Results[k].RpsList = append(resultData.Results[k].RpsList, timeValue)
					timeValue.Value = resultData.Results[k].Tps
					if resultData.Results[k].TpsList == nil {
						resultData.Results[k].TpsList = []report.TimeValue{}
					}
					// 错误数列表
					resultData.Results[k].TpsList = append(resultData.Results[k].TpsList, timeValue)
					timeValue.Value = resultData.Results[k].Concurrency
					if resultData.Results[k].ConcurrencyList == nil {
						resultData.Results[k].ConcurrencyList = []report.TimeValue{}
					}
					// 并发数列表
					resultData.Results[k].ConcurrencyList = append(resultData.Results[k].ConcurrencyList, timeValue)

					// 平均响应时间列表
					timeValue.Value = resultData.Results[k].AvgRequestTime
					if resultData.Results[k].AvgList == nil {
						resultData.Results[k].AvgList = []report.TimeValue{}
					}
					resultData.Results[k].AvgList = append(resultData.Results[k].AvgList, timeValue)

					// 50响应时间列表
					timeValue.Value = resultData.Results[k].FiftyRequestTimelineValue
					if resultData.Results[k].FiftyList == nil {
						resultData.Results[k].FiftyList = []report.TimeValue{}
					}
					resultData.Results[k].FiftyList = append(resultData.Results[k].FiftyList, timeValue)

					// 90响应时间列表
					timeValue.Value = resultData.Results[k].NinetyNineRequestTimeLineValue
					if resultData.Results[k].NinetyList == nil {
						resultData.Results[k].NinetyList = []report.TimeValue{}
					}
					resultData.Results[k].NinetyList = append(resultData.Results[k].NinetyList, timeValue)

					// 95响应时间列表
					timeValue.Value = resultData.Results[k].NinetyFiveRequestTimeLineValue
					if resultData.Results[k].NinetyFiveList == nil {
						resultData.Results[k].NinetyFiveList = []report.TimeValue{}
					}
					resultData.Results[k].NinetyFiveList = append(resultData.Results[k].NinetyFiveList, timeValue)

					// 99响应时间列表
					timeValue.Value = resultData.Results[k].NinetyNineRequestTimeLineValue
					if resultData.Results[k].NinetyNineList == nil {
						resultData.Results[k].NinetyNineList = []report.TimeValue{}
					}
					resultData.Results[k].NinetyNineList = append(resultData.Results[k].NinetyNineList, timeValue)
				}
				log.Logger.Info("NotifyStopStress--组装添加数据完成")
			}
			if resultMsg.End {
				log.Logger.Info("NotifyStopStress--报告已完成，准备入库")
				var by []byte
				by, err = json.Marshal(resultData)
				if err != nil {
					log.Logger.Info("NotifyStopStress--resultData转字节失败，err:", proof.WithError(err))
					return err
				}
				var apiResultTotalMsg = make(map[string]string)
				for _, value := range resultData.Results {
					apiResultTotalMsg[value.ApiName] = fmt.Sprintf("平均响应时间为%0.1fms； 百分之五十响应时间线的值为%0.1fms; 百分之九十响应时间线的值为%0.1fms; 百分之九十五响应时间线的值为%0.1fms; 百分之九十九响应时间线的值为%0.1fms; RPS为%0.1f; SRPS为%0.1f; TPS为%0.1f; STPS为%0.1f",
						value.AvgRequestTime, value.FiftyRequestTimelineValue, value.NinetyRequestTimeLineValue, value.NinetyFiveRequestTimeLineValue, value.NinetyNineRequestTimeLineValue, value.Rps, value.SRps, value.Tps, value.STps)
				}
				dataMap["report_id"] = resultData.ReportId
				dataMap["team_id"] = req.TeamID
				dataMap["plan_id"] = req.PlanID
				dataMap["data"] = string(by)
				by, _ = json.Marshal(apiResultTotalMsg)
				dataMap["analysis"] = string(by)
				dataMap["description"] = ""
				_, err = collection.InsertOne(ctx, dataMap)
				log.Logger.Info("NotifyStopStress--报告数据插入mg库结果，err:", proof.WithError(err))
				if err != nil {
					log.Logger.Info("NotifyStopStress--测试数据写入mongo失败，err:", proof.WithError(err))
					return err
				}
				err = rdb.Del(ctx, key).Err()
				if err != nil {
					log.Logger.Info("NotifyStopStress--删除redis的key：", key, " err:", proof.WithError(err))
					return err
				}
			}
		}
	}

	return nil
}
