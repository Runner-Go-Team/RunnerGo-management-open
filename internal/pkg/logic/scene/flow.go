package scene

import (
	"context"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/errno"
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
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/target"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/packer"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
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

func GetFlow(ctx context.Context, sceneID string) (rao.GetFlowResp, error) {
	res := rao.GetFlowResp{}

	tx := dal.GetQuery().Target
	targetInfo, err := tx.WithContext(ctx).Where(tx.TargetID.Eq(sceneID)).First()
	if err != nil {
		return res, err
	}
	res.TeamID = targetInfo.TeamID
	res.SceneID = targetInfo.TargetID
	res.Version = targetInfo.Version

	ret := mao.Flow{}
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectFlow)
	err = collection.FindOne(ctx, bson.D{{"scene_id", sceneID}}).Decode(&ret)
	if err != nil && err != mongo.ErrNoDocuments {
		return res, err
	}

	if err == mongo.ErrNoDocuments {
		return res, nil
	}

	return packer.TransMaoFlowToRaoGetFowResp(ret), nil
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
		if targetInfo.TargetType == consts.TargetTypeFolder { // 分组目录
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

			_, err = tx.Target.WithContext(ctx).Where(tx.Target.TargetID.In(caseIDs...),
				tx.Target.TargetType.Eq(consts.TargetTypeTestCase)).Delete()
			if err != nil {
				return err
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

			// 删除场景与计划的关系
			collection = dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectSceneCiteRelation)
			if req.Source == consts.TargetSourceScene { // 场景管理
				_, err = collection.DeleteMany(ctx, bson.D{{"old_scene_id", req.TargetID}})
				if err != nil {
					return err
				}
			} else {
				_, err = collection.DeleteMany(ctx, bson.D{{"new_scene_id", req.TargetID}})
				if err != nil {
					return err
				}
			}

			// 删除场景flow里面Node和接口的关系
			collection = dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectTargetCiteRelation)
			_, err = collection.DeleteMany(ctx, bson.D{{"scene_id", req.TargetID}})
			if err != nil {
				return err
			}

		}

		// 记录操作日志
		var operate int32 = 0
		if targetInfo.TargetType == consts.TargetTypeScene {
			operate = record.OperationOperateDeleteScene
		} else if targetInfo.TargetType == consts.TargetTypeFolder {
			operate = record.OperationOperateDeleteFolder
		}
		if err := record.InsertDelete(ctx, targetInfo.TeamID, userID, operate, targetInfo.Name); err != nil {
			return err
		}

		return nil
	})
}

func ChangeDisabledStatus(ctx *gin.Context, req *rao.ChangeDisabledStatusReq) error {
	tx := dal.GetQuery().Target
	_, err := tx.WithContext(ctx).Where(tx.TargetID.Eq(req.TargetID)).UpdateSimple(tx.IsDisabled.Value(req.IsDisabled))
	if err != nil {
		return fmt.Errorf("修改禁用状态失败")
	}
	return nil
}

