package autoPlan

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/errno"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/log"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/mail"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/record"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/uuid"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/conf"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/mao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/run_plan"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/packer"
	"github.com/go-omnibus/proof"
	"github.com/go-resty/resty/v2"
	"github.com/shirou/gopsutil/load"
	"go.mongodb.org/mongo-driver/bson"
	"sort"
	"strings"
	"sync"
	"time"
)

type Baton struct {
	Ctx             context.Context
	PlanID          string
	TeamID          string
	UserID          string
	SceneIDs        []string
	RunType         int
	reportID        string
	plan            *model.AutoPlan
	scenes          []*model.Target
	testCase        []*model.Target
	testCaseIDs     []string
	ConfigTask      ConfigTask // 任务配置
	globalVariables run_plan.GlobalVariable
	sceneFlows      map[string]*mao.Flow
	sceneCaseFlows  map[string]*mao.SceneCaseFlow
	sceneVariables  map[string]run_plan.GlobalVariable
	importVariables map[string][]*model.VariableImport
	reports         []*model.Report
	balance         *DispatchMachineBalance
	stress          []*run_plan.Stress
	MachineList     []*HeartBeat
	RealRunParam    RealRunParam
}

type RealRunParam struct {
	PlanId         string                  `json:"plan_id" bson:"plan_id"`             // 计划id
	PlanName       string                  `json:"plan_name" bson:"plan_name"`         // 计划名称
	ReportId       string                  `json:"report_id" bson:"report_id"`         // 报告名称
	TeamId         string                  `json:"team_id" bson:"team_id"`             // 团队id
	ReportName     string                  `json:"report_name" bson:"report_name"`     // 报告名称
	MachineNum     int64                   `json:"machine_num" bson:"machine_num"`     // 使用的机器数量
	ConfigTask     ConfigTask              `json:"config_task" bson:"config_task"`     // 任务配置
	Variable       []PlanKv                `json:"variable" bson:"variable"`           // 全局变量
	Scenes         []Scene                 `json:"scenes" bson:"scenes"`               // 场景
	Configuration  Configuration           `json:"configuration" bson:"configuration"` // 场景配置
	GlobalVariable run_plan.GlobalVariable `json:"global_variable" bson:"global_variable"`
}

// ConfigTask 任务配置
type ConfigTask struct {
	TaskType     int64  `json:"task_type" bson:"task_type"`           // 任务类型：0. 普通任务； 1. 定时任务； 2. cicd任务
	TaskMode     int64  `json:"task_mode" bson:"task_mode"`           // 1. 按用例执行
	SceneRunMode int64  `json:"scene_run_mode" bson:"scene_run_mode"` // 2. 同时执行； 1. 顺序执行
	CaseRunMode  int64  `json:"case_run_mode" bson:"case_run_mode"`   // 2. 同时执行； 1. 顺序执行
	Remark       string `json:"remark" bson:"remark"`                 // 备注
}

type PlanKv struct {
	Var string `json:"Var"`
	Val string `json:"Val"`
}

type Scene struct {
	PlanId                  string                  `json:"plan_id" bson:"plan_id"`
	SceneId                 string                  `json:"scene_id" bson:"scene_id"`     // 场景Id
	IsChecked               int32                   `json:"is_checked" bson:"is_checked"` // 是否启用
	ParentId                string                  `json:"parentId" bson:"parent_id"`
	CaseId                  string                  `json:"case_id" bson:"case_id"`
	Partition               int32                   `json:"partition"`
	MachineNum              int64                   `json:"machine_num" bson:"machine_num"` // 使用的机器数量
	ReportId                string                  `json:"report_id" bson:"report_id"`
	TeamId                  string                  `json:"team_id" bson:"team_id"`
	SceneName               string                  `json:"scene_name" bson:"scene_name"` // 场景名称
	Version                 int64                   `json:"version" bson:"version"`
	Debug                   string                  `json:"debug" bson:"debug"`
	EnablePlanConfiguration bool                    `json:"enable_plan_configuration" bson:"enable_plan_configuration"` // 是否启用计划的任务配置，默认为true，
	ConfigTask              ConfigTask              `json:"config_task" bson:"config_task"`                             // 任务配置
	Configuration           Configuration           `json:"configuration" bson:"configuration"`                         // 场景配置
	Variable                []KV                    `json:"variable" bson:"variable"`                                   // 场景配置
	Cases                   []Scene                 `json:"cases" bson:"cases"`
	NodesRound              [][]rao.Node            `json:"nodes_round" bson:"nodes_round"`
	GlobalVariable          run_plan.GlobalVariable `json:"global_variable"` // 全局变量
	Prepositions            []rao.Preposition       `json:"prepositions"`    // 前置条件
}

type Configuration struct {
	ParameterizedFile ParameterizedFile       `json:"parameterizedFile" bson:"parameterizedFile"`
	SceneVariable     run_plan.GlobalVariable `json:"scene_variable"`
	//Variable          []KV              `json:"variable" bson:"variable"` // todo 弃用
}

// ParameterizedFile 参数化文件
type ParameterizedFile struct {
	Paths         []FileList     `json:"paths"` // 文件地址
	RealPaths     []string       `json:"real_paths"`
	VariableNames *VariableNames `json:"variable_names"` // 存储变量及数据的map
}

type FileList struct {
	IsChecked int64  `json:"is_checked"` // 1 开， 2： 关
	Path      string `json:"path"`
}

type VariableNames struct {
	VarMapList map[string][]string `json:"var_map_list"`
	Index      int                 `json:"index"`
	Mu         sync.Mutex          `json:"mu"`
}

type KV struct {
	Key   string      `json:"key" bson:"key"`
	Value interface{} `json:"value" bson:"value"`
}

