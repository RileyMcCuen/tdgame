package asset

import (
	"bytes"
	"image"
	_ "image/png"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"tdgame/core"
	"tdgame/util"

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
		image.Image
	}
	CenteredAsset struct {
		StaticAsset
		p core.Point
	}
	Sprite struct {
		image.Image
		frames                   []image.Image
		total, delay, cur, width int
		t                        *core.Ticker
	}
	AssetAtlas struct {
		assets map[core.Kind]Asset
		prefix string
	}
)

var (
	SpriteRegEx   = regexp.MustCompile(`(\w+)_(\d+)_(\d+)`) //tag_delay_width
	CenteredRegEx = regexp.MustCompile(`centered_(\w+)`)    //centered_tag
)

func (a *StaticAsset) Draw(con *gg.Context, l core.Location) {
	sz := a.Bounds().Size()
	con.Push()
	con.RotateAbout(gg.Radians(float64(l.Rot())), float64(l.X()+sz.X), float64(l.Y()+sz.Y))
	con.DrawImage(a.Image, l.X(), l.Y())
	con.Pop()
}

func (a *StaticAsset) Copy() Asset {
	return &StaticAsset{a.Image}
}

func (s *StaticAsset) Process(ticks int) {}

func (s *StaticAsset) Done() bool { return false }

func (a *StaticAsset) Reset() {}

func (c *CenteredAsset) Draw(con *gg.Context, l core.Location) {
	x, y := l.X()+c.p.X(), l.Y()+c.p.Y()
	con.Push()
	con.RotateAbout(gg.Radians(float64(l.Rot())), float64(y), float64(x))
	con.DrawImage(c.Image, x, y)
	con.Pop()
}

func (c *CenteredAsset) Copy() Asset {
	return &CenteredAsset{c.StaticAsset, c.p}
}

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

func MakeAssetAtlas(prefix string) AssetAtlas {
	ret := AssetAtlas{
		make(map[core.Kind]Asset),
		prefix,
	}
	entries, err := os.ReadDir(prefix)
	util.Check(err)
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		ret.LoadAsset(entry.Name())
	}
	return ret
}

func (aa AssetAtlas) Blank() Asset {
	return aa.assets[core.Blk]
}

func (aa AssetAtlas) Asset(k core.Kind) Asset {
	return aa.assets[k].Copy()
}

func (aa AssetAtlas) Sprite(k core.Kind) *Sprite {
	return aa.Asset(k).(*Sprite)
}

func (aa AssetAtlas) PutAsset(k core.Kind, a Asset) {
	aa.assets[k] = a
}

func readPNG(file string) image.Image {
	data, err := os.ReadFile(file)
	util.Check(err)
	img, _, err := image.Decode(bytes.NewBuffer(data))
	util.Check(err)
	return img
}

func (aa AssetAtlas) HandleStaticAsset(name string, img image.Image) {
	kinds := strings.Split(name, "_")
	as := &StaticAsset{img}
	for _, k := range kinds {
		aa.assets[core.Kind(k)] = as
	}
}

func (aa AssetAtlas) HandleCenteredAsset(name string, img image.Image) {
	matches := CenteredRegEx.FindStringSubmatch(name)
	k := core.Kind(matches[1])
	sz := img.Bounds().Size()
	x, y := (core.TileSizeInt-sz.X)/2, (core.TileSizeInt-sz.Y)/2
	aa.assets[k] = &CenteredAsset{StaticAsset{img}, core.Pt(x, y)}
}

func (aa AssetAtlas) HandleSprite(name string, img image.Image) {
	matches := SpriteRegEx.FindStringSubmatch(name)
	k := core.Kind(matches[1])
	delay, err := strconv.Atoi(matches[2])
	util.Check(err)
	width, err := strconv.Atoi(matches[3])
	util.Check(err)
	total := img.Bounds().Max.X / width
	imgs := make([]image.Image, total)
	y0, y1 := 0, img.Bounds().Max.Y
	for i := range imgs {
		r := image.Rect(i*width, y0, (i+1)*width, y1)
		imgs[i] = img.(SubImager).SubImage(r)
	}
	t := core.NewTicker(delay)
	aa.assets[k] = &Sprite{img, imgs, total, delay, 0, width, t}
}

func (aa AssetAtlas) LoadAsset(file string) {
	name, fullFile := strings.TrimSuffix(path.Base(file), path.Ext(file)), path.Join(aa.prefix, file)
	img := readPNG(fullFile)
	if SpriteRegEx.MatchString(name) {
		aa.HandleSprite(name, img)
	} else if CenteredRegEx.MatchString(name) {
		aa.HandleCenteredAsset(name, img)
	} else {
		aa.HandleStaticAsset(name, img)
	}
}
