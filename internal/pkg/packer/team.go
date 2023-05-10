package packer

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
)

type TeamMemberCount struct {
	TeamID string
	Cnt    int64
}

func TransTeamsModelToRaoTeam(teams []*model.Team, userTeams []*model.UserTeam, teamCnt []*TeamMemberCount, users []*model.User) []*rao.Team {
	ret := make([]*rao.Team, 0)

	memo := make(map[string]*model.UserTeam)
	for _, team := range userTeams {
		memo[team.TeamID] = team
	}

	cntMemo := make(map[string]int64)
	for _, count := range teamCnt {
		cntMemo[count.TeamID] = count.Cnt
	}

	userMemo := make(map[string]*model.User)
	for _, user := range users {
		userMemo[user.UserID] = user
	}

	teamMemo := make(map[string]*model.Team)
	for _, t := range teams {
		teamMemo[t.TeamID] = t
	}

	for _, userTeamInfo := range userTeams {
		ret = append(ret, &rao.Team{
			Name:            teamMemo[userTeamInfo.TeamID].Name,
			Type:            teamMemo[userTeamInfo.TeamID].Type,
			Sort:            userTeamInfo.Sort,
			TeamID:          userTeamInfo.TeamID,
			RoleID:          userTeamInfo.RoleID,
			CreatedUserID:   teamMemo[userTeamInfo.TeamID].CreatedUserID,
			CreatedUserName: userMemo[teamMemo[userTeamInfo.TeamID].CreatedUserID].Nickname,
			CreatedTimeSec:  teamMemo[userTeamInfo.TeamID].CreatedAt.Unix(),
			Cnt:             cntMemo[userTeamInfo.TeamID],
		})
	}

	return ret
}
