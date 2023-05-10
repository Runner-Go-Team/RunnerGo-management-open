package handler

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/errno"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/response"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/group"

	"github.com/gin-gonic/gin"
)

// SaveGroup 创建/保存分组
func SaveGroup(ctx *gin.Context) {
	var req rao.SaveGroupReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	err := group.Save(ctx, &req)
	if err != nil {
		if err.Error() == "名称已存在" {
			response.ErrorWithMsg(ctx, errno.ErrGroupNameAlreadyExist, err.Error())
		} else {
			response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		}

		return
	}

	response.Success(ctx)
	return
}

// GetGroup 获取分组
func GetGroup(ctx *gin.Context) {
	var req rao.GetGroupReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	g, err := group.GetByTargetID(ctx, req.TeamID, req.TargetID, req.Source)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.GetGroupResp{Group: g})
	return
}
