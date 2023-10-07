package prometheus

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/valyala/fasthttp"

	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/conf"
)

func GetMemRangeUsage(ip string, s, e int64) ([][]interface{}, error) {
	u := url.URL{
		Scheme:   "http",
		Host:     fmt.Sprintf("%s:%d", conf.Conf.Prometheus.Host, conf.Conf.Prometheus.Port),
		Path:     "/api/v1/query_range",
		RawQuery: fmt.Sprintf("start=%d&end=%d&step=15&query=(node_memory_MemTotal_bytes-node_memory_MemAvailable_bytes)/node_memory_MemTotal_bytes{instance=\"%s:9100\"}", s, e, ip),
	}

	uu := u.String()
	statusCode, body, err := fasthttp.Get(nil, uu)
	if err != nil {

	}
	if statusCode != http.StatusOK {

	}

	var resp Response
	if err := json.Unmarshal(body, &resp); err != nil {

	}

	return resp.Data.Result[0].Values, nil

}
