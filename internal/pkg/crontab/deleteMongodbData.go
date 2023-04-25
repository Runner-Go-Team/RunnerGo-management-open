package crontab

import (
	"context"
	"fmt"
	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/bson"
	"kp-management/internal/pkg/biz/consts"
	"kp-management/internal/pkg/biz/log"
	"kp-management/internal/pkg/conf"
	"kp-management/internal/pkg/dal"
	"time"
)

func DeleteMongodbData() {

	crontab := cron.New()
	EntryID, err := crontab.AddFunc("* 3 * * *", DeleteDebugData)
	fmt.Println(time.Now(), EntryID, err)

	crontab.Start()
	time.Sleep(time.Minute * 5)

}

func DeleteDebugData() {
	ctx := context.Background()
	// 删除Api调试数据
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectAPIDebug)
	findFilter := bson.D{{}}
	_, err := collection.DeleteOne(ctx, findFilter)
	if err != nil {
		log.Logger.Info("删除操作日志--api_debug日志删除失败，err:", err)
	}
	log.Logger.Info("删除操作日志--api_debug删除成功")

	// 删除场景调试数据
	collection = dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectSceneDebug)
	findFilter = bson.D{{}}
	_, err = collection.DeleteOne(ctx, findFilter)
	if err != nil {
		log.Logger.Info("删除操作日志--scene_debug日志删除失败，err:", err)
	}
	log.Logger.Info("删除操作日志--scene_debug删除成功")

	// 删除性能压测debug日志，仅保存一个月以内的数据
	keepDataTime := conf.Conf.KeepStressDebugLogTime
	now := time.Now()                               // 获取当前时间
	oneMonthAgo := now.AddDate(0, -keepDataTime, 0) // 获取一个月前的时间
	tx := dal.GetQuery().StressPlanReport
	reportList, err := tx.WithContext(ctx).Where(tx.CreatedAt.Lt(oneMonthAgo)).Find()
	reportIDs := make([]string, 0, len(reportList))
	for _, reportInfo := range reportList {
		reportIDs = append(reportIDs, reportInfo.ReportID)
	}

	collection = dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectStressDebug)
	findFilter = bson.D{{"report_id", bson.D{{"$in", reportIDs}}}}
	_, err = collection.DeleteMany(ctx, findFilter)
	if err != nil {
		log.Logger.Info("删除debug日志--debug日志删除失败，err:", err)
	}
	log.Logger.Info("删除debug日志--debug删除成功")
}
