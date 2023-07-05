package handler

import (
	"encoding/json"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/jwt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/log"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/mail"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/packer"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"math"
	"strings"

	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/errno"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/response"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/report"
)

// ListReports 测试报告列表
func ListReports(ctx *gin.Context) {
	var req rao.ListReportsReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	isExist := strings.Index(req.Keyword, "%")
	if isExist >= 0 {
		response.SuccessWithData(ctx, rao.ListReportsResp{})
		return
	}

	reports, total, err := report.ListByTeamID2(ctx, req.TeamID, req.Size, (req.Page-1)*req.Size,
		req.Keyword, req.StartTimeSec, req.EndTimeSec, req.TaskType, req.TaskMode, req.Status, req.Sort)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.ListReportsResp{
		Reports: reports,
		Total:   total,
	})
	return
}

// DeleteReport 删除报告
func DeleteReport(ctx *gin.Context) {
	var req rao.DeleteReportReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	err := report.DeleteReport(ctx, &req, jwt.GetUserIDByCtx(ctx))
	if err != nil {
		if err.Error() == "运行中的报告不能删除" {
			response.ErrorWithMsg(ctx, errno.ErrCannotDeleteRunningReport, err.Error())
		} else {
			response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		}
		return
	}

	response.Success(ctx)
	return
}

// ReportDetail 报告详情
func ReportDetail(ctx *gin.Context) {
	var req rao.GetReportReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	result, err := report.GetReportDetail(ctx, req)
	if err != nil {
		if err.Error() == "报告不存在" {
			response.ErrorWithMsg(ctx, errno.ErrReportNotFound, err.Error())
		} else {
			response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		}
		return
	}
	response.SuccessWithData(ctx, result)
	return
}

// GetReportTaskDetail 获取报告任务详情
func GetReportTaskDetail(ctx *gin.Context) {
	var req rao.GetReportTaskDetailReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	ret, err := report.GetTaskDetail(ctx, req)
	if err != nil {
		if err.Error() == "报告不存在" {
			response.ErrorWithMsg(ctx, errno.ErrReportNotFound, err.Error())
		} else {
			response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		}
		return
	}

	response.SuccessWithData(ctx, rao.GetReportTaskDetailResp{Report: ret})
	return
}

// DebugDetail 查询报告debug状态
func DebugDetail(ctx *gin.Context) {
	var req rao.GetReportReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	result := report.GetReportDebugStatus(ctx, req)
	response.SuccessWithData(ctx, result)
}

// GetDebug 获取debug日志
func GetDebug(ctx *gin.Context) {
	var req rao.GetReportReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	result := report.GetReportDebugLog(ctx, req)
	response.SuccessWithData(ctx, result)
}

// DebugSetting 开启debug模式
func DebugSetting(ctx *gin.Context) {
	var req rao.DebugSettingReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectDebugStatus)
	filter := bson.D{{"report_id", req.ReportID}, {"team_id", req.TeamID}, {"plan_id", req.PlanID}}
	singleResult := collection.FindOne(ctx, filter)
	result, err := singleResult.DecodeBytes()
	if err != nil {
		debug := bson.D{{"report_id", req.ReportID}, {"team_id", req.TeamID}, {"plan_id", req.PlanID}, {"debug", req.Setting}}
		_, err = collection.InsertOne(ctx, debug)
		if err != nil {
			response.ErrorWithMsg(ctx, errno.ErrMongoFailed, err.Error())
			return
		}
	} else {
		_, err = result.Elements()
		if err != nil {
			//debug := bson.D{{"report_id", req.ReportID}, {"debug", req.Setting}}
			debug := bson.D{{"$set", bson.D{{"report_id", req.ReportID}, {"team_id", req.TeamID}, {"plan_id", req.PlanID}, {"debug", req.Setting}}}}
			_, err = collection.InsertOne(ctx, debug)
			if err != nil {
				response.ErrorWithMsg(ctx, errno.ErrMongoFailed, err.Error())
				return
			}
		} else {
			//debug := bson.D{{"report_id", req.ReportID}, {"debug", req.Setting}}
			debug := bson.D{{"$set", bson.D{{"report_id", req.ReportID}, {"team_id", req.TeamID}, {"plan_id", req.PlanID}, {"debug", req.Setting}}}}
			_, err = collection.UpdateOne(ctx, filter, debug)
			if err != nil {
				response.ErrorWithMsg(ctx, errno.ErrMongoFailed, err.Error())
				return
			}
		}

	}

	// 发送debug状态变更消息
	statusChangeKey := consts.SubscriptionStressPlanStatusChange + req.ReportID
	statusChangeKeyValue := rao.SubscriptionStressPlanStatusChange{
		Type:  2,
		Debug: req.Setting,
	}
	statusChangeValueString, err := json.Marshal(statusChangeKeyValue)
	if err == nil {
		// 发送计划相关信息到redis频道
		_, err = dal.GetRDB().Publish(ctx, statusChangeKey, string(statusChangeValueString)).Result()
		if err != nil {
			log.Logger.Info("设置debug--发送压测计划状态变更到对应频道失败")
		}
	} else {
		log.Logger.Info("设置debug--发送压测计划状态变更到对应频道，压缩数据失败")
	}

	response.Success(ctx)
	return
}

