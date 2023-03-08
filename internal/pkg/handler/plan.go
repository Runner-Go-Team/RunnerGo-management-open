package handler

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"kp-management/internal/pkg/biz/consts"
	"kp-management/internal/pkg/biz/errno"
	"kp-management/internal/pkg/biz/jwt"
	"kp-management/internal/pkg/biz/log"
	"kp-management/internal/pkg/biz/mail"
	"kp-management/internal/pkg/biz/record"
	"kp-management/internal/pkg/biz/response"
	"kp-management/internal/pkg/dal"
	"kp-management/internal/pkg/dal/model"
	"kp-management/internal/pkg/dal/rao"
	"kp-management/internal/pkg/logic/plan"
	"kp-management/internal/pkg/logic/stress"
	"strings"
	"sync"
	"time"

	"kp-management/internal/pkg/dal/query"
)

type RunStressReq struct {
	PlanID  string   `json:"plan_id"`
	TeamID  string   `json:"team_id"`
	SceneID []string `json:"scene_id"`
	UserID  string   `json:"user_id"`
	RunType int      `json:"source"` // 0，1-普通，2-定时
}

// RunPlan 启动计划
func RunPlan(ctx *gin.Context) {
	var req rao.RunPlanReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	// 调用controller方法改成本地
	runStressParams := RunStressReq{
		PlanID:  req.PlanID,
		TeamID:  req.TeamID,
		UserID:  jwt.GetUserIDByCtx(ctx),
		RunType: 1,
	}

	errnoNum, runErr := RunStress(ctx, runStressParams)
	if runErr != nil {
		response.ErrorWithMsg(ctx, errnoNum, runErr.Error())
		return
	}

	px := dal.GetQuery().StressPlan
	p, err := px.WithContext(ctx).Where(px.TeamID.Eq(req.TeamID), px.PlanID.Eq(req.PlanID)).First()
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	// 插入操作日志
	if p.TaskType == consts.PlanTaskTypeNormal || runStressParams.RunType == 2 {
		if err := record.InsertRun(ctx, req.TeamID, jwt.GetUserIDByCtx(ctx), record.OperationOperateRunPlan, p.PlanName); err != nil {
			response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
			return
		}
	} else {
		if err := record.InsertExecute(ctx, req.TeamID, jwt.GetUserIDByCtx(ctx), record.OperationOperateExecPlan, p.PlanName); err != nil {
			response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
			return
		}
	}

	tx := dal.GetQuery().StressPlanEmail
	emails, err := tx.WithContext(ctx).Where(tx.TeamID.Eq(req.TeamID), tx.PlanID.Eq(req.PlanID)).Find()
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrHttpFailed, err.Error())
		return
	}
	if len(emails) > 0 {
		px := dal.GetQuery().StressPlan
		planInfo, err := px.WithContext(ctx).Where(px.TeamID.Eq(req.TeamID), px.PlanID.Eq(req.PlanID)).First()
		if err != nil {
			response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
			return
		}

		ttx := dal.GetQuery().Team
		team, err := ttx.WithContext(ctx).Where(ttx.TeamID.Eq(req.TeamID)).First()
		if err != nil {
			response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
			return
		}

		rx := dal.GetQuery().StressPlanReport
		reports, err := rx.WithContext(ctx).Where(rx.TeamID.Eq(req.TeamID), rx.PlanID.Eq(req.PlanID), rx.CreatedAt.Gt(emails[0].CreatedAt)).Find()
		if err != nil {
			response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
			return
		}

		ux := dal.GetQuery().User
		user, err := ux.WithContext(ctx).Where(ux.UserID.Eq(jwt.GetUserIDByCtx(ctx))).First()
		if err != nil {
			response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
			return
		}

		var userIDs []string
		for _, report := range reports {
			userIDs = append(userIDs, report.RunUserID)
		}
		runUsers, err := ux.WithContext(ctx).Where(ux.UserID.In(userIDs...)).Find()
		if err != nil {
			response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
			return
		}

		for _, email := range emails {
			if err := mail.SendPlanEmail(email.Email, planInfo.PlanName, team.Name, user.Nickname, reports, runUsers); err != nil {
				if err.Error() == "请配置邮件相关环境变量" {
					response.ErrorWithMsg(ctx, errno.ErrNotEmailConfig, err.Error())
				} else {
					response.ErrorWithMsg(ctx, errno.ErrHttpFailed, err.Error())
				}
				return
			}
		}
	}

	response.Success(ctx)
	return
}

