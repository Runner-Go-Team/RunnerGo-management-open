package handler

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/errno"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/record"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/response"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/uuid"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/query"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/gin-gonic/gin"
	"github.com/go-omnibus/omnibus"
	"time"
)

// OpenRunStressPlan 运行计划
func OpenRunStressPlan(ctx *gin.Context) {
	var req rao.OpenRunStressPlanReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	// 根据用户账密，查出用户id
	userInfo, err := GetUserInfoByAccountAndPassword(ctx, req.Account, req.Password)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrAuthFailed, err.Error())
		return
	}

	// 调用controller方法改成本地
	runStressParams := RunStressReq{
		PlanID:  req.PlanID,
		TeamID:  req.TeamID,
		UserID:  userInfo.UserID,
		RunType: 1,
	}

	errnoNum, newReportIDs, runErr := RunStress(ctx, runStressParams)
	if runErr != nil {
		response.ErrorWithMsg(ctx, errnoNum, runErr.Error())
		return
	}

	px := dal.GetQuery().StressPlan
	planInfo, err := px.WithContext(ctx).Where(px.TeamID.Eq(req.TeamID), px.PlanID.Eq(req.PlanID)).First()
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	// 插入操作日志
	if planInfo.TaskType == consts.PlanTaskTypeNormal || runStressParams.RunType == 2 {
		if err := record.InsertRun(ctx, req.TeamID, userInfo.UserID, record.OperationOperateRunPlan, planInfo.PlanName); err != nil {
			response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
			return
		}
	} else {
		if err := record.InsertExecute(ctx, req.TeamID, userInfo.UserID, record.OperationOperateExecPlan, planInfo.PlanName); err != nil {
			response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
			return
		}
	}

	// 执行计划当次操作和报告的关系
	planRunUUID := uuid.GetUUID()
	planRunUUIDRedisKey := consts.RedisPlanRunUUIDRelateReports + planRunUUID
	_ = dal.GetRDB().SAdd(ctx, planRunUUIDRedisKey, newReportIDs).Err()
	_ = dal.GetRDB().Expire(ctx, planRunUUIDRedisKey, time.Second*86400).Err()
	for _, r := range newReportIDs {
		reportPlanRunRedisKey := consts.RedisReportPlanRunUUID + r
		_ = dal.GetRDB().Set(ctx, reportPlanRunRedisKey, planRunUUID, time.Second*86400).Err()
	}

	response.Success(ctx)
	return
}

func GetUserInfoByAccountAndPassword(ctx *gin.Context, account, password string) (*model.User, error) {
	userInfo := new(model.User)
	var err error
	err = dal.GetQuery().Transaction(func(tx *query.Query) error {
		userInfo, err = tx.User.WithContext(ctx).Where(tx.User.Account.Eq(account)).First()
		if err != nil {
			return err
		}

		if err = omnibus.CompareBcryptHashAndPassword(userInfo.Password, password); err != nil {
			return err
		}
		return nil
	})

	return userInfo, nil
}

// OpenRunAutoPlan 运行自动化测试计划
func OpenRunAutoPlan(ctx *gin.Context) {
	var req rao.OpenRunAutoPlanReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	// 根据用户账密，查出用户id
	userInfo, err := GetUserInfoByAccountAndPassword(ctx, req.Account, req.Password)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrAuthFailed, err.Error())
		return
	}

	// 调用controller方法改成本地
	runAutoPlanParams := RunAutoPlanReq{
		PlanID:  req.PlanID,
		TeamID:  req.TeamID,
		SceneID: req.SceneID,
		UserID:  userInfo.UserID,
	}

	errnoNum, runErr := RunAutoPlanDetail(ctx, runAutoPlanParams)
	if runErr != nil {
		response.ErrorWithMsg(ctx, errnoNum, runErr.Error())
		return
	}

	px := dal.GetQuery().AutoPlan
	autoPlanInfo, err := px.WithContext(ctx).Where(px.TeamID.Eq(req.TeamID), px.PlanID.Eq(req.PlanID)).First()
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	// 添加操作日志
	if autoPlanInfo.TaskType == consts.PlanTaskTypeNormal || runAutoPlanParams.RunType == 2 {
		if err := record.InsertRun(ctx, req.TeamID, userInfo.UserID, record.OperationOperateRunPlan, autoPlanInfo.PlanName); err != nil {
			response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
			return
		}
	} else {
		if err := record.InsertExecute(ctx, req.TeamID, userInfo.UserID, record.OperationOperateExecPlan, autoPlanInfo.PlanName); err != nil {
			response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
			return
		}
	}
	response.SuccessWithData(ctx, rao.RunAutoPlanResp{
		TaskType: autoPlanInfo.TaskType,
	})
	return
}
