package mao

import "time"

type TcpDetail struct {
	TargetID    string    `bson:"target_id"`
	TeamID      string    `bson:"team_id"`
	EnvInfo     EnvInfo   `bson:"env_info"`
	Url         string    `bson:"url"`
	SendMessage string    `bson:"send_message"`
	MessageType string    `bson:"message_type"`
	TcpConfig   TcpConfig `bson:"tcp_config"`
	CreatedAt   time.Time `bson:"created_at"`
}

type TcpConfig struct {
	ConnectType         int `bson:"connect_type"`           // 连接类型：1-长连接，2-短连接
	IsAutoSend          int `bson:"is_auto_send"`           // 是否自动发送消息：0-非自动，1-自动
	ConnectDurationTime int `bson:"connect_duration_time"`  // 连接持续时长
	SendMsgDurationTime int `bson:"send_msg_duration_time"` // 发送消息间隔时长
	ConnectTimeoutTime  int `bson:"connect_timeout_time"`   // 连接超时时间，单位：毫秒
	RetryNum            int `bson:"retry_num"`              // 重连次数
	RetryInterval       int `bson:"retry_interval"`         // 重连间隔时间，单位：毫秒
}

type TcpDebug struct {
	TargetID     string `bson:"target_id"`
	TeamID       string `bson:"team_id"`
	Uuid         string `bson:"uuid"`
	Name         string `bson:"name"`
	IsStop       bool   `bson:"is_stop"`
	Type         string `bson:"type"`
	RequestBody  string `bson:"request_body"`
	ResponseBody string `bson:"response_body"`
	Status       string `bson:"status"`
}
