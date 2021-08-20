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
	EnemyAttributes struct {
		Asset     core.Kind
		Animation core.Kind
		Effect    core.Kind
		Health    int
		Speed     int
		Points    int
	}
	EnemySpec struct {
		core.Meta
		EnemyAttributes `yaml:"attributes"`
	}
	EnemyAtlas map[core.Kind]Enemy
	Damageable interface {
		Health() int
		Damage(int)
		Heal(int)
		Destroyed() bool
	}
	Particle interface {
		core.Processor
		core.Locator
		core.PoolItem
		core.ListItem
		core.Drawer
		Finalize() asset.Effect
	}
	Enemy interface {
		CopyAt(l core.Location) Enemy
		Spec() *EnemySpec
		Damageable
		core.Locator
		Active() bool
		LocationAt(tick int) (core.Location, bool)
	}
	HealthBar struct {
		max, health int
	}
	BasicEnemy struct {
		*EnemySpec
		*graph.TileLocation
		*HealthBar
		anim   *animator.PrecalculatedAnimator
		sprite *asset.Sprite
		effect *asset.SpriteEffect
		e      *list.Element
		active bool
	}
	ParticlePool struct {
		*core.BasicPool
	}
	ParticleList struct {
		*list.List
	}
)

const (
	EnemyType    = "enemy"
	BasicVariety = "basic"
)

func (es *EnemySpec) String() string {
	return core.StructToYaml(es)
}

var _ core.DeclarationHandler = EnemyAtlas{}

func NewEnemyAtlas() EnemyAtlas {
	return EnemyAtlas{}
}

func (ea EnemyAtlas) Enemy(l core.Location, k core.Kind) Enemy {
	return ea[k].CopyAt(l)
}

func (ea EnemyAtlas) AddEnemy(e Enemy) {
	ea[e.Spec().Name] = e
}

func (ea EnemyAtlas) Type() core.Kind {
	return EnemyType
}

func (ea EnemyAtlas) Match(pm *core.PreMeta) (core.Kinder, int) {
	switch pm.Variety {
	case BasicVariety:
		return &EnemySpec{}, 5
	default:
		panic("variety of enemy does not exist")
	}
}

func (ea EnemyAtlas) PreLoad(d *core.Declarations) {

}

func (ea EnemyAtlas) Load(spec core.Kinder, d *core.Declarations) {
	assets := d.Get("asset").(asset.AssetAtlas)
	anims := d.Get("animator").(animator.AnimatorAtlas)
	g := d.Get("graph").(graph.GraphAtlas).Graph("map").(graph.CachedImageGraph)
	switch es := spec.(type) {
	case *EnemySpec:
		ea[es.Name] = EnemyFromSpec(es, assets, anims, g)
	default:
		panic("variety of enemy does not exist")
	}
}

func NewParticlePool(size int, c func() Particle) *ParticlePool {
	return &ParticlePool{core.NewPool(size, func() core.PoolItem { return c() })}
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
	if p != nil {
		p.SetElem(pl.List.PushBack(p))
	}
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
	hb.health = core.MaxInt(0, hb.health)
}

func (hb *HealthBar) Destroyed() bool {
	return hb.health == 0
}

func (hb *HealthBar) Reset() {
	hb.health = hb.max
}

func (hb *HealthBar) Heal(amount int) {
	hb.health += amount
	hb.health = core.MinInt(hb.health, hb.max)
}

func (hb *HealthBar) Copy() *HealthBar {
	return NewHealthBar(hb.health)
}

func EnemyFromSpec(es *EnemySpec, assets asset.AssetAtlas, anims animator.AnimatorAtlas, g graph.CachedImageGraph) Enemy {
	switch es.Variety {
	case "basic":
		sp := assets.Sprite(es.Asset)
		return &BasicEnemy{
			es,
			g.TLoc(sp.Offset(), sp.Size()),
			NewHealthBar(es.Health),
			anims.PrecalculatedAnimator(es.Animation),
			sp,
			asset.NewSpriteEffect(core.ZeroLoc, assets.Sprite(es.Effect)),
			nil,
			false,
		}
	default:
		panic("variety of enemy does not exist")
	}
}

func (e *BasicEnemy) Finalize() asset.Effect {
	if e.Destroyed() {
		e.effect.SetLocation(e.Location())
		return e.effect
	}
	return nil
}

func (e *BasicEnemy) Spec() *EnemySpec {
	return e.EnemySpec
}

func (e *BasicEnemy) Active() bool {
	return e.active
}

func (e *BasicEnemy) Init() {
	e.active = true
	e.effect.Reset()
}

func (e *BasicEnemy) Reset() {
	e.active = false
	e.sprite.Reset()
	e.anim.Reset()
	e.HealthBar.Reset()
	e.e = nil
	e.SetLocation(core.ZeroLoc)
}

func (e *BasicEnemy) Process(ticks int) {
	e.sprite.Process(ticks)
	e.anim.Animate(e)
}

func (e *BasicEnemy) Speed() int {
	return e.EnemySpec.Speed
}

func (e *BasicEnemy) Draw(con *gg.Context) {
	e.sprite.Draw(con, e.Location())
}

func (e *BasicEnemy) Elem() *list.Element {
	return e.e
}

func (e *BasicEnemy) SetElem(el *list.Element) {
	e.e = el
}

func (e *BasicEnemy) Done() bool {
	return e.anim.Done()
}

func (e *BasicEnemy) LocationAt(tick int) (core.Location, bool) {
	return e.anim.LocationOffset(tick * e.Speed())
}

func (e *BasicEnemy) Radius() int {
	return e.Size.X() / 2
}

func (e *BasicEnemy) Near(col graph.Collider) bool {
	return e.Location().Near(col.Location().Point, e.Radius()+col.Radius())
}

func (e *BasicEnemy) CopyAt(l core.Location) Enemy {
	ret := &BasicEnemy{
		e.EnemySpec,
		e.TileLocation.Copy(),
		e.HealthBar.Copy(),
		e.anim.Copy().(*animator.PrecalculatedAnimator),
		e.sprite.Copy().(*asset.Sprite),
		e.effect.CopyAt(core.ZeroLoc).(*asset.SpriteEffect),
		nil,
		false,
	}
	ret.TileLocation.Move(l, ret)
	return ret
}
