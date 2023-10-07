package target

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/jwt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/log"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/record"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/uuid"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/mao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/query"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/runner"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/api"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/packer"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gen"
	"time"
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

func SendScene(ctx *gin.Context, teamID string, sceneID string, userID string) (string, error) {
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

	// 把调试信息入库
	InsertTargetDebugLog(ctx, teamID, sceneID, t.TargetType)

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
	InsertTargetDebugLog(ctx, teamID, targetID, targetInfo.TargetType)

	//targetDebugLog := dal.GetQuery().TargetDebugLog
	//insertData := &model.TargetDebugLog{
	//	TargetID:   targetID,
	//	TargetType: consts.TargetDebugLogApi,
	//	TeamID:     teamID,
	//}
	//err = targetDebugLog.WithContext(ctx).Create(insertData)
	//if err != nil {
	//	return "", err
	//}

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

func InsertTargetDebugLog(ctx *gin.Context, teamID, targetID, TargetType string) {
	logData := mao.TargetDebugLog{
		TeamID:     teamID,
		TargetID:   targetID,
		TargetType: TargetType,
		CreatedAt:  time.Now().Local(),
	}
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectTargetDebugLog)
	_, err := collection.InsertOne(ctx, logData)
	if err != nil {
		log.Logger.Error("调试日志入库失败")
	}
	return
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
	InsertTargetDebugLog(ctx, req.TeamID, req.TargetID, targetInfo.TargetType)

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
	sqlDebug := mao.APIDebug{}
	err := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectAPIDebug).
		FindOne(ctx, bson.D{{"uuid", req.RetID}}).Decode(&sqlDebug)
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, err, ""
	}

	if err == mongo.ErrNoDocuments {
		return nil, nil, ""
	}

	assertArr := make([]rao.AssertionMsg, 0, len(sqlDebug.Assert.AssertionMsgs))
	for _, a := range sqlDebug.Assert.AssertionMsgs {
		assertArr = append(assertArr, rao.AssertionMsg{
			Type:      a.Type,
			Code:      a.Code,
			IsSucceed: a.IsSucceed,
			Msg:       a.Msg,
		})
	}

	regexArr := make([]rao.Reg, 0, len(sqlDebug.Regex.Regs))
	for _, regInfo := range sqlDebug.Regex.Regs {
		regexArr = append(regexArr, rao.Reg{
			Key:   regInfo.Key,
			Value: regInfo.Value,
		})
	}

	res := &rao.GetSendSqlResultResp{
		RetID:       sqlDebug.UUID,
		RequestTime: sqlDebug.RequestTime,
		Status:      sqlDebug.Status,
		Assert: rao.AssertObj{
			AssertionMsgs: assertArr,
		},
		Regex: rao.RegexObj{
			Regs: regexArr,
		},
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
	InsertTargetDebugLog(ctx, req.TeamID, req.TargetID, targetInfo.TargetType)

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
	InsertTargetDebugLog(ctx, req.TeamID, req.TargetID, targetInfo.TargetType)

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
	InsertTargetDebugLog(ctx, req.TeamID, req.TargetID, targetInfo.TargetType)

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
	dubboDebug := mao.APIDebug{}
	err := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectAPIDebug).
		FindOne(ctx, bson.D{{"uuid", req.RetID}}).Decode(&dubboDebug)
	if err != nil && err != mongo.ErrNoDocuments {
		return rao.GetSendDubboResultResp{}, err
	}

	assertions := make([]rao.AssertionMsg, 0, len(dubboDebug.Assert.AssertionMsgs))
	for _, a := range dubboDebug.Assert.AssertionMsgs {
		assertions = append(assertions, rao.AssertionMsg{
			Code:      a.Code,
			IsSucceed: a.IsSucceed,
			Msg:       a.Msg,
		})
	}

	regexs := make([]rao.Reg, 0, len(dubboDebug.Regex.Regs))
	for _, r := range dubboDebug.Regex.Regs {
		regexs = append(regexs, rao.Reg{
			Key:   r.Key,
			Value: r.Value,
		})
	}

	res := rao.GetSendDubboResultResp{
		TargetID:     dubboDebug.ApiID,
		TeamID:       dubboDebug.TeamID,
		Uuid:         dubboDebug.UUID,
		Name:         dubboDebug.APIName,
		RequestType:  dubboDebug.RequestType,
		RequestBody:  dubboDebug.RequestBody,
		ResponseBody: dubboDebug.ResponseBody,
		Assert: rao.AssertObj{
			AssertionMsgs: assertions,
		},
		Regex: rao.RegexObj{
			Regs: regexs,
		},
		Status: dubboDebug.Status,
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
	InsertTargetDebugLog(ctx, req.TeamID, req.TargetID, targetInfo.TargetType)

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

func JustSendTarget(ctx *gin.Context, req *rao.SaveTargetReq) (string, error) {
	// 获取全局变量
	globalVariable, err := GetGlobalVariable(ctx, req.TeamID)

	retID, err := runner.RunTarget(packer.GetJustSendTargetParam(req, globalVariable))
	if err != nil {
		return "", fmt.Errorf("调试接口返回非200状态")
	}
	return retID, err
}

func SaveTargetHistoryRecord(ctx *gin.Context, req *rao.SaveTargetReq) (string, bool, error) {
	userID := jwt.GetUserIDByCtx(ctx)

	// 定义排序选项
	opts := options.Find().SetSort(bson.D{{"created_at", -1}}) // 按创建时间升序排序
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectTargetHistoryRecord)
	cursor, err := collection.Find(ctx, bson.D{{"target_id", req.TargetID}}, opts)
	if err != nil {
		log.Logger.Error("保存测试对象历史记录失败，err:", err)
		return "", false, err
	}
	historyRecord := make([]mao.TargetHistoryRecord, 0, 22)
	if err = cursor.All(ctx, &historyRecord); err != nil {
		return "", false, err
	}

	detailTemp, _ := json.Marshal(req)
	detail := string(detailTemp)

	md5Hash := md5.Sum(detailTemp)
	hashString := hex.EncodeToString(md5Hash[:])

	if len(historyRecord) > 0 {
		// 如果没有变化，则不保存此次记录
		if historyRecord[0].Hash == hashString {
			return historyRecord[0].Hash, historyRecord[0].IsSaveTag, nil
		}
	}

	if len(historyRecord) == 20 {
		// 需要删除老的记录
		_, _ = collection.DeleteOne(ctx, bson.D{{"uuid", historyRecord[19].Uuid}})
	}

	historyUuid := uuid.GetUUID()
	operationLog := mao.TargetHistoryRecord{
		TargetID:  req.TargetID,
		TeamID:    req.TeamID,
		UserID:    userID,
		Detail:    detail,
		Uuid:      historyUuid,
		Hash:      hashString,
		CreatedAt: time.Now().Local(),
	}
	if _, err = collection.InsertOne(ctx, operationLog); err != nil {
		return "", false, err
	}
	return historyUuid, false, nil
}

func GetHistoryRecord(ctx *gin.Context, req *rao.GetHistoryRecordReq) ([]rao.GetHistoryRecordResp, error) {
	userID := jwt.GetUserIDByCtx(ctx)

	opts := options.Find().SetSort(bson.D{{"created_at", -1}}) // 按创建时间降序排序
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectTargetHistoryRecord)
	cursor, err := collection.Find(ctx, bson.D{{"target_id", req.TargetID}}, opts)
	if err != nil {
		log.Logger.Error("获取测试对象历史记录失败，err:", err)
		return nil, err
	}
	historyRecord := make([]mao.TargetHistoryRecord, 0, 22)
	if err = cursor.All(ctx, &historyRecord); err != nil {
		return nil, err
	}

	userIDs := make([]string, 0, len(historyRecord))
	for _, v := range historyRecord {
		userIDs = append(userIDs, v.UserID)
	}

	// 查询用户信息
	userTB := dal.GetQuery().User
	userList, err := userTB.WithContext(ctx).Where(userTB.UserID.In(userIDs...)).Find()
	if err != nil {
		return nil, err
	}
	userMap := make(map[string]*model.User, len(userList))
	for _, v := range userList {
		userMap[v.UserID] = v
	}

	targetType := ""
	targetDetail := rao.SaveTargetReq{}
	if len(historyRecord) > 0 {
		err = json.Unmarshal([]byte(historyRecord[0].Detail), &targetDetail)
		if err == nil {
			targetType = targetDetail.TargetType
		}
	}

	dateResMap := make(map[string][]rao.GetHistoryRecordBase, 0)
	for _, v := range historyRecord {
		avatar := ""
		nickname := ""
		if userInfo, ok := userMap[v.UserID]; ok {
			avatar = userInfo.Avatar
			nickname = userInfo.Nickname
		}

		dateString := v.CreatedAt.Local().Format("2006-01-02")
		timeString := v.CreatedAt.Local().Format("15:04:05")

		isMyself := false
		if userID == v.UserID {
			isMyself = true
		}

		temp := rao.GetHistoryRecordBase{
			TargetID:   v.TargetID,
			TargetType: targetType,
			Uuid:       v.Uuid,
			UserID:     v.UserID,
			Nickname:   nickname,
			Avatar:     avatar,
			IsSaveTag:  v.IsSaveTag,
			Detail:     v.Detail,
			CreatedAt:  v.CreatedAt.Local().Unix(),
			DateString: dateString,
			TimeString: timeString,
			IsMyself:   isMyself,
		}

		dateResMap[dateString] = append(dateResMap[dateString], temp)
	}

	res := make([]rao.GetHistoryRecordResp, 0, 100)
	for k, v := range dateResMap {
		temp := rao.GetHistoryRecordResp{
			DataInt:       dateResMap[k][0].CreatedAt,
			DataString:    k,
			HistoryRecord: v,
		}
		res = append(res, temp)
	}
	return res, nil
}

func MarkTagVersion(ctx *gin.Context, req *rao.MarkTagVersionReq) error {
	userID := jwt.GetUserIDByCtx(ctx)

	// 查询当前测试对象的tag名称是否存在
	tagVersion := mao.TargetTagVersion{}
	collection1 := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectTargetTagVersion)
	err := collection1.FindOne(ctx, bson.D{{"target_id", req.TargetID},
		{"tag_name", req.TagName}}).Decode(&tagVersion)
	if err == nil {
		return fmt.Errorf("名称已存在")
	}

	// 查询历史记录信息
	historyRecord := mao.TargetHistoryRecord{}
	collection2 := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectTargetHistoryRecord)
	if req.Uuid != "" {
		err = collection2.FindOne(ctx, bson.D{{"uuid", req.Uuid}}).Decode(&historyRecord)
		if err != nil {
			return fmt.Errorf("没有查到历史记录详细信息")
		}
	} else { // 打最新的历史记录为tag
		opts := options.FindOne().SetSort(bson.D{{"created_at", -1}}) // 按创建时间降序排序
		err = collection2.FindOne(ctx, bson.D{{"target_id", req.TargetID}}, opts).Decode(&historyRecord)
		if err != nil {
			log.Logger.Error("获取测试对象历史记录失败，err:", err)
			return fmt.Errorf("没有查到历史记录详细信息")
		}
	}

	// 判断是否打过tag
	if historyRecord.IsSaveTag == true {
		return fmt.Errorf("已经打过tag版本")
	}

	// 组装tag数据
	tagVersionData := mao.TargetTagVersion{
		TagName:           req.TagName,
		TargetID:          req.TargetID,
		TeamID:            req.TeamID,
		UserID:            userID,
		Detail:            historyRecord.Detail,
		Uuid:              uuid.GetUUID(),
		HistoryRecordUuid: req.Uuid,
		CreatedAt:         time.Now().Local(),
	}

	_, err = collection1.InsertOne(ctx, tagVersionData)
	if err != nil {
		log.Logger.Info("标记tag失败，err:", err)
		return err
	}

	// 修改历史记录打标记状态
	filter := bson.D{{"uuid", historyRecord.Uuid}}
	update := bson.D{{"$set", bson.D{{"is_save_tag", true}}}}
	_, err = collection2.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

