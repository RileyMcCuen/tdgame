package game

import (
	"fmt"
	"tdgame/animator"
	"tdgame/asset"
	"tdgame/core"
	"tdgame/graph"
	"tdgame/td"

	"github.com/fogleman/gg"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type (
	Game struct {
		core.Layers
		// graph.CachedImageGraph
		// Particles   *td.ParticleList
		// Projectiles *td.ProjectileList
		// Effects     *asset.EffectList
		// Towers      []td.Tower
		// r           *td.Round
		// *ui.UI
		*core.Declarations
		t *core.Ticker
	}
)

func (g *Game) HandleInput() {
	// if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
	// 	if g.r == nil {
	// 		g.r = g.NewRound(g.UI.Round(), (g.UI.Round()+1)*10)
	// 	}
	// }
	// if inpututil.IsKeyJustPressed(ebiten.KeyG) {
	// 	core.Grid = !core.Grid
	// }
	// if inpututil.IsKeyJustPressed(ebiten.KeyF) {
	// 	ebiten.SetFullscreen(!ebiten.IsFullscreen())
	// }
}

func (g *Game) PostUpdate() error {
	// if g.r != nil && g.r.Done() && g.Particles.Len() == 0 {
	// 	g.r = nil
	// 	g.UI.IncrementRound()
	// }
	// g.t.Tick()
	// if g.UI.Health() == 0 {
	// 	panic("game over")
	// }
	return nil
}

func (g *Game) Update() error {
	// check for user input
	g.HandleInput()
	// process everything
	g.Layers.Process(g.t.Ticks(), nil)
	return g.PostUpdate()
}

// func (g *Game) UpdateOld() error {
// 	// check for user input
// 	g.HandleInput()

// 	// process everything
// 	ticks := g.t.Ticks()
// 	// fmt.Println(ticks)
// 	if g.r != nil && !g.r.Done() {
// 		if newEnemy := g.r.Spawn(); newEnemy != nil {
// 			e := newEnemy.(td.Particle)
// 			e.Init()
// 			g.Particles.Push(e)
// 		}
// 		g.r.Process(ticks, con)
// 	}
// 	// fmt.Println("round")
// 	for _, t := range g.Towers {
// 		t.Process(ticks, con)
// 		if proj := t.Spawn(g.Particles); proj != nil {
// 			g.Projectiles.Push(proj.(td.Projectile))
// 		}
// 	}
// 	// fmt.Println("towers")
// 	g.Projectiles.For(func(_ int, p td.Projectile) bool {
// 		p.Process(ticks, con)
// 		if p.Done() {
// 			g.Effects.Push(g.Projectiles.Remove(p.Elem()).Finalize())
// 		}
// 		return false
// 	})
// 	// fmt.Println("projectiles")
// 	g.Particles.For(func(_ int, p td.Particle) bool {
// 		p.Process(ticks, con)
// 		e := p.(td.Enemy)
// 		if e.Destroyed() {
// 			g.UI.AddScore(e.Spec().Points)
// 			// enemy was destroyed, delete and give points to player
// 			g.Effects.Push(p.Finalize())
// 			g.Particles.Remove(p.Elem()).Reset()
// 		}
// 		if p.Done() {
// 			g.UI.AddHealth(-1)
// 			// enemy made it to the end of the path, delete and do damage to player
// 			g.Effects.Push(p.Finalize())
// 			g.Particles.Remove(p.Elem()).Reset()
// 		}
// 		return false
// 	})
// 	// fmt.Println("particles")
// 	g.Effects.For(func(_ int, p asset.Effect) bool {
// 		p.Process(ticks, con)
// 		if p.Done() {
// 			g.Effects.Remove(p.Elem())
// 		}
// 		return false
// 	})
// 	g.Process(ticks, con)
// 	// fmt.Println("effects")
// 	// perform any post update checks
// 	return g.PostUpdate()
// }

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *Game) Draw(screen *ebiten.Image) {
	con := gg.NewContext(screen.Size())
	g.Layers.Draw(con)
	// create ebiten image and copy it to screen
	eimg := ebiten.NewImageFromImage(con.Image())
	screen.DrawImage(eimg, &ebiten.DrawImageOptions{})
	ebitenutil.DebugPrint(screen, fmt.Sprint(ebiten.CurrentTPS()))
	// // draw the background before drawing everything else
	// // g.CachedImageGraph.Draw(screen)
	// g.CachedImageGraph.Draw(con)
	// // draw all of the processables
	// g.Particles.ForReverse(func(_ int, p td.Particle) bool {
	// 	p.Draw(con)
	// 	return false
	// })
	// g.Projectiles.For(func(_ int, p td.Projectile) bool {
	// 	p.Draw(con)
	// 	return false
	// })
	// for _, t := range g.Towers {
	// 	t.Draw(con)
	// }
	// g.Effects.For(func(_ int, p asset.Effect) bool {
	// 	p.Draw(con)
	// 	return false
	// })
	// // draw ui
	// g.UI.Draw(con)

	// // create ebiten image and copy it to screen
	// eimg := ebiten.NewImageFromImage(con.Image())
	// screen.DrawImage(eimg, &ebiten.DrawImageOptions{})
	// ebitenutil.DebugPrint(screen, fmt.Sprint(ebiten.CurrentTPS()))
}

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
// If you don't have to adjust the screen size with the outside size, just return a fixed size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	// w, h := g.Size()
	// return w + g.UI.Width(), h
	// TODO: make layout size attribute
	return 100, 00
}

func (g *Game) NewRound(round, points int) *td.Round {
	es := make([]td.Enemy, points)
	for i := range es {
		if i%2 == 0 {
			es[i] = g.Declarations.Get("enemy").(td.EnemyAtlas).Enemy(core.ZeroLoc, "slug")
			// es[i] = g.Declarations.Get("enemy").(td.EnemyAtlas).Enemy(g.StartLoc(), "slug")
		} else {
			es[i] = g.Declarations.Get("enemy").(td.EnemyAtlas).Enemy(core.ZeroLoc, "spider")
			// es[i] = g.Declarations.Get("enemy").(td.EnemyAtlas).Enemy(g.StartLoc(), "spider")
		}
	}
	return &td.Round{core.MinGameObject, 0, 96, round, points, core.NewTicker(0), es}
}

func NewGame(declarations string) *Game {
	decs := core.NewDeclarations()
	decs.RegisterHandlers(
		asset.NewAssetAtlas(),
		graph.NewGraphAtlas(),
		animator.DefaultAnimatorAtlas,
		td.NewTowerAtlas(),
		td.NewEnemyAtlas(),
	).AddDir(
		declarations,
	).Load()
	g := &Game{
		core.NewLayers(int(core.NumberOfLayers)),
		// decs.Get(graph.GraphType).(graph.GraphAtlas).Graph("map").(graph.CachedImageGraph),
		// td.NewParticleList(),
		// td.NewProjectileList(),
		// asset.NewEffectList(),
		// []td.Tower{
		// 	decs.Get("tower").(*td.TowerAtlas).Tower(core.Loc(core.Pt(128, 0), 0), "cannon"),
		// 	decs.Get("tower").(*td.TowerAtlas).Tower(core.Loc(core.Pt(256, 128), 0), "cannon"),
		// 	decs.Get("tower").(*td.TowerAtlas).Tower(core.Loc(core.Pt(384, 256), 0), "cannon"),
		// },
		decs,
		// nil,
		// ui.NewUI(),
		core.NewTicker(-1), // nearly infinite ticker, ticks until int overflow happens
	}
	return g
}
