package rao

type GetCaseAssembleListReq struct {
	SceneID  string `json:"scene_id" binding:"required,gt=0"`
	CaseName string `json:"case_name"`
	Sort     int32  `form:"sort"`
}

type CaseAssembleListResp struct {
	CaseAssembleList []*CaseAssembleDetailResp `json:"case_assemble_list"`
	Total            int64                     `json:"total"`
}

type CaseAssembleDetailResp struct {
	CaseID      string `json:"case_id"`
	TeamID      string `json:"team_id"`
	SceneID     string `json:"scene_id"`
	CaseName    string `json:"case_name"`
	Sort        int32  `json:"sort"`
	IsChecked   int32  `json:"is_checked"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
	Status      int32  `json:"status"`
	Description string `json:"description"`
}

type CopyAssembleReq struct {
	CaseID string `json:"case_id" binding:"required,gt=0"`
	TeamID string `json:"team_id" binding:"required,gt=0"`
}

type CaseAssembleDetailReq struct {
	CaseID string `json:"case_id" binding:"required,gt=0"`
}

type SaveCaseAssembleReq struct {
	CaseID      string `json:"case_id"`
	TeamID      string `json:"team_id" binding:"required,gt=0"`
	SceneID     string `json:"scene_id" binding:"required,gt=0"`
	Name        string `json:"name" binding:"required,min=1"`
	Sort        int32  `json:"sort"`
	Version     int32  `json:"version"`
	Source      int32  `json:"source"`
	PlanID      string `json:"plan_id"`
	Description string `json:"description"`
}

type SaveSceneCaseFlowReq struct {
	SceneID     string `json:"scene_id" binding:"required,gt=0"`
	SceneCaseID string `json:"scene_case_id" binding:"required,gt=0"`
	TeamID      string `json:"team_id" binding:"required,gt=0"`
	Version     int32  `json:"version"`

	Nodes           []Node `json:"nodes"`
	Edges           []Edge `json:"edges"`
	MultiLevelNodes string `json:"multi_level_nodes"`
}

type GetSceneCaseFlowReq struct {
	CaseID string `json:"case_id" binding:"required"`
}

type GetSceneCaseFlowResp struct {
	SceneID         string `json:"scene_id"`
	SceneCaseID     string `json:"scene_case_id"`
	TeamID          string `json:"team_id"`
	Version         int32  `json:"version"`
	Nodes           []Node `json:"nodes"`
	Edges           []Edge `json:"edges"`
	MultiLevelNodes []byte `json:"multi_level_nodes"`
	EnvID           int64  `json:"env_id"`
}

type DelCaseAssembleReq struct {
	CaseID string `json:"case_id" binding:"required,gt=0"`
	TeamID string `json:"team_id" binding:"required,gt=0"`
}

type ChangeCaseAssembleCheckReq struct {
	CaseID    string `json:"case_id" binding:"required,gt=0"`
	TeamID    string `json:"team_id" binding:"required,gt=0"`
	IsChecked int32  `json:"is_checked" binding:"required"`
}

type SendSceneCaseReq struct {
	TeamID      string `json:"team_id" binding:"required,gt=0"`
	SceneID     string `json:"scene_id" binding:"required,gt=0"`
	SceneCaseID string `json:"scene_case_id" binding:"required,gt=0"`
}

type SendSceneCaseResp struct {
	RetID string `json:"ret_id"`
}

type StopSceneCaseReq struct {
	SceneID     string `json:"scene_id" binding:"required,gt=0"`
	SceneCaseID string `json:"scene_case_id" binding:"required,gt=0"`
	TeamID      string `json:"team_id" binding:"required,gt=0"`
}

type StopSceneCaseResp struct {
}

type SceneCaseFlow struct {
	SceneID        string         `json:"scene_id"`
	SceneCaseID    string         `bson:"scene_case_id"`
	SceneCaseName  string         `json:"scene_case_name"`
	TeamID         string         `json:"team_id"`
	Configuration  Configuration  `json:"configuration"`
	Variable       []KVVariable   `json:"variable"` // 全局变量
	NodesRound     [][]Node       `json:"nodes_round"`
	GlobalVariable GlobalVariable `json:"global_variable"`
	Prepositions   []Preposition  `json:"prepositions"` // 前置条件
}

type ChangeCaseSortReq struct {
	CaseList []SortCase `json:"case_list"`
}

type SortCase struct {
	TeamID  string `json:"team_id"`
	CaseID  string `json:"case_id"`
	Sort    int32  `json:"sort"`
	SceneID string `json:"scene_id"`
}
