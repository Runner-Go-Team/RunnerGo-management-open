// Package errno 定义所有错误码
package errno

const (
	Ok                                        = 0
	ErrParam                                  = 10001
	ErrServer                                 = 10002
	ErrNonce                                  = 10003
	ErrTimeStamp                              = 10004
	ErrRPCFailed                              = 10005
	ErrInvalidToken                           = 10006
	ErrMarshalFailed                          = 10007
	ErrUnMarshalFailed                        = 10008
	ErrOperationFail                          = 10009
	ErrMustDID                                = 10011
	ErrMustSN                                 = 10012
	ErrHttpFailed                             = 10013
	ErrRedisFailed                            = 10100
	ErrMongoFailed                            = 10101
	ErrMysqlFailed                            = 10102
	ErrRecordNotFound                         = 10103
	ErrSignError                              = 20001
	ErrRepeatRequest                          = 20002
	ErrMustLogin                              = 20003
	ErrAuthFailed                             = 20004
	ErrYetRegister                            = 20005
	ErrURLExpired                             = 20006
	ErrExistsTeam                             = 20007
	ErrMustTaskInit                           = 20008
	ErrResourceNotEnough                      = 20009
	ErrEmptyScene                             = 20010
	ErrYetPreinstall                          = 20011
	ErrReportNotFound                         = 20012
	ErrInviteCodeFailed                       = 20013
	ErrDefaultTeamFailed                      = 20014
	ErrRecordExists                           = 20015
	ErrEmptyTestCase                          = 20016
	ErrSceneCaseNameIsExist                   = 20017
	ErrApiNameAlreadyExist                    = 20018
	ErrGroupNameAlreadyExist                  = 20019
	ErrFolderNameAlreadyExist                 = 20020
	ErrSceneNameAlreadyExist                  = 20021
	ErrPlanNameAlreadyExist                   = 20022
	ErrEnvNameIsExist                         = 20023
	ErrReportInRun                            = 20024
	ErrMobileYetRegister                      = 20025
	ErrSmsCodeSendIllegal                     = 20026
	ErrSmsCodeVerifyFail                      = 20027
	ErrAuthFailedNotRegistered                = 20030
	ErrSmsCodeSend                            = 20031
	ErrTeamNotExist                           = 20034
	ErrPreinstallNameIsExist                  = 20043
	ErrAddEmailUserNumOvertopLimit            = 20045
	ErrMachineMonitorDataPastDue              = 20047
	ErrInPlanSceneNameAlreadyExist            = 20048
	ErrPlanNameNotEmpty                       = 20049
	ErrInPlanFolderNameAlreadyExist           = 20050
	ErrVerifyFail                             = 20051
	ErrTimedTaskOverdue                       = 20052
	ErrWechatLoginQrCodeOverdue               = 20054
	ErrCannotDeleteRunningPlan                = 20055
	ErrCannotBatchDeleteRunningPlan           = 20056
	ErrMaxConcurrencyLessThanStartConcurrency = 20058
	ErrNotEmailConfig                         = 20059
	ErrEmptySceneFlow                         = 20061
	ErrEmptyTestCaseFlow                      = 20062
	ErrNameOverLength                         = 20063
	ErrTargetSortNameAlreadyExist             = 20064
	ErrEnvNameAlreadyExist                    = 20065
	ErrServiceNameAlreadyExist                = 20066
	ErrExecSqlErr                             = 20067
	ErrCannotDeleteRunningReport              = 20068
	ErrCannotBatchDeleteRunningReport         = 20069
	ErrMockPathExists                         = 20070
	ErrYetAccountRegister                     = 20071
	ErrAccountDel                             = 20072
	ErrNoticeBatchReportLimit                 = 20073
	ErrNoticeConfigError                      = 20074
	ErrMockPathNotNull                        = 20080
)

