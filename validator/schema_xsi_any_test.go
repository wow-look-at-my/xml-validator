package validator

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaXSIAttributesAccepted(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root" type="xs:string"/>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:noNamespaceSchemaLocation="test.xsd">hello</root>`, xsd)
	mustSchemaValid(t, `<?xml version="1.1"?><root xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://example.com test.xsd">hello</root>`, xsd)
	mustSchemaValid(t, `<?xml version="1.1"?><root xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">hello</root>`, xsd)
}

// TestSchemaAnyInAllGroupWrongNamespace covers the namespace-constraint
// branch of xs:any: an element whose namespace does not match the wildcard
// is "unexpected" regardless of whether anything else is declared.
func TestSchemaAnyInAllGroupWrongNamespace(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" targetNamespace="http://example.com/main" xmlns:tns="http://example.com/main" elementFormDefault="qualified">
  <xs:element name="event">
    <xs:complexType>
      <xs:all>
        <xs:element name="date" type="xs:string"/>
        <xs:any namespace="http://example.com/ext" minOccurs="0" maxOccurs="unbounded"/>
      </xs:all>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><tns:event xmlns:tns="http://example.com/main" xmlns:bad="http://wrong.example"><tns:date>2025-01-01</tns:date><bad:note>test</bad:note></tns:event>`, xsd, "unexpected element")
}

// TestSchemaAnyNamespaceLocal exercises namespace="##local" by declaring the
// matched element as a global no-namespace element and validating against it.
func TestSchemaAnyNamespaceLocal(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="extra" type="xs:string"/>
  <xs:element name="root">
    <xs:complexType>
      <xs:all>
        <xs:element name="name" type="xs:string"/>
        <xs:any namespace="##local" minOccurs="0" maxOccurs="unbounded"/>
      </xs:all>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root><name>hi</name><extra>ok</extra></root>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><root xmlns:ns="http://x"><name>hi</name><ns:extra>no</ns:extra></root>`, xsd, "unexpected element")
}

