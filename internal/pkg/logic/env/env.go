package env

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gen"
	"gorm.io/gorm"
	"kp-management/internal/pkg/biz/consts"
	"kp-management/internal/pkg/biz/jwt"
	"kp-management/internal/pkg/biz/log"
	"kp-management/internal/pkg/biz/record"
	"kp-management/internal/pkg/dal/model"
	"kp-management/internal/pkg/packer"
	"strconv"
	"strings"
	"sync"

	"kp-management/internal/pkg/dal/query"

	"kp-management/internal/pkg/dal"
	"kp-management/internal/pkg/dal/rao"
)

func GetList(ctx *gin.Context, req *rao.EnvListReq) ([]rao.EnvListResp, error) {

	teDB := query.Use(dal.DB()).TeamEnv

	conditions := make([]gen.Condition, 0)
	conditions = append(conditions, teDB.TeamID.Eq(req.TeamID))
	//conditions = append(conditions, teDB.Status.Eq(consts.TeamEnvStatusNormal))

	if req.Name != "" {
		conditions = append(conditions, teDB.Name.Like(fmt.Sprintf("%%%s%%", req.Name)))
	}

	envListData, envListErr := teDB.WithContext(ctx).Select(teDB.ID, teDB.Name, teDB.TeamID, teDB.Sort, teDB.CreatedUserID).Where(conditions...).Order(teDB.CreatedAt.Desc()).Find()
	if envListErr != nil {
		log.Logger.Info("环境列表--获取列表失败，err:", envListErr)
		return nil, envListErr
	}

	var envServiceList []rao.ServiceListResp

	if len(envListData) > 0 {

		var envIDs []int64
		for _, envV := range envListData {
			envIDs = append(envIDs, envV.ID)
		}
		tesDB := query.Use(dal.DB()).TeamEnvService

		envServiceListData, _ := tesDB.WithContext(ctx).Select(tesDB.ID, tesDB.TeamEnvID, tesDB.Name, tesDB.Content).Where(tesDB.TeamEnvID.In(envIDs...)).Where(tesDB.TeamID.Eq(req.TeamID)).Where(tesDB.Status.Eq(consts.TeamEnvServiceStatusNormal)).Order(tesDB.CreatedAt).Find()

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

// EnvNameIsExist 判断环境名称在同一团队下是否已存在
func EnvNameIsExist(ctx *gin.Context, req *rao.SaveEnvReq) (bool, error) {

	teTB := dal.GetQuery().TeamEnv

	conditions := make([]gen.Condition, 0)
	conditions = append(conditions, teTB.TeamID.Eq(req.TeamID))
	conditions = append(conditions, teTB.Name.Eq(req.Name))
	if req.ID != 0 {
		conditions = append(conditions, teTB.ID.Neq(req.ID))
	}

	existName, err := teTB.WithContext(ctx).Where(conditions...).Find()

	if err != nil {
		return false, err
	}
	if len(existName) != 0 {
		return true, err
	}

	return false, err
}

func SaveEnv(ctx *gin.Context, req *rao.SaveEnvReq) (*rao.SaveEnvResp, error) {

	//teDB := query.Use(dal.DB()).TeamEnv

	userID := jwt.GetUserIDByCtx(ctx)

	var respDetail *rao.SaveEnvResp

	query.Use(dal.DB()).Transaction(func(tb *query.Query) error {

		if req.ID != 0 {

			detail, detailErr := tb.TeamEnv.WithContext(ctx).Where(tb.TeamEnv.ID.Eq(req.ID)).Where(tb.TeamEnv.TeamID.Eq(req.TeamID)).Where(tb.TeamEnv.Status.Eq(consts.TeamEnvStatusNormal)).First()
			if detailErr != nil {
				log.Logger.Info("编辑环境失败,环境不存在 --编辑数据失败，err:", detailErr)
				return detailErr
			}

			updateData := &model.TeamEnv{
				Name: req.Name,
			}

			if _, err := tb.TeamEnv.WithContext(ctx).Omit(tb.TeamEnv.CreatedUserID).Where(tb.TeamEnv.ID.Eq(req.ID)).Updates(updateData); err != nil {
				log.Logger.Info("编辑环境失败 --编辑数据失败，err:", err)
				return err
			}

			//if _, err := tb.TeamEnvService.WithContext(ctx).Omit(tb.TeamEnvService.CreatedUserID).Where(tb.TeamEnvService.TeamEnvID.Eq(detail.ID)).Delete(); err != nil {
			//	log.Logger.Info("清空环境服务失败 --清空数据失败，err:", err)
			//	return err
			//}

			//遍历添加新服务
			for _, ServiceListV := range req.ServiceList {
				if ServiceListV.ID == 0 {
					insertServiceData := &model.TeamEnvService{
						TeamEnvID: detail.ID,
						TeamID:    req.TeamID,
						Name:      ServiceListV.Name,
						Content:   ServiceListV.Content,
						//Sort: req.Sort,
						CreatedUserID: userID,
					}

					err := tb.TeamEnvService.WithContext(ctx).Create(insertServiceData)
					if err != nil {
						log.Logger.Info("保存环境服务失败 --保存数据失败，err:", err)
						return err
					}

				} else {
					updateServiceData := &model.TeamEnvService{
						Name:    ServiceListV.Name,
						Content: ServiceListV.Content,
					}

					_, updateServiceErr := tb.TeamEnvService.WithContext(ctx).Omit(tb.TeamEnvService.CreatedUserID).Where(tb.TeamEnvService.ID.Eq(ServiceListV.ID)).Updates(updateServiceData)
					if updateServiceErr != nil {
						log.Logger.Info("编辑环境服务失败 --编辑数据失败，err:", updateServiceErr)
						return updateServiceErr
					}
				}
			}

			respDetail = &rao.SaveEnvResp{
				ID:     detail.ID,
				Name:   req.Name,
				TeamID: detail.TeamID,
				Sort:   detail.Sort,
			}

			return record.InsertUpdate(ctx, detail.TeamID, userID, record.OperationOperateUpdateEnv, detail.Name)

		} else {
			insertData := &model.TeamEnv{
				Name:   req.Name,
				TeamID: req.TeamID,
				//Sort: req.Sort,
				CreatedUserID: userID,
			}

			err := tb.TeamEnv.WithContext(ctx).Create(insertData)
			if err != nil {
				log.Logger.Info("保存环境失败 --保存数据失败，err:", err)
				return err
			}

			for _, ServiceListV := range req.ServiceList {
				insertServiceData := &model.TeamEnvService{
					TeamEnvID: insertData.ID,
					TeamID:    req.TeamID,
					Name:      ServiceListV.Name,
					Content:   ServiceListV.Content,
					//Sort: req.Sort,
					CreatedUserID: userID,
				}

				err := tb.TeamEnvService.WithContext(ctx).Create(insertServiceData)
				if err != nil {
					log.Logger.Info("保存环境服务失败 --保存数据失败，err:", err)
					return err
				}
			}

			respDetail = &rao.SaveEnvResp{
				ID:     insertData.ID,
				Name:   insertData.Name,
				TeamID: insertData.TeamID,
				Sort:   insertData.Sort,
			}

			return record.InsertCreate(ctx, insertData.TeamID, userID, record.OperationOperateSaveEnv, insertData.Name)
		}
	})

	return respDetail, nil
}

func CopyEnv(ctx *gin.Context, req *rao.CopyEnvReq) (*rao.CopyEnvResp, error) {

	teDB := query.Use(dal.DB()).TeamEnv

	userID := jwt.GetUserIDByCtx(ctx)

	detail, detailErr := teDB.WithContext(ctx).Where(teDB.ID.Eq(req.ID)).Where(teDB.TeamID.Eq(req.TeamID)).First()
	if detailErr != nil {
		log.Logger.Info("复制环境失败,环境不存在 --复制数据失败，err:", detailErr)
		return nil, detailErr
	}

	oldName := detail.Name
	newName := ""

	var lock sync.Mutex
	lock.Lock()
	defer lock.Unlock()
	list, err := teDB.WithContext(ctx).Where(teDB.TeamID.Eq(detail.TeamID)).Where(teDB.Name.Like(fmt.Sprintf("%s%%", oldName+"_"))).Find()
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Logger.Info("复制用例--查询错误，err:", err)
		return nil, err
	} else if err == gorm.ErrRecordNotFound {
		newName = oldName + "_1"
	} else { // 有复制过得配置
		maxNum := 0
		for _, caseInfo := range list {
			nameTmp := caseInfo.Name
			postfixSlice := strings.Split(nameTmp, "_")
			if len(postfixSlice) != 2 {
				continue
			}
			currentNum, err := strconv.Atoi(postfixSlice[1])
			if err != nil {
				log.Logger.Info("复制用例--类型转换失败，err:", err)
			}
			if currentNum > maxNum {
				maxNum = currentNum
			}
		}
		newName = oldName + "_" + fmt.Sprintf("%d", maxNum+1)
	}

	var copyEnvResp *rao.CopyEnvResp

	query.Use(dal.DB()).Transaction(func(tb *query.Query) error {

		insertData := &model.TeamEnv{
			TeamID:        detail.TeamID,
			Name:          newName,
			CreatedUserID: userID,
		}
		err := tb.TeamEnv.WithContext(ctx).Create(insertData)
		if err != nil {
			log.Logger.Info("复制环境服务失败 --复制数据失败，err:", err)
			return err
		}

		ServiceList, ServiceListErr := tb.TeamEnvService.WithContext(ctx).Where(tb.TeamEnvService.TeamEnvID.Eq(detail.ID)).Find()
		if ServiceListErr != nil {
			log.Logger.Info("复制环境失败,环境不存在 --复制数据失败，err:", ServiceListErr)
			return ServiceListErr
		}

		if len(ServiceList) > 0 {
			for _, ServiceListV := range ServiceList {
				insertServiceData := &model.TeamEnvService{
					TeamID:    ServiceListV.TeamID,
					TeamEnvID: insertData.ID,
					Name:      ServiceListV.Name,
					Content:   ServiceListV.Content,
				}
				err := tb.TeamEnvService.WithContext(ctx).Create(insertServiceData)
				if err != nil {
					log.Logger.Info("复制环境服务失败 --复制数据失败，err:", err)
					return err
				}
			}
		}

		copyEnvResp = &rao.CopyEnvResp{
			ID:     insertData.ID,
			Name:   insertData.Name,
			TeamID: insertData.TeamID,
		}

		return record.InsertCreate(ctx, insertData.TeamID, userID, record.OperationOperateCopyEnv, insertData.Name)
	})

	return copyEnvResp, nil
}

func DelEnv(ctx *gin.Context, req *rao.DelEnvReq) error {

	teDB := query.Use(dal.DB()).TeamEnv
	tesDB := query.Use(dal.DB()).TeamEnvService

	userID := jwt.GetUserIDByCtx(ctx)

	detail, detailErr := teDB.WithContext(ctx).Where(teDB.ID.Eq(req.ID)).Where(teDB.TeamID.Eq(req.TeamID)).First()
	if detailErr != nil {
		log.Logger.Info("删除环境失败,环境不存在 --编辑数据失败，err:", detailErr)
		return detailErr
	}

	//updateData := make(map[string]interface{}, 1)
	//updateData["status"] = consts.TeamEnvStatusDel
	//updateData["recent_user_id"] = userID

	_, err := teDB.WithContext(ctx).Where(teDB.ID.Eq(req.ID)).Where(teDB.TeamID.Eq(req.TeamID)).Delete()
	if err != nil {
		return err
	}

	////清除原环境下所有服务
	//updateServiceData := &model.TeamEnvService{
	//	Status:       consts.TeamEnvServiceStatusDel,
	//	RecentUserID: userID,
	//}

	if _, err := tesDB.WithContext(ctx).Omit(tesDB.CreatedUserID).Where(tesDB.TeamEnvID.Eq(detail.ID)).Delete(); err != nil {
		log.Logger.Info("清空环境服务失败 --清空数据失败，err:", err)
		return err
	}

	record.InsertCreate(ctx, req.TeamID, userID, record.OperationOperateDeleteEnv, detail.Name)

	return nil
}

func DelService(ctx *gin.Context, req *rao.DelServiceReq) error {
	tesDB := query.Use(dal.DB()).TeamEnvService

	userID := jwt.GetUserIDByCtx(ctx)

	detail, detailErr := tesDB.WithContext(ctx).Where(tesDB.ID.Eq(req.ID)).Where(tesDB.TeamEnvID.Eq(req.EnvID)).Where(tesDB.TeamID.Eq(req.TeamID)).First()
	if detailErr != nil {
		log.Logger.Info("删除环境服务失败,环境服务不存在 --编辑数据失败，err:", detailErr)
		return detailErr
	}

	//updateData := make(map[string]interface{}, 1)
	//updateData["status"] = consts.TeamEnvServiceStatusDel
	//updateData["recent_user_id"] = userID

	_, err := tesDB.WithContext(ctx).Where(tesDB.ID.Eq(req.ID)).Where(tesDB.TeamID.Eq(req.TeamID)).Delete()
	if err != nil {
		return err
	}

	record.InsertCreate(ctx, req.TeamID, userID, record.OperationOperateDeleteEnvService, detail.Name)

	return nil
}
