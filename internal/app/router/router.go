package router

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/go-omnibus/proof"

	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/app/middleware"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/handler"
)

func RegisterRouter(r *gin.Engine) {
	// cors
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"POST", "GET", "OPTIONS", "DELETE", "PUT", "PATCH"},
		AllowHeaders:     []string{"Authorization", "Content-Type", "Upgrade", "Origin", "Connection", "Accept-Encoding", "Accept-Language", "Host", "x-requested-with", "CurrentTeamID"},
		ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.Use(ginzap.Ginzap(proof.Logger.Z, time.RFC3339, true))

	r.Use(ginzap.RecoveryWithZap(proof.Logger.Z, true))

	r.Use(middleware.RecoverPanic()) // 恢复因接口内部错误导致的panic

	// 探活接口
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// 独立报告页面接口
	html := r.Group("/html/api/v1/report")
	html.GET("/debug", handler.GetDebug)
	html.GET("/debug/detail", handler.DebugDetail)
	html.GET("/detail", handler.ReportDetail)
	html.GET("/machine", handler.ListMachines)
	html.GET("/task_detail", handler.GetReportTaskDetail)

	// 独立报告页面接口
	htmlAutoPlan := r.Group("/html/api/v1/auto_plan")
	htmlAutoPlan.POST("get_report_detail", handler.GetAutoPlanReportDetail)

	// routers
	api := r.Group("/management/api")
	api.POST("/v1/plan/notify_stop_stress", handler.NotifyStopStress)
	api.POST("/v1/auto_plan/notify_run_finish", handler.NotifyRunFinish)

	// 邀请链接相关接口
	invite := api.Group("/v1/team/")
	invite.POST("get_invite_user_info", handler.GetInviteUserInfo)

	// 用户鉴权
	auth := api.Group("/v1/auth/")
	auth.POST("user_register", handler.UserRegister)
	auth.POST("user_login", handler.AuthLogin)          //手机号和邮箱密码登录二合一
	auth.POST("/mobile_login", handler.MobileAuthLogin) //手机号密码登录
	auth.POST("/refresh_token", handler.RefreshToken)
	auth.POST("/get_sms_code", handler.GetSmsCode)                     //获取短信验证码
	auth.POST("/verify_sms_code", handler.VerifySmsCode)               //校验短信验证码
	auth.POST("/sms_code_login", handler.SmsCodeLogin)                 //手机号验证码登录
	auth.POST("check_email_is_register", handler.CheckEmailIsRegister) //检查邮箱是否注册
	auth.POST("forget_password", handler.AuthForgetPassword)
	auth.POST("reset_password", handler.AuthResetPassword)

	// 新注册登录相关接口
	auth.POST("check_user_is_register", handler.CheckUserIsRegister) // 检查手机号或邮箱是否注册
	auth.POST("set_user_password", handler.SetUserPassword)          // 设置用户密码

	// 微信登录相关接口
	auth.POST("get_wechat_login_qr_code", handler.GetWechatLoginQrCode)       // 获取微信登录二维码
	auth.POST("get_wechat_login_result", handler.GetWechatLoginResult)        // 获取微信登录结果
	auth.POST("check_wechat_is_change_bind", handler.CheckWechatIsChangeBind) // 检查当前用户是否需要换绑微信

	// 开启接口鉴权
	api.Use(middleware.JWT())

	user := api.Group("/v1/user")
	user.POST("/update_password", handler.UpdatePassword)
	user.POST("/update_nickname", handler.UpdateNickname)
	user.POST("/update_avatar", handler.UpdateAvatar)
	user.POST("/verify_password", handler.VerifyPassword)
	user.POST("collect_user_info", handler.CollectUserInfo)        // 用户信息收集
	user.POST("get_collect_user_info", handler.GetCollectUserInfo) // 判断是否需要手机用户信息
	user.POST("update_email", handler.UpdateEmail)                 // 修改用户邮箱
	user.POST("update_account", handler.UpdateAccount)

	// 用户配置
	setting := api.Group("/v1/setting")
	setting.GET("/get", handler.GetUserSettings)
	setting.POST("/set", handler.SetUserSettings)

	// 团队
	team := api.Group("/v1/team")
	team.POST("/save", handler.SaveTeam)
	team.GET("/list", handler.ListTeam)
	team.GET("/members", handler.ListTeamMembers)
	team.POST("/invite", handler.InviteMember)
	team.GET("/invite/url", handler.GetInviteMemberURL)
	team.POST("/invite/url", handler.CheckInviteMemberURL) // 目前已弃用
	team.POST("/role", handler.SetUserTeamRole)
	team.GET("/role", handler.GetUserTeamRole)
	team.POST("/remove", handler.RemoveMember)
	team.POST("/quit", handler.QuitTeam)
	team.POST("/disband", handler.DisbandTeam)
	team.POST("/transfer", handler.TransferTeam)
	team.POST("/invite/login", handler.InviteLogin)
	team.POST("get_invite_email_is_exist", handler.GetInviteEmailIsExist) // 查询当前邀请邮箱是否存在

	// 全局变量
	variable := api.Group("/v1/variable")
	variable.POST("/save", handler.SaveVariable)
	variable.POST("/delete", handler.DeleteVariable)
	variable.POST("/sync", handler.SyncGlobalVariables)
	variable.GET("/list", handler.ListGlobalVariables)
	// 场景变量
	variable.POST("/scene/sync", handler.SyncSceneVariables)
	variable.GET("/scene/list", handler.ListSceneVariables)
	variable.POST("/scene/import", handler.ImportSceneVariables)
	variable.POST("/scene/import/delete", handler.DeleteImportSceneVariables)
	variable.GET("/scene/import/list", handler.ListImportSceneVariables)
	variable.POST("/scene/import/update", handler.UpdateImportSceneVariables)

	// 首页
	dashboard := api.Group("/v1/dashboard/")
	dashboard.GET("default", handler.DashboardDefault)
	dashboard.POST("home_page", handler.HomePage)
	dashboard.GET("underway_plans", handler.ListUnderwayPlan)
	dashboard.GET("public_function_list", handler.GetPublicFunctionList)

	// 文件夹
	folder := api.Group("/v1/folder")
	folder.POST("/save", handler.SaveFolder)
	folder.GET("/detail", handler.GetFolder)

	// 测试对象
	target := api.Group("/v1/target")
	target.POST("/send", handler.SendTarget) // 接口调试
	target.GET("/result", handler.GetSendTargetResult)
	target.POST("/save", handler.SaveTarget) // 测试对象保存
	target.POST("save_import_api", handler.SaveImportApi)
	target.POST("/sort", handler.SortTarget)
	target.GET("/list", handler.ListFolderAPI)
	target.GET("/detail", handler.BatchGetTarget)
	// sql 调试相关接口
	target.POST("get_sql_database_list", handler.GetSqlDatabaseList) // 获取当前团队下sql数据库列表
	target.POST("send_sql", handler.SendSql)                         // 调试sql语句
	target.POST("connection_database", handler.ConnectionDatabase)   // 测试连接数据库接口
	target.POST("get_send_sql_result", handler.GetSendSqlResult)     // 获取运行sql语句结果
	// Tcp 调试相关接口
	target.POST("send_tcp", handler.SendTcp)                              // 调试tcp接口
	target.POST("get_send_tcp_result", handler.GetSendTcpResult)          // 获取运行tcp结果
	target.POST("tcp_send_or_stop_message", handler.TcpSendOrStopMessage) // 发送或者停止ws消息
	// Websocket 调试相关接口
	target.POST("send_websocket", handler.SendWebsocket)                     // 调试websocket接口
	target.POST("get_send_websocket_result", handler.GetSendWebsocketResult) // 获取运行websocket结果
	target.POST("ws_send_or_stop_message", handler.WsSendOrStopMessage)      // 发送或者停止ws消息
	// Dubbo 调试相关接口
	target.POST("send_dubbo", handler.SendDubbo)                     // 调试Dubbo接口
	target.POST("get_send_dubbo_result", handler.GetSendDubboResult) // 获取运行Dubbo结果
	// Mqtt 调试相关接口
	target.POST("send_mqtt", handler.SendMqtt)                     // 调试Mqtt接口
	target.POST("get_send_mqtt_result", handler.GetSendMqttResult) // 获取运行Mqtt结果

	// 接口回收站
	target.GET("/trash_list", handler.TrashTargetList)
	target.POST("/trash", handler.TrashTarget)
	target.POST("/recall", handler.RecallTarget)
	target.POST("/delete", handler.DeleteTarget)

	// 分组
	group := api.Group("/v1/group")
	group.POST("/save", handler.SaveGroup)
	group.GET("/detail", handler.GetGroup)

	// 场景
	scene := api.Group("/v1/scene")
	// 场景调试
	scene.POST("/send", handler.SendScene)
	scene.POST("/stop", handler.StopScene)
	scene.POST("/api/send", handler.SendSceneAPI)
	scene.GET("/result", handler.GetSendSceneResult)
	scene.POST("/delete", handler.DeleteScene)
	scene.POST("send_mysql", handler.SendMysql) // 调试场景里面的mysql
	// 场景管理
	scene.POST("/save", handler.SaveScene)
	scene.GET("/list", handler.ListGroupScene)
	scene.POST("/detail", handler.BatchGetScene)
	scene.GET("/flow/get", handler.GetFlow)
	scene.GET("/flow/batch/get", handler.BatchGetFlow)
	scene.POST("/flow/save", handler.SaveFlow)
	scene.POST("change_disabled_status", handler.ChangeDisabledStatus) // 修改场景禁用状态

	//用例集管理
	caseAssemble := api.Group("/v1/case")
	caseAssemble.POST("/list", handler.GetCaseAssembleList)               //用例列表
	caseAssemble.POST("/copy", handler.CopyCaseAssemble)                  //copy用例
	caseAssemble.POST("/save", handler.SaveCaseAssemble)                  //保存用例
	caseAssemble.POST("/save/scene/case/flow", handler.SaveSceneCaseFlow) //保存用例执行流
	caseAssemble.POST("/del", handler.DelCaseAssemble)                    //删除用例
	caseAssemble.POST("/flow/detail", handler.GetSceneCaseFlow)           //获取用例执行流
	caseAssemble.POST("/send", handler.SendSceneCase)                     //调试用例
	caseAssemble.POST("/stop", handler.StopSceneCase)                     //停止调试用例
	caseAssemble.POST("/change/check", handler.ChangeCaseAssembleCheck)   //用例启用/关闭
	caseAssemble.POST("change_case_sort", handler.ChangeCaseSort)         //修改用例的排序

	// 测试计划
	plan := api.Group("/v1/plan/")
	plan.POST("run", handler.RunPlan)
	plan.POST("stop", handler.StopPlan)
	plan.POST("clone", handler.ClonePlan)
	plan.GET("list", handler.ListPlans)
	plan.POST("save", handler.SavePlan)
	plan.GET("detail", handler.GetPlan)
	plan.POST("task/save", handler.SavePlanTask)
	plan.GET("task/detail", handler.GetPlanTask)
	plan.POST("delete", handler.DeletePlan)
	plan.POST("email_notify", handler.PlanAddEmail)
	plan.POST("email_delete", handler.PlanDeleteEmail)
	plan.GET("email_list", handler.PlanListEmail)
	plan.POST("import_scene", handler.ImportScene)
	plan.POST("batch_delete", handler.BatchDeletePlan)

	// 测试报告
	report := api.Group("/v1/report/")
	report.GET("list", handler.ListReports)
	report.POST("detail", handler.ReportDetail)
	report.GET("machine", handler.ListMachines)
	report.POST("delete", handler.DeleteReport)
	report.GET("debug", handler.GetDebug)
	report.GET("task_detail", handler.GetReportTaskDetail)
	report.POST("debug/setting", handler.DebugSetting)
	report.POST("stop", handler.StopReport)
	report.GET("debug/detail", handler.DebugDetail)
	report.POST("email_notify", handler.ReportEmail)
	report.POST("change_task_conf_run", handler.ChangeTaskConfRun) // 编辑报告配置并执行
	report.POST("compare_report", handler.CompareReport)           // 对比报告
	report.POST("update/description", handler.UpdateDescription)   // 保存或更新测试结果描述
	report.POST("batch_delete", handler.BatchDeleteReport)         // 批量删除性能计划报告
	report.POST("update_report_name", handler.UpdateReportName)    // 修改计划报告名称

	// 操作日志
	operation := api.Group("/v1/operation")
	operation.GET("/list", handler.ListOperations)

	// 机器管理
	machine := api.Group("/v1/machine/")
	machine.POST("machine_list", handler.GetMachineList)              // 获取压力机列表
	machine.POST("change_machine_on_off", handler.ChangeMachineOnOff) // 启用或停用机器

	// 计划预设配置
	preinstall := api.Group("/v1/preinstall/")
	preinstall.POST("save", handler.SavePreinstall)
	preinstall.POST("list", handler.GetPreinstallList)
	preinstall.POST("detail", handler.GetPreinstallDetail)
	preinstall.POST("delete", handler.DeletePreinstall)
	preinstall.POST("copy", handler.CopyPreinstall)
	preinstall.POST("get_available_machine_list", handler.GetAvailableMachineList)

	// 自动化测试
	// 计划相关
	autoPlan := api.Group("/v1/auto_plan/")
	autoPlan.POST("run", handler.RunAutoPlan)
	autoPlan.POST("save", handler.SaveAutoPlan)
	autoPlan.POST("list", handler.GetAutoPlanList)
	autoPlan.POST("detail", handler.GetAutoPlanDetail)
	autoPlan.POST("delete", handler.DeleteAutoPlan)
	autoPlan.POST("copy", handler.CopyAutoPlan)
	autoPlan.POST("update", handler.UpdateAutoPlan)
	autoPlan.POST("batch_delete", handler.BatchDeleteAutoPlan)
	autoPlan.POST("clone_scene", handler.CloneAutoPlanScene) // 克隆场景
	autoPlan.POST("stop_auto_plan", handler.StopAutoPlan)

	// 邮件相关
	autoPlan.POST("email_list", handler.GetEmailList)
	autoPlan.POST("add_email", handler.AddEmail)
	autoPlan.POST("delete_email", handler.DeleteEmail)

	// 配置相关
	autoPlan.POST("save_task_conf", handler.SaveTaskConf)
	autoPlan.POST("get_task_conf", handler.GetTaskConf)

	// 报告相关
	autoPlan.POST("get_report_list", handler.GetAutoPlanReportList)
	autoPlan.POST("batch_delete_report", handler.BatchDeleteAutoPlanReport)
	autoPlan.POST("get_report_detail", handler.GetAutoPlanReportDetail)
	autoPlan.POST("report_email_notify", handler.ReportEmailNotify)
	autoPlan.POST("get_report_api_detail", handler.GetReportApiDetail)    // 获取自动化报告里面的接口详情
	autoPlan.POST("send_report_api", handler.SendReportApi)               // 运行自动化报告里面的接口
	autoPlan.POST("update_report_name", handler.UpdateAutoPlanReportName) // 修改自动化计划报告名称

	//环境管理
	env := api.Group("/v1/env/")
	env.POST("list", handler.EnvList)                        //获取环境列表 (待废弃)
	env.POST("update_env", handler.UpdateEnv)                //更新环境名称
	env.POST("create_env", handler.CreateEnv)                //新建环境名称
	env.POST("copy_env", handler.CopyEnv)                    //克隆环境信息
	env.POST("del_env", handler.DelEnv)                      //删除环境
	env.POST("del_env_service", handler.DelEnvService)       //删除环境下服务
	env.POST("del_env_database", handler.DelEnvDatabase)     //删除环境下数据库
	env.POST("get_env_list", handler.GetEnvList)             //获取环境列表
	env.POST("get_service_list", handler.GetServiceList)     //获取环境下服务列表
	env.POST("get_database_list", handler.GetDatabaseList)   //获取环境下数据库列表
	env.POST("save_env_service", handler.CreateEnvService)   //新建/修改环境下服务
	env.POST("save_env_database", handler.CreateEnvDatabase) //新建/修改环境下数据库

	// 企业管理相关接口
	company := r.Group("management/api/company")
	company.POST("get_newest_stress_plan_list", handler.GetNewestStressPlanList) // 获取团队最新性能计划列表
	company.POST("get_newest_auto_plan_list", handler.GetNewestAutoPlanList)     // 获取团队最新自动化计划列表

	// 权限相关接口
	permission := api.Group("permission")
	permission.POST("get_team_company_members", handler.GetTeamCompanyMembers) // 获取当前团队和企业成员关系
	permission.POST("team_members_save", handler.TeamMembersSave)              // 添加团队成员
	permission.POST("get_role_member_info", handler.GetRoleMemberInfo)         // 获取我的角色信息
	permission.POST("user_get_marks", handler.UserGetMarks)                    // 获取用户的全部角色对应的mark
	permission.GET("get_notice_group_list", handler.GetNoticeGroupList)        // 获取通知组列表
	permission.GET("get_notice_third_users", handler.GetNoticeThirdUsers)      // 获取三方组织架构成员

	// mock 相关接口
	mock := api.Group("/v1/mock/")
	mock.GET("get", handler.MockInfo)
	mock.POST("save_to_target", handler.MockSaveToTarget) // mock 接口同步至测试对象

	mockFolder := mock.Group("folder/") // 文件夹
	mockFolder.POST("save", handler.MockSaveFolder)
	mockFolder.GET("detail", handler.MockGetFolder)

	mockTarget := mock.Group("target/")             // mock 测试对象
	mockTarget.POST("save", handler.MockSaveTarget) // 测试对象保存
	mockTarget.POST("send", handler.MockSendTarget) // 接口调试
	mockTarget.GET("result", handler.MockGetSendTargetResult)
	mockTarget.GET("detail", handler.MockBatchGetTarget)
	mockTarget.POST("sort", handler.MockSortTarget)
	mockTarget.GET("list", handler.MockListFolderAPI)
	mockTarget.POST("trash", handler.MockTrashTarget)

	// 三方通知
	notice := api.Group("/v1/notice")
	notice.POST("send", handler.SendNotice)                    // 发送通知
	notice.POST("save_event", handler.SaveNoticeEvent)         // 三方通知绑定
	notice.GET("get_group_event", handler.GetGroupNoticeEvent) // 获取通知事件对应通知组ID
}
