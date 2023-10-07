package uiScene

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/uuid"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/query"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/errmsg"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/packer"
	"github.com/gin-gonic/gin"
)

func FolderReqSave(ctx *gin.Context, userID string, req *rao.UISceneSaveFolderReq) error {
	req.SceneID = uuid.GetUUID()
	scene := packer.TransSaveUISceneFolderReqToUISceneModel(req, userID)

	// 名称不能存在
	us := query.Use(dal.DB()).UIScene
	if _, err := us.WithContext(ctx).Where(
		us.TeamID.Eq(req.TeamID),
		us.Name.Eq(req.Name),
		us.SceneType.Eq(consts.UISceneTypeFolder),
		us.ParentID.Eq(req.ParentID),
		us.Status.Eq(consts.TargetStatusNormal),
		us.Source.Eq(req.Source),
		us.PlanID.Eq(req.PlanID),
	).First(); err == nil {
		return errmsg.ErrUISceneFolderNameRepeat
	}

	return query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		if err := tx.UIScene.WithContext(ctx).Create(scene); err != nil {
			return err
		}

		//if err := record.InsertCreate(ctx, scene.TeamID, userID, record.OperationOperateCreateUISceneFolder, scene.Name); err != nil {
		//	return err
		//}

		return nil
	})
}

func FolderReqUpdate(ctx *gin.Context, userID string, req *rao.UISceneSaveFolderReq) error {
	scene := packer.TransSaveUISceneFolderReqToUISceneModel(req, userID)

	// 名称不能存在
	us := query.Use(dal.DB()).UIScene
	if _, err := us.WithContext(ctx).Where(
		us.TeamID.Eq(req.TeamID),
		us.Name.Eq(req.Name),
		us.SceneType.Eq(consts.UISceneTypeFolder),
		us.ParentID.Eq(req.ParentID),
		us.SceneID.Neq(req.SceneID),
		us.Status.Eq(consts.TargetStatusNormal),
		us.Source.Eq(req.Source),
		us.PlanID.Eq(req.PlanID),
	).First(); err == nil {
		return errmsg.ErrUISceneFolderNameRepeat
	}

	return query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		if _, err := tx.UIScene.WithContext(ctx).Where(tx.UIScene.SceneID.Eq(req.SceneID)).Updates(scene); err != nil {
			return err
		}

		//if err := record.InsertUpdate(ctx, scene.TeamID, userID, record.OperationOperateUpdateUISceneFolder, scene.Name); err != nil {
		//	return err
		//}

		return nil
	})
}

func GetBySceneID(ctx *gin.Context, teamID string, sceneID string) (*rao.UISceneFolder, error) {
	us := query.Use(dal.DB()).UIScene
	scene, err := us.WithContext(ctx).Where(
		us.SceneID.Eq(sceneID),
		us.TeamID.Eq(teamID),
		us.SceneType.Eq(consts.UISceneTypeFolder),
		us.Status.Eq(consts.UISceneStatusNormal),
	).First()

	if err != nil {
		return nil, err
	}

	return packer.TransUISceneToRaoMockUISceneFolder(scene), nil
}
