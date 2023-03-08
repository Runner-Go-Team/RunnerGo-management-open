package consts

const (
	SmsLogTypeSignUp         = 1 // 注册验证
	SmsLogTypeLogin          = 2 // 登录验证
	SmsLogTypeForgetPassword = 3 // 找回密码验证

	SmsLogSendStatusSuc  = 1 //发送成功
	SmsLogSendStatusFail = 2 //发送失败

	SmsLogVerifyStatusWait   = 1 //未校验
	SmsLogVerifyStatusFinish = 2 //已校验

	SmsPerHourSendCountLimit = 40 //短信条数限制 每小时
)