func GetTagVersionList(ctx *gin.Context, req *rao.GetTagVersionListReq) ([]rao.GetTagVersionListResp, error) {
	opts := options.Find().SetSort(bson.D{{"created_at", -1}}) // 按创建时间降序排序
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectTargetTagVersion)
	cur, err := collection.Find(ctx, bson.D{{"target_id", req.TargetID}}, opts)
	if err != nil {
		return nil, err
	}

	targetTagVersionList := make([]mao.TargetTagVersion, 0, 100)
	if err = cur.All(ctx, &targetTagVersionList); err != nil {
		return nil, err
	}

	userIDs := make([]string, 0, len(targetTagVersionList))
	for _, v := range targetTagVersionList {
		userIDs = append(userIDs, v.UserID)
	}

	tx := dal.GetQuery().User
	userList, err := tx.WithContext(ctx).Where(tx.UserID.In(userIDs...)).Find()
	if err != nil {
		return nil, err
	}

	userMap := make(map[string]*model.User, len(userList))
	for _, v := range userList {
		userMap[v.UserID] = v
	}

	targetType := ""
	targetDetail := rao.SaveTargetReq{}
	if len(targetTagVersionList) > 0 {
		err = json.Unmarshal([]byte(targetTagVersionList[0].Detail), &targetDetail)
		if err == nil {
			targetType = targetDetail.TargetType
		}
	}

	dateResMap := make(map[string][]rao.GetTargetTagVersionBase, 0)
	for _, v := range targetTagVersionList {
		avatar := ""
		nickname := ""
		if userInfo, ok := userMap[v.UserID]; ok {
			avatar = userInfo.Avatar
			nickname = userInfo.Nickname
		}

		dateString := v.CreatedAt.Local().Format("2006-01-02")
		timeString := v.CreatedAt.Local().Format("15:04:05")

		temp := rao.GetTargetTagVersionBase{
			TargetID:   v.TargetID,
			TargetType: targetType,
			Uuid:       v.Uuid,
			UserID:     v.UserID,
			Nickname:   nickname,
			Avatar:     avatar,
			TagName:    v.TagName,
			Detail:     v.Detail,
			CreatedAt:  v.CreatedAt.Local().Unix(),
			DateString: dateString,
			TimeString: timeString,
		}

		dateResMap[dateString] = append(dateResMap[dateString], temp)
	}

	res := make([]rao.GetTagVersionListResp, 0, 100)
	for k, v := range dateResMap {
		temp := rao.GetTagVersionListResp{
			DataInt:    dateResMap[k][0].CreatedAt,
			DataString: k,
			TagVersion: v,
		}
		res = append(res, temp)
	}
	return res, nil
}

func UpdateTagVersion(ctx *gin.Context, req *rao.UpdateTagVersionReq) error {
	filter := bson.D{{"uuid", req.Uuid}}
	update := bson.D{{"$set", bson.D{{"tag_name", req.TagName}}}}
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectTargetTagVersion)
	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}

func DeleteTagVersion(ctx *gin.Context, req *rao.DeleteTagVersionReq) error {
	// 查询当前测试对象的tag名称是否存在
	tagVersion := mao.TargetTagVersion{}
	collection1 := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectTargetTagVersion)
	err := collection1.FindOne(ctx, bson.D{{"uuid", req.Uuid}}).Decode(&tagVersion)
	if err != nil {
		return err
	}

	// 删除tag版本
	_, err = collection1.DeleteOne(ctx, bson.D{{"uuid", req.Uuid}})
	if err != nil {
		return err
	}

	// 修改历史记录打标记状态
	filter := bson.D{{"uuid", tagVersion.HistoryRecordUuid}}
	update := bson.D{{"$set", bson.D{{"is_save_tag", false}}}}
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectTargetHistoryRecord)
	_, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}

func RestoreHistoryOrTag(ctx *gin.Context, req *rao.RestoreHistoryOrTagReq) error {
	saveTargetReq := rao.SaveTargetReq{}
	// 判断恢复的来源
	if req.RestoreType == 1 { // 历史记录
		// 查询历史记录信息
		historyRecord := mao.TargetHistoryRecord{}
		collection2 := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectTargetHistoryRecord)
		err := collection2.FindOne(ctx, bson.D{{"uuid", req.Uuid}}).Decode(&historyRecord)
		if err != nil {
			return fmt.Errorf("没有查到历史记录详细信息")
		}

		err = json.Unmarshal([]byte(historyRecord.Detail), &saveTargetReq)
		if err != nil {
			return fmt.Errorf("解析历史记录详情数据失败")
		}
	} else { // tag
		// 查询当前测试对象的tag名称是否存在
		tagVersion := mao.TargetTagVersion{}
		collection1 := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectTargetTagVersion)
		err := collection1.FindOne(ctx, bson.D{{"uuid", req.Uuid}}).Decode(&tagVersion)
		if err != nil {
			return fmt.Errorf("名称已存在")
		}
		err = json.Unmarshal([]byte(tagVersion.Detail), &saveTargetReq)
		if err != nil {
			return fmt.Errorf("解析历史记录详情数据失败")
		}
	}

	if saveTargetReq.TargetType == consts.TargetTypeSql { // sql
		_, err := api.SaveSql(ctx, &saveTargetReq)
		if err != nil {
			return err
		}
	} else if saveTargetReq.TargetType == consts.TargetTypeTcp { // tcp
		_, err := api.SaveTcp(ctx, &saveTargetReq)
		if err != nil {
			return err
		}
	} else if saveTargetReq.TargetType == consts.TargetTypeWebsocket { // websocket
		_, err := api.SaveWebsocket(ctx, &saveTargetReq)
		if err != nil {
			return err
		}
	} else if saveTargetReq.TargetType == consts.TargetTypeMQTT { // Mqtt
		_, err := api.SaveMQTT(ctx, &saveTargetReq)
		if err != nil {
			return err
		}
	} else if saveTargetReq.TargetType == consts.TargetTypeDubbo { // Dubbo
		_, err := api.SaveDubbo(ctx, &saveTargetReq)
		if err != nil {
			return err
		}
	} else { // api
		_, err := api.Save(ctx, &saveTargetReq, jwt.GetUserIDByCtx(ctx))
		if err != nil {
			return err
		}
	}

	// 保存历史记录
	_, _, err := SaveTargetHistoryRecord(ctx, &saveTargetReq)
	if err != nil {
		return err
	}
	return nil
}

func GetCanSyncData(ctx *gin.Context, req *rao.GetCanSyncDataReq) ([]rao.SyncChildren, error) {
	// 组装最后返回值
	res := make([]rao.SyncChildren, 0, 3)
	// 查询当前测试对象被引用关系数据
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectTargetCiteRelation)
	cursor, err := collection.Find(ctx, bson.D{{"target_id", req.TargetID}})
	if err != nil {
		return nil, err
	}
	targetCiteRelations := make([]mao.TargetCiteRelation, 0, 1000)
	if err = cursor.All(ctx, &targetCiteRelations); err != nil {
		return res, err
	}

	if len(targetCiteRelations) == 0 {
		return res, nil
	}

	allStressPlanIDs := make([]string, 0, len(targetCiteRelations))
	allAutoPlanIDs := make([]string, 0, len(targetCiteRelations))
	allSceneIDs := make([]string, 0, len(targetCiteRelations))
	allCaseIDs := make([]string, 0, len(targetCiteRelations))
	for _, v := range targetCiteRelations {
		allSceneIDs = append(allSceneIDs, v.SceneID)

		if v.PlanID != "" {
			if v.Source == consts.TargetSourcePlan {
				allStressPlanIDs = append(allStressPlanIDs, v.PlanID)
			}

			if v.Source == consts.TargetSourceAutoPlan {
				allAutoPlanIDs = append(allAutoPlanIDs, v.PlanID)
			}
		}

		if v.CaseID != "" {
			allCaseIDs = append(allCaseIDs, v.CaseID)
		}
	}

	isShowTargetManage := false
	isShowSceneManage := false
	isShowStressPlan := false
	isShowAutoPlan := false

	tx := dal.GetQuery()
	// 查询所有场景基本数据
	sceneList, err := tx.Target.WithContext(ctx).Where(tx.Target.TargetID.In(allSceneIDs...)).Find()
	if err != nil {
		return res, err
	}
	sceneMap := make(map[string]*model.Target, len(sceneList))
	for _, v := range sceneList {
		sceneMap[v.TargetID] = v
	}

	// 查询所有用例基本数据
	caseList, err := tx.Target.WithContext(ctx).Where(tx.Target.TargetID.In(allCaseIDs...)).Find()
	if err != nil {
		return res, err
	}
	caseMap := make(map[string]*model.Target, len(caseList))
	for _, v := range caseList {
		caseMap[v.TargetID] = v
	}

	// 查询所有性能计划基本数据
	stressPlanList, err := tx.StressPlan.WithContext(ctx).Where(tx.StressPlan.PlanID.In(allStressPlanIDs...)).Find()
	if err != nil {
		return res, err
	}

	// 查询所有自动换计划基本数据
	autoPlanList, err := tx.AutoPlan.WithContext(ctx).Where(tx.AutoPlan.PlanID.In(allAutoPlanIDs...)).Find()
	if err != nil {
		return res, err
	}

	// 查询所有场景flow数据
	collection = dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectFlow)
	cursor, err = collection.Find(ctx, bson.D{{"scene_id", bson.D{{"$in", allSceneIDs}}}})
	if err != nil {
		return res, err
	}
	allFlowList := make([]mao.Flow, 0, len(allSceneIDs))
	if err = cursor.All(ctx, &allFlowList); err != nil {
		return res, err
	}
	sceneFlowMap := make(map[string]mao.Flow, len(allFlowList))
	for _, v := range allFlowList {
		sceneFlowMap[v.SceneID] = v
	}

	// 查询所有用例flow数据
	collection = dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectSceneCaseFlow)
	cursor, err = collection.Find(ctx, bson.D{{"scene_case_id", bson.D{{"$in", allCaseIDs}}}})
	if err != nil {
		return res, err
	}
	allCaseFlowList := make([]mao.SceneCaseFlow, 0, len(allCaseIDs))
	if err = cursor.All(ctx, &allCaseFlowList); err != nil {
		return res, err
	}
	caseFlowMap := make(map[string]mao.SceneCaseFlow, len(allCaseFlowList))
	for _, v := range allCaseFlowList {
		caseFlowMap[v.SceneCaseID] = v
	}

	// 组装场景管理里面的数据
	sceneHaveApiMap := make(map[string][]rao.SyncChildren, len(sceneFlowMap))
	for _, v := range sceneFlowMap {
		nodes := mao.Node{}
		if err := bson.Unmarshal(v.Nodes, &nodes); err != nil {
			log.Logger.Info("flow.nodes bson unmarshal err %w", err)
			continue
		}
		if len(nodes.Nodes) != 0 {
			for _, vv := range nodes.Nodes {
				if vv.API.TargetID == req.TargetID {
					temp := rao.SyncChildren{
						Title:   "[测试对象] " + vv.API.Name,
						Key:     vv.ID,
						Source:  sceneMap[v.SceneID].Source,
						PlanID:  sceneMap[v.SceneID].PlanID,
						SceneID: v.SceneID,
						Type:    "target",
					}
					sceneHaveApiMap[v.SceneID] = append(sceneHaveApiMap[v.SceneID], temp)
				}
			}
		}
	}

	// 组装每个用例下所有接口数据
	caseHaveApiMap := make(map[string][]rao.SyncChildren, len(caseFlowMap))
	for _, v := range caseFlowMap {
		nodes := mao.Node{}
		if err := bson.Unmarshal(v.Nodes, &nodes); err != nil {
			log.Logger.Info("flow.nodes bson unmarshal err %w", err)
			continue
		}
		if len(nodes.Nodes) != 0 {
			for _, vv := range nodes.Nodes {
				if vv.API.TargetID == req.TargetID {
					temp := rao.SyncChildren{
						Title:   "[测试对象] " + vv.API.Name,
						Key:     vv.ID,
						Source:  4,
						PlanID:  sceneMap[v.SceneID].PlanID,
						SceneID: v.SceneID,
						CaseID:  v.SceneCaseID,
						Type:    "target",
					}
					caseHaveApiMap[v.SceneCaseID] = append(caseHaveApiMap[v.SceneCaseID], temp)
				}
			}
		}
	}

	if req.Source != consts.TargetSourceApi {
		isShowTargetManage = true
	}

	for _, v := range sceneList {
		if v.Source == consts.TargetSourceScene { // 场景管理
			_, ok := sceneHaveApiMap[v.TargetID]
			if ok && req.Source != consts.TargetSourceScene {
				isShowSceneManage = true
			}
		}

		if v.Source == consts.TargetSourcePlan { // 性能计划
			_, ok := sceneHaveApiMap[v.TargetID]
			if ok && req.Source != consts.TargetSourcePlan {
				isShowStressPlan = true
			}
		}

		if v.Source == consts.TargetSourceAutoPlan { // 自动化计划
			_, ok := sceneHaveApiMap[v.TargetID]
			if ok && req.Source != consts.TargetSourceAutoPlan {
				isShowAutoPlan = true
			}
		}
	}

	sceneChildren := make([]rao.SyncChildren, 0, len(sceneList))
	stressPlanChildren := make([]rao.SyncChildren, 0, len(sceneList))
	autoPlanChildren := make([]rao.SyncChildren, 0, len(sceneList))
	for _, v := range sceneList {
		// 场景管理下面数据
		if v.Source == consts.TargetSourceScene {
			// 场景管理下的接口
			apiTemp := rao.SyncChildren{
				Title:    "[场景] " + v.Name,
				Key:      v.TargetID,
				Source:   v.Source,
				Children: sceneHaveApiMap[v.TargetID],
				Type:     v.TargetType,
			}

			// 场景下的用例
			for _, vv := range caseList {
				if vv.ParentID == v.TargetID {
					caseTemp := rao.SyncChildren{
						Title:    "[用例] " + vv.Name,
						Key:      vv.TargetID,
						Source:   4,
						Children: caseHaveApiMap[vv.TargetID],
						Type:     v.TargetType,
					}
					apiTemp.Children = append(apiTemp.Children, caseTemp)
				}
			}
			sceneChildren = append(sceneChildren, apiTemp)
		}

		// 性能计划下面数据
		if v.Source == consts.TargetSourcePlan {
			// 场景下的接口
			apiTemp := rao.SyncChildren{
				Title:    "[场景] " + v.Name,
				Key:      v.TargetID,
				Source:   v.Source,
				Children: sceneHaveApiMap[v.TargetID],
				Type:     v.TargetType,
			}
			stressPlanChildren = append(stressPlanChildren, apiTemp)
		}

		// 自动化下面数据
		if v.Source == consts.TargetSourceAutoPlan {
			// 场景下的接口
			apiTemp := rao.SyncChildren{
				Title:    "[场景] " + v.Name,
				Key:      v.TargetID,
				Source:   v.Source,
				Children: sceneHaveApiMap[v.TargetID],
				Type:     v.TargetType,
			}

			// 场景下的用例
			for _, vv := range caseList {
				if vv.ParentID == v.TargetID {
					caseTemp := rao.SyncChildren{
						Title:    "[用例] " + vv.Name,
						Key:      vv.TargetID,
						Source:   4,
						Children: caseHaveApiMap[vv.TargetID],
						Type:     v.TargetType,
					}
					apiTemp.Children = append(apiTemp.Children, caseTemp)
				}
			}
			autoPlanChildren = append(autoPlanChildren, apiTemp)
		}
	}

	if req.Source != consts.TargetSourceApi {
		isShowSceneManage = false
		isShowStressPlan = false
		isShowAutoPlan = false
	}

	// 组装场景管理
	sceneRes := rao.SyncChildren{}
	if isShowSceneManage {
		sceneRes.Title = "场景管理"
		sceneRes.Key = uuid.GetUUID()
		sceneRes.Source = consts.TargetSourceScene
		sceneRes.Children = sceneChildren
		sceneRes.Type = "scene_manage"
	}

	// 组装性能计划数据
	stressPlanRes := rao.SyncChildren{}
	if isShowStressPlan {
		stressPlanRes.Title = "性能计划"
		stressPlanRes.Key = uuid.GetUUID()
		stressPlanRes.Source = consts.TargetSourcePlan
		for _, v := range stressPlanList {
			temp := rao.SyncChildren{
				Title:    "[计划] " + v.PlanName,
				Key:      v.PlanID,
				Source:   consts.TargetSourcePlan,
				Children: stressPlanChildren,
				Type:     "stress_plan",
			}
			stressPlanRes.Children = append(stressPlanRes.Children, temp)
		}
	}

	// 组装自动化计划数据
	autoPlanRes := rao.SyncChildren{}
	if isShowAutoPlan {
		autoPlanRes.Title = "自动化计划"
		autoPlanRes.Key = uuid.GetUUID()
		autoPlanRes.Source = consts.TargetSourceAutoPlan
		for _, v := range autoPlanList {
			temp := rao.SyncChildren{
				Title:    "[计划] " + v.PlanName,
				Key:      v.PlanID,
				Source:   consts.TargetSourceAutoPlan,
				Children: autoPlanChildren,
				Type:     "auto_plan",
			}
			autoPlanRes.Children = append(autoPlanRes.Children, temp)
		}
	}

	// 查询target基本信息
	targetInfo, err := tx.Target.WithContext(ctx).Where(tx.Target.TargetID.Eq(req.TargetID)).First()
	if err != nil {
		isShowTargetManage = false
	}

	// 测试对象数据
	targetRes := rao.SyncChildren{}
	if isShowTargetManage {
		targetChildren := rao.SyncChildren{
			Title:  "[测试对象] " + targetInfo.Name,
			Key:    targetInfo.TargetID,
			Source: targetInfo.Source,
			Type:   "target",
		}

		targetRes.Title = "测试对象"
		targetRes.Key = uuid.GetUUID()
		targetRes.Source = consts.MockTargetSourceApi
		targetRes.Children = append(targetRes.Children, targetChildren)
	}

	if targetRes.Title != "" {
		res = append(res, targetRes)
	}
	if sceneRes.Title != "" {
		res = append(res, sceneRes)
	}
	if stressPlanRes.Title != "" {
		res = append(res, stressPlanRes)
	}
	if autoPlanRes.Title != "" {
		res = append(res, autoPlanRes)
	}
	return res, nil
}

