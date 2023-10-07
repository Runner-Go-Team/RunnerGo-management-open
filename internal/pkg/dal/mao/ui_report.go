package mao

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/api/ui"
	"go.mongodb.org/mongo-driver/bson"
)

type SendReport struct {
	ReportID string   `bson:"report_id"`
	Detail   bson.Raw `bson:"detail"`
}

type SendReportDetail struct {
	Detail *ui.RunRequest `bson:"detail"`
}
