package autoPlan

import (
	"context"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/errno"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/jwt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/log"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/record"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/response"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/uuid"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/mao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/query"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/runner"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/notice"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/packer"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/public"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
	"math"
	"strconv"
	"strings"
	"time"
)

func SaveAutoPlan(ctx *gin.Context, req *rao.SaveAutoPlanReq) (string, int, error) {
	// 用户信息
	userID := jwt.GetUserIDByCtx(ctx)
	var rankID int64 = 1
	planID := uuid.GetUUID()

	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		// 查询当前计划名称是否存在
		_, err := tx.AutoPlan.WithContext(ctx).Where(tx.AutoPlan.TeamID.Eq(req.TeamID)).Where(tx.AutoPlan.PlanName.Eq(req.PlanName)).First()
		if err != nil && err != gorm.ErrRecordNotFound {
			log.Logger.Info("保存自动化测试计划，err:", err)
			return err
		}

		if err == nil { // 查到了
			return fmt.Errorf("名称已存在")
		}

		autoPlanInfo, err := tx.AutoPlan.WithContext(ctx).Where(tx.AutoPlan.TeamID.Eq(req.TeamID)).Order(tx.AutoPlan.RankID.Desc()).Limit(1).First()
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}

		if err == nil {
			rankID = autoPlanInfo.RankID + 1
		}

		// 不存在，则创建数据
		insertData := &model.AutoPlan{
			PlanName:     req.PlanName,
			RankID:       rankID,
			PlanID:       planID,
			TeamID:       req.TeamID,
			CreateUserID: userID,
			RunUserID:    userID,
			Remark:       req.Remark,
		}

		err = tx.AutoPlan.WithContext(ctx).Create(insertData)
		if err != nil {
			return err
		}

		// 保存一份默认的配置
		insertTaskConf := &model.AutoPlanTaskConf{
			PlanID:           planID,
			TeamID:           req.TeamID,
			TaskType:         consts.PlanTaskTypeNormal,
			TaskMode:         consts.AutoPlanTaskRunMode,
			SceneRunOrder:    consts.AutoPlanSceneRunModeOrder,
			TestCaseRunOrder: consts.AutoPlanTestCaseRunModeOrder,
			RunUserID:        userID,
		}
		err = tx.AutoPlanTaskConf.WithContext(ctx).Create(insertTaskConf)
		if err != nil {
			return err
		}

		if err := record.InsertDelete(ctx, req.TeamID, userID, record.OperationOperateCreatePlan, req.PlanName); err != nil {
			return err
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

func GetAutoPlanList(ctx *gin.Context, req *rao.GetAutoPlanListReq) ([]*rao.AutoPlanDetailResp, int64, error) {
	// 查询数据库
	tx := dal.GetQuery().AutoPlan
	// 查询数据库
	limit := req.Size
	offset := (req.Page - 1) * req.Size
	sort := make([]field.Expr, 0, 6)
	if req.Sort == 0 || req.Sort == 2 { // 默认排序， 创建时间倒序
		sort = append(sort, tx.CreatedAt.Desc())
	} else if req.Sort == 1 { // 创建时间升序
		sort = append(sort, tx.CreatedAt)
	} else if req.Sort == 3 { // 最后修改时间升序
		sort = append(sort, tx.UpdatedAt)
	} else { // 最后修改时间倒序
		sort = append(sort, tx.UpdatedAt.Desc())
	}

	conditions := make([]gen.Condition, 0)
	conditions = append(conditions, tx.TeamID.Eq(req.TeamId))

	if req.PlanName != "" {
		conditions = append(conditions, tx.PlanName.Like(fmt.Sprintf("%%%s%%", req.PlanName)))
		// 先查询出来用户id
		userTable := dal.GetQuery().User
		userList, err := userTable.WithContext(ctx).Where(userTable.Nickname.Like(fmt.Sprintf("%%%s%%", req.PlanName))).Find()
		if err == nil {
			tempUserIDs := make([]string, 0, len(userList))
			for _, userInfo := range userList {
				tempUserIDs = append(tempUserIDs, userInfo.UserID)
			}

			// 查询属于当前团队的用户
			userTeamTable := dal.GetQuery().UserTeam
			userTeamList, err := userTeamTable.WithContext(ctx).Where(userTeamTable.TeamID.Eq(req.TeamId),
				userTeamTable.UserID.In(tempUserIDs...)).Find()
			if err == nil {
				userIDs := make([]string, 0, len(userList))
				for _, vutInfo := range userTeamList {
					userIDs = append(userIDs, vutInfo.UserID)
				}
				if len(userIDs) > 0 {
					conditions[1] = tx.RunUserID.In(userIDs...)
				}
			}
		}
	}

	if req.TaskType != 0 {
		conditions = append(conditions, tx.TaskType.Eq(req.TaskType))
	}

	if req.Status != 0 {
		conditions = append(conditions, tx.Status.Eq(req.Status))
	}

	if (req.StartTimeSec > 0 && req.EndTimeSec > 0) && (req.EndTimeSec > req.StartTimeSec) {
		startTime := time.Unix(req.StartTimeSec, 0)
		endTime := time.Unix(req.EndTimeSec, 0)
		conditions = append(conditions, tx.CreatedAt.Between(startTime, endTime))
	}

	list, total, err := tx.WithContext(ctx).Where(conditions...).Order(sort...).FindByPage(offset, limit)
	if err != nil {
		log.Logger.Info("自动化计划列表--获取列表失败，err:", err)
		return nil, 0, err
	}

	// 获取所有操作人id
	runUserIDs := make([]string, 0, len(list))
	for _, detail := range list {
		runUserIDs = append(runUserIDs, detail.RunUserID)
	}

	userTable := dal.GetQuery().User
	userList, err := userTable.WithContext(ctx).Where(userTable.UserID.In(runUserIDs...)).Find()
	if err != nil {
		return nil, 0, err
	}
	// 用户id和名称映射
	userMap := make(map[string]string, len(userList))
	for _, userValue := range userList {
		userMap[userValue.UserID] = userValue.Nickname
	}

	res := make([]*rao.AutoPlanDetailResp, 0, len(list))
	for _, detail := range list {
		detailTmp := &rao.AutoPlanDetailResp{
			RankID:    detail.RankID,
			PlanID:    detail.PlanID,
			TeamID:    detail.TeamID,
			PlanName:  detail.PlanName,
			TaskType:  detail.TaskType,
			CreatedAt: detail.CreatedAt.Unix(),
			UpdatedAt: detail.UpdatedAt.Unix(),
			Status:    detail.Status,
			Remark:    detail.Remark,
			UserName:  userMap[detail.RunUserID],
		}
		res = append(res, detailTmp)
	}
	return res, total, nil
}

func DeleteAutoPlan(ctx *gin.Context, req *rao.DeleteAutoPlanReq) error {
	userID := jwt.GetUserIDByCtx(ctx)
	err := dal.GetQuery().Transaction(func(tx *query.Query) error {
		planInfo, err := tx.AutoPlan.WithContext(ctx).Where(tx.AutoPlan.TeamID.Eq(req.TeamID)).Where(tx.AutoPlan.PlanID.Eq(req.PlanID)).First()
		if err != nil {
			return err
		}

		if planInfo.Status == consts.PlanStatusUnderway {
			return fmt.Errorf("该计划正在运行，无法删除")
		}

		// 删除计划基本信息
		_, err = tx.AutoPlan.WithContext(ctx).Where(tx.AutoPlan.TeamID.Eq(req.TeamID)).Where(tx.AutoPlan.PlanID.Eq(req.PlanID)).Delete()
		if err != nil {
			return err
		}

		// 查询计划下所创建的数据
		targetList, err := tx.Target.WithContext(ctx).Where(tx.Target.TeamID.Eq(req.TeamID),
			tx.Target.PlanID.Eq(req.PlanID), tx.Target.Source.Eq(consts.TargetSourceAutoPlan),
		).Find()
		if err == nil {
			// 组装场景集合
			sceneIDs := make([]string, 0, len(targetList))
			caseIDs := make([]string, 0, len(targetList))

			needDeleteIDs := make([]string, 0, len(targetList))
			for _, targetInfo := range targetList {
				needDeleteIDs = append(needDeleteIDs, targetInfo.TargetID)

				if targetInfo.TargetType == "scene" {
					sceneIDs = append(sceneIDs, targetInfo.TargetID)
				}
				if targetInfo.TargetType == "test_case" {
					caseIDs = append(caseIDs, targetInfo.TargetID)
				}
			}

			// 删除计划下所有创建的数据
			_, err := tx.Target.WithContext(ctx).Where(tx.Target.TargetID.In(needDeleteIDs...)).Delete()
			if err != nil {
				return err
			}

			// 删除计划下所有场景
			if len(sceneIDs) > 0 {
				if _, err = tx.Target.WithContext(ctx).Where(tx.Target.TeamID.Eq(req.TeamID),
					tx.Target.PlanID.Eq(req.PlanID), tx.Target.TargetID.In(sceneIDs...)).Delete(); err != nil {
					return err
				}

				// 从mg里面删除当前场景对应的flow
				collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectFlow)
				_, err = collection.DeleteMany(ctx, bson.D{{"scene_id", bson.D{{"$in", sceneIDs}}}})

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

			// 删除计划下所有场景下的所有用例
			if len(caseIDs) > 0 {
				if _, err = tx.Target.WithContext(ctx).Where(tx.Target.TeamID.Eq(req.TeamID),
					tx.Target.PlanID.Eq(req.PlanID), tx.Target.TargetID.In(caseIDs...)).Delete(); err != nil {
					return err
				}

				// 从mg里面删除当前场景对应的所有用例的flow
				collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectSceneCaseFlow)
				_, err = collection.DeleteMany(ctx, bson.D{{"scene_case_id", bson.D{{"$in", caseIDs}}}})
				if err != nil {
					return err
				}
			}

		}

		//删除计划下的配置--普通任务
		_, err = tx.AutoPlanTaskConf.WithContext(ctx).Where(tx.AutoPlanTaskConf.TeamID.Eq(req.TeamID),
			tx.AutoPlanTaskConf.PlanID.Eq(req.PlanID)).Delete()
		if err != nil {
			return err
		}

		//删除计划下的配置--定时任务任务
		_, err = tx.AutoPlanTimedTaskConf.WithContext(ctx).Where(tx.AutoPlanTimedTaskConf.TeamID.Eq(req.TeamID),
			tx.AutoPlanTimedTaskConf.PlanID.Eq(req.PlanID)).Delete()
		if err != nil {
			return err
		}

		return record.InsertDelete(ctx, req.TeamID, userID, record.OperationOperateDeletePlan, planInfo.PlanName)
	})
	return err
}

