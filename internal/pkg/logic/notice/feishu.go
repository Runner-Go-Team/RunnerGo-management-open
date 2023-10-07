package notice

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/log"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/public"
	"github.com/go-resty/resty/v2"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"net/http"
	"strconv"
	"time"
)

type MessageSecret struct {
	MsgType   string      `json:"msg_type"`
	Card      interface{} `json:"card"`
	Timestamp string      `json:"timestamp,omitempty"`
	Sign      string      `json:"sign,omitempty"`
}

type SendFeiShuBotResp struct {
	Code int `json:"code"`
	Data struct {
	} `json:"data"`
	Msg string `json:"msg"`
}

func SendFeiShuBot(ctx context.Context, robot *rao.FeiShuRobot, card string) error {
	webhookURL := robot.WebhookURL
	secret := robot.Secret

	message := MessageSecret{
		MsgType: larkim.MsgTypeInteractive,
		Card:    card,
	}
	// 计算签名值
	if len(secret) > 0 {
		timestamp := time.Now().Unix()
		signature, err := genSign(secret, timestamp)
		if err != nil {
			return err
		}
		message.Sign = signature
		message.Timestamp = strconv.FormatInt(timestamp, 10)
	}

	client := resty.New()
	resp, err := client.R().
		SetBody(message).
		Post(webhookURL)

	var sendFeiShuBotResp *SendFeiShuBotResp
	if err := json.Unmarshal(resp.Body(), &sendFeiShuBotResp); err != nil {
		return errors.New("飞书 Hook 地址" + webhookURL + "err :" + string(resp.Body()))
	}

	if sendFeiShuBotResp.Code != 0 {
		return errors.New("飞书 Hook 地址" + webhookURL + "err :" + sendFeiShuBotResp.Msg)
	}

	log.Logger.Info("third -- SendFeiShuBot -- resp：", resp)

	if err != nil {
		return err
	}

	return nil
}

func genSign(secret string, timestamp int64) (string, error) {
	//timestamp + key 做sha256, 再进行base64 encode
	stringToSign := fmt.Sprintf("%v", timestamp) + "\n" + secret
	var data []byte
	h := hmac.New(sha256.New, []byte(stringToSign))
	_, err := h.Write(data)
	if err != nil {
		return "", err
	}
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return signature, nil
}

func SendFeiShuApp(ctx context.Context, app *rao.FeiShuApp, openIDs []string, message Content) error {
	openIDs = public.SliceUnique(openIDs)
	// 创建 Client
	client := lark.NewClient(app.AppID, app.AppSecret)
	// 构建body
	body := map[string]interface{}{}
	body["msg_type"] = larkim.MsgTypeInteractive
	body["card"] = message
	body["open_ids"] = openIDs

	// 发起请求
	resp, err := client.Do(ctx,
		&larkcore.ApiReq{
			HttpMethod:                http.MethodPost,
			ApiPath:                   "/open-apis/message/v4/batch_send",
			Body:                      body,
			SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant},
		},
	)

	log.Logger.Info("third -- SendFeiShuApp -- resp：", resp)

	if resp.StatusCode != 200 {
		return errors.New("request failed")
	}

	// 错误处理
	if err != nil {
		return err
	}

	return nil
}
