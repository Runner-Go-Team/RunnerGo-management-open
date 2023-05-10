package mail

import (
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/encrypt"
	"math/rand"
	"net/smtp"
	"time"

	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/conf"
)

const (
	inviteHTMLTemplate = `<!DOCTYPE html>
<html lang="en">

<head>
    <meta charset="UTF-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Document</title>
    <style>
        * {
            padding: 0 !important;
            margin: 0 !important;
        }

        .email {
            width: 100vw !important;
            height: 100vh !important;
            background-color: #f2f2f2 !important;
            display: flex !important;
            flex-direction: column !important;
            align-items: center !important;
            justify-content: center !important;
        }

        .logo {
            width: 241px !important;
            height: 66px !important;
        }

        .title {
            margin-top: 30px !important;
            font-size: 30px !important;
            color: #000 !important;
            font-family: 'PingFang SC' !important;
            font-style: normal !important;
            font-weight: 600 !important;
        }

        .slogn {
            margin-top: 30px !important;
            font-family: 'PingFang SC' !important;
            font-style: normal !important;
            font-weight: 400 !important;
            font-size: 18px !important;
            color: #999999 !important;
        }

        .email-body {
            max-width: 908px !important;
            height: 178px !important;
            background-color: #f8f8f8 !important;
            border-radius: 15px !important;
            display: flex !important;
            flex-direction: column !important;
            align-items: center !important;
            padding: 0 38px !important; 
            margin-top: 77px !important;
            padding-top: 24px !important;
            box-sizing: border-box !important;
        }

        .email-body>.p1 {
            font-family: 'PingFang SC' !important;
            font-style: normal !important;
            font-weight: 600 !important;
            font-size: 16px !important;
            color: #000 !important;
            white-space: nowrap !important;
        }

        .email-body>.p2 {
            font-family: 'PingFang SC' !important;
            font-style: normal !important;
            font-weight: 400 !important;
            font-size: 14px !important;
            line-height: 20px !important;
            color: #999999 !important;
            margin: 24px 0 !important;
        }

        .email-body>button {
            background: #054BB9 !important;
            border-radius: 5px !important;
            width: 335px !important;
            height: 41px !important;
            color: #fff !important;
            border: none !important;
        }

        a {
            text-decoration: none !important;
            color: #fff !important;
        }
    </style>
</head>

<body>
    <div class="email">
        <img class="logo" src="https://apipost.oss-cn-beijing.aliyuncs.com/kunpeng/logo_black.png" alt="">
        <p class="title">全栈测试平台</p>
        <p class="slogn">为研发赋能，让测试更简单！</p>
        <div class="email-body">
            <p class="p1">您已被【%s】成功邀请加入【%s】团队</p>
            <p class="p2">点击下方登录查看团队</p>
            <button><a href="%s">接受邀请</a></button>
        </div>
    </div>
</body>

</html>`
)

func SendInviteEmail(toEmail string, inviteUserID string, userName, teamName string, teamID string, roleID int64, isRegister bool) error {
	host := conf.Conf.SMTP.Host
	port := conf.Conf.SMTP.Port
	email := conf.Conf.SMTP.Email
	password := conf.Conf.SMTP.Password
	if host == "" || port == 0 || email == "" || password == "" {
		return fmt.Errorf("请配置邮件相关环境变量")
	}

	header := make(map[string]string)
	header["From"] = "RunnerGo" + "<" + email + ">"
	header["To"] = toEmail
	header["Subject"] = fmt.Sprintf("团队邀请 您的同事【%s】邀请您加入【%s】团队", userName, teamName)
	header["Content-Type"] = "text/html; charset=UTF-8"
	path := "#/login"
	//if !isRegister {
	//	path = "#/register"
	//}

	// 把用户信息加密
	rand.Seed(time.Now().UnixNano())
	rNum := rand.Intn(1000000)
	userInfo := fmt.Sprintf("%s_%d_%s_%d", teamID, roleID, inviteUserID, rNum)
	userInfoEncryptCode := encrypt.AesEncrypt(userInfo, conf.Conf.InviteData.AesSecretKey)

	//body := fmt.Sprintf(inviteHTMLTemplate, userName, teamName, conf.Conf.Base.Domain+path+"?email="+toEmail)
	body := fmt.Sprintf(inviteHTMLTemplate, userName, teamName, conf.Conf.Base.Domain+path+"?invite_verify_code="+userInfoEncryptCode)
	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body
	auth := smtp.PlainAuth(
		"",
		email,
		password,
		host,
	)
	return SendMailUsingTLS(
		fmt.Sprintf("%s:%d", host, port),
		auth,
		email,
		[]string{toEmail},
		[]byte(message),
	)
}
