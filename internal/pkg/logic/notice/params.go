package notice

import (
	"encoding/json"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/conf"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/gin-gonic/gin"
	"net/url"
)

type Content struct {
	Config   Config    `json:"config"`
	Header   Header    `json:"header"`
	Elements []Element `json:"elements"`
}

type Config struct {
	WideScreenMode bool `json:"wide_screen_mode"`
}

type Title struct {
	Tag     string `json:"tag"`
	Content string `json:"content"`
}

type Header struct {
	Template string `json:"template"`
	Title    Title  `json:"title"`
}

type MultiURL struct {
	URL        string `json:"url"`
	PCURL      string `json:"pc_url"`
	AndroidURL string `json:"android_url"`
	IOSURL     string `json:"ios_url"`
}

type Text struct {
	Tag     string `json:"tag"`
	Content string `json:"content"`
}

type ActionButton struct {
	Tag      string   `json:"tag"`
	Text     Text     `json:"text"`
	Type     string   `json:"type"`
	MultiURL MultiURL `json:"multi_url"`
}

type Element struct {
	Tag     string         `json:"tag"`
	Content string         `json:"content,omitempty"`
	Actions []ActionButton `json:"actions,omitempty"`
}

// getCard 飞书发送模版
func getCard(params *rao.SendCardParams) string {
	marshal, err := json.Marshal(getCardContent(params))
	if err != nil {
		return ""
	}

	return string(marshal)
}

// getCardContent 飞书发送模版
func getCardContent(params *rao.SendCardParams) Content {
	elements := make([]Element, 0)
	actionButtons := make([]ActionButton, 0)

	if params.ReportType == consts.PlanStress {
		for _, report := range params.StressPlanReports {
			actionButton := ActionButton{
				Tag: "button",
				Text: Text{
					Tag:     "plain_text",
					Content: report.ReportName,
				},
				Type: "primary",
				MultiURL: MultiURL{
					URL: fmt.Sprintf(conf.Conf.Base.Domain + "#/email/report?report_id=" + fmt.Sprintf("%s", report.ReportID) + "&team_id=" + fmt.Sprintf("%s", params.Team.TeamID)),
				},
			}
			actionButtons = append(actionButtons, actionButton)
		}
	}

	if params.ReportType == consts.PlanAuto {
		for _, report := range params.AutoPlanReports {
			actionButton := ActionButton{
				Tag: "button",
				Text: Text{
					Tag:     "plain_text",
					Content: report.ReportName,
				},
				Type: "primary",
				MultiURL: MultiURL{
					URL: fmt.Sprintf(conf.Conf.Base.Domain + "#/email/autoReport?report_id=" + fmt.Sprintf("%s", report.ReportID) + "&team_id=" + fmt.Sprintf("%s", params.Team.TeamID)),
				},
			}
			actionButtons = append(actionButtons, actionButton)
		}
	}

	element1 := Element{
		Tag: "markdown",
		Content: fmt.Sprintf("**所在团队**：%s\n**计划名称**：%s\n**任务类型**：%s\n\n**执行者**：%s\n",
			params.Team.Name, params.PlanName, params.TaskTypeName, params.RunUserName),
	}

	element2 := Element{
		Tag: "hr",
	}

	element3 := Element{
		Tag:     "action",
		Actions: actionButtons,
	}

	elements = append(elements, element1)
	elements = append(elements, element2)
	elements = append(elements, element3)

	content := Content{
		Config: Config{
			WideScreenMode: true,
		},
		Header: Header{
			Template: "blue",
			Title: Title{
				Tag:     "plain_text",
				Content: "测试报告",
			},
		},
		Elements: elements,
	}

	return content
}

