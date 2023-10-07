package machine

import (
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/log"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/gin-gonic/gin"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"time"
)

func GetMachineList(ctx *gin.Context, req rao.GetMachineListParam) ([]*rao.MachineList, int64, error) {
	// 查询机器列表
	tx := dal.GetQuery().Machine

	conditions := make([]gen.Condition, 0)
	if req.Name != "" {
		conditions = append(conditions, tx.Name.Like(fmt.Sprintf("%%%s%%", req.Name)))
	}
	if req.ServerType != 0 {
		conditions = append(conditions, tx.ServerType.Eq(req.ServerType))
	}

	timeStr := time.Now().Format("2006-01-02")
	todayZero, err := time.Parse("2006-01-02", timeStr)
	if err == nil {
		conditions = append(conditions, tx.UpdatedAt.Gte(todayZero))
	}

	// 排序
	sort := make([]field.Expr, 0, 6)
	if req.SortTag == 0 { // 默认排序(创建时间)
		sort = append(sort, tx.CreatedAt.Desc())
	}
	if req.SortTag == 1 { // CPU使用率升序
		sort = append(sort, tx.CPUUsage)
	}
	if req.SortTag == 2 { // CPU使用率降序
		sort = append(sort, tx.CPUUsage.Desc())
	}
	if req.SortTag == 3 { // 内存使用率升序
		sort = append(sort, tx.MemUsage)
	}
	if req.SortTag == 4 { // 内存使用率降序
		sort = append(sort, tx.MemUsage.Desc())
	}
	if req.SortTag == 5 { // 磁盘使用率升序
		sort = append(sort, tx.DiskUsage)
	}
	if req.SortTag == 6 { // 磁盘使用率降序
		sort = append(sort, tx.DiskUsage.Desc())
	}
	// 查询数据库
	limit := req.Size
	offset := (req.Page - 1) * req.Size
	machineList, count, err := tx.WithContext(ctx).Where(conditions...).Order(sort...).FindByPage(offset, limit)
	if err != nil {
		log.Logger.Error("机器列表--获取机器列表数据失败，err:", err)
		return nil, 0, err
	}

	res := make([]*rao.MachineList, 0, len(machineList))
	for _, machineInfo := range machineList {
		machineTmp := &rao.MachineList{
			ID:                machineInfo.ID,
			Name:              machineInfo.Name,
			CPUUsage:          machineInfo.CPUUsage,
			CPULoadOne:        machineInfo.CPULoadOne,
			CPULoadFive:       machineInfo.CPULoadFive,
			CPULoadFifteen:    machineInfo.CPULoadFifteen,
			MemUsage:          machineInfo.MemUsage,
			DiskUsage:         machineInfo.DiskUsage,
			MaxGoroutines:     machineInfo.MaxGoroutines,
			CurrentGoroutines: machineInfo.CurrentGoroutines,
			ServerType:        machineInfo.ServerType,
			Status:            machineInfo.Status,
			CreatedAt:         machineInfo.CreatedAt,
			UpdatedAt:         machineInfo.UpdatedAt,
		}
		res = append(res, machineTmp)
	}

	return res, count, nil
}

// ChangeMachineOnOff 启用或卸载机器
func ChangeMachineOnOff(ctx *gin.Context, req rao.ChangeMachineOnOff) error {
	// 查询机器列表
	tx := dal.GetQuery().Machine
	res, err := tx.WithContext(ctx).Where(tx.ID.Eq(req.ID)).Update(tx.Status, req.Status)
	if err != nil || res.RowsAffected == 0 {
		log.Logger.Error("启用或卸载机器--修改数据库失败，err:", err)
		return err
	}
	return nil
}
