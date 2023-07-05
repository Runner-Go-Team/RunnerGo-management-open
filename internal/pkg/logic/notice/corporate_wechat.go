package notice

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/log"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/go-resty/resty/v2"
)

type WechatTemplateCard struct {
	MsgType      string      `json:"msgtype"`
	TemplateCard CardContent `json:"template_card"`
}

type CardContent struct {
	CardType              string     `json:"card_type"`
	MainTitle             TitleDesc  `json:"main_title"`
	QuoteArea             QuoteArea  `json:"quote_area"`
	SubTitleText          string     `json:"sub_title_text"`
	HorizontalContentList []KeyValue `json:"horizontal_content_list"`
	JumpList              []JumpItem `json:"jump_list"`
	CardAction            CardAction `json:"card_action"`
}

type TitleDesc struct {
	Title string `json:"title"`
	Desc  string `json:"desc"`
}

type QuoteArea struct {
	Type      int    `json:"type"`
	URL       string `json:"url"`
	AppID     string `json:"appid"`
	PagePath  string `json:"pagepath"`
	Title     string `json:"title"`
	QuoteText string `json:"quote_text"`
}

type KeyValue struct {
	KeyName string `json:"keyname"`
	Value   string `json:"value"`
	Type    int    `json:"type,omitempty"`
	URL     string `json:"url,omitempty"`
	MediaID string `json:"media_id,omitempty"`
}

type JumpItem struct {
	Type     int    `json:"type"`
	URL      string `json:"url,omitempty"`
	AppID    string `json:"appid,omitempty"`
	PagePath string `json:"pagepath,omitempty"`
	Title    string `json:"title"`
}

type CardAction struct {
	Type     int    `json:"type"`
	URL      string `json:"url,omitempty"`
	AppID    string `json:"appid,omitempty"`
	PagePath string `json:"pagepath,omitempty"`
}

type SendWechatBotResp struct {
	Code int `json:"code"`
	Data struct {
	} `json:"data"`
	Msg string `json:"msg"`
}

func SendWechatBot(ctx context.Context, robot *rao.WechatRobot, card WechatTemplateCard) error {
	webhookUrl := robot.WebhookURL

	jsonBytes, _ := json.Marshal(card)
	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(jsonBytes).
		Post(webhookUrl)
	log.Logger.Info("third -- SendWechatBot -- resp：", resp)

	if err != nil {
		return err
	}

	var sendWechatBotResp *SendWechatBotResp
	if err := json.Unmarshal(resp.Body(), &sendWechatBotResp); err != nil {
		return errors.New("飞书 Hook 地址" + webhookUrl + string(resp.Body()))
	}

	if sendWechatBotResp.Code != 0 {
		return errors.New(sendWechatBotResp.Msg)
	}

	defer resp.RawResponse.Body.Close()

	return nil
}
