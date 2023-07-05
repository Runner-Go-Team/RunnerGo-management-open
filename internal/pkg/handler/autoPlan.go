package handler

import (
	"context"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/errno"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/jwt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/mail"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/record"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/response"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/autoPlan"
	"github.com/gin-gonic/gin"
	"strings"
)

type RunAutoPlanReq struct {
	PlanID  string   `json:"plan_id"`
	TeamID  string   `json:"team_id"`
	SceneID []string `json:"scene_id"`
	UserID  string   `json:"user_id"`
	RunType int      `json:"source"` // 0，1-普通，2-定时
}

// RunAutoPlan 运行自动化测试计划
func RunAutoPlan(ctx *gin.Context) {
	var req rao.RunAutoPlanReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	// 调用controller方法改成本地
	runAutoPlanParams := RunAutoPlanReq{
		PlanID:  req.PlanID,
		TeamID:  req.TeamID,
		SceneID: req.SceneID,
		UserID:  jwt.GetUserIDByCtx(ctx),
	}

	errnoNum, runErr := RunAutoPlanDetail(ctx, runAutoPlanParams)
	if runErr != nil {
		response.ErrorWithMsg(ctx, errnoNum, runErr.Error())
		return
	}

	px := dal.GetQuery().AutoPlan
	autoPlanInfo, err := px.WithContext(ctx).Where(px.TeamID.Eq(req.TeamID), px.PlanID.Eq(req.PlanID)).First()
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	// 添加操作日志
	if autoPlanInfo.TaskType == consts.PlanTaskTypeNormal || runAutoPlanParams.RunType == 2 {
		if err := record.InsertRun(ctx, req.TeamID, jwt.GetUserIDByCtx(ctx), record.OperationOperateRunPlan, autoPlanInfo.PlanName); err != nil {
			response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
			return
		}
	} else {
		if err := record.InsertExecute(ctx, req.TeamID, jwt.GetUserIDByCtx(ctx), record.OperationOperateExecPlan, autoPlanInfo.PlanName); err != nil {
			response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
			return
		}
	}

	//if autoPlanInfo.TaskType == consts.PlanTaskTypeNormal {
	//	rx := dal.GetQuery().AutoPlanReport
	//	reportInfo, err := rx.WithContext(ctx).Where(rx.TeamID.Eq(req.TeamID), rx.PlanID.Eq(req.PlanID)).Order(rx.CreatedAt.Desc()).First()
	//	if err != nil {
	//		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
	//		return
	//	}
	//
	//	tx := dal.GetQuery().AutoPlanEmail
	//	emails, err := tx.WithContext(ctx).Where(tx.TeamID.Eq(req.TeamID), tx.PlanID.Eq(req.PlanID)).Find()
	//	if err == nil && len(emails) > 0 {
	//		ttx := dal.GetQuery().Team
	//		teamInfo, err := ttx.WithContext(ctx).Where(ttx.TeamID.Eq(req.TeamID)).First()
	//		if err != nil {
	//			response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
	//			return
	//		}
	//
	//		ux := dal.GetQuery().User
	//		user, err := ux.WithContext(ctx).Where(ux.UserID.Eq(reportInfo.RunUserID)).First()
	//		if err != nil {
	//			response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
	//			return
	//		}
	//
	//		for _, email := range emails {
	//			if err := mail.SendAutoPlanEmail(email.Email, autoPlanInfo, teamInfo, user.Nickname, reportInfo.ReportID); err != nil {
	//				response.ErrorWithMsg(ctx, errno.ErrHttpFailed, err.Error())
	//				return
	//			}
	//		}
	//	}
	//}
	response.SuccessWithData(ctx, rao.RunAutoPlanResp{
		TaskType: autoPlanInfo.TaskType,
	})
	return
}

