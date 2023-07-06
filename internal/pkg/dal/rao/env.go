package rao

type EnvListReq struct {
	TeamID string `json:"team_id" binding:"required,gt=0"`
	Name   string `json:"name"`
}

type EnvListResp struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	TeamID string `json:"team_id"`
	Sort   int32  `json:"sort"`
	//Status int8   `json:"status"`
	ServiceList []ServiceListResp `json:"service_list"`
}

type GetEnvListReq struct {
	TeamID string `json:"team_id" binding:"required"`
	Name   string `json:"name"`
}

type GetEnvListResp struct {
	TeamID  string `json:"team_id"`
	EnvID   int64  `json:"env_id"`
	EnvName string `json:"env_name"`
}

type UpdateEnvReq struct {
	EnvID   int64  `json:"env_id"`
	EnvName string `json:"env_name" binding:"required"`
	TeamID  string `json:"team_id" binding:"required"`
}

type CreateEnvReq struct {
	TeamID string `json:"team_id" binding:"required"`
}
type CreateEnvResp struct {
	EnvID   int64  `json:"env_id"`
	EnvName string `json:"env_name"`
}

type CopyEnvReq struct {
	EnvID  int64  `json:"env_id" binding:"required"`
	TeamID string `json:"team_id" binding:"required"`
}

type SaveServiceListReq struct {
	ID        int64  `json:"id"`
	TeamEnvID int64  `json:"env_id"`
	Name      string `json:"name" binding:"required"`
	Content   string `json:"content" binding:"required"`
	Sort      int32  `json:"sort"`
}

type UpdateEnvResp struct {
	EnvID   int64  `json:"env_id"`
	EnvName string `json:"env_name"`
	TeamID  string `json:"team_id"`
}

type CopyEnvResp struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	TeamID string `json:"team_id"`
	Sort   int32  `json:"sort"`
}

type DelEnvReq struct {
	EnvID  int64  `json:"env_id" binding:"required"`
	TeamID string `json:"team_id" binding:"required"`
}

type DelEnvResp struct {
}

type ServiceListReq struct {
	EnvID  int64  `json:"env_id" binding:"required,gt=0"`
	TeamID string `json:"team_id" binding:"required,gt=0"`
	Name   string `json:"name"`
}

type ServiceListResp struct {
	ID        int64  `json:"id"`
	TeamEnvID int64  `json:"env_id"`
	TeamID    string `json:"team_id"`
	Name      string `json:"name"`
	Content   string `json:"content"`
}

type SaveServiceReq struct {
	ID     int64  `json:"id"`
	Name   string `json:"name" binding:"required"`
	EnvID  int64  `json:"env_id" binding:"required,gt=0"`
	TeamID string `json:"team_id" binding:"required,gt=0"`
	Sort   int32  `json:"sort"`
}

type SaveServiceResp struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	EnvID  int64  `json:"emv_id"`
	TeamID string `json:"team_id"`
	Sort   int32  `json:"sort"`
}

type DelEnvServiceReq struct {
	ServiceID int64  `json:"service_id" binding:"required"`
	EnvID     int64  `json:"env_id" binding:"required"`
	TeamID    string `json:"team_id" binding:"required"`
}

type DelEnvDatabaseReq struct {
	DatabaseID int64  `json:"database_id" binding:"required"`
	EnvID      int64  `json:"env_id" binding:"required"`
	TeamID     string `json:"team_id" binding:"required"`
}

type GetServiceListReq struct {
	TeamID string `json:"team_id"`
	EnvID  int64  `json:"env_id"`
	Page   int    `json:"page"`
	Size   int    `json:"size"`
}

type ServiceList struct {
	ServiceID   int64  `json:"service_id"`
	TeamID      string `json:"team_id"`
	EnvID       int64  `json:"env_id"`
	ServiceName string `json:"service_name"`
	Content     string `json:"content"`
}

type GetServiceListResp struct {
	ServiceList []ServiceList `json:"service_list"`
	Total       int64         `json:"total"`
}

type GetDatabaseListReq struct {
	TeamID string `json:"team_id"`
	EnvID  int64  `json:"env_id"`
	Page   int    `json:"page"`
	Size   int    `json:"size"`
}

type GetDatabaseListResp struct {
	DatabaseList []DatabaseList `json:"database_list"`
	Total        int64          `json:"total"`
}

type DatabaseList struct {
	DatabaseID int64  `json:"database_id"`
	TeamID     string `json:"team_id"`
	EnvID      int64  `json:"env_id"`
	Type       string `json:"type"`
	ServerName string `json:"server_name"`
	Host       string `json:"host"`
	User       string `json:"user"`
	Password   string `json:"password"`
	Port       int32  `json:"port"`
	DbName     string `json:"db_name"`
	Charset    string `json:"charset"`
}

type CreateEnvServiceReq struct {
	ServiceID   int64  `json:"service_id"`
	TeamID      string `json:"team_id"`
	EnvID       int64  `json:"env_id"`
	ServiceName string `json:"service_name"`
	Content     string `json:"content"`
}

type CreateEnvDatabaseReq struct {
	DatabaseID int64  `json:"database_id"`
	TeamID     string `json:"team_id"`
	EnvID      int64  `json:"env_id"`
	Type       string `json:"type"`
	ServerName string `json:"server_name"`
	Host       string `json:"host"`
	User       string `json:"user"`
	Password   string `json:"password"`
	Port       int32  `json:"port"`
	DbName     string `json:"db_name"`
	Charset    string `json:"charset"`
}
