package env

import (
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/jwt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/log"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/record"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/packer"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/public"
	"github.com/gin-gonic/gin"
	"gorm.io/gen"
	"strconv"
	"strings"
	"time"

	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/query"

	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
)

func GetList(ctx *gin.Context, req *rao.EnvListReq) ([]rao.EnvListResp, error) {
	teDB := query.Use(dal.DB()).TeamEnv
	conditions := make([]gen.Condition, 0)
	conditions = append(conditions, teDB.TeamID.Eq(req.TeamID))
	if req.Name != "" {
		conditions = append(conditions, teDB.Name.Like(fmt.Sprintf("%%%s%%", req.Name)))
	}
	envListData, err := teDB.WithContext(ctx).Select(teDB.ID, teDB.Name, teDB.TeamID, teDB.CreatedUserID).Where(conditions...).
		Order(teDB.CreatedAt.Desc()).Find()
	if err != nil {
		log.Logger.Info("环境列表--获取列表失败，err:", err)
		return nil, err
	}

	var envServiceList []rao.ServiceListResp

	if len(envListData) > 0 {
		envIDs := make([]int64, 0, len(envListData))
		for _, envV := range envListData {
			envIDs = append(envIDs, envV.ID)
		}

		tesDB := query.Use(dal.DB()).TeamEnvService
		envServiceListData, _ := tesDB.WithContext(ctx).Select(tesDB.ID, tesDB.TeamEnvID, tesDB.Name, tesDB.Content).
			Where(tesDB.TeamID.Eq(req.TeamID), tesDB.TeamEnvID.In(envIDs...)).Order(tesDB.CreatedAt).Find()

		if len(envServiceListData) > 0 {
			for _, envServiceListDataV := range envServiceListData {
				envServiceList = append(envServiceList, rao.ServiceListResp{
					ID:        envServiceListDataV.ID,
					TeamEnvID: envServiceListDataV.TeamEnvID,
					Name:      envServiceListDataV.Name,
					Content:   envServiceListDataV.Content,
				})
			}
		}

	}
	return packer.TransEnvDataToRaoEnvList(envListData, envServiceList), nil
}

func GetEnvList(ctx *gin.Context, req *rao.GetEnvListReq) ([]rao.GetEnvListResp, error) {
	tx := dal.GetQuery().TeamEnv
	conditions := make([]gen.Condition, 0)
	conditions = append(conditions, tx.TeamID.Eq(req.TeamID))
	if req.Name != "" {
		conditions = append(conditions, tx.Name.Like(fmt.Sprintf("%%%s%%", req.Name)))
	}
	envList, err := tx.WithContext(ctx).Select(tx.ID, tx.TeamID, tx.Name).Where(conditions...).Find()
	if err != nil {
		return nil, err
	}
	res := make([]rao.GetEnvListResp, 0, len(envList))
	for _, envInfo := range envList {
		temp := rao.GetEnvListResp{
			EnvID:   envInfo.ID,
			EnvName: envInfo.Name,
			TeamID:  envInfo.TeamID,
		}
		res = append(res, temp)
	}
	return res, err
}

func UpdateEnv(ctx *gin.Context, req *rao.UpdateEnvReq) error {
	userID := jwt.GetUserIDByCtx(ctx)
	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		_, err := tx.TeamEnv.WithContext(ctx).Where(tx.TeamEnv.TeamID.Eq(req.TeamID),
			tx.TeamEnv.ID.Neq(req.EnvID), tx.TeamEnv.Name.Eq(req.EnvName)).First()
		if err == nil { // 存在重名
			return fmt.Errorf("名称已存在")
		}

		// 修改名称
		_, err = tx.TeamEnv.WithContext(ctx).Where(tx.TeamEnv.ID.Eq(req.EnvID)).UpdateSimple(tx.TeamEnv.Name.Value(req.EnvName),
			tx.TeamEnv.CreatedUserID.Value(userID))
		if err != nil {
			log.Logger.Info("更新环境失败，err:", err)
			return err
		}
		return record.InsertUpdate(ctx, req.TeamID, userID, record.OperationOperateUpdateEnv, req.EnvName)
	})
	return err
}

