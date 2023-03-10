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
	if sortTag == 0 { // ????????????
		sort = append(sort, tx.CreatedAt.Desc())
	}
	if sortTag == 1 { // ??????????????????
		sort = append(sort, tx.CreatedAt.Desc())
	}
	if sortTag == 2 { // ??????????????????
		sort = append(sort, tx.CreatedAt)
	}
	if sortTag == 3 { // ??????????????????
		sort = append(sort, tx.UpdatedAt.Desc())
	}
	if sortTag == 4 { // ??????????????????
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
		return req.PlanID, errno.ErrPlanNameNotEmpty, fmt.Errorf("????????????????????????")
	}

	// ????????????
	userID := jwt.GetUserIDByCtx(ctx)
	planID := req.PlanID
	var rankID int64 = 1
	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		// ????????????????????????????????????
		_, err := tx.StressPlan.WithContext(ctx).Where(tx.StressPlan.TeamID.Eq(req.TeamID),
			tx.StressPlan.PlanName.Eq(req.PlanName), tx.StressPlan.PlanID.Neq(req.PlanID)).First()
		if err != nil && err != gorm.ErrRecordNotFound {
			log.Logger.Info("???????????????????????????err:", err)
			return err
		}

		if err == nil { // ?????????
			return fmt.Errorf("???????????????")
		}

		// ??????????????????plan_id
		if req.PlanID == "" { // ????????????
			// ??????????????????????????????plan_id???
			StressPlanInfo, err := tx.StressPlan.WithContext(ctx).Where(tx.StressPlan.TeamID.Eq(req.TeamID)).Order(tx.StressPlan.RankID.Desc()).Limit(1).First()
			if err != nil && err != gorm.ErrRecordNotFound {
				return err
			}
			if err == nil {
				rankID = StressPlanInfo.RankID + 1
			}

			planID = uuid.GetUUID()

			// ???????????????????????????
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
		} else { // ????????????
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
		if err.Error() == "???????????????" {
			return "", errno.ErrPlanNameAlreadyExist, err
		}
		return "", errno.ErrMysqlFailed, err
	}

	return planID, errno.Ok, nil
}