// RunAutoPlanDetail 调度压力测试机进行压测的方法
func RunAutoPlanDetail(ctx context.Context, req RunAutoPlanReq) (int, error) {
	baton := &autoPlan.Baton{
		Ctx:      ctx,
		PlanID:   req.PlanID,
		TeamID:   req.TeamID,
		SceneIDs: req.SceneID,
		UserID:   req.UserID,
		RunType:  req.RunType,
	}

	// 1、校验团队是否过期
	//errnoNum, err := autoPlan.CheckTeamIsOverdue(baton)
	//if err != nil {
	//	return errnoNum, err
	//}

	// 2、组装计划相关数据
	errnoNum, err := autoPlan.AssemblePlan(baton)
	if err != nil {
		return errnoNum, err
	}

	// 组装场景相关数据
	errnoNum, err = autoPlan.AssembleScenes(baton)
	if err != nil {
		return errnoNum, err
	}

	// 组装场景下所有测试用例相关数据
	errnoNum, err = autoPlan.AssembleTestCase(baton)
	if err != nil {
		return errnoNum, err
	}

	// 组装计划配置数据
	errnoNum, err = autoPlan.AssembleTask(baton)
	if err != nil {
		return errnoNum, err
	}

	// 检查是否是定时计划
	errnoNum, err = autoPlan.CheckAutoPlanTaskType(baton)
	if err != nil {
		if err.Error() == "定时任务已经开启" {
			return errnoNum, nil
		}
		return errnoNum, err
	}

	// 1、检查运行计划机器情况
	errnoNum, err = autoPlan.CheckIdleMachine(baton)
	if err != nil {
		return errnoNum, err
	}

	// 组装所有场景flow
	errnoNum, err = autoPlan.AssembleSceneFlows(baton)
	if err != nil {
		return errnoNum, err
	}

	// 组装所有测试用例flow
	errnoNum, err = autoPlan.AssembleTestCaseFlows(baton)
	if err != nil {
		return errnoNum, err
	}

	// 组装全局变量数据
	errnoNum, err = autoPlan.AssembleGlobalVariables(baton)
	if err != nil {
		return errnoNum, err
	}

	// 组装场景变量数据
	errnoNum, err = autoPlan.AssembleVariable(baton)
	if err != nil {
		return errnoNum, err
	}

	// 7、组装导入变量数据
	errnoNum, err = autoPlan.AssembleImportVariables(baton)
	if err != nil {
		return errnoNum, err
	}

	// 生成报告列表信息
	errnoNum, err = autoPlan.MakeReport(baton)
	if err != nil {
		return errnoNum, err
	}

	// 组装运行计划最终参数
	errnoNum, err = autoPlan.AssembleRunPlanRealParams(baton)
	if err != nil {
		return errnoNum, err
	}

	// 运行自动化计划
	errnoNum, err = autoPlan.RunAutoPlan(baton)
	if err != nil {
		return errnoNum, err
	}

	// 当为定时任务运行时，运行完成发送邮件
	//if req.RunType == 2 {
	//	errnoNum, err = autoPlan.RunAutoPlanSendEmail(baton)
	//	if err != nil {
	//		return errnoNum, err
	//	}
	//}

	return errnoNum, err
}

// SaveAutoPlan 保存自动化测试计划
func SaveAutoPlan(ctx *gin.Context) {
	var req rao.SaveAutoPlanReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	planID, errNum, err := autoPlan.SaveAutoPlan(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errNum, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.SaveAutoPlanResp{PlanID: planID})
	return
}

// GetAutoPlanList 获取计划列表
func GetAutoPlanList(ctx *gin.Context) {
	var req rao.GetAutoPlanListReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	if req.Page == 0 {
		req.Page = 1
	}
	if req.Size == 0 {
		req.Size = 10
	}

	isExist := strings.Index(req.PlanName, "%")
	if isExist >= 0 {
		response.SuccessWithData(ctx, rao.AutoPlanListResp{})
		return
	}

	list, total, err := autoPlan.GetAutoPlanList(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.AutoPlanListResp{
		AutoPlanList: list,
		Total:        total,
	})
	return
}

