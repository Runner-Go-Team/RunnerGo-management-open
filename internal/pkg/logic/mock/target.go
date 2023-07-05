package mock

import (
	"context"
	"errors"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/api/v1alpha1"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/record"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/clients"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/mao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/query"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	lcapi "github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/api"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/packer"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/public"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"gorm.io/gorm"
)

func Save(ctx context.Context, req *rao.MockSaveTargetReq, userID string) (string, error) {
	target := packer.TransSaveTargetReqToMockTargetModel(req, userID)
	api := packer.TransSaveMockTargetReqToMaoAPI(req)
	mockApiCase := packer.TransExpectToMaoMockCase(req.Expects)
	mock := packer.TransSaveMockTargetReqToMaoMock(req, mockApiCase)

	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectAPI)
	collectionMock := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectMock)

	if len(mock.Path) <= 0 {
		return "", fmt.Errorf("路径不能为空")
	}

	// mock path 不能重复
	cursor, err := collectionMock.Find(ctx, bson.D{{"path", mock.Path}, {"method", mock.Method}, {"target_id", bson.D{{"$ne", mock.TargetID}}}})
	if err != nil {
		return "", err
	}

	var mocks []*mao.Mock
	if err := cursor.All(ctx, &mocks); err != nil {
		return "", err
	}

	if len(mocks) > 0 {
		mt := query.Use(dal.DB()).MockTarget
		var repeatTargetIDs []string
		for _, mock := range mocks {
			repeatTargetIDs = append(repeatTargetIDs, mock.TargetID)
		}
		targets, err := mt.WithContext(ctx).Where(
			mt.TargetID.In(repeatTargetIDs...),
			mt.Status.Eq(consts.TargetStatusNormal),
			mt.TargetID.Neq(target.TargetID),
		).Find()
		if err != nil {
			return "", err
		}
		if len(targets) > 0 {
			return "", fmt.Errorf("路径已存在，不能重复")
		}
	}

	err = query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		// 接口名排重
		_, err := tx.MockTarget.WithContext(ctx).Where(tx.MockTarget.TeamID.Eq(req.TeamID), tx.MockTarget.Name.Eq(req.Name),
			tx.MockTarget.TargetType.Eq(consts.TargetTypeAPI), tx.MockTarget.TargetID.Neq(req.TargetID),
			tx.MockTarget.Status.Eq(consts.TargetStatusNormal), tx.MockTarget.ParentID.Eq(req.ParentID),
			tx.MockTarget.Source.Eq(req.Source)).First()
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
			api.TargetID = target.TargetID
			if _, err = collection.InsertOne(ctx, api); err != nil {
				return err
			}

			if _, err = collectionMock.InsertOne(ctx, mock); err != nil {
				return err
			}

			if err = record.InsertCreate(ctx, target.TeamID, userID, record.OperationOperateCreateMockAPI, target.Name); err != nil {
				return err
			}
		} else { // 修改
			if _, err = tx.MockTarget.WithContext(ctx).Where(tx.MockTarget.TargetID.Eq(req.TargetID)).Updates(target); err != nil {
				return err
			}

			if _, err = collection.UpdateOne(ctx, bson.D{{"target_id", target.TargetID}}, bson.M{"$set": api}); err != nil {
				return err
			}

			// mock 数据
			if _, err = collectionMock.UpdateOne(ctx, bson.D{{"target_id", target.TargetID}}, bson.M{"$set": mock}); err != nil {
				return err
			}

			if err = record.InsertUpdate(ctx, target.TeamID, userID, record.OperationOperateUpdateMockAPI, target.Name); err != nil {
				return err
			}
		}

		// 同步 mock 服务
		var mockApi v1alpha1.MockAPI
		mockApi.UniqueKey = mock.UniqueKey
		mockApi.Path = mock.Path
		mockApi.Method = mock.Method
		mockApi.Cases = mockApiCase
		if mock.IsMockOpen == consts.IsMockOpenOff {
			if err = clients.DeleteMockAPI(&mockApi); err != nil {
				return err
			}
		} else {
			if err = clients.SaveMockAPI(&mockApi); err != nil {
				return err
			}
		}

		return nil
	})

	// 保存并同步测试对象
	if req.OperateType == consts.MockOperateTypeSaveAndTarget {
		// step1:查询 当前 targetID 的 parentID 递归查询
		// step2:调用 SaveToTarget
		var targetIDs []string
		targetIDs = append(targetIDs, req.TargetID)
		parentId := req.ParentID
		// 避免死循环 限制10次
		var nums [10][0]int //  空数组，不占用内存大小，不用额外开销内存
		for range nums {
			// 查询当前 target 目录
			tx := query.Use(dal.DB()).MockTarget
			t, err := tx.WithContext(ctx).Where(
				tx.TargetID.Eq(parentId),
				tx.TeamID.Eq(req.TeamID),
				tx.TargetType.Eq(consts.TargetTypeFolder),
				tx.Status.Eq(consts.TargetStatusNormal),
			).First()
			if errors.Is(err, gorm.ErrRecordNotFound) {
				break
			}
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return "", err
			}
			targetIDs = append(targetIDs, t.TargetID)
			if len(t.ParentID) <= 0 {
				break
			}

			parentId = t.ParentID
		}

		targetIDs = public.ReverseSlice(targetIDs)
		if err := SaveToTarget(ctx, userID, targetIDs, req.TeamID); err != nil {
			return "", err
		}

		if err = record.InsertUpdate(ctx, target.TeamID, userID, record.OperationOperateMockTarget, target.Name); err != nil {
			return "", err
		}
	}

	// 保存并同步
	if req.OperateType == consts.MockOperateTypeSaveAndSync {
		t := query.Use(dal.DB()).Target
		targetInfos, err := t.WithContext(ctx).Where(
			t.SourceID.Eq(req.TargetID),
			t.Status.Eq(consts.TargetStatusNormal),
		).Find()
		if err != nil {
			return "", err
		}

		if len(targetInfos) > 0 {
			for _, t := range targetInfos {
				req.TargetID = t.TargetID
				req.TargetType = t.TargetType
				req.Name = t.Name
				req.ParentID = t.ParentID
				req.Sort = t.Sort
				req.TypeSort = t.TypeSort
				req.Version = t.Version
				req.SourceID = target.TargetID
				req.Source = consts.TargetSourceApi
				if _, err := lcapi.Save(ctx, req.SaveTargetReq, userID); err != nil {
					return "", err
				}
			}
		}

		if err = record.InsertUpdate(ctx, target.TeamID, userID, record.OperationOperateMockSyncApi, target.Name); err != nil {
			return "", err
		}
	}

	return target.TargetID, err
}

