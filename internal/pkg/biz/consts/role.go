package consts

const (
	RoleTypeCompany = 1 // 角色分类（1：企业  2：团队）
	RoleTypeTeam    = 2 // 角色分类（1：企业  2：团队）

	RoleLevelSuperManager = 1 // 超管/团队管理员  // 企业角色三级  团队角色二级
	RoleLevelManager      = 2 // 管理员/团队成员 / 自定义角色
	RoleLevelGeneral      = 3 // 自定义角色

	RoleDefault = 1 // 默认角色
	RoleCustom  = 2 // 自定义角色
)
