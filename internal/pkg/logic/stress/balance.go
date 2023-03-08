package stress

// DispatchMachineBalance 压力机调度
type DispatchMachineBalance struct {
	rss []WeightNode
}

type WeightNode struct {
	addr             string // 服务器addr(IP+PORT)
	usableGoroutines int64
}

func (r *DispatchMachineBalance) AddMachine(addr string, usableGoroutines int64) error {
	if addr != "" {
		r.rss = append(r.rss, WeightNode{
			addr:             addr,
			usableGoroutines: usableGoroutines,
		})
	}
	return nil
}

func (r *DispatchMachineBalance) GetMachine(curIndex int) string {
	if curIndex < 0 || curIndex >= len(r.rss) {
		return ""
	}
	return r.rss[curIndex].addr
}

func (r *DispatchMachineBalance) GetAllFreeMachine(curIndex int) []WeightNode {
	freeMachineList := make([]WeightNode, 0, len(r.rss))
	for i := 0; i < len(r.rss); i++ {
		freeMachineList = append(freeMachineList, r.rss[curIndex])
		curIndex++
		if curIndex == len(r.rss) {
			curIndex = 0
		}
	}

	return freeMachineList
}
