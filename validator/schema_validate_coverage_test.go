package validator

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateWithSchemaIOReader(t *testing.T) {
	xml := strings.NewReader(`<?xml version="1.1"?><root>hello</root>`)
	xsd := strings.NewReader(simpleXSD)
	require.NoError(t, ValidateWithSchema(xml, xsd))
}

func TestValidateWithSchemaBytesNilResolver(t *testing.T) {
	require.NoError(t, ValidateWithSchemaBytes([]byte(`<?xml version="1.1"?><root>hi</root>`), []byte(simpleXSD)))
}

func TestSchemaParseMalformedRoot(t *testing.T) {
	err := schemaValidate(t, `<?xml version="1.1"?><r/>`, `not xml at all`)
	require.Error(t, err)
}

func TestSchemaParseMalformedXMLDoc(t *testing.T) {
	err := schemaValidate(t, `<?xml version="1.1"?><r`, simpleXSD)
	require.Error(t, err)
}

func TestSchemaInvalidMinOccurs(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="x" type="xs:string" minOccurs="abc"/>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	err := schemaValidate(t, `<?xml version="1.1"?><root><x>v</x></root>`, xsd)
	require.Error(t, err)
}

func TestSchemaInvalidMaxOccurs(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="x" type="xs:string" maxOccurs="xyz"/>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	err := schemaValidate(t, `<?xml version="1.1"?><root><x>v</x></root>`, xsd)
	require.Error(t, err)
}

func TestSchemaSimpleTypeRestrictionBaseNonBuiltin(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:simpleType name="alphaType">
    <xs:restriction base="xs:string">
      <xs:enumeration value="A"/>
    </xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="alphaSubset">
    <xs:restriction base="alphaType">
      <xs:enumeration value="A"/>
    </xs:restriction>
  </xs:simpleType>
  <xs:element name="root" type="alphaSubset"/>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root>A</root>`, xsd)
}

// TestTypeInterfaceStubs invokes typeName, contentModel, and particle stubs
// on the various schema model types.
func TestTypeInterfaceStubs(t *testing.T) {
	var bt Type = &BuiltinType{name: "string"}
	assert.Equal(t, "string", bt.typeName())

	ct := &ComplexType{Name: "ct"}
	var ctType Type = ct
	assert.Equal(t, "ct", ctType.typeName())

	st := &SimpleType{Name: "st"}
	var stType Type = st
	assert.Equal(t, "st", stType.typeName())

	var cm ContentModel = &Sequence{}
	cm.contentModel()
	cm = &Choice{}
	cm.contentModel()
	cm = &All{}
	cm.contentModel()

	var p Particle = &ElementDecl{}
	p.particle()
	p = &Sequence{}
	p.particle()
	p = &Choice{}
	p.particle()
	p = &All{}
	p.particle()
	p = &AnyParticle{}
	p.particle()
}

func TestSchemaValidateNoTypeElement(t *testing.T) {
	// An <xs:element name="x"/> without a type validates anything (validator
	// returns early when decl.Type is nil).
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="anything"/>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root><anything>whatever</anything></root>`, xsd)
}

func TestSchemaValidateListEmpty(t *testing.T) {
	// A list with no items and no facets constraining the item count.
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:simpleType name="intList">
    <xs:list itemType="xs:integer"/>
  </xs:simpleType>
  <xs:element name="root" type="intList"/>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root></root>`, xsd)
	mustSchemaValid(t, `<?xml version="1.1"?><root>   </root>`, xsd)
}

func TestSchemaSimpleTextSimpleTypeError(t *testing.T) {
	// simpleContent extension over a SimpleType whose base builtin rejects
	// non-numeric text. Exercises the SimpleType branch of validateSimpleText.
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:simpleType name="codeType">
    <xs:restriction base="xs:integer"/>
  </xs:simpleType>
  <xs:element name="root">
    <xs:complexType>
      <xs:simpleContent>
        <xs:extension base="codeType">
          <xs:attribute name="kind" type="xs:string"/>
        </xs:extension>
      </xs:simpleContent>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><root kind="k">abc</root>`, xsd, "integer")
}

func TestSchemaSimpleTextEnumFacet(t *testing.T) {
	// simpleContent restriction declares the facets inline so they are
	// attached directly to the SimpleType used as SimpleText.
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:simpleContent>
        <xs:restriction base="xs:string">
          <xs:enumeration value="A"/>
          <xs:enumeration value="B"/>
          <xs:maxLength value="1"/>
          <xs:attribute name="kind" type="xs:string"/>
        </xs:restriction>
      </xs:simpleContent>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root kind="k">A</root>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><root kind="k">C</root>`, xsd, "not one of the allowed values")
}

func TestSchemaAttributeProhibited(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:attribute name="forbid" type="xs:string" use="prohibited"/>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root/>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><root forbid="x"/>`, xsd, "unexpected attribute")
}

