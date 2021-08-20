package asset

import (
	"container/list"
	"tdgame/core"

	"github.com/fogleman/gg"
)

type (
	Effect interface {
		core.Processor
		core.ListItem
		core.PoolItem
		core.Locator
		core.Drawer
		Length() int
		CopyAt(l core.Location) Effect
	}
	EffectList struct {
		*list.List
	}
	SpriteEffect struct {
		*core.LocationWrapper
		s             *Sprite
		el            *list.Element
		started, done bool
	}
)

func NewSpriteEffect(l core.Location, s *Sprite) *SpriteEffect {
	return &SpriteEffect{core.LocWrapper(l), s, nil, false, false}
}

func (s *SpriteEffect) Length() int {
	return s.s.total * s.s.delay
}

func (s *SpriteEffect) Process(ticks int, con core.Context) bool {
	s.s.Process(ticks, con)
	s.done = s.started && s.s.cur == 0 && s.s.t.Ticks() == 0
	return false
}

func (s *SpriteEffect) Draw(con *gg.Context) {
	s.started = true
	s.s.Draw(con, s.Location())
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

func (s *SpriteEffect) CopyAt(l core.Location) Effect {
	return NewSpriteEffect(l, s.s.Copy().(*Sprite))
}

func (s *SpriteEffect) CopySprite() *Sprite {
	return s.s.Copy().(*Sprite)
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
