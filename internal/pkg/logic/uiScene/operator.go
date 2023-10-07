package uiScene

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/log"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/uuid"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/mao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/query"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/errmsg"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/packer"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/public"
	"github.com/gin-gonic/gin"
	"github.com/go-omnibus/proof"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
	"reflect"
	"strconv"
)

// IsPointerStructNil 判断指针结构体中的值是否为nil
func IsPointerStructNil(ptr interface{}) bool {
	if ptr == nil {
		return true
	}

	v := reflect.ValueOf(ptr)
	if v.Kind() != reflect.Ptr {
		return false
	}

	if v.IsNil() {
		return true
	}

	val := v.Elem()
	if val.Kind() != reflect.Struct {
		return false
	}

	zero := reflect.Zero(val.Type()).Interface()

	return reflect.DeepEqual(val.Interface(), zero)
}

func CheckSaveOperatorAndGetNameReq(ctx *gin.Context, req *rao.UISceneSaveOperatorReq) (string, error) {
	var (
		name string
	)
	if req.Action == consts.UISceneOptTypeOpenPage {
		if IsPointerStructNil(req.ActionDetail.OpenPage) {
			log.Logger.Error("UISceneOptTypeOpenPage CheckStructIsEmpty StructIsEmpty")
			return "", errmsg.ErrUISceneOpenPageEmpty
		}

		name = fmt.Sprintf("%s", req.ActionDetail.OpenPage.Url)
	}

	if req.Action == consts.UISceneOptTypeClosePage {
		if IsPointerStructNil(req.ActionDetail.ClosePage) {
			log.Logger.Error("UISceneOptTypeClosePage CheckStructIsEmpty StructIsEmpty")
			return "", errmsg.ErrUISceneClosePageEmpty
		}

		if t, ok := consts.ToggleWindowType[req.ActionDetail.ClosePage.WindowAction]; ok {
			name = fmt.Sprintf("%s%s", t, req.ActionDetail.ClosePage.InputContent)
		}
	}

	if req.Action == consts.UISceneOptTypeToggleWindow {
		toggleWindowType := req.ActionDetail.ToggleWindow.Type
		if toggleWindowType == consts.UISceneOptTypeSwitchPage {
			if IsPointerStructNil(req.ActionDetail.ToggleWindow.SwitchPage) {
				log.Logger.Error("UISceneOptTypeSwitchPage CheckStructIsEmpty StructIsEmpty")
				return "", errmsg.ErrUISceneSwitchPageEmpty
			}

			if t, ok := consts.ToggleWindowType[req.ActionDetail.ToggleWindow.SwitchPage.WindowAction]; ok {
				name = fmt.Sprintf("切换到%s", t)
			}
		}

		if toggleWindowType == consts.UISceneOptTypeExitFrame {
			name = "退出当前frame"
		}

		if toggleWindowType == consts.UISceneOptTypeSwitchFrameByIndex {
			if IsPointerStructNil(req.ActionDetail.ToggleWindow.SwitchFrameByIndex) {
				log.Logger.Error("UISceneOptTypeSwitchFrameByIndex CheckStructIsEmpty StructIsEmpty")
				return "", errmsg.ErrUISceneSwitchFrameByIndexEmpty
			}
			name = fmt.Sprintf("切换到索引号为%s的frame", strconv.Itoa(req.ActionDetail.ToggleWindow.SwitchFrameByIndex.FrameIndex))
		}

		if toggleWindowType == consts.UISceneOptTypeSwitchToParentFrame {
			if IsPointerStructNil(req.ActionDetail.ToggleWindow.SwitchToParentFrame) {
				log.Logger.Error("UISceneOptTypeSwitchFrameByIndex SwitchToParentFrame StructIsEmpty")
				return "", errmsg.ErrUISceneSwitchToParentFrame
			}
			name = "切换到上一层父级 frame"
		}

		if toggleWindowType == consts.UISceneOptTypeSwitchFrameByLocator {
			if IsPointerStructNil(req.ActionDetail.ToggleWindow.SwitchFrameByLocator) {
				log.Logger.Error("UISceneOptTypeSwitchFrameByIndex SwitchFrameByLocator StructIsEmpty")
				return "", errmsg.ErrUISceneSwitchFrameByLocatorEmpty
			}
			targetElement := req.ActionDetail.ToggleWindow.SwitchFrameByLocator.Element
			elementName := targetElement.Name
			if targetElement.TargetType == consts.UISceneElementTypeSelect {
				if len(targetElement.ElementID) == 0 {
					return "", errmsg.ErrElementLocatorNotFound
				}
			}
			if targetElement.TargetType == consts.UISceneElementTypeCustom {
				for _, locator := range targetElement.CustomLocators {
					if len(locator.Value) == 0 {
						return "", errmsg.ErrElementLocatorNotFound
					}
					method := "选择器"
					if elementMethod, ok := consts.ElementMethodType[locator.Method]; ok {
						method = elementMethod
					}

					elementName = fmt.Sprintf("{%s_%s_%s}", method, locator.Type, locator.Value)
				}
			}
			name = fmt.Sprintf("根据 %s 切换 frame", elementName)
		}
	}

	if req.Action == consts.UISceneOptTypeForward {
		name = "前进"
	}

	if req.Action == consts.UISceneOptTypeBack {
		name = "后退"
	}

	if req.Action == consts.UISceneOptTypeRefresh {
		name = "刷新"
	}

	if req.Action == consts.UISceneOptTypeSetWindowSize {
		if IsPointerStructNil(req.ActionDetail.SetWindowSize) {
			log.Logger.Error("UISceneOptTypeSwitchFrameByIndex SetWindowSize StructIsEmpty")
			return "", errmsg.ErrUISceneSetWindowSizeEmpty
		}
	}

	if req.Action == consts.UISceneOptTypeMouseClicking {
		if IsPointerStructNil(req.ActionDetail.MouseClicking) {
			log.Logger.Error("UISceneOptTypeSwitchFrameByIndex MouseClicking StructIsEmpty")
			return "", errmsg.ErrUISceneMouseClickingEmpty
		}
		if IsPointerStructNil(req.ActionDetail.MouseClicking.Element) {
			log.Logger.Error("UISceneOptTypeSwitchFrameByIndex MouseClicking StructIsEmpty")
			return "", errmsg.ErrUISceneMouseClickingEmpty
		}

		if t, ok := consts.MouseClickingType[req.ActionDetail.MouseClicking.Type]; ok {
			targetElement := req.ActionDetail.MouseClicking.Element
			elementName := targetElement.Name
			if targetElement.TargetType == consts.UISceneElementTypeSelect {
				if len(targetElement.ElementID) == 0 {
					return "", errmsg.ErrElementLocatorNotFound
				}
			}
			if targetElement.TargetType == consts.UISceneElementTypeCustom {
				for _, locator := range targetElement.CustomLocators {
					if len(locator.Value) == 0 {
						return "", errmsg.ErrElementLocatorNotFound
					}
					method := "选择器"
					if elementMethod, ok := consts.ElementMethodType[locator.Method]; ok {
						method = elementMethod
					}

					elementName = fmt.Sprintf("{%s_%s_%s}", method, locator.Type, locator.Value)
				}
			}
			name = fmt.Sprintf("%s %s", t, elementName)
		}
	}

	if req.Action == consts.UISceneOptTypeMouseScrolling {
		if IsPointerStructNil(req.ActionDetail.MouseScrolling) {
			log.Logger.Error("UISceneOptTypeSwitchFrameByIndex MouseScrolling StructIsEmpty")
			return "", errmsg.ErrUISceneMouseScrollingEmpty
		}
		if t, ok := consts.MouseScrollingType[req.ActionDetail.MouseScrolling.Type]; ok {
			name = fmt.Sprintf("%s %s", t, strconv.Itoa(req.ActionDetail.MouseScrolling.ScrollDistance))
		}

		if req.ActionDetail.MouseScrolling.Type == consts.MouseScrollingTypeMouseElementAppears {
			targetElement := req.ActionDetail.MouseScrolling.Element
			elementName := targetElement.Name
			if targetElement.TargetType == consts.UISceneElementTypeSelect {
				if len(targetElement.ElementID) == 0 {
					return "", errmsg.ErrElementLocatorNotFound
				}
			}
			if targetElement.TargetType == consts.UISceneElementTypeCustom {
				for _, locator := range targetElement.CustomLocators {
					if len(locator.Value) == 0 {
						return "", errmsg.ErrElementLocatorNotFound
					}
					method := "选择器"
					if elementMethod, ok := consts.ElementMethodType[locator.Method]; ok {
						method = elementMethod
					}

					elementName = fmt.Sprintf("{%s_%s_%s}", method, locator.Type, locator.Value)
				}
			}
			name = fmt.Sprintf("鼠标滚动到 %s 出现", elementName)
		}
	}

	if req.Action == consts.UISceneOptTypeMouseMovement {
		if IsPointerStructNil(req.ActionDetail.MouseMovement) {
			log.Logger.Error("UISceneOptTypeSwitchFrameByIndex MouseMovement StructIsEmpty")
			return "", errmsg.ErrUISceneMouseMovementEmpty
		}
	}

	if req.Action == consts.UISceneOptTypeMouseDragging {
		if IsPointerStructNil(req.ActionDetail.MouseDragging) {
			log.Logger.Error("UISceneOptTypeSwitchFrameByIndex MouseDragging StructIsEmpty")
			return "", errmsg.ErrUISceneMouseDraggingEmpty
		}

		if req.ActionDetail.MouseDragging.Type == consts.MouseDraggingTypeDragElement {
			targetElement := req.ActionDetail.MouseDragging.Element
			elementName := targetElement.Name
			if targetElement.TargetType == consts.UISceneElementTypeSelect {
				if len(targetElement.ElementID) == 0 {
					return "", errmsg.ErrElementLocatorNotFound
				}
			}
			if targetElement.TargetType == consts.UISceneElementTypeCustom {
				for _, locator := range targetElement.CustomLocators {
					if len(locator.Value) == 0 {
						return "", errmsg.ErrElementLocatorNotFound
					}
					method := "选择器"
					if elementMethod, ok := consts.ElementMethodType[locator.Method]; ok {
						method = elementMethod
					}

					elementName = fmt.Sprintf("{%s_%s_%s}", method, locator.Type, locator.Value)
				}
			}

			targetTargetElement := req.ActionDetail.MouseDragging.TarGetElement
			targetElementName := targetTargetElement.Name
			if targetTargetElement.TargetType == consts.UISceneElementTypeSelect {
				if len(targetTargetElement.ElementID) > 0 {
					return "", errmsg.ErrElementLocatorNotFound
				}
			}
			if targetTargetElement.TargetType == consts.UISceneElementTypeCustom {
				for _, locator := range targetTargetElement.CustomLocators {
					if len(locator.Value) == 0 {
						return "", errmsg.ErrElementLocatorNotFound
					}
					method := "选择器"
					if elementMethod, ok := consts.ElementMethodType[locator.Method]; ok {
						method = elementMethod
					}

					targetElementName = fmt.Sprintf("{%s_%s_%s}", method, locator.Type, locator.Value)
				}
			}

			name = fmt.Sprintf("将 %s 拖动至 %s", elementName, targetElementName)
		}

		if req.ActionDetail.MouseDragging.Type == consts.MouseDraggingTypeDragByPointCoordinates {
			name = fmt.Sprintf("拖动至 %s、%s",
				strconv.FormatFloat(req.ActionDetail.MouseDragging.EndPointCoordinates.X, 'f', -1, 64),
				strconv.FormatFloat(req.ActionDetail.MouseDragging.EndPointCoordinates.Y, 'f', -1, 64),
			)
		}
	}

	if req.Action == consts.UISceneOptTypeInputOperations {
		if IsPointerStructNil(req.ActionDetail.InputOperations) {
			log.Logger.Error("UISceneOptTypeSwitchFrameByIndex InputOperations StructIsEmpty")
			return "", errmsg.ErrUISceneInputOperationsEmpty
		}

		if t, ok := consts.InputOperationsType[req.ActionDetail.InputOperations.Type]; ok {
			name = fmt.Sprintf("%s %s", t, req.ActionDetail.InputOperations.InputContent)
		}

		if req.ActionDetail.InputOperations.Type == consts.InputOperationsOnElement {
			targetElement := req.ActionDetail.InputOperations.Element
			elementName := targetElement.Name
			if targetElement.TargetType == consts.UISceneElementTypeSelect {
				if len(targetElement.ElementID) == 0 {
					return "", errmsg.ErrElementLocatorNotFound
				}
			}
			if targetElement.TargetType == consts.UISceneElementTypeCustom {
				for _, locator := range targetElement.CustomLocators {
					if len(locator.Value) == 0 {
						return "", errmsg.ErrElementLocatorNotFound
					}
					method := "选择器"
					if elementMethod, ok := consts.ElementMethodType[locator.Method]; ok {
						method = elementMethod
					}

					elementName = fmt.Sprintf("{%s_%s_%s}", method, locator.Type, locator.Value)
				}
			}
			name = fmt.Sprintf("在 %s 上输入 %s", elementName, req.ActionDetail.InputOperations.InputContent)
		}
	}

	if req.Action == consts.UISceneOptTypeWaitEvents {
		if IsPointerStructNil(req.ActionDetail.WaitEvents) {
			log.Logger.Error("UISceneOptTypeSwitchFrameByIndex WaitEvents StructIsEmpty")
			return "", errmsg.ErrUISceneWaitEventsEmpty
		}

		if t, ok := consts.WaitEventsType[req.ActionDetail.WaitEvents.Type]; ok {
			if req.ActionDetail.WaitEvents.Type == consts.WaitEventsFixedTime {
				name = fmt.Sprintf("%s %d", t, req.ActionDetail.WaitEvents.WaitTime)
			}
			if req.ActionDetail.WaitEvents.Type == consts.WaitEventsElementExist ||
				req.ActionDetail.WaitEvents.Type == consts.WaitEventsElementNonExist ||
				req.ActionDetail.WaitEvents.Type == consts.WaitEventsElementDisplayed ||
				req.ActionDetail.WaitEvents.Type == consts.WaitEventsElementNotDisplayed ||
				req.ActionDetail.WaitEvents.Type == consts.WaitEventsElementEditable ||
				req.ActionDetail.WaitEvents.Type == consts.WaitEventsElementNotEditable {

				targetElement := req.ActionDetail.WaitEvents.Element
				elementName := targetElement.Name
				if targetElement.TargetType == consts.UISceneElementTypeSelect {
					if len(targetElement.ElementID) == 0 {
						return "", errmsg.ErrElementLocatorNotFound
					}
				}
				if targetElement.TargetType == consts.UISceneElementTypeCustom {
					for _, locator := range targetElement.CustomLocators {
						if len(locator.Value) == 0 {
							return "", errmsg.ErrElementLocatorNotFound
						}
						method := "选择器"
						if elementMethod, ok := consts.ElementMethodType[locator.Method]; ok {
							method = elementMethod
						}

						elementName = fmt.Sprintf("{%s_%s_%s}", method, locator.Type, locator.Value)
					}
				}
				name = fmt.Sprintf("%s %s", t, elementName)
			}
			if req.ActionDetail.WaitEvents.Type == consts.WaitEventsTextAppearance ||
				req.ActionDetail.WaitEvents.Type == consts.WaitEventsTextDisappearance {
				name = fmt.Sprintf("%s", t)
			}
		}
	}

	if req.Action == consts.UISceneOptTypeIfCondition {
		if IsPointerStructNil(req.ActionDetail.IfCondition) {
			log.Logger.Error("IfCondition StructIsEmpty")
			return "", errmsg.ErrUISceneIfConditionEmpty
		}

		name = fmt.Sprintf("%s", "条件步骤")
	}

	if req.Action == consts.UISceneOptTypeForLoop {
		if IsPointerStructNil(req.ActionDetail.ForLoop) {
			log.Logger.Error("UISceneOptTypeForLoop StructIsEmpty")
			return "", errmsg.ErrUISceneForLoopEmpty
		}

		if t, ok := consts.ForLoopType[req.ActionDetail.ForLoop.Type]; ok {
			name = fmt.Sprintf("%s", t)
		}
	}

	if req.Action == consts.UISceneOptTypeWhileLoop {
		if IsPointerStructNil(req.ActionDetail.WhileLoop) {
			log.Logger.Error("UISceneOptTypeWhileLoop StructIsEmpty")
			return "", errmsg.ErrUISceneWhileLoopEmpty
		}

		name = fmt.Sprintf("%s", "条件步骤")
	}

	if req.Action == consts.UISceneOptTypeAssert {
		if IsPointerStructNil(req.ActionDetail.Assert) {
			log.Logger.Error("UISceneOptTypeAssert StructIsEmpty")
			return "", errmsg.ErrUISceneAssertEmpty
		}
		if err := AssertCheckSaveOperatorReq(req.ActionDetail.Assert); err != nil {
			log.Logger.Error("AssertCheckSaveOperatorReq StructIsEmpty")
			return "", err
		}
		if t, ok := consts.AssertType[req.ActionDetail.Assert.Type]; ok {
			name = fmt.Sprintf("%s", t)
		}
	}

	if req.Action == consts.UISceneOptTypeDataWithdraw {
		if err := WithdrawCheckSaveOperatorReq(req.ActionDetail.DataWithdraw); err != nil {
			log.Logger.Error("WithdrawCheckSaveOperatorReq StructIsEmpty")
			return "", err
		}
		name = fmt.Sprintf("%s", req.ActionDetail.DataWithdraw.Name)
	}

	for _, a := range req.Asserts {
		if err := AssertCheckSaveOperatorReq(a); err != nil {
			log.Logger.Error("AssertCheckSaveOperatorReq StructIsEmpty")
			return "", err
		}
	}

	for _, a := range req.DataWithdraws {
		if err := WithdrawCheckSaveOperatorReq(a); err != nil {
			log.Logger.Error("WithdrawCheckSaveOperatorReq StructIsEmpty")
			return "", err
		}
	}

	return name, nil
}

