package script

import (
	"database/sql"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/conf"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func DataMigrations() {
	mysqlConf := conf.Conf.MySQL
	// 读取数据库配置
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", mysqlConf.Username, mysqlConf.Password, mysqlConf.Host, mysqlConf.Port, mysqlConf.DBName))
	if err != nil {
		fmt.Println("连接数据库失败", err)
	}
	defer func(db *sql.DB) {
		err = db.Close()
		if err != nil {
			fmt.Println("数据库关闭失败")
		}
	}(db)

	// 创建迁移记录表
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS `migrations` (\n  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',\n  `version` varchar(50) NOT NULL COMMENT '版本号',\n  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',\n  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',\n  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',\n  PRIMARY KEY (`id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;")
	if err != nil {
		fmt.Println("创建迁移记录表失败：", err)
	}

	// 获取已执行的脚本版本号
	var versions []string
	rows, err := db.Query("SELECT version FROM migrations")
	if err != nil {
		fmt.Println("查询数据迁移记录失败：", err)
	}
	defer func(rows *sql.Rows) {
		err = rows.Close()
		if err != nil {
			fmt.Println("关闭查询迁移记录数据失败：", err)
		}
	}(rows)
	for rows.Next() {
		var version string
		if err = rows.Scan(&version); err != nil {
			fmt.Println("遍历版本迁移记录数据失败：", err)
		}
		versions = append(versions, version)
	}
	if err = rows.Err(); err != nil {
		fmt.Println("迭代数据迁移记录表失败：", err)
	}

	// 获取所有更新sql
	AlterSqlMap := GetAllAlterSqlMap()

	// 执行未执行的脚本文件
	for version, allSql := range AlterSqlMap {
		if contains(versions, version) {
			continue
		}

		// 执行SQL语句
		for _, statement := range allSql {
			if strings.TrimSpace(statement) == "" {
				continue
			}
			_, err := db.Exec(statement)
			if err != nil {
				fmt.Println("执行sql语句失败", err)
				continue
			}
		}
		_, err = db.Exec("INSERT INTO migrations (version) VALUES (?)", version)
		if err != nil {
			fmt.Println("添加版本迁移记录失败：", err)
		}
		fmt.Println("数据迁移成功，迁移的版本号为：", version)
	}
	return
}

func contains(versions []string, version string) bool {
	for _, v := range versions {
		if v == version {
			return true
		}
	}
	return false
}
