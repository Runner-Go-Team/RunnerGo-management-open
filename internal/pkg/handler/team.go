package handler

import (
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/encrypt"
	"math/rand"
	"time"

	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/errno"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/jwt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/response"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/conf"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/team"
	"github.com/gin-gonic/gin"
)

// SaveTeam 创建或修改团队
func SaveTeam(ctx *gin.Context) {
	var req rao.SaveTeamReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	teamID, err := team.SaveTeam(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.SaveTeamResp{
		TeamID: teamID,
	})
	return
}

// ListTeam 团队列表
func ListTeam(ctx *gin.Context) {
	teams, err := team.ListByUserID(ctx, jwt.GetUserIDByCtx(ctx))
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.ListTeamResp{Teams: teams})
	return
}

// ListTeamMembers 团队成员列表
func ListTeamMembers(ctx *gin.Context) {
	var req rao.ListMembersReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	members, total, err := team.ListMembersByTeamID(ctx, req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.ListMembersResp{
		Members: members,
		Total:   total,
	})
	return
}

// GetUserTeamRole 获取用户所在团队的角色
func GetUserTeamRole(ctx *gin.Context) {
	var req rao.GetTeamRoleReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	tx := dal.GetQuery().UserTeam
	ut, err := tx.WithContext(ctx).Where(tx.TeamID.Eq(req.TeamID), tx.UserID.Eq(jwt.GetUserIDByCtx(ctx))).First()
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.GetTeamRoleResp{
		RoleID: ut.RoleID,
	})
	return
}

// GetInviteMemberURL 获取邀请链接
func GetInviteMemberURL(ctx *gin.Context) {
	var req rao.GetInviteMemberURLReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	userID := jwt.GetUserIDByCtx(ctx)

	tx := dal.GetQuery().UserTeam
	_, err := tx.WithContext(ctx).Where(tx.UserID.Eq(jwt.GetUserIDByCtx(ctx)), tx.RoleID.In(consts.RoleTypeAdmin, consts.RoleTypeOwner)).First()
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	// 把用户信息加密
	rand.Seed(time.Now().UnixNano())
	rNum := rand.Intn(1000000)
	userInfo := fmt.Sprintf("%s_%d_%s_%d", req.TeamID, req.RoleID, jwt.GetUserIDByCtx(ctx), rNum)
	userInfoEncryptCode := encrypt.AesEncrypt(userInfo, conf.Conf.InviteData.AesSecretKey)

	// 给邀请链接设置过期时间
	k := fmt.Sprintf("invite:url:%s:%d:%s:%s", req.TeamID, req.RoleID, userID, userInfoEncryptCode)
	_, err = dal.GetRDB().Set(ctx, k, fmt.Sprintf("%s", jwt.GetUserIDByCtx(ctx)), 24*time.Hour).Result()
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrRedisFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, &rao.GetInviteMemberURLResp{
		URL:     fmt.Sprintf("%s%s#/invitatePage?invite_verify_code=%s", "邀请您加入RunnerGo团队", conf.Conf.Base.Domain, userInfoEncryptCode),
		Expired: time.Now().Add(time.Hour * 24).Unix(),
	})
	return
}

// CheckInviteMemberURL 检查邀请链接
func CheckInviteMemberURL(ctx *gin.Context) {
	var req rao.CheckInviteMemberURLReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	k := fmt.Sprintf("invite:url:%s:%d", req.TeamID, req.RoleID)
	inviteUserID, err := dal.GetRDB().Get(ctx, k).Result()
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrURLExpired, err.Error())
		return
	}
	if inviteUserID == "" {
		response.ErrorWithMsg(ctx, errno.ErrURLExpired, "")
		return
	}

	tx := dal.GetQuery().UserTeam
	cnt, err := tx.WithContext(ctx).Where(tx.TeamID.Eq(req.TeamID), tx.UserID.Eq(jwt.GetUserIDByCtx(ctx))).Count()
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}
	if cnt > 0 {
		response.ErrorWithMsg(ctx, errno.ErrExistsTeam, "")
		return
	}

	err = tx.WithContext(ctx).Create(&model.UserTeam{
		UserID:       jwt.GetUserIDByCtx(ctx),
		TeamID:       req.TeamID,
		RoleID:       req.RoleID,
		InviteUserID: inviteUserID,
	})
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	sx := dal.GetQuery().Setting
	_, err = sx.WithContext(ctx).Where(sx.UserID.Eq(jwt.GetUserIDByCtx(ctx))).UpdateColumn(sx.TeamID, req.TeamID)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}

// InviteMember 邀请成员
func InviteMember(ctx *gin.Context) {
	var req rao.InviteMemberReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	resp, err := team.InviteMember(ctx, jwt.GetUserIDByCtx(ctx), req.TeamID, req.Members)
	if err != nil {
		if err.Error() == "请配置邮件相关环境变量" {
			response.ErrorWithMsg(ctx, errno.ErrNotEmailConfig, err.Error())
		} else {
			response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		}
		return
	}

	response.SuccessWithData(ctx, resp)
	return
}

// SetUserTeamRole 设置用户团队角色
func SetUserTeamRole(ctx *gin.Context) {
	var req rao.RoleUserReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	if err := team.SetTeamRole(ctx, req.TeamID, req.UserID, req.RoleID); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}

// RemoveMember 移除成员
func RemoveMember(ctx *gin.Context) {
	var req rao.RemoveMemberReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	if err := team.RemoveMember(ctx, req.TeamID, jwt.GetUserIDByCtx(ctx), req.MemberID); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}

// QuitTeam 退出团队
func QuitTeam(ctx *gin.Context) {
	var req rao.QuitTeamReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	teamID, err := team.QuitTeam(ctx, req.TeamID, jwt.GetUserIDByCtx(ctx))
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.QuitTeamResp{TeamID: teamID})
	return
}

// DisbandTeam 解散团队
func DisbandTeam(ctx *gin.Context) {
	var req rao.DisbandTeamReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	teamID, err := team.DisbandTeam(ctx, req.TeamID, jwt.GetUserIDByCtx(ctx))
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.DisbandTeamResp{TeamID: teamID})
	return
}

func TransferTeam(ctx *gin.Context) {
	var req rao.TransferTeamReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
}

// GetInviteUserInfo 获取邀请链接人的用户信息
func GetInviteUserInfo(ctx *gin.Context) {
	var req rao.GetInviteUserInfoReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	userInfo, err := team.GetInviteUserInfo(ctx, &req)
	if err != nil {
		if err.Error() == "团队不存在或已解散" {
			response.ErrorWithMsg(ctx, errno.ErrTeamNotExist, err.Error())
		} else if err.Error() == "邀请链接已过期" {
			response.ErrorWithMsg(ctx, errno.ErrURLExpired, err.Error())
		} else {
			response.ErrorWithMsg(ctx, errno.ErrInviteCodeFailed, err.Error())
		}
		return
	}

	response.SuccessWithData(ctx, userInfo)
	return
}

// InviteLogin 邀请登录接口
func InviteLogin(ctx *gin.Context) {
	var req rao.InviteLoginReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	err := team.InviteLogin(ctx, req.InviteVerifyCode, jwt.GetUserIDByCtx(ctx))
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrInviteCodeFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}

func GetInviteEmailIsExist(ctx *gin.Context) {
	var req rao.GetInviteEmailIsExistReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	resBool, err := team.GetInviteEmailIsExist(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())

		return
	}
	response.SuccessWithData(ctx, rao.GetInviteEmailIsExistResp{
		EmailIsExist: resBool,
	})
	return
}
