package scene

import (
	"context"
	"fmt"
	"gorm.io/gen"

	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/record"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/query"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/packer"
)

func Save(ctx context.Context, req *rao.SaveSceneReq, userID string) (string, string, error) {
	target := packer.TransSaveSceneReqToTargetModel(req, userID)
	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		// 接口名排重
		conditions := make([]gen.Condition, 0)
		conditions = append(conditions, tx.Target.TeamID.Eq(req.TeamID))
		conditions = append(conditions, tx.Target.TargetType.Eq(consts.TargetTypeScene))
		conditions = append(conditions, tx.Target.TargetID.Neq(req.TargetID))
		conditions = append(conditions, tx.Target.Source.Eq(req.Source))
		conditions = append(conditions, tx.Target.Name.Eq(req.Name))
		conditions = append(conditions, tx.Target.ParentID.Eq(req.ParentID))
		conditions = append(conditions, tx.Target.Status.Eq(consts.TargetStatusNormal))
		if req.Source == consts.TargetSourcePlan || req.Source == consts.TargetSourceAutoPlan {
			conditions = append(conditions, tx.Target.PlanID.Eq(req.PlanID))
		}
		_, err := tx.Target.WithContext(ctx).Where(conditions...).First()
		if err == nil {
			return fmt.Errorf("名称已存在")
		}

		// 查询是否存在
		_, err = tx.Target.WithContext(ctx).Where(tx.Target.TargetID.Eq(req.TargetID)).First()
		if err != nil {
			if err := tx.Target.WithContext(ctx).Create(target); err != nil {
				return err
			}

			return record.InsertCreate(ctx, target.TeamID, userID, record.OperationOperateCreateScene, target.Name)
		}

		// 修改场景数据
		if _, err := tx.Target.WithContext(ctx).Where(tx.Target.TargetID.Eq(req.TargetID)).UpdateSimple(tx.Target.Name.Value(req.Name),
			tx.Target.Description.Value(req.Description)); err != nil {
			return err
		}
		return record.InsertUpdate(ctx, target.TeamID, userID, record.OperationOperateUpdateScene, target.Name)
	})
	return target.TargetID, target.Name, err
}

func BatchGetByTargetID(ctx context.Context, teamID string, TargetID []string, source int32) ([]*rao.Scene, error) {
	tx := query.Use(dal.DB()).Target
	t, err := tx.WithContext(ctx).Where(
		tx.TargetID.In(TargetID...),
		tx.TeamID.Eq(teamID),
		tx.TargetType.Eq(consts.TargetTypeScene),
		tx.Status.Eq(consts.TargetStatusNormal),
		tx.Source.Eq(source),
	).Find()

	if err != nil {
		return nil, err
	}

	return packer.TransTargetToRaoScene(t), nil
}
