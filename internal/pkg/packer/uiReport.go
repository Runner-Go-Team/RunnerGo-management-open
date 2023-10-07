package packer

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
)

func TransUIReportModelToRaoReportList(reports []*model.UIPlanReport, users []*model.User) []*rao.UIPlanReport {
	ret := make([]*rao.UIPlanReport, 0)

	memo := make(map[string]*model.User)
	for _, user := range users {
		memo[user.UserID] = user
	}

	for _, r := range reports {
		ret = append(ret, &rao.UIPlanReport{
			ReportID:      r.ReportID,
			ReportName:    r.ReportName,
			RankID:        r.RankID,
			TaskType:      r.TaskType,
			Status:        r.Status,
			RunTimeSec:    r.CreatedAt.Unix(),
			LastTimeSec:   r.UpdatedAt.Unix(),
			RunUserID:     r.RunUserID,
			RunUserName:   memo[r.RunUserID].Nickname,
			SceneRunOrder: r.SceneRunOrder,
			TeamID:        r.TeamID,
			PlanID:        r.PlanID,
			PlanName:      r.PlanName,
		})
	}
	return ret
}