func GetAutoPlanDetail(ctx *gin.Context, req *rao.GetAutoPlanDetailReq) (*rao.GetAutoPlanDetailResp, error) {
	// 获取计划详情
	tx := dal.GetQuery().AutoPlan
	planInfo, err := tx.WithContext(ctx).Where(tx.TeamID.Eq(req.TeamID)).Where(tx.PlanID.Eq(req.PlanID)).First()
	if err != nil {
		return nil, err
	}

	// 查询创建人
	tableUser := dal.GetQuery().User
	userInfo, err := tableUser.WithContext(ctx).Where(tableUser.UserID.Eq(planInfo.CreateUserID)).First()
	if err != nil {
		return nil, err
	}

	res := &rao.GetAutoPlanDetailResp{
		PlanID:    planInfo.PlanID,
		TeamID:    planInfo.TeamID,
		PlanName:  planInfo.PlanName,
		CreatedAt: planInfo.CreatedAt.Unix(),
		UpdatedAt: planInfo.UpdatedAt.Unix(),
		Remark:    planInfo.Remark,
		Status:    planInfo.Status,
		UserName:  userInfo.Nickname,
		Avatar:    userInfo.Avatar,
	}
	return res, nil
}

func CopyAutoPlan(ctx *gin.Context, req *rao.CopyAutoPlanReq) error {

	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		// 1、查询原来计划的信息
		autoPlanTable := tx.AutoPlan
		oldPlanInfo, err := autoPlanTable.WithContext(ctx).Where(autoPlanTable.TeamID.Eq(req.TeamID)).Where(autoPlanTable.PlanID.Eq(req.PlanID)).First()
		if err != nil {
			return err
		}

		oldPlanName := oldPlanInfo.PlanName
		newPlanName := oldPlanName + "_1"

		// 查询老配置相关的
		list, err := autoPlanTable.WithContext(ctx).Where(autoPlanTable.TeamID.Eq(req.TeamID)).Where(autoPlanTable.PlanName.Like(fmt.Sprintf("%s%%", oldPlanName+"_"))).Find()
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
					log.Logger.Info("复制自动化计划--类型转换失败，err:", err)
					continue
				}
				if currentNum > maxNum {
					maxNum = currentNum
				}
			}
			newPlanName = oldPlanName + fmt.Sprintf("_%d", maxNum+1)
		}

		nameLength := public.GetStringNum(newPlanName)
		if nameLength > 30 { // 场景名称限制30个字符
			return fmt.Errorf("名称过长！不可超出30字符")
		}

		// 查询当前团队内的计划最大
		var rankID int64 = 1
		autoPlanInfo, err := autoPlanTable.WithContext(ctx).Where(autoPlanTable.TeamID.Eq(req.TeamID)).Order(autoPlanTable.RankID.Desc()).Limit(1).First()
		if err == nil { // 查到了
			rankID = autoPlanInfo.RankID + 1
		}

		// 用户信息
		userID := jwt.GetUserIDByCtx(ctx)

		newPlanID := uuid.GetUUID()

		oldPlanInfo.ID = 0
		oldPlanInfo.PlanID = newPlanID
		oldPlanInfo.RankID = rankID
		oldPlanInfo.PlanName = newPlanName
		oldPlanInfo.Status = consts.PlanStatusNormal
		oldPlanInfo.RunUserID = userID
		oldPlanInfo.CreateUserID = userID
		oldPlanInfo.CreatedAt = time.Now()
		oldPlanInfo.UpdatedAt = time.Now()
		err = autoPlanTable.WithContext(ctx).Create(oldPlanInfo)
		if err != nil {
			log.Logger.Info("复制计划--复制计划基本数据失败，err:", err)
			return err
		}

		// 复制计划下场景,分组
		targetTable := tx.Target
		oldTargetList, err := targetTable.WithContext(ctx).Where(targetTable.TeamID.Eq(req.TeamID),
			targetTable.PlanID.Eq(req.PlanID), targetTable.Source.Eq(consts.TargetSourceAutoPlan),
			targetTable.Status.Eq(consts.TargetStatusNormal), targetTable.TargetType.In(consts.TargetTypeScene, consts.TargetTypeFolder),
		).Order(targetTable.ParentID).Find()

		oldSceneIDs := make([]string, 0, len(oldTargetList))
		sceneIDOldNewMap := make(map[string]string)
		if err == nil {
			for _, oldTargetInfo := range oldTargetList {
				if oldTargetInfo.TargetType == consts.TargetTypeScene {
					oldSceneIDs = append(oldSceneIDs, oldTargetInfo.TargetID)
				}

				// 新的sceneID
				newSceneID := uuid.GetUUID()

				oldTargetID := oldTargetInfo.TargetID
				oldTargetInfo.ID = 0
				oldTargetInfo.TargetID = newSceneID
				oldTargetInfo.ParentID = sceneIDOldNewMap[oldTargetInfo.ParentID]
				oldTargetInfo.PlanID = newPlanID
				oldTargetInfo.CreatedUserID = userID
				oldTargetInfo.RecentUserID = userID
				oldTargetInfo.CreatedAt = time.Now()
				oldTargetInfo.UpdatedAt = time.Now()
				if err := tx.Target.WithContext(ctx).Create(oldTargetInfo); err != nil {
					return err
				}

				sceneIDOldNewMap[oldTargetID] = newSceneID
			}
		}

		// 复制场景详情flow
		for _, oldSceneID := range oldSceneIDs {
			flow := mao.Flow{}
			c1 := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectFlow)
			err = c1.FindOne(ctx, bson.D{{"scene_id", oldSceneID}}).Decode(&flow)
			if err == nil {
				flow.SceneID = sceneIDOldNewMap[oldSceneID]
				// 更新api的uuid
				err := packer.ChangeSceneNodeUUID(&flow)
				if err != nil {
					log.Logger.Info("复制计划--替换event_id失败")
					return err
				}
				if _, err := c1.InsertOne(ctx, flow); err != nil {
					return err
				}
			}
		}

		// 复制测试用例
		testCaseList, err := targetTable.WithContext(ctx).Where(targetTable.TeamID.Eq(req.TeamID),
			targetTable.PlanID.Eq(req.PlanID), targetTable.ParentID.In(oldSceneIDs...),
			targetTable.TargetType.Eq(consts.TargetTypeTestCase)).Find()
		if err == nil {
			oldAndNewCaseIDMap := make(map[string]string, len(testCaseList))
			oldTestCaseIDs := make([]string, 0, len(testCaseList))
			oldCaseAndNewParentIDMap := make(map[string]string)
			for _, testCaseInfo := range testCaseList {
				oldTestCaseIDs = append(oldTestCaseIDs, testCaseInfo.TargetID)
				oldCaseID := testCaseInfo.TargetID
				oldCaseAndNewParentIDMap[oldCaseID] = sceneIDOldNewMap[testCaseInfo.ParentID]

				//新的caseID
				newCaseID := uuid.GetUUID()

				testCaseInfo.ID = 0
				testCaseInfo.TargetID = newCaseID
				testCaseInfo.ParentID = sceneIDOldNewMap[testCaseInfo.ParentID]
				testCaseInfo.PlanID = newPlanID
				testCaseInfo.CreatedUserID = userID
				testCaseInfo.RecentUserID = userID
				testCaseInfo.CreatedAt = time.Now()
				testCaseInfo.UpdatedAt = time.Now()
				if err := targetTable.WithContext(ctx).Create(testCaseInfo); err != nil {
					return err
				}
				oldAndNewCaseIDMap[oldCaseID] = newCaseID
			}

			// 克隆用例的详情
			for _, oldCaseID := range oldTestCaseIDs {
				var sceneCaseFlow mao.SceneCaseFlow
				collection3 := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectSceneCaseFlow)
				err := collection3.FindOne(ctx, bson.D{{"scene_case_id", oldCaseID}}).Decode(&sceneCaseFlow)
				if err == nil {
					sceneCaseFlow.SceneCaseID = oldAndNewCaseIDMap[oldCaseID]
					sceneCaseFlow.SceneID = oldCaseAndNewParentIDMap[oldCaseID]
					// 更新testCase的uuid
					err = packer.ChangeCaseNodeUUID(&sceneCaseFlow)
					if err != nil {
						log.Logger.Info("复制计划--替换用例event_id失败")
						return err
					}
					if _, err := collection3.InsertOne(ctx, sceneCaseFlow); err != nil {
						return err
					}
				}
			}
		}

		// 克隆场景变量
		for _, oldSceneID := range oldSceneIDs {
			collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectSceneParam)
			cur, err := collection.Find(ctx, bson.D{{"team_id", req.TeamID}, {"scene_id", oldSceneID}})
			var sceneParamDataArr []*mao.SceneParamData
			if err == nil {
				if err := cur.All(ctx, &sceneParamDataArr); err != nil {
					return fmt.Errorf("场景参数数据获取失败")
				}
				for _, sv := range sceneParamDataArr {
					sv.SceneID = sceneIDOldNewMap[oldSceneID]
					if _, err := collection.InsertOne(ctx, sv); err != nil {
						return err
					}
				}
			}
		}

		// 克隆导入变量
		variableImportTable := tx.VariableImport
		variableImportList, err := variableImportTable.WithContext(ctx).Where(variableImportTable.SceneID.In(oldSceneIDs...)).Find()
		if err == nil {
			for _, variableImportInfo := range variableImportList {
				variableImportInfo.ID = 0
				variableImportInfo.SceneID = sceneIDOldNewMap[variableImportInfo.SceneID]
				variableImportInfo.CreatedAt = time.Now()
				variableImportInfo.UpdatedAt = time.Now()
				if err := tx.VariableImport.WithContext(ctx).Create(variableImportInfo); err != nil {
					return err
				}
			}
		}

		// 判断当前计划的任务类型
		if oldPlanInfo.TaskType == consts.PlanTaskTypeNormal { // 普通任务
			// 复制普通任务配置
			autoPlanTaskConfTable := tx.AutoPlanTaskConf
			autoPlanTaskConfInfo, err := autoPlanTaskConfTable.WithContext(ctx).Where(autoPlanTaskConfTable.TeamID.Eq(req.TeamID), autoPlanTaskConfTable.PlanID.Eq(req.PlanID)).First()
			if err == nil {
				autoPlanTaskConfInfo.ID = 0
				autoPlanTaskConfInfo.PlanID = newPlanID
				autoPlanTaskConfInfo.RunUserID = userID
				autoPlanTaskConfInfo.CreatedAt = time.Now()
				autoPlanTaskConfInfo.UpdatedAt = time.Now()
				if err := autoPlanTaskConfTable.WithContext(ctx).Create(autoPlanTaskConfInfo); err != nil {
					return err
				}
			}
		} else { // 定时任务
			// 复制普通任务配置
			autoPlanTimedTaskConfTable := tx.AutoPlanTimedTaskConf
			autoPlanTimedTaskConfInfo, err := autoPlanTimedTaskConfTable.WithContext(ctx).Where(autoPlanTimedTaskConfTable.TeamID.Eq(req.TeamID), autoPlanTimedTaskConfTable.PlanID.Eq(req.PlanID)).First()
			if err == nil {
				autoPlanTimedTaskConfInfo.ID = 0
				autoPlanTimedTaskConfInfo.PlanID = newPlanID
				autoPlanTimedTaskConfInfo.Status = 0
				autoPlanTimedTaskConfInfo.RunUserID = userID
				autoPlanTimedTaskConfInfo.CreatedAt = time.Now()
				autoPlanTimedTaskConfInfo.UpdatedAt = time.Now()
				if err := autoPlanTimedTaskConfTable.WithContext(ctx).Create(autoPlanTimedTaskConfInfo); err != nil {
					return err
				}
			}
		}

		return nil
	})

	return err
}

