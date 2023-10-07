package handler

import (
	"errors"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/errno"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/jwt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/response"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/errmsg"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/uiScene"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/public"
	"github.com/gin-gonic/gin"
)

// UISceneSaveFolder 创建/修改文件夹
func UISceneSaveFolder(ctx *gin.Context) {
	var req rao.UISceneSaveFolderReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	err := uiScene.FolderReqSave(ctx, jwt.GetUserIDByCtx(ctx), &req)
	if err != nil {
		if errors.Is(err, errmsg.ErrUISceneFolderNameRepeat) {
			response.ErrorWithMsg(ctx, errno.ErrUISceneFolderNameRepeat, "")
			return
		}
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}

// UISceneUpdateFolder 修改文件夹
func UISceneUpdateFolder(ctx *gin.Context) {
	var req rao.UISceneSaveFolderReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	err := uiScene.FolderReqUpdate(ctx, jwt.GetUserIDByCtx(ctx), &req)
	if err != nil {
		if errors.Is(err, errmsg.ErrUISceneFolderNameRepeat) {
			response.ErrorWithMsg(ctx, errno.ErrUISceneFolderNameRepeat, "")
			return
		}
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}

// UISceneGetFolder 获取文件夹
func UISceneGetFolder(ctx *gin.Context) {
	var req rao.UISceneGetFolderReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	f, err := uiScene.GetBySceneID(ctx, req.TeamID, req.SceneID)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.UISceneGetFolderResp{Folder: f})
	return
}

// UISceneSave 保存UI场景
func UISceneSave(ctx *gin.Context) {
	var req rao.UISceneSaveReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	if len(req.Name) == 0 {
		response.ErrorWithMsg(ctx, errno.ErrUISceneNameNotEmpty, "")
		return
	}
	if public.GetStringNum(req.Name) > 30 {
		response.ErrorWithMsg(ctx, errno.ErrUISceneNameLong, "")
		return
	}

	err := uiScene.Save(ctx, jwt.GetUserIDByCtx(ctx), &req)
	if err != nil {
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

// UISceneUpdate 保存UI场景
func UISceneUpdate(ctx *gin.Context) {
	var req rao.UISceneSaveReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	if len(req.Name) == 0 {
		response.ErrorWithMsg(ctx, errno.ErrUISceneNameNotEmpty, "")
		return
	}
	if public.GetStringNum(req.Name) > 30 {
		response.ErrorWithMsg(ctx, errno.ErrUISceneNameLong, "")
		return
	}

	err := uiScene.Update(ctx, jwt.GetUserIDByCtx(ctx), &req)
	if err != nil {
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

// UISceneDetail 获取接口详情
func UISceneDetail(ctx *gin.Context) {
	var req rao.UISceneDetailReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	scene, err := uiScene.DetailBySceneID(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.UISceneDetailResp{
		Scene: scene,
	})
	return
}

// SendUIScene 调试UI场景
func SendUIScene(ctx *gin.Context) {
	var req rao.SendUISceneReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	runID, err := uiScene.Send(ctx, jwt.GetUserIDByCtx(ctx), req.TeamID, req.SceneID, req.PlanID, req.OperatorIDs)
	if err != nil {
		if errors.Is(err, errmsg.ErrSendLinuxNotQTMode) {
			response.ErrorWithMsg(ctx, errno.ErrSendLinuxNotQTMode, "")
			return
		}
		response.ErrorWithMsg(ctx, errno.ErrUIEngineError, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.SendUISceneResp{RunID: runID})
	return
}

// StopUIScene 停止调试场景
func StopUIScene(ctx *gin.Context) {
	var req rao.StopUISceneReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	err := uiScene.StopScene(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrHttpFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}

// UISceneList 场景列表
func UISceneList(ctx *gin.Context) {
	var req rao.UISceneListReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	scenes, err := uiScene.List(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.UISceneListResp{
		Scenes: scenes,
	})
	return
}

func UIScenesSort(ctx *gin.Context) {
	var req rao.UIScenesSortReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	err := uiScene.Sort(ctx, jwt.GetUserIDByCtx(ctx), &req)
	if err != nil {
		if errors.Is(err, errmsg.ErrTargetSortNameAlreadyExist) {
			response.ErrorWithMsg(ctx, errno.ErrTargetSortNameAlreadyExist, "")
			return
		}
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}

// UISceneCopy 复制
func UISceneCopy(ctx *gin.Context) {
	var req rao.UISceneCopyReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	err := uiScene.Copy(ctx, jwt.GetUserIDByCtx(ctx), &req)
	if err != nil {
		if errors.Is(err, errmsg.ErrUISceneNameLong) {
			response.ErrorWithMsg(ctx, errno.ErrUISceneNameLong, "")
			return
		}
		response.ErrorWithMsg(ctx, errno.ErrOperationFail, err.Error())
		return
	}
	response.Success(ctx)
	return
}

// UISceneTrash 移入回收站
func UISceneTrash(ctx *gin.Context) {
	var req rao.UISceneTrashReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	if err := uiScene.Trash(ctx, &req, jwt.GetUserIDByCtx(ctx)); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}

// UISceneTrashList 回收站列表
func UISceneTrashList(ctx *gin.Context) {
	var req rao.UISceneTrashListReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	scenes, total, err := uiScene.TrashList(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.UISceneTrashListResp{
		Total:     total,
		TrashList: scenes,
	})
	return
}

// UISceneRecall 回复回收站
func UISceneRecall(ctx *gin.Context) {
	var req rao.UISceneRecallReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	if err := uiScene.Recall(ctx, req.SceneIDs, req.TeamID, jwt.GetUserIDByCtx(ctx)); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}

// UISceneDelete 删除
func UISceneDelete(ctx *gin.Context) {
	var req rao.UISceneDeleteReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	if err := uiScene.Delete(ctx, req.SceneIDs, req.TeamID, jwt.GetUserIDByCtx(ctx)); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}

// UISceneEngineMachine 获取机器信息
func UISceneEngineMachine(ctx *gin.Context) {
	var req rao.UISceneEngineMachineReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	machineList, err := uiScene.GetUiEngineMachineList(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrUISceneGetMachineError, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.UISceneEngineMachineResp{
		List: machineList,
	})
	return
}
