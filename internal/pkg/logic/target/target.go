package target

import (
	"context"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/jwt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/log"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/record"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/mao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/query"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/runner"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/packer"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gen"
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

	// 获取全局变量
	globalVariable, err := GetGlobalVariable(ctx, teamID)

	// 获取场景变量
	sceneVariable := rao.GlobalVariable{}
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectSceneParam)
	cur, err := collection.Find(ctx, bson.D{{"team_id", teamID}, {"scene_id", sceneID}})
	var sceneParamDataArr []*mao.SceneParamData
	if err == nil {
		if err := cur.All(ctx, &sceneParamDataArr); err != nil {
			return "", fmt.Errorf("场景参数数据获取失败")
		}
	}

	sceneCookieParam := make([]rao.CookieParam, 0, 100)
	sceneHeaderParam := make([]rao.HeaderParam, 0, 100)
	sceneVariableParam := make([]rao.VariableParam, 0, 100)
	sceneAssertParam := make([]rao.AssertParam, 0, 100)
	for _, sceneParamInfo := range sceneParamDataArr {
		if sceneParamInfo.ParamType == 1 {
			err = json.Unmarshal([]byte(sceneParamInfo.DataDetail), &sceneCookieParam)
			if err != nil {
				return "", err
			}
			parameter := make([]rao.Parameter, 0, len(sceneCookieParam))
			for _, v := range sceneCookieParam {
				temp := rao.Parameter{
					IsChecked: v.IsChecked,
					Key:       v.Key,
					Value:     v.Value,
				}
				parameter = append(parameter, temp)
			}
			sceneVariable.Cookie.Parameter = parameter
		}
		if sceneParamInfo.ParamType == 2 {
			err = json.Unmarshal([]byte(sceneParamInfo.DataDetail), &sceneHeaderParam)
			if err != nil {
				return "", err
			}

			parameter := make([]rao.Parameter, 0, len(sceneHeaderParam))
			for _, v := range sceneHeaderParam {
				temp := rao.Parameter{
					IsChecked: v.IsChecked,
					Key:       v.Key,
					Value:     v.Value,
				}
				parameter = append(parameter, temp)
			}
			sceneVariable.Header.Parameter = parameter

		}
		if sceneParamInfo.ParamType == 3 {
			err = json.Unmarshal([]byte(sceneParamInfo.DataDetail), &sceneVariableParam)
			if err != nil {
				return "", err
			}

			parameter := make([]rao.VarForm, 0, len(sceneVariableParam))
			for _, v := range sceneVariableParam {
				temp := rao.VarForm{
					IsChecked: int64(v.IsChecked),
					Key:       v.Key,
					Value:     v.Value,
				}
				parameter = append(parameter, temp)
			}
			sceneVariable.Variable = parameter
		}
		if sceneParamInfo.ParamType == 4 {
			err = json.Unmarshal([]byte(sceneParamInfo.DataDetail), &sceneAssertParam)
			if err != nil {
				return "", err
			}

			parameter := make([]rao.AssertionText, 0, len(sceneAssertParam))
			for _, v := range sceneAssertParam {
				temp := rao.AssertionText{
					IsChecked:    int(v.IsChecked),
					ResponseType: int8(v.ResponseType),
					Compare:      v.Compare,
					Var:          v.Var,
					Val:          v.Val,
				}
				parameter = append(parameter, temp)
			}
			sceneVariable.Assert = parameter
		}
	}

	for _, node := range n.Nodes {
		if node.ID == nodeID {
			node.API.GlobalVariable = globalVariable
			node.API.Configuration.SceneVariable = sceneVariable
			if len(fileList) > 0 {
				node.API.Configuration.ParameterizedFile.Paths = fileList
			}
			node.API.Request.PreUrl = ""
			if node.API.EnvInfo.EnvID != 0 {
				node.API.Request.PreUrl = node.API.EnvInfo.PreUrl
			}
			if node.API.Request.Method == "" {
				node.API.Request.Method = node.API.Method
			}

			return runner.RunAPI(node.API)
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

	// 获取全局变量
	globalVariable, err := GetGlobalVariable(ctx, teamID)

	// 获取场景变量
	sceneVariable := rao.GlobalVariable{}
	collection = dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectSceneParam)
	cur, err := collection.Find(ctx, bson.D{{"team_id", teamID}, {"scene_id", sceneID}})
	var sceneParamDataArr []*mao.SceneParamData
	if err == nil {
		if err := cur.All(ctx, &sceneParamDataArr); err != nil {
			return "", fmt.Errorf("场景参数数据获取失败")
		}
	}

	sceneCookieParam := make([]rao.CookieParam, 0, 100)
	sceneHeaderParam := make([]rao.HeaderParam, 0, 100)
	sceneVariableParam := make([]rao.VariableParam, 0, 100)
	sceneAssertParam := make([]rao.AssertParam, 0, 100)
	for _, sceneParamInfo := range sceneParamDataArr {
		if sceneParamInfo.ParamType == 1 {
			err = json.Unmarshal([]byte(sceneParamInfo.DataDetail), &sceneCookieParam)
			if err != nil {
				return "", err
			}
			parameter := make([]rao.Parameter, 0, len(sceneCookieParam))
			for _, v := range sceneCookieParam {
				temp := rao.Parameter{
					IsChecked: v.IsChecked,
					Key:       v.Key,
					Value:     v.Value,
				}
				parameter = append(parameter, temp)
			}
			sceneVariable.Cookie.Parameter = parameter
		}
		if sceneParamInfo.ParamType == 2 {
			err = json.Unmarshal([]byte(sceneParamInfo.DataDetail), &sceneHeaderParam)
			if err != nil {
				return "", err
			}

			parameter := make([]rao.Parameter, 0, len(sceneHeaderParam))
			for _, v := range sceneHeaderParam {
				temp := rao.Parameter{
					IsChecked: v.IsChecked,
					Key:       v.Key,
					Value:     v.Value,
				}
				parameter = append(parameter, temp)
			}
			sceneVariable.Header.Parameter = parameter

		}
		if sceneParamInfo.ParamType == 3 {
			err = json.Unmarshal([]byte(sceneParamInfo.DataDetail), &sceneVariableParam)
			if err != nil {
				return "", err
			}

			parameter := make([]rao.VarForm, 0, len(sceneVariableParam))
			for _, v := range sceneVariableParam {
				temp := rao.VarForm{
					IsChecked: int64(v.IsChecked),
					Key:       v.Key,
					Value:     v.Value,
				}
				parameter = append(parameter, temp)
			}
			sceneVariable.Variable = parameter
		}
		if sceneParamInfo.ParamType == 4 {
			err = json.Unmarshal([]byte(sceneParamInfo.DataDetail), &sceneAssertParam)
			if err != nil {
				return "", err
			}

			parameter := make([]rao.AssertionText, 0, len(sceneAssertParam))
			for _, v := range sceneAssertParam {
				temp := rao.AssertionText{
					IsChecked:    int(v.IsChecked),
					ResponseType: int8(v.ResponseType),
					Compare:      v.Compare,
					Var:          v.Var,
					Val:          v.Val,
				}
				parameter = append(parameter, temp)
			}
			sceneVariable.Assert = parameter
		}
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

	runSceneParam := packer.TransMaoFlowToRaoSceneFlow(t, &f, vis, sceneVariable, globalVariable)
	return runner.RunScene(runSceneParam)
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
	targetInfo, err := tx.WithContext(ctx).Where(tx.TargetID.Eq(targetID)).First()
	if err != nil {
		return "", err
	}

	var apiInfo mao.API
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectAPI)
	err = collection.FindOne(ctx, bson.D{{"target_id", targetID}}).Decode(&apiInfo)
	if err != nil {
		return "", err
	}

	// 获取全局变量
	globalVariable, err := GetGlobalVariable(ctx, teamID)

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
	if err := record.InsertDebug(ctx, teamID, userID, record.OperationOperateDebugApi, targetInfo.Name); err != nil {
		return "", err
	}

	retID, err := runner.RunTarget(packer.GetRunTargetParam(targetInfo, globalVariable, &apiInfo))
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

func ListFolderAPI(ctx context.Context, req *rao.ListFolderAPIReq) ([]*rao.FolderAPI, error) {
	tx := query.Use(dal.DB()).Target
	targets, err := tx.WithContext(ctx).Where(
		tx.TeamID.Eq(req.TeamID),
		tx.TargetType.In(consts.TargetTypeFolder, consts.TargetTypeAPI,
			consts.TargetTypeSql, consts.TargetTypeTcp, consts.TargetTypeWebsocket,
			consts.TargetTypeDubbo),
		tx.Status.Eq(consts.TargetStatusNormal),
		tx.Source.Eq(req.Source)).Order(tx.Sort, tx.CreatedAt.Desc()).Find()

	if err != nil {
		return nil, err
	}

	return packer.TransTargetToRaoFolderAPIList(targets), nil
}

func SortTarget(ctx context.Context, req *rao.SortTargetReq) error {
	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		targetNameMap := make(map[string]int, len(req.Targets))
		for _, target := range req.Targets {
			if _, ok := targetNameMap[target.Name]; ok {
				return fmt.Errorf("存在重名，无法操作")
			} else {
				targetNameMap[target.Name] = 1
			}
		}

		for _, target := range req.Targets {
			_, err := tx.Target.WithContext(ctx).Where(tx.Target.TeamID.Eq(target.TeamID),
				tx.Target.TargetID.Eq(target.TargetID)).UpdateSimple(tx.Target.Sort.Value(target.Sort), tx.Target.ParentID.Value(target.ParentID))
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func ListGroupScene(ctx context.Context, req *rao.ListGroupSceneReq) ([]*rao.GroupScene, error) {
	tx := query.Use(dal.DB()).Target

	condition := make([]gen.Condition, 0)
	condition = append(condition, tx.TeamID.Eq(req.TeamID))
	condition = append(condition, tx.TargetType.In(consts.TargetTypeFolder, consts.TargetTypeScene))
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

func Trash(ctx *gin.Context, targetID string, userID string) error {
	return dal.GetQuery().Transaction(func(tx *query.Query) error {
		t, err := tx.Target.WithContext(ctx).Where(tx.Target.TargetID.Eq(targetID)).First()
		if err != nil {
			return err
		}

		// 删除
		_ = getAllSonTargetID(ctx, targetID, t.TargetType)

		var operate int32 = 0
		if t.TargetType == consts.TargetTypeFolder {
			operate = record.OperationOperateDeleteFolder
		} else if t.TargetType == consts.TargetTypeSql {
			operate = record.OperationLogDeleteSql
		} else if t.TargetType == consts.TargetTypeTcp {
			operate = record.OperationLogDeleteTcp
		} else if t.TargetType == consts.TargetTypeWebsocket {
			operate = record.OperationLogDeleteWebsocket
		} else if t.TargetType == consts.TargetTypeMQTT {
			operate = record.OperationLogDeleteMqtt
		} else if t.TargetType == consts.TargetTypeDubbo {
			operate = record.OperationLogDeleteDubbo
		} else {
			operate = record.OperationOperateDeleteApi
		}
		if err := record.InsertDelete(ctx, t.TeamID, userID, operate, t.Name); err != nil {
			return err
		}
		return nil
	})
}

func getAllSonTargetID(ctx *gin.Context, targetID string, targetType string) error {
	tx := dal.GetQuery().Target
	if targetType == consts.TargetTypeFolder {
		// 查询这个目录下是否还有别的目录或文件
		targetList, err := tx.WithContext(ctx).Where(tx.ParentID.Eq(targetID)).Find()
		if err != nil {
			return err
		}
		if len(targetList) > 0 {
			for _, tInfo := range targetList {
				_ = getAllSonTargetID(ctx, tInfo.TargetID, tInfo.TargetType)
			}
		}
		// 删除目录本身
		_, err = tx.WithContext(ctx).Where(tx.TargetID.Eq(targetID)).UpdateSimple(tx.Status.Value(consts.TargetStatusTrash))
		if err != nil {
			return err
		}
	} else {
		_, err := tx.WithContext(ctx).Where(tx.TargetID.Eq(targetID)).UpdateSimple(tx.Status.Value(consts.TargetStatusTrash))
		if err != nil {
			return err
		}
	}
	return nil
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

		if _, err := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectAPI).DeleteOne(ctx, filter); err != nil {
			return err
		}

		return nil
	})
}

func SendSql(ctx *gin.Context, req *rao.SendSqlReq) (string, error) {
	tx := dal.GetQuery().Target
	targetInfo, err := tx.WithContext(ctx).Where(tx.TargetID.Eq(req.TargetID)).First()
	if err != nil {
		return "", err
	}

	// 获取全局变量
	globalVariable, _ := GetGlobalVariable(ctx, req.TeamID)

	retID := ""
	runSqlParam := rao.RunTargetParam{}

	sqlDetailInfo := mao.SqlDetailForMg{}
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectSqlDetail)
	err = collection.FindOne(ctx, bson.D{{"target_id", req.TargetID}}).Decode(&sqlDetailInfo)
	if err != nil {
		return retID, err
	}
	runSqlParam = packer.TransRunSqlParam(targetInfo, &sqlDetailInfo, globalVariable)

	// 把调试信息入库
	targetDebugLog := dal.GetQuery().TargetDebugLog
	insertData := &model.TargetDebugLog{
		TargetID:   req.TargetID,
		TargetType: consts.TargetDebugLogApi,
		TeamID:     req.TeamID,
	}
	_ = targetDebugLog.WithContext(ctx).Create(insertData)

	userID := jwt.GetUserIDByCtx(ctx)
	if err := record.InsertDebug(ctx, req.TeamID, userID, record.OperationOperateDebugApi, targetInfo.Name); err != nil {
		return retID, err
	}

	retID, err = runner.RunTarget(runSqlParam)
	if err != nil {
		return retID, fmt.Errorf("调试返回非200状态")
	}
	return retID, err
}