// CodeAlertMap 错图码映射错误提示，展示给用户
var CodeAlertMap = map[int]string{
	Ok:                              "成功",
	ErrServer:                       "服务器错误",
	ErrParam:                        "参数校验错误",
	ErrSignError:                    "签名错误",
	ErrRepeatRequest:                "重放请求",
	ErrNonce:                        "_nonce参数错误",
	ErrTimeStamp:                    "_timestamp参数错误",
	ErrRecordNotFound:               "数据库记录不存在",
	ErrRPCFailed:                    "请求下游服务失败",
	ErrInvalidToken:                 "无效的token",
	ErrMarshalFailed:                "序列化失败",
	ErrUnMarshalFailed:              "反序列化失败",
	ErrOperationFail:                "操作失败",
	ErrRedisFailed:                  "redis操作失败",
	ErrMongoFailed:                  "mongo操作失败",
	ErrMysqlFailed:                  "mysql操作失败",
	ErrMustLogin:                    "没有获取到登录态",
	ErrMustDID:                      "缺少设备DID信息",
	ErrMustSN:                       "缺少设备SN信息",
	ErrHttpFailed:                   "请求下游Http服务失败",
	ErrAuthFailed:                   "用户名或密码错误",
	ErrYetRegister:                  "用户邮箱已注册",
	ErrURLExpired:                   "邀请链接已过期",
	ErrExistsTeam:                   "用户已在此团队",
	ErrMustTaskInit:                 "请填写任务配置并保存",
	ErrResourceNotEnough:            "资源不足",
	ErrEmptyScene:                   "场景不能为空",
	ErrYetPreinstall:                "预设配置名称已存在",
	ErrReportNotFound:               "报告不存在",
	ErrInviteCodeFailed:             "邀请码验证失败",
	ErrDefaultTeamFailed:            "当前默认团队错误",
	ErrRecordExists:                 "数据库记录已存在",
	ErrEmptyTestCase:                "场景用例不能为空",
	ErrSceneCaseNameIsExist:         "同一场景下用例名称不能重复",
	ErrApiNameAlreadyExist:          "名称已存在",
	ErrGroupNameAlreadyExist:        "分组名称已存在",
	ErrFolderNameAlreadyExist:       "文件夹名称已存在",
	ErrSceneNameAlreadyExist:        "场景名称已存在",
	ErrPlanNameAlreadyExist:         "计划名称已存在",
	ErrEnvNameIsExist:               "环境名称已存在",
	ErrReportInRun:                  "报告数据正在生成中，请稍后再查看",
	ErrMobileYetRegister:            "手机号已注册",
	ErrSmsCodeSendIllegal:           "验证码发送不合法",
	ErrSmsCodeVerifyFail:            "验证码不正确",
	ErrAuthFailedNotRegistered:      "账号未注册",
	ErrSmsCodeSend:                  "短信验证码非法操作",
	ErrTeamNotExist:                 "团队人数已经超过上限",
	ErrPreinstallNameIsExist:        "预设配置名称已存在",
	ErrAddEmailUserNumOvertopLimit:  "单次只可添加1-50个收件人进行发送",
	ErrMachineMonitorDataPastDue:    "只能查询15天以内的压力机监控数据",
	ErrInPlanSceneNameAlreadyExist:  "计划内场景不可重名",
	ErrPlanNameNotEmpty:             "计划名称不能为空",
	ErrInPlanFolderNameAlreadyExist: "计划内目录不可重名",
	ErrVerifyFail:                   "验证失败",
	ErrTimedTaskOverdue:             "开始或结束时间不能早于当前时间",
	ErrWechatLoginQrCodeOverdue:     "当前微信二维码过期",
	ErrCannotDeleteRunningPlan:      "该计划正在运行，无法删除",
	ErrCannotBatchDeleteRunningPlan: "存在运行中的计划，无法删除",
	ErrMaxConcurrencyLessThanStartConcurrency: "最大并发数不能小于起始并发数",
	ErrNotEmailConfig:                         "请配置邮件相关环境变量",
	ErrEmptySceneFlow:                         "场景flow不能为空",
	ErrEmptyTestCaseFlow:                      "场景用例flow不能为空",
	ErrNameOverLength:                         "名称过长！不可超出30字符",
	ErrTargetSortNameAlreadyExist:             "存在重名，无法操作",
	ErrEnvNameAlreadyExist:                    "环境名称已存在",
	ErrServiceNameAlreadyExist:                "服务名称已存在",
	ErrExecSqlErr:                             "执行sql语句失败",
	ErrCannotDeleteRunningReport:              "运行中的报告不能删除",
	ErrCannotBatchDeleteRunningReport:         "存在运行中的报告，无法删除",
	ErrMockPathExists:                         "mock 路径已存在，不能重复",
	ErrYetAccountRegister:                     "用户账户已注册",
	ErrAccountDel:                             "用户不存在或已删除",
	ErrMockPathNotNull:                        "路径不能为空",
	ErrNoticeBatchReportLimit:                 "批量操作最多支持选择10条数据",
	ErrNoticeConfigError:                      "三方配置有误，请检查",
}

