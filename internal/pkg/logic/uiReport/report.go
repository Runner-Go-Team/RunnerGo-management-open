package uiReport

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/log"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/uuid"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/clients"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/mao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/query"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/errmsg"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/packer"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/public"
	"github.com/gin-gonic/gin"
	"github.com/go-omnibus/proof"
	"github.com/jinzhu/copier"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"math"
	"math/rand"
	"strings"
	"time"
)

func ListByTeamID2(ctx *gin.Context, req *rao.ListUIReportsReq) ([]*rao.UIPlanReport, int64, error) {
	if req.Size == 0 {
		req.Size = 20
	}
	if req.Page == 0 {
		req.Page = 1
	}
	teamID := req.TeamID
	limit := req.Size
	offset := (req.Page - 1) * req.Size
	reportName := req.ReportName
	planName := req.PlanName
	startTime := req.StartTime
	endTime := req.EndTime
	taskType := req.TaskType
	status := req.Status
	sortTag := req.Sort
	sceneRunOrder := req.SceneRunOrder
	runUserID := req.RunUserID

	tx := query.Use(dal.DB()).UIPlanReport

	conditions := make([]gen.Condition, 0)
	conditions = append(conditions, tx.TeamID.Eq(teamID))

	if len(reportName) > 0 {
		conditions = append(conditions, tx.ReportName.Like(fmt.Sprintf("%%%s%%", reportName)))
	}

	if len(planName) > 0 {
		conditions = append(conditions, tx.PlanName.Like(fmt.Sprintf("%%%s%%", planName)))
	}

	if len(startTime) > 0 && len(endTime) > 0 {
		layout := "2006-01-02 15:04:05"
		start, _ := time.Parse(layout, startTime)
		t, _ := time.Parse(layout, endTime)

		newTime := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, t.Location())
		end := t.Add(newTime.Sub(t))
		conditions = append(conditions, tx.CreatedAt.Between(start, end))
	}

	if taskType > 0 {
		conditions = append(conditions, tx.TaskType.Eq(taskType))
	}

	if status > 0 {
		conditions = append(conditions, tx.Status.Eq(status))
	}

	if sceneRunOrder > 0 {
		conditions = append(conditions, tx.SceneRunOrder.Eq(sceneRunOrder))
	}

	if len(runUserID) > 0 {
		conditions = append(conditions, tx.RunUserID.Eq(runUserID))
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

	return packer.TransUIReportModelToRaoReportList(reports, users), cnt, nil
}

