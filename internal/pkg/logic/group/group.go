package group

import (
	"context"
	"fmt"

	"kp-management/internal/pkg/biz/consts"
	"kp-management/internal/pkg/biz/record"
	"kp-management/internal/pkg/dal"
	"kp-management/internal/pkg/dal/query"
	"kp-management/internal/pkg/dal/rao"
	"kp-management/internal/pkg/packer"
)

func Save(ctx context.Context, req *rao.SaveGroupReq, userID string) error {
	target := packer.TransSaveGroupReqToTargetModel(req, userID)
	return query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		// 分组名排重
		_, err := tx.Target.WithContext(ctx).Where(tx.Target.TeamID.Eq(req.TeamID), tx.Target.Name.Eq(req.Name),
			tx.Target.TargetType.Eq(consts.TargetTypeGroup), tx.Target.Source.Eq(req.Source), tx.Target.TargetID.Neq(req.TargetID)).First()
		if err == nil {
			return fmt.Errorf("名称已存在")
		}

		// 查询是否存在
		_, err = tx.Target.WithContext(ctx).Where(tx.Target.TargetID.Eq(req.TargetID)).First()
		if err != nil {
			if err := tx.Target.WithContext(ctx).Create(target); err != nil {
				return err
			}
			return record.InsertCreate(ctx, target.TeamID, userID, record.OperationOperateCreateGroup, target.Name)
		}

		if _, err := tx.Target.WithContext(ctx).Where(tx.Target.TargetID.Eq(req.TargetID)).Updates(target); err != nil {
			return err
		}
		return record.InsertUpdate(ctx, target.TeamID, userID, record.OperationOperateUpdateGroup, target.Name)
	})
}

func GetByTargetID(ctx context.Context, teamID string, targetID string, source int32) (*rao.Group, error) {
	tx := query.Use(dal.DB()).Target
	t, err := tx.WithContext(ctx).Where(
		tx.TargetID.Eq(targetID),
		tx.TeamID.Eq(teamID),
		tx.TargetType.Eq(consts.TargetTypeGroup),
		tx.Status.Eq(consts.TargetStatusNormal),
		tx.Source.Eq(source),
	).First()

	if err != nil {
		return nil, err
	}

	return packer.TransTargetToRaoGroup(t, nil), nil
}
