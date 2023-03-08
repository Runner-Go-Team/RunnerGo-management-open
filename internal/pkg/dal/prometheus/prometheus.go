package prometheus

type Response struct {
	Status string        `json:"status"`
	Data   *ResponseData `json:"data"`
}

type ResponseData struct {
	ResultType string            `json:"resultType"`
	Result     []*ResponseResult `json:"result"`
}

type ResponseResult struct {
	Metric *ResponseMetric `json:"metric"`
	Value  []interface{}   `json:"value"`
	Values [][]interface{} `json:"values"`
}

type ResponseMetric struct {
	Instance string `json:"instance"`
	Job      string `json:"job"`
}