// TestSchemaAnyNamespaceTargetNamespace exercises namespace="##targetNamespace"
// by declaring the matched element in the schema's target namespace.
func TestSchemaAnyNamespaceTargetNamespace(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" targetNamespace="http://example.com/main" xmlns:tns="http://example.com/main" elementFormDefault="qualified">
  <xs:element name="extra" type="xs:string"/>
  <xs:element name="root">
    <xs:complexType>
      <xs:all>
        <xs:element name="name" type="xs:string"/>
        <xs:any namespace="##targetNamespace" minOccurs="0" maxOccurs="unbounded"/>
      </xs:all>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><tns:root xmlns:tns="http://example.com/main"><tns:name>hi</tns:name><tns:extra>ok</tns:extra></tns:root>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><tns:root xmlns:tns="http://example.com/main" xmlns:ext="http://other"><tns:name>hi</tns:name><ext:extra>no</ext:extra></tns:root>`, xsd, "unexpected element")
}

// TestWildcardMatchesNS pins the namespace-constraint matcher for every
// supported form (##any, ##local, ##other, ##targetNamespace, explicit URI,
// multi-token). The matcher is shared by xs:any and xs:anyAttribute, so this
// keeps coverage even though the wildcard tests above only exercise a few
// constraint values directly.
func TestWildcardMatchesNS(t *testing.T) {
	sv := &schemaValidator{schema: &Schema{TargetNamespace: "http://example.com/tns"}}
	cases := []struct {
		constraint string
		ns         string
		want       bool
	}{
		{"##any", "http://x", true},
		{"##any", "", true},
		{"", "http://x", true},
		{"##local", "", true},
		{"##local", "http://x", false},
		{"##other", "http://other", true},
		{"##other", "http://example.com/tns", false},
		{"##other", "", false},
		{"##targetNamespace", "http://example.com/tns", true},
		{"##targetNamespace", "http://other", false},
		{"http://allowed", "http://allowed", true},
		{"http://allowed", "http://nope", false},
		{"##local http://allowed", "", true},
		{"##local http://allowed", "http://allowed", true},
		{"##local http://allowed", "http://nope", false},
		{"##targetNamespace http://other", "http://example.com/tns", true},
		{"##targetNamespace http://other", "http://other", true},
		{"##targetNamespace http://other", "http://elsewhere", false},
	}
	for _, c := range cases {
		got := sv.wildcardMatchesNS(c.constraint, c.ns)
		assert.Equal(t, c.want, got, "wildcardMatchesNS(%q, %q)", c.constraint, c.ns)
	}
}

// TestSchemaAnyStrictRejectsUndeclared confirms that xs:any (strict) rejects
// any matched element with no global declaration the validator can find.
func TestSchemaAnyStrictRejectsUndeclared(t *testing.T) {
	mainXSD := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
           xmlns:v="https://example.com/v"
           targetNamespace="https://example.com/p"
           xmlns:p="https://example.com/p"
           elementFormDefault="qualified">
  <xs:import namespace="https://example.com/v" schemaLocation="v.xsd"/>
  <xs:element name="root">
    <xs:complexType>
      <xs:sequence>
        <xs:any namespace="https://example.com/v" minOccurs="0" maxOccurs="unbounded"/>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	importedXSD := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" targetNamespace="https://example.com/v">
  <xs:element name="known" type="xs:string"/>
</xs:schema>`
	resolver := func(_, loc string) ([]byte, error) {
		if loc == "v.xsd" {
			return []byte(importedXSD), nil
		}
		return nil, fmt.Errorf("unexpected %q", loc)
	}

	good := `<?xml version="1.1"?><p:root xmlns:p="https://example.com/p" xmlns:v="https://example.com/v"><v:known>hello</v:known></p:root>`
	require.NoError(t, ValidateWithSchemaResolver([]byte(good), []byte(mainXSD), resolver))

	bad := `<?xml version="1.1"?><p:root xmlns:p="https://example.com/p" xmlns:v="https://example.com/v"><v:unknown>hello</v:unknown></p:root>`
	err := ValidateWithSchemaResolver([]byte(bad), []byte(mainXSD), resolver)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no declaration found")
}