// AssertCheckSaveOperatorReq 检查断言中类型的值是否为空
func AssertCheckSaveOperatorReq(a *rao.AutomationAssert) error {
	if a.Type == consts.UISceneOptElementExists {
		if IsPointerStructNil(a.Element) {
			return errmsg.ErrUISceneAssertElementExistsEmpty
		}
		targetElement := a.Element
		if targetElement.TargetType == consts.UISceneElementTypeSelect {
			if len(targetElement.ElementID) == 0 {
				return errmsg.ErrElementLocatorNotFound
			}
		}
		if targetElement.TargetType == consts.UISceneElementTypeCustom {
			for _, locator := range targetElement.CustomLocators {
				if len(locator.Value) == 0 {
					return errmsg.ErrElementLocatorNotFound
				}
			}
		}
	}

	if a.Type == consts.UISceneOptElementNotExists {
		if IsPointerStructNil(a.Element) {
			return errmsg.ErrUISceneAssertElementNotExistsEmpty
		}
		targetElement := a.Element
		if targetElement.TargetType == consts.UISceneElementTypeSelect {
			if len(targetElement.ElementID) == 0 {
				return errmsg.ErrElementLocatorNotFound
			}
		}
		if targetElement.TargetType == consts.UISceneElementTypeCustom {
			for _, locator := range targetElement.CustomLocators {
				if len(locator.Value) == 0 {
					return errmsg.ErrElementLocatorNotFound
				}
			}
		}
	}

	if a.Type == consts.UISceneOptElementDisplayed {
		if IsPointerStructNil(a.Element) {
			return errmsg.ErrUISceneAssertElementDisplayedEmpty
		}
		targetElement := a.Element
		if targetElement.TargetType == consts.UISceneElementTypeSelect {
			if len(targetElement.ElementID) == 0 {
				return errmsg.ErrElementLocatorNotFound
			}
		}
		if targetElement.TargetType == consts.UISceneElementTypeCustom {
			for _, locator := range targetElement.CustomLocators {
				if len(locator.Value) == 0 {
					return errmsg.ErrElementLocatorNotFound
				}
			}
		}
	}

	if a.Type == consts.UISceneOptElementNotDisplayed {
		if IsPointerStructNil(a.Element) {
			return errmsg.ErrUISceneAssertElementNotDisplayedEmpty
		}
		targetElement := a.Element
		if targetElement.TargetType == consts.UISceneElementTypeSelect {
			if len(targetElement.ElementID) == 0 {
				return errmsg.ErrElementLocatorNotFound
			}
		}
		if targetElement.TargetType == consts.UISceneElementTypeCustom {
			for _, locator := range targetElement.CustomLocators {
				if len(locator.Value) == 0 {
					return errmsg.ErrElementLocatorNotFound
				}
			}
		}
	}

	if a.Type == consts.UISceneOptTextExists {
		if IsPointerStructNil(a.TextExists) {
			return errmsg.ErrUISceneAssertTextExistsEmpty
		}
	}

	if a.Type == consts.UISceneOptTextNotExists {
		if IsPointerStructNil(a.TextNotExists) {
			return errmsg.ErrUISceneAssertTextNotExistsEmpty
		}
	}

	if a.Type == consts.UISceneOptVariableAssertion {
		if IsPointerStructNil(a.VariableAssertion) {
			return errmsg.ErrUISceneAssertVariableAssertionEmpty
		}
	}

	if a.Type == consts.UISceneOptExpressionAssertion {
		if IsPointerStructNil(a.ExpressionAssertion) {
			return errmsg.ErrUISceneAssertExpressionAssertionEmpty
		}
	}

	if a.Type == consts.UISceneOptElementAttributeAssertion {
		if IsPointerStructNil(a.ElementAttributeAssert) {
			return errmsg.ErrUISceneAssertElementAttributeAssertEmpty
		}
		targetElement := a.Element
		if targetElement.TargetType == consts.UISceneElementTypeSelect {
			if len(targetElement.ElementID) == 0 {
				return errmsg.ErrElementLocatorNotFound
			}
		}
		if targetElement.TargetType == consts.UISceneElementTypeCustom {
			for _, locator := range targetElement.CustomLocators {
				if len(locator.Value) == 0 {
					return errmsg.ErrElementLocatorNotFound
				}
			}
		}
	}

	if a.Type == consts.UISceneOptPageAttributeAssertion {
		if IsPointerStructNil(a.PageAttributeAssert) {
			return errmsg.ErrUISceneAssertPageAttributeAssertEmpty
		}
	}

	return nil
}

