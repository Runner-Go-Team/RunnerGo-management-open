package mao

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
)

type API struct {
	TargetID     string       `bson:"target_id"`
	URL          string       `bson:"url"`
	EnvInfo      EnvInfo      `bson:"env_info"`
	Header       bson.Raw     `bson:"header"`
	Query        bson.Raw     `bson:"query"`
	Cookie       bson.Raw     `bson:"cookie"`
	Body         bson.Raw     `bson:"body"`
	Auth         bson.Raw     `bson:"auth"`
	Description  string       `bson:"description"`
	Assert       bson.Raw     `bson:"assert"`
	Regex        bson.Raw     `bson:"regex"`
	HttpApiSetup HttpApiSetup `bson:"http_api_setup"`
}

type EnvInfo struct {
	EnvID       int64  `bson:"env_id"`
	EnvName     string `bson:"env_name"`
	ServiceID   int64  `bson:"service_id"`
	ServiceName string `bson:"service_name"`
	PreUrl      string `bson:"pre_url"`
	DatabaseID  int64  `bson:"database_id"`
	ServerName  string `bson:"server_name"`
}

type HttpApiSetup struct {
	IsRedirects         int    `bson:"is_redirects"`  // 是否跟随重定向 0: 是   1：否
	RedirectsNum        int    `bson:"redirects_num"` // 重定向次数>= 1; 默认为3
	ReadTimeOut         int    `bson:"read_time_out"` // 请求超时时间
	WriteTimeOut        int    `bson:"write_time_out"`
	ClientName          string `bson:"client_name"`
	KeepAlive           bool   `bson:"keep_alive"`
	MaxIdleConnDuration int32  `bson:"max_idle_conn_duration"`
	MaxConnPerHost      int32  `bson:"max_conn_per_host"`
	UserAgent           bool   `bson:"user_agent"`
	MaxConnWaitTimeout  int64  `bson:"max_conn_wait_timeout"`
}

type Assert struct {
	Assert []rao.Assert `bson:"assert"`
}

type Regex struct {
	Regex []rao.Regex `bson:"regex"`
}
