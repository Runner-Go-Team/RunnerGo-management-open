package rao

type SaveGroupReq struct {
	TeamID      string `json:"team_id" binding:"required,gt=0"`
	TargetID    string `json:"target_id"`
	ParentID    string `json:"parent_id"`
	Name        string `json:"name" binding:"required,min=1"`
	Method      string `json:"method"`
	Sort        int32  `json:"sort"`
	TypeSort    int32  `json:"type_sort"`
	Version     int32  `json:"version" binding:"required,gt=0"`
	Source      int32  `json:"source"`
	PlanID      string `json:"plan_id"`
	Description string `json:"description"`
	//Request  *Request `json:"request"`
	//Script   *Script  `json:"script"`
}

type SaveGroupResp struct {
}

type GetGroupReq struct {
	TeamID   string `form:"team_id"`
	TargetID string `form:"target_id"`
	Source   int32  `form:"source" json:"source"`
}

type GetGroupResp struct {
	Group *Group `json:"group"`
}

type Group struct {
	TeamID      string   `json:"team_id"`
	TargetID    string   `json:"target_id"`
	ParentID    string   `json:"parent_id"`
	Name        string   `json:"name"`
	Method      string   `json:"method"`
	Sort        int32    `json:"sort"`
	TypeSort    int32    `json:"type_sort"`
	Version     int32    `json:"version"`
	Request     *Request `json:"request"`
	Script      *Script  `json:"script"`
	Source      int32    `json:"source"`
	PlanID      string   `json:"plan_id"`
	Description string   `json:"description"`
}