type Event struct {
	Id                string   `json:"id" bson:"id"`
	ReportId          int64    `json:"report_id" bson:"report_id"`
	TeamId            int64    `json:"team_id" bson:"team_id"`
	IsCheck           bool     `json:"is_check" bson:"is_check"`
	Type              string   `json:"type" bson:"type"` //   事件类型 "request" "controller"
	PreList           []string `json:"pre_list" bson:"pre_list"`
	NextList          []string `json:"next_list"   bson:"next_list"`
	Tag               bool     `json:"tag" bson:"tag"` // Tps模式下，该标签代表以该接口为准
	Debug             string   `json:"debug" bson:"debug"`
	Mode              int64    `json:"mode"`                 // 模式类型
	RequestThreshold  int64    `json:"request_threshold"`    // Rps（每秒请求数）阈值
	ResponseThreshold int64    `json:"response_threshold"`   // 响应时间阈值
	ErrorThreshold    float32  `json:"error_threshold"`      // 错误率阈值
	PercentAge        int64    `json:"percent_age"`          // 响应时间线
	Weight            int64    `json:"weight" bson:"weight"` // 权重，并发分配的比例
	Api               Api      `json:"api"`
	Var               string   `json:"var"`     // if控制器key，值某个变量
	Compare           string   `json:"compare"` // 逻辑运算符
	Val               string   `json:"val"`     // key对应的值
	Name              string   `json:"name"`    // 控制器名称
	WaitTime          int      `json:"wait_ms"` // 等待时长，ms
}

