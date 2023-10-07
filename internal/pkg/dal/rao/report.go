package rao

type ListUnderwayReportReq struct {
	TeamID string `form:"team_id" binding:"required,gt=0"`
	Page   int    `form:"page,default=1"`
	Size   int    `form:"size,default=10"`
}

type ListUnderwayReportResp struct {
	Reports []*StressPlanReport `json:"reports"`
	Total   int64               `json:"total"`
}

type ListReportsReq struct {
	TeamID string `form:"team_id" binding:"required,gt=0"`
	Page   int    `form:"page,default=1"`
	Size   int    `form:"size,default=10"`

	Keyword      string `form:"keyword"`
	TaskType     int32  `form:"task_type"`
	TaskMode     int32  `form:"task_mode"`
	Status       int32  `form:"status"`
	StartTimeSec int64  `form:"start_time_sec"`
	EndTimeSec   int64  `form:"end_time_sec"`
	Sort         int32  `form:"sort"`
}

type ListReportsResp struct {
	Reports []*StressPlanReport `json:"reports"`
	Total   int64               `json:"total"`
}

type StressPlanReport struct {
	ReportID    string `json:"report_id"`
	ReportName  string `json:"report_name"`
	RankID      int64  `json:"rank_id"`
	TeamID      string `json:"team_id"`
	TaskMode    int32  `json:"task_mode"`
	TaskType    int32  `json:"task_type"`
	Status      int32  `json:"status"`
	RunTimeSec  int64  `json:"run_time_sec"`
	LastTimeSec int64  `json:"last_time_sec"`
	RunUserID   string `json:"run_user_id"`
	RunUserName string `json:"run_user_name"`
	PlanID      string `json:"plan_id"`
	PlanName    string `json:"plan_name"`
	SceneID     string `json:"scene_id"`
	SceneName   string `json:"scene_name"`
}

type DeleteReportReq struct {
	TeamID   string `json:"team_id"`
	ReportID string `json:"report_id"`
}

type DeleteReportResp struct {
}

type StopReportReq struct {
	TeamID    string   `json:"team_id"`
	PlanID    string   `json:"plan_id"`
	ReportIDs []string `json:"report_ids"`
}

type StopReportResp struct {
}

type ListMachineReq struct {
	ReportID string `form:"report_id" binding:"required,gt=0" json:"report_id"`
	TeamID   string `form:"team_id" json:"team_id"`
	PlanID   string `form:"plan_id" json:"plan_id"`
}

type ListMachineResp struct {
	StartTimeSec int64    `json:"start_time_sec"`
	EndTimeSec   int64    `json:"end_time_sec"`
	ReportStatus int32    `json:"report_status"`
	Metrics      []Metric `json:"metrics"`
}

type Metric struct {
	IP          string          `json:"ip"`
	MachineName string          `json:"machine_name"`
	Region      string          `json:"region"`
	Concurrency int64           `json:"concurrency"`
	CPU         [][]interface{} `json:"cpu"`
	Mem         [][]interface{} `json:"mem"`
	NetIO       [][]interface{} `json:"net_io"`
	DiskIO      [][]interface{} `json:"disk_io"`
}

type GetReportReq struct {
	TeamID   string `form:"team_id" json:"team_id"`
	PlanID   string `form:"plan_id" json:"plan_id"`
	ReportID string `form:"report_id" json:"report_id"`
}

type GetReportResp struct {
	Report *ReportTask `json:"report"`
}

type GetReportTaskDetailReq struct {
	TeamID   string `form:"team_id" json:"team_id"`
	ReportID string `form:"report_id" json:"report_id"`
}

type GetReportTaskDetailResp struct {
	Report *ReportTask `json:"report"`
}

