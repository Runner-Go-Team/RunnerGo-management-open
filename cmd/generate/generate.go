package main

import (
	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"
)

func main() {
	dsn := "runnergo_open:czYNsm6LmfZ0XU3E@tcp(rm-2zem14s80lyu5c4z7.mysql.rds.aliyuncs.com:3306)/runnergo_open?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	g := gen.NewGenerator(gen.Config{
		//OutPath: "./internal/pkg/dal/query",
		OutPath: "../../internal/pkg/dal/query",
	})

	g.UseDB(db)

	g.ApplyBasic(
		g.GenerateModel("target"),
		g.GenerateModel("operation"),
		g.GenerateModel("user"),
		g.GenerateModel("team"),
		g.GenerateModel("team_env"),
		g.GenerateModel("team_env_service"),
		g.GenerateModel("user_team"),
		g.GenerateModel("setting"),
		g.GenerateModel("sms_log"),
		g.GenerateModel("report_machine"),
		g.GenerateModel("variable"),
		g.GenerateModel("variable_import"),
		g.GenerateModel("team_user_queue"),
		g.GenerateModel("machine"),
		g.GenerateModel("preinstall_conf"),
		g.GenerateModel("auto_plan"),
		g.GenerateModel("auto_plan_email"),
		g.GenerateModel("auto_plan_timed_task_conf"),
		g.GenerateModel("auto_plan_task_conf"),
		g.GenerateModel("auto_plan_task_conf"),
		g.GenerateModel("auto_plan_report"),
		g.GenerateModel("stress_plan"),
		g.GenerateModel("stress_plan_task_conf"),
		g.GenerateModel("stress_plan_report"),
		g.GenerateModel("stress_plan_email"),
		g.GenerateModel("stress_plan_timed_task_conf"),
		g.GenerateModel("target_debug_log"),
		g.GenerateModel("invoice"),
		g.GenerateModel("order"),
		g.GenerateModel("team_buy_version"),
		g.GenerateModel("vum_use_log"),
		g.GenerateModel("vum_buy_version"),
		g.GenerateModel("user_collect_info"),
	)

	g.Execute()
}
