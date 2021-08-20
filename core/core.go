package core

import (
	"container/list"
	"fmt"
	"log"
	"math"
	"regexp"

	"github.com/fogleman/gg"
)

type (
	Direction int
	Kind      string
	Point     struct {
		x, y int
	}
	Location struct {
		Point
		rot int
	}
	LocationWrapper struct {
		l *Location
	}
	Locator interface {
		Location() Location
		SetLocation(Location)
	}
	ListItem interface {
		Elem() *list.Element
		SetElem(*list.Element)
	}
)

const (
	Zero float64   = 0
	N    Direction = iota
	E
	S
	W
	Bl Kind = "BL" // blank (not part of the path)
	// Straights
	NN Kind = "NN"
	SS Kind = "SS"
	EE Kind = "EE"
	WW Kind = "WW"
	// Turns
	CL Kind = "CL" // Clockwise
	CC Kind = "CC" // CounterClockwise
	NE Kind = "NE"
	NW Kind = "NW"
	SE Kind = "SE"
	SW Kind = "SW"
	EN Kind = "EN"
	ES Kind = "ES"
	WN Kind = "WN"
	WS Kind = "WS"
)

var (
	TileSizeInt int     = 64
	TileSize    float64 = 64
	ZeroPt              = Point{0, 0}
	ZeroLoc             = Loc(ZeroPt, 0)
	TileSizePt          = Pt(TileSizeInt, TileSizeInt)
	Grid                = true
	Directions          = []Direction{N, E, S, W}
	PointRegEx          = regexp.MustCompile(`(\d+),(\d+)`)
)

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

func StringToDirection(s string) Direction {
	switch s {
	case "N":
		return N
	case "E":
		return E
	case "S":
		return S
	case "W":
		return W
	default:
		log.Printf("'%s'\n", s)
		panic("cannot convert string to direction")
	}
}

func TileIndexToCoordinate(idx int) float64 {
	return (float64(idx) * TileSize)
}

func (d Direction) Opposite() Direction {
	switch d {
	case N:
		return S
	case E:
		return W
	case S:
		return N
	case W:
		return E
	default:
		panic("cannot take opposite of unknown direction")
	}
}

func (d Direction) String() string {
	switch d {
	case N:
		return "N"
	case E:
		return "E"
	case S:
		return "S"
	case W:
		return "W"
	default:
		panic("cannot stringify an unknown direction")
	}
}

func DirectionsToKind(entry, exit Direction) Kind {
	if entry.Opposite() == exit {
		panic("entry and exit cannot be opposite in Kind")
	}
	return Kind(fmt.Sprintf("%s%s", entry, exit))
}

func Pt(x, y int) Point {
	return Point{x, y}
}

func (p Point) North() Point {
	return Pt(p.x, North(p.y))
}

func (p Point) East() Point {
	return Pt(East(p.x), p.y)
}

func (p Point) South() Point {
	return Pt(p.x, South(p.y))
}

func (p Point) West() Point {
	return Pt(West(p.x), p.y)
}

func (p Point) Neighbor(d Direction) Point {
	switch d {
	case N:
		return Point{p.x, p.y - 1}
	case E:
		return Point{p.x + 1, p.y}
	case S:
		return Point{p.x, p.y + 1}
	case W:
		return Point{p.x - 1, p.y}
	default:
		panic("invalid direction supplied")
	}
}

func (p Point) X() int {
	return p.x
}

func (p Point) Y() int {
	return p.y
}

func (p Point) XY() (int, int) {
	return p.x, p.y
}

func (p Point) Coordinates() (float64, float64, float64, float64) {
	return TileIndexToCoordinate(p.x),
		TileIndexToCoordinate(p.y),
		TileSize,
		TileSize
}

func (p Point) Near(o Point, maxDist int) bool {
	return p.DistanceSquared(o) <= Square(maxDist)
}

func (p Point) Reduce(r int) Point {
	return Pt(p.x/r, p.y/r)
}

func (p Point) Scale(s int) Point {
	return Pt(p.x*s, p.y*s)
}

func (p Point) Add(o Point) Point {
	return Pt(p.x+o.x, p.y+o.y)
}

func (p Point) Subtract(o Point) Point {
	return Pt(p.x-o.x, p.y-o.y)
}

func (p Point) Multiply(o Point) Point {
	return Pt(p.x*o.x, p.y*o.y)
}

func (p Point) DistanceSquared(o Point) int {
	p = p.Subtract(o)
	return Square(p.x) + Square(p.y)
}

func (p Point) String() string {
	return fmt.Sprintf("(%d,%d)", p.x, p.y)
}

func (p Point) TileIndex() Point {
	return p.Reduce(TileSizeInt)
}

func LocWrapper(l Location) *LocationWrapper {
	return &LocationWrapper{&Location{l.Point, l.rot}}
}

func (l *LocationWrapper) Location() Location {
	return *l.l
}

func (l *LocationWrapper) SetLocation(loc Location) {
	*l.l = loc
}

func (l *LocationWrapper) SetX(x int) {
	l.l.x = x
}

func (l *LocationWrapper) SetY(y int) {
	l.l.y = y
}

func (l *LocationWrapper) SetRot(rot int) {
	l.l.rot = rot
}

func (l *LocationWrapper) Copy() *LocationWrapper {
	return LocWrapper(*l.l)
}

func Loc(p Point, rot int) Location {
	return Location{p, rot}
}

func (l Location) RotateByATan2(o Location) Location {
	curx, cury := l.XY()
	ex, ey := o.XY()
	dx, dy := float64(ex-curx), float64(cury-ey)
	return Loc(l.Point, int(gg.Degrees(math.Atan2(dx, dy))))
}

func (l Location) Rot() int {
	return l.rot
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