// GetGlobalVariable 获取全局变量
func GetGlobalVariable(ctx context.Context, teamID string) (rao.GlobalVariable, error) {
	// 获取全局变量
	globalVariable := rao.GlobalVariable{}
	// 查询全局变量
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectGlobalParam)
	cur, err := collection.Find(ctx, bson.D{{"team_id", teamID}})
	var globalParamDataArr []*mao.GlobalParamData
	if err == nil {
		if err := cur.All(ctx, &globalParamDataArr); err != nil {
			return rao.GlobalVariable{}, fmt.Errorf("全局参数数据获取失败")
		}
	}

	cookieParam := make([]rao.CookieParam, 0, 100)
	headerParam := make([]rao.HeaderParam, 0, 100)
	variableParam := make([]rao.VariableParam, 0, 100)
	assertParam := make([]rao.AssertParam, 0, 100)
	for _, globalParamInfo := range globalParamDataArr {
		if globalParamInfo.ParamType == 1 {
			err = json.Unmarshal([]byte(globalParamInfo.DataDetail), &cookieParam)
			if err != nil {
				return rao.GlobalVariable{}, err
			}
			parameter := make([]rao.Parameter, 0, len(cookieParam))
			for _, v := range cookieParam {
				temp := rao.Parameter{
					IsChecked: v.IsChecked,
					Key:       v.Key,
					Value:     v.Value,
				}
				parameter = append(parameter, temp)
			}
			globalVariable.Cookie.Parameter = parameter
		}
		if globalParamInfo.ParamType == 2 {
			err = json.Unmarshal([]byte(globalParamInfo.DataDetail), &headerParam)
			if err != nil {
				return rao.GlobalVariable{}, err
			}

			parameter := make([]rao.Parameter, 0, len(headerParam))
			for _, v := range headerParam {
				temp := rao.Parameter{
					IsChecked: v.IsChecked,
					Key:       v.Key,
					Value:     v.Value,
				}
				parameter = append(parameter, temp)
			}
			globalVariable.Header.Parameter = parameter

		}
		if globalParamInfo.ParamType == 3 {
			err = json.Unmarshal([]byte(globalParamInfo.DataDetail), &variableParam)
			if err != nil {
				return rao.GlobalVariable{}, err
			}

			parameter := make([]rao.VarForm, 0, len(variableParam))
			for _, v := range variableParam {
				temp := rao.VarForm{
					IsChecked: int64(v.IsChecked),
					Key:       v.Key,
					Value:     v.Value,
				}
				parameter = append(parameter, temp)
			}
			globalVariable.Variable = parameter

		}
		if globalParamInfo.ParamType == 4 {
			err = json.Unmarshal([]byte(globalParamInfo.DataDetail), &assertParam)
			if err != nil {
				return rao.GlobalVariable{}, err
			}

			parameter := make([]rao.AssertionText, 0, len(assertParam))
			for _, v := range assertParam {
				temp := rao.AssertionText{
					IsChecked:    int(v.IsChecked),
					ResponseType: int8(v.ResponseType),
					Compare:      v.Compare,
					Var:          v.Var,
					Val:          v.Val,
				}
				parameter = append(parameter, temp)
			}
			globalVariable.Assert = parameter
		}
	}
	return globalVariable, err
}

