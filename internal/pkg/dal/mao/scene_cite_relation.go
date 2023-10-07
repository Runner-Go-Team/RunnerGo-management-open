package mao

type SceneCiteRelation struct {
	OldSceneID string `bson:"old_scene_id" json:"old_scene_id,omitempty"`
	NewSceneID string `bson:"new_scene_id" json:"new_scene_id"`
	PlanID     string `bson:"plan_id" json:"plan_id,omitempty"`
	TeamID     string `bson:"team_id" json:"team_id"`
	Source     int32  `bson:"source" json:"source"`
}
