package rao

type RegisterOrLoginReq struct {
	Mobile           string `json:"mobile"`
	VerifyCode       string `json:"verify_code"`
	IsAutoLogin      bool   `json:"is_auto_login"`
	UtmSource        string `json:"utm_source"`
	InviteVerifyCode string `json:"invite_verify_code"`
}

type AuthSignupReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type AuthSignupResp struct {
	Token         string `json:"token"`
	ExpireTimeSec int64  `json:"expire_time_sec"`
}

type AuthLoginReq struct {
	Email            string `json:"email" binding:"required"`
	Password         string `json:"password" binding:"required,min=6,max=32"`
	IsAutoLogin      bool   `json:"is_auto_login"`
	InviteVerifyCode string `json:"invite_verify_code"`
}

type Captcha struct {
	CaptchaId     string `json:"captcha_id"`
	LotNumber     string `json:"lot_number"`
	PassToken     string `json:"pass_token"`
	GenTime       string `json:"gen_time"`
	CaptchaOutput string `json:"captcha_output"`
}

type AuthLoginResp struct {
	Token         string `json:"token"`
	ExpireTimeSec int64  `json:"expire_time_sec"`
	TeamID        string `json:"team_id"`
	IsRegister    bool   `json:"is_register"`
}

type MobileAuthLoginReq struct {
	Mobile      string `json:"mobile" binding:"required"`
	Password    string `json:"password" binding:"required,min=6,max=32"`
	IsAutoLogin bool   `json:"is_auto_login"`
	VerifyCode  string `json:"verify_code"`
}

type MobileAuthLoginResp struct {
	Token         string `json:"token"`
	ExpireTimeSec int64  `json:"expire_time_sec"`
	TeamID        string `json:"team_id"`
	IsRegister    bool   `json:"is_register"`
}

type SetUserSettingsReq struct {
	UserSettings UserSettings `json:"settings"`
}

type SetUserSettingsResp struct {
}

type GetUserSettingsReq struct {
}

type GetUserSettingsResp struct {
	UserSettings *UserSettings `json:"settings"`
	UserInfo     *UserInfo     `json:"user_info"`
}

type UserSettings struct {
	CurrentTeamID string `json:"current_team_id" binding:"required,gt=0"`
}

type UserInfo struct {
	ID              int64  `json:"id"`
	Email           string `json:"email"`
	Mobile          string `json:"mobile"`
	Nickname        string `json:"nickname"`
	Avatar          string `json:"avatar"`
	RoleID          string `json:"role_id"`
	Account         string `json:"account"`
	RoleName        string `json:"role_name"`
	UserID          string `json:"user_id"`
	CompanyRoleID   string `json:"company_role_id"`
	CompanyRoleName string `json:"company_role_name"`
}

type ForgetPasswordReq struct {
	Mobile  string  `json:"mobile" binding:"required"`
	Captcha Captcha `json:"captcha"`
}

type ForgetPasswordResp struct {
}

type AuthResetPasswordReq struct {
	Password       string `json:"password"`
	RepeatPassword string `json:"repeat_password"`
}

type AuthResetPasswordResp struct {
}

type UpdatePasswordReq struct {
	NewPassword    string `json:"new_password" binding:"required,min=6,eqfield=RepeatPassword"`
	RepeatPassword string `json:"repeat_password" binding:"required,min=6"`
}

type UpdatePasswordResp struct {
}

type UpdateNicknameReq struct {
	Nickname string `json:"nickname" binding:"required,min=2"`
}

type UpdateNicknameResp struct {
}

type UpdateAvatarReq struct {
	AvatarURL string `json:"avatar_url" binding:"required"`
}

type UpdateAvatarResp struct {
}

type VerifyPasswordReq struct {
	Password string `json:"password"`
}

type VerifyPasswordResp struct {
	IsMatch bool `json:"is_match"`
}

type ResetPasswordReq struct {
	//U              string `json:"u"`
	Mobile         string `json:"mobile"`
	NewPassword    string `json:"new_password"`
	RepeatPassword string `json:"repeat_password"`
	VerifyCode     string `json:"verify_code"`
}

