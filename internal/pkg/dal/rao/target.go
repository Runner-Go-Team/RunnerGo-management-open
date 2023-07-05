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
	TargetID     string       `json:"target_id"`
	ParentID     string       `json:"parent_id"`
	TeamID       string       `json:"team_id" binding:"required,gt=0"`
	Mark         string       `json:"mark"`
	Name         string       `json:"name" binding:"required,min=1"`
	Method       string       `json:"method" binding:"required"`
	PreUrl       string       `json:"pre_url"`
	URL          string       `json:"url"`
	Sort         int32        `json:"sort"`
	TypeSort     int32        `json:"type_sort"`
	Request      Request      `json:"request"`
	Source       int32        `json:"source"`
	Version      int32        `json:"version"`
	Description  string       `json:"description"`
	Assert       []Assert     `json:"assert"` // 断言
	Regex        []Regex      `json:"regex"`  // 关联提取
	HttpApiSetup HttpApiSetup `json:"http_api_setup" bson:"http_api_setup"`
	EnvInfo      EnvInfo      `json:"env_info"`
	// 为了导入接口而新增的一些字段
	TargetType  string `json:"target_type"`
	OldTargetID string `json:"old_target_id"`
	OldParentID string `json:"old_parent_id"`
	// 更多协议的调试元素
	SqlDetail       SqlDetail       `json:"sql_detail"`       // mysql详情数据
	TcpDetail       TcpDetail       `json:"tcp_detail"`       // tcp详情数据
	WebsocketDetail WebsocketDetail `json:"websocket_detail"` // Websocket详情数据
	MqttDetail      MqttDetail      `json:"mqtt_detail"`      // Mqtt详情数据
	DubboDetail     DubboDetail     `json:"dubbo_detail"`     // Dubbo详情数据
	SourceID        string          `json:"source_id"`
}

type DubboDetail struct {
	ApiName       string        `json:"api_name"`
	FunctionName  string        `json:"function_name"`
	DubboProtocol string        `json:"dubbo_protocol"`
	DubboParam    []DubboParam  `json:"dubbo_param"`
	DubboAssert   []DubboAssert `json:"dubbo_assert"`
	DubboRegex    []DubboRegex  `json:"dubbo_regex"`
	DubboConfig   DubboConfig   `json:"dubbo_config"`
}

type DubboConfig struct {
	RegistrationCenterName    string `json:"registration_center_name"`
	RegistrationCenterAddress string `json:"registration_center_address"`
	Version                   string `json:"version"`
}

type DubboAssert struct {
	IsChecked    int    `json:"is_checked"`
	ResponseType int32  `json:"response_type"`
	Var          string `json:"var"`
	Compare      string `json:"compare"`
	Val          string `json:"val"`
	Index        int    `json:"index"` // 正则时提取第几个值
}

type DubboRegex struct {
	IsChecked int    `json:"is_checked"` // 1 选中, -1未选
	Type      int    `json:"type"`       // 0 正则  1 json
	Var       string `json:"var"`
	Express   string `json:"express"`
	Val       string `json:"val"`
	Index     int    `json:"index"` // 正则时提取第几个值
}

type DubboParam struct {
	IsChecked int32  `json:"is_checked"`
	ParamType string `json:"param_type"`
	Var       string `json:"var"`
	Val       string `json:"val"`
}

type MqttDetail struct {
	Topic        string       `json:"topic"`
	SendMessage  string       `json:"send_message"`
	CommonConfig CommonConfig `json:"common_config"`
	HigherConfig HigherConfig `json:"higher_config"`
	WillConfig   WillConfig   `json:"will_config"`
}

type CommonConfig struct {
	ClientName string   `json:"client_name"` // 客户端名称
	Username   string   `json:"username"`    // 用户名
	Password   string   `json:"password"`    // 密码
	IsEncrypt  bool     `json:"is_encrypt"`  // 是否开启加密
	AuthFile   AuthFile `json:"auth_file"`   // 认证文件
}
type AuthFile struct {
	FileName string `json:"file_name"`
	FileUrl  string `json:"file_url"`
}

type HigherConfig struct {
	ConnectTimeoutTime  int    `json:"connect_timeout_time"`   // 连接超时时间，单位：秒
	KeepAliveTime       int    `json:"keep_alive_time"`        // 保持连接时长，单位：秒
	IsAutoRetry         bool   `json:"is_auto_retry"`          // 是否开启自动重连，true-开启，false-关闭
	RetryNum            int    `json:"retry_num"`              // 重连次数
	RetryInterval       int    `json:"retry_interval"`         // 重连间隔时间，单位：秒
	MqttVersion         string `json:"mqtt_version"`           // mqtt版本号
	DialogueTimeout     int    `json:"dialogue_timeout"`       // 会话过期时间，单位：秒
	IsSaveMessage       bool   `json:"is_save_message"`        // 是否保留消息
	ServiceQuality      int    `json:"service_quality"`        // 服务质量:0-至多一次，1-至少一次，2-确保只有一次
	SendMsgIntervalTime int    `json:"send_msg_interval_time"` // 发送消息间隔时间，单位：秒
}

