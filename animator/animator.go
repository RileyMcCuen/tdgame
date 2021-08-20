package animator

import (
	"fmt"
	"tdgame/core"
	"tdgame/graph"
)

type (
	AnimatorAttributes struct {
		Dummy string
	}
	AnimatorSpec struct {
		core.Meta
		AnimatorAttributes
	}
	AnimatorAtlas struct {
		anims map[core.Kind]Animator
	}
	Animatable interface {
		core.Locator
		Speed() int
	}
	Animator interface {
		Animate(a Animatable)
		Kind() core.Kind
		Done() bool
		Copy() Animator
		Reset()
	}
	animatableLoc struct {
		*core.LocationWrapper
	}
	// Performs animation for a single tile distance
	TileAnimator struct {
		k      core.Kind
		action func(tick int, l core.Location) core.Location
		t      *core.Ticker
	}
	SerialAnimator struct {
		k     core.Kind
		anims []Animator
		cur   int
	}
	PrecalculatedAnimator struct {
		k    core.Kind
		locs []core.Location
		t    *core.Ticker
	}
)

const (
	RotationTime = 30
	AnimatorType = "animator"
	PathVariety  = "path"
)

func (animatableLoc) Speed() int {
	return 1
}

func BlankAnimator(ticks int) Animator {
	return NewTileAnimator(ticks, core.Bl, func(tick int, l core.Location) core.Location { return l })
}

func NewTileAnimator(max int, k core.Kind, action func(tick int, l core.Location) core.Location) Animator {
	return &TileAnimator{k, action, core.NewTicker(max)}
}

func (ba *TileAnimator) Kind() core.Kind {
	return ba.k
}

func (ba *TileAnimator) Animate(a Animatable) {
	a.SetLocation(ba.action(ba.t.Ticks(), a.Location()))
	ba.t.Tick()
}

func (ba *TileAnimator) Done() bool {
	return ba.t.Done()
}

func (ba *TileAnimator) Copy() Animator {
	return NewTileAnimator(ba.t.Max(), ba.k, ba.action)
}

func (ba *TileAnimator) Reset() {
	ba.t.Reset()
}

func NewSerialAnimator(k core.Kind, anims ...Animator) Animator {
	return &SerialAnimator{k, anims, 0}
}

func (sa *SerialAnimator) Reset() {
	sa.cur = 0
	for _, a := range sa.anims {
		a.Reset()
	}
}

func (sa *SerialAnimator) Kind() core.Kind {
	return sa.k
}

func (sa *SerialAnimator) CurrentAnimator() Animator {
	return sa.anims[sa.cur]
}

func (sa *SerialAnimator) Animate(a Animatable) {
	anim := sa.CurrentAnimator()
	anim.Animate(a)
	if anim.Done() {
		sa.cur++
	}
}

func (sa *SerialAnimator) Done() bool {
	return sa.cur == len(sa.anims)
}

func (sa *SerialAnimator) Copy() Animator {
	anims := make([]Animator, len(sa.anims))
	for i, anim := range sa.anims {
		anims[i] = anim.Copy()
	}
	return NewSerialAnimator(sa.k, anims...)
}

func NewPrecalculatedAnimator(k core.Kind, start core.Location, a Animator) Animator {
	l, locs := animatableLoc{core.LocWrapper(start)}, make([]core.Location, 0)
	for !a.Done() {
		a.Animate(l)
		locs = append(locs, l.Location())
	}
	// // fmt.Println(locs)
	return &PrecalculatedAnimator{
		k,
		locs,
		core.NewTicker(len(locs)),
	}
}

func AnimatorFromLine(start, end core.Point, ticks int) *PrecalculatedAnimator {
	flTicks := float64(ticks)
	locs := make([]core.Location, ticks)
	xdif, ydif := float64(end.X()-start.X())/flTicks, float64(end.Y()-start.Y())/flTicks
	curx, cury := float64(start.X()), float64(start.Y())
	for i := 0; i < ticks; i++ {
		curx, cury = curx+xdif, cury+ydif
		locs[i] = core.Loc(core.Pt(int(curx), int(cury)), 0)
	}
	// fmt.Println(locs)
	locs[len(locs)-1] = core.Loc(end, 0)
	return &PrecalculatedAnimator{"lineanim", locs, core.NewTicker(ticks)}
}

func (pa *PrecalculatedAnimator) Location(tick int) core.Location {
	if tick >= pa.t.Max() {
		return core.Loc(core.Pt(-1, -1), -1)
	}
	return pa.locs[tick]
}

func (pa *PrecalculatedAnimator) LocationOffset(tick int) (core.Location, bool) {
	tick += pa.t.Ticks()
	if tick >= pa.t.Max() {
		return core.Loc(core.Pt(-2048, -2048), 0), false
	}
	return pa.locs[tick], true
}

func (pa *PrecalculatedAnimator) LastLocation() (core.Location, int) {
	maxTicks := pa.t.Max()
	return pa.locs[maxTicks-1], maxTicks - pa.t.Ticks()
}

func (pa *PrecalculatedAnimator) Animate(a Animatable) {
	if pa.Done() {
		return
	}
	a.SetLocation(pa.locs[pa.t.Ticks()])
	pa.t.TickBy(a.Speed())
}

