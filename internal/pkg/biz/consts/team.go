package consts

const (
	TeamTypePrivate = 1 // 私有团队
	TeamTypeNormal  = 2 // 普通团队

	PrivateTeamTrialTime  = 360  //私有团队试用期7天 单位:h
	PrivateTeamDefaultVum = 5000 //私有团队试用期默认赠送vum数

	IsVipTypeNo  = 1 // 否
	IsVipTypeYes = 2 // 是

	// 个人版试用期内赠送成员数
	TrialTeamGiftUserNum = 1

	// 所有团队类型的最大并发数，固定是50000
	AllTeamMaxConcurrency = 50000

	// 团队是否可用
	TeamNotCanUse = 0 // 不可用
	TeamCanUse    = 1 // 可用
)
