package packer

import (
	"github.com/go-omnibus/proof"
	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson"

	"kp-management/internal/pkg/dal/mao"
	"kp-management/internal/pkg/dal/model"
	"kp-management/internal/pkg/dal/rao"
)

func TransSaveFlowReqToMaoFlow(req *rao.SaveFlowReq) *mao.Flow {

	nodes, err := bson.Marshal(mao.Node{Nodes: req.Nodes})
	if err != nil {
		proof.Errorf("flow.nodes bson marshal err %w", err)
	}

	edges, err := bson.Marshal(mao.Edge{Edges: req.Edges})
	if err != nil {
		proof.Errorf("flow.edges bson marshal err %w", err)
	}

	return &mao.Flow{
		SceneID: req.SceneID,
		TeamID:  req.TeamID,
		Version: req.Version,
		Nodes:   nodes,
		Edges:   edges,
		//MultiLevelNodes: req.MultiLevelNodes,
	}
}

func TransMaoFlowToRaoSceneFlow(t *model.Target, f *mao.Flow, vis []*model.VariableImport, sceneVariables, variables []*model.Variable) *rao.SceneFlow {
	var n mao.Node
	if err := bson.Unmarshal(f.Nodes, &n); err != nil {
		proof.Errorf("flow.nodes bson unmarshal err %w", err)
	}

	var fileList []rao.FileList
	for _, vi := range vis {
		fileList = append(fileList, rao.FileList{
			IsChecked: int64(vi.Status),
			Path:      vi.URL,
		})
	}

	var v []rao.KV
	for _, variable := range sceneVariables {
		v = append(v, rao.KV{
			Key:   variable.Var,
			Value: variable.Val,
		})
	}

	var globalVariables []*rao.KVVariable
	for _, variable := range variables {
		globalVariables = append(globalVariables, &rao.KVVariable{
			Key:   variable.Var,
			Value: variable.Val,
		})
	}

	return &rao.SceneFlow{
		SceneID:   t.TargetID,
		SceneName: t.Name,
		TeamID:    t.TeamID,
		Nodes:     n.Nodes,
		Configuration: rao.Configuration{
			ParameterizedFile: rao.ParameterizedFile{
				Paths: fileList,
			},
			Variable: v,
		},
		Variable: globalVariables,
	}
}

func TransMaoFlowToRaoGetFowResp(f *mao.Flow) *rao.GetFlowResp {

	var n mao.Node
	if err := bson.Unmarshal(f.Nodes, &n); err != nil {
		proof.Errorf("flow.nodes bson unmarshal err %w", err)
	}

	var e mao.Edge
	if err := bson.Unmarshal(f.Edges, &e); err != nil {
		proof.Errorf("flow.edges bson unmarshal err %w", err)
	}

	return &rao.GetFlowResp{
		SceneID: f.SceneID,
		TeamID:  f.TeamID,
		Version: f.Version,
		Nodes:   n.Nodes,
		Edges:   e.Edges,
		//MultiLevelNodes: f.MultiLevelNodes,
	}
}

func TransMaoFlowsToRaoFlows(flows []*mao.Flow) []*rao.Flow {
	ret := make([]*rao.Flow, 0)
	for _, f := range flows {
		var n mao.Node
		if err := bson.Unmarshal(f.Nodes, &n); err != nil {
			proof.Errorf("flow.nodes bson unmarshal err %w", err)
		}

		var e mao.Edge
		if err := bson.Unmarshal(f.Edges, &e); err != nil {
			proof.Errorf("flow.edges bson unmarshal err %w", err)
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
	var node mao.Node
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
