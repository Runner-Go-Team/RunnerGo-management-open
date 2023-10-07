package mao

import "time"

type TargetTagVersion struct {
	TagName           string    `bson:"tag_name"`
	TargetID          string    `bson:"target_id"`
	TeamID            string    `bson:"team_id"`
	UserID            string    `bson:"user_id"`
	Detail            string    `bson:"detail"`
	Uuid              string    `bson:"uuid"`
	HistoryRecordUuid string    `bson:"history_record_uuid"`
	CreatedAt         time.Time `bson:"created_at"`
}
