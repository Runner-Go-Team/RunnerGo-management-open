package consts

const (
	UIPlanSceneSyncModeAuto            = 1 // 场景同步方式 1：自动   2： 手动 已场景为准  3：已计划为准
	UIPlanSceneSyncModeHandTargetScene = 2
	UIPlanSceneSyncModeHandTargetPlan  = 3

	UIPlanSceneSyncModeTargetScene = 1
	UIPlanSceneSyncModeTargetPlan  = 2 // 手动同步时以什么为准   1:已场景为准  2:已计划为准

	UIPlanTaskTypeNormal  = 1 // 普通任务
	UIPlanTaskTypeCronjob = 2 // 定时任务

	// 定时任务的几个状态
	UIPlanTimedTaskWaitEnable = 0 // 未启用
	UIPlanTimedTaskInExec     = 1 // 运行中
	UIPlanTimedTaskTimeout    = 2 // 已过期

	// 定时任务的执行频次
	UIPlanFrequencyOnce      = 0
	UIPlanFrequencyEveryday  = 1
	UIPlanFrequencyWeekly    = 2
	UIPlanFrequencyMonthly   = 3
	UIPlanFrequencyFixedTime = 4

	// 场景运行模式
	UIPlanSceneRunModeOrder    = 1 // 顺序执行
	UIPlanSceneRunModeMeantime = 2 // 同时执行
)