func ExecSyncData(ctx *gin.Context, req *rao.ExecSyncDataReq) error {
	if req.Source != consts.TargetSourceApi {
		err := ExecSyncDataFlowToTarget(ctx, req)
		return err
	}

	tx := dal.GetQuery()
	collectionFlow := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectFlow)
	collectionCaseFlow := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectSceneCaseFlow)
	collectionApi := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectAPI)
	if req.SyncType == consts.TargetSyncDataTypePush { // 推送
		targetInfo, err := tx.Target.WithContext(ctx).Where(tx.Target.TargetID.Eq(req.TargetID)).First()
		if err != nil {
			return err
		}

		targetDetail := mao.API{}
		collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectAPI)
		err = collection.FindOne(ctx, bson.D{{"target_id", req.TargetID}}).Decode(&targetDetail)
		if err != nil {
			return err
		}

		for _, v := range req.SyncApiInfo {
			if v.CaseID == "" { // 场景管理或计划下面的flow
				sceneFlow := mao.Flow{}
				err = collectionFlow.FindOne(ctx, bson.D{{"scene_id", v.SceneID}}).Decode(&sceneFlow)
				if err != nil {
					log.Logger.Error("获取flow数据失败， err:", err)
					continue
				}

				nodes := mao.Node{}
				if err = bson.Unmarshal(sceneFlow.Nodes, &nodes); err != nil {
					log.Logger.Error("解析nodes失败， err:", err)
					continue
				}

				for kk, vv := range nodes.Nodes {
					if vv.ID == v.NodeID {
						for _, vvv := range req.SyncContent {
							if vvv == consts.TargetSyncMethodID {
								nodes.Nodes[kk].API.Method = targetInfo.Method
							}

							if vvv == consts.TargetSyncUrlID {
								nodes.Nodes[kk].API.URL = targetDetail.URL
								nodes.Nodes[kk].API.Request.URL = targetDetail.URL
							}

							if vvv == consts.TargetSyncCookieID {
								cookie := rao.Cookie{}
								err = bson.Unmarshal(targetDetail.Cookie, &cookie)
								if err != nil {
									log.Logger.Error("解析cookie失败，err:", err)
									continue
								}
								nodes.Nodes[kk].API.Request.Cookie = cookie
							}

							if vvv == consts.TargetSyncHeaderID {
								header := rao.Header{}
								err = bson.Unmarshal(targetDetail.Header, &header)
								if err != nil {
									log.Logger.Error("解析header失败，err:", err)
									continue
								}
								nodes.Nodes[kk].API.Request.Header = header
							}

							if vvv == consts.TargetSyncQueryID {
								queryList := rao.Query{}
								err = bson.Unmarshal(targetDetail.Query, &queryList)
								if err != nil {
									log.Logger.Error("解析query失败，err:", err)
									continue
								}
								nodes.Nodes[kk].API.Request.Query = queryList
							}

							if vvv == consts.TargetSyncBodyID {
								body := rao.Body{}
								err = bson.Unmarshal(targetDetail.Body, &body)
								if err != nil {
									log.Logger.Error("解析body失败，err:", err)
									continue
								}
								nodes.Nodes[kk].API.Request.Body = body

							}

							if vvv == consts.TargetSyncAuthID {
								auth := rao.Auth{}
								err = bson.Unmarshal(targetDetail.Auth, &auth)
								if err != nil {
									log.Logger.Error("解析auth失败，err:", err)
									continue
								}
								nodes.Nodes[kk].API.Request.Auth = auth
							}

							if vvv == consts.TargetSyncAssertID {
								assert := mao.Assert{}
								err = bson.Unmarshal(targetDetail.Assert, &assert)
								if err != nil {
									log.Logger.Error("解析assert失败，err:", err)
									continue
								}

								assertTemp := make([]rao.Assert, 0, len(assert.Assert))
								for _, assertInfo := range assert.Assert {
									temp := rao.Assert{
										ResponseType: assertInfo.ResponseType,
										Var:          assertInfo.Var,
										Compare:      assertInfo.Compare,
										Val:          assertInfo.Val,
										IsChecked:    assertInfo.IsChecked,
										Index:        assertInfo.Index,
									}
									assertTemp = append(assertTemp, temp)
								}
								nodes.Nodes[kk].API.Request.Assert = assertTemp
							}

							if vvv == consts.TargetSyncRegexID {
								regex := mao.Regex{}
								err = bson.Unmarshal(targetDetail.Regex, &regex)
								if err != nil {
									log.Logger.Error("解析regex失败，err:", err)
									continue
								}

								regexTemp := make([]rao.Regex, 0, len(regex.Regex))
								for _, regexInfo := range regex.Regex {
									temp := rao.Regex{
										IsChecked: regexInfo.IsChecked,
										Type:      regexInfo.Type,
										Var:       regexInfo.Var,
										Val:       regexInfo.Val,
										Express:   regexInfo.Express,
										Index:     regexInfo.Index,
									}
									regexTemp = append(regexTemp, temp)
								}
								nodes.Nodes[kk].API.Request.Regex = regexTemp
							}

							if vvv == consts.TargetSyncConfigID {
								httpApiSetupTemp := rao.HttpApiSetup{
									IsRedirects:         targetDetail.HttpApiSetup.IsRedirects,
									RedirectsNum:        targetDetail.HttpApiSetup.RedirectsNum,
									ReadTimeOut:         targetDetail.HttpApiSetup.ReadTimeOut,
									WriteTimeOut:        targetDetail.HttpApiSetup.WriteTimeOut,
									ClientName:          targetDetail.HttpApiSetup.ClientName,
									KeepAlive:           targetDetail.HttpApiSetup.KeepAlive,
									MaxIdleConnDuration: targetDetail.HttpApiSetup.MaxIdleConnDuration,
									MaxConnPerHost:      targetDetail.HttpApiSetup.MaxConnPerHost,
									UserAgent:           targetDetail.HttpApiSetup.UserAgent,
									MaxConnWaitTimeout:  targetDetail.HttpApiSetup.MaxConnWaitTimeout,
								}
								nodes.Nodes[kk].API.Request.HttpApiSetup = httpApiSetupTemp
							}
						}
					}
				}

				nodesData, err := bson.Marshal(mao.Node{Nodes: nodes.Nodes})
				if err != nil {
					log.Logger.Info("flow.nodes bson marshal err %w", err)
				}
				sceneFlow.Nodes = nodesData
				_, err = collectionFlow.UpdateOne(ctx, bson.D{{"scene_id", v.SceneID}}, bson.M{"$set": sceneFlow})
				if err != nil {
					return err
				}
			} else { // 测试用例flow里面的接口
				caseFlow := mao.SceneCaseFlow{}
				err = collectionCaseFlow.FindOne(ctx, bson.D{{"scene_case_id", v.CaseID}}).Decode(&caseFlow)
				if err != nil {
					continue
				}

				nodes := mao.SceneCaseFlowNode{}
				if err = bson.Unmarshal(caseFlow.Nodes, &nodes); err != nil {
					log.Logger.Error("解析nodes失败， err:", err)
					continue
				}

				for kk, vv := range nodes.Nodes {
					if vv.ID == v.NodeID {
						for _, vvv := range req.SyncContent {
							if vvv == consts.TargetSyncMethodID {
								nodes.Nodes[kk].API.Method = targetInfo.Method
							}

							if vvv == consts.TargetSyncUrlID {
								nodes.Nodes[kk].API.URL = targetDetail.URL
								nodes.Nodes[kk].API.Request.URL = targetDetail.URL
							}

							if vvv == consts.TargetSyncCookieID {
								cookie := rao.Cookie{}
								err = bson.Unmarshal(targetDetail.Cookie, &cookie)
								if err != nil {
									log.Logger.Error("解析cookie失败，err:", err)
									continue
								}
								nodes.Nodes[kk].API.Request.Cookie = cookie
							}

							if vvv == consts.TargetSyncHeaderID {
								header := rao.Header{}
								err = bson.Unmarshal(targetDetail.Header, &header)
								if err != nil {
									log.Logger.Error("解析header失败，err:", err)
									continue
								}
								nodes.Nodes[kk].API.Request.Header = header
							}

							if vvv == consts.TargetSyncQueryID {
								queryList := rao.Query{}
								err = bson.Unmarshal(targetDetail.Query, &queryList)
								if err != nil {
									log.Logger.Error("解析query失败，err:", err)
									continue
								}
								nodes.Nodes[kk].API.Request.Query = queryList
							}

							if vvv == consts.TargetSyncBodyID {
								body := rao.Body{}
								err = bson.Unmarshal(targetDetail.Body, &body)
								if err != nil {
									log.Logger.Error("解析body失败，err:", err)
									continue
								}
								nodes.Nodes[kk].API.Request.Body = body
							}

							if vvv == consts.TargetSyncAuthID {
								auth := rao.Auth{}
								err = bson.Unmarshal(targetDetail.Auth, &auth)
								if err != nil {
									log.Logger.Error("解析auth失败，err:", err)
									continue
								}
								nodes.Nodes[kk].API.Request.Auth = auth
							}

							if vvv == consts.TargetSyncAssertID {
								assert := mao.Assert{}
								err = bson.Unmarshal(targetDetail.Assert, &assert)
								if err != nil {
									log.Logger.Error("解析assert失败，err:", err)
									continue
								}
								assertArr := make([]rao.Assert, 0, len(assert.Assert))
								for _, info := range assert.Assert {
									temp := rao.Assert{
										ResponseType: info.ResponseType,
										Var:          info.Var,
										Compare:      info.Compare,
										Val:          info.Val,
										IsChecked:    info.IsChecked,
										Index:        info.Index,
									}
									assertArr = append(assertArr, temp)
								}
								nodes.Nodes[kk].API.Request.Assert = assertArr
							}

							if vvv == consts.TargetSyncRegexID {
								regex := mao.Regex{}
								err = bson.Unmarshal(targetDetail.Regex, &regex)
								if err != nil {
									log.Logger.Error("解析regex失败，err:", err)
									continue
								}
								regexArr := make([]rao.Regex, 0, len(regex.Regex))
								for _, info := range regex.Regex {
									temp := rao.Regex{
										IsChecked: info.IsChecked,
										Type:      info.Type,
										Var:       info.Var,
										Val:       info.Val,
										Express:   info.Express,
										Index:     info.Index,
									}
									regexArr = append(regexArr, temp)
								}
								nodes.Nodes[kk].API.Request.Regex = regexArr
							}

							if vvv == consts.TargetSyncConfigID {
								httpApiSetupTemp := rao.HttpApiSetup{
									IsRedirects:         targetDetail.HttpApiSetup.IsRedirects,
									RedirectsNum:        targetDetail.HttpApiSetup.RedirectsNum,
									ReadTimeOut:         targetDetail.HttpApiSetup.ReadTimeOut,
									WriteTimeOut:        targetDetail.HttpApiSetup.WriteTimeOut,
									ClientName:          targetDetail.HttpApiSetup.ClientName,
									KeepAlive:           targetDetail.HttpApiSetup.KeepAlive,
									MaxIdleConnDuration: targetDetail.HttpApiSetup.MaxIdleConnDuration,
									MaxConnPerHost:      targetDetail.HttpApiSetup.MaxConnPerHost,
									UserAgent:           targetDetail.HttpApiSetup.UserAgent,
									MaxConnWaitTimeout:  targetDetail.HttpApiSetup.MaxConnWaitTimeout,
								}
								nodes.Nodes[kk].API.Request.HttpApiSetup = httpApiSetupTemp
							}
						}
					}
				}

				nodesData, err := bson.Marshal(mao.SceneCaseFlowNode{Nodes: nodes.Nodes})
				if err != nil {
					log.Logger.Info("flow.nodes bson marshal err %w", err)
				}
				caseFlow.Nodes = nodesData
				_, err = collectionCaseFlow.UpdateOne(ctx, bson.D{{"scene_case_id", v.CaseID}}, bson.M{"$set": caseFlow})
				if err != nil {
					return err
				}
			}
		}

	} else { // 拉取
		if req.SyncApiInfo[0].CaseID == "" { // 场景下的接口
			sceneFlow := mao.Flow{}
			err := collectionFlow.FindOne(ctx, bson.D{{"scene_id", req.SyncApiInfo[0].SceneID}}).Decode(&sceneFlow)
			if err != nil {
				return err
			}

			nodes := mao.Node{}
			if err = bson.Unmarshal(sceneFlow.Nodes, &nodes); err != nil {
				log.Logger.Error("解析nodes失败， err:", err)
				return err
			}

			// 查询测试对象详情
			apiDetail := mao.API{}
			err = collectionApi.FindOne(ctx, bson.D{{"target_id", req.TargetID}}).Decode(&apiDetail)
			if err != nil {
				return err
			}

			for _, v := range nodes.Nodes {
				if v.ID == req.SyncApiInfo[0].NodeID {
					for _, vvv := range req.SyncContent {
						if vvv == consts.TargetSyncMethodID {
							_, err = tx.Target.WithContext(ctx).Where(tx.Target.TargetID.Eq(req.TargetID)).
								UpdateSimple(tx.Target.Method.Value(v.API.Method))
							if err != nil {
								log.Logger.Error("同步method失败，err:", err)
							}
						}

						if vvv == consts.TargetSyncUrlID {
							apiDetail.URL = v.API.URL
						}

						if vvv == consts.TargetSyncCookieID {
							cookieTemp := rao.Cookie{}
							parameterTemp := make([]rao.Parameter, 0, 20)
							for _, vvvv := range v.API.Request.Cookie.Parameter {
								temp := rao.Parameter{
									IsChecked:   vvvv.IsChecked,
									Type:        vvvv.Type,
									Key:         vvvv.Key,
									Value:       vvvv.Value,
									NotNull:     vvvv.NotNull,
									Description: vvvv.Description,
									FileBase64:  vvvv.FileBase64,
									FieldType:   vvvv.FieldType,
								}
								parameterTemp = append(parameterTemp, temp)
							}
							cookieTemp.Parameter = parameterTemp
							cookieData, err := bson.Marshal(cookieTemp)
							if err != nil {
								log.Logger.Info("拉取cookie失败 err:", err)
							}
							apiDetail.Cookie = cookieData
						}

						if vvv == consts.TargetSyncHeaderID {
							headerTemp := rao.Header{}
							parameterTemp := make([]rao.Parameter, 0, 20)
							for _, vvvv := range v.API.Request.Header.Parameter {
								temp := rao.Parameter{
									IsChecked:   vvvv.IsChecked,
									Type:        vvvv.Type,
									Key:         vvvv.Key,
									Value:       vvvv.Value,
									NotNull:     vvvv.NotNull,
									Description: vvvv.Description,
									FileBase64:  vvvv.FileBase64,
									FieldType:   vvvv.FieldType,
								}
								parameterTemp = append(parameterTemp, temp)
							}
							headerTemp.Parameter = parameterTemp
							headerData, err := bson.Marshal(headerTemp)
							if err != nil {
								log.Logger.Info("拉取header失败 err:", err)
							}
							apiDetail.Header = headerData
						}

						if vvv == consts.TargetSyncQueryID {
							queryTemp := rao.Query{}
							parameterTemp := make([]rao.Parameter, 0, 20)
							for _, vvvv := range v.API.Request.Query.Parameter {
								temp := rao.Parameter{
									IsChecked:   vvvv.IsChecked,
									Type:        vvvv.Type,
									Key:         vvvv.Key,
									Value:       vvvv.Value,
									NotNull:     vvvv.NotNull,
									Description: vvvv.Description,
									FileBase64:  vvvv.FileBase64,
									FieldType:   vvvv.FieldType,
								}
								parameterTemp = append(parameterTemp, temp)
							}
							queryTemp.Parameter = parameterTemp
							queryData, err := bson.Marshal(&queryTemp)
							if err != nil {
								log.Logger.Info("拉取query失败 err:", err)
							}
							apiDetail.Query = queryData
						}

						if vvv == consts.TargetSyncBodyID {
							bodyData, err := bson.Marshal(&v.API.Request.Body)
							if err != nil {
								log.Logger.Info("拉取body失败 err:", err)
							}

							apiDetail.Body = bodyData
						}

						if vvv == consts.TargetSyncAuthID {
							authData, err := bson.Marshal(&v.API.Request.Auth)
							if err != nil {
								log.Logger.Info("拉取auth失败 err:", err)
							}
							apiDetail.Auth = authData
						}

						if vvv == consts.TargetSyncAssertID {
							assertData, err := bson.Marshal(&mao.Assert{Assert: v.API.Request.Assert})
							if err != nil {
								log.Logger.Info("拉取assert失败 err:", err)
							}
							apiDetail.Assert = assertData
						}

						if vvv == consts.TargetSyncRegexID {
							regexData, err := bson.Marshal(mao.Regex{Regex: v.API.Request.Regex})
							if err != nil {
								log.Logger.Info("拉取regex失败 err:", err)
							}
							apiDetail.Regex = regexData
						}

						if vvv == consts.TargetSyncConfigID {
							httpApiSetupTemp := mao.HttpApiSetup{
								IsRedirects:         v.API.Request.HttpApiSetup.IsRedirects,
								RedirectsNum:        v.API.Request.HttpApiSetup.RedirectsNum,
								ReadTimeOut:         v.API.Request.HttpApiSetup.ReadTimeOut,
								WriteTimeOut:        v.API.Request.HttpApiSetup.WriteTimeOut,
								ClientName:          v.API.Request.HttpApiSetup.ClientName,
								KeepAlive:           v.API.Request.HttpApiSetup.KeepAlive,
								MaxIdleConnDuration: v.API.Request.HttpApiSetup.MaxIdleConnDuration,
								MaxConnPerHost:      v.API.Request.HttpApiSetup.MaxConnPerHost,
								UserAgent:           v.API.Request.HttpApiSetup.UserAgent,
								MaxConnWaitTimeout:  v.API.Request.HttpApiSetup.MaxConnWaitTimeout,
							}
							apiDetail.HttpApiSetup = httpApiSetupTemp
						}
					}

					// 更新数据
					update := bson.M{"$set": apiDetail}
					_, err = collectionApi.UpdateOne(ctx, bson.D{{"target_id", req.TargetID}}, update)
					if err != nil {
						log.Logger.Error("拉取数据失败，err:", err)
						return err
					}
					break
				}
			}
		} else { // 用例下的接口
			caseFlow := mao.SceneCaseFlow{}
			err := collectionCaseFlow.FindOne(ctx, bson.D{{"scene_case_id", req.SyncApiInfo[0].CaseID}}).Decode(&caseFlow)
			if err != nil {
				return err
			}

			nodes := mao.SceneCaseFlowNode{}
			if err = bson.Unmarshal(caseFlow.Nodes, &nodes); err != nil {
				log.Logger.Error("解析nodes失败， err:", err)
				return err
			}

			// 查询测试对象详情
			apiDetail := mao.API{}
			err = collectionApi.FindOne(ctx, bson.D{{"target_id", req.TargetID}}).Decode(&apiDetail)
			if err != nil {
				return err
			}

			for _, v := range nodes.Nodes {
				if v.ID == req.SyncApiInfo[0].NodeID {
					for _, vvv := range req.SyncContent {
						if vvv == consts.TargetSyncMethodID {
							_, err = tx.Target.WithContext(ctx).Where(tx.Target.TargetID.Eq(req.TargetID)).
								UpdateSimple(tx.Target.Method.Value(v.API.Method))
							if err != nil {
								log.Logger.Error("同步method失败，err:", err)
							}
						}

						if vvv == consts.TargetSyncUrlID {
							apiDetail.URL = v.API.URL
						}

						if vvv == consts.TargetSyncCookieID {
							cookieTemp := rao.Cookie{}
							parameterTemp := make([]rao.Parameter, 0, 20)
							for _, vvvv := range v.API.Request.Cookie.Parameter {
								temp := rao.Parameter{
									IsChecked:   vvvv.IsChecked,
									Type:        vvvv.Type,
									Key:         vvvv.Key,
									Value:       vvvv.Value,
									NotNull:     vvvv.NotNull,
									Description: vvvv.Description,
									FileBase64:  vvvv.FileBase64,
									FieldType:   vvvv.FieldType,
								}
								parameterTemp = append(parameterTemp, temp)
							}
							cookieTemp.Parameter = parameterTemp
							cookieData, err := bson.Marshal(cookieTemp)
							if err != nil {
								log.Logger.Info("拉取cookie失败 err:", err)
							}
							apiDetail.Cookie = cookieData
						}

						if vvv == consts.TargetSyncHeaderID {
							headerTemp := rao.Header{}
							parameterTemp := make([]rao.Parameter, 0, 20)
							for _, vvvv := range v.API.Request.Header.Parameter {
								temp := rao.Parameter{
									IsChecked:   vvvv.IsChecked,
									Type:        vvvv.Type,
									Key:         vvvv.Key,
									Value:       vvvv.Value,
									NotNull:     vvvv.NotNull,
									Description: vvvv.Description,
									FileBase64:  vvvv.FileBase64,
									FieldType:   vvvv.FieldType,
								}
								parameterTemp = append(parameterTemp, temp)
							}
							headerTemp.Parameter = parameterTemp
							headerData, err := bson.Marshal(&headerTemp)
							if err != nil {
								log.Logger.Info("拉取header失败 err:", err)
							}
							apiDetail.Header = headerData
						}

						if vvv == consts.TargetSyncQueryID {
							queryTemp := rao.Query{}
							parameterTemp := make([]rao.Parameter, 0, 20)
							for _, vvvv := range v.API.Request.Query.Parameter {
								temp := rao.Parameter{
									IsChecked:   vvvv.IsChecked,
									Type:        vvvv.Type,
									Key:         vvvv.Key,
									Value:       vvvv.Value,
									NotNull:     vvvv.NotNull,
									Description: vvvv.Description,
									FileBase64:  vvvv.FileBase64,
									FieldType:   vvvv.FieldType,
								}
								parameterTemp = append(parameterTemp, temp)
							}
							queryTemp.Parameter = parameterTemp
							queryData, err := bson.Marshal(queryTemp)
							if err != nil {
								log.Logger.Info("拉取query失败 err:", err)
							}
							apiDetail.Query = queryData
						}

						if vvv == consts.TargetSyncBodyID {
							bodyData, err := bson.Marshal(v.API.Request.Body)
							if err != nil {
								log.Logger.Info("拉取body失败 err:", err)
							}

							apiDetail.Body = bodyData
						}

						if vvv == consts.TargetSyncAuthID {
							authData, err := bson.Marshal(v.API.Request.Auth)
							if err != nil {
								log.Logger.Info("拉取auth失败 err:", err)
							}
							apiDetail.Auth = authData
						}

						if vvv == consts.TargetSyncAssertID {
							assertTemp := mao.Assert{}
							for _, assertInfo := range v.API.Request.Assert {
								temp := rao.Assert{
									ResponseType: assertInfo.ResponseType,
									Var:          assertInfo.Var,
									Compare:      assertInfo.Compare,
									Val:          assertInfo.Val,
									IsChecked:    assertInfo.IsChecked,
									Index:        assertInfo.Index,
								}
								assertTemp.Assert = append(assertTemp.Assert, temp)
							}

							assertData, err := bson.Marshal(assertTemp)
							if err != nil {
								log.Logger.Info("拉取assert失败 err:", err)
							}
							apiDetail.Assert = assertData
						}

						if vvv == consts.TargetSyncRegexID {
							regexTemp := mao.Regex{}
							for _, regexInfo := range v.API.Request.Regex {
								temp := rao.Regex{
									IsChecked: regexInfo.IsChecked,
									Type:      regexInfo.Type,
									Var:       regexInfo.Var,
									Val:       regexInfo.Val,
									Express:   regexInfo.Express,
									Index:     regexInfo.Index,
								}
								regexTemp.Regex = append(regexTemp.Regex, temp)
							}

							regexData, err := bson.Marshal(regexTemp)
							if err != nil {
								log.Logger.Info("拉取regex失败 err:", err)
							}
							apiDetail.Regex = regexData
						}

						if vvv == consts.TargetSyncConfigID {
							httpApiSetupTemp := mao.HttpApiSetup{
								IsRedirects:         v.API.Request.HttpApiSetup.IsRedirects,
								RedirectsNum:        v.API.Request.HttpApiSetup.RedirectsNum,
								ReadTimeOut:         v.API.Request.HttpApiSetup.ReadTimeOut,
								WriteTimeOut:        v.API.Request.HttpApiSetup.WriteTimeOut,
								ClientName:          v.API.Request.HttpApiSetup.ClientName,
								KeepAlive:           v.API.Request.HttpApiSetup.KeepAlive,
								MaxIdleConnDuration: v.API.Request.HttpApiSetup.MaxIdleConnDuration,
								MaxConnPerHost:      v.API.Request.HttpApiSetup.MaxConnPerHost,
								UserAgent:           v.API.Request.HttpApiSetup.UserAgent,
								MaxConnWaitTimeout:  v.API.Request.HttpApiSetup.MaxConnWaitTimeout,
							}
							apiDetail.HttpApiSetup = httpApiSetupTemp
						}
					}

					// 更新数据
					update := bson.M{"$set": apiDetail}
					_, err = collectionApi.UpdateOne(ctx, bson.D{{"target_id", req.TargetID}}, update)
					if err != nil {
						log.Logger.Error("拉取数据失败，err:", err)
						return err
					}
					break
				}
			}
		}

		// 保存历史记录
		saveHistoryParam, err := AssembleSaveHistoryParam(ctx, req.TargetID)
		if err == nil {
			_, _, err = SaveTargetHistoryRecord(ctx, saveHistoryParam)
			if err != nil {
				log.Logger.Info("同步数据--保存历史记录失败，err:", err)
			}
		}
	}
	return nil
}

