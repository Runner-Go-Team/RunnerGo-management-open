package rao

type SendSceneReq struct {
	TeamID  string `json:"team_id" binding:"required,gt=0"`
	SceneID string `json:"scene_id" binding:"required,gt=0"`
}

type SendSceneResp struct {
	RetID string `json:"ret_id"`
}

type StopSceneReq struct {
	SceneID string `json:"scene_id" binding:"required,gt=0"`
	TeamID  string `json:"team_id" binding:"required,gt=0"`
}

type StopSceneResp struct {
}

type SendSceneAPIReq struct {
	TeamID      string `json:"team_id" binding:"required,gt=0"`
	SceneID     string `json:"scene_id" binding:"required,gt=0"`
	NodeID      string `json:"node_id" binding:"required"`
	SceneCaseID string `json:"scene_case_id"`
}

type SendSceneAPIResp struct {
	RetID string `json:"ret_id"`
}

type GetSendSceneResultReq struct {
	RetID string `form:"ret_id" binding:"required,gt=0"`
}

type GetSendSceneResultResp struct {
	Scenes []*SceneDebug `json:"scenes"`
}

type SaveSceneReq struct {
	//ImportSceneID int64  `json:"import_scene_id"`
	TeamID      string `json:"team_id" binding:"required,gt=0"`
	TargetID    string `json:"target_id"`
	ParentID    string `json:"parent_id"`
	Name        string `json:"name" binding:"required,min=1"`
	Method      string `json:"method"`
	Sort        int32  `json:"sort"`
	TypeSort    int32  `json:"type_sort"`
	Version     int32  `json:"version"`
	Source      int32  `json:"source"`
	PlanID      string `json:"plan_id"`
	Description string `json:"description"`
	//Request  *Request `json:"request"`
	//Script   *Script  `json:"script"`
}

type SaveSceneResp struct {
	TargetID   string `json:"target_id"`
	TargetName string `json:"target_name"`
}

type GetSceneReq struct {
	TeamID   string   `form:"team_id" json:"team_id" binding:"required,gt=0"`
	TargetID []string `form:"target_id" json:"target_id" binding:"required,gt=0"`
	Source   int32    `form:"source" json:"source"`
}

type GetSceneResp struct {
	Scenes []*Scene `json:"scenes"`
}

type Scene struct {
	TeamID      string   `json:"team_id"`
	TargetID    string   `json:"target_id"`
	ParentID    string   `json:"parent_id"`
	Name        string   `json:"name"`
	Method      string   `json:"method"`
	Sort        int32    `json:"sort"`
	TypeSort    int32    `json:"type_sort"`
	Version     int32    `json:"version"`
	Request     *Request `json:"request"`
	Script      *Script  `json:"script"`
	Description string   `json:"description"`
}

type SaveFlowReq struct {
	SceneID         string `json:"scene_id" binding:"required,gt=0"`
	TeamID          string `json:"team_id" binding:"required,gt=0"`
	Version         int32  `json:"version"`
	Nodes           []Node `json:"nodes"`
	Edges           []Edge `json:"edges"`
	MultiLevelNodes string `json:"multi_level_nodes"`
	Prepositions    []Node `json:"prepositions"` // 前置条件
}

