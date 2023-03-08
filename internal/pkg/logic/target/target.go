package target

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gen"
	"kp-management/internal/pkg/biz/consts"
	"kp-management/internal/pkg/biz/jwt"
	"kp-management/internal/pkg/biz/record"
	"kp-management/internal/pkg/dal"
	"kp-management/internal/pkg/dal/mao"
	"kp-management/internal/pkg/dal/model"
	"kp-management/internal/pkg/dal/query"
	"kp-management/internal/pkg/dal/rao"
	"kp-management/internal/pkg/dal/runner"
	"kp-management/internal/pkg/packer"
)

func SendSceneAPI(ctx context.Context, teamID string, sceneID string, nodeID string, sceneCaseID string) (string, error) {
	var n mao.Node

	// 判断是场景还是用例
	if sceneCaseID != "" { // 用例
		var f mao.SceneCaseFlow
		collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectSceneCaseFlow)
		err := collection.FindOne(ctx, bson.D{{"scene_case_id", sceneCaseID}}).Decode(&f)
		if err != nil {
			return "", err
		}
		if err = bson.Unmarshal(f.Nodes, &n); err != nil {
			return "", err
		}
	} else { // 场景
		var f mao.Flow
		collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectFlow)
		err := collection.FindOne(ctx, bson.D{{"scene_id", sceneID}}).Decode(&f)
		if err != nil {
			return "", err
		}
		if err = bson.Unmarshal(f.Nodes, &n); err != nil {
			return "", err
		}
	}

	// 获取上传的变量文件地址
	vi := dal.GetQuery().VariableImport
	vis, err := vi.WithContext(ctx).Where(vi.TeamID.Eq(teamID), vi.SceneID.Eq(sceneID)).Find()
	if err != nil {
		return "", err
	}

	var fileList []rao.FileList
	for _, viInfo := range vis {
		fileList = append(fileList, rao.FileList{
			IsChecked: int64(viInfo.Status),
			Path:      viInfo.URL,
		})
	}

	for _, node := range n.Nodes {
		if node.ID == nodeID {

			tx := dal.GetQuery().Variable
			variables, err := tx.WithContext(ctx).Where(tx.TeamID.Eq(teamID),
				tx.Status.Eq(consts.VariableStatusOpen)).Order(tx.Type.Desc()).Find()
			if err != nil {
				return "", err
			}

			var vs []*rao.KVVariable
			for _, v := range variables {
				vs = append(vs, &rao.KVVariable{
					Key:   v.Var,
					Value: v.Val,
				})
			}

			node.API.Variable = vs
			if len(fileList) > 0 {
				node.API.Configuration.ParameterizedFile.Paths = fileList
			}
			return runner.RunAPI(ctx, node.API)
		}
	}

	return "", nil
}

func SendScene(ctx context.Context, teamID string, sceneID string, userID string) (string, error) {
	tx := dal.GetQuery().Target
	t, err := tx.WithContext(ctx).Where(tx.TargetID.Eq(sceneID), tx.TargetType.Eq(consts.TargetTypeScene)).First()
	if err != nil {
		return "", err
	}

	var f mao.Flow
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectFlow)
	err = collection.FindOne(ctx, bson.D{{"scene_id", sceneID}}).Decode(&f)
	if err != nil {
		return "", err
	}

	vi := dal.GetQuery().VariableImport
	vis, err := vi.WithContext(ctx).Where(vi.SceneID.Eq(sceneID), vi.Status.Eq(consts.VariableStatusOpen)).Limit(5).Find()
	if err != nil {
		return "", err
	}

	sv := dal.GetQuery().Variable
	sceneVariables, err := sv.WithContext(ctx).Where(sv.SceneID.Eq(sceneID), sv.Status.Eq(consts.VariableStatusOpen)).Find()
	if err != nil {
		return "", err
	}

	variables, err := sv.WithContext(ctx).Where(sv.TeamID.Eq(teamID),
		sv.Type.Eq(consts.VariableTypeGlobal),
		sv.Status.Eq(consts.VariableStatusOpen)).Find()
	if err != nil {
		return "", err
	}

	// 增加调试场景日志
	targetDebugLog := dal.GetQuery().TargetDebugLog
	insertData := &model.TargetDebugLog{
		TargetID:   sceneID,
		TargetType: consts.TargetDebugLogScene,
		TeamID:     teamID,
	}
	if err := targetDebugLog.WithContext(ctx).Create(insertData); err != nil {
		return "", err
	}

	if err := record.InsertDebug(ctx, teamID, userID, record.OperationOperateDebugScene, t.Name); err != nil {
		return "", err
	}

	req := packer.TransMaoFlowToRaoSceneFlow(t, &f, vis, sceneVariables, variables)
	return runner.RunScene(ctx, req)
}

