package mao

import "time"

type DubboDetail struct {
	TargetID      string        `bson:"target_id"`
	TeamID        string        `bson:"team_id"`
	EnvInfo       EnvInfo       `bson:"env_info"`
	ApiName       string        `bson:"api_name"`
	FunctionName  string        `bson:"function_name"`
	DubboProtocol string        `bson:"dubbo_protocol"`
	DubboParam    []DubboParam  `bson:"dubbo_param"`
	DubboAssert   []DubboAssert `bson:"dubbo_assert"`
	DubboRegex    []DubboRegex  `bson:"dubbo_regex"`
	DubboConfig   DubboConfig   `bson:"dubbo_config"`
	CreatedAt     time.Time     `bson:"created_at"`
}

type DubboParam struct {
	IsChecked int32  `bson:"is_checked"`
	ParamType string `bson:"param_type"`
	Var       string `bson:"var"`
	Val       string `bson:"val"`
}

type DubboConfig struct {
	RegistrationCenterName    string `bson:"registration_center_name"`
	RegistrationCenterAddress string `bson:"registration_center_address"`
	Version                   string `bson:"version"`
}

type DubboAssert struct {
	IsChecked    int    `bson:"is_checked"`
	ResponseType int32  `bson:"response_type"`
	Var          string `bson:"var"`
	Compare      string `bson:"compare"`
	Val          string `bson:"val"`
	Index        int    `bson:"index"` // 正则时提取第几个值
}

type DubboRegex struct {
	IsChecked int    `bson:"is_checked"` // 1 选中, -1未选
	Type      int    `bson:"type"`       // 0 正则  1 json
	Var       string `bson:"var"`
	Express   string `bson:"express"`
	Val       string `bson:"val"`
	Index     int    `bson:"index"` // 正则时提取第几个值
}

type DubboDebug struct {
	TargetID     string                   `bson:"api_id"`
	TeamID       string                   `bson:"team_id"`
	Uuid         string                   `bson:"uuid"`
	Name         string                   `bson:"api_name"`
	Status       string                   `bson:"status"`
	RequestType  string                   `bson:"request_type"`
	RequestBody  string                   `bson:"request_body"`
	ResponseBody string                   `bson:"response_body"`
	Assert       []DebugAssertion         `bson:"assert"`
	Regex        []map[string]interface{} `bson:"regex"`
}

type DebugAssertion struct {
	Code      int    `bson:"code"`
	IsSucceed bool   `bson:"is_succeed"`
	Msg       string `bson:"msg"`
	Type      string `bson:"type"`
}
