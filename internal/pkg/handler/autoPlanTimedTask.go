package handler

import (
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/log"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"golang.org/x/net/context"
	"gorm.io/gen"
	"time"
)

func AutoPlanTimedTaskExec() {
	// 开启定时任务轮询
	for {
		ctx := context.Background()
		tx := dal.GetQuery().AutoPlanTimedTaskConf
		// 组装查询条件
		conditions := make([]gen.Condition, 0)
		conditions = append(conditions, tx.Status.Eq(consts.TimedTaskInExec))
		// 从数据库当中，查出当前需要执行的定时任务
		timedTaskData, err := tx.WithContext(ctx).Select(tx.TeamID, tx.PlanID,
			tx.TaskExecTime, tx.TaskCloseTime, tx.Frequency, tx.RunUserID).Where(conditions...).Find()

		if err != nil {
			log.Logger.Info("自动化测试--定时任务查询数据库出错，err：", err)
			continue
		}

		if len(timedTaskData) == 0 {
			// 睡眠一分钟，再循环执行
			time.Sleep(60 * time.Second)
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

		log.Logger.Info("自动化测试--定时任务查到了数据：", timedTaskData)
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
				_, err := tx.WithContext(ctx).Where(tx.PlanID.Eq(timedTaskInfo.PlanID)).
					UpdateColumn(tx.Status, consts.TimedTaskWaitEnable)
				if err != nil {
					log.Logger.Info("自动化测试--定时任务过期状态修改失败，err：", err)
				}
				log.Logger.Info("自动化测试--定时任务设置为过期：", timedTaskInfo.TaskCloseTime, " 当前时间：", nowTime)
				continue
			}

			// 根据不同的任务频次，进行不同的运行逻辑
			switch timedTaskInfo.Frequency {
			case 0: // 一次
				if taskYear != nowYear || taskMonth != nowMonth || taskDay != nowDay || taskHour != nowHour || taskMinute != nowMinute {
					continue
				}
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
			timedTaskKey := "AutoPlanTimeTaskRun:" + fmt.Sprintf("%s:%s", timedTaskInfo.TeamID, timedTaskInfo.PlanID)
			setRedisRes := dal.GetRDB().SetNX(ctx, timedTaskKey, 1, time.Second*180)
			if setRedisRes.Val() == false {
				continue
			}

			// 执行定时任务计划
			err := runAutoTimedTask(ctx, timedTaskInfo)
			if err != nil {
				log.Logger.Info("自动化测试--定时任务运行失败，任务信息：", timedTaskInfo, " err：", err)
			}
		}

		// 睡眠一分钟，再循环执行
		time.Sleep(60 * time.Second)
	}
}

func runAutoTimedTask(ctx context.Context, timedTaskInfo *model.AutoPlanTimedTaskConf) error {
	// 开始执行计划
	runAutoPlanParam := RunAutoPlanReq{
		PlanID:  timedTaskInfo.PlanID,
		TeamID:  timedTaskInfo.TeamID,
		UserID:  timedTaskInfo.RunUserID,
		RunType: 2,
	}
	log.Logger.Info("自动化测试--定时任务开始执行计划，参数：", runAutoPlanParam)
	// 进入执行计划方法
	_, runErr := RunAutoPlanDetail(ctx, runAutoPlanParam)
	log.Logger.Info("自动化测试定时任务--执行结果，runErr：", runErr)
	if runErr != nil {
		return runErr
	}
	return nil
}
