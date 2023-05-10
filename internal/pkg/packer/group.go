package packer

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/mao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
)

func TransSaveGroupReqToMaoGroup(group *rao.SaveGroupReq) *mao.Group {
	return &mao.Group{
		TargetID: group.TargetID,
	}
}

func TransTargetToRaoGroup(t *model.Target, g *mao.Group) *rao.Group {
	return &rao.Group{
		TeamID:      t.TeamID,
		TargetID:    t.TargetID,
		ParentID:    t.ParentID,
		Name:        t.Name,
		Method:      t.Method,
		Sort:        t.Sort,
		TypeSort:    t.TypeSort,
		Version:     t.Version,
		Source:      t.Source,
		PlanID:      t.PlanID,
		Description: t.Description,
	}
}
