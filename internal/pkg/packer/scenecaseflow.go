package packer

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/log"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/mao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"go.mongodb.org/mongo-driver/bson"
)

func TransSaveSceneCaseFlowReqToMaoFlow(req *rao.SaveSceneCaseFlowReq) *mao.SceneCaseFlow {

	nodes, err := bson.Marshal(mao.Node{Nodes: req.Nodes})
	if err != nil {
		log.Logger.Errorf("flow.nodes bson marshal err %w", err)
	}

	edges, err := bson.Marshal(mao.Edge{Edges: req.Edges})
	if err != nil {
		log.Logger.Errorf("flow.edges bson marshal err %w", err)
	}

	return &mao.SceneCaseFlow{
		SceneID:     req.SceneID,
		SceneCaseID: req.SceneCaseID,
		TeamID:      req.TeamID,
		Version:     req.Version,
		Nodes:       nodes,
		Edges:       edges,
		//MultiLevelNodes: req.MultiLevelNodes,
	}
}

func TransSaveCaseAssembleToTargetModel(caseAssemble *rao.SaveCaseAssembleReq, userID string) *model.Target {
	return &model.Target{
		TargetID:      caseAssemble.CaseID,
		TeamID:        caseAssemble.TeamID,
		TargetType:    consts.TargetTypeTestCase,
		Name:          caseAssemble.Name,
		ParentID:      caseAssemble.SceneID,
		Method:        caseAssemble.Method,
		Sort:          caseAssemble.Sort,
		TypeSort:      caseAssemble.TypeSort,
		Status:        consts.TargetStatusNormal,
		Version:       caseAssemble.Version,
		CreatedUserID: userID,
		RecentUserID:  userID,
		Source:        caseAssemble.Source,
		PlanID:        caseAssemble.PlanID,
		Description:   caseAssemble.Description,
	}
}

func TransMaoSceneCaseFlowToRaoGetFowResp(f *mao.SceneCaseFlow) *rao.GetSceneCaseFlowResp {

	var n mao.Node
	if err := bson.Unmarshal(f.Nodes, &n); err != nil {
		log.Logger.Errorf("flow.nodes bson unmarshal err %w", err)
	}

	var e mao.Edge
	if err := bson.Unmarshal(f.Edges, &e); err != nil {
		log.Logger.Errorf("flow.edges bson unmarshal err %w", err)
	}

	return &rao.GetSceneCaseFlowResp{
		SceneID:     f.SceneID,
		SceneCaseID: f.SceneCaseID,
		TeamID:      f.TeamID,
		Version:     f.Version,
		Nodes:       n.Nodes,
		Edges:       e.Edges,
		//MultiLevelNodes: f.MultiLevelNodes,
	}
}

func TransMaoFlowToRaoSceneCaseFlow(t *model.Target, f *mao.Flow, vis []*model.VariableImport, sceneVariable rao.GlobalVariable, globalVariable rao.GlobalVariable) *rao.SceneCaseFlow {
	nodes := mao.Node{}
	if err := bson.Unmarshal(f.Nodes, &nodes); err != nil {
		log.Logger.Errorf("flow.nodes bson unmarshal err %w", err)
	}

	edges := mao.Edge{}
	if err := bson.Unmarshal(f.Edges, &edges); err != nil {
		log.Logger.Errorf("flow.edges bson unmarshal err %w", err)
	}

	fileList := make([]rao.FileList, 0, len(vis))
	for _, vi := range vis {
		fileList = append(fileList, rao.FileList{
			IsChecked: int64(vi.Status),
			Path:      vi.URL,
		})
	}

	//var v []rao.KV
	//for _, variable := range sceneVariables {
	//	v = append(v, rao.KV{
	//		Key:   variable.Var,
	//		Value: variable.Val,
	//	})
	//}
	//
	//var globalVariables []*rao.KVVariable
	//for _, variable := range variables {
	//	globalVariables = append(globalVariables, &rao.KVVariable{
	//		Key:   variable.Var,
	//		Value: variable.Val,
	//	})
	//}

	nodesRound := GetNodesByLevel(nodes.Nodes, edges.Edges)

	return &rao.SceneCaseFlow{
		SceneID:       t.ParentID,
		SceneCaseID:   t.TargetID,
		SceneCaseName: t.Name,
		TeamID:        t.TeamID,
		//Nodes:         nodes.Nodes,
		Configuration: rao.Configuration{
			ParameterizedFile: rao.ParameterizedFile{
				Paths: fileList,
			},
			SceneVariable: sceneVariable,
			//Variable: v,
		},
		//Variable:       globalVariables,
		NodesRound:     nodesRound,
		GlobalVariable: globalVariable,
	}
}

func TransMaoFlowToMaoSceneCaseFlow(flow *mao.Flow, sceneID string) *mao.SceneCaseFlow {

	if flow.Nodes != nil {
		var n mao.Node
		if err := bson.Unmarshal(flow.Nodes, &n); err != nil {
			log.Logger.Errorf("flow.nodes bson unmarshal err %w", err)
		}

		var e mao.Edge
		if err := bson.Unmarshal(flow.Edges, &e); err != nil {
			log.Logger.Errorf("flow.edges bson unmarshal err %w", err)
		}

		for nodeInfoK := range n.Nodes {
			//newUUID := GetUUID()
			//n.Nodes[nodeInfoK].ID = newUUID
			n.Nodes[nodeInfoK].Data.From = "case"
			//n.Nodes[nodeInfoK].Data.ID = newUUID
		}
		flow.Nodes, _ = bson.Marshal(n)
	}

	sceneCaseFlow := mao.SceneCaseFlow{
		SceneID:     flow.SceneID,
		SceneCaseID: sceneID,
		TeamID:      flow.TeamID,
		Version:     flow.Version,
		Nodes:       flow.Nodes,
		Edges:       flow.Edges,
		//MultiLevelNodes: req.MultiLevelNodes,
	}

	err := ChangeCaseNodeUUID(&sceneCaseFlow)
	if err != nil {
		log.Logger.Errorf("sceneCaseFlow change UUID err %w", err)
	}
	return &sceneCaseFlow
}
