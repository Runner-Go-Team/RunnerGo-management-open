package packer

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
)

func TransUsersToRaoMembers(users []*model.User, userTeams []*model.UserTeam, urList []*model.UserRole,
	roleList []*model.Role, companyRoleList []*model.UserRole) []*rao.Member {
	ret := make([]*rao.Member, 0)

	memo := make(map[string]string)
	for _, u := range users {
		memo[u.UserID] = u.Nickname
	}

	roleMap := make(map[string]*model.Role)
	for _, roleInfo := range roleList {
		roleMap[roleInfo.RoleID] = roleInfo
	}

	urMap := make(map[string]string)
	for _, v := range urList {
		urMap[v.TeamID+v.UserID] = v.RoleID
	}

	companyRMap := make(map[string]string)
	for _, v := range companyRoleList {
		companyRMap[v.UserID] = v.RoleID
	}

	for _, ut := range userTeams {
		for _, u := range users {
			if ut.UserID == u.UserID {
				member := &rao.Member{
					Avatar:         u.Avatar,
					Account:        u.Account,
					UserID:         u.UserID,
					Email:          u.Email,
					Nickname:       u.Nickname,
					JoinTimeSec:    ut.CreatedAt.Unix(),
					RoleID:         ut.RoleID,
					TeamRoleName:   roleMap[urMap[ut.TeamID+ut.UserID]].Name,
					InviteUserID:   ut.InviteUserID,
					InviteUserName: memo[ut.InviteUserID],
				}
				if roleID, ok := companyRMap[ut.UserID]; ok {
					if role, ok := roleMap[roleID]; ok {
						member.CompanyRoleLevel = role.Level
					}
				}

				ret = append(ret, member)
			}
		}
	}
	return ret
}
