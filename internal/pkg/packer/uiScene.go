package packer

import (
	"encoding/json"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/api/ui"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/log"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/mao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/go-omnibus/proof"
	"github.com/jinzhu/copier"
	"go.mongodb.org/mongo-driver/bson"
	"strconv"
	"strings"
	"time"
)

func TransUISceneModelToRaoUIScene(s *model.UIScene) *rao.UIScene {
	browsers := make([]*rao.Browser, 0)
	_ = json.Unmarshal([]byte(s.Browsers), &browsers)

	return &rao.UIScene{
		SceneID:      s.SceneID,
		SceneType:    s.SceneType,
		TeamID:       s.TeamID,
		ParentID:     s.ParentID,
		Name:         s.Name,
		Sort:         s.Sort,
		Version:      s.Version,
		Description:  s.Description,
		Browsers:     browsers,
		UIMachineKey: s.UIMachineKey,
		Source:       s.Source,
		PlanID:       s.PlanID,
	}
}

func TransSaveUISceneFolderReqToUISceneModel(folder *rao.UISceneSaveFolderReq, userID string) *model.UIScene {
	browsers, err := json.Marshal("")
	if err != nil {
		log.Logger.Error("TransSaveReqToUISceneModel.Browsers marshal err", proof.WithError(err))
	}
	return &model.UIScene{
		SceneID:       folder.SceneID,
		SceneType:     consts.UISceneTypeFolder,
		TeamID:        folder.TeamID,
		Name:          folder.Name,
		ParentID:      folder.ParentID,
		Sort:          folder.Sort,
		Status:        consts.UISceneStatusNormal,
		Version:       folder.Version,
		CreatedUserID: userID,
		RecentUserID:  userID,
		Description:   folder.Description,
		Browsers:      string(browsers),
		Source:        folder.Source,
		PlanID:        folder.PlanID,
	}
}

func TransUISceneToRaoMockUISceneFolder(t *model.UIScene) *rao.UISceneFolder {
	return &rao.UISceneFolder{
		SceneID:     t.SceneID,
		TeamID:      t.TeamID,
		ParentID:    t.ParentID,
		Name:        t.Name,
		Sort:        t.Sort,
		Version:     t.Version,
		Description: t.Description,
		Source:      t.Source,
		PlanID:      t.PlanID,
	}
}

func TransSaveReqToUISceneModel(scene *rao.UISceneSaveReq, userID string) *model.UIScene {
	browsers, err := json.Marshal(scene.Browsers)
	if err != nil {
		log.Logger.Error("TransSaveReqToUISceneModel.Browsers marshal err", proof.WithError(err))
	}

	var description string
	if scene.Description != nil {
		description = *scene.Description
	}
	return &model.UIScene{
		SceneID:       scene.SceneID,
		TeamID:        scene.TeamID,
		SceneType:     consts.UISceneTypeScene,
		Name:          scene.Name,
		ParentID:      scene.ParentID,
		Sort:          scene.Sort,
		Status:        consts.UISceneStatusNormal,
		Source:        scene.Source,
		Version:       scene.Version,
		CreatedUserID: userID,
		RecentUserID:  userID,
		Description:   description,
		Browsers:      string(browsers),
		UIMachineKey:  scene.UIMachineKey,
		PlanID:        scene.PlanID,
	}
}

func TransSaveOperatorReqToOperatorModel(req *rao.UISceneSaveOperatorReq) *model.UISceneOperator {
	uiSceneOperator := &model.UISceneOperator{
		OperatorID: req.OperatorID,
		SceneID:    req.SceneID,
		Name:       req.Name,
		ParentID:   req.ParentID,
		Sort:       req.Sort,
		Status:     req.Status,
		Type:       req.Type,
		Action:     req.Action,
	}

	return uiSceneOperator
}

func TransSaveOperatorReqToUISceneOperatorMao(scene *rao.UISceneSaveOperatorReq) *mao.SceneOperator {
	detail, err := bson.Marshal(mao.ActionDetail{
		Detail: scene.ActionDetail,
	})
	if err != nil {
		log.Logger.Info("TransSaveReqToUISceneOperatorMao.ActionDetail bson marshal err", proof.WithError(err))
	}

	settings, err := bson.Marshal(mao.Settings{
		Settings: scene.Settings,
	})
	if err != nil {
		log.Logger.Info("TransSaveReqToUISceneOperatorMao.settings bson marshal err", proof.WithError(err))
	}

	asserts, err := bson.Marshal(mao.Asserts{
		Asserts: scene.Asserts,
	})
	if err != nil {
		log.Logger.Info("TransSaveReqToUISceneOperatorMao.asserts bson marshal err", proof.WithError(err))
	}

	dataWithdraws, err := bson.Marshal(mao.DataWithdraws{
		DataWithdraws: scene.DataWithdraws,
	})
	if err != nil {
		log.Logger.Info("TransSaveReqToUISceneOperatorMao.dataWithdraws bson marshal err", proof.WithError(err))
	}

	return &mao.SceneOperator{
		SceneID:       scene.SceneID,
		OperatorID:    scene.OperatorID,
		ActionDetail:  detail,
		Settings:      settings,
		Asserts:       asserts,
		DataWithdraws: dataWithdraws,
	}
}