type ReportTask struct {
	UserID            string           `json:"user_id"`
	UserName          string           `json:"user_name"`
	UserAvatar        string           `json:"user_avatar"`
	PlanID            string           `json:"plan_id"`
	PlanName          string           `json:"plan_name"`
	SceneID           string           `json:"scene_id"`
	SceneName         string           `json:"scene_name"`
	ReportID          string           `json:"report_id"`
	ReportName        string           `json:"report_name"`
	CreatedTimeSec    int64            `json:"created_time_sec"`
	TaskType          int32            `json:"task_type"`
	TaskMode          int32            `json:"task_mode"`
	ControlMode       int32            `json:"control_mode"` // 控制模式
	DebugMode         string           `json:"debug_mode"`
	TaskStatus        int32            `json:"task_status"`
	RunDurationTime   int64            `json:"run_duration_time"` // 报告运行时长
	ModeConf          ModeConf         `json:"mode_conf"`
	ChangeTakeConf    []ChangeTakeConf `json:"change_take_conf"`
	IsOpenDistributed int32            `json:"is_open_distributed"` // 是否开启分布式调度：0-关闭，1-开启
	MachineAllotType  int32            `json:"machine_allot_type"`  // 机器分配方式：0-权重，1-自定义
}

type DebugSettingReq struct {
	ReportID string `json:"report_id"`
	PlanID   string `json:"plan_id"`
	TeamID   string `json:"team_id"`
	Setting  string `json:"setting"`
}

type ReportEmailReq struct {
	TeamID   string   `json:"team_id"`
	ReportID string   `json:"report_id"`
	Emails   []string `json:"emails"`
}

type ReportEmailResp struct {
}

type ChangeTaskConfReq struct {
	ReportID                string                  `json:"report_id"`
	PlanID                  string                  `json:"plan_id"`
	TeamID                  string                  `json:"team_id"`
	ModeConf                *ModeConf               `json:"mode_conf"`
	IsOpenDistributed       int32                   `json:"is_open_distributed"`        // 是否开启分布式调度：0-关闭，1-开启
	MachineDispatchModeConf MachineDispatchModeConf `json:"machine_dispatch_mode_conf"` // 分布式压力机配置
}

type CompareReportReq struct {
	TeamID    string   `json:"team_id"`
	ReportIDs []string `json:"report_ids"`
}

type UpdateDescriptionReq struct {
	ReportID    string `json:"report_id"`
	TeamID      string `json:"team_id"`
	Description string `json:"description"`
}

type BatchDeleteReportReq struct {
	ReportIDs []string `json:"report_ids"`
	TeamID    string   `json:"team_id"`
}

type UpdateReportNameReq struct {
	ReportID   string `json:"report_id"`
	ReportName string `json:"report_name"`
}

type ResultData struct {
	End           bool                     `json:"end" bson:"end"`
	ReportId      string                   `json:"report_id" bson:"report_id"`
	ReportName    string                   `json:"report_name" bson:"report_name"`
	PlanId        string                   `json:"plan_id" bson:"plan_id"`     // 任务ID
	PlanName      string                   `json:"plan_name" bson:"plan_name"` //
	SceneId       string                   `json:"scene_id" bson:"scene_id"`   // 场景
	SceneName     string                   `json:"scene_name" bson:"scene_name"`
	Results       map[string]ResultDataMsg `json:"results" bson:"results"`
	TimeStamp     int64                    `json:"time_stamp" bson:"time_stamp"`
	Analysis      string                   `json:"analysis" bson:"analysis"`
	Description   string                   `json:"description" bson:"description"`
	ReportRunTime int64                    `json:"report_run_time" bson:"report_run_time"`
	Msg           string                   `json:"msg" bson:"msg"`
}

