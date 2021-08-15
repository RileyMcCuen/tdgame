package main

type (
	AnimatorAtlas struct {
		anims map[Kind]Animator
	}
	Animatable interface {
		Locator
	}
	Animator interface {
		Animate(a Animatable)
		Kind() Kind
		Done() bool
		Copy() Animator
		Reset()
	}
	Location struct {
		Point
		rot int
	}
	LocationWrapper struct {
		l Location
	}
	// Performs animation for a single tile distance
	TileAnimator struct {
		k      Kind
		action func(tick int, l Location) Location
		t      *Ticker
	}
	SerialAnimator struct {
		k     Kind
		anims []Animator
		cur   int
	}
	PrecalculatedAnimator struct {
		k    Kind
		locs []Location
		t    *Ticker
	}
)

const (
	RotationTime = 30
)

func LocWrapper(l Location) *LocationWrapper {
	return &LocationWrapper{l}
}

func (l *LocationWrapper) Location() Location {
	return l.l
}

func (l *LocationWrapper) SetLocation(loc Location) {
	l.l = loc
}

func Loc(p Point, rot int) Location {
	return Location{p, rot}
}

func (l Location) North() Location {
	return Loc(l.Point.North(), l.rot)
}

func (l Location) East() Location {
	return Loc(l.Point.East(), l.rot)
}

func (l Location) South() Location {
	return Loc(l.Point.South(), l.rot)
}

func (l Location) West() Location {
	return Loc(l.Point.West(), l.rot)
}

func (l Location) Clockwise(delta int) Location {
	return Loc(l.Point, Clockwise(l.rot, delta))
}

func (l Location) CounterClockwise(delta int) Location {
	return Loc(l.Point, CounterClockwise(l.rot, delta))
}

func BlankAnimator(ticks int) Animator {
	return NewTileAnimator(ticks, Blk, func(tick int, l Location) Location { return l })
}

func MakeAnimatorAtlas() AnimatorAtlas {
	return AnimatorAtlas{make(map[Kind]Animator)}
}

func (aa AnimatorAtlas) Animator(k Kind) Animator {
	return aa.anims[k].Copy()
}

func (aa AnimatorAtlas) PrecalculatedAnimator(k Kind) *PrecalculatedAnimator {
	return aa.Animator(k).(*PrecalculatedAnimator)
}

func (aa AnimatorAtlas) PutAnimator(k Kind, a Animator) {
	aa.anims[k] = a
}

func (aa AnimatorAtlas) SerialAnimatorFromPath(key Kind, p []Kind) Animator {
	anims := make([]Animator, len(p))
	for i, k := range p {
		//fmt.Println(k)
		anims[i] = aa.anims[k]
	}
	anim := NewSerialAnimator(key, anims...)
	aa.anims[key] = anim
	return anim.Copy()
}

func NewTileAnimator(max int, k Kind, action func(tick int, l Location) Location) Animator {
	return &TileAnimator{k, action, NewTicker(max)}
}