// TransSaveOperatorReqToUISceneElementMao 保存获取关联 元素ID
func TransSaveOperatorReqToUISceneElementMao(scene *rao.UISceneSaveOperatorReq) map[string]*rao.Element {
	ret := make(map[string]*rao.Element, 0)
	if scene.Action == consts.UISceneOptTypeToggleWindow {
		if scene.ActionDetail.ToggleWindow.Type == consts.UISceneOptTypeSwitchFrameByLocator {
			elementMode := scene.ActionDetail.ToggleWindow.SwitchFrameByLocator.Element
			ret = elementIDCollect(ret, elementMode)
		}
	}

	if scene.Action == consts.UISceneOptTypeMouseClicking {
		elementMode := scene.ActionDetail.MouseClicking.Element
		ret = elementIDCollect(ret, elementMode)
	}

	if scene.Action == consts.UISceneOptTypeMouseScrolling {
		elementMode := scene.ActionDetail.MouseScrolling.Element
		ret = elementIDCollect(ret, elementMode)
	}

	if scene.Action == consts.UISceneOptTypeMouseDragging {
		elementMode := scene.ActionDetail.MouseDragging.Element
		ret = elementIDCollect(ret, elementMode)
		if scene.ActionDetail.MouseDragging.Type == consts.MouseDraggingTypeDragElement {
			targetElementMode := scene.ActionDetail.MouseDragging.TarGetElement
			ret = elementIDCollect(ret, targetElementMode)
		}
	}

	if scene.Action == consts.UISceneOptTypeInputOperations {
		elementMode := scene.ActionDetail.InputOperations.Element
		ret = elementIDCollect(ret, elementMode)
	}

	if scene.Action == consts.UISceneOptTypeWaitEvents {
		elementMode := scene.ActionDetail.WaitEvents.Element
		ret = elementIDCollect(ret, elementMode)
	}

	if scene.Action == consts.UISceneOptTypeIfCondition {
		if scene.ActionDetail.IfCondition != nil {
			conditionOperators := scene.ActionDetail.IfCondition.ConditionOperators
			if conditionOperators != nil {
				for _, conditionOperator := range conditionOperators {
					if conditionOperator.Type == consts.IfConditionTypeConditionOperatorAssert &&
						conditionOperator.Assert != nil {
						ret = assertElementIDCollect(ret, conditionOperator.Assert)
					}
				}
			}
		}
	}

	if scene.Action == consts.UISceneOptTypeCodeOperation {
		elementMode := scene.ActionDetail.CodeOperation.Element
		ret = elementIDCollect(ret, elementMode)
	}

	if scene.Action == consts.UISceneOptTypeAssert {
		ret = assertElementIDCollect(ret, scene.ActionDetail.Assert)
	}

	if scene.Action == consts.UISceneOptTypeDataWithdraw {
		ret = withdrawElementIDCollect(ret, scene.ActionDetail.DataWithdraw)
	}

	for _, a := range scene.Asserts {
		ret = assertElementIDCollect(ret, a)
	}

	for _, d := range scene.DataWithdraws {
		ret = withdrawElementIDCollect(ret, d)
	}

	return ret
}

// assertElementIDCollect 断言中元素ID集合
func assertElementIDCollect(ret map[string]*rao.Element, a *rao.AutomationAssert) map[string]*rao.Element {
	if a.Type == consts.UISceneOptElementExists {
		element := a.Element
		ret = elementIDCollect(ret, element)
	}
	if a.Type == consts.UISceneOptElementNotExists {
		element := a.Element
		ret = elementIDCollect(ret, element)
	}
	if a.Type == consts.UISceneOptElementDisplayed {
		element := a.Element
		ret = elementIDCollect(ret, element)
	}
	if a.Type == consts.UISceneOptElementNotDisplayed {
		element := a.Element
		ret = elementIDCollect(ret, element)
	}
	if a.Type == consts.UISceneOptElementAttributeAssertion {
		element := a.Element
		ret = elementIDCollect(ret, element)
	}

	return ret
}

// withdrawElementIDCollect 数据提取中元素ID集合
func withdrawElementIDCollect(ret map[string]*rao.Element, a *rao.DataWithdraw) map[string]*rao.Element {
	if a.WithdrawType == consts.UISceneWithdrawTypeElement {
		if a.ElementMethod == nil {
			return ret
		}
		element := a.ElementMethod.Element
		ret = elementIDCollect(ret, element)
	}

	return ret
}

// elementIDCollect 元素ID集合
func elementIDCollect(ret map[string]*rao.Element, actualElement *rao.Element) map[string]*rao.Element {
	element := &rao.Element{}
	if err := copier.Copy(element, actualElement); err != nil {
		log.Logger.Error("elementIDCollect Copy err", proof.WithError(err))
	}
	if len(element.ElementID) > 0 && element.TargetType == consts.UISceneElementTypeSelect {
		ret[element.ElementID] = actualElement
	}

	return ret
}

