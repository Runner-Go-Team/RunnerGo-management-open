package script

import (
	"database/sql"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/log"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/conf"
	_ "github.com/go-sql-driver/mysql"
)

func DataMigrations() {
	mysqlConf := conf.Conf.MySQL
	// 读取数据库配置
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", mysqlConf.Username, mysqlConf.Password, mysqlConf.Host, mysqlConf.Port, mysqlConf.DBName))
	if err != nil {
		log.Logger.Error("连接数据库失败", err)
	}
	defer func(db *sql.DB) {
		err = db.Close()
		if err != nil {
			log.Logger.Error("数据库关闭失败")
		}
	}(db)

	// 初始化所有mysql表结构
	FirstInitMysqlTable(db)

	// 初始化所有内容数据
	FirstInitMysqlContent(db)

	// 执行数据变更迁移sql
	ExecMigrationSql(db)

	return
}
