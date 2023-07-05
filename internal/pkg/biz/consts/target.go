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

	TargetSourceApi      = 0 // 场景管理
	TargetSourceScene    = 1 // 场景管理
	TargetSourcePlan     = 2 // 性能计划来源
	TargetSourceAutoPlan = 3 // 自动化计划来源

	TargetIsCheckedOpen  = 1 // 启用状态：1-开启
	TargetIsCheckedClose = 2 // 启用状态：2-关闭

	TargetDebugLogApi   = 1
	TargetDebugLogScene = 2

	TargetIsDisabledNo  = 0 // 不禁用
	TargetIsDisabledYes = 1 // 禁用
)