// CodeMsgMap 错误码映射错误信息，不展示给用户
var CodeMsgMap = map[int]string{
	Ok:                              "success",
	ErrServer:                       "internal server error",
	ErrParam:                        "param error",
	ErrSignError:                    "signature error",
	ErrRepeatRequest:                "repeat request",
	ErrNonce:                        "nonce error",
	ErrTimeStamp:                    "timestamp error",
	ErrRecordNotFound:               "record not found",
	ErrRPCFailed:                    "rpc failed",
	ErrInvalidToken:                 "invalid token",
	ErrMarshalFailed:                "marshal failed",
	ErrUnMarshalFailed:              "unmarshal failed",
	ErrOperationFail:                "ErrOperationFail",
	ErrRedisFailed:                  "redis operate failed",
	ErrMongoFailed:                  "mongo operate failed",
	ErrMysqlFailed:                  "mysql operate failed",
	ErrMustLogin:                    "must login",
	ErrMustDID:                      "must DID",
	ErrMustSN:                       "must SN",
	ErrHttpFailed:                   "http failed",
	ErrAuthFailed:                   "username/password failed",
	ErrYetRegister:                  "email yet register",
	ErrURLExpired:                   "invite url expired",
	ErrExistsTeam:                   "invite user exists team",
	ErrMustTaskInit:                 "fill in the task allocation and save it",
	ErrResourceNotEnough:            "resource not enough",
	ErrEmptyScene:                   "the scene cannot be empty",
	ErrYetPreinstall:                "preinstall yet exists",
	ErrReportNotFound:               "report not found",
	ErrInviteCodeFailed:             "invite code failed",
	ErrDefaultTeamFailed:            "default team failed",
	ErrRecordExists:                 "record exists",
	ErrEmptyTestCase:                "scenario cases cannot be empty",
	ErrSceneCaseNameIsExist:         "scene case name is exist",
	ErrApiNameAlreadyExist:          "ErrApiNameAlreadyExist",
	ErrGroupNameAlreadyExist:        "group name already exist",
	ErrFolderNameAlreadyExist:       "folder name already exist",
	ErrSceneNameAlreadyExist:        "scene name already exist",
	ErrPlanNameAlreadyExist:         "plan name already exist",
	ErrEnvNameIsExist:               "environment name is exist",
	ErrReportInRun:                  "report in run",
	ErrMobileYetRegister:            "mobile yet register",
	ErrSmsCodeSendIllegal:           "ErrSmsCodeSendIllegal",
	ErrSmsCodeVerifyFail:            "ErrSmsCodeVerifyFail",
	ErrAuthFailedNotRegistered:      "account not registered",
	ErrSmsCodeSend:                  "ErrSmsCodeSend",
	ErrTeamNotExist:                 "ErrTeamNotExist",
	ErrPreinstallNameIsExist:        "preinstall name is exist",
	ErrAddEmailUserNumOvertopLimit:  "ErrAddEmailUserNumOvertopLimit",
	ErrMachineMonitorDataPastDue:    "ErrMachineMonitorDataPastDue",
	ErrInPlanSceneNameAlreadyExist:  "ErrInPlanSceneNameAlreadyExist",
	ErrPlanNameNotEmpty:             "ErrPlanNameNotEmpty",
	ErrInPlanFolderNameAlreadyExist: "ErrInPlanFolderNameAlreadyExist",
	ErrVerifyFail:                   "ErrVerifyFail",
	ErrTimedTaskOverdue:             "ErrTimedTaskOverdue",
	ErrWechatLoginQrCodeOverdue:     "ErrWechatLoginQrCodeOverdue",
	ErrCannotDeleteRunningPlan:      "ErrCannotDeleteRunningPlan",
	ErrCannotBatchDeleteRunningPlan: "ErrCannotBatchDeleteRunningPlan",
	ErrMaxConcurrencyLessThanStartConcurrency: "ErrMaxConcurrencyLessThanStartConcurrency",
	ErrNotEmailConfig:                         "ErrNotEmailConfig",
	ErrEmptySceneFlow:                         "ErrEmptySceneFlow",
	ErrEmptyTestCaseFlow:                      "ErrEmptyTestCaseFlow",
	ErrNameOverLength:                         "ErrNameOverLength",
	ErrTargetSortNameAlreadyExist:             "ErrTargetSortNameAlreadyExist",
	ErrEnvNameAlreadyExist:                    "ErrEnvNameAlreadyExist",
	ErrServiceNameAlreadyExist:                "ErrServiceNameAlreadyExist",
	ErrExecSqlErr:                             "ErrExecSqlErr",
	ErrCannotDeleteRunningReport:              "ErrCannotDeleteRunningReport",
	ErrCannotBatchDeleteRunningReport:         "ErrCannotBatchDeleteRunningReport",
	ErrMockPathExists:                         "ErrMockPathExists",
	ErrYetAccountRegister:                     "ErrYetAccountRegister",
	ErrAccountDel:                             "ErrAccountDel",
	ErrMockPathNotNull:                        "ErrMockPathNotNull",
	ErrNoticeBatchReportLimit:                 "ErrNoticeBatchReportLimit",
	ErrNoticeConfigError:                      "ErrNoticeConfigError",
}
