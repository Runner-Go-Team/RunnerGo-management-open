package mao

import "go.mongodb.org/mongo-driver/bson"

type ReportDetailData struct {
	TeamID           string   `bson:"team_id"`
	ReportID         string   `bson:"report_id"`
	ReportDetailData bson.Raw `bson:"report_detail_data"`
}
