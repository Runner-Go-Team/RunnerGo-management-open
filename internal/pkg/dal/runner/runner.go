package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"kp-management/internal/pkg/biz/consts"
	"kp-management/internal/pkg/biz/errno"
	"kp-management/internal/pkg/biz/log"
	"kp-management/internal/pkg/biz/response"
	"kp-management/internal/pkg/dal"
	"time"

	"kp-management/internal/pkg/conf"
	"kp-management/internal/pkg/dal/rao"
)

type RunAPIResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data string `json:"data"`
}

type StopRunnerReq struct {
	TeamID    string   `json:"team_id"`
	PlanID    string   `json:"plan_id"`
	ReportIds []string `json:"report_ids"`
}

func RunAPI(ctx context.Context, body rao.APIDetail) (string, error) {

	//// 检查是否有引用环境服务URL
	//var requestURLBuilder strings.Builder
	//if body.EnvServiceURL != "" {
	//	requestURLBuilder.WriteString(body.EnvServiceURL)
	//}
	//requestURLBuilder.WriteString(body.URL)
	//requestURL := requestURLBuilder.String()
	//body.URL = requestURL

	bodyByte, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	log.Logger.Infof("body %s", bodyByte)

	var ret RunAPIResp
	_, err = resty.New().R().
		SetHeader("Content-Type", "application/json").
		SetBody(bodyByte).
		SetResult(&ret).
		Post(conf.Conf.Clients.Runner.RunAPI)

	if err != nil {
		return "", err
	}

	if ret.Code != 200 {
		log.Logger.Error("发送接口请求，返回值：", ret)
		return "", fmt.Errorf("ret code not 200")
	}

	return ret.Data, nil
}

func RunScene(ctx context.Context, body *rao.SceneFlow) (string, error) {

	bodyByte, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	log.Logger.Info("body:", string(bodyByte))

	var ret RunAPIResp
	_, err = resty.New().R().
		SetHeader("Content-Type", "application/json").
		SetBody(bodyByte).
		SetResult(&ret).
		Post(conf.Conf.Clients.Runner.RunScene)

	if err != nil {
		return "", err
	}

	if ret.Code != 200 {
		return "", fmt.Errorf("ret code not 200")
	}

	return ret.Data, nil
}

func StopScene(ctx *gin.Context, req *rao.StopSceneReq) error {
	// 停止计划的时候，往redis里面写一条数据
	stopSceneKey := consts.StopScenePrefix + req.TeamID + ":" + req.SceneID
	_, err := dal.GetRDB().Set(ctx, stopSceneKey, "stop", time.Second*3600).Result()
	if err != nil {
		log.Logger.Errorf("停止场景--写入redis数据失败，err:", err)
		response.ErrorWithMsg(ctx, errno.ErrRedisFailed, err.Error())
		return err
	}

	//var ret RunAPIResp
	//_, err := resty.New().R().
	//	SetBody(req).
	//	SetResult(&ret).
	//	Post(conf.Conf.Clients.Runner.StopScene)
	//
	//if err != nil {
	//	return err
	//}
	//
	//if ret.Code != 200 {
	//	return fmt.Errorf("ret code not 200")
	//}

	return nil
}

func RunSceneCaseFlow(ctx context.Context, body *rao.SceneCaseFlow) (string, error) {
	bodyByte, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	log.Logger.Infof("body %s", bodyByte)

	var ret RunAPIResp
	_, err = resty.New().R().
		SetHeader("Content-Type", "application/json").
		SetBody(bodyByte).
		SetResult(&ret).
		Post(conf.Conf.Clients.Runner.RunScene)

	if err != nil {
		return "", err
	}

	if ret.Code != 200 {
		return "", fmt.Errorf("ret code not 200")
	}

	return ret.Data, nil
}

func StopSceneCase(ctx *gin.Context, req *rao.StopSceneCaseReq) error {
	// 停止计划的时候，往redis里面写一条数据
	stopSceneCaseKey := consts.StopScenePrefix + req.TeamID + ":" + req.SceneID + ":" + req.SceneCaseID
	_, err := dal.GetRDB().Set(ctx, stopSceneCaseKey, "stop", time.Second*3600).Result()
	if err != nil {
		log.Logger.Errorf("停止场景用例--写入redis数据失败，err:", err)
		response.ErrorWithMsg(ctx, errno.ErrRedisFailed, err.Error())
		return err
	}

	//var ret RunAPIResp
	//_, err := resty.New().R().
	//	SetBody(req).
	//	SetResult(&ret).
	//	Post(conf.Conf.Clients.Runner.StopScene)
	//
	//if err != nil {
	//	return err
	//}
	//
	//if ret.Code != 200 {
	//	return fmt.Errorf("ret code not 200")
	//}

	return nil
}
