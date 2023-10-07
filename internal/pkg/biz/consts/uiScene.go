package consts

const (
	UISceneTypeFolder = "folder"
	UISceneTypeScene  = "scene"

	UISceneStatusNormal = 1 // 正常状态
	UISceneStatusTrash  = 2 // 回收站

	UISceneOptStatusNormal  = 1 // 正常
	UISceneOptStatusDisable = 2 // 禁用

	UISceneOptSyncModeAuto            = 1 // 元素同步方式 1：自动   2：元素同步到场景  3：场景同步到元素
	UISceneOptSyncModeManual          = 2
	UISceneOptSyncModeManualToElement = 3

	UISceneElementTypeSelect = 1 // 1:选择元素  2:自定义元素
	UISceneElementTypeCustom = 2

	UISceneSource     = 1 // 来源：场景
	UISceneSourcePlan = 2 // 计划

	UIReportStatusChange     = "UIReportStatusChange:"
	UIEngineCurrentRunPrefix = "UIEngineCurrentRun:" // ui机器对应的运行ID
	UIEngineRunAddrPrefix    = "UIEngineRunAddr:"    // 运行ID对应的机器

	UISceneOptTypeOpenPage        = "open_page"
	UISceneOptTypeClosePage       = "close_page"
	UISceneOptTypeToggleWindow    = "toggle_window"
	UISceneOptTypeForward         = "forward"
	UISceneOptTypeBack            = "back"
	UISceneOptTypeRefresh         = "refresh"
	UISceneOptTypeSetWindowSize   = "set_window_size"
	UISceneOptTypeMouseClicking   = "mouse_clicking"
	UISceneOptTypeMouseScrolling  = "mouse_scrolling"
	UISceneOptTypeMouseMovement   = "mouse_movement"
	UISceneOptTypeMouseDragging   = "mouse_dragging"
	UISceneOptTypeInputOperations = "input_operations"
	UISceneOptTypeWaitEvents      = "wait_events"
	UISceneOptTypeIfCondition     = "if_condition"
	UISceneOptTypeForLoop         = "for_loop"
	UISceneOptTypeWhileLoop       = "while_loop"
	UISceneOptTypeAssert          = "assert"
	UISceneOptTypeDataWithdraw    = "data_withdraw"
	UISceneOptTypeCodeOperation   = "code_operation"

	UISceneOptTypeSwitchPage           = "switch_page"
	UISceneOptTypeExitFrame            = "exit_frame"
	UISceneOptTypeSwitchFrameByIndex   = "switch_frame_by_index"
	UISceneOptTypeSwitchToParentFrame  = "switch_to_parent_frame"
	UISceneOptTypeSwitchFrameByLocator = "switch_frame_by_locator"

	UISceneOptElementExists             = "element_exists"
	UISceneOptElementNotExists          = "element_not_exists"
	UISceneOptElementDisplayed          = "element_displayed"
	UISceneOptElementNotDisplayed       = "element_not_displayed"
	UISceneOptTextExists                = "text_exists"
	UISceneOptTextNotExists             = "text_not_exists"
	UISceneOptVariableAssertion         = "variable_assertion"
	UISceneOptExpressionAssertion       = "expression_assertion"
	UISceneOptElementAttributeAssertion = "element_attribute_assertion"
	UISceneOptPageAttributeAssertion    = "page_attribute_assertion"

	UISceneWithdrawTypeElement   = "element_method"
	UISceneWithdrawTypeWebpage   = "webpage_method"
	UISceneWithdrawTypeScrollBar = "scroll_bar_method"

	// "element_text_content",element_source_code,element_value,element_attribute,element_position,website_url,webpage_title,webpage_source_code,webpage_text_content
	UISceneWithdrawMethodElementTextContent = "element_text_content"
	UISceneWithdrawMethodElementSourceCode  = "element_source_code"
	UISceneWithdrawMethodElementValue       = "element_value"
	UISceneWithdrawMethodElementAttribute   = "element_attribute"
	UISceneWithdrawMethodElementPosition    = "element_position"
	UISceneWithdrawMethodWebsiteURL         = "website_url"
	UISceneWithdrawMethodWebpageTitle       = "webpage_title"
	UISceneWithdrawMethodWebpageSourceCode  = "webpage_source_code"
	UISceneWithdrawMethodWebpageTextContent = "webpage_text_content"

	// WindowActionFirst 切换窗口
	WindowActionFirst          = "first"
	WindowActionPrevious       = "previous"
	WindowActionNext           = "next"
	WindowActionLast           = "last"
	WindowActionCustomIndex    = "custom_index"
	WindowActionCustomHandleId = "custom_handle_id"
	WindowActionAll            = "all"

	// MouseClickingTypeSingleClickLeft mouse_clicking => "single_click_left", "single_click_right", "double_click","long_press"
	MouseClickingTypeSingleClickLeft  = "single_click_left"
	MouseClickingTypeSingleClickRight = "single_click_right"
	MouseClickingTypeDoubleClick      = "double_click"
	MouseClickingTypeLongPress        = "long_press"

	// MouseScrollingTypeMouse  "scroll_mouse", "scroll_mouse_element_appears"
	MouseScrollingTypeMouse               = "scroll_mouse"
	MouseScrollingTypeMouseElementAppears = "scroll_mouse_element_appears"

	// MouseMovementTypeEnterElement   "mouse_enter_element","mouse_leave_element"
	MouseMovementTypeEnterElement = "mouse_enter_element"
	MouseMovementTypeLeaveElement = "mouse_leave_element"

	// MouseDraggingTypeDragElement "drag_element" , "drag_by_point_coordinates", "drag_to_element_appears"
	MouseDraggingTypeDragElement            = "drag_element"
	MouseDraggingTypeDragByPointCoordinates = "drag_to_target_point"

	// InputOperationsOnElement "input_on_element", "input_at_cursor_position"
	InputOperationsOnElement        = "input_on_element"
	InputOperationsAtCursorPosition = "input_at_cursor_position"

	// WaitEventsFixedTime "fixed_time","element_exist"
	WaitEventsFixedTime           = "fixed_time"
	WaitEventsElementExist        = "element_exist"
	WaitEventsElementNonExist     = "element_non_exist"
	WaitEventsElementDisplayed    = "element_displayed"
	WaitEventsElementNotDisplayed = "element_not_displayed"
	WaitEventsElementEditable     = "element_editable"
	WaitEventsElementNotEditable  = "element_not_editable"
	WaitEventsTextAppearance      = "text_appearance"
	WaitEventsTextDisappearance   = "text_disappearance"

	// ForLoopForTimes for_times  for_collection
	ForLoopForTimes = "for_times"
	ForLoopForData  = "for_data"

	// IfConditionTypeConditionStep condition_step  expression
	IfConditionTypeConditionStep           = "condition_step"
	IfConditionTypeExpression              = "expression"
	IfConditionTypeConditionOperatorAssert = "assert"
)

