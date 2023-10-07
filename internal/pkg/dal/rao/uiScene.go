package rao

type UISceneFolder struct {
	SceneID     string `json:"scene_id"`
	TeamID      string `json:"team_id"`
	ParentID    string `json:"parent_id"`
	Name        string `json:"name"`
	Sort        int32  `json:"sort"`
	Version     int32  `json:"version"`
	Description string `json:"description"`
	SourceID    string `json:"source_id"`
	PlanID      string `json:"plan_id"`
	Source      int32  `json:"source"`
}

type UIScene struct {
	SceneID       string     `json:"scene_id"`
	SceneType     string     `json:"scene_type"`
	TeamID        string     `json:"team_id" binding:"required,gt=0"`
	ParentID      string     `json:"parent_id"`
	Name          string     `json:"name" binding:"required,min=1"`
	Sort          int32      `json:"sort"`
	Version       int32      `json:"version"`
	Description   string     `json:"description"`
	Browsers      []*Browser `json:"browsers,omitempty"`
	UIMachineKey  string     `json:"ui_machine_key"`
	PlanID        string     `json:"plan_id"`
	Source        int32      `json:"source"`
	SyncMode      int32      `json:"sync_mode"` // 状态：1-实时，2-手动,已场景为准   3-手动,已计划为准
	SourceUIScene *UIScene   `json:"source_ui_scene"`
}

type UISceneOperator struct {
	TeamID        string              `json:"team_id" binding:"required,gt=0"`
	SceneID       string              `json:"scene_id"`
	OperatorID    string              `json:"operator_id" binding:"required,gt=0"`
	ParentID      string              `json:"parent_id"`
	Name          string              `json:"name" binding:"required,min=1"`
	Sort          int32               `json:"sort" binding:"required"`
	Status        int32               `json:"status"`
	Type          string              `json:"type"`
	Action        string              `json:"action"`
	ActionDetail  *ActionDetail       `json:"action_detail,omitempty"`
	Settings      *AutomationSettings `json:"settings,omitempty" bson:"settings,omitempty"`
	Asserts       []*AutomationAssert `json:"asserts,omitempty" bson:"asserts,omitempty"`
	DataWithdraws []*DataWithdraw     `json:"data_withdraws" bson:"data_withdraws"`
}

type UISceneTrash struct {
	SceneID         string `json:"scene_id"`
	TeamID          string `json:"team_id" binding:"required,gt=0"`
	Name            string `json:"name" binding:"required,min=1"`
	SceneType       string `json:"scene_type"`
	CreatedUserID   string `json:"created_user_id"`
	CreatedUserName string `json:"created_user_name,omitempty"`
	CreatedTimeSec  int64  `json:"created_time_sec"`
}

type UiEngineMachineInfo struct {
	Key         string                     `json:"key"`
	IP          string                     `json:"ip"`
	Timestamp   float64                    `json:"timestamp"` // 数据上报时间（时间戳）
	CurrentTask int64                      `json:"current_task"`
	SystemInfo  *UiEngineMachineSystemInfo `json:"system_info"`
}

type UiEngineMachineSystemInfo struct {
	Hostname    string `json:"hostname"`
	Machine     string `json:"machine"`
	Processor   string `json:"processor"`
	SystemBasic string `json:"system_basic"`
}

// UIReportStatusChange 订阅压测计划状态变更
type UIReportStatusChange struct {
	Status string `json:"status"`
}

type UISceneSaveFolderReq struct {
	SceneID     string `json:"scene_id"`
	TeamID      string `json:"team_id" binding:"required,gt=0"`
	ParentID    string `json:"parent_id"`
	Name        string `json:"name" binding:"required,min=1"`
	Sort        int32  `json:"sort"`
	Version     int32  `json:"version"`
	Description string `json:"description"`
	PlanID      string `json:"plan_id"`
	Source      int32  `json:"source"`
}

type UISceneGetFolderReq struct {
	TeamID  string `form:"team_id"`
	SceneID string `form:"scene_id"`
}

type UISceneGetFolderResp struct {
	Folder *UISceneFolder `json:"folder"`
}

type UISceneSaveReq struct {
	SceneID      string     `json:"scene_id"`
	TeamID       string     `json:"team_id" binding:"required,gt=0"`
	ParentID     string     `json:"parent_id"`
	Name         string     `json:"name"`
	Sort         int32      `json:"sort"`
	Version      int32      `json:"version"`
	Description  *string    `json:"description"`
	Browsers     []*Browser `json:"browsers"`
	UIMachineKey string     `json:"ui_machine_key"`
	PlanID       string     `json:"plan_id"`
	Source       int32      `json:"source"`
}

type UISceneDetailReq struct {
	TeamID  string `form:"team_id" binding:"required,gt=0"`
	SceneID string `form:"scene_id" binding:"required,gt=0"`
}

