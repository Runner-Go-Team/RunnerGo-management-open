package rao

import "github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"

type FeiShuRobot struct {
	WebhookURL string `json:"webhook_url"`
	Secret     string `json:"secret"`
}

type FeiShuApp struct {
	AppID     string `json:"app_id"`
	AppSecret string `json:"app_secret"`
}

type WechatRobot struct {
	WebhookURL string `json:"webhook_url"`
}

type WechatApp struct {
	ID        string `json:"id"`
	AppID     string `json:"app_id"`
	AppSecret string `json:"app_secret"`
}

type SMTPEmail struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type DingTalkRobot struct {
	WebhookURL string `json:"webhook_url"`
	Secret     string `json:"secret"`
}

type DingTalkApp struct {
	AgentId   string `json:"agent_id"`
	AppKey    string `json:"app_key"`
	AppSecret string `json:"app_secret"`
}

type SendNoticeParams struct {
	EventID        int32    `json:"event_id" binding:"required"`
	TeamID         string   `json:"team_id"`
	ReportIDs      []string `json:"report_ids"`
	NoticeGroupIDs []string `json:"notice_group_ids" binding:"required"`
}

type SendCardParams struct {
	PlanName          string                    `json:"plan_name"`
	TaskTypeName      string                    `json:"task_type_name"`
	RunUserName       string                    `json:"run_user_name"`
	ReportType        int32                     `json:"report_type"`
	StressPlanReports []*model.StressPlanReport `json:"stress_plan_reports"`
	AutoPlanReports   []*model.AutoPlanReport   `json:"auto_plan_report"`
	Team              *model.Team               `json:"team"`
}

type NoticeGroupRelate struct {
	NoticeID string                   `json:"notice_id"`
	Params   *NoticeGroupRelateParams `json:"params"`
}

type NoticeGroupRelateParams struct {
	UserIDs []string `json:"user_ids,omitempty"`
	Emails  []string `json:"emails,omitempty"`
}

type SaveNoticeEventReq struct {
	EventID  int32    `json:"event_id" binding:"required"`
	GroupIDs []string `json:"notice_group_ids" binding:"required"`
	PlanIDs  []string `json:"plan_ids"`
	TeamID   string   `json:"team_id"`
}

type GetGroupNoticeEventReq struct {
	EventID int32  `form:"event_id" binding:"required"`
	PlanID  string `form:"plan_id"`
	TeamID  string `form:"team_id"`
}

type GetGroupNoticeEventResp struct {
	NoticeGroupIDs []string `json:"notice_group_ids"`
}

type SendNoticeReq struct {
	EventID        int32    `json:"event_id" binding:"required"`
	TeamID         string   `json:"team_id"`
	ReportIDs      []string `json:"report_ids"`
	NoticeGroupIDs []string `json:"notice_group_ids" binding:"required"`
}

type ThirdUserInfo struct {
	OpenID string `json:"open_id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

type ThirdDepartmentInfo struct {
	DepartmentID   string                `json:"department_id"`
	Name           string                `json:"name"`
	MemberCount    int                   `json:"member_count"`
	UserList       []ThirdUserInfo       `json:"user_list"`
	DepartmentList []ThirdDepartmentInfo `json:"department_list"`
}

type ThirdCompanyUsers struct {
	DepartmentList []ThirdDepartmentInfo `json:"department_list"`
	UserList       []ThirdUserInfo       `json:"user_list"`
}
