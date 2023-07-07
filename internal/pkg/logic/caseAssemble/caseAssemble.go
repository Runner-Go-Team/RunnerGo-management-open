package caseAssemble

import (
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
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/target"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/packer"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/public"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gen"
	"strconv"
	"strings"
)

func GetCaseAssembleList(ctx *gin.Context, req *rao.GetCaseAssembleListReq) ([]*rao.CaseAssembleDetailResp, error) {
	// target表
	targetTB := dal.GetQuery().Target
	conditions := make([]gen.Condition, 0)
	if req.CaseName != "" {
		conditions = append(conditions, targetTB.Name.Like(fmt.Sprintf("%%%s%%", req.CaseName)))
	}
	conditions = append(conditions, targetTB.TargetType.Eq(consts.TargetTypeTestCase))
	conditions = append(conditions, targetTB.ParentID.Eq(req.SceneID))
	//conditions = append(conditions, targetTB.Source.Eq(consts.TargetSourceAutoPlan))
	list, err := targetTB.WithContext(ctx).Where(conditions...).Order(targetTB.Sort).Find()
	if err != nil {
		log.Logger.Info("用例集列表--获取列表失败，err:", err)
		return nil, err
	}

	res := make([]*rao.CaseAssembleDetailResp, 0, len(list))
	for _, detail := range list {
		detailTmp := &rao.CaseAssembleDetailResp{
			CaseID:      detail.TargetID,
			TeamID:      detail.TeamID,
			SceneID:     detail.ParentID,
			CaseName:    detail.Name,
			Sort:        detail.Sort,
			IsChecked:   detail.IsChecked,
			CreatedAt:   detail.CreatedAt.Unix(),
			UpdatedAt:   detail.UpdatedAt.Unix(),
			Status:      detail.Status,
			Description: detail.Description,
		}
		res = append(res, detailTmp)
	}
	return res, nil
}

func GetCaseAssembleDetail(ctx *gin.Context, req *rao.CopyAssembleReq) (*rao.CaseAssembleDetailResp, error) {
	// target表
	targetTB := dal.GetQuery().Target

	conditions := make([]gen.Condition, 0)
	conditions = append(conditions, targetTB.TargetID.Eq(req.CaseID))
	conditions = append(conditions, targetTB.TargetType.Eq(consts.TargetTypeTestCase))

	detail, err := targetTB.WithContext(ctx).Where(conditions...).First()
	if err != nil {
		log.Logger.Info("用例详情--获取详情失败，err:", err)
		return nil, err
	}

	res := &rao.CaseAssembleDetailResp{
		CaseID:   detail.TargetID,
		TeamID:   detail.TeamID,
		CaseName: detail.Name,
		//Sort:      detail.Sort,
		//TypeSort:  detail.TypeSort,
		CreatedAt:   detail.CreatedAt.Unix(),
		UpdatedAt:   detail.UpdatedAt.Unix(),
		Status:      detail.Status,
		Description: detail.Description,
	}
	return res, nil
}

func CopyCaseAssemble(ctx *gin.Context, req *rao.CopyAssembleReq) error {
	err := dal.GetQuery().Transaction(func(tx *query.Query) error {
		// target表
		targetTB := dal.GetQuery().Target

		conditions := make([]gen.Condition, 0)
		conditions = append(conditions, targetTB.TargetID.Eq(req.CaseID))
		conditions = append(conditions, targetTB.TargetType.Eq(consts.TargetTypeTestCase))
		detail, err := targetTB.WithContext(ctx).Where(conditions...).First()
		if err != nil {
			log.Logger.Info("用例详情--获取详情失败，err:", err)
			return err
		}

		oldCaseName := detail.Name
		newCaseName := oldCaseName + "_1"

		list, err := targetTB.WithContext(ctx).Where(targetTB.TeamID.Eq(req.TeamID),
			targetTB.ParentID.Eq(detail.ParentID), targetTB.Source.Eq(consts.TargetSourceAutoPlan)).Where(targetTB.Name.Like(fmt.Sprintf("%s%%", oldCaseName+"_"))).Find()
		if err == nil && len(list) > 0 { // 有复制过得配置
			maxNum := 0
			for _, caseInfo := range list {
				nameTmp := caseInfo.Name
				postfixSlice := strings.Split(nameTmp, "_")
				if len(postfixSlice) < 2 {
					continue
				}
				currentNum, err := strconv.Atoi(postfixSlice[len(postfixSlice)-1])
				if err != nil {
					log.Logger.Info("复制用例--类型转换失败，err:", err)
				}
				if currentNum > maxNum {
					maxNum = currentNum
				}
			}
			newCaseName = oldCaseName + fmt.Sprintf("_%d", maxNum+1)
		}

		nameLength := public.GetStringNum(newCaseName)
		if nameLength > 30 { // 场景名称限制30个字符
			return fmt.Errorf("名称过长！不可超出30字符")
		}

		userID := jwt.GetUserIDByCtx(ctx)
		newCase := &model.Target{
			TargetID:      uuid.GetUUID(),
			TeamID:        detail.TeamID,
			TargetType:    detail.TargetType,
			Name:          newCaseName,
			ParentID:      detail.ParentID,
			Method:        detail.Method,
			Sort:          detail.Sort,
			TypeSort:      detail.TypeSort,
			Status:        detail.Status,
			CreatedUserID: userID,
			RecentUserID:  userID,
			Source:        consts.TargetSourceAutoPlan,
			PlanID:        detail.PlanID,
			Version:       1,
			Description:   detail.Description,
		}
		err = targetTB.WithContext(ctx).Create(newCase)
		if err != nil {
			log.Logger.Info("复制用例--复制数据失败，err:", err)
			return err
		}
		req.CaseID = newCase.TargetID

		//复制mongo中的用例执行流
		var oldCaseSceneCaseFlow mao.SceneCaseFlow

		collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectSceneCaseFlow)
		collectionErr := collection.FindOne(ctx, bson.D{{"scene_case_id", detail.TargetID}}).Decode(&oldCaseSceneCaseFlow)
		if collectionErr == nil {
			oldCaseSceneCaseFlow.SceneCaseID = newCase.TargetID
			// 更新api的uuid
			err := packer.ChangeCaseNodeUUID(&oldCaseSceneCaseFlow)
			if err != nil {
				log.Logger.Info("复制用例--替换event_id失败")
				return err
			}
			if _, err := collection.InsertOne(ctx, oldCaseSceneCaseFlow); err != nil {
				log.Logger.Info("复制用例--插入SceneCaseFlow到mg库失败")
				return err
			}
		}
		return nil
	})
	return err
}

