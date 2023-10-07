package rao

type OpenRunStressPlanReq struct {
	Account  string   `json:"account"`
	Password string   `json:"password"`
	PlanID   string   `json:"plan_id"`
	TeamID   string   `json:"team_id"`
	SceneID  []string `json:"scene_id"`
}

type OpenRunAutoPlanReq struct {
	Account  string   `json:"account"`
	Password string   `json:"password"`
	PlanID   string   `json:"plan_id"`
	TeamID   string   `json:"team_id"`
	SceneID  []string `json:"scene_id"`
}
