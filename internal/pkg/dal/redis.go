package dal

import (
	"fmt"
	"github.com/go-redis/redis/v8"

	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/conf"
)

var rdb *redis.Client

func MustInitRedis() {
	fmt.Println("redis initialized")
	rdb = redis.NewClient(&redis.Options{
		Addr:     conf.Conf.Redis.Address,
		Password: conf.Conf.Redis.Password,
		DB:       conf.Conf.Redis.DB,
	})
}

func GetRDB() *redis.Client {
	return rdb
}