func UpdateAutoPlan(ctx *gin.Context, req *rao.UpdateAutoPlanReq) error {

	return dal.GetQuery().Transaction(func(tx *query.Query) error {
		_, err := tx.AutoPlan.WithContext(ctx).Where(tx.AutoPlan.TeamID.Eq(req.TeamID)).Where(tx.AutoPlan.PlanID.Neq(req.PlanID)).Where(tx.AutoPlan.PlanName.Eq(req.PlanName)).First()
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		if err == nil {
			return fmt.Errorf("名称已存在")
		}

		updateData := make(map[string]interface{}, 1)
		if req.PlanName != "" {
			updateData["plan_name"] = req.PlanName
		}
		if req.Remark != "" {
			updateData["remark"] = req.Remark
		}
		if len(updateData) <= 0 {
			return nil
		}
		_, err = tx.AutoPlan.WithContext(ctx).Where(tx.AutoPlan.TeamID.Eq(req.TeamID)).Where(tx.AutoPlan.PlanID.Eq(req.PlanID)).Updates(updateData)
		if err != nil {
			return err
		}
		return nil
	})
}

func AddEmail(ctx *gin.Context, req *rao.AddEmailReq) error {
	tx := dal.GetQuery().AutoPlanEmail
	emailCount, err := tx.WithContext(ctx).Where(tx.TeamID.Eq(req.TeamID), tx.PlanID.Eq(req.PlanID), tx.Email.In(req.Emails...)).Count()
	if err != nil {
		return err
	}

	if emailCount > 0 {
		return fmt.Errorf("邮箱已存在")
	}

	insertData := make([]*model.AutoPlanEmail, 0, len(req.Emails))
	for _, email := range req.Emails {
		insertDataTmp := &model.AutoPlanEmail{
			PlanID: req.PlanID,
			TeamID: req.TeamID,
			Email:  email,
		}
		insertData = append(insertData, insertDataTmp)
	}

	if err := tx.WithContext(ctx).CreateInBatches(insertData, 5); err != nil {
		return err
	}
	return nil
}

func GetEmailList(ctx *gin.Context, req *rao.GetEmailListReq) ([]*rao.AutoPlanEmail, error) {
	tx := dal.GetQuery().AutoPlanEmail
	emailList, err := tx.WithContext(ctx).Where(tx.TeamID.Eq(req.TeamID)).Where(tx.PlanID.Eq(req.PlanID)).Find()
	if err != nil {
		return nil, err
	}

	res := make([]*rao.AutoPlanEmail, 0, len(emailList))
	for _, emailInfo := range emailList {
		tmpData := &rao.AutoPlanEmail{
			ID:     emailInfo.ID,
			TeamID: emailInfo.TeamID,
			PlanID: emailInfo.PlanID,
			Email:  emailInfo.Email,
		}
		res = append(res, tmpData)
	}
	return res, nil
}

func DeleteEmail(ctx *gin.Context, req *rao.DeleteEmailReq) error {
	// 删除邮箱
	tx := dal.GetQuery().AutoPlanEmail
	_, err := tx.WithContext(ctx).Where(tx.ID.Eq(req.ID)).Delete()
	if err != nil {
		return err
	}
	return nil
}

func BatchDeleteAutoPlan(ctx *gin.Context, req *rao.BatchDeleteAutoPlanReq) error {
	// 删除计划
	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		planList, err := tx.AutoPlan.WithContext(ctx).Where(tx.AutoPlan.PlanID.In(req.PlanIDs...)).Find()
		if err != nil {
			return err
		}

		if len(planList) > 0 {
			for _, planInfo := range planList {
				if planInfo.Status == consts.PlanStatusUnderway {
					return fmt.Errorf("存在运行中的计划，无法删除")
				}
			}
		}

		_, err = tx.AutoPlan.WithContext(ctx).Where(tx.AutoPlan.TeamID.Eq(req.TeamID), tx.AutoPlan.PlanID.In(req.PlanIDs...)).Delete()
		if err != nil {
			return err
		}

		targetList, err := tx.Target.WithContext(ctx).Where(tx.Target.PlanID.In(req.PlanIDs...),
			tx.Target.Source.Eq(consts.TargetSourceAutoPlan),
			tx.Target.TargetType.In(consts.TargetTypeTestCase, consts.TargetTypeScene)).Find()
		if err != nil {
			return err
		}
		sceneIDs := make([]string, 0, len(targetList))
		caseIDs := make([]string, 0, len(targetList))
		for _, targetInfo := range targetList {
			if targetInfo.TargetType == consts.TargetTypeScene {
				sceneIDs = append(sceneIDs, targetInfo.TargetID)
			}
			if targetInfo.TargetType == consts.TargetTypeTestCase {
				caseIDs = append(caseIDs, targetInfo.TargetID)
			}
		}

		//删除所有场景内的接口详情
		if len(sceneIDs) > 0 {
			// 删除场景下的flow
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

		// 删除计划下所有场景下的所有用例
		if len(caseIDs) > 0 {
			// 从mg里面删除当前场景对应的所有用例的flow
			collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectSceneCaseFlow)
			_, err = collection.DeleteMany(ctx, bson.D{{"scene_case_id", bson.D{{"$in", caseIDs}}}})
			if err != nil {
				return err
			}
		}

		_, err = tx.Target.WithContext(ctx).Where(tx.Target.PlanID.In(req.PlanIDs...), tx.Target.Source.Eq(consts.TargetSourceAutoPlan)).Delete()
		if err != nil {
			return err
		}

		return nil
	})
	return err
}

