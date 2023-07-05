package mock

import (
	"context"
	"errors"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/record"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/uuid"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/mao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/query"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/folder"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/packer"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/public"
	"go.mongodb.org/mongo-driver/bson"
	"gorm.io/gorm"
)

func FolderReqSave(ctx context.Context, userID string, req *rao.MockSaveFolderReq) error {
	target := packer.TransSaveFolderReqToMockTargetModel(req, userID)
	return query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		// 接口名排重
		_, err := tx.MockTarget.WithContext(ctx).Where(tx.MockTarget.TeamID.Eq(req.TeamID), tx.MockTarget.Name.Eq(req.Name),
			tx.MockTarget.TargetType.Eq(consts.TargetTypeFolder), tx.MockTarget.TargetID.Neq(req.TargetID),
			tx.MockTarget.Status.Eq(consts.TargetStatusNormal), tx.MockTarget.Source.Eq(consts.MockTargetSourceApi),
			tx.MockTarget.ParentID.Eq(req.ParentID)).First()
		if err == nil {
			return fmt.Errorf("名称已存在")
		}

		// 查询当前接口是否存在
		_, err = tx.MockTarget.WithContext(ctx).Where(tx.MockTarget.TargetID.Eq(req.TargetID)).First()
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if errors.Is(err, gorm.ErrRecordNotFound) { // 需新增
			if err = tx.MockTarget.WithContext(ctx).Create(target); err != nil {
				return err
			}

			if err = record.InsertCreate(ctx, target.TeamID, userID, record.OperationOperateCreateMockFolder, target.Name); err != nil {
				return err
			}
		} else {
			if _, err = tx.MockTarget.WithContext(ctx).Where(tx.MockTarget.TargetID.Eq(req.TargetID)).Updates(target); err != nil {
				return err
			}

			if err = record.InsertUpdate(ctx, target.TeamID, userID, record.OperationOperateUpdateMockFolder, target.Name); err != nil {
				return err
			}
		}

		return nil
	})
}

// SaveToTarget mock 接口保存至测试对象
func SaveToTarget(ctx context.Context, userID string, targetIDs []string, teamID string) error {
	// step1:查询 MockTarget 全部的target
	// step2:需要替换 targetID parentID
	// step3:更新数据
	if len(targetIDs) <= 0 {
		return nil
	}
	t := query.Use(dal.DB()).Target
	mt := query.Use(dal.DB()).MockTarget
	mockTargets, err := mt.WithContext(ctx).Where(
		mt.TeamID.Eq(teamID),
		mt.TargetID.In(targetIDs...),
		mt.TargetType.Eq(consts.TargetTypeFolder),
	).Order(mt.Sort, mt.CreatedAt.Desc()).Find()
	if err != nil {
		return err
	}

	mockTargetsApi, err := mt.WithContext(ctx).Where(
		mt.TeamID.Eq(teamID),
		mt.TargetID.In(targetIDs...),
		mt.TargetType.Eq(consts.TargetTypeAPI),
	).Order(mt.Sort, mt.CreatedAt.Desc()).Find()
	if err != nil {
		return err
	}

	mockTargets = append(mockTargets, mockTargetsApi...)
	if len(mockTargets) <= 0 {
		return nil
	}

	newTargetIDs := make(map[string]string, len(mockTargets))
	newTargetIDs["0"] = "0"
	for _, target := range mockTargets {
		newTargetIDs[target.TargetID] = uuid.GetUUID()
	}

	return query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		t = tx.Target
		mt = tx.MockTarget
		collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectAPI)

		// 排序
		var sort int32
		targetSort, err := t.WithContext(ctx).Select(t.Sort).Where(t.TeamID.Eq(teamID)).Order(t.Sort.Desc()).First()
		if err == nil {
			sort = targetSort.Sort + 1
		}

		for k, mtv := range mockTargets {
			if mtv.TargetType == consts.TargetTypeFolder {
				sort = sort + int32(k)
				saveFolder := &model.Target{
					TargetID:      newTargetIDs[mtv.TargetID],
					TeamID:        mtv.TeamID,
					TargetType:    mtv.TargetType,
					Name:          consts.MockFolderSaveNamePrefix + mtv.Name,
					ParentID:      newTargetIDs[mtv.ParentID],
					Method:        mtv.Method,
					Sort:          sort,
					TypeSort:      mtv.TypeSort,
					Status:        consts.TargetStatusNormal,
					Version:       1,
					CreatedUserID: userID,
					RecentUserID:  userID,
					Source:        consts.TargetSourceApi,
					Description:   mtv.Description,
					SourceID:      mtv.TargetID,
				}

				// 判断是否存在当前名称
				targetInfo, err := t.WithContext(ctx).Where(
					t.TeamID.Eq(saveFolder.TeamID),
					t.Name.Eq(saveFolder.Name),
					t.TargetID.Neq(saveFolder.TargetID),
					t.TargetType.Eq(consts.TargetTypeFolder),
					t.Status.Eq(consts.TargetStatusNormal),
					t.Source.Eq(consts.TargetSourceApi),
					t.ParentID.Eq(saveFolder.ParentID),
				).First()
				if errors.Is(err, gorm.ErrRecordNotFound) {
					if err := t.WithContext(ctx).Create(saveFolder); err != nil {
						return err
					}
				}
				if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
					return err
				}

				if targetInfo != nil {
					newTargetIDs[mtv.TargetID] = targetInfo.TargetID
					continue
				}

			}

			if mtv.TargetType == consts.TargetTypeAPI {
				sort = sort + int32(k)
				saveTarget := &model.Target{
					TargetID:      newTargetIDs[mtv.TargetID],
					TeamID:        mtv.TeamID,
					TargetType:    mtv.TargetType,
					Name:          consts.MockSaveNamePrefix + mtv.Name,
					ParentID:      newTargetIDs[mtv.ParentID],
					Method:        mtv.Method,
					Sort:          sort,
					TypeSort:      mtv.TypeSort,
					Status:        consts.TargetStatusNormal,
					Version:       1,
					CreatedUserID: userID,
					RecentUserID:  userID,
					Source:        consts.TargetSourceApi,
					Description:   mtv.Description,
					SourceID:      mtv.TargetID,
				}

				// 判断是否存在当前名称
				var names []string
				if err = t.WithContext(ctx).Where(
					t.TeamID.Eq(saveTarget.TeamID),
					t.Name.Like(fmt.Sprintf("%s%%", saveTarget.Name)),
					t.TargetID.Neq(saveTarget.TargetID),
					t.TargetType.Eq(consts.TargetTypeAPI),
					t.Status.Eq(consts.TargetStatusNormal),
					t.Source.Eq(consts.TargetSourceApi),
					t.ParentID.Eq(saveTarget.ParentID),
				).Pluck(t.Name, &names); err != nil {
					return err
				}
				//	[xxx-01,xxx-02,xxx-04]
				if len(names) > 0 {
					for _, s := range consts.MockNameOrder {
						newName := saveTarget.Name + s
						if !public.ContainsStringSlice(names, newName) {
							saveTarget.Name = newName
							break
						}
					}
				}

				// 添加 target
				if err := t.WithContext(ctx).Create(saveTarget); err != nil {
					return err
				}

				// 添加 api
				cur, err := collection.Find(ctx, bson.D{{"target_id", mtv.TargetID}})
				if err != nil {
					return err
				}
				var apis []*mao.API
				if err := cur.All(ctx, &apis); err != nil {
					return fmt.Errorf("api数据获取失败")
				}

				for _, api := range apis {
					api.TargetID = saveTarget.TargetID
					if _, err := collection.InsertOne(ctx, api); err != nil {
						return err
					}
				}

			}
		}

		return nil
	})
}

