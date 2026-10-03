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
	// Attributes are the global xs:attribute declarations.
	Attributes map[string]*AttrDecl
	Imports    []*Import
	// identity holds every xs:key and xs:unique by resolved name.
	identity     map[string]*IdentityConstraint
	identityRefs []*IdentityConstraint
	// BlockDefault is the schema's blockDefault attribute.
	BlockDefault string
	// prefixes are the namespace declarations this schema document made.
	prefixes map[string]string
}

// Import is an xs:import directive recorded on the schema.
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
	MaxOccurs int
	Default   string
	Fixed     string
	Nillable  bool
	Ref       string
	// SubstitutionGroup names the global element this may stand in for.
	SubstitutionGroup string
	Abstract          bool
	Block             string
	// substitutes are the declarations that may appear where a reference to this element does, transitively.
	substitutes []*ElementDecl
	// Alternatives are the xs:alternative type choices, in schema order.
	Alternatives []*TypeAlternative
	// Constraints are the xs:key, xs:keyref, and xs:unique declarations on this element.
	Constraints []*IdentityConstraint
	// compiled records that the constraint XPaths were compiled and registered on the schema.
	compiled bool
}

// IdentityConstraint is one xs:key, xs:keyref, or xs:unique. Selector and
// Fields hold the compiled XPaths; a keyref also names the key it points at.
type IdentityConstraint struct {
	Kind  string // "key", "keyref", or "unique"
	Name  string
	Refer string

	selectorXPath string
	fieldXPaths   []string
	selector      []idPath
	fields        [][]idPath
	referKey      string
}

type Type interface {
	typeName() string
}

type ComplexType struct {
	Name         string
	Mixed        bool
	Content      ContentModel
	Attributes   []*AttrDecl
	AnyAttribute *AnyAttrDecl
	SimpleText   Type // non-nil for simpleContent
	// baseName and derivation record an xs:complexContent derivation until the base type is resolvable.
	baseName      string
	derivation    string // "extension" or "restriction"
	attrGroupRefs []string
	// baseType is the type this derived from, kept after resolution folded the base in.
	baseType Type
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
func (*GroupRef) particle()    {}

// GroupRef is an xs:group ref particle.
type GroupRef struct {
	Ref       string
	MinOccurs int
	MaxOccurs int
}

// AnyParticle is an xs:any wildcard. processContents is always "strict".
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
	// Namespace is the target namespace of the schema declaring this global attribute.
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
