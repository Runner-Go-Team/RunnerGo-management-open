package handler

import (
	"errors"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/log"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/conf"
	"gorm.io/gorm"
	"time"

	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/errno"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/jwt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/response"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/auth"
	"github.com/go-omnibus/omnibus"

	"github.com/gin-gonic/gin"
)

// UserRegister 注册
func UserRegister(ctx *gin.Context) {
	var req rao.AuthSignupReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	tx := dal.GetQuery().User
	cnt, err := tx.WithContext(ctx).Where(tx.Email.Eq(req.Email)).Count()
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}
	if cnt > 0 {
		response.ErrorWithMsg(ctx, errno.ErrYetRegister, "")
		return
	}

	_, err = auth.SignUp(ctx, req.Email, req.Password)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}

// AuthLogin 登录
func AuthLogin(ctx *gin.Context) {
	var req rao.AuthLoginReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	userInfo, err := auth.Login(ctx, req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrAuthFailed, "")
		return
	}

	// 开始生成token
	expireTime := conf.Conf.AboutTimeConfig.DefaultTokenExpireTime
	d := expireTime * time.Hour
	if req.IsAutoLogin {
		d = 30 * 24 * time.Hour
	}

	token, exp, err := jwt.GenerateTokenByTime(userInfo.UserID, d)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrInvalidToken, err.Error())
		return
	}

	if err := auth.UpdateLoginTime(ctx, userInfo.UserID); err != nil {
		log.Logger.Errorf("update login time err %s", err)
	}

	defaultTeamID := ""
	if req.InviteVerifyCode == "" {
		defaultTeamID, _ = auth.GetAvailTeamID(ctx, userInfo.UserID)
		userSettings := rao.UserSettings{
			CurrentTeamID: defaultTeamID,
		}
		_ = auth.SetUserSettings(ctx, userInfo.UserID, &userSettings)
	}

	response.SuccessWithData(ctx, rao.AuthLoginResp{
		Token:         token,
		ExpireTimeSec: exp.Unix(),
		TeamID:        defaultTeamID,
		IsRegister:    true,
	})
	return
}

// MobileAuthLogin 手机号密码登录
func MobileAuthLogin(ctx *gin.Context) {
	var req rao.MobileAuthLoginReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	u, err := auth.MobileLogin(ctx, req.Mobile, req.Password, req.VerifyCode)
	if err != nil {
		if err.Error() == "record not found" {
			response.ErrorWithMsg(ctx, errno.ErrAuthFailedNotRegistered, "")
		}
		if err.Error() == "crypto/bcrypt: hashedPassword is not the hash of the given password" {
			response.ErrorWithMsg(ctx, errno.ErrAuthFailed, "")
		}
		return
	}

	d := 7 * 24 * time.Hour
	if req.IsAutoLogin {
		d = 30 * 24 * time.Hour
	}

	token, exp, err := jwt.GenerateTokenByTime(u.UserID, d)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrInvalidToken, err.Error())
		return
	}

	if err := auth.UpdateLoginTime(ctx, u.UserID); err != nil {
		log.Logger.Errorf("update login time err %s", err)
	}

	availTeamID, _ := auth.GetAvailTeamID(ctx, u.UserID)
	userSettings := rao.UserSettings{
		CurrentTeamID: availTeamID,
	}
	_ = auth.SetUserSettings(ctx, u.UserID, &userSettings)

	response.SuccessWithData(ctx, rao.MobileAuthLoginResp{
		Token:         token,
		ExpireTimeSec: exp.Unix(),
		TeamID:        availTeamID,
	})
	return
}

// RefreshToken 续期
func RefreshToken(ctx *gin.Context) {
	tokenString := ctx.GetHeader("Authorization")

	token, exp, err := jwt.RefreshToken(tokenString)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, rao.AuthLoginResp{
		Token:         token,
		ExpireTimeSec: exp.Unix(),
	})
	return
}

func UpdatePassword(ctx *gin.Context) {
	var req rao.UpdatePasswordReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	if req.NewPassword != req.RepeatPassword {
		response.ErrorWithMsg(ctx, errno.ErrParam, "两次密码输入不一致")
		return
	}

	userID := jwt.GetUserIDByCtx(ctx)

	tx := dal.GetQuery().User
	_, err := tx.WithContext(ctx).Where(tx.UserID.Eq(userID)).First()
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, "查询用户信息失败")
		return
	}

	hashedPassword, err := omnibus.GenerateBcryptFromPassword(req.NewPassword)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, "获取加密密码失败")
		return
	}
	if _, err := tx.WithContext(ctx).Where(tx.UserID.Eq(userID)).UpdateSimple(tx.Password.Value(hashedPassword)); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, "密码修改失败")
		return
	}

	response.Success(ctx)
	return
}

