package packer

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
)

func TransSaveFolderReqToTargetModel(folder *rao.SaveFolderReq, userID string) *model.Target {
	return &model.Target{
		TargetID:      folder.TargetID,
		TeamID:        folder.TeamID,
		TargetType:    consts.TargetTypeFolder,
		Name:          folder.Name,
		ParentID:      folder.ParentID,
		Method:        folder.Method,
		Sort:          folder.Sort,
		TypeSort:      folder.TypeSort,
		Status:        consts.TargetStatusNormal,
		Version:       folder.Version,
		CreatedUserID: userID,
		RecentUserID:  userID,
		Source:        folder.Source,
		Description:   folder.Description,
	}
}

func TransSaveTargetReqToTargetModel(target *rao.SaveTargetReq, userID string) *model.Target {
	return &model.Target{
		TargetID:      target.TargetID,
		TeamID:        target.TeamID,
		TargetType:    target.TargetType,
		Name:          target.Name,
		ParentID:      target.ParentID,
		Method:        target.Method,
		Sort:          target.Sort,
		TypeSort:      target.TypeSort,
		Status:        consts.TargetStatusNormal,
		Version:       target.Version,
		CreatedUserID: userID,
		RecentUserID:  userID,
		Source:        target.Source,
		Description:   target.Description,
		SourceID:      target.SourceID,
	}
}

func TransSaveImportFolderReqToTargetModel(folder rao.SaveTargetReq, teamID string, userID string) *model.Target {
	return &model.Target{
		TargetID:      folder.OldTargetID,
		TeamID:        teamID,
		TargetType:    consts.TargetTypeFolder,
		Name:          folder.Name,
		ParentID:      folder.OldParentID,
		Status:        consts.TargetStatusNormal,
		CreatedUserID: userID,
		RecentUserID:  userID,
		Source:        consts.TargetSourceApi,
		Description:   folder.Description,
	}
}

func TransSaveImportTargetReqToTargetModel(target rao.SaveTargetReq, teamID string, userID string) *model.Target {
	return &model.Target{
		TargetID:      target.OldTargetID,
		TeamID:        teamID,
		TargetType:    consts.TargetTypeAPI,
		Name:          target.Name,
		ParentID:      target.OldParentID,
		Method:        target.Method,
		Status:        consts.TargetStatusNormal,
		CreatedUserID: userID,
		RecentUserID:  userID,
		Source:        consts.TargetSourceApi,
		Description:   target.Request.Description,
	}
}

func TransSaveGroupReqToTargetModel(group *rao.SaveGroupReq, userID string) *model.Target {
	return &model.Target{
		TargetID:      group.TargetID,
		TeamID:        group.TeamID,
		TargetType:    consts.TargetTypeFolder,
		Name:          group.Name,
		ParentID:      group.ParentID,
		Method:        group.Method,
		Sort:          group.Sort,
		TypeSort:      group.TypeSort,
		Status:        consts.TargetStatusNormal,
		Version:       group.Version,
		CreatedUserID: userID,
		RecentUserID:  userID,
		Source:        group.Source,
		PlanID:        group.PlanID,
		Description:   group.Description,
	}
}

func TransSaveSceneReqToTargetModel(scene *rao.SaveSceneReq, userID string) *model.Target {
	return &model.Target{
		TargetID:      scene.TargetID,
		TeamID:        scene.TeamID,
		TargetType:    consts.TargetTypeScene,
		Name:          scene.Name,
		ParentID:      scene.ParentID,
		Method:        scene.Method,
		Sort:          scene.Sort,
		TypeSort:      scene.TypeSort,
		Status:        consts.TargetStatusNormal,
		Version:       scene.Version,
		CreatedUserID: userID,
		RecentUserID:  userID,
		Source:        scene.Source,
		PlanID:        scene.PlanID,
		Description:   scene.Description,
	}
}