func (ba *TileAnimator) Kind() Kind {
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

func NewSerialAnimator(k Kind, anims ...Animator) Animator {
	return &SerialAnimator{k, anims, 0}
}

func (sa *SerialAnimator) Reset() {
	sa.cur = 0
	for _, a := range sa.anims {
		a.Reset()
	}
}

func (sa *SerialAnimator) Kind() Kind {
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

func NewPrecalculatedAnimator(k Kind, start Location, a Animator) Animator {
	l, locs := LocWrapper(start), make([]Location, 0)
	for !a.Done() {
		a.Animate(l)
		locs = append(locs, l.Location())
	}
	// fmt.Println(locs)
	return &PrecalculatedAnimator{
		k,
		locs,
		NewTicker(len(locs)),
	}
}

func AnimatorFromLine(start, end Point, ticks int) Animator {
	flTicks := float64(ticks)
	locs := make([]Location, ticks)
	xdif, ydif := float64(end.x-start.x)/flTicks, float64(end.y-start.y)/flTicks
	curx, cury := float64(start.x), float64(start.y)
	for i := 0; i < ticks; i++ {
		curx, cury = curx+xdif, cury+ydif
		locs[i] = Loc(Pt(int(curx), int(cury)), 0)
	}
	// fmt.Println(locs)
	locs[len(locs)-1] = Loc(end, 0)
	return &PrecalculatedAnimator{"lineanim", locs, NewTicker(ticks)}
}

func (pa *PrecalculatedAnimator) Location(tick int) Location {
	if tick >= pa.t.max {
		return Loc(Pt(-1, -1), -1)
	}
	return pa.locs[tick]
}

func (pa *PrecalculatedAnimator) Animate(a Animatable) {
	if pa.Done() {
		return
	}
	a.SetLocation(pa.locs[pa.t.Ticks()])
	pa.t.Tick()
}

func (pa *PrecalculatedAnimator) Kind() Kind {
	return pa.k
}

func (pa *PrecalculatedAnimator) Done() bool {
	return pa.t.Done()
}

func (pa *PrecalculatedAnimator) Copy() Animator {
	return &PrecalculatedAnimator{
		pa.k,
		pa.locs,
		NewTicker(pa.t.max),
	}
}

func (pa *PrecalculatedAnimator) Reset() {
	pa.t.Reset()
}

func North(y int) int {
	return y - 1
}

func East(x int) int {
	return x + 1
}

func South(y int) int {
	return y + 1
}

func West(x int) int {
	return x - 1
}

func Clockwise(rot, delta int) int {
	return rot + delta
}

func CounterClockwise(rot, delta int) int {
	return rot - delta
}

var (
	// Straight anims (4 total)
	//// Vertical Turns
	NNSAnim = NewTileAnimator(TileSizeInt, NNS, func(_ int, l Location) Location {
		return l.North()
	})
	SSSAnim = NewTileAnimator(TileSizeInt, SSS, func(_ int, l Location) Location {
		return l.South()
	})
	//// Horizontal Turns
	EESAnim = NewTileAnimator(TileSizeInt, EES, func(_ int, l Location) Location {
		return l.East()
	})
	WWSAnim = NewTileAnimator(TileSizeInt, WWS, func(_ int, l Location) Location {
		return l.West()
	})
	// Turn anims (10 total - primary rotation anims)
	ClockwiseAnim = NewTileAnimator(RotationTime, CLW, func(_ int, l Location) Location {
		return l.Clockwise(3)
	})
	CounterClockwiseAnim = NewTileAnimator(RotationTime, CLW, func(_ int, l Location) Location {
		return l.CounterClockwise(3)
	})
	//// North Turns
	NETAnim = NewSerialAnimator(NET, NNSAnim, ClockwiseAnim)
	NWTAnim = NewSerialAnimator(NWT, NNSAnim, CounterClockwiseAnim)
	//// South Turns
	SETAnim = NewSerialAnimator(SET, SSSAnim, CounterClockwiseAnim)
	SWTAnim = NewSerialAnimator(SWT, SSSAnim, ClockwiseAnim)
	// East Turns
	ENTAnim = NewSerialAnimator(ENT, EESAnim, CounterClockwiseAnim)
	ESTAnim = NewSerialAnimator(EST, EESAnim, ClockwiseAnim)
	// West Turns
	WNTAnim              = NewSerialAnimator(WNT, WWSAnim, ClockwiseAnim)
	WSTAnim              = NewSerialAnimator(WST, WWSAnim, CounterClockwiseAnim)
	DefaultAnimatorAtlas = AnimatorAtlas{
		map[Kind]Animator{
			NNS: NNSAnim,
			SSS: SSSAnim,
			EES: EESAnim,
			WWS: WWSAnim,
			NET: NETAnim,
			NWT: NWTAnim,
			SET: SETAnim,
			SWT: SWTAnim,
			ENT: ENTAnim,
			EST: ESTAnim,
			WNT: WNTAnim,
			WST: WSTAnim,
		},
	}
)
