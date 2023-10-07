package consts

const (
	TargetTypeFolder    = "folder"
	TargetTypeAPI       = "api"
	TargetTypeScene     = "scene"
	TargetTypeTestCase  = "test_case"
	TargetTypeSql       = "sql"
	TargetTypeTcp       = "tcp"
	TargetTypeWebsocket = "websocket"
	TargetTypeMQTT      = "mqtt"
	TargetTypeDubbo     = "dubbo"

	TargetStatusNormal = 1 // 正常状态
	TargetStatusTrash  = 2 // 回收站

	TargetSourceApi      = 0 // 测试对象
	TargetSourceScene    = 1 // 场景管理
	TargetSourcePlan     = 2 // 性能计划来源
	TargetSourceAutoPlan = 3 // 自动化计划来源

	TargetIsCheckedOpen  = 1 // 启用状态：1-开启
	TargetIsCheckedClose = 2 // 启用状态：2-关闭

	TargetDebugLogApi   = 1
	TargetDebugLogScene = 2

	TargetIsDisabledNo  = 0 // 不禁用
	TargetIsDisabledYes = 1 // 禁用

	// 测试对象同步内容字段
	TargetSyncMethodID = 10001 // 请求方法
	TargetSyncUrlID    = 10002 // URL
	TargetSyncCookieID = 10003 // Cookie
	TargetSyncHeaderID = 10004 // Header
	TargetSyncQueryID  = 10005 // Query
	TargetSyncBodyID   = 10006 // Body
	TargetSyncAuthID   = 10007 // 认证
	TargetSyncAssertID = 10008 // 断言
	TargetSyncRegexID  = 10009 // 关联提取
	TargetSyncConfigID = 10010 // 接口设置

	TargetSyncDataTypePush = 1 // 推送
	TargetSyncDataTypePull = 2 // 拉取
)