// ListMachines 施压机列表
func ListMachines(ctx *gin.Context) {
	var req rao.ListMachineReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	resp, err := report.ListMachines(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}
	response.SuccessWithData(ctx, resp)
	return
}

func StopReport(ctx *gin.Context) {
	var req rao.StopReportReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	err := report.StopReport(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrOperationFail, err.Error())
		return
	}
	response.Success(ctx)
	return
}

func ReportEmail(ctx *gin.Context) {
	var req rao.ReportEmailReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	// 单次限制添加50条
	if len(req.Emails) > 50 {
		response.ErrorWithMsg(ctx, errno.ErrAddEmailUserNumOvertopLimit, "单次只可添加1-50个收件人进行发送")
		return
	}

	tx := dal.GetQuery().Team
	team, err := tx.WithContext(ctx).Where(tx.TeamID.Eq(req.TeamID)).First()
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	ux := dal.GetQuery().User
	user, err := ux.WithContext(ctx).Where(ux.UserID.Eq(jwt.GetUserIDByCtx(ctx))).First()
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	rx := dal.GetQuery().StressPlanReport
	reportInfo, err := rx.WithContext(ctx).Where(rx.TeamID.Eq(req.TeamID), rx.ReportID.Eq(req.ReportID)).First()
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	for _, email := range req.Emails {
		if err := mail.SendReportEmail(email, req.ReportID, team, user, reportInfo); err != nil {
			if err.Error() == "请配置邮件相关环境变量" {
				response.ErrorWithMsg(ctx, errno.ErrNotEmailConfig, err.Error())
			} else {
				response.ErrorWithMsg(ctx, errno.ErrHttpFailed, err.Error())
			}
			return
		}
	}

	response.Success(ctx)
	return
}

