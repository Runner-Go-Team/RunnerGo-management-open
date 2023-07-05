package scene

import (
	"context"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/errno"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/record"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/mao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/query"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/runner"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/target"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/packer"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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

			//// 删除自动化场景对应的任务配置
			//if targetInfo.Source == consts.TargetSourceAutoPlan { // 自动化下的场景
			//	// 查询计划信息
			//	planInfo, err := tx.AutoPlan.WithContext(ctx).Where(tx.AutoPlan.PlanID.Eq(req.PlanID)).First()
			//	if err != nil {
			//		return err
			//	}
			//
			//	if planInfo.TaskType == consts.PlanTaskTypeNormal { // 普通任务
			//		_, err = tx.AutoPlanTaskConf.WithContext(ctx).Where(tx.AutoPlanTaskConf.PlanID.Eq(req.PlanID)).Delete()
			//		if err != nil {
			//			return err
			//		}
			//	} else {
			//		_, err = tx.AutoPlanTimedTaskConf.WithContext(ctx).Where(tx.AutoPlanTimedTaskConf.PlanID.Eq(req.PlanID)).Delete()
			//		if err != nil {
			//			return err
			//		}
			//	}
			//}
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