// StopPlan 停止计划
func StopPlan(ctx *gin.Context) {
	var req rao.StopPlanReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	tx := dal.GetQuery().StressPlanReport
	reports, err := tx.WithContext(ctx).Where(tx.TeamID.Eq(req.TeamID), tx.PlanID.In(req.PlanIDs...)).Find()
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	var reportIDs []string
	for _, report := range reports {
		reportIDs = append(reportIDs, report.ReportID)
	}

	// 停止计划的时候，往redis里面写一条数据
	reportIDsString := reportIDs
	for _, reportID := range reportIDsString {
		// 发送停止计划状态变更信息
		statusChangeKey := consts.SubscriptionStressPlanStatusChange + reportID
		statusChangeValue := rao.SubscriptionStressPlanStatusChange{
			Type:     1,
			StopPlan: "stop",
		}
		statusChangeValueString, err := json.Marshal(statusChangeValue)
		if err == nil {
			// 发送计划相关信息到redis频道
			_, err = dal.GetRDB().Publish(ctx, statusChangeKey, string(statusChangeValueString)).Result()
			if err != nil {
				log.Logger.Info("停止计划--发送压测计划状态变更到对应频道失败")
				continue
			}
		} else {
			log.Logger.Info("停止计划--发送压测计划状态变更到对应频道，压缩数据失败")
			continue
		}
		log.Logger.Info("停止计划--发送性能计划停止消息成功")
	}

	px := dal.GetQuery().StressPlan
	_, err = px.WithContext(ctx).Where(px.TeamID.Eq(req.TeamID), px.PlanID.In(req.PlanIDs...)).UpdateColumn(px.Status, consts.PlanStatusNormal)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}

// ClonePlan 克隆计划
func ClonePlan(ctx *gin.Context) {
	var req rao.ClonePlanReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	if err := plan.ClonePlan(ctx, &req, jwt.GetUserIDByCtx(ctx)); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}

// ListUnderwayPlan 运行中的计划
func ListUnderwayPlan(ctx *gin.Context) {
	var req rao.ListUnderwayPlanReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	runPlanNum, err := plan.ListByStatus(ctx, req.TeamID)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.ListUnderwayPlanResp{
		RunPlanNum: runPlanNum,
	})
	return
}

// ListPlans 测试计划列表
func ListPlans(ctx *gin.Context) {
	var req rao.ListPlansReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	isExist := strings.Index(req.Keyword, "%")
	if isExist >= 0 {
		response.SuccessWithData(ctx, rao.ListPlansResp{})
		return
	}

	plans, total, err := plan.ListByTeamID(ctx, req.TeamID, req.Size, (req.Page-1)*req.Size,
		req.Keyword, req.StartTimeSec, req.EndTimeSec, req.TaskType, req.TaskMode, req.Status, req.Sort)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.ListPlansResp{
		Plans: plans,
		Total: total,
	})
	return
}

// SavePlan 创建修改计划
func SavePlan(ctx *gin.Context) {
	var req rao.SavePlanReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	planID, errNum, err := plan.Save(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errNum, err.Error())
		return
	}
	response.SuccessWithData(ctx, rao.SavePlanResp{PlanID: planID})
	return
}

