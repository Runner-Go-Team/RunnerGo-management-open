package permission

import (
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/jwt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/log"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/conf"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/goccy/go-json"
)

const (
	GetTeamCompanyMembersUri = "/permission/api/open/v1/team/company/members"
	TeamMemberSaveUri        = "/permission/api/open/v1/team/member/save"
	GetRoleMemberInfoUri     = "/permission/api/open/v1/role/member/info"
	UserGetMarksUri          = "/permission/api/open/v1/permission/user/get_marks"
	NoticeGroupListUri       = "/permission/api/open/v1/notice/group/list"
	NoticeThirdUsersUri      = "/permission/api/open/v1/notice/get_third_users"
)

func GetTeamCompanyMembers(ctx *gin.Context, req *rao.GetTeamCompanyMembersReq) (interface{}, error) {
	userID := jwt.GetUserIDByCtx(ctx)

	log.Logger.Info("权限接口--获取当前团队和企业成员关系接口--参数", req.TeamID, req.Keyword, req.Page, req.Size)

	response, err := resty.New().R().
		SetHeader("Content-Type", "application/json").
		Get(fmt.Sprintf("%s%s?team_id=%s&user_id=%s&keyword=%s&page=%d&size=%d",
			conf.Conf.Clients.Permission.PermissionDomain,
			GetTeamCompanyMembersUri,
			req.TeamID,
			userID,
			req.Keyword,
			req.Page,
			req.Size,
		))
	if err != nil {
		return "", err
	}

	resp := rao.SendPermissionApiResp{}
	err = json.Unmarshal(response.Body(), &resp)
	if err != nil {
		log.Logger.Info("权限接口--获取当前团队和企业成员关系接口--返回值解析失败：err：", err)
		return nil, err
	}
	return resp.Data, nil
}

func TeamMembersSave(ctx *gin.Context, req *rao.TeamMembersSaveReq) (interface{}, error) {
	userID := jwt.GetUserIDByCtx(ctx)

	log.Logger.Info("权限接口--添加团队成员--参数", req)
	req.UserID = userID
	bodyByte, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	response, err := resty.New().R().
		SetHeader("Content-Type", "application/json").
		SetBody(bodyByte).
		Post(conf.Conf.Clients.Permission.PermissionDomain + TeamMemberSaveUri)
	if err != nil {
		return "", err
	}

	resp := rao.SendPermissionApiResp{}
	err = json.Unmarshal(response.Body(), &resp)
	if err != nil {
		log.Logger.Info("权限接口--添加团队成员--返回值解析失败：err：", err)
		return nil, err
	}
	return resp.Data, nil
}

func GetRoleMemberInfo(ctx *gin.Context, req *rao.GetRoleMemberInfoReq) (interface{}, error) {
	userID := jwt.GetUserIDByCtx(ctx)

	uc := dal.GetQuery().UserCompany
	userCompanyInfo, err := uc.WithContext(ctx).Where(uc.UserID.Eq(userID)).First()
	if err != nil {
		return nil, err
	}

	log.Logger.Info("权限接口--获取我的角色信息--参数", req)
	response, err := resty.New().R().
		SetHeader("Content-Type", "application/json").
		Get(fmt.Sprintf(conf.Conf.Clients.Permission.PermissionDomain+GetRoleMemberInfoUri+"?user_id=%s&team_id=%s&company_id=%s&role_id=%s",
			userID, req.TeamID, userCompanyInfo.CompanyID, req.RoleID))
	if err != nil {
		return "", err
	}

	resp := rao.SendPermissionApiResp{}
	err = json.Unmarshal(response.Body(), &resp)
	if err != nil {
		log.Logger.Info("权限接口--获取我的角色信息--返回值解析失败：err：", err)
		return nil, err
	}
	return resp.Data, nil
}

func UserGetMarks(ctx *gin.Context) (interface{}, error) {
	userID := jwt.GetUserIDByCtx(ctx)
	log.Logger.Info("权限接口--获取用户的全部角色对应的mark--参数", userID)
	response, err := resty.New().R().
		SetHeader("Content-Type", "application/json").
		Get(conf.Conf.Clients.Permission.PermissionDomain + UserGetMarksUri + "?user_id=" + userID)
	if err != nil {
		return "", err
	}

	resp := rao.SendPermissionApiResp{}
	err = json.Unmarshal(response.Body(), &resp)
	if err != nil {
		log.Logger.Info("权限接口--获取用户的全部角色对应的mark--返回值解析失败：err：", err)
		return nil, err
	}
	return resp.Data, nil
}

func GetNoticeGroupList(ctx *gin.Context, req *rao.GetNoticeGroupListReq) (interface{}, error) {
	userID := jwt.GetUserIDByCtx(ctx)
	log.Logger.Info("权限接口--获取通知组列表--参数", userID)
	response, err := resty.New().R().
		SetHeader("Content-Type", "application/json").
		Get(fmt.Sprintf(conf.Conf.Clients.Permission.PermissionDomain+NoticeGroupListUri+"?user_id=%s&keyword=%s&channel_id=%d",
			userID, req.Keyword, req.ChannelID))
	if err != nil {
		return "", err
	}

	resp := rao.SendPermissionApiResp{}
	err = json.Unmarshal(response.Body(), &resp)
	if err != nil {
		log.Logger.Info("权限接口--获取通知组列表--返回值解析失败：err：", err)
		return nil, err
	}
	return resp.Data, nil
}

func GetNoticeThirdUsers(ctx *gin.Context, req *rao.GetNoticeThirdUsersReq) (interface{}, error) {
	userID := jwt.GetUserIDByCtx(ctx)
	log.Logger.Info("权限接口--获取三方通知--参数", userID)
	response, err := resty.New().R().
		SetHeader("Content-Type", "application/json").
		Get(fmt.Sprintf(conf.Conf.Clients.Permission.PermissionDomain+NoticeThirdUsersUri+"?user_id=%s&notice_id=%s",
			userID, req.NoticeID))
	if err != nil {
		return "", err
	}

	resp := rao.SendPermissionApiResp{}
	err = json.Unmarshal(response.Body(), &resp)
	if err != nil {
		log.Logger.Info("权限接口--获取通知组列表--返回值解析失败：err：", err)
		return nil, err
	}
	return resp.Data, nil
}
