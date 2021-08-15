package main

import (
	"container/list"
)

type (
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
		size int
		c    Creator
		l    *list.List
	}
)

func NewPool(size int, c Creator) *BasicPool {
	p := &BasicPool{size, c, list.New()}
	for i := 0; i < size; i++ {
		p.l.PushBack(c())
	}
	return p
}

func (p *BasicPool) Item() PoolItem {
	if p.l.Len() == 0 {
		p.size++
		c := p.c()
		c.Init()
		// fmt.Printf("%p\n", c)
		return c
	}
	ret := p.l.Remove(p.l.Front()).(PoolItem)
	ret.Init()
	// fmt.Printf("%p\n", ret)
	return ret
}

func (p *BasicPool) Return(i PoolItem) {
	i.Reset()
	p.l.PushBack(i)
}