func AssembleSaveHistoryParam(ctx *gin.Context, targetID string) (*rao.SaveTargetReq, error) {
	tx := dal.GetQuery().Target
	targetInfo, err := tx.WithContext(ctx).Where(tx.TargetID.Eq(targetID)).First()
	if err != nil {
		return nil, err
	}

	// 查询target详情
	// 查询测试对象详情
	apiDetail := mao.API{}
	collectionApi := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectAPI)
	err = collectionApi.FindOne(ctx, bson.D{{"target_id", targetID}}).Decode(&apiDetail)
	if err != nil {
		return nil, err
	}

	auth := rao.Auth{}
	err = bson.Unmarshal(apiDetail.Auth, &auth)
	if err != nil {
		return nil, err
	}

	body := rao.Body{}
	err = bson.Unmarshal(apiDetail.Body, &body)
	if err != nil {
		return nil, err
	}

	header := rao.Header{}
	err = bson.Unmarshal(apiDetail.Header, &header)
	if err != nil {
		return nil, err
	}

	queryData := rao.Query{}
	err = bson.Unmarshal(apiDetail.Query, &queryData)
	if err != nil {
		return nil, err
	}

	cookie := rao.Cookie{}
	err = bson.Unmarshal(apiDetail.Cookie, &queryData)
	if err != nil {
		return nil, err
	}

	assert := mao.Assert{}
	err = bson.Unmarshal(apiDetail.Assert, &assert)
	if err != nil {
		return nil, err
	}
	assertArr := make([]rao.Assert, 0, len(assert.Assert))
	for _, v := range assert.Assert {
		temp := rao.Assert{
			ResponseType: v.ResponseType,
			Var:          v.Var,
			Compare:      v.Compare,
			Val:          v.Val,
			IsChecked:    v.IsChecked,
			Index:        v.Index,
		}
		assertArr = append(assertArr, temp)
	}

	regex := mao.Regex{}
	err = bson.Unmarshal(apiDetail.Regex, &regex)
	if err != nil {
		return nil, err
	}
	regexArr := make([]rao.Regex, 0, len(regex.Regex))
	for _, v := range regex.Regex {
		temp := rao.Regex{
			IsChecked: v.IsChecked,
			Type:      v.Type,
			Var:       v.Var,
			Val:       v.Val,
			Express:   v.Express,
			Index:     v.Index,
		}
		regexArr = append(regexArr, temp)
	}

	HttpApiSetup := rao.HttpApiSetup{
		IsRedirects:         apiDetail.HttpApiSetup.IsRedirects,
		RedirectsNum:        apiDetail.HttpApiSetup.RedirectsNum,
		ReadTimeOut:         apiDetail.HttpApiSetup.ReadTimeOut,
		WriteTimeOut:        apiDetail.HttpApiSetup.WriteTimeOut,
		ClientName:          apiDetail.HttpApiSetup.ClientName,
		KeepAlive:           apiDetail.HttpApiSetup.KeepAlive,
		MaxIdleConnDuration: apiDetail.HttpApiSetup.MaxIdleConnDuration,
		MaxConnPerHost:      apiDetail.HttpApiSetup.MaxConnPerHost,
		UserAgent:           apiDetail.HttpApiSetup.UserAgent,
		MaxConnWaitTimeout:  apiDetail.HttpApiSetup.MaxConnWaitTimeout,
	}

	request := rao.Request{
		PreUrl:       apiDetail.EnvInfo.PreUrl,
		URL:          apiDetail.URL,
		Method:       targetInfo.Method,
		Description:  targetInfo.Description,
		Auth:         auth,
		Body:         body,
		Header:       header,
		Query:        queryData,
		Cookie:       cookie,
		Assert:       assertArr,
		Regex:        regexArr,
		HttpApiSetup: HttpApiSetup,
	}

	envInfoTemp := rao.EnvInfo{
		EnvID:       apiDetail.EnvInfo.EnvID,
		EnvName:     apiDetail.EnvInfo.EnvName,
		ServiceID:   apiDetail.EnvInfo.ServiceID,
		ServiceName: apiDetail.EnvInfo.ServiceName,
		PreUrl:      apiDetail.EnvInfo.PreUrl,
		DatabaseID:  apiDetail.EnvInfo.DatabaseID,
		ServerName:  apiDetail.EnvInfo.ServerName,
	}

	res := &rao.SaveTargetReq{
		TargetID:    targetID,
		ParentID:    targetInfo.ParentID,
		TeamID:      targetInfo.TeamID,
		Name:        targetInfo.Name,
		Method:      targetInfo.Method,
		URL:         apiDetail.URL,
		Sort:        targetInfo.Sort,
		TypeSort:    targetInfo.TypeSort,
		Request:     request,
		Source:      targetInfo.Source,
		Version:     targetInfo.Version,
		Description: targetInfo.Description,
		EnvInfo:     envInfoTemp,
		// 为了导入接口而新增的一些字段
		TargetType: targetInfo.TargetType,
	}
	return res, nil
}