type ResetPasswordResp struct {
}

type GetVerifyCodeReq struct {
	Content string `json:"content" binding:"required"`
}

type GetSmsCodeReq struct {
	Mobile  string  `json:"mobile" binding:"required,min=11,max=11"`
	Type    int32   `json:"type"`
	Captcha Captcha `json:"captcha"`
}

type VerifySmsCodeReq struct {
	Mobile string `json:"mobile" binding:"required,min=11,max=11"`
	Code   string `json:"code" binding:"required"`
	Type   int32  `json:"type"`
}

type SmsCodeLoginReq struct {
	Mobile      string `json:"mobile" binding:"required"`
	Code        string `json:"code" binding:"required"`
	IsAutoLogin bool   `json:"is_auto_login"`
}

type VerifySmsCodeResp struct {
	U string `json:"u"`
}

type CheckEmailIsRegisterReq struct {
	Email string `json:"email"`
}

type CheckEmailIsRegisterResp struct {
	RegisterStatus bool `json:"register_status"`
}

type CheckUserIsRegisterReq struct {
	Email   string  `json:"email"`
	Mobile  string  `json:"mobile"`
	Type    int32   `json:"type"`
	Captcha Captcha `json:"captcha"`
}

type CheckUserIsRegisterResp struct {
	RegisterStatus bool `json:"register_status"`
}

type SetUserPasswordReq struct {
	Mobile   string `json:"mobile" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type CollectUserInfoReq struct {
	TeamID   string `json:"team_id"`
	TeamName string `json:"team_name"`
	Nickname string `json:"nickname"`
	Industry string `json:"industry"`
	TeamSize string `json:"team_size"`
	WorkType string `json:"work_type"`
}

type GetCollectUserInfoReq struct {
	TeamID string `json:"team_id"`
}

type GetCollectUserInfoResp struct {
	Nickname          string `json:"nickname"`
	TeamName          string `json:"team_name"`
	IsCollectUserInfo bool   `json:"is_collect_user_info"`
}

type GetWechatLoginQrCodeResult struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Ticket        string `json:"ticket"`
		ExpireSeconds int    `json:"expire_seconds"`
		Url           string `json:"url"`
	} `json:"data"`
}

type GetWechatLoginQrCodeParam struct {
	ExpireSeconds int64 `json:"expire_seconds"`
}

type GetWechatLoginQrCodeResp struct {
	WechatLoginUrl string `json:"wechat_login_url"`
	Ticket         string `json:"ticket"`
}

type GetWechatLoginResultReq struct {
	Ticket string `json:"ticket" binding:"required"`
}

type GetWechatLoginResult struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		OpenId string `json:"openid"`
	} `json:"data"`
}

type GetWechatLoginResultResp struct {
	ScanStatus int    `json:"scan_status"`
	Openid     string `json:"openid"`
	IsRegister bool   `json:"is_register"`
	UserID     string `json:"user_id"`
}
type WechatRegisterOrLoginReq struct {
	Mobile           string `json:"mobile"`
	VerifyCode       string `json:"verify_code"`
	Ticket           string `json:"ticket"`
	IsAutoLogin      bool   `json:"is_auto_login"`
	UtmSource        string `json:"utm_source"`
	InviteVerifyCode string `json:"invite_verify_code"`
}

type GetWechatLoginResp struct {
	Token         string `json:"token"`
	ExpireTimeSec int64  `json:"expire_time_sec"`
	TeamID        string `json:"team_id"`
	IsRegister    bool   `json:"is_register"`
	ScanStatus    int    `json:"scan_status"`
}

type CheckWechatIsChangeBindReq struct {
	Mobile string `json:"mobile" binding:"required"`
}

type CheckWechatIsChangeBindResp struct {
	IsChangeBind bool `json:"is_change_bind"`
}

type UpdateEmailReq struct {
	Email string `json:"email"`
}

type UpdateAccountReq struct {
	Account string `json:"account" binding:"required,min=6,max=30"`
}