func SendMysql(ctx context.Context, req *rao.SendMysqlReq) (string, error) {
	prepositions := mao.Preposition{}
	// 场景
	f := mao.Flow{}
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectFlow)
	err := collection.FindOne(ctx, bson.D{{"scene_id", req.SceneID}}).Decode(&f)
	if err != nil {
		return "", err
	}

	if err = bson.Unmarshal(f.Prepositions, &prepositions); err != nil {
		return "", err
	}

	// 获取上传的变量文件地址
	vi := dal.GetQuery().VariableImport
	vis, err := vi.WithContext(ctx).Where(vi.SceneID.Eq(req.SceneID)).Find()
	if err != nil {
		return "", err
	}

	fileList := make([]rao.FileList, 0, len(vis))
	for _, viInfo := range vis {
		fileList = append(fileList, rao.FileList{
			IsChecked: int64(viInfo.Status),
			Path:      viInfo.URL,
		})
	}

	// 获取全局变量
	globalVariable, err := target.GetGlobalVariable(ctx, req.TeamID)

	// 获取场景变量
	sceneVariable := rao.GlobalVariable{}
	collection = dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectSceneParam)
	cur, err := collection.Find(ctx, bson.D{{"scene_id", req.SceneID}})
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

	// 组装场景变量
	configurationData := rao.Configuration{
		ParameterizedFile: rao.ParameterizedFile{
			Paths: fileList,
		},
		SceneVariable: rao.GlobalVariable{
			Cookie:   sceneVariable.Cookie,
			Header:   sceneVariable.Header,
			Variable: sceneVariable.Variable,
			Assert:   sceneVariable.Assert,
		},
	}

	for _, prepositionsInfo := range prepositions.Prepositions {
		if prepositionsInfo.ID == req.NodeID {
			dbType := "mysql"
			if prepositionsInfo.API.Method == "ORACLE" {
				dbType = "oracle"
			} else if prepositionsInfo.API.Method == "PgSQL" {
				dbType = "postgresql"
			}

			runMysqlParam := rao.RunTargetParam{
				TargetID:   prepositionsInfo.ID,
				Name:       prepositionsInfo.API.Name,
				TeamID:     req.TeamID,
				TargetType: prepositionsInfo.API.TargetType,
				SqlDetail: rao.SqlDetail{
					SqlString: prepositionsInfo.API.SqlDetail.SqlString,
					SqlDatabaseInfo: rao.SqlDatabaseInfo{
						Type:     dbType,
						Host:     prepositionsInfo.API.SqlDetail.SqlDatabaseInfo.Host,
						User:     prepositionsInfo.API.SqlDetail.SqlDatabaseInfo.User,
						Password: prepositionsInfo.API.SqlDetail.SqlDatabaseInfo.Password,
						Port:     prepositionsInfo.API.SqlDetail.SqlDatabaseInfo.Port,
						DbName:   prepositionsInfo.API.SqlDetail.SqlDatabaseInfo.DbName,
						Charset:  prepositionsInfo.API.SqlDetail.SqlDatabaseInfo.Charset,
					},
					Assert: prepositionsInfo.API.SqlDetail.Assert,
					Regex:  prepositionsInfo.API.SqlDetail.Regex,
				},
				Configuration:  configurationData,
				GlobalVariable: globalVariable,
			}
			return runner.RunTarget(runMysqlParam)
		}
	}
	return "", nil
}

func SaveTargetCiteRelation(ctx *gin.Context, nodes []rao.Node, sceneID, planID, teamID, caseID string, source int32) error {
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectTargetCiteRelation)
	var filter interface{}
	if caseID != "" {
		filter = bson.D{{"case_id", caseID}}
	} else {
		filter = bson.D{{"scene_id", sceneID}, {"case_id", ""}}
	}
	// 保存引入接口和被引入地址的关系
	_, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		return err
	}

	targetCiteRelation := make([]interface{}, 0, len(nodes))
	for _, v := range nodes {
		if v.API.TargetID != "" {
			temp := mao.TargetCiteRelation{
				SceneID:  sceneID,
				TargetID: v.API.TargetID,
				NodeID:   v.ID,
				PlanID:   planID,
				TeamID:   teamID,
				CaseID:   caseID,
				Source:   source,
			}
			targetCiteRelation = append(targetCiteRelation, temp)
		}
	}
	_, err = collection.InsertMany(ctx, targetCiteRelation)
	if err != nil {
		return err
	}
	return nil
}

func SaveSceneCiteRelation(ctx *gin.Context, oldSceneID, newSceneID, planID, teamID string, source int32) error {
	sceneCiteRelation := mao.SceneCiteRelation{
		OldSceneID: oldSceneID,
		NewSceneID: newSceneID,
		PlanID:     planID,
		TeamID:     teamID,
		Source:     source,
	}
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectSceneCiteRelation)
	_, err := collection.InsertOne(ctx, sceneCiteRelation)
	if err != nil {
		log.Logger.Info("存储导入场景与计划关系数据失败,err:", err)
	}
	return nil
}

