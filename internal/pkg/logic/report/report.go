package report

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/jwt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/log"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strconv"
	"time"

	"github.com/shopspring/decimal"

	"gorm.io/gen/field"

	"github.com/go-omnibus/proof"
	"go.mongodb.org/mongo-driver/bson"
	"gorm.io/gen"

	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/record"

	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/mao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/query"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/packer"
)

func ListByTeamID2(ctx context.Context, teamID string, limit, offset int, keyword string, startTimeSec, endTimeSec int64, taskType, taskMode, status, sortTag int32) ([]*rao.StressPlanReport, int64, error) {

	tx := query.Use(dal.DB()).StressPlanReport

	conditions := make([]gen.Condition, 0)
	conditions = append(conditions, tx.TeamID.Eq(teamID))

	if keyword != "" {
		var reportIDs []string

		planReportIDs, err := KeywordFindPlan(ctx, teamID, keyword)
		if err != nil {
			return nil, 0, err
		}
		reportIDs = append(reportIDs, planReportIDs...)

		sceneReportIDs, err := KeywordFindScene(ctx, teamID, keyword)
		if err != nil {
			return nil, 0, err
		}
		reportIDs = append(reportIDs, sceneReportIDs...)

		userReportIDs, err := KeywordFindUser(ctx, teamID, keyword)
		if err != nil {
			return nil, 0, err
		}
		reportIDs = append(reportIDs, userReportIDs...)

		if len(reportIDs) > 0 {
			conditions = append(conditions, tx.ReportID.In(reportIDs...))
		} else {
			conditions = append(conditions, tx.ReportID.In(""))
		}
	}

	if startTimeSec > 0 && endTimeSec > 0 {
		startTime := time.Unix(startTimeSec, 0)
		endTime := time.Unix(endTimeSec, 0)
		conditions = append(conditions, tx.CreatedAt.Between(startTime, endTime))
	}

	if taskType > 0 {
		conditions = append(conditions, tx.TaskType.Eq(taskType))
	}

	if taskMode > 0 {
		conditions = append(conditions, tx.TaskMode.Eq(taskMode))
	}

	if status > 0 {
		conditions = append(conditions, tx.Status.Eq(status))
	}

	sort := make([]field.Expr, 0)

	if sortTag == 0 { // 默认排序
		sort = append(sort, tx.CreatedAt.Desc())
	}
	if sortTag == 1 { // 创建时间倒序
		sort = append(sort, tx.CreatedAt.Desc())
	}
	if sortTag == 2 { // 创建时间正序
		sort = append(sort, tx.CreatedAt)
	}
	if sortTag == 3 { // 修改时间倒序
		sort = append(sort, tx.UpdatedAt.Desc())
	}
	if sortTag == 4 { // 修改时间正序
		sort = append(sort, tx.UpdatedAt)
	}

	reports, cnt, err := tx.WithContext(ctx).Where(conditions...).
		Order(sort...).
		FindByPage(offset, limit)

	if err != nil {
		return nil, 0, err
	}

	var userIDs []string
	for _, r := range reports {
		userIDs = append(userIDs, r.RunUserID)
	}

	u := query.Use(dal.DB()).User
	users, err := u.WithContext(ctx).Where(u.UserID.In(userIDs...)).Find()
	if err != nil {
		return nil, 0, err
	}

	return packer.TransReportModelToRaoReportList(reports, users), cnt, nil
}

func KeywordFindPlan(ctx context.Context, teamID string, keyword string) ([]string, error) {
	var planIDs []string

	p := dal.GetQuery().StressPlan
	err := p.WithContext(ctx).Where(p.TeamID.Eq(teamID), p.PlanName.Like(fmt.Sprintf("%%%s%%", keyword))).Pluck(p.PlanID, &planIDs)
	if err != nil {
		return nil, err
	}

	if len(planIDs) == 0 {
		return nil, nil
	}

	var reportIDs []string
	r := dal.GetQuery().StressPlanReport
	err = r.WithContext(ctx).Where(r.TeamID.Eq(teamID), r.PlanID.In(planIDs...)).Pluck(r.ReportID, &reportIDs)
	if err != nil {
		return nil, err
	}

	return reportIDs, nil
}

func KeywordFindScene(ctx context.Context, teamID string, keyword string) ([]string, error) {
	var sceneIDs []string

	s := dal.GetQuery().Target
	err := s.WithContext(ctx).Where(s.TeamID.Eq(teamID), s.Name.Like(fmt.Sprintf("%%%s%%", keyword))).Pluck(s.TargetID, &sceneIDs)
	if err != nil {
		return nil, err
	}

	if len(sceneIDs) == 0 {
		return nil, nil
	}

	var reportIDs []string
	r := dal.GetQuery().StressPlanReport
	err = r.WithContext(ctx).Where(r.TeamID.Eq(teamID), r.SceneID.In(sceneIDs...)).Pluck(r.ReportID, &reportIDs)
	if err != nil {
		return nil, err
	}

	return reportIDs, nil
}

func KeywordFindUser(ctx context.Context, teamID string, keyword string) ([]string, error) {
	var userIDs []string

	u := query.Use(dal.DB()).User
	err := u.WithContext(ctx).Where(u.Nickname.Like(fmt.Sprintf("%%%s%%", keyword))).Pluck(u.UserID, &userIDs)
	if err != nil {
		return nil, err
	}

	if len(userIDs) == 0 {
		return nil, nil
	}

	var reportIDs []string
	r := dal.GetQuery().StressPlanReport
	err = r.WithContext(ctx).Where(r.TeamID.Eq(teamID), r.RunUserID.In(userIDs...)).Pluck(r.ReportID, &reportIDs)
	if err != nil {
		return nil, err
	}

	return reportIDs, nil
}

