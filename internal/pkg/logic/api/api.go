package api

import (
	"context"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/jwt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/log"
	"github.com/gin-gonic/gin"
	"gorm.io/gen"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/record"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/mao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/query"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/packer"
)

func Save(ctx context.Context, req *rao.SaveTargetReq, userID string) (string, error) {
	target := packer.TransSaveTargetReqToTargetModel(req, userID)
	apiDetail := packer.TransSaveTargetReqToMaoAPI(req)

	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectAPI)

	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		// 接口名排重
		_, err := tx.Target.WithContext(ctx).Where(
			tx.Target.TeamID.Eq(req.TeamID),
			tx.Target.Name.Eq(req.Name),
			tx.Target.TargetType.Eq(consts.TargetTypeAPI),
			tx.Target.TargetID.Neq(req.TargetID),
			tx.Target.Status.Eq(consts.TargetStatusNormal),
			tx.Target.ParentID.Eq(req.ParentID),
			tx.Target.Source.Eq(req.Source)).First()
		if err == nil {
			return fmt.Errorf("名称已存在")
		}

		// 查询当前接口是否存在
		_, err = tx.Target.WithContext(ctx).Where(tx.Target.TargetID.Eq(req.TargetID)).First()
		if err != nil { // 需新增
			if err := tx.Target.WithContext(ctx).Create(target); err != nil {
				return err
			}
			apiDetail.TargetID = target.TargetID
			_, err := collection.InsertOne(ctx, apiDetail)
			if err != nil {
				return err
			}
			return record.InsertCreate(ctx, target.TeamID, userID, record.OperationOperateCreateAPI, target.Name)
		}

		if _, err := tx.Target.WithContext(ctx).Where(tx.Target.TargetID.Eq(req.TargetID)).Updates(target); err != nil {
			return err
		}

		_, err = collection.UpdateOne(ctx, bson.D{{"target_id", target.TargetID}}, bson.M{"$set": apiDetail})
		if err != nil {
			return err
		}
		return record.InsertUpdate(ctx, target.TeamID, userID, record.OperationOperateUpdateAPI, target.Name)
	})
	return target.TargetID, err
}