// SavePlanTask 创建/修改计划配置
func SavePlanTask(ctx *gin.Context) {
	var req rao.SavePlanConfReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	// 必填项判断
	if (req.ModeConf.Duration == 0 && req.ModeConf.RoundNum == 0) || (req.Mode == 1 && req.ModeConf.Concurrency == 0) {
		response.ErrorWithMsg(ctx, errno.ErrMustTaskInit, "必填项不能为空")
		return
	}

	if req.Mode != 1 { // 非并发模式参数校验
		if req.ModeConf.StartConcurrency == 0 || req.ModeConf.Step == 0 || req.ModeConf.StepRunTime == 0 || req.ModeConf.MaxConcurrency == 0 || req.ModeConf.Duration == 0 {
			response.ErrorWithMsg(ctx, errno.ErrMustTaskInit, "必填项不能为空")
			return
		}
		if req.ModeConf.MaxConcurrency < req.ModeConf.StartConcurrency {
			response.ErrorWithMsg(ctx, errno.ErrMaxConcurrencyLessThanStartConcurrency, "")
			return
		}
	}

	if req.TaskType == 2 {
		if req.TimedTaskConf.Frequency == 0 {
			if req.TimedTaskConf.TaskExecTime == 0 {
				response.ErrorWithMsg(ctx, errno.ErrMustTaskInit, "必填项不能为空")
				return
			}
			req.TimedTaskConf.TaskCloseTime = req.TimedTaskConf.TaskExecTime + 120
		} else {
			if req.TimedTaskConf.TaskExecTime == 0 || req.TimedTaskConf.TaskCloseTime == 0 {
				response.ErrorWithMsg(ctx, errno.ErrMustTaskInit, "必填项不能为空")
				return
			}
		}
	}

	errNum, err := plan.SaveTask(ctx, &req, jwt.GetUserIDByCtx(ctx))
	if err != nil {
		response.ErrorWithMsg(ctx, errNum, err.Error())
		return
	}

	response.Success(ctx)
	return
}

func GetPlanTask(ctx *gin.Context) {
	var req rao.GetPlanTaskReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	pt, err := plan.GetPlanTask(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.GetPlanTaskResp{PlanTask: pt})
	return
}

// GetPlan 获取计划
func GetPlan(ctx *gin.Context) {
	var req rao.GetPlanConfReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	p, err := plan.GetByPlanID(ctx, req.TeamID, req.PlanID)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.GetPlanResp{Plan: p})
	return
}

// DeletePlan 删除计划
func DeletePlan(ctx *gin.Context) {
	var req rao.DeletePlanReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	err := plan.DeleteByPlanID(ctx, req.TeamID, req.PlanID, jwt.GetUserIDByCtx(ctx))
	if err != nil {
		if err.Error() == "不能删除正在运行的计划" {
			response.ErrorWithMsg(ctx, errno.ErrCannotDeleteRunningPlan, err.Error())
		} else {
			response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		}
		return
	}

	response.Success(ctx)
	return
}

// ImportScene 导入场景
func ImportScene(ctx *gin.Context) {
	var req rao.ImportSceneReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	if len(req.TargetIDList) == 0 {
		response.ErrorWithMsg(ctx, errno.ErrParam, "导入场景不能为空")
		return
	}

	scenes, err := plan.ImportScene(ctx, jwt.GetUserIDByCtx(ctx), &req)
	if err != nil {
		if err.Error() == "计划内分组不可重名" {
			response.ErrorWithMsg(ctx, errno.ErrInPlanGroupNameAlreadyExist, err.Error())
		} else if err.Error() == "计划内场景不可重名" {
			response.ErrorWithMsg(ctx, errno.ErrInPlanSceneNameAlreadyExist, err.Error())
		} else {
			response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		}
		return
	}

	response.SuccessWithData(ctx, rao.ImportSceneResp{
		Scenes: scenes,
	})
	return
}

// PlanAddEmail 添加计划收件人
func PlanAddEmail(ctx *gin.Context) {
	var req rao.PlanEmailReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	var planEmails []*model.StressPlanEmail
	for _, email := range req.Emails {
		planEmails = append(planEmails, &model.StressPlanEmail{
			PlanID: req.PlanID,
			TeamID: req.TeamID,
			Email:  email,
		})
	}

	tx := dal.GetQuery().StressPlanEmail
	cnt, err := tx.WithContext(ctx).Where(tx.TeamID.Eq(req.TeamID), tx.PlanID.Eq(req.PlanID), tx.Email.In(req.Emails...)).Count()
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	if cnt > 0 {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, "email exists")
		return
	}

	if err := tx.WithContext(ctx).CreateInBatches(planEmails, 5); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}
	response.Success(ctx)
	return
}

