package td

import (
	"image/color"
	"tdgame/asset"
	"tdgame/core"
	"tdgame/util"

	"github.com/fogleman/gg"
)

type (
	Tower interface {
		Spec() *TowerSpec
		core.Processor
		core.Locator
		asset.Drawer
		Spawn(enemies *ParticleList) Particle
		CopyAt(loc core.Location) Tower
	}
	ShootingTower struct {
		*TowerSpec
		*core.LocationWrapper
		// TODO: ProjectilePool
		enemyLoc *core.Location
		sprite   *asset.Sprite
		t        *core.Ticker
		proj     Projectile
	}
)

var _ Tower = (*ShootingTower)(nil)

func TowerFromSpec(ts *TowerSpec, assets asset.AssetAtlas, anims core.AnimatorAtlas) Tower {
	switch ts.Variety {
	case "shooting":
		proj := NewBullet(
			&ts.ProjectileAttributes,
			assets.Asset(ts.ProjectileAttributes.Asset),
			core.ZeroLoc,
			nil,
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
			if t.t.Ticks() == t.sprite.Length() {
				t.enemyLoc = nil
			}
		}
		t.t.Tick()
	}
}

func (t *ShootingTower) Done() bool { return false }

func (t *ShootingTower) calculateTrajectory(e Enemy) Projectile {
	// fmt.Println("\tstart")
	tPoint := t.Location().Point
	// if enemy is within range
	if e.Location().DistanceSquared(tPoint) >= util.Square(t.Max*t.Speed) {
		return nil
	}
	// fmt.Println("\tclose enough")
	// find tick to hit the enemy
	for i := t.Min; i <= t.Max; i += t.ExplosionRadius {
		// fmt.Println("\t\t", i)
		eLoc, ok := e.LocationAt(i)
		if !ok {
			return nil
		}
		if eLoc.Near(tPoint, util.Square(i*t.Speed)) {
			// fmt.Println(t.proj, t.Location())
			proj := t.proj.CopyAt(t.Location())
			proj.UpdateTarget(core.AnimatorFromLine(tPoint, eLoc.Point, i), e)
			return proj
		}
	}
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
	pl.For(func(idx int, p Particle) bool {
		// fmt.Println(idx)
		enemy := p.(Enemy)
		np := t.calculateTrajectory(enemy)
		// fmt.Println("traj")
		if np != nil {
			ret = np
			dest := np.Destination()
			t.enemyLoc = &dest
			t.LocationWrapper.SetLocation(t.Location().RotateByATan2(dest))
			return true
		}
		// fmt.Println(idx)
		return false
	})
	return ret
}

func (t *ShootingTower) Draw(con *gg.Context) {
	// draw circle radius of test tower
	con.SetColor(color.Black)
	con.DrawCircle(float64(t.Location().X()), float64(t.Location().Y()), float64(t.Max*t.Speed))
	con.Stroke()
	con.SetRGBA(0, .95, 0, 0.2)
	con.DrawCircle(float64(t.Location().X()), float64(t.Location().Y()), float64(t.Max*t.Speed))
	con.Fill()
	t.sprite.Draw(con, t.Location())
}

func (t *ShootingTower) CopyAt(l core.Location) Tower {
	return &ShootingTower{
		t.TowerSpec,
		core.LocWrapper(l),
		nil,
		t.sprite.Copy().(*asset.Sprite),
		core.NewTicker(t.Delay),
		t.proj.CopyAt(l),
	}
}

func (t *ShootingTower) Spec() *TowerSpec {
	return t.TowerSpec
}
