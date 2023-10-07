package rao

type Browser struct {
	Headless    bool        `json:"headless"`
	BrowserType string      `json:"browser_type"`
	SizeType    string      `json:"size_type,omitempty"`
	SetSize     *WindowSize `json:"set_size,omitempty"`
}

type ActionDetail struct {
	OpenPage        *OpenPage         `json:"open_page,omitempty" bson:"open_page,omitempty"`
	ClosePage       *ClosePage        `json:"close_page,omitempty" bson:"close_page,omitempty"`
	ToggleWindow    *ToggleWindow     `json:"toggle_window,omitempty" bson:"toggle_window,omitempty"`
	SetWindowSize   *SetWindowSize    `json:"set_window_size,omitempty" bson:"set_window_size,omitempty"`
	Forward         *Forward          `json:"forward,omitempty" bson:"forward,omitempty"`
	Back            *Back             `json:"back,omitempty" bson:"back,omitempty"`
	Refresh         *Refresh          `json:"refresh,omitempty" bson:"refresh,omitempty"`
	MouseClicking   *MouseClick       `json:"mouse_clicking,omitempty" bson:"mouse_clicking,omitempty"`
	MouseScrolling  *MouseScroll      `json:"mouse_scrolling,omitempty" bson:"mouse_scrolling,omitempty"`
	MouseMovement   *MouseMove        `json:"mouse_movement,omitempty" bson:"mouse_movement,omitempty"`
	MouseDragging   *MouseDrag        `json:"mouse_dragging,omitempty" bson:"mouse_dragging,omitempty"`
	InputOperations *InputOperations  `json:"input_operations,omitempty" bson:"input_operations,omitempty"`
	WaitEvents      *WaitEvent        `json:"wait_events,omitempty" bson:"wait_events,omitempty"`
	IfCondition     *IfCondition      `json:"if_condition,omitempty" bson:"if_condition,omitempty"`
	ForLoop         *ForLoop          `json:"for_loop,omitempty" bson:"for_loop,omitempty"`
	WhileLoop       *WhileLoop        `json:"while_loop,omitempty" bson:"while_loop,omitempty"`
	Assert          *AutomationAssert `json:"assert,omitempty" bson:"assert,omitempty"`
	DataWithdraw    *DataWithdraw     `json:"data_withdraw,omitempty" bson:"data_withdraw,omitempty"`
	CodeOperation   *CodeOperation    `json:"code_operation,omitempty" bson:"code_operation,omitempty"`
}

type CodeOperation struct {
	Type          string   `json:"type"` // javascript
	Element       *Element `json:"element"`
	OperationType string   `json:"operation_type"` // element | page
	CodeText      string   `json:"code_text"`
}

type DataWithdraw struct {
	Status          int32                    `json:"status,omitempty"`
	Name            string                   `json:"name"`
	VariableType    string                   `json:"variable_type"`
	WithdrawType    string                   `json:"withdraw_type"` // "element_method","webpage_method","scroll_bar_method"
	ElementMethod   *WithdrawElementMethod   `json:"element_method"`
	WebpageMethod   *WithdrawWebpageMethod   `json:"webpage_method"`
	ScrollBarMethod *WithdrawScrollBarMethod `json:"scroll_bar_method"`
}

type WithdrawElementMethod struct {
	Method        string   `json:"method"`
	Element       *Element `json:"element"`
	AttributeName string   `json:"attribute_name"`
	PositionType  string   `json:"position_type"`
}

type WithdrawWebpageMethod struct {
	Method string `json:"method"`
	Value  string `json:"value"`
}

type WithdrawScrollBarMethod struct {
	Method         string `json:"method"`
	ScrollPosition string `json:"scroll_position"`
}

type OpenPage struct {
	Url       string `json:"url,omitempty" binding:"required"`
	IsNewPage bool   `json:"is_new_page,omitempty"`
}