func GetSceneCanSyncData(ctx *gin.Context, req *rao.GetSceneCanSyncDataReq) ([]rao.SyncChildren, error) {
	tx := dal.GetQuery()
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectSceneCiteRelation)
	filter := bson.D{}
	if req.Source == consts.TargetSourceScene { // 是场景管理
		filter = bson.D{
			{"old_scene_id", req.SceneID},
			//{"source", bson.M{"$ne": req.Source}},
		}
	} else {
		filter = bson.D{{"new_scene_id", req.SceneID}}
	}
	cursor, err := collection.Find(ctx, filter)
	sceneCiteRelationList := make([]mao.SceneCiteRelation, 0, 100)
	if err = cursor.All(ctx, &sceneCiteRelationList); err != nil {
		return nil, err
	}

	allSceneIDs := make([]string, 0, len(sceneCiteRelationList))
	allStressPlanIDs := make([]string, 0, len(sceneCiteRelationList))
	allAutoPlanIDs := make([]string, 0, len(sceneCiteRelationList))
	if req.Source == consts.TargetSourceScene { // 场景管理
		for _, v := range sceneCiteRelationList {
			allSceneIDs = append(allSceneIDs, v.NewSceneID)
			if v.Source == consts.TargetSourcePlan {
				allStressPlanIDs = append(allStressPlanIDs, v.PlanID)
			}

			if v.Source == consts.TargetSourceAutoPlan {
				allAutoPlanIDs = append(allAutoPlanIDs, v.PlanID)
			}
		}
	} else {
		for _, v := range sceneCiteRelationList {
			allSceneIDs = append(allSceneIDs, v.OldSceneID)
			break
		}
	}

	// 查询所有性能计划信息
	stressPlanList, err := tx.StressPlan.WithContext(ctx).Where(tx.StressPlan.PlanID.In(allStressPlanIDs...),
		tx.StressPlan.Status.Eq(consts.PlanStatusNormal)).Find()
	if err != nil {
		return nil, err
	}

	// 查询所有自动化计划信息
	autoPlanList, err := tx.AutoPlan.WithContext(ctx).Where(tx.AutoPlan.PlanID.In(allAutoPlanIDs...),
		tx.AutoPlan.Status.Eq(consts.PlanStatusNormal)).Find()
	if err != nil {
		return nil, err
	}

	// 是否展示某个模块
	isShowSceneManage := false
	isShowStressPlan := false
	isShowAutoPlan := false

	// 查询所有场景基本信息
	sceneList, err := tx.Target.WithContext(ctx).Where(tx.Target.TargetID.In(allSceneIDs...),
		tx.Target.Status.Eq(consts.TargetStatusNormal)).Find()
	if err != nil {
		return nil, err
	}

	planSceneMap := make(map[string][]rao.SyncChildren, len(sceneList))
	for _, v := range sceneList {
		temp := rao.SyncChildren{
			Title:   v.Name,
			Key:     v.TargetID,
			Source:  v.Source,
			PlanID:  v.PlanID,
			SceneID: v.TargetID,
			Type:    v.TargetType,
		}
		if v.PlanID != "" {
			planSceneMap[v.PlanID] = append(planSceneMap[v.PlanID], temp)
		}
	}

	sceneBaseMap := make(map[string]*model.Target, len(sceneList))
	for _, v := range sceneList {
		sceneBaseMap[v.TargetID] = v
	}

	if req.Source == consts.TargetSourceScene { // 场景管理
		isShowStressPlan = true
		isShowAutoPlan = true
	} else if req.Source == consts.TargetSourcePlan { // 性能计划
		isShowSceneManage = true
		//isShowAutoPlan = true
	} else if req.Source == consts.TargetSourceAutoPlan { // 自动化计划
		isShowSceneManage = true
		//isShowStressPlan = true
	}

	sceneRes := rao.SyncChildren{}
	if isShowSceneManage {
		sceneRes.Title = "场景管理"
		sceneRes.Key = uuid.GetUUID()
		sceneRes.Source = consts.TargetSourceScene
		children := make([]rao.SyncChildren, 0, len(sceneList))
		for _, v := range sceneList {
			if v.Source == consts.TargetSourceScene {
				temp := rao.SyncChildren{
					Title:   v.Name,
					Key:     v.TargetID,
					Source:  v.Source,
					PlanID:  v.PlanID,
					SceneID: v.TargetID,
					Type:    v.TargetType,
				}
				children = append(children, temp)
			}
		}
		sceneRes.Children = children
	}

	// 组装性能计划数据
	stressPlanRes := rao.SyncChildren{}
	if isShowStressPlan {
		stressPlanRes.Title = "性能计划"
		stressPlanRes.Key = uuid.GetUUID()
		stressPlanRes.Source = consts.TargetSourcePlan
		for _, v := range stressPlanList {
			temp := rao.SyncChildren{
				Title:    v.PlanName,
				Key:      v.PlanID,
				Source:   consts.TargetSourceScene,
				PlanID:   v.PlanID,
				Type:     "stress_plan",
				Children: planSceneMap[v.PlanID],
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
				Title:    v.PlanName,
				Key:      v.PlanID,
				Source:   consts.TargetSourceAutoPlan,
				PlanID:   v.PlanID,
				Type:     "auto_plan",
				Children: planSceneMap[v.PlanID],
			}
			autoPlanRes.Children = append(autoPlanRes.Children, temp)
		}
	}

	// 最后返回值
	res := make([]rao.SyncChildren, 0, 3)
	if isShowSceneManage && len(sceneRes.Children) > 0 {
		res = append(res, sceneRes)
	}

	if isShowStressPlan && len(stressPlanRes.Children) > 0 {
		res = append(res, stressPlanRes)
	}

	if isShowAutoPlan && len(autoPlanRes.Children) > 0 {
		res = append(res, autoPlanRes)
	}

	return res, nil
}

