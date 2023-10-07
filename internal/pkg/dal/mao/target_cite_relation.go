package mao

type TargetCiteRelation struct {
	SceneID  string `bson:"scene_id" json:"scene_id,omitempty"`
	CaseID   string `bson:"case_id" json:"case_id"`
	TargetID string `bson:"target_id" json:"target_id,omitempty"`
	NodeID   string `bson:"node_id" json:"node_id,omitempty"`
	PlanID   string `bson:"plan_id" json:"plan_id,omitempty"`
	TeamID   string `bson:"team_id" json:"team_id"`
	Source   int32  `bson:"source" json:"source"`
}
