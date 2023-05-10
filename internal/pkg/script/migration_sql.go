package script

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

	return alterSqlMap
}