// SaveTaskConf 保存计划配置
func SaveTaskConf(ctx *gin.Context, req *rao.SaveTaskConfReq) error {
	userID := jwt.GetUserIDByCtx(ctx)
	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		if req.TaskType == consts.PlanTaskTypeNormal { // 普通任务
			// 删除定时任务配置
			_, err := tx.AutoPlanTimedTaskConf.WithContext(ctx).Where(tx.AutoPlanTimedTaskConf.TeamID.Eq(req.TeamID),
				tx.AutoPlanTimedTaskConf.PlanID.Eq(req.PlanID)).Delete()
			if err != nil {
				return err
			}

			taskConfInfo, err := tx.AutoPlanTaskConf.WithContext(ctx).Where(tx.AutoPlanTaskConf.TeamID.Eq(req.TeamID), tx.AutoPlanTaskConf.PlanID.Eq(req.PlanID)).First()
			if err != nil && err != gorm.ErrRecordNotFound {
				return err
			}

			if err == nil { // 查到已存在，则修改
				updateData := make(map[string]interface{}, 10)
				updateData["task_mode"] = req.TaskMode
				updateData["scene_run_order"] = req.SceneRunOrder
				updateData["test_case_run_order"] = req.TestCaseRunOrder
				updateData["run_user_id"] = userID
				_, err := tx.AutoPlanTaskConf.WithContext(ctx).Where(tx.AutoPlanTaskConf.ID.Eq(taskConfInfo.ID)).Updates(updateData)
				if err != nil {
					return err
				}
			} else { // 没查到则新增
				newData := &model.AutoPlanTaskConf{
					PlanID:           req.PlanID,
					TeamID:           req.TeamID,
					TaskType:         req.TaskType,
					TaskMode:         req.TaskMode,
					SceneRunOrder:    req.SceneRunOrder,
					TestCaseRunOrder: req.TestCaseRunOrder,
					RunUserID:        userID,
				}
				err := tx.AutoPlanTaskConf.WithContext(ctx).Create(newData)
				if err != nil {
					return err
				}
			}
		} else { // 定时任务
			// 删除普通任务配置
			_, err := tx.AutoPlanTaskConf.WithContext(ctx).Where(tx.AutoPlanTaskConf.TeamID.Eq(req.TeamID),
				tx.AutoPlanTaskConf.PlanID.Eq(req.PlanID)).Delete()
			if err != nil {
				return err
			}

			timedTaskConfInfo, err := tx.AutoPlanTimedTaskConf.WithContext(ctx).Where(tx.AutoPlanTimedTaskConf.TeamID.Eq(req.TeamID), tx.AutoPlanTimedTaskConf.PlanID.Eq(req.PlanID)).First()
			if err != nil && err != gorm.ErrRecordNotFound {
				return err
			}

			// 检测定时任务时间
			nowTime := time.Now().Unix()
			if req.Frequency == 0 {
				if req.TaskExecTime < nowTime {
					return fmt.Errorf("开始或结束时间不能早于当前时间")
				}
			} else {
				if req.TaskCloseTime < nowTime {
					return fmt.Errorf("开始或结束时间不能早于当前时间")
				}
			}

			if err == nil { // 查到已存在，则修改
				updateData := make(map[string]interface{}, 10)
				updateData["task_mode"] = req.TaskMode
				updateData["frequency"] = req.Frequency
				updateData["task_exec_time"] = req.TaskExecTime
				updateData["task_close_time"] = req.TaskCloseTime
				updateData["scene_run_order"] = req.SceneRunOrder
				updateData["test_case_run_order"] = req.TestCaseRunOrder
				updateData["Status"] = 0
				updateData["run_user_id"] = userID
				_, err := tx.AutoPlanTimedTaskConf.WithContext(ctx).Where(tx.AutoPlanTimedTaskConf.ID.Eq(timedTaskConfInfo.ID)).Updates(updateData)
				if err != nil {
					return err
				}
			} else { // 没查到则新增
				newData := &model.AutoPlanTimedTaskConf{
					PlanID:           req.PlanID,
					TeamID:           req.TeamID,
					TaskType:         req.TaskType,
					TaskMode:         req.TaskMode,
					SceneRunOrder:    req.SceneRunOrder,
					TestCaseRunOrder: req.TestCaseRunOrder,
					Frequency:        req.Frequency,
					TaskExecTime:     req.TaskExecTime,
					TaskCloseTime:    req.TaskCloseTime,
					Status:           0,
					RunUserID:        userID,
				}
				err := tx.AutoPlanTimedTaskConf.WithContext(ctx).Create(newData)
				if err != nil {
					return err
				}
			}
		}

		// 修改计划类型
		_, err := tx.AutoPlan.WithContext(ctx).Where(tx.AutoPlan.TeamID.Eq(req.TeamID), tx.AutoPlan.PlanID.Eq(req.PlanID)).UpdateSimple(tx.AutoPlan.TaskType.Value(req.TaskType))
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// GetTaskConf 获取计划配置
func GetTaskConf(ctx *gin.Context, req *rao.GetTaskConfReq) (*rao.GetTaskConfResp, error) {
	// 获取计划信息
	tx := dal.GetQuery().AutoPlan
	planInfo, err := tx.WithContext(ctx).Where(tx.TeamID.Eq(req.TeamID), tx.PlanID.Eq(req.PlanID)).First()
	if err != nil {
		return nil, err
	}

	res := &rao.GetTaskConfResp{}
	if planInfo.TaskType == consts.PlanTaskTypeNormal { // 普通任务
		tx := dal.GetQuery().AutoPlanTaskConf
		taskConfInfo, err := tx.WithContext(ctx).Where(tx.TeamID.Eq(req.TeamID), tx.PlanID.Eq(req.PlanID)).First()
		if err != nil {
			return nil, err
		}
		res = &rao.GetTaskConfResp{
			PlanID:           taskConfInfo.PlanID,
			TeamID:           taskConfInfo.TeamID,
			TaskType:         taskConfInfo.TaskType,
			TaskMode:         taskConfInfo.TaskMode,
			SceneRunOrder:    taskConfInfo.SceneRunOrder,
			TestCaseRunOrder: taskConfInfo.TestCaseRunOrder,
		}
	} else { // 定时任务
		tx := dal.GetQuery().AutoPlanTimedTaskConf
		timedTaskConfInfo, err := tx.WithContext(ctx).Where(tx.TeamID.Eq(req.TeamID), tx.PlanID.Eq(req.PlanID)).First()
		if err != nil {
			return nil, err
		}
		res = &rao.GetTaskConfResp{
			PlanID:           timedTaskConfInfo.PlanID,
			TeamID:           timedTaskConfInfo.TeamID,
			TaskType:         timedTaskConfInfo.TaskType,
			TaskMode:         timedTaskConfInfo.TaskMode,
			SceneRunOrder:    timedTaskConfInfo.SceneRunOrder,
			TestCaseRunOrder: timedTaskConfInfo.TestCaseRunOrder,
			Frequency:        timedTaskConfInfo.Frequency,
			TaskExecTime:     timedTaskConfInfo.TaskExecTime,
			TaskCloseTime:    timedTaskConfInfo.TaskCloseTime,
			Status:           timedTaskConfInfo.Status,
		}

		if timedTaskConfInfo.Frequency == 0 { // 频次一次
			res.TaskCloseTime = 0
		}
	}
	return res, nil
}

func GetAutoPlanReportList(ctx *gin.Context, req *rao.GetAutoPlanReportListReq) ([]*rao.GetAutoPlanReportList, int64, error) {
	// 查询数据库
	tx := dal.GetQuery().AutoPlanReport
	// 查询数据库
	limit := req.Size
	offset := (req.Page - 1) * req.Size
	sort := make([]field.Expr, 0, 6)
	if req.Sort == 0 || req.Sort == 2 { // 默认排序， 创建时间倒序
		sort = append(sort, tx.CreatedAt.Desc())
	} else if req.Sort == 1 { // 创建时间升序
		sort = append(sort, tx.CreatedAt)
	}

	conditions := make([]gen.Condition, 0)
	conditions = append(conditions, tx.TeamID.Eq(req.TeamId))
	if req.PlanName != "" {
		conditions = append(conditions, tx.PlanName.Like(fmt.Sprintf("%%%s%%", req.PlanName)))
		// 先查询出来用户id
		userTable := dal.GetQuery().User
		userList, err := userTable.WithContext(ctx).Where(userTable.Nickname.Like(fmt.Sprintf("%%%s%%", req.PlanName))).Find()
		if err == nil && len(userList) > 0 {
			tempUserIDs := make([]string, 0, len(userList))
			for _, userInfo := range userList {
				tempUserIDs = append(tempUserIDs, userInfo.UserID)
			}

			// 查询属于当前团队的用户
			userTeamTable := dal.GetQuery().UserTeam
			userTeamList, err := userTeamTable.WithContext(ctx).Where(userTeamTable.TeamID.Eq(req.TeamId),
				userTeamTable.UserID.In(tempUserIDs...)).Find()
			if err == nil && len(userTeamList) > 0 {
				userIDs := make([]string, 0, len(userList))
				for _, vutInfo := range userTeamList {
					userIDs = append(userIDs, vutInfo.UserID)
				}
				if len(userIDs) > 0 {
					conditions[1] = tx.RunUserID.In(userIDs...)
				}
			}
		}
	}

	if req.TaskType != 0 {
		conditions = append(conditions, tx.TaskType.Eq(req.TaskType))
	}

	if req.TaskType != 0 {
		conditions = append(conditions, tx.TaskType.Eq(req.TaskType))
	}

	if req.Status != 0 {
		conditions = append(conditions, tx.Status.Eq(req.Status))
	}

	if (req.StartTimeSec > 0 && req.EndTimeSec > 0) && (req.EndTimeSec > req.StartTimeSec) {
		startTime := time.Unix(req.StartTimeSec, 0)
		endTime := time.Unix(req.EndTimeSec, 0)
		conditions = append(conditions, tx.CreatedAt.Between(startTime, endTime))
	}

	list, total, err := tx.WithContext(ctx).Where(conditions...).Order(sort...).FindByPage(offset, limit)
	if err != nil {
		log.Logger.Info("自动化计划报告列表--获取列表失败，err:", err)
		return nil, 0, err
	}

	// 获取所有创建人id
	createUserIDs := make([]string, 0, len(list))
	for _, detail := range list {
		createUserIDs = append(createUserIDs, detail.RunUserID)
	}

	userTable := dal.GetQuery().User
	userList, err := userTable.WithContext(ctx).Where(userTable.UserID.In(createUserIDs...)).Find()
	if err != nil {
		return nil, 0, err
	}
	// 用户id和名称映射
	userMap := make(map[string]string, len(userList))
	for _, userValue := range userList {
		userMap[userValue.UserID] = userValue.Nickname
	}

	res := make([]*rao.GetAutoPlanReportList, 0, len(list))
	for _, detail := range list {
		detailTmp := &rao.GetAutoPlanReportList{
			RankID:           detail.RankID,
			ReportID:         detail.ReportID,
			ReportName:       detail.ReportName,
			PlanID:           detail.PlanID,
			TeamID:           detail.TeamID,
			PlanName:         detail.PlanName,
			TaskType:         detail.TaskType,
			TaskMode:         detail.TaskMode,
			SceneRunOrder:    detail.SceneRunOrder,
			TestCaseRunOrder: detail.TestCaseRunOrder,
			StartTimeSec:     detail.CreatedAt.Unix(),
			EndTimeSec:       detail.UpdatedAt.Unix(),
			Status:           detail.Status,
			Remark:           detail.Remark,
			RunUserName:      userMap[detail.RunUserID],
		}
		res = append(res, detailTmp)
	}
	return res, total, nil
}

func BatchDeleteAutoPlanReport(ctx *gin.Context, req *rao.BatchDeleteAutoPlanReportReq) error {
	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		_, err := tx.AutoPlanReport.WithContext(ctx).Where(tx.AutoPlanReport.TeamID.Eq(req.TeamID),
			tx.AutoPlanReport.ReportID.In(req.ReportIDs...)).Delete()
		if err != nil {
			return err
		}

		// 删除报告详情
		reportStringSLice := make([]string, 0, len(req.ReportIDs))
		for _, rId := range req.ReportIDs {
			reportStringSLice = append(reportStringSLice, fmt.Sprintf("%s", rId))
		}

		// 获取所有用例运行结果
		collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectAutoReport)
		_, err = collection.DeleteMany(ctx, bson.D{{"report_id", bson.D{{"$in", reportStringSLice}}}})
		if err != nil {
			return err
		}

		// 获取所有用例运行结果
		collection = dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectAutoReportDetailData)
		_, err = collection.DeleteMany(ctx, bson.D{{"report_id", bson.D{{"$in", reportStringSLice}}}})
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

