package ui

import (
	"fmt"
	"image/color"
	"tdgame/core"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
)

type (
	UI struct {
		state *UIState
		has   *RoundHealthScoreUI
		c     color.Color
	}
	RoundHealthScoreUI struct {
		*RoundHealthScoreState
		f    font.Face
		size int
		c    color.Color
	}
)

type (
	UIState struct {
		has RoundHealthScoreState
	}
	RoundHealthScoreState struct {
		round, health, score int
	}
)

const (
	large = 24
	width = 150
)

var (
	gray     = color.RGBA{50, 50, 50, 255}
	offWhite = color.RGBA{245, 245, 245, 255}
)

func NewUI() *UI {
	return &UI{
		&UIState{},
		NewHealthAndScoreUI(),
		gray,
	}
}

func (ui *UI) Width() int {
	return width
}

func (ui *UI) Height() int {
	return width
}

func (ui *UI) Draw(con *gg.Context) {
	con.SetColor(ui.c)
	con.DrawRectangle(float64(con.Width()-ui.Width()), 0, float64(ui.Width()), float64(con.Height()))
	con.Fill()
	ui.has.Draw(con)
}

func (ui *UI) IncrementRound() {
	ui.has.round += 1
}

func (ui *UI) AddHealth(health int) {
	ui.has.RoundHealthScoreState.health += health
}

func (ui *UI) AddScore(score int) {
	ui.has.RoundHealthScoreState.score += score
}

func (ui *UI) Round() int {
	return ui.has.round
}

func (ui *UI) Health() int {
	return ui.has.health
}

func NewHealthAndScoreUI() *RoundHealthScoreUI {
	font, err := truetype.Parse(goregular.TTF)
	core.Check(err)
	face := truetype.NewFace(font, &truetype.Options{Size: large})
	return &RoundHealthScoreUI{&RoundHealthScoreState{0, 100, 0}, face, large, offWhite}
}

func (has *RoundHealthScoreUI) Draw(con *gg.Context) {
	con.SetFontFace(has.f)
	con.SetColor(has.c)
	con.DrawString(fmt.Sprintf("Round:  %d", has.round), float64(con.Width()-(width-10)), float64(con.Height()/2)-30)
	con.DrawString(fmt.Sprintf("Health: %d", has.health), float64(con.Width()-(width-10)), float64(con.Height()/2))
	con.DrawString(fmt.Sprintf("Score:  %d", has.score), float64(con.Width()-(width-10)), float64(con.Height()/2)+30)
}
