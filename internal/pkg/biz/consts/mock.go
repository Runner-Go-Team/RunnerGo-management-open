package consts

const (
	MockRequestScope         = "$request"
	MockRequestJsonScope     = "$request-json"
	MockFolderSaveNamePrefix = "Mock-"
	MockSaveNamePrefix       = "Mock-"

	MockSplit = "."

	MockContentTypeJson = "json"
	MockContentTypeText = "text"
	MockContentTypeHtml = "html"

	MockTargetSourceApi = 0 // mock 场景管理

	MockTargetTypeAPI = "api"

	MockTargetStatusNormal = 1 // 正常状态
	MockTargetStatusTrash  = 2 // 回收站

	IsMockOpenOn  = 1 // 是否开启mock   1:启用   2:禁用
	IsMockOpenOff = 2

	MockOperateTypeSave          = 1 //  1:保存  2:保存并添加到测试对象  3:保存并同步
	MockOperateTypeSaveAndTarget = 2
	MockOperateTypeSaveAndSync   = 3

	MockConditionJsonPath = "body-json"
)

var MockContentType = map[string]string{
	MockContentTypeJson: "application/json",
	MockContentTypeText: "text/plain; charset=utf-8",
	MockContentTypeHtml: "text/html; charset=utf-8",
}

var MockNameOrder = []string{
	"-01", "-02", "-03", "-04", "-05",
	"-06", "-07", "-08", "-09", "-10",
	"-11", "-12", "-13", "-14", "-15",
	"-16", "-17", "-18", "-19", "-20",
}
