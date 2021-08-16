package main

import (
	"container/list"

	"github.com/fogleman/gg"
)

type (
	Effect interface {
		Processor
		ListItem
		PoolItem
		Locator
		Drawer
	}
	EffectList struct {
		*list.List
	}
	SpriteEffect struct {
		l             Location
		s             *Sprite
		el            *list.Element
		started, done bool
	}
)

func NewSpriteEffect(l Location, s *Sprite) *SpriteEffect {
	return &SpriteEffect{l, s, nil, false, false}
}

func (s *SpriteEffect) Location() Location {
	return s.l
}

func (s *SpriteEffect) SetLocation(l Location) {
	s.l = l
}

func (s *SpriteEffect) Process(ticks int) {
	s.s.Process(ticks)
	s.done = s.started && s.s.cur == 0 && s.s.t.cur == 0
}

func (s *SpriteEffect) Draw(con *gg.Context) {
	s.started = true
	s.s.Draw(con, s.l)
}

func (s *SpriteEffect) Done() bool {
	return s.done
}

func (s *SpriteEffect) Active() bool {
	return !s.done
}

func (s *SpriteEffect) Init() {
	s.started = false
	s.done = false
}

func (s *SpriteEffect) Reset() {
	s.s.Reset()
}

func (s *SpriteEffect) SetElem(e *list.Element) {
	s.el = e
}

func (s *SpriteEffect) Elem() *list.Element {
	return s.el
}

func NewEffectList() *EffectList {
	return &EffectList{list.New()}
}

func (pl *EffectList) Push(p Effect) {
	if p != nil {
		p.SetElem(pl.List.PushFront(p))
	}
}

func (pl *EffectList) Peek() Effect {
	return pl.List.Back().Value.(Effect)
}

func (pl *EffectList) Pop() Effect {
	return pl.Remove(pl.List.Back())
}

func (pl *EffectList) Remove(e *list.Element) Effect {
	return pl.List.Remove(e).(Effect)
}

func (pl *EffectList) For(f func(idx int, p Effect) bool) {
	for idx, cur := 0, pl.List.Front(); cur != nil; idx, cur = idx+1, cur.Next() {
		if f(idx, cur.Value.(Effect)) {
			return
		}
	}
}
