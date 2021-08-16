package main

import (
	"bytes"
	"image"
	_ "image/png"
	"math"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/fogleman/gg"
	"github.com/hajimehoshi/ebiten/v2"
)

type (
	Drawer interface {
		Draw(con *gg.Context)
	}
	SubImager interface {
		SubImage(r image.Rectangle) image.Image
	}
	Asset interface {
		Processor
		Draw(con *gg.Context, l Location)
		Copy() Asset
		Reset()
	}
	StaticAsset struct {
		image.Image
	}
	CenteredAsset struct {
		StaticAsset
		p Point
	}
	Sprite struct {
		image.Image
		frames                   []image.Image
		total, speed, cur, width int
		t                        *Ticker
	}
	AssetAtlas struct {
		assets map[Kind]Asset
		prefix string
	}
)

var (
	SpriteRegEx   = regexp.MustCompile(`(\w+)_(\d+)_(\d+)`)
	CenteredRegEx = regexp.MustCompile(`centered_(\w+)`)
)

func (a *StaticAsset) Draw(con *gg.Context, l Location) {
	sz := a.Bounds().Size()
	con.Push()
	con.RotateAbout(gg.Radians(float64(l.rot)), float64(l.x+sz.X), float64(l.y+sz.Y))
	con.DrawImage(a.Image, l.x, l.y)
	con.Pop()
}

func (a *StaticAsset) Copy() Asset {
	return &StaticAsset{a.Image}
}

func (s *StaticAsset) Process(ticks int) {}

func (s *StaticAsset) Done() bool { return false }

func (a *StaticAsset) Reset() {}

func (c *CenteredAsset) Draw(con *gg.Context, l Location) {
	x, y := l.x+c.p.x, l.y+c.p.y
	con.Push()
	con.RotateAbout(gg.Radians(float64(l.rot)), float64(y), float64(x))
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
		s.speed,
		0,
		s.width,
		NewTicker(s.t.Max()),
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

func (s *Sprite) Draw(con *gg.Context, l Location) {
	con.Push()
	con.RotateAbout(gg.Radians(float64(l.rot)), float64(l.x+32), float64(l.y+32))
	img := s.CurrentFrame()
	con.DrawImage(img, l.x-img.Bounds().Min.X, l.y-img.Bounds().Min.Y)
	con.RotateAbout(-gg.Radians(float64(l.rot)), float64(l.x+32), float64(l.y+32))
	con.Pop()
}

func MakeAssetAtlas(prefix string) AssetAtlas {
	ret := AssetAtlas{
		make(map[Kind]Asset),
		prefix,
	}
	entries, err := os.ReadDir(prefix)
	Check(err)
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		ret.LoadAsset(entry.Name())
	}
	return ret
}

func (aa AssetAtlas) Blank() Asset {
	return aa.assets[Blk]
}

func (aa AssetAtlas) Asset(k Kind) Asset {
	return aa.assets[k].Copy()
}

func (aa AssetAtlas) Sprite(k Kind) *Sprite {
	return aa.Asset(k).(*Sprite)
}

func (aa AssetAtlas) PutAsset(k Kind, a Asset) {
	aa.assets[k] = a
}

func readPNG(file string) image.Image {
	data, err := os.ReadFile(file)
	Check(err)
	img, _, err := image.Decode(bytes.NewBuffer(data))
	Check(err)
	return img
}

func (aa AssetAtlas) HandleStaticAsset(name string, img image.Image) {
	kinds := strings.Split(name, "_")
	as := &StaticAsset{img}
	for _, k := range kinds {
		aa.assets[Kind(k)] = as
	}
}

func (aa AssetAtlas) HandleCenteredAsset(name string, img image.Image) {
	matches := CenteredRegEx.FindStringSubmatch(name)
	k := Kind(matches[1])
	sz := img.Bounds().Size()
	x, y := (TileSizeInt-sz.X)/2, (TileSizeInt-sz.Y)/2
	aa.assets[k] = &CenteredAsset{StaticAsset{img}, Pt(x, y)}
}

func (aa AssetAtlas) HandleSprite(name string, img image.Image) {
	matches := SpriteRegEx.FindStringSubmatch(name)
	k := Kind(matches[1])
	speed, err := strconv.Atoi(matches[2])
	Check(err)
	width, err := strconv.Atoi(matches[3])
	Check(err)
	total := img.Bounds().Max.X / width
	imgs := make([]image.Image, total)
	y0, y1 := 0, img.Bounds().Max.Y
	for i := range imgs {
		r := image.Rect(i*width, y0, (i+1)*width, y1)
		imgs[i] = img.(SubImager).SubImage(r)
	}
	t := NewTicker(int(math.RoundToEven(float64(ebiten.MaxTPS() / speed))))
	aa.assets[k] = &Sprite{img, imgs, total, speed, 0, width, t}
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