func CreateEnv(ctx *gin.Context, req *rao.CreateEnvReq) (*rao.CreateEnvResp, error) {
	userID := jwt.GetUserIDByCtx(ctx)
	var envID int64 = 0
	envName := "默认环境"
	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		// 查询老配置相关的
		list, err := tx.TeamEnv.WithContext(ctx).Where(tx.TeamEnv.TeamID.Eq(req.TeamID),
			tx.TeamEnv.Name.Like(fmt.Sprintf("%s%%", envName+"_"))).Find()
		if err == nil {
			// 有复制过得配置
			maxNum := 0
			for _, envInfo := range list {
				nameTmp := envInfo.Name
				postfixSlice := strings.Split(nameTmp, "_")
				if len(postfixSlice) < 2 {
					continue
				}
				currentNum, err := strconv.Atoi(postfixSlice[len(postfixSlice)-1])
				if err != nil {
					log.Logger.Info("新建环境--类型转换失败，err:", err)
					continue
				}
				if currentNum > maxNum {
					maxNum = currentNum
				}
			}
			envName = envName + fmt.Sprintf("_%d", maxNum+1)
		}

		nameLength := public.GetStringNum(envName)
		if nameLength > 30 { // 场景名称限制30个字符
			return fmt.Errorf("名称过长！不可超出30字符")
		}

		insertData := model.TeamEnv{
			TeamID:        req.TeamID,
			Name:          envName,
			CreatedUserID: userID,
		}
		err = tx.TeamEnv.WithContext(ctx).Save(&insertData)
		if err != nil {
			return err
		}
		envID = insertData.ID

		// 查找是否有已存在的环境
		oldEnvInfo, err := tx.TeamEnv.WithContext(ctx).Where(tx.TeamEnv.TeamID.Eq(req.TeamID)).First()
		if err == nil { // 查到了
			// 新建环境下服务
			serviceList, err := tx.TeamEnvService.WithContext(ctx).Where(tx.TeamEnvService.TeamID.Eq(req.TeamID),
				tx.TeamEnvService.TeamEnvID.Eq(oldEnvInfo.ID)).Find()
			if err != nil {
				return err
			}
			for _, serviceInfo := range serviceList {
				serviceInfo.ID = 0
				serviceInfo.TeamEnvID = envID
				serviceInfo.CreatedUserID = userID
				serviceInfo.CreatedAt = time.Now()
				serviceInfo.UpdatedAt = time.Now()
				err = tx.TeamEnvService.WithContext(ctx).Create(serviceInfo)
				if err != nil {
					log.Logger.Info("新建环境--新建环境下服务失败，err:", err)
					return err
				}
			}

			// 克隆环境下服务
			databaseList, err := tx.TeamEnvDatabase.WithContext(ctx).Where(tx.TeamEnvDatabase.TeamID.Eq(req.TeamID),
				tx.TeamEnvDatabase.TeamEnvID.Eq(oldEnvInfo.ID)).Find()
			if err != nil {
				return err
			}
			for _, databaseInfo := range databaseList {
				databaseInfo.ID = 0
				databaseInfo.TeamEnvID = envID
				databaseInfo.CreatedUserID = userID
				databaseInfo.CreatedAt = time.Now()
				databaseInfo.UpdatedAt = time.Now()
				err = tx.TeamEnvDatabase.WithContext(ctx).Create(databaseInfo)
				if err != nil {
					log.Logger.Info("克隆环境--克隆环境下数据库失败，err:", err)
					return err
				}
			}

		}

		return record.InsertUpdate(ctx, req.TeamID, userID, record.OperationOperateSaveEnv, envName)
	})
	res := &rao.CreateEnvResp{
		EnvID:   envID,
		EnvName: envName,
	}
	return res, err
}

