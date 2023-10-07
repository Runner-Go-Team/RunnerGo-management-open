package record

import (
	"context"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/mao"
	"time"
)

func InsertCreate(ctx context.Context, teamID string, userID string, operate int32, name string) error {
	return insert(ctx, teamID, userID, name, consts.OperationCategoryCreate, operate)
}

func InsertUpdate(ctx context.Context, teamID string, userID string, operate int32, name string) error {
	return insert(ctx, teamID, userID, name, consts.OperationCategoryUpdate, operate)
}

func InsertDelete(ctx context.Context, teamID string, userID string, operate int32, name string) error {
	return insert(ctx, teamID, userID, name, consts.OperationCategoryDelete, operate)
}

func InsertRun(ctx context.Context, teamID string, userID string, operate int32, name string) error {
	return insert(ctx, teamID, userID, name, consts.OperationCategoryRun, operate)
}

func InsertDebug(ctx context.Context, teamID string, userID string, operate int32, name string) error {
	return insert(ctx, teamID, userID, name, consts.OperationCategoryDebug, operate)
}

func InsertExecute(ctx context.Context, teamID string, userID string, operate int32, name string) error {
	return insert(ctx, teamID, userID, name, consts.OperationCategoryExecute, operate)
}

func insert(ctx context.Context, teamID string, userID string, name string, category, operate int32) error {
	//return query.Use(dal.DB()).Operation.WithContext(ctx).Create(&model.Operation{
	//	TeamID:    teamID,
	//	UserID:    userID,
	//	Category:  category,
	//	Name:      name,
	//	Operate:   operate,
	//	CreatedAt: time.Now(),
	//})
	nowTimeInt := time.Now().Unix()
	nowTimeDate := time.Now().Local().Format("2006-01-02 15:04:05")
	operationLog := mao.OperationLog{
		TeamID:      teamID,
		UserID:      userID,
		Category:    category,
		Operate:     operate,
		Name:        name,
		CreatedAt:   nowTimeInt,
		CreatedDate: nowTimeDate,
	}
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectOperationLog)
	if _, err := collection.InsertOne(ctx, operationLog); err != nil {
		return err
	}
	return nil
}