// DeleteAutoPlan 删除自动化测试计划
func DeleteAutoPlan(ctx *gin.Context) {
	var req rao.DeleteAutoPlanReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	err := autoPlan.DeleteAutoPlan(ctx, &req)
	if err != nil {
		if err.Error() == "该计划正在运行，无法删除" {
			response.ErrorWithMsg(ctx, errno.ErrCannotDeleteRunningPlan, err.Error())
		} else {
			response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		}
		return
	}
	response.Success(ctx)
	return
}

// GetAutoPlanDetail 获取计划详情
func GetAutoPlanDetail(ctx *gin.Context) {
	var req rao.GetAutoPlanDetailReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	detail, err := autoPlan.GetAutoPlanDetail(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}
	response.SuccessWithData(ctx, detail)
	return
}

// CopyAutoPlan 复制计划
func CopyAutoPlan(ctx *gin.Context) {
	var req rao.CopyAutoPlanReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	err := autoPlan.CopyAutoPlan(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}
	response.Success(ctx)
	return
}

// UpdateAutoPlan 更新计划
func UpdateAutoPlan(ctx *gin.Context) {
	var req rao.UpdateAutoPlanReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	err := autoPlan.UpdateAutoPlan(ctx, &req)
	if err != nil {
		if err.Error() == "名称已存在" {
			response.ErrorWithMsg(ctx, errno.ErrPlanNameAlreadyExist, err.Error())
		} else {
			response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		}
		return
	}
	response.Success(ctx)
	return
}

// AddEmail 添加收件人
func AddEmail(ctx *gin.Context) {
	var req rao.AddEmailReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	// 单次限制添加50条
	if len(req.Emails) > 50 {
		response.ErrorWithMsg(ctx, errno.ErrAddEmailUserNumOvertopLimit, "单次只可添加1-50个收件人进行发送")
		return
	}

	err := autoPlan.AddEmail(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}
	response.Success(ctx)
	return
}

// GetEmailList 收件人邮箱列表
func GetEmailList(ctx *gin.Context) {
	var req rao.GetEmailListReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	if req.PlanCategory == 0 {
		req.PlanCategory = 1
	}

	emailList, err := autoPlan.GetEmailList(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}
	response.SuccessWithData(ctx, rao.GetEmailListResp{Emails: emailList})
	return
}

// DeleteEmail 删除收件人邮箱
func DeleteEmail(ctx *gin.Context) {
	var req rao.DeleteEmailReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	err := autoPlan.DeleteEmail(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}
	response.Success(ctx)
	return
}

