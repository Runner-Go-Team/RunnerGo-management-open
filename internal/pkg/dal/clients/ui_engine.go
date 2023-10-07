package clients

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Runner-Go-Team/RunnerGo-management-open/api/ui"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/log"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/errmsg"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"math/rand"
	"strings"
	"time"
)

// RandUiEngineMachineID 随机获取一个机器ID
func RandUiEngineMachineID() (string, error) {
	var addr string
	machineList, err := GetUiEngineMachineList()
	if err != nil {
		return "", err
	}

	if len(machineList) > 0 {
		machineInfo := machineList[rand.Intn(len(machineList))]
		addr = machineInfo.IP
	}

	return addr, nil
}

// GetUiEngineMachineList 获取机器列表
func GetUiEngineMachineList() ([]*rao.UiEngineMachineInfo, error) {
	ctx := context.Background()
	// 从Redis获取压力机列表
	machineListRes := dal.GetRDB().HGetAll(ctx, consts.UiEngineMachineListRedisKey)
	if len(machineListRes.Val()) == 0 || machineListRes.Err() != nil {
		log.Logger.Error("获取 UiEngine，err:", machineListRes.Err())
		return nil, machineListRes.Err()
	}

	uiEngineMachineList := make([]*rao.UiEngineMachineInfo, 0, len(machineListRes.Val()))
	// 有数据，则入库
	for machineAddr, machineDetail := range machineListRes.Val() {
		// 获取机器IP，端口号，区域
		machineAddrSlice := strings.Split(machineAddr, "_")
		if len(machineAddrSlice) != 2 {
			continue
		}

		// 把机器详情信息解析成格式化数据
		var runnerMachineInfo *rao.UiEngineMachineInfo
		err := json.Unmarshal([]byte(machineDetail), &runnerMachineInfo)
		if err != nil {
			log.Logger.Info("压力机数据入库--压力机详情数据解析失败，err：", err)
			continue
		}

		// 当前时间
		nowTime := time.Now().Unix()
		if int64(runnerMachineInfo.Timestamp) < nowTime-120 {
			dal.GetRDB().HDel(ctx, consts.UiEngineMachineListRedisKey, machineAddr)
			continue
		}

		runnerMachineInfo.Key = machineAddr
		ip := machineAddrSlice[0]
		port := machineAddrSlice[1]
		ipAddr := ip + ":" + port

		currentTask, _ := dal.GetRDB().SCard(ctx, consts.UIEngineCurrentRunPrefix+ipAddr).Result()
		runnerMachineInfo.CurrentTask = currentTask
		runnerMachineInfo.IP = ipAddr

		uiEngineMachineList = append(uiEngineMachineList, runnerMachineInfo)
	}

	return uiEngineMachineList, nil
}

// RunUiEngine 运行自动化
func RunUiEngine(ctx context.Context, addr string, in *ui.RunRequest) (map[string]string, error) {
	// 调用 json.Marshal 函数，将结构体转换成 JSON
	b, err := json.MarshalIndent(in, "", "  ")
	if err != nil {
		log.Logger.Errorf("in RunRequest json.Marshal err:%s", err)
	}
	log.Logger.Info("in RunRequest----", string(b))

	// 连接到server端，此处禁用安全传输
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		if status.Code(err) == codes.DeadlineExceeded {
			return nil, errmsg.ErrSendUIEngineDeadline
		}
		log.Logger.Error("conn err：", err)
		return nil, err
	}
	defer conn.Close()
	c := ui.NewUiEngineClient(conn)

	if err != nil {
		if status.Code(err) == codes.DeadlineExceeded {
			return nil, errmsg.ErrSendUIEngineDeadline
		}
		log.Logger.Error("conn err：", err)
		return nil, err
	}

	// 执行RPC调用并打印收到的响应数据
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	resp, err := c.Run(ctx, in)
	if err != nil {
		if status.Code(err) == codes.DeadlineExceeded {
			return nil, errmsg.ErrSendUIEngineDeadline
		}
		log.Logger.Error("ui RunUiEngine err：", err)
		return nil, err
	}

	log.Logger.Info("ui RunUiEngine：", resp)
	if resp.Code != 10000 {
		log.Logger.Error("ui RunUiEngine message：", resp.Message)
		return nil, errors.New(resp.Message)
	}
	log.Logger.Info("ui RunUiEngine resp.Data：", resp.Data)

	return resp.Data, nil
}