func ConnectionDatabase(req *rao.ConnectionDatabaseReq) (bool, error) {
	res, err := runner.RunConnectionDatabase(*req)
	if err != nil {
		log.Logger.Info("连接数据库失败，err:", err)
		return res, err
	}
	return res, nil
}

func GetSendSqlResult(ctx *gin.Context, req *rao.GetSendSqlResultReq) (*rao.GetSendSqlResultResp, error, string) {
	sqlDebug := mao.SqlDebug{}
	err := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectAPIDebug).
		FindOne(ctx, bson.D{{"uuid", req.RetID}}).Decode(&sqlDebug)
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, err, ""
	}

	if err == mongo.ErrNoDocuments {
		return nil, nil, ""
	}

	res := &rao.GetSendSqlResultResp{
		RetID:        sqlDebug.Uuid,
		RequestTime:  sqlDebug.RequestTime,
		Status:       sqlDebug.Status,
		Regex:        sqlDebug.Regex,
		Assert:       sqlDebug.Assert,
		ResponseBody: sqlDebug.ResponseBody,
	}
	return res, nil, ""
}

func SendTcp(ctx *gin.Context, req *rao.SendTcpReq) (string, error) {
	tx := dal.GetQuery().Target
	targetInfo, err := tx.WithContext(ctx).Where(tx.TargetID.Eq(req.TargetID)).First()
	if err != nil {
		return "", err
	}

	tcpDetailInfo := mao.TcpDetail{}
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectTcpDetail)
	err = collection.FindOne(ctx, bson.D{{"target_id", req.TargetID}}).Decode(&tcpDetailInfo)
	if err != nil {
		return "", err
	}

	// 获取全局变量
	globalVariable, err := GetGlobalVariable(ctx, req.TeamID)

	// 把调试信息入库
	targetDebugLog := dal.GetQuery().TargetDebugLog
	insertData := &model.TargetDebugLog{
		TargetID:   req.TargetID,
		TargetType: consts.TargetDebugLogApi,
		TeamID:     req.TeamID,
	}
	err = targetDebugLog.WithContext(ctx).Create(insertData)
	if err != nil {
		return "", err
	}

	userID := jwt.GetUserIDByCtx(ctx)
	if err := record.InsertDebug(ctx, req.TeamID, userID, record.OperationOperateDebugApi, targetInfo.Name); err != nil {
		return "", err
	}

	retID, err := runner.RunTarget(packer.GetSendTcpParam(targetInfo, &tcpDetailInfo, globalVariable))
	if err != nil {
		return "", fmt.Errorf("调试TCP返回非200状态")
	}
	return retID, err
}

