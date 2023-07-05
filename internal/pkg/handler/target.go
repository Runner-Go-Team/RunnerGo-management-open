package handler

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/errno"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/jwt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/response"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/api"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/target"

	"github.com/gin-gonic/gin"
)

// SendTarget 发送接口
func SendTarget(ctx *gin.Context) {
	var req rao.SendTargetReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	retID, err := target.SendAPI(ctx, req.TeamID, req.TargetID)
	if err != nil {
		if err.Error() == "调试接口返回非200状态" {
			response.ErrorWithMsg(ctx, errno.ErrHttpFailed, err.Error())
		} else {
			response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		}
		return
	}

	response.SuccessWithData(ctx, rao.SendTargetResp{RetID: retID})
	return
}

// GetSendTargetResult 获取发送接口结果
func GetSendTargetResult(ctx *gin.Context) {
	var req rao.GetSendTargetResultReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	r, err := target.GetSendAPIResult(ctx, req.RetID)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, r)
	return
}

// SaveTarget 创建/修改接口
func SaveTarget(ctx *gin.Context) {
	var req rao.SaveTargetReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	targetID := req.TargetID
	var err error

	if req.TargetType == consts.TargetTypeSql { // sql
		targetID, err = api.SaveSql(ctx, &req)
		if err != nil {
			if err.Error() == "名称已存在" {
				response.ErrorWithMsg(ctx, errno.ErrApiNameAlreadyExist, err.Error())
			} else {
				response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
			}
			return
		}
	} else if req.TargetType == consts.TargetTypeTcp { // tcp
		targetID, err = api.SaveTcp(ctx, &req)
		if err != nil {
			if err.Error() == "名称已存在" {
				response.ErrorWithMsg(ctx, errno.ErrApiNameAlreadyExist, err.Error())
			} else {
				response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
			}
			return
		}
	} else if req.TargetType == consts.TargetTypeWebsocket { // websocket
		targetID, err = api.SaveWebsocket(ctx, &req)
		if err != nil {
			if err.Error() == "名称已存在" {
				response.ErrorWithMsg(ctx, errno.ErrApiNameAlreadyExist, err.Error())
			} else {
				response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
			}
			return
		}
	} else if req.TargetType == consts.TargetTypeMQTT { // Mqtt
		targetID, err = api.SaveMQTT(ctx, &req)
		if err != nil {
			if err.Error() == "名称已存在" {
				response.ErrorWithMsg(ctx, errno.ErrApiNameAlreadyExist, err.Error())
			} else {
				response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
			}
			return
		}
	} else if req.TargetType == consts.TargetTypeDubbo { // Dubbo
		targetID, err = api.SaveDubbo(ctx, &req)
		if err != nil {
			if err.Error() == "名称已存在" {
				response.ErrorWithMsg(ctx, errno.ErrApiNameAlreadyExist, err.Error())
			} else {
				response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
			}
			return
		}
	} else { // api
		targetID, err = api.Save(ctx, &req, jwt.GetUserIDByCtx(ctx))
		if err != nil {
			if err.Error() == "名称已存在" {
				response.ErrorWithMsg(ctx, errno.ErrApiNameAlreadyExist, err.Error())
			} else {
				response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
			}
			return
		}
	}

	response.SuccessWithData(ctx, rao.SaveTargetResp{TargetID: targetID})
	return
}

// SaveImportApi 导入接口
func SaveImportApi(ctx *gin.Context) {
	var req rao.SaveImportApiReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	err := api.SaveImportApi(ctx, &req, jwt.GetUserIDByCtx(ctx))
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}
	response.Success(ctx)
	return
}

// SortTarget 排序
func SortTarget(ctx *gin.Context) {
	var req rao.SortTargetReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	err := target.SortTarget(ctx, &req)
	if err != nil {
		if err.Error() == "存在重名，无法操作" {
			response.ErrorWithMsg(ctx, errno.ErrTargetSortNameAlreadyExist, err.Error())
		} else {
			response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		}
		return
	}

	response.Success(ctx)
	return
}

// TrashTargetList 文件夹/接口回收站列表
func TrashTargetList(ctx *gin.Context) {
	var req rao.ListTrashTargetReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	targets, total, err := target.ListTrashFolderAPI(ctx, req.TeamID, req.Size, (req.Page-1)*req.Size)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.ListTrashTargetResp{
		Targets: targets,
		Total:   total,
	})
	return
}

// TrashTarget 移入回收站
func TrashTarget(ctx *gin.Context) {
	var req rao.DeleteTargetReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	if err := target.Trash(ctx, req.TargetID, jwt.GetUserIDByCtx(ctx)); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}

// RecallTarget 从回收站恢复
func RecallTarget(ctx *gin.Context) {
	var req rao.RecallTargetReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	err := target.Recall(ctx, req.TargetID)
	if err != nil {
		if err.Error() == "文件夹名称已存在" {
			response.ErrorWithMsg(ctx, errno.ErrFolderNameAlreadyExist, err.Error())
		} else if err.Error() == "接口名称已存在" {
			response.ErrorWithMsg(ctx, errno.ErrApiNameAlreadyExist, err.Error())
		} else {
			response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		}
		return
	}

	response.Success(ctx)
	return
}

// DeleteTarget 回收站彻底删除
func DeleteTarget(ctx *gin.Context) {
	var req rao.DeleteTargetReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	if err := target.Delete(ctx, req.TargetID); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}

