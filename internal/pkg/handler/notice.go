package handler

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/errno"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/jwt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/response"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/notice"
	"github.com/gin-gonic/gin"
)

// SaveNoticeEvent 三方通知绑定
func SaveNoticeEvent(ctx *gin.Context) {
	var req rao.SaveNoticeEventReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	err := notice.SaveNoticeEvent(ctx, jwt.GetUserIDByCtx(ctx), &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}

// GetGroupNoticeEvent 获取通知事件对应通知组ID
func GetGroupNoticeEvent(ctx *gin.Context) {
	var req rao.GetGroupNoticeEventReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	groupIDs, err := notice.GetGroupNoticeEvent(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.GetGroupNoticeEventResp{
		NoticeGroupIDs: groupIDs,
	})
	return
}

// SendNotice 发送通知
func SendNotice(ctx *gin.Context) {
	var req rao.SendNoticeReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	if len(req.ReportIDs) > 10 {
		response.ErrorWithMsg(ctx, errno.ErrNoticeBatchReportLimit, "")
		return
	}

	for _, reportID := range req.ReportIDs {
		reportIDs := make([]string, 0, 1)
		sendNoticeReq := &rao.SendNoticeParams{
			EventID:        req.EventID,
			TeamID:         req.TeamID,
			ReportIDs:      append(reportIDs, reportID),
			NoticeGroupIDs: req.NoticeGroupIDs,
		}
		cardParams, err := notice.GetSendCardParamsByReq(ctx, sendNoticeReq)
		if err != nil {
			response.ErrorWithMsg(ctx, errno.ErrServer, err.Error())
			return
		}

		for _, groupID := range req.NoticeGroupIDs {
			if err := notice.SendNoticeByGroup(ctx, groupID, cardParams); err != nil {
				response.ErrorWithMsg(ctx, errno.ErrNoticeConfigError, err.Error())
				return
			}
		}
	}

	response.Success(ctx)
	return
}
