package main

import (
	"math"

	"github.com/fogleman/gg"
)

type (
	Tower interface {
		Processor
		Locator
		Drawer
		Spawn(enemies *ParticleList) Particle
	}
	ShootingTower struct {
		fireRate, minRange, maxRange, projectileSpeed int
		AssetAtlas
		// ParticlePool
		enemyLoc *Location
		Asset
		*LocationWrapper
		*Ticker
	}
)

var _ Tower = (*ShootingTower)(nil)

func NewTower(k Kind, l Location, aa AssetAtlas) *ShootingTower {
	return &ShootingTower{2, 32, 96, 2, aa, nil, aa.Asset(k), &LocationWrapper{l}, NewTicker(32)}
}

func (t *ShootingTower) Process(ticks int) {
	if !t.Done() {
		if t.enemyLoc != nil {
			t.Asset.Process(ticks)
			if t.Ticks() > 15 {
				t.Asset.Reset()
				t.enemyLoc = nil
			}
		}
		t.Tick()
	}
}

func (t *ShootingTower) Done() bool { return false }

func (t *ShootingTower) Spawn(pl *ParticleList) Particle {
	var ret Projectile = nil
	if t.Ticker.Done() {
		t.Ticker.Reset()
	} else {
		return ret
	}
	pl.For(func(_ int, p Particle) bool {
		enemy := p.(Enemy)
		np := NewProjectile(t.l, t.AssetAtlas.Asset("bullet"), enemy, t.AssetAtlas.Sprite("explosion"))
		if np != nil {
			ret = np
			dest := np.Destination()
			t.enemyLoc = &dest
			dx, dy := float64(t.enemyLoc.x-t.l.x), float64(t.l.y-t.enemyLoc.y)
			t.l.rot = int(gg.Degrees(math.Atan2(dx, dy)))
			return true
		}
		return false
	})
	return ret
}

func (t *ShootingTower) Draw(con *gg.Context) {
	t.Asset.Draw(con, t.l)
}