func CloneAutoPlanScene(ctx *gin.Context, req *rao.CloneAutoPlanSceneReq) error {
	userID := jwt.GetUserIDByCtx(ctx)
	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		// 1、查询被复制的场景信息
		targetTable := tx.Target

		//组装条件
		conditions := make([]gen.Condition, 0)
		conditions = append(conditions, targetTable.TeamID.Eq(req.TeamID))
		conditions = append(conditions, targetTable.TargetID.Eq(req.SceneID))
		if req.PlanID != "" {
			conditions = append(conditions, targetTable.PlanID.Eq(req.PlanID))
		}
		conditions = append(conditions, targetTable.Source.Eq(req.Source))
		conditions = append(conditions, targetTable.TargetType.Eq(consts.TargetTypeScene))

		// 老的场景基本信息
		oldSceneInfo, err := targetTable.WithContext(ctx).Where(conditions...).First()
		if err != nil {
			return err
		}

		oldSceneName := oldSceneInfo.Name   // 老场景名称
		newSceneName := oldSceneName + "_1" // 新场景名称

		// 查询重名
		conditions2 := make([]gen.Condition, 0)
		conditions2 = append(conditions2, targetTable.TeamID.Eq(req.TeamID))
		if req.PlanID != "" {
			conditions2 = append(conditions2, targetTable.PlanID.Eq(req.PlanID))
		}
		conditions2 = append(conditions2, targetTable.Source.Eq(req.Source))
		conditions2 = append(conditions2, targetTable.TargetType.Eq(consts.TargetTypeScene))
		// 查询老配置相关的
		targetNameList, err := targetTable.WithContext(ctx).Where(conditions2...).Where(targetTable.Name.Like(fmt.Sprintf("%s%%", oldSceneName+"_"))).Find()
		if err == nil {
			// 有复制过得配置
			maxNum := 0
			for _, targetInfo := range targetNameList {
				nameTmp := targetInfo.Name
				postfixSlice := strings.Split(nameTmp, "_")
				if len(postfixSlice) < 2 {
					continue
				}
				currentNum, err := strconv.Atoi(postfixSlice[len(postfixSlice)-1])
				if err != nil {
					log.Logger.Info("复制自动化计划--类型转换失败，err:", err)
					continue
				}
				if currentNum > maxNum {
					maxNum = currentNum
				}
			}
			newSceneName = oldSceneName + fmt.Sprintf("_%d", maxNum+1)
		}

		nameLength := public.GetStringNum(newSceneName)
		if nameLength > 30 { // 场景名称限制30个字符
			return fmt.Errorf("名称过长！不可超出30字符")
		}

		// 组装新场景基本信息数据
		oldSceneInfo.ID = 0
		oldSceneInfo.TargetID = uuid.GetUUID()
		oldSceneInfo.Name = newSceneName
		oldSceneInfo.Sort = oldSceneInfo.Sort + 1
		oldSceneInfo.CreatedAt = time.Now()
		oldSceneInfo.UpdatedAt = time.Now()
		oldSceneInfo.CreatedUserID = userID
		oldSceneInfo.RecentUserID = userID

		err = targetTable.WithContext(ctx).Create(oldSceneInfo)
		if err != nil {
			return err
		}

		// 新的场景ID
		newSceneID := oldSceneInfo.TargetID

		// 获取场景变量
		collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectSceneParam)
		cur, err := collection.Find(ctx, bson.D{{"team_id", req.TeamID}, {"scene_id", req.SceneID}})
		var sceneParamDataArr []*mao.SceneParamData
		if err == nil {
			if err := cur.All(ctx, &sceneParamDataArr); err != nil {
				return fmt.Errorf("场景参数数据获取失败")
			}
			for _, sv := range sceneParamDataArr {
				sv.SceneID = newSceneID
				if _, err := collection.InsertOne(ctx, sv); err != nil {
					return err
				}
			}
		}

		// 3、克隆导入变量
		variableImportList, err := tx.VariableImport.WithContext(ctx).Where(tx.VariableImport.SceneID.Eq(req.SceneID)).Find()
		if err == nil {
			for _, variableImport := range variableImportList {
				variableImport.ID = 0
				variableImport.SceneID = newSceneID
				variableImport.CreatedAt = time.Now()
				variableImport.UpdatedAt = time.Now()
				if err := tx.VariableImport.WithContext(ctx).Create(variableImport); err != nil {
					return err
				}
			}
		}

		// 4、克隆流程
		flow := mao.Flow{}
		c1 := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectFlow)
		err = c1.FindOne(ctx, bson.D{{"scene_id", req.SceneID}}).Decode(&flow)
		if err == nil {
			flow.SceneID = newSceneID
			// 更新api的uuid
			err := packer.ChangeSceneNodeUUID(&flow)
			if err != nil {
				log.Logger.Info("克隆场景--替换event_id失败")
				return err
			}
			if _, err := c1.InsertOne(ctx, flow); err != nil {
				return err
			}
		}

		// 场景管理和自动化测试复制测试用例
		if req.Source == consts.TargetSourceScene || req.Source == consts.TargetSourceAutoPlan {
			// 7、克隆测试用例
			testCaseList, err := targetTable.WithContext(ctx).Where(targetTable.ParentID.Eq(req.SceneID), targetTable.TargetType.Eq(consts.TargetTypeTestCase)).Find()
			if err == nil && len(testCaseList) > 0 {
				for _, testCaseInfo := range testCaseList {

					oldCaseID := testCaseInfo.TargetID
					testCaseInfo.ID = 0
					testCaseInfo.TargetID = uuid.GetUUID()
					testCaseInfo.ParentID = newSceneID
					testCaseInfo.CreatedUserID = userID
					testCaseInfo.RecentUserID = userID
					testCaseInfo.CreatedAt = time.Now()
					testCaseInfo.UpdatedAt = time.Now()
					if err := targetTable.WithContext(ctx).Create(testCaseInfo); err != nil {
						return err
					}
					newCaseID := testCaseInfo.TargetID

					// 查询老的测试用例信息
					// 克隆用例的详情
					var sceneCaseFlow mao.SceneCaseFlow
					collection3 := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectSceneCaseFlow)
					err := collection3.FindOne(ctx, bson.D{{"scene_case_id", oldCaseID}}).Decode(&sceneCaseFlow)
					if err == nil {
						sceneCaseFlow.SceneCaseID = newCaseID
						sceneCaseFlow.SceneID = newSceneID
						// 更新testCase的uuid
						err = packer.ChangeCaseNodeUUID(&sceneCaseFlow)
						if err != nil {
							log.Logger.Info("复制计划--替换用例event_id失败")
							return err
						}
						if _, err := collection3.InsertOne(ctx, sceneCaseFlow); err != nil {
							return err
						}
					}

				}
			}

		}

		if req.PlanID != "" {
			if req.Source == consts.TargetSourcePlan { // 性能计划场景
				// 5、克隆任务配置
				// 查询老的任务对应配置
				oldTaskConfInfo, err := tx.StressPlanTaskConf.WithContext(ctx).Where(tx.StressPlanTaskConf.TeamID.Eq(req.TeamID),
					tx.StressPlanTaskConf.PlanID.Eq(req.PlanID), tx.StressPlanTaskConf.SceneID.Eq(req.SceneID)).First()
				if err == nil {
					oldTaskConfInfo.ID = 0
					oldTaskConfInfo.PlanID = req.PlanID
					oldTaskConfInfo.SceneID = newSceneID
					oldTaskConfInfo.RunUserID = userID
					oldTaskConfInfo.CreatedAt = time.Now()
					err := tx.StressPlanTaskConf.WithContext(ctx).Create(oldTaskConfInfo)
					if err != nil {
						return err
					}
				}

				// 6、克隆定时任务
				timedTaskInfo, err := tx.StressPlanTimedTaskConf.WithContext(ctx).Where(tx.StressPlanTimedTaskConf.PlanID.Eq(req.PlanID),
					tx.StressPlanTimedTaskConf.TeamID.Eq(req.TeamID), tx.StressPlanTimedTaskConf.SceneID.Eq(req.SceneID)).First()
				if err == nil {
					timedTaskInfo.ID = 0
					timedTaskInfo.SceneID = newSceneID
					timedTaskInfo.Status = consts.TimedTaskWaitEnable
					timedTaskInfo.RunUserID = userID
					timedTaskInfo.CreatedAt = time.Now()
					if err := tx.StressPlanTimedTaskConf.WithContext(ctx).Create(timedTaskInfo); err != nil {
						return err
					}
				}
			}
		}
		return nil
	})
	return err
}