func SaveTask(ctx *gin.Context, req *rao.SavePlanConfReq, userID string) (int, error) {
	// ????????????????????????
	var err error

	if req.TaskType == consts.PlanTaskTypeNormal { // ????????????
		err = query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
			// 1????????????????????????????????????
			_, err := tx.StressPlanTimedTaskConf.WithContext(ctx).
				Where(tx.StressPlanTimedTaskConf.TeamID.Eq(req.TeamID)).
				Where(tx.StressPlanTimedTaskConf.PlanID.Eq(req.PlanID)).
				Where(tx.StressPlanTimedTaskConf.SceneID.Eq(req.SceneID)).Delete()
			if err != nil {
				log.Logger.Info("????????????--??????????????????????????????mysql??????,err:", err)
			}

			// ????????????????????????
			_, err = tx.StressPlanTaskConf.WithContext(ctx).
				Where(tx.StressPlanTaskConf.TeamID.Eq(req.TeamID),
					tx.StressPlanTaskConf.PlanID.Eq(req.PlanID),
					tx.StressPlanTaskConf.SceneID.Eq(req.SceneID)).First()
			if err != nil && err != gorm.ErrRecordNotFound {
				return err
			}

			// ????????????????????????
			modeConfString, err2 := json.Marshal(*req.ModeConf)
			if err2 != nil {
				log.Logger.Info("??????????????????--??????????????????????????????")
				return err2
			}

			if err == nil { // ????????? ?????????
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
			} else { // ?????????????????????
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
	} else { // ????????????
		err = query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
			// 1????????????????????????????????????
			_, err := tx.StressPlanTaskConf.WithContext(ctx).
				Where(tx.StressPlanTaskConf.TeamID.Eq(req.TeamID),
					tx.StressPlanTaskConf.PlanID.Eq(req.PlanID),
					tx.StressPlanTaskConf.SceneID.Eq(req.SceneID)).Delete()
			if err != nil {
				return err
			}

			// ????????????????????????????????????
			// ????????????????????????????????????
			_, err = tx.StressPlanTimedTaskConf.WithContext(ctx).
				Where(tx.StressPlanTimedTaskConf.TeamID.Eq(req.TeamID)).
				Where(tx.StressPlanTimedTaskConf.PlanID.Eq(req.PlanID)).
				Where(tx.StressPlanTimedTaskConf.SceneID.Eq(req.SceneID)).First()
			if err != nil && err != gorm.ErrRecordNotFound { // ????????????
				log.Logger.Info("????????????--?????????????????????????????????err:", req)
				return err
			} else if err == gorm.ErrRecordNotFound { // ???????????????
				// ????????????
				timingTaskConfig, err := packer.TransSaveTimingTaskConfigReqToModelData(req, userID)
				if err != nil {
					log.Logger.Info("????????????--??????mode_conf?????????????????????", err)
					return err
				}
				err = tx.StressPlanTimedTaskConf.WithContext(ctx).Create(timingTaskConfig)
				if err != nil {
					log.Logger.Info("????????????--????????????????????????????????????err???", err)
					return err
				}
			} else {
				// ???mode_conf??????????????????
				modeConfString, err := json.Marshal(req.ModeConf)
				if err != nil {
					log.Logger.Info("????????????--??????mode_conf?????????????????????", err)
					return err
				}

				// ????????????
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
					log.Logger.Info("????????????--?????????????????????????????????err:", err)
					return err
				}
			}
			// ???????????????
			return nil
		})
	}
	if err != nil {
		log.Logger.Info("????????????--?????????????????????????????????err:", err)
		return errno.ErrMysqlFailed, err
	}

	// ????????????????????????
	tx := dal.GetQuery()
	var planMode int32 = 0
	if req.TaskType == consts.PlanTaskTypeNormal { // ????????????
		tasks, err := tx.StressPlanTaskConf.WithContext(ctx).
			Where(tx.StressPlanTaskConf.TeamID.Eq(req.TeamID),
				tx.StressPlanTaskConf.PlanID.Eq(req.PlanID)).Find()
		if err != nil {
			log.Logger.Info("????????????--??????????????????????????????????????????????????????err:", err)
			return errno.ErrMysqlFailed, err
		}
		if len(tasks) > 0 {
			// ??????
			planMode = tasks[0].TaskMode
			for i, t := range tasks {
				if i > 0 && t.TaskMode != planMode && planMode != 0 {
					planMode = consts.PlanModeMix
					break
				}
			}
		}
	} else {
		// ??????????????????????????????????????????
		timedTaskInfo, err := tx.StressPlanTimedTaskConf.WithContext(ctx).Where(tx.StressPlanTimedTaskConf.TeamID.Eq(req.TeamID),
			tx.StressPlanTimedTaskConf.PlanID.Eq(req.PlanID)).First()
		if err != nil {
			log.Logger.Info("????????????--??????????????????????????????????????????????????????err:", err)
			return errno.ErrMysqlFailed, err
		}

		planMode = timedTaskInfo.TaskMode
	}

	_, err = tx.StressPlan.WithContext(ctx).Where(tx.StressPlan.TeamID.Eq(req.TeamID), tx.StressPlan.PlanID.Eq(req.PlanID)).UpdateSimple(tx.StressPlan.TaskMode.Value(planMode))
	if err != nil {
		log.Logger.Info("????????????--???????????????????????????????????????????????????err:", err)
		return errno.ErrMysqlFailed, err
	}
	// ???????????????
	return errno.Ok, nil
}

