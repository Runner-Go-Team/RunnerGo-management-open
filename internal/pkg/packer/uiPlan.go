package packer

import (
	"encoding/json"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/log"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/go-omnibus/proof"
	"strings"
)

func TransSaveReqToUIPlanModel(plan *rao.UIPlanSaveReq, userID string) *model.UIPlan {
	browsers, err := json.Marshal(plan.Browsers)
	if err != nil {
		log.Logger.Error("TransSaveReqToUIPlanModel.Browsers marshal err", proof.WithError(err))
	}

	var description string
	if plan.Description != nil {
		description = *plan.Description
	}
	return &model.UIPlan{
		PlanID:       plan.PlanID,
		TeamID:       plan.TeamID,
		RankID:       plan.RankID,
		Name:         plan.Name,
		TaskType:     plan.TaskType,
		CreateUserID: userID,
		HeadUserID:   strings.Join(plan.HeadUserIDs, ","),
		InitStrategy: plan.InitStrategy,
		Description:  description,
		Browsers:     string(browsers),
		UIMachineKey: plan.UIMachineKey,
	}
}

func TransUIPlanToRaoPlanList(plans []*model.UIPlan, users []*model.User, taskMap map[string]int32) []*rao.UIPlan {
	ret := make([]*rao.UIPlan, 0)

	memo := make(map[string]*model.User)
	for _, user := range users {
		memo[user.UserID] = user
	}

	for _, p := range plans {
		var timedStatus int32 = 0
		if p.TaskType == consts.UIPlanTaskTypeCronjob {
			if status, ok := taskMap[p.PlanID]; ok {
				timedStatus = status
			}
		}

		ret = append(ret, &rao.UIPlan{
			PlanID:            p.PlanID,
			RankID:            p.RankID,
			TeamID:            p.TeamID,
			Name:              p.Name,
			TaskType:          p.TaskType,
			CreatedUserID:     p.CreateUserID,
			CreatedUserName:   memo[p.CreateUserID].Nickname,
			CreatedUserAvatar: memo[p.CreateUserID].Avatar,
			HeadUserIDs:       strings.Split(p.HeadUserID, ","),
			Description:       p.Description,
			CreatedTimeSec:    p.CreatedAt.Unix(),
			UpdatedTimeSec:    p.UpdatedAt.Unix(),
			TimedStatus:       timedStatus,
		})
	}
	return ret
}

func TransUIPlanToRaoPlan(p *model.UIPlan, users []*model.User) *rao.UIPlan {
	memo := make(map[string]*model.User)
	for _, user := range users {
		memo[user.UserID] = user
	}

	browsers := make([]*rao.Browser, 0)
	_ = json.Unmarshal([]byte(p.Browsers), &browsers)

	return &rao.UIPlan{
		PlanID:            p.PlanID,
		RankID:            p.RankID,
		TeamID:            p.TeamID,
		Name:              p.Name,
		TaskType:          p.TaskType,
		CreatedUserID:     p.CreateUserID,
		CreatedUserName:   memo[p.CreateUserID].Nickname,
		CreatedUserAvatar: memo[p.CreateUserID].Avatar,
		HeadUserIDs:       strings.Split(p.HeadUserID, ","),
		Description:       p.Description,
		InitStrategy:      p.InitStrategy,
		CreatedTimeSec:    p.CreatedAt.Unix(),
		UpdatedTimeSec:    p.UpdatedAt.Unix(),
		Browsers:          browsers,
		UIMachineKey:      p.UIMachineKey,
	}
}