// ListFolderAPI 文件夹/接口列表
func ListFolderAPI(ctx *gin.Context) {
	var req rao.ListFolderAPIReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	targets, err := target.ListFolderAPI(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.ListFolderAPIResp{
		Targets: targets,
	})
	return
}

// ListGroupScene 分组/场景列表
func ListGroupScene(ctx *gin.Context) {
	var req rao.ListGroupSceneReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	targets, err := target.ListGroupScene(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.ListGroupSceneResp{
		Targets: targets,
	})
	return
}

// BatchGetTarget 获取接口详情
func BatchGetTarget(ctx *gin.Context) {
	var req rao.BatchGetDetailReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	targets, err := api.DetailByTargetIDs(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.BatchGetDetailResp{
		Targets: targets,
	})
	return
}

func GetSqlDatabaseList(ctx *gin.Context) {
	var req rao.GetSqlDatabaseListReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	dbList, err := api.GetSqlDatabaseList(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}
	response.SuccessWithData(ctx, dbList)
	return
}

func SendSql(ctx *gin.Context) {
	var req rao.SendSqlReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	retID, err := target.SendSql(ctx, &req)
	if err != nil {
		if err.Error() == "调试接口返回非200状态" {
			response.ErrorWithMsg(ctx, errno.ErrHttpFailed, err.Error())
		} else {
			response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		}
		return
	}

	response.SuccessWithData(ctx, rao.SendSqlResp{RetID: retID})
	return
}

func ConnectionDatabase(ctx *gin.Context) {
	var req rao.ConnectionDatabaseReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	res, err := target.ConnectionDatabase(&req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrHttpFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, res)
	return
}

func GetSendSqlResult(ctx *gin.Context) {
	var req rao.GetSendSqlResultReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	res, err, sqlErr := target.GetSendSqlResult(ctx, &req)
	if err != nil {
		if err.Error() == "sql执行失败" {
			response.ErrorWithMsg(ctx, errno.ErrExecSqlErr, sqlErr)
		} else {
			response.ErrorWithMsg(ctx, errno.ErrMongoFailed, err.Error())
		}
		return
	}
	response.SuccessWithData(ctx, res)
	return
}

func SendTcp(ctx *gin.Context) {
	var req rao.SendTcpReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	retID, err := target.SendTcp(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMongoFailed, err.Error())
		return
	}
	response.SuccessWithData(ctx, rao.SendTcpResp{
		RetID: retID,
	})
	return
}

func GetSendTcpResult(ctx *gin.Context) {
	var req rao.GetSendTcpResultReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	res, err := target.GetSendTcpResult(ctx, &req)
	if err != nil {
		response.ErrorWithMsgAndData(ctx, errno.ErrMongoFailed, err.Error(), res)
		return
	}
	response.SuccessWithData(ctx, res)
	return
}

func TcpSendOrStopMessage(ctx *gin.Context) {
	var req rao.TcpSendOrStopMessageReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	err := target.TcpSendOrStopMessage(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrOperationFail, err.Error())
		return
	}
	response.Success(ctx)
	return
}

func SendWebsocket(ctx *gin.Context) {
	var req rao.SendWebsocketReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	retID, err := target.SendWebsocket(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMongoFailed, err.Error())
		return
	}
	response.SuccessWithData(ctx, rao.SendTcpResp{
		RetID: retID,
	})
	return
}

func GetSendWebsocketResult(ctx *gin.Context) {
	var req rao.GetSendWebsocketResultReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	res, err := target.GetSendWebsocketResult(ctx, &req)
	if err != nil {
		response.ErrorWithMsgAndData(ctx, errno.ErrMongoFailed, err.Error(), res)
		return
	}
	response.SuccessWithData(ctx, res)
	return
}

func WsSendOrStopMessage(ctx *gin.Context) {
	var req rao.WsSendOrStopMessageReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	err := target.WsSendOrStopMessage(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMongoFailed, err.Error())
		return
	}
	response.Success(ctx)
	return
}

func SendDubbo(ctx *gin.Context) {
	var req rao.SendDubboReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	retID, err := target.SendDubbo(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrOperationFail, err.Error())
		return
	}
	response.SuccessWithData(ctx, rao.SendTcpResp{
		RetID: retID,
	})
	return
}

func GetSendDubboResult(ctx *gin.Context) {
	var req rao.GetSendDubboResultReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	res, err := target.GetSendDubboResult(ctx, &req)
	if err != nil {
		response.ErrorWithMsgAndData(ctx, errno.ErrMongoFailed, err.Error(), res)
		return
	}
	response.SuccessWithData(ctx, res)
	return
}

func SendMqtt(ctx *gin.Context) {
	var req rao.SendMqttReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	retID, err := target.SendMqtt(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrOperationFail, err.Error())
		return
	}
	response.SuccessWithData(ctx, rao.SendTcpResp{
		RetID: retID,
	})
	return
}

func GetSendMqttResult(ctx *gin.Context) {
	var req rao.GetSendMqttResultReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	res, err := target.GetSendMqttResult(ctx, &req)
	if err != nil {
		response.ErrorWithMsgAndData(ctx, errno.ErrMongoFailed, err.Error(), res)
		return
	}
	response.SuccessWithData(ctx, res)
	return
}