func CopyEnv(ctx *gin.Context, req *rao.CopyEnvReq) error {
	userID := jwt.GetUserIDByCtx(ctx)
	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		oldEnvInfo, err := tx.TeamEnv.WithContext(ctx).Where(tx.TeamEnv.ID.Eq(req.EnvID)).First()
		if err != nil {
			log.Logger.Info("克隆环境失败--环境不存在，err:", err)
			return err
		}
		oldEnvName := oldEnvInfo.Name
		newEnvName := oldEnvInfo.Name + "_1"

		// 查询老配置相关的
		list, err := tx.TeamEnv.WithContext(ctx).Where(tx.TeamEnv.TeamID.Eq(req.TeamID)).Where(tx.TeamEnv.Name.Like(fmt.Sprintf("%s%%", oldEnvName+"_"))).Find()
		if err == nil {
			// 有复制过得配置
			maxNum := 0
			for _, envInfo := range list {
				nameTmp := envInfo.Name
				postfixSlice := strings.Split(nameTmp, "_")
				if len(postfixSlice) < 2 {
					continue
				}
				currentNum, err := strconv.Atoi(postfixSlice[len(postfixSlice)-1])
				if err != nil {
					log.Logger.Info("克隆环境--类型转换失败，err:", err)
					continue
				}
				if currentNum > maxNum {
					maxNum = currentNum
				}
			}
			newEnvName = newEnvName + fmt.Sprintf("_%d", maxNum+1)
		}

		// 克隆环境数据
		oldEnvInfo.ID = 0
		oldEnvInfo.Name = newEnvName
		oldEnvInfo.CreatedUserID = userID
		oldEnvInfo.CreatedAt = time.Now()
		oldEnvInfo.UpdatedAt = time.Now()
		err = tx.TeamEnv.WithContext(ctx).Create(oldEnvInfo)
		if err != nil {
			log.Logger.Info("克隆环境--克隆环境基本数据失败，err:", err)
			return err
		}

		// 克隆环境下服务
		serviceList, err := tx.TeamEnvService.WithContext(ctx).Where(tx.TeamEnvService.TeamID.Eq(req.TeamID),
			tx.TeamEnvService.TeamEnvID.Eq(req.EnvID)).Find()
		if err != nil {
			return err
		}
		for _, serviceInfo := range serviceList {
			serviceInfo.ID = 0
			serviceInfo.TeamEnvID = oldEnvInfo.ID
			serviceInfo.CreatedUserID = userID
			serviceInfo.CreatedAt = time.Now()
			serviceInfo.UpdatedAt = time.Now()
			err = tx.TeamEnvService.WithContext(ctx).Create(serviceInfo)
			if err != nil {
				log.Logger.Info("克隆环境--克隆环境下服务失败，err:", err)
				return err
			}
		}

		// 克隆环境下服务
		databaseList, err := tx.TeamEnvDatabase.WithContext(ctx).Where(tx.TeamEnvDatabase.TeamID.Eq(req.TeamID),
			tx.TeamEnvDatabase.TeamEnvID.Eq(req.EnvID)).Find()
		if err != nil {
			return err
		}
		for _, databaseInfo := range databaseList {
			databaseInfo.ID = 0
			databaseInfo.TeamEnvID = oldEnvInfo.ID
			databaseInfo.CreatedUserID = userID
			databaseInfo.CreatedAt = time.Now()
			databaseInfo.UpdatedAt = time.Now()
			err = tx.TeamEnvDatabase.WithContext(ctx).Create(databaseInfo)
			if err != nil {
				log.Logger.Info("克隆环境--克隆环境下数据库失败，err:", err)
				return err
			}
		}
		return record.InsertCreate(ctx, req.TeamID, userID, record.OperationOperateCopyEnv, oldEnvInfo.Name)
	})
	return err
}