func DeleteReport(ctx context.Context, req *rao.DeleteReportReq, userID string) error {
	allErr := dal.GetQuery().Transaction(func(tx *query.Query) error {
		reportInfo, err := tx.StressPlanReport.WithContext(ctx).Where(tx.StressPlanReport.TeamID.Eq(req.TeamID), tx.StressPlanReport.ReportID.Eq(req.ReportID)).First()
		if err != nil {
			return err
		}

		if reportInfo.Status == consts.ReportStatusNormal {
			return fmt.Errorf("运行中的报告不能删除")
		}

		if _, err := tx.StressPlanReport.WithContext(ctx).Where(tx.StressPlanReport.TeamID.Eq(req.TeamID), tx.StressPlanReport.ReportID.Eq(req.ReportID)).Delete(); err != nil {
			return err
		}

		// 删除报告对应机器信息
		if _, err := tx.ReportMachine.WithContext(ctx).Where(tx.ReportMachine.TeamID.Eq(reportInfo.TeamID),
			tx.ReportMachine.PlanID.Eq(reportInfo.PlanID),
			tx.ReportMachine.ReportID.Eq(reportInfo.ReportID)).Delete(); err != nil {
			return err
		}

		// 把mongodb库里面的报告详情数据删掉
		collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectReportData)
		findFilter := bson.D{{"report_id", req.ReportID}, {"team_id", req.TeamID}}
		_, err = collection.DeleteOne(ctx, findFilter)
		if err != nil {
			log.Logger.Info("report_data：删除失败")
		}

		// 把mongodb库里面的报告任务配置详情数据删掉
		collection = dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectReportTask)
		findFilter = bson.D{{"report_id", req.ReportID}, {"team_id", req.TeamID}}
		_, err = collection.DeleteOne(ctx, findFilter)
		if err != nil {
			log.Logger.Info("report_task：删除失败")
		}

		// 把mongodb库里面的编辑报告详情数据删掉
		collection = dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectChangeReportConf)
		findFilter = bson.D{{"report_id", req.ReportID}, {"team_id", req.TeamID}}
		_, err = collection.DeleteOne(ctx, findFilter)
		if err != nil {
			log.Logger.Info("change_report_conf：删除失败")
		}

		if err := record.InsertDelete(ctx, req.TeamID, userID, record.OperationOperateDeleteReport, fmt.Sprintf("%s %s", reportInfo.PlanName, reportInfo.SceneName)); err != nil {
			return err
		}
		return nil
	})

	if allErr != nil {
		log.Logger.Info("DeleteReport：删除失败")
		return allErr
	}

	return nil
}

func GetTaskDetail(ctx context.Context, req rao.GetReportTaskDetailReq) (*rao.ReportTask, error) {
	// 查询报告是否被删除
	tx := dal.GetQuery().StressPlanReport
	reportInfo, err := tx.WithContext(ctx).Where(tx.TeamID.Eq(req.TeamID), tx.ReportID.Eq(req.ReportID)).First()
	if err != nil {
		log.Logger.Info("报告详情--查询报告基本信息失败，err:")
		errNew := fmt.Errorf("报告不存在")
		return nil, errNew
	}

	detail := mao.ReportTask{}
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectReportTask)

	err = collection.FindOne(ctx, bson.D{{"report_id", req.ReportID}}).Decode(&detail)
	if err != nil {
		log.Logger.Info("mongo decode err", proof.WithError(err))
		return nil, err
	}

	r := query.Use(dal.DB()).StressPlanReport
	ru, err := r.WithContext(ctx).Where(r.TeamID.Eq(req.TeamID), r.ReportID.Eq(req.ReportID)).First()
	if err != nil {
		log.Logger.Info("req not found err", proof.WithError(err))
		return nil, err
	}

	u := query.Use(dal.DB()).User
	user, err := u.WithContext(ctx).Where(u.UserID.Eq(ru.RunUserID)).First()
	if err != nil {
		log.Logger.Info("user not found err", proof.WithError(err))
		return nil, err
	}

	// 从mongo查出编辑报告的数据列表
	collection = dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectChangeReportConf)
	ChangeTaskConfDetail, _ := collection.Find(ctx, bson.D{{"report_id", req.ReportID}})

	changeTaskConf := make([]mao.ChangeTaskConf, 0, 10)
	if err := ChangeTaskConfDetail.All(ctx, &changeTaskConf); err != nil {
		log.Logger.Info("没有查到编辑报告列表数据", proof.WithError(err))
	}

	modeConf := rao.ModeConf{
		RoundNum:         detail.ModeConf.RoundNum,
		Concurrency:      detail.ModeConf.Concurrency,
		ThresholdValue:   detail.ModeConf.ThresholdValue,
		StartConcurrency: detail.ModeConf.StartConcurrency,
		Step:             detail.ModeConf.Step,
		StepRunTime:      detail.ModeConf.StepRunTime,
		MaxConcurrency:   detail.ModeConf.MaxConcurrency,
		Duration:         detail.ModeConf.Duration,
		CreatedTimeSec:   ru.CreatedAt.Unix(),
	}

	usableMachineList := make([]rao.UsableMachineInfo, 0, len(detail.MachineDispatchModeConf.UsableMachineList))
	for _, v := range detail.MachineDispatchModeConf.UsableMachineList {
		temp := rao.UsableMachineInfo{
			MachineStatus:    v.MachineStatus,
			MachineName:      v.MachineName,
			Region:           v.Region,
			Ip:               v.Ip,
			Weight:           v.Weight,
			RoundNum:         v.RoundNum,
			Concurrency:      v.Concurrency,
			ThresholdValue:   v.ThresholdValue,
			StartConcurrency: v.StartConcurrency,
			Step:             v.Step,
			StepRunTime:      v.StepRunTime,
			MaxConcurrency:   v.MaxConcurrency,
			Duration:         v.Duration,
			CreatedTimeSec:   v.CreatedTimeSec,
		}
		usableMachineList = append(usableMachineList, temp)
	}

	changeTaskConfData := rao.ChangeTakeConf{
		RoundNum:          detail.ModeConf.RoundNum,
		Concurrency:       detail.ModeConf.Concurrency,
		ThresholdValue:    detail.ModeConf.ThresholdValue,
		StartConcurrency:  detail.ModeConf.StartConcurrency,
		Step:              detail.ModeConf.Step,
		StepRunTime:       detail.ModeConf.StepRunTime,
		MaxConcurrency:    detail.ModeConf.MaxConcurrency,
		Duration:          detail.ModeConf.Duration,
		CreatedTimeSec:    ru.CreatedAt.Unix(),
		UsableMachineList: usableMachineList,
	}

	res := &rao.ReportTask{
		UserID:            user.UserID,
		UserName:          user.Nickname,
		UserAvatar:        user.Avatar,
		PlanID:            detail.PlanID,
		PlanName:          detail.PlanName,
		ReportID:          detail.ReportID,
		ReportName:        reportInfo.ReportName,
		SceneID:           ru.SceneID,
		SceneName:         ru.SceneName,
		CreatedTimeSec:    ru.CreatedAt.Unix(),
		TaskType:          detail.TaskType,
		TaskMode:          detail.TaskMode,
		ControlMode:       detail.ControlMode,
		DebugMode:         detail.DebugMode,
		TaskStatus:        ru.Status,
		ModeConf:          modeConf,
		IsOpenDistributed: detail.IsOpenDistributed,
		MachineAllotType:  detail.MachineDispatchModeConf.MachineAllotType,
	}

	res.ChangeTakeConf = append(res.ChangeTakeConf, changeTaskConfData)

	if len(changeTaskConf) > 0 {
		for _, changeTaskConfTmp := range changeTaskConf {
			usableMachineListTemp := make([]rao.UsableMachineInfo, 0, len(changeTaskConfTmp.MachineDispatchModeConf.UsableMachineList))
			for _, v := range changeTaskConfTmp.MachineDispatchModeConf.UsableMachineList {
				temp := rao.UsableMachineInfo{
					MachineStatus:    v.MachineStatus,
					MachineName:      v.MachineName,
					Region:           v.Region,
					Ip:               v.Ip,
					Weight:           v.Weight,
					RoundNum:         v.RoundNum,
					Concurrency:      v.Concurrency,
					ThresholdValue:   v.ThresholdValue,
					StartConcurrency: v.StartConcurrency,
					Step:             v.Step,
					StepRunTime:      v.StepRunTime,
					MaxConcurrency:   v.MaxConcurrency,
					Duration:         v.Duration,
					CreatedTimeSec:   v.CreatedTimeSec,
				}
				usableMachineListTemp = append(usableMachineListTemp, temp)
			}

			tmp := rao.ChangeTakeConf{
				RoundNum:          changeTaskConfTmp.ModeConf.RoundNum,
				Concurrency:       changeTaskConfTmp.ModeConf.Concurrency,
				ThresholdValue:    changeTaskConfTmp.ModeConf.ThresholdValue,
				StartConcurrency:  changeTaskConfTmp.ModeConf.StartConcurrency,
				Step:              changeTaskConfTmp.ModeConf.Step,
				StepRunTime:       changeTaskConfTmp.ModeConf.StepRunTime,
				MaxConcurrency:    changeTaskConfTmp.ModeConf.MaxConcurrency,
				Duration:          changeTaskConfTmp.ModeConf.Duration,
				CreatedTimeSec:    changeTaskConfTmp.ModeConf.CreatedTimeSec,
				UsableMachineList: usableMachineListTemp,
			}
			res.ChangeTakeConf = append(res.ChangeTakeConf, tmp)
		}
	}
	return res, nil
}