func SaveImportApi(ctx *gin.Context, req *rao.SaveImportApiReq, userID string) error {
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectAPI)
	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		// 文件夹相关数据
		folderData := make([]rao.SaveTargetReq, 0, len(req.Apis))
		apiData := make([]rao.SaveTargetReq, 0, len(req.Apis))
		for _, info := range req.Apis {
			if info.TargetType == "folder" {
				folderData = append(folderData, info)
			}
			if info.TargetType == "api" {
				apiData = append(apiData, info)
			}
		}

		// 老的targetID对应新的targetID字典
		oldAndNewTargetIDMap := make(map[string]string)
		// 先把文件夹入库
		if len(folderData) > 0 {
			for _, folderInfo := range folderData {
				insertTargetInfo := packer.TransSaveImportFolderReqToTargetModel(folderInfo, req.TeamID, userID)
				newFolderName := insertTargetInfo.Name
				// 文件夹名排重
				_, err := tx.Target.WithContext(ctx).Where(tx.Target.TeamID.Eq(req.TeamID), tx.Target.Name.Eq(insertTargetInfo.Name),
					tx.Target.TargetType.Eq(consts.TargetTypeFolder), tx.Target.Status.Eq(consts.TargetStatusNormal),
					tx.Target.ParentID.Eq(folderInfo.ParentID)).First()
				if err == nil {
					newFolderName = newFolderName + "_1"
					// 查询老配置相关的
					list, err := tx.Target.WithContext(ctx).Where(tx.Target.TeamID.Eq(req.TeamID), tx.Target.Name.Like(fmt.Sprintf("%s%%", insertTargetInfo.Name+"_"))).Find()
					if err == nil && len(list) > 0 {
						// 有复制过得配置
						maxNum := 0
						for _, targetInfo := range list {
							nameTmp := targetInfo.Name
							postfixSlice := strings.Split(nameTmp, "_")
							if len(postfixSlice) < 2 {
								continue
							}
							currentNum, err := strconv.Atoi(postfixSlice[len(postfixSlice)-1])
							if err != nil {
								log.Logger.Info("复制自动化计划--类型转换失败，err:", err)
								continue
							}
							if currentNum > maxNum {
								maxNum = currentNum
							}
						}
						newFolderName = insertTargetInfo.Name + fmt.Sprintf("_%d", maxNum+1)
					}

				}
				// 插入target表
				insertTargetInfo.Name = newFolderName
				if folderInfo.OldParentID != "" {
					if newTargetID, ok := oldAndNewTargetIDMap[folderInfo.OldParentID]; ok {
						insertTargetInfo.ParentID = newTargetID
					}
				}

				err = tx.Target.WithContext(ctx).Create(insertTargetInfo)
				if err != nil {
					return err
				}
				oldAndNewTargetIDMap[folderInfo.OldTargetID] = insertTargetInfo.TargetID
			}
		}

		if len(apiData) > 0 {
			for _, apiInfo := range apiData {
				insertTargetInfo := packer.TransSaveImportTargetReqToTargetModel(apiInfo, req.TeamID, userID)
				newApiName := insertTargetInfo.Name
				// 接口名排重
				_, err := tx.Target.WithContext(ctx).Where(tx.Target.TeamID.Eq(req.TeamID), tx.Target.Name.Eq(insertTargetInfo.Name),
					tx.Target.TargetType.Eq(consts.TargetTypeAPI), tx.Target.Status.Eq(consts.TargetStatusNormal),
					tx.Target.ParentID.Eq(apiInfo.ParentID)).First()
				if err == nil {
					newApiName = newApiName + "_1"
					// 查询老配置相关的
					list, err := tx.Target.WithContext(ctx).Where(tx.Target.TeamID.Eq(req.TeamID), tx.Target.Name.Like(fmt.Sprintf("%s%%", insertTargetInfo.Name+"_"))).Find()
					if err == nil {
						// 有复制过得配置
						maxNum := 0
						for _, targetInfo := range list {
							nameTmp := targetInfo.Name
							postfixSlice := strings.Split(nameTmp, "_")
							if len(postfixSlice) < 2 {
								continue
							}
							currentNum, err := strconv.Atoi(postfixSlice[len(postfixSlice)-1])
							if err != nil {
								log.Logger.Info("复制自动化计划--类型转换失败，err:", err)
								continue
							}
							if currentNum > maxNum {
								maxNum = currentNum
							}
						}
						newApiName = insertTargetInfo.Name + fmt.Sprintf("_%d", maxNum+1)
					}

				}
				// 插入target表
				insertTargetInfo.Name = newApiName

				if apiInfo.OldParentID != "" {
					if newTargetID, ok := oldAndNewTargetIDMap[apiInfo.OldParentID]; ok {
						insertTargetInfo.ParentID = newTargetID
					}
				}

				err = tx.Target.WithContext(ctx).Create(insertTargetInfo)
				if err != nil {
					return err
				}

				apiDetail := packer.TransSaveTargetReqToMaoAPI(&apiInfo)
				// 把接口详情插入mongodb数据库
				apiDetail.TargetID = insertTargetInfo.TargetID
				_, err = collection.InsertOne(ctx, apiDetail)
				if err != nil {
					return err
				}
			}
		}

		return nil
	})
	return err
}

