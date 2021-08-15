package main

import (
	"container/list"

	"github.com/fogleman/gg"
)

type (
	Projectile interface {
		Particle
	}
	Bullet struct {
		Asset
		*LocationWrapper
		Animator
		explode func(l Location)
		e       Enemy
		el      *list.Element
		active  bool
	}
	ProjectileList struct {
		*list.List
	}
)

func NewProjectileList() *ProjectileList {
	return &ProjectileList{list.New()}
}

func (pl *ProjectileList) Push(p Projectile) {
	p.SetElem(pl.List.PushFront(p))
}

func (pl *ProjectileList) Peek() Projectile {
	return pl.List.Back().Value.(Projectile)
}

func (pl *ProjectileList) Pop() Projectile {
	return pl.Remove(pl.List.Back())
}

func (pl *ProjectileList) Remove(e *list.Element) Projectile {
	return pl.List.Remove(e).(Projectile)
}

func (pl *ProjectileList) For(f func(idx int, p Projectile) bool) {
	for idx, cur := 0, pl.List.Front(); cur != nil; idx, cur = idx+1, cur.Next() {
		if f(idx, cur.Value.(Projectile)) {
			return
		}
	}
}

func NewProjectile(start Location, a Asset, e Enemy, explode func(l Location)) Projectile {
	speed := 2
	max := 96
	// if enemy is within range
	if e.Location().DistanceSquared(start.Point) >= Square(max*speed) {
		return nil
	}
	// find tick to hit the enemy
	i := 16
	for i = 16; i < max; i += 4 {
		if e.LocationAt(i).Near(start.Point, Square(i*speed)) {
			return &Bullet{
				a,
				&LocationWrapper{start},
				AnimatorFromLine(start.Point, e.LocationAt(i).Point, i),
				explode,
				e,
				nil,
				true,
			}
		}
	}
	// if no tick found, return nil
	return nil
}

func (b *Bullet) Process(tick int) {
	if !b.Done() {
		b.Animate(b)
	}
}

func (b *Bullet) Active() bool {
	return b.active
}

func (b *Bullet) Init() {
	b.active = true
}

func (b *Bullet) Reset() {
	b.active = false
	b.Animator.Reset()
	b.Asset.Reset()
	b.LocationWrapper.SetLocation(Loc(Pt(-100, -100), 0)) // put off screen for the moment
	b.e = nil
	b.el = nil
}

func (b *Bullet) Draw(con *gg.Context) {
	b.Asset.Draw(con, b.Location())
}

func (b *Bullet) Near(o Point) bool {
	return b.l.Near(o, 6*6)
}

func (b *Bullet) Elem() *list.Element {
	return b.el
}

func (b *Bullet) SetElem(e *list.Element) {
	b.el = e
}

func (b *Bullet) Done() bool {
	return b.Animator.Done()
}

func (b *Bullet) Finalize() {
	if b.e.Active() {
		// TODO: variable damage
		b.e.Damage(3)
	}
	b.explode(Loc(b.l.Add(Pt(24, 24)), 0))
}
