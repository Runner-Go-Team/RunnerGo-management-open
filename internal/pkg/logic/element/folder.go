package element

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/uuid"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/query"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/errmsg"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/packer"
	"github.com/gin-gonic/gin"
)

func FolderSave(ctx *gin.Context, userID string, req *rao.ElementSaveFolderReq) (string, error) {
	// 目录名称不能重复
	e := query.Use(dal.DB()).Element
	if _, err := e.WithContext(ctx).Where(e.TeamID.Eq(req.TeamID), e.Name.Eq(req.Name),
		e.ElementType.Eq(consts.ElementTypeFolder),
		e.Source.Eq(req.Source), e.ParentID.Eq(req.ParentID),
	).First(); err == nil {
		return "", errmsg.ErrElementFolderNameRepeat
	}

	elementID := uuid.GetUUID()
	element := packer.TransSaveFolderReqToElementModel(req, elementID, userID)

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

func FolderUpdate(ctx *gin.Context, userID string, req *rao.ElementSaveFolderReq) (string, error) {
	// 目录名称不能重复
	e := query.Use(dal.DB()).Element
	if _, err := e.WithContext(ctx).Where(e.TeamID.Eq(req.TeamID), e.Name.Eq(req.Name),
		e.ElementType.Eq(consts.ElementTypeFolder), e.ElementID.Neq(req.ElementID),
		e.Source.Eq(req.Source), e.ParentID.Eq(req.ParentID),
	).First(); err == nil {
		return "", errmsg.ErrElementFolderNameRepeat
	}

	element := packer.TransSaveFolderReqToElementModel(req, req.ElementID, userID)
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

func FolderList(ctx *gin.Context, teamID string) ([]*rao.ElementFolder, error) {
	e := query.Use(dal.DB()).Element
	list, err := e.WithContext(ctx).Where(
		e.ElementType.Eq(consts.ElementTypeFolder),
		e.TeamID.Eq(teamID),
	).Find()
	if err != nil {
		return nil, err
	}

	elementList := make([]*rao.ElementFolder, 0, len(list))
	for _, e := range list {
		element := &rao.ElementFolder{
			ElementID:   e.ElementID,
			ElementType: e.ElementType,
			TeamID:      e.TeamID,
			ParentID:    e.ParentID,
			Name:        e.Name,
			Sort:        e.Sort,
			Version:     e.Version,
			Description: e.Description,
			Source:      e.Source,
		}
		elementList = append(elementList, element)
	}

	return elementList, err
}

func FolderRemove(ctx *gin.Context, userID string, teamID string, elementIDs []string) error {
	// 递归查询元素
	var recursiveQuery func(string) error
	recursiveQuery = func(parentID string) error {
		e := dal.GetQuery().Element
		list, err := e.WithContext(ctx).Select(e.ElementID, e.ParentID).Where(e.ParentID.Eq(parentID)).Find()
		if err != nil {
			return err
		}
		if len(list) == 0 {
			return nil
		}
		for _, e := range list {
			elementIDs = append(elementIDs, e.ElementID)
			if e.ParentID != "0" {
				if err := recursiveQuery(e.ElementID); err != nil {
					return err
				}
			}
		}
		return nil
	}

	// 查询目录下的元素
	for _, elementID := range elementIDs {
		if err := recursiveQuery(elementID); err != nil {
			return err
		}
	}

	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
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

func SortFolder(ctx *gin.Context, userID string, req *rao.ElementSortFolderReq) error {
	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		for _, e := range req.Elements {
			_, err := tx.Element.WithContext(ctx).Where(
				tx.Element.TeamID.Eq(e.TeamID),
				tx.Element.ElementID.Neq(e.ElementID),
				tx.Element.ElementType.Eq(consts.ElementTypeFolder),
				tx.Element.Name.Eq(e.Name),
				tx.Element.ParentID.Eq(e.ParentID),
			).First()
			if err == nil {
				return errmsg.ErrTargetSortNameAlreadyExist
			}

			_, err = tx.Element.WithContext(ctx).Where(tx.Element.TeamID.Eq(e.TeamID),
				tx.Element.ElementID.Eq(e.ElementID)).UpdateSimple(
				tx.Element.Sort.Value(e.Sort),
				tx.Element.ParentID.Value(e.ParentID))
			if err != nil {
				return err
			}
		}
		return nil
	})

	return err
}
