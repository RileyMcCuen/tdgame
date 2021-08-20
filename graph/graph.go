package graph

import (
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"path"
	"sort"
	"strconv"
	"strings"
	"tdgame/asset"
	"tdgame/core"

	"github.com/fogleman/gg"
)

type (
	Graph interface {
		core.Processor
		Spec() *GraphSpec
		Width() int
		Height() int
		Node(core.Point) *Node
		TLoc(offset, size core.Point) *TileLocation
	}
	GraphAttributes struct {
		File string
	}
	GraphSpec struct {
		core.Meta
		GraphAttributes
		FilePath string
	}
	GraphAtlas map[core.Kind]Graph
)

type (
	TileLocation struct {
		g Graph
		*core.LocationWrapper
		Offset, Size core.Point
		Tiles        [4]core.Point
	}
	Collider interface {
		Location() core.Location
		Radius() int
		Near(Collider) bool
		Kind() core.Kind
	}
	Damageable interface {
		Collider
		TakeDamage(amount int)
	}
	Damager interface {
		Collider
		DoDamage(Damageable)
	}
	Node struct {
		dables        []Damageable
		dmgers        []Damager
		distanceToEnd int
		core.Point
		k core.Kind
		a asset.Asset
	}
	NodeDirection struct {
		core.Direction
		*Node
	}
	Nodes            []*Node
	BasicGraph       []Nodes
	CachedImageGraph struct {
		*GraphSpec
		// imageWithGrid, image *ebiten.Image
		imageWithGrid, image image.Image
		start                core.Point
		path                 []core.Kind
		BasicGraph
	}
)

const (
	Priority      = 1
	GraphType     = "graph"
	CachedVariety = "cached"
)

var _ core.DeclarationHandler = GraphAtlas{}

func NewGraphAtlas() GraphAtlas {
	ret := make(GraphAtlas)
	return ret
}

func (ga GraphAtlas) Type() core.Kind {
	return GraphType
}

func (ga GraphAtlas) Match(pm *core.PreMeta) (spec core.Kinder, priority int) {
	switch pm.Variety {
	case CachedVariety:
		return &GraphSpec{FilePath: pm.FilePath}, 2
	default:
		panic("variety of graph does not exist")
	}
}

func (ga GraphAtlas) PreLoad(d *core.Declarations) {

}

func (ga GraphAtlas) Load(spec core.Kinder, decs *core.Declarations) {
	switch spec.(type) {
	case *GraphSpec:
		g := spec.(*GraphSpec)
		ga[g.Name] = GraphFromSpec(g, decs.Get(asset.AssetType).(asset.AssetAtlas))
	default:
		panic("variety of graph does not exist")
	}
}

func (ga GraphAtlas) Graph(k core.Kind) Graph {
	return ga[k]
}

func (g CachedImageGraph) TLoc(offset, size core.Point) *TileLocation {
	return &TileLocation{
		g,
		nil,
		offset,
		size,
		[4]core.Point{},
	}
}

func (t *TileLocation) calculateTiles(col Collider) {
	center, half := t.LocationWrapper.Location().Add(t.Offset), t.Size.Reduce(2)
	ohalf := half.Multiply(core.Pt(1, -1))
	nw := center.Subtract(half)
	ne := center.Add(ohalf)
	sw := center.Subtract(ohalf)
	se := center.Add(half)
	tiles := [4]core.Point{nw.TileIndex(), ne.TileIndex(), sw.TileIndex(), se.TileIndex()}
	for i := 0; i < 4; i++ {
		tile, match := t.Tiles[i], false
		for j := 0; j < 4; j++ {
			match = match || tile == tiles[j]
		}
		if !match {
			// do remove, tile is not in new tiles
			t.g.Node(tile).Remove(col)
		}
	}
	for i := 0; i < 4; i++ {
		tile, match := tiles[i], false
		for j := 0; j < 4; j++ {
			match = match || tile == t.Tiles[j]
		}
		if !match {
			// do add, tile is not in old tiles
			t.g.Node(tile).Add(col)
		}
	}
	t.Tiles = tiles
}

func (t *TileLocation) Move(l core.Location, col Collider) {
	if t.LocationWrapper.Location().Point == l.Point {
		t.SetRot(l.Rot())
		return
	}
	t.SetLocation(l)
	t.calculateTiles(col)
}

func (t *TileLocation) Copy() *TileLocation {
	return t.g.TLoc(t.Offset, t.Size)
}

func BlankNode(p core.Point) *Node {
	return &Node{make([]Damageable, 0), make([]Damager, 0), 0, p, core.Bl, &asset.StaticAsset{}}
}

func Nd(dist int, p core.Point, k core.Kind, a asset.Asset) *Node {
	return &Node{make([]Damageable, 0), make([]Damager, 0), dist, p, k, a}
}