// getWechatCardContent 获取企业微信发送模版
func getWechatCardContent(params *rao.SendCardParams) WechatTemplateCard {
	jumpItems := make([]JumpItem, 0)

	url := ""
	if params.ReportType == consts.PlanStress {
		for _, report := range params.StressPlanReports {
			jumpItem := JumpItem{
				Type:  1,
				URL:   fmt.Sprintf(conf.Conf.Base.Domain + "#/email/report?report_id=" + fmt.Sprintf("%s", report.ReportID) + "&team_id=" + fmt.Sprintf("%s", params.Team.TeamID)),
				Title: report.ReportName,
			}
			url = jumpItem.URL
			jumpItems = append(jumpItems, jumpItem)
		}
	}

	if params.ReportType == consts.PlanAuto {
		for _, report := range params.AutoPlanReports {
			jumpItem := JumpItem{
				Type:  1,
				URL:   fmt.Sprintf(conf.Conf.Base.Domain + "#/email/autoReport?report_id=" + fmt.Sprintf("%s", report.ReportID) + "&team_id=" + fmt.Sprintf("%s", params.Team.TeamID)),
				Title: report.ReportName,
			}
			url = jumpItem.URL
			jumpItems = append(jumpItems, jumpItem)
		}
	}

	message := WechatTemplateCard{
		MsgType: "template_card",
		TemplateCard: CardContent{
			CardType: "text_notice",
			MainTitle: TitleDesc{
				Title: "测试报告",
				Desc:  "",
			},
			SubTitleText: fmt.Sprintf("所在团队：%s\n计划名称：%s\n任务类型：%s\n\n执行者：%s\n",
				params.Team.Name, params.PlanName, params.TaskTypeName, params.RunUserName),
			HorizontalContentList: nil,
			JumpList:              jumpItems,
			CardAction: CardAction{
				Type:     1,
				URL:      url,
				AppID:    "",
				PagePath: "",
			},
		},
	}

	return message
}

// getDingCardContent 获取钉钉发送模版
func getDingCardContent(params *rao.SendCardParams) DingActionCard {
	dingButtons := make([]DingButton, 0)

	if params.ReportType == consts.PlanStress {
		for _, report := range params.StressPlanReports {
			dingButton := DingButton{
				Title:     report.ReportName,
				ActionURL: "dingtalk://dingtalkclient/page/link?url=" + url.QueryEscape(fmt.Sprintf(conf.Conf.Base.Domain+"#/email/report?report_id="+fmt.Sprintf("%s", report.ReportID)+"&team_id="+fmt.Sprintf("%s", params.Team.TeamID))) + "&pc_slide=false",
			}
			dingButtons = append(dingButtons, dingButton)
		}
	}

	if params.ReportType == consts.PlanAuto {
		for _, report := range params.AutoPlanReports {
			dingButton := DingButton{
				Title:     report.ReportName,
				ActionURL: "dingtalk://dingtalkclient/page/link?url=" + url.QueryEscape(fmt.Sprintf(conf.Conf.Base.Domain+"#/email/autoReport?report_id="+fmt.Sprintf("%s", report.ReportID)+"&team_id="+fmt.Sprintf("%s", params.Team.TeamID))) + "&pc_slide=false",
			}
			dingButtons = append(dingButtons, dingButton)
		}
	}

	message := DingActionCard{
		MsgType: "actionCard",
		ActionCard: DingCardContent{
			Title: "测试报告",
			Text: fmt.Sprintf("**测试报告** \n\n\n\n所在团队：%s\n\n计划名称：%s\n\n任务类型：%s\n\n\n\n执行者：%s\n\n",
				params.Team.Name, params.PlanName, params.TaskTypeName, params.RunUserName),
			BtnOrientation: "0",
			Btns:           dingButtons,
		},
	}

	return message
}

type DingAppActionCard struct {
	BtnJsonList    []DingAppBtnJson `json:"btn_json_list"`
	SingleURL      string           `json:"single_url"`
	BtnOrientation string           `json:"btn_orientation"`
	SingleTitle    string           `json:"single_title"`
	Markdown       string           `json:"markdown"`
	Title          string           `json:"title"`
}

type DingAppBtnJson struct {
	ActionURL string `json:"action_url"`
	Title     string `json:"title"`
}

type DingAppMessage struct {
	ActionCard DingAppActionCard `json:"action_card"`
	MsgType    string            `json:"msgtype"`
}