type ResultDataMsg struct {
	ApiName                        string      `json:"api_name" bson:"api_name"`
	Concurrency                    int64       `json:"concurrency" bson:"concurrency"`
	TotalRequestNum                uint64      `json:"total_request_num" bson:"total_request_num"`   // 总请求数
	TotalRequestTime               float64     `json:"total_request_time" bson:"total_request_time"` // 总响应时间
	SuccessNum                     uint64      `json:"success_num" bson:"success_num"`
	ErrorRate                      float64     `json:"error_rate" bson:"error_rate"`
	ErrorNum                       uint64      `json:"error_num" bson:"error_num"`               // 错误数
	AvgRequestTime                 float64     `json:"avg_request_time" bson:"avg_request_time"` // 平均响应事件
	MaxRequestTime                 float64     `json:"max_request_time" bson:"max_request_time"`
	MinRequestTime                 float64     `json:"min_request_time" bson:"min_request_time"`     // 毫秒
	PercentAge                     int64       `json:"percent_age" bson:"percent_age"`               // 响应时间线
	ErrorThreshold                 float64     `json:"error_threshold" bson:"error_threshold"`       // 自定义错误率
	RequestThreshold               int64       `json:"request_threshold" bson:"request_threshold"`   // Rps（每秒请求数）阈值
	ResponseThreshold              int64       `json:"response_threshold" bson:"response_threshold"` // 响应时间阈值
	CustomRequestTimeLine          int64       `json:"custom_request_time_line" bson:"custom_request_time_line"`
	FiftyRequestTimeline           int64       `json:"fifty_request_time_line" bson:"fifty_request_time_line"`
	NinetyRequestTimeLine          int64       `json:"ninety_request_time_line" bson:"ninety_request_time_line"`
	NinetyFiveRequestTimeLine      int64       `json:"ninety_five_request_time_line" bson:"ninety_five_request_time_line"`
	NinetyNineRequestTimeLine      int64       `json:"ninety_nine_request_time_line" bson:"ninety_nine_request_time_line"`
	CustomRequestTimeLineValue     float64     `json:"custom_request_time_line_value" bson:"custom_request_time_line_value"`
	FiftyRequestTimelineValue      float64     `json:"fifty_request_time_line_value" bson:"fifty_request_time_line_value"`
	NinetyRequestTimeLineValue     float64     `json:"ninety_request_time_line_value" bson:"ninety_request_time_line_value"`
	NinetyFiveRequestTimeLineValue float64     `json:"ninety_five_request_time_line_value" bson:"ninety_five_request_time_line_value"`
	NinetyNineRequestTimeLineValue float64     `json:"ninety_nine_request_time_line_value" bson:"ninety_nine_request_time_line_value"`
	SendBytes                      float64     `json:"send_bytes" bson:"send_bytes"`         // 发送字节数
	ReceivedBytes                  float64     `json:"received_bytes" bson:"received_bytes"` // 接收字节数
	Rps                            float64     `json:"rps" bson:"rps"`
	SRps                           float64     `json:"srps" bson:"srps"`
	Tps                            float64     `json:"tps" bson:"tps"`
	STps                           float64     `json:"stps" bson:"stps"`
	ConcurrencyList                []TimeValue `json:"concurrency_list" bson:"concurrency_list"`
	RpsList                        []TimeValue `json:"rps_list" bson:"rps_list"`
	TpsList                        []TimeValue `json:"tps_list" bson:"tps_list"`
	AvgList                        []TimeValue `json:"avg_list" bson:"avg_list"`
	FiftyList                      []TimeValue `json:"fifty_list" bson:"fifty_list"`
	NinetyList                     []TimeValue `json:"ninety_list" bson:"ninety_list"`
	NinetyFiveList                 []TimeValue `json:"ninety_five_list" bson:"ninety_five_list"`
	NinetyNineList                 []TimeValue `json:"ninety_nine_list" bson:"ninety_nine_list"`
}

type TimeValue struct {
	TimeStamp int64       `json:"time_stamp" bson:"time_stamp"`
	Value     interface{} `json:"value" bson:"value"`
}

