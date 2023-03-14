package packer

import (
	"encoding/json"
	"kp-management/internal/pkg/biz/consts"
	"kp-management/internal/pkg/dal/mao"
	"kp-management/internal/pkg/dal/model"
	"kp-management/internal/pkg/dal/rao"
	"time"
)

func TransPlansToRaoPlanList(plans []*model.StressPlan, users []*model.User) []*rao.StressPlan {
	ret := make([]*rao.StressPlan, 0)

	memo := make(map[string]*model.User)
	for _, user := range users {
		memo[user.UserID] = user
	}

	for _, p := range plans {
		ret = append(ret, &rao.StressPlan{
			ID:                p.ID,
			PlanID:            p.PlanID,
			RankID:            p.RankID,
			TeamID:            p.TeamID,
			PlanName:          p.PlanName,
			TaskType:          p.TaskType,
			TaskMode:          p.TaskMode,
			Status:            p.Status,
			CreatedUserName:   memo[p.CreateUserID].Nickname,
			CreatedUserAvatar: memo[p.CreateUserID].Avatar,
			CreatedUserID:     p.CreateUserID,
			Remark:            p.Remark,
			CreatedTimeSec:    p.CreatedAt.Unix(),
			UpdatedTimeSec:    p.UpdatedAt.Unix(),
		})
	}
	return ret
}

func TransTaskToRaoPlan(p *model.StressPlan, t rao.ModeConf, u *model.User) *rao.StressPlan {
	mc := rao.ModeConf{
		ReheatTime:       t.ReheatTime,
		RoundNum:         t.RoundNum,
		Concurrency:      t.Concurrency,
		ThresholdValue:   t.ThresholdValue,
		StartConcurrency: t.StartConcurrency,
		Step:             t.Step,
		StepRunTime:      t.StepRunTime,
		MaxConcurrency:   t.MaxConcurrency,
		Duration:         t.Duration,
	}

	return &rao.StressPlan{
		PlanID:            p.PlanID,
		TeamID:            p.TeamID,
		PlanName:          p.PlanName,
		TaskType:          p.TaskType,
		TaskMode:          p.TaskMode,
		Status:            p.Status,
		CreatedUserID:     p.CreateUserID,
		CreatedUserAvatar: u.Avatar,
		CreatedUserName:   u.Nickname,
		Remark:            p.Remark,
		CreatedTimeSec:    p.CreatedAt.Unix(),
		UpdatedTimeSec:    p.UpdatedAt.Unix(),
		ModeConf:          &mc,
	}
}

func TransSaveTimingTaskConfigReqToModelData(req *rao.SavePlanConfReq, userID string) (*model.StressPlanTimedTaskConf, error) {
	// 把mode_conf压缩成字符串
	modeConfString, err := json.Marshal(req.ModeConf)
	if err != nil {
		return nil, err
	}
	return &model.StressPlanTimedTaskConf{
		PlanID:        req.PlanID,
		SceneID:       req.SceneID,
		TeamID:        req.TeamID,
		UserID:        userID,
		Frequency:     req.TimedTaskConf.Frequency,
		TaskExecTime:  req.TimedTaskConf.TaskExecTime,
		TaskCloseTime: req.TimedTaskConf.TaskCloseTime,
		TaskType:      req.TaskType,
		TaskMode:      req.Mode,
		ControlMode:   req.ControlMode,
		ModeConf:      string(modeConfString),
		Status:        consts.TimedTaskWaitEnable,
	}, nil
}

func TransChangeReportConfRunToMao(req rao.ChangeTaskConfReq) *mao.ChangeTaskConf {
	return &mao.ChangeTaskConf{
		ReportID: req.ReportID,
		TeamID:   req.TeamID,
		PlanID:   req.PlanID,
		ModeConf: &mao.ModeConf{
			ReheatTime:       req.ModeConf.ReheatTime,
			RoundNum:         req.ModeConf.RoundNum,
			Concurrency:      req.ModeConf.Concurrency,
			ThresholdValue:   req.ModeConf.ThresholdValue,
			StartConcurrency: req.ModeConf.StartConcurrency,
			Step:             req.ModeConf.Step,
			StepRunTime:      req.ModeConf.StepRunTime,
			MaxConcurrency:   req.ModeConf.MaxConcurrency,
			Duration:         req.ModeConf.Duration,
			CreatedTimeSec:   time.Now().Unix(),
		},
	}
}
