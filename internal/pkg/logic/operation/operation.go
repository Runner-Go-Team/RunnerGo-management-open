package operation

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"kp-management/internal/pkg/biz/consts"
	"kp-management/internal/pkg/dal"
	"kp-management/internal/pkg/dal/mao"
	"kp-management/internal/pkg/dal/query"
	"kp-management/internal/pkg/dal/rao"
	"kp-management/internal/pkg/packer"
)

func List(ctx *gin.Context, teamID string, limit, offset int) ([]*rao.Operation, int64, error) {
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectOperationLog)
	findOptions := new(options.FindOptions)
	if limit > 0 {
		findOptions.SetLimit(int64(limit))
		findOptions.SetSkip(int64(offset))
		sort := bson.D{{"created_time_sec", -1}}
		findOptions.SetSort(sort)
	}

	cur1, err := collection.Find(ctx, bson.D{{"team_id", teamID}})
	if err != nil {
		return nil, 0, err
	}

	var operationLog []*mao.OperationLog
	if err := cur1.All(ctx, &operationLog); err != nil {
		return nil, 0, err
	}

	total := int64(len(operationLog))

	cur, err := collection.Find(ctx, bson.D{{"team_id", teamID}}, findOptions)
	if err != nil {
		return nil, 0, err
	}

	if err := cur.All(ctx, &operationLog); err != nil {
		return nil, 0, err
	}

	var userIDs []string
	for _, olInfo := range operationLog {
		userIDs = append(userIDs, olInfo.UserID)
	}

	u := query.Use(dal.DB()).User
	users, err := u.WithContext(ctx).Where(u.UserID.In(userIDs...)).Find()
	if err != nil {
		return nil, 0, err
	}

	return packer.TransOperationsToRaoOperationList(operationLog, users), total, nil

}
