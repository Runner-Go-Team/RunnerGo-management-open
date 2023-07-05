package packer

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
)

func TransTargetToRaoScene(targets []*model.Target) []*rao.Scene {
	ret := make([]*rao.Scene, 0)
	for _, t := range targets {
		ret = append(ret, &rao.Scene{
			TeamID:      t.TeamID,
			TargetID:    t.TargetID,
			ParentID:    t.ParentID,
			Name:        t.Name,
			Method:      t.Method,
			Sort:        t.Sort,
			TypeSort:    t.TypeSort,
			Version:     t.Version,
			Description: t.Description,
		})
	}
	return ret
}
