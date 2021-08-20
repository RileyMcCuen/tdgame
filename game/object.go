package game

import (
	"tdgame/core"

	"github.com/fogleman/gg"
)

type (
	Flag         struct{}
	DrawNoop     struct{}
	ProcessNoop  struct{}
	FinalizeNoop struct{}
	GameObject   interface {
		core.Processor
		core.Drawer
		Finalize() (layer int, effect GameObject)
	}
	GameObjectSet map[GameObject]Flag
	GameObjects   interface {
		Add(GameObject)
		Remove(GameObject)
		Process(ticks int, layers Layers)
		Draw(con *gg.Context)
	}
	Layers []GameObjects
)

var (
	On = struct{}{}
)

func (DrawNoop) Draw(con *gg.Context)            {}
func (ProcessNoop) Process(ticks int)            {}
func (ProcessNoop) Done() bool                   { return false }
func (FinalizeNoop) Finalize() (int, GameObject) { return -1, nil }

func NewGameObjects() GameObjects {
	return make(GameObjectSet)
}

func (dos GameObjectSet) Draw(con *gg.Context) {
	for d := range dos {
		d.Draw(con)
	}
}

func (dos GameObjectSet) Process(ticks int, layers Layers) {
	for d := range dos {
		d.Process(ticks)
		if d.Done() {
			dos.Remove(d)
			if layer, effect := d.Finalize(); effect != nil {
				layers.Add(layer, effect)
			}
		}
	}
}

func (dos GameObjectSet) Add(g GameObject) {
	dos[g] = On
}

func (dos GameObjectSet) Remove(g GameObject) {
	delete(dos, g)
}

func NewLayers(size int) Layers {
	ret := make(Layers, size)
	for i := range ret {
		ret[i] = NewGameObjects()
	}
	return ret
}

func (l Layers) Add(layer int, g GameObject) {
	l[layer].Add(g)
}

func (l Layers) Remove(layer int, g GameObject) {
	l[layer].Remove(g)
}

func (l Layers) Draw(con *gg.Context) {
	for _, layer := range l {
		layer.Draw(con)
	}
}

func (l Layers) Process(ticks int) {
	for _, layer := range l {
		layer.Process(ticks, l)
	}
}
