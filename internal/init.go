package internal

import (
	"go.uber.org/zap"
	"kp-management/internal/pkg/biz/log"
	"kp-management/internal/pkg/biz/proof"
	"kp-management/internal/pkg/conf"
	"kp-management/internal/pkg/dal"
)

func InitProjects(readConfMode int, configFile string) {
	if readConfMode == 1 {
		conf.MustInitConfByEnv()
	} else {
		conf.MustInitConf(configFile)
	}

	//conf.MustInitConf()
	// 初始化各种中间件
	dal.MustInitMySQL()
	dal.MustInitMongo()
	//dal.MustInitElasticSearch()
	proof.MustInitProof()
	//dal.MustInitGRPC()
	dal.MustInitRedis()
	dal.MustInitRedisForReport()
	dal.MustInitBigCache()
	// 初始化logger
	zap.S().Debug("初始化logger")
	log.InitLogger()
}
