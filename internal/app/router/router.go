package router

import (
	"time"

	"github.com/gin-contrib/cors"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/go-omnibus/proof"

	"kp-management/internal/app/middleware"
	"kp-management/internal/pkg/handler"
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

	// 文件夹
	folder := api.Group("/v1/folder")
	folder.POST("/save", handler.SaveFolder)
	folder.GET("/detail", handler.GetFolder)

	// 接口
	target := api.Group("/v1/target")
	// 接口调试
	target.POST("/send", handler.SendTarget)
	target.GET("/result", handler.GetSendTargetResult)
	// 接口保存
	target.POST("/save", handler.SaveTarget)
	target.POST("save_import_api", handler.SaveImportApi)
	target.POST("/sort", handler.SortTarget)
	target.GET("/list", handler.ListFolderAPI)
	target.GET("/detail", handler.BatchGetTarget)
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

	// 场景管理
	scene.POST("/save", handler.SaveScene)
	scene.GET("/list", handler.ListGroupScene)
	scene.POST("/detail", handler.BatchGetScene)
	scene.GET("/flow/get", handler.GetFlow)
	scene.GET("/flow/batch/get", handler.BatchGetFlow)
	scene.POST("/flow/save", handler.SaveFlow)

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
	report.GET("detail", handler.ReportDetail)
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

	//环境管理
	env := api.Group("/v1/env/")
	env.POST("list", handler.EnvList)           //获取环境列表
	env.POST("save", handler.SaveEnv)           //保存/编辑环境信息
	env.POST("copy", handler.CopyEnv)           //复制环境信息
	env.POST("del", handler.DelEnv)             //删除环境
	env.POST("del_service", handler.DelService) //删除环境下服务
	//env.POST("get_service_list", handler.ServiceList) //获取环境下服务列表
	//env.POST("save_service", handler.SaveService)     //保存/编辑环境下服务信息
}
