package scene

import (
	"context"
	"kp-management/internal/pkg/biz/errno"
	"kp-management/internal/pkg/biz/record"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

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

func DeleteScene(ctx context.Context, req *rao.DeleteSceneReq, userID string) error {
	return dal.GetQuery().Transaction(func(tx *query.Query) error {
		targetInfo, err := tx.Target.WithContext(ctx).Where(tx.Target.TargetID.Eq(req.TargetID)).First()
		if err != nil {
			return err
		}

		if _, err := tx.Target.WithContext(ctx).Where(tx.Target.TargetID.Eq(req.TargetID)).Delete(); err != nil {
			return err
		}

		if _, err = tx.Target.WithContext(ctx).Where(tx.Target.ParentID.Eq(req.TargetID)).Delete(); err != nil {
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
				_, err = tx.StressPlanTaskConf.WithContext(ctx).Where(tx.StressPlanTaskConf.TeamID.Eq(req.TeamID),
					tx.StressPlanTaskConf.PlanID.Eq(req.PlanID), tx.StressPlanTaskConf.SceneID.Eq(req.TargetID)).Delete()
				if err != nil {
					return err
				}
			} else {
				_, err = tx.StressPlanTimedTaskConf.WithContext(ctx).Where(tx.StressPlanTimedTaskConf.TeamID.Eq(req.TeamID),
					tx.StressPlanTimedTaskConf.PlanID.Eq(req.PlanID), tx.StressPlanTimedTaskConf.SceneID.Eq(req.TargetID)).Delete()
				if err != nil {
					return err
				}
			}
		}

		// 从mg里面删除当前场景对应的flow
		collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectFlow)
		_, err = collection.DeleteMany(ctx, bson.D{{"scene_id", req.TargetID}})
		if err != nil {
			return err
		}

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
