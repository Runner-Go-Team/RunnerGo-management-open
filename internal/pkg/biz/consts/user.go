package consts

const (
	RoleTypeOwner  = 1 // 超级管理员
	RoleTypeMember = 2 // 普通成员
	RoleTypeAdmin  = 3 // 管理员

	WechatQrCodeExpiresTime = 300 // 微信登录二维码过期时间，单位：秒
)

var (
	DefaultAvatarMemo = map[int]string{
		0: "https://apipost.oss-cn-beijing.aliyuncs.com/kunpeng/avatar/default1.png",
		1: "https://apipost.oss-cn-beijing.aliyuncs.com/kunpeng/avatar/default2.png",
		2: "https://apipost.oss-cn-beijing.aliyuncs.com/kunpeng/avatar/default3.png",
		3: "https://apipost.oss-cn-beijing.aliyuncs.com/kunpeng/avatar/default4.png",
	}
)