func ExecSyncDataFlowToTarget(ctx *gin.Context, req *rao.ExecSyncDataReq) error {
	err := dal.GetQuery().Transaction(func(tx *query.Query) error {
		// 查询测试对象基本信息
		targetInfo, err := tx.Target.WithContext(ctx).Where(tx.Target.TargetID.Eq(req.SyncApiInfo[0].NodeID)).First()
		if err != nil {
			return err
		}

		// 查询接口详情数据
		apiDetail := mao.API{}
		collectionApi := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectAPI)
		err = collectionApi.FindOne(ctx, bson.D{{"target_id", req.SyncApiInfo[0].NodeID}}).Decode(&apiDetail)
		if err != nil {
			return err
		}

		if req.Source == consts.TargetSourceScene || req.Source == consts.TargetSourcePlan ||
			req.Source == consts.TargetSourceAutoPlan {
			collectionFlow := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectFlow)
			// 查询基本的flow
			baseFlow := mao.Flow{}
			err := collectionFlow.FindOne(ctx, bson.D{{"scene_id", req.SceneID}}).Decode(&baseFlow)
			if err != nil {
				return err
			}

			baseNodes := mao.Node{}
			err = bson.Unmarshal(baseFlow.Nodes, &baseNodes)
			if err != nil {
				return err
			}

			// 获取基本node节点的信息
			baseNodeInfo := rao.Node{}
			for _, v := range baseNodes.Nodes {
				if v.ID == req.NodeID {
					baseNodeInfo = v
				}
			}

			if req.SyncType == consts.TargetSyncDataTypePush { // 推送
				for _, val := range req.SyncContent {
					if val == consts.TargetSyncMethodID {
						_, err = tx.Target.WithContext(ctx).Where(tx.Target.TargetID.Eq(req.SyncApiInfo[0].NodeID)).
							UpdateSimple(tx.Target.Method.Value(baseNodeInfo.API.Method))
						if err != nil {
							log.Logger.Error("同步method失败，err:", err)
						}
					}

					if val == consts.TargetSyncUrlID {
						apiDetail.URL = baseNodeInfo.API.URL
					}

					if val == consts.TargetSyncCookieID {
						cookieData, err := bson.Marshal(baseNodeInfo.API.Request.Cookie)
						if err != nil {
							log.Logger.Info("推送cookie失败 err:", err)
						}
						apiDetail.Cookie = cookieData
					}

					if val == consts.TargetSyncHeaderID {
						headerData, err := bson.Marshal(baseNodeInfo.API.Request.Header)
						if err != nil {
							log.Logger.Info("推送header失败 err:", err)
						}
						apiDetail.Header = headerData
					}

					if val == consts.TargetSyncQueryID {
						queryData, err := bson.Marshal(baseNodeInfo.API.Request.Query)
						if err != nil {
							log.Logger.Info("推送query失败 err:", err)
						}
						apiDetail.Query = queryData
					}

					if val == consts.TargetSyncBodyID {
						bodyData, err := bson.Marshal(baseNodeInfo.API.Request.Body)
						if err != nil {
							log.Logger.Info("推送body失败 err:", err)
						}

						apiDetail.Body = bodyData
					}

					if val == consts.TargetSyncAuthID {
						authData, err := bson.Marshal(baseNodeInfo.API.Request.Auth)
						if err != nil {
							log.Logger.Info("推送auth失败 err:", err)
						}
						apiDetail.Auth = authData
					}

					if val == consts.TargetSyncAssertID {
						assertData, err := bson.Marshal(&mao.Assert{Assert: baseNodeInfo.API.Request.Assert})
						if err != nil {
							log.Logger.Info("推送assert失败 err:", err)
						}
						apiDetail.Assert = assertData
					}

					if val == consts.TargetSyncRegexID {
						regexData, err := bson.Marshal(mao.Regex{Regex: baseNodeInfo.API.Request.Regex})
						if err != nil {
							log.Logger.Info("推送regex失败 err:", err)
						}
						apiDetail.Regex = regexData
					}

					if val == consts.TargetSyncConfigID {
						httpApiSetupTemp := mao.HttpApiSetup{
							IsRedirects:         baseNodeInfo.API.Request.HttpApiSetup.IsRedirects,
							RedirectsNum:        baseNodeInfo.API.Request.HttpApiSetup.RedirectsNum,
							ReadTimeOut:         baseNodeInfo.API.Request.HttpApiSetup.ReadTimeOut,
							WriteTimeOut:        baseNodeInfo.API.Request.HttpApiSetup.WriteTimeOut,
							ClientName:          baseNodeInfo.API.Request.HttpApiSetup.ClientName,
							KeepAlive:           baseNodeInfo.API.Request.HttpApiSetup.KeepAlive,
							MaxIdleConnDuration: baseNodeInfo.API.Request.HttpApiSetup.MaxIdleConnDuration,
							MaxConnPerHost:      baseNodeInfo.API.Request.HttpApiSetup.MaxConnPerHost,
							UserAgent:           baseNodeInfo.API.Request.HttpApiSetup.UserAgent,
							MaxConnWaitTimeout:  baseNodeInfo.API.Request.HttpApiSetup.MaxConnWaitTimeout,
						}
						apiDetail.HttpApiSetup = httpApiSetupTemp
					}
				}
				// 更新数据
				update := bson.M{"$set": apiDetail}
				_, err = collectionApi.UpdateOne(ctx, bson.D{{"target_id", req.SyncApiInfo[0].NodeID}}, update)
				if err != nil {
					log.Logger.Error("拉取数据失败，err:", err)
					return err
				}

				// 保存历史记录
				saveHistoryParam, err := AssembleSaveHistoryParam(ctx, req.TargetID)
				if err == nil {
					_, _, err = SaveTargetHistoryRecord(ctx, saveHistoryParam)
					if err != nil {
						log.Logger.Info("同步数据--保存历史记录失败，err:", err)
					}
				}

			} else { // 拉取
				for key, val := range baseNodes.Nodes {
					if val.ID == req.NodeID {
						for _, val2 := range req.SyncContent {
							if val2 == consts.TargetSyncMethodID {
								baseNodes.Nodes[key].API.Method = targetInfo.Method
							}

							if val2 == consts.TargetSyncUrlID {
								baseNodes.Nodes[key].API.URL = apiDetail.URL
								baseNodes.Nodes[key].API.Request.URL = apiDetail.URL
							}

							if val2 == consts.TargetSyncCookieID {
								cookie := mao.Cookie{}
								err = bson.Unmarshal(apiDetail.Cookie, &cookie)
								if err != nil {
									log.Logger.Error("解析cookie失败，err:", err)
									continue
								}
								cookieTemp := rao.Cookie{}
								for _, info := range cookie.Parameter {
									temp := rao.Parameter{
										IsChecked:   info.IsChecked,
										Type:        info.Type,
										Key:         info.Key,
										Value:       info.Value,
										NotNull:     info.NotNull,
										Description: info.Description,
										FileBase64:  info.FileBase64,
										FieldType:   info.FieldType,
									}
									cookieTemp.Parameter = append(cookieTemp.Parameter, temp)
								}
								baseNodes.Nodes[key].API.Request.Cookie = cookieTemp
							}

							if val2 == consts.TargetSyncHeaderID {
								header := mao.Header{}
								err = bson.Unmarshal(apiDetail.Header, &header)
								if err != nil {
									log.Logger.Error("解析header失败，err:", err)
									continue
								}
								headerTemp := rao.Header{}
								for _, info := range header.Parameter {
									temp := rao.Parameter{
										IsChecked:   info.IsChecked,
										Type:        info.Type,
										Key:         info.Key,
										Value:       info.Value,
										NotNull:     info.NotNull,
										Description: info.Description,
										FileBase64:  info.FileBase64,
										FieldType:   info.FieldType,
									}
									headerTemp.Parameter = append(headerTemp.Parameter, temp)
								}
								baseNodes.Nodes[key].API.Request.Header = headerTemp
							}

							if val2 == consts.TargetSyncQueryID {
								queryData := mao.Query{}
								err = bson.Unmarshal(apiDetail.Query, &queryData)
								if err != nil {
									log.Logger.Error("解析query失败，err:", err)
									continue
								}
								queryTemp := rao.Query{}
								for _, info := range queryData.Parameter {
									temp := rao.Parameter{
										IsChecked:   info.IsChecked,
										Type:        info.Type,
										Key:         info.Key,
										Value:       info.Value,
										NotNull:     info.NotNull,
										Description: info.Description,
										FileBase64:  info.FileBase64,
										FieldType:   info.FieldType,
									}
									queryTemp.Parameter = append(queryTemp.Parameter, temp)
								}
								baseNodes.Nodes[key].API.Request.Query = queryTemp
							}

							if val2 == consts.TargetSyncBodyID {
								body := rao.Body{}
								err = bson.Unmarshal(apiDetail.Body, &body)
								if err != nil {
									log.Logger.Error("解析body失败，err:", err)
									continue
								}
								baseNodes.Nodes[key].API.Request.Body = body
							}

							if val2 == consts.TargetSyncAuthID {
								auth := rao.Auth{}
								err = bson.Unmarshal(apiDetail.Auth, &auth)
								if err != nil {
									log.Logger.Error("解析auth失败，err:", err)
									continue
								}
								baseNodes.Nodes[key].API.Request.Auth = auth
							}

							if val2 == consts.TargetSyncAssertID {
								assert := mao.Assert{}
								err = bson.Unmarshal(apiDetail.Assert, &assert)
								if err != nil {
									log.Logger.Error("解析assert失败，err:", err)
									continue
								}
								assertTemp := make([]rao.Assert, 0, len(assert.Assert))
								for _, assertInfo := range assert.Assert {
									temp := rao.Assert{
										ResponseType: assertInfo.ResponseType,
										Var:          assertInfo.Var,
										Compare:      assertInfo.Compare,
										Val:          assertInfo.Val,
										IsChecked:    assertInfo.IsChecked,
										Index:        assertInfo.Index,
									}
									assertTemp = append(assertTemp, temp)
								}
								baseNodes.Nodes[key].API.Request.Assert = assertTemp
							}

							if val2 == consts.TargetSyncRegexID {
								regex := mao.Regex{}
								err = bson.Unmarshal(apiDetail.Regex, &regex)
								if err != nil {
									log.Logger.Error("解析assert失败，err:", err)
									continue
								}
								regexTemp := make([]rao.Regex, 0, len(regex.Regex))
								for _, regexInfo := range regex.Regex {
									temp := rao.Regex{
										IsChecked: regexInfo.IsChecked,
										Type:      regexInfo.Type,
										Var:       regexInfo.Var,
										Val:       regexInfo.Val,
										Express:   regexInfo.Express,
										Index:     regexInfo.Index,
									}
									regexTemp = append(regexTemp, temp)
								}
								baseNodes.Nodes[key].API.Request.Regex = regexTemp
							}

							if val2 == consts.TargetSyncConfigID {
								httpApiSetupTemp := rao.HttpApiSetup{
									IsRedirects:         apiDetail.HttpApiSetup.IsRedirects,
									RedirectsNum:        apiDetail.HttpApiSetup.RedirectsNum,
									ReadTimeOut:         apiDetail.HttpApiSetup.ReadTimeOut,
									WriteTimeOut:        apiDetail.HttpApiSetup.WriteTimeOut,
									ClientName:          apiDetail.HttpApiSetup.ClientName,
									KeepAlive:           apiDetail.HttpApiSetup.KeepAlive,
									MaxIdleConnDuration: apiDetail.HttpApiSetup.MaxIdleConnDuration,
									MaxConnPerHost:      apiDetail.HttpApiSetup.MaxConnPerHost,
									UserAgent:           apiDetail.HttpApiSetup.UserAgent,
									MaxConnWaitTimeout:  apiDetail.HttpApiSetup.MaxConnWaitTimeout,
								}
								baseNodes.Nodes[key].API.Request.HttpApiSetup = httpApiSetupTemp
							}
						}
					}
				}

				nodesTemp, err := bson.Marshal(baseNodes)
				if err != nil {
					return err
				}
				baseFlow.Nodes = nodesTemp

				// 更新数据
				update := bson.M{"$set": baseFlow}
				_, err = collectionFlow.UpdateOne(ctx, bson.D{{"scene_id", req.SceneID}}, update)
				if err != nil {
					log.Logger.Error("拉取数据失败，err:", err)
					return err
				}
			}
		} else { // 用例
			collectionCaseFlow := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectSceneCaseFlow)
			// 查询基本的flow
			baseFlow := mao.SceneCaseFlow{}
			err := collectionCaseFlow.FindOne(ctx, bson.D{{"scene_case_id", req.CaseID}}).Decode(&baseFlow)
			if err != nil {
				return err
			}

			baseNodes := mao.Node{}
			err = bson.Unmarshal(baseFlow.Nodes, &baseNodes)
			if err != nil {
				return err
			}

			// 获取基本node节点的信息
			baseNodeInfo := rao.Node{}
			for _, v := range baseNodes.Nodes {
				if v.ID == req.NodeID {
					baseNodeInfo = v
				}
			}

			if req.SyncType == consts.TargetSyncDataTypePush { // 推送
				for _, val := range req.SyncContent {
					if val == consts.TargetSyncMethodID {
						_, err = tx.Target.WithContext(ctx).Where(tx.Target.TargetID.Eq(req.SyncApiInfo[0].NodeID)).
							UpdateSimple(tx.Target.Method.Value(baseNodeInfo.API.Method))
						if err != nil {
							log.Logger.Error("同步method失败，err:", err)
						}
					}

					if val == consts.TargetSyncUrlID {
						apiDetail.URL = baseNodeInfo.API.URL
					}

					if val == consts.TargetSyncCookieID {
						cookieData, err := bson.Marshal(baseNodeInfo.API.Request.Cookie)
						if err != nil {
							log.Logger.Info("推送cookie失败 err:", err)
						}
						apiDetail.Cookie = cookieData
					}

					if val == consts.TargetSyncHeaderID {
						headerData, err := bson.Marshal(baseNodeInfo.API.Request.Header)
						if err != nil {
							log.Logger.Info("推送header失败 err:", err)
						}
						apiDetail.Header = headerData
					}

					if val == consts.TargetSyncQueryID {
						queryData, err := bson.Marshal(baseNodeInfo.API.Request.Query)
						if err != nil {
							log.Logger.Info("推送query失败 err:", err)
						}
						apiDetail.Query = queryData
					}

					if val == consts.TargetSyncBodyID {
						bodyData, err := bson.Marshal(baseNodeInfo.API.Request.Body)
						if err != nil {
							log.Logger.Info("推送body失败 err:", err)
						}

						apiDetail.Body = bodyData
					}

					if val == consts.TargetSyncAuthID {
						authData, err := bson.Marshal(baseNodeInfo.API.Request.Auth)
						if err != nil {
							log.Logger.Info("推送auth失败 err:", err)
						}
						apiDetail.Auth = authData
					}

					if val == consts.TargetSyncAssertID {
						assertData, err := bson.Marshal(&mao.Assert{Assert: baseNodeInfo.API.Request.Assert})
						if err != nil {
							log.Logger.Info("推送assert失败 err:", err)
						}
						apiDetail.Assert = assertData
					}

					if val == consts.TargetSyncRegexID {
						regexData, err := bson.Marshal(mao.Regex{Regex: baseNodeInfo.API.Request.Regex})
						if err != nil {
							log.Logger.Info("推送regex失败 err:", err)
						}
						apiDetail.Regex = regexData
					}

					if val == consts.TargetSyncConfigID {
						httpApiSetupTemp := mao.HttpApiSetup{
							IsRedirects:         baseNodeInfo.API.Request.HttpApiSetup.IsRedirects,
							RedirectsNum:        baseNodeInfo.API.Request.HttpApiSetup.RedirectsNum,
							ReadTimeOut:         baseNodeInfo.API.Request.HttpApiSetup.ReadTimeOut,
							WriteTimeOut:        baseNodeInfo.API.Request.HttpApiSetup.WriteTimeOut,
							ClientName:          baseNodeInfo.API.Request.HttpApiSetup.ClientName,
							KeepAlive:           baseNodeInfo.API.Request.HttpApiSetup.KeepAlive,
							MaxIdleConnDuration: baseNodeInfo.API.Request.HttpApiSetup.MaxIdleConnDuration,
							MaxConnPerHost:      baseNodeInfo.API.Request.HttpApiSetup.MaxConnPerHost,
							UserAgent:           baseNodeInfo.API.Request.HttpApiSetup.UserAgent,
							MaxConnWaitTimeout:  baseNodeInfo.API.Request.HttpApiSetup.MaxConnWaitTimeout,
						}
						apiDetail.HttpApiSetup = httpApiSetupTemp
					}
				}
				// 更新数据
				update := bson.M{"$set": apiDetail}
				_, err = collectionApi.UpdateOne(ctx, bson.D{{"target_id", req.SyncApiInfo[0].NodeID}}, update)
				if err != nil {
					log.Logger.Error("拉取数据失败，err:", err)
					return err
				}
			} else { // 拉取
				for key, val := range baseNodes.Nodes {
					if val.ID == req.SyncApiInfo[0].NodeID {
						for _, val2 := range req.SyncContent {
							if val2 == consts.TargetSyncMethodID {
								baseNodes.Nodes[key].API.Method = targetInfo.Method
							}

							if val2 == consts.TargetSyncUrlID {
								baseNodes.Nodes[key].API.URL = apiDetail.URL
								baseNodes.Nodes[key].API.Request.URL = apiDetail.URL
							}

							if val2 == consts.TargetSyncCookieID {
								cookie := rao.Cookie{}
								err = bson.Unmarshal(apiDetail.Cookie, &cookie)
								if err != nil {
									log.Logger.Error("解析cookie失败，err:", err)
									continue
								}
								baseNodes.Nodes[key].API.Request.Cookie = cookie
							}

							if val2 == consts.TargetSyncHeaderID {
								header := rao.Header{}
								err = bson.Unmarshal(apiDetail.Header, &header)
								if err != nil {
									log.Logger.Error("解析header失败，err:", err)
									continue
								}
								baseNodes.Nodes[key].API.Request.Header = header
							}

							if val2 == consts.TargetSyncQueryID {
								queryList := rao.Query{}
								err = bson.Unmarshal(apiDetail.Query, &queryList)
								if err != nil {
									log.Logger.Error("解析query失败，err:", err)
									continue
								}
								baseNodes.Nodes[key].API.Request.Query = queryList
							}

							if val2 == consts.TargetSyncBodyID {
								body := rao.Body{}
								err = bson.Unmarshal(apiDetail.Body, &body)
								if err != nil {
									log.Logger.Error("解析body失败，err:", err)
									continue
								}
								baseNodes.Nodes[key].API.Request.Body = body

							}

							if val2 == consts.TargetSyncAuthID {
								auth := rao.Auth{}
								err = bson.Unmarshal(apiDetail.Auth, &auth)
								if err != nil {
									log.Logger.Error("解析auth失败，err:", err)
									continue
								}
								baseNodes.Nodes[key].API.Request.Auth = auth
							}

							if val2 == consts.TargetSyncAssertID {
								assert := make([]rao.Assert, 0, 20)
								err = bson.Unmarshal(apiDetail.Assert, &assert)
								if err != nil {
									log.Logger.Error("解析assert失败，err:", err)
									continue
								}
								baseNodes.Nodes[key].API.Request.Assert = assert
							}

							if val2 == consts.TargetSyncRegexID {
								regex := make([]rao.Regex, 0, 20)
								err = bson.Unmarshal(apiDetail.Regex, &regex)
								if err != nil {
									log.Logger.Error("解析assert失败，err:", err)
									continue
								}
								baseNodes.Nodes[key].API.Request.Regex = regex
							}

							if val2 == consts.TargetSyncConfigID {
								httpApiSetupTemp := rao.HttpApiSetup{
									IsRedirects:         apiDetail.HttpApiSetup.IsRedirects,
									RedirectsNum:        apiDetail.HttpApiSetup.RedirectsNum,
									ReadTimeOut:         apiDetail.HttpApiSetup.ReadTimeOut,
									WriteTimeOut:        apiDetail.HttpApiSetup.WriteTimeOut,
									ClientName:          apiDetail.HttpApiSetup.ClientName,
									KeepAlive:           apiDetail.HttpApiSetup.KeepAlive,
									MaxIdleConnDuration: apiDetail.HttpApiSetup.MaxIdleConnDuration,
									MaxConnPerHost:      apiDetail.HttpApiSetup.MaxConnPerHost,
									UserAgent:           apiDetail.HttpApiSetup.UserAgent,
									MaxConnWaitTimeout:  apiDetail.HttpApiSetup.MaxConnWaitTimeout,
								}
								baseNodes.Nodes[key].API.Request.HttpApiSetup = httpApiSetupTemp
							}
						}
					}
				}

				nodesTemp, err := bson.Marshal(baseNodes)
				if err != nil {
					return err
				}
				baseFlow.Nodes = nodesTemp

				// 更新数据
				update := bson.M{"$set": baseFlow}
				_, err = collectionCaseFlow.UpdateOne(ctx, bson.D{{"scene_case_id", req.CaseID}}, update)
				if err != nil {
					log.Logger.Error("拉取数据失败，err:", err)
					return err
				}
			}
		}
		return nil
	})
	return err
}

