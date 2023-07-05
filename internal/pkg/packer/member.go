package packer

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
)

func TransUsersToRaoMembers(users []*model.User, userTeams []*model.UserTeam, urList []*model.UserRole,
	roleList []*model.Role) []*rao.Member {
	ret := make([]*rao.Member, 0)

	memo := make(map[string]string)
	for _, u := range users {
		memo[u.UserID] = u.Nickname
	}

	roleMap := make(map[string]string)
	for _, roleInfo := range roleList {
		roleMap[roleInfo.RoleID] = roleInfo.Name
	}

	urMap := make(map[string]string)
	for _, v := range urList {
		urMap[v.TeamID+v.UserID] = v.RoleID
	}

	for _, ut := range userTeams {
		for _, u := range users {
			if ut.UserID == u.UserID {
				ret = append(ret, &rao.Member{
					Avatar:         u.Avatar,
					Account:        u.Account,
					UserID:         u.UserID,
					Email:          u.Email,
					Nickname:       u.Nickname,
					JoinTimeSec:    ut.CreatedAt.Unix(),
					RoleID:         ut.RoleID,
					TeamRoleName:   roleMap[urMap[ut.TeamID+ut.UserID]],
					InviteUserID:   ut.InviteUserID,
					InviteUserName: memo[ut.InviteUserID],
				})
			}
		}
	}
	return ret
}
