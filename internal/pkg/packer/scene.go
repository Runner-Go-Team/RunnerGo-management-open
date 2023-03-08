package packer

import (
	"kp-management/internal/pkg/dal/mao"
	"kp-management/internal/pkg/dal/model"
	"kp-management/internal/pkg/dal/rao"
)

func TransSaveSceneReqToMaoScene(scene *rao.SaveSceneReq) *mao.Scene {
	//request, err := bson.Marshal(scene.Request)
	//if err != nil {
	//	fmt.Println(fmt.Errorf("scene.request json marshal err %w", err))
	//}
	//
	//script, err := bson.Marshal(scene.Script)
	//if err != nil {
	//	fmt.Println(fmt.Errorf("scene.script json marshal err %w", err))
	//}

	return &mao.Scene{
		TargetID: scene.TargetID,
		//Request:  request,
		//Script:   script,
	}
}

func TransTargetToRaoScene(targets []*model.Target) []*rao.Scene {
	ret := make([]*rao.Scene, 0)
	for _, t := range targets {
		ret = append(ret, &rao.Scene{
			TeamID:      t.TeamID,
			TargetID:    t.TargetID,
			ParentID:    t.ParentID,
			Name:        t.Name,
			Method:      t.Method,
			Sort:        t.Sort,
			TypeSort:    t.TypeSort,
			Version:     t.Version,
			Description: t.Description,
		})
	}
	return ret
}