// ChangeTaskConfRun 报告里面编辑任务配置并执行
func ChangeTaskConfRun(ctx *gin.Context) {
	var req rao.ChangeTaskConfReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	if req.ModeConf.MaxConcurrency < req.ModeConf.StartConcurrency {
		response.ErrorWithMsg(ctx, errno.ErrMaxConcurrencyLessThanStartConcurrency, "")
		return
	}

	// 根据报告id，查询出来机器ip
	rm := dal.GetQuery().ReportMachine
	reportMachineInfo, err := rm.WithContext(ctx).Where(rm.TeamID.Eq(req.TeamID),
		rm.ReportID.Eq(req.ReportID)).Order(rm.CreatedAt.Desc()).First()
	if err != nil {
		log.Logger.Info("编辑报告-查询报告对应的机器失败，err：", err)
		response.ErrorWithMsg(ctx, errno.ErrOperationFail, err.Error()+" 报告对应的机器IP信息没有查到")
		return
	}

	// 查询报告基本信息
	sprTB := dal.GetQuery().StressPlanReport
	reportInfo, err := sprTB.WithContext(ctx).Where(sprTB.ReportID.Eq(req.ReportID)).First()
	if err != nil {
		log.Logger.Info("编辑报告-查询报告基本信息失败，err：", err)
		response.ErrorWithMsg(ctx, errno.ErrOperationFail, err.Error()+" 报告基本信息没有查到")
		return
	}

	machineTB := dal.GetQuery().Machine
	// 发送debug状态变更消息
	machineModeConfArr := make([]rao.MachineModeConf, 0, len(req.MachineDispatchModeConf.UsableMachineList))
	// 判断是否是分布式任务
	if req.IsOpenDistributed == 1 { // 分布式
		// 判断分布式类型
		if req.MachineDispatchModeConf.MachineAllotType == 0 { // 权重
			// 判断压测模式
			if reportInfo.TaskMode == consts.PlanModeConcurrence { // 并发模式
				for _, v := range req.MachineDispatchModeConf.UsableMachineList {
					addrInfo, err := machineTB.WithContext(ctx).Where(machineTB.IP.Eq(v.Ip)).First()
					if err != nil {
						log.Logger.Info("没有查到配置的压力机信息：", " 机器ip为：", v.Ip)
						continue
					} else { // 查到了
						concurrencyNum := int64(math.Ceil(float64(v.Weight) * float64(req.ModeConf.Concurrency) / 100))
						modeConfTemp := rao.ChangeTakeConf{
							RoundNum:    req.ModeConf.RoundNum,
							Duration:    req.ModeConf.Duration,
							Concurrency: concurrencyNum,
						}
						temp := rao.MachineModeConf{
							Machine:  addrInfo.IP,
							ModeConf: modeConfTemp,
						}
						machineModeConfArr = append(machineModeConfArr, temp)
					}
				}
			} else { // 非并发模式
				for _, v := range req.MachineDispatchModeConf.UsableMachineList {
					addrInfo, err := machineTB.WithContext(ctx).Where(machineTB.IP.Eq(v.Ip)).First()
					if err != nil {
						log.Logger.Info("没有查到配置的压力机信息：", " 机器ip为：", v.Ip)
						continue
					} else { // 查到了
						modeConfTemp := rao.ChangeTakeConf{
							StartConcurrency: int64(math.Ceil(float64(v.StartConcurrency) * float64(v.Weight) / 100)),
							Step:             int64(math.Ceil(float64(v.Step) * float64(v.Weight) / 100)),
							StepRunTime:      v.StepRunTime,
							MaxConcurrency:   int64(math.Ceil(float64(v.MaxConcurrency) * float64(v.Weight) / 100)),
							Duration:         v.Duration,
						}
						temp := rao.MachineModeConf{
							Machine:  addrInfo.IP,
							ModeConf: modeConfTemp,
						}
						machineModeConfArr = append(machineModeConfArr, temp)
					}
				}
			}
		} else { // 自定义
			if reportInfo.TaskMode == consts.PlanModeConcurrence { // 并发模式
				for _, v := range req.MachineDispatchModeConf.UsableMachineList {
					addrInfo, err := machineTB.WithContext(ctx).Where(machineTB.IP.Eq(v.Ip)).First()
					if err != nil {
						log.Logger.Info("没有查到配置的压力机信息：", " 机器ip为：", v.Ip)
						continue
					} else { // 查到了
						modeConfTemp := rao.ChangeTakeConf{
							RoundNum:    req.ModeConf.RoundNum,
							Duration:    req.ModeConf.Duration,
							Concurrency: v.Concurrency,
						}
						temp := rao.MachineModeConf{
							Machine:  addrInfo.IP,
							ModeConf: modeConfTemp,
						}
						machineModeConfArr = append(machineModeConfArr, temp)
					}
				}
			} else { // 非并发模式
				for _, v := range req.MachineDispatchModeConf.UsableMachineList {
					addrInfo, err := machineTB.WithContext(ctx).Where(machineTB.IP.Eq(v.Ip)).First()
					if err != nil {
						log.Logger.Info("没有查到配置的压力机信息：", " 机器ip为：", v.Ip)
						continue
					} else { // 查到了
						modeConfTemp := rao.ChangeTakeConf{
							StartConcurrency: v.StartConcurrency,
							Step:             v.Step,
							StepRunTime:      v.StepRunTime,
							MaxConcurrency:   v.MaxConcurrency,
							Duration:         v.Duration,
						}
						temp := rao.MachineModeConf{
							Machine:  addrInfo.IP,
							ModeConf: modeConfTemp,
						}
						machineModeConfArr = append(machineModeConfArr, temp)
					}
				}
			}
		}
	} else { // 智能调度
		// 把新编辑的任务配置保存到redis当中，供压力机执行使用
		value := rao.ChangeTakeConf{
			RoundNum:         req.ModeConf.RoundNum,
			Concurrency:      req.ModeConf.Concurrency,
			StartConcurrency: req.ModeConf.StartConcurrency,
			Step:             req.ModeConf.Step,
			StepRunTime:      req.ModeConf.StepRunTime,
			MaxConcurrency:   req.ModeConf.MaxConcurrency,
			Duration:         req.ModeConf.Duration,
		}
		// 发送debug状态变更消息
		MachineModeConf := rao.MachineModeConf{
			Machine:  reportMachineInfo.IP,
			ModeConf: value,
		}
		machineModeConfArr = append(machineModeConfArr, MachineModeConf)
	}

	// 组装修改的配置数据，保存到mg
	changeReportConf := packer.TransChangeReportConfRunToMao(req)
	// 操作mongodb
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectChangeReportConf)
	_, err = collection.InsertOne(ctx, changeReportConf)
	if err != nil {
		log.Logger.Info("编辑报告保存配置项失败，err：", err)
		response.ErrorWithMsg(ctx, errno.ErrMongoFailed, err.Error())
		return
	}

	// 发送消息
	for _, machineModeConfInfo := range machineModeConfArr {
		statusChangeKey := consts.SubscriptionStressPlanStatusChange + req.ReportID
		statusChangeKeyValue := rao.SubscriptionStressPlanStatusChange{
			Type:            3,
			MachineModeConf: machineModeConfInfo,
		}
		statusChangeValueString, err := json.Marshal(statusChangeKeyValue)
		if err == nil {
			// 发送计划相关信息到redis频道
			_, err = dal.GetRDB().Publish(ctx, statusChangeKey, string(statusChangeValueString)).Result()
			if err != nil {
				log.Logger.Info("编辑报告--发送压测计划状态变更到对应频道失败")
			}
		} else {
			log.Logger.Info("编辑报告--发送压测计划状态变更到对应频道，压缩数据失败")
		}
	}
	log.Logger.Info("编辑报告--发送压测计划状态变更到对应频道成功")
	response.Success(ctx)
	return
}

// CompareReport 对比报告
func CompareReport(ctx *gin.Context) {
	var req rao.CompareReportReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	res, err := report.GetCompareReportData(ctx, req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	response.SuccessWithData(ctx, res)
	return

}

// UpdateDescription 保存或更新测试结果描述
func UpdateDescription(ctx *gin.Context) {
	var req rao.UpdateDescriptionReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	err := report.UpdateDescription(ctx, req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMongoFailed, err.Error())
		return
	}
	response.Success(ctx)
	return
}

// BatchDeleteReport 批量删除报告
func BatchDeleteReport(ctx *gin.Context) {
	var req rao.BatchDeleteReportReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	err := report.BatchDeleteReport(ctx, &req)
	if err != nil {
		if err.Error() == "存在运行中的报告，无法删除" {
			response.ErrorWithMsg(ctx, errno.ErrCannotBatchDeleteRunningReport, err.Error())
		} else {
			response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		}
		return
	}
	response.Success(ctx)
	return
}

func UpdateReportName(ctx *gin.Context) {
	var req rao.UpdateReportNameReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	err := report.UpdateReportName(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}
	response.Success(ctx)
	return
}
