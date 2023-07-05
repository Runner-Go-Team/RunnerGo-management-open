package mail

import (
	"context"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"net/smtp"

	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/conf"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
)

const (
	planHTMLTemplate = `<!DOCTYPE html>
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
            margin-top: 20px !important;
            font-size: 30px !important;
            color: #000 !important;
            font-family: 'PingFang SC' !important;
            font-style: normal !important;
            font-weight: 600 !important;
        }

        .slogn {
            margin-top: 10px !important;
            font-family: 'PingFang SC' !important;
            font-style: normal !important;
            font-weight: 400 !important;
            font-size: 18px !important;
            color: #999999 !important;
        }

        .email-body {
            width: 908px !important;
            /* height: 135px; */
            background-color: #f8f8f8 !important;
            border-radius: 15px !important;
            display: flex !important;
            flex-direction: column !important;
            align-items: center !important;
            margin-top: 30px !important;
            padding-top: 24px !important;
            box-sizing: border-box !important;
            /* overflow: hidden; */
        }

        .email-body>.plan-name {
            font-family: 'PingFang SC' !important;
            font-style: normal !important;
            font-weight: 600 !important;
            font-size: 16px !important;
            color: #000 !important;
        }

        .report-list {
            display: flex !important;
            flex-direction: column !important;
            align-items: center !important;
            justify-content: center !important;
            position: relative !important;
            margin-top: 13px !important;
            padding-bottom: 41px !important;
        }

        .line {
            width: 817px !important;
            height: 10px !important;
            background-color: #1A1A1D !important;
            border-radius: 4.5px !important;
            position: absolute !important;
            top: 15px !important;
        }

        .list {
            display: flex !important;
            flex-direction: column !important;
            padding-bottom: 20px !important;
            margin-top: 20px !important;
            background-color: #FEFEFE !important;
            border-width: 0px 2px 2px 2px !important;
            border-style: solid !important;
            border-color: #054BB9 !important;
            width: 805px !important;
            max-height: 300px !important;
            overflow-y: scroll !important;
            box-sizing: border-box !important;
            z-index: 20 !important;
        }

        .list-item {
            display: flex !important;
            box-sizing: border-box !important;
            justify-content: space-between !important;
            padding: 10px 0 !important;
            font-size: 12px !important;
            margin: 0 26px !important;
            border-bottom: 1px solid #000 !important;
            cursor: pointer !important;
        }

        .list-item>p:nth-child(1) {
            font-size: 14px !important;
        }

        .list-item:hover {
            color: #054BB9 !important;
            border-color: #054BB9 !important;
        }

        .list-item>.handle {
            color: #4052EC;
        }


        .team {
            font-size: 20px !important;
            margin-top: 36px !important;
        }

        .a_text {
            text-decoration: none !important;
            color: #000 !important;
        }

        .to_login {
            margin-top: 20px !important;
            height: 28px !important;
            width: 82px !important;
            background: rgba(64, 82, 236, 0.1) !important;
            border: 1px solid #4052EC !important;
            border-radius: 42px !important;
            color: #4052EC !important;
            line-height: 28px;
            text-align: center;
        }

        a {
            text-decoration: none;
        }
    </style>
</head>

<body>
    <div class="email">
        <a href="https://runnergo.com" target="_blank">
            <img class="logo" src="https://apipost.oss-cn-beijing.aliyuncs.com/kunpeng/logo_black.png" alt="" />
        </a>
        <p class="title">全栈测试平台</p>
        <p class="slogn">为研发赋能，让测试更简单！</p>
        <a class="to_login" href="%s" target="_blank">去登录</a>
        <p class="team">【%s】</p>
        <div class="email-body">
            <p class="plan-name">【%s】By %s</p>
            <div class="report-list">
                <div class="line"></div>
                <div class="list">
                   %s
                </div>
            </div>
        </div>
    </div>


</body>

</html>`

	reportListHTMLTemplate = `<a class="a_text" href="%s">
                        <div class="list-item">
                            <p>【%s】</p>
                            <p>执行者: %s</p>
                            <p>%s</p>
                            <p class="handle">查看报告</p>
                        </div>
                    </a>`
)

func SendPlanEmail(toEmail string, planName, teamName, userName string, reports []*model.StressPlanReport, runUsers *model.User) error {
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
	header["Subject"] = fmt.Sprintf("测试报告 【%s】的【%s】给您发送了【%s】的测试报告，点击查看", teamName, userName, planName)
	header["Content-Type"] = "text/html; charset=UTF-8"

	var r string
	for _, report := range reports {
		r += fmt.Sprintf(reportListHTMLTemplate,
			conf.Conf.Base.Domain+"#/email/report?report_id="+fmt.Sprintf("%s", report.ReportID)+"&team_id="+fmt.Sprintf("%s", report.TeamID),
			report.SceneName, runUsers.Nickname, report.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	domainUrl := conf.Conf.Base.Domain
	body := fmt.Sprintf(planHTMLTemplate, domainUrl, teamName, planName, userName, r)
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

func SendPlanNoticeEmail(ctx context.Context, emailConf *rao.SMTPEmail, toEmail string, params *rao.SendCardParams) error {
	host := emailConf.Host
	port := emailConf.Port
	email := emailConf.Email
	password := emailConf.Password
	if host == "" || port == 0 || email == "" || password == "" {
		return fmt.Errorf("请配置邮件相关环境变量")
	}

	header := make(map[string]string)
	header["From"] = "RunnerGo" + "<" + email + ">"
	header["To"] = toEmail
	header["Subject"] = fmt.Sprintf("测试报告 【%s】的【%s】给您发送了【%s】的测试报告，点击查看", params.Team.Name, params.RunUserName, params.PlanName)
	header["Content-Type"] = "text/html; charset=UTF-8"

	reports := params.StressPlanReports
	var r string
	for _, report := range reports {
		r += fmt.Sprintf(reportListHTMLTemplate,
			conf.Conf.Base.Domain+"#/email/report?report_id="+fmt.Sprintf("%s", report.ReportID)+"&team_id="+fmt.Sprintf("%s", report.TeamID),
			report.SceneName, params.RunUserName, report.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	domainUrl := conf.Conf.Base.Domain
	body := fmt.Sprintf(planHTMLTemplate, domainUrl, params.Team.Name, params.PlanName, params.RunUserName, r)
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
