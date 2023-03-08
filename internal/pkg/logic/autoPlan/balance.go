package autoPlan

// DispatchMachineBalance 压力机调度
type DispatchMachineBalance struct {
	rss []*WeightNode
}

type WeightNode struct {
	addr string // 服务器addr(IP+PORT)
}

func (r *DispatchMachineBalance) AddMachine(addr string) error {
	if addr != "" {
		r.rss = append(r.rss, &WeightNode{
			addr: addr,
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