type WillConfig struct {
	WillTopic      string `json:"will_topic"`      // 遗愿主题
	IsOpenWill     bool   `json:"is_open_will"`    // 是否开启遗愿
	ServiceQuality int    `json:"service_quality"` // 服务质量:0-至多一次，1-至少一次，2-确保只有一次
}

type EnvInfo struct {
	EnvID       int64  `json:"env_id"`
	EnvName     string `json:"env_name"`
	ServiceID   int64  `json:"service_id"`
	ServiceName string `json:"service_name"`
	PreUrl      string `json:"pre_url"`
	DatabaseID  int64  `json:"database_id"`
	ServerName  string `json:"server_name"`
}

type WebsocketDetail struct {
	Url         string    `json:"url"`
	SendMessage string    `json:"send_message"`
	MessageType string    `json:"message_type"`
	WsHeader    []WsQuery `json:"ws_header"`
	WsParam     []WsQuery `json:"ws_param"`
	WsEvent     []WsQuery `json:"ws_event"`
	WsConfig    WsConfig  `json:"ws_config"`
}

type WsConfig struct {
	ConnectType         int `json:"connect_type"`           // 连接类型：1-长连接，2-短连接
	IsAutoSend          int `json:"is_auto_send"`           // 是否自动发送消息：0-非自动，1-自动
	ConnectDurationTime int `json:"connect_duration_time"`  // 连接持续时长，单位：秒
	SendMsgDurationTime int `json:"send_msg_duration_time"` // 发送消息间隔时长，单位：毫秒
	ConnectTimeoutTime  int `json:"connect_timeout_time"`   // 连接超时时间，单位：毫秒
	RetryNum            int `json:"retry_num"`              // 重连次数
	RetryInterval       int `json:"retry_interval"`         // 重连间隔时间，单位：毫秒
}

type WsQuery struct {
	IsChecked int32  `json:"is_checked"`
	Var       string `json:"var"`
	Val       string `json:"val"`
}

type TcpDetail struct {
	Url         string    `json:"url"`
	MessageType string    `json:"message_type"`
	SendMessage string    `json:"send_message"`
	TcpConfig   TcpConfig `json:"tcp_config"`
}

type TcpConfig struct {
	ConnectType         int `json:"connect_type"`           // 连接类型：1-长连接，2-短连接
	IsAutoSend          int `json:"is_auto_send"`           // 是否自动发送消息：0-非自动，1-自动
	ConnectDurationTime int `json:"connect_duration_time"`  // 连接持续时长，单位：秒
	SendMsgDurationTime int `json:"send_msg_duration_time"` // 发送消息间隔时长，单位：毫秒
	ConnectTimeoutTime  int `json:"connect_timeout_time"`   // 连接超时时间，单位：毫秒
	RetryNum            int `json:"retry_num"`              // 重连次数
	RetryInterval       int `json:"retry_interval"`         // 重连间隔时间，单位：毫秒
}

type SqlDetail struct {
	SqlString       string          `json:"sql_string"`        // sql语句
	Assert          []SqlAssert     `json:"assert"`            // 断言
	Regex           []SqlRegex      `json:"regex"`             // 关联提取
	SqlDatabaseInfo SqlDatabaseInfo `json:"sql_database_info"` // 使用的数据库信息
}

type SqlAssert struct {
	IsChecked int    `json:"is_checked"`
	Field     string `json:"field"`
	Compare   string `json:"compare"`
	Val       string `json:"val"`
	Index     int    `json:"index"` // 断言时提取第几个值
}

type SqlRegex struct {
	IsChecked int    `json:"is_checked"` // 1 选中, -1未选
	Var       string `json:"var"`
	Field     string `json:"field"`
	Index     int    `json:"index"` // 正则时提取第几个值
}

type SqlDatabaseInfo struct {
	Type       string `json:"type"`
	ServerName string `json:"server_name"`
	Host       string `json:"host"`
	User       string `json:"user"`
	Password   string `json:"password"`
	Port       int32  `json:"port"`
	DbName     string `json:"db_name"`
	Charset    string `json:"charset"`
}

