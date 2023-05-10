package homePage

import (
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/jwt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/mao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/query"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

func HomePage(ctx *gin.Context, req *rao.HomePageReq) (*rao.HomePageResp, error) {
	userID := jwt.GetUserIDByCtx(ctx)

	res := &rao.HomePageResp{}
	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		// 1、接口管理数量统计
		targetTable := tx.Target
		targetList, err := targetTable.WithContext(ctx).Where(targetTable.TeamID.Eq(req.TeamID), targetTable.Status.Eq(consts.TargetStatusNormal),
			targetTable.TargetType.In(consts.TargetTypeAPI, consts.TargetTypeScene, consts.TargetTypeTestCase)).Find()
		if err != nil {
			return err
		}

		allSceneID := make([]string, 0, len(targetList)) // 当前团队下--所有场景id
		allSceneIDMap := make(map[string]int, len(targetList))

		allCaseID := make([]string, 0, len(targetList))    // 当前团队下--所有用例id
		allApiMap := make(map[string]int, len(targetList)) // 当前团队下--所有接口的字典

		autoPlanSceneIDs := make([]string, 0, len(targetList))   // 自动化下的所有场景
		autoPlanCaseIDs := make([]string, 0, len(targetList))    // 自动化下所有测试用例
		stressPlanSceneIDs := make([]string, 0, len(targetList)) // 性能下的所有场景
		stressPlanCaseIDs := make([]string, 0, len(targetList))  // 性能下所有测试用例

		for _, tlInfo := range targetList {
			if tlInfo.TargetType == "scene" {
				allSceneIDMap[tlInfo.TargetID] = 1
				allSceneID = append(allSceneID, tlInfo.TargetID)
				if tlInfo.Source == consts.TargetSourceAutoPlan && tlInfo.PlanID != "" {
					autoPlanSceneIDs = append(autoPlanSceneIDs, tlInfo.TargetID)
				}
				if tlInfo.Source == consts.TargetSourcePlan && tlInfo.PlanID != "" {
					stressPlanSceneIDs = append(stressPlanSceneIDs, tlInfo.TargetID)
				}
			}
			if tlInfo.TargetType == "test_case" {
				allCaseID = append(allCaseID, tlInfo.TargetID)
				if tlInfo.Source == consts.TargetSourceAutoPlan && tlInfo.PlanID != "" {
					autoPlanCaseIDs = append(autoPlanCaseIDs, tlInfo.TargetID)
				}
				if tlInfo.Source == consts.TargetSourcePlan && tlInfo.PlanID != "" {
					stressPlanCaseIDs = append(stressPlanCaseIDs, tlInfo.TargetID)
				}
			}

			if tlInfo.TargetType == "api" {
				allApiMap[tlInfo.TargetID] = 1
			}
		}

		allSiteApiMap := make(map[string]int, 100)
		// 获取当前团队下所有场景flow
		allSceneFlows, err := GetFlowBySceneIDs(ctx, allSceneID, req.TeamID)
		if err == nil {
			for _, flowInfo := range allSceneFlows {
				var node mao.Node
				err := bson.Unmarshal(flowInfo.Nodes, &node)
				if err == nil {
					for _, nodeTemp := range node.Nodes {
						if nodeTemp.API.TargetType == "api" && nodeTemp.API.TargetID != "" {
							allSiteApiMap[nodeTemp.API.TargetID]++
						}
					}
				}
			}
		}

		// 获取自动化下面的所有接口
		autoPlanSiteApiMap := make(map[string]int64)
		autoPlanTotalApiNum := 0
		autoPlanSceneFlows, err := GetFlowBySceneIDs(ctx, autoPlanSceneIDs, req.TeamID)
		if err == nil {
			for _, flowInfo := range autoPlanSceneFlows {
				var node mao.Node
				err := bson.Unmarshal(flowInfo.Nodes, &node)
				if err == nil {
					for _, nodeTemp := range node.Nodes {
						if nodeTemp.API.TargetType == "api" {
							if nodeTemp.API.TargetID != "" {
								autoPlanSiteApiMap[nodeTemp.API.TargetID]++
							}
							autoPlanTotalApiNum++
						}

					}
				}
			}
		}

		// 获取性能下面的所有接口
		stressPlanSiteApiMap := make(map[string]int64)
		stressPlanTotalApiNum := 0
		stressPlanSceneFlows, err := GetFlowBySceneIDs(ctx, stressPlanSceneIDs, req.TeamID)
		if err == nil {
			for _, flowInfo := range stressPlanSceneFlows {
				var node mao.Node
				err := bson.Unmarshal(flowInfo.Nodes, &node)
				if err == nil {
					for _, nodeTemp := range node.Nodes {
						if nodeTemp.API.TargetType == "api" {
							if nodeTemp.API.TargetID != "" {
								stressPlanSiteApiMap[nodeTemp.API.TargetID]++
							}
							stressPlanTotalApiNum++
						}

					}
				}
			}
		}

		// 获取当前团队下所有用例 caseFlow
		allCaseFlows, err := GetCaseFlowByCaseIDs(ctx, allCaseID, req.TeamID)
		if err == nil {
			for _, flowInfo := range allCaseFlows {
				var node mao.SceneCaseFlowNode
				err3 := bson.Unmarshal(flowInfo.Nodes, &node)
				if err3 == nil {
					for _, nodeTemp := range node.Nodes {
						if nodeTemp.API.TargetType == "api" && nodeTemp.API.TargetID != "" {
							allSiteApiMap[nodeTemp.API.TargetID]++
						}
					}
				}
			}
		}

		// 获取自动化下面所有的用例 caseFlow
		autoPlanCaseFlows, err := GetCaseFlowByCaseIDs(ctx, autoPlanCaseIDs, req.TeamID)
		if err == nil {
			for _, flowInfo := range autoPlanCaseFlows {
				var node mao.SceneCaseFlowNode
				err3 := bson.Unmarshal(flowInfo.Nodes, &node)
				if err3 == nil {
					for _, nodeTemp := range node.Nodes {
						if nodeTemp.API.TargetType == "api" {
							if nodeTemp.API.TargetID != "" {
								autoPlanSiteApiMap[nodeTemp.API.TargetID]++
							}
							autoPlanTotalApiNum++
						}
					}
				}
			}
		}

		// 获取性能下面所有的用例 caseFlow
		stressPlanCaseFlows, err := GetCaseFlowByCaseIDs(ctx, stressPlanCaseIDs, req.TeamID)
		if err == nil {
			for _, flowInfo := range stressPlanCaseFlows {
				var node mao.SceneCaseFlowNode
				err3 := bson.Unmarshal(flowInfo.Nodes, &node)
				if err3 == nil {
					for _, nodeTemp := range node.Nodes {
						if nodeTemp.API.TargetType == "api" {
							if nodeTemp.API.TargetID != "" {
								stressPlanSiteApiMap[nodeTemp.API.TargetID]++
							}
							stressPlanTotalApiNum++
						}
					}
				}
			}
		}

		// 查询当前用户所属的所有团队
		userTeamTable := tx.UserTeam
		teamList, err := userTeamTable.WithContext(ctx).Where(userTeamTable.UserID.Eq(userID)).Find()
		if err != nil {
			return err
		}

		allTeamIDs := make([]string, 0, len(teamList))
		for _, teamInfo := range teamList {
			allTeamIDs = append(allTeamIDs, teamInfo.TeamID)
		}

		// teamID与名字映射
		teamIDNameMap := make(map[string]string)
		// 查询所有团队名称信息
		teamBaseList, err := tx.Team.WithContext(ctx).Where(tx.Team.TeamID.In(allTeamIDs...)).Find()
		if err != nil {
			return err
		}

		for _, teamBaseInfo := range teamBaseList {
			teamIDNameMap[teamBaseInfo.TeamID] = teamBaseInfo.Name
		}

		// 性能测试计划总览
		// 查询当前用户下所有团队的所有性能测试计划
		stressPlanList, err := tx.StressPlan.WithContext(ctx).Where(tx.StressPlan.TeamID.In(allTeamIDs...)).Find()
		if err != nil {
			return err
		}

		stressPlanTotalNumMap := make(map[string]int)  // 性能计划统计map
		stressPlanExecCountMap := make(map[string]int) // 性能计划统计map
		for _, spInfo := range stressPlanList {
			stressPlanTotalNumMap[spInfo.TeamID]++
			if spInfo.RunCount > 0 {
				stressPlanExecCountMap[spInfo.TeamID]++
			}
		}

		// 自动化测试计划总览
		// 查询当前用户下所有团队的所有自动化测试计划
		autoPlanList, err := tx.AutoPlan.WithContext(ctx).Where(tx.AutoPlan.TeamID.In(allTeamIDs...)).Find()
		if err != nil {
			return err
		}

		autoPlanTotalNumMap := make(map[string]int)  // 自动化计划数量统计map
		autoPlanExecCountMap := make(map[string]int) // 自动化计划数量统计map
		for _, apInfo := range autoPlanList {
			autoPlanTotalNumMap[apInfo.TeamID]++
			if apInfo.RunCount > 0 {
				autoPlanExecCountMap[apInfo.TeamID]++
			}
		}

		teamOverview := make([]rao.TeamOverview, 0, len(teamList))
		for _, tInfo := range teamList {
			tempData := rao.TeamOverview{
				TeamName:           teamIDNameMap[tInfo.TeamID],
				AutoPlanTotalNum:   autoPlanTotalNumMap[tInfo.TeamID],
				AutoPlanExecNum:    autoPlanExecCountMap[tInfo.TeamID],
				StressPlanTotalNum: stressPlanTotalNumMap[tInfo.TeamID],
				StressPlanExecNum:  stressPlanExecCountMap[tInfo.TeamID],
			}
			teamOverview = append(teamOverview, tempData)
		}

		//当前团队下的数据统计
		apiTotalCount := 0 // 总的接口数

		scentTotalCount := 0                 // 总的场景数
		sceneCiteMap := make(map[string]int) // 被引用的场景map

		currentTeamTotalCaseNum := 0 // 当前团队下总用例数

		stressTotalApiNum := 0             // 性能计划下所有接口数量
		stressTotalSceneNum := 0           // 性能计划下所有场景数量
		stressPlanTotalImportSceneNum := 0 // 性能所有引入进来的场景数量

		autoPlanTotalSceneNum := 0       // 自动化测试里面的所有场景数
		autoPlanTotalImportSceneNum := 0 // 自动化所有引入进来的场景数量

		autoPlanTestCaseIDs := make([]string, 0, len(targetList)) // 自动化计划下面所有测试用例id集合

		// 时间相关数据
		todayTimeStr := time.Now().Format("2006-01-02")
		todayTimeObj, _ := time.Parse("2006-01-02", todayTimeStr)
		endTime := todayTimeObj.Unix() - 8*3600
		startTime := endTime - 7*24*3600
		startTimeTmp := endTime - 7*24*3600

		// 时间格式对应7日新增api
		sevenDayAddApiMap := make(map[string]int, 7)
		// 时间格式对应7日新增api
		sevenDayAddSceneMap := make(map[string]int, 7)
		// 时间格式对应7日新增用例数量map
		sevenDayAddCaseMap := make(map[string]int, 7)

		for i := 1; i <= 7; i++ {
			timeObj := time.Unix(startTimeTmp, 0)
			//返回string
			dateStr := timeObj.Format("01.02")
			sevenDayAddApiMap[dateStr] = 0
			sevenDayAddSceneMap[dateStr] = 0
			sevenDayAddCaseMap[dateStr] = 0
			startTimeTmp = startTimeTmp + 24*3600
		}

		for _, targetInfo := range targetList {
			targetDataTime := targetInfo.CreatedAt.Unix()

			if targetInfo.TargetType == "api" {
				if targetInfo.Source == consts.TargetSourceApi {
					apiTotalCount++
					if targetDataTime >= startTime && targetDataTime <= endTime {
						apiStartTime := startTime
						for i := 1; i <= 7; i++ {
							if targetDataTime >= apiStartTime && targetDataTime < apiStartTime+(24*3600) {
								timeObj := time.Unix(apiStartTime, 0)
								//返回string
								dateStr := timeObj.Format("01.02")
								sevenDayAddApiMap[dateStr]++
								break
							}
							apiStartTime = apiStartTime + 24*3600
						}
					}
				}

				if targetInfo.Source == consts.TargetSourcePlan && targetInfo.PlanID != "" { // 性能计划
					stressTotalApiNum++
				}

			}

			if targetInfo.TargetType == "scene" {
				if targetInfo.Source == consts.TargetSourceScene {
					scentTotalCount++
					// 获取最近7日内数据
					if targetDataTime >= startTime && targetDataTime <= endTime {
						sceneStartTime := startTime
						for i := 1; i <= 7; i++ {
							if targetDataTime >= sceneStartTime && targetDataTime < sceneStartTime+(24*3600) {
								timeObj := time.Unix(sceneStartTime, 0)
								//返回string
								dateStr := timeObj.Format("01.02")
								sevenDayAddSceneMap[dateStr]++
								break
							}
							sceneStartTime = sceneStartTime + 24*3600
						}
					}
				}

				if targetInfo.Source != consts.TargetSourceScene {
					sceneCiteMap[targetInfo.SourceID]++
				}

				if targetInfo.Source == 2 && targetInfo.PlanID != "" {
					stressTotalSceneNum++
					if targetInfo.SourceID != "" { // 引入进来的场景
						stressPlanTotalImportSceneNum++
					}
				}

				if targetInfo.Source == 3 && targetInfo.PlanID != "" {
					autoPlanTotalSceneNum++
					if targetInfo.SourceID != "" { // 引入进来的场景
						autoPlanTotalImportSceneNum++
					}
				}

			}

			if targetInfo.TargetType == "test_case" { // 用例统计
				if targetInfo.Source == consts.TargetSourceScene || targetInfo.Source == consts.TargetSourceAutoPlan {
					if targetDataTime >= startTime && targetDataTime <= endTime {
						caseStartTime := startTime
						for i := 1; i <= 7; i++ {
							if targetDataTime >= caseStartTime && targetDataTime < caseStartTime+(24*3600) {
								timeObj := time.Unix(caseStartTime, 0)
								//返回string
								dateStr := timeObj.Format("01.02")
								sevenDayAddCaseMap[dateStr]++
								break
							}
							caseStartTime = caseStartTime + 24*3600
						}
					}
				}

				if (targetInfo.Source == consts.TargetSourceAutoPlan) && targetInfo.PlanID != "" {
					currentTeamTotalCaseNum++ //自动化计划里面的用例总数
					autoPlanTestCaseIDs = append(autoPlanTestCaseIDs, targetInfo.TargetID)
				}

			}

		}

		// 查询已经测试过的用例
		sceneCaseReport, err := GetAutoReportByCaseIDs(ctx, autoPlanTestCaseIDs, req.TeamID)
		isTestCaseMap := make(map[string]int)         // 所有已经测试过的用例字典
		failAndNotTestCaseMap := make(map[string]int) // 包含失败接口的测试用例
		for _, caseReportDetail := range sceneCaseReport {
			caseID := caseReportDetail["case_id"].(string)
			isTestCaseMap[caseID]++

			if caseReportDetail["status"] == "failed" {
				failAndNotTestCaseMap[caseID]++
			}
		}

		// 计算未测、未通过率
		failAndNotTestPercent := 0
		passCasePercent := 0
		if len(isTestCaseMap) > 0 {
			failAndNotTestPercent = (len(failAndNotTestCaseMap) * 100) / len(isTestCaseMap)
			passCasePercent = 100 - failAndNotTestPercent
		} else {
			failAndNotTestPercent = 0
			passCasePercent = 100 - failAndNotTestPercent
		}

		// 当前团队所有自动化测试相关数据
		// 查询所有性能计划报告信息
		stressPlanReportCount, err := tx.StressPlanReport.WithContext(ctx).Where(tx.StressPlanReport.TeamID.Eq(req.TeamID)).Count()
		if err != nil {
			return err
		}

		// 查询所有自动化计划报告信息
		autoPlanReportCount, err := tx.AutoPlanReport.WithContext(ctx).Where(tx.AutoPlanReport.TeamID.Eq(req.TeamID)).Count()
		if err != nil {
			return err
		}

		// 查询性能计划下所有普通任务数量
		stressPlanNormalTaskNum, err := tx.StressPlanTaskConf.WithContext(ctx).Where(tx.StressPlanTaskConf.TeamID.Eq(req.TeamID)).Count()
		if err != nil {
			return err
		}
		// 查询性能计划下所有定时任务数量
		stressPlanTimedTaskNum, err := tx.StressPlanTimedTaskConf.WithContext(ctx).Where(tx.StressPlanTimedTaskConf.TeamID.Eq(req.TeamID)).Count()
		if err != nil {
			return err
		}

		// 获取自动化计划下面的报告列表
		autoPlanReportList, err := tx.AutoPlanReport.WithContext(ctx).Where(tx.AutoPlanReport.TeamID.Eq(req.TeamID)).Order(tx.AutoPlanReport.CreatedAt.Desc()).Limit(10).Find()
		if err != nil {
			return err
		}
		runUserIDs := make([]string, 0, len(autoPlanReportList))
		for _, aprInfo := range autoPlanReportList {
			runUserIDs = append(runUserIDs, aprInfo.RunUserID)
		}

		// 获取性能计划下面的报告列表
		stressPlanReportList, err := tx.StressPlanReport.WithContext(ctx).Where(tx.StressPlanReport.TeamID.Eq(req.TeamID)).Order(tx.StressPlanReport.CreatedAt.Desc()).Limit(10).Find()
		if err != nil {
			return err
		}
		for _, sprInfo := range stressPlanReportList {
			runUserIDs = append(runUserIDs, sprInfo.RunUserID)
		}

		userList, err := tx.User.WithContext(ctx).Where(tx.User.UserID.In(runUserIDs...)).Find()
		if err != nil {
			return err
		}
		userIDAndNameMap := make(map[string]string, len(userList))
		for _, userInfo := range userList {
			userIDAndNameMap[userInfo.UserID] = userInfo.Nickname
		}

		autoPlanReportResp := make([]rao.LatelyReportList, 0, len(autoPlanReportList))
		for _, aprInfo := range autoPlanReportList {
			temp := rao.LatelyReportList{
				ReportID:    aprInfo.ReportID,
				RankID:      aprInfo.RankID,
				PlanName:    aprInfo.PlanName,
				TaskType:    aprInfo.TaskType,
				RunUserName: userIDAndNameMap[aprInfo.RunUserID],
				Status:      aprInfo.Status,
			}
			autoPlanReportResp = append(autoPlanReportResp, temp)
		}

		stressPlanReportResp := make([]rao.LatelyReportList, 0, len(stressPlanReportList))
		for _, sprInfo := range stressPlanReportList {
			temp := rao.LatelyReportList{
				ReportID:    sprInfo.ReportID,
				RankID:      sprInfo.RankID,
				PlanName:    sprInfo.PlanName,
				TaskType:    sprInfo.TaskType,
				TaskMode:    sprInfo.TaskMode,
				RunUserName: userIDAndNameMap[sprInfo.RunUserID],
				Status:      sprInfo.Status,
			}
			stressPlanReportResp = append(stressPlanReportResp, temp)
		}

		sumTaskNum := stressPlanNormalTaskNum + stressPlanTimedTaskNum
		normalTaskPercent := 0
		timedTaskPercent := 0
		if sumTaskNum > 0 {
			normalTaskPercent = int((stressPlanNormalTaskNum * 100) / sumTaskNum)
			timedTaskPercent = 100 - normalTaskPercent
		} else {
			normalTaskPercent = 0
			timedTaskPercent = 100 - normalTaskPercent
		}

		// 查询调试日志
		debugLog := tx.TargetDebugLog
		debugLogList, err := debugLog.WithContext(ctx).Where(debugLog.TeamID.Eq(req.TeamID)).Find()
		debugStartTimeTmp := startTime
		sevenDayDebugApiMap := make(map[string]int, len(debugLogList))
		sevenDayDebugSceneMap := make(map[string]int, len(debugLogList))
		for i := 1; i <= 7; i++ {
			timeObj := time.Unix(debugStartTimeTmp, 0)
			//返回string
			dateStr := timeObj.Format("01.02")
			sevenDayDebugApiMap[dateStr] = 0
			sevenDayDebugSceneMap[dateStr] = 0
			debugStartTimeTmp = debugStartTimeTmp + 24*3600
		}

		for _, debugInfo := range debugLogList {
			targetDataTime := debugInfo.CreatedAt.Unix()
			if debugInfo.TargetType == 1 && targetDataTime >= startTime && targetDataTime <= endTime {
				debugApiStartTime := startTime
				for i := 1; i <= 7; i++ {
					if targetDataTime >= debugApiStartTime && targetDataTime < debugApiStartTime+(24*3600) {
						timeObj := time.Unix(debugApiStartTime, 0)
						//返回string
						dateStr := timeObj.Format("01.02")
						sevenDayDebugApiMap[dateStr]++
						break
					}
					debugApiStartTime = debugApiStartTime + 24*3600
				}
			}

			if debugInfo.TargetType == 2 && targetDataTime >= startTime && targetDataTime <= endTime {
				debugSceneStartTime := startTime
				for i := 1; i <= 7; i++ {
					if targetDataTime >= debugSceneStartTime && targetDataTime < debugSceneStartTime+(24*3600) {
						timeObj := time.Unix(debugSceneStartTime, 0)
						//返回string
						dateStr := timeObj.Format("01.02")
						sevenDayDebugSceneMap[dateStr]++
						break
					}
					debugSceneStartTime = debugSceneStartTime + 24*3600
				}
			}
		}

		// 计算被引用接口
		validSiteApiMap := make(map[string]int, len(allSiteApiMap))
		for apiID := range allSiteApiMap {
			if _, ok := allApiMap[apiID]; ok {
				validSiteApiMap[apiID] = 1
			}
		}

		validSiteSceneMap := make(map[string]int, len(allSiteApiMap))
		for sceneID := range sceneCiteMap {
			if _, ok := allSceneIDMap[sceneID]; ok {
				validSiteSceneMap[sceneID] = 1
			}
		}

		res = &rao.HomePageResp{
			TeamName: teamIDNameMap[req.TeamID],
			ApiManageData: rao.ApiManageData{
				ApiCiteCount:  len(validSiteApiMap),
				ApiTotalCount: apiTotalCount,
				ApiDebugCount: sevenDayDebugApiMap,
				ApiAddCount:   sevenDayAddApiMap,
			},
			SceneManageData: rao.SceneManageData{
				SceneCiteCount:  len(validSiteSceneMap),
				SceneTotalCount: scentTotalCount,
				SceneDebugCount: sevenDayDebugSceneMap,
				SceneAddCount:   sevenDayAddSceneMap,
			},
			CaseAddSevenData: sevenDayAddCaseMap,
			TeamOverview:     teamOverview,
			AutoPlanData: rao.AutoPlanData{
				PlanNum:                   autoPlanTotalNumMap[req.TeamID],
				ReportNum:                 autoPlanReportCount,
				CaseTotalNum:              currentTeamTotalCaseNum,
				CaseExecNum:               len(isTestCaseMap),
				CasePassNum:               len(isTestCaseMap) - len(failAndNotTestCaseMap),
				CiteApiNum:                len(autoPlanSiteApiMap),
				TotalApiNum:               autoPlanTotalApiNum,
				CiteSceneNum:              autoPlanTotalImportSceneNum,
				TotalSceneNum:             autoPlanTotalSceneNum,
				CasePassPercent:           float64(passCasePercent),
				CaseNotTestAndPassPercent: float64(failAndNotTestPercent),
				LatelyReportList:          autoPlanReportResp,
			},
			StressPlanData: rao.StressPlanData{
				PlanNum:          stressPlanTotalNumMap[req.TeamID],
				ReportNum:        int(stressPlanReportCount),
				ApiNum:           stressPlanTotalApiNum,
				SceneNum:         stressTotalSceneNum,
				CiteApiNum:       len(stressPlanSiteApiMap),
				TotalApiNum:      stressPlanTotalApiNum,
				CiteSceneNum:     stressPlanTotalImportSceneNum,
				TotalSceneNum:    stressTotalSceneNum,
				TimedPlanNum:     float64(timedTaskPercent),
				NormalPlanNum:    float64(normalTaskPercent),
				LatelyReportList: stressPlanReportResp,
			},
		}

		return nil
	})
	return res, err
}