func DelEnv(ctx *gin.Context, req *rao.DelEnvReq) error {
	userID := jwt.GetUserIDByCtx(ctx)
	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		envInfo, err := tx.TeamEnv.WithContext(ctx).Where(tx.TeamEnv.ID.Eq(req.EnvID)).Where(tx.TeamEnv.TeamID.Eq(req.TeamID)).First()
		if err != nil {
			log.Logger.Info("删除环境失败,环境不存在,err:", err)
			return err
		}

		_, err = tx.TeamEnv.WithContext(ctx).Where(tx.TeamEnv.TeamID.Eq(req.TeamID), tx.TeamEnv.ID.Eq(req.EnvID)).Delete()
		if err != nil {
			return err
		}

		// 删除环境下所有服务
		_, err = tx.TeamEnvService.WithContext(ctx).Where(tx.TeamEnvService.TeamID.Eq(req.TeamID), tx.TeamEnvService.TeamEnvID.Eq(req.EnvID)).Delete()
		if err != nil {
			return err
		}

		// 删除环境下所有数据库
		_, err = tx.TeamEnvDatabase.WithContext(ctx).Where(tx.TeamEnvDatabase.TeamID.Eq(req.TeamID), tx.TeamEnvDatabase.TeamEnvID.Eq(req.EnvID)).Delete()
		if err != nil {
			return err
		}
		return record.InsertCreate(ctx, req.TeamID, userID, record.OperationOperateDeleteEnv, envInfo.Name)
	})
	return err
}

func DelEnvService(ctx *gin.Context, req *rao.DelEnvServiceReq) error {
	userID := jwt.GetUserIDByCtx(ctx)
	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		serviceInfo, err := tx.TeamEnvService.WithContext(ctx).Where(tx.TeamEnvService.ID.Eq(req.ServiceID),
			tx.TeamEnvService.TeamID.Eq(req.TeamID), tx.TeamEnvService.TeamEnvID.Eq(req.EnvID)).First()
		if err != nil {
			log.Logger.Info("删除环境服务失败,环境服务不存在，err:", err)
			return err
		}

		// 删除当前团队下，别的环境下同名的服务
		_, err = tx.TeamEnvService.WithContext(ctx).Where(tx.TeamEnvService.TeamID.Eq(req.TeamID),
			tx.TeamEnvService.Name.Eq(serviceInfo.Name)).Delete()
		if err != nil {
			return err
		}

		_, err = tx.TeamEnvService.WithContext(ctx).Where(tx.TeamEnvService.ID.Eq(req.ServiceID), tx.TeamEnvService.TeamID.Eq(req.TeamID)).Delete()
		if err != nil {
			return err
		}
		return record.InsertCreate(ctx, req.TeamID, userID, record.OperationOperateDeleteEnvService, serviceInfo.Name)
	})
	return err
}

func DelEnvDatabase(ctx *gin.Context, req *rao.DelEnvDatabaseReq) error {
	userID := jwt.GetUserIDByCtx(ctx)
	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		dbInfo, err := tx.TeamEnvDatabase.WithContext(ctx).Where(tx.TeamEnvDatabase.ID.Eq(req.DatabaseID),
			tx.TeamEnvDatabase.TeamID.Eq(req.TeamID), tx.TeamEnvDatabase.TeamEnvID.Eq(req.EnvID)).First()
		if err != nil {
			log.Logger.Info("删除环境服务失败,环境服务不存在，err:", err)
			return err
		}

		// 删除当前团队下，别的环境下同名的服务
		_, err = tx.TeamEnvDatabase.WithContext(ctx).Where(tx.TeamEnvDatabase.TeamID.Eq(req.TeamID),
			tx.TeamEnvDatabase.ServerName.Eq(dbInfo.ServerName)).Delete()
		if err != nil {
			return err
		}

		_, err = tx.TeamEnvDatabase.WithContext(ctx).Where(tx.TeamEnvDatabase.ID.Eq(req.DatabaseID),
			tx.TeamEnvDatabase.TeamID.Eq(req.TeamID)).Delete()
		if err != nil {
			return err
		}
		return record.InsertCreate(ctx, req.TeamID, userID, record.OperationLogDeleteDatabase, dbInfo.ServerName)
	})
	return err
}

func GetServiceList(ctx *gin.Context, req *rao.GetServiceListReq) ([]rao.ServiceList, int64, error) {
	tx := dal.GetQuery().TeamEnvService

	limit := req.Size
	offset := (req.Page - 1) * req.Size

	serviceList, total, err := tx.WithContext(ctx).Where(tx.TeamID.Eq(req.TeamID), tx.TeamEnvID.Eq(req.EnvID)).FindByPage(offset, limit)
	if err != nil {
		return nil, 0, err
	}
	res := make([]rao.ServiceList, 0, len(serviceList))
	for _, serviceInfo := range serviceList {
		temp := rao.ServiceList{
			ServiceID:   serviceInfo.ID,
			TeamID:      serviceInfo.TeamID,
			EnvID:       serviceInfo.TeamEnvID,
			ServiceName: serviceInfo.Name,
			Content:     serviceInfo.Content,
		}
		res = append(res, temp)
	}
	return res, total, err
}

