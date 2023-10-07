package uiScene

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/api/ui"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/log"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/uuid"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/clients"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/mao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/query"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/errmsg"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/packer"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/public"
	"github.com/gin-gonic/gin"
	"github.com/go-omnibus/proof"
	"github.com/jinzhu/copier"
	"go.mongodb.org/mongo-driver/bson"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"
)

func Save(ctx *gin.Context, userID string, req *rao.UISceneSaveReq) error {
	// 名称不能存在
	req.SceneID = uuid.GetUUID()
	us := query.Use(dal.DB()).UIScene
	if _, err := us.WithContext(ctx).Where(
		us.TeamID.Eq(req.TeamID),
		us.Name.Eq(req.Name),
		us.SceneType.Eq(consts.UISceneTypeScene),
		us.ParentID.Eq(req.ParentID),
		us.Status.Eq(consts.UISceneStatusNormal),
		us.Source.Eq(req.Source),
		us.PlanID.Eq(req.PlanID),
	).First(); err == nil {
		return errmsg.ErrUISceneNameRepeat
	}

	// 随机获取一个机器 key
	if len(req.UIMachineKey) == 0 {
		machineKey, _ := clients.RandUiEngineMachineID()
		req.UIMachineKey = machineKey
	}
	scene := packer.TransSaveReqToUISceneModel(req, userID)

	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		// 查询当前接口是否存在
		_, err := tx.UIScene.WithContext(ctx).Where(tx.UIScene.SceneID.Eq(req.SceneID)).First()
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if errors.Is(err, gorm.ErrRecordNotFound) { // 需新增
			if err = tx.UIScene.WithContext(ctx).Create(scene); err != nil {
				return err
			}

			//if err = record.InsertCreate(ctx, req.TeamID, userID, record.OperationOperateCreateUISceneAPI, req.Name); err != nil {
			//	return err
			//}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func Update(ctx *gin.Context, userID string, req *rao.UISceneSaveReq) error {
	// 名称不能存在
	us := query.Use(dal.DB()).UIScene
	if _, err := us.WithContext(ctx).Where(
		us.TeamID.Eq(req.TeamID),
		us.Name.Eq(req.Name),
		us.SceneType.Eq(consts.UISceneTypeScene),
		us.ParentID.Eq(req.ParentID),
		us.SceneID.Neq(req.SceneID),
		us.Status.Eq(consts.UISceneStatusNormal),
		us.Source.Eq(req.Source),
		us.PlanID.Eq(req.PlanID),
	).First(); err == nil {
		return errmsg.ErrUISceneNameRepeat
	}

	// 随机获取一个机器 key
	if len(req.UIMachineKey) == 0 {
		machineKey, _ := clients.RandUiEngineMachineID()
		req.UIMachineKey = machineKey
	}
	scene := packer.TransSaveReqToUISceneModel(req, userID)

	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		// 查询当前接口是否存在
		_, err := tx.UIScene.WithContext(ctx).Where(tx.UIScene.SceneID.Eq(req.SceneID)).First()
		if err != nil {
			return err
		}

		if _, err = tx.UIScene.WithContext(ctx).Where(tx.UIScene.SceneID.Eq(req.SceneID)).Updates(scene); err != nil {
			return err
		}

		// 为空字段处理
		fields := make([]field.AssignExpr, 0)
		if req.Description != nil {
			fields = append(fields, tx.UIScene.Description.Value(*req.Description))
		}
		if len(fields) > 0 {
			if _, err = tx.UIScene.WithContext(ctx).Where(tx.UIScene.SceneID.Eq(req.SceneID)).UpdateColumnSimple(fields...); err != nil {
				return err
			}
		}

		//if err = record.InsertUpdate(ctx, req.TeamID, userID, record.OperationOperateUpdateUISceneAPI, req.Name); err != nil {
		//	return err
		//}

		return nil
	})

	// 同步绑定的场景
	if err = SyncBindScene(ctx, req.SceneID, req.TeamID, userID); err != nil {
		log.Logger.Error("SyncBindScene err", proof.WithError(err))
	}

	if err != nil {
		return err
	}

	return nil
}

func DetailBySceneID(ctx *gin.Context, req *rao.UISceneDetailReq) (*rao.UIScene, error) {
	us := query.Use(dal.DB()).UIScene
	s, err := us.WithContext(ctx).Where(
		us.SceneID.Eq(req.SceneID),
		us.TeamID.Eq(req.TeamID),
		us.Status.Eq(consts.TargetStatusNormal),
	).First()

	if err != nil {
		return nil, err
	}
	uiScene := packer.TransUISceneModelToRaoUIScene(s)

	// 查询当前场景是否有同步关系
	ss := query.Use(dal.DB()).UISceneSync
	sceneSync, err := ss.WithContext(ctx).Where(
		ss.SceneID.Eq(req.SceneID),
		ss.TeamID.Eq(req.TeamID),
	).First()
	if sceneSync != nil {
		syncs, err := us.WithContext(ctx).Where(
			us.SceneID.Eq(sceneSync.SourceSceneID),
			us.TeamID.Eq(req.TeamID),
			us.Status.Eq(consts.TargetStatusNormal),
		).First()

		if err == nil {
			uiScene.SyncMode = sceneSync.SyncMode
			uiScene.SourceUIScene = packer.TransUISceneModelToRaoUIScene(syncs)
		}
	}

	return uiScene, nil
}

