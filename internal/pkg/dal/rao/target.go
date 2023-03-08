package rao

type SendTargetReq struct {
	TargetID string `json:"target_id" binding:"required,gt=0"`
	TeamID   string `json:"team_id" binding:"required,gt=0"`
}

type SendTargetResp struct {
	RetID string `json:"ret_id"`
}

type GetSendTargetResultReq struct {
	RetID string `form:"ret_id" binding:"required,gt=0"`
}

type GetSendTargetResultResp struct {
}

type SaveTargetReq struct {
	TargetID      string   `json:"target_id"`
	ParentID      string   `json:"parent_id"`
	TeamID        string   `json:"team_id" binding:"required,gt=0"`
	Mark          string   `json:"mark"`
	Name          string   `json:"name" binding:"required,min=1"`
	Method        string   `json:"method" binding:"required"`
	PreUrl        string   `json:"pre_url"`
	URL           string   `json:"url"`
	EnvServiceID  int64    `json:"env_service_id"`
	EnvServiceURL string   `json:"env_service_url"`
	Sort          int32    `json:"sort"`
	TypeSort      int32    `json:"type_sort"`
	Request       *Request `json:"request"`
	//Response      *Response `json:"response"`
	Version      int32        `json:"version"`
	Description  string       `json:"description"`
	Assert       []*Assert    `json:"assert"`
	Regex        []*Regex     `json:"regex"`
	HttpApiSetup HttpApiSetup `json:"http_api_setup" bson:"http_api_setup"`

	// 为了导入接口而新增的一些字段
	TargetType  string `json:"target_type"`
	OldTargetID string `json:"old_target_id"`
	OldParentID string `json:"old_parent_id"`
}

type HttpApiSetup struct {
	IsRedirects  int `json:"is_redirects"`  // 是否跟随重定向 0: 是   1：否
	RedirectsNum int `json:"redirects_num"` // 重定向次数>= 1; 默认为3
	ReadTimeOut  int `json:"read_time_out"` // 请求超时时间
	WriteTimeOut int `json:"write_time_out"`
}

type SaveImportApiReq struct {
	Project Project         `json:"project"`
	Apis    []SaveTargetReq `json:"apis"`
	TeamID  string          `json:"team_id"`
}

type Project struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type SaveTargetResp struct {
	TargetID string `json:"target_id"`
}

type SortTargetReq struct {
	Targets []*SortTarget `json:"targets"`
}

type SortTarget struct {
	TeamID   string `json:"team_id"`
	TargetID string `json:"target_id"`
	Sort     int32  `json:"sort"`
	ParentID string `json:"parent_id"`
}

type SortTargetResp struct {
}

type TrashTargetReq struct {
	TargetID string `json:"target_id" binding:"required,gt=0"`
}

type TrashTargetResp struct {
}

type RecallTargetReq struct {
	TargetID string `json:"target_id" binding:"required,gt=0"`
}

type RecallTargetResp struct {
}

type DeleteTargetReq struct {
	TargetID string `json:"target_id" binding:"required,gt=0"`
}

type DeleteTargetResp struct {
}

type ListTrashTargetReq struct {
	TeamID string `form:"team_id" binding:"required,gt=0"`
	Page   int    `form:"page,default=1"`
	Size   int    `form:"size,default=10"`
}

type ListTrashTargetResp struct {
	Targets []*FolderAPI `json:"targets"`
	Total   int64        `json:"total"`
}

type ListFolderAPIReq struct {
	TeamID string `form:"team_id" binding:"required,gt=0"`
	PlanID int64  `json:"plan_id" form:"plan_id"`
	Source int32  `json:"source" form:"source"`
	//Page   int   `form:"page,default=1"`
	//Size   int   `form:"size,default=10"`
}

type ListFolderAPIResp struct {
	Targets []*FolderAPI `json:"targets"`
	Total   int64        `json:"total"`
}

type FolderAPI struct {
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
	CreatedUserID string `json:"created_user_id"`
	RecentUserID  string `json:"recent_user_id"`
}

type ListGroupSceneReq struct {
	TeamID string `form:"team_id" binding:"required,gt=0"`
	Source int32  `form:"source,default=1"`
	PlanID string `form:"plan_id"`
}

type ListGroupSceneResp struct {
	Targets []*GroupScene `json:"targets"`
	Total   int64         `json:"total"`
}

type GroupScene struct {
	TargetID      string `json:"target_id"`
	TeamID        string `json:"team_id"`
	TargetType    string `json:"target_type"`
	Name          string `json:"name"`
	ParentID      string `json:"parent_id"`
	Method        string `json:"method"`
	Sort          int32  `json:"sort"`
	TypeSort      int32  `json:"type_sort"`
	Version       int32  `json:"version"`
	CreatedUserID string `json:"created_user_id"`
	RecentUserID  string `json:"recent_user_id"`
	Description   string `json:"description"`
}

type BatchGetDetailReq struct {
	TeamID    string   `form:"team_id" binding:"required,gt=0"`
	TargetIDs []string `form:"target_ids" binding:"required,gt=0"`
}

type BatchGetDetailResp struct {
	Targets []*APIDetail `json:"targets"`
}

type APIDetail struct {
	TargetID       string        `json:"target_id"`
	ParentID       string        `json:"parent_id"`
	TargetType     string        `json:"target_type"`
	TeamID         string        `json:"team_id"`
	Name           string        `json:"name"`
	Method         string        `json:"method"`
	EnvServiceID   int64         `json:"env_service_id"`
	EnvServiceURL  string        `json:"env_service_url"`
	URL            string        `json:"url"`
	Sort           int32         `json:"sort"`
	TypeSort       int32         `json:"type_sort"`
	Request        *Request      `json:"request"`
	Response       *Response     `json:"response"`
	Version        int32         `json:"version"`
	Description    string        `json:"description"`
	CreatedTimeSec int64         `json:"created_time_sec"`
	UpdatedTimeSec int64         `json:"updated_time_sec"`
	Assert         []*Assert     `json:"assert"`
	Regex          []*Regex      `json:"regex"`
	Variable       []*KVVariable `json:"variable"`      // 全局变量
	Configuration  Configuration `json:"configuration"` // 场景配置
	HttpApiSetup   HttpApiSetup  `json:"http_api_setup"`
}

type KVVariable struct {
	Key   string `json:"key" bson:"key"`
	Value string `json:"value" bson:"value"`
}
