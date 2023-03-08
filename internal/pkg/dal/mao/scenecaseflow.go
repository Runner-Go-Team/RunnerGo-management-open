package mao

import (
	"go.mongodb.org/mongo-driver/bson"

	"kp-management/internal/pkg/dal/rao"
)

type SceneCaseFlow struct {
	SceneID     string `bson:"scene_id"`
	SceneCaseID string `bson:"scene_case_id"`
	TeamID      string `bson:"team_id"`
	Version     int32  `bson:"version"`
	//Flows   string `bson:"flows"`
	Nodes bson.Raw `bson:"nodes"`
	Edges bson.Raw `bson:"edges"`
	//MultiLevelNodes bson.Raw `bson:"multi_level_nodes"`
}

type SceneCaseFlowNode struct {
	Nodes []*rao.Node `bson:"nodes"`
}

type SceneCaseFlowEdge struct {
	Edges []*rao.Edge `bson:"edges"`
}
