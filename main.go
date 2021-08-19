package main

import (
	"fmt"
	"tdgame/animator"
	"tdgame/asset"
	"tdgame/core"
	"tdgame/graph"
	"tdgame/td"
	"tdgame/ui"
	"tdgame/util"

	"github.com/fogleman/gg"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type (
	Game struct {
		graph.CachedImageGraph
		Particles   *td.ParticleList
		Projectiles *td.ProjectileList
		Effects     *asset.EffectList
		Towers      []td.Tower
		*core.Declarations
		r *td.Round
		*ui.UI
		t *core.Ticker
	}
)

const (
	// MaxTPS = 32 * 512
	MaxTPS = 32
)

func (g *Game) HandleInput() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		if g.r == nil {
			g.r = g.NewRound(g.UI.Round(), (g.UI.Round()+1)*10)
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyG) {
		core.Grid = !core.Grid
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}
}

func (g *Game) PostUpdate() error {
	if g.r != nil && g.r.Done() && g.Particles.Len() == 0 {
		g.r = nil
		g.UI.IncrementRound()
	}
	g.t.Tick()
	if g.UI.Health() == 0 {
		panic("game over")
	}
	return nil
}

func (g *Game) Update() error {
	// check for user input
	g.HandleInput()

	// process everything
	ticks := g.t.Ticks()
	// fmt.Println(ticks)
	if g.r != nil && !g.r.Done() {
		if newEnemy := g.r.Spawn(); newEnemy != nil {
			e := newEnemy.(td.Particle)
			e.Init()
			g.Particles.Push(e)
		}
		g.r.Process(ticks)
	}
	// fmt.Println("round")
	for _, t := range g.Towers {
		t.Process(ticks)
		if proj := t.Spawn(g.Particles); proj != nil {
			g.Projectiles.Push(proj.(td.Projectile))
		}
	}
	// fmt.Println("towers")
	g.Projectiles.For(func(_ int, p td.Projectile) bool {
		p.Process(ticks)
		if p.Done() {
			g.Effects.Push(g.Projectiles.Remove(p.Elem()).Finalize())
		}
		return false
	})
	// fmt.Println("projectiles")
	g.Particles.For(func(_ int, p td.Particle) bool {
		p.Process(ticks)
		e := p.(td.Enemy)
		if e.Destroyed() {
			g.UI.AddScore(e.Spec().Points)
			// enemy was destroyed, delete and give points to player
			g.Effects.Push(p.Finalize())
			g.Particles.Remove(p.Elem()).Reset()
		}
		if p.Done() {
			g.UI.AddHealth(-1)
			// enemy made it to the end of the path, delete and do damage to player
			g.Effects.Push(p.Finalize())
			g.Particles.Remove(p.Elem()).Reset()
		}
		return false
	})
	// fmt.Println("particles")
	g.Effects.For(func(_ int, p asset.Effect) bool {
		p.Process(ticks)
		if p.Done() {
			g.Effects.Remove(p.Elem())
		}
		return false
	})
	// fmt.Println("effects")
	// perform any post update checks
	return g.PostUpdate()
}

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *Game) Draw(screen *ebiten.Image) {
	con := gg.NewContext(screen.Size())
	// draw the background before drawing everything else
	// g.CachedImageGraph.Draw(screen)
	g.CachedImageGraph.Draw(con)
	// draw all of the processables
	g.Particles.ForReverse(func(_ int, p td.Particle) bool {
		p.Draw(con)
		return false
	})
	g.Projectiles.For(func(_ int, p td.Projectile) bool {
		p.Draw(con)
		return false
	})
	for _, t := range g.Towers {
		t.Draw(con)
	}
	g.Effects.For(func(_ int, p asset.Effect) bool {
		p.Draw(con)
		return false
	})
	// draw ui
	g.UI.Draw(con)

	// create ebiten image and copy it to screen
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

func (g *Game) NewRound(round, points int) *td.Round {
	es := make([]td.Enemy, points)
	for i := range es {
		if i%2 == 0 {
			es[i] = g.Declarations.Get("enemy").(td.EnemyAtlas).Enemy(g.StartLoc(), "slug")
		} else {
			es[i] = g.Declarations.Get("enemy").(td.EnemyAtlas).Enemy(g.StartLoc(), "spider")
		}
	}
	return &td.Round{0, 96, round, points, core.NewTicker(0), es}
}

func configureEbiten() {
	ebiten.SetMaxTPS(MaxTPS)
	ebiten.SetWindowTitle("Tower Defense")
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
		decs.Get(graph.GraphType).(graph.GraphAtlas).Graph("map").(graph.CachedImageGraph),
		td.NewParticleList(),
		td.NewProjectileList(),
		asset.NewEffectList(),
		[]td.Tower{
			decs.Get("tower").(td.TowerAtlas).Tower(core.Loc(core.Pt(128, 0), 0), "cannon"),
			decs.Get("tower").(td.TowerAtlas).Tower(core.Loc(core.Pt(256, 128), 0), "cannon"),
			decs.Get("tower").(td.TowerAtlas).Tower(core.Loc(core.Pt(384, 256), 0), "cannon"),
		},
		decs,
		nil,
		ui.NewUI(),
		core.NewTicker(-1), // nearly infinite ticker, ticks until int overflow happens
	}
	return g
}

func main() {
	configureEbiten()
	g := NewGame("./0_gamedata/declarations")
	// f, err := os.Create("poolprofile")
	// util.Check(err)
	// pprof.StartCPUProfile(f)
	// defer pprof.StopCPUProfile()
	util.Check(ebiten.RunGame(g))
}
