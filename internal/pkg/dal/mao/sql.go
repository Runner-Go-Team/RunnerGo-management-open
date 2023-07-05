package mao

import (
	"time"
)

type SqlDetailForMg struct {
	TargetID        string          `bson:"target_id"`
	TeamID          string          `bson:"team_id"`
	SqlString       string          `bson:"sql_string"`        // sql语句
	Assert          []SqlAssert     `bson:"assert"`            // 验证的方法(断言)
	Regex           []SqlRegex      `bson:"regex"`             // 正则表达式
	SqlDatabaseInfo SqlDatabaseInfo `bson:"sql_database_info"` // 数据库详情
	EnvInfo         EnvInfo         `bson:"env_info"`
	CreatedAt       time.Time       `bson:"created_at"`
}

type SqlDatabaseInfo struct {
	Type       string `bson:"type"`
	ServerName string `bson:"server_name"`
	Host       string `bson:"host"`
	User       string `bson:"user"`
	Password   string `bson:"password"`
	Port       int32  `bson:"port"`
	DbName     string `bson:"db_name"`
	Charset    string `bson:"charset"`
}

type SqlAssert struct {
	IsChecked int    `bson:"is_checked"`
	Field     string `bson:"field"`
	Compare   string `bson:"compare"`
	Val       string `bson:"val"`
	Index     int    `bson:"index"` // 断言时提取第几个值
}

type SqlRegex struct {
	IsChecked int    `bson:"is_checked"` // 1 选中, -1未选
	Var       string `bson:"var"`
	Field     string `bson:"field"`
	Index     int    `bson:"index"` // 正则时提取第几个值
}
