package handler

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/errno"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/jwt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/response"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/conf"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/mock"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/target"
	"github.com/gin-gonic/gin"
)

func MockInfo(ctx *gin.Context) {
	var req rao.MockInfoReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	httpPreUrl := conf.Conf.Clients.Mock.HttpServer
	if len(req.TeamID) > 0 {
		httpPreUrl += "/" + req.TeamID + "/"
	}
	response.SuccessWithData(ctx, rao.MockInfoResp{HttpPreUrl: httpPreUrl})
	return
}

func MockSaveToTarget(ctx *gin.Context) {
	var req rao.MockSaveToTargetReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	err := mock.SaveToTarget(ctx, jwt.GetUserIDByCtx(ctx), req.TargetIDs, req.TeamID)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())

		return
	}

	response.Success(ctx)
	return
}

// MockSaveFolder 创建/修改文件夹
func MockSaveFolder(ctx *gin.Context) {
	var req rao.MockSaveFolderReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	err := mock.FolderReqSave(ctx, jwt.GetUserIDByCtx(ctx), &req)
	if err != nil {
		if err.Error() == "名称已存在" {
			response.ErrorWithMsg(ctx, errno.ErrFolderNameAlreadyExist, err.Error())
		} else {
			response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		}

		return
	}

	response.Success(ctx)
	return
}

// MockGetFolder 获取文件夹
func MockGetFolder(ctx *gin.Context) {
	var req rao.MockGetFolderReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	f, err := mock.GetByTargetID(ctx, req.TeamID, req.TargetID)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.MockGetFolderResp{Folder: f})
	return
}

// MockSaveTarget 创建/修改接口
func MockSaveTarget(ctx *gin.Context) {
	var req rao.MockSaveTargetReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	targetID := req.TargetID
	var err error

	targetID, err = mock.Save(ctx, &req, jwt.GetUserIDByCtx(ctx))
	if err != nil {
		if err.Error() == "名称已存在" {
			response.ErrorWithMsg(ctx, errno.ErrApiNameAlreadyExist, err.Error())
			return
		}
		if err.Error() == "路径已存在，不能重复" {
			response.ErrorWithMsg(ctx, errno.ErrMockPathExists, err.Error())
			return
		}
		if err.Error() == "路径不能为空" {
			response.ErrorWithMsg(ctx, errno.ErrMockPathNotNull, err.Error())
			return
		}
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.MockSaveTargetResp{TargetID: targetID})
	return
}

// MockSendTarget 发送接口
func MockSendTarget(ctx *gin.Context) {
	var req rao.MockSendTargetReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	retID, err := mock.SendAPI(ctx, req.TeamID, req.TargetID)
	if err != nil {
		if err.Error() == "调试接口返回非200状态" {
			response.ErrorWithMsg(ctx, errno.ErrHttpFailed, err.Error())
		} else {
			response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		}
		return
	}

	response.SuccessWithData(ctx, rao.MockSendTargetResp{RetID: retID})
	return
}

// MockGetSendTargetResult 获取发送接口结果
func MockGetSendTargetResult(ctx *gin.Context) {
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

// MockBatchGetTarget 获取接口详情
func MockBatchGetTarget(ctx *gin.Context) {
	var req rao.MockBatchGetDetailReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	targets, err := mock.DetailByTargetIDs(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.MockBatchGetDetailResp{
		Targets: targets,
	})
	return
}

// MockSortTarget 排序
func MockSortTarget(ctx *gin.Context) {
	var req rao.MockSortTargetReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	err := mock.SortTarget(ctx, &req)
	if err != nil {
		if err.Error() == "存在重名，无法操作" {
			response.ErrorWithMsg(ctx, errno.ErrTargetSortNameAlreadyExist, err.Error())
		} else {
			response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		}
		return
	}

	response.Success(ctx)
	return
}

// MockListFolderAPI 文件夹/接口列表
func MockListFolderAPI(ctx *gin.Context) {
	var req rao.MockListFolderAPIReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	targets, err := mock.ListFolderAPI(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.MockListFolderAPIResp{
		Targets: targets,
	})
	return
}

// MockTrashTarget 移入回收站
func MockTrashTarget(ctx *gin.Context) {
	var req rao.MockDeleteTargetReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	if err := mock.Trash(ctx, req.TargetID, jwt.GetUserIDByCtx(ctx)); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}
