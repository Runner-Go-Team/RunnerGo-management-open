package consts

const (
	NoticeEventStressPlan       = 101 // 性能计划（可能有多个报告）
	NoticeEventStressPlanReport = 102 // 性能计划报告
	NoticeEventAuthPlan         = 103 // 自动计划（可能有多个报告）
	NoticeEventAuthPlanReport   = 104 // 自动计划报告

	NoticeStatusNormal = 1 // 通知状态 1:启用 2:禁用
	NoticeStatusClose  = 2

	NoticeChannelIDFRobot    = 1 // 1:飞书群机器人
	NoticeChannelIDFApp      = 2 // 2:飞书企业应用
	NoticeChannelIDWxApp     = 3 // 3:企业微信应用
	NoticeChannelIDWxRobot   = 4 // 4:企业微信机器人
	NoticeChannelIDEmail     = 5 // 5:邮箱
	NoticeChannelIDDingRobot = 6 // 6:钉钉群机器人s
	NoticeChannelIDDingApp   = 7 // 7:钉钉应用
)