// TransSceneOperatorToRaoSceneOperator 步骤详情
func TransSceneOperatorToRaoSceneOperator(
	scene *model.UIScene,
	operator *model.UISceneOperator,
	maoOperator *mao.SceneOperator,
	elementList []*model.Element,
) *rao.UISceneOperator {
	var d *mao.ActionDetail
	if err := bson.Unmarshal(maoOperator.ActionDetail, &d); err != nil {
		log.Logger.Errorf("maoOperator.ActionDetail bson unmarshal err %w", err)
	}

	var s *mao.Settings
	if err := bson.Unmarshal(maoOperator.Settings, &s); err != nil {
		log.Logger.Errorf("mao.Settings bson unmarshal err %w", err)
	}

	elementMemo := make(map[string]*model.Element)
	for _, e := range elementList {
		elementMemo[e.ElementID] = e
	}

	if operator.Action == consts.UISceneOptTypeToggleWindow {
		if d.Detail.ToggleWindow.Type == consts.UISceneOptTypeSwitchFrameByLocator {
			elementMode := d.Detail.ToggleWindow.SwitchFrameByLocator.Element
			elementMode = elementModeToNewestElement(elementMode, elementMemo, s.Settings)
		}
	}

	if operator.Action == consts.UISceneOptTypeMouseClicking {
		elementMode := d.Detail.MouseClicking.Element
		elementMode = elementModeToNewestElement(elementMode, elementMemo, s.Settings)
	}

	if operator.Action == consts.UISceneOptTypeMouseScrolling {
		elementMode := d.Detail.MouseScrolling.Element
		elementMode = elementModeToNewestElement(elementMode, elementMemo, s.Settings)
	}

	if operator.Action == consts.UISceneOptTypeMouseDragging {
		elementMode := d.Detail.MouseDragging.Element
		elementMode = elementModeToNewestElement(elementMode, elementMemo, s.Settings)
		if d.Detail.MouseDragging.Type == consts.MouseDraggingTypeDragElement {
			targetElementMode := d.Detail.MouseDragging.TarGetElement
			targetElementMode = elementModeToNewestElement(targetElementMode, elementMemo, s.Settings)
		}
	}

	if operator.Action == consts.UISceneOptTypeInputOperations {
		elementMode := d.Detail.InputOperations.Element
		elementMode = elementModeToNewestElement(elementMode, elementMemo, s.Settings)
	}

	if operator.Action == consts.UISceneOptTypeWaitEvents {
		elementMode := d.Detail.WaitEvents.Element
		elementMode = elementModeToNewestElement(elementMode, elementMemo, s.Settings)
	}

	if operator.Action == consts.UISceneOptTypeIfCondition {
		if d.Detail.IfCondition != nil {
			conditionOperators := d.Detail.IfCondition.ConditionOperators
			if conditionOperators != nil {
				for _, conditionOperator := range conditionOperators {
					if conditionOperator.Type == consts.IfConditionTypeConditionOperatorAssert &&
						conditionOperator.Assert != nil {
						assertElementModeToNewestElement(conditionOperator.Assert, elementMemo, s.Settings)
					}
				}
			}
		}
	}

	if operator.Action == consts.UISceneOptTypeCodeOperation {
		elementMode := d.Detail.CodeOperation.Element
		elementMode = elementModeToNewestElement(elementMode, elementMemo, s.Settings)
	}

	if operator.Action == consts.UISceneOptTypeAssert {
		assertElementModeToNewestElement(d.Detail.Assert, elementMemo, s.Settings)
	}

	if operator.Action == consts.UISceneOptTypeDataWithdraw {
		withdrawElementModeToNewestElement(d.Detail.DataWithdraw, elementMemo, s.Settings)
	}

	var a *mao.Asserts
	if err := bson.Unmarshal(maoOperator.Asserts, &a); err != nil {
		log.Logger.Errorf("mao.Asserts bson unmarshal err %w", err)
	}

	for _, assert := range a.Asserts {
		assertElementModeToNewestElement(assert, elementMemo, s.Settings)
	}

	var w *mao.DataWithdraws
	if err := bson.Unmarshal(maoOperator.DataWithdraws, &w); err != nil {
		log.Logger.Errorf("mao.DataWithdraws bson unmarshal err %w", err)
	}

	for _, withdraw := range w.DataWithdraws {
		withdrawElementModeToNewestElement(withdraw, elementMemo, s.Settings)
	}

	uiSceneOperator := &rao.UISceneOperator{
		TeamID:        scene.TeamID,
		SceneID:       operator.SceneID,
		OperatorID:    operator.OperatorID,
		ParentID:      operator.ParentID,
		Name:          operator.Name,
		Sort:          operator.Sort,
		Status:        operator.Status,
		Type:          operator.Type,
		Action:        operator.Action,
		ActionDetail:  d.Detail,
		Settings:      s.Settings,
		Asserts:       a.Asserts,
		DataWithdraws: w.DataWithdraws,
	}

	return uiSceneOperator
}

// assertElementModeToNewestElement 获取最新的元素
func assertElementModeToNewestElement(assert *rao.AutomationAssert, elementMemo map[string]*model.Element, setting *rao.AutomationSettings) {
	if assert.Type == consts.UISceneOptElementExists {
		elementMode := assert.Element
		elementMode = elementModeToNewestElement(elementMode, elementMemo, setting)
	}
	if assert.Type == consts.UISceneOptElementNotExists {
		elementMode := assert.Element
		elementMode = elementModeToNewestElement(elementMode, elementMemo, setting)
	}
	if assert.Type == consts.UISceneOptElementDisplayed {
		elementMode := assert.Element
		elementMode = elementModeToNewestElement(elementMode, elementMemo, setting)
	}
	if assert.Type == consts.UISceneOptElementNotDisplayed {
		elementMode := assert.Element
		elementMode = elementModeToNewestElement(elementMode, elementMemo, setting)
	}
	if assert.Type == consts.UISceneOptElementAttributeAssertion {
		elementMode := assert.Element
		elementMode = elementModeToNewestElement(elementMode, elementMemo, setting)
	}
}

func withdrawElementModeToNewestElement(a *rao.DataWithdraw, elementMemo map[string]*model.Element, setting *rao.AutomationSettings) {
	if a.WithdrawType == consts.UISceneWithdrawTypeElement {
		if a.ElementMethod != nil {
			elementMode := a.ElementMethod.Element
			elementMode = elementModeToNewestElement(elementMode, elementMemo, setting)
		}
	}
}

// elementModeToNewestElement 获取最新的元素
func elementModeToNewestElement(
	element *rao.Element,
	elementMemo map[string]*model.Element,
	setting *rao.AutomationSettings,
) *rao.Element {
	if element == nil {
		return &rao.Element{}
	}

	if setting != nil && setting.ElementSyncMode == consts.UISceneOptSyncModeAuto &&
		element.TargetType == consts.UISceneElementTypeSelect {
		if e, ok := elementMemo[element.ElementID]; ok {
			raoElement := TransModelElementToRaoElement(e)
			raoElement.TargetType = element.TargetType
			raoElement.CustomLocators = element.CustomLocators
			if err := copier.Copy(element, raoElement); err != nil {
				log.Logger.Error("elementModeToNewestElement Copy err", proof.WithError(err))
			}
		}
	}

	return element
}