func GetSendTcpResult(ctx *gin.Context, req *rao.GetSendTcpResultReq) ([]rao.GetSendTcpResultResp, error) {
	cur, err := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectAPIDebug).
		Find(ctx, bson.D{{"uuid", req.RetID}})

	tcpDebug := make([]mao.TcpDebug, 0, 10)
	if err == nil {
		if err = cur.All(ctx, &tcpDebug); err != nil {
			return nil, fmt.Errorf("tcp结果解析失败")
		}
	}

	res := make([]rao.GetSendTcpResultResp, 0, len(tcpDebug))
	for _, v := range tcpDebug {
		temp := rao.GetSendTcpResultResp{
			TargetID:     v.TargetID,
			TeamID:       v.TeamID,
			Uuid:         v.Uuid,
			Name:         v.Name,
			IsStop:       v.IsStop,
			Type:         v.Type,
			RequestBody:  v.RequestBody,
			ResponseBody: v.ResponseBody,
			Status:       v.Status,
		}
		res = append(res, temp)
	}

	return res, nil
}

func SendWebsocket(ctx *gin.Context, req *rao.SendWebsocketReq) (string, error) {
	tx := dal.GetQuery().Target
	targetInfo, err := tx.WithContext(ctx).Where(tx.TargetID.Eq(req.TargetID)).First()
	if err != nil {
		return "", err
	}

	wsDetailInfo := mao.WebsocketDetail{}
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectWebsocketDetail)
	err = collection.FindOne(ctx, bson.D{{"target_id", req.TargetID}}).Decode(&wsDetailInfo)
	if err != nil {
		return "", err
	}

	// 获取全局变量
	globalVariable, err := GetGlobalVariable(ctx, req.TeamID)

	// 把调试信息入库
	targetDebugLog := dal.GetQuery().TargetDebugLog
	insertData := &model.TargetDebugLog{
		TargetID:   req.TargetID,
		TargetType: consts.TargetDebugLogApi,
		TeamID:     req.TeamID,
	}
	err = targetDebugLog.WithContext(ctx).Create(insertData)
	if err != nil {
		return "", err
	}

	userID := jwt.GetUserIDByCtx(ctx)
	if err := record.InsertDebug(ctx, req.TeamID, userID, record.OperationOperateDebugApi, targetInfo.Name); err != nil {
		return "", err
	}

	retID, err := runner.RunTarget(packer.GetSendWebsocketParam(targetInfo, &wsDetailInfo, globalVariable))
	if err != nil {
		return "", fmt.Errorf("调试TCP返回非200状态")
	}
	return retID, err
}