func DetailByTargetIDs(ctx context.Context, req *rao.BatchGetDetailReq) ([]rao.APIDetail, error) {
	tx := query.Use(dal.DB()).Target
	targets, err := tx.WithContext(ctx).Where(
		tx.TargetID.In(req.TargetIDs...),
		tx.TeamID.Eq(req.TeamID),
		tx.Status.Eq(consts.TargetStatusNormal),
	).Order(tx.Sort.Desc(), tx.CreatedAt.Desc()).Find()

	if err != nil {
		return nil, err
	}

	if len(targets) == 0 {
		return nil, fmt.Errorf("删除失败")
	}

	apiIDs := make([]string, 0, len(targets))
	sqlIDs := make([]string, 0, len(targets))
	tcpIDs := make([]string, 0, len(targets))
	websocketIDs := make([]string, 0, len(targets))
	//mqttIDs := make([]string, 0, len(targets))
	dubboIDs := make([]string, 0, len(targets))
	for _, targetInfo := range targets {
		if targetInfo.TargetType == consts.TargetTypeAPI {
			apiIDs = append(apiIDs, targetInfo.TargetID)
		}
		if targetInfo.TargetType == consts.TargetTypeSql {
			sqlIDs = append(sqlIDs, targetInfo.TargetID)
		}
		if targetInfo.TargetType == consts.TargetTypeTcp {
			tcpIDs = append(tcpIDs, targetInfo.TargetID)
		}
		if targetInfo.TargetType == consts.TargetTypeWebsocket {
			websocketIDs = append(websocketIDs, targetInfo.TargetID)
		}
		//if targetInfo.TargetType == consts.TargetTypeMQTT {
		//	mqttIDs = append(mqttIDs, targetInfo.TargetID)
		//}
		if targetInfo.TargetType == consts.TargetTypeDubbo {
			dubboIDs = append(dubboIDs, targetInfo.TargetID)
		}
	}

	// 查询接口详情数据
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectAPI)
	cursor, err := collection.Find(ctx, bson.D{{"target_id", bson.D{{"$in", apiIDs}}}})
	if err != nil {
		return nil, err
	}
	var apis []*mao.API
	if err = cursor.All(ctx, &apis); err != nil {
		return nil, err
	}

	apiMap := make(map[string]*mao.API, len(apis))
	for _, apiInfo := range apis {
		apiMap[apiInfo.TargetID] = apiInfo
	}

	// 查询mysql详情数据
	collection = dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectSqlDetail)
	cursor, err = collection.Find(ctx, bson.D{{"target_id", bson.D{{"$in", sqlIDs}}}})
	if err != nil {
		return nil, err
	}
	var sqls []*mao.SqlDetailForMg
	if err = cursor.All(ctx, &sqls); err != nil {
		return nil, err
	}
	sqlMap := make(map[string]*mao.SqlDetailForMg, len(sqls))
	for _, sqlInfo := range sqls {
		sqlMap[sqlInfo.TargetID] = sqlInfo
	}

	// 查询Tcp详情数据
	collection = dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectTcpDetail)
	cursor, err = collection.Find(ctx, bson.D{{"target_id", bson.D{{"$in", tcpIDs}}}})
	if err != nil {
		return nil, err
	}
	var tcps []*mao.TcpDetail
	if err = cursor.All(ctx, &tcps); err != nil {
		return nil, err
	}
	tcpMap := make(map[string]*mao.TcpDetail, len(tcps))
	for _, tcpInfo := range tcps {
		tcpMap[tcpInfo.TargetID] = tcpInfo
	}

	// 查询websocket详情数据
	collection = dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectWebsocketDetail)
	cursor, err = collection.Find(ctx, bson.D{{"target_id", bson.D{{"$in", websocketIDs}}}})
	if err != nil {
		return nil, err
	}

	websockets := make([]*mao.WebsocketDetail, 0)
	if err = cursor.All(ctx, &websockets); err != nil {
		return nil, err
	}
	websocketMap := make(map[string]*mao.WebsocketDetail, len(websockets))
	for _, websocketInfo := range websockets {
		websocketMap[websocketInfo.TargetID] = websocketInfo
	}

	//// 查询mqtt详情数据
	//collection = dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectMqttDetail)
	//cursor, err = collection.Find(ctx, bson.D{{"target_id", bson.D{{"$in", mqttIDs}}}})
	//if err != nil {
	//	return nil, err
	//}

	//mqtts := make([]*mao.MqttDetail, 0)
	//if err = cursor.All(ctx, &mqtts); err != nil {
	//	return nil, err
	//}
	//mqttMap := make(map[string]*mao.MqttDetail, len(websockets))
	//for _, mqttInfo := range mqtts {
	//	mqttMap[mqttInfo.TargetID] = mqttInfo
	//}

	// 查询dubbo详情数据
	collection = dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectDubboDetail)
	cursor, err = collection.Find(ctx, bson.D{{"target_id", bson.D{{"$in", dubboIDs}}}})
	if err != nil {
		return nil, err
	}
	dubbos := make([]*mao.DubboDetail, len(dubboIDs))
	if err = cursor.All(ctx, &dubbos); err != nil {
		return nil, err
	}
	dubboMap := make(map[string]*mao.DubboDetail, len(dubboIDs))
	for _, dubboInfo := range dubbos {
		dubboMap[dubboInfo.TargetID] = dubboInfo
	}

	return packer.TransTargetsToRaoAPIDetails(targets, apiMap, sqlMap, tcpMap, websocketMap, dubboMap), nil
}