// GetFlowBySceneIDs 根据场景ID集合，获取对应的flow详情
func GetFlowBySceneIDs(ctx *gin.Context, sceneIDs []string, teamID string) ([]mao.Flow, error) {
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectFlow)
	cur, err := collection.Find(ctx, bson.D{{"team_id", teamID}, {"scene_id", bson.D{{"$in", sceneIDs}}}})
	if err != nil {
		return nil, err
	}

	var flows []mao.Flow
	if err := cur.All(ctx, &flows); err != nil {
		return nil, err
	}
	return flows, nil
}

// GetCaseFlowByCaseIDs 根据用例ID集合，获取对应的flow详情
func GetCaseFlowByCaseIDs(ctx *gin.Context, caseIDs []string, teamID string) ([]mao.SceneCaseFlow, error) {
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectSceneCaseFlow)
	cur, err := collection.Find(ctx, bson.D{{"team_id", teamID}, {"scene_case_id", bson.D{{"$in", caseIDs}}}})
	if err != nil {
		return nil, err
	}

	var caseFlows []mao.SceneCaseFlow
	if err := cur.All(ctx, &caseFlows); err != nil {
		return nil, err
	}
	return caseFlows, nil
}

// GetAutoReportByCaseIDs 根据用例ID集合，获取对应的报告详情
func GetAutoReportByCaseIDs(ctx *gin.Context, caseIDs []string, teamID string) ([]map[string]interface{}, error) {
	// 获取所有用例运行结果
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectAutoReport)
	cur, err := collection.Find(ctx, bson.D{{"team_id", teamID}, {"case_id", bson.D{{"$in", caseIDs}}}, {"type", "api"}})
	if err != nil {
		return nil, fmt.Errorf("获取所有运行用例结果数据为空")
	}
	var sceneCaseReport []map[string]interface{}
	if err := cur.All(ctx, &sceneCaseReport); err != nil {
		return nil, fmt.Errorf("获取所有运行用例结果数据失败")
	}
	return sceneCaseReport, nil
}