// List 列表
func List(ctx *gin.Context, req *rao.UISceneListReq) ([]*rao.UIScene, error) {
	s := query.Use(dal.DB()).UIScene

	source := int32(consts.UISceneSource)
	if req.Source > 0 {
		source = req.Source
	}
	conditions := make([]gen.Condition, 0)
	conditions = append(conditions, s.TeamID.Eq(req.TeamID))
	conditions = append(conditions, s.Status.Eq(consts.UISceneStatusNormal))
	conditions = append(conditions, s.Source.Eq(source))
	if len(req.PlanID) > 0 && req.Source == consts.UISceneSourcePlan {
		conditions = append(conditions, s.PlanID.Eq(req.PlanID))
	}

	scenes, err := s.WithContext(ctx).Where(conditions...).Order(s.Sort, s.CreatedAt.Desc()).Find()
	if err != nil {
		return nil, err
	}

	ret := make([]*rao.UIScene, 0, len(scenes))
	for _, s := range scenes {
		scene := &rao.UIScene{
			SceneID:     s.SceneID,
			SceneType:   s.SceneType,
			TeamID:      s.TeamID,
			ParentID:    s.ParentID,
			Name:        s.Name,
			Sort:        s.Sort,
			Version:     s.Version,
			Description: s.Description,
		}

		ret = append(ret, scene)
	}

	return ret, nil
}

// StopScene 停止记录
func StopScene(ctx *gin.Context, req *rao.StopUISceneReq) error {
	// 发送停止计划状态变更信息
	statusChangeKey := consts.UIReportStatusChange + req.RunID
	statusChangeValue := rao.UIReportStatusChange{
		Status: "stop",
	}
	statusChangeValueString, err := json.Marshal(statusChangeValue)
	if err == nil {
		// 发送计划相关信息到redis频道
		_, err = dal.GetRDB().Publish(ctx, statusChangeKey, string(statusChangeValueString)).Result()
		if err != nil {
			log.Logger.Error("停止--发送压测计划状态变更到对应频道失败")
			return err
		}
	} else {
		log.Logger.Error("停止--发送压测计划状态变更到对应频道，压缩数据失败")
		return err
	}

	// 删除 Redis 对应关系
	redisKey := consts.UIEngineRunAddrPrefix + req.RunID
	addr, err := dal.GetRDB().Get(ctx, redisKey).Result()
	if err != nil {
		log.Logger.Info("StopScene--运行ID对应的机器失败：", err)
	}
	dal.GetRDB().SRem(ctx, consts.UIEngineCurrentRunPrefix+addr, req.RunID)
	dal.GetRDB().Del(ctx, consts.UIEngineRunAddrPrefix+req.RunID)

	return nil
}

// FormatRunUiEngineByScene 通过场景组装 UI 发送的数据
func FormatRunUiEngineByScene(
	ctx context.Context,
	runID,
	teamID,
	sceneID string,
	sendOperatorIDs []string,
) (*ui.Scene, *mao.UISendScene, error) {
	us := query.Use(dal.DB()).UIScene
	scene, err := us.WithContext(ctx).Where(
		us.SceneID.Eq(sceneID),
		us.TeamID.Eq(teamID),
		us.Status.Eq(consts.TargetStatusNormal),
	).First()
	if err != nil {
		return nil, nil, err
	}

	// 查询步骤
	usp := query.Use(dal.DB()).UISceneOperator
	conditions := make([]gen.Condition, 0)
	conditions = append(conditions, usp.SceneID.Eq(sceneID))
	if len(sendOperatorIDs) > 0 {
		conditions = append(conditions, usp.OperatorID.In(sendOperatorIDs...))
	}
	operators, err := usp.WithContext(ctx).Where(conditions...).Order(usp.Sort).Find()
	if err != nil {
		return nil, nil, err
	}

	var operatorIDs []string
	for _, o := range operators {
		operatorIDs = append(operatorIDs, o.OperatorID)
	}

	if len(operatorIDs) == 0 {
		return nil, nil, err
	}

	// 查询接口详情数据
	var sceneOperators []*mao.SceneOperator
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectUISceneOperator)
	cursor, err := collection.Find(ctx, bson.D{{"operator_id", bson.D{{"$in", operatorIDs}}}})
	if err != nil {
		return nil, nil, err
	}
	if err = cursor.All(ctx, &sceneOperators); err != nil {
		return nil, nil, err
	}

	// 获取当前步骤依赖的元素
	se := query.Use(dal.DB()).UISceneElement
	sceneElementList, err := se.WithContext(ctx).Where(
		se.OperatorID.In(operatorIDs...),
		se.SceneID.Eq(sceneID),
		se.Status.Eq(consts.UISceneStatusNormal),
	).Find()
	if err != nil {
		return nil, nil, err
	}
	elementIDs := make([]string, 0, len(sceneElementList))
	for _, e := range sceneElementList {
		elementIDs = append(elementIDs, e.ElementID)
	}

	e := query.Use(dal.DB()).Element
	elementList, err := e.WithContext(ctx).Where(e.ElementID.In(elementIDs...)).Find()
	if err != nil {
		return nil, nil, err
	}

	uiOperators, uiSendSceneOperators := packer.TransSceneOperatorsToRaoSendOperators(operators, runID, scene, sceneOperators, elementList)

	// 排序
	sort.Sort(ByParentSort(uiOperators))
	parentMap := make(map[string][]*ui.Operator)
	for _, op := range uiOperators {
		parentMap[op.ParentId] = append(parentMap[op.ParentId], op)
	}

	formatUiOperators := parentMap["0"]
	addChild(formatUiOperators, parentMap)

	uiScene := &ui.Scene{
		SceneId:   scene.SceneID,
		SceneName: scene.Name,
		Operators: formatUiOperators,
	}

	// 统计报告中的数量
	var assertTotalNum int
	for _, reportOperator := range uiSendSceneOperators {
		assertTotalNum += reportOperator.AssertTotalNum
	}

	uiReportScene := &mao.UISendScene{
		SceneID:            scene.SceneID,
		Name:               scene.Name,
		OperatorTotalNum:   len(uiOperators),
		OperatorSuccessNum: 0,
		OperatorErrorNum:   0,
		OperatorUnExecNum:  len(uiOperators),
		Operators:          uiSendSceneOperators,
		AssertTotalNum:     assertTotalNum,
	}

	return uiScene, uiReportScene, nil
}