// TransSceneOperatorsToRaoSendOperators 组装发送数据和操作记录
func TransSceneOperatorsToRaoSendOperators(
	operators []*model.UISceneOperator,
	runID string,
	scene *model.UIScene,
	sceneOperators []*mao.SceneOperator,
	elementList []*model.Element,
) ([]*ui.Operator, []*mao.UISendSceneOperator) {
	var (
		err                  error
		uiOperators          = make([]*ui.Operator, 0, len(operators))
		uiSendSceneOperators = make([]*mao.UISendSceneOperator, 0, len(operators))
	)

	operatorMap := make(map[string]*mao.SceneOperator, len(sceneOperators))
	for _, o := range sceneOperators {
		operatorMap[o.OperatorID] = o
	}

	elementMemo := make(map[string]*model.Element)
	for _, e := range elementList {
		elementMemo[e.ElementID] = e
	}

	for _, o := range operators {
		// 禁用的不发送
		if o.Status == consts.UISceneOptStatusDisable {
			continue
		}
		if d, ok := operatorMap[o.OperatorID]; ok {
			var assertTotalNum int
			uiEngineAssertions := make([]*rao.UIEngineAssertion, 0, len(d.Asserts))
			uiEngineDataWithdraws := make([]*rao.UIEngineDataWithdraw, 0, len(d.DataWithdraws))
			uiEngineResultDataMsgs := make([]*rao.UIEngineResultDataMsg, 0)

			settings := mao.Settings{}
			if err := bson.Unmarshal(d.Settings, &settings); err != nil {
				log.Logger.Error("api.Settings bson Unmarshal err", proof.WithError(err))
			}
			uiSettings := &ui.Settings{}
			// 将 o.Settings 的字段复制到 settings 中
			if err = copier.Copy(uiSettings, settings.Settings); err != nil {
				log.Logger.Error("ui.Settings Copy err", proof.WithError(err))
			}

			assert := mao.Asserts{}
			if err := bson.Unmarshal(d.Asserts, &assert); err != nil {
				log.Logger.Error("api.Asserts bson Unmarshal err", proof.WithError(err))
			}
			uiAsserts := make([]*ui.Assert, 0, len(d.Asserts))
			for _, a := range assert.Asserts {
				if a.Status != consts.UISceneOptStatusDisable {
					uiAssert := &ui.Assert{
						Type: a.Type,
					}
					uiEngineResultName := assertElementModeToUiAssert(a, elementMemo, uiAssert, settings.Settings)

					uiAsserts = append(uiAsserts, uiAssert)
					assertTotalNum++

					uiEngineAssertion := &rao.UIEngineAssertion{
						Name: uiEngineResultName,
					}
					uiEngineAssertions = append(uiEngineAssertions, uiEngineAssertion)
				}
			}

			dataWithdraws := mao.DataWithdraws{}
			if err := bson.Unmarshal(d.DataWithdraws, &dataWithdraws); err != nil {
				log.Logger.Error("api.DataWithdraws bson Unmarshal err", proof.WithError(err))
			}
			uiDataWithdraws := make([]*ui.DataWithdraw, 0, len(d.DataWithdraws))
			for _, w := range dataWithdraws.DataWithdraws {
				if w.Status != consts.UISceneOptStatusDisable {
					dataWithdraw := &ui.DataWithdraw{
						ElementMethod: &ui.WithdrawElementMethod{
							Element: &ui.Element{Locators: nil},
						},
					}
					if err = copier.Copy(dataWithdraw, w); err != nil {
						log.Logger.Error("ui.DataWithdraw Copy err", proof.WithError(err))
					}
					if dataWithdraw.WithdrawType == consts.UISceneWithdrawTypeElement {
						elementMode := w.ElementMethod.Element
						uiLocators := elementModeToUiLocators(elementMode, elementMemo, settings.Settings)

						if dataWithdraw.ElementMethod != nil {
							dataWithdraw.ElementMethod.Element.Locators = uiLocators
						}
					}

					uiDataWithdraws = append(uiDataWithdraws, dataWithdraw)

					uiEngineDataWithdraw := &rao.UIEngineDataWithdraw{}
					uiEngineDataWithdraws = append(uiEngineDataWithdraws, uiEngineDataWithdraw)
				}
			}

			operator := &ui.Operator{
				OperatorId:    o.OperatorID,
				Sort:          o.Sort,
				OperatorType:  o.Type,
				ParentId:      o.ParentID,
				Action:        o.Action,
				Settings:      uiSettings,
				Asserts:       uiAsserts,
				DataWithdraws: uiDataWithdraws,
			}

			actionDetail := mao.ActionDetail{}
			if err := bson.Unmarshal(d.ActionDetail, &actionDetail); err != nil {
				log.Logger.Info("api.Settings bson Unmarshal err", proof.WithError(err))
			}

			if o.Action == consts.UISceneOptTypeOpenPage {
				openPage := &ui.OpenPage{}
				if err = copier.Copy(openPage, actionDetail.Detail.OpenPage); err != nil {
					log.Logger.Error("ui.OpenPage Copy err", proof.WithError(err))
				}
				operator.OpenPage = openPage
			}

			if o.Action == consts.UISceneOptTypeClosePage {
				closePage := &ui.ClosePage{}
				if err = copier.Copy(closePage, actionDetail.Detail.ClosePage); err != nil {
					log.Logger.Error("UISceneOptTypeClosePage Copy err", proof.WithError(err))
				}
				operator.ClosePage = closePage
			}

			if o.Action == consts.UISceneOptTypeToggleWindow {
				toggleWindowType := actionDetail.Detail.ToggleWindow.Type
				uiToggleWindow := &ui.ToggleWindow{
					Type: toggleWindowType,
				}
				if toggleWindowType == consts.UISceneOptTypeSwitchPage {
					switchPage := &ui.SwitchPage{}
					if err = copier.Copy(switchPage, actionDetail.Detail.ToggleWindow.SwitchPage); err != nil {
						log.Logger.Error("UISceneOptTypeSwitchPage Copy err", proof.WithError(err))
					}
					uiToggleWindow.SwitchPage = switchPage
				}

				if toggleWindowType == consts.UISceneOptTypeExitFrame {
					exitFrame := &ui.ExitFrame{}
					if err = copier.Copy(exitFrame, actionDetail.Detail.ToggleWindow.ExitFrame); err != nil {
						log.Logger.Error("UISceneOptTypeExitFrame Copy err", proof.WithError(err))
					}
					uiToggleWindow.ExitFrame = exitFrame
				}

				if toggleWindowType == consts.UISceneOptTypeSwitchFrameByIndex {
					switchFrameByIndex := &ui.SwitchFrameByIndex{}
					if err = copier.Copy(switchFrameByIndex, actionDetail.Detail.ToggleWindow.SwitchFrameByIndex); err != nil {
						log.Logger.Error("UISceneOptTypeSwitchFrameByIndex Copy err", proof.WithError(err))
					}
					uiToggleWindow.SwitchFrameByIndex = switchFrameByIndex
				}

				if toggleWindowType == consts.UISceneOptTypeSwitchToParentFrame {
					switchToParentFrame := &ui.SwitchToParentFrame{}
					if err = copier.Copy(switchToParentFrame, actionDetail.Detail.ToggleWindow.SwitchToParentFrame); err != nil {
						log.Logger.Error("UISceneOptTypeSwitchToParentFrame Copy err", proof.WithError(err))
					}
					uiToggleWindow.SwitchToParentFrame = switchToParentFrame
				}

				if toggleWindowType == consts.UISceneOptTypeSwitchFrameByLocator {
					switchFrameByLocator := &ui.SwitchFrameByLocator{
						Element: &ui.Element{Locators: nil},
					}
					if err = copier.Copy(switchFrameByLocator, actionDetail.Detail.ToggleWindow.SwitchFrameByLocator); err != nil {
						log.Logger.Error("UISceneOptTypeSwitchFrameByLocator Copy err", proof.WithError(err))
					}
					elementMode := actionDetail.Detail.ToggleWindow.SwitchFrameByLocator.Element
					uiLocators := elementModeToUiLocators(elementMode, elementMemo, settings.Settings)

					if switchFrameByLocator.Element != nil {
						switchFrameByLocator.Element.Locators = uiLocators
					}

					uiToggleWindow.SwitchFrameByLocator = switchFrameByLocator
				}
				operator.ToggleWindow = uiToggleWindow
			}

			if o.Action == consts.UISceneOptTypeSetWindowSize {
				setWindowSize := &ui.SetWindowSize{}
				if err = copier.Copy(setWindowSize, actionDetail.Detail.SetWindowSize); err != nil {
					log.Logger.Error("UISceneOptTypeSetWindowSize Copy err", proof.WithError(err))
				}

				operator.SetWindowSize = setWindowSize
			}

			if o.Action == consts.UISceneOptTypeMouseClicking {
				mouseClick := &ui.MouseClick{
					Type:          "",
					Element:       &ui.Element{Locators: nil},
					ClickPosition: nil,
				}
				if err = copier.Copy(mouseClick, actionDetail.Detail.MouseClicking); err != nil {
					log.Logger.Error("UISceneOptTypeMouseClicking Copy err", proof.WithError(err))
				}
				clickPosition := &ui.ClickPosition{}
				if err = copier.Copy(clickPosition, actionDetail.Detail.MouseClicking.ClickPosition); err != nil {
					log.Logger.Error("UISceneOptTypeMouseClicking Copy err", proof.WithError(err))
				}
				elementMode := actionDetail.Detail.MouseClicking.Element
				uiLocators := elementModeToUiLocators(elementMode, elementMemo, settings.Settings)

				if mouseClick.Element != nil {
					mouseClick.Element.Locators = uiLocators
				}

				mouseClick.ClickPosition = clickPosition

				operator.MouseClicking = mouseClick
			}

			if o.Action == consts.UISceneOptTypeMouseScrolling {
				mouseScroll := &ui.MouseScroll{
					Type:                 "",
					Element:              &ui.Element{Locators: nil},
					ScrollDistance:       0,
					SingleScrollDistance: 0,
				}
				if err = copier.Copy(mouseScroll, actionDetail.Detail.MouseScrolling); err != nil {
					log.Logger.Error("UISceneOptTypeMouseScrolling Copy err", proof.WithError(err))
				}
				elementMode := actionDetail.Detail.MouseScrolling.Element
				uiLocators := elementModeToUiLocators(elementMode, elementMemo, settings.Settings)

				if mouseScroll.Element != nil {
					mouseScroll.Element.Locators = uiLocators
				}

				operator.MouseScrolling = mouseScroll
			}

			if o.Action == consts.UISceneOptTypeMouseMovement {
				mouseMove := &ui.MouseMove{
					Type: "",
				}
				if err = copier.Copy(mouseMove, actionDetail.Detail.MouseMovement); err != nil {
					log.Logger.Error("UISceneOptTypeMouseMovement Copy err", proof.WithError(err))
				}

				operator.MouseMovement = mouseMove
			}

			if o.Action == consts.UISceneOptTypeMouseDragging {
				mouseDrag := &ui.MouseDrag{
					Type:                "",
					Element:             &ui.Element{Locators: nil},
					TargetElement:       &ui.Element{Locators: nil},
					EndPointCoordinates: nil,
				}
				if err = copier.Copy(mouseDrag, actionDetail.Detail.MouseDragging); err != nil {
					log.Logger.Error("UISceneOptTypeMouseDragging Copy err", proof.WithError(err))
				}
				elementMode := actionDetail.Detail.MouseDragging.Element
				uiLocators := elementModeToUiLocators(elementMode, elementMemo, settings.Settings)

				if mouseDrag.Element != nil {
					mouseDrag.Element.Locators = uiLocators
				}

				if actionDetail.Detail.MouseDragging.Type == consts.MouseDraggingTypeDragElement {
					targetElementMode := actionDetail.Detail.MouseDragging.TarGetElement
					targetUILocators := elementModeToUiLocators(targetElementMode, elementMemo, settings.Settings)

					if mouseDrag.TargetElement != nil {
						mouseDrag.TargetElement.Locators = targetUILocators
					}
				}

				dragPointCoordinates := &ui.DragPointCoordinates{}
				if err = copier.Copy(dragPointCoordinates, actionDetail.Detail.MouseDragging.EndPointCoordinates); err != nil {
					log.Logger.Error("UISceneOptTypeMouseDragging EndPointCoordinates Copy err", proof.WithError(err))
				}
				mouseDrag.EndPointCoordinates = dragPointCoordinates

				operator.MouseDragging = mouseDrag
			}

			if o.Action == consts.UISceneOptTypeInputOperations {
				// 复制 UI  所需数据
				inputOperations := &ui.InputOperations{
					Type:         "",
					Element:      &ui.Element{Locators: nil},
					InputContent: "",
				}
				if err = copier.Copy(inputOperations, actionDetail.Detail.InputOperations); err != nil {
					log.Logger.Error("ui.InputOperations Copy err", proof.WithError(err))
				}

				elementMode := actionDetail.Detail.InputOperations.Element
				uiLocators := elementModeToUiLocators(elementMode, elementMemo, settings.Settings)

				if inputOperations.Element != nil {
					inputOperations.Element.Locators = uiLocators
				}

				operator.InputOperations = inputOperations
			}

			if o.Action == consts.UISceneOptTypeWaitEvents {
				// 复制 UI  所需数据
				waitEvent := &ui.WaitEvent{
					Element: &ui.Element{Locators: nil},
				}
				if err = copier.Copy(waitEvent, actionDetail.Detail.WaitEvents); err != nil {
					log.Logger.Error("ui.WaitEvents Copy err", proof.WithError(err))
				}
				waitEvent.WaitTime = strconv.FormatInt(int64(actionDetail.Detail.WaitEvents.WaitTime), 10)

				elementMode := actionDetail.Detail.WaitEvents.Element
				uiLocators := elementModeToUiLocators(elementMode, elementMemo, settings.Settings)

				if waitEvent.Element != nil {
					waitEvent.Element.Locators = uiLocators
				}

				targetTexts := make([]string, 0)
				for _, text := range actionDetail.Detail.WaitEvents.TargetTexts {
					targetTexts = append(targetTexts, text)
				}

				waitEvent.TargetTexts = targetTexts

				operator.WaitEvents = waitEvent
			}

			if o.Action == consts.UISceneOptTypeIfCondition {
				// 组织 grpc 所用到的元素信息
				uiIf := &ui.IfCondition{
					ConditionRelate:    "",
					ConditionOperators: nil,
				}
				if err = copier.Copy(uiIf, actionDetail.Detail.IfCondition); err != nil {
					log.Logger.Error("ui.If Copy err", proof.WithError(err))
				}

				// ConditionOperators 组织数据
				uiConditionOperators := make([]*ui.ConditionOperator, 0)
				if actionDetail.Detail.IfCondition != nil {
					conditionOperators := actionDetail.Detail.IfCondition.ConditionOperators
					if conditionOperators != nil {
						for _, conditionOperator := range conditionOperators {
							if conditionOperator.Status != consts.UISceneOptStatusDisable {
								uiConditionOperator := &ui.ConditionOperator{
									Type:       conditionOperator.Type,
									AssertInfo: nil,
								}
								if conditionOperator.Type == consts.IfConditionTypeConditionOperatorAssert &&
									conditionOperator.Assert != nil {
									uiAssert := &ui.Assert{
										Type: conditionOperator.Assert.Type,
									}
									assertElementModeToUiAssert(conditionOperator.Assert, elementMemo, uiAssert, settings.Settings)
									uiConditionOperator.AssertInfo = uiAssert
								}
								uiConditionOperators = append(uiConditionOperators, uiConditionOperator)
							}
						}
					}
				}
				uiIf.ConditionOperators = uiConditionOperators

				operator.IfCondition = uiIf
			}

			if o.Action == consts.UISceneOptTypeCodeOperation {
				codeOperation := &ui.CodeOperation{
					Element: &ui.Element{Locators: nil},
				}
				if err = copier.Copy(codeOperation, actionDetail.Detail.CodeOperation); err != nil {
					log.Logger.Error("UISceneOptTypeCodeOperation Copy err", proof.WithError(err))
				}
				elementMode := actionDetail.Detail.CodeOperation.Element
				uiLocators := elementModeToUiLocators(elementMode, elementMemo, settings.Settings)

				if codeOperation.Element != nil {
					codeOperation.Element.Locators = uiLocators
				}

				operator.CodeOperation = codeOperation
			}

			if o.Action == consts.UISceneOptTypeAssert {
				uiAssert := &ui.Assert{
					Type: actionDetail.Detail.Assert.Type,
				}
				assertElementModeToUiAssert(actionDetail.Detail.Assert, elementMemo, uiAssert, settings.Settings)
				operator.AssertInfo = uiAssert
			}

			if o.Action == consts.UISceneOptTypeForLoop {
				// 复制 UI  所需数据
				forLoop := &ui.ForLoop{}
				if err = copier.Copy(forLoop, actionDetail.Detail.ForLoop); err != nil {
					log.Logger.Error("ui.ForLoop Copy err", proof.WithError(err))
				}
				uiBaseFiles := make([]*ui.BaseFile, 0)
				if len(actionDetail.Detail.ForLoop.Files) > 0 {
					for _, file := range actionDetail.Detail.ForLoop.Files {
						if file.Status != consts.UISceneOptStatusDisable {
							uiBaseFile := &ui.BaseFile{}
							if err = copier.Copy(uiBaseFile, file); err != nil {
								log.Logger.Error("ui.BaseFile Copy err", proof.WithError(err))
							}
							uiBaseFiles = append(uiBaseFiles, uiBaseFile)
						}
					}
				}

				forLoop.Files = uiBaseFiles
				operator.ForLoop = forLoop
			}

			if o.Action == consts.UISceneOptTypeWhileLoop {
				// 复制 UI  所需数据
				whileLoop := &ui.WhileLoop{
					ConditionRelate:    "",
					ConditionOperators: nil,
					MaxCount:           0,
				}
				if err = copier.Copy(whileLoop, actionDetail.Detail.WhileLoop); err != nil {
					log.Logger.Error("ui.WhileLoop Copy err", proof.WithError(err))
				}
				// ConditionOperators 组织数据
				uiConditionOperators := make([]*ui.ConditionOperator, 0)
				if actionDetail.Detail.WhileLoop != nil {
					conditionOperators := actionDetail.Detail.WhileLoop.ConditionOperators
					if conditionOperators != nil {
						for _, conditionOperator := range conditionOperators {
							if conditionOperator.Status != consts.UISceneOptStatusDisable {
								uiConditionOperator := &ui.ConditionOperator{
									Type:       conditionOperator.Type,
									AssertInfo: nil,
								}
								if conditionOperator.Type == consts.IfConditionTypeConditionOperatorAssert &&
									conditionOperator.Assert != nil {
									uiAssert := &ui.Assert{
										Type: conditionOperator.Assert.Type,
									}
									assertElementModeToUiAssert(conditionOperator.Assert, elementMemo, uiAssert, settings.Settings)
									uiConditionOperator.AssertInfo = uiAssert
								}
								uiConditionOperators = append(uiConditionOperators, uiConditionOperator)
							}
						}
					}
				}
				whileLoop.ConditionOperators = uiConditionOperators

				operator.WhileLoop = whileLoop
			}

			if o.Action == consts.UISceneOptTypeDataWithdraw {
				dataWithdraw := &ui.DataWithdraw{
					ElementMethod: &ui.WithdrawElementMethod{
						Element: &ui.Element{Locators: nil},
					},
				}
				if err = copier.Copy(dataWithdraw, actionDetail.Detail.DataWithdraw); err != nil {
					log.Logger.Error("ui.DataWithdraw Copy err", proof.WithError(err))
				}
				if dataWithdraw.WithdrawType == consts.UISceneWithdrawTypeElement {
					elementMode := actionDetail.Detail.DataWithdraw.ElementMethod.Element
					uiLocators := elementModeToUiLocators(elementMode, elementMemo, settings.Settings)

					if dataWithdraw.ElementMethod.Element != nil {
						dataWithdraw.ElementMethod.Element.Locators = uiLocators
					}
				}

				operator.DataWithdraw = dataWithdraw
			}

			uiOperators = append(uiOperators, operator)

			uiOperatorDetail, err := bson.Marshal(mao.UIOperatorDetail{
				Operator: d,
			})
			if err != nil {
				log.Logger.Info("TransSceneOperatorsToRaoSendOperators.uiOperatorDetail bson marshal err", proof.WithError(err))
			}

			assertResults, err := bson.Marshal(mao.AssertResults{
				Asserts: uiEngineAssertions,
			})
			if err != nil {
				log.Logger.Info("TransSceneOperatorsToRaoSendOperators.uiOperatorDetail bson marshal err", proof.WithError(err))
			}

			withdrawResults, err := bson.Marshal(mao.WithdrawResults{
				Withdraws: uiEngineDataWithdraws,
			})
			if err != nil {
				log.Logger.Info("TransSceneOperatorsToRaoSendOperators.uiOperatorDetail bson marshal err", proof.WithError(err))
			}

			multiResults, err := bson.Marshal(mao.MultiResults{
				MultiResults: uiEngineResultDataMsgs,
			})
			if err != nil {
				log.Logger.Info("TransSceneOperatorsToRaoSendOperators.uiOperatorDetail bson marshal err", proof.WithError(err))
			}

			uiSendSceneOperator := &mao.UISendSceneOperator{
				ReportId:        runID,
				TeamID:          scene.TeamID,
				SceneID:         scene.SceneID,
				SceneName:       scene.Name,
				OperatorID:      o.OperatorID,
				ParentID:        o.ParentID,
				Name:            o.Name,
				Sort:            o.Sort,
				Type:            o.Type,
				Action:          o.Action,
				RunStatus:       1,
				ExecTime:        0,
				RunEndTimes:     0,
				Status:          "",
				Msg:             "",
				Screenshot:      "",
				End:             false,
				IsMulti:         false,
				AssertTotalNum:  assertTotalNum,
				CreatedAt:       time.Now(),
				Detail:          uiOperatorDetail,
				AssertResults:   assertResults,
				MultiResults:    multiResults,
				WithdrawResults: withdrawResults,
			}

			uiSendSceneOperators = append(uiSendSceneOperators, uiSendSceneOperator)
		}
	}

	return uiOperators, uiSendSceneOperators
}