func UpdateNickname(ctx *gin.Context) {
	var req rao.UpdateNicknameReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	tx := dal.GetQuery().User
	if _, err := tx.WithContext(ctx).Where(tx.UserID.Eq(jwt.GetUserIDByCtx(ctx))).UpdateColumn(tx.Nickname, req.Nickname); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, "nickname failed")
		return
	}

	response.Success(ctx)
	return
}

func UpdateAvatar(ctx *gin.Context) {
	var req rao.UpdateAvatarReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	tx := dal.GetQuery().User
	if _, err := tx.WithContext(ctx).Where(tx.UserID.Eq(jwt.GetUserIDByCtx(ctx))).UpdateColumn(tx.Avatar, req.AvatarURL); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, "password = new password")
		return
	}

	response.Success(ctx)
	return
}

// SetUserSettings 设置用户配置
func SetUserSettings(ctx *gin.Context) {
	var req rao.SetUserSettingsReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	if err := auth.SetUserSettings(ctx, jwt.GetUserIDByCtx(ctx), &req.UserSettings); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, "")
		return
	}

	response.Success(ctx)
	return
}

// GetUserSettings 获取用户配置
func GetUserSettings(ctx *gin.Context) {
	res, err := auth.GetUserSettings(ctx, jwt.GetUserIDByCtx(ctx))
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.SuccessWithData(ctx, res)
	return
}

func AuthForgetPassword(ctx *gin.Context) {
	var req rao.ForgetPasswordReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	// 先验证极验验证码
	geeTestRes := auth.FastVerify(req.Captcha)
	if geeTestRes == false {
		response.ErrorWithMsg(ctx, errno.ErrVerifyFail, "验证失败")
		return
	}

	tx := dal.GetQuery().User
	_, err := tx.WithContext(ctx).Where(tx.Mobile.Eq(req.Mobile)).First()
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrAuthFailedNotRegistered, err.Error())
		return
	}

	// 发送手机验证码
	if req.Mobile != "" {
		_ = auth.SendSmsCode(ctx, req.Mobile)
	}

	response.Success(ctx)
	return
}

func AuthResetPassword(ctx *gin.Context) {
	var req rao.ResetPasswordReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	//校验短信验证码是否正确
	verifySuc, _ := auth.VerifySmsCode(ctx, req.Mobile, req.VerifyCode)
	if verifySuc == false {
		response.ErrorWithMsg(ctx, errno.ErrSmsCodeVerifyFail, "")
		return
	}

	if req.Mobile == "" || req.NewPassword == "" || req.RepeatPassword == "" {
		response.ErrorWithMsg(ctx, errno.ErrParam, "")
		return
	}

	hashedPassword, err := omnibus.GenerateBcryptFromPassword(req.NewPassword)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	tx := dal.GetQuery().User
	if _, err := tx.WithContext(ctx).Where(tx.Mobile.Eq(req.Mobile)).UpdateColumn(tx.Password, hashedPassword); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	response.Success(ctx)
	return
}

func VerifyPassword(ctx *gin.Context) {
	var req rao.VerifyPasswordReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	tx := dal.GetQuery().User
	u, err := tx.WithContext(ctx).Where(tx.UserID.Eq(jwt.GetUserIDByCtx(ctx))).First()
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	err = omnibus.CompareBcryptHashAndPassword(u.Password, req.Password)

	response.SuccessWithData(ctx, rao.VerifyPasswordResp{IsMatch: err == nil})
	return
}

