package uiPlan

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/api/ui"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/log"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/uuid"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/clients"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/mao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/query"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/errmsg"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/uiScene"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/packer"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/public"
	"github.com/gin-gonic/gin"
	"github.com/go-omnibus/proof"
	"github.com/jinzhu/copier"
	"go.mongodb.org/mongo-driver/bson"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func Save(ctx *gin.Context, userID string, req *rao.UIPlanSaveReq) error {
	// 名称不能存在
	req.PlanID = uuid.GetUUID()
	us := query.Use(dal.DB()).UIPlan
	if _, err := us.WithContext(ctx).Where(
		us.TeamID.Eq(req.TeamID),
		us.Name.Eq(req.Name),
	).First(); err == nil {
		return errmsg.ErrUISceneNameRepeat
	}

	// 查询当前团队下最大的plan_id数
	info, err := us.WithContext(ctx).Where(us.TeamID.Eq(req.TeamID)).Order(us.RankID.Desc()).Limit(1).First()
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	var rankID int64 = 1
	if err == nil {
		rankID = info.RankID + 1
	}
	req.RankID = rankID

	// 随机获取一个机器 key
	if len(req.UIMachineKey) == 0 {
		machineKey, _ := clients.RandUiEngineMachineID()
		req.UIMachineKey = machineKey
	}
	planInfo := packer.TransSaveReqToUIPlanModel(req, userID)

	err = query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		if err = tx.UIPlan.WithContext(ctx).Create(planInfo); err != nil {
			return err
		}

		//if err = record.InsertCreate(ctx, req.TeamID, userID, record.OperationOperateCreateUISceneAPI, req.Name); err != nil {
		//	return err
		//}

		// 保存一份默认的配置
		insertTaskConf := &model.UIPlanTaskConf{
			PlanID:        req.PlanID,
			TeamID:        req.TeamID,
			TaskType:      consts.UIPlanTaskTypeNormal,
			SceneRunOrder: consts.UIPlanSceneRunModeOrder,
			RunUserID:     userID,
		}
		err = tx.UIPlanTaskConf.WithContext(ctx).Create(insertTaskConf)
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

func Update(ctx *gin.Context, userID string, req *rao.UIPlanSaveReq) error {
	// 名称不能存在
	us := query.Use(dal.DB()).UIPlan
	if _, err := us.WithContext(ctx).Where(
		us.TeamID.Eq(req.TeamID),
		us.Name.Eq(req.Name),
		us.PlanID.Neq(req.PlanID),
	).First(); err == nil {
		return errmsg.ErrUISceneNameRepeat
	}

	// 随机获取一个机器 key
	if len(req.UIMachineKey) == 0 {
		machineKey, _ := clients.RandUiEngineMachineID()
		req.UIMachineKey = machineKey
	}
	planInfo := packer.TransSaveReqToUIPlanModel(req, userID)
	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		// 查询当前接口是否存在
		_, err := tx.UIPlan.WithContext(ctx).Where(tx.UIPlan.PlanID.Eq(req.PlanID)).First()
		if err != nil {
			return err
		}

		if _, err = tx.UIPlan.WithContext(ctx).Where(tx.UIPlan.PlanID.Eq(req.PlanID)).Updates(planInfo); err != nil {
			return err
		}

		// 为空字段处理
		fields := make([]field.AssignExpr, 0)
		if req.Description != nil {
			fields = append(fields, tx.UIPlan.Description.Value(*req.Description))
		}
		if len(fields) > 0 {
			if _, err = tx.UIPlan.WithContext(ctx).Where(tx.UIPlan.PlanID.Eq(req.PlanID)).UpdateColumnSimple(fields...); err != nil {
				return err
			}
		}

		//if err = record.InsertUpdate(ctx, req.TeamID, userID, record.OperationOperateUpdateUISceneAPI, req.Name); err != nil {
		//	return err
		//}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func ListByTeamID(ctx *gin.Context, req *rao.ListUIPlanReq) ([]*rao.UIPlan, int64, error) {
	if req.Size == 0 {
		req.Size = 10
	}
	if req.Page == 0 {
		req.Page = 1
	}
	tx := query.Use(dal.DB()).UIPlan
	conditions := make([]gen.Condition, 0)
	conditions = append(conditions, tx.TeamID.Eq(req.TeamID))

	if len(req.Name) > 0 {
		conditions = append(conditions, tx.Name.Like(fmt.Sprintf("%%%s%%", req.Name)))
	}

	if len(req.CreatedUserName) > 0 {
		userIDs, _ := KeywordFindPlanForUserID(ctx, req.CreatedUserName)
		conditions = append(conditions, tx.CreateUserID.In(userIDs...))
	}

	if len(req.UpdatedTime) == 2 {
		layout := "2006-01-02 15:04:05"
		start, _ := time.Parse(layout, req.UpdatedTime[0])
		t, _ := time.Parse(layout, req.UpdatedTime[1])

		newTime := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, t.Location())
		end := t.Add(newTime.Sub(t))
		conditions = append(conditions, tx.UpdatedAt.Between(start, end))
	}

	if len(req.CreatedTime) == 2 {
		layout := "2006-01-02 15:04:05"
		start, _ := time.Parse(layout, req.CreatedTime[0])
		end, _ := time.Parse(layout, req.CreatedTime[1])
		conditions = append(conditions, tx.CreatedAt.Between(start, end))
	}

	if req.TaskType > 0 {
		conditions = append(conditions, tx.TaskType.Eq(req.TaskType))
	}

	if len(req.HeadUserName) > 0 {
		userIDs, _ := KeywordFindPlanForUserID(ctx, req.HeadUserName)
		list, err := tx.WithContext(ctx).Where(conditions...).Find()
		if err != nil {
			log.Logger.Error("req.HeadUserName err", proof.WithError(err))
		}
		planIDs := make([]string, 0)
		for _, plan := range list {
			for _, userID := range strings.Split(plan.HeadUserID, ",") {
				if public.ContainsStringSlice(userIDs, userID) {
					planIDs = append(planIDs, plan.PlanID)
				}
			}
		}
		conditions = append(conditions, tx.PlanID.In(planIDs...))
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
	timedPlanIDs := make([]string, 0, len(ret))
	for _, r := range ret {
		userIDs = append(userIDs, r.CreateUserID)
		userIDs = append(userIDs, strings.Split(r.HeadUserID, ",")...)
		if r.TaskType == consts.UIPlanTaskTypeCronjob { // 定时计划
			timedPlanIDs = append(timedPlanIDs, r.PlanID)
		}
	}

	u := query.Use(dal.DB()).User
	users, err := u.WithContext(ctx).Where(u.UserID.In(userIDs...)).Find()
	if err != nil {
		return nil, 0, err
	}

	timedConfTB := dal.GetQuery().UIPlanTimedTaskConf
	timedConfList, err := timedConfTB.WithContext(ctx).Where(timedConfTB.PlanID.In(timedPlanIDs...)).Find()
	if err != nil {
		return nil, 0, err
	}
	// 定时任务配置map
	taskMap := make(map[string]int32, len(timedConfList))
	for _, v := range timedConfList {
		taskMap[v.PlanID] = v.Status
	}

	return packer.TransUIPlanToRaoPlanList(ret, users, taskMap), cnt, nil
}

func KeywordFindPlanForUserID(ctx *gin.Context, keyword string) ([]string, error) {
	userIDs := make([]string, 0, 100)

	u := query.Use(dal.DB()).User
	err := u.WithContext(ctx).Where(u.Nickname.Like(fmt.Sprintf("%%%s%%", keyword))).Pluck(u.UserID, &userIDs)
	if err != nil {
		return nil, err
	}

	return userIDs, nil
}

func Detail(ctx *gin.Context, req *rao.UIPlanDetailReq) (*rao.UIPlan, error) {
	// 获取计划详情
	tx := dal.GetQuery().UIPlan
	planInfo, err := tx.WithContext(ctx).Where(tx.TeamID.Eq(req.TeamID)).Where(tx.PlanID.Eq(req.PlanID)).First()
	if err != nil {
		return nil, err
	}

	// 查询创建人
	var userIDs = make([]string, 0)
	userIDs = append(userIDs, planInfo.CreateUserID)
	headUserIDs := strings.Split(planInfo.HeadUserID, ",")
	userIDs = append(userIDs, headUserIDs...)

	tableUser := dal.GetQuery().User
	users, err := tableUser.WithContext(ctx).Where(tableUser.UserID.In(userIDs...)).Find()
	if err != nil {
		return nil, err
	}

	return packer.TransUIPlanToRaoPlan(planInfo, users), nil
}

func Delete(ctx *gin.Context, userID string, req *rao.UIPlanDeleteReq) error {
	err := dal.GetQuery().Transaction(func(tx *query.Query) error {
		for _, planID := range req.PlanIDs {
			// 删除计划基本信息
			if _, err := tx.UIPlan.WithContext(ctx).Where(tx.UIPlan.TeamID.Eq(req.TeamID)).Where(tx.UIPlan.PlanID.Eq(planID)).Delete(); err != nil {
				return err
			}

			// 查询计划下所创建的数据
			var sceneIDs = make([]string, 0)
			if err := tx.UIScene.WithContext(ctx).Where(
				tx.UIScene.TeamID.Eq(req.TeamID),
				tx.UIScene.PlanID.Eq(planID),
				tx.UIScene.Source.Eq(consts.UISceneSourcePlan),
			).Pluck(tx.UIScene.SceneID, &sceneIDs); err != nil {
				return err
			}

			if _, err := tx.UIScene.WithContext(ctx).Where(
				tx.UIScene.TeamID.Eq(req.TeamID),
				tx.UIScene.PlanID.Eq(planID),
				tx.UIScene.Source.Eq(consts.UISceneSourcePlan),
			).Delete(); err != nil {
				return err
			}

			// 删除步骤
			if _, err := tx.UISceneOperator.WithContext(ctx).Where(
				tx.UISceneOperator.SceneID.In(sceneIDs...),
			).Delete(); err != nil {
				return err
			}

			// 删除绑定关系
			if _, err := tx.UISceneSync.WithContext(ctx).Where(
				tx.UISceneSync.SceneID.In(sceneIDs...),
				tx.UISceneSync.TeamID.Eq(req.TeamID),
			).Delete(); err != nil {
				return err
			}

			//删除计划下的配置--普通任务
			if _, err := tx.UIPlanTaskConf.WithContext(ctx).Where(
				tx.UIPlanTaskConf.TeamID.Eq(req.TeamID),
				tx.UIPlanTaskConf.PlanID.Eq(planID)).Delete(); err != nil {
				return err
			}

			//删除计划下的配置--定时任务任务
			if _, err := tx.UIPlanTimedTaskConf.WithContext(ctx).Where(
				tx.UIPlanTimedTaskConf.TeamID.Eq(req.TeamID),
				tx.UIPlanTimedTaskConf.PlanID.Eq(planID)).Delete(); err != nil {
				return err
			}

		}
		return nil
	})

	return err
}

// Copy 复制
func Copy(ctx *gin.Context, userID string, req *rao.UIPlanCopyReq) error {
	newTime := time.Now()
	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		// step1 查询原来计划的信息
		oldPlanInfo, err := tx.UIPlan.WithContext(ctx).Where(
			tx.UIPlan.TeamID.Eq(req.TeamID),
			tx.UIPlan.PlanID.Eq(req.PlanID)).First()
		if err != nil {
			return err
		}

		oldPlanName := oldPlanInfo.Name
		newPlanName := oldPlanName + "_1"

		// 查询老配置相关的
		list, err := tx.UIPlan.WithContext(ctx).Where(
			tx.UIPlan.TeamID.Eq(req.TeamID),
			tx.UIPlan.Name.Like(fmt.Sprintf("%s%%", oldPlanName+"_"))).Find()
		if err == nil {
			// 有复制过得配置
			maxNum := 0
			for _, p := range list {
				nameTmp := p.Name
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
		autoPlanInfo, err := tx.UIPlan.WithContext(ctx).
			Where(tx.UIPlan.TeamID.Eq(req.TeamID)).
			Order(tx.UIPlan.RankID.Desc()).Limit(1).First()
		if err == nil { // 查到了
			rankID = autoPlanInfo.RankID + 1
		}

		// step2 插入计划
		newPlanID := uuid.GetUUID()

		oldPlanInfo.ID = 0
		oldPlanInfo.PlanID = newPlanID
		oldPlanInfo.RankID = rankID
		oldPlanInfo.Name = newPlanName
		oldPlanInfo.CreateUserID = userID
		oldPlanInfo.CreatedAt = newTime
		oldPlanInfo.UpdatedAt = newTime
		err = tx.UIPlan.WithContext(ctx).Create(oldPlanInfo)
		if err != nil {
			log.Logger.Info("复制计划--复制计划基本数据失败，err:", err)
			return err
		}

		// step3 插入计划配置任务
		if oldPlanInfo.TaskType == consts.UIPlanTaskTypeNormal { // 普通任务
			planTaskConfInfo, err := tx.UIPlanTaskConf.WithContext(ctx).Where(
				tx.UIPlanTaskConf.TeamID.Eq(req.TeamID),
				tx.UIPlanTaskConf.PlanID.Eq(req.PlanID)).First()
			if err == nil {
				planTaskConfInfo.ID = 0
				planTaskConfInfo.PlanID = newPlanID
				planTaskConfInfo.RunUserID = userID
				planTaskConfInfo.CreatedAt = newTime
				planTaskConfInfo.UpdatedAt = newTime
				if err := tx.UIPlanTaskConf.WithContext(ctx).Create(planTaskConfInfo); err != nil {
					return err
				}
			}
		} else { // 定时任务
			// 复制普通任务配置
			uiPlanTimedTaskConfInfo, err := tx.UIPlanTimedTaskConf.WithContext(ctx).Where(
				tx.UIPlanTimedTaskConf.TeamID.Eq(req.TeamID),
				tx.UIPlanTimedTaskConf.PlanID.Eq(req.PlanID)).First()
			if err == nil {
				uiPlanTimedTaskConfInfo.ID = 0
				uiPlanTimedTaskConfInfo.PlanID = newPlanID
				uiPlanTimedTaskConfInfo.Status = 0
				uiPlanTimedTaskConfInfo.RunUserID = userID
				uiPlanTimedTaskConfInfo.CreatedAt = newTime
				uiPlanTimedTaskConfInfo.UpdatedAt = newTime
				if err := tx.UIPlanTimedTaskConf.WithContext(ctx).Create(uiPlanTimedTaskConfInfo); err != nil {
					return err
				}
			}
		}

		// step4 复制计划下场景,分组
		oldTargetList, err := tx.UIScene.WithContext(ctx).Where(
			tx.UIScene.TeamID.Eq(req.TeamID),
			tx.UIScene.PlanID.Eq(req.PlanID),
			tx.UIScene.Source.Eq(consts.UISceneSourcePlan),
			tx.UIScene.Status.Eq(consts.UISceneStatusNormal),
			tx.UIScene.SceneType.In(consts.UISceneTypeFolder, consts.UISceneTypeScene),
		).Order(tx.UIScene.ParentID).Find()

		oldSceneIDs := make([]string, 0, len(oldTargetList))
		sceneIDOldNewMap := make(map[string]string)
		if err == nil {
			for _, oldTargetInfo := range oldTargetList {
				if oldTargetInfo.SceneType == consts.UISceneTypeScene {
					oldSceneIDs = append(oldSceneIDs, oldTargetInfo.SceneID)
				}

				// 新的sceneID
				newSceneID := uuid.GetUUID()

				oldTargetID := oldTargetInfo.SceneID
				oldTargetInfo.ID = 0
				oldTargetInfo.SceneID = newSceneID
				oldTargetInfo.ParentID = sceneIDOldNewMap[oldTargetInfo.ParentID]
				oldTargetInfo.PlanID = newPlanID
				oldTargetInfo.CreatedUserID = userID
				oldTargetInfo.RecentUserID = userID
				oldTargetInfo.CreatedAt = newTime
				oldTargetInfo.UpdatedAt = newTime
				if err := tx.UIScene.WithContext(ctx).Create(oldTargetInfo); err != nil {
					return err
				}

				sceneIDOldNewMap[oldTargetID] = newSceneID
			}
		}

		// step5 复制场景步骤
		for _, oldSceneID := range oldSceneIDs {
			operators, err := tx.UISceneOperator.WithContext(ctx).Where(
				tx.UISceneOperator.SceneID.Eq(oldSceneID)).
				Order(tx.UISceneOperator.ParentID).Find()
			if err != nil || len(operators) == 0 {
				continue
			}
			sourceOperatorIDs := make([]string, 0)
			for _, o := range operators {
				sourceOperatorIDs = append(sourceOperatorIDs, o.OperatorID)
			}

			// step4: 查询源场景步骤详细
			collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectUISceneOperator)
			cursor, err := collection.Find(ctx, bson.D{{"operator_id", bson.D{{"$in", sourceOperatorIDs}}}})
			if err != nil {
				log.Logger.Error("collection.Find err", proof.WithError(err))
				return err
			}
			var sceneOperators []*mao.SceneOperator
			if err = cursor.All(ctx, &sceneOperators); err != nil {
				log.Logger.Error("cursor.All err", proof.WithError(err))
				return err
			}

			// step5: 生成新步骤ID
			newOperatorIDs := make(map[string]string)
			newOperatorIDs["0"] = "0"
			for _, o := range operators {
				newOperatorIDs[o.OperatorID] = uuid.GetUUID()
			}

			// step6: 新步骤ID基本数据
			var uiSceneOperators = make([]*model.UISceneOperator, 0, len(newOperatorIDs))
			for _, o := range operators {
				uiSceneOperator := &model.UISceneOperator{
					OperatorID: newOperatorIDs[o.OperatorID],
					SceneID:    sceneIDOldNewMap[oldSceneID],
					Name:       o.Name,
					ParentID:   newOperatorIDs[o.ParentID],
					Sort:       o.Sort,
					Status:     o.Status,
					Type:       o.Type,
					Action:     o.Action,
				}
				uiSceneOperators = append(uiSceneOperators, uiSceneOperator)
			}

			// step7: 新步骤ID详细数据
			collectSceneOperators := make([]interface{}, 0, len(sceneOperators))
			for _, so := range sceneOperators {
				so.OperatorID = newOperatorIDs[so.OperatorID]
				so.SceneID = sceneIDOldNewMap[oldSceneID]
				collectSceneOperators = append(collectSceneOperators, so)
			}

			// step8: 开启事务：添加操作 MySQL、MongoDB
			if err := tx.UISceneOperator.WithContext(ctx).Create(uiSceneOperators...); err != nil {
				return err
			}

			if _, err := collection.InsertMany(ctx, collectSceneOperators); err != nil {
				return err
			}
		}

		// step6 复制同步关系
		oldSyncList, _ := tx.UISceneSync.WithContext(ctx).Where(
			tx.UISceneSync.TeamID.Eq(req.TeamID),
			tx.UISceneSync.SceneID.In(oldSceneIDs...)).Find()
		for _, sync := range oldSyncList {
			sync.ID = 0
			sync.SceneID = sceneIDOldNewMap[sync.SceneID]
			sync.CreatedAt = newTime
			sync.UpdatedAt = newTime
			if err := tx.UISceneSync.WithContext(ctx).Create(sync); err != nil {
				return err
			}
		}

		// step7 复制元素与场景关联关系
		for _, oldSceneID := range oldSceneIDs {
			oldElements := make([]string, 0)
			if err = tx.UISceneElement.WithContext(ctx).Where(
				tx.UISceneElement.TeamID.Eq(req.TeamID),
				tx.UISceneElement.SceneID.Eq(oldSceneID)).Pluck(tx.UISceneElement.ElementID, &oldElements); err != nil {
				return err
			}
			if len(oldElements) > 0 {
				oldElements = public.SliceUnique(oldElements)
				newElements := make([]*model.UISceneElement, 0, len(oldElements))
				for _, elementID := range oldElements {
					newElement := &model.UISceneElement{
						SceneID:   sceneIDOldNewMap[oldSceneID],
						ElementID: elementID,
						TeamID:    req.TeamID,
						Status:    consts.UISceneStatusNormal,
						CreatedAt: newTime,
						UpdatedAt: newTime,
					}
					newElements = append(newElements, newElement)
				}
				if err = tx.UISceneElement.WithContext(ctx).Create(newElements...); err != nil {
					return err
				}
			}
		}

		return nil
	})

	return err
}

