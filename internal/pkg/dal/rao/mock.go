package rao

// Expect mock 期望
type Expect struct {
	Name       string             `json:"name"`
	ExpectId   string             `json:"expect_id"`
	Conditions []*ExpectCondition `json:"conditions"`
	Response   *ExpectResponse    `json:"response"`
}

type ExpectCondition struct {
	Path           string `json:"path"`
	ParameterName  string `json:"parameter_name"`
	Compare        string `json:"compare"`
	ParameterValue string `json:"parameter_value"`
}

type ExpectResponse struct {
	ContentType string `json:"content_type"`
	JsonSchema  string `json:"json_schema"`
	Json        string `json:"json"`
	Raw         string `json:"raw"`
}

type MockAPIDetail struct {
	*APIDetail
	IsMockOpen int32     `json:"is_mock_open"` // 是否开启mock
	MockPath   string    `json:"mock_path"`    //
	Expects    []*Expect `json:"expects"`      // mock 期望
}

type MockSaveFolderReq struct {
	TargetID    string `json:"target_id"`
	TeamID      string `json:"team_id" binding:"required,gt=0"`
	ParentID    string `json:"parent_id"`
	Name        string `json:"name" binding:"required,min=1"`
	Method      string `json:"method"`
	Sort        int32  `json:"sort"`
	TypeSort    int32  `json:"type_sort"`
	Version     int32  `json:"version"`
	Description string `json:"description"`
	Source      int32  `json:"source"`
	//Request  *Request `json:"request"`
	//Script   *Script  `json:"script"`
}

type MockGetFolderReq struct {
	TeamID   string `form:"team_id"`
	TargetID string `form:"target_id"`
}

type MockGetFolderResp struct {
	Folder *MockFolder `json:"folder"`
}

type MockInfoReq struct {
	TeamID string `form:"team_id"  binding:"required"`
}

type MockInfoResp struct {
	HttpPreUrl string `json:"http_pre_url"`
}

type MockFolder struct {
	TargetID    string `json:"target_id"`
	TeamID      string `json:"team_id"`
	ParentID    string `json:"parent_id"`
	Name        string `json:"name"`
	Method      string `json:"method"`
	Sort        int32  `json:"sort"`
	TypeSort    int32  `json:"type_sort"`
	Version     int32  `json:"version"`
	Source      int32  `json:"source"`
	Description string `json:"description"`
	SourceID    string `json:"source_id"`
}

type MockFolderAPI struct {
	TargetID      string `json:"target_id"`
	TeamID        string `json:"team_id"`
	TargetType    string `json:"target_type"`
	Name          string `json:"name"`
	Url           string `json:"url"`
	ParentID      string `json:"parent_id"`
	Method        string `json:"method"`
	Sort          int32  `json:"sort"`
	TypeSort      int32  `json:"type_sort"`
	Version       int32  `json:"version"`
	Source        int32  `json:"source"`
	IsMockOpen    int32  `json:"is_mock_open"` // 是否开启mock
	CreatedUserID string `json:"created_user_id"`
	RecentUserID  string `json:"recent_user_id"`
}

type MockSaveTargetReq struct {
	*SaveTargetReq
	OperateType int32     `json:"operate_type"` // 1:保存  2:保存并添加到测试对象  3:保存并同步
	IsMockOpen  int32     `json:"is_mock_open"`
	MockPath    string    `json:"mock_path"`
	Expects     []*Expect `json:"expects"`
}

type MockSaveTargetResp struct {
	TargetID string `json:"target_id"`
}

type MockSaveToTargetReq struct {
	TargetIDs []string `json:"targetIDs" binding:"required"`
	TeamID    string   `json:"team_id" binding:"required,gt=0"`
}

type MockSendTargetReq struct {
	TargetID string `json:"target_id" binding:"required,gt=0"`
	TeamID   string `json:"team_id" binding:"required,gt=0"`
}

type MockSendTargetResp struct {
	RetID string `json:"ret_id"`
}

type MockBatchGetDetailReq struct {
	TeamID    string   `form:"team_id" binding:"required,gt=0"`
	TargetIDs []string `form:"target_ids" binding:"required,gt=0"`
}

type MockBatchGetDetailResp struct {
	Targets []MockAPIDetail `json:"targets"`
}

type MockSortTargetReq struct {
	Targets []*SortTarget `json:"targets"`
}

type MockListFolderAPIReq struct {
	TeamID string `form:"team_id" binding:"required,gt=0"`
	PlanID int64  `json:"plan_id" form:"plan_id"`
	Source int32  `json:"source" form:"source"`
}

type MockListFolderAPIResp struct {
	Targets []*MockFolderAPI `json:"targets"`
	Total   int64            `json:"total"`
}

type MockDeleteTargetReq struct {
	TargetID string `json:"target_id" binding:"required,gt=0"`
}