// assertElementModeToNewestElement 获取最新的元素
func assertElementModeToUiAssert(
	a *rao.AutomationAssert,
	elementMemo map[string]*model.Element,
	uiAssert *ui.Assert,
	setting *rao.AutomationSettings) string {
	var name string
	if a.Type == consts.UISceneOptElementExists {
		elementMode := a.Element
		uiLocators := elementModeToUiLocators(elementMode, elementMemo, setting)

		uiAssert.ElementExists = &ui.ElementAssertion{
			Element: &ui.Element{Locators: uiLocators},
		}

		name = fmt.Sprintf("断言 %s 元素存在", elementMode.Name)
	}

	if a.Type == consts.UISceneOptElementNotExists {
		// 查询最新的元素数据
		elementMode := a.Element
		uiLocators := elementModeToUiLocators(elementMode, elementMemo, setting)

		uiAssert.ElementNotExists = &ui.ElementAssertion{
			Element: &ui.Element{Locators: uiLocators},
		}

		name = fmt.Sprintf("断言 %s 元素不存在", elementMode.Name)
	}

	if a.Type == consts.UISceneOptElementDisplayed {
		// 查询最新的元素数据
		elementMode := a.Element
		uiLocators := elementModeToUiLocators(elementMode, elementMemo, setting)

		uiAssert.ElementDisplayed = &ui.ElementAssertion{
			Element: &ui.Element{Locators: uiLocators},
		}

		name = fmt.Sprintf("断言 %s 元素显示", elementMode.Name)
	}

	if a.Type == consts.UISceneOptElementNotDisplayed {
		// 查询最新的元素数据
		elementMode := a.Element
		uiLocators := elementModeToUiLocators(elementMode, elementMemo, setting)

		uiAssert.ElementNotDisplayed = &ui.ElementAssertion{
			Element: &ui.Element{Locators: uiLocators},
		}

		name = fmt.Sprintf("断言 %s 元素不显示", elementMode.Name)
	}

	if a.Type == consts.UISceneOptTextExists {
		targetTexts := make([]string, 0)
		for _, text := range a.TextExists.TargetTexts {
			targetTexts = append(targetTexts, text)
		}

		uiAssert.TextExists = &ui.TextExists{TargetTexts: targetTexts}

		name = fmt.Sprintf("断言 %s 存在", strings.Join(targetTexts, ","))
	}

	if a.Type == consts.UISceneOptTextNotExists {
		targetTexts := make([]string, 0)
		for _, text := range a.TextNotExists.TargetTexts {
			targetTexts = append(targetTexts, text)
		}

		uiAssert.TextNotExists = &ui.TextNotExists{TargetTexts: targetTexts}

		name = fmt.Sprintf("断言 %s 不存在", strings.Join(targetTexts, ","))
	}

	if a.Type == consts.UISceneOptVariableAssertion {
		variableAssertion := &ui.VariableAssertion{}
		if err := copier.Copy(variableAssertion, a.VariableAssertion); err != nil {
			log.Logger.Error("UISceneOptVariableAssertion Copy err", proof.WithError(err))
		}

		uiAssert.VariableAssertion = variableAssertion

		relationOptions := ""
		if r, ok := consts.RelationOptions[variableAssertion.RelationOptions]; ok {
			relationOptions = r
		}
		name = fmt.Sprintf("变量断言:断言 %s %s %s", variableAssertion.ActualValue, relationOptions, variableAssertion.ExpectedValue)
	}

	if a.Type == consts.UISceneOptExpressionAssertion {
		expressionAssertion := &ui.ExpressionAssertion{}
		if err := copier.Copy(expressionAssertion, a.ExpressionAssertion); err != nil {
			log.Logger.Error("UISceneOptExpressionAssertion Copy err", proof.WithError(err))
		}

		uiAssert.ExpressionAssertion = expressionAssertion

		name = fmt.Sprintf("表达式断言： %s", expressionAssertion.ExpectedValue)

	}

	if a.Type == consts.UISceneOptElementAttributeAssertion {
		// 复制 UI  所需数据
		elementAttributeAssert := &ui.ElementAttributeAssert{
			RelationOptions: "",
			Element:         &ui.Element{Locators: nil},
			ConditionType:   "",
		}
		if err := copier.Copy(elementAttributeAssert, a.ElementAttributeAssert); err != nil {
			log.Logger.Error("UISceneOptElementAttributeAssertion Copy err", proof.WithError(err))
		}

		elementMode := a.Element
		uiLocators := elementModeToUiLocators(elementMode, elementMemo, setting)

		if elementAttributeAssert.Element != nil {
			elementAttributeAssert.Element.Locators = uiLocators
		}

		uiAssert.ElementAttributeAssertion = elementAttributeAssert
		//      3. 断言元素属性：断言 xx元素（目标元素）的标签名称（元素属性） 等于（断言关系） 111（期望值）
		relationOptions := ""
		if r, ok := consts.RelationOptions[elementAttributeAssert.RelationOptions]; ok {
			relationOptions = r
		}
		name = fmt.Sprintf("断言元素属性：断言  %s 的 %s %s %s",
			elementMode.Name, elementAttributeAssert.ConditionType, relationOptions, elementAttributeAssert.ExpectedValue)
	}

	if a.Type == consts.UISceneOptPageAttributeAssertion {
		// 复制 UI  所需数据
		pageAttributeAssert := &ui.PageAttributeAssert{}
		if err := copier.Copy(pageAttributeAssert, a.PageAttributeAssert); err != nil {
			log.Logger.Error("UISceneOptPageAttributeAssertion Copy err", proof.WithError(err))
		}

		uiAssert.PageAttributeAssertion = pageAttributeAssert

		//      6. 断言页面属性：断言 页面url（断言属性） 包含（断言关系） 111（期望值）
		relationOptions := ""
		if r, ok := consts.RelationOptions[pageAttributeAssert.RelationOptions]; ok {
			relationOptions = r
		}
		name = fmt.Sprintf("断言页面属性：断言  %s %s %s",
			pageAttributeAssert.AssertAttribute, relationOptions, pageAttributeAssert.ExpectedValue)
	}

	return name
}

