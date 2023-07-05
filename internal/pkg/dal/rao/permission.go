package rao

type SendPermissionApiResp struct {
	Code int         `json:"code"`
	Em   string      `json:"em"`
	Et   string      `json:"et"`
	Data interface{} `json:"data"`
}

type GetTeamCompanyMembersReq struct {
	TeamID  string `form:"team_id" json:"team_id"`
	Keyword string `form:"keyword" json:"keyword"`
	Page    int    `form:"page,default=1"`
	Size    int    `form:"size,default=20"`
}

type TeamMembersSaveReq struct {
	TeamID  string        `json:"team_id"`
	UserID  string        `json:"user_id"`
	Members []MembersList `json:"members"`
}

type MembersList struct {
	UserID     string `json:"user_id"`
	TeamRoleID string `json:"team_role_id"`
}

type GetRoleMemberInfoReq struct {
	TeamID    string `json:"team_id"`
	RoleID    string `json:"role_id"`
	UserID    string `json:"user_id"`
	CompanyId string `json:"company_id"`
}

type GetNoticeGroupListReq struct {
	Keyword   string `form:"keyword"`
	ChannelID int64  `form:"channel_id"`
}

type GetNoticeThirdUsersReq struct {
	NoticeID string `form:"notice_id"`
}