func GetSendWebsocketResult(ctx *gin.Context, req *rao.GetSendWebsocketResultReq) ([]rao.GetSendWebsocketResultResp, error) {
	cur, err := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectAPIDebug).
		Find(ctx, bson.D{{"uuid", req.RetID}})

	websocketDebug := make([]mao.WebsocketDebug, 0, 10)
	res := make([]rao.GetSendWebsocketResultResp, 0, len(websocketDebug))

	if err == nil {
		if err = cur.All(ctx, &websocketDebug); err != nil {
			return nil, fmt.Errorf("websocket结果解析失败")
		}
	}

	if len(websocketDebug) > 0 {
		for _, v := range websocketDebug {
			temp := rao.GetSendWebsocketResultResp{
				TargetID:            v.TargetID,
				TeamID:              v.TeamID,
				Uuid:                v.Uuid,
				Name:                v.Name,
				IsStop:              v.IsStop,
				Type:                v.Type,
				Status:              v.Status,
				RequestBody:         v.RequestBody,
				ResponseBody:        v.ResponseBody,
				ResponseMessageType: v.ResponseMessageType,
			}
			res = append(res, temp)
		}
	}

	return res, nil
}

func SendDubbo(ctx *gin.Context, req *rao.SendDubboReq) (string, error) {
	tx := dal.GetQuery().Target
	targetInfo, err := tx.WithContext(ctx).Where(tx.TargetID.Eq(req.TargetID)).First()
	if err != nil {
		return "", err
	}

	dubboDetailInfo := mao.DubboDetail{}
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectDubboDetail)
	err = collection.FindOne(ctx, bson.D{{"target_id", req.TargetID}}).Decode(&dubboDetailInfo)
	if err != nil {
		return "", err
	}

	// 获取全局变量
	globalVariable, err := GetGlobalVariable(ctx, req.TeamID)

	// 把调试信息入库
	targetDebugLog := dal.GetQuery().TargetDebugLog
	insertData := &model.TargetDebugLog{
		TargetID:   req.TargetID,
		TargetType: consts.TargetDebugLogApi,
		TeamID:     req.TeamID,
	}
	err = targetDebugLog.WithContext(ctx).Create(insertData)
	if err != nil {
		return "", err
	}

	userID := jwt.GetUserIDByCtx(ctx)
	if err := record.InsertDebug(ctx, req.TeamID, userID, record.OperationOperateDebugApi, targetInfo.Name); err != nil {
		return "", err
	}

	retID, err := runner.RunTarget(packer.GetSendDubboParam(targetInfo, &dubboDetailInfo, globalVariable))
	if err != nil {
		return "", fmt.Errorf("调试dubbo返回非200状态")
	}
	return retID, err
}

