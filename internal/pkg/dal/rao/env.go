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

type SaveEnvReq struct {
	ID          int64                `json:"id"`
	Name        string               `json:"name" binding:"required"`
	TeamID      string               `json:"team_id" binding:"required,gt=0"`
	Sort        int32                `json:"sort"`
	ServiceList []SaveServiceListReq `json:"service_list"`
}

type CopyEnvReq struct {
	ID     int64  `json:"id"`
	TeamID string `json:"team_id" binding:"required,gt=0"`
}

type SaveServiceListReq struct {
	ID        int64  `json:"id"`
	TeamEnvID int64  `json:"team_env_id"`
	Name      string `json:"name" binding:"required"`
	Content   string `json:"content" binding:"required"`
	Sort      int32  `json:"sort"`
}

type SaveEnvResp struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	TeamID string `json:"team_id"`
	Sort   int32  `json:"sort"`
}

type CopyEnvResp struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	TeamID string `json:"team_id"`
	Sort   int32  `json:"sort"`
}

type DelEnvReq struct {
	ID     int64  `json:"id" binding:"required,gt=0"`
	TeamID string `json:"team_id" binding:"required,gt=0"`
}

type DelEnvResp struct {
}

type ServiceListReq struct {
	EnvID  int64  `json:"team_env_id" binding:"required,gt=0"`
	TeamID string `json:"team_id" binding:"required,gt=0"`
	Name   string `json:"name"`
}

type ServiceListResp struct {
	ID        int64  `json:"id"`
	TeamEnvID int64  `json:"team_env_id"`
	TeamID    string `json:"team_id"`
	Name      string `json:"name"`
	Content   string `json:"content"`
}

type SaveServiceReq struct {
	ID     int64  `json:"id"`
	Name   string `json:"name" binding:"required"`
	EnvID  int64  `json:"team_env_id" binding:"required,gt=0"`
	TeamID string `json:"team_id" binding:"required,gt=0"`
	Sort   int32  `json:"sort"`
}

type SaveServiceResp struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	EnvID  int64  `json:"team_env_id"`
	TeamID string `json:"team_id"`
	Sort   int32  `json:"sort"`
}

type DelServiceReq struct {
	ID     int64  `json:"id" binding:"required,gt=0"`
	EnvID  int64  `json:"team_env_id" binding:"required,gt=0"`
	TeamID string `json:"team_id" binding:"required,gt=0"`
}
