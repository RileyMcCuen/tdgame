package core

import (
	"log"
	"os"
	"path"
	"reflect"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type (
	PreMeta struct {
		Meta
		FilePath   string
		Attributes yaml.Node
	}
	Meta struct {
		Type    Kind
		Variety Kind
		Name    Kind
	}
	Range struct {
		Min int
		Max int
	}
	AssetSpec struct {
		Meta
	}
	Kinder interface {
		Kind() Kind
	}
	DeclarationHandler interface {
		Type() Kind
		Match(pm *PreMeta) (spec Kinder, priority int)
		// PreLoad is called on each DeclarationHandler after all matching has been done
		PreLoad(d *Declarations)
		Load(spec Kinder, decs *Declarations)
	}
	Spec struct {
		spec     Kinder
		priority int
	}
	Declarations struct {
		specs    []Spec
		handlers map[Kind]DeclarationHandler
	}
)

func (m Meta) Kind() Kind {
	return m.Type
}

func StructToYaml(in interface{}) string {
	out, err := yaml.Marshal(in)
	Check(err)
	return string(out)
}

func (d *Declarations) HandlePreMeta(pm *PreMeta) {
	handler := d.handlers[pm.Kind()]
	if spec, prior := handler.Match(pm); spec != nil {
		val := reflect.ValueOf(spec)
		if val.Kind() != reflect.Ptr {
			panic("spec returned by handler is not a pointer")
		}
		derefVal := val.Elem()
		metaVal := derefVal.FieldByName("Meta")
		metaVal.Set(reflect.ValueOf(pm.Meta))
		attrVal := derefVal.FieldByNameFunc(func(s string) bool {
			return strings.Contains(s, "Attributes")
		})
		Check(pm.Attributes.Decode(attrVal.Addr().Interface()))
		d.specs = append(d.specs, Spec{spec, prior})
	} else {
		panic("declaration handler returned nil")
	}
}

func NewDeclarations() *Declarations {
	ret := &Declarations{make([]Spec, 0), make(map[Kind]DeclarationHandler)}
	return ret
}

func (d *Declarations) RegisterHandlers(dhs ...DeclarationHandler) *Declarations {
	for _, dh := range dhs {
		d.handlers[dh.Type()] = dh
	}
	return d
}

func (d *Declarations) AddDir(dir string) *Declarations {
	entries, err := os.ReadDir(dir)
	Check(err)
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		fullPath := path.Join(dir, entry.Name())
		d.AddFile(fullPath)
	}
	return d
}

func (d *Declarations) AddFile(fil string) *Declarations {
	log.Println("Adding file:", fil)
	data, err := os.ReadFile(fil)
	Check(err)
	pm := &PreMeta{Meta{}, fil, yaml.Node{}}
	Check(yaml.Unmarshal(data, pm))
	d.HandlePreMeta(pm)
	return d
}

func (d *Declarations) Load() *Declarations {
	for _, handler := range d.handlers {
		handler.PreLoad(d)
	}
	// sort specs by priority
	sort.Slice(d.specs, func(i, j int) bool {
		return d.specs[i].priority < d.specs[j].priority
	})
	// load each handler in order of priority
	for _, spec := range d.specs {
		log.Printf("%T\n", spec.spec)
		d.handlers[spec.spec.Kind()].Load(spec.spec, d)
	}
	return d
}

func (d *Declarations) Get(k Kind) DeclarationHandler {
	return d.handlers[k]
}
