package main

import (
	"fmt"
	"image/color"
	"os"
	"runtime/pprof"
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
		asset.AssetAtlas
		core.AnimatorAtlas
		Particles   *td.ParticleList
		Projectiles *td.ProjectileList
		Effects     *asset.EffectList
		Towers      []td.Tower
		*td.Declarations
		r *td.Round
		*ui.UI
		t *core.Ticker
	}
)

const (
	MaxTPS = 32
)

func (g *Game) HandleInput() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		// fmt.Println("pressed", g.r)
		if g.r == nil {
			g.r = g.NewRound(g.UI.Round(), (g.UI.Round()+1)*10)
		}
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
		if p.(td.Enemy).Destroyed() {
			g.UI.AddScore(1)
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
	g.CachedImageGraph.Draw(screen)
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

	// draw circle radius of test tower
	con.SetColor(color.Black)
	con.DrawCircle(128, 0, 192)
	con.Stroke()
	con.SetRGBA(0, .95, 0, 0.2)
	con.DrawCircle(128, 0, 192)
	con.Fill()

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
		es[i] = g.Declarations.EnemyAtlas.Enemy(g.StartLoc(), "slug")
	}
	return &td.Round{0, 128, round, points, core.NewTicker(0), es}
}

func configureEbiten() {
	ebiten.SetMaxTPS(MaxTPS)
	ebiten.SetWindowTitle("Tower Defense")
}

func NewGame(assets, declarations, gameMap string) *Game {
	asAtlas := asset.MakeAssetAtlas(assets)
	graph := graph.GraphFromFile(gameMap, asAtlas)
	anAtlas := core.DefaultAnimatorAtlas
	anAtlas.CreatePathAnimator(graph.StartLoc(), graph.Path())
	decs := td.NewDeclarations(declarations, asAtlas, anAtlas)
	g := &Game{
		graph,
		asAtlas,
		core.DefaultAnimatorAtlas,
		td.NewParticleList(),
		td.NewProjectileList(),
		asset.NewEffectList(),
		[]td.Tower{
			decs.Tower(core.Loc(core.Pt(128, 0), 0), "cannon"),
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
	g := NewGame(
		"./0_gamedata/assets",
		"./0_gamedata/declarations",
		"./0_gamedata/maps/8x8.txt",
	)
	f, err := os.Create("poolprofile")
	util.Check(err)
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()
	util.Check(ebiten.RunGame(g))
}
