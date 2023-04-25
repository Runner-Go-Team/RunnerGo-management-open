package mao

import (
	"go.mongodb.org/mongo-driver/bson"

	"kp-management/internal/pkg/dal/rao"
)

type Flow struct {
	SceneID string `bson:"scene_id"`
	TeamID  string `bson:"team_id"`
	Version int32  `bson:"version"`
	//Flows   string `bson:"flows"`
	Nodes bson.Raw `bson:"nodes"`
	Edges bson.Raw `bson:"edges"`
	//MultiLevelNodes bson.Raw `bson:"multi_level_nodes"`
}

type Node struct {
	Nodes []rao.Node `bson:"nodes"`
}

type Edge struct {
	Edges []rao.Edge `bson:"edges"`
}
