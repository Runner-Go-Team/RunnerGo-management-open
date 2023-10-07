package mao

import "time"

type TargetDebugLog struct {
	TeamID     string    `bson:"team_id"`
	TargetID   string    `bson:"target_id"`
	TargetType string    `bson:"target_type"`
	CreatedAt  time.Time `bson:"created_at"`
}
