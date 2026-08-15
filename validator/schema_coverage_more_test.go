package validator

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaGroupChoiceContent(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:group name="picker">
    <xs:choice>
      <xs:element name="a" type="xs:string"/>
      <xs:element name="b" type="xs:string"/>
    </xs:choice>
  </xs:group>
  <xs:element name="root">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="x" type="xs:string"/>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root><x>val</x></root>`, xsd)
}

func TestSchemaGroupAllContent(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:group name="bundle">
    <xs:all>
      <xs:element name="a" type="xs:string"/>
      <xs:element name="b" type="xs:string"/>
    </xs:all>
  </xs:group>
  <xs:element name="root">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="x" type="xs:string"/>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root><x>val</x></root>`, xsd)
}

func TestSchemaParticlesGroupRef(t *testing.T) {
	// parseParticles "group" ref branch is a no-op placeholder; just make sure
	// it doesn't break parsing.
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:group name="pair">
    <xs:sequence>
      <xs:element name="key" type="xs:string"/>
      <xs:element name="value" type="xs:string"/>
    </xs:sequence>
  </xs:group>
  <xs:element name="root">
    <xs:complexType>
      <xs:sequence>
        <xs:group ref="pair"/>
        <xs:element name="extra" type="xs:string" minOccurs="0"/>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root><extra>v</extra></root>`, xsd)
}

func TestSchemaParticlesNestedGroups(t *testing.T) {
	// parseParticles dispatches to parseSequence, parseChoice, parseAll for
	// nested compositors. Make sure all three nested branches parse.
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:sequence>
        <xs:sequence>
          <xs:element name="a" type="xs:string"/>
        </xs:sequence>
        <xs:choice>
          <xs:element name="b" type="xs:string"/>
          <xs:element name="c" type="xs:string"/>
        </xs:choice>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root><a>1</a><b>2</b></root>`, xsd)
}

func TestSchemaParticlesNestedAll(t *testing.T) {
	// parseParticles also needs to handle nested <xs:all>; an all-group can
	// appear inside a sequence at most once.
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:all>
        <xs:element name="a" type="xs:string"/>
        <xs:element name="b" type="xs:string"/>
      </xs:all>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root><a>1</a><b>2</b></root>`, xsd)
}

func TestSchemaComplexContentRestrictionWithAttr(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:complexType name="baseType">
    <xs:sequence>
      <xs:element name="a" type="xs:string" maxOccurs="unbounded"/>
    </xs:sequence>
  </xs:complexType>
  <xs:element name="root">
    <xs:complexType>
      <xs:complexContent>
        <xs:restriction base="baseType">
          <xs:choice>
            <xs:element name="a" type="xs:string"/>
            <xs:element name="b" type="xs:string"/>
          </xs:choice>
          <xs:attribute name="kind" type="xs:string"/>
        </xs:restriction>
      </xs:complexContent>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root kind="x"><a>val</a></root>`, xsd)
}

func TestSchemaComplexContentExtensionAll(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:complexType name="baseType">
    <xs:sequence>
      <xs:element name="a" type="xs:string"/>
    </xs:sequence>
  </xs:complexType>
  <xs:element name="root">
    <xs:complexType>
      <xs:complexContent>
        <xs:extension base="baseType">
          <xs:all>
            <xs:element name="b" type="xs:string"/>
            <xs:element name="c" type="xs:string"/>
          </xs:all>
        </xs:extension>
      </xs:complexContent>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root><b>v1</b><c>v2</c></root>`, xsd)
}

func TestSchemaComplexContentExtensionAttrGroup(t *testing.T) {
	// parseComplexContent extension branch with attributeGroup ref. The
	// extension must define the content directly because base content is not
	// inherited by this validator.
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:attributeGroup name="commonAttrs">
    <xs:attribute name="id" type="xs:string" use="required"/>
  </xs:attributeGroup>
  <xs:complexType name="baseType"/>
  <xs:element name="root">
    <xs:complexType>
      <xs:complexContent>
        <xs:extension base="baseType">
          <xs:sequence>
            <xs:element name="payload" type="xs:string"/>
          </xs:sequence>
          <xs:attributeGroup ref="commonAttrs"/>
        </xs:extension>
      </xs:complexContent>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root id="x"><payload>v</payload></root>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><root><payload>v</payload></root>`, xsd, "required attribute")
}