type HttpApiSetup struct {
	IsRedirects         int    `json:"is_redirects"`  // 是否跟随重定向 0: 是   1：否
	RedirectsNum        int    `json:"redirects_num"` // 重定向次数>= 1; 默认为3
	ReadTimeOut         int    `json:"read_time_out"` // 请求超时时间
	WriteTimeOut        int    `json:"write_time_out"`
	ClientName          string `json:"client_name"`
	KeepAlive           bool   `json:"keep_alive"`
	MaxIdleConnDuration int32  `json:"max_idle_conn_duration"`
	MaxConnPerHost      int32  `json:"max_conn_per_host"`
	UserAgent           bool   `json:"user_agent"`
	MaxConnWaitTimeout  int64  `json:"max_conn_wait_timeout"`
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
	Name     string `json:"name"`
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
	Source        int32  `json:"source"`
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
	Source        int32  `json:"source"`
	IsDisabled    int32  `json:"is_disabled"`
}

type BatchGetDetailReq struct {
	TeamID    string   `form:"team_id" binding:"required,gt=0"`
	TargetIDs []string `form:"target_ids" binding:"required,gt=0"`
}

type BatchGetDetailResp struct {
	Targets []APIDetail `json:"targets"`
}

type APIDetail struct {
	TargetID        string          `json:"target_id"`
	ParentID        string          `json:"parent_id"`
	TargetType      string          `json:"target_type"`
	TeamID          string          `json:"team_id"`
	Name            string          `json:"name"`
	Method          string          `json:"method"`
	URL             string          `json:"url"`
	Sort            int32           `json:"sort"`
	TypeSort        int32           `json:"type_sort"`
	Request         Request         `json:"request"`
	Version         int32           `json:"version"`
	Description     string          `json:"description"`
	CreatedTimeSec  int64           `json:"created_time_sec"`
	UpdatedTimeSec  int64           `json:"updated_time_sec"`
	Variable        []KVVariable    `json:"variable"`      // 全局变量
	Configuration   Configuration   `json:"configuration"` // 场景配置
	GlobalVariable  GlobalVariable  `json:"global_variable"`
	SqlDetail       SqlDetail       `json:"sql_detail"`       // mysql数据库详情
	TcpDetail       TcpDetail       `json:"tcp_detail"`       // tcp数据库详情
	WebsocketDetail WebsocketDetail `json:"websocket_detail"` // websocket数据库详情
	MqttDetail      MqttDetail      `json:"mqtt_detail"`      // mqtt数据库详情
	DubboDetail     DubboDetail     `json:"dubbo_detail"`     // dubbo数据库详情
	EnvInfo         EnvInfo         `json:"env_info"`
}

type KVVariable struct {
	Key   string `json:"key" bson:"key"`
	Value string `json:"value" bson:"value"`
}

type GetSqlDatabaseListReq struct {
	TeamID string `json:"team_id"`
	EnvID  int64  `json:"env_id"`
}

type GetSqlDatabaseListResp struct {
	MysqlID    int64  `json:"mysql_id"`
	Type       string `json:"type"`
	ServerName string `json:"server_name"`
	Host       string `json:"host"`
	Port       int32  `json:"port"`
	User       string `json:"user"`
	Password   string `json:"password"`
	DbName     string `json:"db_name"`
	Charset    string `json:"charset"`
}

type GetAllEnvListReq struct {
	TeamID string `json:"team_id"`
}

type GetAllEnvListResp struct {
	EnvID   int64  `json:"env_id"`
	TeamID  string `json:"team_id"`
	EnvName string `json:"env_name"`
}

type SendSqlReq struct {
	TargetID string `json:"target_id" binding:"required"`
	TeamID   string `json:"team_id" binding:"required"`
}

type SendSqlResp struct {
	RetID string `json:"ret_id"`
}

type ConnectionDatabaseReq struct {
	Type     string `json:"type"`
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
	Port     int32  `json:"port"`
	DbName   string `json:"db_name"`
	Charset  string `json:"charset"`
}

type GetSendSqlResultReq struct {
	RetID      string `json:"ret_id"`
	TargetType string `json:"target_type"`
}

type SendTcpReq struct {
	TargetID string `json:"target_id" binding:"required"`
	TeamID   string `json:"team_id" binding:"required"`
}

type SendTcpResp struct {
	RetID string `json:"ret_id"`
}

type GetSendTcpResultReq struct {
	RetID string `json:"ret_id"`
}

type TcpSendOrStopMessageReq struct {
	RetID                  string                 `json:"ret_id"`
	ConnectionStatusChange ConnectionStatusChange `json:"connection_status_change"`
}

type GetSendTcpResultResp struct {
	TargetID     string `json:"target_id"`
	TeamID       string `json:"team_id"`
	Uuid         string `json:"uuid"`
	Name         string `json:"name"`
	IsStop       bool   `json:"is_stop"`
	Type         string `json:"type"`
	RequestBody  string `json:"request_body"`
	ResponseBody string `json:"response_body"`
	Status       string `json:"status"`
}

