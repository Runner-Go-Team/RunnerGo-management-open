package packer

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/log"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/mao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson"
)

func TransSaveFlowReqToMaoFlow(req *rao.SaveFlowReq) *mao.Flow {

	nodes, err := bson.Marshal(mao.Node{Nodes: req.Nodes})
	if err != nil {
		log.Logger.Info("flow.nodes bson marshal err %w", err)
	}

	edges, err := bson.Marshal(mao.Edge{Edges: req.Edges})
	if err != nil {
		log.Logger.Info("flow.edges bson marshal err %w", err)
	}

	prepositions, err := bson.Marshal(mao.Node{Nodes: req.Prepositions})
	if err != nil {
		log.Logger.Info("flow.prepositions bson marshal err %w", err)
	}

	return &mao.Flow{
		SceneID:      req.SceneID,
		TeamID:       req.TeamID,
		Version:      req.Version,
		Nodes:        nodes,
		Edges:        edges,
		Prepositions: prepositions,
	}
}

func TransMaoFlowToRaoSceneFlow(t *model.Target, f *mao.Flow, vis []*model.VariableImport,
	sceneVariable rao.GlobalVariable, globalVariable rao.GlobalVariable) *rao.SceneFlow {
	nodes := mao.Node{}
	if err := bson.Unmarshal(f.Nodes, &nodes); err != nil {
		log.Logger.Info("flow.nodes bson unmarshal err %w", err)
	}

	edges := mao.Edge{}
	if err := bson.Unmarshal(f.Edges, &edges); err != nil {
		log.Logger.Info("flow.edges bson unmarshal err %w", err)
	}

	prepositions := mao.Preposition{}
	if err := bson.Unmarshal(f.Prepositions, &prepositions); err != nil {
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

	fileList := make([]rao.FileList, 0, len(vis))
	for _, vi := range vis {
		fileList = append(fileList, rao.FileList{
			IsChecked: int64(vi.Status),
			Path:      vi.URL,
		})
	}

	for k, nodeInfo := range nodes.Nodes {
		if f.EnvID != 0 {
			nodes.Nodes[k].API.Request.PreUrl = nodeInfo.API.EnvInfo.PreUrl
		} else {
			nodes.Nodes[k].API.Request.PreUrl = ""
		}
	}

	nodesRound := GetNodesByLevel(nodes.Nodes, edges.Edges)

	return &rao.SceneFlow{
		SceneID:   t.TargetID,
		SceneName: t.Name,
		TeamID:    t.TeamID,
		Configuration: rao.SceneConfiguration{
			ParameterizedFile: rao.SceneVariablePath{
				Paths: fileList,
			},
			SceneVariable: sceneVariable,
		},
		NodesRound:     nodesRound,
		GlobalVariable: globalVariable,
		Prepositions:   prepositionsArr,
	}
}

func GetNodesByLevel(nodes []rao.Node, edges []rao.Edge) [][]rao.Node {
	arr := make([][]rao.Node, 0, len(nodes))
	for len(nodes) > 0 {
		var currentLayer []rao.Node
		for _, node := range nodes {
			isTarget := false
			for _, edge := range edges {
				if edge.Target == node.ID {
					isTarget = true
					break
				}
			}
			if !isTarget {
				currentLayer = append(currentLayer, node)
			}
		}
		arr = append(arr, currentLayer)
		for _, node := range currentLayer {
			var filteredEdges []rao.Edge
			for _, edge := range edges {
				if edge.Source != node.ID {
					filteredEdges = append(filteredEdges, edge)
				}
			}
			edges = filteredEdges
		}
		var remainingNodes []rao.Node
		for _, node := range nodes {
			if !containsNode(currentLayer, node) {
				remainingNodes = append(remainingNodes, node)
			}
		}
		nodes = remainingNodes
	}
	return arr
}

func containsNode(nodes []rao.Node, node rao.Node) bool {
	for _, n := range nodes {
		if n.ID == node.ID {
			return true
		}
	}
	return false
}

func TransMaoFlowToRaoGetFowResp(f mao.Flow) rao.GetFlowResp {
	n := mao.Node{}
	if err := bson.Unmarshal(f.Nodes, &n); err != nil {
		log.Logger.Info("flow.nodes bson unmarshal err %w", err)
	}
	if n.Nodes == nil || len(n.Nodes) == 0 {
		n.Nodes = make([]rao.Node, 0)
	}

	e := mao.Edge{}
	if err := bson.Unmarshal(f.Edges, &e); err != nil {
		log.Logger.Info("flow.edges bson unmarshal err %w", err)
	}
	if e.Edges == nil || len(e.Edges) == 0 {
		e.Edges = make([]rao.Edge, 0)
	}

	prepositions := mao.Preposition{}
	if err := bson.Unmarshal(f.Prepositions, &prepositions); err != nil {
		log.Logger.Info("flow.prepositions bson unmarshal err %w", err)
	}
	if prepositions.Prepositions == nil || len(prepositions.Prepositions) == 0 {
		prepositions.Prepositions = make([]rao.Node, 0)
	}

	// 把前置条件放到node里面
	for _, v := range prepositions.Prepositions {
		n.Nodes = append(n.Nodes, v)
	}

	for k, v := range n.Nodes {
		if v.API.Request.Method == "" {
			n.Nodes[k].API.Request.Method = v.API.Method
		}
	}

	return rao.GetFlowResp{
		SceneID: f.SceneID,
		TeamID:  f.TeamID,
		Version: f.Version,
		Nodes:   n.Nodes,
		Edges:   e.Edges,
		EnvID:   f.EnvID,
	}
}

func TransMaoFlowsToRaoFlows(flows []*mao.Flow) []*rao.Flow {
	ret := make([]*rao.Flow, 0)
	for _, f := range flows {
		var n mao.Node
		if err := bson.Unmarshal(f.Nodes, &n); err != nil {
			log.Logger.Info("flow.nodes bson unmarshal err %w", err)
		}

		var e mao.Edge
		if err := bson.Unmarshal(f.Edges, &e); err != nil {
			log.Logger.Info("flow.edges bson unmarshal err %w", err)
		}

		ret = append(ret, &rao.Flow{
			SceneID: f.SceneID,
			TeamID:  f.TeamID,
			Version: f.Version,
			Nodes:   n.Nodes,
			Edges:   e.Edges,
			//MultiLevelNodes: nil,
		})
	}
	return ret
}

// ChangeSceneNodeUUID 更换接口的uuid
func ChangeSceneNodeUUID(data *mao.Flow) error {
	// 新老uuid映射关系
	oldAndNewUUIDMap := make(map[string]string)

	// 替换node里面的uuid
	var node mao.Node
	err := bson.Unmarshal(data.Nodes, &node)
	if err != nil {
		return err
	}
	for k, nodeInfo := range node.Nodes {
		if _, ok := oldAndNewUUIDMap[nodeInfo.ID]; !ok {
			oldAndNewUUIDMap[nodeInfo.ID] = uuid.NewV4().String()
		}
		node.Nodes[k].ID = oldAndNewUUIDMap[nodeInfo.ID]
		node.Nodes[k].Data.ID = node.Nodes[k].ID
		for k2, oldPreID := range nodeInfo.PreList {
			if _, ok := oldAndNewUUIDMap[oldPreID]; !ok {
				oldAndNewUUIDMap[oldPreID] = uuid.NewV4().String()
			}
			node.Nodes[k].PreList[k2] = oldAndNewUUIDMap[oldPreID]
		}
		for k3, oldNextID := range nodeInfo.NextList {
			if _, ok := oldAndNewUUIDMap[oldNextID]; !ok {
				oldAndNewUUIDMap[oldNextID] = uuid.NewV4().String()
			}
			node.Nodes[k].NextList[k3] = oldAndNewUUIDMap[oldNextID]
		}
	}
	newNode, err := bson.Marshal(node)
	if err != nil {
		return err
	}
	data.Nodes = newNode

	// 替换prepositions里面的uuid
	prepositions := mao.Preposition{}
	err = bson.Unmarshal(data.Prepositions, &prepositions)
	if err != nil {
		return err
	}
	for k, nodeInfo := range prepositions.Prepositions {
		if _, ok := oldAndNewUUIDMap[nodeInfo.ID]; !ok {
			oldAndNewUUIDMap[nodeInfo.ID] = uuid.NewV4().String()
		}
		prepositions.Prepositions[k].ID = oldAndNewUUIDMap[nodeInfo.ID]
		prepositions.Prepositions[k].Data.ID = prepositions.Prepositions[k].ID
		for k2, oldPreID := range nodeInfo.PreList {
			if _, ok := oldAndNewUUIDMap[oldPreID]; !ok {
				oldAndNewUUIDMap[oldPreID] = uuid.NewV4().String()
			}
			prepositions.Prepositions[k].PreList[k2] = oldAndNewUUIDMap[oldPreID]
		}
		for k3, oldNextID := range nodeInfo.NextList {
			if _, ok := oldAndNewUUIDMap[oldNextID]; !ok {
				oldAndNewUUIDMap[oldNextID] = uuid.NewV4().String()
			}
			prepositions.Prepositions[k].NextList[k3] = oldAndNewUUIDMap[oldNextID]
		}
	}
	newPrepositions, err := bson.Marshal(prepositions)
	if err != nil {
		return err
	}
	data.Prepositions = newPrepositions

	// 替换edges里面的uuid
	var oldEdges mao.Edge
	err = bson.Unmarshal(data.Edges, &oldEdges)
	if err != nil {
		return err
	}
	for k, edgesInfo := range oldEdges.Edges {
		if _, ok := oldAndNewUUIDMap[edgesInfo.Source]; !ok {
			oldAndNewUUIDMap[edgesInfo.Source] = uuid.NewV4().String()
		}
		oldEdges.Edges[k].Source = oldAndNewUUIDMap[edgesInfo.Source]
		if _, ok := oldAndNewUUIDMap[edgesInfo.Target]; !ok {
			oldAndNewUUIDMap[edgesInfo.Target] = uuid.NewV4().String()
		}
		oldEdges.Edges[k].Target = oldAndNewUUIDMap[edgesInfo.Target]
	}

	newEdges, err := bson.Marshal(oldEdges)
	if err != nil {
		return err
	}
	data.Edges = newEdges
	return nil
}

// ChangeCaseNodeUUID 更换接口的uuid
func ChangeCaseNodeUUID(data *mao.SceneCaseFlow) error {
	var node mao.SceneCaseFlowNode
	err := bson.Unmarshal(data.Nodes, &node)
	if err != nil {
		return err
	}
	oldAndNewUUIDMap := make(map[string]string)
	for k, nodeInfo := range node.Nodes {
		if _, ok := oldAndNewUUIDMap[nodeInfo.ID]; !ok {
			oldAndNewUUIDMap[nodeInfo.ID] = uuid.NewV4().String()
		}

		node.Nodes[k].ID = oldAndNewUUIDMap[nodeInfo.ID]
		node.Nodes[k].Data.ID = node.Nodes[k].ID
		for k2, oldPreID := range nodeInfo.PreList {
			if _, ok := oldAndNewUUIDMap[oldPreID]; !ok {
				oldAndNewUUIDMap[oldPreID] = uuid.NewV4().String()
			}
			node.Nodes[k].PreList[k2] = oldAndNewUUIDMap[oldPreID]
		}
		for k3, oldNextID := range nodeInfo.NextList {
			if _, ok := oldAndNewUUIDMap[oldNextID]; !ok {
				oldAndNewUUIDMap[oldNextID] = uuid.NewV4().String()
			}
			node.Nodes[k].NextList[k3] = oldAndNewUUIDMap[oldNextID]
		}
	}
	newNode, err := bson.Marshal(node)
	if err != nil {
		return err
	}
	data.Nodes = newNode

	var oldEdges mao.SceneCaseFlowEdge
	err = bson.Unmarshal(data.Edges, &oldEdges)
	if err != nil {
		return err
	}
	for k, edgesInfo := range oldEdges.Edges {
		if _, ok := oldAndNewUUIDMap[edgesInfo.Source]; !ok {
			oldAndNewUUIDMap[edgesInfo.Source] = uuid.NewV4().String()
		}
		oldEdges.Edges[k].Source = oldAndNewUUIDMap[edgesInfo.Source]
		if _, ok := oldAndNewUUIDMap[edgesInfo.Target]; !ok {
			oldAndNewUUIDMap[edgesInfo.Target] = uuid.NewV4().String()
		}
		oldEdges.Edges[k].Target = oldAndNewUUIDMap[edgesInfo.Target]
	}

	newEdges, err := bson.Marshal(oldEdges)
	if err != nil {
		return err
	}
	data.Edges = newEdges
	return nil
}
