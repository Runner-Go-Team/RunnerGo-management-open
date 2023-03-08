package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"kp-management/internal"
	"kp-management/internal/app/router"
	"kp-management/internal/pkg/conf"
	"kp-management/internal/pkg/crontab"
	"kp-management/internal/pkg/handler"
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

	if err := r.Run(fmt.Sprintf(":%d", conf.Conf.Http.Port)); err != nil {
		panic(err)
	}
}