func SaveSql(ctx *gin.Context, req *rao.SaveTargetReq) (string, error) {
	userID := jwt.GetUserIDByCtx(ctx)
	targetData := packer.TransSaveTargetReqToTargetModel(req, userID)
	sqlDetail := packer.TransSaveTargetReqToMaoSqlDetail(req)

	// mongodb表
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectSqlDetail)

	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		// 接口名排重
		_, err := tx.Target.WithContext(ctx).Where(tx.Target.TeamID.Eq(req.TeamID), tx.Target.Name.Eq(req.Name),
			tx.Target.TargetType.Eq(req.TargetType), tx.Target.TargetID.Neq(req.TargetID),
			tx.Target.Status.Eq(consts.TargetStatusNormal), tx.Target.ParentID.Eq(req.ParentID),
			tx.Target.Source.Eq(req.Source)).First()
		if err == nil {
			return fmt.Errorf("名称已存在")
		}

		// 查询当前Mysql是否存在
		_, err = tx.Target.WithContext(ctx).Where(tx.Target.TargetID.Eq(req.TargetID)).First()
		if err != nil { // 新增
			if err := tx.Target.WithContext(ctx).Create(targetData); err != nil {
				return err
			}
			_, err := collection.InsertOne(ctx, sqlDetail)
			if err != nil {
				return err
			}
			return record.InsertCreate(ctx, targetData.TeamID, userID, record.OperationLogCreateSql, targetData.Name)
		} else { // 修改
			if _, err := tx.Target.WithContext(ctx).Where(tx.Target.TargetID.Eq(req.TargetID)).Updates(targetData); err != nil {
				return err
			}

			_, err = collection.UpdateOne(ctx, bson.D{{"target_id", req.TargetID}}, bson.M{"$set": sqlDetail})
			if err != nil {
				return err
			}
			return record.InsertUpdate(ctx, targetData.TeamID, userID, record.OperationLogUpdateSql, targetData.Name)
		}
	})
	return req.TargetID, err
}

func GetSqlDatabaseList(ctx *gin.Context, req *rao.GetSqlDatabaseListReq) ([]rao.GetSqlDatabaseListResp, error) {
	tx := dal.GetQuery().TeamEnvDatabase
	conditions := make([]gen.Condition, 0)
	conditions = append(conditions, tx.TeamID.Eq(req.TeamID))
	if req.EnvID != 0 {
		conditions = append(conditions, tx.TeamEnvID.Eq(req.EnvID))
	}
	dbList, err := tx.WithContext(ctx).Where(conditions...).Find()
	if err != nil {
		return nil, err
	}

	res := make([]rao.GetSqlDatabaseListResp, 0, len(dbList))
	for _, dbInfo := range dbList {
		temp := rao.GetSqlDatabaseListResp{
			MysqlID:    dbInfo.ID,
			Type:       dbInfo.Type,
			ServerName: dbInfo.ServerName,
			Host:       dbInfo.Host,
			Port:       dbInfo.Port,
			User:       dbInfo.User,
			Password:   dbInfo.Password,
			DbName:     dbInfo.DbName,
			Charset:    dbInfo.Charset,
		}
		res = append(res, temp)
	}
	return res, nil
}

