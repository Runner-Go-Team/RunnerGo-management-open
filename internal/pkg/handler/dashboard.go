package handler

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/errno"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/jwt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/response"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/homePage"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/user"

	"github.com/gin-gonic/gin"
)

// DashboardDefault 首页控制台
func DashboardDefault(ctx *gin.Context) {
	var req rao.DashboardDefaultReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	u, err := user.FirstByUserID(ctx, jwt.GetUserIDByCtx(ctx), req.TeamID)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.DashboardDefaultResp{
		User:   u,
		Mobile: u.Mobile,
	})
	return
}

// HomePage 新的首页
func HomePage(ctx *gin.Context) {
	var req rao.HomePageReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	homePageData, err := homePage.HomePage(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, homePageData)
	return
}
