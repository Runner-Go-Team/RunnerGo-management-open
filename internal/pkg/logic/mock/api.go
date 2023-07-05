package mock

import (
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/jwt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/record"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/mao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/runner"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/target"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/packer"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func SendAPI(ctx *gin.Context, teamID string, targetID string) (string, error) {
	tx := dal.GetQuery().MockTarget
	targetInfo, err := tx.WithContext(ctx).Where(tx.TargetID.Eq(targetID)).First()
	if err != nil {
		return "", err
	}

	var apiInfo mao.API
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectAPI)
	err = collection.FindOne(ctx, bson.D{{"target_id", targetID}}).Decode(&apiInfo)
	if err != nil {
		return "", err
	}

	// 获取全局变量
	globalVariable, err := target.GetGlobalVariable(ctx, teamID)

	// 把调试信息入库
	targetDebugLog := dal.GetQuery().MockTargetDebugLog
	insertData := &model.MockTargetDebugLog{
		TargetID:   targetID,
		TargetType: consts.TargetDebugLogApi,
		TeamID:     teamID,
	}
	err = targetDebugLog.WithContext(ctx).Create(insertData)
	if err != nil {
		return "", err
	}

	userID := jwt.GetUserIDByCtx(ctx)
	if err := record.InsertDebug(ctx, teamID, userID, record.OperationOperateDebugMockApi, targetInfo.Name); err != nil {
		return "", err
	}

	retID, err := runner.RunTarget(packer.GetRunMockTargetParam(targetInfo, globalVariable, &apiInfo))
	if err != nil {
		return "", fmt.Errorf("调试接口返回非200状态")
	}

	return retID, err
}
