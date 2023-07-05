package plan

import (
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/jwt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/log"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/uuid"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/packer"
	"github.com/gin-gonic/gin"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/mao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/query"
)

func ImportScene(ctx *gin.Context, req *rao.ImportSceneReq) ([]*model.Target, error) {
	userID := jwt.GetUserIDByCtx(ctx)
	targetList := make([]*model.Target, 0)
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectFlow)

	targetIDMap := make(map[string]int, len(req.TargetIDList))
	for _, tID := range req.TargetIDList {
		targetIDMap[tID] = 1
	}

	err := dal.GetQuery().Transaction(func(tx *query.Query) error {
		targets, err := tx.Target.WithContext(ctx).Where(tx.Target.TargetID.In(req.TargetIDList...)).Find()
		if err != nil {
			return err
		}

		if len(targets) == 0 {
			return fmt.Errorf("导入场景不能为空")
		}

		// 查询要导入的场景基本信息
		oldSceneIDs := make([]string, 0, len(targets))

		groupIDOldToNewMap := make(map[string]string, len(targets))
		for _, targetInfo := range targets {
			// 分组名称
			if targetInfo.TargetType == consts.TargetTypeFolder {
				newGroupID := uuid.GetUUID()
				groupIDOldToNewMap[targetInfo.TargetID] = newGroupID

				if targetInfo.ParentID == "0" {
					// 查询当前目录名称在当前目录下是否存在
					_, err = tx.Target.WithContext(ctx).Where(tx.Target.PlanID.Eq(req.PlanID),
						tx.Target.ParentID.Eq("0"), tx.Target.Source.Eq(req.Source), tx.Target.Name.Eq(targetInfo.Name)).First()
					if err == nil {
						return fmt.Errorf("计划内目录不可重名")
					}
				}
			}
			// 场景名称
			if targetInfo.TargetType == consts.TargetTypeScene {
				oldSceneIDs = append(oldSceneIDs, targetInfo.TargetID)

				if targetInfo.ParentID == "0" {
					// 查询当前场景名称在当前目录下是否存在
					_, err = tx.Target.WithContext(ctx).Where(tx.Target.PlanID.Eq(req.PlanID),
						tx.Target.ParentID.Eq("0"), tx.Target.Source.Eq(req.Source), tx.Target.Name.Eq(targetInfo.Name)).First()
					if err == nil {
						return fmt.Errorf("计划内场景不可重名")
					}
				}
			}
		}

		// 根据source 判断来源
		dataFrom := ""
		if req.Source == 2 {
			dataFrom = "plan"
		} else {
			dataFrom = "auto_plan"
		}

		// 先导入顶层分组
		for _, targetInfo := range targets {
			if targetInfo.TargetType == consts.TargetTypeFolder {
				oldGroupID := targetInfo.TargetID
				oldParentID := targetInfo.ParentID
				targetInfo.ID = 0
				targetInfo.TargetID = groupIDOldToNewMap[oldGroupID]
				targetInfo.PlanID = req.PlanID
				targetInfo.ParentID = ""
				targetInfo.CreatedUserID = userID
				targetInfo.RecentUserID = userID
				targetInfo.Source = req.Source
				targetInfo.SourceID = oldGroupID

				// 判断父级ID
				if _, ok := targetIDMap[oldParentID]; ok {
					targetInfo.ParentID = groupIDOldToNewMap[oldParentID]
				}

				if err := tx.Target.WithContext(ctx).Create(targetInfo); err != nil {
					return err
				}
			}
		}

		// 导入场景数据
		sceneIDMap := make(map[string]string, len(targets))
		for _, targetInfo := range targets {
			if targetInfo.TargetType == consts.TargetTypeScene {
				oldSceneID := targetInfo.TargetID
				oldParentID := targetInfo.ParentID
				targetInfo.ID = 0
				targetInfo.TargetID = uuid.GetUUID()
				targetInfo.PlanID = req.PlanID
				targetInfo.ParentID = ""
				targetInfo.CreatedUserID = userID
				targetInfo.RecentUserID = userID
				targetInfo.Source = req.Source
				targetInfo.SourceID = oldSceneID

				if _, ok := targetIDMap[oldParentID]; ok {
					targetInfo.ParentID = groupIDOldToNewMap[oldParentID]
				}

				if err := tx.Target.WithContext(ctx).Create(targetInfo); err != nil {
					return err
				}
				sceneIDMap[oldSceneID] = targetInfo.TargetID

				// 复制场景里面的flow
				flow := mao.Flow{}
				err = collection.FindOne(ctx, bson.D{{"scene_id", oldSceneID}}).Decode(&flow)
				if err != nil && err != mongo.ErrNoDocuments {
					return err
				}
				if err != mongo.ErrNoDocuments {
					// 修改node来源from字段
					ns := mao.Node{}
					if err := bson.Unmarshal(flow.Nodes, &ns); err != nil {
						return err
					}
					for k := range ns.Nodes {
						ns.Nodes[k].Data.From = dataFrom
					}
					nodes, err := bson.Marshal(ns)
					if err != nil {
						return err
					}
					flow.SceneID = targetInfo.TargetID
					flow.Nodes = nodes

					// 修改前置条件来源from字段
					prepositions := mao.Preposition{}
					if err := bson.Unmarshal(flow.Prepositions, &prepositions); err != nil {
						return err
					}
					for k := range prepositions.Prepositions {
						prepositions.Prepositions[k].Data.From = dataFrom
					}
					newPrepositions, err := bson.Marshal(prepositions)
					if err != nil {
						return err
					}

					flow.SceneID = targetInfo.TargetID
					flow.Nodes = nodes
					flow.Prepositions = newPrepositions

					// 更新api的uuid
					err = packer.ChangeSceneNodeUUID(&flow)
					if err != nil {
						log.Logger.Info("导入场景--替换event_id失败")
						return err
					}
					if _, err := collection.InsertOne(ctx, flow); err != nil {
						return err
					}
				}
				targetList = append(targetList, targetInfo)
			}
		}

		// 复制场景变量
		for _, sceneID := range oldSceneIDs {
			collection = dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectSceneParam)
			cur, err := collection.Find(ctx, bson.D{{"team_id", req.TeamID}, {"scene_id", sceneID}})
			var sceneParamDataArr []*mao.SceneParamData
			if err == nil {
				if err := cur.All(ctx, &sceneParamDataArr); err != nil {
					return fmt.Errorf("场景参数数据获取失败")
				}
				for _, sv := range sceneParamDataArr {
					sv.SceneID = sceneIDMap[sceneID]
					if _, err := collection.InsertOne(ctx, sv); err != nil {
						return err
					}
				}
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
