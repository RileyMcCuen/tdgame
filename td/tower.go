package td

import (
	"fmt"
	"image/color"
	"tdgame/animator"
	"tdgame/asset"
	"tdgame/core"
	"tdgame/graph"

	"github.com/fogleman/gg"
)

type (
	TowerAttributes struct {
		ProjectileAttributes `yaml:"projectile"`
		Asset                core.Kind
		core.Range           // min, max ticks for projectile to reach enemy; min*speed, max*speed pixels donut radii
		Delay                int
		Cost                 int
	}
	TowerSpec struct {
		core.Meta
		TowerAttributes `yaml:"attributes"`
	}
	TowerAtlas struct {
		tows   map[core.Kind]Tower
		assets asset.AssetAtlas
		anims  animator.AnimatorAtlas
		graphs graph.GraphAtlas
	}
	Tower interface {
		Spec() *TowerSpec
		core.Processor
		core.Locator
		core.Drawer
		Spawn(enemies *ParticleList) Particle
		CopyAt(loc core.Location, ta *TowerAtlas) Tower
	}
	ShootingTower struct {
		*TowerSpec
		*core.LocationWrapper
		nodes []*graph.Node
		// TODO: ProjectilePool
		enemyLoc *core.Location
		sprite   *asset.Sprite
		t        *core.Ticker
		proj     Projectile
	}
)

const (
	TowerType       = "tower"
	ShootingVariety = "shooting"
)

func (p ProjectileAttributes) Kind() core.Kind {
	return p.Asset
}

func (ts *TowerSpec) String() string {
	return core.StructToYaml(ts)
}

var _ core.DeclarationHandler = (*TowerAtlas)(nil)

func NewTowerAtlas() *TowerAtlas {
	return &TowerAtlas{
		tows: make(map[core.Kind]Tower),
	}
}

func (ta *TowerAtlas) Tower(l core.Location, k core.Kind) Tower {
	return ta.tows[k].CopyAt(l, ta)
}

func (ta *TowerAtlas) AddTower(t Tower) {
	ta.tows[t.Spec().Name] = t
}

func (ta *TowerAtlas) Type() core.Kind {
	return TowerType
}

func (ta *TowerAtlas) Match(pm *core.PreMeta) (spec core.Kinder, priority int) {
	switch pm.Variety {
	case ShootingVariety:
		return &TowerSpec{}, 5
	default:
		panic("variety of tower does not exist")
	}
}

func (ta *TowerAtlas) PreLoad(d *core.Declarations) {
	ta.assets = d.Get(asset.AssetType).(asset.AssetAtlas)
	ta.anims = d.Get(animator.AnimatorType).(animator.AnimatorAtlas)
	ta.graphs = d.Get(graph.GraphType).(graph.GraphAtlas)
}

func (ta *TowerAtlas) Load(spec core.Kinder, d *core.Declarations) {
	fmt.Println(ta.graphs)
	graph := ta.graphs.Graph("map").(graph.CachedImageGraph)
	switch ts := spec.(type) {
	case *TowerSpec:
		ta.tows[ts.Name] = TowerFromSpec(ts, ta.assets, ta.anims, graph)
	default:
		panic("variety of tower does not exist")
	}
}

var _ core.DeclarationHandler = (*TowerAtlas)(nil)
var _ Tower = (*ShootingTower)(nil)

func TowerFromSpec(ts *TowerSpec, assets asset.AssetAtlas, anims animator.AnimatorAtlas, g graph.CachedImageGraph) Tower {
	switch ts.Variety {
	case "shooting":
		projAsset := assets.Asset(ts.ProjectileAttributes.Asset)
		proj := NewBullet(
			&ts.ProjectileAttributes,
			projAsset,
			g.TLoc(projAsset.Offset(), projAsset.Size()),
			nil,
			asset.NewSpriteEffect(
				core.ZeroLoc,
				assets.Sprite(ts.ProjectileAttributes.Effect),
			),
		)
		return &ShootingTower{
			ts,
			core.LocWrapper(core.ZeroLoc),
			nil,
			nil,
			assets.Sprite(ts.Asset),
			core.NewTicker(ts.Delay),
			proj,
		}
	default:
		panic("variety of tower does not exist")
	}
}

func (t *ShootingTower) Process(ticks int) {
	if !t.Done() {
		if t.enemyLoc != nil {
			t.sprite.Process(ticks)
			if t.t.Ticks() == t.sprite.Length()-1 {
				t.enemyLoc = nil
			}
		}
		t.t.Tick()
	}
}

func (t *ShootingTower) Done() bool { return false }

func (t *ShootingTower) calculateTrajectory(e Enemy) Projectile {
	// fmt.Println("\tstart")
	// tPoint := t.Location().Point
	// // if enemy is within range
	// if e.Location().DistanceSquared(tPoint) >= util.Square(t.Max*t.Speed) {
	// 	return nil
	// }
	// // fmt.Println("\tclose enough")
	// // find tick to hit the enemy
	// for i := t.Min; i <= t.Max; i++ { // i += t.ExplosionRadius {
	// 	// fmt.Println("\t\t", i)
	// 	eLoc, ok := e.LocationAt(i)
	// 	if !ok {
	// 		return nil
	// 	}
	// 	if eLoc.Near(tPoint, util.Square(i*t.Speed)) {
	// 		// fmt.Println(t.proj, t.Location())
	// 		proj := t.proj.CopyAt(t.Location())
	// 		proj.UpdateTarget(animator.AnimatorFromLine(tPoint, eLoc.Point, i), e)
	// 		return proj
	// 	}
	// }
	// fmt.Println("\tfinished for")
	return nil
}

func (t *ShootingTower) Spawn(pl *ParticleList) Particle {
	var ret Projectile = nil
	if t.t.Done() {
		t.t.Reset()
	} else {
		return ret
	}
	// TODO: new projectile spawn logic
	return ret
}

func (t *ShootingTower) Draw(con *gg.Context) {
	// draw circle radius of test tower
	if core.Grid {
		centered := t.Location().Add(core.Pt(32, 32))
		con.SetColor(color.Black)
		con.DrawCircle(float64(centered.X()), float64(centered.Y()), float64(t.Max*t.Speed))
		con.Stroke()
		con.SetRGBA(.9, .9, .9, 0.2)
		con.DrawCircle(float64(centered.X()), float64(centered.Y()), float64(t.Max*t.Speed))
		con.Fill()
	}
	t.sprite.Draw(con, t.Location())
}

func (t *ShootingTower) CopyAt(l core.Location, ta *TowerAtlas) Tower {
	ter := core.NewTicker(t.Delay)
	ter.TickBy(t.Delay)
	return &ShootingTower{
		t.TowerSpec,
		core.LocWrapper(l),
		ta.graphs.Graph("map").(graph.CachedImageGraph).TilesAround(l.Point, t.Range),
		nil,
		t.sprite.Copy().(*asset.Sprite),
		ter,
		t.proj.CopyAt(l, ta.graphs.Graph("map")),
	}
}

func (t *ShootingTower) Spec() *TowerSpec {
	return t.TowerSpec
}
