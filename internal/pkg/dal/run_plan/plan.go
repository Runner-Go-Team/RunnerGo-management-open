package run_plan

import "sync"

type Stress struct {
	PlanID        string         `json:"plan_id"`
	PlanName      string         `json:"plan_name"`
	ReportID      string         `json:"report_id"`
	MachineNum    int32          `json:"machine_num"`
	TeamID        string         `json:"team_id"`
	ReportName    string         `json:"report_name"`
	ConfigTask    *ConfigTask    `json:"config_task"`
	Variable      []*Variable    `json:"variable"`
	Scene         *Scene         `json:"scene"`
	Partition     int32          `json:"partition"`
	Configuration *Configuration `json:"configuration" bson:"configuration"` // 变量
	IsRun         int            `json:"is_run"`
	Addr          string         `json:"addr"`
}

type Configuration struct {
	ParameterizedFile *ParameterizedFile `json:"parameterizedFile" bson:"parameterizedFile"`
	Variable          []*KV              `json:"variable" bson:"variable"`
	Mu                sync.Mutex         `json:"mu" bson:"mu"`
}

// ParameterizedFile 参数化文件
type ParameterizedFile struct {
	Paths         []FileList     `json:"paths"`
	RealPaths     []string       `json:"real_paths"`
	VariableNames *VariableNames `json:"variable_names"` // 存储变量及数据的map
}

type VariableNames struct {
	VarMapList map[string][]string `json:"var_map_list"`
	Index      int                 `json:"index"`
	Mu         sync.Mutex          `json:"mu"`
}

type ConfigTask struct {
	TaskType    int32     `json:"task_type"`
	Mode        int32     `json:"mode"`
	ControlMode int32     `json:"control_mode"`
	Remark      string    `json:"remark"`
	CronExpr    string    `json:"cron_expr"`
	ModeConf    *ModeConf `json:"mode_conf"`
}

type ModeConf struct {
	ReheatTime       int64 `json:"reheat_time"`
	RoundNum         int64 `json:"round_num"`
	Concurrency      int64 `json:"concurrency"`
	ThresholdValue   int64 `json:"threshold_value"`
	StartConcurrency int64 `json:"start_concurrency"`
	Step             int64 `json:"step"`
	StepRunTime      int64 `json:"step_run_time"`
	MaxConcurrency   int64 `json:"max_concurrency"`
	Duration         int64 `json:"duration"`
	CreatedTimeSec   int64 `json:"created_time_sec"` // 创建时间
}

type Variable struct {
	Var string `json:"Var"`
	Val string `json:"Val"`
}

type Scene struct {
	SceneID                 string              `json:"scene_id"`
	EnablePlanConfiguration bool                `json:"enable_plan_configuration"`
	SceneName               string              `json:"scene_name"`
	TeamID                  string              `json:"team_id"`
	Nodes                   []*Node             `json:"nodes"`
	Configuration           *SceneConfiguration `json:"configuration"`
}

type SceneConfiguration struct {
	ParameterizedFile *SceneVariablePath `json:"parameterizedFile"`
	Variable          []*Variable        `json:"variable"`
}

type SceneVariablePath struct {
	Paths []FileList `json:"paths"`
}

type FileList struct {
	IsChecked int64  `json:"is_checked"` // 1 开， 2： 关
	Path      string `json:"path"`
}

type Assert struct {
	ResponseType int    `json:"response_type"`
	Var          string `json:"var"`
	Compare      string `json:"compare"`
	Val          string `json:"val"`
	IsChecked    int    `json:"is_checked"` // 1 选中, 2未选
}

type Regex struct {
	Var       string `json:"var"`
	Express   string `json:"express"`
	Val       string `json:"val"`
	IsChecked int    `json:"is_checked"` // 1 选中, 2未选
	Type      int    `json:"type"`       // 0 正则  1 json
}

type Request struct {
	PreUrl      string  `json:"pre_url"`
	URL         string  `json:"url"`
	Description string  `json:"description"`
	Auth        *Auth   `json:"auth"`
	Body        *Body   `json:"body"`
	Header      *Header `json:"header"`
	Query       *Query  `json:"query"`
	Cookie      *Cookie `json:"cookie"`
	Resful      *Resful `json:"resful"`
}

type Auth struct {
	Type     string    `json:"type"`
	Kv       *KV       `json:"kv"`
	Bearer   *Bearer   `json:"bearer"`
	Basic    *Basic    `json:"basic"`
	Digest   *Digest   `json:"digest"`
	Hawk     *Hawk     `json:"hawk"`
	Awsv4    *AwsV4    `json:"awsv4"`
	Ntlm     *Ntlm     `json:"ntlm"`
	Edgegrid *Edgegrid `json:"edgegrid"`
	Oauth1   *Oauth1   `json:"oauth1"`
}

type Bearer struct {
	Key string `json:"key"`
}

type KV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Basic struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Digest struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	Realm     string `json:"realm"`
	Nonce     string `json:"nonce"`
	Algorithm string `json:"algorithm"`
	Qop       string `json:"qop"`
	Nc        string `json:"nc"`
	Cnonce    string `json:"cnonce"`
	Opaque    string `json:"opaque"`
}

type Hawk struct {
	AuthID             string `json:"authId"`
	AuthKey            string `json:"authKey"`
	Algorithm          string `json:"algorithm"`
	User               string `json:"user"`
	Nonce              string `json:"nonce"`
	ExtraData          string `json:"extraData"`
	App                string `json:"app"`
	Delegation         string `json:"delegation"`
	Timestamp          string `json:"timestamp"`
	IncludePayloadHash int    `json:"includePayloadHash"`
}