// Api 请求数据
type Api struct {
	TargetId   int64   `json:"target_id" bson:"target_id"`
	Name       string  `json:"name" bson:"name"`
	TeamId     int64   `json:"team_id" bson:"team_id"`
	TargetType string  `json:"target_type" bson:"target_type"` // api/webSocket/tcp/grpc
	Method     string  `json:"method" bson:"method"`           // 方法 GET/POST/PUT
	Request    Request `json:"request" bson:"request"`
	//Parameters    *sync.Map            `json:"parameters" bson:"parameters"`
	Assert        []*AssertionText     `json:"assert" bson:"assert"`         // 验证的方法(断言)
	Timeout       int64                `json:"timeout" bson:"timeout"`       // 请求超时时间
	Regex         []*RegularExpression `json:"regex" bson:"regex"`           // 正则表达式
	Debug         string               `json:"debug" bson:"debug"`           // 是否开启Debug模式
	Connection    int64                `json:"connection" bson:"connection"` // 0:websocket长连接
	Configuration *Configuration       `json:"configuration" bson:"configuration"`
	Variable      []*KV                `json:"variable" bson:"variable"` // 全局变量
	HttpApiSetup  HttpApiSetup         `json:"http_api_setup" bson:"http_api_setup"`
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

type Request struct {
	PreUrl    string     `json:"pre_url" bson:"pre_url"`
	URL       string     `json:"url" bson:"url"`
	Parameter []*VarForm `json:"parameter" bson:"parameter"`
	Header    *Header    `json:"header" bson:"header"` // Headers
	Query     *Query     `json:"query" bson:"query"`
	Body      *Body      `json:"body" bson:"body"`
	Auth      *Auth      `json:"auth" bson:"auth"`
	Cookie    *Cookie    `json:"cookie" bson:"cookie"`
}
type Header struct {
	Parameter []*VarForm `json:"parameter" bson:"parameter"`
}
type Query struct {
	Parameter []*VarForm `json:"parameter" bson:"parameter"`
}

type Cookie struct {
	Parameter []*VarForm
}

type Auth struct {
	Type          string    `json:"type" bson:"type"`
	KV            *KV       `json:"kv" bson:"kv"`
	Bearer        *Bearer   `json:"bearer" bson:"bearer"`
	Basic         *Basic    `json:"basic" bson:"basic"`
	Digest        *Digest   `json:"digest"`
	Hawk          *Hawk     `json:"hawk"`
	Awsv4         *AwsV4    `json:"awsv4"`
	Ntlm          *Ntlm     `json:"ntlm"`
	Edgegrid      *Edgegrid `json:"edgegrid"`
	Oauth1        *Oauth1   `json:"oauth1"`
	Bidirectional TLS       `json:"bidirectional"`
}

type TLS struct {
	CaCert     string `json:"ca_cert"`
	CaCertName string `json:"ca_cert_name"`
}

type Bearer struct {
	Key string `json:"key" bson:"key"`
}

type Basic struct {
	UserName string `json:"username" bson:"username"`
	Password string `json:"password" bson:"password"`
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
	Mode      string     `json:"mode" bson:"mode"`
	Raw       string     `json:"raw" bson:"raw"`
	Parameter []*VarForm `json:"parameter" bson:"parameter"`
}

type RegularExpression struct {
	IsChecked int         `json:"is_checked"` // 1 选中, -1未选
	Type      int         `json:"type"`       // 0 正则  1 json
	Var       string      `json:"var"`        // 变量
	Express   string      `json:"express"`    // 表达式
	Val       interface{} `json:"val"`        // 值
}

// AssertionText 文本断言 0
type AssertionText struct {
	IsChecked    int    `json:"is_checked"`    // 1 选中  -1 未选
	ResponseType int8   `json:"response_type"` //  1:ResponseHeaders; 2:ResponseData; 3: ResponseCode;
	Compare      string `json:"compare"`       // Includes、UNIncludes、Equal、UNEqual、GreaterThan、GreaterThanOrEqual、LessThan、LessThanOrEqual、Includes、UNIncludes、NULL、NotNULL、OriginatingFrom、EndIn
	Var          string `json:"var"`
	Val          string `json:"val"`
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

//////////
type UsableMachineMap struct {
	IP               string // IP地址(包含端口号)
	Region           string // 机器所属区域
	Weight           int64  // 权重
	UsableGoroutines int64  // 可用协程数
}

// 压力机心跳上报数据
type HeartBeat struct {
	Name              string        `json:"name"`               // 机器名称
	CpuUsage          float64       `json:"cpu_usage"`          // CPU使用率
	CpuLoad           *load.AvgStat `json:"cpu_load"`           // CPU负载信息
	MemInfo           []MemInfo     `json:"mem_info"`           // 内存使用情况
	Networks          []Network     `json:"networks"`           // 网络连接情况
	DiskInfos         []DiskInfo    `json:"disk_infos"`         // 磁盘IO情况
	MaxGoroutines     int64         `json:"max_goroutines"`     // 当前机器支持最大协程数
	CurrentGoroutines int64         `json:"current_goroutines"` // 当前已用协程数
	ServerType        int64         `json:"server_type"`        // 压力机类型：0-主力机器，1-备用机器
	CreateTime        int64         `json:"create_time"`        // 数据上报时间（时间戳）
}

type MemInfo struct {
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"usedPercent"`
}

type DiskInfo struct {
	Total       uint64  `json:"total"`
	Free        uint64  `json:"free"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"usedPercent"`
}

type Network struct {
	Name        string `json:"name"`
	BytesSent   uint64 `json:"bytesSent"`
	BytesRecv   uint64 `json:"bytesRecv"`
	PacketsSent uint64 `json:"packetsSent"`
	PacketsRecv uint64 `json:"packetsRecv"`
}

// CheckAutoPlanTaskType 检查计划类型，如果是定时任务，直接修改状态，返回成功
func CheckAutoPlanTaskType(baton *Baton) (int, error) {
	if baton.RunType != 2 {
		tx := dal.GetQuery().AutoPlan
		planInfo, err := tx.WithContext(baton.Ctx).Where(tx.TeamID.Eq(baton.TeamID),
			tx.PlanID.Eq(baton.PlanID)).First()
		if err != nil {
			return errno.ErrMysqlFailed, fmt.Errorf("计划没查到")
		}
		if planInfo.TaskType == 2 {
			timedTaskConfTable := dal.GetQuery().AutoPlanTimedTaskConf
			ttcInfo, err := timedTaskConfTable.WithContext(baton.Ctx).Where(timedTaskConfTable.PlanID.Eq(baton.PlanID)).First()
			if err != nil {
				return errno.ErrMustTaskInit, err
			}

			// 检查定时任务时间是否过期
			nowTime := time.Now().Unix()
			var taskCloseTime int64 = 0
			if ttcInfo.Frequency == 0 {
				taskCloseTime = ttcInfo.TaskExecTime
			} else {
				taskCloseTime = ttcInfo.TaskCloseTime
			}

			if taskCloseTime <= nowTime {
				return errno.ErrTimedTaskOverdue, fmt.Errorf("开始或结束时间不能早于当前时间")
			}

			// 修改定时任务状态
			autoPlanTimedTaskConfTable := dal.GetQuery().AutoPlanTimedTaskConf
			_, err = autoPlanTimedTaskConfTable.WithContext(baton.Ctx).Where(autoPlanTimedTaskConfTable.TeamID.Eq(baton.TeamID),
				autoPlanTimedTaskConfTable.PlanID.Eq(baton.PlanID)).UpdateSimple(autoPlanTimedTaskConfTable.Status.Value(consts.TimedTaskInExec), autoPlanTimedTaskConfTable.RunUserID.Value(baton.UserID))
			if err != nil {
				return errno.ErrMysqlFailed, fmt.Errorf("定时任务状态修改失败")
			}

			// 修改计划的状态
			ap := dal.GetQuery().AutoPlan
			_, err = ap.WithContext(baton.Ctx).Where(ap.TeamID.Eq(baton.TeamID), ap.PlanID.Eq(baton.PlanID)).UpdateSimple(ap.Status.Value(consts.PlanStatusUnderway))
			if err != nil {
				return errno.ErrMysqlFailed, fmt.Errorf("计划状态修改失败")
			}
			return errno.Ok, fmt.Errorf("定时任务已经开启")
		}
	}
	return errno.Ok, nil
}

// CheckIdleMachine 1、检查运行机器情况
func CheckIdleMachine(baton *Baton) (int, error) {
	// 从Redis获取压力机列表
	machineListRes := dal.GetRDB().HGetAll(baton.Ctx, consts.MachineListRedisKey)
	if len(machineListRes.Val()) == 0 || machineListRes.Err() != nil {
		// todo 后面可能增加兜底策略
		log.Logger.Info("没有上报上来的空闲压力机可用")
		return errno.ErrResourceNotEnough, fmt.Errorf("资源不足")
	}

	baton.balance = &DispatchMachineBalance{}

	usableMachineMap := UsableMachineMap{}                                       // 单个压力机基本数据
	usableMachineSlice := make([]UsableMachineMap, 0, len(machineListRes.Val())) // 所有上报过来的压力机切片
	var minWeight int64                                                          // 所有可用压力机里面最小的权重的值
	var inUseMachineNum int                                                      // 所有有任务在运行的压力机数量

	var breakFor = false

	tx := dal.GetQuery().Machine
	// 查到了机器列表，然后判断可用性
	var runnerMachineInfo HeartBeat
	for machineAddr, machineDetail := range machineListRes.Val() {
		// 把机器详情信息解析成格式化数据
		err := json.Unmarshal([]byte(machineDetail), &runnerMachineInfo)
		if err != nil {
			log.Logger.Info("runner_machine_detail 数据解析失败 err：", err)
			continue
		}

		baton.MachineList = append(baton.MachineList, &runnerMachineInfo)

		// 压力机数据上报时间超过10秒，则认为服务不可用，不参与本次压力测试
		nowTime := time.Now().Unix()
		if nowTime-runnerMachineInfo.CreateTime > int64(conf.Conf.MachineConfig.MachineAliveTime) {
			log.Logger.Info("当前压力机上报心跳数据超时，暂不可用，机器信息：", machineAddr)
			continue
		}

		// 判断当前压力机性能是否爆满,如果某个指标爆满，则不参与本次压力测试
		if runnerMachineInfo.CpuUsage >= float64(conf.Conf.MachineConfig.CpuTopLimit) { // CPU使用判断
			log.Logger.Info("CPU超过使用阈值，阈值为：", conf.Conf.MachineConfig.CpuTopLimit, "当前cpu使用率为：", runnerMachineInfo.CpuUsage, "机器信息为：", machineAddr)
			continue
		}
		for _, memInfo := range runnerMachineInfo.MemInfo { // 内存使用判断
			if memInfo.UsedPercent >= float64(conf.Conf.MachineConfig.MemoryTopLimit) {
				log.Logger.Info("内存超过使用阈值，阈值为：", conf.Conf.MachineConfig.MemoryTopLimit, "当前内存使用率为：", memInfo.UsedPercent, "机器信息为：", machineAddr)
				breakFor = true
				break
			}
		}
		for _, diskInfo := range runnerMachineInfo.DiskInfos { // 磁盘使用判断
			if diskInfo.UsedPercent >= float64(conf.Conf.MachineConfig.DiskTopLimit) {
				log.Logger.Info("磁盘超过使用阈值，阈值为：", conf.Conf.MachineConfig.DiskTopLimit, "当前磁盘使用率为：", diskInfo.UsedPercent, "机器信息为：", machineAddr)
				breakFor = true
				break
			}
		}

		// 最后判断是否结束当前循环
		if breakFor {
			continue
		}

		machineAddrSlice := strings.Split(machineAddr, "_")
		if len(machineAddrSlice) != 3 {
			continue
		}

		// 判断当前压力机是否被停用，如果停用，则不参与压测
		machineInfo, err := tx.WithContext(baton.Ctx).Where(tx.IP.Eq(machineAddrSlice[0])).First()
		if err != nil {
			log.Logger.Info("运行计划--没有查到当前压力机数据，err:", err)
			continue
		}
		if machineInfo.Status == 2 { // 已停用
			continue
		}

		// 当前机器可用协程数
		usableGoroutines := runnerMachineInfo.MaxGoroutines - runnerMachineInfo.CurrentGoroutines

		// 组装可用机器结构化数据
		usableMachineMap.IP = machineAddrSlice[0] + ":" + machineAddrSlice[1]
		usableMachineMap.UsableGoroutines = usableGoroutines
		usableMachineMap.Weight = usableGoroutines
		usableMachineSlice = append(usableMachineSlice, usableMachineMap)

		// 获取当前压力机当中最小的权重值
		if minWeight == 0 || minWeight > usableGoroutines {
			minWeight = usableGoroutines
		}

		// 获取当前机器是否使用当中
		machineUseStateKey := consts.MachineUseStatePrefix + machineAddrSlice[0] + ":" + machineAddrSlice[1]
		useStateVal, _ := dal.GetRDB().Get(baton.Ctx, machineUseStateKey).Result()
		if useStateVal != "" {
			inUseMachineNum++
		}

	}

	for _, machineInfo := range usableMachineSlice {
		if inUseMachineNum < len(usableMachineSlice) {
			// 获取当前机器是否使用当中
			machineUseStateKey := consts.MachineUseStatePrefix + machineInfo.IP
			useStateVal, _ := dal.GetRDB().Get(baton.Ctx, machineUseStateKey).Result()
			if useStateVal != "" {
				machineInfo.UsableGoroutines = int64(minWeight) - 1
				if machineInfo.UsableGoroutines <= 0 {
					machineInfo.UsableGoroutines = 1
				}
			}
		}
	}

	sort.Slice(usableMachineSlice, func(i, j int) bool {
		return usableMachineSlice[i].UsableGoroutines > usableMachineSlice[j].UsableGoroutines
	})

	// 按当前顺序把机器放到备用列表
	for _, machineInfo := range usableMachineSlice {
		addErr := baton.balance.AddMachine(fmt.Sprintf("%s", machineInfo.IP))
		if addErr != nil {
			continue
		}
	}

	if len(baton.balance.rss) == 0 {
		log.Logger.Info("当前没有空闲压力机可用")
		return errno.ErrResourceNotEnough, fmt.Errorf("资源不足")
	}

	return errno.Ok, nil
}

// AssemblePlan 2、组装计划
func AssemblePlan(baton *Baton) (int, error) {
	tx := dal.GetQuery().AutoPlan
	autoPlanInfo, err := tx.WithContext(baton.Ctx).Where(tx.TeamID.Eq(baton.TeamID), tx.PlanID.Eq(baton.PlanID)).First()
	if err != nil {
		return errno.ErrMysqlFailed, fmt.Errorf("组装计划失败")
	}
	baton.plan = autoPlanInfo
	return errno.Ok, err
}

// AssembleScenes 3、组装计划下场景
func AssembleScenes(baton *Baton) (int, error) {
	tx := dal.GetQuery().Target
	targetList, err := tx.WithContext(baton.Ctx).Where(tx.TeamID.Eq(baton.TeamID),
		tx.PlanID.Eq(baton.PlanID),
		tx.TargetType.Eq(consts.TargetTypeScene),
		tx.Source.Eq(consts.TargetSourceAutoPlan),
		tx.IsDisabled.Eq(consts.TargetIsDisabledNo)).Find()
	if err != nil {
		return errno.ErrMysqlFailed, fmt.Errorf("组装场景失败")
	}
	sceneIDs := make([]string, 0, len(baton.scenes))
	for _, sceneInfo := range targetList {
		sceneIDs = append(sceneIDs, sceneInfo.TargetID)
	}
	baton.scenes = targetList
	baton.SceneIDs = sceneIDs
	return errno.Ok, err
}

// AssembleTask 4、组装计划的配置
func AssembleTask(baton *Baton) (int, error) {
	// 查询计划类型
	tx := dal.GetQuery()
	if baton.plan.TaskType == consts.PlanTaskTypeNormal || baton.plan.TaskType == 0 { // 普通任务
		autoPlanTaskConf, err := tx.AutoPlanTaskConf.WithContext(baton.Ctx).Where(tx.AutoPlanTaskConf.TeamID.Eq(baton.TeamID), tx.AutoPlanTaskConf.PlanID.Eq(baton.PlanID)).First()
		if err != nil {
			return errno.ErrMustTaskInit, fmt.Errorf("没有查到计划配置")
		}
		baton.ConfigTask.TaskType = int64(autoPlanTaskConf.TaskType)
		baton.ConfigTask.TaskMode = int64(autoPlanTaskConf.TaskMode)
		baton.ConfigTask.SceneRunMode = int64(autoPlanTaskConf.SceneRunOrder)
		baton.ConfigTask.CaseRunMode = int64(autoPlanTaskConf.TestCaseRunOrder)
		baton.ConfigTask.Remark = baton.plan.Remark
	} else { // 定时任务
		AutoPlanTimedTaskConf, err := tx.AutoPlanTimedTaskConf.WithContext(baton.Ctx).Where(tx.AutoPlanTimedTaskConf.TeamID.Eq(baton.TeamID), tx.AutoPlanTimedTaskConf.PlanID.Eq(baton.PlanID)).First()
		if err != nil {
			return errno.ErrMustTaskInit, fmt.Errorf("没有查到计划配置")
		}
		baton.ConfigTask.TaskType = int64(AutoPlanTimedTaskConf.TaskType)
		baton.ConfigTask.TaskMode = int64(AutoPlanTimedTaskConf.TaskMode)
		baton.ConfigTask.SceneRunMode = int64(AutoPlanTimedTaskConf.SceneRunOrder)
		baton.ConfigTask.CaseRunMode = int64(AutoPlanTimedTaskConf.TestCaseRunOrder)
		baton.ConfigTask.Remark = baton.plan.Remark
	}
	return errno.Ok, nil
}

// AssembleTestCase 组装测试用例
func AssembleTestCase(baton *Baton) (int, error) {
	// 查询所有场景下的所有用例
	tx := dal.GetQuery().Target
	testCaseList, err := tx.WithContext(baton.Ctx).Where(
		tx.ParentID.In(baton.SceneIDs...),
		tx.TargetType.Eq(consts.TargetTypeTestCase),
		tx.IsChecked.Eq(consts.TargetIsCheckedOpen)).Find()
	if err != nil {
		return errno.ErrMysqlFailed, fmt.Errorf("查询场景用例出错")
	}

	if len(testCaseList) == 0 {
		return errno.ErrEmptyTestCase, fmt.Errorf("场景用例不能为空")
	}

	// 收集测试用例的id集合
	testCaseIDs := make([]string, 0, len(testCaseList))
	for _, v := range testCaseList {
		testCaseIDs = append(testCaseIDs, v.TargetID)
	}
	baton.testCaseIDs = testCaseIDs
	baton.testCase = testCaseList
	return errno.Ok, nil
}

// AssembleSceneFlows 组装所有场景flow
func AssembleSceneFlows(baton *Baton) (int, error) {
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectFlow)
	cur, err := collection.Find(baton.Ctx, bson.D{{"scene_id", bson.D{{"$in", baton.SceneIDs}}}})
	if err != nil {
		return errno.ErrMongoFailed, fmt.Errorf("场景flow查询失败")
	}

	flows := make([]*mao.Flow, 0, 1)
	if err := cur.All(baton.Ctx, &flows); err != nil {
		return errno.ErrMongoFailed, fmt.Errorf("场景flow获取失败")
	}
	if len(flows) != len(baton.SceneIDs) {
		log.Logger.Info("场景flow不能为空")
		return errno.ErrEmptySceneFlow, fmt.Errorf("场景flow不能为空")
	}

	sceneFlowsMap := make(map[string]*mao.Flow)
	for _, flowInfo := range flows {
		sceneFlowsMap[flowInfo.SceneID] = flowInfo
	}
	baton.sceneFlows = sceneFlowsMap
	return errno.Ok, nil
}

// AssembleTestCaseFlows 组装所有测试用例的flow
func AssembleTestCaseFlows(baton *Baton) (int, error) {
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectSceneCaseFlow)
	cur, err := collection.Find(baton.Ctx, bson.D{{"scene_case_id", bson.D{{"$in", baton.testCaseIDs}}}})
	if err != nil {
		return errno.ErrMongoFailed, fmt.Errorf("测试用例flow查询失败")
	}
	sceneCaseFlows := make([]*mao.SceneCaseFlow, 0, 1)
	if err := cur.All(baton.Ctx, &sceneCaseFlows); err != nil {
		return errno.ErrMongoFailed, fmt.Errorf("测试用例flow获取失败")
	}

	// 判断用例flow是否为空
	if len(sceneCaseFlows) > 0 {
		for _, caseFlow := range sceneCaseFlows {
			sceneCaseFlowNodeTemp := mao.SceneCaseFlowNode{}
			err := bson.Unmarshal(caseFlow.Nodes, &sceneCaseFlowNodeTemp)
			if err != nil {
				return errno.ErrMongoFailed, fmt.Errorf("测试用例flow解析失败")
			}
			if len(sceneCaseFlowNodeTemp.Nodes) == 0 {
				return errno.ErrEmptyTestCaseFlow, fmt.Errorf("测试用例flow为空")
			}
		}
	}

	if len(sceneCaseFlows) != len(baton.testCaseIDs) {
		log.Logger.Info("测试用例flow不能为空")
		return errno.ErrEmptyTestCaseFlow, fmt.Errorf("测试用例flow不能为空")
	}

	sceneCaseFlowsMap := make(map[string]*mao.SceneCaseFlow)
	for _, sceneCaseFlowInfo := range sceneCaseFlows {
		sceneCaseFlowsMap[sceneCaseFlowInfo.SceneCaseID] = sceneCaseFlowInfo
	}
	baton.sceneCaseFlows = sceneCaseFlowsMap
	return errno.Ok, nil
}

// AssembleGlobalVariables 组装全局变量数据
func AssembleGlobalVariables(baton *Baton) (int, error) {
	//tx := dal.GetQuery().Variable
	//variables, err := tx.WithContext(baton.Ctx).Where(
	//	tx.TeamID.Eq(baton.TeamID),
	//	tx.Type.Eq(consts.VariableTypeGlobal),
	//	tx.Status.Eq(consts.VariableStatusOpen),
	//).Find()
	//
	//if err == nil {
	//	baton.globalVariables = variables
	//}

	globalVariable := run_plan.GlobalVariable{}
	// 查询全局变量
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectGlobalParam)
	cur, err := collection.Find(baton.Ctx, bson.D{{"team_id", baton.TeamID}})
	var globalParamDataArr []*mao.GlobalParamData
	if err == nil {
		if err := cur.All(baton.Ctx, &globalParamDataArr); err != nil {
			return errno.ErrMongoFailed, fmt.Errorf("全局参数数据获取失败")
		}
	}

	cookieParam := make([]rao.CookieParam, 0, 100)
	headerParam := make([]rao.HeaderParam, 0, 100)
	variableParam := make([]rao.VariableParam, 0, 100)
	assertParam := make([]rao.AssertParam, 0, 100)
	for _, globalParamInfo := range globalParamDataArr {
		if globalParamInfo.ParamType == 1 {
			err = json.Unmarshal([]byte(globalParamInfo.DataDetail), &cookieParam)
			if err != nil {
				return errno.ErrUnMarshalFailed, err
			}
			parameter := make([]*run_plan.Parameter, 0, len(cookieParam))
			for _, v := range cookieParam {
				temp := &run_plan.Parameter{
					IsChecked: v.IsChecked,
					Key:       v.Key,
					Value:     v.Value,
				}
				parameter = append(parameter, temp)
			}
			globalVariable.Cookie.Parameter = parameter
		}
		if globalParamInfo.ParamType == 2 {
			err = json.Unmarshal([]byte(globalParamInfo.DataDetail), &headerParam)
			if err != nil {
				return errno.ErrUnMarshalFailed, err
			}

			parameter := make([]*run_plan.Parameter, 0, len(headerParam))
			for _, v := range headerParam {
				temp := &run_plan.Parameter{
					IsChecked: v.IsChecked,
					Key:       v.Key,
					Value:     v.Value,
				}
				parameter = append(parameter, temp)
			}
			globalVariable.Header.Parameter = parameter

		}
		if globalParamInfo.ParamType == 3 {
			err = json.Unmarshal([]byte(globalParamInfo.DataDetail), &variableParam)
			if err != nil {
				return errno.ErrUnMarshalFailed, err
			}

			parameter := make([]run_plan.VarForm, 0, len(variableParam))
			for _, v := range variableParam {
				temp := run_plan.VarForm{
					IsChecked: int64(v.IsChecked),
					Key:       v.Key,
					Value:     v.Value,
				}
				parameter = append(parameter, temp)
			}
			globalVariable.Variable = parameter

		}
		if globalParamInfo.ParamType == 4 {
			err = json.Unmarshal([]byte(globalParamInfo.DataDetail), &assertParam)
			if err != nil {
				return errno.ErrUnMarshalFailed, err
			}

			parameter := make([]run_plan.AssertionText, 0, len(assertParam))
			for _, v := range assertParam {
				temp := run_plan.AssertionText{
					IsChecked:    int(v.IsChecked),
					ResponseType: int8(v.ResponseType),
					Compare:      v.Compare,
					Var:          v.Var,
					Val:          v.Val,
				}
				parameter = append(parameter, temp)
			}
			globalVariable.Assert = parameter
		}
	}
	baton.globalVariables = globalVariable
	return errno.Ok, nil
}

// AssembleVariable 组装场景变量数据
func AssembleVariable(baton *Baton) (int, error) {
	//查询当前计划下所有场景的所有变量
	//tx := dal.GetQuery().Variable
	//variables, err := tx.WithContext(baton.Ctx).Where(
	//	tx.TeamID.Eq(baton.TeamID),
	//	tx.SceneID.In(baton.SceneIDs...),
	//	tx.Type.Eq(consts.VariableTypeScene),
	//	tx.Status.Eq(consts.VariableStatusOpen),
	//).Find()
	//
	//tempData := make(map[string][]*model.Variable)
	//if err == nil {
	//	for _, variablesInfo := range variables {
	//		tempData[variablesInfo.SceneID] = append(tempData[variablesInfo.SceneID], variablesInfo)
	//	}
	//}
	//baton.sceneVariables = tempData

	sceneVariableMap := make(map[string]run_plan.GlobalVariable, len(baton.scenes))
	// 查询全局变量
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectSceneParam)
	for _, sceneInfo := range baton.scenes {
		sceneVariable := run_plan.GlobalVariable{}
		cur, err := collection.Find(baton.Ctx, bson.D{{"team_id", baton.TeamID}, {"scene_id", sceneInfo.TargetID}})
		var sceneParamDataArr []*mao.SceneParamData
		if err == nil {
			if err := cur.All(baton.Ctx, &sceneParamDataArr); err != nil {
				return errno.ErrMongoFailed, fmt.Errorf("场景参数数据获取失败")
			}
		}

		cookieParam := make([]rao.CookieParam, 0, 100)
		headerParam := make([]rao.HeaderParam, 0, 100)
		variableParam := make([]rao.VariableParam, 0, 100)
		assertParam := make([]rao.AssertParam, 0, 100)
		for _, sceneParamInfo := range sceneParamDataArr {
			if sceneParamInfo.ParamType == 1 {
				err = json.Unmarshal([]byte(sceneParamInfo.DataDetail), &cookieParam)
				if err != nil {
					return errno.ErrUnMarshalFailed, err
				}
				parameter := make([]*run_plan.Parameter, 0, len(cookieParam))
				for _, v := range cookieParam {
					temp := &run_plan.Parameter{
						IsChecked: v.IsChecked,
						Key:       v.Key,
						Value:     v.Value,
					}
					parameter = append(parameter, temp)
				}
				sceneVariable.Cookie.Parameter = parameter
			}
			if sceneParamInfo.ParamType == 2 {
				err = json.Unmarshal([]byte(sceneParamInfo.DataDetail), &headerParam)
				if err != nil {
					return errno.ErrUnMarshalFailed, err
				}

				parameter := make([]*run_plan.Parameter, 0, len(headerParam))
				for _, v := range headerParam {
					temp := &run_plan.Parameter{
						IsChecked: v.IsChecked,
						Key:       v.Key,
						Value:     v.Value,
					}
					parameter = append(parameter, temp)
				}
				sceneVariable.Header.Parameter = parameter

			}
			if sceneParamInfo.ParamType == 3 {
				err = json.Unmarshal([]byte(sceneParamInfo.DataDetail), &variableParam)
				if err != nil {
					return errno.ErrUnMarshalFailed, err
				}

				parameter := make([]run_plan.VarForm, 0, len(variableParam))
				for _, v := range variableParam {
					temp := run_plan.VarForm{
						IsChecked: int64(v.IsChecked),
						Key:       v.Key,
						Value:     v.Value,
					}
					parameter = append(parameter, temp)
				}
				sceneVariable.Variable = parameter
			}
			if sceneParamInfo.ParamType == 4 {
				err = json.Unmarshal([]byte(sceneParamInfo.DataDetail), &assertParam)
				if err != nil {
					return errno.ErrUnMarshalFailed, err
				}

				parameter := make([]run_plan.AssertionText, 0, len(assertParam))
				for _, v := range assertParam {
					temp := run_plan.AssertionText{
						IsChecked:    int(v.IsChecked),
						ResponseType: int8(v.ResponseType),
						Compare:      v.Compare,
						Var:          v.Var,
						Val:          v.Val,
					}
					parameter = append(parameter, temp)
				}
				sceneVariable.Assert = parameter
			}
		}
		sceneVariableMap[sceneInfo.TargetID] = sceneVariable
	}
	baton.sceneVariables = sceneVariableMap
	return errno.Ok, nil
}

// AssembleImportVariables 组装导入变量数据
func AssembleImportVariables(baton *Baton) (int, error) {
	// 查询所有导入的场景变量
	tx := dal.GetQuery().VariableImport
	variableImport, err := tx.WithContext(baton.Ctx).Where(
		tx.TeamID.Eq(baton.TeamID),
		tx.SceneID.In(baton.SceneIDs...),
		tx.Status.Eq(consts.VariableStatusOpen),
	).Find()

	tempData := make(map[string][]*model.VariableImport)
	if err == nil {
		for _, variableImportInfo := range variableImport {
			tempData[variableImportInfo.SceneID] = append(tempData[variableImportInfo.SceneID], variableImportInfo)
		}
	}
	baton.importVariables = tempData
	return errno.Ok, nil

}

// MakeReport 生成报告
func MakeReport(baton *Baton) (int, error) {
	var rankID int64 = 1
	newReportID := uuid.GetUUID()
	// 获取当前团队下最大的报告id
	tx := dal.GetQuery().AutoPlanReport
	autoPlanReport, err := tx.WithContext(baton.Ctx).Where(tx.TeamID.Eq(baton.TeamID)).Order(tx.RankID.Desc()).Limit(1).First()
	if err == nil {
		rankID = autoPlanReport.RankID + 1
	}

	insertData := &model.AutoPlanReport{
		ReportID:         newReportID,
		ReportName:       baton.plan.PlanName,
		RankID:           rankID,
		PlanID:           baton.PlanID,
		PlanName:         baton.plan.PlanName,
		TeamID:           baton.TeamID,
		TaskType:         baton.plan.TaskType,
		TaskMode:         int32(baton.ConfigTask.TaskMode),
		SceneRunOrder:    int32(baton.ConfigTask.SceneRunMode),
		TestCaseRunOrder: int32(baton.ConfigTask.CaseRunMode),
		Status:           consts.ReportStatusNormal,
		RunUserID:        baton.UserID,
		Remark:           baton.plan.Remark,
	}
	err = tx.WithContext(baton.Ctx).Create(insertData)
	if err != nil {
		return errno.ErrMysqlFailed, fmt.Errorf("创建报告失败")
	}
	baton.reportID = newReportID
	return errno.Ok, nil
}

// AssembleRunPlanRealParams 组装最后请求运行计划的参数
func AssembleRunPlanRealParams(baton *Baton) (int, error) {
	baton.RealRunParam.PlanId = baton.PlanID
	baton.RealRunParam.PlanName = baton.plan.PlanName
	baton.RealRunParam.ReportId = baton.reportID
	baton.RealRunParam.TeamId = baton.TeamID
	baton.RealRunParam.MachineNum = 1
	baton.RealRunParam.ConfigTask = baton.ConfigTask

	//组装全局变量
	baton.RealRunParam.GlobalVariable = baton.globalVariables

	// 组装场景
	for _, sceneInfo := range baton.scenes {
		tempData := Scene{
			PlanId:     sceneInfo.PlanID,
			SceneId:    sceneInfo.TargetID,
			IsChecked:  sceneInfo.IsChecked,
			ParentId:   sceneInfo.ParentID,
			ReportId:   baton.reportID,
			TeamId:     baton.TeamID,
			SceneName:  sceneInfo.Name,
			Version:    int64(sceneInfo.Version),
			ConfigTask: baton.ConfigTask, // 任务配置
		}

		// 组装前置条件
		if flowInfo, ok := baton.sceneFlows[sceneInfo.TargetID]; ok {
			prepositions := mao.Preposition{}
			if err := bson.Unmarshal(flowInfo.Prepositions, &prepositions); err != nil {
				log.Logger.Errorf("prepositions bson unmarshal err:%v", err)
				continue
			}
			prepositionsArr := make([]rao.Preposition, 0, len(prepositions.Prepositions))
			for _, v := range prepositions.Prepositions {
				temp := rao.Preposition{
					Type:  v.Type,
					Event: v,
				}
				prepositionsArr = append(prepositionsArr, temp)
			}
			tempData.Prepositions = prepositionsArr
		}

		// 场景导入变量
		for _, iv := range baton.importVariables[sceneInfo.TargetID] {
			temp := FileList{
				IsChecked: int64(iv.Status),
				Path:      iv.URL,
			}
			tempData.Configuration.ParameterizedFile.Paths = append(tempData.Configuration.ParameterizedFile.Paths, temp)
		}

		// 全局变量
		tempData.GlobalVariable = baton.globalVariables

		// 场景变量
		tempData.Configuration.SceneVariable = baton.sceneVariables[sceneInfo.TargetID]

		// 拼装场景下所有测试用例
		for _, tc := range baton.testCase {
			if tc.ParentID == sceneInfo.TargetID {
				testCase, err := getTestCase(tc, baton)
				if err != nil {
					return errno.ErrUnMarshalFailed, fmt.Errorf("测试用例flow数据解析失败")
				}
				tempData.Cases = append(tempData.Cases, testCase)
			}
		}
		baton.RealRunParam.Scenes = append(baton.RealRunParam.Scenes, tempData)
	}
	return errno.Ok, nil
}

// 拼装测试用例数据
func getTestCase(tc *model.Target, baton *Baton) (Scene, error) {
	res := Scene{
		PlanId:    tc.PlanID,
		IsChecked: tc.IsChecked,
		ParentId:  tc.ParentID,
		CaseId:    tc.TargetID,
		ReportId:  baton.reportID,
		TeamId:    tc.TeamID,
		SceneName: tc.Name,
		Version:   int64(tc.Version),
	}

	// 拼装场景flow
	nodes := mao.Node{}
	if err := bson.Unmarshal(baton.sceneCaseFlows[tc.TargetID].Nodes, &nodes); err != nil {
		log.Logger.Info("测试用例的flow node bson unmarshal err:%v", err)
		return Scene{}, err
	}

	edges := mao.Edge{}
	if err := bson.Unmarshal(baton.sceneCaseFlows[tc.TargetID].Edges, &edges); err != nil {
		log.Logger.Errorf("测试用例的flow edges bson unmarshal err:%v", err)
	}

	nodesRound := packer.GetNodesByLevel(nodes.Nodes, edges.Edges)
	// 把flow里面的pre_url替换成环境变量的
	for k1, v1 := range nodesRound {
		for k2, v2 := range v1 {
			if baton.sceneCaseFlows[tc.TargetID].EnvID != 0 {
				nodesRound[k1][k2].API.Request.PreUrl = v2.API.EnvInfo.PreUrl
			} else {
				nodesRound[k1][k2].API.Request.PreUrl = ""
			}
		}
	}

	res.NodesRound = nodesRound
	return res, nil
}

// RunAutoPlan 运行自动化计划
func RunAutoPlan(baton *Baton) (int, error) {
	addr := baton.balance.GetMachine(0)
	response, err := resty.New().R().SetBody(baton.RealRunParam).Post(fmt.Sprintf("http://%s/runner/run_auto_plan", addr))
	log.Logger.Info("自动化测试运行情况，req：%+v， response:%+v。 err： %+v。", proof.Render("req", baton.RealRunParam), response, err)
	if err != nil {
		_ = DeleteAutoPlanReport(baton)
		log.Logger.Info("请求压力机进行压测失败，err：", err)
		return errno.ErrHttpFailed, err
	}

	// 把计划状态改成运行中
	tx := dal.GetQuery().AutoPlan
	_, err = tx.WithContext(baton.Ctx).Where(tx.TeamID.Eq(baton.TeamID),
		tx.PlanID.Eq(baton.PlanID)).UpdateSimple(tx.Status.Value(consts.PlanStatusUnderway), tx.RunUserID.Value(baton.UserID),
		tx.RunCount.Value(baton.plan.RunCount+1))
	if err != nil {
		return errno.ErrMysqlFailed, fmt.Errorf("更新计划的状态失败")
	}

	return errno.Ok, nil
}

// DeleteAutoPlanReport 删除执行失败的计划下的所有报告
func DeleteAutoPlanReport(baton *Baton) error {
	tx := dal.GetQuery().AutoPlanReport
	_, err := tx.WithContext(baton.Ctx).Where(tx.TeamID.Eq(baton.TeamID), tx.PlanID.Eq(baton.PlanID), tx.ReportID.Eq(baton.reportID)).Delete()
	if err != nil {
		log.Logger.Info("删除报告失败")
		return err
	}
	return nil
}

// RunAutoPlanSendEmail 运行完发送邮件
func RunAutoPlanSendEmail(baton *Baton) (int, error) {
	if err := record.InsertRun(baton.Ctx, baton.TeamID, baton.UserID, record.OperationOperateRunPlan, baton.plan.PlanName); err != nil {
		return errno.ErrMysqlFailed, err
	}

	rx := dal.GetQuery().AutoPlanReport
	reportInfo, err := rx.WithContext(baton.Ctx).Where(rx.TeamID.Eq(baton.TeamID), rx.PlanID.Eq(baton.PlanID)).Order(rx.CreatedAt.Desc()).First()
	if err != nil {
		return errno.ErrMysqlFailed, err
	}

	tx := dal.GetQuery().AutoPlanEmail
	emails, err := tx.WithContext(baton.Ctx).Where(tx.TeamID.Eq(baton.TeamID), tx.PlanID.Eq(baton.PlanID)).Find()
	if err == nil && len(emails) > 0 {
		ttx := dal.GetQuery().Team
		teamInfo, err := ttx.WithContext(baton.Ctx).Where(ttx.TeamID.Eq(baton.TeamID)).First()
		if err != nil {
			return errno.ErrMysqlFailed, err
		}

		ux := dal.GetQuery().User
		user, err := ux.WithContext(baton.Ctx).Where(ux.UserID.Eq(reportInfo.RunUserID)).First()
		if err != nil {
			return errno.ErrMysqlFailed, err
		}

		for _, email := range emails {
			if err := mail.SendAutoPlanEmail(email.Email, baton.plan, teamInfo, user.Nickname, reportInfo.ReportID); err != nil {
				return errno.ErrMysqlFailed, err
			}
		}
	}
	return errno.Ok, nil
}