func TestSchemaAttributeWithXmlPrefix(t *testing.T) {
	// xml: prefixed attributes should be ignored by attribute validation.
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:attribute name="x" type="xs:string"/>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root xml:lang="en" x="x"/>`, xsd)
}

func TestSchemaAttributeNamespacedNoWildcard(t *testing.T) {
	// Namespaced attribute that doesn't match any declared local attribute and
	// where no anyAttribute wildcard is present.
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:attribute name="x" type="xs:string"/>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><root xmlns:ext="http://e" ext:y="z" x="ok"/>`, xsd, "unexpected attribute")
}

func TestSchemaChoiceUnmatchedExceededMaxOccurs(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:choice maxOccurs="1">
        <xs:element name="a" type="xs:string"/>
        <xs:element name="b" type="xs:string"/>
      </xs:choice>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><root><a>1</a><a>2</a></root>`, xsd, "exceeded maxOccurs")
}

func TestSchemaChoiceUnmatched(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:choice maxOccurs="unbounded">
        <xs:element name="a" type="xs:string"/>
        <xs:element name="b" type="xs:string"/>
      </xs:choice>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><root><a>1</a><c>2</c></root>`, xsd, "not allowed here")
}

func TestSchemaAllChildExceedsMaxOccurs(t *testing.T) {
	// validateAll declMap branch when a child's maxOccurs is unbounded.
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:all>
        <xs:element name="a" type="xs:string" maxOccurs="1"/>
      </xs:all>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><root><a>1</a><a>2</a></root>`, xsd, "appears too many times")
}

func TestSchemaAllAnyExceedsMaxOccurs(t *testing.T) {
	// strict xs:any requires matched elements to have global declarations,
	// so the test elements are declared at the top level. The wildcard's
	// maxOccurs=1 bound is what we are testing here.
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="a" type="xs:string"/>
  <xs:element name="b" type="xs:string"/>
  <xs:element name="root">
    <xs:complexType>
      <xs:all>
        <xs:element name="name" type="xs:string"/>
        <xs:any namespace="##local" minOccurs="0" maxOccurs="1"/>
      </xs:all>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><root><name>hi</name><a/><b/></root>`, xsd, "exceeded maxOccurs")
}

func TestSchemaMatchAnyMinOccurs(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:sequence>
        <xs:any namespace="##any" minOccurs="2" maxOccurs="unbounded"/>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><root><a/></root>`, xsd, "xs:any")
}

func TestSchemaMatchSequenceMinOccurs(t *testing.T) {
	// A nested sequence with minOccurs="2" inside the top-level sequence
	// exercises matchSequence (the top-level uses validateSequence which
	// ignores minOccurs).
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:sequence>
        <xs:sequence minOccurs="2" maxOccurs="2">
          <xs:element name="a" type="xs:string"/>
        </xs:sequence>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><root><a>1</a></root>`, xsd, "requires at least")
}

func TestSchemaMatchChoiceMinOccurs(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:sequence>
        <xs:choice minOccurs="2" maxOccurs="2">
          <xs:element name="a" type="xs:string"/>
          <xs:element name="b" type="xs:string"/>
        </xs:choice>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><root><a>1</a></root>`, xsd, "choice")
}

func TestSchemaWildcardTargetNamespaceToken(t *testing.T) {
	// Exercises the multi-token namespace constraint with ##targetNamespace.
	// Under strict, the matched element must have a global declaration in the
	// target namespace -- so "more" is declared at the top level.
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" targetNamespace="https://example.com/main" xmlns:tns="https://example.com/main" elementFormDefault="qualified">
  <xs:element name="more" type="xs:string"/>
  <xs:element name="root">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="x" type="xs:string"/>
        <xs:any namespace="##targetNamespace http://other" minOccurs="0" maxOccurs="unbounded"/>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t,
		`<?xml version="1.1"?><tns:root xmlns:tns="https://example.com/main"><tns:x>v</tns:x><tns:more/></tns:root>`,
		xsd)
}

func TestSchemaResolveSimpleTypeBaseChain(t *testing.T) {
	// Restriction whose base is another named simpleType — exercises the
	// SimpleType-inner branch of resolveSimpleTypeBaseName.
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:simpleType name="codeType">
    <xs:restriction base="xs:string"/>
  </xs:simpleType>
  <xs:simpleType name="narrow">
    <xs:restriction base="codeType">
      <xs:maxLength value="3"/>
    </xs:restriction>
  </xs:simpleType>
  <xs:element name="root" type="narrow"/>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root>abc</root>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><root>abcd</root>`, xsd, "maxLength")
}

