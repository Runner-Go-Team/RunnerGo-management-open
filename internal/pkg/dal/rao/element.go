package rao

type Element struct {
	ElementID      string          `json:"element_id"`
	ElementType    string          `json:"element_type,omitempty"`
	TeamID         string          `json:"team_id"`
	ParentID       string          `json:"parent_id,omitempty"`
	ParentName     string          `json:"parent_name,omitempty"`
	Name           string          `json:"name"`
	Locators       []*Locator      `json:"locators"  binding:"required"`
	Sort           int32           `json:"sort,omitempty"`
	Version        int32           `json:"version,omitempty"`
	Description    string          `json:"description,omitempty"`
	Source         int32           `json:"source,omitempty"`
	RelateScenes   []*UIScene      `json:"relate_scenes"`
	CreatedTimeSec int64           `json:"created_time_sec,omitempty"`
	UpdatedTimeSec int64           `json:"updated_time_sec,omitempty"`
	Setting        *ElementSetting `json:"setting"`
	TargetType     int32           `json:"target_type,omitempty"` // 1:选择元素  2:自定义元素
	CustomLocators []*Locator      `json:"custom_locators,omitempty"`
}

type ElementSetting struct {
	SyncMode int32 `json:"sync_mode"` // 1.实时同步  2.手动同步
}

type ElementFolder struct {
	ElementID   string `json:"element_id"`
	ElementType string `json:"element_type"`
	TeamID      string `json:"team_id"`
	ParentID    string `json:"parent_id"`
	ParentName  string `json:"parent_name"`
	Name        string `json:"name"`
	Sort        int32  `json:"sort"`
	Version     int32  `json:"version"`
	Description string `json:"description"`
	Source      int32  `json:"source"`
}

type Locator struct {
	ID        string `json:"id"`
	Method    string `json:"method"`
	Type      string `json:"type"`
	Key       string `json:"key"`
	Value     string `json:"value"`
	Index     int32  `json:"index"`
	IsChecked int32  `json:"is_checked"`
}

type ElementSaveFolderReq struct {
	ElementID   string `json:"element_id"`
	TeamID      string `json:"team_id" binding:"required,gt=0"`
	ParentID    string `json:"parent_id"`
	Name        string `json:"name" binding:"required,min=1"`
	Sort        int32  `json:"sort"`
	Version     int32  `json:"version"`
	Description string `json:"description"`
	Source      int32  `json:"source"`
}

type ElementSaveFolderResp struct {
	ElementID string `json:"element_id"`
}

type ElementFolderListReq struct {
	TeamID string `form:"team_id" binding:"required,gt=0"`
}

type ElementFolderListResp struct {
	Folder []*ElementFolder `json:"folders"`
}

type ElementSaveReq struct {
	ElementID   string     `json:"element_id"`
	TeamID      string     `json:"team_id" binding:"required,gt=0"`
	ParentID    string     `json:"parent_id"`
	Name        string     `json:"name" binding:"required,min=1"`
	Locators    []*Locator `json:"locators" binding:"required"`
	Sort        int32      `json:"sort"`
	Version     int32      `json:"version"`
	Description string     `json:"description"`
	Source      int32      `json:"source"`
}

type ElementDetailReq struct {
	ElementID string `form:"element_id" binding:"required,gt=0"`
	TeamID    string `form:"team_id" binding:"required,gt=0"`
}

type ElementDetailResp struct {
	Element *Element `json:"element"`
}

type ElementListReq struct {
	Page          int      `json:"page,default=1"`
	Size          int      `json:"size,default=20"`
	TeamID        string   `json:"team_id" binding:"required,gt=0"`
	ParentID      string   `json:"parent_id"`
	Name          string   `json:"name"`
	UpdatedTime   []string `json:"updated_time"`
	LocatorMethod []string `json:"locator_method"`
	LocatorType   []string `json:"locator_type"`
	LocatorValue  string   `json:"locator_value"`
}

type ElementListResp struct {
	Elements []*Element `json:"elements"`
	Total    int64      `json:"total"`
}

type ElementRemoveFolderReq struct {
	ElementIDs []string `json:"element_ids" binding:"required"`
	TeamID     string   `json:"team_id" binding:"required,gt=0"`
}

type ElementRemoveReq struct {
	ElementIDs []string `json:"element_ids" binding:"required"`
	TeamID     string   `json:"team_id" binding:"required,gt=0"`
}

type SortElement struct {
	TeamID    string `json:"team_id"`
	ElementID string `json:"element_id" binding:"required,gt=0"`
	Sort      int32  `json:"sort"`
	ParentID  string `json:"parent_id"`
	Name      string `json:"name"`
}

type ElementSortFolderReq struct {
	Elements []*SortElement `json:"elements"`
}

type ElementSortReq struct {
	ElementIDs []string `json:"element_ids" binding:"required"`
	TeamID     string   `json:"team_id" binding:"required,gt=0"`
	ParentID   string   `json:"parent_id" binding:"required,gt=0"`
}
