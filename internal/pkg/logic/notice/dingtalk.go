package notice

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/log"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/public"
	"github.com/go-resty/resty/v2"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	baseURL = "https://oapi.dingtalk.com"
)

type AccessToken struct {
	AccessToken string `json:"access_token"`
}

type DingActionCard struct {
	MsgType    string          `json:"msgtype"`
	ActionCard DingCardContent `json:"actionCard"`
}

type DingCardContent struct {
	Title          string       `json:"title"`
	Text           string       `json:"text"`
	BtnOrientation string       `json:"btnOrientation"`
	Btns           []DingButton `json:"btns"`
}

type DingButton struct {
	Title     string `json:"title"`
	ActionURL string `json:"actionURL"`
}

func SendDingTalkBot(ctx context.Context, robot *rao.DingTalkRobot, card DingActionCard) error {
	webhookUrl := robot.WebhookURL
	secret := robot.Secret

	if len(secret) > 0 {
		timestamp := time.Now().UnixMilli()
		signature, err := genDingSign(secret, timestamp)
		if err != nil {
			return err
		}
		webhookUrl = webhookUrl + fmt.Sprintf("&timestamp=%d&sign=%s", timestamp, signature)
	}

	jsonBytes, _ := json.Marshal(card)
	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(jsonBytes).
		Post(webhookUrl)
	log.Logger.Info("third -- SendDingBot -- resp：", resp)

	if err != nil {
		return err
	}

	defer resp.RawResponse.Body.Close()

	return nil
}

func genDingSign(secret string, timestamp int64) (string, error) {
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)
	hash := hmac.New(sha256.New, []byte(secret))
	hash.Write([]byte(stringToSign))
	signData := hash.Sum(nil)
	sign := url.QueryEscape(base64.StdEncoding.EncodeToString(signData))

	return sign, nil
}

func getAccessToken(appKey, appSecret string) (string, error) {
	client := resty.New()

	resp, err := client.R().
		SetQueryParams(map[string]string{
			"appkey":    appKey,
			"appsecret": appSecret,
		}).
		Get(baseURL + "/gettoken")

	if err != nil {
		return "", err
	}

	var accessToken AccessToken
	if err := json.Unmarshal(resp.Body(), &accessToken); err != nil {
		return "", err
	}

	return accessToken.AccessToken, nil
}

type SendDingTalkAppParams struct {
	AgentID    int64       `json:"agent_id"`
	UseridList string      `json:"userid_list"`
	Msg        interface{} `json:"msg"`
}

func SendDingTalkApp(ctx context.Context, app *rao.DingTalkApp, openIDs []string, card DingAppMessage) error {
	openIDs = public.SliceUnique(openIDs)

	accessToken, err := getAccessToken(app.AppKey, app.AppSecret)
	if err != nil {
		return err
	}

	client := resty.New()

	agentID, err := strconv.ParseInt(app.AgentId, 10, 64)
	if err != nil {
		return err
	}

	params := SendDingTalkAppParams{
		AgentID:    agentID,
		UseridList: strings.Join(openIDs, ", "),
		Msg:        card,
	}
	bodyByte, err := json.Marshal(params)
	if err != nil {
		return err
	}

	resp, err := client.R().
		SetQueryParams(map[string]string{
			"access_token": accessToken,
		}).
		SetBody(bodyByte).
		Post("https://oapi.dingtalk.com/topapi/message/corpconversation/asyncsend_v2")

	log.Logger.Info("third -- SendDingTalkApp -- resp：", resp)
	if err != nil {
		return err
	}

	return nil
}
