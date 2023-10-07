package packer

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
)

func TransModelVariablesToRaoVariables(vs []*model.Variable) []*rao.Variable {
	ret := make([]*rao.Variable, 0)
	for _, v := range vs {
		ret = append(ret, &rao.Variable{
			VarID:       v.ID,
			TeamID:      v.TeamID,
			Var:         v.Var,
			Val:         v.Val,
			Status:      v.Status,
			Description: v.Description,
		})
	}
	return ret
}

func TransRaoVariablesToModelVariables(teamID string, vs []*rao.Variable) []*model.Variable {
	ret := make([]*model.Variable, 0)
	for _, v := range vs {
		ret = append(ret, &model.Variable{
			TeamID:      teamID,
			Var:         v.Var,
			Val:         v.Val,
			Description: v.Description,
			Type:        consts.VariableTypeGlobal,
		})
	}
	return ret
}

func TransSceneRaoVariablesToModelVariables(teamID string, sceneID string, vs []*rao.Variable) []*model.Variable {
	ret := make([]*model.Variable, 0)
	for _, v := range vs {
		ret = append(ret, &model.Variable{
			TeamID:      teamID,
			SceneID:     sceneID,
			Var:         v.Var,
			Val:         v.Val,
			Description: v.Description,
			Type:        consts.VariableTypeScene,
		})
	}
	return ret
}

func TransImportVariablesToRaoImportVariables(vi []*model.VariableImport) []*rao.Import {
	ret := make([]*rao.Import, 0)
	for _, v := range vi {
		ret = append(ret, &rao.Import{
			ID:             v.ID,
			TeamID:         v.TeamID,
			SceneID:        v.SceneID,
			Name:           v.Name,
			URL:            v.URL,
			Status:         v.Status,
			CreatedTimeSec: v.CreatedAt.Unix(),
		})
	}
	return ret
}
