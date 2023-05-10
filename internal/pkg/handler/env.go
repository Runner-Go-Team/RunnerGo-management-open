package handler

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/errno"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/response"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/env"
	"github.com/gin-gonic/gin"
	"strings"
)

// EnvList 获取环境列表
func EnvList(ctx *gin.Context) {
	var req rao.EnvListReq

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	isExist := strings.Index(req.Name, "%")
	if isExist >= 0 {
		response.SuccessWithData(ctx, []rao.EnvListResp{})
		return
	}

	resp, err := env.GetList(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, resp)
	return

}

// SaveEnv 保存/编辑环境
func SaveEnv(ctx *gin.Context) {
	var req rao.SaveEnvReq

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	//判断环境名称在同一团队下是否已存在
	envNameIsExist, _ := env.EnvNameIsExist(ctx, &req)
	if envNameIsExist == true {
		response.ErrorWithMsg(ctx, errno.ErrEnvNameIsExist, "")
		return
	}

	resp, err := env.SaveEnv(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, resp)
	return

}

// CopyEnv 复制环境
func CopyEnv(ctx *gin.Context) {
	var req rao.CopyEnvReq

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	resp, err := env.CopyEnv(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, resp)
	return

}

// DelEnv 保存/编辑环境
func DelEnv(ctx *gin.Context) {
	var req rao.DelEnvReq

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	err := env.DelEnv(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}

// DelService 删除环境
func DelService(ctx *gin.Context) {
	var req rao.DelServiceReq

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	err := env.DelService(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.Success(ctx)
	return

}
