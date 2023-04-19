package handler

import (
	"fmt"
	"golang.org/x/net/context"
	"gorm.io/gen"
	"kp-management/internal/pkg/biz/consts"
	"kp-management/internal/pkg/biz/log"
	"kp-management/internal/pkg/dal"
	"kp-management/internal/pkg/dal/model"
	"time"
)

func TimedTaskExec() {
	// 开启定时任务轮询
	for {
		//// 睡眠一分钟，再循环执行
		time.Sleep(60 * time.Second)

		ctx := context.Background()
		tx := dal.GetQuery().StressPlanTimedTaskConf
		// 组装查询条件
		conditions := make([]gen.Condition, 0)
		conditions = append(conditions, tx.Status.Eq(consts.TimedTaskInExec))
		// 从数据库当中，查出当前需要执行的定时任务
		timedTaskData, err := tx.WithContext(ctx).Where(conditions...).Find()

		if err != nil {
			log.Logger.Info("性能测试--定时任务查询数据库出错，err：", err)
			continue
		}

		if len(timedTaskData) == 0 {
			continue
		}

		// 当前时间的 时，分
		// 当前时间
		nowTime := time.Now().Unix()
		nowTimeInfo := time.Unix(nowTime, 0)
		nowYear := nowTimeInfo.Year()
		nowMonth := nowTimeInfo.Month()
		nowDay := nowTimeInfo.Day()
		nowHour := nowTimeInfo.Hour()
		nowMinute := nowTimeInfo.Minute()
		nowWeekday := nowTimeInfo.Weekday()

		log.Logger.Info("定时任务--查到了数据：", timedTaskData)
		// 组装运行计划参数
		for _, timedTaskInfo := range timedTaskData {
			// 获取定时任务的执行时间相关数据
			tm := time.Unix(timedTaskInfo.TaskExecTime, 0)
			taskYear := tm.Year()
			taskMonth := tm.Month()
			taskDay := tm.Day()
			taskHour := tm.Hour()
			taskMinute := tm.Minute()
			taskWeekday := tm.Weekday()

			// 排除过期的定时任务
			if timedTaskInfo.TaskCloseTime < nowTime {
				// 把当前定时任务状态变成未开始状态
				_, err := tx.WithContext(ctx).Where(tx.TeamID.Eq(timedTaskInfo.TeamID)).
					Where(tx.PlanID.Eq(timedTaskInfo.PlanID)).
					Where(tx.SceneID.Eq(timedTaskInfo.SceneID)).
					UpdateColumn(tx.Status, consts.TimedTaskWaitEnable)
				if err != nil {
					log.Logger.Info("定时任务过期状态修改失败，err：", err)
				}

				// 把当前定时计划的状态变成未运行
				planTable := dal.GetQuery().StressPlan
				_, err = planTable.WithContext(ctx).Where(planTable.TeamID.Eq(timedTaskInfo.TeamID)).
					Where(planTable.PlanID.Eq(timedTaskInfo.PlanID)).
					UpdateColumn(planTable.Status, consts.PlanStatusNormal)
				if err != nil {
					log.Logger.Info("定时计划修改为未运行状态失败，err：", err)
				}

				log.Logger.Info("定时任务--设置为过期：", timedTaskInfo.TaskCloseTime, " 当前时间：", nowTime)
				continue
			}

			// 根据不同的任务频次，进行不同的运行逻辑
			switch timedTaskInfo.Frequency {
			case 0: // 一次
				if taskYear != nowYear || taskMonth != nowMonth || taskDay != nowDay || taskHour != nowHour || taskMinute != nowMinute {
					continue
				}
				log.Logger.Info("定时任务--频次一次：通过可运行")
			case 1: // 每天
				// 比较当前时间是否等于定时任务的时间
				if taskHour != nowHour || taskMinute != nowMinute {
					continue
				}

			case 2: // 每周
				// 比较当前周几是否等于定时任务的时间
				if taskWeekday != nowWeekday || taskHour != nowHour || taskMinute != nowMinute {
					continue
				}

			case 3: // 每月
				// 比较当前每月几号是否等于定时任务的时间
				if taskDay != nowDay || taskHour != nowHour || taskMinute != nowMinute {
					continue
				}
			}

			// 给当前任务加分布式锁，防止重复执行
			timedTaskKey := "TimeTaskRun:" + fmt.Sprintf("%s", timedTaskInfo.SceneID)
			setRedisRes := dal.GetRDB().SetNX(ctx, timedTaskKey, 1, time.Second*180)
			if setRedisRes.Val() == false {
				continue
			}

			// 执行定时任务计划
			err := runTimedTask(ctx, timedTaskInfo)
			if err != nil {
				log.Logger.Info("定时任务运行失败，任务信息：", timedTaskInfo, " err：", err)
			}
		}
	}
}

func runTimedTask(ctx context.Context, timedTaskInfo *model.StressPlanTimedTaskConf) error {
	sceneIDs := make([]string, 0, 1)
	sceneIDs = append(sceneIDs, timedTaskInfo.SceneID)
	// 开始执行计划
	runStressParams := RunStressReq{
		PlanID:  timedTaskInfo.PlanID,
		TeamID:  timedTaskInfo.TeamID,
		UserID:  timedTaskInfo.UserID,
		SceneID: sceneIDs,
		RunType: 2,
	}
	log.Logger.Info("定时任务--开始执行计划，参数：", runStressParams)
	// 进入执行计划方法
	_, runErr := RunStress(ctx, runStressParams)
	log.Logger.Info("定时任务--执行结果，runErr：", runErr)
	if runErr != nil {
		return runErr
	}
	return nil
}