// BatchDeleteAutoPlan 批量删除自动化测试计划
func BatchDeleteAutoPlan(ctx *gin.Context) {
	var req rao.BatchDeleteAutoPlanReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	err := autoPlan.BatchDeleteAutoPlan(ctx, &req)
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

// SaveTaskConf 保存计划配置--普通任务
func SaveTaskConf(ctx *gin.Context) {
	var req rao.SaveTaskConfReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	err := autoPlan.SaveTaskConf(ctx, &req)
	if err != nil {
		if err.Error() == "开始或结束时间不能早于当前时间" {
			response.ErrorWithMsg(ctx, errno.ErrTimedTaskOverdue, err.Error())
		} else {
			response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		}

		return
	}
	response.Success(ctx)
	return
}

func GetTaskConf(ctx *gin.Context) {
	var req rao.GetTaskConfReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	taskConf, err := autoPlan.GetTaskConf(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}
	response.SuccessWithData(ctx, *taskConf)
	return
}

// GetAutoPlanReportList 获取报告列表
func GetAutoPlanReportList(ctx *gin.Context) {
	var req rao.GetAutoPlanReportListReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	if req.Page == 0 {
		req.Page = 1
	}
	if req.Size == 0 {
		req.Size = 10
	}

	isExist := strings.Index(req.PlanName, "%")
	if isExist >= 0 {
		response.SuccessWithData(ctx, rao.GetAutoPlanReportListResp{})
		return
	}

	list, total, err := autoPlan.GetAutoPlanReportList(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.GetAutoPlanReportListResp{
		AutoPlanReportList: list,
		Total:              total,
	})
	return
}

// BatchDeleteAutoPlanReport 批量删除报告
func BatchDeleteAutoPlanReport(ctx *gin.Context) {
	var req rao.BatchDeleteAutoPlanReportReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	err := autoPlan.BatchDeleteAutoPlanReport(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}
	response.Success(ctx)
	return
}

func CloneAutoPlanScene(ctx *gin.Context) {
	var req rao.CloneAutoPlanSceneReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	err := autoPlan.CloneAutoPlanScene(ctx, &req)
	if err != nil {
		if err.Error() == "名称过长！不可超出30字符" {
			response.ErrorWithMsg(ctx, errno.ErrNameOverLength, err.Error())
		} else {
			response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		}

		return
	}
	response.Success(ctx)
	return
}

// NotifyRunFinish 自动化计划结束回调接口
func NotifyRunFinish(ctx *gin.Context) {
	var req rao.NotifyRunFinishReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	err := autoPlan.NotifyRunFinish(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}
	response.Success(ctx)
	return
}

// StopAutoPlan 停止计划
func StopAutoPlan(ctx *gin.Context) {
	var req rao.StopAutoPlanReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	err := autoPlan.StopAutoPlan(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}
	response.Success(ctx)
	return
}

// GetAutoPlanReportDetail 获取报告详情
func GetAutoPlanReportDetail(ctx *gin.Context) {
	var req rao.GetAutoPlanReportDetailReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	reportDetail, err := autoPlan.GetAutoPlanReportDetail(ctx, &req)
	//reportDetail, err := autoPlan.MakeAutoPlanReportDetail(ctx, &req) //动态查询报告详情
	if err != nil {
		if err.Error() == "计划正在运行中" {
			response.ErrorWithMsg(ctx, errno.ErrReportInRun, err.Error())
		} else if err.Error() == "报告不存在" {
			response.ErrorWithMsg(ctx, errno.ErrReportNotFound, err.Error())
		} else {
			response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		}
		return
	}
	response.SuccessWithData(ctx, *reportDetail)
	return
}

// ReportEmailNotify 报告详情邮件通知
func ReportEmailNotify(ctx *gin.Context) {
	var req rao.ReportEmailNotifyReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	tx := dal.GetQuery().Team
	team, err := tx.WithContext(ctx).Where(tx.TeamID.Eq(req.TeamID)).First()
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

	rx := dal.GetQuery().AutoPlanReport
	reportInfo, err := rx.WithContext(ctx).Where(rx.TeamID.Eq(req.TeamID), rx.ReportID.Eq(req.ReportID)).First()
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	for _, email := range req.Emails {
		if err := mail.SendAutoPlanReportEmail(email, req.ReportID, team, user, reportInfo); err != nil {
			if err.Error() == "请配置邮件相关环境变量" {
				response.ErrorWithMsg(ctx, errno.ErrNotEmailConfig, err.Error())
			} else {
				response.ErrorWithMsg(ctx, errno.ErrHttpFailed, err.Error())
			}
			return
		}
	}

	response.Success(ctx)
	return
}

func GetReportApiDetail(ctx *gin.Context) {
	var req rao.GetReportApiDetailReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	apiDetail, err := autoPlan.GetReportApiDetail(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrReportInRun, err.Error())
		return
	}
	response.SuccessWithData(ctx, apiDetail)
	return
}

func SendReportApi(ctx *gin.Context) {
	var req rao.SendReportApiReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	retID, err := autoPlan.SendReportApi(&req)
	if err != nil {
		if err.Error() == "调试接口返回非200状态" {
			response.ErrorWithMsg(ctx, errno.ErrHttpFailed, err.Error())
		} else {
			response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		}
		return
	}

	response.SuccessWithData(ctx, rao.SendReportApiResp{RetID: retID})
	return
}

func UpdateAutoPlanReportName(ctx *gin.Context) {
	var req rao.UpdateAutoPlanReportNameReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	err := autoPlan.UpdateAutoPlanReportName(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}
	response.Success(ctx)
	return
}