// GetSmsCode 获取短信验证码
func GetSmsCode(ctx *gin.Context) {
	var req rao.GetSmsCodeReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	// 先验证极验验证码
	geeTestRes := auth.FastVerify(req.Captcha)
	if geeTestRes == false {
		response.ErrorWithMsg(ctx, errno.ErrVerifyFail, "验证失败")
		return
	}

	//IP防刷 同一个IP同一个手机号 每小时最多可以发送20条短信验证码
	isIllegal, illegalErr := auth.IsIllegalSend(ctx, &req)
	if illegalErr != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, "")
		return
	}

	if isIllegal == true {
		response.ErrorWithMsg(ctx, errno.ErrSmsCodeSendIllegal, "")
		return
	}

	////验证码类型：1-注册验证 2-登录验证 3-找回密码验证
	//if req.Type == consts.SmsLogTypeSignUp { //注册时 判断该手机号是否已注册 已注册时则抛异常
	//	_, userInfoErr := auth.GetUserByMobile(ctx, req.Mobile)
	//	if userInfoErr == nil {
	//		response.ErrorWithMsg(ctx, errno.ErrMobileYetRegister, "")
	//		return
	//	}
	//} else { //登录和找回密码时 判断该手机号是否已经注册过 未注册则抛异常
	//	_, userInfoErr := auth.GetUserByMobile(ctx, req.Mobile)
	//	if userInfoErr != nil {
	//		response.ErrorWithMsg(ctx, errno.ErrAuthFailedNotRegistered, "")
	//		return
	//	}
	//}

	_ = auth.SendSmsCode(ctx, req.Mobile)
	response.Success(ctx)
	return
}

// VerifySmsCode 校验短信验证码
func VerifySmsCode(ctx *gin.Context) {
	var req rao.VerifySmsCodeReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	//校验短信验证码是否正确
	verifySuc, _ := auth.VerifySmsCode(ctx, req.Mobile, req.Code)
	if verifySuc == false {
		response.ErrorWithMsg(ctx, errno.ErrSmsCodeVerifyFail, "")
		return
	}

	//如果是找回密码的验证码校验 则需要返回用户ID
	if req.Type == consts.SmsLogTypeForgetPassword {
		userInfo, userErr := auth.GetUserByMobile(ctx, req.Mobile)
		if userErr != nil {
			response.ErrorWithMsg(ctx, errno.ErrParam, userErr.Error())
		}
		response.SuccessWithData(ctx, rao.VerifySmsCodeResp{
			U: userInfo.UserID,
		})
	} else {
		response.Success(ctx)
	}

	return
}

// SmsCodeLogin 手机号验证码登录
func SmsCodeLogin(ctx *gin.Context) {
	var req rao.SmsCodeLoginReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	//校验短信验证码是否正确
	verifySuc, _ := auth.VerifySmsCode(ctx, req.Mobile, req.Code)
	if verifySuc == false {
		response.ErrorWithMsg(ctx, errno.ErrSmsCodeVerifyFail, "")
		return
	}

	//根据手机号获取用户信息
	u, err := auth.SmsCodeLogin(ctx, req.Mobile)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrAuthFailed, err.Error())
		return
	}

	d := 7 * 24 * time.Hour
	if req.IsAutoLogin {
		d = 30 * 24 * time.Hour
	}

	token, exp, err := jwt.GenerateTokenByTime(u.UserID, d)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrInvalidToken, err.Error())
		return
	}

	if err := auth.UpdateLoginTime(ctx, u.UserID); err != nil {
		log.Logger.Errorf("update login time err %s", err)
	}

	availTeamID, _ := auth.GetAvailTeamID(ctx, u.UserID)
	userSettings := rao.UserSettings{
		CurrentTeamID: availTeamID,
	}
	_ = auth.SetUserSettings(ctx, u.UserID, &userSettings)

	response.SuccessWithData(ctx, rao.AuthLoginResp{
		Token:         token,
		ExpireTimeSec: exp.Unix(),
		TeamID:        availTeamID,
	})
	return
}

func CheckEmailIsRegister(ctx *gin.Context) {
	var req rao.CheckEmailIsRegisterReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	//校验短信验证码是否正确
	res, err := auth.CheckEmailIsRegister(ctx, req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}
	response.SuccessWithData(ctx, res)
	return
}

func CheckUserIsRegister(ctx *gin.Context) {
	var req rao.CheckUserIsRegisterReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	if req.Mobile == "" && req.Email == "" {
		response.ErrorWithMsg(ctx, errno.ErrParam, "")
		return
	}

	//// 先验证极验验证码
	//geeTestRes := auth.FastVerify(req.Captcha)
	//if geeTestRes == false {
	//	response.ErrorWithMsg(ctx, errno.ErrVerifyFail, "验证失败")
	//	return
	//}

	//IP防刷 同一个IP同一个手机号 每小时最多可以发送20条短信验证码
	reqTemp := rao.GetSmsCodeReq{
		Mobile:  req.Mobile,
		Captcha: req.Captcha,
	}
	isIllegal, illegalErr := auth.IsIllegalSend(ctx, &reqTemp)
	if illegalErr != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, "")
		return
	}

	if isIllegal == true {
		response.ErrorWithMsg(ctx, errno.ErrSmsCodeSendIllegal, "")
		return
	}

	if req.Mobile != "" {
		// 发送手机验证码
		_ = auth.SendSmsCode(ctx, req.Mobile)
	}

	//检查当前用户是否注册过
	res, err := auth.CheckUserIsRegister(ctx, req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}
	response.SuccessWithData(ctx, res)
	return
}

