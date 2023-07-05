package packer

import (
	"encoding/json"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/mao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
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

	// 把mode_conf压缩成字符串
	machineDispatchModeConfString, err := json.Marshal(req.MachineDispatchModeConf)
	if err != nil {
		return nil, err
	}

	return &model.StressPlanTimedTaskConf{
		PlanID:                  req.PlanID,
		SceneID:                 req.SceneID,
		TeamID:                  req.TeamID,
		UserID:                  userID,
		Frequency:               req.TimedTaskConf.Frequency,
		TaskExecTime:            req.TimedTaskConf.TaskExecTime,
		TaskCloseTime:           req.TimedTaskConf.TaskCloseTime,
		TaskType:                req.TaskType,
		TaskMode:                req.Mode,
		ControlMode:             req.ControlMode,
		DebugMode:               req.DebugMode,
		ModeConf:                string(modeConfString),
		IsOpenDistributed:       req.IsOpenDistributed,
		MachineDispatchModeConf: string(machineDispatchModeConfString),
		Status:                  consts.TimedTaskWaitEnable,
		RunUserID:               userID,
	}, nil
}

func TransChangeReportConfRunToMao(req rao.ChangeTaskConfReq) mao.ChangeTaskConf {
	usableMachineList := make([]mao.UsableMachineInfo, 0, len(req.MachineDispatchModeConf.UsableMachineList))
	for _, v := range req.MachineDispatchModeConf.UsableMachineList {
		temp := mao.UsableMachineInfo{
			MachineStatus:    v.MachineStatus,
			MachineName:      v.MachineName,
			Region:           v.Region,
			Ip:               v.Ip,
			Weight:           v.Weight,
			RoundNum:         v.RoundNum,
			Concurrency:      v.Concurrency,
			ThresholdValue:   v.ThresholdValue,
			StartConcurrency: v.StartConcurrency,
			Step:             v.Step,
			StepRunTime:      v.StepRunTime,
			MaxConcurrency:   v.MaxConcurrency,
			Duration:         v.Duration,
			CreatedTimeSec:   v.CreatedTimeSec,
		}
		usableMachineList = append(usableMachineList, temp)
	}

	return mao.ChangeTaskConf{
		ReportID: req.ReportID,
		TeamID:   req.TeamID,
		PlanID:   req.PlanID,
		ModeConf: &mao.ModeConf{
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
		IsOpenDistributed: req.IsOpenDistributed,
		MachineDispatchModeConf: mao.MachineDispatchModeConf{
			MachineAllotType:  req.MachineDispatchModeConf.MachineAllotType,
			UsableMachineList: usableMachineList,
		},
	}
}

func TransNewestPlansToRaoPlanList(plans []*model.StressPlan, users []*model.User) []rao.GetNewestStressPlanListResp {
	ret := make([]rao.GetNewestStressPlanListResp, 0)

	userMap := make(map[string]*model.User)
	for _, user := range users {
		userMap[user.UserID] = user
	}

	for _, planInfo := range plans {
		ret = append(ret, rao.GetNewestStressPlanListResp{
			TeamID:     planInfo.TeamID,
			PlanID:     planInfo.PlanID,
			PlanName:   planInfo.PlanName,
			PlanType:   "stress",
			Username:   userMap[planInfo.CreateUserID].Nickname,
			UserAvatar: userMap[planInfo.CreateUserID].Avatar,
			UpdatedAt:  planInfo.UpdatedAt.Unix(),
		})
	}
	return ret
}
