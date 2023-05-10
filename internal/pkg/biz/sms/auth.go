package sms

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/conf"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi20170525 "github.com/alibabacloud-go/dysmsapi-20170525/v3/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/goccy/go-json"
)

type TemplateParam struct {
	Code string `json:"code"`
}

func createClient(accessKeyId *string, accessKeySecret *string) (_result *dysmsapi20170525.Client, _err error) {
	config := &openapi.Config{
		// 必填，您的 AccessKey ID
		AccessKeyId: accessKeyId,
		// 必填，您的 AccessKey Secret
		AccessKeySecret: accessKeySecret,
	}
	// 访问的域名
	config.Endpoint = tea.String("dysmsapi.aliyuncs.com")
	_result = &dysmsapi20170525.Client{}
	_result, _err = dysmsapi20170525.NewClient(config)
	return _result, _err
}

type SendSmsResponse struct {
	Headers    map[string]*string   `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32               `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *SendSmsResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

type SendSmsResponseBody struct {
	BizId     *string `json:"BizId,omitempty" xml:"BizId,omitempty"`
	Code      *string `json:"Code,omitempty" xml:"Code,omitempty"`
	Message   *string `json:"Message,omitempty" xml:"Message,omitempty"`
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func sendCodeHandler(mobile string, code string) (_result *SendSmsResponse, _err error) {

	smsID := conf.Conf.Sms.ID
	smsSecret := conf.Conf.Sms.Secret

	client, _err := createClient(tea.String(smsID), tea.String(smsSecret))
	if _err != nil {
		return nil, _err
	}

	templateParam := TemplateParam{
		Code: code,
	}
	templateParamJson, _ := json.Marshal(&templateParam)
	templateParamRequest := string(templateParamJson)

	sendSmsRequest := &dysmsapi20170525.SendSmsRequest{
		PhoneNumbers:  tea.String(mobile),
		SignName:      tea.String("RunnerGo"),
		TemplateCode:  tea.String("SMS_133215123"),
		TemplateParam: tea.String(templateParamRequest),
	}
	var sendSmsRst *SendSmsResponse
	tryErr := func() (_e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		// 复制代码运行请自行打印 API 的返回值
		_, _err := client.SendSmsWithOptions(sendSmsRequest, &util.RuntimeOptions{})
		//fmt.Println("sendSmsRst ==== ", sendSmsRst)

		if _err != nil {
			return _err
		}

		return nil
	}()

	//fmt.Println("sendSmsRst ==== ", sendSmsRst)

	if tryErr != nil {
		var error = &tea.SDKError{}
		if _t, ok := tryErr.(*tea.SDKError); ok {
			error = _t
		} else {
			error.Message = tea.String(tryErr.Error())
		}
		// 如有需要，请打印 error
		_, _err = util.AssertAsString(error.Message)
		if _err != nil {
			return nil, _err
		}
	}
	return sendSmsRst, _err
}

func SendCode(mobile string, code string) (_result *SendSmsResponse, _err error) {
	sendSmsRst, err := sendCodeHandler(mobile, code)
	if err != nil {
		panic(err)
	}

	return sendSmsRst, err
}
