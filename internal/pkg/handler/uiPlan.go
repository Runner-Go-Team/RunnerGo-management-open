package handler

import (
	"errors"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/errno"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/jwt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/response"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/errmsg"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/uiPlan"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/uiReport"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/public"
	"github.com/gin-gonic/gin"
	"time"
)

// UIPlanSave 创建
func UIPlanSave(ctx *gin.Context) {
	var req rao.UIPlanSaveReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	if len(req.Name) == 0 {
		response.ErrorWithMsg(ctx, errno.ErrUISceneNameNotEmpty, "")
		return
	}
	if public.GetStringNum(req.Name) > 30 {
		response.ErrorWithMsg(ctx, errno.ErrUISceneNameLong, "")
		return
	}
	if len(req.HeadUserIDs) == 0 {
		response.ErrorWithMsg(ctx, errno.ErrCheckParams, "负责人不能为空")
		return
	}

	err := uiPlan.Save(ctx, jwt.GetUserIDByCtx(ctx), &req)
	if err != nil {
		if errors.Is(err, errmsg.ErrUISceneFolderNameRepeat) {
			response.ErrorWithMsg(ctx, errno.ErrUISceneFolderNameRepeat, "")
			return
		}
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}

// UIPlanUpdate 修改
func UIPlanUpdate(ctx *gin.Context) {
	var req rao.UIPlanSaveReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	if len(req.Name) == 0 {
		response.ErrorWithMsg(ctx, errno.ErrUISceneNameNotEmpty, "")
		return
	}
	if public.GetStringNum(req.Name) > 30 {
		response.ErrorWithMsg(ctx, errno.ErrUISceneNameLong, "")
		return
	}
	if len(req.HeadUserIDs) == 0 {
		response.ErrorWithMsg(ctx, errno.ErrCheckParams, "负责人不能为空")
		return
	}

	err := uiPlan.Update(ctx, jwt.GetUserIDByCtx(ctx), &req)
	if err != nil {
		if errors.Is(err, errmsg.ErrUISceneFolderNameRepeat) {
			response.ErrorWithMsg(ctx, errno.ErrUISceneFolderNameRepeat, "")
			return
		}
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}

// UIPlanList 测试计划列表
func UIPlanList(ctx *gin.Context) {
	var req rao.ListUIPlanReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	plans, total, err := uiPlan.ListByTeamID(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.ListUIPlanResp{
		Plans: plans,
		Total: total,
	})
	return
}

// UIPlanImportScene 导入场景
func UIPlanImportScene(ctx *gin.Context) {
	var req rao.UIPlanImportSceneReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	if len(req.SceneIDs) == 0 {
		response.ErrorWithMsg(ctx, errno.ErrParam, "导入场景不能为空")
		return
	}

	userID := jwt.GetUserIDByCtx(ctx)
	err := uiPlan.ImportScene(ctx, userID, &req)
	if err != nil {
		if errors.Is(err, errmsg.ErrUISceneFolderNameRepeat) {
			response.ErrorWithMsg(ctx, errno.ErrUISceneFolderNameRepeat, "")
			return
		}
		if errors.Is(err, errmsg.ErrUISceneNameRepeat) {
			response.ErrorWithMsg(ctx, errno.ErrUISceneNameRepeat, "")
			return
		}
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}

// UIPlanSetSceneSyncMode 修改同步方式
func UIPlanSetSceneSyncMode(ctx *gin.Context) {
	var req rao.UIPlanSetSceneSyncModeReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	userID := jwt.GetUserIDByCtx(ctx)
	err := uiPlan.SetSceneSyncMode(ctx, userID, &req)
	if err != nil {
		if errors.Is(err, errmsg.ErrUISceneFolderNameRepeat) {
			response.ErrorWithMsg(ctx, errno.ErrUISceneFolderNameRepeat, "")
			return
		}
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}

// UIPlanHandSyncLastData 手动同步
func UIPlanHandSyncLastData(ctx *gin.Context) {
	var req rao.UIPlanSetSceneHandSyncReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	userID := jwt.GetUserIDByCtx(ctx)
	err := uiPlan.HandSyncLastData(ctx, userID, &req)
	if err != nil {
		if errors.Is(err, errmsg.ErrUISceneFolderNameRepeat) {
			response.ErrorWithMsg(ctx, errno.ErrUISceneFolderNameRepeat, "")
			return
		}
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}

// UIPlanDetail 获取计划详情
func UIPlanDetail(ctx *gin.Context) {
	var req rao.UIPlanDetailReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	detail, err := uiPlan.Detail(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}
	response.SuccessWithData(ctx, rao.UIPlanDetailResp{
		Plan: detail,
	})
	return
}

// UIPlanDelete 删除
func UIPlanDelete(ctx *gin.Context) {
	var req rao.UIPlanDeleteReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	err := uiPlan.Delete(ctx, jwt.GetUserIDByCtx(ctx), &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}
	response.Success(ctx)
	return
}

// UIPlanCopy 复制计划
func UIPlanCopy(ctx *gin.Context) {
	var req rao.UIPlanCopyReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	err := uiPlan.Copy(ctx, jwt.GetUserIDByCtx(ctx), &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrOperationFail, err.Error())
		return
	}
	response.Success(ctx)
	return
}

// UIPlanSaveTaskConf 保存计划配置--普通任务
func UIPlanSaveTaskConf(ctx *gin.Context) {
	var req rao.UIPlanSaveTaskConfReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	// 检测定时任务时间
	nowTime := time.Now().Unix()
	if req.TaskType == consts.UIPlanTaskTypeCronjob {
		if req.Frequency == consts.UIPlanFrequencyOnce {
			if req.TaskExecTime == 0 {
				response.ErrorWithMsg(ctx, errno.ErrMustTaskInit, "必填项不能为空")
				return
			}
			if req.TaskExecTime < nowTime {
				response.ErrorWithMsg(ctx, errno.ErrParam, "开始时间不能早于当前时间")
				return
			}

			req.TaskCloseTime = req.TaskExecTime + 120
		} else if req.Frequency > consts.UIPlanFrequencyOnce && req.Frequency < consts.UIPlanFrequencyFixedTime {
			if req.TaskExecTime == 0 || req.TaskCloseTime == 0 {
				response.ErrorWithMsg(ctx, errno.ErrMustTaskInit, "必填项不能为空")
				return
			}
			if req.TaskExecTime < nowTime {
				response.ErrorWithMsg(ctx, errno.ErrParam, "开始时间不能早于当前时间")
				return
			}

			if req.TaskExecTime > req.TaskCloseTime {
				response.ErrorWithMsg(ctx, errno.ErrParam, "结束时间不能小于开始时间")
				return
			}
		} else {
			if req.FixedIntervalStartTime == 0 || req.FixedIntervalTime == 0 ||
				req.FixedRunNum == 0 {
				response.ErrorWithMsg(ctx, errno.ErrMustTaskInit, "必填项不能为空")
				return
			}
			if req.FixedIntervalStartTime < nowTime {
				response.ErrorWithMsg(ctx, errno.ErrParam, "开始时间不能小于当前时间")
				return
			}
		}
	}

	userID := jwt.GetUserIDByCtx(ctx)
	err := uiPlan.SaveTaskConf(ctx, userID, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrOperationFail, err.Error())
		return
	}
	response.Success(ctx)
	return
}

// UIPlanGetTaskConf 获取配置
func UIPlanGetTaskConf(ctx *gin.Context) {
	var req rao.UIPlanGetTaskConfReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	taskConf, err := uiPlan.GetTaskConf(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}
	response.SuccessWithData(ctx, *taskConf)
	return
}

// RunUIPlan 启动计划
func RunUIPlan(ctx *gin.Context) {
	var req rao.RunUIPlanReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	userID := jwt.GetUserIDByCtx(ctx)
	reportID, err := uiPlan.RunOrStartCron(ctx, userID, &req)
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

// StopUIPlan 停止计划
func StopUIPlan(ctx *gin.Context) {
	var req rao.StopUIPlanReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	userID := jwt.GetUserIDByCtx(ctx)
	err := uiReport.StopPlan(ctx, userID, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrOperationFail, err.Error())
		return
	}

	response.Success(ctx)
	return
}
