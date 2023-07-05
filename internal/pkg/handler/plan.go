package handler

import (
	"context"
	"encoding/json"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/errno"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/jwt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/log"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/record"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/response"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/uuid"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/notice"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/plan"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/stress"
	"github.com/gin-gonic/gin"
	"strings"
	"sync"
	"time"

	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/query"
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

	errnoNum, newReportIDs, runErr := RunStress(ctx, runStressParams)
	if runErr != nil {
		response.ErrorWithMsg(ctx, errnoNum, runErr.Error())
		return
	}

	px := dal.GetQuery().StressPlan
	planInfo, err := px.WithContext(ctx).Where(px.TeamID.Eq(req.TeamID), px.PlanID.Eq(req.PlanID)).First()
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	// 插入操作日志
	if planInfo.TaskType == consts.PlanTaskTypeNormal || runStressParams.RunType == 2 {
		if err := record.InsertRun(ctx, req.TeamID, jwt.GetUserIDByCtx(ctx), record.OperationOperateRunPlan, planInfo.PlanName); err != nil {
			response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
			return
		}
	} else {
		if err := record.InsertExecute(ctx, req.TeamID, jwt.GetUserIDByCtx(ctx), record.OperationOperateExecPlan, planInfo.PlanName); err != nil {
			response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
			return
		}
	}

	// 执行计划当次操作和报告的关系
	planRunUUID := uuid.GetUUID()
	planRunUUIDRedisKey := consts.RedisPlanRunUUIDRelateReports + planRunUUID
	_ = dal.GetRDB().SAdd(ctx, planRunUUIDRedisKey, newReportIDs).Err()
	_ = dal.GetRDB().Expire(ctx, planRunUUIDRedisKey, time.Second*86400).Err()
	for _, r := range newReportIDs {
		reportPlanRunRedisKey := consts.RedisReportPlanRunUUID + r
		_ = dal.GetRDB().Set(ctx, reportPlanRunRedisKey, planRunUUID, time.Second*86400).Err()
	}

	//tx := dal.GetQuery().StressPlanEmail
	//emails, err := tx.WithContext(ctx).Where(tx.TeamID.Eq(req.TeamID), tx.PlanID.Eq(req.PlanID)).Find()
	//if err != nil {
	//	response.ErrorWithMsg(ctx, errno.ErrHttpFailed, err.Error())
	//	return
	//}
	//if len(emails) > 0 {
	//	ttx := dal.GetQuery().Team
	//	team, err := ttx.WithContext(ctx).Where(ttx.TeamID.Eq(req.TeamID)).First()
	//	if err != nil {
	//		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
	//		return
	//	}
	//
	//	rx := dal.GetQuery().StressPlanReport
	//	reports, err := rx.WithContext(ctx).Where(rx.ReportID.In(newReportIDs...)).Find()
	//	if err != nil {
	//		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
	//		return
	//	}
	//
	//	ux := dal.GetQuery().User
	//	userInfo, err := ux.WithContext(ctx).Where(ux.UserID.Eq(jwt.GetUserIDByCtx(ctx))).First()
	//	if err != nil {
	//		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
	//		return
	//	}
	//
	//	if len(reports) > 0 {
	//		for _, email := range emails {
	//			if err := mail.SendPlanEmail(email.Email, planInfo.PlanName, team.Name, userInfo.Nickname, reports, userInfo); err != nil {
	//				if err.Error() == "请配置邮件相关环境变量" {
	//					response.ErrorWithMsg(ctx, errno.ErrNotEmailConfig, err.Error())
	//				} else {
	//					response.ErrorWithMsg(ctx, errno.ErrHttpFailed, err.Error())
	//				}
	//				return
	//			}
	//		}
	//	}
	//}

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

	// 判断是否是分布式
	if req.IsOpenDistributed == 1 { // 分布式
		if req.MachineDispatchModeConf.MachineAllotType == 0 { // 权重
			// 校验权重之和
			totalWeight := 0
			for _, v := range req.MachineDispatchModeConf.UsableMachineList {
				totalWeight = totalWeight + v.Weight
			}
			if totalWeight != 100 {
				response.ErrorWithMsg(ctx, errno.ErrMustTaskInit, "所有压力机权重值相加必须等于100")
				return
			}
			// 校验必填项
			if req.Mode == consts.PlanModeConcurrence || req.Mode == consts.PlanModeRound { // 并发模式或轮次模式
				if (req.ModeConf.Duration == 0 && req.ModeConf.RoundNum == 0) || req.ModeConf.Concurrency == 0 {
					response.ErrorWithMsg(ctx, errno.ErrMustTaskInit, "必填项不能为空")
					return
				}
			} else { // 非并发模式
				if req.ModeConf.StartConcurrency == 0 || req.ModeConf.Step == 0 || req.ModeConf.StepRunTime == 0 || req.ModeConf.MaxConcurrency == 0 || req.ModeConf.Duration == 0 {
					response.ErrorWithMsg(ctx, errno.ErrMustTaskInit, "必填项不能为空")
					return
				}
			}
		} else { // 自定义
			// 校验必填项
			if req.Mode == consts.PlanModeConcurrence || req.Mode == consts.PlanModeRound { // 并发模式或轮次模式
				// 校验总并发数之和
				var totalConcurrency int64 = 0
				for _, v := range req.MachineDispatchModeConf.UsableMachineList {
					totalConcurrency = totalConcurrency + v.Concurrency
				}
				if totalConcurrency == 0 {
					response.ErrorWithMsg(ctx, errno.ErrMustTaskInit, "所有压力机并发数相加不能等于0")
					return
				}

				// 校验持续时长或轮次之和
				for _, v := range req.MachineDispatchModeConf.UsableMachineList {
					if v.MachineStatus == 1 {
						if v.RoundNum == 0 && v.Duration == 0 {
							response.ErrorWithMsg(ctx, errno.ErrMustTaskInit, "必填项不能为空")
							return
						}
					}
				}
			} else { // 非并发模式
				// 校验持续时长或轮次之和
				for _, v := range req.MachineDispatchModeConf.UsableMachineList {
					if v.StartConcurrency == 0 || v.Step == 0 || v.StepRunTime == 0 || v.MaxConcurrency == 0 || v.Duration == 0 {
						response.ErrorWithMsg(ctx, errno.ErrMustTaskInit, "必填项不能为空")
						return
					}
				}
			}
		}
		if len(req.MachineDispatchModeConf.UsableMachineList) == 0 {
			response.ErrorWithMsg(ctx, errno.ErrMustTaskInit, "分布式配置下，必须选择至少一个可用压力机")
			return
		}
	} else { // 智能调度
		if req.Mode == 1 { // 并发模式
			if req.ModeConf.Duration == 0 || req.ModeConf.Concurrency == 0 {
				response.ErrorWithMsg(ctx, errno.ErrMustTaskInit, "必填项不能为空")
				return
			}
		}

		if req.Mode == 6 { // 轮次模式
			if req.ModeConf.RoundNum == 0 || req.ModeConf.Concurrency == 0 {
				response.ErrorWithMsg(ctx, errno.ErrMustTaskInit, "必填项不能为空")
				return
			}
		}

		if req.Mode > 1 && req.Mode < 6 { // 非并发模式参数校验
			if req.ModeConf.StartConcurrency == 0 || req.ModeConf.Step == 0 || req.ModeConf.StepRunTime == 0 || req.ModeConf.MaxConcurrency == 0 || req.ModeConf.Duration == 0 {
				response.ErrorWithMsg(ctx, errno.ErrMustTaskInit, "必填项不能为空")
				return
			}
			if req.ModeConf.MaxConcurrency < req.ModeConf.StartConcurrency {
				response.ErrorWithMsg(ctx, errno.ErrMaxConcurrencyLessThanStartConcurrency, "")
				return
			}
		}
	}

	errNum, err := plan.SaveTask(ctx, &req)
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

	scenes, err := plan.ImportScene(ctx, &req)
	if err != nil {
		if err.Error() == "计划内目录不可重名" {
			response.ErrorWithMsg(ctx, errno.ErrInPlanFolderNameAlreadyExist, err.Error())
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
func RunStress(ctx context.Context, req RunStressReq) (int, []string, error) {
	rms := &stress.RunMachineStress{}

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

	batonData := &stress.Baton{
		Ctx:      ctx,
		PlanID:   req.PlanID,
		TeamID:   req.TeamID,
		UserID:   req.UserID,
		SceneIDs: req.SceneID,
		RunType:  req.RunType,
	}
	errnoNum, err := ctt.Execute(batonData)

	return errnoNum, batonData.ReportIDs, err
}

// NotifyStopStress 压力机回调压测状态和结果
func NotifyStopStress(ctx *gin.Context) {
	var req rao.NotifyStopStressReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	log.Logger.Info("NotifyStopStress--性能测试回调接口参数：", req)

	// 设置锁，防止同一个计划下的报告，并发回调当前接口
	notifyStopStressKey := "NotifyStopStress:" + req.PlanID
	defer func() {
		dal.GetRDB().Del(ctx, notifyStopStressKey)
	}()
	// 设置锁
	for i := 0; i < 3; i++ {
		err := dal.GetRDB().SetNX(ctx, notifyStopStressKey, 1, 3*time.Second).Err()
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}

	// 修改报告状态
	tx := dal.GetQuery().StressPlanReport
	_, err := tx.WithContext(ctx).Where(tx.TeamID.Eq(req.TeamID),
		tx.PlanID.Eq(req.PlanID),
		tx.ReportID.Eq(req.ReportID)).UpdateSimple(tx.Status.Value(consts.ReportStatusFinish))
	if err != nil {
		log.Logger.Info("NotifyStopStress--修改报告状态失败")
		response.ErrorWithMsg(ctx, errno.ErrOperationFail, err.Error())
		return
	}
	// 是否发送通知
	var (
		isSendNotice bool
		newReportIDs []string
	)

	reportInfo := new(model.StressPlanReport)
	allErr := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		// 查询当前计划信息
		planInfo, err := tx.StressPlan.WithContext(ctx).Where(tx.StressPlan.TeamID.Eq(req.TeamID),
			tx.StressPlan.PlanID.Eq(req.PlanID)).First()
		if err != nil {
			return err
		}

		// 查找报告对应计划
		report, err := tx.StressPlanReport.WithContext(ctx).Where(tx.StressPlanReport.TeamID.Eq(req.TeamID),
			tx.StressPlanReport.PlanID.Eq(req.PlanID), tx.StressPlanReport.ReportID.Eq(req.ReportID)).First()
		if err != nil {
			log.Logger.Info("NotifyStopStress--查找报告对应计划失败")
			return err
		}
		reportInfo = report

		if planInfo.TaskType == consts.PlanTaskTypeNormal {
			// 统计报告是否全部完成
			reportCnt, err := tx.StressPlanReport.WithContext(ctx).Where(tx.StressPlanReport.TeamID.Eq(req.TeamID),
				tx.StressPlanReport.PlanID.Eq(req.PlanID)).Count()
			if err != nil {
				log.Logger.Info("NotifyStopStress--统计当前计划下所有的报告数量--失败")
				return err
			}
			finishReportCnt, err := tx.StressPlanReport.WithContext(ctx).Where(tx.StressPlanReport.TeamID.Eq(req.TeamID),
				tx.StressPlanReport.PlanID.Eq(report.PlanID), tx.StressPlanReport.Status.Eq(consts.ReportStatusFinish)).Count()
			if err != nil {
				log.Logger.Info("NotifyStopStress--统计当前计划下所有成功的报告--失败")
				return err
			}

			if finishReportCnt == reportCnt { // 报告全部完成则计划也完成
				_, err := tx.StressPlan.WithContext(ctx).Where(tx.StressPlan.TeamID.Eq(req.TeamID),
					tx.StressPlan.PlanID.Eq(report.PlanID)).UpdateSimple(tx.StressPlan.Status.Value(consts.PlanStatusNormal))
				if err != nil {
					log.Logger.Info("NotifyStopStress计划下所有报告并未全部完成")
					return err
				}
				isSendNotice = true
				reportPlanRunRedisKey := consts.RedisReportPlanRunUUID + req.ReportID
				planRunUUID, err := dal.GetRDB().Get(ctx, reportPlanRunRedisKey).Result()
				if err != nil {
					log.Logger.Info("NotifyStopStress--获取报告执行的操作ID失败：", err)
				}
				planRunUUIDRedisKey := consts.RedisPlanRunUUIDRelateReports + planRunUUID
				planRunReportIDs, err := dal.GetRDB().SMembers(ctx, planRunUUIDRedisKey).Result()
				if err != nil {
					log.Logger.Info("NotifyStopStress--获取执行操作ID报告失败：", err)
				}
				newReportIDs = append(newReportIDs, planRunReportIDs...)
			}
		} else { // 定时任务
			// 判断定时任务频次
			TimedTaskConfInfo, err := tx.StressPlanTimedTaskConf.WithContext(ctx).Where(tx.StressPlanTimedTaskConf.TeamID.Eq(req.TeamID),
				tx.StressPlanTimedTaskConf.PlanID.Eq(req.PlanID), tx.StressPlanTimedTaskConf.SceneID.Neq(reportInfo.SceneID)).First()
			nowTime := time.Now().Unix()
			if err == nil {
				if TimedTaskConfInfo.Frequency == 0 || (TimedTaskConfInfo.Frequency != 0 && TimedTaskConfInfo.TaskCloseTime <= nowTime) {
					// 查到定时任务配置了,如果任务配置过期时间小于当前时间，则把计划状态改为未运行
					_, err := tx.StressPlan.WithContext(ctx).Where(tx.StressPlan.TeamID.Eq(reportInfo.TeamID),
						tx.StressPlan.PlanID.Eq(reportInfo.PlanID)).UpdateSimple(tx.StressPlan.Status.Value(consts.PlanStatusNormal))
					if err != nil {
						log.Logger.Info("NotifyStopStress--修改定时计划状态失败")
						response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
						return err
					}
				}
			}
			isSendNotice = true
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
		response.ErrorWithMsg(ctx, errno.ErrOperationFail, allErr.Error())
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

	// 发送通知
	if isSendNotice {
		noticeGroupIDs := make([]string, 0)
		nge := dal.GetQuery().ThirdNoticeGroupEvent
		if err := nge.WithContext(ctx).Where(
			nge.TeamID.Eq(req.TeamID),
			nge.PlanID.Eq(req.PlanID),
			nge.EventID.Eq(consts.NoticeEventStressPlan)).Pluck(nge.GroupID, &noticeGroupIDs); err != nil {
			log.Logger.Error("NotifyStopStress--查询通知组失败：", err)
		}
		if len(noticeGroupIDs) > 0 {
			if len(newReportIDs) <= 0 {
				newReportIDs = append(newReportIDs, req.ReportID)
			}
			sendNoticeReq := &rao.SendNoticeParams{
				EventID:        consts.NoticeEventStressPlan,
				TeamID:         req.TeamID,
				ReportIDs:      newReportIDs,
				NoticeGroupIDs: noticeGroupIDs,
			}
			params, err := notice.GetSendCardParamsByReq(ctx, sendNoticeReq)
			if err != nil {
				log.Logger.Error("NotifyStopStress--获取通知参数失败：", err)
			}
			for _, groupID := range noticeGroupIDs {
				if err := notice.SendNoticeByGroup(ctx, groupID, params); err != nil {
					log.Logger.Error("NotifyStopStress--发送通知失败：", err)
				}
			}
		}
	}

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

// GetPublicFunctionList 获取公共函数列表
func GetPublicFunctionList(ctx *gin.Context) {
	res, err := plan.GetPublicFunctionList(ctx)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}
	response.SuccessWithData(ctx, res)
	return
}