// GetReportDetail 获取报告详情
func GetReportDetail(ctx *gin.Context, req *rao.UIReportDetailReq) (*rao.UIReportDetail, error) {
	// step1: 查询基本信息
	r := query.Use(dal.DB()).UIPlanReport
	detail, err := r.WithContext(ctx).Where(
		r.ReportID.Eq(req.ReportID),
		r.TeamID.Eq(req.TeamID)).First()
	if err != nil {
		return nil, err
	}

	// step2: 用户信息
	u := query.Use(dal.DB()).User
	user, err := u.WithContext(ctx).Where(u.UserID.Eq(detail.RunUserID)).First()
	if err != nil {
		return nil, err
	}

	// step3: 查询全部的数据
	opts := options.Find().SetSort(bson.D{{"created_at", 1}})
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectUISendSceneOperator)
	cursor, err := collection.Find(ctx, bson.D{{"report_id", req.ReportID}}, opts)
	if err != nil {
		log.Logger.Error("collection.Find err", proof.WithError(err))
		return nil, err
	}
	var sceneOperators []*mao.UISendSceneOperator
	if err = cursor.All(ctx, &sceneOperators); err != nil {
		log.Logger.Error("cursor.All err", proof.WithError(err))
		return nil, err
	}

	uiReportScenes := make([]*rao.UIReportScene, 0)
	uiReportScenesMemo := make(map[string][]*rao.UIReportSceneOperator, 0)
	uiReportSceneIDs := make([]string, 0)
	for _, operator := range sceneOperators {
		uiReportSceneOperator := &rao.UIReportSceneOperator{}
		if err := copier.Copy(uiReportSceneOperator, operator); err != nil {
			log.Logger.Error("UIReportSceneOperator Copy err", proof.WithError(err))
		}

		assertResults := &mao.AssertResults{}
		if err := bson.Unmarshal(operator.AssertResults, &assertResults); err != nil {
			log.Logger.Info("mao.AssertResults bson Unmarshal err", proof.WithError(err))
		}
		uiReportSceneOperator.Assertions = assertResults.Asserts

		withdrawResults := &mao.WithdrawResults{}
		if err := bson.Unmarshal(operator.WithdrawResults, &withdrawResults); err != nil {
			log.Logger.Info("mao.WithdrawResults bson Unmarshal err", proof.WithError(err))
		}
		uiReportSceneOperator.Withdraws = withdrawResults.Withdraws

		multiResults := &mao.MultiResults{}
		if err := bson.Unmarshal(operator.MultiResults, &multiResults); err != nil {
			log.Logger.Info("mao.MultiResults bson Unmarshal err", proof.WithError(err))
		}
		uiReportSceneOperator.MultiResult = multiResults.MultiResults
		if len(multiResults.MultiResults) > 0 {
			for _, multi := range multiResults.MultiResults {
				uiReportSceneOperator.ExecTime += multi.ExecTime
				multi.ExecTime = math.Round(multi.ExecTime*100) / 100
			}
		}
		uiReportSceneOperator.ExecTime = math.Round(uiReportSceneOperator.ExecTime*100) / 100

		uiReportScenesMemo[operator.SceneID] = append(uiReportScenesMemo[operator.SceneID], uiReportSceneOperator)
		uiReportSceneIDs = append(uiReportSceneIDs, operator.SceneID)
	}

	uiReportSceneIDs = public.SliceUnique(uiReportSceneIDs)
	for _, sceneID := range uiReportSceneIDs {
		operators := uiReportScenesMemo[sceneID]

		var operatorTotalNum int
		var operatorSuccessNum int
		var operatorErrorNum int
		var operatorUnExecNum int
		var assertTotalNum int
		var assertSuccessNum int
		var assertErrorNum int
		var assertUnExecNum int
		var sceneName string
		var runStatus int
		for _, operator := range operators {
			// 区分批量和单个
			if len(operator.MultiResult) == 0 {
				operatorTotalNum++
				assertTotalNum += operator.AssertTotalNum
				if operator.RunStatus == consts.ReportRunStatusNo {
					operatorUnExecNum++
					assertUnExecNum += operator.AssertTotalNum
				}
				if operator.RunStatus == consts.ReportRunStatusSuccess {
					operatorSuccessNum++
					assertSuccessNum += operator.AssertTotalNum
				}
				if operator.RunStatus == consts.ReportRunStatusError {
					operatorErrorNum++
					assertErrorNum += operator.AssertTotalNum
				}
			}
			if len(operator.MultiResult) > 0 {
				for _, result := range operator.MultiResult {
					operatorTotalNum++
					assertTotalNum += operator.AssertTotalNum
					if result.RunStatus == consts.ReportRunStatusNo || result.RunStatus == consts.ReportRunStatusZero {
						operatorUnExecNum++
						assertUnExecNum += operator.AssertTotalNum
					}
					if result.RunStatus == consts.ReportRunStatusSuccess {
						operatorSuccessNum++
						assertSuccessNum += operator.AssertTotalNum
					}
					if result.RunStatus == consts.ReportRunStatusError {
						operatorErrorNum++
						assertErrorNum += operator.AssertTotalNum
					}
				}
			}
			sceneName = operator.SceneName
		}

		// 场景状态
		if operatorUnExecNum == len(operators) {
			runStatus = consts.ReportRunStatusNo
		} else if operatorSuccessNum == len(operators) {
			runStatus = consts.ReportRunStatusSuccess
		} else {
			runStatus = consts.ReportRunStatusError
		}
		uiReportScene := &rao.UIReportScene{
			SceneID:            sceneID,
			Name:               sceneName,
			RunStatus:          runStatus,
			OperatorTotalNum:   operatorTotalNum,
			OperatorSuccessNum: operatorSuccessNum,
			OperatorErrorNum:   operatorErrorNum,
			OperatorUnExecNum:  operatorUnExecNum,
			AssertTotalNum:     assertTotalNum,
			AssertSuccessNum:   assertSuccessNum,
			AssertErrorNum:     assertErrorNum,
			AssertUnExecNum:    assertUnExecNum,
			Operators:          operators,
		}
		uiReportScenes = append(uiReportScenes, uiReportScene)
	}

	var allSceneErrorNum int
	var allSceneUnExecNum int
	var allOperatorTotalNum int
	var allOperatorSuccessNum int
	var allOperatorErrorNum int
	var allOperatorUnExecNum int
	var allAssertTotalNum int
	var allAssertSuccessNum int
	var allAssertErrorNum int
	var allAssertUnExecNum int
	for _, scene := range uiReportScenes {
		if scene.OperatorUnExecNum == len(scene.Operators) {
			allSceneUnExecNum++
		}
		if scene.OperatorErrorNum > 0 {
			allSceneErrorNum++
		}
		allOperatorTotalNum += scene.OperatorTotalNum
		allOperatorSuccessNum += scene.OperatorSuccessNum
		allOperatorErrorNum += scene.OperatorErrorNum
		allOperatorUnExecNum += scene.OperatorUnExecNum

		allAssertTotalNum += scene.AssertTotalNum
		allAssertSuccessNum += scene.AssertSuccessNum
		allAssertErrorNum += scene.AssertErrorNum
		allAssertUnExecNum += scene.AssertUnExecNum
	}

	runDurationTime := detail.RunDurationTime
	if runDurationTime == 0 && detail.Status == consts.UIReportStatusIng {
		durationTime := time.Now().Unix() - detail.CreatedAt.Unix()
		runDurationTime = durationTime
	}

	return &rao.UIReportDetail{
		RunUserID:          user.UserID,
		RunUserName:        user.Nickname,
		UserAvatar:         user.Avatar,
		PlanID:             detail.PlanID,
		PlanName:           detail.PlanName,
		ReportID:           detail.ReportID,
		ReportName:         detail.ReportName,
		CreatedTimeSec:     detail.CreatedAt.Unix(),
		EndTimeSec:         detail.CreatedAt.Unix() + detail.RunDurationTime,
		TaskType:           detail.TaskType,
		SceneRunOrder:      detail.SceneRunOrder,
		Status:             detail.Status,
		RunDurationTime:    runDurationTime,
		RunTimeSec:         detail.CreatedAt.Unix(),
		LastTimeSec:        detail.UpdatedAt.Unix(),
		SceneTotalNum:      len(uiReportScenes),
		SceneSuccessNum:    len(uiReportScenes) - allSceneErrorNum - allSceneUnExecNum,
		SceneErrorNum:      allSceneErrorNum,
		SceneUnExecNum:     allSceneUnExecNum,
		OperatorTotalNum:   allOperatorTotalNum,
		OperatorSuccessNum: allOperatorSuccessNum,
		OperatorErrorNum:   allOperatorErrorNum,
		OperatorUnExecNum:  allOperatorUnExecNum,
		AssertTotalNum:     allAssertTotalNum,
		AssertSuccessNum:   allAssertSuccessNum,
		AssertErrorNum:     allAssertErrorNum,
		AssertUnExecNum:    allAssertUnExecNum,
		Scenes:             uiReportScenes,
	}, nil
}