func GetSendDubboResult(ctx *gin.Context, req *rao.GetSendDubboResultReq) (rao.GetSendDubboResultResp, error) {
	dubboDebug := mao.DubboDebug{}
	err := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectAPIDebug).
		FindOne(ctx, bson.D{{"uuid", req.RetID}}).Decode(&dubboDebug)
	if err != nil && err != mongo.ErrNoDocuments {
		return rao.GetSendDubboResultResp{}, err
	}

	assertions := make([]rao.DebugAssert, 0, len(dubboDebug.Assert))
	for _, a := range dubboDebug.Assert {
		assertions = append(assertions, rao.DebugAssert{
			Code:      a.Code,
			IsSucceed: a.IsSucceed,
			Msg:       a.Msg,
		})
	}

	regexs := make([]map[string]interface{}, len(dubboDebug.Regex))
	for _, r := range dubboDebug.Regex {
		regexs = append(regexs, r)
	}

	res := rao.GetSendDubboResultResp{
		TargetID:     dubboDebug.TargetID,
		TeamID:       dubboDebug.TeamID,
		Uuid:         dubboDebug.Uuid,
		Name:         dubboDebug.Name,
		RequestType:  dubboDebug.RequestType,
		RequestBody:  dubboDebug.RequestBody,
		ResponseBody: dubboDebug.ResponseBody,
		Assert:       assertions,
		Regex:        regexs,
		Status:       dubboDebug.Status,
	}

	return res, nil
}