func ExecSyncSceneData(ctx *gin.Context, req *rao.ExecSyncSceneDataReq) error {
	userID := jwt.GetUserIDByCtx(ctx)
	err := dal.GetQuery().Transaction(func(tx *query.Query) error {
		allSceneIDs := make([]string, 0, len(req.SyncSceneInfo)+1)

		for _, v := range req.SyncSceneInfo {
			allSceneIDs = append(allSceneIDs, v.SceneID)
		}
		allSceneIDs = append(allSceneIDs, req.SceneID)

		collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectFlow)
		filter := bson.D{{"scene_id", bson.M{"$in": allSceneIDs}}}
		cursor, err := collection.Find(ctx, filter)
		sceneFlowList := make([]mao.Flow, 0, 100)
		if err = cursor.All(ctx, &sceneFlowList); err != nil {
			return err
		}

		sceneFlowMap := make(map[string]mao.Flow, len(sceneFlowList))
		for _, v := range sceneFlowList {
			sceneFlowMap[v.SceneID] = v
		}

		// 查询基本的场景信息
		baseSceneInfo, err := tx.Target.WithContext(ctx).Where(tx.Target.TargetID.Eq(req.SceneID)).First()
		if err != nil {
			return err
		}

		// 查询基本场景下面的所有用例
		baseCaseList, err := tx.Target.WithContext(ctx).Where(tx.Target.ParentID.Eq(req.SceneID),
			tx.Target.TargetType.Eq(consts.TargetTypeTestCase)).Find()
		if err != nil {
			return err
		}
		baseCaseIDs := make([]string, 0, len(baseCaseList))
		for _, v := range baseCaseList {
			baseCaseIDs = append(baseCaseIDs, v.TargetID)
		}

		collection2 := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectSceneCaseFlow)

		// 查询基本用例的flow
		cursor, err = collection2.Find(ctx, bson.D{{"scene_case_id", bson.D{{"$in", baseCaseIDs}}}})
		if err != nil {
			return err
		}
		baseCaseFlows := make([]mao.SceneCaseFlow, 0, len(baseCaseIDs))
		if err = cursor.All(ctx, &baseCaseFlows); err != nil {
			return err
		}

		baseCaseFlowMap := make(map[string]mao.SceneCaseFlow, len(baseCaseFlows))
		for _, v := range baseCaseFlows {
			baseCaseFlowMap[v.SceneCaseID] = v
		}

		baseFlow := sceneFlowMap[req.SceneID]
		// 判断是推送还是拉取
		if req.SyncType == consts.TargetSyncDataTypePush { // 推送
			for _, v := range req.SyncSceneInfo {
				// 判断基础flow是否存在
				if _, ok := sceneFlowMap[req.SceneID]; !ok {
					_, err = collection.DeleteOne(ctx, bson.D{{"scene_id", v.SceneID}})
					continue
				}

				nodes := mao.Node{}
				err = bson.Unmarshal(baseFlow.Nodes, &nodes)
				if err != nil {
					log.Logger.Info("同步场景推送--解析node失败")
					continue
				}
				for kk := range nodes.Nodes {
					if v.Source == consts.TargetSourceScene {
						nodes.Nodes[kk].Data.From = "scene"
					}
					if v.Source == consts.TargetSourcePlan {
						nodes.Nodes[kk].Data.From = "plan"
					}
					if v.Source == consts.TargetSourceAutoPlan {
						nodes.Nodes[kk].Data.From = "auto_plan"
					}
				}

				nodesTemp, err := bson.Marshal(nodes)
				if err != nil {
					log.Logger.Info("同步场景推送--压缩node失败")
					continue
				}

				// 前置条件修改
				prepositions := mao.Preposition{}
				err = bson.Unmarshal(baseFlow.Prepositions, &prepositions)
				if err != nil {
					log.Logger.Info("同步场景推送--解析prepositions失败")
					continue
				}
				for key := range prepositions.Prepositions {
					if v.Source == consts.TargetSourceScene {
						prepositions.Prepositions[key].Data.From = "scene"
					}
					if v.Source == consts.TargetSourcePlan {
						prepositions.Prepositions[key].Data.From = "plan"
					}
					if v.Source == consts.TargetSourceAutoPlan {
						prepositions.Prepositions[key].Data.From = "auto_plan"
					}
				}
				prepositionsTemp, err := bson.Marshal(prepositions)
				if err != nil {
					log.Logger.Info("同步场景推送--压缩prepositions失败")
					continue
				}

				// 判断需要被推送的flow数据是否存在
				if _, ok := sceneFlowMap[v.SceneID]; !ok { // 如果不存在,则新增flow
					baseFlow.Nodes = nodesTemp
					baseFlow.PlanID = v.PlanID
					baseFlow.SceneID = v.SceneID
					baseFlow.Prepositions = prepositionsTemp
					// 更新api的uuid
					err = packer.ChangeSceneNodeUUID(&baseFlow)
					if err != nil {
						log.Logger.Info("同步场景拉取--替换node_id失败")
						continue
					}
					_, err := collection.InsertOne(ctx, baseFlow)
					if err != nil {
						return err
					}
				} else {
					// 推送场景数据
					needSyncFlow := sceneFlowMap[v.SceneID]
					needSyncFlow.Nodes = nodesTemp
					needSyncFlow.Edges = baseFlow.Edges
					needSyncFlow.Prepositions = prepositionsTemp
					needSyncFlow.EnvID = baseFlow.EnvID
					// 更新api的uuid
					err = packer.ChangeSceneNodeUUID(&needSyncFlow)
					if err != nil {
						log.Logger.Info("同步场景推送--替换node_id失败")
						continue
					}
					if _, err = collection.UpdateOne(ctx, bson.D{{"scene_id", v.SceneID}}, bson.M{"$set": needSyncFlow}); err != nil {
						log.Logger.Info("同步场景推送--同步失败")
					}
				}

				if len(baseCaseList) > 0 && v.Source != consts.TargetSourcePlan {
					// 删除老的用例测试用例数据
					_, err = tx.Target.WithContext(ctx).Where(tx.Target.ParentID.Eq(v.SceneID),
						tx.Target.TargetType.Eq(consts.TargetTypeTestCase)).Delete()
					if err != nil {
						return err
					}

					_, err = collection2.DeleteMany(ctx, bson.D{{"scene_id", v.SceneID}})
					if err != nil {
						continue
					}
					// 推送测试用例数据
					for _, oldCaseInfo := range baseCaseList {
						oldCaseID := oldCaseInfo.TargetID
						newCaseID := uuid.GetUUID()
						oldCaseInfo.ID = 0
						oldCaseInfo.TargetID = newCaseID
						oldCaseInfo.ParentID = v.SceneID
						oldCaseInfo.PlanID = v.PlanID
						oldCaseInfo.Source = v.Source
						oldCaseInfo.CreatedUserID = userID
						oldCaseInfo.CreatedAt = time.Now().Local()
						oldCaseInfo.UpdatedAt = time.Now().Local()
						oldCaseInfo.SourceID = oldCaseID
						err = tx.Target.WithContext(ctx).Create(oldCaseInfo)
						if err != nil {
							return err
						}

						tempFlow := baseCaseFlowMap[oldCaseID]
						err = packer.ChangeCaseNodeUUID(&tempFlow)
						if err != nil {
							log.Logger.Info("同步场景推送--替换用例node_id失败")
							continue
						}
						tempFlow.SceneID = v.SceneID
						tempFlow.SceneCaseID = newCaseID
						tempFlow.PlanID = v.PlanID
						_, err = collection2.InsertOne(ctx, tempFlow)
						if err != nil {
							log.Logger.Info("同步场景推送--插入用例数据失败,err:", err)
							continue
						}
					}
				}
			}
		} else { // 拉取
			for _, v := range req.SyncSceneInfo {
				// 判断被拉取的场景flow是否为空
				if _, ok := sceneFlowMap[v.SceneID]; !ok {
					// 如果被拉取的场景flow为空，则删除当前场景的flow
					if _, err = collection.DeleteOne(ctx, bson.D{{"scene_id", req.SceneID}}); err != nil {
						log.Logger.Info("同步场景拉取--同步失败")
						return err
					}
					return nil
				}

				// 同步场景数据
				pullFlow := sceneFlowMap[v.SceneID]
				nodes := mao.Node{}
				err = bson.Unmarshal(pullFlow.Nodes, &nodes)
				if err != nil {
					log.Logger.Info("同步场景拉取--解析node失败")
					continue
				}
				for kk := range nodes.Nodes {
					if req.Source == consts.TargetSourceScene {
						nodes.Nodes[kk].Data.From = "scene"
					}
					if req.Source == consts.TargetSourcePlan {
						nodes.Nodes[kk].Data.From = "plan"
					}
					if req.Source == consts.TargetSourceAutoPlan {
						nodes.Nodes[kk].Data.From = "auto_plan"
					}
				}

				nodesTemp, err := bson.Marshal(mao.Node{Nodes: nodes.Nodes})
				if err != nil {
					log.Logger.Info("同步场景拉取--压缩node失败")
					continue
				}

				// 前置条件修改
				prepositions := mao.Preposition{}
				err = bson.Unmarshal(pullFlow.Prepositions, &prepositions)
				if err != nil {
					log.Logger.Info("同步场景拉取--解析prepositions失败")
					continue
				}
				for key := range prepositions.Prepositions {
					if req.Source == consts.TargetSourceScene {
						prepositions.Prepositions[key].Data.From = "scene"
					}
					if req.Source == consts.TargetSourcePlan {
						prepositions.Prepositions[key].Data.From = "plan"
					}
					if req.Source == consts.TargetSourceAutoPlan {
						prepositions.Prepositions[key].Data.From = "auto_plan"
					}
				}
				prepositionsTemp, err := bson.Marshal(prepositions)
				if err != nil {
					log.Logger.Info("同步场景拉取--压缩prepositions失败")
					continue
				}

				// 判断需要拉取的场景flow是否存在
				if _, ok := sceneFlowMap[req.SceneID]; !ok { // 如果不存在,则新增flow
					pullFlow.Nodes = nodesTemp
					pullFlow.PlanID = baseSceneInfo.PlanID
					pullFlow.SceneID = req.SceneID
					pullFlow.Prepositions = prepositionsTemp
					// 更新api的uuid
					err = packer.ChangeSceneNodeUUID(&pullFlow)
					if err != nil {
						log.Logger.Info("同步场景拉取--替换node_id失败")
						continue
					}
					_, err := collection.InsertOne(ctx, pullFlow)
					if err != nil {
						return err
					}
				} else {
					baseFlow.Nodes = nodesTemp
					baseFlow.Edges = pullFlow.Edges
					baseFlow.Prepositions = prepositionsTemp
					baseFlow.EnvID = pullFlow.EnvID
					// 更新api的uuid
					err = packer.ChangeSceneNodeUUID(&baseFlow)
					if err != nil {
						log.Logger.Info("同步场景拉取--替换node_id失败")
						continue
					}
					if _, err = collection.UpdateOne(ctx, bson.D{{"scene_id", req.SceneID}}, bson.M{"$set": baseFlow}); err != nil {
						log.Logger.Info("同步场景拉取--同步失败")
					}
				}

				// 同步用例数据
				// 查询基本场景下面的所有用例
				originCaseList, err := tx.Target.WithContext(ctx).Where(tx.Target.ParentID.Eq(v.SceneID),
					tx.Target.TargetType.Eq(consts.TargetTypeTestCase)).Find()
				if err != nil {
					return err
				}

				if len(originCaseList) > 0 && req.Source != consts.TargetSourcePlan {
					// 删除老的用例基本信息
					_, err = tx.Target.WithContext(ctx).Where(tx.Target.ParentID.Eq(req.SceneID),
						tx.Target.TargetType.Eq(consts.TargetTypeTestCase)).Delete()
					if err != nil {
						return err
					}

					// 删除老的用例flow
					_, err = collection2.DeleteMany(ctx, bson.D{{"scene_id", req.SceneID}})
					if err != nil {
						continue
					}

					originCaseIDs := make([]string, 0, len(originCaseList))
					for _, vv := range originCaseList {
						originCaseIDs = append(originCaseIDs, vv.TargetID)
					}

					// 获取远端用例flow
					cursor1, err := collection2.Find(ctx, bson.D{{"scene_case_id", bson.D{{"$in", originCaseIDs}}}})
					if err != nil {
						return err
					}
					originCaseFlows := make([]mao.SceneCaseFlow, 0, len(originCaseIDs))
					if err = cursor1.All(ctx, &originCaseFlows); err != nil {
						return err
					}
					originCaseFlowMap := make(map[string]mao.SceneCaseFlow, len(originCaseFlows))
					for _, vv := range originCaseFlows {
						originCaseFlowMap[vv.SceneCaseID] = vv
					}

					// 推送测试用例数据
					for _, oldCaseInfo := range originCaseList {
						oldCaseID := oldCaseInfo.TargetID
						newCaseID := uuid.GetUUID()
						oldCaseInfo.ID = 0
						oldCaseInfo.TargetID = newCaseID
						oldCaseInfo.ParentID = req.SceneID
						oldCaseInfo.PlanID = baseSceneInfo.PlanID
						oldCaseInfo.Source = baseSceneInfo.Source
						oldCaseInfo.CreatedUserID = userID
						oldCaseInfo.CreatedAt = time.Now().Local()
						oldCaseInfo.UpdatedAt = time.Now().Local()
						oldCaseInfo.SourceID = oldCaseID
						err = tx.Target.WithContext(ctx).Create(oldCaseInfo)
						if err != nil {
							return err
						}

						tempFlow := originCaseFlowMap[oldCaseID]
						err = packer.ChangeCaseNodeUUID(&tempFlow)
						if err != nil {
							log.Logger.Info("同步场景拉取--替换用例node_id失败")
							continue
						}
						tempFlow.SceneID = req.SceneID
						tempFlow.SceneCaseID = newCaseID
						tempFlow.PlanID = baseSceneInfo.PlanID
						_, err = collection2.InsertOne(ctx, tempFlow)
						if err != nil {
							log.Logger.Info("同步场景拉取--插入用例数据失败,err:", err)
							continue
						}
					}
				}
				break
			}
		}
		return nil
	})
	return err
}

