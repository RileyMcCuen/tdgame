package main

import (
	"fmt"
	"image/color"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"

	"github.com/fogleman/gg"
	"github.com/hajimehoshi/ebiten/v2"
)

type (
	Direction int
	Kind      string
	Point     struct {
		x, y int
	}
	Node struct {
		Point
		k Kind
		a Asset
	}
	NodeDirection struct {
		Direction
		*Node
	}
	Nodes            []*Node
	Graph            []Nodes
	CachedImageGraph struct {
		image *ebiten.Image
		start Point
		path  []Kind
		Graph
	}
)

const (
	TileSizeInt int       = 64
	TileSize    float64   = 64
	Zero        float64   = 0
	N           Direction = iota
	E
	S
	W
	Blk Kind = "BLK" // blank (not part of the path)
	// Straights
	NNS Kind = "NNS"
	SSS Kind = "SSS"
	EES Kind = "EES"
	WWS Kind = "WWS"
	// Turns
	CLW Kind = "CLW" // Clockwise
	CCW Kind = "CCW" // CounterClockwise
	NET Kind = "NET"
	NWT Kind = "NWT"
	SET Kind = "SET"
	SWT Kind = "SWT"
	ENT Kind = "ENT"
	EST Kind = "EST"
	WNT Kind = "WNT"
	WST Kind = "WST"
)

var (
	Grid       = true
	Directions = []Direction{N, E, S, W}
	PointRegEx = regexp.MustCompile(`(\d+),(\d+)`)
)

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
	if entry == exit { // straight
		return Kind(fmt.Sprintf("%s%sS", entry, exit))
	} else { // turn
		return Kind(fmt.Sprintf("%s%sT", entry, exit))
	}
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

func (p Point) Coordinates() (float64, float64, float64, float64) {
	return TileIndexToCoordinate(p.x),
		TileIndexToCoordinate(p.y),
		TileSize,
		TileSize
}

