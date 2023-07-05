package mao

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
)

type SceneCaseFlow struct {
	SceneID     string   `bson:"scene_id"`
	SceneCaseID string   `bson:"scene_case_id"`
	TeamID      string   `bson:"team_id"`
	Version     int32    `bson:"version"`
	Nodes       bson.Raw `bson:"nodes"`
	Edges       bson.Raw `bson:"edges"`
	EnvID       int64    `bson:"env_id"`
}

type SceneCaseFlowNode struct {
	Nodes []*rao.Node `bson:"nodes"`
}

type SceneCaseFlowEdge struct {
	Edges []*rao.Edge `bson:"edges"`
}
