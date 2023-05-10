package dal

import (
	"fmt"
	"github.com/go-redis/redis/v8"

	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/conf"
)

var rdbReport *redis.Client

func MustInitRedisForReport() {
	fmt.Println("redis_report initialized")
	rdbReport = redis.NewClient(&redis.Options{
		Addr:     conf.Conf.RedisReport.Address,
		Password: conf.Conf.RedisReport.Password,
		DB:       conf.Conf.RedisReport.DB,
	})
}

func GetRDBForReport() *redis.Client {
	return rdbReport
}