func GetReportDebugStatus(ctx context.Context, req rao.GetReportReq) string {
	reportId := req.ReportID
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectDebugStatus)
	filter := bson.D{{"report_id", reportId}, {"team_id", req.TeamID}, {"plan_id", req.PlanID}}
	//fmt.Println("filter:", filter)
	cur := collection.FindOne(ctx, filter)
	result, err := cur.DecodeBytes()
	if err != nil {
		return consts.StopDebug
	}
	list, err := result.Elements()
	if err != nil {
		return consts.StopDebug
	}
	for _, value := range list {
		if value.Key() == "debug" {
			return value.Value().StringValue()
		}
	}
	return consts.StopDebug
}

func GetReportDebugLog(ctx context.Context, report rao.GetReportReq) []map[string]interface{} {
	var limit int64 = 1000
	debugMsgList := make([]map[string]interface{}, 0, limit)

	// 查询运行报告时，是否开启debug模式
	tx := dal.GetQuery().StressPlanReport
	reportInfo, err := tx.WithContext(ctx).Where(tx.ReportID.Eq(report.ReportID)).First()
	if err != nil || reportInfo.DebugMode == "stop" {
		return debugMsgList
	}

	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectStressDebug)
	filter := bson.D{{"report_id", report.ReportID}}
	findOptions := options.Find().SetSort(bson.M{"response_time": -1}).SetLimit(limit).SetAllowDiskUse(true)
	cur, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		log.Logger.Info("debug日志查询失败", err)
		return debugMsgList
	}
	for cur.Next(ctx) {
		debugMsg := make(map[string]interface{})
		err = cur.Decode(&debugMsg)
		if err != nil {
			log.Logger.Info("debug日志转换失败", err)
			return debugMsgList
		}
		if debugMsg["end"] != true {
			// 删除无用的字段
			delete(debugMsg, "_id")
			delete(debugMsg, "api_id")
			delete(debugMsg, "assert_failed_num")
			delete(debugMsg, "assert_num")
			delete(debugMsg, "case_id")
			delete(debugMsg, "next_list")
			delete(debugMsg, "parent_id")
			delete(debugMsg, "plan_id")
			delete(debugMsg, "report_id")
			delete(debugMsg, "scene_id")
			delete(debugMsg, "team_id")
			delete(debugMsg, "type")
			//delete(debugMsg, "uuid")
			debugMsgList = append(debugMsgList, debugMsg)
		}
	}

	// 限制debug日志数量
	if len(debugMsgList) > int(limit) {
		startIndex := len(debugMsgList) - int(limit)
		debugMsgList = debugMsgList[startIndex:]
	}
	return debugMsgList
}