// Send 调试
func Send(ctx *gin.Context, userID string, teamID string, sceneID string, planID string, sendOperatorIDs []string) (string, error) {
	var (
		uiMachineKey  string
		modelBrowsers string
	)

	us := query.Use(dal.DB()).UIScene
	scene, err := us.WithContext(ctx).Where(
		us.SceneID.Eq(sceneID),
		us.TeamID.Eq(teamID),
		us.Status.Eq(consts.TargetStatusNormal),
	).First()
	if err != nil {
		return "", err
	}

	// 调试如果传计划ID，按计划的配置信息走
	uiMachineKey = scene.UIMachineKey
	modelBrowsers = scene.Browsers
	if len(planID) > 0 {
		up := query.Use(dal.DB()).UIPlan
		plan, err := up.WithContext(ctx).Where(
			up.PlanID.Eq(planID),
			up.TeamID.Eq(teamID),
		).First()
		if err != nil {
			return "", err
		}
		uiMachineKey = plan.UIMachineKey
		modelBrowsers = plan.Browsers
	}

	// step1: 获取发送机器
	uiEngineList, err := clients.GetUiEngineMachineList()
	if err != nil {
		return "", errors.New("get ui_engine empty" + err.Error())
	}

	uiEngine := &rao.UiEngineMachineInfo{}
	for _, info := range uiEngineList {
		if info.Key == uiMachineKey {
			uiEngine = info
		}
	}
	if len(uiEngine.IP) == 0 {
		uiEngine = uiEngineList[rand.Intn(len(uiEngineList))]
	}

	// step4: 获取浏览器信息
	browsers := make([]*rao.Browser, 0)
	_ = json.Unmarshal([]byte(modelBrowsers), &browsers)
	if strings.Contains(strings.ToLower(uiEngine.SystemInfo.SystemBasic), "linux") {
		for _, b := range browsers {
			if !b.Headless {
				return "", errmsg.ErrSendLinuxNotQTMode
			}
		}
	}

	// step2: 组织场景操作数据
	runID := uuid.GetUUID()
	uiScene, uiSendScene, err := FormatRunUiEngineByScene(ctx, runID, teamID, sceneID, sendOperatorIDs)
	if err != nil {
		log.Logger.Error("FormatRunUiEngineByScene err", err)
		return "", err
	}

	// step3: 生成操作记录
	var docs []interface{}
	for _, run := range uiSendScene.Operators {
		docs = append(docs, run)
	}
	if _, err := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectUISendSceneOperatorDebug).InsertMany(ctx, docs); err != nil {
		log.Logger.Error("调试日志入库失败", err)
		return "", err
	}

	uiBrowsers := make([]*ui.Browser, 0)
	for _, browser := range browsers {
		b := &ui.Browser{Headless: false}
		if err = copier.Copy(b, browser); err != nil {
			log.Logger.Error("ui.Browser Copy err", proof.WithError(err))
		}
		uiBrowsers = append(uiBrowsers, b)
	}

	uiScenes := make([]*ui.Scene, 0)
	uiScenes = append(uiScenes, uiScene)

	// step5: 调用自动化接口
	if _, err = clients.RunUiEngine(ctx, uiEngine.IP, &ui.RunRequest{
		Topic:         runID,
		SceneRunOrder: consts.UIPlanSceneRunModeOrder,
		UserId:        userID,
		Browsers:      uiBrowsers,
		Scenes:        uiScenes,
	}); err != nil {
		return "", err
	}

	// step6: 添加自动化记录
	dal.GetRDB().SAdd(ctx, consts.UIEngineCurrentRunPrefix+uiEngine.IP, runID)
	dal.GetRDB().Set(ctx, consts.UIEngineRunAddrPrefix+runID, uiEngine.IP, time.Second*3600)

	return runID, err
}

