package rao

type SaveVariableReq struct {
	VarID       int64  `json:"var_id"`
	TeamID      string `json:"team_id" binding:"required,gt=0"`
	Var         string `json:"var" binding:"required,min=1"`
	Val         string `json:"val" binding:"required,min=1"`
	Status      int32  `json:"status"`
	Description string `json:"description"`
}

type SaveVariableResp struct {
}

type DeleteVariableReq struct {
	TeamID string `json:"team_id"`
	VarID  int64  `json:"var_id"`
}

type DeleteVariableResp struct {
}

type SyncVariablesReq struct {
	TeamID    string      `json:"team_id" binding:"required,gt=0"`
	Variables []*Variable `json:"variables"`
}

type SyncVariablesResp struct {
}

type SyncSceneVariablesReq struct {
	TeamID    string      `json:"team_id" binding:"required,gt=0"`
	SceneID   string      `json:"scene_id" binding:"required,gt=0"`
	Variables []*Variable `json:"variables"`
}

type SyncSceneVariablesResp struct {
}

type ListSceneVariablesReq struct {
	TeamID  string `form:"team_id" binding:"required,gt=0"`
	SceneID string `form:"scene_id" binding:"required,gt=0"`
	Page    int    `form:"page"`
	Size    int    `form:"size"`
}

type ListVariablesReq struct {
	TeamID string `form:"team_id" binding:"required,gt=0"`
	Page   int    `form:"page"`
	Size   int    `form:"size"`
}

type ListVariablesResp struct {
	Variables []*Variable `json:"variables"`
	Total     int64       `json:"total"`
}

type Variable struct {
	VarID       int64  `json:"var_id,omitempty"`
	TeamID      string `json:"team_id,omitempty"`
	Var         string `json:"var"`
	Val         string `json:"val"`
	Status      int32  `json:"status"`
	Description string `json:"description"`
}

type ImportVariablesReq struct {
	TeamID  string `json:"team_id" binding:"required,gt=0"`
	SceneID string `json:"scene_id" binding:"required,gt=0"`
	Name    string `json:"name"`
	URL     string `json:"url"`
}

type ImportVariablesResp struct {
}

type DeleteImportSceneVariablesReq struct {
	TeamID  string `json:"team_id" binding:"required,gt=0"`
	SceneID string `json:"scene_id" binding:"required,gt=0"`
	Name    string `json:"name"`
}

type ListImportVariablesReq struct {
	TeamID  string `form:"team_id" binding:"required,gt=0"`
	SceneID string `form:"scene_id" binding:"required,gt=0"`
}

type ListImportVariablesResp struct {
	Imports []*Import `json:"imports"`
}

type Import struct {
	ID             int64  `json:"id"`
	TeamID         string `json:"team_id"`
	SceneID        string `json:"scene_id"`
	Name           string `json:"name"`
	URL            string `json:"url"`
	Status         int32  `json:"status"`
	CreatedTimeSec int64  `json:"created_time_sec"`
}

type UpdateImportSceneVariablesReq struct {
	ID     int64 `json:"id"`
	Status int32 `json:"status"`
}