func SendMqtt(ctx *gin.Context, req *rao.SendMqttReq) (string, error) {
	tx := dal.GetQuery().Target
	targetInfo, err := tx.WithContext(ctx).Where(tx.TargetID.Eq(req.TargetID)).First()
	if err != nil {
		return "", err
	}

	mqttDetailInfo := mao.MqttDetail{}
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectMqttDetail)
	err = collection.FindOne(ctx, bson.D{{"target_id", req.TargetID}}).Decode(&mqttDetailInfo)
	if err != nil {
		return "", err
	}

	// 获取全局变量
	globalVariable, err := GetGlobalVariable(ctx, req.TeamID)

	// 把调试信息入库
	targetDebugLog := dal.GetQuery().TargetDebugLog
	insertData := &model.TargetDebugLog{
		TargetID:   req.TargetID,
		TargetType: consts.TargetDebugLogApi,
		TeamID:     req.TeamID,
	}
	err = targetDebugLog.WithContext(ctx).Create(insertData)
	if err != nil {
		return "", err
	}

	userID := jwt.GetUserIDByCtx(ctx)
	if err := record.InsertDebug(ctx, req.TeamID, userID, record.OperationOperateDebugApi, targetInfo.Name); err != nil {
		return "", err
	}

	retID, err := runner.RunMqtt(packer.GetSendMqttParam(targetInfo, &mqttDetailInfo, globalVariable))
	if err != nil {
		return "", fmt.Errorf("调试mqtt返回非200状态")
	}
	return retID, err
}

