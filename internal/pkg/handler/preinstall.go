package handler

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/errno"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/response"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/preinstall"
	"github.com/gin-gonic/gin"
	"strings"
)

// SavePreinstall 保存预设设置
func SavePreinstall(ctx *gin.Context) {
	var req rao.SavePreinstallReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	errNum, err := preinstall.SavePreinstall(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errNum, err.Error())
		return
	}

	response.Success(ctx)
	return
}

// GetPreinstallDetail 获取预设设置
func GetPreinstallDetail(ctx *gin.Context) {
	var req rao.GetPreinstallDetailReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	preinstallDetail, err := preinstall.GetPreinstallDetail(ctx, req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}
	response.SuccessWithData(ctx, preinstallDetail)
	return
}

// GetPreinstallList 获取预设配置列表
func GetPreinstallList(ctx *gin.Context) {
	var req rao.GetPreinstallListReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	isExist := strings.Index(req.ConfName, "%")
	if isExist >= 0 {
		response.SuccessWithData(ctx, rao.GetPreinstallResponse{})
		return
	}

	list, total, err := preinstall.GetPreinstallList(ctx, req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}
	response.SuccessWithData(ctx, rao.GetPreinstallResponse{
		PreinstallList: list,
		Total:          total,
	})
	return
}

// DeletePreinstall 删除预设配置
func DeletePreinstall(ctx *gin.Context) {
	var req rao.DeletePreinstallReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	err := preinstall.DeletePreinstall(ctx, req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}
	response.Success(ctx)
	return
}

// CopyPreinstall 复制预设配置数据
func CopyPreinstall(ctx *gin.Context) {
	var req rao.CopyPreinstallReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	err := preinstall.CopyPreinstall(ctx, req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}
	response.Success(ctx)
	return
}

func GetAvailableMachineList(ctx *gin.Context) {
	var req rao.GetAvailableMachineListReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	list, err := preinstall.GetAvailableMachineList(ctx)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}
	response.SuccessWithData(ctx, list)
	return
}