// Sort 排序
func Sort(ctx *gin.Context, userID string, req *rao.UIScenesSortReq) error {
	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		for _, s := range req.Scenes {
			_, err := tx.UIScene.WithContext(ctx).Where(
				tx.UIScene.TeamID.Eq(s.TeamID),
				tx.UIScene.SceneID.Neq(s.SceneID),
				tx.UIScene.Status.Eq(consts.UISceneStatusNormal),
				tx.UIScene.Name.Eq(s.Name),
				tx.UIScene.ParentID.Eq(s.ParentID),
				tx.UIScene.PlanID.Eq(s.PlanID),
				tx.UIScene.Source.Eq(s.Source),
			).First()
			if err == nil {
				return errmsg.ErrTargetSortNameAlreadyExist
			}

			_, err = tx.UIScene.WithContext(ctx).Where(
				tx.UIScene.TeamID.Eq(s.TeamID),
				tx.UIScene.SceneID.Eq(s.SceneID),
			).UpdateSimple(
				tx.UIScene.Sort.Value(s.Sort),
				tx.UIScene.ParentID.Value(s.ParentID))
			if err != nil {
				return err
			}
		}
		return nil
	})

	return err
}

// Trash 移动到回收站
func Trash(ctx *gin.Context, req *rao.UISceneTrashReq, userID string) error {
	sceneID := req.SceneID
	teamID := req.TeamID
	return dal.GetQuery().Transaction(func(tx *query.Query) error {
		t, err := tx.UIScene.WithContext(ctx).Where(tx.UIScene.SceneID.Eq(sceneID)).First()
		if err != nil {
			return err
		}

		// 删除
		_ = getAllSonSceneID(ctx, sceneID, teamID, userID, t.SceneType)

		//var operate int32 = 0
		//if t.SceneType == consts.UISceneTypeFolder {
		//	operate = record.OperationOperateDeleteUISceneFolder
		//} else {
		//	operate = record.OperationOperateDeleteUISceneApi
		//}
		//if err := record.InsertDelete(ctx, t.TeamID, userID, operate, t.Name); err != nil {
		//	return err
		//}

		return nil
	})
}

func getAllSonSceneID(ctx *gin.Context, sceneID string, teamID string, userID string, sceneType string) error {
	tx := dal.GetQuery().UIScene
	if sceneType == consts.UISceneTypeFolder {
		// 查询这个目录下是否还有别的目录或文件
		list, err := tx.WithContext(ctx).Where(tx.ParentID.Eq(sceneID)).Find()
		if err != nil {
			return err
		}
		if len(list) > 0 {
			for _, tInfo := range list {
				_ = getAllSonSceneID(ctx, tInfo.SceneID, tInfo.TeamID, userID, tInfo.SceneType)
			}
		}
		// 删除目录本身
		if _, err = tx.WithContext(ctx).Where(tx.SceneID.Eq(sceneID)).UpdateSimple(tx.Status.Value(consts.UISceneStatusTrash)); err != nil {
			return err
		}
	} else {
		if _, err := tx.WithContext(ctx).Where(tx.SceneID.Eq(sceneID)).UpdateSimple(tx.Status.Value(consts.UISceneStatusTrash)); err != nil {
			return err
		}
		// 删除关联关系
		se := dal.GetQuery().UISceneElement
		if _, err := se.WithContext(ctx).Where(
			se.SceneID.Eq(sceneID),
			se.TeamID.Eq(teamID),
		).Update(se.Status, consts.UISceneStatusTrash); err != nil {
			return err
		}
	}

	// 添加到回收站
	st := dal.GetQuery().UISceneTrash
	sceneTrash := &model.UISceneTrash{
		SceneID:       sceneID,
		TeamID:        teamID,
		CreatedUserID: userID,
	}
	if err := st.WithContext(ctx).Create(sceneTrash); err != nil {
		return err
	}
	return nil
}

