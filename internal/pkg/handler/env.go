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

// GetEnvList  获取环境列表
func GetEnvList(ctx *gin.Context) {
	var req rao.GetEnvListReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	isExist := strings.Index(req.Name, "%")
	if isExist >= 0 {
		response.SuccessWithData(ctx, []rao.EnvListResp{})
		return
	}
	resp, err := env.GetEnvList(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}
	response.SuccessWithData(ctx, resp)
	return
}

// UpdateEnv 编辑环境
func UpdateEnv(ctx *gin.Context) {
	var req rao.UpdateEnvReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	err := env.UpdateEnv(ctx, &req)
	if err != nil {
		if err.Error() == "名称已存在" {
			response.ErrorWithMsg(ctx, errno.ErrEnvNameAlreadyExist, err.Error())
		} else {
			response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		}
		return
	}
	response.Success(ctx)
	return
}

// CreateEnv 新建环境
func CreateEnv(ctx *gin.Context) {
	var req rao.CreateEnvReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	envInfo, err := env.CreateEnv(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}
	response.SuccessWithData(ctx, envInfo)
	return
}

// CopyEnv 复制环境
func CopyEnv(ctx *gin.Context) {
	var req rao.CopyEnvReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	err := env.CopyEnv(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}
	response.Success(ctx)
	return
}

// DelEnv 删除环境
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

// DelEnvService 删除环境
func DelEnvService(ctx *gin.Context) {
	var req rao.DelEnvServiceReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	err := env.DelEnvService(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}
	response.Success(ctx)
	return
}

// DelEnvDatabase 删除环境下的数据库
func DelEnvDatabase(ctx *gin.Context) {
	var req rao.DelEnvDatabaseReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	err := env.DelEnvDatabase(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}
	response.Success(ctx)
	return
}

func GetServiceList(ctx *gin.Context) {
	var req rao.GetServiceListReq
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

	serviceList, total, err := env.GetServiceList(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}
	response.SuccessWithData(ctx, rao.GetServiceListResp{
		ServiceList: serviceList,
		Total:       total,
	})
	return
}

func GetDatabaseList(ctx *gin.Context) {
	var req rao.GetDatabaseListReq
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

	dbList, total, err := env.GetDatabaseList(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}
	response.SuccessWithData(ctx, rao.GetDatabaseListResp{
		DatabaseList: dbList,
		Total:        total,
	})
	return
}

func CreateEnvService(ctx *gin.Context) {
	var req rao.CreateEnvServiceReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	err := env.CreateEnvService(ctx, &req)
	if err != nil {
		if err.Error() == "名称已存在" {
			response.ErrorWithMsg(ctx, errno.ErrServiceNameAlreadyExist, err.Error())
		} else {
			response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		}
		return
	}
	response.Success(ctx)
	return
}

func CreateEnvDatabase(ctx *gin.Context) {
	var req rao.CreateEnvDatabaseReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	err := env.CreateEnvDatabase(ctx, &req)
	if err != nil {
		if err.Error() == "名称已存在" {
			response.ErrorWithMsg(ctx, errno.ErrServiceNameAlreadyExist, err.Error())
		} else {
			response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		}
		return
	}
	response.Success(ctx)
	return
}