type SaveFlowResp struct {
}

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Node struct {
	ID               string   `json:"id"`
	Type             string   `json:"type"`
	IsCheck          bool     `json:"is_check"`
	IsDisabled       int      `json:"is_disabled"` // 0-不禁用，1-禁用该接口
	PositionAbsolute Point    `json:"positionAbsolute"`
	Position         Point    `json:"position"`
	PreList          []string `json:"pre_list"`
	NextList         []string `json:"next_list"`
	Width            int      `json:"width"`
	Height           int      `json:"height"`
	Selected         bool     `json:"selected"`
	Dragging         bool     `json:"dragging"`
	DragHandle       string   `json:"dragHandle"`
	Data             struct {
		ID   string `json:"id"`
		From string `json:"from"`
	} `json:"data"`
	Weight            int       `json:"weight,omitempty"`
	Mode              int       `json:"mode,omitempty"`
	ErrorThreshold    float64   `json:"error_threshold,omitempty"`
	ResponseThreshold int       `json:"response_threshold,omitempty"`
	RequestThreshold  int       `json:"request_threshold,omitempty"`
	PercentAge        int       `json:"percent_age,omitempty"`
	API               APIDetail `json:"api,omitempty"`
	Assets            []string  `json:"assets,omitempty"`  // 全局断言
	WaitMs            int       `json:"wait_ms,omitempty"` // 等待控制器
	Var               string    `json:"var,omitempty"`     // 条件控制器
	Compare           string    `json:"compare,omitempty"`
	Val               string    `json:"val,omitempty"`
	Remark            string    `json:"remark"`
}

type Edge struct {
	ID           string   `json:"id"`
	Source       string   `json:"source"`
	SourceHandle string   `json:"sourceHandle"`
	Target       string   `json:"target"`
	TargetHandle string   `json:"targetHandle"`
	Type         string   `json:"type"`
	Data         EdgeData `json:"data"`
}
type EdgeData struct {
	From string `json:"from"`
}

type GetFlowReq struct {
	SceneID string `form:"scene_id" binding:"required,gt=0"`
	TeamID  string `form:"team_id" binding:"required,gt=0"`
}

type GetFlowResp struct {
	SceneID         string `json:"scene_id"`
	TeamID          string `json:"team_id"`
	Version         int32  `json:"version"`
	Nodes           []Node `json:"nodes"`
	Edges           []Edge `json:"edges"`
	MultiLevelNodes []byte `json:"multi_level_nodes"`
	EnvID           int64  `json:"env_id"`
}

type BatchGetFlowReq struct {
	TeamID  string   `form:"team_id" binding:"required,gt=0"`
	SceneID []string `form:"scene_id" binding:"required"`
}

type BatchGetFlowResp struct {
	Flows []*Flow `json:"flows"`
}

type Flow struct {
	SceneID string `json:"scene_id"`
	TeamID  string `json:"team_id"`
	Version int32  `json:"version"`

	Nodes           []Node `json:"nodes"`
	Edges           []Edge `json:"edges"`
	MultiLevelNodes []byte `json:"multi_level_nodes"`
}

type SceneFlow struct {
	SceneID        string             `json:"scene_id"`
	SceneName      string             `json:"scene_name"`
	TeamID         string             `json:"team_id"`
	Configuration  SceneConfiguration `json:"configuration"`
	Variable       []KVVariable       `json:"variable"` // 全局变量
	NodesRound     [][]Node           `json:"nodes_round"`
	GlobalVariable GlobalVariable     `json:"global_variable"`
	Prepositions   []Preposition      `json:"prepositions"` // 前置条件
}

type Preposition struct {
	Type       string `json:"type"`
	ValueType  string `json:"value_type"`
	Key        string `json:"key"`
	Scope      int32  `json:"scope"`
	JsScript   string `json:"js_script"`
	IsDisabled int    `json:"is_disabled"` // 0-不禁用，1-禁用该接口
	Event      Node   `json:"event"`
}

type SceneConfiguration struct {
	ParameterizedFile SceneVariablePath `json:"parameterizedFile"`
	SceneVariable     GlobalVariable    `json:"scene_variable"`
	//Variable          []*Variable        `json:"variable"` // todo 已废弃
}

type GlobalVariable struct {
	Cookie   Cookie          `json:"cookie"`
	Header   Header          `json:"header"`
	Variable []VarForm       `json:"variable"`
	Assert   []AssertionText `json:"assert"` // 验证的方法(断言)
}

type SceneVariablePath struct {
	Paths []FileList `json:"paths"`
}

