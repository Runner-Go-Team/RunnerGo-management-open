package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/app/router"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/conf"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/crontab"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/handler"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/script"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var readConfMode int
var configFile string

func main() {
	flag.IntVar(&readConfMode, "m", 0, "读取环境变量还是读取配置文件")
	flag.StringVar(&configFile, "c", "./configs/dev.yaml", "app config file.")
	flag.Parse()

	internal.InitProjects(readConfMode, configFile)

	// 数据库初始化和数据迁移任务
	go func() {
		script.DataMigrations()
	}()

	//异步执行性能定时任务
	go func() {
		handler.TimedTaskExec()
	}()

	//异步执行自动化测试定时任务
	go func() {
		handler.AutoPlanTimedTaskExec()
	}()

	// 把压力机机器心跳信息定时写入数据库
	go func() {
		handler.MachineDataInsert()
	}()

	// 把上报心跳数据超时的压力机机器从数据库当中移除
	go func() {
		handler.DeleteLostConnectMachine()
	}()

	// 把压力机监控数据定时写入数据库
	go func() {
		handler.MachineMonitorInsert()
	}()

	// 异步写入压力机所需分区总数
	go func() {
		handler.InitTotalKafkaPartition()
	}()

	// 每天凌晨3点执行的任务
	go func() {
		crontab.DeleteOperationLogBeforeSevenDay()
		crontab.DeleteMongodbData()
	}()

	//// 性能分析
	//go func() {
	//	_, err := pyroscope.Start(
	//		pyroscope.Config{
	//			ApplicationName: "RunnerGo-management-open",
	//			ServerAddress:   "http://192.168.1.205:4040/",
	//			//Logger:          pyroscope.StandardLogger,
	//			ProfileTypes: []pyroscope.ProfileType{
	//				pyroscope.ProfileCPU,
	//				pyroscope.ProfileAllocObjects,
	//				pyroscope.ProfileAllocSpace,
	//				pyroscope.ProfileInuseObjects,
	//				pyroscope.ProfileInuseSpace,
	//			},
	//		})
	//	if err != nil {
	//		log.Logger.Errorf("性能统计设置失败")
	//	}
	//}()

	// 创建 Gin 引擎实例
	engine := gin.Default()
	router.RegisterRouter(engine)

	// 创建 HTTP 服务器
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", conf.Conf.Http.Port),
		Handler: engine,
	}

	// 启动 HTTP 服务器
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务启动失败: %v", err)
		}
	}()

	// 优雅退出
	gracefulExit(server)
}

// 释放连接池资源，优雅退出
func gracefulExit(server *http.Server) {
	// 等待中断信号，然后优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("服务关闭中...")

	// 创建一个 5 秒的上下文，用于等待请求处理完成
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 关闭服务器，等待现有的请求处理完成
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("服务关闭失败: %v", err)
	}

	log.Println("服务已经优雅的关闭~~~")
}
