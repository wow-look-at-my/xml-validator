package validator

const xsdNS = "http://www.w3.org/2001/XMLSchema"
const xsiNS = "http://www.w3.org/2001/XMLSchema-instance"

type Schema struct {
	TargetNamespace      string
	ElementFormDefault   string
	AttributeFormDefault string
	Elements             map[string]*ElementDecl
	Types                map[string]Type
	Groups               map[string]*Group
	AttrGroups           map[string]*AttrGroup
	// Attributes are the global xs:attribute declarations, keyed by local name
	// like Elements. A qualified attribute in an instance document is validated
	// against the declaration here whose Namespace matches its own; without one
	// there is nothing to validate against, so the attribute is an error.
	Attributes map[string]*AttrDecl
	Imports    []*Import
}

// Import is an xs:import directive recorded on the schema. The Namespace is
// the imported target namespace (may be empty) and SchemaLocation is the URI
// hint provided to locate the imported schema (may be empty).
type Import struct {
	Namespace      string
	SchemaLocation string
}

type ElementDecl struct {
	Name      string
	Namespace string // target namespace of the schema declaring this global element
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
	Name          string
	Mixed         bool
	Content       ContentModel
	Attributes    []*AttrDecl
	AnyAttribute  *AnyAttrDecl
	SimpleText    Type // non-nil for simpleContent
	attrGroupRefs []string
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

// AnyParticle is an xs:any wildcard. processContents is always "strict";
// the schema parser rejects any other value because this validator does not
// support a no-validation mode.
type AnyParticle struct {
	Namespace string
	MinOccurs int
	MaxOccurs int
}

// AnyAttrDecl is an xs:anyAttribute wildcard. processContents is always
// "strict" -- see [AnyParticle] for the rationale.
type AnyAttrDecl struct {
	Namespace string
}

type AttrDecl struct {
	Name string
	// Namespace is the target namespace of the schema declaring this global
	// attribute. It is empty on a local declaration, which is unqualified
	// unless the schema says otherwise.
	Namespace string
	TypeName  string
	Type      Type
	Use       string // required, optional, prohibited
	Default   string
	Fixed     string
	Ref       string
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
	Name         string
	Attributes   []*AttrDecl
	AnyAttribute *AnyAttrDecl
}