func GetPlanTask(ctx context.Context, req *rao.GetPlanTaskReq) (*rao.PlanTaskResp, error) {
	// ??????????????????
	planTaskConf := &rao.PlanTaskResp{
		PlanID:        req.PlanID,
		SceneID:       req.SceneID,
		TaskType:      req.TaskType,
		Mode:          consts.PlanModeConcurrence,
		ModeConf:      rao.ModeConf{},
		TimedTaskConf: rao.TimedTaskConf{},
	}

	tx := dal.GetQuery()
	if req.TaskType == consts.PlanTaskTypeNormal { // ????????????
		taskConfInfo, err := tx.StressPlanTaskConf.WithContext(ctx).
			Where(tx.StressPlanTaskConf.TeamID.Eq(req.TeamID), tx.StressPlanTaskConf.PlanID.Eq(req.PlanID),
				tx.StressPlanTaskConf.SceneID.Eq(req.SceneID)).First()
		if err == nil { // ????????????????????????
			// ??????????????????
			var taskConfDetail rao.ModeConf
			err := json.Unmarshal([]byte(taskConfInfo.ModeConf), &taskConfDetail)
			if err != nil {
				log.Logger.Info("??????????????????--??????????????????")
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
	} else { // ????????????
		// ????????????????????????
		timingTaskConfigInfo, err := tx.StressPlanTimedTaskConf.WithContext(ctx).Where(tx.StressPlanTimedTaskConf.TeamID.Eq(req.TeamID),
			tx.StressPlanTimedTaskConf.PlanID.Eq(req.PlanID),
			tx.StressPlanTimedTaskConf.SceneID.Eq(req.SceneID)).First()
		if err == nil {
			var modeConf rao.ModeConf
			err := json.Unmarshal([]byte(timingTaskConfigInfo.ModeConf), &modeConf)
			if err != nil {
				log.Logger.Info("????????????????????????--???????????????????????????????????????err:", err)
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

			if timingTaskConfigInfo.Frequency == 0 { // ????????????
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

	// ??????????????????
	u := query.Use(dal.DB()).User
	user, err := u.WithContext(ctx).Where(u.UserID.Eq(planInfo.CreateUserID)).First()
	if err != nil {
		return nil, err
	}

	// ??????????????????
	taskConfTable := dal.GetQuery().StressPlanTaskConf
	taskConfInfo, err := taskConfTable.WithContext(ctx).Where(taskConfTable.TeamID.Eq(teamID), taskConfTable.PlanID.Eq(planID)).Order(taskConfTable.SceneID).First()
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	var taskConf rao.ModeConf
	if err == nil {
		err := json.Unmarshal([]byte(taskConfInfo.ModeConf), &taskConf)
		if err != nil {
			log.Logger.Info("????????????--?????????????????????????????????????????????", taskConfInfo.ModeConf)
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
			return fmt.Errorf("?????????????????????????????????")
		}

		// ??????????????????
		if _, err := tx.StressPlan.WithContext(ctx).Where(tx.StressPlan.TeamID.Eq(teamID),
			tx.StressPlan.PlanID.Eq(planID)).Delete(); err != nil {
			return err
		}

		// ??????????????????????????????
		sceneList, err := tx.Target.WithContext(ctx).Where(tx.Target.TeamID.Eq(teamID), tx.Target.PlanID.Eq(planID),
			tx.Target.Source.Eq(consts.TargetSourcePlan)).Find()
		if err != nil {
			return err
		}
		//????????????????????????????????????
		if len(sceneList) > 0 {
			sceneIDs := make([]string, 0, len(sceneList))
			for _, sceneInfo := range sceneList {
				sceneIDs = append(sceneIDs, sceneInfo.TargetID)
			}

			// ??????????????????flow
			collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectFlow)
			_, err = collection.DeleteMany(ctx, bson.D{{"team_id", teamID}, {"scene_id", bson.D{{"$in", sceneIDs}}}})
			if err != nil {
				return err
			}

			// ????????????????????????
			_, err = tx.Variable.WithContext(ctx).Where(tx.Variable.SceneID.In(sceneIDs...)).Delete()
			if err != nil {
				return err
			}

			// ??????????????????????????????
			_, err = tx.VariableImport.WithContext(ctx).Where(tx.VariableImport.SceneID.In(sceneIDs...)).Delete()
			if err != nil {
				return err
			}
		}

		// ?????????????????????
		if _, err = tx.Target.WithContext(ctx).Where(tx.Target.TeamID.Eq(teamID), tx.Target.PlanID.Eq(planID),
			tx.Target.Source.Eq(consts.TargetSourcePlan)).Delete(); err != nil {
			return err
		}

		if planInfo.TaskType == consts.PlanTaskTypeNormal {
			// ???????????????????????????????????????
			if _, err = tx.StressPlanTaskConf.WithContext(ctx).Where(tx.StressPlanTaskConf.TeamID.Eq(teamID),
				tx.StressPlanTaskConf.PlanID.Eq(planID)).Delete(); err != nil {
				return err
			}
		} else {
			// ???????????????????????????????????????
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
		//????????????
		oldPlanInfo, err := tx.StressPlan.WithContext(ctx).Where(tx.StressPlan.TeamID.Eq(req.TeamID), tx.StressPlan.PlanID.Eq(req.PlanID)).First()
		if err != nil {
			return err
		}

		oldPlanName := oldPlanInfo.PlanName
		newPlanName := oldPlanName + "_1"

		// ????????????????????????
		list, err := tx.StressPlan.WithContext(ctx).Where(tx.StressPlan.TeamID.Eq(req.TeamID)).Where(tx.StressPlan.PlanName.Like(fmt.Sprintf("%s%%", oldPlanName+"_"))).Find()
		if err == nil {
			// ?????????????????????
			maxNum := 0
			for _, autoPlanInfo := range list {
				nameTmp := autoPlanInfo.PlanName
				postfixSlice := strings.Split(nameTmp, "_")
				if len(postfixSlice) < 2 {
					continue
				}
				currentNum, err := strconv.Atoi(postfixSlice[len(postfixSlice)-1])
				if err != nil {
					log.Logger.Info("??????????????????--?????????????????????err:", err)
					continue
				}
				if currentNum > maxNum {
					maxNum = currentNum
				}
			}
			newPlanName = oldPlanName + fmt.Sprintf("_%d", maxNum+1)
		}

		// ????????????????????????????????????
		newPlanID := uuid.GetUUID()
		var rankID int64 = 1
		stressPlanInfo, err := tx.StressPlan.WithContext(ctx).Where(tx.StressPlan.TeamID.Eq(req.TeamID)).Order(tx.StressPlan.RankID.Desc()).Limit(1).First()
		if err == nil { // ?????????
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
		// ?????????????????????
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
			// ??????????????????
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

			// ??????????????????
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

			// ????????????
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

		// ??????????????????
		if oldPlanInfo.TaskType == consts.PlanTaskTypeNormal {
			// ??????????????????????????????
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
			// ??????????????????
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
				return fmt.Errorf("???????????????????????????????????????")
			}
		}

		// ??????????????????
		if _, err := tx.StressPlan.WithContext(ctx).Where(tx.StressPlan.TeamID.Eq(req.TeamID),
			tx.StressPlan.PlanID.In(req.PlanIDs...)).Delete(); err != nil {
			return err
		}

		// ??????????????????????????????
		sceneList, err := tx.Target.WithContext(ctx).Where(tx.Target.TeamID.Eq(req.TeamID), tx.Target.PlanID.In(req.PlanIDs...),
			tx.Target.Source.Eq(consts.TargetSourcePlan)).Find()
		if err != nil {
			return err
		}
		//????????????????????????????????????
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

			// ????????????????????????
			_, err = tx.Variable.WithContext(ctx).Where(tx.Variable.SceneID.In(sceneIDs...)).Delete()
			if err != nil {
				return err
			}

			// ??????????????????????????????
			_, err = tx.VariableImport.WithContext(ctx).Where(tx.VariableImport.SceneID.In(sceneIDs...)).Delete()
			if err != nil {
				return err
			}
		}

		// ?????????????????????
		if _, err = tx.Target.WithContext(ctx).Where(tx.Target.TeamID.Eq(req.TeamID), tx.Target.PlanID.In(req.PlanIDs...),
			tx.Target.Source.Eq(consts.TargetSourcePlan)).Delete(); err != nil {
			return err
		}

		// ???????????????????????????????????????
		if _, err = tx.StressPlanTaskConf.WithContext(ctx).Where(tx.StressPlanTaskConf.TeamID.Eq(req.TeamID),
			tx.StressPlanTaskConf.PlanID.In(req.PlanIDs...)).Delete(); err != nil {
			return err
		}

		// ???????????????????????????????????????
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

	// ????????????????????????
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectReportData)
	filter := bson.D{{"team_id", req.TeamID}, {"report_id", req.ReportID}}
	var resultMsg report.SceneTestResultDataMsg
	var dataMap = make(map[string]interface{})
	err := collection.FindOne(ctx, filter).Decode(&dataMap)
	_, ok := dataMap["data"]
	log.Logger.Info("NotifyStopStress--???MongoDB?????????????????????????????????err:", err, " ok:", ok)
	if err != nil || !ok {
		log.Logger.Info("NotifyStopStress--???redis????????????mg???")
		rdb := dal.GetRDBForReport()
		key := fmt.Sprintf("reportData:%s", req.ReportID)
		dataList := rdb.LRange(ctx, key, 0, -1).Val()
		if len(dataList) < 1 {
			log.Logger.Info("NotifyStopStress--redis???????????????????????????????????????err:", proof.WithError(err))
			return nil
		}
		log.Logger.Info("NotifyStopStress--redis???????????????????????????????????????", len(dataList))
		for i := len(dataList) - 1; i >= 0; i-- {
			resultMsgString := dataList[i]
			err = json.Unmarshal([]byte(resultMsgString), &resultMsg)
			if err != nil {
				log.Logger.Info("NotifyStopStress--json?????????????????????err:", proof.WithError(err))
			}
			if resultData.Results == nil {
				resultData.Results = make(map[string]*report.ResultDataMsg)
			}
			log.Logger.Info("NotifyStopStress--?????????????????????????????????id??????", resultMsg.ReportId)
			resultData.ReportId = resultMsg.ReportId
			resultData.End = resultMsg.End
			resultData.ReportName = resultMsg.ReportName
			resultData.PlanId = resultMsg.PlanId
			resultData.PlanName = resultMsg.PlanName
			resultData.SceneId = resultMsg.SceneId
			resultData.SceneName = resultMsg.SceneName
			resultData.TimeStamp = resultMsg.TimeStamp
			if resultMsg.Results != nil && len(resultMsg.Results) > 0 {
				log.Logger.Info("NotifyStopStress--resultMsg.Results?????????end?????????", resultMsg.End)
				for k, apiResult := range resultMsg.Results {
					//log.Logger.Info("NotifyStopStress--????????????????????????")
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
					// qps??????
					timeValue.Value = resultData.Results[k].Rps
					resultData.Results[k].RpsList = append(resultData.Results[k].RpsList, timeValue)
					timeValue.Value = resultData.Results[k].Tps
					if resultData.Results[k].TpsList == nil {
						resultData.Results[k].TpsList = []report.TimeValue{}
					}
					// ???????????????
					resultData.Results[k].TpsList = append(resultData.Results[k].TpsList, timeValue)
					timeValue.Value = resultData.Results[k].Concurrency
					if resultData.Results[k].ConcurrencyList == nil {
						resultData.Results[k].ConcurrencyList = []report.TimeValue{}
					}
					// ???????????????
					resultData.Results[k].ConcurrencyList = append(resultData.Results[k].ConcurrencyList, timeValue)

					// ????????????????????????
					timeValue.Value = resultData.Results[k].AvgRequestTime
					if resultData.Results[k].AvgList == nil {
						resultData.Results[k].AvgList = []report.TimeValue{}
					}
					resultData.Results[k].AvgList = append(resultData.Results[k].AvgList, timeValue)

					// 50??????????????????
					timeValue.Value = resultData.Results[k].FiftyRequestTimelineValue
					if resultData.Results[k].FiftyList == nil {
						resultData.Results[k].FiftyList = []report.TimeValue{}
					}
					resultData.Results[k].FiftyList = append(resultData.Results[k].FiftyList, timeValue)

					// 90??????????????????
					timeValue.Value = resultData.Results[k].NinetyNineRequestTimeLineValue
					if resultData.Results[k].NinetyList == nil {
						resultData.Results[k].NinetyList = []report.TimeValue{}
					}
					resultData.Results[k].NinetyList = append(resultData.Results[k].NinetyList, timeValue)

					// 95??????????????????
					timeValue.Value = resultData.Results[k].NinetyFiveRequestTimeLineValue
					if resultData.Results[k].NinetyFiveList == nil {
						resultData.Results[k].NinetyFiveList = []report.TimeValue{}
					}
					resultData.Results[k].NinetyFiveList = append(resultData.Results[k].NinetyFiveList, timeValue)

					// 99??????????????????
					timeValue.Value = resultData.Results[k].NinetyNineRequestTimeLineValue
					if resultData.Results[k].NinetyNineList == nil {
						resultData.Results[k].NinetyNineList = []report.TimeValue{}
					}
					resultData.Results[k].NinetyNineList = append(resultData.Results[k].NinetyNineList, timeValue)
				}
				log.Logger.Info("NotifyStopStress--????????????????????????")
			}
			if resultMsg.End {
				log.Logger.Info("NotifyStopStress--??????????????????????????????")
				var by []byte
				by, err = json.Marshal(resultData)
				if err != nil {
					log.Logger.Info("NotifyStopStress--resultData??????????????????err:", proof.WithError(err))
					return err
				}
				var apiResultTotalMsg = make(map[string]string)
				for _, value := range resultData.Results {
					apiResultTotalMsg[value.ApiName] = fmt.Sprintf("?????????????????????%0.1fms??? ???????????????????????????????????????%0.1fms; ???????????????????????????????????????%0.1fms; ??????????????????????????????????????????%0.1fms; ??????????????????????????????????????????%0.1fms; RPS???%0.1f; SRPS???%0.1f; TPS???%0.1f; STPS???%0.1f",
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
				log.Logger.Info("NotifyStopStress--??????????????????mg????????????err:", proof.WithError(err))
				if err != nil {
					log.Logger.Info("NotifyStopStress--??????????????????mongo?????????err:", proof.WithError(err))
					return err
				}
				err = rdb.Del(ctx, key).Err()
				if err != nil {
					log.Logger.Info("NotifyStopStress--??????redis???key???", key, " err:", proof.WithError(err))
					return err
				}
			}
		}
	}

	return nil
}
