package mao

type APIDebug struct {
	ApiID                 string                   `bson:"api_id"`
	APIName               string                   `bson:"api_name"`
	Assertion             []*Assertion             `bson:"assertion"`
	EventID               string                   `bson:"event_id"`
	Regex                 []map[string]interface{} `bson:"regex"`
	RequestBody           string                   `bson:"request_body"`
	RequestCode           int64                    `bson:"request_code"`
	RequestHeader         string                   `bson:"request_header"`
	RequestTime           int64                    `bson:"request_time"`
	ResponseBody          string                   `bson:"response_body"`
	ResponseBytes         float64                  `bson:"response_bytes"`
	ResponseHeader        string                   `bson:"response_header"`
	ResponseTime          string                   `bson:"response_time"`
	ResponseLen           int32                    `bson:"response_len"`
	ResponseStatusMessage string                   `bson:"response_status_message"`
	UUID                  string                   `bson:"uuid"`
}

type Assertion struct {
	Code      int    `bson:"code"`
	IsSucceed bool   `bson:"isSucceed"`
	Msg       string `bson:"msg"`
}