// GetReportDetail 从redis获取测试数据
func GetReportDetail(ctx context.Context, req rao.GetReportReq) (ResultData, error) {
	resultData := ResultData{}
	// 查询报告是否被删除
	tx := dal.GetQuery().StressPlanReport
	reportBaseInfo, err := tx.WithContext(ctx).Where(tx.TeamID.Eq(req.TeamID), tx.ReportID.Eq(req.ReportID)).First()
	if err != nil {
		log.Logger.Info("报告详情--查询报告基本信息失败，err:", err)
		err = fmt.Errorf("报告不存在")
		return resultData, err
	}

	// 查询报告详情数据
	resultMsg := SceneTestResultDataMsg{}
	dataMap := make(map[string]interface{}, 0)
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectReportData)
	filter := bson.D{{"team_id", req.TeamID}, {"report_id", req.ReportID}}
	err = collection.FindOne(ctx, filter).Decode(&dataMap)
	_, ok := dataMap["data"]
	if err != nil || !ok {
		log.Logger.Info("mango数据为空，开始查询redis")
		rdb := dal.GetRDBForReport()
		key := fmt.Sprintf("reportData:%s", req.ReportID)
		dataList := rdb.LRange(ctx, key, 0, -1).Val()
		log.Logger.Info("查询redis报告数据，报告数据的Key:", key, "，数组长度为：", len(dataList), dataList)
		if len(dataList) < 1 {
			log.Logger.Info("redis里面没有查到报告详情数据")
			err = nil
			return resultData, err
		}
		for i := len(dataList) - 1; i >= 0; i-- {
			resultMsgString := dataList[i]
			err = json.Unmarshal([]byte(resultMsgString), &resultMsg)
			if err != nil {
				log.Logger.Info("json转换格式错误：", err)
				continue
			}
			if resultData.Results == nil {
				resultData.Results = make(map[string]*ResultDataMsg, len(dataList))
			}
			resultData.ReportId = resultMsg.ReportId
			resultData.End = resultMsg.End
			resultData.ReportName = resultMsg.ReportName
			resultData.PlanId = resultMsg.PlanId
			resultData.PlanName = resultMsg.PlanName
			resultData.SceneId = resultMsg.SceneId
			resultData.SceneName = resultMsg.SceneName
			resultData.TimeStamp = resultMsg.TimeStamp
			if len(resultMsg.Results) > 0 {
				for k, apiResult := range resultMsg.Results {
					if resultData.Results[k] == nil {
						resultData.Results[k] = new(ResultDataMsg)
					}
					resultData.Results[k].ApiName = apiResult.Name
					resultData.Results[k].Concurrency = apiResult.Concurrency
					resultData.Results[k].TotalRequestNum = apiResult.TotalRequestNum
					resultData.Results[k].TotalRequestTime, _ = decimal.NewFromFloat(float64(apiResult.TotalRequestTime) / float64(time.Second)).Round(2).Float64()
					resultData.Results[k].SuccessNum = apiResult.SuccessNum
					resultData.Results[k].ErrorNum = apiResult.ErrorNum
					if apiResult.TotalRequestNum != 0 {
						errRate := float64(apiResult.ErrorNum) / float64(apiResult.TotalRequestNum)
						resultData.Results[k].ErrorRate, _ = strconv.ParseFloat(fmt.Sprintf("%0.2f", errRate), 64)
					}
					resultData.Results[k].PercentAge = apiResult.PercentAge
					resultData.Results[k].ErrorThreshold = apiResult.ErrorThreshold
					resultData.Results[k].ResponseThreshold = apiResult.ResponseThreshold
					resultData.Results[k].RequestThreshold = apiResult.RequestThreshold
					resultData.Results[k].AvgRequestTime, _ = decimal.NewFromFloat(apiResult.AvgRequestTime / float64(time.Millisecond)).Round(1).Float64()
					resultData.Results[k].MaxRequestTime, _ = decimal.NewFromFloat(apiResult.MaxRequestTime / float64(time.Millisecond)).Round(1).Float64()
					resultData.Results[k].MinRequestTime, _ = decimal.NewFromFloat(apiResult.MinRequestTime / float64(time.Millisecond)).Round(1).Float64()
					resultData.Results[k].CustomRequestTimeLine = apiResult.CustomRequestTimeLine
					resultData.Results[k].CustomRequestTimeLineValue, _ = decimal.NewFromFloat(apiResult.CustomRequestTimeLineValue / float64(time.Millisecond)).Round(1).Float64()
					resultData.Results[k].FiftyRequestTimelineValue, _ = decimal.NewFromFloat(apiResult.FiftyRequestTimelineValue / float64(time.Millisecond)).Round(1).Float64()
					resultData.Results[k].NinetyRequestTimeLine = apiResult.NinetyRequestTimeLine
					resultData.Results[k].NinetyRequestTimeLineValue, _ = decimal.NewFromFloat(apiResult.NinetyRequestTimeLineValue / float64(time.Millisecond)).Round(1).Float64()
					resultData.Results[k].NinetyFiveRequestTimeLine = apiResult.NinetyFiveRequestTimeLine
					resultData.Results[k].NinetyFiveRequestTimeLineValue, _ = decimal.NewFromFloat(apiResult.NinetyFiveRequestTimeLineValue / float64(time.Millisecond)).Round(1).Float64()
					resultData.Results[k].NinetyNineRequestTimeLine = apiResult.NinetyNineRequestTimeLine
					resultData.Results[k].NinetyNineRequestTimeLineValue, _ = decimal.NewFromFloat(apiResult.NinetyNineRequestTimeLineValue / float64(time.Millisecond)).Round(1).Float64()
					resultData.Results[k].SendBytes, _ = decimal.NewFromFloat(apiResult.SendBytes).Round(1).Float64()
					resultData.Results[k].ReceivedBytes, _ = decimal.NewFromFloat(apiResult.ReceivedBytes).Round(1).Float64()
					resultData.Results[k].Rps = apiResult.Rps
					resultData.Results[k].SRps = apiResult.SRps
					resultData.Results[k].Tps = apiResult.Tps
					resultData.Results[k].STps = apiResult.STps
					if resultData.Results[k].RpsList == nil {
						resultData.Results[k].RpsList = make([]TimeValue, 0)
					}

					timeValue := TimeValue{}
					timeValue.TimeStamp = resultData.TimeStamp
					// qps列表
					timeValue.Value = resultData.Results[k].Rps
					resultData.Results[k].RpsList = append(resultData.Results[k].RpsList, timeValue)
					timeValue.Value = resultData.Results[k].Tps
					if resultData.Results[k].TpsList == nil {
						resultData.Results[k].TpsList = make([]TimeValue, 0)
					}
					// 错误数列表
					resultData.Results[k].TpsList = append(resultData.Results[k].TpsList, timeValue)
					timeValue.Value = resultData.Results[k].Concurrency
					if resultData.Results[k].ConcurrencyList == nil {
						resultData.Results[k].ConcurrencyList = make([]TimeValue, 0)
					}
					// 并发数列表
					resultData.Results[k].ConcurrencyList = append(resultData.Results[k].ConcurrencyList, timeValue)

					// 平均响应时间列表
					timeValue.Value = resultData.Results[k].AvgRequestTime
					if resultData.Results[k].AvgList == nil {
						resultData.Results[k].AvgList = make([]TimeValue, 0)
					}
					resultData.Results[k].AvgList = append(resultData.Results[k].AvgList, timeValue)

					// 50响应时间列表
					timeValue.Value = resultData.Results[k].FiftyRequestTimelineValue
					if resultData.Results[k].FiftyList == nil {
						resultData.Results[k].FiftyList = make([]TimeValue, 0)
					}
					resultData.Results[k].FiftyList = append(resultData.Results[k].FiftyList, timeValue)

					// 90响应时间列表
					timeValue.Value = resultData.Results[k].NinetyRequestTimeLineValue
					if resultData.Results[k].NinetyList == nil {
						resultData.Results[k].NinetyList = make([]TimeValue, 0)
					}
					resultData.Results[k].NinetyList = append(resultData.Results[k].NinetyList, timeValue)

					// 95响应时间列表
					timeValue.Value = resultData.Results[k].NinetyFiveRequestTimeLineValue
					if resultData.Results[k].NinetyFiveList == nil {
						resultData.Results[k].NinetyFiveList = make([]TimeValue, 0)
					}
					resultData.Results[k].NinetyFiveList = append(resultData.Results[k].NinetyFiveList, timeValue)

					// 99响应时间列表
					timeValue.Value = resultData.Results[k].NinetyNineRequestTimeLineValue
					if resultData.Results[k].NinetyNineList == nil {
						resultData.Results[k].NinetyNineList = make([]TimeValue, 0)
					}
					resultData.Results[k].NinetyNineList = append(resultData.Results[k].NinetyNineList, timeValue)
				}
			}
			if resultMsg.End {
				log.Logger.Info("报告end为true")
				by := make([]byte, 0)
				by, err = json.Marshal(resultData)
				if err != nil {
					log.Logger.Info("resultData转字节失败：：    ", err)
					return resultData, err
				}
				var apiResultTotalMsg = make(map[string]string)
				for _, value := range resultData.Results {
					apiResultTotalMsg[value.ApiName] = fmt.Sprintf("平均响应时间为%0.1fms； 百分之五十响应时间线的值为%0.1fms; 百分之九十响应时间线的值为%0.1fms; 百分之九十五响应时间线的值为%0.1fms; 百分之九十九响应时间线的值为%0.1fms; RPS为%0.1f; SRPS为%0.1f; TPS为%0.1f; STPS为%0.1f",
						value.AvgRequestTime, value.FiftyRequestTimelineValue, value.NinetyRequestTimeLineValue, value.NinetyFiveRequestTimeLineValue, value.NinetyNineRequestTimeLineValue, value.Rps, value.SRps, value.Tps, value.STps)
				}
				dataMap["report_id"] = resultData.ReportId
				dataMap["team_id"] = req.TeamID
				dataMap["plan_id"] = reportBaseInfo.PlanID
				dataMap["data"] = string(by)
				by, _ = json.Marshal(apiResultTotalMsg)
				dataMap["analysis"] = string(by)
				dataMap["description"] = ""
				_, err = collection.InsertOne(ctx, dataMap)
				if err != nil {
					log.Logger.Info("测试数据写入mongo失败:", err)
					return resultData, err
				}
				err = rdb.Del(ctx, key).Err()
				if err != nil {
					log.Logger.Info(fmt.Sprintf("删除redis的key：%s:", key), err)
					return resultData, err
				}
			}
		}
	} else {
		log.Logger.Info("从mongo查到了数据，直接返回结果")
		data := dataMap["data"].(string)
		err = json.Unmarshal([]byte(data), &resultData)
		resultData.Analysis = dataMap["analysis"].(string)
		resultData.Description = dataMap["description"].(string)
		return resultData, err
	}
	err = nil
	return resultData, err
}