func WithdrawCheckSaveOperatorReq(a *rao.DataWithdraw) error {
	if a.WithdrawType == consts.UISceneWithdrawTypeElement {
		if a.ElementMethod == nil {
			return errmsg.ErrUISceneWithdrawElementExistsEmpty
		}
		if IsPointerStructNil(a.ElementMethod.Element) {
			return errmsg.ErrUISceneWithdrawElementExistsEmpty
		}
		targetElement := a.ElementMethod.Element
		if targetElement.TargetType == consts.UISceneElementTypeSelect {
			if len(targetElement.ElementID) == 0 {
				return errmsg.ErrElementLocatorNotFound
			}
		}
		if targetElement.TargetType == consts.UISceneElementTypeCustom {
			for _, locator := range targetElement.CustomLocators {
				if len(locator.Value) == 0 {
					return errmsg.ErrElementLocatorNotFound
				}
			}
		}
	}

	if len(a.Name) == 0 {
		return errmsg.ErrUISceneRequired
	}

	return nil
}

func SaveOperator(ctx *gin.Context, userID string, req *rao.UISceneSaveOperatorReq) (string, error) {
	req.Name, _ = CheckSaveOperatorAndGetNameReq(ctx, req)
	req.OperatorID = uuid.GetUUID()
	operator := packer.TransSaveOperatorReqToOperatorModel(req)
	operatorMao := packer.TransSaveOperatorReqToUISceneOperatorMao(req)
	elementMap := packer.TransSaveOperatorReqToUISceneElementMao(req)
	// 添加关联元素
	sceneElements := make([]*model.UISceneElement, 0, len(elementMap))
	for elementID, _ := range elementMap {
		sceneElement := &model.UISceneElement{
			SceneID:    operator.SceneID,
			OperatorID: operator.OperatorID,
			ElementID:  elementID,
			TeamID:     req.TeamID,
			Status:     consts.UISceneStatusNormal,
		}
		sceneElements = append(sceneElements, sceneElement)
	}

	// 1: 从底层插入，获取最大 sort + 1
	// 2: 为空从1开始
	if req.Sort == 0 {
		so := query.Use(dal.DB()).UISceneOperator
		lastOperator, err := so.WithContext(ctx).Where(so.SceneID.Eq(req.SceneID), so.ParentID.Eq(operator.ParentID)).
			Order(so.Sort.Desc()).First()
		if err != nil {
			operator.Sort = 1
		} else {
			operator.Sort = lastOperator.Sort + 1
		}
	}

	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectUISceneOperator)
	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		// 查询当前接口是否存在
		_, err := tx.UISceneOperator.WithContext(ctx).Where(tx.UISceneOperator.OperatorID.Eq(req.OperatorID)).First()
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			// step1 :新增步骤
			if err = tx.UISceneOperator.WithContext(ctx).Create(operator); err != nil {
				return err
			}
			// step2 :新增集合
			if _, err = collection.InsertOne(ctx, operatorMao); err != nil {
				return err
			}
			// step3 :插入日志
			//if err = record.InsertCreate(ctx, req.TeamID, userID, record.OperationOperateCreateUISceneAPI, req.Name); err != nil {
			//	return err
			//}
			// step4 :添加关联元素
			if len(sceneElements) > 0 {
				if err = tx.UISceneElement.WithContext(ctx).Create(sceneElements...); err != nil {
					return err
				}
			}

			// step5 :同级数据重新排序
			if req.IsReSort {
				so := tx.UISceneOperator
				operatorList, err := so.WithContext(ctx).Where(so.SceneID.Eq(req.SceneID), so.ParentID.Eq(operator.ParentID)).
					Order(so.Sort, so.CreatedAt).Find()
				if err != nil {
					return err
				}

				var operatorSort = make(map[string]int32)
				for index, o := range operatorList {
					operatorSort[o.OperatorID] = int32(index + 1)
				}
				for _, o := range operatorList {
					_, err = so.WithContext(ctx).Where(so.OperatorID.Eq(o.OperatorID), so.SceneID.Eq(o.SceneID)).
						UpdateSimple(so.Sort.Value(operatorSort[o.OperatorID]), so.ParentID.Value(o.ParentID))
					if err != nil {
						return err
					}
				}
			}

			// step6 :修改元素属性
			if len(elementMap) > 0 {
				for elementID, element := range elementMap {
					marshal, _ := json.Marshal(element.Locators)
					locators := string(marshal)
					if _, err = tx.Element.WithContext(ctx).Where(
						tx.Element.ElementID.Eq(elementID),
						tx.Element.TeamID.Eq(req.TeamID),
					).Update(tx.Element.Locators, locators); err != nil {
						return err
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	// 同步绑定的场景
	if err = SyncBindSceneOperator(ctx, req.SceneID, req.TeamID, userID); err != nil {
		log.Logger.Error("SyncBindSceneOperator err", proof.WithError(err))
	}

	return req.OperatorID, nil
}

// UpdateOperator 修改场景步骤
func UpdateOperator(ctx *gin.Context, userID string, req *rao.UISceneSaveOperatorReq) error {
	req.Name, _ = CheckSaveOperatorAndGetNameReq(ctx, req)
	operator := packer.TransSaveOperatorReqToOperatorModel(req)
	operatorMao := packer.TransSaveOperatorReqToUISceneOperatorMao(req)
	elementMap := packer.TransSaveOperatorReqToUISceneElementMao(req)

	elementIDs := make([]string, 0, len(elementMap))
	for elementID := range elementMap {
		elementIDs = append(elementIDs, elementID)
	}
	elementIDs = public.SliceUnique(elementIDs)
	oldElementIDs := make([]string, 0)
	se := dal.GetQuery().UISceneElement
	if err := se.WithContext(ctx).Where(
		se.OperatorID.Eq(operator.OperatorID),
		se.SceneID.Eq(operator.SceneID),
		se.TeamID.Eq(req.TeamID),
	).Pluck(se.ElementID, &oldElementIDs); err != nil {
		return err
	}

	createIDs := public.SliceDiff(elementIDs, oldElementIDs)
	delIDs := public.SliceDiff(oldElementIDs, elementIDs)

	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectUISceneOperator)

	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		// 查询当前接口是否存在
		_, err := tx.UISceneOperator.WithContext(ctx).Where(tx.UISceneOperator.OperatorID.Eq(req.OperatorID)).First()
		if err != nil {
			return err
		}

		if _, err = tx.UISceneOperator.WithContext(ctx).Where(tx.UISceneOperator.OperatorID.Eq(req.OperatorID)).Updates(operator); err != nil {
			return err
		}

		if _, err = collection.UpdateOne(ctx, bson.D{{"operator_id", operator.OperatorID}}, bson.M{"$set": operatorMao}); err != nil {
			return err
		}

		//if err = record.InsertUpdate(ctx, req.TeamID, userID, record.OperationOperateUpdateUISceneAPI, req.Name); err != nil {
		//	return err
		//}

		// 新增元素
		if len(createIDs) > 0 {
			sceneElements := make([]*model.UISceneElement, 0, len(elementIDs))
			for _, id := range elementIDs {
				sceneElement := &model.UISceneElement{
					SceneID:    operator.SceneID,
					OperatorID: operator.OperatorID,
					ElementID:  id,
					TeamID:     req.TeamID,
				}
				sceneElements = append(sceneElements, sceneElement)
			}
			if err = tx.UISceneElement.WithContext(ctx).Create(sceneElements...); err != nil {
				return err
			}
		}

		// 删除元素
		if len(delIDs) > 0 {
			if _, err = tx.UISceneElement.WithContext(ctx).Where(tx.UISceneElement.ElementID.In(delIDs...),
				tx.UISceneElement.OperatorID.Eq(operator.OperatorID),
				tx.UISceneElement.SceneID.Eq(operator.SceneID),
				tx.UISceneElement.TeamID.Eq(req.TeamID),
			).Delete(); err != nil {
				return err
			}
		}

		// step6 :修改元素属性
		if len(elementMap) > 0 {
			for elementID, element := range elementMap {
				marshal, _ := json.Marshal(element.Locators)
				locators := string(marshal)
				if _, err = tx.Element.WithContext(ctx).Where(
					tx.Element.ElementID.Eq(elementID),
					tx.Element.TeamID.Eq(req.TeamID),
				).Update(tx.Element.Locators, locators); err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	// 同步绑定的场景
	if err = SyncBindSceneOperator(ctx, req.SceneID, req.TeamID, userID); err != nil {
		log.Logger.Error("SyncBindSceneOperator err", proof.WithError(err))
	}

	return nil
}

// DetailOperator 步骤详情
func DetailOperator(ctx *gin.Context, req *rao.UISceneDetailOperatorReq) (*rao.UISceneOperator, error) {
	us := query.Use(dal.DB()).UIScene
	scene, err := us.WithContext(ctx).Where(us.SceneID.Eq(req.SceneID)).First()
	if err != nil {
		return nil, err
	}

	uso := query.Use(dal.DB()).UISceneOperator
	operator, err := uso.WithContext(ctx).Where(
		uso.SceneID.Eq(req.SceneID),
		uso.OperatorID.Eq(req.OperatorID),
	).First()
	if err != nil {
		return nil, err
	}

	// 查询接口详情数据
	ret := &mao.SceneOperator{}
	collection := dal.GetMongo().Database(dal.MongoDB()).Collection(consts.CollectUISceneOperator)
	err = collection.FindOne(ctx, bson.D{{"operator_id", req.OperatorID}}).Decode(&ret)
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, err
	}

	if err == mongo.ErrNoDocuments {
		return nil, nil
	}

	// 获取当前步骤依赖的元素
	se := query.Use(dal.DB()).UISceneElement
	sceneElementList, err := se.WithContext(ctx).Where(
		se.OperatorID.Eq(req.OperatorID),
		se.SceneID.Eq(req.SceneID),
		se.Status.Eq(consts.UISceneStatusNormal),
	).Find()
	if err != nil {
		return nil, err
	}
	elementIDs := make([]string, 0, len(sceneElementList))
	for _, e := range sceneElementList {
		elementIDs = append(elementIDs, e.ElementID)
	}

	e := query.Use(dal.DB()).Element
	elementList, err := e.WithContext(ctx).Where(e.ElementID.In(elementIDs...)).Find()
	if err != nil {
		return nil, err
	}

	return packer.TransSceneOperatorToRaoSceneOperator(scene, operator, ret, elementList), nil
}

// ListOperator 列表
func ListOperator(ctx *gin.Context, teamID string, sceneID string) ([]*rao.UISceneOperator, error) {
	s := query.Use(dal.DB()).UIScene
	scene, err := s.WithContext(ctx).Where(
		s.SceneID.Eq(sceneID), s.TeamID.Eq(teamID),
	).First()

	so := query.Use(dal.DB()).UISceneOperator
	operators, err := so.WithContext(ctx).Where(
		so.SceneID.Eq(sceneID),
	).Order(so.Sort, so.ParentID).Find()
	if err != nil {
		return nil, err
	}

	ret := make([]*rao.UISceneOperator, 0, len(operators))
	for _, o := range operators {
		operator := &rao.UISceneOperator{
			TeamID:     scene.TeamID,
			SceneID:    o.SceneID,
			OperatorID: o.OperatorID,
			ParentID:   o.ParentID,
			Name:       o.Name,
			Sort:       o.Sort,
			Status:     o.Status,
			Type:       o.Type,
			Action:     o.Action,
		}

		ret = append(ret, operator)
	}

	return ret, nil
}

// OperatorSort 排序
func OperatorSort(ctx *gin.Context, userID string, req *rao.UIScenesOperatorStepReq) error {
	var (
		sceneID string
		teamID  string
	)

	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		for _, s := range req.Operators {
			_, err := tx.UISceneOperator.WithContext(ctx).Where(
				tx.UISceneOperator.OperatorID.Eq(s.OperatorID),
				tx.UISceneOperator.SceneID.Eq(s.SceneID),
			).UpdateSimple(
				tx.UISceneOperator.Sort.Value(s.Sort),
				tx.UISceneOperator.ParentID.Value(s.ParentID))
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	// 同步绑定的场景
	if err = SyncBindSceneOperator(ctx, sceneID, teamID, userID); err != nil {
		log.Logger.Error("SyncBindSceneOperator err", proof.WithError(err))
	}

	return nil
}

// OperatorCopy 复制步骤
func OperatorCopy(ctx *gin.Context, userID string, req *rao.UISceneCopyOperatorReq) error {
	// step1:查询步骤复制一遍
	// step2:修改 sort
	s := query.Use(dal.DB()).UIScene
	uiScene, err := s.WithContext(ctx).Where(s.SceneID.Eq(req.SceneID)).First()
	if err != nil {
		return err
	}

	so := query.Use(dal.DB()).UISceneOperator
	operator, err := so.WithContext(ctx).Where(so.SceneID.Eq(req.SceneID), so.OperatorID.Eq(req.OperatorID)).First()
	if err != nil {
		return err
	}

	if err = HandleCopyOperator(ctx, userID, uiScene, operator, uiScene, operator.ParentID); err != nil {
		return err
	}

	// 同步绑定的场景
	if err = SyncBindSceneOperator(ctx, req.SceneID, uiScene.TeamID, userID); err != nil {
		log.Logger.Error("SyncBindSceneOperator err", proof.WithError(err))
	}

	return nil
}

func HandleCopyOperator(
	ctx *gin.Context,
	userID string,
	uiScene *model.UIScene,
	operator *model.UISceneOperator,
	toUiScene *model.UIScene,
	parentID string,
) error {
	// step1 :查询步骤详情
	so := query.Use(dal.DB()).UISceneOperator
	d, err := DetailOperator(ctx, &rao.UISceneDetailOperatorReq{
		TeamID:     uiScene.TeamID,
		SceneID:    operator.SceneID,
		OperatorID: operator.OperatorID,
	})

	// step2 :添加步骤 && 重新排序
	saveOperatorReq := &rao.UISceneSaveOperatorReq{
		TeamID:       toUiScene.TeamID,
		SceneID:      toUiScene.SceneID,
		ParentID:     parentID,
		Name:         d.Name,
		Sort:         d.Sort,
		Status:       d.Status,
		Type:         d.Type,
		Action:       d.Action,
		ActionDetail: d.ActionDetail,
		Settings:     d.Settings,
		Asserts:      d.Asserts,
	}
	parentID, err = SaveOperator(ctx, userID, saveOperatorReq)
	if err != nil {
		return err
	}

	// step3 :查询是否有子集、递归操作
	sonOperators, err := so.WithContext(ctx).Where(so.SceneID.Eq(uiScene.SceneID), so.ParentID.Eq(operator.OperatorID)).Find()
	if err != nil {
		return err
	}
	// 如果 sonOperators 不等于 nil
	if len(sonOperators) > 0 {
		for _, sonOperator := range sonOperators {
			if err = HandleCopyOperator(ctx, userID, uiScene, sonOperator, toUiScene, parentID); err != nil {
				return err
			}
		}
	}

	return nil
}

func OperatorSetStatus(ctx *gin.Context, userID string, req *rao.UISceneSetStatusOperatorReq) error {
	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		_, err := tx.UISceneOperator.WithContext(ctx).Where(
			tx.UISceneOperator.OperatorID.In(req.OperatorIDs...),
			tx.UISceneOperator.SceneID.Eq(req.SceneID),
		).UpdateSimple(tx.UISceneOperator.Status.Value(req.Status))
		if err != nil {
			return err
		}
		return nil
	})

	// 同步绑定的场景
	if err = SyncBindSceneOperator(ctx, req.SceneID, req.TeamID, userID); err != nil {
		log.Logger.Error("SyncBindSceneOperator err", proof.WithError(err))
	}

	return err
}

func OperatorDelete(ctx *gin.Context, userID string, req *rao.UISceneDeleteOperatorReq) error {
	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		var delParentIDs = make([]string, 0)
		if err := tx.UISceneOperator.WithContext(ctx).Where(
			tx.UISceneOperator.OperatorID.In(req.OperatorIDs...),
			tx.UISceneOperator.SceneID.Eq(req.SceneID),
		).Pluck(tx.UISceneOperator.ParentID, &delParentIDs); err != nil {
			return err
		}
		delParentIDs = public.SliceUnique(delParentIDs)

		_, err := tx.UISceneOperator.WithContext(ctx).Where(
			tx.UISceneOperator.OperatorID.In(req.OperatorIDs...),
			tx.UISceneOperator.SceneID.Eq(req.SceneID),
		).Delete()
		if err != nil {
			return err
		}

		_, err = tx.UISceneElement.WithContext(ctx).Where(
			tx.UISceneElement.OperatorID.In(req.OperatorIDs...),
			tx.UISceneElement.SceneID.Eq(req.SceneID),
			tx.UISceneElement.TeamID.Eq(req.TeamID),
		).Delete()
		if err != nil {
			return err
		}

		// step5 :同级数据重新排序
		for _, d := range delParentIDs {
			so := tx.UISceneOperator
			operatorList, err := so.WithContext(ctx).Where(so.SceneID.Eq(req.SceneID), so.ParentID.Eq(d)).
				Order(so.Sort, so.CreatedAt).Find()
			if err != nil {
				return err
			}

			var operatorSort = make(map[string]int32)
			for index, o := range operatorList {
				operatorSort[o.OperatorID] = int32(index + 1)
			}
			for _, o := range operatorList {
				_, err = so.WithContext(ctx).Where(so.OperatorID.Eq(o.OperatorID), so.SceneID.Eq(o.SceneID)).
					UpdateSimple(so.Sort.Value(operatorSort[o.OperatorID]), so.ParentID.Value(o.ParentID))
				if err != nil {
					return err
				}
			}
		}

		return nil
	})

	// 同步绑定的场景
	if err = SyncBindSceneOperator(ctx, req.SceneID, req.TeamID, userID); err != nil {
		log.Logger.Error("SyncBindSceneOperator err", proof.WithError(err))
	}

	return err
}