func GetDatabaseList(ctx *gin.Context, req *rao.GetDatabaseListReq) ([]rao.DatabaseList, int64, error) {
	tx := dal.GetQuery().TeamEnvDatabase

	limit := req.Size
	offset := (req.Page - 1) * req.Size

	mysqlList, total, err := tx.WithContext(ctx).Where(tx.TeamID.Eq(req.TeamID), tx.TeamEnvID.Eq(req.EnvID)).FindByPage(offset, limit)
	if err != nil {
		return nil, 0, err
	}
	res := make([]rao.DatabaseList, 0, len(mysqlList))
	for _, dbInfo := range mysqlList {
		temp := rao.DatabaseList{
			DatabaseID: dbInfo.ID,
			TeamID:     dbInfo.TeamID,
			EnvID:      dbInfo.TeamEnvID,
			Type:       dbInfo.Type,
			ServerName: dbInfo.ServerName,
			Host:       dbInfo.Host,
			User:       dbInfo.User,
			Password:   dbInfo.Password,
			Port:       dbInfo.Port,
			DbName:     dbInfo.DbName,
			Charset:    dbInfo.Charset,
		}
		res = append(res, temp)
	}
	return res, total, err
}

func CreateEnvService(ctx *gin.Context, req *rao.CreateEnvServiceReq) error {
	userID := jwt.GetUserIDByCtx(ctx)
	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		_, err := tx.TeamEnvService.WithContext(ctx).Where(tx.TeamEnvService.TeamID.Eq(req.TeamID),
			tx.TeamEnvService.TeamEnvID.Eq(req.EnvID), tx.TeamEnvService.ID.Neq(req.ServiceID),
			tx.TeamEnvService.Name.Eq(req.ServiceName)).First()
		if err == nil {
			return fmt.Errorf("名称已存在")
		}

		// 查询当前团队下所有环境
		envList, err := tx.TeamEnv.WithContext(ctx).Where(tx.TeamEnv.TeamID.Eq(req.TeamID)).Find()
		if err != nil {
			return err
		}

		if req.ServiceID == 0 { // 新增
			for _, envInfo := range envList {
				InsertData := model.TeamEnvService{
					TeamID:        req.TeamID,
					TeamEnvID:     envInfo.ID,
					Name:          req.ServiceName,
					Content:       req.Content,
					CreatedUserID: userID,
				}
				err = tx.TeamEnvService.WithContext(ctx).Create(&InsertData)
				if err != nil {
					return err
				}
			}

			return record.InsertCreate(ctx, req.TeamID, userID, record.OperationLogCreateService, req.ServiceName)
		} else { // 修改
			// 查询原来的服务数据
			serviceInfo, err := tx.TeamEnvService.WithContext(ctx).Where(tx.TeamEnvService.ID.Eq(req.ServiceID)).First()
			if err != nil {
				return err
			}

			for _, envInfo := range envList {
				if envInfo.ID == req.EnvID {
					_, err = tx.TeamEnvService.WithContext(ctx).Where(tx.TeamEnvService.TeamID.Eq(req.TeamID), tx.TeamEnvService.ID.Eq(req.ServiceID)).UpdateSimple(
						tx.TeamEnvService.Name.Value(req.ServiceName), tx.TeamEnvService.Content.Value(req.Content),
						tx.TeamEnvService.CreatedUserID.Value(userID))
					if err != nil {
						return err
					}
				} else {
					_, err = tx.TeamEnvService.WithContext(ctx).Where(tx.TeamEnvService.TeamID.Eq(req.TeamID),
						tx.TeamEnvService.TeamEnvID.Eq(envInfo.ID),
						tx.TeamEnvService.Name.Eq(serviceInfo.Name)).UpdateSimple(
						tx.TeamEnvService.Name.Value(req.ServiceName), tx.TeamEnvService.CreatedUserID.Value(userID))
					if err != nil {
						return err
					}
				}
			}
			return record.InsertCreate(ctx, req.TeamID, userID, record.OperationLogUpdateService, req.ServiceName)
		}
	})
	return err
}

