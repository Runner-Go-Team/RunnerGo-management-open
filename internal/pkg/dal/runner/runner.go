package runner

import (
	"encoding/json"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/errno"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/log"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/response"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"time"

	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/conf"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
)

const (
	EngineRunApi             = "/runner/run_api"
	EngineRunScene           = "/runner/run_scene"
	EngineStopScene          = "/runner/stop_scene"
	EngineRunPlan            = "/runner/run_plan"
	EngineStopPlan           = "/runner/stop"
	EngineRunSql             = "/runner/run_sql"
	EngineConnectionDatabase = "/runner/sql_connection"
	EngineRunTcp             = "/runner/run_tcp"
	EngineRunWebsocket       = "/runner/run_ws"
	EngineRunDubbo           = "/runner/run_dubbo"
	EngineRunMqtt            = "/runner/run_mt"
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

func RunAPI(body rao.APIDetail) (string, error) {
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
		Post(conf.Conf.Clients.Runner.EngineDomain + EngineRunApi)

	if err != nil {
		return "", err
	}

	if ret.Code != 200 {
		log.Logger.Error("发送接口请求，返回值：", ret)
		return "", fmt.Errorf("ret code not 200")
	}

	return ret.Data, nil
}

func RunTarget(body rao.RunTargetParam) (string, error) {
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
		Post(conf.Conf.Clients.Runner.EngineDomain + EngineRunApi)

	if err != nil {
		return "", err
	}

	if ret.Code != 200 {
		log.Logger.Error("发送接口请求，返回值：", ret)
		return "", fmt.Errorf("ret code not 200")
	}

	return ret.Data, nil
}

func RunScene(body *rao.SceneFlow) (string, error) {

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
		Post(conf.Conf.Clients.Runner.EngineDomain + EngineRunScene)

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
	return nil
}

func RunSceneCaseFlow(body *rao.SceneCaseFlow) (string, error) {
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
		Post(conf.Conf.Clients.Runner.EngineDomain + EngineRunScene)

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
	return nil
}

func ChangeCaseSort(ctx *gin.Context, req *rao.ChangeCaseSortReq) error {
	tx := dal.GetQuery().Target
	for _, target := range req.CaseList {
		_, err := tx.WithContext(ctx).Where(tx.TeamID.Eq(target.TeamID),
			tx.TargetID.Eq(target.CaseID)).UpdateSimple(tx.Sort.Value(target.Sort), tx.ParentID.Value(target.SceneID))
		if err != nil {
			return err
		}
	}
	return nil
}

//func RunSql(body rao.RunSqlParam) (string, error) {
//	bodyByte, err := json.Marshal(body)
//	if err != nil {
//		return "", err
//	}
//	log.Logger.Infof("body %s", bodyByte)
//
//	var ret RunAPIResp
//	_, err = resty.New().R().
//		SetHeader("Content-Type", "application/json").
//		SetBody(bodyByte).
//		SetResult(&ret).
//		Post(conf.Conf.Clients.Runner.EngineDomain + EngineRunSql)
//
//	if err != nil {
//		return "", err
//	}
//
//	if ret.Code != 200 {
//		log.Logger.Error("发送调试mysql语句请求，返回值：", ret)
//		return "", fmt.Errorf("ret code not 200")
//	}
//
//	return ret.Data, nil
//}

func RunConnectionDatabase(body rao.ConnectionDatabaseReq) (bool, error) {
	bodyByte, err := json.Marshal(body)
	if err != nil {
		return false, err
	}
	log.Logger.Infof("body %s", bodyByte)

	var ret RunAPIResp
	_, err = resty.New().R().
		SetHeader("Content-Type", "application/json").
		SetBody(bodyByte).
		SetResult(&ret).
		Post(conf.Conf.Clients.Runner.EngineDomain + EngineConnectionDatabase)

	if err != nil {
		return false, err
	}

	if ret.Code != 200 {
		log.Logger.Error("测试链接数据库失败，返回码不为200，返回值：", ret)
		return false, fmt.Errorf(ret.Data)
	}

	return true, nil
}

//func RunTcp(body rao.RunTcpParam) (string, error) {
//	bodyByte, err := json.Marshal(body)
//	if err != nil {
//		return "", err
//	}
//	log.Logger.Infof("body %s", bodyByte)
//
//	var ret RunAPIResp
//	_, err = resty.New().R().
//		SetHeader("Content-Type", "application/json").
//		SetBody(bodyByte).
//		SetResult(&ret).
//		Post(conf.Conf.Clients.Runner.EngineDomain + EngineRunTcp)
//
//	if err != nil {
//		return "", err
//	}
//
//	if ret.Code != 200 {
//		log.Logger.Error("发送调试tcp接口失败，返回值：", ret)
//		return "", fmt.Errorf("返回code码非200")
//	}
//
//	return ret.Data, nil
//}

//func RunWebsocket(body rao.RunWebsocketParam) (string, error) {
//	bodyByte, err := json.Marshal(body)
//	if err != nil {
//		return "", err
//	}
//	log.Logger.Infof("body %s", bodyByte)
//
//	var ret RunAPIResp
//	_, err = resty.New().R().
//		SetHeader("Content-Type", "application/json").
//		SetBody(bodyByte).
//		SetResult(&ret).
//		Post(conf.Conf.Clients.Runner.EngineDomain + EngineRunWebsocket)
//	if err != nil {
//		return "", err
//	}
//
//	if ret.Code != 200 {
//		log.Logger.Error("发送调试websocket接口失败，返回值：", ret)
//		return "", fmt.Errorf("返回code码非200")
//	}
//
//	return ret.Data, nil
//}

func RunDubbo(body rao.RunDubboParam) (string, error) {
	bodyByte, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	log.Logger.Infof("body %s", bodyByte)

	ret := RunAPIResp{}
	_, err = resty.New().R().
		SetHeader("Content-Type", "application/json").
		SetBody(bodyByte).
		SetResult(&ret).
		Post(conf.Conf.Clients.Runner.EngineDomain + EngineRunDubbo)

	if err != nil {
		return "", err
	}

	if ret.Code != 200 {
		log.Logger.Error("发送调试websocket接口失败，返回值：", ret)
		return "", fmt.Errorf("返回code码非200")
	}

	return ret.Data, nil
}

func RunMqtt(body rao.RunMqttParam) (string, error) {
	bodyByte, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	log.Logger.Infof("body %s", bodyByte)

	ret := RunAPIResp{}
	_, err = resty.New().R().
		SetHeader("Content-Type", "application/json").
		SetBody(bodyByte).
		SetResult(&ret).
		Post(conf.Conf.Clients.Runner.EngineDomain + EngineRunMqtt)

	if err != nil {
		return "", err
	}

	if ret.Code != 200 {
		log.Logger.Error("发送调试mqtt接口失败，返回值：", ret)
		return "", fmt.Errorf("返回code码非200")
	}

	return ret.Data, nil
}