type SceneTestResultDataMsg struct {
	End        bool                             `json:"end" bson:"end"`
	ReportId   string                           `json:"report_id" bson:"report_id"`
	ReportName string                           `json:"report_name" bson:"report_name"`
	PlanId     string                           `json:"plan_id" bson:"plan_id"`     // 任务ID
	PlanName   string                           `json:"plan_name" bson:"plan_name"` //
	SceneId    string                           `json:"scene_id" bson:"scene_id"`   // 场景
	SceneName  string                           `json:"scene_name" bson:"scene_name"`
	Results    map[string]*ApiTestResultDataMsg `json:"results" bson:"results"`
	Machine    map[string]int64                 `json:"machine" bson:"machine"`
	TimeStamp  int64                            `json:"time_stamp" bson:"time_stamp"`
}

// ApiTestResultDataMsg 接口测试数据经过计算后的测试结果
type ApiTestResultDataMsg struct {
	Name                           string  `json:"name" bson:"name"`
	Concurrency                    int64   `json:"concurrency" bson:"concurrency"`
	TotalRequestNum                uint64  `json:"total_request_num" bson:"total_request_num"`   // 总请求数
	TotalRequestTime               uint64  `json:"total_request_time" bson:"total_request_time"` // 总响应时间
	SuccessNum                     uint64  `json:"success_num" bson:"success_num"`
	ErrorNum                       uint64  `json:"error_num" bson:"error_num"`                   // 错误数
	ErrorThreshold                 float64 `json:"error_threshold" bson:"error_threshold"`       // 自定义错误率
	RequestThreshold               int64   `json:"request_threshold" bson:"request_threshold"`   // Rps（每秒请求数）阈值
	ResponseThreshold              int64   `json:"response_threshold" bson:"response_threshold"` // 响应时间阈值
	PercentAge                     int64   `json:"percent_age" bson:"percent_age"`               // 响应时间线
	AvgRequestTime                 float64 `json:"avg_request_time" bson:"avg_request_time"`     // 平均响应事件
	MaxRequestTime                 float64 `json:"max_request_time" bson:"max_request_time"`
	MinRequestTime                 float64 `json:"min_request_time" bson:"min_request_time"` // 毫秒
	CustomRequestTimeLine          int64   `json:"custom_request_time_line" bson:"custom_request_time_line"`
	FiftyRequestTimeline           int64   `json:"fifty_request_time_line" bson:"fifty_request_time_line"`
	NinetyRequestTimeLine          int64   `json:"ninety_request_time_line" bson:"ninety_request_time_line"`
	NinetyFiveRequestTimeLine      int64   `json:"ninety_five_request_time_line" bson:"ninety_five_request_time_line"`
	NinetyNineRequestTimeLine      int64   `json:"ninety_nine_request_time_line" bson:"ninety_nine_request_time_line"`
	FiftyRequestTimelineValue      float64 `json:"fifty_request_time_line_value"`
	CustomRequestTimeLineValue     float64 `json:"custom_request_time_line_value" bson:"custom_request_time_line_value"`
	NinetyRequestTimeLineValue     float64 `json:"ninety_request_time_line_value" bson:"ninety_request_time_line_value"`
	NinetyFiveRequestTimeLineValue float64 `json:"ninety_five_request_time_line_value" bson:"ninety_five_request_time_line_value"`
	NinetyNineRequestTimeLineValue float64 `json:"ninety_nine_request_time_line_value" bson:"ninety_nine_request_time_line_value"`
	SendBytes                      float64 `json:"send_bytes" bson:"send_bytes"`         // 发送字节数
	ReceivedBytes                  float64 `json:"received_bytes" bson:"received_bytes"` // 接收字节数
	Rps                            float64 `json:"rps" bson:"rps"`
	SRps                           float64 `json:"srps" bson:"srps"`
	Tps                            float64 `json:"tps" bson:"tps"`
	STps                           float64 `json:"stps" bson:"stps"`
	ApiName                        string  `json:"api_name" bson:"api_name"`
	ErrorRate                      float64 `json:"error_rate" bson:"error_rate"`
}

