package td

import (
	"os"
	"path"
	"strings"
	"tdgame/asset"
	"tdgame/core"
	"tdgame/util"

	"gopkg.in/yaml.v3"
)

type (
	PreMeta struct {
		Meta
		Attributes yaml.Node
	}
	Meta struct {
		Type    string
		Variety string
		Name    core.Kind
	}
	EnemyAttributes struct {
		Asset     core.Kind
		Animation core.Kind
		Effect    core.Kind
		Health    int
		Speed     int
		Points    int
	}
	EnemySpec struct {
		Meta
		EnemyAttributes `yaml:"attributes"`
	}
	Range struct {
		Min int
		Max int
	}
	ProjectileAttributes struct {
		Asset           core.Kind
		Effect          core.Kind
		PoolSize        int `yaml:"poolSize"`
		Speed           int
		Damage          int
		ExplosionRadius int `yaml:"explosionRadius"`
	}
	TowerAttributes struct {
		ProjectileAttributes `yaml:"projectile"`
		Asset                core.Kind
		Range                // min, max ticks for projectile to reach enemy; min*speed, max*speed pixels donut radii
		Delay                int
		Cost                 int
	}
	TowerSpec struct {
		Meta
		TowerAttributes `yaml:"attributes"`
	}
	TowerAtlas   map[core.Kind]Tower
	EnemyAtlas   map[core.Kind]Enemy
	Declarations struct {
		assets asset.AssetAtlas
		anims  core.AnimatorAtlas
		TowerAtlas
		EnemyAtlas
	}
)

const (
	TowerType = "tower"
	EnemyType = "enemy"
)

func (m Meta) Kind() core.Kind {
	return m.Name
}

func (p ProjectileAttributes) Kind() core.Kind {
	return p.Asset
}

func StructToYaml(in interface{}) string {
	out, err := yaml.Marshal(in)
	util.Check(err)
	return string(out)
}

func (ts *TowerSpec) String() string {
	return StructToYaml(ts)
}

func (es *EnemySpec) String() string {
	return StructToYaml(es)
}

func (d *Declarations) HandlePreMeta(pm *PreMeta) {
	switch pm.Meta.Type {
	case TowerType:
		tSpec := &TowerSpec{pm.Meta, TowerAttributes{}}
		util.Check(pm.Attributes.Decode(&tSpec.TowerAttributes))
		// fmt.Println(tSpec)
		if tSpec.ExplosionRadius == 0 {
			panic("explosionradius cannot be zero")
		}
		tSpec.ExplosionRadius = util.MaxInt(1, tSpec.ExplosionRadius)
		// fmt.Println(*tSpec)
		d.TowerAtlas[pm.Kind()] = TowerFromSpec(tSpec, d.assets, d.anims)
	case EnemyType:
		eSpec := &EnemySpec{pm.Meta, EnemyAttributes{}}
		util.Check(pm.Attributes.Decode(&eSpec.EnemyAttributes))
		// fmt.Println(*eSpec)
		d.EnemyAtlas[pm.Kind()] = EnemyFromSpec(eSpec, d.assets, d.anims)
	default:
		panic("cannot handle PreMeta due to unknown type")
	}
}

func NewDeclarations(pathToDeclarations string, assets asset.AssetAtlas, anims core.AnimatorAtlas) *Declarations {
	ret := &Declarations{assets, anims, NewTowerAtlas(), NewEnemyAtlas()}
	entries, err := os.ReadDir(pathToDeclarations)
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		fullPath := path.Join(pathToDeclarations, entry.Name())
		data, err := os.ReadFile(fullPath)
		util.Check(err)
		pm := &PreMeta{}
		util.Check(yaml.Unmarshal(data, pm))
		ret.HandlePreMeta(pm)
	}
	return ret
}

func NewTowerAtlas() TowerAtlas {
	return TowerAtlas{}
}

func (ta TowerAtlas) Tower(l core.Location, k core.Kind) Tower {
	return ta[k].CopyAt(l)
}

func (ta TowerAtlas) AddTower(t Tower) {
	ta[t.Spec().Name] = t
}

func NewEnemyAtlas() EnemyAtlas {
	return EnemyAtlas{}
}

func (ea EnemyAtlas) Enemy(l core.Location, k core.Kind) Enemy {
	return ea[k].CopyAt(l)
}

func (ea EnemyAtlas) AddEnemy(e Enemy) {
	ea[e.Spec().Name] = e
}