func (n Node) IsBlank() bool {
	return n.k == core.Bl
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

func (n Node) Add(col Collider) {
	switch d := col.(type) {
	case Damageable:
		n.dables = append(n.dables, d)
	case Damager:
		n.dmgers = append(n.dmgers, d)
	default:
		panic("cannot add non Collider, not Damager nor Damageable")
	}
}

func (n Node) Remove(col Collider) {
	switch d := col.(type) {
	case Damageable:
		for i, dable := range n.dables {
			if dable == d {
				n.dables = append(n.dables[:i], n.dables[i+1:]...)
			}
		}
	case Damager:
		for i, dmger := range n.dmgers {
			if dmger == d {
				n.dmgers = append(n.dmgers[:i], n.dmgers[i+1:]...)
			}
		}
	default:
		panic("cannot add non Collider, not Damager nor Damageable")
	}
}

func (n Node) Process(ticks int) {
	for _, dmger := range n.dmgers {
		for _, dable := range n.dables {
			if dmger.Near(dable) {

			}
		}
	}
}

func (n Node) Done() bool { return false }

func NewGraph(width, height int) BasicGraph {
	if width <= 0 || height <= 0 {
		panic("graph must have width and height of >= 1")
	}
	ret := make(BasicGraph, height)
	for i := range ret {
		row := make(Nodes, width)
		for j := range row {
			row[j] = BlankNode(core.Pt(j, i))
		}
		ret[i] = row
	}
	return ret
}

func GraphFromSpec(spec *GraphSpec, aa asset.AssetAtlas) CachedImageGraph {
	data, err := ioutil.ReadFile(path.Join(spec.FilePath, spec.File))
	core.Check(err)
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
	core.Check(err)
	y, err := strconv.Atoi(pStrs[2])
	core.Check(err)
	p, dirs := core.Pt(x, y), make([]core.Direction, len(lines))
	for i, line := range lines {
		dirs[i] = core.StringToDirection(strings.TrimSpace(line))
	}
	return GraphFromPath(spec, p, dirs, aa)
}

func GraphFromPath(spec *GraphSpec, start core.Point, dirs []core.Direction, aa asset.AssetAtlas) CachedImageGraph {
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
		width, height = core.MaxInt(width, p.X()), core.MaxInt(height, p.Y())
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
		g[p.Y()][p.X()] = Nd(len(dirs)-i, p, k, aa.Asset(k))
		p = p.Neighbor(exit)
	}
	kinds = append(kinds, core.DirectionsToKind(dirs[len(dirs)-1], dirs[len(dirs)-1]))
	blankAsset := aa.Blank()
	for _, row := range g {
		for _, n := range row {
			if n.IsBlank() {
				n.a = blankAsset
			}
		}
	}
	core.Grid = false
	con := gg.NewContext(g.Size())
	g.Draw(con)
	eimg := con.Image() // ebiten.NewImageFromImage(con.Image())
	core.Grid = true
	con = gg.NewContext(g.Size())
	g.Draw(con)
	eimgWithGrid := con.Image() // ebiten.NewImageFromImage(con.Image())
	return CachedImageGraph{spec, eimgWithGrid, eimg, start, kinds, g}
}

func (g BasicGraph) Process(ticks int) {
	for _, row := range g {
		for _, nd := range row {
			nd.Process(ticks)
		}
	}
}

func (g BasicGraph) Done() bool { return false }

func (g BasicGraph) Height() int {
	return len(g)
}

func (g BasicGraph) Width() int {
	return len(g[0])
}

func (g BasicGraph) Contains(p core.Point) bool {
	return p.X() >= 0 && p.X() < g.Width() && p.Y() >= 0 && p.Y() < g.Height()
}

func (g BasicGraph) Node(p core.Point) *Node {
	if p.X() < 0 || p.Y() < 0 || p.X() >= g.Width() || p.Y() >= g.Height() {
		return nil
	}
	return g[p.Y()][p.X()]
}

func (g BasicGraph) Neighbors(n Node) []NodeDirection {
	ret := make([]NodeDirection, 0)
	for _, d := range core.Directions {
		p := n.Point.Neighbor(d)
		if g.Contains(p) {
			ret = append(ret, NodeDirection{d, g.Node(p)})
		}
	}
	return ret
}

func (g BasicGraph) Size() (int, int) {
	return g.Width() * core.TileSizeInt, g.Height() * core.TileSizeInt
}

func (g BasicGraph) String() string {
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

func (g BasicGraph) TilesAround(p core.Point, rng core.Range) []*Node {
	tp, ret := p.TileIndex(), make([]*Node, 0)
	for dist := rng.Min; dist < rng.Max; dist++ {
		for i := -dist; i <= dist; i++ {
			for j := -dist; j <= dist; j++ {
				ntp := tp.Add(core.Pt(i, j))
				if nd := g.Node(ntp); nd != nil {
					ret = append(ret, nd)
				}
			}
		}
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i].distanceToEnd < ret[j].distanceToEnd })
	return ret
}

func (g BasicGraph) Draw(con *gg.Context) {
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

func (g CachedImageGraph) Spec() *GraphSpec {
	return g.GraphSpec
}

// func (g CachedImageGraph) Draw(screen *ebiten.Image) {
func (g CachedImageGraph) Draw(con *gg.Context) {
	if core.Grid {
		con.DrawImage(g.imageWithGrid, 0, 0)
		// screen.DrawImage(g.imageWithGrid, &ebiten.DrawImageOptions{})
	} else {
		con.DrawImage(g.image, 0, 0)
		// screen.DrawImage(g.imageWithGrid, &ebiten.DrawImageOptions{})
	}
}

func (g CachedImageGraph) InitialRotation() int {
	rot := 0 /// SSS is first dir
	if g.path[0] == core.EE {
		rot = core.CounterClockwise(rot, 90)
	}
	return rot
}

func (g CachedImageGraph) InitialPoint() core.Point {
	ret := g.start
	if g.path[0] == core.SS {
		ret = ret.Add(core.Pt(0, -1))
	} else {
		ret = ret.Add(core.Pt(-1, 0))
	}
	return ret
}

func (g CachedImageGraph) StartLoc() core.Location {
	return core.Loc(g.InitialPoint().Scale(core.TileSizeInt), g.InitialRotation())
}
