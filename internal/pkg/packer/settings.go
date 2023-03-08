package packer

import (
	"kp-management/internal/pkg/dal/model"
	"kp-management/internal/pkg/dal/rao"
)

func TransUserSettingsToRaoUserSettings(s *model.Setting, utInfo *model.UserTeam, userInfo *model.User) *rao.GetUserSettingsResp {
	return &rao.GetUserSettingsResp{
		UserSettings: &rao.UserSettings{
			CurrentTeamID: s.TeamID,
		},
		UserInfo: &rao.UserInfo{
			ID:       userInfo.ID,
			Email:    userInfo.Email,
			Mobile:   userInfo.Mobile,
			Nickname: userInfo.Nickname,
			Avatar:   userInfo.Avatar,
			RoleID:   utInfo.RoleID,
		},
	}
}