type SendWebsocketReq struct {
	TargetID string `json:"target_id" binding:"required"`
	TeamID   string `json:"team_id" binding:"required"`
}

type GetSendWebsocketResultReq struct {
	RetID string `json:"ret_id"`
}

type WsSendOrStopMessageReq struct {
	RetID                  string                 `json:"ret_id"`
	ConnectionStatusChange ConnectionStatusChange `json:"connection_status_change"`
}

type ConnectionStatusChange struct {
	Type        int32  `json:"type"` // 1: 断开连接； 2： 发送消息
	MessageType string `json:"message_type"`
	Message     string `json:"message"`
}

type GetSendWebsocketResultResp struct {
	TargetID            string `json:"target_id"`
	TeamID              string `json:"team_id"`
	Uuid                string `json:"uuid"`
	Name                string `json:"name"`
	IsStop              bool   `json:"is_stop"`
	Type                string `json:"type"`
	Status              string `json:"status"`
	RequestBody         string `json:"request_body"`
	ResponseBody        string `json:"response_body"`
	ResponseMessageType int32  `json:"response_message_type"`
}

type SendDubboReq struct {
	TargetID string `json:"target_id" binding:"required"`
	TeamID   string `json:"team_id" binding:"required"`
}

type RunDubboParam struct {
	TargetID       string         `json:"target_id"`
	Name           string         `json:"name"`
	TeamID         string         `json:"team_id"`
	Debug          string         `json:"debug"`
	DubboProtocol  string         `json:"dubbo_protocol"`
	ApiName        string         `json:"api_name"`
	FunctionName   string         `json:"function_name"`
	Version        string         `json:"version"`
	DubboParam     []DubboParam   `json:"dubbo_param"`
	DubboAssert    []DubboAssert  `json:"dubbo_assert"`
	DubboRegex     []DubboRegex   `json:"dubbo_regex"`
	DubboConfig    DubboConfig    `json:"dubbo_config"`
	GlobalVariable GlobalVariable `json:"global_variable"` // 全局变量
}

type GetSendDubboResultReq struct {
	RetID string `json:"ret_id"`
}

type GetSendDubboResultResp struct {
	TargetID     string                   `json:"target_id"`
	TeamID       string                   `json:"team_id"`
	Uuid         string                   `json:"uuid"`
	Name         string                   `json:"name"`
	RequestType  string                   `json:"request_type"`
	RequestBody  string                   `json:"request_body"`
	ResponseBody string                   `json:"response_body"`
	Assert       []DebugAssert            `json:"assert"`
	Regex        []map[string]interface{} `json:"regex"`
	Status       string                   `json:"status"`
}

type SendMqttReq struct {
	TargetID string `json:"target_id" binding:"required"`
	TeamID   string `json:"team_id" binding:"required"`
}

type RunMqttParam struct {
	TargetID       string         `json:"target_id"`
	Name           string         `json:"name" bson:"name"`
	TeamID         string         `json:"team_id" bson:"team_id"`
	TargetType     string         `json:"target_type" bson:"target_type"` // api/webSocket/tcp/grpc
	MQTTConfig     MqttDetail     `json:"mqtt_config"`
	GlobalVariable GlobalVariable `json:"global_variable"` // 全局变量
}

type GetSendMqttResultReq struct {
	RetID string `json:"ret_id"`
}

type GetSendMqttResultResp struct {
	TargetID      string  `json:"target_id"`
	TeamID        string  `json:"team_id"`
	Uuid          string  `json:"uuid"`
	Name          string  `json:"name"`
	RequestTime   int64   `json:"request_time"`
	ErrorType     int64   `json:"error_type"`
	IsSucceed     bool    `json:"is_succeed"`
	SendBytes     float64 `json:"send_bytes"`
	ReceivedBytes float64 `json:"received_bytes"`
	ErrorMsg      string  `json:"error_msg"`
	Timestamp     int64   `json:"timestamp"`
	StartTime     int64   `json:"start_time"`
	EndTime       int64   `json:"end_time"`
}

type RunTargetParam struct {
	TargetID        string          `json:"target_id"`
	ParentID        string          `json:"parent_id"`
	TargetType      string          `json:"target_type"`
	TeamID          string          `json:"team_id"`
	Name            string          `json:"name"`
	Request         Request         `json:"request"`
	Configuration   Configuration   `json:"configuration"` // 场景配置
	GlobalVariable  GlobalVariable  `json:"global_variable"`
	SqlDetail       SqlDetail       `json:"sql_detail"`   // mysql数据库详情
	TcpDetail       TcpDetail       `json:"tcp_detail"`   // tcp数据库详情
	WebsocketDetail WebsocketDetail `json:"ws_detail"`    // websocket数据库详情
	DubboDetail     DubboDetail     `json:"dubbo_detail"` // dubbo数据库详情
}