func Delete(ctx *gin.Context, userID string, req *rao.UIReportDeleteReq) error {
	err := dal.GetQuery().Transaction(func(tx *query.Query) error {
		for _, reportID := range req.ReportIDs {
			// 删除计划基本信息
			if _, err := tx.UIPlanReport.WithContext(ctx).Where(
				tx.UIPlanReport.TeamID.Eq(req.TeamID),
				tx.UIPlanReport.ReportID.Eq(reportID)).Delete(); err != nil {
				return err
			}
		}
		return nil
	})

	return err
}

// Stop 停止
func Stop(ctx *gin.Context, userID string, req *rao.StopUIReportReq) error {
	// 发送停止计划状态变更信息
	statusChangeKey := consts.UIReportStatusChange + req.ReportID
	statusChangeValue := rao.UIReportStatusChange{
		Status: "stop",
	}
	statusChangeValueString, err := json.Marshal(statusChangeValue)
	if err == nil {
		// 发送计划相关信息到redis频道
		_, err = dal.GetRDB().Publish(ctx, statusChangeKey, string(statusChangeValueString)).Result()
		if err != nil {
			log.Logger.Error("停止--发送压测计划状态变更到对应频道失败")
			return err
		}
	} else {
		log.Logger.Error("停止--发送压测计划状态变更到对应频道，压缩数据失败")
		return err
	}

	// 是否存在报告，存在报告数据
	pr := query.Use(dal.DB()).UIPlanReport
	planReport, err := pr.WithContext(ctx).Where(
		pr.ReportID.Eq(req.ReportID),
		pr.TeamID.Eq(req.TeamID),
	).First()
	var upDataPlanReport = make(map[string]interface{}, 0)
	durationTime := time.Now().Unix() - planReport.CreatedAt.Unix()
	upDataPlanReport["run_duration_time"] = durationTime
	upDataPlanReport["status"] = consts.UIReportStatusEnd

	if err = query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		if len(upDataPlanReport) > 0 {
			if _, err = tx.UIPlanReport.WithContext(ctx).Where(
				tx.UIPlanReport.ReportID.Eq(req.ReportID),
				tx.UIPlanReport.TeamID.Eq(req.TeamID)).Updates(upDataPlanReport); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return err
	}

	// 删除 Redis 对应关系
	redisKey := consts.UIEngineRunAddrPrefix + req.ReportID
	addr, err := dal.GetRDB().Get(ctx, redisKey).Result()
	if err != nil {
		log.Logger.Info("StopScene--运行ID对应的机器失败：", err)
	}
	dal.GetRDB().SRem(ctx, consts.UIEngineCurrentRunPrefix+addr, req.ReportID)
	dal.GetRDB().Del(ctx, consts.UIEngineRunAddrPrefix+req.ReportID)

	return nil
}

