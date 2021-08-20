package asset

import (
	"image"
	_ "image/png"
	"regexp"
	"tdgame/core"

	"github.com/fogleman/gg"
)

type (
	SubImager interface {
		SubImage(r image.Rectangle) image.Image
	}
	Asset interface {
		core.Processor
		Size() core.Point
		Offset() core.Point
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
		offset, size             core.Point
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

func (a *StaticAsset) Offset() core.Point {
	return a.offset
}

func (a *StaticAsset) Size() core.Point {
	sz := a.Image.Bounds().Size()
	return core.Pt(sz.X, sz.Y)
}

func (a *StaticAsset) Copy() Asset {
	return &StaticAsset{a.offset, a.Image}
}

func (s *StaticAsset) Process(ticks int, con core.Context) bool { return false }

func (s *StaticAsset) Done() bool { return false }

func (a *StaticAsset) Reset() {}

func (s *Sprite) Process(ticks int, con core.Context) bool {
	if s.t.Tick() {
		s.IncrementFrame()
		s.t.Reset()
	}
	return false
}

func (s *Sprite) Done() bool {
	return false
}

func (s *Sprite) Offset() core.Point {
	return s.offset
}

func (s *Sprite) Size() core.Point {
	return s.size
}

func (s *Sprite) Copy() Asset {
	return &Sprite{
		s.Image,
		s.frames,
		s.offset,
		s.size,
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
	img := s.CurrentFrame()
	con.Push()
	// con.RotateAbout(gg.Radians(float64(l.Rot())), float64(l.X()+32), float64(l.Y()+32))
	con.RotateAbout(gg.Radians(float64(l.Rot())), float64(l.X()+(s.size.X()/2)), float64(l.Y()+(s.size.Y()/2)))
	con.DrawImage(img, (l.X()-img.Bounds().Min.X)+s.offset.X(), (l.Y()-img.Bounds().Min.Y)+s.offset.Y())
	con.Pop()
}

func (s *Sprite) Length() int {
	return s.delay * s.total
}
