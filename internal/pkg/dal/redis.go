package dal

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"strings"

	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/conf"
)

var rdb *redis.ClusterClient

func MustInitRedis() {
	fmt.Println("redis initialized")
	rdb = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    strings.Split(conf.Conf.Redis.ClusterAddress, ";"),
		Password: conf.Conf.Redis.Password,
	})
}

func GetRDB() *redis.ClusterClient {
	return rdb
}