// ToggleWindowType 1. 切换窗口 - 切换到{第一个/第二个/自定义的索引}
// 2. 切换窗口 - 退出当前frame
// 3. 切换窗口 - 切换到索引号为0的frame
// 4. 切换窗口 - 切换到上一层父级frame
// 5. 切换窗口 - 根据定位当时切换frame
// 6. 切换窗口 - 前进/后退/刷新
var ToggleWindowType = map[string]string{
	WindowActionFirst:          "第一个",
	WindowActionPrevious:       "上一个",
	WindowActionNext:           "下一个",
	WindowActionLast:           "最后一个",
	WindowActionCustomIndex:    "自定义索引",
	WindowActionCustomHandleId: "自定义句柄ID",
	WindowActionAll:            "全部关闭",
}

var MouseClickingType = map[string]string{
	MouseClickingTypeSingleClickLeft:  "单击（左击）",
	MouseClickingTypeSingleClickRight: "单击（右击）",
	MouseClickingTypeDoubleClick:      "双击",
	MouseClickingTypeLongPress:        "长按",
}

// MouseScrollingType 1. 单击（左击/右击）/双击/长按 - 元素名称/{定位方式_定位类型}
// 2. 鼠标滚动 - 滚动距离
// 3. 鼠标滚动到元素出现 - 目标元素名称
// 4. 鼠标移入元素 - 元素名称
// 5. 鼠标移出元素 - 元素名称
// 6. 拖动元素 - 将{元素名称}拖动至终点(x,y)
// 7. 按点位坐标 - 从起点(x,y)拖动至终点(x,y)
// 8. 拖动到元素出现 - 元素名称
var MouseScrollingType = map[string]string{
	MouseScrollingTypeMouse:               "鼠标滚动",
	MouseScrollingTypeMouseElementAppears: "鼠标滚动到元素出现",
}