func TestSchemaComplexContentExtensionAnyAttribute(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:complexType name="baseType"/>
  <xs:element name="root">
    <xs:complexType>
      <xs:complexContent>
        <xs:extension base="baseType">
          <xs:sequence>
            <xs:element name="payload" type="xs:string"/>
          </xs:sequence>
          <xs:anyAttribute namespace="##any"/>
        </xs:extension>
      </xs:complexContent>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root xmlns:x="http://x" x:tag="ok"><payload>v</payload></root>`, xsd)
}

func TestSchemaComplexContentMixed(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:complexType name="baseType"/>
  <xs:element name="root">
    <xs:complexType>
      <xs:complexContent mixed="true">
        <xs:extension base="baseType">
          <xs:sequence>
            <xs:element name="b" type="xs:string"/>
          </xs:sequence>
        </xs:extension>
      </xs:complexContent>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root>text<b>v</b></root>`, xsd)
}

func TestSchemaSimpleContentRestrictionAttribute(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:simpleContent>
        <xs:restriction base="xs:string">
          <xs:minLength value="1"/>
          <xs:attribute name="kind" type="xs:string"/>
        </xs:restriction>
      </xs:simpleContent>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root kind="k">hello</root>`, xsd)
}

func TestSchemaSimpleContentRestrictionAnyAttribute(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:simpleContent>
        <xs:restriction base="xs:string">
          <xs:anyAttribute namespace="##any"/>
        </xs:restriction>
      </xs:simpleContent>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root xmlns:x="http://x" x:tag="t">hello</root>`, xsd)
}

func TestSchemaSimpleContentRestrictionAttrGroup(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:attributeGroup name="grp">
    <xs:attribute name="kind" type="xs:string" use="required"/>
  </xs:attributeGroup>
  <xs:element name="root">
    <xs:complexType>
      <xs:simpleContent>
        <xs:restriction base="xs:string">
          <xs:attributeGroup ref="grp"/>
        </xs:restriction>
      </xs:simpleContent>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root kind="k">hello</root>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><root>hello</root>`, xsd, "required attribute")
}

func TestSchemaSimpleContentExtensionAttrGroup(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:attributeGroup name="grp">
    <xs:attribute name="kind" type="xs:string" use="required"/>
  </xs:attributeGroup>
  <xs:element name="root">
    <xs:complexType>
      <xs:simpleContent>
        <xs:extension base="xs:string">
          <xs:attributeGroup ref="grp"/>
        </xs:extension>
      </xs:simpleContent>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root kind="k">hello</root>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><root>hello</root>`, xsd, "required attribute")
}

func TestSchemaSimpleContentExtensionAnyAttribute(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:simpleContent>
        <xs:extension base="xs:string">
          <xs:anyAttribute namespace="##any"/>
        </xs:extension>
      </xs:simpleContent>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root xmlns:x="http://x" x:tag="t">hello</root>`, xsd)
}

func TestSchemaSimpleContentRejectsChildElement(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:simpleContent>
        <xs:extension base="xs:string">
          <xs:attribute name="kind" type="xs:string"/>
        </xs:extension>
      </xs:simpleContent>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><root kind="k"><child/></root>`, xsd, "simpleContent")
}

