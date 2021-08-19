package asset

import (
	"image"
	_ "image/png"
	"regexp"
	"tdgame/core"

	"github.com/fogleman/gg"
)

type (
	Drawer interface {
		Draw(con *gg.Context)
	}
	SubImager interface {
		SubImage(r image.Rectangle) image.Image
	}
	Asset interface {
		core.Processor
		Draw(con *gg.Context, l core.Location)
		Copy() Asset
		Reset()
	}
	StaticAsset struct {
		offset core.Point
		image.Image
	}
	Sprite struct {
		image.Image
		frames                   []image.Image
		total, delay, cur, width int
		t                        *core.Ticker
	}
)

var (
	SpriteRegEx   = regexp.MustCompile(`(\w+)_(\d+)_(\d+)`) //tag_delay_width
	CenteredRegEx = regexp.MustCompile(`centered_(\w+)`)    //centered_tag
)

func (a *StaticAsset) Draw(con *gg.Context, l core.Location) {
	sz := a.Bounds().Size()
	con.Push()
	con.RotateAbout(gg.Radians(float64(l.Rot())), float64(l.X()+sz.X+a.offset.X()), float64(l.Y()+sz.Y+a.offset.Y()))
	con.DrawImage(a.Image, (l.X()-a.Bounds().Min.X)+a.offset.X(), (l.Y()-a.Bounds().Min.Y)+a.offset.X())
	con.Pop()
}

func (a *StaticAsset) Copy() Asset {
	return &StaticAsset{a.offset, a.Image}
}

func (s *StaticAsset) Process(ticks int) {}

func (s *StaticAsset) Done() bool { return false }

func (a *StaticAsset) Reset() {}

func (s *Sprite) Process(ticks int) {
	if s.t.Tick() {
		s.IncrementFrame()
		s.t.Reset()
	}
}

func (s *Sprite) Done() bool {
	return false
}

func (s *Sprite) Copy() Asset {
	return &Sprite{
		s.Image,
		s.frames,
		s.total,
		s.delay,
		0,
		s.width,
		core.NewTicker(s.delay),
	}
}

func (s *Sprite) CurrentFrame() image.Image {
	return s.frames[s.cur]
}

func (s *Sprite) IncrementFrame() {
	if s.cur == s.total-1 {
		s.cur = 0
	} else {
		s.cur++
	}
}

func (s *Sprite) Reset() {
	s.cur = 0
	s.t.Reset()
}

func (s *Sprite) Draw(con *gg.Context, l core.Location) {
	con.Push()
	con.RotateAbout(gg.Radians(float64(l.Rot())), float64(l.X()+32), float64(l.Y()+32))
	img := s.CurrentFrame()
	con.DrawImage(img, l.X()-img.Bounds().Min.X, l.Y()-img.Bounds().Min.Y)
	// con.RotateAbout(-gg.Radians(float64(l.Rot())), float64(l.X()+32), float64(l.Y()+32))
	con.Pop()
}

func (s *Sprite) Length() int {
	return s.delay * s.total
}
