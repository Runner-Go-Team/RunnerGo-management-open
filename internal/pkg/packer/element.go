package packer

import (
	"encoding/json"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/log"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
)

func TransSaveFolderReqToElementModel(folder *rao.ElementSaveFolderReq, elementID string, userID string) *model.Element {
	marshal, _ := json.Marshal("")
	locators := string(marshal)
	return &model.Element{
		ElementID:     elementID,
		ElementType:   consts.ElementTypeFolder,
		TeamID:        folder.TeamID,
		Name:          folder.Name,
		ParentID:      folder.ParentID,
		Locators:      locators,
		Sort:          folder.Sort,
		Version:       folder.Version,
		CreatedUserID: userID,
		Description:   folder.Description,
		Source:        folder.Source,
	}
}

func TransSaveElementReqToElementModel(element *rao.ElementSaveReq, elementID string, userID string) *model.Element {
	marshal, _ := json.Marshal(element.Locators)
	locators := string(marshal)
	return &model.Element{
		ElementID:     elementID,
		ElementType:   consts.ElementTypeDefault,
		TeamID:        element.TeamID,
		Name:          element.Name,
		ParentID:      element.ParentID,
		Locators:      locators,
		Sort:          element.Sort,
		Version:       element.Version,
		CreatedUserID: userID,
		Description:   element.Description,
		Source:        element.Source,
	}
}

func TransModelElementToRaoElement(e *model.Element) *rao.Element {
	locators := make([]*rao.Locator, 0)
	if err := json.Unmarshal([]byte(e.Locators), &locators); err != nil {
		log.Logger.Error("Element TransModelElementToRaoElement err:", err)
	}
	element := &rao.Element{
		ElementID:      e.ElementID,
		ElementType:    e.ElementType,
		TeamID:         e.TeamID,
		ParentID:       e.ParentID,
		Name:           e.Name,
		Locators:       locators,
		Sort:           e.Sort,
		Version:        e.Version,
		Description:    e.Description,
		Source:         e.Source,
		CreatedTimeSec: e.CreatedAt.Unix(),
		UpdatedTimeSec: e.UpdatedAt.Unix(),
	}

	return element
}

func TransModelElementToRaoElementFolders(
	list []*model.Element,
	parentList []*model.Element,
	sceneElementList []*model.UISceneElement,
	sceneList []*model.UIScene,
) []*rao.Element {
	elements := make([]*rao.Element, 0, len(list))

	parentListMemo := make(map[string]*model.Element)
	for _, p := range parentList {
		parentListMemo[p.ElementID] = p
	}

	sceneElementIDsMemo := make(map[string][]string)
	for _, e := range sceneElementList {
		sceneElementIDsMemo[e.ElementID] = append(sceneElementIDsMemo[e.ElementID], e.SceneID)
	}

	uiSceneMemo := make(map[string]*model.UIScene)
	for _, s := range sceneList {
		uiSceneMemo[s.SceneID] = s
	}

	for _, e := range list {
		element := TransModelElementToRaoElement(e)
		if e.ParentID == "0" {
			element.ParentName = "根目录"
		} else {
			if p, ok := parentListMemo[e.ParentID]; ok {
				element.ParentName = p.Name
			}
		}
		// 使用 map 进行去重
		seen := make(map[string]bool)
		uiScenes := make([]*rao.UIScene, 0)
		if s, ok := sceneElementIDsMemo[e.ElementID]; ok {
			for _, sceneID := range s {
				if s, ok := uiSceneMemo[sceneID]; ok {
					uiScene := &rao.UIScene{
						SceneID: s.SceneID,
						TeamID:  s.TeamID,
						Name:    s.Name,
					}
					if !seen[sceneID] {
						seen[sceneID] = true
						uiScenes = append(uiScenes, uiScene)
					}
				}
			}
		}
		element.RelateScenes = uiScenes
		elements = append(elements, element)
	}

	return elements
}