func SaveTcp(ctx *gin.Context, req *rao.SaveTargetReq) (string, error) {
	userID := jwt.GetUserIDByCtx(ctx)
	targetData := packer.TransSaveTargetReqToTargetModel(req, userID)
	tcpDetail := packer.TransSaveTargetReqToMaoTcpDetail(req)

	// mongodb表
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectTcpDetail)

	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		// 接口名排重
		_, err := tx.Target.WithContext(ctx).Where(tx.Target.TeamID.Eq(req.TeamID), tx.Target.Name.Eq(req.Name),
			tx.Target.TargetType.Eq(req.TargetType), tx.Target.TargetID.Neq(req.TargetID),
			tx.Target.Status.Eq(consts.TargetStatusNormal), tx.Target.ParentID.Eq(req.ParentID),
			tx.Target.Source.Eq(req.Source)).First()
		if err == nil {
			return fmt.Errorf("名称已存在")
		}

		// 查询当前Mysql是否存在
		_, err = tx.Target.WithContext(ctx).Where(tx.Target.TargetID.Eq(req.TargetID)).First()
		if err != nil { // 新增
			if err := tx.Target.WithContext(ctx).Create(targetData); err != nil {
				return err
			}
			_, err := collection.InsertOne(ctx, tcpDetail)
			if err != nil {
				return err
			}
			return record.InsertCreate(ctx, targetData.TeamID, userID, record.OperationLogCreateTcp, targetData.Name)
		} else { // 修改
			if _, err := tx.Target.WithContext(ctx).Where(tx.Target.TargetID.Eq(req.TargetID)).Updates(targetData); err != nil {
				return err
			}
			_, err = collection.UpdateOne(ctx, bson.D{{"target_id", req.TargetID}}, bson.M{"$set": tcpDetail})
			if err != nil {
				return err
			}
			return record.InsertUpdate(ctx, targetData.TeamID, userID, record.OperationLogUpdateTcp, targetData.Name)
		}
	})
	return req.TargetID, err
}

func SaveWebsocket(ctx *gin.Context, req *rao.SaveTargetReq) (string, error) {
	userID := jwt.GetUserIDByCtx(ctx)
	targetData := packer.TransSaveTargetReqToTargetModel(req, userID)
	websocketDetail := packer.TransSaveTargetReqToMaoWebsocketDetail(req)

	// mongodb表
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectWebsocketDetail)

	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		// 接口名排重
		_, err := tx.Target.WithContext(ctx).Where(tx.Target.TeamID.Eq(req.TeamID), tx.Target.Name.Eq(req.Name),
			tx.Target.TargetType.Eq(req.TargetType), tx.Target.TargetID.Neq(req.TargetID),
			tx.Target.Status.Eq(consts.TargetStatusNormal), tx.Target.ParentID.Eq(req.ParentID),
			tx.Target.Source.Eq(req.Source)).First()
		if err == nil {
			return fmt.Errorf("名称已存在")
		}

		// 查询当前Mysql是否存在
		_, err = tx.Target.WithContext(ctx).Where(tx.Target.TargetID.Eq(req.TargetID)).First()
		if err != nil { // 新增
			if err := tx.Target.WithContext(ctx).Create(targetData); err != nil {
				return err
			}
			_, err := collection.InsertOne(ctx, websocketDetail)
			if err != nil {
				return err
			}
			return record.InsertCreate(ctx, targetData.TeamID, userID, record.OperationLogCreateWebsocket, targetData.Name)
		} else { // 修改
			if _, err := tx.Target.WithContext(ctx).Where(tx.Target.TargetID.Eq(req.TargetID)).Updates(targetData); err != nil {
				return err
			}

			_, err = collection.UpdateOne(ctx, bson.D{{"target_id", req.TargetID}}, bson.M{"$set": websocketDetail})
			if err != nil {
				return err
			}
			return record.InsertUpdate(ctx, targetData.TeamID, userID, record.OperationLogUpdateWebsocket, targetData.Name)
		}
	})
	return req.TargetID, err
}