// ResultDataMsg 前端展示各个api数据
type ResultDataMsg struct {
	ApiName                        string      `json:"api_name" bson:"api_name"`
	Concurrency                    int64       `json:"concurrency" bson:"concurrency"`
	TotalRequestNum                uint64      `json:"total_request_num" bson:"total_request_num"`   // 总请求数
	TotalRequestTime               float64     `json:"total_request_time" bson:"total_request_time"` // 总响应时间
	SuccessNum                     uint64      `json:"success_num" bson:"success_num"`
	ErrorRate                      float64     `json:"error_rate" bson:"error_rate"`
	ErrorNum                       uint64      `json:"error_num" bson:"error_num"`               // 错误数
	AvgRequestTime                 float64     `json:"avg_request_time" bson:"avg_request_time"` // 平均响应事件
	MaxRequestTime                 float64     `json:"max_request_time" bson:"max_request_time"`
	MinRequestTime                 float64     `json:"min_request_time" bson:"min_request_time"`     // 毫秒
	PercentAge                     int64       `json:"percent_age" bson:"percent_age"`               // 响应时间线
	ErrorThreshold                 float64     `json:"error_threshold" bson:"error_threshold"`       // 自定义错误率
	RequestThreshold               int64       `json:"request_threshold" bson:"request_threshold"`   // Rps（每秒请求数）阈值
	ResponseThreshold              int64       `json:"response_threshold" bson:"response_threshold"` // 响应时间阈值
	CustomRequestTimeLine          int64       `json:"custom_request_time_line" bson:"custom_request_time_line"`
	FiftyRequestTimeline           int64       `json:"fifty_request_time_line" bson:"fifty_request_time_line"`
	NinetyRequestTimeLine          int64       `json:"ninety_request_time_line" bson:"ninety_request_time_line"`
	NinetyFiveRequestTimeLine      int64       `json:"ninety_five_request_time_line" bson:"ninety_five_request_time_line"`
	NinetyNineRequestTimeLine      int64       `json:"ninety_nine_request_time_line" bson:"ninety_nine_request_time_line"`
	CustomRequestTimeLineValue     float64     `json:"custom_request_time_line_value" bson:"custom_request_time_line_value"`
	FiftyRequestTimelineValue      float64     `json:"fifty_request_time_line_value" bson:"fifty_request_time_line_value"`
	NinetyRequestTimeLineValue     float64     `json:"ninety_request_time_line_value" bson:"ninety_request_time_line_value"`
	NinetyFiveRequestTimeLineValue float64     `json:"ninety_five_request_time_line_value" bson:"ninety_five_request_time_line_value"`
	NinetyNineRequestTimeLineValue float64     `json:"ninety_nine_request_time_line_value" bson:"ninety_nine_request_time_line_value"`
	SendBytes                      float64     `json:"send_bytes" bson:"send_bytes"`         // 发送字节数
	ReceivedBytes                  float64     `json:"received_bytes" bson:"received_bytes"` // 接收字节数
	Rps                            float64     `json:"rps" bson:"rps"`
	SRps                           float64     `json:"srps" bson:"srps"`
	Tps                            float64     `json:"tps" bson:"tps"`
	STps                           float64     `json:"stps" bson:"stps"`
	ConcurrencyList                []TimeValue `json:"concurrency_list" bson:"concurrency_list"`
	RpsList                        []TimeValue `json:"rps_list" bson:"rps_list"`
	TpsList                        []TimeValue `json:"tps_list" bson:"tps_list"`
	AvgList                        []TimeValue `json:"avg_list" bson:"avg_list"`
	FiftyList                      []TimeValue `json:"fifty_list" bson:"fifty_list"`
	NinetyList                     []TimeValue `json:"ninety_list" bson:"ninety_list"`
	NinetyFiveList                 []TimeValue `json:"ninety_five_list" bson:"ninety_five_list"`
	NinetyNineList                 []TimeValue `json:"ninety_nine_list" bson:"ninety_nine_list"`
}

type reportDataFmt struct {
	ReportID string `json:"report_id" bson:"report_id"`
	//Data     ResultData `json:"data" bson:"data"`
	Data string `json:"data" bson:"data"`
}
type ResultData struct {
	End         bool                      `json:"end" bson:"end"`
	ReportId    string                    `json:"report_id" bson:"report_id"`
	ReportName  string                    `json:"report_name" bson:"report_name"`
	PlanId      string                    `json:"plan_id" bson:"plan_id"`     // 任务ID
	PlanName    string                    `json:"plan_name" bson:"plan_name"` //
	SceneId     string                    `json:"scene_id" bson:"scene_id"`   // 场景
	SceneName   string                    `json:"scene_name" bson:"scene_name"`
	Results     map[string]*ResultDataMsg `json:"results" bson:"results"`
	TimeStamp   int64                     `json:"time_stamp" bson:"time_stamp"`
	Analysis    string                    `json:"analysis" bson:"analysis"`
	Description string                    `json:"description" bson:"description"`
	Msg         string                    `json:"msg" bson:"msg"`
}

type TimeValue struct {
	TimeStamp int64       `json:"time_stamp" bson:"time_stamp"`
	Value     interface{} `json:"value" bson:"value"`
}