func TrashList(ctx *gin.Context, req *rao.UISceneTrashListReq) ([]*rao.UISceneTrash, int64, error) {
	limit := req.Size
	offset := (req.Page - 1) * req.Size
	keyword := strings.TrimSpace(req.Keyword)

	// 获取当前时间
	now := time.Now()
	// 计算30天前的时间
	thirtyDaysAgo := now.AddDate(0, 0, -30)

	s := query.Use(dal.DB()).UIScene
	ut := query.Use(dal.DB()).UISceneTrash
	conditions := make([]gen.Condition, 0)
	conditions = append(conditions, ut.TeamID.Eq(req.TeamID))
	conditions = append(conditions, ut.CreatedAt.Gte(thirtyDaysAgo))

	if len(keyword) > 0 {
		var keywordSceneID = make([]string, 0)
		if err := s.WithContext(ctx).Where(s.Name.Like(fmt.Sprintf("%%%s%%", keyword))).Pluck(s.SceneID, &keywordSceneID); err != nil {
			return nil, 0, err
		}
		conditions = append(conditions, ut.SceneID.In(keywordSceneID...))
	}

	trashList, total, err := ut.WithContext(ctx).Where(conditions...).Order(ut.ID.Desc()).FindByPage(offset, limit)
	if err != nil {
		return nil, 0, err
	}

	sceneIDs := make([]string, 0, len(trashList))
	userIDs := make([]string, 0, len(trashList))
	for _, t := range trashList {
		sceneIDs = append(sceneIDs, t.SceneID)
		userIDs = append(userIDs, t.CreatedUserID)
	}

	sceneIDs = public.SliceUnique(sceneIDs)
	userIDs = public.SliceUnique(userIDs)
	sceneList, err := s.WithContext(ctx).Select(s.Name, s.SceneID, s.SceneType).Where(s.SceneID.In(sceneIDs...)).Find()
	if err != nil {
		return nil, 0, err
	}

	u := query.Use(dal.DB()).User
	users, err := u.WithContext(ctx).Select(u.UserID, u.Nickname).Where(u.UserID.In(userIDs...)).Find()
	if err != nil {
		return nil, 0, err
	}

	return packer.TransSceneTrashListToRaoTrash(trashList, sceneList, users), total, nil
}

// Recall 恢复
func Recall(ctx *gin.Context, sceneIDs []string, teamID string, userID string) error {
	s := query.Use(dal.DB()).UIScene

	uiSceneList, err := s.WithContext(ctx).Where(s.SceneID.In(sceneIDs...)).Find()
	if err != nil {
		return err
	}

	return dal.GetQuery().Transaction(func(tx *query.Query) error {
		for _, scene := range uiSceneList {
			if err := RecallSonScene(ctx, scene.SceneID, scene.TeamID, userID, scene.SceneType); err != nil {
				return err
			}
		}

		return nil
	})
}

// RecallSonScene 恢复目录，若目录有场景一起恢复
func RecallSonScene(ctx *gin.Context, sceneID string, teamID string, userID string, sceneType string) error {
	tx := dal.GetQuery().UIScene
	if sceneType == consts.UISceneTypeFolder {
		// 查询这个目录下是否还有别的目录或文件
		list, err := tx.WithContext(ctx).Where(tx.ParentID.Eq(sceneID)).Find()
		if err != nil {
			return err
		}
		// 恢复目录本身
		if err = HandleRecall(ctx, teamID, sceneID); err != nil {
			return err
		}
		if len(list) > 0 {
			for _, tInfo := range list {
				_ = RecallSonScene(ctx, tInfo.SceneID, tInfo.TeamID, userID, tInfo.SceneType)
			}
		}
	} else {
		if err := HandleRecall(ctx, teamID, sceneID); err != nil {
			return err
		}
	}

	return nil
}

// HandleRecall 处理恢复
func HandleRecall(ctx *gin.Context, teamID, sceneID string) error {
	tx := dal.GetQuery().UIScene
	txt := dal.GetQuery().UISceneTrash
	scene, err := tx.WithContext(ctx).Where(
		tx.TeamID.Eq(teamID),
		tx.SceneID.Eq(sceneID)).First()
	if err != nil {
		return err
	}
	newSceneName := scene.Name
	_, err = tx.WithContext(ctx).Where(
		tx.TeamID.Eq(teamID),
		tx.Name.Eq(newSceneName),
		tx.SceneType.Eq(scene.SceneType),
		tx.SceneID.Neq(scene.SceneID),
		tx.Source.Eq(scene.Source),
		tx.ParentID.Eq(scene.ParentID),
		tx.PlanID.Eq(scene.PlanID),
		tx.Status.Eq(consts.UISceneStatusNormal)).First()
	if err == nil { // 重名
		newSceneName = scene.Name + "_恢复"
		list, err := tx.WithContext(ctx).Where(
			tx.TeamID.Eq(teamID),
			tx.Name.Like(fmt.Sprintf("%s%%", newSceneName)),
			tx.SceneType.Eq(scene.SceneType),
			tx.SceneID.Neq(scene.SceneID),
			tx.Source.Eq(scene.Source),
			tx.ParentID.Eq(scene.ParentID),
			tx.PlanID.Eq(scene.PlanID),
			tx.Status.Eq(consts.UISceneStatusNormal),
		).Find()
		if err == nil && len(list) > 0 {
			// 有复制过得配置
			maxNum := 0
			for _, targetInfo := range list {
				nameTmp := targetInfo.Name
				postfixSlice := strings.Split(nameTmp, "_恢复")
				if len(postfixSlice) < 2 {
					continue
				}
				currentNum, err := strconv.Atoi(postfixSlice[len(postfixSlice)-1])
				if err != nil {
					log.Logger.Info("复制自动化计划--类型转换失败，err:", err)
					currentNum = 1
				}
				if currentNum > maxNum {
					maxNum = currentNum
				}
			}
			newSceneName = scene.Name + fmt.Sprintf("_恢复%d", maxNum+1)
		}
	}

	// 若不存在父级目录，恢复到根目录
	parentID := scene.ParentID
	if scene.SceneType == consts.UISceneTypeScene {
		if _, err = tx.WithContext(ctx).Where(
			tx.TeamID.Eq(teamID),
			tx.SceneType.Eq(consts.UISceneTypeFolder),
			tx.SceneID.Eq(scene.ParentID),
			tx.Source.Eq(scene.Source),
			tx.PlanID.Eq(scene.PlanID),
			tx.Status.Eq(consts.UISceneStatusNormal)).First(); err != nil {
			parentID = "0"
		}
	}

	updateData := make(map[string]interface{}, 2)
	updateData["status"] = consts.UISceneStatusNormal
	updateData["name"] = newSceneName
	updateData["parent_id"] = parentID
	if _, err := tx.WithContext(ctx).Where(tx.SceneID.Eq(sceneID)).Updates(updateData); err != nil {
		return err
	}

	if _, err = txt.WithContext(ctx).Where(
		txt.SceneID.Eq(sceneID),
		txt.TeamID.Eq(teamID)).Delete(); err != nil {
		return err
	}

	se := dal.GetQuery().UISceneElement
	if _, err := se.WithContext(ctx).Where(
		se.SceneID.Eq(sceneID),
		se.TeamID.Eq(teamID),
	).Update(se.Status, consts.UISceneStatusNormal); err != nil {
		return err
	}

	return err
}

