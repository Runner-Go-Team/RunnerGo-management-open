package packer

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
)

func TransTargetsToRaoGroupSceneList(targets []*model.Target) []*rao.GroupScene {
	ret := make([]*rao.GroupScene, 0)
	for _, t := range targets {
		ret = append(ret, &rao.GroupScene{
			TargetID:      t.TargetID,
			TeamID:        t.TeamID,
			TargetType:    t.TargetType,
			Name:          t.Name,
			ParentID:      t.ParentID,
			Method:        t.Method,
			Sort:          t.Sort,
			TypeSort:      t.TypeSort,
			Version:       t.Version,
			CreatedUserID: t.CreatedUserID,
			RecentUserID:  t.RecentUserID,
			Description:   t.Description,
			Source:        t.Source,
			IsDisabled:    t.IsDisabled,
		})
	}
	return ret
}
