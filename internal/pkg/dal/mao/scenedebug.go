package mao

type SceneDebug struct {
	ApiID          string      `bson:"api_id"`
	APIName        string      `bson:"api_name"`
	Assert         AssertObj   `bson:"assert"`
	EventID        string      `bson:"event_id"`
	NextList       []string    `bson:"next_list"`
	Regex          RegexObj    `bson:"regex"`
	RequestBody    string      `bson:"request_body"`
	RequestCode    int64       `bson:"request_code"`
	RequestHeader  string      `bson:"request_header"`
	ResponseBody   interface{} `bson:"response_body"`
	ResponseBytes  float64     `bson:"response_bytes"`
	ResponseHeader string      `bson:"response_header"`
	Status         string      `bson:"status"`
	UUID           string      `bson:"uuid"`
	ResponseTime   string      `bson:"response_time"`
	RequestType    string      `bson:"request_type"`
}
