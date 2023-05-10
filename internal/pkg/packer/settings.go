package packer

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
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
