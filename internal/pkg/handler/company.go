package handler

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/errno"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/response"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/autoPlan"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/plan"
	"github.com/gin-gonic/gin"
)

// GetNewestStressPlanList 测试计划列表
func GetNewestStressPlanList(ctx *gin.Context) {
	var req rao.GetNewestStressPlanListReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	if req.Page == 0 {
		req.Page = 1
	}

	if req.Size == 0 {
		req.Size = 10
	}

	plans, _ := plan.GetNewestStressPlanList(ctx, &req)
	response.SuccessWithData(ctx, plans)
	return
}

// GetNewestAutoPlanList 获取最近自动化计划列表
func GetNewestAutoPlanList(ctx *gin.Context) {
	var req rao.GetNewestAutoPlanListReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	if req.Page == 0 {
		req.Page = 1
	}

	if req.Size == 0 {
		req.Size = 10
	}

	plans, _ := autoPlan.GetNewestAutoPlanList(ctx, &req)
	response.SuccessWithData(ctx, plans)
	return
}