func TestSchemaImportMultipleNoLocation(t *testing.T) {
	// xs:import without a schemaLocation is accepted without a resolver.
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:import namespace="http://x"/>
  <xs:import namespace="http://y"/>
  <xs:element name="root" type="xs:string"/>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root>v</root>`, xsd)
}

func TestSchemaParseUnknownTopLevel(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:weird-thing name="oops"/>
  <xs:element name="root" type="xs:string"/>
</xs:schema>`
	err := schemaValidate(t, `<?xml version="1.1"?><root>v</root>`, xsd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported schema element")
}

func TestSchemaParseUnsupportedRedefine(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:redefine schemaLocation="x.xsd"/>
  <xs:element name="root" type="xs:string"/>
</xs:schema>`
	err := schemaValidate(t, `<?xml version="1.1"?><root>v</root>`, xsd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "redefine")
}

func TestSchemaParseUnsupportedOverride(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:override schemaLocation="x.xsd"/>
  <xs:element name="root" type="xs:string"/>
</xs:schema>`
	err := schemaValidate(t, `<?xml version="1.1"?><root>v</root>`, xsd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "override")
}

func TestSchemaResolverPassedNamespace(t *testing.T) {
	// Resolver should receive the namespace of the import.
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:import namespace="https://imported.example" schemaLocation="i.xsd"/>
  <xs:element name="root" type="xs:string"/>
</xs:schema>`
	gotNS := ""
	resolver := func(ns, _ string) ([]byte, error) {
		gotNS = ns
		return []byte(`<?xml version="1.0"?><xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" targetNamespace="https://imported.example"/>`), nil
	}
	require.NoError(t, ValidateWithSchemaResolver([]byte(`<?xml version="1.1"?><root>v</root>`), []byte(xsd), resolver))
	assert.Equal(t, "https://imported.example", gotNS)
}

func TestSchemaIncludeBadIncludedSchema(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:include schemaLocation="bad.xsd"/>
  <xs:element name="root" type="xs:string"/>
</xs:schema>`
	// included document is not a schema root
	resolver := func(_, _ string) ([]byte, error) {
		return []byte(`<?xml version="1.0"?><notschema/>`), nil
	}
	err := ValidateWithSchemaResolver([]byte(`<?xml version="1.1"?><root>v</root>`), []byte(xsd), resolver)
	require.Error(t, err)
}

func TestSchemaImportBadIncludedSchema(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:import namespace="http://other" schemaLocation="bad.xsd"/>
  <xs:element name="root" type="xs:string"/>
</xs:schema>`
	resolver := func(_, _ string) ([]byte, error) {
		return []byte(`<?xml version="1.0"?><notschema/>`), nil
	}
	err := ValidateWithSchemaResolver([]byte(`<?xml version="1.1"?><root>v</root>`), []byte(xsd), resolver)
	require.Error(t, err)
}

func TestSchemaImportMalformedSchemaContent(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:import namespace="http://other" schemaLocation="bad.xsd"/>
  <xs:element name="root" type="xs:string"/>
</xs:schema>`
	resolver := func(_, _ string) ([]byte, error) {
		return []byte(`<<<not xml`), nil
	}
	err := ValidateWithSchemaResolver([]byte(`<?xml version="1.1"?><root>v</root>`), []byte(xsd), resolver)
	require.Error(t, err)
}

func TestRefElementWithoutMatch(t *testing.T) {
	// element ref that resolves to a global decl
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="leaf" type="xs:integer"/>
  <xs:element name="root">
    <xs:complexType>
      <xs:sequence>
        <xs:element ref="leaf"/>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root><leaf>42</leaf></root>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><root><leaf>abc</leaf></root>`, xsd, "integer")
}

func TestSchemaInvalidElementMinOccurs(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="x" type="xs:string"/>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><root/>`, xsd, "requires at least")
}

func TestSchemaListItemError(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:simpleType name="intList">
    <xs:list itemType="xs:integer"/>
  </xs:simpleType>
  <xs:element name="root" type="intList"/>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><root>1 abc 3</root>`, xsd, "integer")
}

func TestSchemaAttrAnyAttrUnmatched(t *testing.T) {
	// anyAttribute with a constraint that doesn't match — should error.
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:anyAttribute namespace="http://allowed"/>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><root xmlns:bad="http://bad" bad:a="x"/>`, xsd, "unexpected attribute")
}
