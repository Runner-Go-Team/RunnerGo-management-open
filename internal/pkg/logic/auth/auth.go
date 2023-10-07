package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/jwt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/log"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/sms"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/uuid"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/conf"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/team"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/tools"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-omnibus/omnibus"

	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/query"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
)

func SignUp(ctx context.Context, email, password string) (*model.User, error) {
	hashedPassword, err := omnibus.GenerateBcryptFromPassword(password)
	if err != nil {
		return nil, err
	}

	// 截取邮箱前缀，用作用户昵称
	emailArrTemp := strings.Split(email, "@")
	emailArr := []rune(emailArrTemp[0])
	nickName := string(emailArr)
	if len(emailArr) > 26 {
		nickName = string(emailArr[0:26])
	}

	rand.Seed(time.Now().UnixNano())
	user := model.User{
		UserID:   uuid.GetUUID(),
		Email:    email,
		Password: hashedPassword,
		Nickname: nickName,
		Avatar:   consts.DefaultAvatarMemo[rand.Intn(3)],
	}

	teamInfo := model.Team{
		TeamID: uuid.GetUUID(),
		Name:   "默认团队",
		Type:   consts.TeamTypePrivate,
	}

	err = query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		if err := tx.User.WithContext(ctx).Create(&user); err != nil {
			return err
		}

		teamInfo.CreatedUserID = user.UserID
		if err := tx.Team.WithContext(ctx).Create(&teamInfo); err != nil {
			return err
		}

		err := tx.UserTeam.WithContext(ctx).Create(&model.UserTeam{
			UserID: user.UserID,
			TeamID: teamInfo.TeamID,
			RoleID: consts.RoleTypeOwner,
		})

		setTable := dal.GetQuery().Setting
		setInsertData := &model.Setting{
			UserID: user.UserID,
			TeamID: teamInfo.TeamID,
		}
		err = setTable.WithContext(ctx).Create(setInsertData)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func Login(ctx *gin.Context, req rao.AuthLoginReq) (*model.User, error) {
	tx := query.Use(dal.DB()).User
	user, err := tx.WithContext(ctx).Where(tx.Email.Eq(req.Email)).First()
	if err != nil {
		return nil, fmt.Errorf("账号未注册")
	}

	if err := omnibus.CompareBcryptHashAndPassword(user.Password, req.Password); err != nil {
		return nil, fmt.Errorf("密码校验失败")
	}

	if req.InviteVerifyCode != "" {
		err := team.InviteLogin(ctx, req.InviteVerifyCode, user.UserID)
		if err != nil {
			return nil, err
		}
	}
	return user, nil
}

// FastVerify 极验验证方法
func FastVerify(req rao.Captcha) bool {

	URL := conf.Conf.GeeTest.ApiServer + "/validate" + "?captcha_id=" + req.CaptchaId
	// 生成签名
	// Generate signature
	// 生成签名使用标准的hmac算法，使用用户当前完成验证的流水号lot_number作为原始消息message，使用客户验证私钥作为key
	// use standard hmac algorithms to generate signatures, and take the user's current verification serial number lot_number as the original message, and the client's verification private key as the key
	// 采用sha256散列算法将message和key进行单向散列生成最终的 “sign_token” 签名
	// use sha256 hash algorithm to hash message and key in one direction to generate the final signature
	//sign_token := hmac_encode(CAPTCHA_KEY, lot_number)
	signToken := HmacEncode(conf.Conf.GeeTest.CaptchaKey, req.LotNumber)

	// 向极验转发前端数据 + “sign_token” 签名
	// send front end parameter + "sign_token" signature to geetest
	formData := make(url.Values)
	formData["lot_number"] = []string{req.LotNumber}
	formData["captcha_output"] = []string{req.CaptchaOutput}
	formData["pass_token"] = []string{req.PassToken}
	formData["gen_time"] = []string{req.GenTime}
	formData["sign_token"] = []string{signToken}

	// 发起post请求
	// initialize a post request
	// 设置5s超时
	// set a 5 seconds timeout
	cli := http.Client{Timeout: time.Second * 5}
	resp, err := cli.PostForm(URL, formData)
	if err != nil || resp.StatusCode != 200 {
		// 当请求发生异常时，应放行通过，以免阻塞业务。
		// when geetest server interface exceptions occur, the request should pass in order not to interrupt the website's business
		fmt.Println("服务接口异常: ")
		fmt.Println(err)
		return false
	}

	resJson, _ := ioutil.ReadAll(resp.Body)
	var resMap map[string]interface{}
	// 根据极验返回的用户验证状态, 网站主进行自己的业务逻辑
	// taking the user authentication status returned from geetest into consideration, the website owner follows his own business logic
	// 响应json数据如：{"result": "success", "reason": "", "captcha_args": {}}
	// respond to json data, such as {"result": "success", "reason": "", "captcha_args": {}}

	if err = json.Unmarshal(resJson, &resMap); err != nil {
		fmt.Println("Json数据解析错误")
		return false
	}

	result := resMap["result"]
	if result == "success" {
		fmt.Println("验证通过")
	} else {
		fmt.Print("验证失败: ")
		return false
	}
	return true
}

// hmac-sha256 加密：  CAPTCHA_KEY,lot_number
// hmac-sha256 encrypt: CAPTCHA_KEY, lot_number
func HmacEncode(key string, data string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

func MobileLogin(ctx *gin.Context, mobile, password, verifyCode string) (*model.User, error) {
	tx := query.Use(dal.DB()).User
	user, err := tx.WithContext(ctx).Where(tx.Mobile.Eq(mobile)).First()
	if err != nil {
		return nil, err
	}

	if err := omnibus.CompareBcryptHashAndPassword(user.Password, password); err != nil {
		return nil, err
	}

	if verifyCode != "" {
		err := team.InviteLogin(ctx, verifyCode, user.UserID)
		if err != nil {
			return nil, err
		}
	}
	return user, nil
}

func GetUserByMobile(ctx *gin.Context, mobile string) (*model.User, error) {
	tx := query.Use(dal.DB()).User
	user, err := tx.WithContext(ctx).Where(tx.Mobile.Eq(mobile)).First()
	if err != nil {
		return nil, err
	}

	return user, nil
}

func SmsCodeLogin(ctx *gin.Context, mobile string) (*model.User, error) {
	tx := query.Use(dal.DB()).User
	user, err := tx.WithContext(ctx).Where(tx.Mobile.Eq(mobile)).First()
	if err != nil {
		return nil, err
	}

	return user, nil
}

func UpdateLoginTime(ctx context.Context, userID string) error {
	tx := query.Use(dal.DB()).User
	_, err := tx.WithContext(ctx).Where(tx.UserID.Eq(userID)).UpdateColumn(tx.LastLoginAt, time.Now())
	return err
}

func IsIllegalSend(ctx context.Context, req *rao.GetSmsCodeReq) (bool, error) {

	//IP防刷 同一个IP同一个手机号 每小时最多可以发送20条短信验证码
	redisKey := fmt.Sprintf("sendSmsCode:%s-ip:%s", req.Mobile, tools.GetLocalIP())
	theSameSendCountStr, err := dal.GetRDB().Get(ctx, redisKey).Result()
	if err != nil {
		setRedisErr := dal.GetRDB().SetNX(ctx, redisKey, int8(1), time.Minute*5).Err()
		if setRedisErr != nil {
			return false, setRedisErr
		}
		return false, nil
	}

	theSameSendCount, _ := strconv.Atoi(theSameSendCountStr)

	if theSameSendCount > consts.SmsPerHourSendCountLimit {
		return true, nil
	}

	_, setRedisErr := dal.GetRDB().Incr(ctx, redisKey).Result()
	if setRedisErr != nil {
		return false, setRedisErr
	}

	return false, nil
}

func SendSmsCode(ctx *gin.Context, mobile string) error {

	////验证码类型：1-注册验证 2-登录验证 3-找回密码验证 如果是注册验证 则先判断手机号是否已注册
	//userTable := query.Use(dal.DB()).User
	//_, userErr := userTable.WithContext(ctx).Where(userTable.Mobile.Eq(mobile)).First()
	//
	//if smsType == consts.SmsLogTypeSignUp { //注册验证
	//	if userErr != nil {
	//		return userErr
	//	}
	//} else { //登录验证或者找回密码验证
	//	if userErr == nil {
	//		return userErr
	//	}
	//}

	code := fmt.Sprintf("%04v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(10000))

	sendResult, sendErr := sms.SendCode(mobile, code)
	if sendErr != nil {
		return sendErr
	}

	redisKey := fmt.Sprintf("SmsCode-Mobile:%s", mobile)
	dal.GetRDB().Del(ctx, redisKey).Result()
	setRedisErr := dal.GetRDB().SetNX(ctx, redisKey, code, time.Minute*5).Err()
	if setRedisErr != nil {
		return setRedisErr
	}

	slTable := query.Use(dal.DB()).SmsLog

	currentTime := time.Now()
	trialTimeAtDuration, _ := time.ParseDuration(fmt.Sprintf("+%dm", 5))
	VerifyCodeExpirationTimeUnix := currentTime.Add(trialTimeAtDuration)

	sendResponseJson, _ := json.Marshal(sendResult)
	sendResponseString := string(sendResponseJson)

	smsLog := &model.SmsLog{
		Mobile:                   mobile,
		VerifyCode:               code,
		VerifyCodeExpirationTime: VerifyCodeExpirationTimeUnix,
		ClientIP:                 tools.GetLocalIP(),
		SendResponse:             sendResponseString,
	}

	if err := slTable.WithContext(ctx).Create(smsLog); err != nil {
		return err
	}

	return nil
}

func VerifySmsCode(ctx *gin.Context, mobile string, code string) (bool, error) {

	redisKey := fmt.Sprintf("SmsCode-Mobile:%s", mobile)
	smsCode, err := dal.GetRDB().Get(ctx, redisKey).Result()
	if err != nil {
		return false, err
	}

	if code == smsCode {
		dal.GetRDB().Del(ctx, redisKey).Result()
		slTable := query.Use(dal.DB()).SmsLog
		updateData := make(map[string]interface{}, 1)
		updateData["verify_status"] = consts.SmsLogVerifyStatusFinish
		slTable.WithContext(ctx).Where(slTable.Mobile.Eq(mobile)).Where(slTable.VerifyCode.Eq(code)).Where(slTable.VerifyStatus.Eq(consts.SmsLogVerifyStatusWait)).Updates(updateData)
		return true, nil
	}

	return false, nil
}

func CheckEmailIsRegister(ctx *gin.Context, req rao.CheckEmailIsRegisterReq) (rao.CheckEmailIsRegisterResp, error) {
	// 检查当前邮箱是否注册
	tx := dal.GetQuery().User
	_, err := tx.WithContext(ctx).Select(tx.ID).Where(tx.Email.Eq(req.Email)).First()
	if err != nil {
		return rao.CheckEmailIsRegisterResp{RegisterStatus: false}, nil
	}
	return rao.CheckEmailIsRegisterResp{RegisterStatus: true}, nil
}

func CheckUserIsRegister(ctx *gin.Context, req rao.CheckUserIsRegisterReq) (rao.CheckUserIsRegisterResp, error) {
	// 检查当前邮箱是否注册
	tx := dal.GetQuery().User
	var err error
	if req.Mobile != "" {
		_, err = tx.WithContext(ctx).Select(tx.ID).Where(tx.Mobile.Eq(req.Mobile)).First()
	} else {
		_, err = tx.WithContext(ctx).Select(tx.ID).Where(tx.Email.Eq(req.Email)).First()
	}
	if err != nil {
		return rao.CheckUserIsRegisterResp{RegisterStatus: false}, nil
	}
	return rao.CheckUserIsRegisterResp{RegisterStatus: true}, nil
}

func SetUserPassword(ctx *gin.Context, req *rao.SetUserPasswordReq) error {
	hashedPassword, err := omnibus.GenerateBcryptFromPassword(req.Password)
	if err != nil {
		return err
	}

	// 设置密码
	tx := dal.GetQuery().User
	_, err = tx.WithContext(ctx).Where(tx.Mobile.Eq(req.Mobile)).UpdateSimple(tx.Password.Value(hashedPassword))
	if err != nil {
		return err
	}
	return nil
}

func CollectUserInfo(ctx *gin.Context, req *rao.CollectUserInfoReq) error {
	userID := jwt.GetUserIDByCtx(ctx)

	allErr := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		// 修改用户昵称
		if req.Nickname != "" {
			_, err := tx.User.WithContext(ctx).Where(tx.User.UserID.Eq(userID)).UpdateSimple(tx.User.Nickname.Value(req.Nickname))
			if err != nil {
				return err
			}
		}

		// 修改团队名称
		if req.TeamName != "" {
			_, err := tx.Team.WithContext(ctx).Where(tx.Team.TeamID.Eq(req.TeamID)).UpdateSimple(tx.Team.Name.Value(req.TeamName))
			if err != nil {
				return err
			}
		}

		// 收集用户信息
		insertData := &model.UserCollectInfo{
			UserID:   userID,
			Industry: req.Industry,
			TeamSize: req.TeamSize,
			WorkType: req.WorkType,
		}

		err := tx.UserCollectInfo.WithContext(ctx).Create(insertData)
		if err != nil {
			return err
		}
		return nil
	})
	return allErr
}

func GetCollectUserInfo(ctx *gin.Context, req *rao.GetCollectUserInfoReq) (*rao.GetCollectUserInfoResp, error) {
	userID := jwt.GetUserIDByCtx(ctx)
	// 查询用户昵称
	tx := dal.GetQuery()
	userInfo, err := tx.User.WithContext(ctx).Where(tx.User.UserID.Eq(userID)).First()
	if err != nil {
		return nil, err
	}

	// 查询团队名称
	teamInfo, err := tx.Team.WithContext(ctx).Where(tx.Team.TeamID.Eq(req.TeamID)).First()
	if err != nil {
		return nil, err
	}

	// 查询是否存在用户收集信息数据
	isCollectUserInfo := true
	_, err = tx.UserCollectInfo.WithContext(ctx).Where(tx.UserCollectInfo.UserID.Eq(userID)).First()
	if err == nil {
		isCollectUserInfo = false
	}

	res := &rao.GetCollectUserInfoResp{
		Nickname:          userInfo.Nickname,
		TeamName:          teamInfo.Name,
		IsCollectUserInfo: isCollectUserInfo,
	}
	return res, nil
}

func GetWechatLoginQrCode() (*rao.GetWechatLoginQrCodeResp, error) {
	// 获取微信登录二维码

	log.Logger.Info("获取微信登录二维码--请求接口参数：", consts.WechatQrCodeExpiresTime)
	rawResponse, err := resty.New().R().Get(conf.Conf.WechatLogin.WechatLoginQrCodeApi + "?expire_seconds=" + fmt.Sprintf("%d", consts.WechatQrCodeExpiresTime))
	if err != nil {
		log.Logger.Info("获取微信登录二维码--请求接口失败：err：", err)
		return nil, err
	}

	var getWechatLoginQrCodeResult rao.GetWechatLoginQrCodeResult
	err = json.Unmarshal(rawResponse.Body(), &getWechatLoginQrCodeResult)
	if err != nil {
		log.Logger.Info("获取微信登录二维码--请求二维码接口返回值解析失败：err：", err)
		return nil, err
	}

	log.Logger.Info("获取微信登录二维码--接口返回结果：resp：", getWechatLoginQrCodeResult)
	if getWechatLoginQrCodeResult.Code != 10000 {
		log.Logger.Info("获取微信登录二维码--请求接口返回结果不对", getWechatLoginQrCodeResult)
		return nil, fmt.Errorf("请求获取微信登录二维码失败")
	}
	res := &rao.GetWechatLoginQrCodeResp{
		WechatLoginUrl: getWechatLoginQrCodeResult.Data.Url,
		Ticket:         getWechatLoginQrCodeResult.Data.Ticket,
	}
	return res, nil
}

func GetWechatLoginResult(ctx *gin.Context, req *rao.GetWechatLoginResultReq) (*rao.GetWechatLoginResultResp, error) {
	// 获取微信登录结果
	log.Logger.Info("获取微信登录结果--请求接口参数：", req.Ticket)
	rawResponse, err := resty.New().R().Get(conf.Conf.WechatLogin.WechatScanResultApi + "?ticket=" + req.Ticket)
	if err != nil {
		log.Logger.Info("获取微信登录结果--请求接口失败：err：", err)
		return nil, err
	}

	var getWechatLoginResult rao.GetWechatLoginResult
	err = json.Unmarshal(rawResponse.Body(), &getWechatLoginResult)
	if err != nil {
		log.Logger.Info("获取微信登录结果--请求二维码接口返回值解析失败：err：", err)
		return nil, err
	}

	scanStatus := 1
	log.Logger.Info("获取微信登录结果--接口返回结果：resp：", getWechatLoginResult)
	if getWechatLoginResult.Code == 11003 {
		log.Logger.Info("获取微信登录结果--尚未扫码", getWechatLoginResult)
		scanStatus = 2
	} else if getWechatLoginResult.Code == 11002 {
		log.Logger.Info("获取微信登录结果--二维码过期", getWechatLoginResult)
		scanStatus = 3
	}

	isRegister := false

	// 判断当前微信是否绑定过
	tx := dal.GetQuery().User
	userID := ""
	userInfo, err := tx.WithContext(ctx).Where(tx.WechatOpenID.Eq(getWechatLoginResult.Data.OpenId)).First()
	if err == nil { // 绑定过
		isRegister = true
		userID = userInfo.UserID
	}

	res := &rao.GetWechatLoginResultResp{
		ScanStatus: scanStatus,
		Openid:     getWechatLoginResult.Data.OpenId,
		IsRegister: isRegister,
		UserID:     userID,
	}
	return res, nil
}

func CheckWechatIsChangeBind(ctx *gin.Context, req *rao.CheckWechatIsChangeBindReq) (bool, error) {
	// 查询当前手机号是否注册过
	tx := dal.GetQuery().User
	userInfo, err := tx.WithContext(ctx).Where(tx.Mobile.Eq(req.Mobile)).First()
	if err != nil {
		return false, nil
	}

	if userInfo.WechatOpenID == "" {
		return false, err
	}

	return true, err
}

func UpdateEmail(ctx *gin.Context, req *rao.UpdateEmailReq) error {
	userID := jwt.GetUserIDByCtx(ctx)
	allErr := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		// 查询当前邮箱
		_, err := tx.User.WithContext(ctx).Where(tx.User.UserID.Neq(userID), tx.User.Email.Eq(req.Email)).First()
		if err == nil {
			return fmt.Errorf("用户邮箱已注册")
		}

		_, err = tx.User.WithContext(ctx).Where(tx.User.UserID.Eq(userID)).UpdateColumn(tx.User.Email, req.Email)
		if err != nil {
			return err
		}
		return nil
	})
	return allErr
}