// Update 修改  目前只能修改名称
func Update(ctx *gin.Context, userID string, req *rao.UIReportUpdateReq) error {
	err := dal.GetQuery().Transaction(func(tx *query.Query) error {
		var upDataPlanReport = make(map[string]interface{}, 0)
		upDataPlanReport["report_name"] = req.ReportName
		// 删除计划基本信息
		if _, err := tx.UIPlanReport.WithContext(ctx).Where(
			tx.UIPlanReport.TeamID.Eq(req.TeamID),
			tx.UIPlanReport.ReportID.Eq(req.ReportID)).Updates(upDataPlanReport); err != nil {
			return err
		}
		return nil
	})

	return err
}

// StopPlan 停止计划
func StopPlan(ctx *gin.Context, userID string, req *rao.StopUIPlanReq) error {
	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		ap := tx.UIPlan
		planInfo, err := ap.WithContext(ctx).Where(ap.TeamID.Eq(req.TeamID), ap.PlanID.Eq(req.PlanID)).First()
		if err != nil {
			return err
		}

		// 判断任务时定时任务还是普通任务
		if planInfo.TaskType == consts.UIPlanTaskTypeCronjob { // 定时任务
			apttc := dal.GetQuery().UIPlanTimedTaskConf
			_, err = apttc.WithContext(ctx).Where(apttc.TeamID.Eq(req.TeamID), apttc.PlanID.Eq(req.PlanID)).UpdateSimple(apttc.Status.Value(consts.UIPlanTimedTaskWaitEnable))
			if err != nil {
				return err
			}
		}

		apr := tx.UIPlanReport
		aprList, err := apr.WithContext(ctx).Where(apr.TeamID.Eq(req.TeamID), apr.PlanID.Eq(req.PlanID), apr.Status.Eq(consts.UIReportStatusIng)).Find()
		if err != nil {
			return err
		}

		_, err = apr.WithContext(ctx).Where(apr.TeamID.Eq(req.TeamID), apr.PlanID.Eq(req.PlanID)).UpdateSimple(apr.Status.Value(consts.UIReportStatusEnd))
		if err != nil {
			return err
		}

		if len(aprList) > 0 {
			for _, aprInfo := range aprList {
				if err = Stop(ctx, userID, &rao.StopUIReportReq{
					TeamID:   req.TeamID,
					ReportID: aprInfo.ReportID,
				}); err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func Run(ctx *gin.Context, userID string, req *rao.RunUIReportReq) (string, error) {
	// step1: 查询计划下的场景
	// step2: 组装数据
	// step3: 生成报告
	r := query.Use(dal.DB()).UIPlanReport
	report, err := r.WithContext(ctx).Where(
		r.ReportID.Eq(req.ReportID),
		r.TeamID.Eq(req.TeamID),
	).First()
	if err != nil {
		return "", err
	}

	// step1: 获取发送机器
	uiEngineList, err := clients.GetUiEngineMachineList()
	if err != nil {
		return "", errors.New("get ui_engine empty" + err.Error())
	}

	uiEngine := &rao.UiEngineMachineInfo{}
	for _, info := range uiEngineList {
		if info.Key == report.UIMachineKey {
			uiEngine = info
		}
	}
	if len(uiEngine.IP) == 0 {
		uiEngine = uiEngineList[rand.Intn(len(uiEngineList))]
	}

	browsers := make([]*rao.Browser, 0)
	_ = json.Unmarshal([]byte(report.Browsers), &browsers)
	if strings.Contains(strings.ToLower(uiEngine.SystemInfo.SystemBasic), "linux") {
		for _, b := range browsers {
			if !b.Headless {
				return "", errmsg.ErrSendLinuxNotQTMode
			}
		}
	}

	runID := uuid.GetUUID()

	// step3: 查询全部的数据
	operatorCollection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectUISendSceneOperator)
	cursor, err := operatorCollection.Find(ctx, bson.D{{"report_id", req.ReportID}})
	if err != nil {
		log.Logger.Error("collection.Find err", proof.WithError(err))
		return "", err
	}
	var sceneOperators []*mao.UISendSceneOperator
	if err = cursor.All(ctx, &sceneOperators); err != nil {
		log.Logger.Error("cursor.All err", proof.WithError(err))
		return "", err
	}

	assertResults, err := bson.Marshal(mao.AssertResults{Asserts: nil})
	if err != nil {
		log.Logger.Info("TransSceneOperatorsToRaoSendOperators.uiOperatorDetail bson marshal err", proof.WithError(err))
	}

	withdrawResults, err := bson.Marshal(mao.WithdrawResults{Withdraws: nil})
	if err != nil {
		log.Logger.Info("TransSceneOperatorsToRaoSendOperators.uiOperatorDetail bson marshal err", proof.WithError(err))
	}

	multiResults, err := bson.Marshal(mao.MultiResults{MultiResults: nil})
	if err != nil {
		log.Logger.Info("TransSceneOperatorsToRaoSendOperators.uiOperatorDetail bson marshal err", proof.WithError(err))
	}
	for _, operator := range sceneOperators {
		operator.ReportId = runID
		operator.RunStatus = 1
		operator.ExecTime = 0
		operator.RunEndTimes = 0
		operator.Status = ""
		operator.Msg = ""
		operator.Screenshot = ""
		operator.End = false
		operator.IsMulti = false
		operator.AssertResults = assertResults
		operator.MultiResults = multiResults
		operator.WithdrawResults = withdrawResults
	}

	// step3: 生成操作记录
	var docs []interface{}
	for _, run := range sceneOperators {
		docs = append(docs, run)
	}
	if _, err := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectUISendSceneOperator).InsertMany(ctx, docs); err != nil {
		log.Logger.Error("调试日志入库失败", err)
		return "", err
	}

	// 查询接口详情数据
	sendReport := &mao.SendReport{}
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectUISendReport)
	err = collection.FindOne(ctx, bson.D{{"report_id", req.ReportID}}).Decode(&sendReport)
	if err != nil {
		log.Logger.Error("查询之前的报告信息失败", err)
	}
	var sendReportDetail *mao.SendReportDetail
	if err = bson.Unmarshal(sendReport.Detail, &sendReportDetail); err != nil {
		log.Logger.Errorf("mao.SendReportDetail bson unmarshal err %w", err)
	}
	runRequest := sendReportDetail.Detail
	runRequest.Topic = runID
	runRequest.UserId = userID
	runRequest.SceneRunOrder = report.SceneRunOrder
	if _, err = clients.RunUiEngine(ctx, uiEngine.IP, runRequest); err != nil {
		return "", err
	}

	// step5: 添加发送记录
	detail, err := bson.Marshal(mao.SendReportDetail{Detail: runRequest})
	if err != nil {
		log.Logger.Error("mao.SendReportDetail bson marshal err", proof.WithError(err))
	}
	sendReport = &mao.SendReport{
		ReportID: runID,
		Detail:   detail,
	}
	if _, err = collection.InsertOne(ctx, sendReport); err != nil {
		return "", err
	}

	// step6: 添加自动化记录
	dal.GetRDB().SAdd(ctx, consts.UIEngineCurrentRunPrefix+uiEngine.IP, runID)
	dal.GetRDB().Set(ctx, consts.UIEngineRunAddrPrefix+runID, uiEngine.IP, time.Second*3600)

	log.Logger.Info("运行计划--创建报告", req.TeamID, req.ReportID)
	reportData := model.UIPlanReport{
		ReportID:        runID,
		ReportName:      report.ReportName,
		PlanID:          report.PlanID,
		PlanName:        report.PlanName,
		TeamID:          report.TeamID,
		TaskType:        report.TaskType,
		SceneRunOrder:   report.SceneRunOrder,
		RunDurationTime: 0,
		Status:          consts.ReportStatusNormal,
		RunUserID:       userID,
		Remark:          "",
		Browsers:        report.Browsers,
		UIMachineKey:    report.UIMachineKey,
	}
	pr := query.Use(dal.DB()).UIPlanReport
	if err = pr.WithContext(ctx).Create(&reportData); err != nil {
		return "", err
	}

	return runID, nil
}
