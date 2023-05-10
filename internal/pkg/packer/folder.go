package packer

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/mao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
)

func TransSaveFolderReqToMaoFolder(folder *rao.SaveFolderReq) *mao.Folder {
	return &mao.Folder{
		TargetID: folder.TargetID,
	}
}

func TransTargetToRaoFolder(t *model.Target) *rao.Folder {
	return &rao.Folder{
		TargetID:    t.TargetID,
		TeamID:      t.TeamID,
		ParentID:    t.ParentID,
		Name:        t.Name,
		Method:      t.Method,
		Sort:        t.Sort,
		TypeSort:    t.TypeSort,
		Version:     t.Version,
		Description: t.Description,
		//Request:  &r,
		//Script:   &s,
	}
}