type ClosePage struct {
	WindowAction string `json:"window_action,omitempty"`
	CustomIndex  int    `json:"custom_index,omitempty"`
	InputContent string `json:"input_content,omitempty"`
}

type ToggleWindow struct {
	Type                 string                `json:"type,omitempty"`
	SwitchPage           *SwitchPage           `json:"switch_page,omitempty" bson:"switch_page,omitempty"`
	ExitFrame            *ExitFrame            `json:"exit_frame,omitempty" bson:"exit_frame,omitempty"`
	SwitchFrameByIndex   *SwitchFrameByIndex   `json:"switch_frame_by_index,omitempty" bson:"switch_frame_by_index,omitempty"`
	SwitchToParentFrame  *SwitchToParentFrame  `json:"switch_to_parent_frame,omitempty" bson:"switch_to_parent_frame,omitempty"`
	SwitchFrameByLocator *SwitchFrameByLocator `json:"switch_frame_by_locator,omitempty" bson:"switch_frame_by_locator,omitempty"`
}

type SwitchPage struct {
	WindowAction string `json:"window_action,omitempty"`
	InputContent string `json:"input_content,omitempty"`
}

type ExitFrame struct{}

type SwitchFrameByIndex struct {
	FrameIndex int `json:"frame_index"`
}

type SwitchToParentFrame struct{}

type SwitchFrameByLocator struct {
	Element *Element `json:"element"`
}

type Forward struct{}

type Back struct{}

type Refresh struct{}

type SetWindowSize struct {
	Type    string      `json:"type,omitempty"`
	SetSize *WindowSize `json:"set_size,omitempty"`
}

type WindowSize struct {
	X float64 `json:"x,omitempty"`
	Y float64 `json:"y,omitempty"`
}

type AutomationSettings struct {
	WaitBeforeExec   int32  `json:"wait_before_exec,omitempty"`
	Timeout          int32  `json:"timeout,omitempty"`
	ErrorHandling    string `json:"error_handling,omitempty"`
	ScreenshotConfig string `json:"screenshot_config,omitempty"`
	ElementSyncMode  int32  `json:"element_sync_mode"` // 1.实时同步  2.手动同步
}

type AutomationAssert struct {
	Type                   string                  `json:"type,omitempty"`
	Status                 int32                   `json:"status,omitempty"`
	Element                *Element                `json:"element"`
	TextExists             *TextExists             `json:"text_exists,omitempty"`
	TextNotExists          *TextNotExists          `json:"text_not_exists,omitempty"`
	VariableAssertion      *VariableAssertion      `json:"variable_assertion,omitempty"`
	ExpressionAssertion    *ExpressionAssertion    `json:"expression_assertion,omitempty"`
	ElementAttributeAssert *ElementAttributeAssert `json:"element_attribute_assertion,omitempty"`
	PageAttributeAssert    *PageAttributeAssert    `json:"page_attribute_assertion,omitempty"`
}

type ElementExists struct {
	Element *Element `json:"element"`
}

type ElementNotExists struct {
	Element *Element `json:"element"`
}

type ElementDisplayed struct {
	Element *Element `json:"element"`
}

type ElementNotDisplayed struct {
	Element *Element `json:"element"`
}

type ElementMode struct {
	SyncMode int32    `json:"sync_mode"` // 1.实时同步  2.手动同步
	Element  *Element `json:"element"`
}

type TextExists struct {
	TargetTexts []string `json:"target_texts,omitempty"`
}

type TextNotExists struct {
	TargetTexts []string `json:"target_texts,omitempty"`
}

type ExpressionAssertion struct {
	ExpectedValue string `json:"expected_value,omitempty"`
}

type ElementAttributeAssert struct {
	RelationOptions string `json:"relation_options,omitempty"`
	ConditionType   string `json:"condition_type,omitempty"`
	ExpectedValue   string `json:"expected_value,omitempty"`
}