// Delete 删除
func Delete(ctx *gin.Context, sceneIDs []string, teamID string, userID string) error {
	return dal.GetQuery().Transaction(func(tx *query.Query) error {
		if _, err := tx.UIScene.WithContext(ctx).Where(
			tx.UIScene.SceneID.In(sceneIDs...),
			tx.UIScene.TeamID.Eq(teamID),
		).Delete(); err != nil {
			return err
		}

		if _, err := tx.UISceneTrash.WithContext(ctx).Where(
			tx.UISceneTrash.SceneID.In(sceneIDs...),
			tx.UISceneTrash.TeamID.Eq(teamID),
		).Delete(); err != nil {
			return err
		}

		if _, err := tx.UISceneOperator.WithContext(ctx).Where(
			tx.UISceneOperator.SceneID.In(sceneIDs...),
		).Delete(); err != nil {
			return err
		}

		if _, err := tx.UISceneElement.WithContext(ctx).Where(
			tx.UISceneElement.SceneID.In(sceneIDs...),
			tx.UISceneElement.TeamID.Eq(teamID),
		).Delete(); err != nil {
			return err
		}

		return nil
	})
}

// SyncBindScene 同步绑定的场景
func SyncBindScene(ctx *gin.Context, sceneID, teamID, userID string) error {
	return nil
	// step1: 查询当前场景是否有实时同步
	//var syncSceneIDs = make([]string, 0)
	//ss := query.Use(dal.DB()).UISceneSync
	//if err := ss.WithContext(ctx).Where(
	//	ss.SourceSceneID.Eq(sceneID),
	//	ss.TeamID.Eq(teamID),
	//	ss.SyncMode.Eq(consts.UISceneOptSyncModeAuto),
	//).Pluck(ss.SceneID, &syncSceneIDs); err != nil {
	//	return err
	//}
	//
	//if len(syncSceneIDs) == 0 {
	//	return nil
	//}
	//
	//// 查询原数据
	//scene, err := DetailBySceneID(ctx, &rao.UISceneDetailReq{
	//	TeamID:  teamID,
	//	SceneID: sceneID,
	//})
	//if err != nil {
	//	return err
	//}
	//
	//s := query.Use(dal.DB()).UIScene
	//sceneList, err := s.WithContext(ctx).Where(s.SceneID.In(syncSceneIDs...), s.TeamID.Eq(teamID)).Find()
	//if err != nil {
	//	return err
	//}
	//for _, s := range sceneList {
	//	req := &rao.UISceneSaveReq{
	//		SceneID:     s.SceneID,
	//		TeamID:      s.TeamID,
	//		ParentID:    s.ParentID,
	//		Name:        scene.Name,
	//		Sort:        s.Sort,
	//		Version:     scene.Version,
	//		Description: scene.Description,
	//		Browsers:    scene.Browsers,
	//		PlanID:      s.PlanID,
	//		Source:      s.Source,
	//	}
	//	if err := Update(ctx, userID, req); err != nil {
	//		return err
	//	}
	//}

	//return nil
}

