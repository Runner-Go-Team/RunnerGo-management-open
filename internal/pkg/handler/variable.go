package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/errno"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/jwt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/response"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/variable"
)

// SaveVariable 保存变量
func SaveVariable(ctx *gin.Context) {
	var req rao.SaveVariableReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	if req.Status == 0 {
		req.Status = 1
	}

	if err := variable.SaveVariable(ctx, &req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}

// DeleteVariable 删除变量
func DeleteVariable(ctx *gin.Context) {
	var req rao.DeleteVariableReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	if err := variable.DeleteVariable(ctx, req.TeamID, req.VarID); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}

// ListGlobalVariables 变量列表
func ListGlobalVariables(ctx *gin.Context) {
	var req rao.ListVariablesReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	v, cnt, err := variable.ListGlobalVariables(ctx, req.TeamID, req.Size, (req.Page-1)*req.Size)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.ListVariablesResp{Variables: v, Total: cnt})
	return
}

// SyncGlobalVariables 同步变量
func SyncGlobalVariables(ctx *gin.Context) {
	var req rao.SyncVariablesReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	//variables := make([]*rao.Variable, 0)
	//for _, v := range req.Variables {
	//	if v.Var != "" && v.Val != "" {
	//		variables = append(variables, v)
	//	}
	//}

	if err := variable.SyncGlobalVariables(ctx, req.TeamID, req.Variables); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}

// ListSceneVariables 场景变量列表
func ListSceneVariables(ctx *gin.Context) {
	var req rao.ListSceneVariablesReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	v, cnt, err := variable.ListSceneVariables(ctx, req.TeamID, req.SceneID, req.Size, (req.Page-1)*req.Size)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.ListVariablesResp{Variables: v, Total: cnt})
	return
}

// SyncSceneVariables 同步场景变量
func SyncSceneVariables(ctx *gin.Context) {
	var req rao.SyncSceneVariablesReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	if err := variable.SyncSceneVariables(ctx, req.TeamID, req.SceneID, req.Variables); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}

// ImportSceneVariables 导入场景变量
func ImportSceneVariables(ctx *gin.Context) {
	var req rao.ImportVariablesReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	if err := variable.ImportSceneVariables(ctx, &req, jwt.GetUserIDByCtx(ctx)); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}

// DeleteImportSceneVariables 删除导入场景变量
func DeleteImportSceneVariables(ctx *gin.Context) {
	var req rao.DeleteImportSceneVariablesReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	if err := variable.DeleteImportSceneVariables(ctx, &req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}

// ListImportSceneVariables 导入场景变量列表
func ListImportSceneVariables(ctx *gin.Context) {
	var req rao.ListImportVariablesReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	i, err := variable.ListImportSceneVariables(ctx, req.TeamID, req.SceneID)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.ListImportVariablesResp{
		Imports: i,
	})
	return
}

// UpdateImportSceneVariables 编辑导入变量开关
func UpdateImportSceneVariables(ctx *gin.Context) {
	var req rao.UpdateImportSceneVariablesReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	if err := variable.UpdateImportSceneVariables(ctx, &req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}
	response.Success(ctx)
	return
}