// TestSchemaAnyStrictValidatesImported confirms strict enforces the schema
// definition of a declared element matched through the wildcard, including
// when the declaration lives in an imported schema.
func TestSchemaAnyStrictValidatesImported(t *testing.T) {
	mainXSD := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
           xmlns:v="https://example.com/v"
           targetNamespace="https://example.com/p"
           xmlns:p="https://example.com/p"
           elementFormDefault="qualified">
  <xs:import namespace="https://example.com/v" schemaLocation="v.xsd"/>
  <xs:element name="root">
    <xs:complexType>
      <xs:sequence>
        <xs:any namespace="https://example.com/v" minOccurs="0" maxOccurs="unbounded"/>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	importedXSD := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" targetNamespace="https://example.com/v">
  <xs:simpleType name="codeType">
    <xs:restriction base="xs:string">
      <xs:enumeration value="A"/>
      <xs:enumeration value="B"/>
    </xs:restriction>
  </xs:simpleType>
  <xs:element name="code" type="codeType"/>
</xs:schema>`
	resolver := func(_, loc string) ([]byte, error) {
		if loc == "v.xsd" {
			return []byte(importedXSD), nil
		}
		return nil, fmt.Errorf("unexpected %q", loc)
	}

	good := `<?xml version="1.1"?><p:root xmlns:p="https://example.com/p" xmlns:v="https://example.com/v"><v:code>A</v:code></p:root>`
	require.NoError(t, ValidateWithSchemaResolver([]byte(good), []byte(mainXSD), resolver))

	bad := `<?xml version="1.1"?><p:root xmlns:p="https://example.com/p" xmlns:v="https://example.com/v"><v:code>X</v:code></p:root>`
	err := ValidateWithSchemaResolver([]byte(bad), []byte(mainXSD), resolver)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not one of the allowed values")
}

// TestSchemaAnyValidatesImportedAttributeEnforced is the haynie repro:
// a strict xs:any wildcard targeting an imported namespace must enforce
// the imported element's declaration (including its attribute facets).
func TestSchemaAnyValidatesImportedAttributeEnforced(t *testing.T) {
	mainXSD := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
           xmlns:v="https://example.com/v"
           targetNamespace="https://example.com/p"
           xmlns:p="https://example.com/p"
           elementFormDefault="qualified">
  <xs:import namespace="https://example.com/v" schemaLocation="v.xsd"/>
  <xs:element name="parents">
    <xs:complexType>
      <xs:sequence>
        <xs:any namespace="https://example.com/v" minOccurs="0" maxOccurs="unbounded"/>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	importedXSD := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
           xmlns:v="https://example.com/v"
           targetNamespace="https://example.com/v"
           elementFormDefault="qualified">
  <xs:simpleType name="statusType">
    <xs:restriction base="xs:string">
      <xs:enumeration value="confirmed"/>
      <xs:enumeration value="unresolved"/>
    </xs:restriction>
  </xs:simpleType>
  <xs:element name="evidence">
    <xs:complexType>
      <xs:simpleContent>
        <xs:extension base="xs:string">
          <xs:attribute name="status" type="v:statusType" use="required"/>
          <xs:attribute name="cite"   type="xs:string"   use="required"/>
        </xs:extension>
      </xs:simpleContent>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	resolver := func(_, loc string) ([]byte, error) {
		if loc == "v.xsd" {
			return []byte(importedXSD), nil
		}
		return nil, fmt.Errorf("unexpected schemaLocation %q", loc)
	}

	good := `<?xml version="1.1"?><p:parents xmlns:p="https://example.com/p" xmlns:v="https://example.com/v"><v:evidence status="confirmed" cite="src">data</v:evidence></p:parents>`
	require.NoError(t, ValidateWithSchemaResolver([]byte(good), []byte(mainXSD), resolver))

	bad := `<?xml version="1.1"?><p:parents xmlns:p="https://example.com/p" xmlns:v="https://example.com/v"><v:evidence status="totally-fake-status" cite="src">data</v:evidence></p:parents>`
	err := ValidateWithSchemaResolver([]byte(bad), []byte(mainXSD), resolver)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not one of the allowed values")
}

func TestSchemaAnyRejectsSkip(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:sequence>
        <xs:any namespace="##any" processContents="skip" minOccurs="0" maxOccurs="unbounded"/>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	err := schemaValidate(t, `<?xml version="1.1"?><root/>`, xsd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `processContents="skip"`)
}

func TestSchemaAnyRejectsLax(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:sequence>
        <xs:any namespace="##any" processContents="lax" minOccurs="0" maxOccurs="unbounded"/>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	err := schemaValidate(t, `<?xml version="1.1"?><root/>`, xsd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `processContents="lax"`)
}

func TestSchemaAnyAttributeRejectsSkip(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:anyAttribute namespace="##any" processContents="skip"/>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	err := schemaValidate(t, `<?xml version="1.1"?><root/>`, xsd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `processContents="skip"`)
}

func TestSchemaAnyAttributeRejectsLax(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:anyAttribute namespace="##any" processContents="lax"/>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	err := schemaValidate(t, `<?xml version="1.1"?><root/>`, xsd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `processContents="lax"`)
}

func TestSchemaAnyRejectsUnknownProcessContents(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:sequence>
        <xs:any namespace="##any" processContents="loose" minOccurs="0" maxOccurs="unbounded"/>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	err := schemaValidate(t, `<?xml version="1.1"?><root/>`, xsd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid processContents")
}
