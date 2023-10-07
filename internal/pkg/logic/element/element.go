package element

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/uuid"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/query"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/errmsg"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/packer"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/public"
	"github.com/gin-gonic/gin"
	"gorm.io/gen"
	"gorm.io/gorm"
	"strings"
	"time"
)

func Save(ctx *gin.Context, userID string, req *rao.ElementSaveReq) (string, error) {
	// 元素名称不能重复
	e := query.Use(dal.DB()).Element
	if _, err := e.WithContext(ctx).Where(e.TeamID.Eq(req.TeamID), e.Name.Eq(req.Name),
		e.ElementType.Eq(consts.ElementTypeDefault),
		e.Source.Eq(req.Source), e.ParentID.Eq(req.ParentID),
	).First(); err == nil {
		return "", errmsg.ErrElementNameRepeat
	}

	// 验证值
	if len(req.Locators) == 0 {
		return "", errmsg.ErrElementLocatorNotFound
	}
	for _, locator := range req.Locators {
		if len(locator.Value) == 0 {
			return "", errmsg.ErrElementLocatorNotFound
		}
	}

	elementID := uuid.GetUUID()
	element := packer.TransSaveElementReqToElementModel(req, elementID, userID)

	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		if err := tx.Element.WithContext(ctx).Create(element); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return elementID, err
}

func Update(ctx *gin.Context, userID string, req *rao.ElementSaveReq) (string, error) {
	// 元素名称不能重复
	e := query.Use(dal.DB()).Element
	if _, err := e.WithContext(ctx).Where(e.TeamID.Eq(req.TeamID), e.Name.Eq(req.Name),
		e.ElementType.Eq(consts.ElementTypeDefault), e.ElementID.Neq(req.ElementID),
		e.Source.Eq(req.Source), e.ParentID.Eq(req.ParentID),
	).First(); err == nil {
		return "", errmsg.ErrElementNameRepeat
	}

	// 验证值
	if len(req.Locators) == 0 {
		return "", errmsg.ErrElementLocatorNotFound
	}
	for _, locator := range req.Locators {
		if len(locator.Value) == 0 {
			return "", errmsg.ErrElementLocatorNotFound
		}
	}

	element := packer.TransSaveElementReqToElementModel(req, req.ElementID, userID)

	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		if _, err := tx.Element.WithContext(ctx).Where(tx.Element.ElementID.Eq(req.ElementID)).Updates(element); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return req.ElementID, err
}

func Detail(ctx *gin.Context, elementID string, teamID string) (*rao.Element, error) {
	e := query.Use(dal.DB()).Element
	info, err := e.WithContext(ctx).Where(e.ElementID.Eq(elementID), e.TeamID.Eq(teamID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errmsg.ErrElementNotFound
		}
		return nil, err
	}

	return packer.TransModelElementToRaoElement(info), nil
}