// PlanListEmail 计划收件人列表
func PlanListEmail(ctx *gin.Context) {
	var req rao.PlanListEmailReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	tx := dal.GetQuery().StressPlanEmail
	emails, err := tx.WithContext(ctx).Where(tx.TeamID.Eq(req.TeamID), tx.PlanID.Eq(req.PlanID)).Order(tx.CreatedAt).Find()
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	ret := make([]*rao.PlanEmail, 0)
	for _, e := range emails {
		ret = append(ret, &rao.PlanEmail{
			ID:            e.ID,
			PlanID:        e.PlanID,
			TeamID:        e.TeamID,
			Email:         e.Email,
			CreateTimeSec: e.CreatedAt.Unix(),
		})
	}
	response.SuccessWithData(ctx, rao.PlanListEmailResp{Emails: ret})
	return
}

// PlanDeleteEmail 删除计划收件人
func PlanDeleteEmail(ctx *gin.Context) {
	var req rao.PlanDeleteEmailReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	tx := dal.GetQuery().StressPlanEmail
	_, err := tx.WithContext(ctx).Where(tx.ID.Eq(req.EmailID)).Delete()
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}

// RunStress 调度压力测试机进行压测的方法
func RunStress(ctx context.Context, req RunStressReq) (int, error) {
	rms := &stress.RunMachineStress{}

	//siv := &stress.SplitImportVariable{}
	//siv.SetNext(rms)

	ss := &stress.SplitStress{}
	ss.SetNext(rms)

	ms := &stress.MakeStress{}
	ms.SetNext(ss)

	mr := &stress.MakeReport{}
	mr.SetNext(ms)

	iv := &stress.AssembleImportVariables{}
	iv.SetNext(mr)

	sv := &stress.AssembleSceneVariables{}
	sv.SetNext(iv)

	f := &stress.AssembleFlows{}
	f.SetNext(sv)

	v := &stress.AssembleGlobalVariables{}
	v.SetNext(f)

	t := &stress.AssembleTask{}
	t.SetNext(v)

	s := &stress.AssembleScenes{}
	s.SetNext(t)

	p := &stress.AssemblePlan{}
	p.SetNext(s)

	m := &stress.CheckIdleMachine{}
	m.SetNext(p)

	ctt := &stress.CheckStressPlanTaskType{}
	ctt.SetNext(m)

	errnoNum, err := ctt.Execute(&stress.Baton{
		Ctx:      ctx,
		PlanID:   req.PlanID,
		TeamID:   req.TeamID,
		UserID:   req.UserID,
		SceneIDs: req.SceneID,
		RunType:  req.RunType,
	})

	return errnoNum, err
}