func (pa *PrecalculatedAnimator) Kind() core.Kind {
	return pa.k
}

func (pa *PrecalculatedAnimator) Done() bool {
	return pa.t.Done()
}

func (pa *PrecalculatedAnimator) Copy() Animator {
	return &PrecalculatedAnimator{
		pa.k,
		pa.locs,
		core.NewTicker(pa.t.Max()),
	}
}

func (pa *PrecalculatedAnimator) Reset() {
	pa.t.Reset()
}

var _ core.DeclarationHandler = AnimatorAtlas{}

func NewAnimatorAtlas() AnimatorAtlas {
	return AnimatorAtlas{make(map[core.Kind]Animator)}
}

func (aa AnimatorAtlas) Type() core.Kind {
	return AnimatorType
}

func (aa AnimatorAtlas) Match(pm *core.PreMeta) (core.Kinder, int) {
	switch pm.Variety {
	case PathVariety:
		return &AnimatorSpec{}, 3
	default:
		panic("variety of animator does not exist")
	}
}

func (aa AnimatorAtlas) PreLoad(d *core.Declarations) {

}

func (aa AnimatorAtlas) Load(spec core.Kinder, d *core.Declarations) {
	g := d.Get(graph.GraphType).(graph.GraphAtlas).Graph("map").(graph.CachedImageGraph)
	switch spec.(type) {
	case *AnimatorSpec:
		aa.CreatePathAnimator(g.StartLoc(), g.Path())
	default:
		panic("variety of animator does not exist")
	}
}

func (aa AnimatorAtlas) Animator(k core.Kind) Animator {
	return aa.anims[k].Copy()
}

func (aa AnimatorAtlas) PrecalculatedAnimator(k core.Kind) *PrecalculatedAnimator {
	fmt.Println(k)
	return aa.Animator(k).(*PrecalculatedAnimator)
}

func (aa AnimatorAtlas) PutAnimator(k core.Kind, a Animator) {
	aa.anims[k] = a
}

func (aa AnimatorAtlas) SerialAnimatorFromPath(key core.Kind, p []core.Kind) Animator {
	anims := make([]Animator, len(p))
	for i, k := range p {
		//// fmt.Println(k)
		anims[i] = aa.anims[k]
	}
	anim := NewSerialAnimator(key, anims...)
	aa.anims[key] = anim
	return anim.Copy()
}

func (aa AnimatorAtlas) CreatePathAnimator(startLoc core.Location, path []core.Kind) {
	loc := startLoc
	sanim := aa.SerialAnimatorFromPath(core.Kind("path"), path)
	aa.PutAnimator("prepath", NewPrecalculatedAnimator("prepath", loc, sanim))
}

var (
	// Straight anims (4 total)
	//// Vertical Turns
	NNAnim = NewTileAnimator(core.TileSizeInt, core.NN, func(_ int, l core.Location) core.Location {
		return l.North()
	})
	SSAnim = NewTileAnimator(core.TileSizeInt, core.SS, func(_ int, l core.Location) core.Location {
		return l.South()
	})
	//// Horizontal Turns
	EEAnim = NewTileAnimator(core.TileSizeInt, core.EE, func(_ int, l core.Location) core.Location {
		return l.East()
	})
	WWAnim = NewTileAnimator(core.TileSizeInt, core.WW, func(_ int, l core.Location) core.Location {
		return l.West()
	})
	// Turn anims (10 total - primary rotation anims)
	ClockwiseAnim = NewTileAnimator(RotationTime, core.CL, func(_ int, l core.Location) core.Location {
		return l.Clockwise(3)
	})
	CounterClockwiseAnim = NewTileAnimator(RotationTime, core.CL, func(_ int, l core.Location) core.Location {
		return l.CounterClockwise(3)
	})
	//// North Turns
	NEAnim = NewSerialAnimator(core.NE, NNAnim, ClockwiseAnim)
	NWAnim = NewSerialAnimator(core.NW, NNAnim, CounterClockwiseAnim)
	//// South Turns
	SEAnim = NewSerialAnimator(core.SE, SSAnim, CounterClockwiseAnim)
	SWAnim = NewSerialAnimator(core.SW, SSAnim, ClockwiseAnim)
	// East Turns
	ENAnim = NewSerialAnimator(core.EN, EEAnim, CounterClockwiseAnim)
	ESAnim = NewSerialAnimator(core.ES, EEAnim, ClockwiseAnim)
	// WES Turns
	WNAnim               = NewSerialAnimator(core.WN, WWAnim, ClockwiseAnim)
	WSAnim               = NewSerialAnimator(core.WS, WWAnim, CounterClockwiseAnim)
	DefaultAnimatorAtlas = AnimatorAtlas{
		map[core.Kind]Animator{
			core.NN: NNAnim,
			core.SS: SSAnim,
			core.EE: EEAnim,
			core.WW: WWAnim,
			core.NE: NEAnim,
			core.NW: NWAnim,
			core.SE: SEAnim,
			core.SW: SWAnim,
			core.EN: ENAnim,
			core.ES: ESAnim,
			core.WN: WNAnim,
			core.WS: WSAnim,
		},
	}
)