func GetSendSceneResult(ctx context.Context, retID string) ([]*rao.SceneDebug, error) {
	cur, err := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectSceneDebug).
		Find(ctx, bson.D{{"uuid", retID}})
	if err != nil {
		return nil, err
	}

	var sds []*mao.SceneDebug
	if err := cur.All(ctx, &sds); err != nil {
		return nil, err
	}

	if len(sds) == 0 {
		return nil, nil
	}

	return packer.TransMaoSceneDebugsToRaoSceneDebugs(sds), nil

}

func SendAPI(ctx *gin.Context, teamID string, targetID string) (string, error) {
	tx := dal.GetQuery().Target
	t, err := tx.WithContext(ctx).Where(tx.TargetID.Eq(targetID)).First()
	if err != nil {
		return "", err
	}

	var apiInfo mao.API
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectAPI)
	err = collection.FindOne(ctx, bson.D{{"target_id", targetID}}).Decode(&apiInfo)
	if err != nil {
		return "", err
	}

	v := dal.GetQuery().Variable
	variables, err := v.WithContext(ctx).Where(v.TeamID.Eq(teamID), v.Status.Eq(consts.VariableStatusOpen)).Find()
	if err != nil {
		return "", err
	}

	// 把调试信息入库
	targetDebugLog := dal.GetQuery().TargetDebugLog
	insertData := &model.TargetDebugLog{
		TargetID:   targetID,
		TargetType: consts.TargetDebugLogApi,
		TeamID:     teamID,
	}
	err = targetDebugLog.WithContext(ctx).Create(insertData)
	if err != nil {
		return "", err
	}

	userID := jwt.GetUserIDByCtx(ctx)
	if err := record.InsertDebug(ctx, teamID, userID, record.OperationOperateDebugApi, t.Name); err != nil {
		return "", err
	}

	retID, err := runner.RunAPI(ctx, packer.TransTargetToRaoAPIDetail(t, &apiInfo, variables))
	if err != nil {
		return "", fmt.Errorf("调试接口返回非200状态")
	}

	return retID, err
}

func GetSendAPIResult(ctx context.Context, retID string) (*rao.APIDebug, error) {
	var ad mao.APIDebug
	err := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectAPIDebug).
		FindOne(ctx, bson.D{{"uuid", retID}}).Decode(&ad)
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, err
	}

	if err == mongo.ErrNoDocuments {
		return nil, nil
	}

	return packer.TransMaoAPIDebugToRaoAPIDebug(&ad), nil
}

func ListFolderAPI(ctx context.Context, teamID string) ([]*rao.FolderAPI, error) {
	tx := query.Use(dal.DB()).Target
	targets, err := tx.WithContext(ctx).Where(
		tx.TeamID.Eq(teamID),
		tx.TargetType.In(consts.TargetTypeFolder, consts.TargetTypeAPI),
		tx.Status.Eq(consts.TargetStatusNormal),
		tx.Source.Eq(consts.TargetSourceNormal)).Order(tx.Sort, tx.CreatedAt.Desc()).Find()

	if err != nil {
		return nil, err
	}

	return packer.TransTargetToRaoFolderAPIList(targets), nil
}

func SortTarget(ctx context.Context, req *rao.SortTargetReq) error {
	tx := dal.GetQuery().Target

	for _, target := range req.Targets {
		_, err := tx.WithContext(ctx).Where(tx.TeamID.Eq(target.TeamID), tx.TargetID.Eq(target.TargetID)).UpdateSimple(tx.Sort.Value(target.Sort), tx.ParentID.Value(target.ParentID))
		if err != nil {
			return err
		}
	}

	return nil
}

func ListGroupScene(ctx context.Context, req *rao.ListGroupSceneReq) ([]*rao.GroupScene, error) {
	tx := query.Use(dal.DB()).Target

	condition := make([]gen.Condition, 0)
	condition = append(condition, tx.TeamID.Eq(req.TeamID))
	condition = append(condition, tx.TargetType.In(consts.TargetTypeGroup, consts.TargetTypeScene))
	condition = append(condition, tx.Status.Eq(consts.TargetStatusNormal))
	condition = append(condition, tx.Source.Eq(req.Source))

	if req.Source == consts.TargetSourcePlan {
		condition = append(condition, tx.PlanID.Eq(req.PlanID))
	}
	if req.Source == consts.TargetSourceAutoPlan {
		condition = append(condition, tx.PlanID.Eq(req.PlanID))
		condition = append(condition, tx.TeamID.Eq(req.TeamID))
	}

	targets, err := tx.WithContext(ctx).Where(condition...).
		Order(tx.Sort.Desc(), tx.CreatedAt.Desc()).Find()

	if err != nil {
		return nil, err
	}

	return packer.TransTargetsToRaoGroupSceneList(targets), nil
}