func SortTarget(ctx context.Context, req *rao.MockSortTargetReq) error {
	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		targetNameMap := make(map[string]int, len(req.Targets))
		for _, target := range req.Targets {
			if _, ok := targetNameMap[target.Name]; ok {
				return fmt.Errorf("存在重名，无法操作")
			} else {
				targetNameMap[target.Name] = 1
			}
		}

		for _, target := range req.Targets {
			_, err := tx.MockTarget.WithContext(ctx).Where(tx.MockTarget.TeamID.Eq(target.TeamID),
				tx.MockTarget.TargetID.Eq(target.TargetID)).UpdateSimple(tx.MockTarget.Sort.Value(target.Sort), tx.MockTarget.ParentID.Value(target.ParentID))
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func ListFolderAPI(ctx context.Context, req *rao.MockListFolderAPIReq) ([]*rao.MockFolderAPI, error) {
	tx := query.Use(dal.DB()).MockTarget
	targets, err := tx.WithContext(ctx).Where(
		tx.TeamID.Eq(req.TeamID),
		tx.TargetType.In(consts.TargetTypeFolder, consts.TargetTypeAPI,
			consts.TargetTypeSql, consts.TargetTypeTcp, consts.TargetTypeWebsocket,
			consts.TargetTypeMQTT, consts.TargetTypeDubbo),
		tx.Status.Eq(consts.TargetStatusNormal),
		tx.Source.Eq(req.Source)).Order(tx.Sort, tx.CreatedAt.Desc()).Find()

	// 查询 mock 详情数据
	targetIDs := make([]string, 0, len(targets))
	for _, target := range targets {
		targetIDs = append(targetIDs, target.TargetID)
	}
	collectionMock := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectMock)
	cursorMock, err := collectionMock.Find(ctx, bson.D{{"target_id", bson.D{{"$in", targetIDs}}}})
	if err != nil {
		return nil, err
	}
	var mocks []*mao.Mock
	if err = cursorMock.All(ctx, &mocks); err != nil {
		return nil, err
	}

	mockMap := make(map[string]*mao.Mock, len(mocks))
	for _, mockInfo := range mocks {
		mockMap[mockInfo.TargetID] = mockInfo
	}

	if err != nil {
		return nil, err
	}

	return packer.TransMockTargetToRaoFolderAPIList(targets, mockMap), nil
}

func Trash(ctx *gin.Context, targetID string, userID string) error {
	return dal.GetQuery().Transaction(func(tx *query.Query) error {
		t, err := tx.MockTarget.WithContext(ctx).Where(tx.MockTarget.TargetID.Eq(targetID)).First()
		if err != nil {
			return err
		}

		// 删除
		_ = getAllSonTargetID(ctx, targetID, t.TargetType)

		var operate int32 = 0
		if t.TargetType == consts.TargetTypeFolder {
			operate = record.OperationOperateDeleteMockFolder
		} else {
			operate = record.OperationOperateDeleteMockApi
		}
		if err := record.InsertDelete(ctx, t.TeamID, userID, operate, t.Name); err != nil {
			return err
		}

		// 同步 mock 服务
		var mockApi v1alpha1.MockAPI
		mockApi.UniqueKey = targetID
		if err = clients.DeleteMockAPI(&mockApi); err != nil {
			return err
		}

		return nil
	})
}

func getAllSonTargetID(ctx *gin.Context, targetID string, targetType string) error {
	tx := dal.GetQuery().MockTarget
	if targetType == consts.TargetTypeFolder {
		// 查询这个目录下是否还有别的目录或文件
		targetList, err := tx.WithContext(ctx).Where(tx.ParentID.Eq(targetID)).Find()
		if err != nil {
			return err
		}
		if len(targetList) > 0 {
			for _, tInfo := range targetList {
				_ = getAllSonTargetID(ctx, tInfo.TargetID, tInfo.TargetType)
			}
		}
		// 删除目录本身
		_, err = tx.WithContext(ctx).Where(tx.TargetID.Eq(targetID)).UpdateSimple(tx.Status.Value(consts.TargetStatusTrash))
		if err != nil {
			return err
		}
	} else {
		_, err := tx.WithContext(ctx).Where(tx.TargetID.Eq(targetID)).UpdateSimple(tx.Status.Value(consts.TargetStatusTrash))
		if err != nil {
			return err
		}
	}
	return nil
}

func DetailByTargetIDs(ctx context.Context, req *rao.MockBatchGetDetailReq) ([]rao.MockAPIDetail, error) {
	tx := query.Use(dal.DB()).MockTarget
	targets, err := tx.WithContext(ctx).Where(
		tx.TargetID.In(req.TargetIDs...),
		tx.TeamID.Eq(req.TeamID),
		tx.Status.Eq(consts.TargetStatusNormal),
	).Order(tx.Sort.Desc(), tx.CreatedAt.Desc()).Find()

	if err != nil {
		return nil, err
	}

	if len(targets) == 0 {
		return nil, fmt.Errorf("删除失败")
	}

	apiIDs := make([]string, 0, len(targets))
	for _, targetInfo := range targets {
		if targetInfo.TargetType == consts.MockTargetTypeAPI {
			apiIDs = append(apiIDs, targetInfo.TargetID)
		}
	}

	// 查询接口详情数据
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectAPI)
	cursor, err := collection.Find(ctx, bson.D{{"target_id", bson.D{{"$in", apiIDs}}}})
	if err != nil {
		return nil, err
	}
	var apis []*mao.API
	if err = cursor.All(ctx, &apis); err != nil {
		return nil, err
	}

	apiMap := make(map[string]*mao.API, len(apis))
	for _, apiInfo := range apis {
		apiMap[apiInfo.TargetID] = apiInfo
	}

	// 查询 mock 详情数据
	collectionMock := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectMock)
	cursorMock, err := collectionMock.Find(ctx, bson.D{{"target_id", bson.D{{"$in", apiIDs}}}})
	if err != nil {
		return nil, err
	}
	var mocks []*mao.Mock
	if err = cursorMock.All(ctx, &mocks); err != nil {
		return nil, err
	}

	mockMap := make(map[string]*mao.Mock, len(apis))
	for _, mockInfo := range mocks {
		mockMap[mockInfo.TargetID] = mockInfo
	}

	return packer.TransMockTargetsToRaoAPIDetails(targets, apiMap, mockMap), nil
}
