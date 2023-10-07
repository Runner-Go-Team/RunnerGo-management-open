package clients

import (
	"context"
	"github.com/Runner-Go-Team/RunnerGo-management-open/api/v1alpha1"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/log"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/conf"
	"google.golang.org/grpc"
	"time"
)

func SaveMockAPI(api *v1alpha1.MockAPI) error {
	addr := conf.Conf.Clients.Mock.ApiManager.GrpcDomain
	// 连接到server端，此处禁用安全传输
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Logger.Error("conn err：", err)
		return err
	}
	defer conn.Close()
	c := v1alpha1.NewMockClient(conn)

	if err != nil {
		log.Logger.Error("conn err：", err)
		return err
	}

	// 执行RPC调用并打印收到的响应数据
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err = c.SaveMockAPI(ctx, &v1alpha1.SaveMockAPIRequest{Data: api})
	if err != nil {
		log.Logger.Error("mock SaveMockAPI err：", err)
		return err
	}

	return nil
}

func DeleteMockAPI(api *v1alpha1.MockAPI) error {
	addr := conf.Conf.Clients.Mock.ApiManager.GrpcDomain
	// 连接到server端，此处禁用安全传输
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Logger.Error("conn err：", err)
		return err
	}
	defer conn.Close()
	c := v1alpha1.NewMockClient(conn)

	if err != nil {
		log.Logger.Error("conn err：", err)
		return err
	}

	// 执行RPC调用并打印收到的响应数据
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err = c.DeleteMockAPI(ctx, &v1alpha1.DeleteMockAPIRequest{
		UniqueKey: api.UniqueKey,
	})
	if err != nil {
		log.Logger.Error("mock SaveMockAPI err：", err)
		return err
	}

	return nil
}
