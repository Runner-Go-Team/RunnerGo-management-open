package mao

type GlobalParamData struct {
	TeamID     string `bson:"team_id" json:"team_id,omitempty"`
	ParamType  int32  `bson:"param_type" json:"param_type,omitempty"`
	DataDetail string `bson:"data_detail" json:"data_detail,omitempty"`
}

type SceneParamData struct {
	TeamID     string `bson:"team_id" json:"team_id,omitempty"`
	SceneID    string `bson:"scene_id" json:"scene_id"`
	ParamType  int32  `bson:"param_type" json:"param_type,omitempty"`
	DataDetail string `bson:"data_detail" json:"data_detail,omitempty"`
}

type CookieParam struct {
	IsChecked int32  `bson:"is_checked" json:"is_checked,omitempty"`
	Key       string `bson:"key" json:"key,omitempty"`
	Value     string `bson:"value" json:"value,omitempty"`
}

type HeaderParam struct {
	IsChecked   int32  `bson:"is_checked" json:"is_checked,omitempty"`
	Key         string `bson:"key" json:"key,omitempty"`
	Value       string `bson:"value" json:"value,omitempty"`
	Description string `bson:"description" json:"description,omitempty"`
}

type VariableParam struct {
	IsChecked   int32  `bson:"is_checked" json:"is_checked,omitempty"`
	Key         string `bson:"key" json:"key,omitempty"`
	Value       string `bson:"value" json:"value,omitempty"`
	Description string `bson:"description" json:"description,omitempty"`
}

type AssertParam struct {
	IsChecked    int32  `bson:"is_checked" json:"is_checked,omitempty"`
	ResponseType int32  `bson:"response_type" json:"response_type,omitempty"`
	Var          string `bson:"var" json:"var,omitempty"`
	Compare      string `bson:"compare" json:"compare,omitempty"`
	Val          string `bson:"val" json:"val,omitempty"`
}
