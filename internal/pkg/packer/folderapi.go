package packer

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
)

func TransTargetToRaoFolderAPIList(targets []*model.Target) []*rao.FolderAPI {
	ret := make([]*rao.FolderAPI, 0)
	for _, t := range targets {
		ret = append(ret, &rao.FolderAPI{
			TargetID:      t.TargetID,
			TeamID:        t.TeamID,
			TargetType:    t.TargetType,
			Name:          t.Name,
			ParentID:      t.ParentID,
			Method:        t.Method,
			Sort:          t.Sort,
			TypeSort:      t.TypeSort,
			Version:       t.Version,
			Source:        t.Source,
			CreatedUserID: t.CreatedUserID,
			RecentUserID:  t.RecentUserID,
		})
	}
	return ret
}

func TransTargetToRaoTrashFolderAPIList(targets []*model.Target, apiIDUrlMap map[string]string) []*rao.FolderAPI {
	ret := make([]*rao.FolderAPI, 0)
	for _, t := range targets {
		ret = append(ret, &rao.FolderAPI{
			TargetID:      t.TargetID,
			TeamID:        t.TeamID,
			TargetType:    t.TargetType,
			Name:          t.Name,
			Url:           apiIDUrlMap[t.TargetID],
			ParentID:      t.ParentID,
			Method:        t.Method,
			Sort:          t.Sort,
			TypeSort:      t.TypeSort,
			Version:       t.Version,
			CreatedUserID: t.CreatedUserID,
			RecentUserID:  t.RecentUserID,
		})
	}
	return ret
}