func (p Point) Near(o Point, rSquared int) bool {
	return p.DistanceSquared(o) <= rSquared
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

func (p Point) DistanceSquared(o Point) int {
	return Square(p.x-o.x) + Square(p.y-o.y)
}

func (p Point) String() string {
	return fmt.Sprintf("(%d,%d)", p.x, p.y)
}

func BlankNode(p Point) *Node {
	return &Node{p, Blk, &StaticAsset{}}
}

func Nd(p Point, k Kind, a Asset) *Node {
	return &Node{p, k, a}
}

func (n Node) IsBlank() bool {
	return n.k == Blk
}

func (n Node) Draw(con *gg.Context) {
	n.a.Draw(con, Loc(Pt(n.x*64, n.y*64), 0))
}

func (n Node) String() string {
	if n.a == nil {
		return fmt.Sprintf("%s:%s is nil", n.Point.String(), n.k)
	} else {
		return fmt.Sprintf("%s:%s no nil", n.Point.String(), n.k)
	}
}

func NewGraph(width, height int) Graph {
	if width <= 0 || height <= 0 {
		panic("graph must have width and height of >= 1")
	}
	ret := make(Graph, height)
	for i := range ret {
		row := make(Nodes, width)
		for j := range row {
			row[j] = BlankNode(Pt(j, i))
		}
		ret[i] = row
	}
	return ret
}

func GraphFromFile(fil string, aa AssetAtlas) CachedImageGraph {
	data, err := ioutil.ReadFile(fil)
	Check(err)
	sdata := string(data)
	lines := strings.Split(sdata, "\n")
	if len(lines) < 1 {
		panic("must have at least 1 line of directions")
	}
	pStrs := PointRegEx.FindStringSubmatch(lines[0])
	if pStrs == nil {
		panic("first line should include start point")
	}
	lines = lines[1:]
	x, err := strconv.Atoi(pStrs[1])
	Check(err)
	y, err := strconv.Atoi(pStrs[2])
	Check(err)
	p, dirs := Pt(x, y), make([]Direction, len(lines))
	for i, line := range lines {
		dirs[i] = StringToDirection(line)
	}
	return GraphFromPath(p, dirs, aa)
}

func GraphFromPath(start Point, dirs []Direction, aa AssetAtlas) CachedImageGraph {
	p, width, height := start, start.x, start.y
	if width < 0 || height < 0 {
		panic("start point coordinates must be positive")
	}
	if width != 0 && height != 0 {
		panic("start point must be on West or North border of map")
	}
	for _, d := range dirs {
		p = p.Neighbor(d)
		if p.x < 0 || p.y < 0 {
			panic("point in path has negative coordinates which are not allowed")
		}
		width, height = MaxInt(width, p.x), MaxInt(height, p.y)
	}
	end, g := p, NewGraph(width+1, height+1)
	if end.x != width && end.y != height {
		panic("end point must be on East or South border of map")
	}
	if start.y == 0 {
		dirs = append([]Direction{S}, dirs...)
	} else { // start.x == 0
		dirs = append([]Direction{E}, dirs...)
	}
	if end.y == height {
		dirs = append(dirs, S)
	} else { // end.x == g.Width()
		dirs = append(dirs, E)
	}
	p = start
	kinds := make([]Kind, len(dirs)-1)
	for i := 1; i < len(dirs); i++ {
		entry, exit := dirs[i-1], dirs[i]
		k := DirectionsToKind(entry, exit)
		kinds[i-1] = k
		g[p.y][p.x] = Nd(p, k, aa.Asset(k))
		p = p.Neighbor(exit)
	}
	blankAsset := aa.Blank()
	for _, row := range g {
		for _, n := range row {
			if n.IsBlank() {
				n.a = blankAsset
			}
		}
	}
	con := gg.NewContext(g.Size())
	g.Draw(con)
	eimg := ebiten.NewImageFromImage(con.Image())
	return CachedImageGraph{eimg, start, kinds, g}
}

func (g Graph) Height() int {
	return len(g)
}

func (g Graph) Width() int {
	return len(g[0])
}

func (g Graph) Contains(p Point) bool {
	return p.x >= 0 && p.x < g.Width() && p.y >= 0 && p.y < g.Height()
}

func (g Graph) Node(p Point) *Node {
	return g[p.y][p.x]
}

func (g Graph) Neighbors(n Node) []NodeDirection {
	ret := make([]NodeDirection, 0)
	for _, d := range Directions {
		p := n.Neighbor(d)
		if g.Contains(p) {
			ret = append(ret, NodeDirection{d, g.Node(p)})
		}
	}
	return ret
}

func (g Graph) Size() (int, int) {
	return g.Width() * int(TileSize), g.Height() * int(TileSize)
}

func (g CachedImageGraph) Draw(screen *ebiten.Image) {
	screen.DrawImage(g.image, &ebiten.DrawImageOptions{})
}

func (g CachedImageGraph) InitialRotation() int {
	rot := 0 /// SSS is first dir
	if g.path[0] == EES {
		rot = CounterClockwise(rot, 90)
	}
	return rot
}

func (g Graph) String() string {
	sb := strings.Builder{}
	for _, row := range g {
		for _, n := range row {
			sb.WriteString(n.String())
			sb.WriteString(" ")
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func (g Graph) Draw(con *gg.Context) {
	for _, row := range g {
		for _, n := range row {
			n.Draw(con)
		}
	}

	if Grid {
		minY, maxY := Zero, TileIndexToCoordinate(g.Height())
		for x := 0; x <= g.Width(); x++ {
			xCoord := TileIndexToCoordinate(x)
			con.DrawLine(xCoord, minY, xCoord, float64(maxY))
		}

		minX, maxX := Zero, TileIndexToCoordinate(g.Width())
		for y := 0; y <= g.Height(); y++ {
			yCoord := TileIndexToCoordinate(y)
			con.DrawLine(minX, yCoord, maxX, yCoord)
		}

		con.SetLineWidth(2)
		con.SetColor(color.Black)
		con.Stroke()
	}
}
