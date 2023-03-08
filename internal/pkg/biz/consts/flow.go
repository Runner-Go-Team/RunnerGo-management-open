package consts

const (
	FlowTypeAPI                 = "api"                  // 接口
	FlowTypeAssert              = "assert"               // 全局断言
	FlowTypeWaitController      = "wait_controller"      // 等待控制器
	FlowTypeConditionController = "condition_controller" // 条件控制器

	FlowAPIModeDefault      = 1 // 默认模式
	FlowAPIModeErrorRate    = 2 // 错误率模式
	FlowAPIModeTPS          = 3 // 每秒事务数模式
	FlowAPIModeResponseTime = 4 // 响应时间模式

	FlowConditionEQ = "eq"
)
