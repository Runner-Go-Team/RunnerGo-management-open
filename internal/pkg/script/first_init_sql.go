package script

import (
	"context"
	"database/sql"
	"errors"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/log"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/uuid"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/conf"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/query"
	"github.com/go-omnibus/omnibus"
	"gorm.io/gorm"
	"math/rand"
	"strings"
	"time"
)

func FirstInitMysqlTable(db *sql.DB) {
	// 所有初始化数据库表结构
	initTableSql := make([]string, 0, 100)
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `auto_plan` (\n`id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',\n`plan_id` varchar(100) NOT NULL COMMENT '计划ID',\n`rank_id` bigint(10) NOT NULL DEFAULT '0' COMMENT '序号ID',\n`team_id` varchar(100) NOT NULL COMMENT '团队ID',\n`plan_name` varchar(255) NOT NULL COMMENT '计划名称',\n`task_type` tinyint(2) NOT NULL DEFAULT '1' COMMENT '计划类型：1-普通任务，2-定时任务',\n`status` tinyint(2) NOT NULL DEFAULT '1' COMMENT '计划状：1-未开始，2-进行中',\n`create_user_id` varchar(100) NOT NULL COMMENT '创建人id',\n`run_user_id` varchar(100) NOT NULL COMMENT '运行人id',\n`remark` text COMMENT '备注',\n`run_count` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '运行次数',\n`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',\n`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',\n`deleted_at` datetime DEFAULT NULL COMMENT '删除时间',\nPRIMARY KEY (`id`),\nKEY `idx_plan_id` (`plan_id`),\nKEY `idx_team_id` (`team_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='自动化测试-计划表';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `auto_plan_email` (\n`id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',\n`plan_id` varchar(100) NOT NULL DEFAULT '0' COMMENT '计划ID',\n`team_id` varchar(100) NOT NULL COMMENT '团队ID',\n`email` varchar(255) NOT NULL COMMENT '邮箱',\n`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',\n`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',\n`deleted_at` datetime DEFAULT NULL COMMENT '删除时间',\nPRIMARY KEY (`id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='自动化测计划—收件人邮箱表';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `auto_plan_report` (\n`id` bigint(20) NOT NULL AUTO_INCREMENT,\n`report_id` varchar(100) NOT NULL COMMENT '报告ID',\n`report_name` varchar(125) NOT NULL COMMENT '报告名称',\n`plan_id` varchar(100) NOT NULL COMMENT '计划ID',\n`rank_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '序号ID',\n`plan_name` varchar(255) NOT NULL COMMENT '计划名称',\n`team_id` varchar(100) NOT NULL COMMENT '团队ID',\n`task_type` int(11) NOT NULL DEFAULT '0' COMMENT '任务类型',\n`task_mode` int(11) NOT NULL DEFAULT '0' COMMENT '运行模式：1-按测试用例运行',\n`control_mode` tinyint(2) NOT NULL DEFAULT '0' COMMENT '控制模式：0-集中模式，1-单独模式',\n`scene_run_order` tinyint(2) NOT NULL DEFAULT '1' COMMENT '场景运行次序：1-顺序执行，2-同时执行',\n`test_case_run_order` tinyint(2) NOT NULL DEFAULT '1' COMMENT '测试用例运行次序：1-顺序执行，2-同时执行',\n`run_duration_time` bigint(20) NOT NULL DEFAULT '0' COMMENT '任务运行持续时长',\n`status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '报告状态1:进行中，2:已完成',\n`run_user_id` varchar(100) NOT NULL DEFAULT '0' COMMENT '启动人id',\n`remark` text NOT NULL COMMENT '备注',\n`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间（执行时间）',\n`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',\n`deleted_at` datetime DEFAULT NULL COMMENT '删除时间',\nPRIMARY KEY (`id`),\nKEY `idx_report_id` (`report_id`),\nKEY `idx_team_id` (`team_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='自动化测试计划-报告表';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `auto_plan_task_conf` (\n`id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '配置ID',\n`plan_id` varchar(100) NOT NULL DEFAULT '0' COMMENT '计划ID',\n`team_id` varchar(100) NOT NULL COMMENT '团队ID',\n`task_type` tinyint(2) NOT NULL DEFAULT '0' COMMENT '任务类型：1-普通模式，2-定时任务',\n`task_mode` tinyint(2) NOT NULL DEFAULT '1' COMMENT '运行模式：1-按照用例执行',\n`scene_run_order` tinyint(2) NOT NULL DEFAULT '1' COMMENT '场景运行次序：1-顺序执行，2-同时执行',\n`test_case_run_order` tinyint(2) NOT NULL DEFAULT '1' COMMENT '用例运行次序：1-顺序执行，2-同时执行',\n`run_user_id` varchar(100) NOT NULL DEFAULT '0' COMMENT '运行人用户ID',\n`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',\n`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',\n`deleted_at` datetime DEFAULT NULL COMMENT '删除时间',\nPRIMARY KEY (`id`),\nKEY `idx_plan_id` (`plan_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='自动化测试—普通任务配置表';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `auto_plan_timed_task_conf` (\n`id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '表id',\n`plan_id` varchar(100) NOT NULL DEFAULT '0' COMMENT '计划id',\n`team_id` varchar(100) NOT NULL COMMENT '团队id',\n`frequency` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '任务执行频次: 0-一次，1-每天，2-每周，3-每月',\n`task_exec_time` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '任务执行时间',\n`task_close_time` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '任务结束时间',\n`task_type` tinyint(2) NOT NULL DEFAULT '2' COMMENT '任务类型：1-普通任务，2-定时任务',\n`task_mode` tinyint(2) NOT NULL DEFAULT '1' COMMENT '运行模式：1-按照用例执行',\n`scene_run_order` tinyint(2) NOT NULL DEFAULT '1' COMMENT '场景运行次序：1-顺序执行，2-同时执行',\n`test_case_run_order` tinyint(2) NOT NULL DEFAULT '1' COMMENT '测试用例运行次序：1-顺序执行，2-同时执行',\n`status` tinyint(2) NOT NULL DEFAULT '0' COMMENT '任务状态：0-未启用，1-运行中，2-已过期',\n`run_user_id` varchar(100) NOT NULL DEFAULT '0' COMMENT '运行人用户ID',\n`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',\n`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',\n`deleted_at` datetime DEFAULT NULL COMMENT '删除时间',\nPRIMARY KEY (`id`),\nKEY `idx_plan_id` (`plan_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='自动化测试-定时任务配置表';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `company` (\n`id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',\n`company_id` varchar(100) NOT NULL COMMENT '企业id',\n`name` varchar(100) NOT NULL DEFAULT '' COMMENT '企业名称',\n`logo` varchar(255) NOT NULL DEFAULT '' COMMENT '企业logo',\n`expire_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '服务到期时间',\n`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,\n`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,\n`deleted_at` datetime DEFAULT NULL,\nPRIMARY KEY (`id`),\nKEY `idx_company_id` (`company_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='企业表';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `global_variable` (\n`id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',\n`team_id` varchar(100) NOT NULL COMMENT '团队id',\n`type` tinyint(2) NOT NULL DEFAULT '0' COMMENT '变量类型：1-全局变量，2-场景变量',\n`var` varchar(255) NOT NULL COMMENT '变量名',\n`val` text NOT NULL COMMENT '变量值',\n`description` text NOT NULL COMMENT '描述',\n`status` tinyint(2) NOT NULL DEFAULT '1' COMMENT '开关状态：1-开启，2-关闭',\n`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',\n`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',\n`deleted_at` datetime DEFAULT NULL COMMENT '删除时间',\nPRIMARY KEY (`id`),\nKEY `idx_team_id` (`team_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='全局变量表';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `machine` (\n`id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',\n`region` varchar(64) NOT NULL COMMENT '所属区域',\n`ip` varchar(16) NOT NULL COMMENT '机器IP',\n`port` int(11) unsigned NOT NULL COMMENT '端口',\n`name` varchar(200) NOT NULL COMMENT '机器名称',\n`cpu_usage` float unsigned NOT NULL DEFAULT '0' COMMENT 'CPU使用率',\n`cpu_load_one` float unsigned NOT NULL DEFAULT '0' COMMENT 'CPU-1分钟内平均负载',\n`cpu_load_five` float unsigned NOT NULL DEFAULT '0' COMMENT 'CPU-5分钟内平均负载',\n`cpu_load_fifteen` float unsigned NOT NULL DEFAULT '0' COMMENT 'CPU-15分钟内平均负载',\n`mem_usage` float unsigned NOT NULL DEFAULT '0' COMMENT '内存使用率',\n`disk_usage` float unsigned NOT NULL DEFAULT '0' COMMENT '磁盘使用率',\n`max_goroutines` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '最大协程数',\n`current_goroutines` bigint(20) NOT NULL DEFAULT '0' COMMENT '已用协程数',\n`server_type` tinyint(2) unsigned NOT NULL DEFAULT '1' COMMENT '机器类型：1-主力机器，2-备用机器',\n`status` tinyint(2) unsigned NOT NULL DEFAULT '1' COMMENT '机器状态：1-使用中，2-已卸载',\n`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',\n`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',\n`deleted_at` datetime DEFAULT NULL COMMENT '删除时间',\nPRIMARY KEY (`id`),\nKEY `machine_region_ip_status_index` (`region`,`ip`,`status`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='压力测试机器表';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `migrations` (\n`id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',\n`version` varchar(50) NOT NULL COMMENT '版本号',\n`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',\n`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',\n`deleted_at` datetime DEFAULT NULL COMMENT '删除时间',\nPRIMARY KEY (`id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `mock_target` (\n`id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'id',\n`target_id` varchar(100) NOT NULL COMMENT '全局唯一ID',\n`team_id` varchar(100) NOT NULL COMMENT '团队id',\n`target_type` varchar(10) NOT NULL COMMENT '类型：文件夹，接口，分组，场景,测试用例',\n`name` varchar(255) NOT NULL COMMENT '名称',\n`parent_id` varchar(100) NOT NULL DEFAULT '0' COMMENT '父级ID',\n`method` varchar(16) NOT NULL COMMENT '方法',\n`sort` int(11) NOT NULL DEFAULT '0' COMMENT '排序',\n`type_sort` int(11) NOT NULL DEFAULT '0' COMMENT '类型排序',\n`status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '回收站状态：1-正常，2-回收站',\n`version` int(11) NOT NULL DEFAULT '0' COMMENT '产品版本号',\n`created_user_id` varchar(100) NOT NULL COMMENT '创建人ID',\n`recent_user_id` varchar(100) NOT NULL COMMENT '最近修改人ID',\n`description` text NOT NULL COMMENT '备注',\n`source` tinyint(4) NOT NULL DEFAULT '0' COMMENT '数据来源：0-mock管理',\n`plan_id` varchar(100) NOT NULL COMMENT '计划id',\n`source_id` varchar(100) NOT NULL COMMENT '引用来源ID',\n`is_checked` tinyint(2) NOT NULL DEFAULT '1' COMMENT '是否开启：1-开启，2-关闭',\n`is_disabled` tinyint(2) NOT NULL DEFAULT '0' COMMENT '运行计划时是否禁用：0-不禁用，1-禁用',\n`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',\n`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',\n`deleted_at` datetime DEFAULT NULL COMMENT '删除时间',\nPRIMARY KEY (`id`),\nKEY `idx_target_id` (`target_id`),\nKEY `idx_plan_id` (`plan_id`),\nKEY `idx_team_id` (`team_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='创建目标';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `mock_target_debug_log` (\n`id` bigint(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',\n`target_id` varchar(100) NOT NULL COMMENT '目标唯一ID',\n`target_type` tinyint(2) NOT NULL COMMENT '目标类型：1-api，2-scene',\n`team_id` varchar(100) NOT NULL COMMENT '团队ID',\n`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',\n`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',\n`deleted_at` datetime DEFAULT NULL COMMENT '删除时间',\nPRIMARY KEY (`id`),\nKEY `idx_target_id` (`target_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='mock目标调试日志表';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `permission` (\n`id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',\n`permission_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '权限ID',\n`title` varchar(100) NOT NULL DEFAULT '' COMMENT '权限内容',\n`mark` varchar(100) NOT NULL DEFAULT '' COMMENT '权限标识',\n`url` varchar(100) NOT NULL DEFAULT '' COMMENT '权限url',\n`type` tinyint(2) NOT NULL DEFAULT '0' COMMENT '类型（1：权限   2：功能）',\n`group_id` int(11) NOT NULL DEFAULT '0' COMMENT '所属权限组（1：企业成员管理  2：团队管理  3：角色管理）',\n`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,\n`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,\n`deleted_at` datetime DEFAULT NULL,\nPRIMARY KEY (`id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='权限表';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `preinstall_conf` (\n`id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',\n`conf_name` varchar(100) NOT NULL COMMENT '配置名称',\n`team_id` varchar(100) NOT NULL COMMENT '团队ID',\n`user_id` varchar(100) NOT NULL DEFAULT '0' COMMENT '用户ID',\n`user_name` varchar(64) NOT NULL COMMENT '用户名称',\n`task_type` tinyint(2) NOT NULL DEFAULT '0' COMMENT '任务类型',\n`task_mode` tinyint(2) NOT NULL DEFAULT '0' COMMENT '压测模式',\n`control_mode` tinyint(2) NOT NULL DEFAULT '0' COMMENT '控制模式：0-集中模式，1-单独模式',\n`debug_mode` varchar(100) NOT NULL DEFAULT 'stop' COMMENT 'debug模式：stop-关闭，all-开启全部日志，only_success-开启仅成功日志，only_error-开启仅错误日志',\n`mode_conf` text NOT NULL COMMENT '压测配置详情',\n`timed_task_conf` text NOT NULL COMMENT '定时任务相关配置',\n`is_open_distributed` tinyint(2) NOT NULL DEFAULT '0' COMMENT '是否开启分布式调度：0-关闭，1-开启',\n`machine_dispatch_mode_conf` text NOT NULL COMMENT '分布式压力机配置',\n`created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',\n`updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',\n`deleted_at` timestamp NULL DEFAULT NULL COMMENT '删除时间',\nPRIMARY KEY (`id`),\nKEY `idx_team_id` (`team_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='预设配置表';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `public_function` (\n`id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',\n`function` varchar(255) NOT NULL COMMENT '函数',\n`function_name` varchar(255) NOT NULL COMMENT '函数名称',\n`remark` text NOT NULL COMMENT '备注',\n`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',\n`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',\n`deleted_at` datetime DEFAULT NULL COMMENT '删除时间',\nPRIMARY KEY (`id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='公共函数表';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `report_machine` (\n`id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',\n`report_id` varchar(100) NOT NULL COMMENT '报告id',\n`plan_id` varchar(100) NOT NULL DEFAULT '0' COMMENT '计划ID',\n`team_id` varchar(100) NOT NULL COMMENT '团队ID',\n`ip` varchar(15) NOT NULL COMMENT '机器ip',\n`concurrency` bigint(20) NOT NULL DEFAULT '0' COMMENT '并发数',\n`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',\n`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',\n`deleted_at` datetime DEFAULT NULL COMMENT '删除时间',\nPRIMARY KEY (`id`),\nKEY `idx_report_id` (`report_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `role` (\n`id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',\n`role_id` varchar(100) NOT NULL COMMENT '角色id',\n`role_type` tinyint(2) NOT NULL DEFAULT '0' COMMENT '角色分类（1：企业  2：团队）',\n`name` varchar(100) NOT NULL DEFAULT '' COMMENT '角色名称',\n`company_id` varchar(100) NOT NULL DEFAULT '' COMMENT '企业id',\n`level` tinyint(2) NOT NULL DEFAULT '0' COMMENT '角色层级（1:超管/团队管理员 2:管理员/团队成员 3:普通成员/只读成员/自定义角色） ',\n`is_default` tinyint(2) NOT NULL DEFAULT '2' COMMENT '是否是默认角色  1：是   2：自定义角色',\n`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,\n`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,\n`deleted_at` datetime DEFAULT NULL,\nPRIMARY KEY (`id`),\nKEY `idx_role_id` (`role_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色表';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `role_permission` (\n`id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',\n`role_id` varchar(100) NOT NULL COMMENT '角色id',\n`permission_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '权限id',\n`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,\n`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,\n`deleted_at` datetime DEFAULT NULL,\nPRIMARY KEY (`id`),\nKEY `idx_role_id` (`role_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色权限表';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `role_type_permission` (\n`id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',\n`role_type` tinyint(2) NOT NULL DEFAULT '0' COMMENT '角色分类（1：企业  2：团队）',\n`permission_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '权限id',\n`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,\n`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,\n`deleted_at` datetime DEFAULT NULL,\nPRIMARY KEY (`id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色分类可拥有的权限';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `scene_variable` (\n`id` bigint(20) NOT NULL AUTO_INCREMENT,\n`team_id` varchar(100) NOT NULL COMMENT '团队id',\n`scene_id` varchar(100) NOT NULL COMMENT '场景ID',\n`type` tinyint(4) NOT NULL DEFAULT '0' COMMENT '使用范围：1-全局变量，2-场景变量',\n`var` varchar(255) NOT NULL COMMENT '变量名',\n`val` text NOT NULL COMMENT '变量值',\n`description` text NOT NULL COMMENT '描述',\n`status` tinyint(2) NOT NULL DEFAULT '1' COMMENT '开关状态：1-开启，2-关闭',\n`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',\n`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',\n`deleted_at` datetime DEFAULT NULL COMMENT '删除时间',\nPRIMARY KEY (`id`),\nKEY `idx_team_id` (`team_id`),\nKEY `idx_scene_id` (`scene_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='设置变量表';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `setting` (\n`id` bigint(20) NOT NULL AUTO_INCREMENT,\n`user_id` varchar(100) NOT NULL COMMENT '用户id',\n`team_id` varchar(100) NOT NULL COMMENT '当前团队id',\n`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,\n`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,\n`deleted_at` datetime DEFAULT NULL,\nPRIMARY KEY (`id`),\nKEY `idx_user_id` (`user_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='设置表';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `sms_log` (\n`id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键id',\n`type` tinyint(2) NOT NULL COMMENT '短信类型: 1-注册，2-登录，3-找回密码',\n`mobile` char(11) NOT NULL DEFAULT '' COMMENT '手机号',\n`content` varchar(200) NOT NULL COMMENT '短信内容',\n`verify_code` varchar(20) NOT NULL COMMENT '验证码',\n`verify_code_expiration_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '验证码有效时间',\n`client_ip` varchar(100) NOT NULL DEFAULT '' COMMENT '客户端IP',\n`send_status` tinyint(2) NOT NULL DEFAULT '1' COMMENT '发送状态：1-成功 2-失败',\n`verify_status` tinyint(2) NOT NULL DEFAULT '1' COMMENT '校验状态：1-未校验 2-已校验',\n`send_response` text NOT NULL COMMENT '短信服务响应',\n`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',\n`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',\n`deleted_at` datetime DEFAULT NULL COMMENT '删除时间',\nPRIMARY KEY (`id`),\nKEY `idx_type_mobile_verify_code` (`type`,`mobile`,`verify_code`,`verify_code_expiration_time`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='短信发送记录表';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `stress_plan` (\n`id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',\n`plan_id` varchar(100) NOT NULL DEFAULT '0' COMMENT '计划ID',\n`team_id` varchar(100) NOT NULL COMMENT '团队ID',\n`rank_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '序号ID',\n`plan_name` varchar(255) NOT NULL COMMENT '计划名称',\n`task_type` tinyint(2) NOT NULL DEFAULT '0' COMMENT '计划类型：1-普通任务，2-定时任务',\n`task_mode` tinyint(2) NOT NULL DEFAULT '0' COMMENT '压测类型: 1-并发模式，2-阶梯模式，3-错误率模式，4-响应时间模式，5-每秒请求数模式，6-每秒事务数模式',\n`status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '计划状态1:未开始,2:进行中',\n`create_user_id` varchar(100) NOT NULL DEFAULT '0' COMMENT '创建人id',\n`run_user_id` varchar(100) NOT NULL DEFAULT '0' COMMENT '运行人id',\n`remark` text NOT NULL COMMENT '备注',\n`run_count` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '运行次数',\n`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',\n`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',\n`deleted_at` datetime DEFAULT NULL COMMENT '删除时间',\nPRIMARY KEY (`id`),\nKEY `idx_plan_id` (`plan_id`),\nKEY `idx_team_id` (`team_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='性能计划表';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `stress_plan_email` (\n`id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',\n`plan_id` varchar(100) NOT NULL COMMENT '计划ID',\n`team_id` varchar(100) NOT NULL COMMENT '团队ID',\n`email` varchar(255) DEFAULT NULL COMMENT '邮箱',\n`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',\n`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',\n`deleted_at` datetime DEFAULT NULL COMMENT '删除时间',\nPRIMARY KEY (`id`),\nKEY `idx_plan_id` (`plan_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='性能计划收件人';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `stress_plan_report` (\n`id` bigint(20) NOT NULL AUTO_INCREMENT,\n`report_id` varchar(100) NOT NULL COMMENT '报告ID',\n`report_name` varchar(125) NOT NULL COMMENT '报告名称',\n`team_id` varchar(100) NOT NULL COMMENT '团队ID',\n`plan_id` varchar(100) NOT NULL COMMENT '计划ID',\n`rank_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '序号ID',\n`plan_name` varchar(255) NOT NULL COMMENT '计划名称',\n`scene_id` varchar(100) NOT NULL COMMENT '场景ID',\n`scene_name` varchar(255) NOT NULL COMMENT '场景名称',\n`task_type` int(11) NOT NULL COMMENT '任务类型',\n`task_mode` int(11) NOT NULL COMMENT '压测模式',\n`control_mode` tinyint(2) NOT NULL DEFAULT '0' COMMENT '控制模式：0-集中模式，1-单独模式',\n`debug_mode` varchar(100) NOT NULL DEFAULT 'stop' COMMENT 'debug模式：stop-关闭，all-开启全部日志，only_success-开启仅成功日志，only_error-开启仅错误日志',\n`status` tinyint(4) NOT NULL COMMENT '报告状态1:进行中，2:已完成',\n`remark` text NOT NULL COMMENT '备注',\n`run_user_id` varchar(100) NOT NULL COMMENT '启动人id',\n`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间（执行时间）',\n`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',\n`deleted_at` datetime DEFAULT NULL COMMENT '删除时间',\nPRIMARY KEY (`id`),\nKEY `idx_report_id` (`report_id`),\nKEY `idx_team_id` (`team_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='性能测试报告表';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `stress_plan_task_conf` (\n`id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '配置ID',\n`plan_id` varchar(100) NOT NULL DEFAULT '0' COMMENT '计划ID',\n`team_id` varchar(100) NOT NULL COMMENT '团队ID',\n`scene_id` varchar(100) NOT NULL COMMENT '场景ID',\n`task_type` tinyint(2) NOT NULL DEFAULT '0' COMMENT '任务类型：1-普通模式，2-定时任务',\n`task_mode` tinyint(2) NOT NULL DEFAULT '0' COMMENT '压测模式：1-并发模式，2-阶梯模式，3-错误率模式，4-响应时间模式，5-每秒请求数模式，6-每秒事务数模式',\n`control_mode` tinyint(2) NOT NULL DEFAULT '0' COMMENT '控制模式：0-集中模式，1-单独模式',\n`debug_mode` varchar(100) NOT NULL DEFAULT 'stop' COMMENT 'debug模式：stop-关闭，all-开启全部日志，only_success-开启仅成功日志，only_error-开启仅错误日志',\n`mode_conf` text NOT NULL COMMENT '压测模式配置详情',\n`is_open_distributed` tinyint(2) NOT NULL DEFAULT '0' COMMENT '是否开启分布式调度：0-关闭，1-开启',\n`machine_dispatch_mode_conf` text NOT NULL COMMENT '分布式压力机配置',\n`run_user_id` varchar(100) NOT NULL DEFAULT '0' COMMENT '运行人用户ID',\n`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',\n`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',\n`deleted_at` datetime DEFAULT NULL COMMENT '删除时间',\nPRIMARY KEY (`id`),\nKEY `idx_plan_id` (`plan_id`),\nKEY `idx_scene_id` (`scene_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='性能计划—普通任务配置表';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `stress_plan_timed_task_conf` (\n`id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '表id',\n`plan_id` varchar(100) NOT NULL COMMENT '计划id',\n`scene_id` varchar(100) NOT NULL COMMENT '场景id',\n`team_id` varchar(100) NOT NULL COMMENT '团队id',\n`user_id` varchar(100) NOT NULL COMMENT '用户ID',\n`frequency` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '任务执行频次: 0-一次，1-每天，2-每周，3-每月',\n`task_exec_time` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '任务执行时间',\n`task_close_time` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '任务结束时间',\n`task_type` tinyint(2) NOT NULL DEFAULT '2' COMMENT '任务类型：1-普通任务，2-定时任务',\n`task_mode` tinyint(2) NOT NULL DEFAULT '1' COMMENT '压测模式：1-并发模式，2-阶梯模式，3-错误率模式，4-响应时间模式，5-每秒请求数模式，6 -每秒事务数模式',\n`control_mode` tinyint(2) NOT NULL DEFAULT '0' COMMENT '控制模式：0-集中模式，1-单独模式',\n`debug_mode` varchar(100) NOT NULL DEFAULT 'stop' COMMENT 'debug模式：stop-关闭，all-开启全部日志，only_success-开启仅成功日志，only_error-开启仅错误日志',\n`mode_conf` text NOT NULL COMMENT '压测详细配置',\n`is_open_distributed` tinyint(2) NOT NULL DEFAULT '0' COMMENT '是否开启分布式调度：0-关闭，1-开启',\n`machine_dispatch_mode_conf` text NOT NULL COMMENT '分布式压力机配置',\n`run_user_id` varchar(100) NOT NULL DEFAULT '0' COMMENT '运行人ID',\n`status` tinyint(11) NOT NULL DEFAULT '0' COMMENT '任务状态：0-未启用，1-运行中，2-已过期',\n`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',\n`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',\n`deleted_at` datetime DEFAULT NULL COMMENT '删除时间',\nPRIMARY KEY (`id`),\nKEY `idx_plan_id` (`plan_id`),\nKEY `idx_scene_id` (`scene_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='性能计划-定时任务配置表';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `target` (\n`id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'id',\n`target_id` varchar(100) NOT NULL COMMENT '全局唯一ID',\n`team_id` varchar(100) NOT NULL COMMENT '团队id',\n`target_type` varchar(10) NOT NULL COMMENT '类型：文件夹，接口，分组，场景,测试用例',\n`name` varchar(255) NOT NULL COMMENT '名称',\n`parent_id` varchar(100) NOT NULL DEFAULT '0' COMMENT '父级ID',\n`method` varchar(16) NOT NULL COMMENT '方法',\n`sort` int(11) NOT NULL DEFAULT '0' COMMENT '排序',\n`type_sort` int(11) NOT NULL DEFAULT '0' COMMENT '类型排序',\n`status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '回收站状态：1-正常，2-回收站',\n`version` int(11) NOT NULL DEFAULT '0' COMMENT '产品版本号',\n`created_user_id` varchar(100) NOT NULL COMMENT '创建人ID',\n`recent_user_id` varchar(100) NOT NULL COMMENT '最近修改人ID',\n`description` text NOT NULL COMMENT '备注',\n`source` tinyint(4) NOT NULL DEFAULT '0' COMMENT '数据来源：0-测试对象，1-场景管理，2-性能，3-自动化测试',\n`plan_id` varchar(100) NOT NULL COMMENT '计划id',\n`source_id` varchar(100) NOT NULL COMMENT '引用来源ID',\n`is_checked` tinyint(2) NOT NULL DEFAULT '1' COMMENT '是否开启：1-开启，2-关闭',\n`is_disabled` tinyint(2) NOT NULL DEFAULT '0' COMMENT '运行计划时是否禁用：0-不禁用，1-禁用',\n`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',\n`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',\n`deleted_at` datetime DEFAULT NULL COMMENT '删除时间',\nPRIMARY KEY (`id`),\nKEY `idx_target_id` (`target_id`),\nKEY `idx_plan_id` (`plan_id`),\nKEY `idx_team_id` (`team_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='创建目标';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `team` (\n`id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',\n`team_id` varchar(100) NOT NULL COMMENT '团队ID',\n`name` varchar(64) NOT NULL COMMENT '团队名称',\n`description` text COMMENT '团队描述',\n`company_id` varchar(100) NOT NULL DEFAULT '' COMMENT '所属企业id',\n`type` tinyint(4) NOT NULL COMMENT '团队类型 1: 私有团队；2: 普通团队',\n`trial_expiration_date` datetime NOT NULL COMMENT '试用有效期',\n`is_vip` tinyint(2) NOT NULL DEFAULT '1' COMMENT '是否为付费团队 1-否 2-是',\n`vip_expiration_date` datetime NOT NULL COMMENT '付费有效期',\n`vum_num` bigint(20) NOT NULL DEFAULT '0' COMMENT '当前可用VUM总数',\n`max_user_num` bigint(20) NOT NULL DEFAULT '0' COMMENT '当前团队最大成员数量',\n`created_user_id` varchar(100) NOT NULL COMMENT '创建者id',\n`team_buy_version_type` int(10) NOT NULL DEFAULT '1' COMMENT '团队套餐类型：1-个人版，2-团队版，3-企业版，4-私有化部署',\n`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,\n`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,\n`deleted_at` datetime DEFAULT NULL,\nPRIMARY KEY (`id`),\nKEY `idx_team_id` (`team_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='团队表';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `team_env` (\n`id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',\n`team_id` varchar(100) NOT NULL COMMENT '团队id',\n`name` varchar(100) NOT NULL COMMENT '环境名称',\n`created_user_id` varchar(100) NOT NULL COMMENT '创建人id',\n`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',\n`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',\n`deleted_at` datetime DEFAULT NULL COMMENT '删除时间',\nPRIMARY KEY (`id`),\nKEY `idx_team_id` (`team_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='环境管理表';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `team_env_database` (\n`id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',\n`team_id` varchar(100) NOT NULL COMMENT '团队id',\n`team_env_id` bigint(20) NOT NULL COMMENT '环境变量id',\n`type` varchar(100) NOT NULL COMMENT '数据库类型',\n`server_name` varchar(100) NOT NULL COMMENT 'mysql服务名称',\n`host` varchar(200) NOT NULL COMMENT '服务地址',\n`port` int(11) NOT NULL COMMENT '端口号',\n`user` varchar(100) NOT NULL COMMENT '账号',\n`password` varchar(200) NOT NULL COMMENT '密码',\n`db_name` varchar(100) NOT NULL COMMENT '数据库名称',\n`charset` varchar(100) NOT NULL DEFAULT 'utf8mb4' COMMENT '字符编码集',\n`created_user_id` varchar(100) NOT NULL COMMENT '创建人id',\n`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',\n`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',\n`deleted_at` datetime DEFAULT NULL COMMENT '删除时间',\nPRIMARY KEY (`id`),\nKEY `idx_team_id` (`team_id`),\nKEY `idx_team_env_id` (`team_env_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Sql数据库服务基础信息表';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `team_env_service` (\n  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',\n  `team_id` varchar(100) NOT NULL COMMENT '团队id',\n  `team_env_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '环境id',\n  `name` varchar(100) NOT NULL COMMENT '服务名称',\n  `content` varchar(200) NOT NULL COMMENT '服务URL',\n  `created_user_id` varchar(100) NOT NULL COMMENT '创建人id',\n  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',\n  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',\n  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',\n  PRIMARY KEY (`id`),\n  KEY `idxx_team_id` (`team_id`)\n) ENGINE=InnoDB AUTO_INCREMENT=1582 DEFAULT CHARSET=utf8mb4 COMMENT='团队环境服务管理';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `team_user_queue` (\n`id` bigint(20) NOT NULL AUTO_INCREMENT,\n`email` varchar(255) NOT NULL COMMENT '邮箱',\n`team_id` varchar(100) NOT NULL COMMENT '团队id',\n`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,\n`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,\n`deleted_at` datetime DEFAULT NULL,\nPRIMARY KEY (`id`),\nKEY `idx_team_id` (`team_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='邀请待注册队列';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `third_notice` (\n`id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',\n`notice_id` varchar(100) NOT NULL COMMENT '通知id',\n`name` varchar(100) NOT NULL DEFAULT '' COMMENT '通知名称',\n`channel_id` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '三方通知渠道id',\n`params` json DEFAULT NULL COMMENT '通知参数',\n`status` tinyint(2) NOT NULL DEFAULT '1' COMMENT '1:启用 2:禁用',\n`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,\n`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,\n`deleted_at` datetime DEFAULT NULL,\nPRIMARY KEY (`id`),\nKEY `idx_notice_id` (`notice_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='三方通知设置';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `third_notice_channel` (\n`id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',\n`name` varchar(100) NOT NULL DEFAULT '' COMMENT '名称',\n`type` tinyint(2) NOT NULL DEFAULT '0' COMMENT '类型 1:飞书  2:企业微信  3:邮箱  4:钉钉',\n`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,\n`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,\n`deleted_at` datetime DEFAULT NULL,\nPRIMARY KEY (`id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='三方通知渠道';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `third_notice_group` (\n`id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',\n`group_id` varchar(100) NOT NULL COMMENT '通知组id',\n`name` varchar(100) NOT NULL DEFAULT '' COMMENT '通知组名称',\n`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,\n`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,\n`deleted_at` datetime DEFAULT NULL,\nPRIMARY KEY (`id`),\nKEY `idx_group_id` (`group_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='三方通知组表';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `third_notice_group_event` (\n`id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',\n`group_id` varchar(100) NOT NULL DEFAULT '' COMMENT '通知组id',\n`event_id` int(11) NOT NULL DEFAULT '0' COMMENT '事件id',\n`plan_id` varchar(100) NOT NULL DEFAULT '' COMMENT '计划ID',\n`team_id` varchar(100) NOT NULL DEFAULT '' COMMENT '团队ID',\n`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,\n`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,\n`deleted_at` datetime DEFAULT NULL,\nPRIMARY KEY (`id`),\nKEY `idx_group_id` (`group_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='三方通知组触发事件表';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `third_notice_group_relate` (\n`id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',\n`group_id` varchar(100) NOT NULL COMMENT '通知组id',\n`notice_id` varchar(100) NOT NULL COMMENT '通知id',\n`params` json DEFAULT NULL COMMENT '通知目标参数',\n`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,\n`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,\n`deleted_at` datetime DEFAULT NULL,\nPRIMARY KEY (`id`),\nKEY `idx_notice_id` (`notice_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='三方通知组通知关联表';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `user` (\n`id` bigint(20) NOT NULL AUTO_INCREMENT,\n`user_id` varchar(100) NOT NULL COMMENT '用户id',\n`account` varchar(100) NOT NULL DEFAULT '' COMMENT '账号',\n`email` varchar(100) NOT NULL COMMENT '邮箱',\n`mobile` char(11) NOT NULL COMMENT '手机号',\n`password` varchar(255) NOT NULL COMMENT '密码',\n`nickname` varchar(64) NOT NULL COMMENT '昵称',\n`avatar` varchar(255) DEFAULT NULL COMMENT '头像',\n`wechat_open_id` varchar(100) NOT NULL COMMENT '微信开放的唯一id',\n`utm_source` varchar(50) NOT NULL COMMENT '渠道来源',\n`last_login_at` datetime DEFAULT NULL COMMENT '最近登录时间',\n`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,\n`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,\n`deleted_at` datetime DEFAULT NULL,\nPRIMARY KEY (`id`),\nKEY `idx_user_id` (`user_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `user_collect_info` (\n`id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',\n`user_id` varchar(100) NOT NULL COMMENT '用户id',\n`industry` varchar(100) NOT NULL COMMENT '所属行业',\n`team_size` varchar(20) NOT NULL COMMENT '团队规模',\n`work_type` varchar(20) NOT NULL COMMENT '工作岗位',\n`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',\n`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',\n`deleted_at` datetime DEFAULT NULL COMMENT '删除时间',\nPRIMARY KEY (`id`),\nKEY `idx_user_id` (`user_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `user_company` (\n`id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',\n`user_id` varchar(100) NOT NULL COMMENT '用户id',\n`company_id` varchar(100) NOT NULL COMMENT '企业id',\n`invite_user_id` varchar(100) NOT NULL DEFAULT '0' COMMENT '邀请人id',\n`invite_time` datetime DEFAULT NULL COMMENT '邀请时间',\n`status` tinyint(2) unsigned NOT NULL DEFAULT '1' COMMENT '状态：1-正常，2-已禁用',\n`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,\n`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,\n`deleted_at` datetime DEFAULT NULL,\nPRIMARY KEY (`id`),\nKEY `idx_company_id` (`company_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户企业关系表';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `user_role` (\n`id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',\n`role_id` varchar(100) NOT NULL COMMENT '角色id',\n`user_id` varchar(100) NOT NULL COMMENT '用户id',\n`company_id` varchar(100) NOT NULL DEFAULT '' COMMENT '企业id',\n`team_id` varchar(100) NOT NULL DEFAULT '' COMMENT '团队id',\n`invite_user_id` varchar(100) NOT NULL DEFAULT '0' COMMENT '邀请人id',\n`invite_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '邀请时间',\n`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,\n`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,\n`deleted_at` datetime DEFAULT NULL,\nPRIMARY KEY (`id`),\nKEY `idx_role_id` (`role_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户角色关联表（企业角色、团队角色）';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `user_team` (\n`id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',\n`user_id` varchar(100) NOT NULL COMMENT '用户ID',\n`team_id` varchar(100) NOT NULL COMMENT '团队id',\n`role_id` bigint(20) NOT NULL COMMENT '角色id1:超级管理员，2成员，3管理员',\n`invite_user_id` varchar(100) NOT NULL DEFAULT '0' COMMENT '邀请人id',\n`invite_time` datetime DEFAULT NULL COMMENT '邀请时间',\n`sort` int(11) NOT NULL DEFAULT '0',\n`is_show` tinyint(2) NOT NULL DEFAULT '1' COMMENT '是否展示到团队列表  1:展示   2:不展示',\n`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,\n`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,\n`deleted_at` datetime DEFAULT NULL,\nPRIMARY KEY (`id`),\nKEY `idx_team_id` (`team_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户团队关系表';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `user_team_collection` (\n`id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',\n`user_id` varchar(100) NOT NULL COMMENT '用户ID',\n`team_id` varchar(100) NOT NULL COMMENT '团队id',\n`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,\n`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,\n`deleted_at` datetime DEFAULT NULL,\nPRIMARY KEY (`id`),\nKEY `idx_user_id` (`user_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户收藏团队表';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `variable` (\n`id` bigint(20) NOT NULL AUTO_INCREMENT,\n`team_id` varchar(100) NOT NULL COMMENT '团队id',\n`scene_id` varchar(100) NOT NULL COMMENT '场景ID',\n`type` tinyint(4) NOT NULL DEFAULT '0' COMMENT '使用范围：1-全局变量，2-场景变量',\n`var` varchar(255) NOT NULL COMMENT '变量名',\n`val` text NOT NULL COMMENT '变量值',\n`description` text NOT NULL COMMENT '描述',\n`status` tinyint(2) NOT NULL DEFAULT '1' COMMENT '开关状态：1-开启，2-关闭',\n`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',\n`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',\n`deleted_at` datetime DEFAULT NULL COMMENT '删除时间',\nPRIMARY KEY (`id`),\nKEY `idx_team_id` (`team_id`),\nKEY `idx_scene_id` (`scene_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='设置变量表';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `variable_import` (\n`id` bigint(20) NOT NULL AUTO_INCREMENT,\n`team_id` varchar(100) NOT NULL COMMENT '团队id',\n`scene_id` varchar(100) NOT NULL DEFAULT '0' COMMENT '场景id',\n`name` varchar(128) NOT NULL COMMENT '文件名称',\n`url` varchar(255) NOT NULL COMMENT '文件地址',\n`uploader_id` varchar(100) NOT NULL COMMENT '上传人id',\n`status` tinyint(2) NOT NULL DEFAULT '1' COMMENT '开关状态：1-开，2-关',\n`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',\n`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',\n`deleted_at` datetime DEFAULT NULL COMMENT '删除时间',\nPRIMARY KEY (`id`),\nKEY `idx_team_id` (`team_id`),\nKEY `idx_scene_id` (`scene_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='导入变量表';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `element` (\n  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'id',\n  `element_id` varchar(100) NOT NULL COMMENT '全局唯一ID',\n  `element_type` varchar(10) NOT NULL COMMENT '类型：文件夹，元素',\n  `team_id` varchar(100) NOT NULL COMMENT '团队id',\n  `name` varchar(255) NOT NULL COMMENT '名称',\n  `parent_id` varchar(100) NOT NULL DEFAULT '0' COMMENT '父级ID',\n  `locators` json DEFAULT NULL COMMENT '定位元素属性',\n  `sort` int(11) NOT NULL DEFAULT '0' COMMENT '排序',\n  `version` int(11) NOT NULL DEFAULT '0' COMMENT '产品版本号',\n  `created_user_id` varchar(100) NOT NULL COMMENT '创建人ID',\n  `description` text NOT NULL COMMENT '备注',\n  `source` tinyint(4) NOT NULL DEFAULT '0' COMMENT '数据来源：0-元素管理，1-场景管理',\n  `source_id` varchar(100) NOT NULL DEFAULT '' COMMENT '引用来源ID',\n  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',\n  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',\n  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',\n  PRIMARY KEY (`id`),\n  KEY `idx_element_id` (`element_id`),\n  KEY `idx_team_id` (`team_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='元素表';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `ui_plan` (\n  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',\n  `plan_id` varchar(100) NOT NULL DEFAULT '0' COMMENT '计划ID',\n  `team_id` varchar(100) NOT NULL COMMENT '团队ID',\n  `rank_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '序号ID',\n  `name` varchar(255) NOT NULL COMMENT '计划名称',\n  `task_type` tinyint(2) NOT NULL DEFAULT '0' COMMENT '计划类型：1-普通任务，2-定时任务',\n  `create_user_id` varchar(100) NOT NULL DEFAULT '0' COMMENT '创建人id',\n  `head_user_id` varchar(1000) NOT NULL DEFAULT '0' COMMENT '负责人id ,用分割',\n  `run_count` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '运行次数',\n  `init_strategy` tinyint(2) NOT NULL DEFAULT '1' COMMENT '初始化策略：1-计划执行前重启浏览器，2-场景执行前重启浏览器，3-无初始化',\n  `description` text NOT NULL COMMENT '备注',\n  `browsers` json DEFAULT NULL COMMENT '浏览器信息',\n  `ui_machine_key` varchar(255) NOT NULL DEFAULT '' COMMENT '指定机器key',\n  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',\n  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',\n  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',\n  PRIMARY KEY (`id`),\n  KEY `idx_plan_id` (`plan_id`),\n  KEY `idx_team_id` (`team_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='UI计划表';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `ui_plan_report` (\n  `id` bigint(20) NOT NULL AUTO_INCREMENT,\n  `report_id` varchar(100) NOT NULL COMMENT '报告ID',\n  `report_name` varchar(125) NOT NULL COMMENT '报告名称',\n  `plan_id` varchar(100) NOT NULL COMMENT '计划ID',\n  `plan_name` varchar(255) NOT NULL COMMENT '计划名称',\n  `team_id` varchar(100) NOT NULL COMMENT '团队ID',\n  `rank_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '序号ID',\n  `task_type` int(11) NOT NULL DEFAULT '0' COMMENT '任务类型',\n  `scene_run_order` tinyint(2) NOT NULL DEFAULT '1' COMMENT '场景运行次序：1-顺序执行，2-同时执行',\n  `run_duration_time` bigint(20) NOT NULL DEFAULT '0' COMMENT '任务运行持续时长',\n  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '报告状态1:进行中，2:已完成',\n  `run_user_id` varchar(100) NOT NULL DEFAULT '0' COMMENT '启动人id',\n  `remark` text NOT NULL COMMENT '备注',\n  `browsers` json DEFAULT NULL COMMENT '浏览器信息',\n  `ui_machine_key` varchar(255) NOT NULL DEFAULT '' COMMENT '指定机器key',\n  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间（执行时间）',\n  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',\n  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',\n  PRIMARY KEY (`id`),\n  KEY `idx_report_id` (`report_id`),\n  KEY `idx_team_id` (`team_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='UI自动化测试计划-报告表';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `ui_plan_task_conf` (\n  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '配置ID',\n  `plan_id` varchar(100) NOT NULL DEFAULT '0' COMMENT '计划ID',\n  `team_id` varchar(100) NOT NULL COMMENT '团队ID',\n  `task_type` tinyint(2) NOT NULL DEFAULT '0' COMMENT '任务类型：1-普通模式，2-定时任务',\n  `scene_run_order` tinyint(2) NOT NULL DEFAULT '1' COMMENT '场景运行次序：1-顺序执行，2-同时执行',\n  `run_user_id` varchar(100) NOT NULL DEFAULT '0' COMMENT '运行人用户ID',\n  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',\n  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',\n  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',\n  PRIMARY KEY (`id`),\n  KEY `idx_plan_id` (`plan_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='UI自动化测试—普通任务配置表';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `ui_plan_timed_task_conf` (\n  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '表id',\n  `plan_id` varchar(100) NOT NULL DEFAULT '0' COMMENT '计划id',\n  `team_id` varchar(100) NOT NULL COMMENT '团队id',\n  `frequency` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '任务执行频次: 0-一次，1-每天，2-每周，3-每月，4-固定时间间隔',\n  `task_exec_time` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '任务执行时间',\n  `task_close_time` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '任务结束时间',\n  `fixed_interval_start_time` bigint(20) NOT NULL DEFAULT '0' COMMENT '固定时间间隔开始时间',\n  `fixed_interval_time` int(10) NOT NULL DEFAULT '0' COMMENT '固定间隔时间',\n  `fixed_run_num` int(10) NOT NULL DEFAULT '0' COMMENT '固定执行次数',\n  `fixed_interval_time_type` int(10) NOT NULL DEFAULT '0' COMMENT '固定间隔时间类型：0-分钟，1-小时',\n  `task_type` tinyint(2) NOT NULL DEFAULT '2' COMMENT '任务类型：1-普通任务，2-定时任务',\n  `scene_run_order` tinyint(2) NOT NULL DEFAULT '1' COMMENT '场景运行次序：1-顺序执行，2-同时执行',\n  `status` tinyint(2) NOT NULL DEFAULT '0' COMMENT '任务状态：0-未启用，1-运行中，2-已过期',\n  `run_user_id` varchar(100) NOT NULL DEFAULT '0' COMMENT '运行人用户ID',\n  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',\n  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',\n  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',\n  PRIMARY KEY (`id`),\n  KEY `idx_plan_id` (`plan_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='UI自动化测试-定时任务配置表';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `ui_scene` (\n  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'id',\n  `scene_id` varchar(100) NOT NULL COMMENT '全局唯一ID',\n  `scene_type` varchar(10) NOT NULL COMMENT '类型：文件夹，场景',\n  `team_id` varchar(100) NOT NULL COMMENT '团队id',\n  `name` varchar(255) NOT NULL COMMENT '名称',\n  `parent_id` varchar(100) NOT NULL DEFAULT '0' COMMENT '父级ID',\n  `sort` int(11) NOT NULL DEFAULT '0' COMMENT '排序',\n  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '回收站状态：1-正常，2-回收站',\n  `version` int(11) NOT NULL DEFAULT '0' COMMENT '产品版本号',\n  `source` tinyint(2) NOT NULL DEFAULT '1' COMMENT '数据来源：1-场景管理，2-计划',\n  `plan_id` varchar(255) NOT NULL DEFAULT '' COMMENT '计划ID',\n  `created_user_id` varchar(100) NOT NULL COMMENT '创建人ID',\n  `recent_user_id` varchar(100) NOT NULL COMMENT '最近修改人ID',\n  `description` text NOT NULL COMMENT '备注',\n  `ui_machine_key` varchar(255) NOT NULL DEFAULT '' COMMENT '指定执行的UI自动化机器key',\n  `source_id` varchar(100) NOT NULL COMMENT '引用来源ID',\n  `browsers` json DEFAULT NULL COMMENT '浏览器信息',\n  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',\n  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',\n  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',\n  PRIMARY KEY (`id`),\n  KEY `idx_scene_id` (`scene_id`),\n  KEY `idx_team_id` (`team_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='UI自动化场景';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `ui_scene_element` (\n  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'id',\n  `scene_id` varchar(100) NOT NULL COMMENT '场景ID',\n  `operator_id` varchar(100) NOT NULL COMMENT '操作ID',\n  `element_id` varchar(100) NOT NULL COMMENT '元素ID',\n  `team_id` varchar(100) NOT NULL COMMENT '团队id',\n  `status` tinyint(2) NOT NULL DEFAULT '1' COMMENT '状态 1：正常  2：回收站',\n  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',\n  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',\n  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',\n  PRIMARY KEY (`id`),\n  KEY `idx_scene_id` (`scene_id`),\n  KEY `idx_team_id` (`team_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='UI自动化场景元素关联表';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `ui_scene_operator` (\n  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'id',\n  `operator_id` varchar(100) NOT NULL COMMENT '全局唯一ID',\n  `scene_id` varchar(100) NOT NULL COMMENT '场景ID',\n  `name` varchar(255) NOT NULL COMMENT '名称',\n  `parent_id` varchar(100) NOT NULL DEFAULT '0' COMMENT '父级ID',\n  `sort` int(11) NOT NULL DEFAULT '0' COMMENT '排序',\n  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '状态：1-正常，2-禁用',\n  `type` varchar(100) NOT NULL DEFAULT '' COMMENT '步骤类型',\n  `action` varchar(100) NOT NULL DEFAULT '' COMMENT '步骤方法',\n  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',\n  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',\n  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',\n  PRIMARY KEY (`id`),\n  KEY `idx_scene_id` (`scene_id`),\n  KEY `idx_operator_id` (`operator_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='UI自动化场景步骤';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `ui_scene_sync` (\n  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',\n  `scene_id` varchar(100) NOT NULL COMMENT '场景ID',\n  `source_scene_id` varchar(100) NOT NULL COMMENT '引用场景ID',\n  `team_id` varchar(100) NOT NULL COMMENT '团队id',\n  `sync_mode` tinyint(2) NOT NULL DEFAULT '0' COMMENT '状态：1-实时，2-手动,已场景为准   3-手动,已计划为准',\n  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',\n  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',\n  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',\n  PRIMARY KEY (`id`),\n  KEY `idx_scene_id` (`scene_id`),\n  KEY `idx_team_id` (`team_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='UI场景同步关系表';")
	initTableSql = append(initTableSql, "CREATE TABLE IF NOT EXISTS `ui_scene_trash` (\n  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'id',\n  `scene_id` varchar(100) NOT NULL COMMENT '场景ID',\n  `team_id` varchar(100) NOT NULL COMMENT '团队id',\n  `created_user_id` varchar(100) NOT NULL COMMENT '创建人ID',\n  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',\n  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',\n  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',\n  PRIMARY KEY (`id`),\n  KEY `idx_scene_id` (`scene_id`),\n  KEY `idx_team_id` (`team_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='UI自动化场景回收站';")

	// 执行sql语句
	for _, sqlInfo := range initTableSql {
		// 执行SQL语句
		if strings.TrimSpace(sqlInfo) == "" {
			continue
		}
		_, err := db.Exec(sqlInfo)
		if err != nil {
			log.Logger.Error("执行初始化表结构sql语句失败，sql:", sqlInfo, " err:", err)
			continue
		}
	}
	return
}

func ExecMysqlContent(db *sql.DB) {
	ctx := context.Background()

	// 是否是新启动的项目
	var isRegister bool

	c := dal.GetQuery().Company
	_, err := c.WithContext(ctx).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		isRegister = true
		// 初始化企业数据
		if err = initCompanyData(ctx); err != nil {
			log.Logger.Error("initCompanyData：", err)
		}
	}

	// 处理权限数据
	handlePermissionData(ctx, db, isRegister)

	// 处理三方通知渠道
	handleNoticeData(ctx, db)
}

// initCompanyData 初始化企业数据
func initCompanyData(ctx context.Context) error {
	account := conf.Conf.CompanyInitConfig.Account
	password := conf.Conf.CompanyInitConfig.Password

	hashedPassword, err := omnibus.GenerateBcryptFromPassword(password)
	if err != nil {
		log.Logger.Error("omnibus.GenerateBcryptFromPassword error：", err)
		return err
	}

	rand.Seed(time.Now().UnixNano())
	user := model.User{
		UserID:   uuid.GetUUID(),
		Email:    "",
		Account:  account,
		Password: hashedPassword,
		Nickname: account,
		Avatar:   consts.DefaultAvatarMemo[rand.Intn(3)],
	}

	cSuperUUID := uuid.GetUUID()
	cManagerUUID := uuid.GetUUID()
	cGeneralUUID := uuid.GetUUID()
	tSuperUUID := uuid.GetUUID()
	tManagerUUID := uuid.GetUUID()

	return query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		// step1: 生成用户
		if err = tx.User.WithContext(ctx).Create(&user); err != nil {
			log.Logger.Error("step1 error：", err)
			return err
		}

		// step2: 生成企业
		company := model.Company{
			CompanyID: uuid.GetUUID(),
			Name:      conf.Conf.CompanyInitConfig.Name,
		}
		if err = tx.Company.WithContext(ctx).Create(&company); err != nil {
			log.Logger.Error("step2 error：", err)
			return err
		}

		// step3: 维护企业用户
		if err = tx.UserCompany.WithContext(ctx).Create(&model.UserCompany{
			UserID:     user.UserID,
			CompanyID:  company.CompanyID,
			InviteTime: time.Now(),
		}); err != nil {
			log.Logger.Error("step3 error：", err)
			return err
		}

		// step4: 创建企业默认角色
		var roles = []*model.Role{
			{
				RoleID:    cSuperUUID,
				RoleType:  consts.RoleTypeCompany,
				Name:      "超管",
				CompanyID: company.CompanyID,
				Level:     consts.RoleLevelSuperManager,
				IsDefault: consts.RoleDefault,
			},
			{
				RoleID:    cManagerUUID,
				RoleType:  consts.RoleTypeCompany,
				Name:      "管理员",
				CompanyID: company.CompanyID,
				Level:     consts.RoleLevelManager,
				IsDefault: consts.RoleDefault,
			},
			{
				RoleID:    cGeneralUUID,
				RoleType:  consts.RoleTypeCompany,
				Name:      "普通成员",
				CompanyID: company.CompanyID,
				Level:     consts.RoleLevelGeneral,
				IsDefault: consts.RoleDefault,
			},
			{
				RoleID:    tSuperUUID,
				RoleType:  consts.RoleTypeTeam,
				Name:      "团队管理员",
				CompanyID: company.CompanyID,
				Level:     consts.RoleLevelSuperManager,
				IsDefault: consts.RoleDefault,
			},
			{
				RoleID:    tManagerUUID,
				RoleType:  consts.RoleTypeTeam,
				Name:      "团队成员",
				CompanyID: company.CompanyID,
				Level:     consts.RoleLevelManager,
				IsDefault: consts.RoleDefault,
			},
		}
		if err = tx.Role.WithContext(ctx).Create(roles...); err != nil {
			log.Logger.Error("step4 error：", err)
			return err
		}

		// step5: 当前用户关联企业角色
		userRoleInfo := model.UserRole{
			RoleID:     cSuperUUID,
			UserID:     user.UserID,
			CompanyID:  company.CompanyID,
			InviteTime: time.Now(),
		}
		if err = tx.UserRole.WithContext(ctx).Create(&userRoleInfo); err != nil {
			log.Logger.Error("step5 error：", err)
			return err
		}

		// step6: 清洗之前的数据
		u := tx.User
		users, err := u.WithContext(ctx).Where(u.Account.Eq("")).Find()
		if err != nil {
			return err
		}
		if len(users) > 0 {
			userID := user.UserID
			companyID := company.CompanyID
			roleIDC2 := cManagerUUID
			roleIDT1 := tSuperUUID
			roleIDT2 := tManagerUUID
			for _, user := range users {
				_, err = u.WithContext(ctx).Where(u.UserID.Eq(user.UserID)).Update(u.Account, u.Email)
				if err != nil {
					return err
				}
			}

			users, err = u.WithContext(ctx).Find()
			if err != nil {
				return err
			}
			userCompany := make([]*model.UserCompany, 0, len(users))
			userRole := make([]*model.UserRole, 0, len(users))
			uc := tx.UserCompany
			ur := tx.UserRole
			for _, causer := range users {
				if causer.UserID != userID {
					_, err = uc.WithContext(ctx).Where(uc.UserID.Eq(causer.UserID), uc.CompanyID.Eq(companyID)).First()
					if errors.Is(err, gorm.ErrRecordNotFound) {
						userCompany = append(userCompany, &model.UserCompany{
							UserID:       causer.UserID,
							CompanyID:    companyID,
							InviteUserID: userID,
							InviteTime:   causer.CreatedAt,
						})
					}

					_, err = ur.WithContext(ctx).Where(ur.UserID.Eq(causer.UserID), ur.CompanyID.Eq(companyID)).First()
					if errors.Is(err, gorm.ErrRecordNotFound) {
						userRole = append(userRole, &model.UserRole{
							RoleID:       roleIDC2,
							UserID:       causer.UserID,
							CompanyID:    companyID,
							InviteUserID: userID,
							InviteTime:   causer.CreatedAt,
						})
					}
				}
			}

			if len(userCompany) > 0 {
				if err = uc.WithContext(ctx).Create(userCompany...); err != nil {
					return err
				}
			}

			if len(userRole) > 0 {
				if err = ur.WithContext(ctx).Create(userRole...); err != nil {
					return err
				}
			}

			t := tx.Team
			if _, err = t.WithContext(ctx).Where(t.ID.Gte(0)).Update(t.CompanyID, companyID); err != nil {
				return err
			}

			s := tx.Setting
			ut := tx.UserTeam
			userRoleTeam := make([]*model.UserRole, 0)
			teams, err := t.WithContext(ctx).Find()
			if err != nil {
				return err
			}

			userTeamsSup := make([]*model.UserTeam, 0)
			// 超管添加之前数据
			for _, teamInfo := range teams {
				_, err = ut.WithContext(ctx).Where(ut.UserID.Eq(userID), ut.TeamID.Eq(teamInfo.TeamID)).First()
				if errors.Is(err, gorm.ErrRecordNotFound) {
					userTeamsSup = append(userTeamsSup, &model.UserTeam{
						TeamID: teamInfo.TeamID,
						UserID: userID,
						IsShow: consts.TeamIsShowFalse,
					})
				}

				_, err = ur.WithContext(ctx).Where(ur.UserID.Eq(userID), ur.TeamID.Eq(teamInfo.TeamID)).First()
				if errors.Is(err, gorm.ErrRecordNotFound) {
					userRoleTeam = append(userRoleTeam, &model.UserRole{
						RoleID:     roleIDT1,
						UserID:     userID,
						TeamID:     teamInfo.TeamID,
						InviteTime: teamInfo.CreatedAt,
					})
				}

				// 设置超管默认团队
				_, err := s.WithContext(ctx).Where(s.UserID.Eq(userID)).First()
				if errors.Is(err, gorm.ErrRecordNotFound) {
					if err = s.WithContext(ctx).Create(&model.Setting{UserID: userID, TeamID: teamInfo.TeamID}); err != nil {
						return err
					}
				}
			}

			// 之前成员添加数据
			userTeams, err := ut.WithContext(ctx).Find()
			if err != nil {
				return err
			}
			for _, tInfo := range userTeams {
				roleID := roleIDT1
				if tInfo.RoleID != 1 {
					roleID = roleIDT2
				}
				_, err = ur.WithContext(ctx).Where(ur.UserID.Eq(tInfo.UserID), ur.TeamID.Eq(tInfo.TeamID)).First()
				if errors.Is(err, gorm.ErrRecordNotFound) {
					userRoleTeam = append(userRoleTeam, &model.UserRole{
						RoleID:       roleID,
						UserID:       tInfo.UserID,
						TeamID:       tInfo.TeamID,
						InviteUserID: tInfo.InviteUserID,
						InviteTime:   tInfo.CreatedAt,
					})
				}
			}
			if len(userTeamsSup) > 0 {
				if err = ut.WithContext(ctx).Create(userTeamsSup...); err != nil {
					return err
				}
			}

			if len(userRoleTeam) > 0 {
				if err = ur.WithContext(ctx).Create(userRoleTeam...); err != nil {
					return err
				}
			}
		}

		return nil
	})
}

// handlePermissionData 处理权限数据
func handlePermissionData(ctx context.Context, db *sql.DB, isRegister bool) {
	var currentPermissionCount int64 = 21

	p := dal.GetQuery().Permission
	count, _ := p.WithContext(ctx).Where(p.PermissionID.Gte(0)).Count()
	if count != currentPermissionCount {
		log.Logger.Info("start reset permission")
		tx, err := db.Begin()
		if err != nil {
			if tx != nil {
				_ = tx.Rollback() // 回滚
			}
			log.Logger.Error("begin trans failed,：", err)
			return
		}
		// 清空权限表  清空权限类型表
		_, err = tx.Exec("TRUNCATE TABLE `permission`")
		if err != nil {
			_ = tx.Rollback()
			log.Logger.Error("清空权限表数据：", err)
		}

		_, err = tx.Exec("TRUNCATE TABLE `role_type_permission`")
		if err != nil {
			_ = tx.Rollback()
			log.Logger.Error("清空权限表数据：", err)
		}

		_, err = tx.Exec("INSERT INTO `permission` (`permission_id`, `title`, `mark`, `url`, `type`, `group_id`, `created_at`, `updated_at`, `deleted_at`)\nVALUES\n\t(101, '创建成员', 'company_save_member', '/permission/api/v1/company/member/save', 1, 1, '2023-05-22 10:31:54', '2023-05-22 14:48:32', NULL),\n\t(102, '批量导入成员', 'company_export_member', '/permission/api/v1/company/member/export', 1, 1, '2023-05-22 10:33:42', '2023-05-24 14:17:06', NULL),\n\t(103, '编辑成员', 'company_update_member', '/permission/api/v1/company/member/update', 1, 1, '2023-05-22 10:33:42', '2023-05-22 14:48:35', NULL),\n\t(104, '删除成员', 'company_remove_member', '/permission/api/v1/company/member/remove', 1, 1, '2023-05-22 10:33:42', '2023-05-22 14:48:37', NULL),\n\t(105, '更改企业角色', 'company_set_role_member', '/permission/api/v1/role/company/set', 1, 1, '2023-05-22 10:33:42', '2023-05-22 16:39:34', NULL),\n\t(201, '新建团队', 'team_save', '/permission/api/v1/team/save', 1, 2, '2023-05-22 10:35:38', '2023-05-22 14:48:40', NULL),\n\t(202, '编辑团队', 'team_update', '/permission/api/v1/team/update', 1, 2, '2023-05-22 10:35:38', '2023-05-22 14:48:41', NULL),\n\t(203, '添加团队成员', 'team_save_member', '/permission/api/v1/team/member/save', 1, 2, '2023-05-22 10:35:38', '2023-05-22 14:48:42', NULL),\n\t(204, '移除团队成员', 'team_remove_member', '/permission/api/v1/team/member/remove', 1, 2, '2023-05-22 10:35:38', '2023-05-22 14:48:44', NULL),\n\t(205, '更改团队角色', 'team_set_role_member', '/permission/api/v1/role/team/set', 1, 2, '2023-05-22 10:35:38', '2023-05-29 14:42:11', NULL),\n\t(206, '解散团队', 'team_disband', '/permission/api/v1/team/disband', 1, 2, '2023-05-22 10:35:38', '2023-05-22 14:48:46', NULL),\n\t(301, '新建角色', 'role_save', '/permission/api/v1/role/save', 1, 3, '2023-05-22 10:36:40', '2023-05-22 14:48:47', NULL),\n\t(302, '设置角色权限', 'role_set', '/permission/api/v1/permission/role/set', 1, 3, '2023-05-22 10:36:40', '2023-05-22 14:48:51', NULL),\n\t(303, '删除角色', 'role_remove', '/permission/api/v1/role/remove', 1, 3, '2023-05-22 10:36:40', '2023-05-22 14:48:54', NULL),\n\t(401, '新建第三方集成', 'notice_save', '/permission/api/v1/notice/save', 1, 4, '2023-07-12 16:56:47', '2023-07-12 16:56:47', NULL),\n\t(402, '修改第三方集成', 'notice_update', '/permission/api/v1/notice/update', 1, 4, '2023-07-12 16:56:47', '2023-07-12 16:56:47', NULL),\n\t(403, '禁用|启用第三方集成', 'notice_set_status', '/permission/api/v1/notice/set_status', 1, 4, '2023-07-12 16:56:47', '2023-07-12 16:56:47', NULL),\n\t(404, '删除第三方集成', 'notice_remove', '/permission/api/v1/notice/remove', 1, 4, '2023-07-12 16:56:47', '2023-07-12 16:56:47', NULL),\n\t(405, '新建通知组', 'notice_group_save', '/permission/api/v1/notice/group/save', 1, 4, '2023-07-12 16:56:47', '2023-07-12 16:56:47', NULL),\n\t(406, '修改通知组', 'notice_group_update', '/permission/api/v1/notice/group/update', 1, 4, '2023-07-12 16:56:47', '2023-07-12 16:56:47', NULL),\n\t(407, '删除通知组', 'notice_group_remove', '/permission/api/v1/notice/group/remove', 1, 4, '2023-07-12 16:56:47', '2023-07-12 16:56:47', NULL);")
		if err != nil {
			_ = tx.Rollback()
			log.Logger.Error("添加权限数据失败：", err)
		}

		_, err = tx.Exec("INSERT INTO `role_type_permission` (`role_type`, `permission_id`, `created_at`, `updated_at`, `deleted_at`)\nVALUES\n\t(1, 101, '2023-05-22 17:13:31', '2023-05-22 17:14:05', NULL),\n\t(1, 102, '2023-05-22 17:16:00', '2023-05-22 17:16:00', NULL),\n\t(1, 103, '2023-05-22 17:16:00', '2023-05-22 17:16:00', NULL),\n\t(1, 104, '2023-05-22 17:16:00', '2023-05-22 17:16:00', NULL),\n\t(1, 105, '2023-05-22 17:16:00', '2023-05-22 17:16:00', NULL),\n\t(1, 201, '2023-05-22 17:16:00', '2023-05-22 17:16:00', NULL),\n\t(1, 202, '2023-05-22 17:16:00', '2023-05-22 17:16:00', NULL),\n\t(1, 203, '2023-05-22 17:16:00', '2023-05-22 17:16:00', NULL),\n\t(1, 204, '2023-05-22 17:16:00', '2023-05-22 17:16:00', NULL),\n\t(1, 205, '2023-05-22 17:16:00', '2023-05-22 17:16:00', NULL),\n  (1, 206, '2023-05-25 18:57:51', '2023-05-25 18:57:51', NULL),\n\t(1, 301, '2023-05-22 17:16:00', '2023-05-22 17:16:00', NULL),\n\t(1, 302, '2023-05-22 17:16:00', '2023-05-22 17:16:00', NULL),\n\t(1, 303, '2023-05-22 17:16:00', '2023-05-22 17:16:00', NULL),\n\t(2, 202, '2023-05-22 17:16:00', '2023-05-22 17:16:00', NULL),\n\t(2, 203, '2023-05-22 17:16:01', '2023-05-22 17:16:01', NULL),\n  (2, 204, '2023-05-24 15:22:43', '2023-05-24 15:22:43', NULL),\n\t(2, 205, '2023-05-22 17:16:01', '2023-05-22 17:16:01', NULL),\n\t(2, 206, '2023-05-22 17:16:01', '2023-05-22 17:16:01', NULL),\n\t(1, 401, '2023-07-12 17:37:50', '2023-07-12 17:37:50', NULL),\n\t(1, 402, '2023-07-12 17:37:50', '2023-07-12 17:37:50', NULL),\n\t(1, 403, '2023-07-12 17:37:50', '2023-07-12 17:37:50', NULL),\n\t(1, 404, '2023-07-12 17:37:50', '2023-07-12 17:37:50', NULL),\n\t(1, 405, '2023-07-12 17:37:50', '2023-07-12 17:37:50', NULL),\n  (1, 406, '2023-07-12 17:37:50', '2023-07-12 17:37:50', NULL),\n\t(1, 407, '2023-07-12 17:37:50', '2023-07-12 17:37:50', NULL);")
		if err != nil {
			_ = tx.Rollback()
			log.Logger.Error("添加权限类型数据失败：", err)
		}

		// 添加不能修改的默认权限
		rtp := dal.GetQuery().RoleTypePermission
		companyRtpList, err := rtp.WithContext(ctx).Where(rtp.RoleType.Eq(consts.RoleTypeCompany)).Find()
		if err != nil {
			log.Logger.Error("查询权限类型数据失败 company：", err)
		}

		teamRtpList, err := rtp.WithContext(ctx).Where(rtp.RoleType.Eq(consts.RoleTypeTeam)).Find()
		if err != nil {
			log.Logger.Error("查询权限类型数据失败 team：", err)
		}

		companyPIDs := make([]int64, 0, len(companyRtpList))
		teamPIDs := make([]int64, 0, len(teamRtpList))

		for _, cp := range companyRtpList {
			companyPIDs = append(companyPIDs, cp.PermissionID)
		}

		for _, tp := range teamRtpList {
			teamPIDs = append(teamPIDs, tp.PermissionID)
		}

		r := dal.GetQuery().Role
		roles, err := r.WithContext(ctx).Where(r.IsDefault.Eq(1)).Find()
		if err != nil {
			log.Logger.Error("查询角色失败：", err)
		}

		var (
			values       []interface{}
			placeholders []string
		)

		for _, r := range roles {
			if r.RoleType == consts.RoleTypeCompany && r.Level <= consts.RoleLevelManager {
				for _, cp := range companyPIDs {
					placeholders = append(placeholders, "(?,?)")
					values = append(values, r.RoleID, cp)
				}
				_, err = tx.Exec("delete from role_permission where role_id = ?", r.RoleID)
				if err != nil {
					_ = tx.Rollback()
					log.Logger.Error("删除 role_permission：", err)
				}
			}

			if r.RoleType == consts.RoleTypeTeam {
				// 团队管理员默认全部权限
				if r.Level == consts.RoleLevelSuperManager {
					for _, tp := range teamPIDs {
						placeholders = append(placeholders, "(?,?)")
						values = append(values, r.RoleID, tp)
					}
					_, err = tx.Exec("delete from role_permission where role_id = ?", r.RoleID)
					if err != nil {
						_ = tx.Rollback()
						log.Logger.Error("删除 role_permission：", err)
					}
				}
				// 初始注册
				if isRegister {
					// 团队成员默认指定权限
					if r.Level == consts.RoleLevelManager {
						var permissionIDs = []int64{203}
						for _, pID := range permissionIDs {
							placeholders = append(placeholders, "(?,?)")
							values = append(values, r.RoleID, pID)
						}
					}
				}
			}
		}

		rolePermissionSql := "INSERT INTO role_permission (role_id, permission_id) VALUES "
		rolePermissionSql += strings.Join(placeholders, ",")
		_, err = tx.Exec(rolePermissionSql, values...)
		if err != nil {
			log.Logger.Error("add role_permission：", err)
			_ = tx.Rollback()
		}

		_ = tx.Commit()
	}

	return
}

// handleNoticeData 处理通知数据
func handleNoticeData(ctx context.Context, db *sql.DB) {
	var currentNoticeChannelCount int64 = 7

	tnc := dal.GetQuery().ThirdNoticeChannel
	count, _ := tnc.WithContext(ctx).Where(tnc.ID.Gte(0)).Count()
	if count != currentNoticeChannelCount {
		_, err := db.Exec("INSERT INTO `third_notice_channel` (`name`, `type`, `created_at`, `updated_at`, `deleted_at`)\nVALUES\n    ('飞书群机器人', 1, '2023-06-21 10:46:03', '2023-06-21 10:46:03', NULL),\n    ('飞书企业应用', 1, '2023-06-21 10:46:25', '2023-06-21 10:46:25', NULL),\n    ('企业微信应用', 2, '2023-06-21 10:46:39', '2023-06-21 10:46:53', NULL),\n    ('企业微信机器人', 2, '2023-06-21 10:47:08', '2023-06-21 10:47:08', NULL),\n    ('邮箱', 3, '2023-06-29 11:03:45', '2023-06-29 11:03:45', NULL),\n    ('钉钉群机器人', 4, '2023-06-29 11:03:55', '2023-06-29 11:04:00', NULL),\n    ('钉钉企业应用', 4, '2023-06-29 11:04:13', '2023-06-29 11:04:13', NULL);")
		if err != nil {
			log.Logger.Error("添加三方通知渠道失败：", err)
		}
	}

	return
}
