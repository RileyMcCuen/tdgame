package main

import (
	"fmt"
	"image/color"

	"github.com/fogleman/gg"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type (
	Game struct {
		CachedImageGraph
		AssetAtlas
		AnimatorAtlas
		*ParticlePool
		Particles   *ParticleList
		Projectiles *ProjectileList
		Effects     *EffectList
		Towers      []Tower
		Spawner     Spawner
		*UI
		t *Ticker
	}
)

const (
	MaxTPS = 32
)

func Check(e error) {
	if e != nil {
		panic(e.Error())
	}
}

func CheckNil(i interface{}) {
	if i == nil {
		panic("object is nil and should not be")
	}
}

func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Square(a int) int {
	return a * a
}

func (g *Game) Return(p Particle) {
	g.Effects.Push(p.Finalize())
	g.ParticlePool.Return(g.Particles.Remove(p.Elem()))
}

func (g *Game) PostUpdate() error {
	g.t.Tick()
	if g.UI.Health() == 0 {
		panic("game over")
	}
	return nil
}

func (g *Game) Update() error {
	ticks := g.t.Ticks()
	for _, t := range g.Towers {
		t.Process(ticks)
		if proj := t.Spawn(g.Particles); proj != nil {
			g.Projectiles.Push(proj.(Projectile))
		}
	}
	g.Projectiles.For(func(_ int, p Projectile) bool {
		p.Process(ticks)
		if p.Done() {
			g.Effects.Push(g.Projectiles.Remove(p.Elem()).Finalize())
		}
		return false
	})
	g.Particles.For(func(_ int, p Particle) bool {
		p.Process(ticks)
		if p.(Enemy).Destroyed() {
			g.UI.AddScore(1)
			// enemy was destroyed, delete and give points to player
			g.Effects.Push(p.Finalize())
			g.ParticlePool.Return(g.Particles.Remove(p.Elem()))
		}
		if p.Done() {
			g.UI.AddHealth(-1)
			// enemy made it to the end of the path, delete and do damage to player
			g.Effects.Push(p.Finalize())
			g.ParticlePool.Return(g.Particles.Remove(p.Elem()))
		}
		return false
	})
	g.Effects.For(func(_ int, p Effect) bool {
		p.Process(ticks)
		if p.Done() {
			g.Effects.Remove(p.Elem())
		}
		return false
	})
	// spawner creating enemies
	if ticks%128 == 0 {
		g.Particles.Push(g.ParticlePool.Item())
	}
	// tower shooting projectiles
	// if ticks%16 == 0 {
	// 	g.Particles.For(func(_ int, p Particle) bool {
	// 		np := NewProjectile(Loc(Pt(128, 0), 0), g.Asset("bullet"), p.(Enemy), g.Sprite("explosion"))
	// 		if np != nil {
	// 			g.Projectiles.Push(np)
	// 			return true
	// 		}
	// 		return false
	// 	})
	// }
	return g.PostUpdate()
}

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *Game) Draw(screen *ebiten.Image) {
	// draw the background before drawing everything else
	g.CachedImageGraph.Draw(screen)
	con := gg.NewContextForImage(screen)
	g.Particles.ForReverse(func(_ int, p Particle) bool {
		p.Draw(con)
		return false
	})
	g.Projectiles.For(func(_ int, p Projectile) bool {
		p.Draw(con)
		return false
	})
	for _, t := range g.Towers {
		t.Draw(con)
	}
	g.Effects.For(func(_ int, p Effect) bool {
		p.Draw(con)
		return false
	})
	g.UI.Draw(con)
	con.SetColor(color.Black)
	con.DrawCircle(128, 0, 192)
	con.Stroke()
	eimg := ebiten.NewImageFromImage(con.Image())
	screen.DrawImage(eimg, &ebiten.DrawImageOptions{})
	ebitenutil.DebugPrint(screen, fmt.Sprint(ebiten.CurrentTPS()))
}

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
// If you don't have to adjust the screen size with the outside size, just return a fixed size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	w, h := g.Size()
	return w + g.UI.Width(), h
}

func main() {
	ebiten.SetMaxTPS(MaxTPS)
	aa := MakeAssetAtlas("./assets")
	graph := GraphFromFile("./maps/8x8.txt", aa)
	g := &Game{
		graph,
		aa,
		DefaultAnimatorAtlas,
		nil, // neet to reference g in order to create the pool
		NewParticleList(),
		NewProjectileList(),
		NewEffectList(),
		[]Tower{
			NewTower("cannon", Loc(Pt(128, 0), 0), aa),
		},
		NewGameSpawner(5, 128, []Enemy{NewEnemy(
			aa.Asset("enemy"),
			DefaultAnimatorAtlas.PrecalculatedAnimator("prepath"),
			10,
			Loc(Pt(graph.start.x*TileSizeInt, (graph.start.y-1)*TileSizeInt), graph.InitialRotation()),
			aa.Sprite("enemydeath"),
		).(Enemy),
			NewEnemy(
				aa.Asset("enemy"),
				DefaultAnimatorAtlas.PrecalculatedAnimator("prepath"),
				10,
				Loc(Pt(graph.start.x*TileSizeInt, (graph.start.y-1)*TileSizeInt), graph.InitialRotation()),
				aa.Sprite("enemydeath"),
			).(Enemy),
			NewEnemy(
				aa.Asset("enemy"),
				DefaultAnimatorAtlas.PrecalculatedAnimator("prepath"),
				10,
				Loc(Pt(graph.start.x*TileSizeInt, (graph.start.y-1)*TileSizeInt), graph.InitialRotation()),
				aa.Sprite("enemydeath"),
			).(Enemy),
			NewEnemy(
				aa.Asset("enemy"),
				DefaultAnimatorAtlas.PrecalculatedAnimator("prepath"),
				10,
				Loc(Pt(graph.start.x*TileSizeInt, (graph.start.y-1)*TileSizeInt), graph.InitialRotation()),
				aa.Sprite("enemydeath"),
			).(Enemy)}),
		NewUI(),
		NewTicker(-1), // nearly infinite ticker, ticks until int overflow happens
	}
	loc := Loc(Pt(g.start.x*TileSizeInt, (g.start.y-1)*TileSizeInt), g.InitialRotation())
	sanim := g.SerialAnimatorFromPath(Kind("path"), g.path)
	g.PutAnimator("prepath", NewPrecalculatedAnimator("prepath", loc, sanim))
	g.ParticlePool = NewParticlePool(100, func() Particle {
		return NewEnemy(
			g.Asset("enemy"),
			g.PrecalculatedAnimator("prepath"),
			10,
			loc,
			g.Sprite("enemydeath"),
		)
	})
	ebiten.SetWindowTitle("Tower Defense")
	// f, err := os.Create("poolprofile")
	// Check(err)
	// pprof.StartCPUProfile(f)
	// defer pprof.StopCPUProfile()
	if err := ebiten.RunGame(g); err != nil {
		panic(err.Error())
	}
}