func TestSchemaSimpleContentBuiltinTypeError(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:simpleContent>
        <xs:extension base="xs:integer">
          <xs:attribute name="kind" type="xs:string"/>
        </xs:extension>
      </xs:simpleContent>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><root kind="k">not-an-int</root>`, xsd, "integer")
}

func TestSchemaImportGroupCollision(t *testing.T) {
	mainXSD := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:import namespace="http://a" schemaLocation="a.xsd"/>
  <xs:import namespace="http://b" schemaLocation="b.xsd"/>
  <xs:element name="root" type="xs:string"/>
</xs:schema>`
	aXSD := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" targetNamespace="http://a">
  <xs:group name="dup">
    <xs:sequence>
      <xs:element name="x" type="xs:string"/>
    </xs:sequence>
  </xs:group>
</xs:schema>`
	bXSD := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" targetNamespace="http://b">
  <xs:group name="dup">
    <xs:choice>
      <xs:element name="y" type="xs:string"/>
    </xs:choice>
  </xs:group>
</xs:schema>`
	resolver := func(_, loc string) ([]byte, error) {
		switch loc {
		case "a.xsd":
			return []byte(aXSD), nil
		case "b.xsd":
			return []byte(bXSD), nil
		}
		return nil, fmt.Errorf("unexpected %q", loc)
	}
	err := ValidateWithSchemaResolver([]byte(`<?xml version="1.1"?><root>v</root>`), []byte(mainXSD), resolver)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `group "dup"`)
	assert.Contains(t, err.Error(), "more than once")
}

func TestSchemaImportAttrGroupCollision(t *testing.T) {
	mainXSD := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:import namespace="http://a" schemaLocation="a.xsd"/>
  <xs:import namespace="http://b" schemaLocation="b.xsd"/>
  <xs:element name="root" type="xs:string"/>
</xs:schema>`
	aXSD := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" targetNamespace="http://a">
  <xs:attributeGroup name="dup">
    <xs:attribute name="x" type="xs:string"/>
  </xs:attributeGroup>
</xs:schema>`
	bXSD := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" targetNamespace="http://b">
  <xs:attributeGroup name="dup">
    <xs:attribute name="y" type="xs:string"/>
  </xs:attributeGroup>
</xs:schema>`
	resolver := func(_, loc string) ([]byte, error) {
		switch loc {
		case "a.xsd":
			return []byte(aXSD), nil
		case "b.xsd":
			return []byte(bXSD), nil
		}
		return nil, fmt.Errorf("unexpected %q", loc)
	}
	err := ValidateWithSchemaResolver([]byte(`<?xml version="1.1"?><root>v</root>`), []byte(mainXSD), resolver)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `attributeGroup "dup"`)
	assert.Contains(t, err.Error(), "more than once")
}

// The same local name in two DIFFERENT namespaces is not a collision: one
// <params> per imported vocabulary is what namespaces are for.
func TestSchemaImportSameNameDifferentNamespaces(t *testing.T) {
	mainXSD := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
           xmlns:a="http://a" xmlns:b="http://b">
  <xs:import namespace="http://a" schemaLocation="a.xsd"/>
  <xs:import namespace="http://b" schemaLocation="b.xsd"/>
  <xs:element name="root">
    <xs:complexType>
      <xs:sequence>
        <xs:element ref="a:dup"/>
        <xs:element ref="b:dup"/>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	aXSD := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" targetNamespace="http://a">
  <xs:element name="dup" type="xs:string"/>
</xs:schema>`
	bXSD := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" targetNamespace="http://b">
  <xs:element name="dup" type="xs:int"/>
</xs:schema>`
	resolver := func(_, loc string) ([]byte, error) {
		switch loc {
		case "a.xsd":
			return []byte(aXSD), nil
		case "b.xsd":
			return []byte(bXSD), nil
		}
		return nil, fmt.Errorf("unexpected %q", loc)
	}
	doc := `<?xml version="1.1"?><root xmlns:a="http://a" xmlns:b="http://b"><a:dup>text</a:dup><b:dup>7</b:dup></root>`
	require.NoError(t, ValidateWithSchemaResolver([]byte(doc), []byte(mainXSD), resolver))

	// Each ref resolved to its OWN namespace's declaration, so the types are
	// not interchangeable.
	bad := `<?xml version="1.1"?><root xmlns:a="http://a" xmlns:b="http://b"><a:dup>text</a:dup><b:dup>text</b:dup></root>`
	err := ValidateWithSchemaResolver([]byte(bad), []byte(mainXSD), resolver)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "integer")
}

// The same name in the SAME namespace is still a collision.
func TestSchemaImportElementCollision(t *testing.T) {
	mainXSD := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:import namespace="http://a" schemaLocation="a.xsd"/>
  <xs:import namespace="http://a" schemaLocation="a2.xsd"/>
  <xs:element name="root" type="xs:string"/>
</xs:schema>`
	aXSD := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" targetNamespace="http://a">
  <xs:element name="dup" type="xs:string"/>
</xs:schema>`
	a2XSD := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" targetNamespace="http://a">
  <xs:element name="dup" type="xs:int"/>
</xs:schema>`
	resolver := func(_, loc string) ([]byte, error) {
		switch loc {
		case "a.xsd":
			return []byte(aXSD), nil
		case "a2.xsd":
			return []byte(a2XSD), nil
		}
		return nil, fmt.Errorf("unexpected %q", loc)
	}
	err := ValidateWithSchemaResolver([]byte(`<?xml version="1.1"?><root>v</root>`), []byte(mainXSD), resolver)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "more than once")
}