func List(ctx *gin.Context, teamID string, req *rao.ElementListReq) ([]*rao.Element, int64, error) {
	if req.Size == 0 {
		req.Size = 20
	}
	if req.Page == 0 {
		req.Page = 1
	}
	limit := req.Size
	offset := (req.Page - 1) * req.Size
	parentID := req.ParentID
	name := strings.TrimSpace(req.Name)
	updatedTime := req.UpdatedTime
	locatorMethod := req.LocatorMethod
	locatorType := req.LocatorType
	locatorValue := strings.TrimSpace(req.LocatorValue)

	// 可能会传 [""] 过来
	filterLocatorType := make([]string, 0)
	for _, s := range locatorType {
		if len(s) > 0 {
			filterLocatorType = append(filterLocatorType, s)
		}
	}
	filterLocatorMethod := make([]string, 0)
	for _, m := range locatorMethod {
		if len(m) > 0 {
			filterLocatorMethod = append(filterLocatorMethod, m)
		}
	}

	e := dal.GetQuery().Element
	conditions := make([]gen.Condition, 0)
	conditions = append(conditions, e.TeamID.Eq(teamID))
	conditions = append(conditions, e.ElementType.Eq(consts.ElementTypeDefault))
	if parentID != "-1" { // -1 代表查询全部
		conditions = append(conditions, e.ParentID.Eq(parentID))
	}
	if len(name) > 0 {
		conditions = append(conditions, e.Name.Like(fmt.Sprintf("%%%s%%", name)))
	}

	if len(updatedTime) == 2 {
		layout := "2006-01-02"
		start, _ := time.Parse(layout, updatedTime[0])
		t, _ := time.Parse(layout, updatedTime[1])

		newTime := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, t.Location())
		end := t.Add(newTime.Sub(t))

		conditions = append(conditions, e.UpdatedAt.Between(start, end))
	}

	locators, err := e.WithContext(ctx).Select(e.ID, e.Locators).Where(conditions...).Find()
	if err != nil {
		return nil, 0, err
	}
	if len(locatorValue) > 0 {
		var ids []int64
		for _, ls := range locators {
			raoLocators := make([]*rao.Locator, 0)
			_ = json.Unmarshal([]byte(ls.Locators), &raoLocators)
			for _, locator := range raoLocators {
				if strings.Contains(locator.Value, locatorValue) {
					ids = append(ids, ls.ID)
				}
			}
		}
		conditions = append(conditions, e.ID.In(ids...))
	}

	if len(filterLocatorType) > 0 {
		var ids []int64
		for _, ls := range locators {
			raoLocators := make([]*rao.Locator, 0)
			_ = json.Unmarshal([]byte(ls.Locators), &raoLocators)
			for _, locator := range raoLocators {
				for _, t := range filterLocatorType {
					if t == locator.Type {
						ids = append(ids, ls.ID)
					}
				}
			}
		}
		conditions = append(conditions, e.ID.In(ids...))
	}

	if len(filterLocatorMethod) > 0 {
		var ids []int64
		for _, ls := range locators {
			raoLocators := make([]*rao.Locator, 0)
			_ = json.Unmarshal([]byte(ls.Locators), &raoLocators)
			for _, locator := range raoLocators {
				for _, m := range filterLocatorMethod {
					if m == locator.Method {
						ids = append(ids, ls.ID)
					}
				}
			}
		}
		conditions = append(conditions, e.ID.In(ids...))
	}

	list, total, err := e.WithContext(ctx).Where(conditions...).Order(e.ID.Desc()).FindByPage(offset, limit)
	if err != nil {
		return nil, 0, err
	}

	parentIDs := make([]string, 0)
	elementIDs := make([]string, 0)
	for _, element := range list {
		parentIDs = append(parentIDs, element.ParentID)
		elementIDs = append(elementIDs, element.ElementID)
	}
	parentIDs = public.SliceUnique(parentIDs)
	parentList, err := e.WithContext(ctx).Select(e.ElementID, e.Name).Where(
		e.ElementType.Eq(consts.ElementTypeFolder),
		e.TeamID.Eq(teamID),
	).Find()
	if err != nil {
		return nil, 0, err
	}

	// 关联场景
	se := query.Use(dal.DB()).UISceneElement
	sceneElementList, err := se.WithContext(ctx).Where(
		se.ElementID.In(elementIDs...),
		se.TeamID.Eq(teamID),
		se.Status.Eq(consts.UISceneStatusNormal),
	).Find()
	if err != nil {
		return nil, 0, err
	}
	sceneIDs := make([]string, 0)
	for _, s := range sceneElementList {
		sceneIDs = append(sceneIDs, s.SceneID)
	}

	s := query.Use(dal.DB()).UIScene
	sceneList, err := s.WithContext(ctx).Where(s.SceneID.In(sceneIDs...), s.TeamID.Eq(teamID)).Find()
	if err != nil {
		return nil, 0, err
	}

	return packer.TransModelElementToRaoElementFolders(list, parentList, sceneElementList, sceneList), total, nil
}

func Remove(ctx *gin.Context, userIDs string, teamID string, elementIDs []string) error {
	// 查询是否有关联关系
	se := query.Use(dal.DB()).UISceneElement
	elements, err := se.WithContext(ctx).Where(se.ElementID.In(elementIDs...), se.TeamID.Eq(teamID)).Find()
	if err != nil {
		return err
	}
	if len(elements) > 0 {
		return errmsg.ErrElementNotDeleteReScene
	}

	err = query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		if _, err := tx.Element.WithContext(ctx).Where(
			tx.Element.ElementID.In(elementIDs...),
			tx.Element.TeamID.Eq(teamID),
		).Delete(); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func Sort(ctx *gin.Context, userIDs string, req *rao.ElementSortReq) error {
	e := query.Use(dal.DB()).Element

	// 名称不能重复
	var newFolderNames []string
	if err := e.WithContext(ctx).Select(e.Name).Where(
		e.ElementType.Eq(consts.ElementTypeDefault),
		e.TeamID.Eq(req.TeamID),
		e.ParentID.Eq(req.ParentID),
	).Pluck(e.Name, &newFolderNames); err != nil {
		return err
	}

	var oldElementNames []string
	if err := e.WithContext(ctx).Select(e.Name).Where(
		e.ElementType.Eq(consts.ElementTypeDefault),
		e.TeamID.Eq(req.TeamID),
		e.ElementID.In(req.ElementIDs...),
	).Pluck(e.Name, &oldElementNames); err != nil {
		return err
	}

	for _, newN := range newFolderNames {
		for _, oldN := range oldElementNames {
			if newN == oldN {
				return errmsg.ErrTargetSortNameAlreadyExist
			}
		}
	}

	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		_, err := tx.Element.WithContext(ctx).Where(tx.Element.TeamID.Eq(req.TeamID),
			tx.Element.ElementID.In(req.ElementIDs...)).UpdateSimple(
			tx.Element.ParentID.Value(req.ParentID))
		if err != nil {
			return err
		}
		return nil
	})

	return err
}