// SyncScene 同步场景
func SyncScene(ctx *gin.Context, sourceSceneID, sceneID, teamID, userID string) error {
	// step1: 查询原数据
	_, err := DetailBySceneID(ctx, &rao.UISceneDetailReq{
		TeamID:  teamID,
		SceneID: sourceSceneID,
	})
	if err != nil {
		return err
	}

	s := query.Use(dal.DB()).UIScene
	scene, err := s.WithContext(ctx).Where(s.SceneID.Eq(sceneID), s.TeamID.Eq(teamID)).First()
	if err != nil {
		return err
	}

	// step2:  同步基本信息
	//req := &rao.UISceneSaveReq{
	//	SceneID:     scene.SceneID,
	//	TeamID:      scene.TeamID,
	//	ParentID:    scene.ParentID,
	//	Name:        sourceScene.Name,
	//	Sort:        scene.Sort,
	//	Version:     sourceScene.Version,
	//	Description: sourceScene.Description,
	//	Browsers:    sourceScene.Browsers,
	//	PlanID:      scene.PlanID,
	//	Source:      scene.Source,
	//}
	//if err := Update(ctx, userID, req); err != nil {
	//	return err
	//}

	// step3: 查询源场景步骤基本信息
	so := query.Use(dal.DB()).UISceneOperator
	sourceOperatorIDs := make([]string, 0)
	operators, err := so.WithContext(ctx).Where(so.SceneID.Eq(sourceSceneID)).Find()
	if err != nil {
		log.Logger.Error("SyncBindSceneOperator ListOperator err", proof.WithError(err))
		return err
	}
	for _, o := range operators {
		sourceOperatorIDs = append(sourceOperatorIDs, o.OperatorID)
	}

	// step4: 查询源场景步骤详细
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectUISceneOperator)
	cursor, err := collection.Find(ctx, bson.D{{"operator_id", bson.D{{"$in", sourceOperatorIDs}}}})
	if err != nil {
		log.Logger.Error("collection.Find err", proof.WithError(err))
		return err
	}
	var sceneOperators []*mao.SceneOperator
	if err = cursor.All(ctx, &sceneOperators); err != nil {
		log.Logger.Error("cursor.All err", proof.WithError(err))
		return err
	}

	//  生成新步骤
	if err = handleNewOperator(ctx, operators, scene, sceneOperators); err != nil {
		return err
	}

	return nil
}

// handleNewOperator  生成新步骤
func handleNewOperator(
	ctx *gin.Context,
	operators []*model.UISceneOperator,
	scene *model.UIScene,
	sceneOperators []*mao.SceneOperator,
) error {
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectUISceneOperator)
	// step5: 生成新步骤ID
	newOperatorIDs := make(map[string]string)
	newOperatorIDs["0"] = "0"
	for _, o := range operators {
		newOperatorIDs[o.OperatorID] = uuid.GetUUID()
	}

	// step6: 新步骤ID基本数据
	var uiSceneOperators = make([]*model.UISceneOperator, 0, len(operators))
	for _, o := range operators {
		uiSceneOperator := &model.UISceneOperator{
			OperatorID: newOperatorIDs[o.OperatorID],
			SceneID:    scene.SceneID,
			Name:       o.Name,
			ParentID:   newOperatorIDs[o.ParentID],
			Sort:       o.Sort,
			Status:     o.Status,
			Type:       o.Type,
			Action:     o.Action,
		}
		uiSceneOperators = append(uiSceneOperators, uiSceneOperator)
	}

	// step7: 新步骤ID详细数据
	collectSceneOperators := make([]interface{}, 0, len(sceneOperators))
	for _, so := range sceneOperators {
		sceneOperator := &mao.SceneOperator{
			SceneID:       scene.SceneID,
			OperatorID:    newOperatorIDs[so.OperatorID],
			ActionDetail:  so.ActionDetail,
			Settings:      so.Settings,
			Asserts:       so.Asserts,
			DataWithdraws: so.DataWithdraws,
		}
		collectSceneOperators = append(collectSceneOperators, sceneOperator)
	}

	// step8: 开启事务：添加操作 MySQL、MongoDB 删除之前的步骤数据
	if err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		if _, err := tx.UISceneOperator.WithContext(ctx).Where(tx.UISceneOperator.SceneID.Eq(scene.SceneID)).Delete(); err != nil {
			return err
		}

		if err := tx.UISceneOperator.WithContext(ctx).Create(uiSceneOperators...); err != nil {
			return err
		}

		// 删除接口用例详情
		_, _ = collection.DeleteMany(ctx, bson.D{{"scene_id", scene.SceneID}})

		if _, err := collection.InsertMany(ctx, collectSceneOperators); err != nil {
			return err
		}

		return nil
	}); err != nil {
		log.Logger.Error("Transaction err", proof.WithError(err))
		return err
	}

	return nil
}