func TestSchemaIncludeNilDataSkips(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:include schemaLocation="skip.xsd"/>
  <xs:element name="root" type="xs:string"/>
</xs:schema>`
	resolver := func(_, _ string) ([]byte, error) {
		return nil, nil
	}
	err := ValidateWithSchemaResolver([]byte(`<?xml version="1.1"?><root>val</root>`), []byte(xsd), resolver)
	require.NoError(t, err)
}

func TestSchemaIncludeResolverError(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:include schemaLocation="missing.xsd"/>
  <xs:element name="root" type="xs:string"/>
</xs:schema>`
	resolver := func(_, _ string) ([]byte, error) {
		return nil, fmt.Errorf("nope")
	}
	err := ValidateWithSchemaResolver([]byte(`<?xml version="1.1"?><root>val</root>`), []byte(xsd), resolver)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing.xsd")
}

func TestSchemaIncludeMalformed(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:include schemaLocation="bad.xsd"/>
  <xs:element name="root" type="xs:string"/>
</xs:schema>`
	resolver := func(_, _ string) ([]byte, error) {
		return []byte("not actually xml"), nil
	}
	err := ValidateWithSchemaResolver([]byte(`<?xml version="1.1"?><root>val</root>`), []byte(xsd), resolver)
	require.Error(t, err)
}

func TestSchemaIncludeChameleon(t *testing.T) {
	// Chameleon include: included schema with no targetNamespace adopts the
	// including schema's namespace. Verify elements are reachable under that
	// namespace via wildcard lookup (lax wildcard with the parent's NS).
	mainXSD := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
           targetNamespace="https://example.com/main"
           xmlns:tns="https://example.com/main"
           elementFormDefault="qualified">
  <xs:include schemaLocation="common.xsd"/>
  <xs:element name="root">
    <xs:complexType>
      <xs:sequence>
        <xs:any namespace="https://example.com/main" minOccurs="0" maxOccurs="unbounded"/>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	commonXSD := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:simpleType name="codeType">
    <xs:restriction base="xs:string">
      <xs:enumeration value="A"/>
      <xs:enumeration value="B"/>
    </xs:restriction>
  </xs:simpleType>
  <xs:element name="code" type="codeType"/>
</xs:schema>`
	resolver := func(_, loc string) ([]byte, error) {
		if loc == "common.xsd" {
			return []byte(commonXSD), nil
		}
		return nil, fmt.Errorf("unexpected %q", loc)
	}

	good := `<?xml version="1.1"?><tns:root xmlns:tns="https://example.com/main"><tns:code>A</tns:code></tns:root>`
	require.NoError(t, ValidateWithSchemaResolver([]byte(good), []byte(mainXSD), resolver))

	bad := `<?xml version="1.1"?><tns:root xmlns:tns="https://example.com/main"><tns:code>X</tns:code></tns:root>`
	err := ValidateWithSchemaResolver([]byte(bad), []byte(mainXSD), resolver)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not one of the allowed values")
}

