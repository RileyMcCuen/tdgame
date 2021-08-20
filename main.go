package main

import (
	"tdgame/core"
	"tdgame/game"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	// MaxTPS = 32 * 512
	MaxTPS = 32
	Title  = "Tower Defense"
)

func configureEbiten() {
	ebiten.SetMaxTPS(MaxTPS)
	ebiten.SetWindowTitle(Title)
}

func main() {
	configureEbiten()
	g := game.NewGame("./0_gamedata/declarations")
	// f, err := os.Create("poolprofile")
	// util.Check(err)
	// pprof.StartCPUProfile(f)
	// defer pprof.StopCPUProfile()
	core.Check(ebiten.RunGame(g))
}
