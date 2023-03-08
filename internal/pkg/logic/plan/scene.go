package plan

import (
	"context"
	"fmt"
	"kp-management/internal/pkg/biz/log"
	"kp-management/internal/pkg/biz/uuid"
	"kp-management/internal/pkg/dal/rao"
	"kp-management/internal/pkg/packer"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"kp-management/internal/pkg/biz/consts"
	"kp-management/internal/pkg/dal"
	"kp-management/internal/pkg/dal/mao"
	"kp-management/internal/pkg/dal/model"
	"kp-management/internal/pkg/dal/query"
)

func ImportScene(ctx context.Context, userID string, req *rao.ImportSceneReq) ([]*model.Target, error) {
	targetList := make([]*model.Target, 0)
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectFlow)

	err := dal.GetQuery().Transaction(func(tx *query.Query) error {
		targets, err := tx.Target.WithContext(ctx).Where(tx.Target.TargetID.In(req.TargetIDList...)).Find()
		if err != nil {
			return err
		}

		if len(targets) == 0 {
			return fmt.Errorf("导入场景不能为空")
		}

		// 查询要导入的场景基本信息
		groupNames := make([]string, 0, len(targets))
		oldGroupsIDs := make([]string, 0, len(targets))

		sceneNames := make([]string, 0, len(targets))
		oldSceneIDs := make([]string, 0, len(targets))
		for _, targetInfo := range targets {
			// 分组名称
			if targetInfo.TargetType == consts.TargetTypeGroup {
				groupNames = append(groupNames, targetInfo.Name)
				oldGroupsIDs = append(oldGroupsIDs, targetInfo.TargetID)
			}
			// 场景名称
			if targetInfo.TargetType == consts.TargetTypeScene {
				sceneNames = append(sceneNames, targetInfo.Name)
				oldSceneIDs = append(oldSceneIDs, targetInfo.TargetID)
			}

		}

		// 检测分组排重
		if len(groupNames) > 0 {
			isExistCount, _ := tx.Target.WithContext(ctx).Where(tx.Target.TeamID.Eq(req.TeamID),
				tx.Target.PlanID.Eq(req.PlanID), tx.Target.Source.Eq(req.Source),
				tx.Target.TargetType.Eq(consts.TargetTypeGroup),
				tx.Target.Name.In(groupNames...)).Count()
			if isExistCount > 0 {
				return fmt.Errorf("计划内分组不可重名")
			}
		}

		// 检查场景名称排重
		if len(sceneNames) > 0 {
			isExistCount, _ := tx.Target.WithContext(ctx).Where(tx.Target.TeamID.Eq(req.TeamID),
				tx.Target.PlanID.Eq(req.PlanID), tx.Target.Source.Eq(req.Source),
				tx.Target.TargetType.Eq(consts.TargetTypeScene),
				tx.Target.Name.In(sceneNames...)).Count()
			if isExistCount > 0 {
				return fmt.Errorf("计划内场景不可重名")
			}
		}
		// 根据source 判断来源
		dataFrom := ""
		if req.Source == 2 {
			dataFrom = "plan"
		} else {
			dataFrom = "auto_plan"
		}

		// 开始导入
		// 先导入分组
		groupIDMap := make(map[string]string, len(targets))
		for _, targetInfo := range targets {
			if targetInfo.TargetType == consts.TargetTypeGroup {
				oldGroupID := targetInfo.TargetID
				targetInfo.ID = 0
				targetInfo.TargetID = uuid.GetUUID()
				targetInfo.PlanID = req.PlanID
				targetInfo.CreatedUserID = userID
				targetInfo.RecentUserID = userID
				targetInfo.Source = req.Source
				targetInfo.SourceID = oldGroupID
				if err := tx.Target.WithContext(ctx).Create(targetInfo); err != nil {
					return err
				}
				groupIDMap[oldGroupID] = targetInfo.TargetID
			}
		}

		// 导入场景数据
		sceneIDMap := make(map[string]string, len(targets))
		for _, targetInfo := range targets {
			if targetInfo.TargetType == consts.TargetTypeScene {
				oldSceneID := targetInfo.TargetID
				targetInfo.ID = 0
				targetInfo.TargetID = uuid.GetUUID()
				targetInfo.PlanID = req.PlanID
				targetInfo.ParentID = groupIDMap[targetInfo.ParentID]
				targetInfo.CreatedUserID = userID
				targetInfo.RecentUserID = userID
				targetInfo.Source = req.Source
				targetInfo.SourceID = oldSceneID
				if err := tx.Target.WithContext(ctx).Create(targetInfo); err != nil {
					return err
				}
				sceneIDMap[oldSceneID] = targetInfo.TargetID

				// 复制场景里面的flow
				var flow mao.Flow
				err = collection.FindOne(ctx, bson.D{{"scene_id", oldSceneID}}).Decode(&flow)
				if err != nil && err != mongo.ErrNoDocuments {
					return err
				}
				if err != mongo.ErrNoDocuments {
					var ns *mao.Node
					if err := bson.Unmarshal(flow.Nodes, &ns); err != nil {
						return err
					}
					for _, n := range ns.Nodes {
						n.Data.From = dataFrom
					}
					nodes, err := bson.Marshal(ns)
					if err != nil {
						return err
					}

					flow.SceneID = targetInfo.TargetID
					flow.Nodes = nodes
					if _, err := collection.InsertOne(ctx, flow); err != nil {
						return err
					}
				}
				targetList = append(targetList, targetInfo)
			}
		}

		// 复制场景变量表
		variableList, err := tx.Variable.WithContext(ctx).Where(tx.Variable.SceneID.In(oldSceneIDs...)).Find()
		if err != nil {
			return err
		}
		var variables []*model.Variable
		for _, variable := range variableList {
			if newSceneID, ok := sceneIDMap[variable.SceneID]; ok {
				variable.ID = 0
				variable.SceneID = newSceneID
				variables = append(variables, variable)
			}
		}
		if len(variables) > 0 {
			if err := tx.Variable.WithContext(ctx).CreateInBatches(variables, 5); err != nil {
				return err
			}
		}

		// 复制导入变量表
		vi, err := tx.VariableImport.WithContext(ctx).Where(tx.VariableImport.SceneID.In(oldSceneIDs...)).Find()
		if err != nil {
			return err
		}
		var variablesImports []*model.VariableImport
		for _, variableImport := range vi {
			if newSceneID, ok := sceneIDMap[variableImport.SceneID]; ok {
				variableImport.ID = 0
				variableImport.SceneID = newSceneID
				variablesImports = append(variablesImports, variableImport)
			}
		}
		if len(variablesImports) > 0 {
			if err := tx.VariableImport.WithContext(ctx).CreateInBatches(variablesImports, 5); err != nil {
				return err
			}
		}

		// 把用例也导入进来
		caseList, err := tx.Target.WithContext(ctx).Where(tx.Target.TeamID.Eq(req.TeamID), tx.Target.ParentID.In(oldSceneIDs...),
			tx.Target.TargetType.Eq(consts.TargetTypeTestCase)).Find()

		if err != nil {
			return err
		}

		oldCastIds := make([]string, 0, len(caseList))
		oldNewCaseIDMap := make(map[string]string)
		if len(caseList) > 0 {
			for _, oldCaseInfo := range caseList {
				oldCastIds = append(oldCastIds, oldCaseInfo.TargetID)
				oldCaseID := oldCaseInfo.TargetID

				oldCaseInfo.ID = 0
				oldCaseInfo.TargetID = uuid.GetUUID()
				oldCaseInfo.ParentID = sceneIDMap[oldCaseInfo.ParentID]
				oldCaseInfo.IsChecked = 1
				oldCaseInfo.PlanID = req.PlanID
				oldCaseInfo.Source = req.Source
				oldCaseInfo.CreatedUserID = userID
				oldCaseInfo.CreatedAt = time.Now()
				oldCaseInfo.UpdatedAt = time.Now()
				oldCaseInfo.SourceID = oldCaseID
				err := tx.Target.WithContext(ctx).Create(oldCaseInfo)
				if err != nil {
					return err
				}
				oldNewCaseIDMap[oldCaseID] = oldCaseInfo.TargetID
			}

			// 获取所有测试用例的flow
			collection = dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectSceneCaseFlow)
			cur, err := collection.Find(ctx, bson.D{{"scene_case_id", bson.D{{"$in", oldCastIds}}}})
			if err == nil {
				var sceneCaseFlows []*mao.SceneCaseFlow
				if err := cur.All(ctx, &sceneCaseFlows); err != nil {
					return fmt.Errorf("测试用例flow获取失败")
				}

				// 复制flow
				for _, sceneCaseFlowInfo := range sceneCaseFlows {
					sceneCaseFlowInfo.SceneCaseID = oldNewCaseIDMap[sceneCaseFlowInfo.SceneCaseID]
					sceneCaseFlowInfo.SceneID = sceneIDMap[sceneCaseFlowInfo.SceneID]
					// 更新testCase的uuid
					err = packer.ChangeCaseNodeUUID(sceneCaseFlowInfo)
					if err != nil {
						log.Logger.Info("克隆场景--替换用例event_id失败")
						return err
					}
					_, err := collection.InsertOne(ctx, sceneCaseFlowInfo)
					if err != nil {
						return err
					}
				}

			}
		}
		return nil
	})

	return targetList, err
}