func NotifyRunFinish(ctx *gin.Context, req *rao.NotifyRunFinishReq) error {
	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		log.Logger.Info("NotifyRunFinish--请求入参", *req)
		// 查询当前计划信息
		planTable := tx.AutoPlan
		planInfo, err := planTable.WithContext(ctx).Where(planTable.TeamID.Eq(req.TeamID),
			planTable.PlanID.Eq(req.PlanID)).First()
		if err != nil {
			return err
		}

		runDurationTime := int64(math.Ceil(float64(req.RunDurationTime) / 1000)) // 运行时长（秒）
		updatedAt := time.Now()
		// 修改报告状态
		ar := tx.AutoPlanReport
		_, err = ar.WithContext(ctx).Where(ar.TeamID.Eq(req.TeamID),
			ar.ReportID.Eq(req.ReportID)).UpdateSimple(ar.Status.Value(consts.ReportStatusFinish),
			ar.RunDurationTime.Value(runDurationTime), ar.UpdatedAt.Value(updatedAt))
		if err != nil {
			log.Logger.Info("NotifyRunFinish--修改报告状态失败")
			return err
		}

		if planInfo.TaskType == consts.PlanTaskTypeNormal {
			ap := tx.AutoPlan
			_, err = ap.WithContext(ctx).Where(ap.TeamID.Eq(req.TeamID), ap.PlanID.Eq(req.PlanID)).UpdateSimple(ap.Status.Value(consts.PlanStatusNormal))
			if err != nil {
				log.Logger.Info("NotifyRunFinish--更新自动化计划状态失败")
				return err
			}
		} else {
			// 判断当前计划下是否还有别的定时任务
			ttc := dal.GetQuery().AutoPlanTimedTaskConf
			TimedTaskConfInfo, err := ttc.WithContext(ctx).Where(ttc.TeamID.Eq(req.TeamID),
				ttc.PlanID.Eq(req.PlanID)).First()
			nowTime := time.Now().Unix()
			if err == nil {
				if TimedTaskConfInfo.Frequency == 0 || (TimedTaskConfInfo.Frequency != 0 && TimedTaskConfInfo.TaskCloseTime <= nowTime) {
					// 查到定时任务配置了,如果任务配置过期时间小于当前时间，则把计划状态改为未运行
					p := dal.GetQuery().AutoPlan
					_, err := p.WithContext(ctx).Where(p.TeamID.Eq(planInfo.TeamID),
						p.PlanID.Eq(planInfo.PlanID)).UpdateSimple(p.Status.Value(consts.PlanStatusNormal))
					if err != nil {
						log.Logger.Info("NotifyRunFinish--修改定时计划状态失败")
						response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
						return err
					}
				}
			}
		}

		// 把报告数据做好快照
		tempReportReq := &rao.GetAutoPlanReportDetailReq{
			TeamID:          req.TeamID,
			ReportID:        req.ReportID,
			RunDurationTime: runDurationTime,
			UpdatedAt:       updatedAt.Unix(),
		}
		_, err = MakeAutoPlanReportDetail(ctx, tempReportReq)
		log.Logger.Info("NotifyRunFinish--创建报告快照数据结果", err)
		if err != nil {
			log.Logger.Info("NotifyRunFinish--报告详情数据mongodb落库失败")
			return err
		}

		// 发送通知
		noticeGroupIDs := make([]string, 0)
		nge := dal.GetQuery().ThirdNoticeGroupEvent
		if err := nge.WithContext(ctx).Where(
			nge.TeamID.Eq(req.TeamID),
			nge.PlanID.Eq(req.PlanID),
			nge.EventID.Eq(consts.NoticeEventAuthPlan)).Pluck(nge.GroupID, &noticeGroupIDs); err != nil {
			log.Logger.Error("NotifyRunFinish--query noticeGroupIDs err:", err)
		}
		if len(noticeGroupIDs) > 0 {
			var reportIDs = make([]string, 0, 1)
			sendNoticeReq := &rao.SendNoticeParams{
				EventID:        consts.NoticeEventAuthPlan,
				TeamID:         req.TeamID,
				ReportIDs:      append(reportIDs, req.ReportID),
				NoticeGroupIDs: noticeGroupIDs,
			}
			params, err := notice.GetSendCardParamsByReq(ctx, sendNoticeReq)
			if err != nil {
				log.Logger.Error("NotifyRunFinish--GetSendCardParamsByReq err:", err)
			}
			for _, groupID := range noticeGroupIDs {
				if err := notice.SendNoticeByGroup(ctx, groupID, params); err != nil {
					log.Logger.Error("NotifyRunFinish--SendNoticeByGroup err:", err)
				}
			}
		}

		// 发邮件
		//rx := dal.GetQuery().AutoPlanReport
		//reportInfo, err := rx.WithContext(ctx).Where(rx.TeamID.Eq(req.TeamID), rx.PlanID.Eq(req.PlanID)).Order(rx.CreatedAt.Desc()).First()
		//if err != nil {
		//	return err
		//}
		//
		//autoPlanInfo, err := tx.AutoPlan.WithContext(ctx).Where(tx.AutoPlan.TeamID.Eq(req.TeamID), tx.AutoPlan.PlanID.Eq(req.PlanID)).First()
		//if err != nil {
		//	return err
		//}
		//
		//emails, err := tx.AutoPlanEmail.WithContext(ctx).Where(tx.AutoPlanEmail.TeamID.Eq(req.TeamID),
		//	tx.AutoPlanEmail.PlanID.Eq(req.PlanID)).Find()
		//if err == nil && len(emails) > 0 {
		//	ttx := dal.GetQuery().Team
		//	teamInfo, err := ttx.WithContext(ctx).Where(ttx.TeamID.Eq(req.TeamID)).First()
		//	if err != nil {
		//		return err
		//	}
		//
		//	ux := dal.GetQuery().User
		//	user, err := ux.WithContext(ctx).Where(ux.UserID.Eq(reportInfo.RunUserID)).First()
		//	if err != nil {
		//		return err
		//	}
		//
		//	for _, email := range emails {
		//		if err := mail.SendAutoPlanEmail(email.Email, autoPlanInfo, teamInfo, user.Nickname, reportInfo.ReportID); err != nil {
		//			log.Logger.Info("自动化计划回调--发送邮件失败")
		//			//return err
		//		}
		//	}
		//}

		return nil
	})
	log.Logger.Info("NotifyRunFinish--mysql操作结果", err)
	if err != nil {
		log.Logger.Info("NotifyRunFinish--操作失败")
		return err
	}

	return nil
}

func StopAutoPlan(ctx *gin.Context, req *rao.StopAutoPlanReq) error {
	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		ap := tx.AutoPlan
		planInfo, err := ap.WithContext(ctx).Where(ap.TeamID.Eq(req.TeamID), ap.PlanID.Eq(req.PlanID)).First()
		if err != nil {
			return err
		}

		_, err = ap.WithContext(ctx).Where(ap.TeamID.Eq(req.TeamID), ap.PlanID.Eq(req.PlanID)).UpdateSimple(ap.Status.Value(consts.PlanStatusNormal))
		if err != nil {
			return err
		}

		// 判断任务时定时任务还是普通任务
		if planInfo.TaskType == 2 { // 定时任务
			apttc := dal.GetQuery().AutoPlanTimedTaskConf
			_, err = apttc.WithContext(ctx).Where(apttc.TeamID.Eq(req.TeamID), apttc.PlanID.Eq(req.PlanID)).UpdateSimple(apttc.Status.Value(consts.TimedTaskWaitEnable))
			if err != nil {
				return err
			}
		}

		apr := tx.AutoPlanReport
		aprList, err := apr.WithContext(ctx).Where(apr.TeamID.Eq(req.TeamID), apr.PlanID.Eq(req.PlanID), apr.Status.Eq(consts.ReportStatusNormal)).Find()
		if err != nil {
			return err
		}

		_, err = apr.WithContext(ctx).Where(apr.TeamID.Eq(req.TeamID), apr.PlanID.Eq(req.PlanID)).UpdateSimple(apr.Status.Value(consts.ReportStatusFinish))
		if err != nil {
			return err
		}

		if len(aprList) > 0 {
			for _, aprInfo := range aprList {
				// 停止计划的时候，往redis里面写一条数据
				stopAutoPlanKey := consts.StopAutoPlanPrefix + req.TeamID + ":" + req.PlanID + ":" + aprInfo.ReportID
				_, err = dal.GetRDB().Set(ctx, stopAutoPlanKey, "stop", time.Second*3600).Result()
				if err != nil {
					log.Logger.Info("停止自动化计划--写入redis数据失败，err:", err)
					return err
				}
			}
		}
		return nil
	})
	return err
}

func GetAutoPlanReportDetail(ctx *gin.Context, req *rao.GetAutoPlanReportDetailReq) (*GetReportDetailResp, error) {
	// 查询当前团队是否解散或删除
	teamTable := dal.GetQuery().Team
	_, err := teamTable.WithContext(ctx).Where(teamTable.TeamID.Eq(req.TeamID)).First()
	if err != nil {
		err = fmt.Errorf("报告不存在")
		return nil, err
	}

	// 查询报告状态
	tx := dal.GetQuery().AutoPlanReport
	reportInfo, err := tx.WithContext(ctx).Where(tx.TeamID.Eq(req.TeamID), tx.ReportID.Eq(req.ReportID)).First()
	if err != nil {
		return nil, fmt.Errorf("报告不存在")
	}

	if reportInfo.Status == consts.ReportStatusNormal { // 进行中
		return nil, fmt.Errorf("计划正在运行中")
	}

	var reportDetailData mao.ReportDetailData
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectAutoReportDetailData)
	err = collection.FindOne(ctx, bson.D{{"team_id", req.TeamID}, {"report_id", req.ReportID}}).Decode(&reportDetailData)
	if err != nil {
		return nil, fmt.Errorf("计划正在运行中")
	}

	var res GetReportDetailResp
	err = bson.Unmarshal(reportDetailData.ReportDetailData, &res)
	if err != nil {
		return nil, err
	}
	res.ReportName = reportInfo.ReportName
	return &res, nil
}

type TestCaseResult struct {
	CaseName   string    `json:"case_name" bson:"case_name"`
	SucceedNum int64     `json:"succeed_num" bson:"succeed_num"`
	TotalNum   int64     `json:"total_num" bson:"total_num"`
	ApiList    []ApiList `json:"api_list" bson:"api_list"`
}
type ApiList struct {
	EventID        string         `json:"event_id" bson:"event_id"`
	TargetID       string         `json:"target_id" bson:"target_id"`
	CaseID         string         `json:"case_id" bson:"case_id"`
	ApiName        string         `json:"api_name" bson:"api_name"`
	Method         string         `json:"method" bson:"method"`
	Url            string         `json:"url" bson:"url"`
	Status         string         `json:"status" bson:"status"`
	ResponseBytes  float64        `json:"response_bytes" bson:"response_bytes"`
	RequestTime    int64          `json:"request_time" bson:"request_time"`
	RequestCode    int32          `json:"request_code" bson:"request_code"`
	RequestHeader  string         `json:"request_header" bson:"request_header"`
	RequestBody    string         `json:"request_body" bson:"request_body"`
	ResponseHeader string         `json:"response_header" bson:"response_header"`
	ResponseBody   string         `json:"response_body" bson:"response_body"`
	AssertionMsg   []AssertionMsg `json:"assert" bson:"assert"`
}

type AssertionMsg struct {
	Type      string `json:"type"`
	Code      int64  `json:"code" bson:"code"`
	IsSucceed bool   `json:"is_succeed" bson:"is_succeed"`
	Msg       string `json:"msg" bson:"msg"`
}

