package mao

type OperationLog struct {
	TeamID      string `bson:"team_id"`
	UserID      string `bson:"user_id"`
	Category    int32  `bson:"category"`
	Operate     int32  `bson:"operate"`
	Name        string `bson:"name"`
	CreatedAt   int64  `bson:"created_time_sec"`
	CreatedDate string `bson:"created_date"`
}