type UISceneDetailResp struct {
	Scene *UIScene `json:"scene"`
}

type SendUISceneReq struct {
	SceneID     string   `json:"scene_id" binding:"required,gt=0"`
	TeamID      string   `json:"team_id" binding:"required,gt=0"`
	PlanID      string   `json:"plan_id"`
	OperatorIDs []string `json:"operator_ids"`
}

type SendUISceneResp struct {
	RunID string `json:"run_id"`
}

type StopUISceneReq struct {
	SceneID string `json:"scene_id" binding:"required,gt=0"`
	TeamID  string `json:"team_id" binding:"required,gt=0"`
	RunID   string `json:"run_id" binding:"required,gt=0"`
}

type UISceneListReq struct {
	TeamID string `form:"team_id" binding:"required,gt=0"`
	Source int32  `form:"source"`
	PlanID string `form:"plan_id"`
}

type UISceneListResp struct {
	Scenes []*UIScene `json:"scenes"`
}

type UIScenesSortReq struct {
	Scenes []*UIScene `json:"scenes"`
}

type UISceneCopyReq struct {
	SceneID string `json:"scene_id" binding:"required,gt=0"`
	TeamID  string `json:"team_id" binding:"required"`
}

type UISceneTrashReq struct {
	SceneID string `json:"scene_id" binding:"required,gt=0"`
	TeamID  string `json:"team_id" binding:"required,gt=0"`
}

type UISceneTrashListReq struct {
	TeamID  string `form:"team_id" binding:"required,gt=0"`
	Page    int    `form:"page"`
	Size    int    `form:"size"`
	Keyword string `form:"keyword"`
}

type UISceneTrashListResp struct {
	Total     int64           `json:"total"`
	TrashList []*UISceneTrash `json:"trash_list"`
}

type UISceneRecallReq struct {
	SceneIDs []string `json:"scene_ids" binding:"required"`
	TeamID   string   `json:"team_id" binding:"required,gt=0"`
}

type UISceneDeleteReq struct {
	SceneIDs []string `json:"scene_ids" binding:"required"`
	TeamID   string   `json:"team_id" binding:"required,gt=0"`
}

type UISceneSaveOperatorReq struct {
	TeamID        string              `json:"team_id" binding:"required,gt=0"`
	SceneID       string              `json:"scene_id"`
	OperatorID    string              `json:"operator_id"`
	ParentID      string              `json:"parent_id"`
	Name          string              `json:"name"`
	Sort          int32               `json:"sort"`
	Status        int32               `json:"status"`
	Type          string              `json:"type"`
	Action        string              `json:"action"`
	ActionDetail  *ActionDetail       `json:"action_detail"`
	Settings      *AutomationSettings `json:"settings" bson:"settings"`
	Asserts       []*AutomationAssert `json:"asserts" bson:"asserts"`
	DataWithdraws []*DataWithdraw     `json:"data_withdraws" bson:"data_withdraws"`
	IsReSort      bool                `json:"is_re_sort"`
}

type UISceneDetailOperatorReq struct {
	TeamID     string `form:"team_id" binding:"required,gt=0"`
	SceneID    string `form:"scene_id" binding:"required,gt=0"`
	OperatorID string `form:"operator_id" binding:"required,gt=0"`
}

type UISceneDetailOperatorResp struct {
	Operator *UISceneOperator `json:"operator"`
}

type UISceneOperatorListReq struct {
	SceneID string `form:"scene_id" binding:"required,gt=0"`
	TeamID  string `form:"team_id" binding:"required,gt=0"`
}

type UISceneOperatorListResp struct {
	Operators []*UISceneOperator `json:"operators"`
}

type UIScenesOperatorStepReq struct {
	Operators []*UISceneOperator `json:"operators"`
}

type UISceneCopyOperatorReq struct {
	SceneID    string `json:"scene_id" binding:"required,gt=0"`
	OperatorID string `json:"operator_id" binding:"required,gt=0"`
}

type UISceneSetStatusOperatorReq struct {
	SceneID     string   `json:"scene_id" binding:"required,gt=0"`
	OperatorIDs []string `json:"operator_ids" binding:"required,gt=0"`
	Status      int32    `json:"status"`
	TeamID      string   `json:"team_id" binding:"required,gt=0"`
}

type UISceneDeleteOperatorReq struct {
	SceneID     string   `json:"scene_id" binding:"required,gt=0"`
	OperatorIDs []string `json:"operator_ids" binding:"required,gt=0"`
	TeamID      string   `json:"team_id" binding:"required,gt=0"`
}

type UISceneEngineMachineReq struct {
	Keyword string `form:"keyword"`
}

type UISceneEngineMachineResp struct {
	List []*UiEngineMachineInfo `json:"list"`
}
