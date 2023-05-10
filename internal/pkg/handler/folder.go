package handler

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/errno"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/jwt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/response"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/folder"

	"github.com/gin-gonic/gin"
)

// SaveFolder 创建/修改文件夹
func SaveFolder(ctx *gin.Context) {
	var req rao.SaveFolderReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	err := folder.Save(ctx, jwt.GetUserIDByCtx(ctx), &req)
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

// GetFolder 获取文件夹
func GetFolder(ctx *gin.Context) {
	var req rao.GetFolderReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	f, err := folder.GetByTargetID(ctx, req.TeamID, req.TargetID)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.GetFolderResp{Folder: f})
	return
}
