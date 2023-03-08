package handler

import (
	"kp-management/internal/pkg/biz/errno"
	"kp-management/internal/pkg/biz/jwt"
	"kp-management/internal/pkg/biz/response"
	"kp-management/internal/pkg/dal/rao"
	"kp-management/internal/pkg/logic/api"
	"kp-management/internal/pkg/logic/target"

	"github.com/gin-gonic/gin"
)

// SendTarget 发送接口
func SendTarget(ctx *gin.Context) {
	var req rao.SendTargetReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	retID, err := target.SendAPI(ctx, req.TeamID, req.TargetID)
	if err != nil {
		if err.Error() == "调试接口返回非200状态" {
			response.ErrorWithMsg(ctx, errno.ErrHttpFailed, err.Error())
		} else {
			response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		}
		return
	}

	response.SuccessWithData(ctx, rao.SendTargetResp{RetID: retID})
	return
}

// GetSendTargetResult 获取发送接口结果
func GetSendTargetResult(ctx *gin.Context) {
	var req rao.GetSendTargetResultReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	r, err := target.GetSendAPIResult(ctx, req.RetID)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, r)
	return
}

// SaveTarget 创建/修改接口
func SaveTarget(ctx *gin.Context) {
	var req rao.SaveTargetReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	targetID, err := api.Save(ctx, &req, jwt.GetUserIDByCtx(ctx))
	if err != nil {
		if err.Error() == "名称已存在" {
			response.ErrorWithMsg(ctx, errno.ErrApiNameAlreadyExist, err.Error())
		} else {
			response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		}

		return
	}

	response.SuccessWithData(ctx, rao.SaveTargetResp{TargetID: targetID})
	return
}

// SaveImportApi 导入接口
func SaveImportApi(ctx *gin.Context) {
	var req rao.SaveImportApiReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	err := api.SaveImportApi(ctx, &req, jwt.GetUserIDByCtx(ctx))
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}
	response.Success(ctx)
	return
}

// SortTarget 排序
func SortTarget(ctx *gin.Context) {
	var req rao.SortTargetReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	if err := target.SortTarget(ctx, &req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}

// TrashTargetList 文件夹/接口回收站列表
func TrashTargetList(ctx *gin.Context) {
	var req rao.ListTrashTargetReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	targets, total, err := target.ListTrashFolderAPI(ctx, req.TeamID, req.Size, (req.Page-1)*req.Size)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.ListTrashTargetResp{
		Targets: targets,
		Total:   total,
	})
	return
}

// TrashTarget 移入回收站
func TrashTarget(ctx *gin.Context) {
	var req rao.DeleteTargetReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	if err := target.Trash(ctx, req.TargetID, jwt.GetUserIDByCtx(ctx)); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}

// RecallTarget 从回收站恢复
func RecallTarget(ctx *gin.Context) {
	var req rao.RecallTargetReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	err := target.Recall(ctx, req.TargetID)
	if err != nil {
		if err.Error() == "文件夹名称已存在" {
			response.ErrorWithMsg(ctx, errno.ErrFolderNameAlreadyExist, err.Error())
		} else if err.Error() == "接口名称已存在" {
			response.ErrorWithMsg(ctx, errno.ErrApiNameAlreadyExist, err.Error())
		} else {
			response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		}
		return
	}

	response.Success(ctx)
	return
}

// DeleteTarget 回收站彻底删除
func DeleteTarget(ctx *gin.Context) {
	var req rao.DeleteTargetReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	if err := target.Delete(ctx, req.TargetID); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}

// ListFolderAPI 文件夹/接口列表
func ListFolderAPI(ctx *gin.Context) {
	var req rao.ListFolderAPIReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	targets, err := target.ListFolderAPI(ctx, req.TeamID)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.ListFolderAPIResp{
		Targets: targets,
	})
	return
}

// ListGroupScene 分组/场景列表
func ListGroupScene(ctx *gin.Context) {
	var req rao.ListGroupSceneReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	targets, err := target.ListGroupScene(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.ListGroupSceneResp{
		Targets: targets,
	})
	return
}

// BatchGetTarget 获取接口详情
func BatchGetTarget(ctx *gin.Context) {
	var req rao.BatchGetDetailReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	targets, err := api.DetailByTargetIDs(ctx, req.TeamID, req.TargetIDs)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.BatchGetDetailResp{
		Targets: targets,
	})
	return
}