func ExportScene(ctx *gin.Context, req *rao.ExportSceneReq) (rao.ExportSceneResp, error) {
	res := rao.ExportSceneResp{}
	collectionFlow := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectFlow)
	collectionCaseFlow := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectSceneCaseFlow)
	err := dal.GetQuery().Transaction(func(tx *query.Query) error {
		// 判断导出类型是什么
		if req.ExportType == 1 || req.ExportType == 3 { // 场景
			// 查询场景基本信息
			sceneList, err := tx.Target.WithContext(ctx).Where(tx.Target.TargetID.In(req.SceneIDs...)).Find()
			if err != nil {
				return err
			}

			// 查询场景flow
			cursor, err := collectionFlow.Find(ctx, bson.D{{"scene_id", bson.D{{"$in", req.SceneIDs}}}})
			if err != nil {
				return err
			}
			sceneFlow := make([]mao.Flow, 0, len(req.SceneIDs))
			if err = cursor.All(ctx, &sceneFlow); err != nil {
				return err
			}

			sceneFlowMap := make(map[string]mao.Flow, len(sceneFlow))
			for _, v := range sceneFlow {
				sceneFlowMap[v.SceneID] = v
			}

			sceneFlowDetailList := make([]map[string]interface{}, 0, len(sceneFlow))
			for _, v := range sceneList {
				tempData := make(map[string]interface{}, 50)
				tempData["target_id"] = v.TargetID
				tempData["scene_name"] = v.Name
				tempData["parent_id"] = v.ParentID
				tempData["target_type"] = v.TargetType
				tempData["target_id"] = v.TargetID
				tempData["target_id"] = v.TargetID
				tempData["target_id"] = v.TargetID
				if v.TargetType == consts.TargetTypeScene {
					nodes := mao.Node{}
					err = bson.Unmarshal(sceneFlowMap[v.TargetID].Nodes, &nodes)
					if err != nil {
						continue
					}

					edges := mao.Edge{}
					err = bson.Unmarshal(sceneFlowMap[v.TargetID].Edges, &edges)
					if err != nil {
						continue
					}

					prepositions := mao.Preposition{}
					err = bson.Unmarshal(sceneFlowMap[v.TargetID].Prepositions, &prepositions)
					if err != nil {
						continue
					}
					tempData["nodes"] = rao.ExportNode{Nodes: nodes.Nodes}
					tempData["edges"] = rao.ExportEdge{Edges: edges.Edges}
					tempData["prepositions"] = rao.ExportPreposition{Prepositions: prepositions.Prepositions}
				}
				sceneFlowDetailList = append(sceneFlowDetailList, tempData)
			}
			res.SceneDetailList = sceneFlowDetailList
		}

		if req.ExportType == 2 || req.ExportType == 3 { // 只导用例
			// 查询用例基本信息列表
			caseList, err := tx.Target.WithContext(ctx).Where(tx.Target.ParentID.In(req.SceneIDs...)).Find()
			if err != nil {
				return err
			}

			caseIDs := make([]string, 0, len(caseList))
			for _, v := range caseList {
				caseIDs = append(caseIDs, v.TargetID)
			}

			// 查询场景flow
			cursor, err := collectionCaseFlow.Find(ctx, bson.D{{"scene_case_id", bson.D{{"$in", caseIDs}}}})
			if err != nil {
				return err
			}
			caseFlow := make([]mao.SceneCaseFlow, 0, len(caseIDs))
			if err = cursor.All(ctx, &caseFlow); err != nil {
				return err
			}
			caseFlowMap := make(map[string]mao.SceneCaseFlow, len(caseFlow))
			for _, v := range caseFlow {
				caseFlowMap[v.SceneCaseID] = v
			}

			caseFlowDetailList := make([]map[string]interface{}, 0, len(caseIDs))
			for _, v := range caseList {
				tempData := make(map[string]interface{}, 50)
				tempData["target_id"] = v.TargetID
				tempData["scene_name"] = v.Name
				tempData["parent_id"] = v.ParentID
				tempData["target_type"] = v.TargetType
				tempData["target_id"] = v.TargetID
				tempData["target_id"] = v.TargetID
				tempData["target_id"] = v.TargetID
				if v.TargetType == consts.TargetTypeTestCase {
					nodes := mao.Node{}
					err = bson.Unmarshal(caseFlowMap[v.TargetID].Nodes, &nodes)
					if err != nil {
						continue
					}

					edges := mao.Edge{}
					err = bson.Unmarshal(caseFlowMap[v.TargetID].Edges, &edges)
					if err != nil {
						continue
					}
					tempData["nodes"] = rao.ExportNode{Nodes: nodes.Nodes}
					tempData["edges"] = rao.ExportEdge{Edges: edges.Edges}
				}

				caseFlowDetailList = append(caseFlowDetailList, tempData)
			}
			res.CaseDetailList = caseFlowDetailList
		}
		return nil
	})
	return res, err
}