var MouseMovementType = map[string]string{
	MouseMovementTypeEnterElement: "鼠标移入元素",
	MouseMovementTypeLeaveElement: "鼠标移出元素",
}

//var MouseDraggingType = map[string]string{
//	MouseDraggingTypeDragElement:            "拖动元素",
//	MouseDraggingTypeDragByPointCoordinates: "按点位坐标",
//	MouseDraggingTypeDragToElementAppears:   "拖动到元素出现",
//}

var InputOperationsType = map[string]string{
	InputOperationsOnElement:        "元素上输入",
	InputOperationsAtCursorPosition: "光标处输入",
}

var WaitEventsType = map[string]string{
	WaitEventsFixedTime:           "等待固定时长",
	WaitEventsElementExist:        "等待元素存在",
	WaitEventsElementNonExist:     "等待元素不存在",
	WaitEventsElementDisplayed:    "等待元素显示",
	WaitEventsElementNotDisplayed: "等待元素不显示",
	WaitEventsElementEditable:     "等待元素可编辑",
	WaitEventsElementNotEditable:  "等待元素不可编辑",
	WaitEventsTextAppearance:      "等待文本出现",
	WaitEventsTextDisappearance:   "等待文本消失",
}

var IfConditionType = map[string]string{
	IfConditionTypeConditionStep: "条件步骤",
	IfConditionTypeExpression:    "表达式",
}

var ForLoopType = map[string]string{
	ForLoopForTimes: "循环次数",
	ForLoopForData:  "循环数据",
}

// AssertType 1. 断言元素存在/不存在/显示/不显示 - 元素名称
//
// 2. 断言文本存在/不存在 - 文本内容
// 3. 变量断言 - {实际值}{断言关系}{期望值}，如：1等于2
// 4. 表达式断言 - 表达式
// 5. 断言元素属性 - {元素名称} {条件类型}{期望值}
// 6. 断言页面属性 - {断言属性} {断言关系}{期望值}
var AssertType = map[string]string{
	UISceneOptElementExists:             "断言元素存在",
	UISceneOptElementNotExists:          "断言元素不存在",
	UISceneOptElementDisplayed:          "断言元素显示",
	UISceneOptElementNotDisplayed:       "断言元素不显示",
	UISceneOptTextExists:                "断言文本存在",
	UISceneOptTextNotExists:             "断言文本不存在",
	UISceneOptVariableAssertion:         "变量断言",
	UISceneOptExpressionAssertion:       "表达式断言",
	UISceneOptElementAttributeAssertion: "断言元素属性",
	UISceneOptPageAttributeAssertion:    "断言页面属性",
}

var RelationOptions = map[string]string{
	"Equal":                "相等",
	"NotEqual":             "不相等",
	"Contains":             "包含",
	"NotContains":          "不包含",
	"GreaterThan":          "大于",
	"LessThan":             "小于",
	"NotEqualTo":           "不等于",
	"GreaterThanorEqualTo": "大于等于",
	"LessThanorEqualTo":    "小于等于",
	"Regex":                "正则断言",
}

type T struct {
	Equal                string `json:"Equal"`
	NotEqual             string `json:"NotEqual"`
	Contains             string `json:"Contains"`
	NotContains          string `json:"NotContains"`
	GreaterThan          string `json:"GreaterThan"`
	LessThan             string `json:"LessThan"`
	NotEqualTo           string `json:"NotEqualTo"`
	GreaterThanorEqualTo string `json:"GreaterThanorEqualTo"`
	LessThanorEqualTo    string `json:"LessThanorEqualTo"`
	Regex                string `json:"Regex"`
}
