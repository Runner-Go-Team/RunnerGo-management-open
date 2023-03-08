package consts

const (
	PlanStatusNormal   = 1 // 未开始
	PlanStatusUnderway = 2 // 进行中

	PlanTaskTypeNormal  = 1 // 普通任务
	PlanTaskTypeCronjob = 2 // 定时任务
	PlanTaskTypeMix     = 3 // 混合模式

	PlanModeConcurrence  = 1 // 并发模式
	PlanModeStep         = 2 // 阶梯模式
	PlanModeErrorRate    = 3 // 错误率模式
	PlanModeResponseTime = 4 // 响应时间模式
	PlanModeRPS          = 5 //每秒请求数模式
	PlanModeTPS          = 6 //每秒事务数模式
	PlanModeMix          = 7 // 混合模式

	// 定时任务的几个状态
	TimedTaskWaitEnable = 0 // 未启用
	TimedTaskInExec     = 1 // 运行中
	TimedTaskTimeout    = 2 // 已过期

	// StopPlanPrefix 停止计划的redis健前缀
	StopPlanPrefix                     = "StopPlan:"
	StopScenePrefix                    = "StopScene:"
	StopAutoPlanPrefix                 = "StopAutoPlan:"
	SubscriptionStressPlanStatusChange = "SubscriptionStressPlanStatusChange:"

	// 单台压力机能快速负载的并发数
	OneMachineCanConcurrence = 5000

	// 以下是自动化测试相关配置
	// 自动化测试计划的运行模式
	AutoPlanTaskRunMode = 1 // 按照用例执行

	// 场景运行模式
	AutoPlanSceneRunModeOrder    = 1 // 顺讯执行
	AutoPlanSceneRunModeMeantime = 2 // 同时执行

	// 用例运行模式
	AutoPlanTestCaseRunModeOrder    = 1 // 顺讯执行
	AutoPlanTestCaseRunModeMeantime = 2 // 同时执行
)
