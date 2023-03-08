package rao

type ListOperationReq struct {
	TeamID string `form:"team_id" binding:"required,gt=0"`
	Page   int    `form:"page,default=1"`
	Size   int    `form:"size,default=20"`
}

type ListOperationResp struct {
	Operations []*Operation `json:"operations"`
	Total      int64        `json:"total"`
}

type Operation struct {
	UserID         string `json:"user_id"`
	UserName       string `json:"user_name"`
	UserAvatar     string `json:"user_avatar"`
	UserStatus     int32  `json:"user_status"`
	Category       int32  `json:"category"`
	Operate        int32  `json:"operate"`
	Name           string `json:"name"`
	CreatedTimeSec int64  `json:"created_time_sec"`
}
