package consts

const (
	TargetTypeFolder   = "folder"
	TargetTypeAPI      = "api"
	TargetTypeGroup    = "group"
	TargetTypeScene    = "scene"
	TargetTypeTestCase = "test_case"

	TargetStatusNormal = 1 // 正常状态
	TargetStatusTrash  = 2 // 回收站

	TargetSourceNormal   = 1 // 正常来源
	TargetSourcePlan     = 2 // 性能计划来源
	TargetSourceAutoPlan = 3 // 自动化计划来源

	TargetIsCheckedOpen  = 1 // 启用状态：1-开启
	TargetIsCheckedClose = 2 // 启用状态：2-关闭

	TargetDebugLogApi   = 1
	TargetDebugLogScene = 2
)
