package packer

import (
	"kp-management/internal/pkg/dal/model"
	"kp-management/internal/pkg/dal/rao"
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
		})
	}
	return ret
}
