package mail

import (
	"context"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"net/smtp"

	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/conf"
)

const (
	autoPlanHTMLTemplate = `<!DOCTYPE html>
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
            font-weight: 400 !important;
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
            width: 386px !important;
            /* height: 135px !important; */
            background-color: #f8f8f8 !important;
            border-radius: 15px !important;
            display: flex !important;
            flex-direction: column !important;
            align-items: center !important;
            margin-top: 36px !important;
            padding-top: 24px !important;
            box-sizing: border-box !important;
            /* overflow: hidden !important; */
        }

        .email-body > .plan-name {
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

        .report-list > button {
            background: #054BB9 !important;
            border-radius: 5px !important;
            width: 335px !important;
            height: 41px !important;
            color: #fff !important;
            margin-top: 24px !important;
            border: none !important;
        }

        .team {
            font-size: 20px !important;
            margin-top: 36px !important;
        }

        .a_text {
            color: #fff !important;
            text-decoration: none !important;
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
                <button><a class="a_text" href="%s">查看测试报告</a></button>
            </div>
        </div>
    </div>
</body>

</html>`
)

func SendAutoPlanEmail(toEmail string, planInfo *model.AutoPlan, teamInfo *model.Team, userName string, reportID string) error {
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
	header["Subject"] = fmt.Sprintf("测试报告 【%s】的【%s】给您发送了【%s】的测试报告，点击查看", teamInfo.Name, userName, planInfo.PlanName)
	header["Content-Type"] = "text/html; charset=UTF-8"

	reportDetailUrl := fmt.Sprintf(conf.Conf.Base.Domain + "#/email/autoReport?report_id=" + fmt.Sprintf("%s", reportID) + "&team_id=" + fmt.Sprintf("%s", teamInfo.TeamID))
	domainUrl := conf.Conf.Base.Domain
	body := fmt.Sprintf(autoPlanHTMLTemplate, domainUrl, teamInfo.Name, planInfo.PlanName, userName, reportDetailUrl)
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

func SendAutoPlanNoticeEmail(ctx context.Context, emailConf *rao.SMTPEmail, toEmail string, params *rao.SendCardParams) error {
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

	reportDetailUrl := ""
	for _, report := range params.AutoPlanReports {
		reportDetailUrl += fmt.Sprintf(conf.Conf.Base.Domain + "#/email/autoReport?report_id=" + fmt.Sprintf("%s", report.ReportID) + "&team_id=" + fmt.Sprintf("%s", params.Team.TeamID))
	}
	domainUrl := conf.Conf.Base.Domain
	body := fmt.Sprintf(autoPlanHTMLTemplate, domainUrl, params.Team.Name, params.PlanName, params.RunUserName, reportDetailUrl)
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