// SceneResult 场景结果
type SceneResult struct {
	SceneID      string `json:"scene_id" bson:"scene_id"`
	SceneName    string `json:"scene_name" bson:"scene_name"`
	CaseFailNum  int    `json:"case_fail_num" bson:"case_fail_num"`
	CaseTotalNum int    `json:"case_total_num" bson:"case_total_num"`
	State        int    `json:"state" bson:"state"` // 1-成功，2-失败
}

// GetReportDetailResp 获取报告详情返回值
type GetReportDetailResp struct {
	PlanName             string                      `json:"plan_name" bson:"plan_name"`
	ReportName           string                      `json:"report_name" bson:"report_name"`
	Avatar               string                      `json:"avatar" bson:"avatar"`
	Nickname             string                      `json:"nickname" bson:"nickname"`
	Remark               string                      `json:"remark" bson:"remark"`
	TaskMode             int32                       `json:"task_mode" bson:"task_mode"`
	SceneRunOrder        int32                       `json:"scene_run_order" bson:"scene_run_order"`
	TestCaseRunOrder     int32                       `json:"test_case_run_order" bson:"test_case_run_order"`
	ReportStatus         int32                       `json:"report_status" bson:"report_status"`
	ReportStartTime      int64                       `json:"report_start_time" bson:"report_start_time"`
	ReportEndTime        int64                       `json:"report_end_time" bson:"report_end_time"`
	ReportRunTime        int64                       `json:"report_run_time" bson:"report_run_time"`
	SceneBaseInfo        SceneBaseInfo               `json:"scene_base_info" bson:"scene_base_info"`
	CaseBaseInfo         CaseBaseInfo                `json:"case_base_info" bson:"case_base_info"`
	ApiBaseInfo          ApiBaseInfo                 `json:"api_base_info" bson:"api_base_info"`
	AssertionBaseInfo    AssertionBaseInfo           `json:"assertion_base_info" bson:"assertion_base_info"`
	SceneResult          []SceneResult               `json:"scene_result" bson:"scene_result"`
	SceneIDCaseResultMap map[string][]TestCaseResult `json:"scene_id_case_result_map" bson:"scene_id_case_result_map"`
}
type AssertionBaseInfo struct {
	AssertionTotalNum int64 `json:"assertion_total_num" bson:"assertion_total_num"`
	SucceedNum        int64 `json:"succeed_num" bson:"succeed_num"`
	FailNum           int64 `json:"fail_num" bson:"fail_num"`
}

type ApiBaseInfo struct {
	ApiTotalNum int64 `json:"api_total_num" bson:"api_total_num"`
	SucceedNum  int64 `json:"succeed_num" bson:"succeed_num"`
	FailNum     int64 `json:"fail_num" bson:"fail_num"`
	NotTestNum  int64 `json:"not_test_num" bson:"not_test_num"`
}

type CaseBaseInfo struct {
	CaseTotalNum int64 `json:"case_total_num" bson:"case_total_num"`
	SucceedNum   int64 `json:"succeed_num" bson:"succeed_num"`
	FailNum      int64 `json:"fail_num" bson:"fail_num"`
}

type SceneBaseInfo struct {
	SceneTotalNum int64 `json:"scene_total_num" bson:"scene_total_num"`
}

func MakeAutoPlanReportDetail(ctx context.Context, req *rao.GetAutoPlanReportDetailReq) (*GetReportDetailResp, error) {
	// 查询报告基本信息
	tx := dal.GetQuery()
	reportInfo, err := tx.AutoPlanReport.WithContext(ctx).Where(tx.AutoPlanReport.TeamID.Eq(req.TeamID),
		tx.AutoPlanReport.ReportID.Eq(req.ReportID)).First()
	if err != nil {
		return nil, err
	}

	// 获取执行计划用户信息
	userInfo, err := tx.User.WithContext(ctx).Where(tx.User.UserID.Eq(reportInfo.RunUserID)).First()
	if err != nil {
		return nil, err
	}

	// 查询当前计划下的所有场景信息
	sceneList, err := tx.Target.WithContext(ctx).Where(tx.Target.TeamID.Eq(req.TeamID),
		tx.Target.PlanID.Eq(reportInfo.PlanID), tx.Target.TargetType.Eq(consts.TargetTypeScene),
		tx.Target.Source.Eq(consts.TargetSourceAutoPlan), tx.Target.IsDisabled.Eq(consts.TargetIsDisabledNo)).Find()
	if err != nil {
		return nil, err
	}

	allSceneIDs := make([]string, 0, len(sceneList))
	for _, sceneInfo := range sceneList {
		allSceneIDs = append(allSceneIDs, sceneInfo.TargetID)
	}

	//查询当前计划下所有场景用例
	testCaseList, err := tx.Target.WithContext(ctx).Where(tx.Target.TeamID.Eq(req.TeamID),
		tx.Target.PlanID.Eq(reportInfo.PlanID), tx.Target.TargetType.Eq(consts.TargetTypeTestCase),
		tx.Target.Source.Eq(consts.TargetSourceAutoPlan), tx.Target.IsChecked.Eq(consts.TargetIsCheckedOpen),
		tx.Target.ParentID.In(allSceneIDs...)).Find()
	if err != nil {
		return nil, err
	}

	allCaseIDs := make([]string, 0, len(testCaseList)) // 所有用例ID
	sceneCaseMap := make(map[string][]string)          // 获取场景与用例的映射
	for _, testCaseInfo := range testCaseList {
		allCaseIDs = append(allCaseIDs, testCaseInfo.TargetID)

		sceneCaseMap[testCaseInfo.ParentID] = append(sceneCaseMap[testCaseInfo.ParentID], testCaseInfo.TargetID)
	}

	// 获取所有测试用例的flow
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectSceneCaseFlow)
	cur, err := collection.Find(ctx, bson.D{{"scene_case_id", bson.D{{"$in", allCaseIDs}}}})
	if err != nil {
		return nil, fmt.Errorf("测试用例flow为空")
	}
	var sceneCaseFlows []*mao.SceneCaseFlow
	if err := cur.All(ctx, &sceneCaseFlows); err != nil {
		return nil, fmt.Errorf("测试用例flow获取失败")
	}

	// 获取运行的结果数据
	// 获取所有用例运行结果
	collection = dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectAutoReport)
	cur, err = collection.Find(ctx, bson.D{{"case_id", bson.D{{"$in", allCaseIDs}}}, {"report_id", req.ReportID}, {"type", "api"}})
	if err != nil {
		return nil, fmt.Errorf("获取所有运行用例结果数据为空")
	}
	var sceneCaseReport []map[string]interface{}
	if err := cur.All(ctx, &sceneCaseReport); err != nil {
		return nil, fmt.Errorf("获取所有运行用例结果数据失败")
	}

	apiTotalNum := 0   // 接口总数
	apiSucceedNum := 0 // 接口成功数
	apiFailNum := 0    // 接口失败数
	apiNotRunNum := 0  // 接口未测数

	var assertionTotalNum int32 = 0
	var assertionFailNum int32 = 0

	// 失败的用例map
	failCaseMap := make(map[string]int)

	sceneFailMap := make(map[string]string) // [sceneID]caseID

	// 组装event_id与接口运行结果映射
	eventApiMap := make(map[string]map[string]interface{})

	caseApiTotalNumMap := make(map[string]int)   // 某个用例的接口总数
	caseApiSucceedNumMap := make(map[string]int) // 某个用例的接口成功数

	// event_id与对应用例id的映射
	eventIDCaseIDMap := make(map[string]string, len(sceneCaseReport))
	eventIDTargetIDMap := make(map[string]string, len(sceneCaseReport))

	for _, caseReportDetail := range sceneCaseReport {
		// 统计排除控制器
		apiName := caseReportDetail["type"].(string)
		if apiName != "api" { // 非接口不参与统计
			continue
		}

		apiTotalNum++ // 统计所有的接口数

		caseID := caseReportDetail["case_id"].(string)
		parentID := caseReportDetail["parent_id"].(string)
		eventID := caseReportDetail["event_id"].(string)
		targetID := caseReportDetail["api_id"].(string)

		eventIDCaseIDMap[eventID] = caseID

		// 获取target_id
		eventIDTargetIDMap[eventID] = targetID

		// 统计断言总数
		if assertionNum, ok := caseReportDetail["assert_num"]; ok {
			assertionTotalNum = assertionTotalNum + assertionNum.(int32)
		}

		// 统计失败断言总数
		if assertionFailedNum, ok := caseReportDetail["assert_failed_num"]; ok {
			assertionFailNum = assertionFailNum + assertionFailedNum.(int32)
		}

		if caseReportDetail["status"] == "success" {
			apiSucceedNum++
			caseApiSucceedNumMap[caseID]++
		}

		var assertionFailedNum int32 = 0
		if _, ok := caseReportDetail["assert_failed_num"]; ok {
			assertionFailedNum = caseReportDetail["assert_failed_num"].(int32)
		}
		if caseReportDetail["status"] == "failed" || assertionFailedNum > 0 {
			apiFailNum++
			failCaseMap[caseID]++
			sceneFailMap[parentID] = caseID
		}
		if caseReportDetail["status"] == "not_run" {
			apiNotRunNum++
		}

		// 组装event_id与接口运行结果映射
		eventApiMap[eventID] = caseReportDetail

		caseApiTotalNumMap[caseID]++
	}

	// 用例成功数
	caseSucceedNum := len(allCaseIDs) - len(failCaseMap)

	// 断言成功总数
	assertionSucceedNum := assertionTotalNum - assertionFailNum

	// 统计场景结果数据
	sceneResultSlice := make([]SceneResult, 0, len(sceneList))
	for _, sceneInfo := range sceneList {
		tempData := SceneResult{
			SceneID:      sceneInfo.TargetID,
			SceneName:    sceneInfo.Name,
			CaseTotalNum: len(sceneCaseMap[sceneInfo.TargetID]),
			State:        1,
		}

		if len(sceneFailMap) > 0 {
			caseFailNum := 0
			for sceneID := range sceneFailMap {
				if sceneID == sceneInfo.TargetID {
					caseFailNum++
				}
			}
			tempData.CaseFailNum = caseFailNum
			if caseFailNum > 0 {
				tempData.State = 2
			}
		}

		sceneResultSlice = append(sceneResultSlice, tempData)
	}

	// 用例id和所有接口的映射
	caseIDApiMap := make(map[string][]ApiList)
	for _, sceneCaseFlowInfo := range sceneCaseFlows {
		var Nodes *mao.SceneCaseFlowNode
		if err := bson.Unmarshal(sceneCaseFlowInfo.Nodes, &Nodes); err != nil {
			continue
		}

		for _, apiInfo := range Nodes.Nodes {
			if apiInfo.Type != "api" {
				continue
			}
			tempData := ApiList{
				EventID: apiInfo.ID,
			}

			tempData.ApiName = apiInfo.API.Name
			tempData.Method = apiInfo.API.Method

			if caseID, ok := eventIDCaseIDMap[apiInfo.ID]; ok {
				tempData.CaseID = caseID
			}

			if targetID, ok := eventIDTargetIDMap[apiInfo.ID]; ok {
				tempData.TargetID = targetID
			}

			if _, ok := eventApiMap[apiInfo.ID]; ok {
				if temp, ok := eventApiMap[apiInfo.ID]["request_url"]; ok {
					tempData.Url = temp.(string)
				}
				if temp, ok := eventApiMap[apiInfo.ID]["status"]; ok {
					tempData.Status = temp.(string)
				}
				if temp, ok := eventApiMap[apiInfo.ID]["response_bytes"]; ok {
					tempData.ResponseBytes = temp.(float64)
				}
				if temp, ok := eventApiMap[apiInfo.ID]["request_time"]; ok {
					tempData.RequestTime = temp.(int64)
				}
				if temp, ok := eventApiMap[apiInfo.ID]["request_code"]; ok {
					tempData.RequestCode = temp.(int32)
				}
				if temp, ok := eventApiMap[apiInfo.ID]["request_header"]; ok {
					tempData.RequestHeader = temp.(string)
				}
				if temp, ok := eventApiMap[apiInfo.ID]["request_body"]; ok {
					tempData.RequestBody = temp.(string)
				}
				if temp, ok := eventApiMap[apiInfo.ID]["response_header"]; ok {
					tempData.ResponseHeader = temp.(string)
				}
				if temp, ok := eventApiMap[apiInfo.ID]["response_body"]; ok {
					tempData.ResponseBody = temp.(string)
				}
				if temp, ok := eventApiMap[apiInfo.ID]["assert"]; ok {
					log.Logger.Info("日志快照--断言数据", temp)
					if temp != nil {
						for _, v := range temp.(primitive.A) {
							if str, ok := v.(map[string]interface{}); ok {
								tempAssertion := AssertionMsg{
									Type:      str["type"].(string),
									Code:      str["code"].(int64),
									IsSucceed: str["is_succeed"].(bool),
									Msg:       str["msg"].(string),
								}
								tempData.AssertionMsg = append(tempData.AssertionMsg, tempAssertion)
							}
						}
					}
				}
			}

			caseIDApiMap[sceneCaseFlowInfo.SceneCaseID] = append(caseIDApiMap[sceneCaseFlowInfo.SceneCaseID], tempData)
		}
	}

	// 组装场景ID与用例结果的映射
	sceneIDCaseResultMap := make(map[string][]TestCaseResult)
	for _, sceneInfo := range sceneList {
		for _, caseInfo := range testCaseList {
			if caseInfo.ParentID == sceneInfo.TargetID {
				tempData := TestCaseResult{
					CaseName:   caseInfo.Name,
					SucceedNum: int64(caseApiSucceedNumMap[caseInfo.TargetID]),
					TotalNum:   int64(caseApiTotalNumMap[caseInfo.TargetID]),
					ApiList:    caseIDApiMap[caseInfo.TargetID],
				}
				sceneIDCaseResultMap[sceneInfo.TargetID] = append(sceneIDCaseResultMap[sceneInfo.TargetID], tempData)
			}
		}
	}

	// 报告运行时长
	var reportRunTime int64 = 0
	if req.RunDurationTime != 0 {
		reportRunTime = req.RunDurationTime
	} else {
		reportRunTime = reportInfo.RunDurationTime
	}

	res := &GetReportDetailResp{
		PlanName:         reportInfo.PlanName,
		ReportName:       reportInfo.ReportName,
		Avatar:           userInfo.Avatar,
		Remark:           reportInfo.Remark,
		Nickname:         userInfo.Nickname,
		TaskMode:         reportInfo.TaskMode,
		SceneRunOrder:    reportInfo.SceneRunOrder,
		TestCaseRunOrder: reportInfo.TestCaseRunOrder,
		ReportStatus:     consts.ReportStatusFinish, // 快照的时候，默认给运行完的状态
		ReportStartTime:  reportInfo.CreatedAt.Unix(),
		ReportEndTime:    req.UpdatedAt,
		ReportRunTime:    reportRunTime,
		SceneBaseInfo: SceneBaseInfo{
			SceneTotalNum: int64(len(sceneList)),
		},
		CaseBaseInfo: CaseBaseInfo{
			CaseTotalNum: int64(len(testCaseList)),
			SucceedNum:   int64(caseSucceedNum),
			FailNum:      int64(len(failCaseMap)),
		},
		ApiBaseInfo: ApiBaseInfo{
			ApiTotalNum: int64(apiTotalNum),
			SucceedNum:  int64(apiSucceedNum),
			FailNum:     int64(apiFailNum),
			NotTestNum:  int64(apiNotRunNum),
		},
		AssertionBaseInfo: AssertionBaseInfo{
			AssertionTotalNum: int64(assertionTotalNum),
			SucceedNum:        int64(assertionSucceedNum),
			FailNum:           int64(assertionFailNum),
		},
		SceneResult:          sceneResultSlice,
		SceneIDCaseResultMap: sceneIDCaseResultMap,
	}
	// 获取所有测试用例的flow
	reportDetailData, err := bson.Marshal(res)
	if err != nil {
		return nil, fmt.Errorf("压缩报告详情数据失败")
	}
	insertData := TransReportDetailDataToMao(req.TeamID, req.ReportID, reportDetailData)
	collection = dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectAutoReportDetailData)
	_, err = collection.InsertOne(ctx, insertData)
	if err != nil {
		return nil, err
	}
	log.Logger.Info("日志快照--断言数据", res)
	return res, nil
}

