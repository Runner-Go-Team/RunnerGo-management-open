package crontab

import (
	"context"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/log"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/conf"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal"
	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/bson"
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
	_, err := collection.DeleteMany(ctx, findFilter)
	if err != nil {
		log.Logger.Info("删除Api调试数据--api_debug日志删除失败，err:", err)
	}
	log.Logger.Info("删除Api调试数据--api_debug删除成功")

	// 删除场景调试数据
	collection = dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectSceneDebug)
	findFilter = bson.D{{}}
	_, err = collection.DeleteMany(ctx, findFilter)
	if err != nil {
		log.Logger.Info("删除场景调试数据--scene_debug日志删除失败，err:", err)
	}
	log.Logger.Info("删除场景调试数据--scene_debug删除成功")

	// 删除性能压测debug日志
	keepDataTime := int(conf.Conf.AboutTimeConfig.KeepStressDebugLogTime)
	now := time.Now() // 获取当前时间
	timeAgo := now.AddDate(0, 0, -keepDataTime)
	tx := dal.GetQuery().StressPlanReport
	reportList, err := tx.WithContext(ctx).Where(tx.CreatedAt.Lt(timeAgo)).Find()
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

	// 删除压力机监控数据
	deleteTime := time.Now().Unix() - (conf.Conf.AboutTimeConfig.KeepMachineMonitorDataTime * 24 * 3600)
	collection = dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectMachineMonitorData)
	_, err = collection.DeleteMany(ctx, bson.D{{"created_at", bson.D{{"$lte", deleteTime}}}})
	if err != nil {
		log.Logger.Info("压力机监控数据--删除压力机监控数据失败")
	}

	// 删除调试日志
	deleteTimeInt := time.Now().Unix() - (9 * 24 * 3600)
	// 转换为时间
	deleteTm := time.Unix(deleteTimeInt, 0)
	collection = dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectTargetDebugLog)
	_, err = collection.DeleteMany(ctx, bson.D{{"created_at", bson.D{{"$lte", deleteTm}}}})
	if err != nil {
		log.Logger.Info("调试日志--删除调试日志数据失败")
	}

	// 删除UI自动化调试数据
	collection = dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectUISendSceneOperatorDebug)
	findFilter = bson.D{{}}
	_, err = collection.DeleteMany(ctx, findFilter)
	if err != nil {
		log.Logger.Info("删除UI自动化调试数据--ui_send_scene_operator_debug日志删除失败，err:", err)
	}
	log.Logger.Info("删除UI自动化调试数据--ui_send_scene_operator_debug删除成功")

}
