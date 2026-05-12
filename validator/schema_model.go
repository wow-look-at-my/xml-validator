package validator

const xsdNS = "http://www.w3.org/2001/XMLSchema"

type Schema struct {
	TargetNamespace      string
	ElementFormDefault   string
	AttributeFormDefault string
	Elements             map[string]*ElementDecl
	Types                map[string]Type
	Groups               map[string]*Group
	AttrGroups           map[string]*AttrGroup
}

type ElementDecl struct {
	Name      string
	TypeName  string
	Type      Type
	MinOccurs int
	MaxOccurs int // -1 = unbounded
	Default   string
	Fixed     string
	Nillable  bool
	Ref       string
}

type Type interface {
	typeName() string
}

type ComplexType struct {
	Name       string
	Mixed      bool
	Content    ContentModel
	Attributes []*AttrDecl
	SimpleText Type // non-nil for simpleContent
}

func (t *ComplexType) typeName() string { return t.Name }

type SimpleType struct {
	Name     string
	Base     string
	BaseType Type
	Facets   []Facet
	List     *SimpleType
	Union    []*SimpleType
}

func (t *SimpleType) typeName() string { return t.Name }

type ContentModel interface{ contentModel() }

type Sequence struct {
	Items     []Particle
	MinOccurs int
	MaxOccurs int
}

func (*Sequence) contentModel() {}

type Choice struct {
	Items     []Particle
	MinOccurs int
	MaxOccurs int
}

func (*Choice) contentModel() {}

type All struct {
	Items     []Particle
	MinOccurs int
	MaxOccurs int
}

func (*All) contentModel() {}

type Particle interface{ particle() }

func (*ElementDecl) particle() {}
func (*Sequence) particle()    {}
func (*Choice) particle()      {}
func (*All) particle()         {}
func (*AnyParticle) particle() {}

type AnyParticle struct {
	Namespace       string
	ProcessContents string
	MinOccurs       int
	MaxOccurs       int
}

type AttrDecl struct {
	Name    string
	TypeName string
	Type    Type
	Use     string // required, optional, prohibited
	Default string
	Fixed   string
	Ref     string
}

type Facet struct {
	Kind  string
	Value string
}

type Group struct {
	Name    string
	Content ContentModel
}

type AttrGroup struct {
	Name       string
	Attributes []*AttrDecl
}
