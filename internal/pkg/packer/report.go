package packer

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
)

func TransReportModelToRaoReportList(reports []*model.StressPlanReport, users []*model.User) []*rao.StressPlanReport {
	ret := make([]*rao.StressPlanReport, 0)

	memo := make(map[string]*model.User)
	for _, user := range users {
		memo[user.UserID] = user
	}

	for _, r := range reports {
		ret = append(ret, &rao.StressPlanReport{
			ReportID:    r.ReportID,
			ReportName:  r.ReportName,
			RankID:      r.RankID,
			TaskType:    r.TaskType,
			TaskMode:    r.TaskMode,
			Status:      r.Status,
			RunTimeSec:  r.CreatedAt.Unix(),
			LastTimeSec: r.UpdatedAt.Unix(),
			RunUserID:   r.RunUserID,
			RunUserName: memo[r.RunUserID].Nickname,
			TeamID:      r.TeamID,
			PlanID:      r.PlanID,
			PlanName:    r.PlanName,
			SceneID:     r.SceneID,
			SceneName:   r.SceneName,
		})
	}
	return ret
}
