package handler

import (
	"errors"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/errno"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/jwt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/response"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/errmsg"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/uiScene"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// UISceneSaveOperator 保存UI场景
func UISceneSaveOperator(ctx *gin.Context) {
	var req rao.UISceneSaveOperatorReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
			return
		}

		response.ErrorWithMsg(ctx, errno.ErrCheckParams, removeTopStruct(errs.Translate(dal.GetTrans())))
		return
	}

	name, err := uiScene.CheckSaveOperatorAndGetNameReq(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrUISceneOptEmpty, err.Error())
		return
	}
	req.Name = name
	req.IsReSort = true

	_, err = uiScene.SaveOperator(ctx, jwt.GetUserIDByCtx(ctx), &req)
	if err != nil {
		if errors.Is(err, errmsg.ErrUISceneRequired) {
			response.ErrorWithMsg(ctx, errno.ErrUISceneRequired, "")
			return
		}
		if errors.Is(err, errmsg.ErrElementLocatorNotFound) {
			response.ErrorWithMsg(ctx, errno.ErrElementLocatorNotFound, "")
			return
		}
		if errors.Is(err, errmsg.ErrUISceneNameRepeat) {
			response.ErrorWithMsg(ctx, errno.ErrUISceneNameRepeat, "")
			return
		}
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}

// UISceneUpdateOperator 保存UI场景
func UISceneUpdateOperator(ctx *gin.Context) {
	var req rao.UISceneSaveOperatorReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
			return
		}

		response.ErrorWithMsg(ctx, errno.ErrCheckParams, removeTopStruct(errs.Translate(dal.GetTrans())))
		return
	}

	name, err := uiScene.CheckSaveOperatorAndGetNameReq(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrUISceneOptEmpty, err.Error())
		return
	}
	req.Name = name
	err = uiScene.UpdateOperator(ctx, jwt.GetUserIDByCtx(ctx), &req)
	if err != nil {
		if errors.Is(err, errmsg.ErrUISceneRequired) {
			response.ErrorWithMsg(ctx, errno.ErrUISceneRequired, "")
			return
		}
		if errors.Is(err, errmsg.ErrElementLocatorNotFound) {
			response.ErrorWithMsg(ctx, errno.ErrElementLocatorNotFound, "")
			return
		}
		if errors.Is(err, errmsg.ErrUISceneNameRepeat) {
			response.ErrorWithMsg(ctx, errno.ErrUISceneNameRepeat, "")
			return
		}
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}

// UISceneDetailOperator 获取步骤详情
func UISceneDetailOperator(ctx *gin.Context) {
	var req rao.UISceneDetailOperatorReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	operator, err := uiScene.DetailOperator(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.UISceneDetailOperatorResp{
		Operator: operator,
	})
	return
}

// UISceneOperatorList 场景列表
func UISceneOperatorList(ctx *gin.Context) {
	var req rao.UISceneOperatorListReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	operators, err := uiScene.ListOperator(ctx, req.TeamID, req.SceneID)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.UISceneOperatorListResp{
		Operators: operators,
	})
	return
}

// UIScenesOperatorSort 步骤排序
func UIScenesOperatorSort(ctx *gin.Context) {
	var req rao.UIScenesOperatorStepReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	err := uiScene.OperatorSort(ctx, jwt.GetUserIDByCtx(ctx), &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}

// UISceneCopyOperator 复制步骤
func UISceneCopyOperator(ctx *gin.Context) {
	var req rao.UISceneCopyOperatorReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	err := uiScene.OperatorCopy(ctx, jwt.GetUserIDByCtx(ctx), &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}

// UISceneSetStatusOperator 设置场景状态
func UISceneSetStatusOperator(ctx *gin.Context) {
	var req rao.UISceneSetStatusOperatorReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	err := uiScene.OperatorSetStatus(ctx, jwt.GetUserIDByCtx(ctx), &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}

func UISceneDeleteOperator(ctx *gin.Context) {
	var req rao.UISceneDeleteOperatorReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	err := uiScene.OperatorDelete(ctx, jwt.GetUserIDByCtx(ctx), &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}