type SceneTestResultDataMsg struct {
	End        bool                            `json:"end" bson:"end"`
	ReportId   string                          `json:"report_id" bson:"report_id"`
	ReportName string                          `json:"report_name" bson:"report_name"`
	PlanId     string                          `json:"plan_id" bson:"plan_id"`     // 任务ID
	PlanName   string                          `json:"plan_name" bson:"plan_name"` //
	SceneId    string                          `json:"scene_id" bson:"scene_id"`   // 场景
	SceneName  string                          `json:"scene_name" bson:"scene_name"`
	Results    map[string]ApiTestResultDataMsg `json:"results" bson:"results"`
	Machine    map[string]int64                `json:"machine" bson:"machine"`
	TimeStamp  int64                           `json:"time_stamp" bson:"time_stamp"`
}

// ApiTestResultDataMsg 接口测试数据经过计算后的测试结果
type ApiTestResultDataMsg struct {
	Name                           string  `json:"name" bson:"name"`
	Concurrency                    int64   `json:"concurrency" bson:"concurrency"`
	TotalRequestNum                uint64  `json:"total_request_num" bson:"total_request_num"`   // 总请求数
	TotalRequestTime               uint64  `json:"total_request_time" bson:"total_request_time"` // 总响应时间
	SuccessNum                     uint64  `json:"success_num" bson:"success_num"`
	ErrorNum                       uint64  `json:"error_num" bson:"error_num"`                   // 错误数
	ErrorThreshold                 float64 `json:"error_threshold" bson:"error_threshold"`       // 自定义错误率
	RequestThreshold               int64   `json:"request_threshold" bson:"request_threshold"`   // Rps（每秒请求数）阈值
	ResponseThreshold              int64   `json:"response_threshold" bson:"response_threshold"` // 响应时间阈值
	PercentAge                     int64   `json:"percent_age" bson:"percent_age"`               // 响应时间线
	AvgRequestTime                 float64 `json:"avg_request_time" bson:"avg_request_time"`     // 平均响应事件
	MaxRequestTime                 float64 `json:"max_request_time" bson:"max_request_time"`
	MinRequestTime                 float64 `json:"min_request_time" bson:"min_request_time"` // 毫秒
	CustomRequestTimeLine          int64   `json:"custom_request_time_line" bson:"custom_request_time_line"`
	FiftyRequestTimeline           int64   `json:"fifty_request_time_line" bson:"fifty_request_time_line"`
	NinetyRequestTimeLine          int64   `json:"ninety_request_time_line" bson:"ninety_request_time_line"`
	NinetyFiveRequestTimeLine      int64   `json:"ninety_five_request_time_line" bson:"ninety_five_request_time_line"`
	NinetyNineRequestTimeLine      int64   `json:"ninety_nine_request_time_line" bson:"ninety_nine_request_time_line"`
	FiftyRequestTimelineValue      float64 `json:"fifty_request_time_line_value"`
	CustomRequestTimeLineValue     float64 `json:"custom_request_time_line_value" bson:"custom_request_time_line_value"`
	NinetyRequestTimeLineValue     float64 `json:"ninety_request_time_line_value" bson:"ninety_request_time_line_value"`
	NinetyFiveRequestTimeLineValue float64 `json:"ninety_five_request_time_line_value" bson:"ninety_five_request_time_line_value"`
	NinetyNineRequestTimeLineValue float64 `json:"ninety_nine_request_time_line_value" bson:"ninety_nine_request_time_line_value"`
	SendBytes                      float64 `json:"send_bytes" bson:"send_bytes"`         // 发送字节数
	ReceivedBytes                  float64 `json:"received_bytes" bson:"received_bytes"` // 接收字节数
	Rps                            float64 `json:"rps" bson:"rps"`
	SRps                           float64 `json:"srps" bson:"srps"`
	Tps                            float64 `json:"tps" bson:"tps"`
	STps                           float64 `json:"stps" bson:"stps"`
	ApiName                        string  `json:"api_name" bson:"api_name"`
	ErrorRate                      float64 `json:"error_rate" bson:"error_rate"`
}