// SceneCaseNameIsExist 判断用例名称在同一场景下是否已存在
func SceneCaseNameIsExist(ctx *gin.Context, req *rao.SaveCaseAssembleReq) (bool, error) {

	targetTB := dal.GetQuery().Target

	conditions := make([]gen.Condition, 0)
	conditions = append(conditions, targetTB.ParentID.Eq(req.SceneID))
	conditions = append(conditions, targetTB.Name.Eq(req.Name))
	conditions = append(conditions, targetTB.TargetType.Eq(consts.TargetTypeTestCase))
	//conditions = append(conditions, targetTB.Status.Eq(consts.TargetStatusNormal))
	if req.CaseID != "" {
		conditions = append(conditions, targetTB.TargetID.Neq(req.CaseID))
	}

	existCase, err := targetTB.WithContext(ctx).Where(conditions...).Find()

	if err != nil {
		return false, err
	}
	if len(existCase) != 0 {
		return true, err
	}

	return false, err
}

func SaveCaseAssemble(ctx *gin.Context, req *rao.SaveCaseAssembleReq) error {
	userID := jwt.GetUserIDByCtx(ctx)
	targetInfo := packer.TransSaveCaseAssembleToTargetModel(req, userID)

	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		// 查询是否存在
		_, err := tx.Target.WithContext(ctx).Where(tx.Target.TargetID.Eq(req.CaseID)).First()
		if err != nil {
			if err := tx.Target.WithContext(ctx).Create(targetInfo); err != nil {
				return err
			}

			//获取场景的执行流flow 然后赋给用例执行流flow
			sceneFlow := mao.Flow{}
			collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectFlow)
			collectionErr := collection.FindOne(ctx, bson.D{{"scene_id", req.SceneID}}).Decode(&sceneFlow)
			if collectionErr != nil && collectionErr != mongo.ErrNoDocuments {
				return collectionErr
			}

			sceneCaseFlow := packer.TransMaoFlowToMaoSceneCaseFlow(&sceneFlow, targetInfo.TargetID)
			sceneCaseFlowCollection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectSceneCaseFlow)
			sceneCaseFlowErr := sceneCaseFlowCollection.FindOne(ctx, bson.D{{"scene_case_id", tx.Target.TargetID}}).Decode(&sceneFlow)
			if sceneCaseFlowErr == mongo.ErrNoDocuments { // 新建
				_, _ = sceneCaseFlowCollection.InsertOne(ctx, sceneCaseFlow)
			} else {
				_, _ = sceneCaseFlowCollection.UpdateOne(ctx, bson.D{
					{"scene_id", sceneCaseFlow.SceneCaseID},
				}, bson.M{"$set": sceneCaseFlow})
			}
			return record.InsertCreate(ctx, targetInfo.TeamID, userID, record.OperationOperateCreateTestCase, targetInfo.Name)
		} else {
			if _, err := tx.Target.WithContext(ctx).Where(tx.Target.TargetID.Eq(req.CaseID)).Updates(targetInfo); err != nil {
				return err
			}
			return record.InsertUpdate(ctx, targetInfo.TeamID, userID, record.OperationOperateUpdateSceneCase, targetInfo.Name)
		}
	})
	return err
}

