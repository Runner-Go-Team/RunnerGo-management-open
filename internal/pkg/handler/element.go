package handler

import (
	"errors"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/errno"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/jwt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/response"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/element"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/errmsg"
	"github.com/gin-gonic/gin"
)

// ElementSaveFolder 创建文件夹
func ElementSaveFolder(ctx *gin.Context) {
	var req rao.ElementSaveFolderReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	elementID, err := element.FolderSave(ctx, jwt.GetUserIDByCtx(ctx), &req)
	if err != nil {
		if errors.Is(err, errmsg.ErrElementFolderNameRepeat) {
			response.ErrorWithMsg(ctx, errno.ErrElementFolderNameRepeat, "")
			return
		}
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())

		return
	}

	response.SuccessWithData(ctx, &rao.ElementSaveFolderResp{ElementID: elementID})
	return
}

// ElementUpdateFolder 修改文件夹
func ElementUpdateFolder(ctx *gin.Context) {
	var req rao.ElementSaveFolderReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	elementID, err := element.FolderUpdate(ctx, jwt.GetUserIDByCtx(ctx), &req)
	if err != nil {
		if errors.Is(err, errmsg.ErrElementFolderNameRepeat) {
			response.ErrorWithMsg(ctx, errno.ErrElementFolderNameRepeat, "")
			return
		}
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())

		return
	}

	response.SuccessWithData(ctx, &rao.ElementSaveFolderResp{ElementID: elementID})
	return
}

// ElementRemoveFolder 元素删除
func ElementRemoveFolder(ctx *gin.Context) {
	var req rao.ElementRemoveFolderReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	err := element.FolderRemove(ctx, jwt.GetUserIDByCtx(ctx), req.TeamID, req.ElementIDs)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}

// ElementSortFolder 元素目录排序
func ElementSortFolder(ctx *gin.Context) {
	var req rao.ElementSortFolderReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	err := element.SortFolder(ctx, jwt.GetUserIDByCtx(ctx), &req)
	if err != nil {
		if errors.Is(err, errmsg.ErrTargetSortNameAlreadyExist) {
			response.ErrorWithMsg(ctx, errno.ErrTargetSortNameAlreadyExist, "")
			return
		}
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}

// ElementListFolder 元素目录列表
func ElementListFolder(ctx *gin.Context) {
	var req rao.ElementFolderListReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	list, err := element.FolderList(ctx, req.TeamID)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, &rao.ElementFolderListResp{Folder: list})
	return
}

// ElementSave 创建元素
func ElementSave(ctx *gin.Context) {
	var req rao.ElementSaveReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	elementID, err := element.Save(ctx, jwt.GetUserIDByCtx(ctx), &req)
	if err != nil {
		if errors.Is(err, errmsg.ErrElementNameRepeat) {
			response.ErrorWithMsg(ctx, errno.ErrElementNameRepeat, "")
			return
		}
		if errors.Is(err, errmsg.ErrElementLocatorNotFound) {
			response.ErrorWithMsg(ctx, errno.ErrElementLocatorNotFound, "")
			return
		}
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())

		return
	}

	response.SuccessWithData(ctx, &rao.ElementSaveFolderResp{ElementID: elementID})
	return
}

// ElementUpdate 修改元素
func ElementUpdate(ctx *gin.Context) {
	var req rao.ElementSaveReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	elementID, err := element.Update(ctx, jwt.GetUserIDByCtx(ctx), &req)
	if err != nil {
		if errors.Is(err, errmsg.ErrElementNameRepeat) {
			response.ErrorWithMsg(ctx, errno.ErrElementNameRepeat, "")
			return
		}
		if errors.Is(err, errmsg.ErrElementLocatorNotFound) {
			response.ErrorWithMsg(ctx, errno.ErrElementLocatorNotFound, "")
			return
		}
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())

		return
	}

	response.SuccessWithData(ctx, &rao.ElementSaveFolderResp{ElementID: elementID})
	return
}

// ElementDetail 元素详情
func ElementDetail(ctx *gin.Context) {
	var req rao.ElementDetailReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	info, err := element.Detail(ctx, req.ElementID, req.TeamID)
	if err != nil {
		if errors.Is(err, errmsg.ErrElementNotFound) {
			response.ErrorWithMsg(ctx, errno.ErrElementNotFound, "")
			return
		}
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, &rao.ElementDetailResp{Element: info})
	return
}

// ElementList 元素列表
func ElementList(ctx *gin.Context) {
	var req rao.ElementListReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	list, total, err := element.List(ctx, req.TeamID, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, &rao.ElementListResp{
		Elements: list,
		Total:    total,
	})
	return
}

// ElementRemove 元素删除
func ElementRemove(ctx *gin.Context) {
	var req rao.ElementRemoveReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	err := element.Remove(ctx, jwt.GetUserIDByCtx(ctx), req.TeamID, req.ElementIDs)
	if err != nil {
		if errors.Is(err, errmsg.ErrElementNotDeleteReScene) {
			response.ErrorWithMsg(ctx, errno.ErrElementNotDeleteReScene, "")
			return
		}
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}

// ElementSort 元素排序
func ElementSort(ctx *gin.Context) {
	var req rao.ElementSortReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	err := element.Sort(ctx, jwt.GetUserIDByCtx(ctx), &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}
