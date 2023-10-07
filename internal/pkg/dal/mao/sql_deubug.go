package mao

type SqlDebug struct {
	Uuid         string                   `bson:"uuid"`
	TeamID       string                   `bson:"team_id"`
	ApiID        string                   `bson:"api_id"`
	RequestTime  int64                    `bson:"request_time"`
	RequestType  string                   `bson:"request_type"`
	Status       string                   `bson:"status"`
	Assert       []map[string]interface{} `bson:"assert"`
	Regex        []map[string]interface{} `bson:"regex"`
	RequestBody  string                   `bson:"request_body"`
	ResponseBody interface{}              `bson:"response_body"`
}
