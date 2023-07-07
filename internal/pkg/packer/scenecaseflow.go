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
	}
}

func TransSaveCaseAssembleToTargetModel(caseAssemble *rao.SaveCaseAssembleReq, userID string) *model.Target {
	return &model.Target{
		TargetID:      caseAssemble.CaseID,
		TeamID:        caseAssemble.TeamID,
		TargetType:    consts.TargetTypeTestCase,
		Name:          caseAssemble.Name,
		ParentID:      caseAssemble.SceneID,
		Sort:          caseAssemble.Sort,
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
		EnvID:       f.EnvID,
	}
}

func TransMaoFlowToRaoSceneCaseFlow(t *model.Target, flow *mao.Flow, caseFlow *mao.Flow,
	vis []*model.VariableImport, sceneVariable rao.GlobalVariable, globalVariable rao.GlobalVariable) *rao.SceneCaseFlow {
	nodes := mao.Node{}
	if err := bson.Unmarshal(caseFlow.Nodes, &nodes); err != nil {
		log.Logger.Errorf("flow.nodes bson unmarshal err %w", err)
	}

	edges := mao.Edge{}
	if err := bson.Unmarshal(caseFlow.Edges, &edges); err != nil {
		log.Logger.Errorf("flow.edges bson unmarshal err %w", err)
	}

	fileList := make([]rao.FileList, 0, len(vis))
	for _, vi := range vis {
		fileList = append(fileList, rao.FileList{
			IsChecked: int64(vi.Status),
			Path:      vi.URL,
		})
	}

	// 前置条件
	prepositions := mao.Preposition{}
	if err := bson.Unmarshal(flow.Prepositions, &prepositions); err != nil {
		log.Logger.Info("flow.prepositions bson unmarshal err %w", err)
	}

	prepositionsArr := make([]rao.Preposition, 0, len(prepositions.Prepositions))
	for _, nodeInfo := range prepositions.Prepositions {
		dbType := "mysql"
		if nodeInfo.API.Method == "ORACLE" {
			dbType = "oracle"
		} else if nodeInfo.API.Method == "PgSQL" {
			dbType = "postgresql"
		}
		nodeInfo.API.SqlDetail.SqlDatabaseInfo.Type = dbType
		temp := rao.Preposition{
			Type:  nodeInfo.Type,
			Event: nodeInfo,
		}
		prepositionsArr = append(prepositionsArr, temp)
	}

	for k, nodeInfo := range nodes.Nodes {
		if caseFlow.EnvID != 0 {
			nodes.Nodes[k].API.Request.PreUrl = nodeInfo.API.EnvInfo.PreUrl
		} else {
			nodes.Nodes[k].API.Request.PreUrl = ""
		}
	}

	nodesRound := GetNodesByLevel(nodes.Nodes, edges.Edges)
	return &rao.SceneCaseFlow{
		SceneID:       t.ParentID,
		SceneCaseID:   t.TargetID,
		SceneCaseName: t.Name,
		TeamID:        t.TeamID,
		Configuration: rao.Configuration{
			ParameterizedFile: rao.ParameterizedFile{
				Paths: fileList,
			},
			SceneVariable: sceneVariable,
		},
		NodesRound:     nodesRound,
		GlobalVariable: globalVariable,
		Prepositions:   prepositionsArr,
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
			n.Nodes[nodeInfoK].Data.From = "case"
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
		EnvID:       flow.EnvID,
	}

	err := ChangeCaseNodeUUID(&sceneCaseFlow)
	if err != nil {
		log.Logger.Errorf("sceneCaseFlow change UUID err %w", err)
	}
	return &sceneCaseFlow
}