func SaveMQTT(ctx *gin.Context, req *rao.SaveTargetReq) (string, error) {
	userID := jwt.GetUserIDByCtx(ctx)
	targetData := packer.TransSaveTargetReqToTargetModel(req, userID)
	mqttDetail := packer.TransSaveTargetReqToMaoMqttDetail(req)

	// mongodb表
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectMqttDetail)

	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		// 接口名排重
		_, err := tx.Target.WithContext(ctx).Where(tx.Target.TeamID.Eq(req.TeamID), tx.Target.Name.Eq(req.Name),
			tx.Target.TargetType.Eq(req.TargetType), tx.Target.TargetID.Neq(req.TargetID),
			tx.Target.Status.Eq(consts.TargetStatusNormal), tx.Target.ParentID.Eq(req.ParentID),
			tx.Target.Source.Eq(req.Source)).First()
		if err == nil {
			return fmt.Errorf("名称已存在")
		}

		// 查询当前target是否存在
		_, err = tx.Target.WithContext(ctx).Where(tx.Target.TargetID.Eq(req.TargetID)).First()
		if err != nil { // 新增
			if err := tx.Target.WithContext(ctx).Create(targetData); err != nil {
				return err
			}
			_, err := collection.InsertOne(ctx, mqttDetail)
			if err != nil {
				return err
			}
			return record.InsertCreate(ctx, targetData.TeamID, userID, record.OperationLogCreateMqtt, targetData.Name)
		} else { // 修改
			if _, err := tx.Target.WithContext(ctx).Where(tx.Target.TargetID.Eq(req.TargetID)).Updates(targetData); err != nil {
				return err
			}

			_, err = collection.UpdateOne(ctx, bson.D{{"target_id", req.TargetID}}, bson.M{"$set": mqttDetail})
			if err != nil {
				return err
			}
			return record.InsertUpdate(ctx, targetData.TeamID, userID, record.OperationLogUpdateMqtt, targetData.Name)
		}
	})
	return req.TargetID, err
}

func SaveDubbo(ctx *gin.Context, req *rao.SaveTargetReq) (string, error) {
	userID := jwt.GetUserIDByCtx(ctx)
	targetData := packer.TransSaveTargetReqToTargetModel(req, userID)
	dubboDetail := packer.TransSaveTargetReqToMaoDubboDetail(req)
	// mongodb表
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectDubboDetail)

	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		// 接口名排重
		_, err := tx.Target.WithContext(ctx).Where(tx.Target.TeamID.Eq(req.TeamID), tx.Target.Name.Eq(req.Name),
			tx.Target.TargetType.Eq(req.TargetType), tx.Target.TargetID.Neq(req.TargetID),
			tx.Target.Status.Eq(consts.TargetStatusNormal), tx.Target.ParentID.Eq(req.ParentID),
			tx.Target.Source.Eq(req.Source)).First()
		if err == nil {
			return fmt.Errorf("名称已存在")
		}

		// 查询当前Dubbo是否存在
		_, err = tx.Target.WithContext(ctx).Where(tx.Target.TargetID.Eq(req.TargetID)).First()
		if err != nil { // 新增
			if err := tx.Target.WithContext(ctx).Create(targetData); err != nil {
				return err
			}
			_, err := collection.InsertOne(ctx, dubboDetail)
			if err != nil {
				return err
			}
			return record.InsertCreate(ctx, targetData.TeamID, userID, record.OperationLogCreateDubbo, targetData.Name)
		} else { // 修改
			if _, err := tx.Target.WithContext(ctx).Where(tx.Target.TargetID.Eq(req.TargetID)).Updates(targetData); err != nil {
				return err
			}

			_, err = collection.UpdateOne(ctx, bson.D{{"target_id", req.TargetID}}, bson.M{"$set": dubboDetail})
			if err != nil {
				return err
			}
			return record.InsertUpdate(ctx, targetData.TeamID, userID, record.OperationLogUpdateDubbo, targetData.Name)
		}
	})
	return req.TargetID, err
}