// GetCompareReportData 获取报告对比数据
func GetCompareReportData(ctx context.Context, req rao.CompareReportReq) (*CompareReportResponse, error) {
	// 获取报告的基本信息
	reportTable := dal.GetQuery().StressPlanReport
	reportBaseList, err := reportTable.WithContext(ctx).Where(reportTable.TeamID.Eq(req.TeamID), reportTable.ReportID.In(req.ReportIDs...)).Find()
	if err != nil {
		return nil, err
	}

	reportNames := make([]string, 0, len(reportBaseList))                                 // 计划和场景名字
	reportBaseResponse := make([]*reportBaseFormat, 0, len(reportBaseList))               // 报告基本信息
	reportCollectAllDataResponse := make([]*reportCollectAllData, 0, len(reportBaseList)) // 报告汇总信息
	reportDetailAllDataResponse := make([]reportDetailAllData, 0, len(reportBaseList))    // 报告详情信息

	// 用户id集合
	var runUserIds []string
	reportNamesMap := make(map[string]string, len(req.ReportIDs))
	reportUserMap := make(map[string]string, len(req.ReportIDs))
	reportRunTimeMap := make(map[string]time.Time, len(req.ReportIDs))
	for _, reportBaseInfo := range reportBaseList {
		// 把报告基本信息设置到map当中
		planAndSceneName := reportBaseInfo.PlanName + "/" + reportBaseInfo.SceneName
		reportNames = append(reportNames, planAndSceneName)
		// 组装报告id和计划/场景名映射
		reportNamesMap[reportBaseInfo.ReportID] = planAndSceneName
		// 组装报告和运行人id的映射
		reportUserMap[reportBaseInfo.ReportID] = reportBaseInfo.RunUserID
		// 组装报告和报告运行时间映射
		reportRunTimeMap[reportBaseInfo.ReportID] = reportBaseInfo.CreatedAt

		// 组装用户id集合
		runUserIds = append(runUserIds, reportBaseInfo.RunUserID)

	}

	// 查出用户信息
	userTable := dal.GetQuery().User
	userList, err := userTable.WithContext(ctx).Where(userTable.UserID.In(runUserIds...)).Find()
	if err != nil {
		log.Logger.Info("对比报告--查询用户信息失败，err:", err)
		return nil, err
	}
	// 用户Id和名称映射数据
	userMap := make(map[string]string)
	for _, userInfo := range userList {
		userMap[userInfo.UserID] = userInfo.Nickname
	}

	// 从mg查询任务对应的配置信息
	var reportTaskConfList []*mao.ReportTask
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectReportTask)
	reportTaskConfListTmp, err := collection.Find(ctx, bson.D{{"team_id", req.TeamID}, {"report_id", bson.D{{"$in", req.ReportIDs}}}})
	if err != nil {
		log.Logger.Info("对比报告--从mongodb查询任务配置信息失败，err:", err)
		return nil, err
	}
	if err := reportTaskConfListTmp.All(ctx, &reportTaskConfList); err != nil {
		log.Logger.Info("对比报告--从mongodb解析任务配置信息失败，err:", err)
		return nil, err
	}

	reportBaseTmp := new(reportBaseFormat)
	for _, reportTaskConfInfo := range reportTaskConfList {
		year := reportRunTimeMap[reportTaskConfInfo.ReportID].Year()
		month := reportRunTimeMap[reportTaskConfInfo.ReportID].Month()
		day := reportRunTimeMap[reportTaskConfInfo.ReportID].Day()
		hour := reportRunTimeMap[reportTaskConfInfo.ReportID].Hour()
		minute := reportRunTimeMap[reportTaskConfInfo.ReportID].Minute()
		second := reportRunTimeMap[reportTaskConfInfo.ReportID].Second()

		reportBaseTmp = &reportBaseFormat{
			ReportID:         reportTaskConfInfo.ReportID,
			Name:             reportNamesMap[reportTaskConfInfo.ReportID],
			Performer:        userMap[reportUserMap[reportTaskConfInfo.ReportID]],
			CreatedTimeSec:   fmt.Sprintf("%d-%d-%d %d:%d:%d", year, month, day, hour, minute, second),
			TaskType:         reportTaskConfInfo.TaskType,
			TaskMode:         reportTaskConfInfo.TaskMode,
			StartConcurrency: reportTaskConfInfo.ModeConf.StartConcurrency,
			Step:             reportTaskConfInfo.ModeConf.Step,
			StepRunTime:      reportTaskConfInfo.ModeConf.StepRunTime,
			MaxConcurrency:   reportTaskConfInfo.ModeConf.MaxConcurrency,
			Duration:         reportTaskConfInfo.ModeConf.Duration,
			Concurrency:      reportTaskConfInfo.ModeConf.Concurrency,
			ReheatTime:       reportTaskConfInfo.ModeConf.ReheatTime,
			RoundNum:         reportTaskConfInfo.ModeConf.RoundNum,
		}
		reportBaseResponse = append(reportBaseResponse, reportBaseTmp)
	}

	// 从mg里面获取报告汇总信息
	var sceneTestResultDataMsgSlice []*reportDataFmt
	collection = dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectReportData)
	reportDataListTmp, err := collection.Find(ctx, bson.D{{"team_id", req.TeamID}, {"report_id", bson.D{{"$in", req.ReportIDs}}}})
	if err != nil {
		log.Logger.Info("对比报告--从mongodb查询报告数据失败，err:", err)
		return nil, err
	}
	if err := reportDataListTmp.All(ctx, &sceneTestResultDataMsgSlice); err != nil {
		log.Logger.Info("对比报告--从mongodb解析报告数据失败，err:", err)
		return nil, err
	}

	// 先组装一个事件id对应报告结果数据
	for _, sceneTestResultData := range sceneTestResultDataMsgSlice {
		var sceneTestResultDataTmp ResultData
		err := json.Unmarshal([]byte(sceneTestResultData.Data), &sceneTestResultDataTmp)
		if err != nil {
			log.Logger.Info("对比报告--解析报告详情数据失败，err:", err)
			return nil, err
		}
		reportCollectDataSlice := make([]reportCollectData, 0, len(req.ReportIDs))
		reportCollectDataMap := make(map[string]*reportDetailData, len(req.ReportIDs))
		for apiId, resultsInfo := range sceneTestResultDataTmp.Results {
			reportCollectDataOne := reportCollectData{
				ApiName:                        resultsInfo.ApiName,
				TotalRequestNum:                resultsInfo.TotalRequestNum,
				TotalRequestTime:               resultsInfo.TotalRequestTime,
				MaxRequestTime:                 resultsInfo.MaxRequestTime,
				MinRequestTime:                 resultsInfo.MinRequestTime,
				AvgRequestTime:                 resultsInfo.AvgRequestTime,
				FiftyRequestTimelineValue:      resultsInfo.FiftyRequestTimelineValue,
				NinetyRequestTimeLineValue:     resultsInfo.NinetyRequestTimeLineValue,
				NinetyFiveRequestTimeLineValue: resultsInfo.NinetyFiveRequestTimeLineValue,
				NinetyNineRequestTimeLineValue: resultsInfo.NinetyNineRequestTimeLineValue,
				Rps:                            resultsInfo.Rps,
				SRps:                           resultsInfo.SRps,
				Tps:                            resultsInfo.Tps,
				STps:                           resultsInfo.STps,
				ErrorRate:                      resultsInfo.ErrorRate,
				ReceivedBytes:                  resultsInfo.ReceivedBytes,
				SendBytes:                      resultsInfo.SendBytes,
			}
			reportCollectDataSlice = append(reportCollectDataSlice, reportCollectDataOne)

			reportCollectDataMap[apiId] = &reportDetailData{
				ApiName:         resultsInfo.ApiName,
				AvgList:         resultsInfo.AvgList,
				RpsList:         resultsInfo.RpsList,
				ConcurrencyList: resultsInfo.ConcurrencyList,
				TpsList:         resultsInfo.TpsList,
				FiftyList:       resultsInfo.FiftyList,
				NinetyList:      resultsInfo.NinetyList,
				NinetyFiveList:  resultsInfo.NinetyFiveList,
				NinetyNineList:  resultsInfo.NinetyNineList,
			}
		}
		reportCollectAllDataResponse = append(reportCollectAllDataResponse, &reportCollectAllData{
			Name: reportNamesMap[sceneTestResultData.ReportID],
			Data: reportCollectDataSlice,
		})

		timeLayout := "2006-01-02 15:04:05"
		createdTime := time.Unix(sceneTestResultDataTmp.TimeStamp, 0).Format(timeLayout)

		reportDetailAllDataTemp := reportDetailAllData{
			Name: reportNamesMap[sceneTestResultData.ReportID],
			Time: createdTime,
			Data: reportCollectDataMap,
		}
		reportDetailAllDataResponse = append(reportDetailAllDataResponse, reportDetailAllDataTemp)
	}

	res := &CompareReportResponse{
		ReportNamesData:     reportNames,
		ReportBaseData:      reportBaseResponse,
		ReportCollectData:   reportCollectAllDataResponse,
		ReportDetailAllData: reportDetailAllDataResponse,
	}

	return res, nil
}

type reportBaseFormat struct {
	ReportID         string `json:"report_id"`
	Name             string `json:"name"`
	Performer        string `json:"performer"`
	CreatedTimeSec   string `json:"created_time_sec"`  // 创建时间
	TaskType         int32  `json:"task_type"`         // 任务类型
	TaskMode         int32  `json:"task_mode"`         // 压测模式
	StartConcurrency int64  `json:"start_concurrency"` // 起始并发数
	Step             int64  `json:"step"`              // 步长
	StepRunTime      int64  `json:"step_run_time"`     // 步长执行时长
	MaxConcurrency   int64  `json:"max_concurrency"`   // 最大并发数
	Duration         int64  `json:"duration"`          // 稳定持续时长，持续时长
	Concurrency      int64  `json:"concurrency"`       // 并发数
	ReheatTime       int64  `json:"reheat_time"`       // 预热时长
	RoundNum         int64  `json:"round_num"`         // 轮次

	ThresholdValue int64 `json:"threshold_value"` // 阈值

}

