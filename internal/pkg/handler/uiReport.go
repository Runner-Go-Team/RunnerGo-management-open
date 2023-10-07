package handler

import (
	"errors"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/errno"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/jwt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/response"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/errmsg"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/uiReport"
	"github.com/gin-gonic/gin"
)

// ListUIReports 报告列表
func ListUIReports(ctx *gin.Context) {
	var req rao.ListUIReportsReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	reports, total, err := uiReport.ListByTeamID2(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.ListUIReportsResp{
		Reports: reports,
		Total:   total,
	})
	return
}

// UIReportDetail 报告详情
func UIReportDetail(ctx *gin.Context) {
	var req rao.UIReportDetailReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	result, err := uiReport.GetReportDetail(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	response.SuccessWithData(ctx, result)
	return
}

// UIReportDelete 删除
func UIReportDelete(ctx *gin.Context) {
	var req rao.UIReportDeleteReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	err := uiReport.Delete(ctx, jwt.GetUserIDByCtx(ctx), &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}
	response.Success(ctx)
	return
}

// RunUIReport 启动
func RunUIReport(ctx *gin.Context) {
	var req rao.RunUIReportReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	userID := jwt.GetUserIDByCtx(ctx)
	reportID, err := uiReport.Run(ctx, userID, &req)
	if err != nil {
		if errors.Is(err, errmsg.ErrSendLinuxNotQTMode) {
			response.ErrorWithMsg(ctx, errno.ErrSendLinuxNotQTMode, "")
			return
		}
		if errors.Is(err, errmsg.ErrSendOperatorNotNull) {
			response.ErrorWithMsg(ctx, errno.ErrSendOperatorNotNull, "")
			return
		}
		if errors.Is(err, errmsg.ErrMustTaskInit) {
			response.ErrorWithMsg(ctx, errno.ErrMustTaskInit, "")
			return
		}
		if errors.Is(err, errmsg.ErrTimedTaskOverdue) {
			response.ErrorWithMsg(ctx, errno.ErrTimedTaskOverdue, "")
			return
		}
		response.ErrorWithMsg(ctx, errno.ErrUIEngineError, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.RunUIPlanResp{ReportID: reportID})
	return
}

// StopUIReport 停止报告
func StopUIReport(ctx *gin.Context) {
	var req rao.StopUIReportReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	userID := jwt.GetUserIDByCtx(ctx)
	err := uiReport.Stop(ctx, userID, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrOperationFail, err.Error())
		return
	}

	response.Success(ctx)
	return
}

// UIReportUpdate 修改报告
func UIReportUpdate(ctx *gin.Context) {
	var req rao.UIReportUpdateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	err := uiReport.Update(ctx, jwt.GetUserIDByCtx(ctx), &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}
	response.Success(ctx)
	return
}