// VarForm 参数表
type VarForm struct {
	IsChecked   int64       `json:"is_checked" bson:"is_checked"`
	Type        string      `json:"type" bson:"type"`
	FileBase64  []string    `json:"fileBase64"`
	Key         string      `json:"key" bson:"key"`
	Value       interface{} `json:"value" bson:"value"`
	NotNull     int64       `json:"not_null" bson:"not_null"`
	Description string      `json:"description" bson:"description"`
	FieldType   string      `json:"field_type" bson:"field_type"`
}

// AssertionText 文本断言 0
type AssertionText struct {
	IsChecked    int    `json:"is_checked"`    // 1 选中  -1 未选
	ResponseType int8   `json:"response_type"` //  1:ResponseHeaders; 2:ResponseData; 3: ResponseCode;
	Compare      string `json:"compare"`       // Includes、UNIncludes、Equal、UNEqual、GreaterThan、GreaterThanOrEqual、LessThan、LessThanOrEqual、Includes、UNIncludes、NULL、NotNULL、OriginatingFrom、EndIn
	Var          string `json:"var"`
	Val          string `json:"val"`
}

type Configuration struct {
	ParameterizedFile ParameterizedFile `json:"parameterizedFile" bson:"parameterizedFile"`
	SceneVariable     GlobalVariable    `json:"scene_variable"`
	Variable          []KV              `json:"variable" bson:"variable"`
}

// ParameterizedFile 参数化文件
type ParameterizedFile struct {
	Paths         []FileList    `json:"paths"` // 文件地址
	RealPaths     []string      `json:"real_paths"`
	VariableNames VariableNames `json:"variable_names"` // 存储变量及数据的map
}

type FileList struct {
	IsChecked int64  `json:"is_checked,omitempty"` // 1 开， 2： 关
	Path      string `json:"path,omitempty"`
}

type VariableNames struct {
	VarMapList map[string][]string `json:"var_map_list,omitempty"`
	Index      int                 `json:"index,omitempty"`
}

type ConfVariable struct {
	Var string `json:"Var"`
	Val string `json:"Val"`
}

type DeleteSceneReq struct {
	TargetID string `json:"target_id" binding:"required,gt=0"`
	TeamID   string `json:"team_id"`
	PlanID   string `json:"plan_id"`
	Source   int64  `json:"source"`
}

type ChangeDisabledStatusReq struct {
	TargetID   string `json:"target_id" binding:"required"`
	IsDisabled int32  `json:"is_disabled"`
}

type SendMysqlReq struct {
	TeamID  string `json:"team_id" binding:"required"`
	SceneID string `json:"scene_id" binding:"required"`
	NodeID  string `json:"node_id" binding:"required"`
}

type GetSceneCanSyncDataReq struct {
	TeamID  string `json:"team_id"`
	SceneID string `json:"scene_id"`
	Source  int32  `json:"source"`
}

type ExecSyncSceneDataReq struct {
	TeamID        string          `json:"team_id"`
	SceneID       string          `json:"scene_id"`
	SyncType      int32           `json:"sync_type"`
	Source        int32           `json:"source"`
	SyncSceneInfo []SyncSceneInfo `json:"sync_scene_info"`
}

type SyncSceneInfo struct {
	Source  int32  `json:"source"`
	PlanID  string `json:"plan_id"`
	SceneID string `json:"scene_id"`
}

type ExportSceneReq struct {
	ExportType int32    `json:"export_type" binding:"required"` // 导出类型：1-场景，2-用例，3-场景及用例
	SceneIDs   []string `json:"scene_ids" binding:"required"`
}

type ExportSceneResp struct {
	SceneDetailList []map[string]interface{} `json:"scene_detail_list,omitempty"`
	CaseDetailList  []map[string]interface{} `json:"case_detail_list,omitempty"`
}

type ExportPreposition struct {
	Prepositions []Node `json:"prepositions"`
}

type ExportNode struct {
	Nodes []Node `json:"nodes"`
}

type ExportEdge struct {
	Edges []Edge `json:"edges"`
}