// NotifyStopStress 压力机回调压测状态和结果
func NotifyStopStress(ctx *gin.Context) {
	var req rao.NotifyStopStressReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	log.Logger.Info("NotifyStopStress--性能测试回调接口参数：", req)

	reportInfo := new(model.StressPlanReport)

	allErr := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		// 查询当前计划信息
		planTable := tx.StressPlan
		planInfo, err := planTable.WithContext(ctx).Where(planTable.TeamID.Eq(req.TeamID),
			planTable.PlanID.Eq(req.PlanID)).First()
		if err != nil {
			return err
		}

		r := tx.StressPlanReport
		// 修改报告状态
		_, err = r.WithContext(ctx).Where(r.TeamID.Eq(req.TeamID), r.PlanID.Eq(req.PlanID), r.ReportID.Eq(req.ReportID)).UpdateSimple(r.Status.Value(consts.ReportStatusFinish))
		if err != nil {
			log.Logger.Info("NotifyStopStress--修改报告状态失败")
			return err
		}

		// 查找报告对应计划
		report, err := r.WithContext(ctx).Where(r.TeamID.Eq(req.TeamID), r.PlanID.Eq(req.PlanID), r.ReportID.Eq(req.ReportID)).First()
		if err != nil {
			log.Logger.Info("NotifyStopStress--查找报告对应计划失败")
			return err
		}
		reportInfo = report

		if planInfo.TaskType == consts.PlanTaskTypeNormal {
			// 统计报告是否全部完成
			reportCnt, err := r.WithContext(ctx).Where(r.TeamID.Eq(req.TeamID), r.PlanID.Eq(req.PlanID)).Count()
			if err != nil {
				log.Logger.Info("NotifyStopStress--统计当前计划下所有的报告数量--失败")
				return err
			}
			finishReportCnt, err := r.WithContext(ctx).Where(r.TeamID.Eq(req.TeamID), r.PlanID.Eq(report.PlanID), r.Status.Eq(consts.ReportStatusFinish)).Count()
			if err != nil {
				log.Logger.Info("NotifyStopStress--统计当前计划下所有成功的报告--失败")
				return err
			}

			if finishReportCnt == reportCnt { // 报告全部完成则计划也完成
				_, err := planTable.WithContext(ctx).Where(planTable.TeamID.Eq(req.TeamID), planTable.PlanID.Eq(report.PlanID)).UpdateSimple(planTable.Status.Value(consts.PlanStatusNormal))
				if err != nil {
					log.Logger.Info("NotifyStopStress计划下所有报告并未全部完成")
					return err
				}
			}
		} else { // 定时任务
			// 判断定时任务频次
			ttc := dal.GetQuery().StressPlanTimedTaskConf
			TimedTaskConfInfo, err := ttc.WithContext(ctx).Where(ttc.TeamID.Eq(req.TeamID),
				ttc.PlanID.Eq(req.PlanID), ttc.SenceID.Neq(reportInfo.SceneID)).First()
			nowTime := time.Now().Unix()
			if err == nil {
				if TimedTaskConfInfo.Frequency == 0 || (TimedTaskConfInfo.Frequency != 0 && TimedTaskConfInfo.TaskCloseTime <= nowTime) {
					// 查到定时任务配置了,如果任务配置过期时间小于当前时间，则把计划状态改为未运行
					p := dal.GetQuery().StressPlan
					_, err := p.WithContext(ctx).Where(p.TeamID.Eq(reportInfo.TeamID),
						p.PlanID.Eq(reportInfo.PlanID)).UpdateSimple(p.Status.Value(consts.PlanStatusNormal))
					if err != nil {
						log.Logger.Info("NotifyStopStress--修改定时计划状态失败")
						response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
						return err
					}
				}
			}
		}

		return nil
	})

	for _, machine := range req.Machines {
		mInfo := strings.Split(machine, "_")
		if len(mInfo) != 2 {
			continue
		}
		machineUseStateKey := consts.MachineUseStatePrefix + mInfo[0] + ":" + mInfo[1]
		dal.GetRDB().Del(ctx, machineUseStateKey)
	}

	if allErr != nil {
		log.Logger.Info("NotifyStopStress整体事务失败")
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, allErr.Error())
		return
	}
	log.Logger.Info("NotifyStopStress--性能测试回调接口结果：", allErr)

	log.Logger.Info("NotifyStopStress--把报告数据从redis写入到MongoDB库，参数为：", req)

	wg := sync.WaitGroup{}
	go func(wg *sync.WaitGroup) {
		wg.Add(1)
		time.Sleep(2 * time.Second)
		_ = plan.InsertReportData(ctx, &req)
		wg.Done()
	}(&wg)
	wg.Wait()

	response.Success(ctx)
	return
}

// BatchDeletePlan 批量删除性能测试计划
func BatchDeletePlan(ctx *gin.Context) {
	var req rao.BatchDeletePlanReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	err := plan.BatchDeletePlan(ctx, &req)
	if err != nil {
		if err.Error() == "存在运行中的计划，无法删除" {
			response.ErrorWithMsg(ctx, errno.ErrCannotBatchDeleteRunningPlan, err.Error())
		} else {
			response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		}

		return
	}
	response.Success(ctx)
	return
}
