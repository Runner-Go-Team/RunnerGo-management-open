package handler

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/errno"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/response"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/permission"
	"github.com/gin-gonic/gin"
)

// GetTeamCompanyMembers 获取当前团队和企业成员关系
func GetTeamCompanyMembers(ctx *gin.Context) {
	var req rao.GetTeamCompanyMembersReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	retData, _ := permission.GetTeamCompanyMembers(ctx, &req)
	response.SuccessWithData(ctx, retData)
	return
}

func TeamMembersSave(ctx *gin.Context) {
	var req rao.TeamMembersSaveReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	retData, _ := permission.TeamMembersSave(ctx, &req)
	response.SuccessWithData(ctx, retData)
	return
}

func GetRoleMemberInfo(ctx *gin.Context) {
	var req rao.GetRoleMemberInfoReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	retData, _ := permission.GetRoleMemberInfo(ctx, &req)
	response.SuccessWithData(ctx, retData)
	return
}

func UserGetMarks(ctx *gin.Context) {
	retData, _ := permission.UserGetMarks(ctx)
	response.SuccessWithData(ctx, retData)
	return
}

func GetNoticeGroupList(ctx *gin.Context) {
	var req rao.GetNoticeGroupListReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	retData, _ := permission.GetNoticeGroupList(ctx, &req)
	response.SuccessWithData(ctx, retData)
	return
}

func GetNoticeThirdUsers(ctx *gin.Context) {
	var req rao.GetNoticeThirdUsersReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	retData, _ := permission.GetNoticeThirdUsers(ctx, &req)
	response.SuccessWithData(ctx, retData)
	return
}
