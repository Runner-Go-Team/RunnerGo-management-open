package uiScene

import "github.com/Runner-Go-Team/RunnerGo-management-open/api/ui"

// ByParentSort Define a custom type to implement sort.Interface
type ByParentSort []*ui.Operator

func (ops ByParentSort) Len() int {
	return len(ops)
}

func (ops ByParentSort) Swap(i, j int) {
	ops[i], ops[j] = ops[j], ops[i]
}

func (ops ByParentSort) Less(i, j int) bool {
	// Compare by parent id first, then by sort
	if ops[i].ParentId == ops[j].ParentId {
		return ops[i].Sort < ops[j].Sort
	}
	return ops[i].ParentId < ops[j].ParentId
}

// Define a helper function to recursively add child slices to each operation
func addChild(ops []*ui.Operator, parentMap map[string][]*ui.Operator) {
	for i, op := range ops {
		if child, ok := parentMap[op.OperatorId]; ok {
			// If the operation has child operations in the map, assign them to its child slice
			ops[i].Operators = child
			// Recursively add child slices to the child operations
			addChild(child, parentMap)
		}
	}
}
