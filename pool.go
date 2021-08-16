package main

type (
	List     []PoolItem
	PoolItem interface {
		Active() bool
		Init()
		Reset()
	}
	Creator func() PoolItem
	Pool    interface {
		Item() PoolItem
		Return(i PoolItem)
	}
	BasicPool struct {
		c Creator
		l *List
	}
)

func NewList(size int, f func() PoolItem) *List {
	items := make(List, size)
	for i := 0; i < size; i++ {
		items[i] = f()
	}
	return &items
}

func (l *List) Len() int {
	return len(*l)
}

func (l *List) Empty() bool {
	return len(*l) == 0
}

func (l *List) Item() PoolItem {
	idx := len(*l) - 1
	ret := (*l)[idx]
	*l = (*l)[0:idx]
	return ret
}

func (l *List) Return(i PoolItem) {
	*l = append(*l, i)
}

func (l *List) Double(f func() PoolItem) {
	listCap, listLen := cap(*l), len(*l)
	newList := make(List, listCap*2)
	copy(newList, *l)
	for i := 0; i < listCap; i++ {
		newList[listLen+i] = f()
	}
	*l = newList
}

func (l *List) Shrink() {
	newList := make(List, len(*l))
	copy(newList, *l)
	*l = newList
}

func NewPool(size int, c Creator) *BasicPool {
	p := &BasicPool{c, NewList(size, c)}
	return p
}

func (p *BasicPool) Item() PoolItem {
	if p.l.Empty() {
		p.l.Double(p.c)
	}
	item := p.l.Item()
	ret := item.(PoolItem)
	ret.Init()
	return ret
}

func (p *BasicPool) Return(i PoolItem) {
	i.Reset()
	p.l.Return(i)
}
