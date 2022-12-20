package scene

type RenderItem struct {
	component RenderComponent
	entity    Entity
}

type RenderQueue []*RenderItem

func (rq RenderQueue) Len() int {
	return len(rq)
}

func (rq RenderQueue) Less(i, j int) bool {
	return rq[i].component.LayerID() > rq[j].component.LayerID()
}

func (rq RenderQueue) Swap(i, j int) {
	rq[i], rq[j] = rq[j], rq[i]
}

func (rq *RenderQueue) Push(item any) {
	ri := item.(*RenderItem)
	*rq = append(*rq, ri)
}

func (rq *RenderQueue) Pop() any {
	old := *rq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	*rq = old[0 : n-1]
	return item
}
