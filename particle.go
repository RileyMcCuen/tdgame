package main

import (
	"container/list"

	"github.com/fogleman/gg"
)

type (
	Processor interface {
		Process(tick int)
		Done() bool
		Finalize()
	}
	Locator interface {
		Location() Location
		SetLocation(Location)
	}
	ListItem interface {
		Elem() *list.Element
		SetElem(*list.Element)
	}
	Damageable interface {
		Health() int
		Damage(int)
		Heal(int)
		Destroyed() bool
	}
	Particle interface {
		Processor
		Locator
		PoolItem
		ListItem
		Draw(*gg.Context)
	}
	Enemy interface {
		Damageable
		Locator
		Active() bool
		LocationAt(tick int) Location
	}
	HealthBar struct {
		max, health int
	}
	BasicEnemy struct {
		Asset
		*PrecalculatedAnimator
		*HealthBar
		e         *list.Element
		startL, l Location
		active    bool
		die       func(l Location)
	}
	ParticlePool struct {
		*BasicPool
	}
	ParticleList struct {
		*list.List
	}
)

func NewParticlePool(size int, c func() Particle) *ParticlePool {
	return &ParticlePool{NewPool(size, func() PoolItem { return c() })}
}

func (p *ParticlePool) Item() Particle {
	return p.BasicPool.Item().(Particle)
}

func (p *ParticlePool) Return(par Particle) {
	p.BasicPool.Return(par)
}

func NewParticleList() *ParticleList {
	return &ParticleList{list.New()}
}

func (pl *ParticleList) Push(p Particle) {
	p.SetElem(pl.List.PushBack(p))
}

func (pl *ParticleList) Peek() Particle {
	return pl.List.Front().Value.(Particle)
}

func (pl *ParticleList) Pop() Particle {
	return pl.Remove(pl.List.Front())
}

func (pl *ParticleList) Remove(e *list.Element) Particle {
	return pl.List.Remove(e).(Particle)
}

func (pl *ParticleList) For(f func(idx int, p Particle) bool) {
	for idx, cur, next := 0, pl.List.Front(), pl.List.Front(); cur != nil; idx, cur = idx+1, next {
		next = next.Next()
		if f(idx, cur.Value.(Particle)) {
			return
		}
	}
}

func (pl *ParticleList) ForReverse(f func(idx int, p Particle) bool) {
	for idx, cur, prev := 0, pl.List.Back(), pl.List.Back(); cur != nil; idx, cur = idx+1, prev {
		prev = prev.Prev()
		if f(idx, cur.Value.(Particle)) {
			return
		}
	}
}

func NewHealthBar(health int) *HealthBar {
	return &HealthBar{health, health}
}

func (hb *HealthBar) Health() int {
	return hb.health
}

func (hb *HealthBar) Damage(amount int) {
	hb.health -= amount
	hb.health = MaxInt(0, hb.health)
}

func (hb *HealthBar) Destroyed() bool {
	return hb.health == 0
}

func (hb *HealthBar) Reset() {
	hb.health = hb.max
}

func (hb *HealthBar) Heal(amount int) {
	hb.health += amount
	hb.health = MinInt(hb.health, hb.max)
}

func NewEnemy(a Asset, anim *PrecalculatedAnimator, health int, l Location, die func(l Location)) Particle {
	return &BasicEnemy{a, anim, NewHealthBar(health), nil, l, l, false, die}
}

func (e *BasicEnemy) Finalize() {
	e.die(e.l)
}

func (e *BasicEnemy) Active() bool {
	return e.active
}

func (e *BasicEnemy) Init() {
	e.active = true
}

func (e *BasicEnemy) Reset() {
	e.active = false
	e.Asset.Reset()
	e.PrecalculatedAnimator.Reset()
	e.HealthBar.Reset()
	e.e = nil
	e.l = e.startL
}

func (e *BasicEnemy) Process(tick int) {
	e.PrecalculatedAnimator.Animate(e)
}

func (e *BasicEnemy) Draw(con *gg.Context) {
	e.Asset.Draw(con, e.l)
}

func (e *BasicEnemy) Location() Location {
	return e.l
}

func (e *BasicEnemy) SetLocation(l Location) {
	e.l = l
}

func (e *BasicEnemy) Elem() *list.Element {
	return e.e
}

func (e *BasicEnemy) SetElem(el *list.Element) {
	e.e = el
}

func (e *BasicEnemy) Done() bool {
	return e.PrecalculatedAnimator.Done()
}

func (e *BasicEnemy) LocationAt(tick int) Location {
	return e.PrecalculatedAnimator.Location(e.PrecalculatedAnimator.t.cur + tick)
}