func CreateEnvDatabase(ctx *gin.Context, req *rao.CreateEnvDatabaseReq) error {
	userID := jwt.GetUserIDByCtx(ctx)
	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		_, err := tx.TeamEnvDatabase.WithContext(ctx).Where(tx.TeamEnvDatabase.TeamID.Eq(req.TeamID),
			tx.TeamEnvDatabase.TeamEnvID.Eq(req.EnvID), tx.TeamEnvDatabase.ID.Neq(req.DatabaseID),
			tx.TeamEnvDatabase.ServerName.Eq(req.ServerName)).First()
		if err == nil {
			return fmt.Errorf("名称已存在")
		}

		// 查询当前团队下所有环境
		envList, err := tx.TeamEnv.WithContext(ctx).Where(tx.TeamEnv.TeamID.Eq(req.TeamID)).Find()
		if err != nil {
			return err
		}

		if req.DatabaseID == 0 { // 新建
			for _, envInfo := range envList {
				InsertData := model.TeamEnvDatabase{
					TeamID:        req.TeamID,
					TeamEnvID:     envInfo.ID,
					Type:          req.Type,
					ServerName:    req.ServerName,
					Host:          req.Host,
					Port:          req.Port,
					User:          req.User,
					Password:      req.Password,
					DbName:        req.DbName,
					Charset:       req.Charset,
					CreatedUserID: userID,
				}
				err = tx.TeamEnvDatabase.WithContext(ctx).Create(&InsertData)
				if err != nil {
					return err
				}
			}

			return record.InsertCreate(ctx, req.TeamID, userID, record.OperationLogCreateDatabase, req.ServerName)
		} else { // 修改
			// 查询原来的服务数据
			dbInfo, err := tx.TeamEnvDatabase.WithContext(ctx).Where(tx.TeamEnvDatabase.ID.Eq(req.DatabaseID)).First()
			if err != nil {
				return err
			}

			for _, envInfo := range envList {
				if envInfo.ID == req.EnvID {
					_, err = tx.TeamEnvDatabase.WithContext(ctx).Where(tx.TeamEnvDatabase.TeamID.Eq(req.TeamID),
						tx.TeamEnvDatabase.TeamEnvID.Eq(req.EnvID),
						tx.TeamEnvDatabase.ID.Eq(req.DatabaseID)).UpdateSimple(
						tx.TeamEnvDatabase.Type.Value(req.Type), tx.TeamEnvDatabase.ServerName.Value(req.ServerName),
						tx.TeamEnvDatabase.Host.Value(req.Host), tx.TeamEnvDatabase.Port.Value(req.Port),
						tx.TeamEnvDatabase.User.Value(req.User), tx.TeamEnvDatabase.Password.Value(req.Password),
						tx.TeamEnvDatabase.DbName.Value(req.DbName),
						tx.TeamEnvDatabase.Charset.Value(req.Charset), tx.TeamEnvDatabase.CreatedUserID.Value(userID))
					if err != nil {
						return err
					}
				} else {
					_, err = tx.TeamEnvDatabase.WithContext(ctx).Where(tx.TeamEnvDatabase.TeamID.Eq(req.TeamID),
						tx.TeamEnvDatabase.TeamEnvID.Eq(envInfo.ID),
						tx.TeamEnvDatabase.ServerName.Eq(dbInfo.ServerName)).UpdateSimple(
						tx.TeamEnvDatabase.ServerName.Value(req.ServerName),
						tx.TeamEnvDatabase.CreatedUserID.Value(userID))
					if err != nil {
						return err
					}
				}
			}

			return record.InsertCreate(ctx, req.TeamID, userID, record.OperationLogUpdateDatabase, req.ServerName)
		}
	})
	return err
}
