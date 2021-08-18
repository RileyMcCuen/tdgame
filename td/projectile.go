package td

import (
	"container/list"
	"tdgame/asset"
	"tdgame/core"
	"tdgame/util"

	"github.com/fogleman/gg"
)

type (
	Projectile interface {
		Particle
		Spec() *ProjectileAttributes
		Destination() core.Location
		CopyAt(l core.Location) Projectile
		UpdateTarget(anim *core.PrecalculatedAnimator, e Enemy)
	}
	Bullet struct {
		*ProjectileAttributes
		*core.LocationWrapper
		asset  asset.Asset
		anim   *core.PrecalculatedAnimator
		e      Enemy
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

func NewBullet(spec *ProjectileAttributes, a asset.Asset, start core.Location, anim *core.PrecalculatedAnimator, e Enemy, effect *asset.SpriteEffect) Projectile {
	return &Bullet{
		spec,
		core.LocWrapper(start),
		a,
		anim,
		e,
		nil,
		false,
		effect,
	}
}

func (b *Bullet) Spec() *ProjectileAttributes {
	return b.ProjectileAttributes
}

func (b *Bullet) Speed() int {
	return b.ProjectileAttributes.Speed
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
	b.e = nil
	b.el = nil
}

func (b *Bullet) Draw(con *gg.Context) {
	b.asset.Draw(con, b.Location())
}

func (b *Bullet) Near(o core.Point) bool {
	return b.Location().Near(o, util.Square(b.ExplosionRadius))
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
	if b.e.Active() {
		b.e.Damage(b.Damage)
	}
	b.effect.SetLocation(core.Loc(b.Location().Add(core.Pt(24, 24)), 0))
	return b.effect
}

func (b *Bullet) CopyAt(l core.Location) Projectile {
	return NewBullet(b.ProjectileAttributes, b.asset.Copy(), b.Location(), nil, nil, b.effect.CopyAt(core.ZeroLoc).(*asset.SpriteEffect))
}

func (b *Bullet) UpdateTarget(anim *core.PrecalculatedAnimator, e Enemy) {
	b.e = e
	b.anim = anim
}

func (b *Bullet) Destination() core.Location {
	l, _ := b.anim.LastLocation()
	return l
}
