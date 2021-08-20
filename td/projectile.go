package td

import (
	"container/list"
	"tdgame/animator"
	"tdgame/asset"
	"tdgame/core"
	"tdgame/graph"

	"github.com/fogleman/gg"
)

type (
	ProjectileAttributes struct {
		Asset           core.Kind
		Effect          core.Kind
		PoolSize        int `yaml:"poolSize"`
		Speed           int
		Damage          int
		ExplosionRadius int `yaml:"explosionRadius"`
	}
	Projectile interface {
		Particle
		Spec() *ProjectileAttributes
		CopyAt(l core.Location, g graph.Graph) Projectile
		UpdateTarget(anim *animator.PrecalculatedAnimator)
	}
	Bullet struct {
		*ProjectileAttributes
		*graph.TileLocation
		asset  asset.Asset
		anim   *animator.PrecalculatedAnimator
		el     *list.Element
		active bool
		effect *asset.SpriteEffect
	}
	ProjectileList struct {
		*list.List
	}
)

func NewProjectileList() *ProjectileList {
	return &ProjectileList{list.New()}
}

func (pl *ProjectileList) Push(p Projectile) {
	if p != nil {
		p.SetElem(pl.List.PushFront(p))
	}
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

func NewBullet(spec *ProjectileAttributes, a asset.Asset, tl *graph.TileLocation, anim *animator.PrecalculatedAnimator, effect *asset.SpriteEffect) Projectile {
	ret := &Bullet{
		spec,
		tl,
		a,
		anim,
		nil,
		false,
		effect,
	}
	return ret
}

func (b *Bullet) Spec() *ProjectileAttributes {
	return b.ProjectileAttributes
}

func (b *Bullet) Speed() int {
	return 1
}

func (b *Bullet) Process(ticks int) {
	if !b.Done() {
		b.asset.Process(ticks)
		b.anim.Animate(b)
	}
}

func (b *Bullet) Active() bool {
	return b.active
}

func (b *Bullet) Init() {
	b.active = true
	b.effect.Reset()
}

func (b *Bullet) Reset() {
	b.active = false
	b.asset.Reset()
	b.LocationWrapper.SetLocation(core.Loc(core.Pt(-100, -100), 0)) // put off screen for the moment
	b.anim = nil
	b.el = nil
}

func (b *Bullet) Draw(con *gg.Context) {
	b.asset.Draw(con, b.Location())
}

func (b *Bullet) Near(o core.Point) bool {
	return b.Location().Near(o, core.Square(b.ExplosionRadius))
}

func (b *Bullet) Elem() *list.Element {
	return b.el
}

func (b *Bullet) SetElem(e *list.Element) {
	b.el = e
}

func (b *Bullet) Done() bool {
	return b.anim.Done()
}

func (b *Bullet) Finalize() asset.Effect {
	b.effect.SetLocation(core.Loc(b.Location().Add(core.Pt(24, 24)), 0))
	return b.effect
}

func (b *Bullet) CopyAt(l core.Location, g graph.Graph) Projectile {
	return NewBullet(b.ProjectileAttributes, b.asset.Copy(), b.TileLocation.Copy(), nil, b.effect.CopyAt(core.ZeroLoc).(*asset.SpriteEffect))
}

func (b *Bullet) UpdateTarget(anim *animator.PrecalculatedAnimator) {
	b.anim = anim
}

func (b *Bullet) Destination() core.Location {
	l, _ := b.anim.LastLocation()
	return l
}

func (b *Bullet) Effect() asset.Effect {
	return b.effect
}