type AwsV4 struct {
	AccessKey          string `json:"accessKey"`
	SecretKey          string `json:"secretKey"`
	Region             string `json:"region"`
	Service            string `json:"service"`
	SessionToken       string `json:"sessionToken"`
	AddAuthDataToQuery int    `json:"addAuthDataToQuery"`
}

type Ntlm struct {
	Username            string `json:"username"`
	Password            string `json:"password"`
	Domain              string `json:"domain"`
	Workstation         string `json:"workstation"`
	DisableRetryRequest int    `json:"disableRetryRequest"`
}

type Edgegrid struct {
	AccessToken   string `json:"accessToken"`
	ClientToken   string `json:"clientToken"`
	ClientSecret  string `json:"clientSecret"`
	Nonce         string `json:"nonce"`
	Timestamp     string `json:"timestamp"`
	BaseURi       string `json:"baseURi"`
	HeadersToSign string `json:"headersToSign"`
}

type Oauth1 struct {
	ConsumerKey          string `json:"consumerKey"`
	ConsumerSecret       string `json:"consumerSecret"`
	SignatureMethod      string `json:"signatureMethod"`
	AddEmptyParamsToSign int    `json:"addEmptyParamsToSign"`
	IncludeBodyHash      int    `json:"includeBodyHash"`
	AddParamsToHeader    int    `json:"addParamsToHeader"`
	Realm                string `json:"realm"`
	Version              string `json:"version"`
	Nonce                string `json:"nonce"`
	Timestamp            string `json:"timestamp"`
	Verifier             string `json:"verifier"`
	Callback             string `json:"callback"`
	TokenSecret          string `json:"tokenSecret"`
	Token                string `json:"token"`
}

type Body struct {
	Mode      string       `json:"mode"`
	Parameter []*Parameter `json:"parameter"`
	Raw       string       `json:"raw"`
}

type Query struct {
	Parameter []*Parameter `json:"parameter"`
}

type Header struct {
	Parameter []*Parameter `json:"parameter"`
}

type Cookie struct {
	Parameter []*Parameter `json:"parameter"`
}

type Resful struct {
	Parameter []*Parameter `json:"parameter"`
}

type Parameter struct {
	IsChecked   int32  `json:"is_checked"`
	Type        string `json:"type"`
	Key         string `json:"key"`
	Value       string `json:"value"`
	NotNull     int32  `json:"not_null"`
	Description string `json:"description"`
}

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Nodes struct {
	Nodes []*Node `bson:"nodes"`
}

type Node struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	IsCheck bool   `json:"is_check"`

	PositionAbsolute *Point   `json:"positionAbsolute"`
	Position         *Point   `json:"position"`
	PreList          []string `json:"pre_list"`
	NextList         []string `json:"next_list"`
	Width            int      `json:"width"`
	Height           int      `json:"height"`
	Selected         bool     `json:"selected"`
	Dragging         bool     `json:"dragging"`
	DragHandle       string   `json:"drag_handle"`
	Data             struct {
		ID   string `json:"id"`
		From string `json:"from"`
	} `json:"data"`

	// 接口
	Weight            int        `json:"weight,omitempty"`
	Mode              int        `json:"mode,omitempty"`
	ErrorThreshold    float64    `json:"error_threshold,omitempty"`
	ResponseThreshold int        `json:"response_threshold,omitempty"`
	RequestThreshold  int        `json:"request_threshold,omitempty"`
	PercentAge        int        `json:"percent_age,omitempty"`
	API               *APIDetail `json:"api,omitempty"`

	// 全局断言
	Assets []string `json:"assets,omitempty"`

	// 等待控制器
	WaitMs int `json:"wait_ms,omitempty"`

	// 条件控制器
	Var     string `json:"var,omitempty"`
	Compare string `json:"compare,omitempty"`
	Val     string `json:"val,omitempty"`

	Controller *Controller `json:"controller"`
}

type APIDetail struct {
	TargetID   string `json:"target_id"`
	Name       string `json:"name"`
	TargetType string `json:"target_type"`
	Method     string `json:"method"`
	//Debug      bool      `json:"debug"`
	Assert       []*Assert    `json:"assert"`
	Regex        []*Regex     `json:"regex"`
	URL          string       `json:"url"`
	Request      *Request     `json:"request"`
	Timeout      int          `json:"timeout"`
	HttpApiSetup HttpApiSetup `json:"http_api_setup"`
}

type HttpApiSetup struct {
	IsRedirects  int `json:"is_redirects"`  // 是否跟随重定向 0: 是   1：否
	RedirectsNum int `json:"redirects_num"` // 重定向次数>= 1; 默认为3
	ReadTimeOut  int `json:"read_time_out"` // 请求超时时间
	WriteTimeOut int `json:"write_time_out"`
}

type Controller struct {
	ControllerType string `json:"controllerType"`
	WaitController struct {
		Name     string `json:"name"`
		WaitTime string `json:"waitTime"`
	} `json:"waitController"`
}

type Task struct {
	PlanID      string    `json:"plan_id"`
	SceneID     string    `json:"scene_id"`
	TaskType    int32     `json:"task_type"`
	TaskMode    int32     `json:"task_mode"`
	ControlMode int32     `json:"control_mode"`
	ModeConf    *ModeConf `json:"mode_conf"`
}