// SaveTaskConf 保存计划配置
func SaveTaskConf(ctx *gin.Context, userID string, req *rao.UIPlanSaveTaskConfReq) error {
	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		if req.TaskType == consts.UIPlanTaskTypeNormal { // 普通任务
			// 删除定时任务配置
			_, err := tx.UIPlanTimedTaskConf.WithContext(ctx).Where(
				tx.UIPlanTimedTaskConf.TeamID.Eq(req.TeamID),
				tx.UIPlanTimedTaskConf.PlanID.Eq(req.PlanID)).Delete()
			if err != nil {
				return err
			}

			taskConfInfo, err := tx.UIPlanTaskConf.WithContext(ctx).Where(
				tx.UIPlanTaskConf.TeamID.Eq(req.TeamID),
				tx.UIPlanTaskConf.PlanID.Eq(req.PlanID)).First()
			if err != nil && err != gorm.ErrRecordNotFound {
				return err
			}

			if err == nil { // 查到已存在，则修改
				updateData := make(map[string]interface{}, 10)
				updateData["scene_run_order"] = req.SceneRunOrder
				updateData["run_user_id"] = userID
				_, err := tx.UIPlanTaskConf.WithContext(ctx).Where(
					tx.UIPlanTaskConf.ID.Eq(taskConfInfo.ID)).Updates(updateData)
				if err != nil {
					return err
				}
			} else { // 没查到则新增
				newData := &model.UIPlanTaskConf{
					PlanID:        req.PlanID,
					TeamID:        req.TeamID,
					TaskType:      req.TaskType,
					SceneRunOrder: req.SceneRunOrder,
					RunUserID:     userID,
				}
				err := tx.UIPlanTaskConf.WithContext(ctx).Create(newData)
				if err != nil {
					return err
				}
			}
		} else { // 定时任务
			// 删除普通任务配置
			_, err := tx.UIPlanTaskConf.WithContext(ctx).Where(
				tx.UIPlanTaskConf.TeamID.Eq(req.TeamID),
				tx.UIPlanTaskConf.PlanID.Eq(req.PlanID)).Delete()
			if err != nil {
				return err
			}

			timedTaskConfInfo, err := tx.UIPlanTimedTaskConf.WithContext(ctx).Where(
				tx.UIPlanTimedTaskConf.TeamID.Eq(req.TeamID),
				tx.UIPlanTimedTaskConf.PlanID.Eq(req.PlanID)).First()
			if err != nil && err != gorm.ErrRecordNotFound {
				return err
			}

			if err == nil { // 查到已存在，则修改
				_, err = tx.UIPlanTimedTaskConf.WithContext(ctx).Where(tx.UIPlanTimedTaskConf.ID.Eq(timedTaskConfInfo.ID)).
					UpdateSimple(
						tx.UIPlanTimedTaskConf.Frequency.Value(req.Frequency),
						tx.UIPlanTimedTaskConf.TaskExecTime.Value(req.TaskExecTime),
						tx.UIPlanTimedTaskConf.TaskCloseTime.Value(req.TaskCloseTime),
						tx.UIPlanTimedTaskConf.FixedIntervalStartTime.Value(req.FixedIntervalStartTime),
						tx.UIPlanTimedTaskConf.FixedIntervalTime.Value(req.FixedIntervalTime),
						tx.UIPlanTimedTaskConf.FixedRunNum.Value(req.FixedRunNum),
						tx.UIPlanTimedTaskConf.FixedIntervalTimeType.Value(req.FixedIntervalTimeType),
						tx.UIPlanTimedTaskConf.SceneRunOrder.Value(req.SceneRunOrder),
						tx.UIPlanTimedTaskConf.Status.Value(consts.UIPlanTimedTaskWaitEnable),
						tx.UIPlanTimedTaskConf.RunUserID.Value(userID),
					)
				if err != nil {
					return err
				}
			} else { // 没查到则新增
				newData := &model.UIPlanTimedTaskConf{
					PlanID:                 req.PlanID,
					TeamID:                 req.TeamID,
					TaskType:               req.TaskType,
					SceneRunOrder:          req.SceneRunOrder,
					Frequency:              req.Frequency,
					TaskExecTime:           req.TaskExecTime,
					TaskCloseTime:          req.TaskCloseTime,
					FixedIntervalStartTime: req.FixedIntervalStartTime,
					FixedIntervalTime:      req.FixedIntervalTime,
					FixedRunNum:            req.FixedRunNum,
					FixedIntervalTimeType:  req.FixedIntervalTimeType,
					Status:                 consts.UIPlanTimedTaskWaitEnable,
					RunUserID:              userID,
				}
				err := tx.UIPlanTimedTaskConf.WithContext(ctx).Create(newData)
				if err != nil {
					return err
				}
			}
		}

		// 修改计划类型
		_, err := tx.UIPlan.WithContext(ctx).Where(
			tx.UIPlan.TeamID.Eq(req.TeamID),
			tx.UIPlan.PlanID.Eq(req.PlanID)).UpdateSimple(tx.UIPlan.TaskType.Value(req.TaskType))
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
func GetTaskConf(ctx *gin.Context, req *rao.UIPlanGetTaskConfReq) (*rao.UIPlanGetTaskConfResp, error) {
	// 获取计划信息
	tx := dal.GetQuery().UIPlan
	planInfo, err := tx.WithContext(ctx).Where(tx.TeamID.Eq(req.TeamID), tx.PlanID.Eq(req.PlanID)).First()
	if err != nil {
		return nil, err
	}

	res := &rao.UIPlanGetTaskConfResp{}
	if planInfo.TaskType == consts.UIPlanTaskTypeNormal { // 普通任务
		tx := dal.GetQuery().UIPlanTaskConf
		taskConfInfo, err := tx.WithContext(ctx).Where(tx.TeamID.Eq(req.TeamID), tx.PlanID.Eq(req.PlanID)).First()
		if err != nil {
			return nil, err
		}
		res = &rao.UIPlanGetTaskConfResp{
			PlanID:        taskConfInfo.PlanID,
			TeamID:        taskConfInfo.TeamID,
			TaskType:      taskConfInfo.TaskType,
			SceneRunOrder: taskConfInfo.SceneRunOrder,
		}
	} else { // 定时任务
		tx := dal.GetQuery().UIPlanTimedTaskConf
		timedTaskConfInfo, err := tx.WithContext(ctx).Where(tx.TeamID.Eq(req.TeamID), tx.PlanID.Eq(req.PlanID)).First()
		if err != nil {
			return nil, err
		}
		res = &rao.UIPlanGetTaskConfResp{
			PlanID:                 timedTaskConfInfo.PlanID,
			TeamID:                 timedTaskConfInfo.TeamID,
			TaskType:               timedTaskConfInfo.TaskType,
			SceneRunOrder:          timedTaskConfInfo.SceneRunOrder,
			Frequency:              timedTaskConfInfo.Frequency,
			TaskExecTime:           timedTaskConfInfo.TaskExecTime,
			TaskCloseTime:          timedTaskConfInfo.TaskCloseTime,
			FixedIntervalStartTime: timedTaskConfInfo.FixedIntervalStartTime,
			FixedIntervalTime:      timedTaskConfInfo.FixedIntervalTime,
			FixedIntervalTimeType:  timedTaskConfInfo.FixedIntervalTimeType,
			FixedRunNum:            timedTaskConfInfo.FixedRunNum,
			Status:                 timedTaskConfInfo.Status,
		}

		if timedTaskConfInfo.Frequency == 0 { // 频次一次
			res.TaskCloseTime = 0
		}
	}
	return res, nil
}

// RunOrStartCron 执行或开启计划任务
func RunOrStartCron(ctx *gin.Context, userID string, req *rao.RunUIPlanReq) (string, error) {
	p := query.Use(dal.DB()).UIPlan
	plan, err := p.WithContext(ctx).Where(
		p.PlanID.Eq(req.PlanID),
		p.TeamID.Eq(req.TeamID),
	).First()
	if err != nil {
		return "", err
	}

	// 如果是定位任务，停止计划
	if plan.TaskType == consts.UIPlanTaskTypeCronjob {
		timedTaskConfTable := dal.GetQuery().UIPlanTimedTaskConf
		ttcInfo, err := timedTaskConfTable.WithContext(ctx).Where(timedTaskConfTable.PlanID.Eq(plan.PlanID)).First()
		if err == gorm.ErrRecordNotFound {
			return "", errmsg.ErrMustTaskInit
		}

		// 检查定时任务时间是否过期
		nowTime := time.Now().Unix()
		var taskCloseTime int64 = 0
		if ttcInfo.Frequency == consts.UIPlanFrequencyOnce {
			taskCloseTime = ttcInfo.TaskExecTime
		} else if ttcInfo.Frequency > consts.UIPlanFrequencyOnce && ttcInfo.Frequency < consts.UIPlanFrequencyFixedTime {
			taskCloseTime = ttcInfo.TaskCloseTime
		} else {
			taskCloseTime = ttcInfo.FixedIntervalStartTime
		}
		if taskCloseTime <= nowTime {
			return "", errmsg.ErrTimedTaskOverdue
		}

		// 修改定时任务状态
		autoPlanTimedTaskConfTable := dal.GetQuery().UIPlanTimedTaskConf
		_, err = autoPlanTimedTaskConfTable.WithContext(ctx).Where(
			autoPlanTimedTaskConfTable.TeamID.Eq(plan.TeamID),
			autoPlanTimedTaskConfTable.PlanID.Eq(plan.PlanID)).UpdateSimple(
			autoPlanTimedTaskConfTable.Status.Value(consts.UIPlanTimedTaskInExec),
			autoPlanTimedTaskConfTable.RunUserID.Value(userID))
		if err != nil {
			return "", err
		}

		return "", nil
	}

	return Run(ctx, userID, req)
}

func Run(ctx context.Context, userID string, req *rao.RunUIPlanReq) (string, error) {
	// step1: 查询计划下的场景
	// step2: 组装数据
	// step3: 生成报告
	p := query.Use(dal.DB()).UIPlan
	plan, err := p.WithContext(ctx).Where(
		p.PlanID.Eq(req.PlanID),
		p.TeamID.Eq(req.TeamID),
	).First()
	if err != nil {
		return "", err
	}

	// step1: 获取发送机器
	uiEngineList, err := clients.GetUiEngineMachineList()
	if err != nil {
		return "", errors.New("get ui_engine empty" + err.Error())
	}

	uiEngine := &rao.UiEngineMachineInfo{}
	for _, info := range uiEngineList {
		if info.Key == plan.UIMachineKey {
			uiEngine = info
		}
	}
	if len(uiEngine.IP) == 0 {
		uiEngine = uiEngineList[rand.Intn(len(uiEngineList))]
	}

	browsers := make([]*rao.Browser, 0)
	_ = json.Unmarshal([]byte(plan.Browsers), &browsers)
	if strings.Contains(strings.ToLower(uiEngine.SystemInfo.SystemBasic), "linux") {
		for _, b := range browsers {
			if !b.Headless {
				return "", errmsg.ErrSendLinuxNotQTMode
			}
		}
	}

	us := query.Use(dal.DB()).UIScene
	scenes, err := us.WithContext(ctx).Where(
		us.TeamID.Eq(req.TeamID),
		us.Status.Eq(consts.UISceneStatusNormal),
		us.SceneType.Eq(consts.UISceneTypeScene),
		us.Source.Eq(consts.UISceneSourcePlan),
		us.PlanID.Eq(req.PlanID),
	).Find()
	if err != nil {
		return "", err
	}

	runID := uuid.GetUUID()

	var (
		operatorTotalNum int
		assertTotalNum   int
	)
	sendOperatorIDs := make([]string, 0)
	uiScenes := make([]*ui.Scene, 0)
	for _, scene := range scenes {
		formatUIScene, uiSendScene, err := uiScene.FormatRunUiEngineByScene(ctx, runID, req.TeamID, scene.SceneID, sendOperatorIDs)
		if err != nil {
			log.Logger.Error("FormatRunUiEngineByScene err", err)
			return "", err
		}

		if formatUIScene == nil || uiSendScene == nil {
			return "", errmsg.ErrSendOperatorNotNull
		}
		uiScenes = append(uiScenes, formatUIScene)
		operatorTotalNum++
		assertTotalNum += uiSendScene.AssertTotalNum

		// step4: 生成 集合
		var docs []interface{}
		for _, run := range uiSendScene.Operators {
			docs = append(docs, run)
		}
		if len(docs) == 0 {
			return "", errmsg.ErrSendOperatorNotNull
		}
		if _, err := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectUISendSceneOperator).InsertMany(ctx, docs); err != nil {
			log.Logger.Error("调试日志入库失败", err)
			return "", err
		}
	}

	uiBrowsers := make([]*ui.Browser, 0)
	for _, browser := range browsers {
		b := &ui.Browser{Headless: false}
		if err = copier.Copy(b, browser); err != nil {
			log.Logger.Error("ui.Browser Copy err", proof.WithError(err))
		}

		uiBrowsers = append(uiBrowsers, b)
	}

	var sceneRunOrder int32
	if plan.TaskType == consts.UIPlanTaskTypeNormal {
		pt := query.Use(dal.DB()).UIPlanTaskConf
		planTaskConf, err := pt.WithContext(ctx).Where(
			pt.PlanID.Eq(req.PlanID),
			pt.TeamID.Eq(req.TeamID),
		).First()
		if err != nil {
			return "", err
		}
		sceneRunOrder = planTaskConf.SceneRunOrder
	}
	if plan.TaskType == consts.UIPlanTaskTypeCronjob {
		ptt := query.Use(dal.DB()).UIPlanTimedTaskConf
		planTimedTaskConf, err := ptt.WithContext(ctx).Where(
			ptt.PlanID.Eq(req.PlanID),
			ptt.TeamID.Eq(req.TeamID),
		).First()
		if err != nil {
			return "", err
		}
		sceneRunOrder = planTimedTaskConf.SceneRunOrder
	}

	runRequest := &ui.RunRequest{
		Topic:         runID,
		SceneRunOrder: sceneRunOrder,
		UserId:        userID,
		Browsers:      uiBrowsers,
		Scenes:        uiScenes,
	}
	if _, err = clients.RunUiEngine(ctx, uiEngine.IP, runRequest); err != nil {
		return "", err
	}

	// step5: 添加发送记录
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectUISendReport)
	detail, err := bson.Marshal(mao.SendReportDetail{Detail: runRequest})
	if err != nil {
		log.Logger.Error("mao.SendReportDetail bson marshal err", proof.WithError(err))
	}
	sendReport := &mao.SendReport{
		ReportID: runID,
		Detail:   detail,
	}
	if _, err = collection.InsertOne(ctx, sendReport); err != nil {
		return "", err
	}

	// step6: 添加自动化记录
	dal.GetRDB().SAdd(ctx, consts.UIEngineCurrentRunPrefix+uiEngine.IP, runID)
	dal.GetRDB().Set(ctx, consts.UIEngineRunAddrPrefix+runID, uiEngine.IP, time.Second*3600)

	log.Logger.Info("运行计划--创建报告", req.TeamID, req.PlanID)
	reportData := model.UIPlanReport{
		ReportID:        runID,
		ReportName:      plan.Name,
		PlanID:          plan.PlanID,
		PlanName:        plan.Name,
		TeamID:          plan.TeamID,
		TaskType:        plan.TaskType,
		SceneRunOrder:   sceneRunOrder,
		RunDurationTime: 0,
		Status:          consts.ReportStatusNormal,
		RunUserID:       userID,
		Remark:          "",
		Browsers:        plan.Browsers,
		UIMachineKey:    plan.UIMachineKey,
	}
	pr := query.Use(dal.DB()).UIPlanReport
	if err = pr.WithContext(ctx).Create(&reportData); err != nil {
		return "", err
	}

	return runID, nil
}