// elementModeToUiLocators
func elementModeToUiLocators(
	element *rao.Element,
	elementMemo map[string]*model.Element,
	setting *rao.AutomationSettings) []*ui.Locator {
	// 组织 grpc 所用到的元素信息
	uiLocators := make([]*ui.Locator, 0)
	if element == nil {
		return uiLocators
	}

	// 查询最新的元素数据
	if setting != nil && setting.ElementSyncMode == consts.UISceneOptSyncModeAuto &&
		element.TargetType == consts.UISceneElementTypeSelect {
		elementID := element.ElementID
		if e, ok := elementMemo[elementID]; ok {
			raoElement := TransModelElementToRaoElement(e)
			element.Name = raoElement.Name
			raoElement.TargetType = element.TargetType
			raoElement.CustomLocators = element.CustomLocators
			if err := copier.Copy(element, raoElement); err != nil {
				log.Logger.Error("elementModeToUiLocators Copy err", proof.WithError(err))
			}
		}
	}

	var targetLocators = make([]*rao.Locator, 0)
	if element.TargetType == consts.UISceneElementTypeCustom {
		targetLocators = append(targetLocators, element.CustomLocators...)
	} else {
		targetLocators = append(targetLocators, element.Locators...)
	}

	for _, l := range targetLocators {
		uiAttribute := &ui.Locator{}
		if err := copier.Copy(uiAttribute, l); err != nil {
			log.Logger.Error("elementModeToUiLocators ui.uiAttribute Copy err", proof.WithError(err))
		}
		uiLocators = append(uiLocators, uiAttribute)
	}

	return uiLocators
}

func TransSceneTrashListToRaoTrash(trashList []*model.UISceneTrash, scenes []*model.UIScene, users []*model.User) []*rao.UISceneTrash {
	ret := make([]*rao.UISceneTrash, 0, len(trashList))

	scenesMemo := make(map[string]*model.UIScene)
	for _, s := range scenes {
		scenesMemo[s.SceneID] = s
	}

	usersMemo := make(map[string]*model.User)
	for _, u := range users {
		usersMemo[u.UserID] = u
	}
	for _, t := range trashList {
		trash := &rao.UISceneTrash{
			SceneID:        t.SceneID,
			TeamID:         t.TeamID,
			CreatedTimeSec: t.CreatedAt.Unix(),
		}
		if s, ok := scenesMemo[t.SceneID]; ok {
			trash.Name = s.Name
			trash.CreatedUserID = s.CreatedUserID
			trash.SceneType = s.SceneType
		}
		if u, ok := usersMemo[t.CreatedUserID]; ok {
			trash.CreatedUserID = u.UserID
			trash.CreatedUserName = u.Nickname
		}

		ret = append(ret, trash)
	}

	return ret
}