type reportCollectAllData struct {
	Name string              `json:"name"` // 计划和场景名称
	Data []reportCollectData `json:"data"`
}
type reportCollectData struct {
	ApiName                        string  `json:"api_name" bson:"api_name"`
	TotalRequestNum                uint64  `json:"total_request_num" bson:"total_request_num"`   // 总请求数
	TotalRequestTime               float64 `json:"total_request_time" bson:"total_request_time"` // 总响应时间
	MaxRequestTime                 float64 `json:"max_request_time" bson:"max_request_time"`
	MinRequestTime                 float64 `json:"min_request_time" bson:"min_request_time"` // 毫秒
	AvgRequestTime                 float64 `json:"avg_request_time" bson:"avg_request_time"` // 平均响应事件
	FiftyRequestTimelineValue      float64 `json:"fifty_request_time_line_value" bson:"fifty_request_time_line_value"`
	NinetyRequestTimeLineValue     float64 `json:"ninety_request_time_line_value" bson:"ninety_request_time_line_value"`
	NinetyFiveRequestTimeLineValue float64 `json:"ninety_five_request_time_line_value" bson:"ninety_five_request_time_line_value"`
	NinetyNineRequestTimeLineValue float64 `json:"ninety_nine_request_time_line_value" bson:"ninety_nine_request_time_line_value"`
	Rps                            float64 `json:"rps" bson:"rps"`
	SRps                           float64 `json:"srps" bson:"srps"`
	Tps                            float64 `json:"tps" bson:"tps"`
	STps                           float64 `json:"stps" bson:"stps"`
	ErrorRate                      float64 `json:"error_rate" bson:"error_rate"`
	ReceivedBytes                  float64 `json:"received_bytes" bson:"received_bytes"` // 接收字节数
	SendBytes                      float64 `json:"send_bytes" bson:"send_bytes"`         // 发送字节数
}

type reportDetailAllData struct {
	Name string                       `json:"name" bson:"name"`
	Time string                       `json:"time" bson:"time"`
	Data map[string]*reportDetailData `json:"data" bson:"data"`
}
type reportDetailData struct {
	ApiName         string      `json:"api_name" bson:"api_name"`
	AvgList         []TimeValue `json:"avg_list" bson:"avg_list"`
	RpsList         []TimeValue `json:"rps_list" bson:"rps_list"`
	ConcurrencyList []TimeValue `json:"concurrency_list" bson:"concurrency_list"`
	TpsList         []TimeValue `json:"tps_list" bson:"tps_list"`
	FiftyList       []TimeValue `json:"fifty_list" bson:"fifty_list"`
	NinetyList      []TimeValue `json:"ninety_list" bson:"ninety_list"`
	NinetyFiveList  []TimeValue `json:"ninety_five_list" bson:"ninety_five_list"`
	NinetyNineList  []TimeValue `json:"ninety_nine_list" bson:"ninety_nine_list"`
}

// CompareReportResponse 对比报告接口返回值
type CompareReportResponse struct {
	ReportNamesData     []string                `json:"report_names_data"`
	ReportBaseData      []*reportBaseFormat     `json:"report_base_data"`
	ReportCollectData   []*reportCollectAllData `json:"report_collect_data"`
	ReportDetailAllData []reportDetailAllData   `json:"report_detail_all_data"`
}

// UpdateDescription 保存或更新报告描述
func UpdateDescription(ctx *gin.Context, req rao.UpdateDescriptionReq) error {
	updateMap := make(map[string]string, 1)
	updateMap["description"] = req.Description
	// 修改mg数据
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectReportData)
	_, err := collection.UpdateOne(ctx, bson.D{{"team_id", req.TeamID}, {"report_id", req.ReportID}}, bson.M{"$set": updateMap})
	if err != nil {
		log.Logger.Info("保存或更新报告描述--操作mg失败，err:", err)
		return err
	}
	return nil
}

func BatchDeleteReport(ctx *gin.Context, req *rao.BatchDeleteReportReq) error {
	userID := jwt.GetUserIDByCtx(ctx)
	allErr := dal.GetQuery().Transaction(func(tx *query.Query) error {
		reportList, err := tx.StressPlanReport.WithContext(ctx).Where(tx.StressPlanReport.TeamID.Eq(req.TeamID), tx.StressPlanReport.ReportID.In(req.ReportIDs...)).Find()
		if err != nil {
			return err
		}

		if len(reportList) == 0 {
			return fmt.Errorf("没有找到要删除的报告")
		}

		// 判断是否存在运行中的计划
		for _, reportInfo := range reportList {
			if reportInfo.Status == consts.ReportStatusNormal {
				return fmt.Errorf("存在运行中的报告，无法删除")
			}
		}

		if _, err := tx.StressPlanReport.WithContext(ctx).Where(tx.StressPlanReport.TeamID.Eq(req.TeamID), tx.StressPlanReport.ReportID.In(req.ReportIDs...)).Delete(); err != nil {
			return err
		}

		// 删除报告对应机器信息
		if _, err := tx.ReportMachine.WithContext(ctx).Where(tx.ReportMachine.TeamID.Eq(req.TeamID),
			tx.ReportMachine.ReportID.In(req.ReportIDs...)).Delete(); err != nil {
			return err
		}

		for _, reportInfo := range reportList {
			// 把mongodb库里面的报告详情数据删掉
			collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectReportData)
			findFilter := bson.D{{"report_id", reportInfo.ReportID}, {"team_id", req.TeamID}}
			_, err = collection.DeleteOne(ctx, findFilter)
			if err != nil {
				log.Logger.Info("report_data：删除失败")
			}

			// 把mongodb库里面的报告详情数据删掉
			collection = dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectReportTask)
			findFilter = bson.D{{"report_id", reportInfo.ReportID}, {"team_id", req.TeamID}}
			_, err = collection.DeleteOne(ctx, findFilter)
			if err != nil {
				log.Logger.Info("report_task：删除失败")
			}

			// 把mongodb库里面的报告详情数据删掉
			collection = dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectChangeReportConf)
			findFilter = bson.D{{"report_id", reportInfo.ReportID}, {"team_id", req.TeamID}}
			_, err = collection.DeleteOne(ctx, findFilter)
			if err != nil {
				log.Logger.Info("change_report_conf：删除失败")
			}

			if err := record.InsertDelete(ctx, req.TeamID, userID, record.OperationOperateDeleteReport, fmt.Sprintf("%s %s", reportInfo.PlanName, reportInfo.SceneName)); err != nil {
				return err
			}
		}
		return nil
	})

	if allErr != nil {
		log.Logger.Info("DeleteReport：删除失败")
		return allErr
	}
	return nil
}

func UpdateReportName(ctx *gin.Context, req *rao.UpdateReportNameReq) error {
	allErr := dal.GetQuery().Transaction(func(tx *query.Query) error {
		_, err := tx.StressPlanReport.WithContext(ctx).Where(tx.StressPlanReport.ReportID.Eq(req.ReportID)).UpdateSimple(tx.StressPlanReport.ReportName.Value(req.ReportName))
		if err != nil {
			return err
		}
		return nil
	})
	return allErr
}