func ListTrashFolderAPI(ctx context.Context, teamID string, limit, offset int) ([]*rao.FolderAPI, int64, error) {
	tx := query.Use(dal.DB()).Target
	targets, cnt, err := tx.WithContext(ctx).Where(
		tx.TeamID.Eq(teamID),
		tx.TargetType.In(consts.TargetTypeFolder, consts.TargetTypeAPI),
		tx.Status.Eq(consts.TargetStatusTrash),
	).Order(tx.Sort.Desc(), tx.CreatedAt.Desc()).FindByPage(offset, limit)

	if err != nil {
		return nil, 0, err
	}

	// 统计接口id
	apiIDs := make([]string, 0, len(targets))
	for _, tInfo := range targets {
		if tInfo.TargetType == "api" {
			apiIDs = append(apiIDs, tInfo.TargetID)
		}
	}

	// 获取api详情
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectAPI)
	cursor, err := collection.Find(ctx, bson.D{{"target_id", bson.D{{"$in", apiIDs}}}})

	if err != nil {
		return nil, 0, err
	}
	var apis []*mao.API
	if err = cursor.All(ctx, &apis); err != nil {
		return nil, 0, err
	}

	// 组装id与url映射
	apiIDUrlMap := make(map[string]string, len(apis))
	for _, apiInfo := range apis {
		apiIDUrlMap[apiInfo.TargetID] = apiInfo.URL
	}

	return packer.TransTargetToRaoTrashFolderAPIList(targets, apiIDUrlMap), cnt, nil
}

func Trash(ctx context.Context, targetID string, userID string) error {
	return dal.GetQuery().Transaction(func(tx *query.Query) error {
		t, err := tx.Target.WithContext(ctx).Where(tx.Target.TargetID.Eq(targetID)).First()
		if err != nil {
			return err
		}

		if _, err := tx.Target.WithContext(ctx).Where(tx.Target.TargetID.Eq(targetID)).UpdateColumn(tx.Target.Status, consts.TargetStatusTrash); err != nil {
			return err
		}

		if _, err = tx.Target.WithContext(ctx).Where(tx.Target.ParentID.Eq(targetID)).UpdateColumn(tx.Target.Status, consts.TargetStatusTrash); err != nil {
			return err
		}

		var operate int32 = 0
		if t.TargetType == consts.TargetTypeFolder {
			operate = record.OperationOperateDeleteFolder
		} else {
			operate = record.OperationOperateDeleteApi
		}
		if err := record.InsertDelete(ctx, t.TeamID, userID, operate, t.Name); err != nil {
			return err
		}
		return nil
	})
}

func Recall(ctx context.Context, targetID string) error {
	tx := query.Use(dal.DB()).Target
	targetInfo, err := tx.WithContext(ctx).Where(tx.TargetID.Eq(targetID)).First()
	if err != nil {
		return err
	}

	// 接口名排重
	_, err = tx.WithContext(ctx).Where(tx.TeamID.Eq(targetInfo.TeamID), tx.Name.Eq(targetInfo.Name),
		tx.TargetType.Eq(targetInfo.TargetType), tx.TargetID.Neq(targetInfo.TargetID),
		tx.Status.Eq(consts.TargetStatusNormal)).First()
	if err == nil {
		if targetInfo.TargetType == consts.TargetTypeFolder {
			return fmt.Errorf("文件夹名称已存在")
		} else {
			return fmt.Errorf("接口名称已存在")
		}
	}

	_, err = tx.WithContext(ctx).Where(tx.TargetID.Eq(targetID)).UpdateColumn(tx.Status, consts.TargetStatusNormal)
	if err != nil {
		return err
	}

	_, err = tx.WithContext(ctx).Where(tx.ParentID.Eq(targetID)).UpdateColumn(tx.Status, consts.TargetStatusNormal)
	if err != nil {
		return err
	}

	return nil
}

func Delete(ctx context.Context, targetID string) error {
	return query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		if _, err := tx.Target.WithContext(ctx).Where(tx.Target.TargetID.Eq(targetID)).Delete(); err != nil {
			return err
		}

		filter := bson.D{{"target_id", targetID}}

		//if _, err := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectFolder).DeleteOne(ctx, filter); err != nil {
		//	return err
		//}

		if _, err := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectAPI).DeleteOne(ctx, filter); err != nil {
			return err
		}

		return nil
	})
}

func APICountByTeamID(ctx context.Context, teamID string) (int64, error) {
	tx := query.Use(dal.DB()).Target

	return tx.WithContext(ctx).Where(
		tx.TargetType.Eq(consts.TargetTypeAPI),
		tx.TeamID.Eq(teamID),
		tx.Status.Eq(consts.TargetStatusNormal),
		tx.Source.Eq(consts.TargetSourceNormal),
	).Count()
}

func SceneCountByTeamID(ctx context.Context, teamID string) (int64, error) {
	tx := query.Use(dal.DB()).Target

	return tx.WithContext(ctx).Where(
		tx.TargetType.Eq(consts.TargetTypeScene),
		tx.Status.Eq(consts.TargetStatusNormal),
		tx.Source.Eq(consts.TargetSourceNormal),
		tx.TeamID.Eq(teamID),
	).Count()
}
