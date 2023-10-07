package consts

const (
	ElementTypeDefault = "element"
	ElementTypeFolder  = "folder"

	ElementSourceDefault = 0 // 元素管理
	ElementSourceScene   = 1 // 场景管理

	ElementPlaywrightSelector = "playwright_selector"
	ElementPlaywrightLocator  = "playwright_locator"
)

var ElementMethodType = map[string]string{
	ElementPlaywrightSelector: "选择器",
	ElementPlaywrightLocator:  "定位器",
}
