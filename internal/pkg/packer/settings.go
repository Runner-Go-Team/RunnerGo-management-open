package packer

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
)

func TransUserSettingsToRaoUserSettings(
	s *model.Setting,
	userInfo *model.User,
	teamRole *model.UserRole,
	companyRole *model.UserRole,
	roles []*model.Role,
) *rao.GetUserSettingsResp {
	rolesMemo := make(map[string]*model.Role)
	for _, role := range roles {
		rolesMemo[role.RoleID] = role
	}

	res := &rao.GetUserSettingsResp{
		UserSettings: &rao.UserSettings{
			CurrentTeamID: s.TeamID,
		},
		UserInfo: &rao.UserInfo{
			ID:       userInfo.ID,
			Email:    userInfo.Email,
			Mobile:   userInfo.Mobile,
			Nickname: userInfo.Nickname,
			Avatar:   userInfo.Avatar,
			Account:  userInfo.Account,
			UserID:   userInfo.UserID,
		},
	}
	if r, ok := rolesMemo[teamRole.RoleID]; ok {
		res.UserInfo.RoleID = r.RoleID
		res.UserInfo.RoleName = r.Name
	}

	if r, ok := rolesMemo[companyRole.RoleID]; ok {
		res.UserInfo.CompanyRoleID = r.RoleID
		res.UserInfo.CompanyRoleName = r.Name
	}

	return res
}