func SetUserPassword(ctx *gin.Context) {
	var req rao.SetUserPasswordReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	//校验短信验证码是否正确
	err := auth.SetUserPassword(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}
	response.Success(ctx)
	return
}

func CollectUserInfo(ctx *gin.Context) {
	var req rao.CollectUserInfoReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	//校验短信验证码是否正确
	err := auth.CollectUserInfo(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}
	response.Success(ctx)
	return
}

func GetCollectUserInfo(ctx *gin.Context) {
	var req rao.GetCollectUserInfoReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	//校验短信验证码是否正确
	res, err := auth.GetCollectUserInfo(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}
	response.SuccessWithData(ctx, res)
	return
}

func GetWechatLoginQrCode(ctx *gin.Context) {
	//校验短信验证码是否正确
	res, err := auth.GetWechatLoginQrCode()
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}
	response.SuccessWithData(ctx, res)
	return
}

func GetWechatLoginResult(ctx *gin.Context) {
	var req rao.GetWechatLoginResultReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	//校验短信验证码是否正确
	wechatLoginResult, err := auth.GetWechatLoginResult(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	res := rao.GetWechatLoginResp{
		ScanStatus: wechatLoginResult.ScanStatus,
	}

	if wechatLoginResult.IsRegister == true {
		d := 1 * 24 * time.Hour
		token, exp, err := jwt.GenerateTokenByTime(wechatLoginResult.UserID, d)
		if err != nil {
			response.ErrorWithMsg(ctx, errno.ErrInvalidToken, err.Error())
			return
		}

		if err := auth.UpdateLoginTime(ctx, wechatLoginResult.UserID); err != nil {
			log.Logger.Errorf("update login time err %s", err)
		}

		availTeamID, _ := auth.GetAvailTeamID(ctx, wechatLoginResult.UserID)
		userSettings := rao.UserSettings{
			CurrentTeamID: availTeamID,
		}
		_ = auth.SetUserSettings(ctx, wechatLoginResult.UserID, &userSettings)
		response.SuccessWithData(ctx, rao.GetWechatLoginResp{
			Token:         token,
			ExpireTimeSec: exp.Unix(),
			TeamID:        availTeamID,
			IsRegister:    true,
			ScanStatus:    wechatLoginResult.ScanStatus,
		})
		return
	}
	response.SuccessWithData(ctx, res)
	return
}

func CheckWechatIsChangeBind(ctx *gin.Context) {
	var req rao.CheckWechatIsChangeBindReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}
	//校验短信验证码是否正确
	isChangeBind, err := auth.CheckWechatIsChangeBind(ctx, &req)
	if err != nil {
		response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		return
	}

	res := rao.CheckWechatIsChangeBindResp{
		IsChangeBind: isChangeBind,
	}

	response.SuccessWithData(ctx, res)
	return
}

func UpdateAccount(ctx *gin.Context) {
	var req rao.UpdateAccountReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	tx := dal.GetQuery().User
	_, err := tx.WithContext(ctx).Where(tx.Account.Eq(req.Account)).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if _, err := tx.WithContext(ctx).Where(tx.UserID.Eq(jwt.GetUserIDByCtx(ctx))).UpdateColumn(tx.Account, req.Account); err != nil {
			response.ErrorWithMsg(ctx, errno.ErrServer, "update account failed")
			return
		}

		response.Success(ctx)
		return
	}

	response.ErrorWithMsg(ctx, errno.ErrYetAccountRegister, "update account failed")
	return
}

func UpdateEmail(ctx *gin.Context) {
	var req rao.UpdateEmailReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(ctx, errno.ErrParam, err.Error())
		return
	}

	if req.Email == "" {
		response.ErrorWithMsg(ctx, errno.ErrParam, "")
		return
	}

	err := auth.UpdateEmail(ctx, &req)
	if err != nil {
		if err.Error() == "用户邮箱已注册" {
			response.ErrorWithMsg(ctx, errno.ErrYetRegister, err.Error())
		} else {
			response.ErrorWithMsg(ctx, errno.ErrMysqlFailed, err.Error())
		}
		return
	}
	response.Success(ctx)
	return
}