func TestSchemaSimpleTypeListNamedBase(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:simpleType name="nameList">
    <xs:list itemType="xs:string"/>
  </xs:simpleType>
  <xs:element name="root" type="nameList"/>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root>alpha beta gamma</root>`, xsd)
}

func TestSchemaSimpleTypeUnionInline(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:simpleType name="numOrBool">
    <xs:union memberTypes="xs:integer xs:boolean"/>
  </xs:simpleType>
  <xs:element name="root" type="numOrBool"/>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root>42</root>`, xsd)
	mustSchemaValid(t, `<?xml version="1.1"?><root>true</root>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><root>hello</root>`, xsd, "does not match any member type")
}

func TestSchemaAttributeFixedFailing(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:attribute name="version" type="xs:string" fixed="2.0"/>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><root version="1.0"/>`, xsd, "fixed value")
}

func TestSchemaAttrTypeBuiltinError(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:attribute name="n" type="xs:integer"/>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><root n="abc"/>`, xsd, "integer")
}

func TestSchemaAttrSimpleTypeBuiltinError(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:attribute name="n">
        <xs:simpleType>
          <xs:restriction base="xs:integer">
            <xs:minInclusive value="0"/>
          </xs:restriction>
        </xs:simpleType>
      </xs:attribute>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><root n="abc"/>`, xsd, "integer")
}

func TestSchemaBuiltinBase64Binary(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="payload" type="xs:base64Binary"/>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><payload>aGVsbG8=</payload>`, xsd)
	mustSchemaValid(t, `<?xml version="1.1"?><payload>SGVsbG8gV29ybGQh</payload>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><payload>not!base64</payload>`, xsd, "base64Binary")
	mustSchemaReject(t, `<?xml version="1.1"?><payload>abc</payload>`, xsd, "base64Binary")
}

func TestSchemaBuiltinIntegerRanges(t *testing.T) {
	for _, tc := range []struct {
		name    string
		typ     string
		good    string
		bad     string
		wantErr string
	}{
		{"byte", "xs:byte", "127", "128", "out of range"},
		{"short", "xs:short", "32767", "32768", "out of range"},
		{"int", "xs:int", "2147483647", "2147483648", "out of range"},
		{"unsignedByte", "xs:unsignedByte", "255", "256", "out of range"},
		{"unsignedShort", "xs:unsignedShort", "65535", "65536", "out of range"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			xsd := fmt.Sprintf(`<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="n" type="%s"/>
</xs:schema>`, tc.typ)
			mustSchemaValid(t, fmt.Sprintf(`<?xml version="1.1"?><n>%s</n>`, tc.good), xsd)
			mustSchemaReject(t, fmt.Sprintf(`<?xml version="1.1"?><n>%s</n>`, tc.bad), xsd, tc.wantErr)
		})
	}
}

func TestSchemaBuiltinFloatSpecials(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="n" type="xs:float"/>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><n>INF</n>`, xsd)
	mustSchemaValid(t, `<?xml version="1.1"?><n>-INF</n>`, xsd)
	mustSchemaValid(t, `<?xml version="1.1"?><n>NaN</n>`, xsd)
}

func TestSchemaCompareIntOpRange(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="n">
    <xs:simpleType>
      <xs:restriction base="xs:integer">
        <xs:minExclusive value="0"/>
        <xs:maxExclusive value="100"/>
      </xs:restriction>
    </xs:simpleType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><n>50</n>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><n>0</n>`, xsd, "must be > 0")
	mustSchemaReject(t, `<?xml version="1.1"?><n>100</n>`, xsd, "must be < 100")
}

func TestSchemaCompareIntOpInclusive(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="n">
    <xs:simpleType>
      <xs:restriction base="xs:integer">
        <xs:minInclusive value="0"/>
        <xs:maxInclusive value="100"/>
      </xs:restriction>
    </xs:simpleType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><n>0</n>`, xsd)
	mustSchemaValid(t, `<?xml version="1.1"?><n>100</n>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><n>-1</n>`, xsd, "must be >= 0")
	mustSchemaReject(t, `<?xml version="1.1"?><n>101</n>`, xsd, "must be <= 100")
}
