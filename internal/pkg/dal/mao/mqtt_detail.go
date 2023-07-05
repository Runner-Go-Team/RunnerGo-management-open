package mao

import "time"

type MqttDetail struct {
	TargetID     string       `bson:"target_id"`
	TeamID       string       `bson:"team_id"`
	EnvInfo      EnvInfo      `bson:"env_info"`
	Topic        string       `bson:"topic"`
	SendMessage  string       `bson:"send_message"`
	CommonConfig CommonConfig `bson:"common_config"`
	HigherConfig HigherConfig `bson:"higher_config"`
	WillConfig   WillConfig   `bson:"will_config"`
	CreatedAt    time.Time    `bson:"created_at"`
}

type CommonConfig struct {
	ClientName string   `bson:"client_name"` // 客户端名称
	Username   string   `bson:"username"`    // 用户名
	Password   string   `bson:"password"`    // 密码
	IsEncrypt  bool     `bson:"is_encrypt"`  // 是否开启加密
	AuthFile   AuthFile `bson:"auth_file"`   // 认证文件
}
type AuthFile struct {
	FileName string `bson:"file_name"`
	FileUrl  string `bson:"file_url"`
}

type HigherConfig struct {
	ConnectTimeoutTime  int    `bson:"connect_timeout_time"`   // 连接超时时间，单位：秒
	KeepAliveTime       int    `bson:"keep_alive_time"`        // 保持连接时长，单位：秒
	IsAutoRetry         bool   `bson:"is_auto_retry"`          // 是否开启自动重连，true-开启，false-关闭
	RetryNum            int    `bson:"retry_num"`              // 重连次数
	RetryInterval       int    `bson:"retry_interval"`         // 重连间隔时间，单位：秒
	MqttVersion         string `bson:"mqtt_version"`           // mqtt版本号
	DialogueTimeout     int    `bson:"dialogue_timeout"`       // 会话过期时间，单位：秒
	IsSaveMessage       bool   `bson:"is_save_message"`        // 是否保留消息
	ServiceQuality      int    `bson:"service_quality"`        // 服务质量:0-至多一次，1-至少一次，2-确保只有一次
	SendMsgIntervalTime int    `bson:"send_msg_interval_time"` // 发送消息间隔时间，单位：秒
}

type WillConfig struct {
	WillTopic      string `bson:"will_topic"`      // 遗愿主题
	IsOpenWill     bool   `bson:"is_open_will"`    // 是否开启遗愿
	ServiceQuality int    `bson:"service_quality"` // 服务质量:0-至多一次，1-至少一次，2-确保只有一次
}

type MqttDebug struct {
	TargetID      string  `bson:"target_id"`
	TeamID        string  `bson:"team_id"`
	Uuid          string  `bson:"uuid"`
	Name          string  `bson:"name"`
	RequestTime   int64   `bson:"request_time"`
	ErrorType     int64   `bson:"error_type"`
	IsSucceed     bool    `bson:"is_succeed"`
	SendBytes     float64 `bson:"send_bytes"`
	ReceivedBytes float64 `bson:"received_bytes"`
	ErrorMsg      string  `bson:"error_msg"`
	Timestamp     int64   `bson:"timestamp"`
	StartTime     int64   `bson:"start_time"`
	EndTime       int64   `bson:"end_time"`
}
