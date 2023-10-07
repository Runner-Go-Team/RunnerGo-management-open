package uiPlan

import (
	"encoding/json"
	"errors"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/log"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/uuid"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/query"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/errmsg"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/uiScene"
	"github.com/gin-gonic/gin"
	"github.com/go-omnibus/proof"
	"gorm.io/gorm"
)

func ImportScene(ctx *gin.Context, userID string, req *rao.UIPlanImportSceneReq) error {
	sceneIDs := req.SceneIDs
	teamID := req.TeamID
	if len(sceneIDs) <= 0 {
		return nil
	}
	mt := query.Use(dal.DB()).UIScene
	scenes, err := mt.WithContext(ctx).Where(
		mt.TeamID.Eq(teamID),
		mt.SceneID.In(sceneIDs...),
	).Order(mt.Sort, mt.CreatedAt.Desc()).Find()
	if err != nil {
		return err
	}

	newSceneIDs := make(map[string]string, len(scenes))
	newSceneIDs["0"] = "0"
	for _, s := range scenes {
		newSceneIDs[s.SceneID] = uuid.GetUUID()
	}

	var syncScenes = make([]*model.UISceneSync, 0)
	return query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		t := tx.UIScene

		for _, mtv := range scenes {
			browsers, err := json.Marshal(mtv.Browsers)
			if err != nil {
				log.Logger.Error("ImportScene.Browsers marshal err", proof.WithError(err))
			}
			scene := &model.UIScene{
				SceneID:       newSceneIDs[mtv.SceneID],
				SceneType:     mtv.SceneType,
				TeamID:        mtv.TeamID,
				Name:          mtv.Name,
				ParentID:      newSceneIDs[mtv.ParentID],
				Sort:          mtv.Sort,
				Status:        mtv.Status,
				Version:       mtv.Version,
				Source:        consts.UISceneSourcePlan,
				CreatedUserID: userID,
				RecentUserID:  mtv.RecentUserID,
				Description:   mtv.Description,
				SourceID:      mtv.SceneID,
				Browsers:      string(browsers),
				PlanID:        req.PlanID,
			}

			if mtv.SceneType == consts.UISceneTypeFolder {
				// 判断是否存在当前名称
				targetInfo, err := t.WithContext(ctx).Where(
					t.TeamID.Eq(scene.TeamID),
					t.Name.Eq(scene.Name),
					t.SceneID.Neq(scene.SceneID),
					t.SceneType.Eq(consts.UISceneTypeFolder),
					t.Status.Eq(consts.UISceneStatusNormal),
					t.Source.Eq(consts.UISceneSourcePlan),
					t.PlanID.Eq(req.PlanID),
					t.ParentID.Eq(scene.ParentID),
				).First()
				if errors.Is(err, gorm.ErrRecordNotFound) {
					if err := t.WithContext(ctx).Create(scene); err != nil {
						return err
					}
				}
				if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
					return err
				}

				if targetInfo != nil {
					newSceneIDs[mtv.SceneID] = targetInfo.SceneID
					continue
				}

			}

			if mtv.SceneType == consts.UISceneTypeScene {
				// 判断是否存在当前名称
				targetInfo, err := t.WithContext(ctx).Where(
					t.TeamID.Eq(scene.TeamID),
					t.Name.Eq(scene.Name),
					t.SceneID.Neq(scene.SceneID),
					t.SceneType.Eq(consts.UISceneTypeScene),
					t.Status.Eq(consts.UISceneStatusNormal),
					t.Source.Eq(consts.UISceneSourcePlan),
					t.PlanID.Eq(req.PlanID),
					t.ParentID.Eq(scene.ParentID),
				).First()
				if errors.Is(err, gorm.ErrRecordNotFound) {
					if err := t.WithContext(ctx).Create(scene); err != nil {
						return err
					}
				}
				if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
					return err
				}
				if targetInfo != nil {
					return errmsg.ErrUISceneNameRepeat
				}

				// 同步场景下的步骤
				operators, err := tx.UISceneOperator.WithContext(ctx).Where(
					tx.UISceneOperator.SceneID.Eq(mtv.SceneID),
					tx.UISceneOperator.ParentID.Eq("0"),
				).Find()
				if err != nil {
					return err
				}
				for _, operator := range operators {
					if err := uiScene.HandleCopyOperator(ctx, userID, mtv, operator, scene, "0"); err != nil {
						return err
					}
				}
				planScene := &model.UISceneSync{
					SceneID:       scene.SceneID,
					SourceSceneID: mtv.SceneID,
					TeamID:        teamID,
					SyncMode:      req.SyncMode,
				}
				syncScenes = append(syncScenes, planScene)
			}
		}

		// 添加场景同步方式
		if len(syncScenes) > 0 {
			if err = tx.UISceneSync.WithContext(ctx).Create(syncScenes...); err != nil {
				return err
			}
		}

		return nil
	})
}

// SetSceneSyncMode 设置场景同步方式
func SetSceneSyncMode(ctx *gin.Context, userID string, req *rao.UIPlanSetSceneSyncModeReq) error {
	ss := query.Use(dal.DB()).UISceneSync
	scene, err := ss.WithContext(ctx).Where(
		ss.TeamID.Eq(req.TeamID),
		ss.SceneID.Eq(req.SceneID),
	).First()
	if err != nil {
		return err
	}

	// 实时同步 源场景
	// 实时同步 源计划
	// 手动同步 源场景
	// 手动同步 源计划
	if _, err = ss.WithContext(ctx).Where(
		ss.TeamID.Eq(req.TeamID),
		ss.SceneID.Eq(req.SceneID)).Update(ss.SyncMode, req.SyncMode); err != nil {
		return err
	}

	// 设置同步状态  自动需立马同步
	if req.SyncMode == consts.UIPlanSceneSyncModeAuto {
		if req.TargetSource == consts.UISceneSource {
			if err = uiScene.SyncScene(ctx, scene.SourceSceneID, scene.SceneID, req.TeamID, userID); err != nil {
				return err
			}
		}
		if req.TargetSource == consts.UISceneSourcePlan {
			if err = uiScene.SyncScene(ctx, scene.SceneID, scene.SourceSceneID, req.TeamID, userID); err != nil {
				return err
			}
		}
	}

	return nil
}

// HandSyncLastData 手动同步最新数据
func HandSyncLastData(ctx *gin.Context, userID string, req *rao.UIPlanSetSceneHandSyncReq) error {
	ss := query.Use(dal.DB()).UISceneSync
	scene, err := ss.WithContext(ctx).Where(
		ss.TeamID.Eq(req.TeamID),
		ss.SceneID.Eq(req.SceneID),
	).First()
	if err != nil {
		return err
	}

	if scene.SyncMode == consts.UIPlanSceneSyncModeAuto {
		return nil
	}

	if scene.SyncMode == consts.UIPlanSceneSyncModeHandTargetScene {
		if err = uiScene.SyncScene(ctx, scene.SourceSceneID, scene.SceneID, req.TeamID, userID); err != nil {
			return err
		}
	}
	if scene.SyncMode == consts.UIPlanSceneSyncModeHandTargetPlan {
		if err = uiScene.SyncScene(ctx, scene.SceneID, scene.SourceSceneID, req.TeamID, userID); err != nil {
			return err
		}
	}

	return nil
}
