package handler

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/errno"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/jwt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/response"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/runner"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/caseAssemble"
	"github.com/gin-gonic/gin"
)

// GetCaseAssembleList 获取用例集列表
func GetCaseAssembleList(ctx *gin.Context) {
	var req rao.GetCaseAssembleListReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	list, err := caseAssemble.GetCaseAssembleList(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.CaseAssembleListResp{
		CaseAssembleList: list,
		//Total:            total,
	})
	return
}

// CopyCaseAssemble 复制用例
func CopyCaseAssemble(ctx *gin.Context) {
	var req rao.CopyAssembleReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	//caseAssembleDetail, err := caseAssemble.GetCaseAssembleDetail(ctx, &req)
	//if err != nil {
	//	response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
	//}
	//
	//if caseAssembleDetail == nil {
	//	response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
	//	return
	//}

	CopyCaseAssembleErr := caseAssemble.CopyCaseAssemble(ctx, &req)
	if CopyCaseAssembleErr != nil {
		if CopyCaseAssembleErr.Error() == "名称过长！不可超出30字符" {
			response.ErrorWithMsg(ctx, errno.ErrNameOverLength, CopyCaseAssembleErr.Error())
		} else {
			response.ErrorWithMsg(ctx, errno.ErrParam, CopyCaseAssembleErr.Error())
		}

		return
	}

	response.SuccessWithData(ctx, req)
	return
}

// SaveCaseAssemble 新建用例
func SaveCaseAssemble(ctx *gin.Context) {
	var req rao.SaveCaseAssembleReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	//判断用例名称在同一场景下是否已存在
	sceneCaseNameIsExist, _ := caseAssemble.SceneCaseNameIsExist(ctx, &req)
	if sceneCaseNameIsExist == true {
		response.ErrorWithMsg(ctx, errno.ErrSceneCaseNameIsExist, "")
		return
	}
	saveCaseAssembleErr := caseAssemble.SaveCaseAssemble(ctx, &req)
	if saveCaseAssembleErr != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, saveCaseAssembleErr.Error())
		return
	}
	response.SuccessWithData(ctx, req)
	return
}

// DelCaseAssemble 删除用例
func DelCaseAssemble(ctx *gin.Context) {
	var req rao.DelCaseAssembleReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	DelCaseAssembleErr := caseAssemble.DelCaseAssemble(ctx, &req)
	if DelCaseAssembleErr != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, DelCaseAssembleErr.Error())
		return
	}

	response.Success(ctx)
	return
}

// ChangeCaseAssembleCheck 开启/关闭用例
func ChangeCaseAssembleCheck(ctx *gin.Context) {

	var req rao.ChangeCaseAssembleCheckReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	ChangeCaseAssembleCheckErr := caseAssemble.ChangeCaseAssembleCheck(ctx, &req)
	if ChangeCaseAssembleCheckErr != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, ChangeCaseAssembleCheckErr.Error())
		return
	}

	response.Success(ctx)
	return
}

// SaveSceneCaseFlow 新建用例执行流
func SaveSceneCaseFlow(ctx *gin.Context) {

	var req rao.SaveSceneCaseFlowReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	SaveSceneCaseFlowErr := caseAssemble.SaveSceneCaseFlow(ctx, &req)
	if SaveSceneCaseFlowErr != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, SaveSceneCaseFlowErr.Error())
		return
	}

	response.Success(ctx)
	return
}

// GetSceneCaseFlow 获取用例执行流
func GetSceneCaseFlow(ctx *gin.Context) {
	var req rao.GetSceneCaseFlowReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	resp, GetSceneCaseFlowErr := caseAssemble.GetSceneCaseFlow(ctx, &req)
	if GetSceneCaseFlowErr != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, GetSceneCaseFlowErr.Error())
		return
	}
	response.SuccessWithData(ctx, resp)
	return
}

// SendSceneCase 调试场景用例
func SendSceneCase(ctx *gin.Context) {
	var req rao.SendSceneCaseReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	retID, err := caseAssemble.SendSceneCase(ctx, req.TeamID, req.SceneID, req.SceneCaseID, jwt.GetUserIDByCtx(ctx))
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.SendSceneResp{RetID: retID})
	return
}

// StopSceneCase 停止调试场景用例
func StopSceneCase(ctx *gin.Context) {
	var req rao.StopSceneCaseReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	err := runner.StopSceneCase(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrHttpFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}

func ChangeCaseSort(ctx *gin.Context) {
	var req rao.ChangeCaseSortReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	err := runner.ChangeCaseSort(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrHttpFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}