// Copy 复制
func Copy(ctx *gin.Context, userID string, req *rao.UISceneCopyReq) error {
	newTime := time.Now()
	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		// step1 复制计划下场景,分组
		oldTarget, err := tx.UIScene.WithContext(ctx).Where(
			tx.UIScene.TeamID.Eq(req.TeamID),
			tx.UIScene.SceneID.Eq(req.SceneID),
			tx.UIScene.Status.Eq(consts.UISceneStatusNormal),
			tx.UIScene.SceneType.In(consts.UISceneTypeScene),
		).Order(tx.UIScene.ParentID).First()
		if err != nil {
			return err
		}

		// step2 重名
		newSceneName := oldTarget.Name + "_1"
		_, err = tx.UIScene.WithContext(ctx).Where(
			tx.UIScene.TeamID.Eq(req.TeamID),
			tx.UIScene.Name.Eq(newSceneName),
			tx.UIScene.SceneType.Eq(consts.UISceneTypeScene),
			tx.UIScene.Status.Eq(consts.UISceneStatusNormal),
			tx.UIScene.ParentID.Eq(oldTarget.ParentID),
			tx.UIScene.Source.Eq(oldTarget.Source),
			tx.UIScene.PlanID.Eq(oldTarget.PlanID),
		).First()
		if err == nil { // 重名了
			list, err := tx.UIScene.WithContext(ctx).Where(
				tx.UIScene.TeamID.Eq(req.TeamID),
				tx.UIScene.Name.Like(fmt.Sprintf("%s%%", oldTarget.Name+"_")),
				tx.UIScene.SceneType.Eq(consts.UISceneTypeScene),
				tx.UIScene.Status.Eq(consts.UISceneStatusNormal),
				tx.UIScene.ParentID.Eq(oldTarget.ParentID),
				tx.UIScene.Source.Eq(oldTarget.Source),
				tx.UIScene.PlanID.Eq(oldTarget.PlanID),
			).Find()
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
				newSceneName = oldTarget.Name + fmt.Sprintf("_%d", maxNum+1)
			}
		}

		nameLength := public.GetStringNum(newSceneName)
		if nameLength > 30 { // 场景名称限制30个字符
			return errmsg.ErrUISceneNameLong
		}

		// 新的sceneID
		newSceneID := uuid.GetUUID()
		oldSceneID := oldTarget.SceneID
		oldTarget.ID = 0
		oldTarget.SceneID = newSceneID
		oldTarget.Name = newSceneName
		oldTarget.CreatedUserID = userID
		oldTarget.RecentUserID = userID
		oldTarget.CreatedAt = newTime
		oldTarget.UpdatedAt = newTime
		if err := tx.UIScene.WithContext(ctx).Create(oldTarget); err != nil {
			return err
		}

		operators, err := tx.UISceneOperator.WithContext(ctx).Where(
			tx.UISceneOperator.SceneID.Eq(oldSceneID)).Order(tx.UISceneOperator.ParentID).Find()
		if err != nil {
			return err
		}

		// 步骤为空  无需处理
		if len(operators) == 0 {
			return nil
		}
		sourceOperatorIDs := make([]string, 0)
		for _, o := range operators {
			sourceOperatorIDs = append(sourceOperatorIDs, o.OperatorID)
		}

		// step4: 查询源场景步骤详细
		collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectUISceneOperator)
		cursor, err := collection.Find(ctx, bson.D{{"operator_id", bson.D{{"$in", sourceOperatorIDs}}}})
		if err != nil {
			log.Logger.Error("collection.Find err", proof.WithError(err))
			return err
		}
		var sceneOperators []*mao.SceneOperator
		if err = cursor.All(ctx, &sceneOperators); err != nil {
			log.Logger.Error("cursor.All err", proof.WithError(err))
			return err
		}

		// step5: 生成新步骤ID
		newOperatorIDs := make(map[string]string)
		newOperatorIDs["0"] = "0"
		for _, o := range operators {
			newOperatorIDs[o.OperatorID] = uuid.GetUUID()
		}

		// step6: 新步骤ID基本数据
		var uiSceneOperators = make([]*model.UISceneOperator, 0, len(newOperatorIDs))
		for _, o := range operators {
			uiSceneOperator := &model.UISceneOperator{
				OperatorID: newOperatorIDs[o.OperatorID],
				SceneID:    newSceneID,
				Name:       o.Name,
				ParentID:   newOperatorIDs[o.ParentID],
				Sort:       o.Sort,
				Status:     o.Status,
				Type:       o.Type,
				Action:     o.Action,
			}
			uiSceneOperators = append(uiSceneOperators, uiSceneOperator)
		}

		// step7: 新步骤ID详细数据
		collectSceneOperators := make([]interface{}, 0, len(sceneOperators))
		for _, so := range sceneOperators {
			so.OperatorID = newOperatorIDs[so.OperatorID]
			so.SceneID = newSceneID
			collectSceneOperators = append(collectSceneOperators, so)
		}

		// step8: 复制元素与场景关联关系
		oldElements := make([]string, 0)
		if err = tx.UISceneElement.WithContext(ctx).Where(
			tx.UISceneElement.TeamID.Eq(req.TeamID),
			tx.UISceneElement.SceneID.Eq(req.SceneID)).Pluck(tx.UISceneElement.ElementID, &oldElements); err != nil {
			return err
		}
		if len(oldElements) > 0 {
			oldElements = public.SliceUnique(oldElements)
			newElements := make([]*model.UISceneElement, 0, len(oldElements))
			for _, elementID := range oldElements {
				newElement := &model.UISceneElement{
					SceneID:   newSceneID,
					ElementID: elementID,
					TeamID:    req.TeamID,
					Status:    consts.UISceneStatusNormal,
					CreatedAt: newTime,
					UpdatedAt: newTime,
				}
				newElements = append(newElements, newElement)
			}
			if err = tx.UISceneElement.WithContext(ctx).Create(newElements...); err != nil {
				return err
			}
		}

		// step9: 开启事务：添加操作 MySQL、MongoDB
		if err := tx.UISceneOperator.WithContext(ctx).Create(uiSceneOperators...); err != nil {
			return err
		}

		if _, err := collection.InsertMany(ctx, collectSceneOperators); err != nil {
			return err
		}

		return nil
	})

	return err
}
