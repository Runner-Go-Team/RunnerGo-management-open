package scene

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"kp-management/internal/pkg/biz/errno"
	"kp-management/internal/pkg/biz/record"

	"kp-management/internal/pkg/biz/consts"
	"kp-management/internal/pkg/dal"
	"kp-management/internal/pkg/dal/mao"
	"kp-management/internal/pkg/dal/query"
	"kp-management/internal/pkg/dal/rao"
	"kp-management/internal/pkg/packer"
)

func SaveFlow(ctx context.Context, req *rao.SaveFlowReq) (int, error) {
	flow := packer.TransSaveFlowReqToMaoFlow(req)
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectFlow)

	err := collection.FindOne(ctx, bson.D{{"scene_id", req.SceneID}}).Err()
	if err == mongo.ErrNoDocuments { // 新建
		_, err := collection.InsertOne(ctx, flow)
		return errno.ErrMysqlFailed, err
	}

	_, err = collection.UpdateOne(ctx, bson.D{
		{"scene_id", flow.SceneID},
	}, bson.M{"$set": flow})

	return errno.Ok, err
}

func GetFlow(ctx context.Context, sceneID string) (*rao.GetFlowResp, error) {
	var ret mao.Flow

	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectFlow)
	err := collection.FindOne(ctx, bson.D{{"scene_id", sceneID}}).Decode(&ret)
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, err
	}

	if err == mongo.ErrNoDocuments {
		return nil, nil
	}

	return packer.TransMaoFlowToRaoGetFowResp(&ret), nil
}

func BatchGetFlow(ctx context.Context, sceneIDs []string) ([]*rao.Flow, error) {

	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectFlow)
	cursor, err := collection.Find(ctx, bson.D{{"scene_id", bson.D{{"$in", sceneIDs}}}})
	if err != nil {
		return nil, err
	}

	var flows []*mao.Flow
	if err := cursor.All(ctx, &flows); err != nil {
		return nil, err
	}

	return packer.TransMaoFlowsToRaoFlows(flows), nil
}

func DeleteScene(ctx *gin.Context, req *rao.DeleteSceneReq, userID string) error {
	return dal.GetQuery().Transaction(func(tx *query.Query) error {
		targetInfo, err := tx.Target.WithContext(ctx).Where(tx.Target.TargetID.Eq(req.TargetID)).First()
		if err != nil {
			return err
		}

		// 判断
		if targetInfo.TargetType == consts.TargetTypeGroup { // 分组目录
			targetList, err := tx.Target.WithContext(ctx).Where(tx.Target.ParentID.Eq(req.TargetID)).Find()
			if err != nil {
				return err
			}

			for _, targetData := range targetList {
				reqTemp := &rao.DeleteSceneReq{
					TargetID: targetData.TargetID,
					TeamID:   req.TeamID,
					PlanID:   req.PlanID,
					Source:   req.Source,
				}
				_ = DeleteScene(ctx, reqTemp, userID)
			}

			// 删除目录自己
			_, err = tx.Target.WithContext(ctx).Where(tx.Target.TargetID.Eq(req.TargetID)).Delete()
			if err != nil {
				return err
			}

		} else { // 场景
			_, err := tx.Target.WithContext(ctx).Where(tx.Target.TargetID.Eq(req.TargetID)).Delete()
			if err != nil {
				return err
			}

			// 从mg里面删除当前场景对应的flow
			collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectFlow)
			_, err = collection.DeleteMany(ctx, bson.D{{"scene_id", req.TargetID}})
			if err != nil {
				return err
			}

			// 查询场景下的用例
			caseList, err := tx.Target.WithContext(ctx).Where(tx.Target.ParentID.Eq(req.TargetID),
				tx.Target.TargetType.Eq(consts.TargetTypeTestCase)).Find()
			caseIDs := make([]string, 0, len(caseList))
			if err == nil && len(caseList) > 0 {
				for _, caseInfo := range caseList {
					caseIDs = append(caseIDs, caseInfo.TargetID)
				}
			}

			// 从mg里面删除当前场景对应的用例flow
			collection = dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectSceneCaseFlow)
			_, err = collection.DeleteMany(ctx, bson.D{{"scene_case_id", bson.D{{"$in", caseIDs}}}})
			if err != nil {
				return err
			}

			// 删除场景对应的场景变量
			if _, err = tx.Variable.WithContext(ctx).Where(tx.Variable.SceneID.Eq(req.TargetID)).Delete(); err != nil {
				return err
			}

			// 删除场景对应的场景变量文件
			if _, err = tx.VariableImport.WithContext(ctx).Where(tx.VariableImport.SceneID.Eq(req.TargetID)).Delete(); err != nil {
				return err
			}

			_, err = tx.Target.WithContext(ctx).Where(tx.Target.ParentID.Eq(req.TargetID)).Delete()
			if err != nil {
				return err
			}

			// 删除场景对应的任务配置
			if targetInfo.Source == consts.TargetSourcePlan { // 性能下的场景
				// 查询计划信息
				planInfo, err := tx.StressPlan.WithContext(ctx).Where(tx.StressPlan.TeamID.Eq(req.TeamID),
					tx.StressPlan.PlanID.Eq(req.PlanID)).First()
				if err != nil {
					return err
				}

				if planInfo.TaskType == consts.PlanTaskTypeNormal { // 普通任务
					_, err = tx.StressPlanTaskConf.WithContext(ctx).Where(tx.StressPlanTaskConf.SceneID.Eq(req.TargetID)).Delete()
					if err != nil {
						return err
					}
				} else {
					_, err = tx.StressPlanTimedTaskConf.WithContext(ctx).Where(tx.StressPlanTimedTaskConf.SceneID.Eq(req.TargetID)).Delete()
					if err != nil {
						return err
					}
				}
			}

			// 删除自动化场景对应的数据
			if targetInfo.Source == consts.TargetSourceAutoPlan { // 自动化下的场景
				// 查询计划信息
				planInfo, err := tx.AutoPlan.WithContext(ctx).Where(tx.AutoPlan.PlanID.Eq(req.PlanID)).First()
				if err != nil {
					return err
				}

				if planInfo.TaskType == consts.PlanTaskTypeNormal { // 普通任务
					_, err = tx.AutoPlanTaskConf.WithContext(ctx).Where(tx.AutoPlanTaskConf.PlanID.Eq(req.PlanID)).Delete()
					if err != nil {
						return err
					}
				} else {
					_, err = tx.AutoPlanTimedTaskConf.WithContext(ctx).Where(tx.AutoPlanTimedTaskConf.PlanID.Eq(req.PlanID)).Delete()
					if err != nil {
						return err
					}
				}
			}
		}

		// 记录操作日志
		var operate int32 = 0
		if targetInfo.TargetType == consts.TargetTypeScene {
			operate = record.OperationOperateDeleteScene
		} else if targetInfo.TargetType == consts.TargetTypeGroup {
			operate = record.OperationOperateDeleteGroup
		}
		if err := record.InsertDelete(ctx, targetInfo.TeamID, userID, operate, targetInfo.Name); err != nil {
			return err
		}

		return nil
	})
}
