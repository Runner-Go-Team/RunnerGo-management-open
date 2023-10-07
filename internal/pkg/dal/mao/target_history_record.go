package mao

import "time"

type TargetHistoryRecord struct {
	TargetID  string    `bson:"target_id"`
	TeamID    string    `bson:"team_id"`
	UserID    string    `bson:"user_id"`
	Detail    string    `bson:"detail"`
	Uuid      string    `bson:"uuid"`
	Hash      string    `bson:"hash"`
	IsSaveTag bool      `bson:"is_save_tag"`
	CreatedAt time.Time `bson:"created_at"`
}