// getDingAppCardContent 获取钉钉发送模版
func getDingAppCardContent(params *rao.SendCardParams) DingAppMessage {
	dingAppBtnJson := make([]DingAppBtnJson, 0)

	title := ""
	urls := ""
	if params.ReportType == consts.PlanStress {
		for _, report := range params.StressPlanReports {
			dingButton := DingAppBtnJson{
				Title:     report.ReportName,
				ActionURL: "dingtalk://dingtalkclient/page/link?url=" + url.QueryEscape(fmt.Sprintf(conf.Conf.Base.Domain+"#/email/report?report_id="+fmt.Sprintf("%s", report.ReportID)+"&team_id="+fmt.Sprintf("%s", params.Team.TeamID))) + "&pc_slide=false",
			}
			urls = dingButton.ActionURL
			title = dingButton.Title
			dingAppBtnJson = append(dingAppBtnJson, dingButton)
		}
	}

	if params.ReportType == consts.PlanAuto {
		for _, report := range params.AutoPlanReports {
			dingButton := DingAppBtnJson{
				Title:     report.ReportName,
				ActionURL: "dingtalk://dingtalkclient/page/link?url=" + url.QueryEscape(fmt.Sprintf(conf.Conf.Base.Domain+"#/email/autoReport?report_id="+fmt.Sprintf("%s", report.ReportID)+"&team_id="+fmt.Sprintf("%s", params.Team.TeamID))) + "&pc_slide=false",
			}
			urls = dingButton.ActionURL
			title = dingButton.Title
			dingAppBtnJson = append(dingAppBtnJson, dingButton)
		}
	}

	message := DingAppMessage{
		ActionCard: DingAppActionCard{
			BtnJsonList:    dingAppBtnJson,
			SingleURL:      urls,
			BtnOrientation: "0",
			SingleTitle:    title,
			Markdown: fmt.Sprintf("**测试报告** \n\n\n\n所在团队：%s\n\n计划名称：%s\n\n任务类型：%s\n\n\n\n执行者：%s\n\n",
				params.Team.Name, params.PlanName, params.TaskTypeName, params.RunUserName),
			Title: "测试报告",
		},
		MsgType: "action_card",
	}

	return message
}

// GetSendCardParamsByReq 获取发送内容所需参数
func GetSendCardParamsByReq(ctx *gin.Context, req *rao.SendNoticeParams) (*rao.SendCardParams, error) {
	tx := dal.GetQuery().Team
	team, err := tx.WithContext(ctx).Where(tx.TeamID.Eq(req.TeamID)).First()
	if err != nil {
		return nil, err
	}

	// 性能计划报告
	if req.EventID == consts.NoticeEventStressPlan || req.EventID == consts.NoticeEventStressPlanReport {
		rx := dal.GetQuery().StressPlanReport
		reports, err := rx.WithContext(ctx).Where(rx.TeamID.Eq(req.TeamID), rx.ReportID.In(req.ReportIDs...)).Find()
		if err != nil {
			return nil, err
		}

		var (
			taskTypeName string = "普通任务"
			runUserID    string
			planName     string
		)
		for _, report := range reports {
			if report.TaskType == consts.PlanTaskTypeCronjob {
				taskTypeName = "定时任务"
			}
			runUserID = report.RunUserID
			planName = report.PlanName
			break
		}

		ux := dal.GetQuery().User
		user, err := ux.WithContext(ctx).Where(ux.UserID.Eq(runUserID)).First()
		if err != nil {
			return nil, err
		}

		params := &rao.SendCardParams{
			PlanName:          planName,
			TaskTypeName:      taskTypeName,
			RunUserName:       user.Nickname,
			Team:              team,
			ReportType:        consts.PlanStress,
			StressPlanReports: reports,
		}

		return params, nil
	}

	// 自动计划报告
	if req.EventID == consts.NoticeEventAuthPlan || req.EventID == consts.NoticeEventAuthPlanReport {
		rx := dal.GetQuery().AutoPlanReport
		reports, err := rx.WithContext(ctx).Where(rx.TeamID.Eq(req.TeamID), rx.ReportID.In(req.ReportIDs...)).Find()
		if err != nil {
			return nil, err
		}

		var (
			taskTypeName string = "普通任务"
			runUserID    string
			planName     string
		)
		for _, report := range reports {
			if report.TaskType == consts.PlanTaskTypeCronjob {
				taskTypeName = "定时任务"
			}
			runUserID = report.RunUserID
			planName = report.PlanName
			break
		}

		ux := dal.GetQuery().User
		user, err := ux.WithContext(ctx).Where(ux.UserID.Eq(runUserID)).First()
		if err != nil {
			return nil, err
		}

		params := &rao.SendCardParams{
			PlanName:        planName,
			TaskTypeName:    taskTypeName,
			RunUserName:     user.Nickname,
			Team:            team,
			ReportType:      consts.PlanAuto,
			AutoPlanReports: reports,
		}

		return params, nil
	}

	return &rao.SendCardParams{}, nil
}
