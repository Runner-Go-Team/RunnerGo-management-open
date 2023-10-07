package errmsg

import "errors"

var (
	ErrTargetSortNameAlreadyExist = errors.New("存在重名，无法操作")

	ErrElementFolderNameRepeat = errors.New("元素目录不能重复")
	ErrElementNameRepeat       = errors.New("元素名称不能重复")
	ErrElementNotFound         = errors.New("元素不存在")
	ErrElementLocatorNotFound  = errors.New("元素属性不能为空")
	ErrElementNotDeleteReScene = errors.New("元素与场景存在关联关系，不可删除")

	ErrUIEngineNotFound     = errors.New("获取 UI-Engine 失败")
	ErrSendUIEngineDeadline = errors.New("请求 UI 服务超时")

	ErrUISceneFolderNameRepeat = errors.New("目录不能重复")
	ErrUISceneNameRepeat       = errors.New("名称不能重复")
	ErrUISceneNameLong         = errors.New("名称过长！不可超出30字符")

	ErrUISceneOpenPageEmpty             = errors.New("OpenPage 参数错误或为空")
	ErrUISceneClosePageEmpty            = errors.New("ClosePage 参数错误或为空")
	ErrUISceneSwitchPageEmpty           = errors.New("SwitchPage 参数错误或为空")
	ErrUISceneSwitchFrameByIndexEmpty   = errors.New("SwitchFrameByIndex 参数错误或为空")
	ErrUISceneSwitchToParentFrame       = errors.New("SwitchToParentFrame 参数错误或为空")
	ErrUISceneSwitchFrameByLocatorEmpty = errors.New("SwitchFrameByLocator 参数错误或为空")
	ErrUISceneSetWindowSizeEmpty        = errors.New("SetWindowSize 参数错误或为空")
	ErrUISceneMouseClickingEmpty        = errors.New("MouseClicking 参数错误或为空")
	ErrUISceneMouseScrollingEmpty       = errors.New("MouseScrolling 参数错误或为空")
	ErrUISceneMouseMovementEmpty        = errors.New("MouseMovement 参数错误或为空")
	ErrUISceneMouseDraggingEmpty        = errors.New("MouseDragging 参数错误或为空")
	ErrUISceneInputOperationsEmpty      = errors.New("InputOperations 参数错误或为空")
	ErrUISceneWaitEventsEmpty           = errors.New("WaitEvents 参数错误或为空")
	ErrUISceneIfConditionEmpty          = errors.New("IfCondition 参数错误或为空")
	ErrUISceneForLoopEmpty              = errors.New("ForLoop 参数错误或为空")
	ErrUISceneWhileLoopEmpty            = errors.New("ErrUISceneWhileLoopEmpty 参数错误或为空")
	ErrUISceneAssertEmpty               = errors.New("assert 参数错误或为空")
	ErrUISceneRequired                  = errors.New("请检查必填项")

	ErrUISceneAssertElementExistsEmpty          = errors.New("AssertElementExists 参数错误或为空")
	ErrUISceneAssertElementNotExistsEmpty       = errors.New("ElementNotExists 参数错误或为空")
	ErrUISceneAssertElementDisplayedEmpty       = errors.New("ElementDisplayed 参数错误或为空")
	ErrUISceneAssertElementNotDisplayedEmpty    = errors.New("ElementNotDisplayed 参数错误或为空")
	ErrUISceneAssertTextExistsEmpty             = errors.New("TextExists 参数错误或为空")
	ErrUISceneAssertTextNotExistsEmpty          = errors.New("TextNotExists 参数错误或为空")
	ErrUISceneAssertVariableAssertionEmpty      = errors.New("VariableAssertion 参数错误或为空")
	ErrUISceneAssertExpressionAssertionEmpty    = errors.New("ExpressionAssertion 参数错误或为空")
	ErrUISceneAssertElementAttributeAssertEmpty = errors.New("ElementAttributeAssert 参数错误或为空")
	ErrUISceneAssertPageAttributeAssertEmpty    = errors.New("PageAttributeAssert 参数错误或为空")

	ErrUISceneWithdrawElementExistsEmpty = errors.New("WithdrawElement 参数错误或为空")

	ErrMustTaskInit     = errors.New("请填写任务配置并保存")
	ErrTimedTaskOverdue = errors.New("开始或结束时间不能早于当前时间")

	ErrSendOperatorNotNull = errors.New("发送的步骤不能为空")
	ErrSendLinuxNotQTMode  = errors.New("linux 机器不支持前台运行模式")
)