func GetSendMqttResult(ctx *gin.Context, req *rao.GetSendMqttResultReq) (rao.GetSendMqttResultResp, error) {
	mqttDebug := mao.MqttDebug{}
	err := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectMqttDebug).
		FindOne(ctx, bson.D{{"uuid", req.RetID}}).Decode(&mqttDebug)
	if err != nil {
		return rao.GetSendMqttResultResp{}, err
	}

	res := rao.GetSendMqttResultResp{
		TargetID:      mqttDebug.TargetID,
		TeamID:        mqttDebug.TeamID,
		Uuid:          mqttDebug.Uuid,
		Name:          mqttDebug.Name,
		RequestTime:   mqttDebug.RequestTime,
		ErrorType:     mqttDebug.ErrorType,
		IsSucceed:     mqttDebug.IsSucceed,
		SendBytes:     mqttDebug.SendBytes,
		ReceivedBytes: mqttDebug.ReceivedBytes,
		ErrorMsg:      mqttDebug.ErrorMsg,
		Timestamp:     mqttDebug.Timestamp,
		StartTime:     mqttDebug.StartTime,
		EndTime:       mqttDebug.EndTime,
	}

	return res, nil
}

func WsSendOrStopMessage(ctx *gin.Context, req *rao.WsSendOrStopMessageReq) error {
	statusChangeKey := fmt.Sprintf("WsStatusChange:%s", req.RetID)
	// 判断是发消息还是断开连接
	var forNum = 1
	execType := "发送websocket消息"
	if req.ConnectionStatusChange.Type == 1 { // 断开消息
		forNum = 2
		execType = "断开websocket连接"
	}

	for i := 0; i < forNum; i++ {
		statusChangeValue := rao.ConnectionStatusChange{
			Type:        req.ConnectionStatusChange.Type,
			MessageType: req.ConnectionStatusChange.MessageType,
			Message:     req.ConnectionStatusChange.Message,
		}
		statusChangeValueString, err := json.Marshal(statusChangeValue)
		if err == nil {
			// 发送计划相关信息到redis频道
			_, err = dal.GetRDB().Publish(ctx, statusChangeKey, string(statusChangeValueString)).Result()
			if err != nil {
				log.Logger.Info(execType + "--发送到对应频道失败")
				continue
			}
		} else {
			log.Logger.Info(execType + "--压缩数据失败")
			continue
		}
	}
	return nil
}

func TcpSendOrStopMessage(ctx *gin.Context, req *rao.TcpSendOrStopMessageReq) error {
	statusChangeKey := fmt.Sprintf("TcpStatusChange:%s", req.RetID)
	// 判断是发消息还是断开连接
	var forNum = 1
	execType := "发送tcp消息"
	if req.ConnectionStatusChange.Type == 1 { // 断开消息
		forNum = 2
		execType = "断开tcp连接"
	}

	for i := 0; i < forNum; i++ {
		statusChangeValue := rao.ConnectionStatusChange{
			Type:        req.ConnectionStatusChange.Type,
			MessageType: req.ConnectionStatusChange.MessageType,
			Message:     req.ConnectionStatusChange.Message,
		}
		statusChangeValueString, err := json.Marshal(statusChangeValue)
		if err == nil {
			// 发送计划相关信息到redis频道
			_, err = dal.GetRDB().Publish(ctx, statusChangeKey, string(statusChangeValueString)).Result()
			if err != nil {
				log.Logger.Info(execType + "--发送到对应频道失败")
				continue
			}
		} else {
			log.Logger.Info(execType + "--压缩数据失败")
			continue
		}
	}
	return nil
}
