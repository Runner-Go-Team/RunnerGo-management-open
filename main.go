package main

import (
	"flag"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/app/router"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/conf"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/crontab"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/handler"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/script"
	"github.com/gin-gonic/gin"
)

var readConfMode int
var configFile string

func main() {
	flag.IntVar(&readConfMode, "m", 0, "读取环境变量还是读取配置文件")
	flag.StringVar(&configFile, "c", "./configs/dev.yaml", "app config file.")
	flag.Parse()

	internal.InitProjects(readConfMode, configFile)

	r := gin.New()
	router.RegisterRouter(r)

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

	// 数据库迁移任务】
	go func() {
		script.DataMigrations()
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

	if err := r.Run(fmt.Sprintf(":%d", conf.Conf.Http.Port)); err != nil {
		panic(err)
	}
}
