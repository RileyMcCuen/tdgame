package core

import (
	"github.com/fogleman/gg"
)

type (
	Flag           struct{}
	GameObjectNoop struct{}
	ContextKey     string
	ContextValue   interface{}
	Attributer     interface {
		Attribute(ContextKey) ContextValue
		SetAttribute(ContextKey, ContextValue)
	}
	Context interface {
		Attributer
		Add(layer Layer, obj GameObject)
		Remove(layer Layer, obj GameObject)
	}
	BasicContext struct {
		Layers
		Attributer
	}
	Processor interface {
		Process(tick int, con Context) bool
	}
	Drawer interface {
		Draw(con *gg.Context)
	}
	GameObject interface {
		Processor
		Drawer
	}
	GameObjectSet map[GameObject]Flag
	GameObjects   interface {
		GameObject
		Add(GameObject)
		Remove(GameObject)
	}
	Layer  int
	Layers []GameObjects
)

const (
	TileLayer Layer = iota
	PrimitiveLayer
	ProjectileLayer
	TowerLayer
	EnemyLayer
	EffectLayer
	NumberOfLayers
)

func (GameObjectNoop) Draw(con *gg.Context)                {}
func (GameObjectNoop) Process(ticks int, con Context) bool { return false }

var (
	On            = struct{}{}
	MinGameObject = GameObjectNoop{}
)

func NewGameObjects() GameObjects {
	return make(GameObjectSet)
}

func (dos GameObjectSet) Draw(con *gg.Context) {
	for d := range dos {
		d.Draw(con)
	}
}

func (dos GameObjectSet) Process(ticks int, con Context) bool {
	for d := range dos {
		if d.Process(ticks, con) {
			dos.Remove(d)
		}
	}
	// GameObjectSet is never done, it should not be cleaned up until the end of the game
	return false
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

func (l Layers) Add(layer Layer, g GameObject) {
	l[layer].Add(g)
}

func (l Layers) Remove(layer Layer, g GameObject) {
	l[layer].Remove(g)
}

func (l Layers) Draw(con *gg.Context) {
	for _, layer := range l {
		layer.Draw(con)
	}
}

func (l Layers) Process(ticks int, a Attributer) {
	for _, layer := range l {
		layer.Process(ticks, &BasicContext{l, a})
	}
}
