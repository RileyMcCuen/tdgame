package graph

import (
	"fmt"
	"image/color"
	"io/ioutil"
	"strconv"
	"strings"
	"tdgame/asset"
	"tdgame/core"
	"tdgame/util"

	"github.com/fogleman/gg"
	"github.com/hajimehoshi/ebiten/v2"
)

type (
	Node struct {
		core.Point
		k core.Kind
		a asset.Asset
	}
	NodeDirection struct {
		core.Direction
		*Node
	}
	Nodes            []*Node
	Graph            []Nodes
	CachedImageGraph struct {
		image *ebiten.Image
		start core.Point
		path  []core.Kind
		Graph
	}
)

func BlankNode(p core.Point) *Node {
	return &Node{p, core.Blk, &asset.StaticAsset{}}
}

func Nd(p core.Point, k core.Kind, a asset.Asset) *Node {
	return &Node{p, k, a}
}

func (n Node) IsBlank() bool {
	return n.k == core.Blk
}

func (n Node) Draw(con *gg.Context) {
	n.a.Draw(con, core.Loc(n.Point.Scale(64), 0))
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
			row[j] = BlankNode(core.Pt(j, i))
		}
		ret[i] = row
	}
	return ret
}

func GraphFromFile(fil string, aa asset.AssetAtlas) CachedImageGraph {
	data, err := ioutil.ReadFile(fil)
	util.Check(err)
	sdata := string(data)
	lines := strings.Split(sdata, "\n")
	if len(lines) < 1 {
		panic("must have at least 1 line of core.Directions")
	}
	pStrs := core.PointRegEx.FindStringSubmatch(lines[0])
	if pStrs == nil {
		panic("first line should include start point")
	}
	lines = lines[1:]
	x, err := strconv.Atoi(pStrs[1])
	util.Check(err)
	y, err := strconv.Atoi(pStrs[2])
	util.Check(err)
	p, dirs := core.Pt(x, y), make([]core.Direction, len(lines))
	for i, line := range lines {
		dirs[i] = core.StringToDirection(strings.TrimSpace(line))
	}
	return GraphFromPath(p, dirs, aa)
}

func GraphFromPath(start core.Point, dirs []core.Direction, aa asset.AssetAtlas) CachedImageGraph {
	p, width, height := start, start.X(), start.Y()
	if width < 0 || height < 0 {
		panic("start point coordinates must be positive")
	}
	if width != 0 && height != 0 {
		panic("start point must be on West or North border of map")
	}
	for _, d := range dirs {
		p = p.Neighbor(d)
		if p.X() < 0 || p.Y() < 0 {
			panic("point in path has negative coordinates which are not allowed")
		}
		width, height = util.MaxInt(width, p.X()), util.MaxInt(height, p.Y())
	}
	end, g := p, NewGraph(width+1, height+1)
	if end.X() != width && end.Y() != height {
		panic("end point must be on East or South border of map")
	}
	if start.Y() == 0 {
		dirs = append([]core.Direction{core.S}, dirs...)
	} else { // start.x == 0
		dirs = append([]core.Direction{core.E}, dirs...)
	}
	if end.Y() == height {
		dirs = append(dirs, core.S)
	} else { // end.x == g.Width()
		dirs = append(dirs, core.E)
	}
	p = start
	kinds := make([]core.Kind, len(dirs)-1)
	for i := 1; i < len(dirs); i++ {
		entry, exit := dirs[i-1], dirs[i]
		k := core.DirectionsToKind(entry, exit)
		kinds[i-1] = k
		g[p.Y()][p.X()] = Nd(p, k, aa.Asset(k))
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

func (g Graph) Contains(p core.Point) bool {
	return p.X() >= 0 && p.X() < g.Width() && p.Y() >= 0 && p.Y() < g.Height()
}

func (g Graph) Node(p core.Point) *Node {
	return g[p.Y()][p.X()]
}

func (g Graph) Neighbors(n Node) []NodeDirection {
	ret := make([]NodeDirection, 0)
	for _, d := range core.Directions {
		p := n.Point.Neighbor(d)
		if g.Contains(p) {
			ret = append(ret, NodeDirection{d, g.Node(p)})
		}
	}
	return ret
}

func (g Graph) Size() (int, int) {
	return g.Width() * core.TileSizeInt, g.Height() * core.TileSizeInt
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

	if core.Grid {
		minY, maxY := core.Zero, core.TileIndexToCoordinate(g.Height())
		for x := 0; x <= g.Width(); x++ {
			xCoord := core.TileIndexToCoordinate(x)
			con.DrawLine(xCoord, minY, xCoord, float64(maxY))
		}

		minX, maxX := core.Zero, core.TileIndexToCoordinate(g.Width())
		for y := 0; y <= g.Height(); y++ {
			yCoord := core.TileIndexToCoordinate(y)
			con.DrawLine(minX, yCoord, maxX, yCoord)
		}

		con.SetLineWidth(2)
		con.SetColor(color.Black)
		con.Stroke()
	}
}

func (g CachedImageGraph) Path() []core.Kind {
	return g.path
}

func (g CachedImageGraph) Draw(screen *ebiten.Image) {
	screen.DrawImage(g.image, &ebiten.DrawImageOptions{})
}

func (g CachedImageGraph) InitialRotation() int {
	rot := 0 /// SSS is first dir
	if g.path[0] == core.EES {
		rot = core.CounterClockwise(rot, 90)
	}
	return rot
}

func (g CachedImageGraph) InitialPoint() core.Point {
	ret := g.start
	if g.path[0] == core.SSS {
		ret = ret.Add(core.Pt(0, -1))
	} else {
		ret = ret.Add(core.Pt(-1, 0))
	}
	return ret
}

func (g CachedImageGraph) StartLoc() core.Location {
	return core.Loc(g.InitialPoint().Scale(core.TileSizeInt), g.InitialRotation())
}