func SaveSceneCaseFlow(ctx *gin.Context, req *rao.SaveSceneCaseFlowReq) error {

	flow := packer.TransSaveSceneCaseFlowReqToMaoFlow(req)
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectSceneCaseFlow)

	err := collection.FindOne(ctx, bson.D{{"scene_case_id", req.SceneCaseID}}).Err()
	if err == mongo.ErrNoDocuments { // 新建
		_, err := collection.InsertOne(ctx, flow)
		return err
	}

	_, err = collection.UpdateOne(ctx, bson.D{
		{"scene_case_id", flow.SceneCaseID},
	}, bson.M{"$set": flow})

	return err
}

func GetSceneCaseFlow(ctx *gin.Context, req *rao.GetSceneCaseFlowReq) (*rao.GetSceneCaseFlowResp, error) {
	// 查询用例基本信息
	tx := dal.GetQuery().Target
	caseInfo, err := tx.WithContext(ctx).Where(tx.TargetID.Eq(req.CaseID)).First()
	if err != nil {
		return nil, err
	}

	res := &rao.GetSceneCaseFlowResp{
		SceneID:     caseInfo.ParentID,
		SceneCaseID: caseInfo.TargetID,
		TeamID:      caseInfo.TeamID,
		Version:     caseInfo.Version,
	}

	ret := mao.SceneCaseFlow{}
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectSceneCaseFlow)
	err = collection.FindOne(ctx, bson.D{{"scene_case_id", req.CaseID}}).Decode(&ret)
	if err != nil {
		return res, nil
	}

	return packer.TransMaoSceneCaseFlowToRaoGetFowResp(&ret), nil
}

func ChangeCaseAssembleCheck(ctx *gin.Context, req *rao.ChangeCaseAssembleCheckReq) error {

	userID := jwt.GetUserIDByCtx(ctx)

	// target表
	targetTB := dal.GetQuery().Target

	updateData := make(map[string]interface{}, 1)
	updateData["recent_user_id"] = userID
	updateData["is_checked"] = req.IsChecked

	fmt.Println("参数", req, updateData)
	_, err := targetTB.WithContext(ctx).Where(targetTB.TargetID.Eq(req.CaseID)).Where(targetTB.TeamID.Eq(req.TeamID)).Updates(updateData)
	if err != nil {
		fmt.Println("没找到")
		return err
	}

	return nil
}

func DelCaseAssemble(ctx *gin.Context, req *rao.DelCaseAssembleReq) error {
	userID := jwt.GetUserIDByCtx(ctx)
	err := dal.GetQuery().Transaction(func(tx *query.Query) error {
		targetTB := dal.GetQuery().Target
		caseInfo, err := targetTB.WithContext(ctx).Where(targetTB.TargetID.Eq(req.CaseID)).First()
		if err != nil {
			return err
		}

		_, err = targetTB.WithContext(ctx).Where(targetTB.TargetID.Eq(req.CaseID)).Where(targetTB.TeamID.Eq(req.TeamID)).Delete()
		if err != nil {
			return err
		}

		// 删除用例对应的flow
		collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectSceneCaseFlow)
		_, err = collection.DeleteMany(ctx, bson.D{{"scene_case_id", req.CaseID}})
		if err != nil {
			return err
		}

		if err := record.InsertDelete(ctx, caseInfo.TeamID, userID, record.OperationOperateDeleteTestCase, caseInfo.Name); err != nil {
			return err
		}
		return nil
	})
	return err
}

func SendSceneCase(ctx *gin.Context, teamID string, sceneID, sceneCaseID string, userID string) (string, error) {
	targetTB := dal.GetQuery().Target
	t, err := targetTB.WithContext(ctx).Where(targetTB.TargetID.Eq(sceneCaseID), targetTB.TargetType.Eq(consts.TargetTypeTestCase)).First()
	if err != nil {
		return "", err
	}

	var flow mao.Flow
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectFlow)
	err = collection.FindOne(ctx, bson.D{{"scene_id", sceneID}}).Decode(&flow)
	if err != nil {
		return "", err
	}

	var caseFlow mao.Flow
	collection = dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectSceneCaseFlow)
	err = collection.FindOne(ctx, bson.D{{"scene_case_id", sceneCaseID}}).Decode(&caseFlow)
	if err != nil {
		return "", err
	}

	vi := dal.GetQuery().VariableImport
	vis, err := vi.WithContext(ctx).Where(vi.SceneID.Eq(sceneID)).Limit(5).Find()
	if err != nil {
		return "", err
	}

	// 获取全局变量
	globalVariable, _ := target.GetGlobalVariable(ctx, teamID)

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

	if err = record.InsertDebug(ctx, teamID, userID, record.OperationOperateRunSceneCase, t.Name); err != nil {
		return "", err
	}

	req := packer.TransMaoFlowToRaoSceneCaseFlow(t, &flow, &caseFlow, vis, sceneVariable, globalVariable)
	return runner.RunSceneCaseFlow(req)

}