func TransReportDetailDataToMao(teamID string, reportID string, reportDetailData []byte) *mao.ReportDetailData {
	return &mao.ReportDetailData{
		TeamID:           teamID,
		ReportID:         reportID,
		ReportDetailData: reportDetailData,
	}
}

func GetReportApiDetail(ctx *gin.Context, req *rao.GetReportApiDetailReq) (*rao.GetReportApiDetailResp, error) {
	var nodes mao.Node
	var flow mao.SceneCaseFlow
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectSceneCaseFlow)
	err := collection.FindOne(ctx, bson.D{{"scene_case_id", req.CaseID}}).Decode(&flow)
	if err != nil {
		return nil, err
	}
	if err = bson.Unmarshal(flow.Nodes, &nodes); err != nil {
		return nil, err
	}

	res := &rao.GetReportApiDetailResp{}

	apiDetail := rao.APIDetail{}

	for _, nodeInfo := range nodes.Nodes {
		if nodeInfo.ID == req.EventID {
			apiDetail = nodeInfo.API
		}
	}
	res.APIDetail = apiDetail
	return res, nil
}

func SendReportApi(req *rao.SendReportApiReq) (string, error) {
	req.ApiDetail.Request.PreUrl = req.ApiDetail.EnvInfo.PreUrl
	retID, err := runner.RunAPI(req.ApiDetail)
	if err != nil {
		return "", fmt.Errorf("调试接口返回非200状态")
	}
	return retID, err
}

func UpdateAutoPlanReportName(ctx *gin.Context, req *rao.UpdateAutoPlanReportNameReq) error {
	allErr := dal.GetQuery().Transaction(func(tx *query.Query) error {
		_, err := tx.AutoPlanReport.WithContext(ctx).Where(tx.AutoPlanReport.ReportID.Eq(req.ReportID)).UpdateSimple(tx.AutoPlanReport.ReportName.Value(req.ReportName))
		if err != nil {
			return err
		}
		return nil
	})
	return allErr
}

func GetNewestAutoPlanList(ctx *gin.Context, req *rao.GetNewestAutoPlanListReq) ([]rao.GetNewestAutoPlanListResp, error) {
	resData := make([]rao.GetNewestAutoPlanListResp, 0, req.Size)
	_ = dal.GetQuery().Transaction(func(tx *query.Query) error {
		// 查询数据库
		limit := req.Size
		offset := (req.Page - 1) * req.Size
		sort := make([]field.Expr, 0, 6)
		sort = append(sort, tx.AutoPlan.CreatedAt.Desc())
		conditions := make([]gen.Condition, 0)
		conditions = append(conditions, tx.AutoPlan.TeamID.Eq(req.TeamID))

		list, _, err := tx.AutoPlan.WithContext(ctx).Where(conditions...).Order(sort...).FindByPage(offset, limit)
		if err != nil {
			log.Logger.Info("自动化计划列表--获取列表失败，err:", err)
			return err
		}

		// 获取所有操作人id
		userIDs := make([]string, 0, len(list))
		for _, detail := range list {
			userIDs = append(userIDs, detail.CreateUserID)
		}

		userTable := dal.GetQuery().User
		userList, err := userTable.WithContext(ctx).Select(tx.User.UserID,
			tx.User.Nickname).Where(userTable.UserID.In(userIDs...)).Find()
		if err != nil {
			return err
		}
		// 用户id和名称映射
		userMap := make(map[string]*model.User, len(userList))
		for _, userInfo := range userList {
			userMap[userInfo.UserID] = userInfo
		}

		for _, detail := range list {
			detailTmp := rao.GetNewestAutoPlanListResp{
				PlanID:     detail.PlanID,
				TeamID:     detail.TeamID,
				PlanName:   detail.PlanName,
				PlanType:   "auto",
				Username:   userMap[detail.CreateUserID].Nickname,
				UserAvatar: userMap[detail.CreateUserID].Avatar,
				UpdatedAt:  detail.UpdatedAt.Unix(),
			}
			resData = append(resData, detailTmp)
		}
		return nil
	})

	return resData, nil
}