//func ExecSyncDataFromFlowTarget(ctx *gin.Context, req *rao.ExecSyncDataReq) error {
//	err := dal.GetQuery().Transaction(func(tx *query.Query) error {
//		if req.Source == consts.TargetSourceScene || req.Source == consts.TargetSourcePlan ||
//			req.Source == consts.TargetSourceAutoPlan {
//
//			collectionFlow := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectFlow)
//			// 查询基本的flow
//			baseFlow := mao.Flow{}
//			err := collectionFlow.FindOne(ctx, bson.D{{"scene_id", req.SceneID}}).Decode(&baseFlow)
//			if err != nil {
//				return err
//			}
//
//			baseNodes := mao.Node{}
//			err = bson.Unmarshal(baseFlow.Nodes, &baseNodes)
//			if err != nil {
//				return err
//			}
//
//			// 获取基本node节点的信息
//			baseNodeInfo := rao.Node{}
//			for _, v := range baseNodes.Nodes {
//				if v.ID == req.NodeID {
//					baseNodeInfo = v
//				}
//			}
//
//			// 获取所有场景id和用例id
//			needSyncTargetID := ""
//			allScnenIds := make([]string, 0, len(req.SyncApiInfo))
//			allCaseIds := make([]string, 0, len(req.SyncApiInfo))
//
//			for _, v := range req.SyncApiInfo {
//				if v.Source == consts.TargetSourceApi { //测试对象
//					needSyncTargetID = v.NodeID
//				} else if v.Source == consts.TargetSourceScene || v.Source == consts.TargetSourcePlan ||
//					v.Source == consts.TargetSourceAutoPlan { // 场景，性能，自动化
//					allScnenIds = append(allScnenIds, v.SceneID)
//				} else { // 测试用例
//					allCaseIds = append(allCaseIds, v.CaseID)
//				}
//			}
//
//			// 查询所有场景flow信息
//			cursor, err := collectionFlow.Find(ctx, bson.D{{"scene_id", bson.D{{"$in", allScnenIds}}}})
//			if err != nil {
//				return err
//			}
//			allSceneFlows := make([]mao.Flow, 0, len(allScnenIds))
//			if err = cursor.All(ctx, &allSceneFlows); err != nil {
//				return err
//			}
//			allSceneFlowMap := make(map[string]mao.Flow, len(allSceneFlows))
//			allSceneNodeMap := make(map[string]rao.Node, 200)
//			for _, v := range allSceneFlows {
//				allSceneFlowMap[v.SceneID] = v
//
//				nodesTemp := mao.Node{}
//				err = bson.Unmarshal(v.Nodes, &nodesTemp)
//				if err != nil {
//					continue
//				}
//
//				for _, vv := range nodesTemp.Nodes {
//					allSceneNodeMap[v.SceneID+"#"+vv.ID] = vv
//				}
//			}
//
//			// 查询所有用例flow信息
//			collectionCaseFlow := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectSceneCaseFlow)
//			cursor, err = collectionCaseFlow.Find(ctx, bson.D{{"scene_case_id", bson.D{{"$in", allCaseIds}}}})
//			if err != nil {
//				return err
//			}
//			allCaseFlows := make([]mao.SceneCaseFlow, 0, len(allCaseIds))
//			if err = cursor.All(ctx, &allCaseFlows); err != nil {
//				return err
//			}
//
//			allCaseFlowMap := make(map[string]mao.SceneCaseFlow, len(allCaseFlows))
//			allCaseNodeMap := make(map[string]rao.Node, len(allCaseFlows))
//			for _, v := range allCaseFlows {
//				allCaseFlowMap[v.SceneCaseID] = v
//
//				nodesTemp := mao.SceneCaseFlowNode{}
//				err = bson.Unmarshal(v.Nodes, &nodesTemp)
//				if err != nil {
//					continue
//				}
//
//				for _, vv := range nodesTemp.Nodes {
//					allCaseNodeMap[v.SceneCaseID+"#"+vv.ID] = vv
//				}
//			}
//
//			// 判断是推送还是拉取
//			if req.SyncType == consts.TargetSyncDataTypePush { // 推送
//				for _, v := range req.SyncApiInfo {
//					if v.Source == consts.TargetSourceApi { // 测试对象
//						// 查询接口详情数据
//						apiDetail := mao.API{}
//						collectionApi := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectAPI)
//						err = collectionApi.FindOne(ctx, bson.D{{"target_id", v.NodeID}}).Decode(&apiDetail)
//						if err != nil {
//							return err
//						}
//
//						for _, val2 := range req.SyncContent {
//							if val2 == consts.TargetSyncMethodID {
//								_, err := tx.Target.WithContext(ctx).Where(tx.Target.TargetID.Eq(v.NodeID)).
//									UpdateSimple(tx.Target.Method.Value(baseNodeInfo.API.Method))
//								if err != nil {
//									log.Logger.Info("同步测试对象--推送到测试对象失败，err:", err)
//									continue
//								}
//							}
//
//							if val2 == consts.TargetSyncUrlID {
//								apiDetail.URL = baseNodeInfo.API.URL
//							}
//
//							if val2 == consts.TargetSyncCookieID {
//								cookieTemp, err := bson.Marshal(baseNodeInfo.API.Request.Cookie)
//								if err != nil {
//									log.Logger.Info("同步测试对象--压缩cookie失败,err:", err)
//								}
//								apiDetail.Cookie = cookieTemp
//							}
//
//							if val2 == consts.TargetSyncHeaderID {
//								headerTemp, err := bson.Marshal(baseNodeInfo.API.Request.Header)
//								if err != nil {
//									log.Logger.Info("同步测试对象--压缩header失败,err:", err)
//								}
//								apiDetail.Header = headerTemp
//							}
//
//							if val2 == consts.TargetSyncQueryID {
//								queryTemp, err := bson.Marshal(baseNodeInfo.API.Request.Query)
//								if err != nil {
//									log.Logger.Info("同步测试对象--压缩query失败,err:", err)
//								}
//								apiDetail.Query = queryTemp
//							}
//
//							if val2 == consts.TargetSyncBodyID {
//								bodyTemp, err := bson.Marshal(baseNodeInfo.API.Request.Body)
//								if err != nil {
//									log.Logger.Info("同步测试对象--压缩body失败,err:", err)
//								}
//								apiDetail.Body = bodyTemp
//
//							}
//
//							if val2 == consts.TargetSyncAuthID {
//								authTemp, err := bson.Marshal(baseNodeInfo.API.Request.Auth)
//								if err != nil {
//									log.Logger.Info("同步测试对象--压缩auth失败,err:", err)
//								}
//								apiDetail.Auth = authTemp
//							}
//
//							if val2 == consts.TargetSyncAssertID {
//								assertTemp, err := bson.Marshal(baseNodeInfo.API.Request.Assert)
//								if err != nil {
//									log.Logger.Info("同步测试对象--压缩assert失败,err:", err)
//								}
//								apiDetail.Assert = assertTemp
//							}
//
//							if val2 == consts.TargetSyncRegexID {
//								regexTemp, err := bson.Marshal(baseNodeInfo.API.Request.Regex)
//								if err != nil {
//									log.Logger.Info("同步测试对象--压缩regex失败,err:", err)
//								}
//								apiDetail.Regex = regexTemp
//							}
//
//							if val2 == consts.TargetSyncConfigID {
//								httpApiSetupTemp := mao.HttpApiSetup{
//									IsRedirects:         baseNodeInfo.API.Request.HttpApiSetup.IsRedirects,
//									RedirectsNum:        baseNodeInfo.API.Request.HttpApiSetup.RedirectsNum,
//									ReadTimeOut:         baseNodeInfo.API.Request.HttpApiSetup.ReadTimeOut,
//									WriteTimeOut:        baseNodeInfo.API.Request.HttpApiSetup.WriteTimeOut,
//									ClientName:          baseNodeInfo.API.Request.HttpApiSetup.ClientName,
//									KeepAlive:           baseNodeInfo.API.Request.HttpApiSetup.KeepAlive,
//									MaxIdleConnDuration: baseNodeInfo.API.Request.HttpApiSetup.MaxIdleConnDuration,
//									MaxConnPerHost:      baseNodeInfo.API.Request.HttpApiSetup.MaxConnPerHost,
//									UserAgent:           baseNodeInfo.API.Request.HttpApiSetup.UserAgent,
//									MaxConnWaitTimeout:  baseNodeInfo.API.Request.HttpApiSetup.MaxConnWaitTimeout,
//								}
//								apiDetail.HttpApiSetup = httpApiSetupTemp
//							}
//						}
//						// 更新数据
//						update := bson.M{"$set": apiDetail}
//						_, err = collectionApi.UpdateOne(ctx, bson.D{{"target_id", v.NodeID}}, update)
//						if err != nil {
//							log.Logger.Error("推送数据失败，err:", err)
//							return err
//						}
//					} else if v.Source == consts.TargetSourceScene || v.Source == consts.TargetSourcePlan ||
//						v.Source == consts.TargetSourceAutoPlan { // 场景，性能，自动化
//						flowInfo := mao.Flow{}
//						ok := false
//						if flowInfo, ok = allSceneFlowMap[v.SceneID]; !ok {
//							continue
//						}
//
//						nodes := mao.Node{}
//						err = bson.Unmarshal(flowInfo.Nodes, &nodes)
//						if err != nil {
//							continue
//						}
//
//						for key2, val2 := range nodes.Nodes {
//							if val2.ID == v.NodeID {
//								for _, vvv := range req.SyncContent {
//									if vvv == consts.TargetSyncMethodID {
//										nodes.Nodes[key2].API.Method = baseNodeInfo.API.Method
//									}
//
//									if vvv == consts.TargetSyncUrlID {
//										nodes.Nodes[key2].API.URL = baseNodeInfo.API.URL
//									}
//
//									if vvv == consts.TargetSyncCookieID {
//										nodes.Nodes[key2].API.Request.Cookie = baseNodeInfo.API.Request.Cookie
//									}
//
//									if vvv == consts.TargetSyncHeaderID {
//										nodes.Nodes[key2].API.Request.Header = baseNodeInfo.API.Request.Header
//									}
//
//									if vvv == consts.TargetSyncQueryID {
//										nodes.Nodes[key2].API.Request.Query = baseNodeInfo.API.Request.Query
//									}
//
//									if vvv == consts.TargetSyncBodyID {
//										nodes.Nodes[key2].API.Request.Body = baseNodeInfo.API.Request.Body
//
//									}
//
//									if vvv == consts.TargetSyncAuthID {
//										nodes.Nodes[key2].API.Request.Auth = baseNodeInfo.API.Request.Auth
//									}
//
//									if vvv == consts.TargetSyncAssertID {
//										nodes.Nodes[key2].API.Request.Assert = baseNodeInfo.API.Request.Assert
//									}
//
//									if vvv == consts.TargetSyncRegexID {
//										nodes.Nodes[key2].API.Request.Regex = baseNodeInfo.API.Request.Regex
//									}
//
//									if vvv == consts.TargetSyncConfigID {
//										nodes.Nodes[key2].API.Request.HttpApiSetup = baseNodeInfo.API.Request.HttpApiSetup
//									}
//								}
//							}
//						}
//
//						nodesData, err := bson.Marshal(mao.Node{Nodes: nodes.Nodes})
//						if err != nil {
//							log.Logger.Info("flow.nodes bson marshal err %w", err)
//						}
//						flowInfo.Nodes = nodesData
//						_, err = collectionFlow.UpdateOne(ctx, bson.D{{"scene_id", v.SceneID}}, bson.M{"$set": flowInfo})
//						if err != nil {
//							return err
//						}
//					} else { // 用例
//						caseFlowInfo := mao.SceneCaseFlow{}
//						ok := false
//						if caseFlowInfo, ok = allCaseFlowMap[v.CaseID]; !ok {
//							continue
//						}
//
//						nodes := mao.SceneCaseFlowNode{}
//						err = bson.Unmarshal(caseFlowInfo.Nodes, &nodes)
//						if err != nil {
//							continue
//						}
//
//						for key2, val2 := range nodes.Nodes {
//							if val2.ID == v.NodeID {
//								for _, vvv := range req.SyncContent {
//									if vvv == consts.TargetSyncMethodID {
//										nodes.Nodes[key2].API.Method = baseNodeInfo.API.Method
//									}
//
//									if vvv == consts.TargetSyncUrlID {
//										nodes.Nodes[key2].API.URL = baseNodeInfo.API.URL
//									}
//
//									if vvv == consts.TargetSyncCookieID {
//										nodes.Nodes[key2].API.Request.Cookie = baseNodeInfo.API.Request.Cookie
//									}
//
//									if vvv == consts.TargetSyncHeaderID {
//										nodes.Nodes[key2].API.Request.Header = baseNodeInfo.API.Request.Header
//									}
//
//									if vvv == consts.TargetSyncQueryID {
//										nodes.Nodes[key2].API.Request.Query = baseNodeInfo.API.Request.Query
//									}
//
//									if vvv == consts.TargetSyncBodyID {
//										nodes.Nodes[key2].API.Request.Body = baseNodeInfo.API.Request.Body
//
//									}
//
//									if vvv == consts.TargetSyncAuthID {
//										nodes.Nodes[key2].API.Request.Auth = baseNodeInfo.API.Request.Auth
//									}
//
//									if vvv == consts.TargetSyncAssertID {
//										nodes.Nodes[key2].API.Request.Assert = baseNodeInfo.API.Request.Assert
//									}
//
//									if vvv == consts.TargetSyncRegexID {
//										nodes.Nodes[key2].API.Request.Regex = baseNodeInfo.API.Request.Regex
//									}
//
//									if vvv == consts.TargetSyncConfigID {
//										nodes.Nodes[key2].API.Request.HttpApiSetup = baseNodeInfo.API.Request.HttpApiSetup
//									}
//								}
//							}
//						}
//
//						nodesData, err := bson.Marshal(mao.SceneCaseFlowNode{Nodes: nodes.Nodes})
//						if err != nil {
//							log.Logger.Info("压缩用例flow失败 err:", err)
//						}
//						caseFlowInfo.Nodes = nodesData
//						_, err = collectionFlow.UpdateOne(ctx, bson.D{{"scene_case_id", v.CaseID}}, bson.M{"$set": caseFlowInfo})
//						if err != nil {
//							return err
//						}
//					}
//				}
//			} else { // 拉取
//				if req.SyncApiInfo[0].Source == consts.TargetSourceApi { // 测试对象
//					targetDetail, err := tx.Target.WithContext(ctx).Where(tx.Target.TargetID.Eq(req.SyncApiInfo[0].NodeID)).First()
//					if err != nil {
//						log.Logger.Info("拉取数据--查询接口基本信息失败，err:", err)
//					}
//					// 查询接口详情数据
//					apiDetail := mao.API{}
//					collectionApi := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectAPI)
//					err = collectionApi.FindOne(ctx, bson.D{{"target_id", req.SyncApiInfo[0].NodeID}}).Decode(&apiDetail)
//					if err != nil {
//						return err
//					}
//
//					nodes := mao.Node{}
//					err = bson.Unmarshal(baseFlow.Nodes, &nodes)
//					if err != nil {
//						log.Logger.Info("基本flow解析失败,err:", err)
//					}
//
//					for key, val := range nodes.Nodes {
//						if val.ID == req.NodeID {
//							for _, val2 := range req.SyncContent {
//								if val2 == consts.TargetSyncMethodID {
//									nodes.Nodes[key].API.Method = targetDetail.Method
//								}
//
//								if val2 == consts.TargetSyncUrlID {
//									nodes.Nodes[key].API.URL = apiDetail.URL
//								}
//
//								if val2 == consts.TargetSyncCookieID {
//									cookie := rao.Cookie{}
//									err = bson.Unmarshal(apiDetail.Cookie, &cookie)
//									if err != nil {
//										log.Logger.Error("解析cookie失败，err:", err)
//										continue
//									}
//									nodes.Nodes[key].API.Request.Cookie = cookie
//								}
//
//								if val2 == consts.TargetSyncHeaderID {
//									header := rao.Header{}
//									err = bson.Unmarshal(apiDetail.Header, &header)
//									if err != nil {
//										log.Logger.Error("解析header失败，err:", err)
//										continue
//									}
//									nodes.Nodes[key].API.Request.Header = header
//								}
//
//								if val2 == consts.TargetSyncQueryID {
//									queryList := rao.Query{}
//									err = bson.Unmarshal(apiDetail.Query, &queryList)
//									if err != nil {
//										log.Logger.Error("解析query失败，err:", err)
//										continue
//									}
//									nodes.Nodes[key].API.Request.Query = queryList
//								}
//
//								if val2 == consts.TargetSyncBodyID {
//									body := rao.Body{}
//									err = bson.Unmarshal(apiDetail.Body, &body)
//									if err != nil {
//										log.Logger.Error("解析body失败，err:", err)
//										continue
//									}
//									nodes.Nodes[key].API.Request.Body = body
//
//								}
//
//								if val2 == consts.TargetSyncAuthID {
//									auth := rao.Auth{}
//									err = bson.Unmarshal(apiDetail.Auth, &auth)
//									if err != nil {
//										log.Logger.Error("解析auth失败，err:", err)
//										continue
//									}
//									nodes.Nodes[key].API.Request.Auth = auth
//								}
//
//								if val2 == consts.TargetSyncAssertID {
//									assert := make([]rao.Assert, 0, 20)
//									err = bson.Unmarshal(apiDetail.Assert, &assert)
//									if err != nil {
//										log.Logger.Error("解析assert失败，err:", err)
//										continue
//									}
//									nodes.Nodes[key].API.Request.Assert = assert
//								}
//
//								if val2 == consts.TargetSyncRegexID {
//									regex := make([]rao.Regex, 0, 20)
//									err = bson.Unmarshal(apiDetail.Regex, &regex)
//									if err != nil {
//										log.Logger.Error("解析assert失败，err:", err)
//										continue
//									}
//									nodes.Nodes[key].API.Request.Regex = regex
//								}
//
//								if val2 == consts.TargetSyncConfigID {
//									httpApiSetupTemp := rao.HttpApiSetup{
//										IsRedirects:         apiDetail.HttpApiSetup.IsRedirects,
//										RedirectsNum:        apiDetail.HttpApiSetup.RedirectsNum,
//										ReadTimeOut:         apiDetail.HttpApiSetup.ReadTimeOut,
//										WriteTimeOut:        apiDetail.HttpApiSetup.WriteTimeOut,
//										ClientName:          apiDetail.HttpApiSetup.ClientName,
//										KeepAlive:           apiDetail.HttpApiSetup.KeepAlive,
//										MaxIdleConnDuration: apiDetail.HttpApiSetup.MaxIdleConnDuration,
//										MaxConnPerHost:      apiDetail.HttpApiSetup.MaxConnPerHost,
//										UserAgent:           apiDetail.HttpApiSetup.UserAgent,
//										MaxConnWaitTimeout:  apiDetail.HttpApiSetup.MaxConnWaitTimeout,
//									}
//									nodes.Nodes[key].API.Request.HttpApiSetup = httpApiSetupTemp
//								}
//							}
//						}
//					}
//
//					nodesData, err := bson.Marshal(mao.Node{Nodes: nodes.Nodes})
//					if err != nil {
//						log.Logger.Info("flow.nodes bson marshal err %w", err)
//					}
//					baseFlow.Nodes = nodesData
//					_, err = collectionFlow.UpdateOne(ctx, bson.D{{"scene_id", req.SyncApiInfo[0].SceneID}}, bson.M{"$set": baseFlow})
//					if err != nil {
//						return err
//					}
//				} else if req.Source == consts.TargetSourceScene ||
//					req.Source == consts.TargetSourcePlan || req.Source == consts.TargetSourceAutoPlan {
//					sceneFlow := mao.Flow{}
//					ok := false
//					if sceneFlow, ok = allSceneFlowMap[req.SyncApiInfo[0].SceneID]; !ok {
//						log.Logger.Info("拉取数据失败--没有找到拉取的接口信息")
//					}
//
//					nodes := mao.Node{}
//					err = bson.Unmarshal(sceneFlow.Nodes, &nodes)
//					if err != nil {
//						log.Logger.Info("拉取数据失败--没有找到拉取的接口信息")
//					}
//
//					needUpdateNode := rao.Node{}
//					for _, val := range nodes.Nodes {
//						if val.ID == req.SyncApiInfo[0].NodeID {
//							needUpdateNode = val
//
//							for _, val3 := range req.SyncContent {
//								if val3 == consts.TargetSyncMethodID {
//									needUpdateNode.API.Method = baseNodeInfo.API.Method
//								}
//
//								if val3 == consts.TargetSyncUrlID {
//									nodes.Nodes[key2].API.URL = baseNodeInfo.API.URL
//								}
//
//								if val3 == consts.TargetSyncCookieID {
//									nodes.Nodes[key2].API.Request.Cookie = baseNodeInfo.API.Request.Cookie
//								}
//
//								if val3 == consts.TargetSyncHeaderID {
//									nodes.Nodes[key2].API.Request.Header = baseNodeInfo.API.Request.Header
//								}
//
//								if val3 == consts.TargetSyncQueryID {
//									nodes.Nodes[key2].API.Request.Query = baseNodeInfo.API.Request.Query
//								}
//
//								if val3 == consts.TargetSyncBodyID {
//									nodes.Nodes[key2].API.Request.Body = baseNodeInfo.API.Request.Body
//
//								}
//
//								if val3 == consts.TargetSyncAuthID {
//									nodes.Nodes[key2].API.Request.Auth = baseNodeInfo.API.Request.Auth
//								}
//
//								if val3 == consts.TargetSyncAssertID {
//									nodes.Nodes[key2].API.Request.Assert = baseNodeInfo.API.Request.Assert
//								}
//
//								if val3 == consts.TargetSyncRegexID {
//									nodes.Nodes[key2].API.Request.Regex = baseNodeInfo.API.Request.Regex
//								}
//
//								if val3 == consts.TargetSyncConfigID {
//									nodes.Nodes[key2].API.Request.HttpApiSetup = baseNodeInfo.API.Request.HttpApiSetup
//								}
//							}
//
//						}
//					}
//
//				}
//
//			}
//
//		} else { // 基本同步接口来自用例
//
//		}
//
//		// 最后的返回
//		return nil
//	})
//	return err
//}