// SyncBindSceneOperator 同步绑定的场景
func SyncBindSceneOperator(ctx *gin.Context, sceneID, teamID, userID string) error {
	// 当前是场景中的scene，查询当前场景是否有实时同步
	// 当前是计划中的scene，查询是否有引用，修改引用, 同上一步
	s := query.Use(dal.DB()).UIScene
	scene, err := s.WithContext(ctx).Where(s.SceneID.Eq(sceneID), s.TeamID.Eq(teamID)).First()
	if err != nil {
		return err
	}
	if scene.Source == consts.UISceneSource {
		if err := SyncSourceSceneOperator(ctx, scene.SceneID, teamID, sceneID); err != nil {
			return err
		}
	}
	if scene.Source == consts.UISceneSourcePlan {
		if err := SyncSceneOperator(ctx, scene.SceneID, teamID); err != nil {
			return err
		}
	}

	return nil
}

// SyncSourceSceneOperator 同步引用这个场景的步骤   ignoreSceneID 不同步场景ID
func SyncSourceSceneOperator(ctx *gin.Context, sceneID, teamID, ignoreSceneID string) error {
	// step1: 查询当前场景是否有实时同步
	ss := query.Use(dal.DB()).UISceneSync
	var syncSceneIDs = make([]string, 0)
	if err := ss.WithContext(ctx).Where(
		ss.SourceSceneID.Eq(sceneID),
		ss.SceneID.Neq(ignoreSceneID),
		ss.TeamID.Eq(teamID),
		ss.SyncMode.Eq(consts.UISceneOptSyncModeAuto),
	).Pluck(ss.SceneID, &syncSceneIDs); err != nil {
		return err
	}

	if len(syncSceneIDs) == 0 {
		return nil
	}

	// step2: 查询要同步的场景
	s := query.Use(dal.DB()).UIScene
	sceneList, err := s.WithContext(ctx).Where(
		s.SceneID.In(syncSceneIDs...),
		s.TeamID.Eq(teamID),
		s.Status.Eq(consts.TargetStatusNormal),
	).Find()
	if err != nil {
		return err
	}

	// step3: 查询源场景步骤基本信息
	so := query.Use(dal.DB()).UISceneOperator
	sourceOperatorIDs := make([]string, 0)
	operators, err := so.WithContext(ctx).Where(so.SceneID.Eq(sceneID)).Find()
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

	// step5: 循环处理要同步的场景步骤
	for _, scene := range sceneList {
		//  生成新步骤
		if err = handleNewOperator(ctx, operators, scene, sceneOperators); err != nil {
			return err
		}
	}

	return nil
}