type PageAttributeAssert struct {
	RelationOptions string `json:"relation_options,omitempty"`
	AssertAttribute string `json:"assert_attribute,omitempty"`
	ExpectedValue   string `json:"expected_value,omitempty"`
}

type AssertElementExists struct {
	Element *Element `json:"element"`
}

type AssertTextExists struct {
	TargetTexts []string `json:"target_texts,omitempty"`
}

type TargetText struct {
	Context string `json:"context,omitempty"`
}

type Expression struct {
	RelationOptions string `json:"relation_options"  binding:"required"`
	ActualValue     string `json:"actual_value"  binding:"required"`
	ExpectedValue   string `json:"expected_value"  binding:"required"`
}

type VariableAssertion struct {
	RelationOptions string `json:"relation_options,omitempty"`
	ActualValue     string `json:"actual_value,omitempty"`
	ExpectedValue   string `json:"expected_value,omitempty"`
}

type MouseClick struct {
	Type          string         `json:"type" binding:"required"`
	Element       *Element       `json:"element" binding:"required"`
	ClickPosition *ClickPosition `json:"click_position,omitempty"`
}

type ClickPosition struct {
	X float64 `json:"x,omitempty"`
	Y float64 `json:"y,omitempty"`
}

type MouseScroll struct {
	Type                 string   `json:"type,omitempty" binding:"required"`
	Direction            string   `json:"direction" binding:"required"` // upAndDown, leftAndRight
	Element              *Element `json:"element" binding:"required_if=Type scroll_mouse_element_appears"`
	ScrollDistance       int      `json:"scroll_distance" binding:"required_if=Type scroll_mouse"`
	SingleScrollDistance int      `json:"single_scroll_distance" binding:"required_if=Type scroll_mouse_element_appears"`
}

type MouseMove struct {
	Type                string                `json:"type" binding:"required"`
	EndPointCoordinates *DragPointCoordinates `json:"end_point_coordinates"`
}

type MouseDrag struct {
	Type                string                `json:"type,omitempty" binding:"required"`
	Element             *Element              `json:"element" binding:"required"`
	TarGetElement       *Element              `json:"target_element"`
	EndPointCoordinates *DragPointCoordinates `json:"end_point_coordinates"`
}

type DragPointCoordinates struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type InputOperations struct {
	Type            string   `json:"type"  binding:"required"`
	Element         *Element `json:"element" binding:"required_if=Type input_on_element"`
	InputContent    string   `json:"input_content" binding:"required"`
	IsAppendContent bool     `json:"is_append_content"`
}

type WaitEvent struct {
	Type        string   `json:"type" binding:"required"`
	WaitTime    int32    `json:"wait_time" binding:"required_if=Type fixed_time"`
	Element     *Element `json:"element"`
	TargetTexts []string `json:"target_texts"`
}

type ConditionOperator struct {
	Type       string            `json:"type"`
	Status     int32             `json:"status,omitempty"`
	Assert     *AutomationAssert `json:"assert"`
	Expression *Expression       `json:"expression,omitempty"`
}

type ForLoop struct {
	Type  string      `json:"type" binding:"required"`
	Count int         `json:"count" binding:"required_if=Type for_times"`
	Files []*BaseFile `json:"files,omitempty"`
}

type BaseFile struct {
	Status   int32  `json:"status"`
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Url      string `json:"url"`
	FileType int64  `json:"file_type"` // 0:普通文件  1:json
	Mark     bool   `json:"mark"`
}

type WhileLoop struct {
	IfCondition
	MaxCount int `json:"max_count,omitempty"`
}

type IfCondition struct {
	ConditionRelate    string               `json:"condition_relate" binding:"required"`
	ConditionOperators []*ConditionOperator `json:"condition_operators,omitempty"`
}

type UIEngineAssertion struct {
	Name   string `json:"name"`
	Status bool   `json:"status"` // 状态
	Msg    string `json:"msg"`
}

type UIEngineDataWithdraw struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}