// FolderSaveCascade 级联同步测试对象目录（存在目录名称忽略）
func FolderSaveCascade(ctx context.Context, userID, parentId, teamID string) (string, error) {
	if len(parentId) <= 0 {
		return parentId, nil
	}

	parentTargets := make([]*model.MockTarget, 0)

	// 避免死循环 限制10次
	var nums [10][0]int //  空数组，不占用内存大小，不用额外开销内存
	for range nums {
		// 查询当前 target 目录
		tx := query.Use(dal.DB()).MockTarget
		t, err := tx.WithContext(ctx).Where(
			tx.TargetID.Eq(parentId),
			tx.TeamID.Eq(teamID),
			tx.TargetType.Eq(consts.TargetTypeFolder),
			tx.Status.Eq(consts.TargetStatusNormal),
		).First()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return parentId, nil
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return "", err
		}

		parentTargets = append(parentTargets, t)

		if len(t.ParentID) <= 0 {
			break
		}

		parentId = t.ParentID
	}

	// 从最大的开始创建
	parentTargets = public.ReverseSlice(parentTargets)
	var newParentId string
	target := query.Use(dal.DB()).Target
	for _, t := range parentTargets {
		newTargetID := uuid.GetUUID()
		saveFolder := &rao.SaveFolderReq{
			TargetID:    newTargetID,
			TeamID:      t.TeamID,
			ParentID:    newParentId,
			Name:        consts.MockFolderSaveNamePrefix + t.Name,
			Method:      t.Method,
			Version:     1,
			Description: t.Description,
			Source:      consts.TargetSourceApi,
		}

		// 判断是否存在当前名称
		targetInfo, err := target.WithContext(ctx).Where(
			target.TeamID.Eq(saveFolder.TeamID),
			target.Name.Eq(saveFolder.Name),
			target.TargetType.Eq(consts.TargetTypeFolder),
			target.Status.Eq(consts.TargetStatusNormal),
			target.Source.Eq(consts.TargetSourceApi),
			target.ParentID.Eq(saveFolder.ParentID),
		).First()
		if err != nil {
			newParentId = targetInfo.TargetID
			continue
		}

		if err := folder.Save(ctx, userID, saveFolder); err != nil {
			return parentId, err
		}
		newParentId = newTargetID
	}

	return newParentId, nil
}

func GetByTargetID(ctx context.Context, teamID string, targetID string) (*rao.MockFolder, error) {
	tx := query.Use(dal.DB()).MockTarget
	t, err := tx.WithContext(ctx).Where(
		tx.TargetID.Eq(targetID),
		tx.TeamID.Eq(teamID),
		tx.TargetType.Eq(consts.TargetTypeFolder),
		tx.Status.Eq(consts.TargetStatusNormal),
	).First()

	if err != nil {
		return nil, err
	}

	return packer.TransTargetToRaoMockFolder(t), nil
}