// SyncSceneOperator 同步计划中的场景，并同步引用这个场景的步骤
func SyncSceneOperator(ctx *gin.Context, sceneID, teamID string) error {
	// step1: 查询当前场景是否有实时同步
	ss := query.Use(dal.DB()).UISceneSync
	scene, _ := ss.WithContext(ctx).Where(
		ss.SceneID.Eq(sceneID),
		ss.TeamID.Eq(teamID),
		ss.SyncMode.Eq(consts.UISceneOptSyncModeAuto),
	).First()
	if scene == nil {
		return nil
	}

	// step2: 查询要同步的场景
	s := query.Use(dal.DB()).UIScene
	sourceScene, err := s.WithContext(ctx).Where(s.SceneID.Eq(scene.SourceSceneID), s.TeamID.Eq(teamID)).First()
	if err != nil {
		return nil
	}

	// step3: 查询源场景步骤基本信息
	so := query.Use(dal.DB()).UISceneOperator
	sourceOperatorIDs := make([]string, 0)
	operators, err := so.WithContext(ctx).Where(so.SceneID.Eq(sceneID)).Find()
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

	// step5: 循环处理要同步的场景步骤
	if err = handleNewOperator(ctx, operators, sourceScene, sceneOperators); err != nil {
		return err
	}

	if err = SyncSourceSceneOperator(ctx, scene.SourceSceneID, teamID, sceneID); err != nil {
		return err
	}

	return nil
}
