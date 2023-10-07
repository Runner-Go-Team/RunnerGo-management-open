package mao

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/api/v1alpha1"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"go.mongodb.org/mongo-driver/bson"
)

type Mock struct {
	TargetID   string   `bson:"target_id"`
	TeamID     string   `bson:"team_id"`
	UniqueKey  string   `bson:"unique_key"`
	Path       string   `bson:"path"`
	Method     string   `bson:"method"`
	MockPath   string   `bson:"mock_path"`
	IsMockOpen int32    `bson:"is_mock_open"`
	Cases      bson.Raw `bson:"cases"`
	Expects    bson.Raw `bson:"expects"`
}

type Cases struct {
	Cases []*v1alpha1.MockAPI_Case `bson:"cases"`
}

type Expects struct {
	Expects []*rao.Expect `bson:"expects"`
}
