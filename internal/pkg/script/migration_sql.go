package script

import (
	"database/sql"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/log"
	"strings"
)

func GetAllAlterSqlMap() map[string][]string {
	alterSqlMap := make(map[string][]string)

	// management版本1.1.2
	v112 := make([]string, 0, 10)
	v112 = append(v112, "alter table team add column company_id varchar(100) not null default '' comment '所属企业id' after `name`;")
	v112 = append(v112, "alter table team add column description text default null comment '团队描述' after `name`;")
	v112 = append(v112, "alter table `user` add column account varchar(100) not null default '' comment '账号' after user_id;")
	v112 = append(v112, "alter table user_team add column invite_time datetime default null comment '邀请时间' after invite_user_id;")
	v112 = append(v112, "alter table user_team add column team_role_id varchar(100) not null default '' comment '角色id (角色表对应)' after role_id;")
	v112 = append(v112, "alter table auto_plan_report add column report_name varchar(125) not null default '' comment '报告名称' after report_id;")
	v112 = append(v112, "alter table stress_plan_report add column report_name varchar(125) not null default '' comment '报告名称' after report_id;")
	v112 = append(v112, "update target set source = 0 where target_type = 'folder' or target_type = 'api';")
	v112 = append(v112, "update target set target_type = 'folder' where target_type = 'group';")
	alterSqlMap["1.1.2"] = v112

	// management版本2.0.0
	v200 := make([]string, 0, 10)
	v200 = append(v200, "ALTER TABLE team_env DROP COLUMN sort;")
	v200 = append(v200, "ALTER TABLE team_env DROP COLUMN `status`;")
	v200 = append(v200, "ALTER TABLE team_env DROP COLUMN recent_user_id;")
	v200 = append(v200, "ALTER TABLE team_env_service DROP COLUMN sort;")
	v200 = append(v200, "ALTER TABLE teamteam_env_service_env DROP COLUMN `status`;")
	v200 = append(v200, "ALTER TABLE teamteam_env_service_env DROP COLUMN recent_user_id;")
	v200 = append(v200, "ALTER TABLE target ADD is_disabled tinyint(2) NOT NULL DEFAULT '0' COMMENT '运行计划时是否禁用：0-不禁用，1-禁用' after is_checked;")
	v200 = append(v200, "ALTER TABLE stress_plan_task_conf ADD machine_dispatch_mode_conf text NOT NULL COMMENT '分布式压力机配置' after mode_conf;")
	v200 = append(v200, "ALTER TABLE stress_plan_task_conf ADD is_open_distributed tinyint(2) NOT NULL DEFAULT '0' COMMENT '是否开启分布式调度：0-关闭，1-开启' after mode_conf;")
	v200 = append(v200, "ALTER TABLE stress_plan_timed_task_conf ADD machine_dispatch_mode_conf text NOT NULL COMMENT '分布式压力机配置' after mode_conf;")
	v200 = append(v200, "ALTER TABLE stress_plan_timed_task_conf ADD is_open_distributed tinyint(2) NOT NULL DEFAULT '0' COMMENT '是否开启分布式调度：0-关闭，1-开启' after mode_conf;")
	v200 = append(v200, "ALTER TABLE preinstall_conf ADD machine_dispatch_mode_conf text NOT NULL COMMENT '分布式压力机配置' after timed_task_conf;")
	v200 = append(v200, "ALTER TABLE preinstall_conf ADD is_open_distributed tinyint(2) NOT NULL DEFAULT '0' COMMENT '是否开启分布式调度：0-关闭，1-开启' after timed_task_conf;")
	v200 = append(v200, "ALTER TABLE target MODIFY COLUMN source tinyint(4) NOT NULL DEFAULT '0' COMMENT '数据来源：0-测试对象，1-场景管理，2-性能，3-自动化测试， 4-mock';")
	v200 = append(v200, "ALTER TABLE report_machine ADD concurrency bigint(20) NOT NULL DEFAULT '0' COMMENT '并发数' after ip;")
	v200 = append(v200, "ALTER TABLE `team` ADD COLUMN `description` text COMMENT '团队描述' AFTER `name`;")
	v200 = append(v200, "ALTER TABLE `team` ADD COLUMN `company_id` varchar(100) NOT NULL DEFAULT '' COMMENT '所属企业id' AFTER `description`;")
	v200 = append(v200, "ALTER TABLE `user` ADD COLUMN `account` varchar(100) NOT NULL DEFAULT '' COMMENT '账号' AFTER `user_id`;")
	v200 = append(v200, "ALTER TABLE `user_team` ADD COLUMN `invite_time` datetime DEFAULT NULL COMMENT '邀请时间' AFTER `invite_user_id`;")
	v200 = append(v200, "ALTER TABLE `user_team` ADD COLUMN `is_show` tinyint(2) NOT NULL DEFAULT 1 COMMENT '是否展示到团队列表  1:展示   2:不展示' AFTER `sort`;")
	alterSqlMap["2.0.0"] = v200

	// management版本2.0.0.1
	v2001 := make([]string, 0, 10)
	v2001 = append(v2001, "ALTER TABLE team_env_service DROP COLUMN protocol_type;")
	alterSqlMap["2.0.0.1"] = v2001

	// management版本2.0.0.1
	v2002 := make([]string, 0, 10)
	v2002 = append(v2002, "ALTER TABLE `user_team` MODIFY COLUMN `invite_time` datetime DEFAULT NULL COMMENT '邀请时间';")
	alterSqlMap["2.0.0.2"] = v2002

	return alterSqlMap
}

func ExecMigrationSql(db *sql.DB) {
	// 创建迁移记录表
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS `migrations` (\n  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',\n  `version` varchar(50) NOT NULL COMMENT '版本号',\n  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',\n  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',\n  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',\n  PRIMARY KEY (`id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;")
	if err != nil {
		log.Logger.Error("创建迁移记录表失败：", err)
	}

	// 获取已执行的脚本版本号
	versions := make([]string, 0)
	rows, err := db.Query("SELECT version FROM migrations")
	if err != nil {
		log.Logger.Error("查询数据迁移记录失败：", err)
	}
	defer func(rows *sql.Rows) {
		err = rows.Close()
		if err != nil {
			log.Logger.Error("关闭查询迁移记录数据失败：", err)
		}
	}(rows)
	for rows.Next() {
		var version string
		if err = rows.Scan(&version); err != nil {
			log.Logger.Error("遍历版本迁移记录数据失败：", err)
		}
		versions = append(versions, version)
	}
	if err = rows.Err(); err != nil {
		log.Logger.Error("迭代数据迁移记录表失败：", err)
	}

	// 获取所有更新sql
	AlterSqlMap := GetAllAlterSqlMap()

	// 执行需要数据迁移版本对应的sql语句
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
				log.Logger.Error("执行sql语句失败", err)
				continue
			}
		}
		_, err = db.Exec("INSERT INTO migrations (version) VALUES (?)", version)
		if err != nil {
			log.Logger.Error("添加版本迁移记录失败：", err)
		}
		log.Logger.Error("数据迁移成功，迁移的版本号为：", version)
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
