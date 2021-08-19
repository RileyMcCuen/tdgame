package asset

import (
	"bytes"
	"image"
	"log"
	"os"
	"path"
	"strings"
	"tdgame/core"
	"tdgame/util"
)

type (
	StaticAttributes struct {
		FilePrefix string `yaml:"filePrefix"`
		Files      []string
	}
	StaticSpec struct {
		core.Meta
		StaticAttributes
		FilePath string
	}
	SpriteDescription struct {
		File  string
		Delay int
		Width int
	}
	SpriteAttributes struct {
		FilePrefix string `yaml:"filePrefix"`
		Files      []SpriteDescription
	}
	SpriteSpec struct {
		core.Meta
		SpriteAttributes
		FilePath string
	}
	MultiAttributes struct {
		File   string
		Width  int
		Height int
		Assets []struct {
			Tags []core.Kind
			X    int
			Y    int
		}
	}
	MultiSpec struct {
		core.Meta
		MultiAttributes
		FilePath string
	}
	AssetAtlas map[core.Kind]Asset
)

const (
	AssetType     = "asset"
	StaticVariety = "static"
	SpriteVariety = "sprite"
	MultiVariety  = "multi"
)

func NewAssetAtlas() AssetAtlas {
	ret := make(AssetAtlas)
	return ret
}

func (aa AssetAtlas) Blank() Asset {
	return aa[core.Bl]
}

func (aa AssetAtlas) Asset(k core.Kind) Asset {
	log.Println(k)
	return aa[k].Copy()
}

func (aa AssetAtlas) Sprite(k core.Kind) *Sprite {
	return aa.Asset(k).(*Sprite)
}

func NewStaticAsset(img image.Image) *StaticAsset {
	sz := img.Bounds().Size()
	x, y := (core.TileSizeInt-sz.X)/2, (core.TileSizeInt-sz.Y)/2
	return &StaticAsset{core.Pt(x, y), img}
}

func (spec *StaticSpec) AddAssets(aa AssetAtlas) {
	for _, fil := range spec.Files {
		name := core.Kind(strings.TrimSuffix(path.Base(fil), ".png"))
		aa[name] = NewStaticAsset(readPNG(path.Join(spec.FilePath, spec.FilePrefix, fil)))
	}
}

func (spec *SpriteSpec) AddAssets(aa AssetAtlas) {
	for _, fil := range spec.Files {
		name := core.Kind(strings.TrimSuffix(path.Base(fil.File), ".png"))
		img := readPNG(path.Join(spec.FilePath, spec.FilePrefix, fil.File))
		total := img.Bounds().Max.X / fil.Width
		imgs := make([]image.Image, total)
		y0, y1 := 0, img.Bounds().Max.Y
		for i := range imgs {
			r := image.Rect(i*fil.Width, y0, (i+1)*fil.Width, y1)
			imgs[i] = img.(SubImager).SubImage(r)
		}
		t := core.NewTicker(fil.Delay)
		aa[name] = &Sprite{img, imgs, total, fil.Delay, 0, fil.Width, t}
	}
}

func (spec *MultiSpec) AddAssets(aa AssetAtlas) {
	img := readPNG(path.Join(spec.FilePath, spec.File))
	simg := img.(SubImager)
	for _, desc := range spec.Assets {
		partImg := simg.SubImage(image.Rect(desc.X, desc.Y, desc.X+spec.Width, desc.Y+spec.Height))
		for _, tag := range desc.Tags {
			log.Println(tag, partImg.Bounds())
			aa[tag] = NewStaticAsset(partImg)
		}
	}
}

// func (aa AssetAtlas) HandleCenteredAsset(name string, img image.Image) {
// 	matches := CenteredRegEx.FindStringSubmatch(name)
// 	k := core.Kind(matches[1])
// 	sz := img.Bounds().Size()
// 	x, y := (core.TileSizeInt-sz.X)/2, (core.TileSizeInt-sz.Y)/2
// 	aa.assets[k] = &CenteredAsset{StaticAsset{img}, core.Pt(x, y)}
// }

// func (aa AssetAtlas) LoadAsset(file string) {
// 	name, fullFile := strings.TrimSuffix(path.Base(file), path.Ext(file)), path.Join(aa.prefix, file)
// 	img := readPNG(fullFile)
// 	if SpriteRegEx.MatchString(name) {
// 		aa.HandleSprite(name, img)
// 	} else if CenteredRegEx.MatchString(name) {
// 		aa.HandleCenteredAsset(name, img)
// 	} else {
// 		aa.HandleStaticAsset(name, img)
// 	}
// }

var _ core.DeclarationHandler = AssetAtlas{}

func (aa AssetAtlas) Type() core.Kind {
	return AssetType
}

func (aa AssetAtlas) Match(pm *core.PreMeta) (core.Kinder, int) {
	switch pm.Variety {
	case StaticVariety:
		return &StaticSpec{FilePath: pm.FilePath}, 0
	case SpriteVariety:
		return &SpriteSpec{FilePath: pm.FilePath}, 0
	case MultiVariety:
		return &MultiSpec{FilePath: pm.FilePath}, 0
	default:
		panic("variety of asset does not exist")
	}
}

func (aa AssetAtlas) Load(spec core.Kinder, decs *core.Declarations) {
	switch s := spec.(type) {
	case *StaticSpec:
		s.AddAssets(aa)
	case *SpriteSpec:
		s.AddAssets(aa)
	case *MultiSpec:
		s.AddAssets(aa)
	default:
		panic("variety of asset does not exist")
	}
}

func readPNG(file string) image.Image {
	data, err := os.ReadFile(file)
	util.Check(err)
	img, _, err := image.Decode(bytes.NewBuffer(data))
	util.Check(err)
	return img
}
