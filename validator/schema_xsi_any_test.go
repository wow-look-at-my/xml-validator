package validator

import (
	"fmt"
	"testing"

	"github.com/wow-look-at-my/testify/assert"
	"github.com/wow-look-at-my/testify/require"
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

func TestSchemaAnyInAllGroup(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" targetNamespace="http://example.com/main" xmlns:tns="http://example.com/main" elementFormDefault="qualified">
  <xs:element name="event">
    <xs:complexType>
      <xs:all>
        <xs:element name="date" type="xs:string"/>
        <xs:element name="place" type="xs:string" minOccurs="0"/>
        <xs:any namespace="http://example.com/ext" processContents="lax" minOccurs="0" maxOccurs="unbounded"/>
      </xs:all>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><tns:event xmlns:tns="http://example.com/main" xmlns:ext="http://example.com/ext"><tns:date>2025-01-01</tns:date><ext:note>test</ext:note></tns:event>`, xsd)
	mustSchemaValid(t, `<?xml version="1.1"?><tns:event xmlns:tns="http://example.com/main" xmlns:ext="http://example.com/ext"><tns:date>2025-01-01</tns:date><ext:note>a</ext:note><ext:detail>b</ext:detail></tns:event>`, xsd)
	mustSchemaValid(t, `<?xml version="1.1"?><tns:event xmlns:tns="http://example.com/main"><tns:date>2025-01-01</tns:date></tns:event>`, xsd)
}

func TestSchemaAnyInAllGroupWrongNamespace(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" targetNamespace="http://example.com/main" xmlns:tns="http://example.com/main" elementFormDefault="qualified">
  <xs:element name="event">
    <xs:complexType>
      <xs:all>
        <xs:element name="date" type="xs:string"/>
        <xs:any namespace="http://example.com/ext" processContents="lax" minOccurs="0" maxOccurs="unbounded"/>
      </xs:all>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><tns:event xmlns:tns="http://example.com/main" xmlns:bad="http://wrong.example"><tns:date>2025-01-01</tns:date><bad:note>test</bad:note></tns:event>`, xsd, "unexpected element")
}

func TestSchemaAnyInAllGroupAnyNamespace(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:all>
        <xs:element name="name" type="xs:string"/>
        <xs:any namespace="##any" processContents="skip" minOccurs="0" maxOccurs="unbounded"/>
      </xs:all>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root xmlns:ext="http://x"><name>hi</name><ext:foo>bar</ext:foo></root>`, xsd)
}

func TestSchemaAnyNamespaceMatchInSequence(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:sequence>
        <xs:any namespace="http://example.com/ok" processContents="skip" minOccurs="0" maxOccurs="unbounded"/>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root xmlns:ok="http://example.com/ok"><ok:a>1</ok:a></root>`, xsd)
	mustSchemaValid(t, `<?xml version="1.1"?><root/>`, xsd)
}

func TestSchemaAnyInAllMinOccurs(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:all>
        <xs:element name="name" type="xs:string"/>
        <xs:any namespace="##any" processContents="skip" minOccurs="1" maxOccurs="unbounded"/>
      </xs:all>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><root><name>hi</name></root>`, xsd, "xs:any wildcard")
	mustSchemaValid(t, `<?xml version="1.1"?><root xmlns:e="http://x"><name>hi</name><e:extra>ok</e:extra></root>`, xsd)
}

func TestSchemaAnyNamespaceLocal(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:all>
        <xs:element name="name" type="xs:string"/>
        <xs:any namespace="##local" processContents="skip" minOccurs="0" maxOccurs="unbounded"/>
      </xs:all>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root><name>hi</name><extra>ok</extra></root>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><root xmlns:ns="http://x"><name>hi</name><ns:extra>no</ns:extra></root>`, xsd, "unexpected element")
}

func TestSchemaAnyNamespaceOther(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" targetNamespace="http://example.com/main" xmlns:tns="http://example.com/main" elementFormDefault="qualified">
  <xs:element name="root">
    <xs:complexType>
      <xs:all>
        <xs:element name="name" type="xs:string"/>
        <xs:any namespace="##other" processContents="skip" minOccurs="0" maxOccurs="unbounded"/>
      </xs:all>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><tns:root xmlns:tns="http://example.com/main" xmlns:ext="http://other"><tns:name>hi</tns:name><ext:extra>ok</ext:extra></tns:root>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><tns:root xmlns:tns="http://example.com/main"><tns:name>hi</tns:name><tns:extra>no</tns:extra></tns:root>`, xsd, "unexpected element")
}

func TestSchemaAnyNamespaceTargetNamespace(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" targetNamespace="http://example.com/main" xmlns:tns="http://example.com/main" elementFormDefault="qualified">
  <xs:element name="root">
    <xs:complexType>
      <xs:all>
        <xs:element name="name" type="xs:string"/>
        <xs:any namespace="##targetNamespace" processContents="skip" minOccurs="0" maxOccurs="unbounded"/>
      </xs:all>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><tns:root xmlns:tns="http://example.com/main"><tns:name>hi</tns:name><tns:extra>ok</tns:extra></tns:root>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><tns:root xmlns:tns="http://example.com/main" xmlns:ext="http://other"><tns:name>hi</tns:name><ext:extra>no</ext:extra></tns:root>`, xsd, "unexpected element")
}

func TestSchemaAnyNamespaceMultiToken(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:all>
        <xs:element name="name" type="xs:string"/>
        <xs:any namespace="##local http://example.com/ext" processContents="skip" minOccurs="0" maxOccurs="unbounded"/>
      </xs:all>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root><name>hi</name><local>ok</local></root>`, xsd)
	mustSchemaValid(t, `<?xml version="1.1"?><root xmlns:ext="http://example.com/ext"><name>hi</name><ext:extra>ok</ext:extra></root>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><root xmlns:bad="http://wrong"><name>hi</name><bad:x>no</bad:x></root>`, xsd, "unexpected element")
}

// TestSchemaAnyLaxValidatesImportedElement covers the XML Schema rule that
// processContents="lax" must validate matched elements against any element
// declaration the processor can find -- including ones reached via xs:import.
func TestSchemaAnyLaxValidatesImportedElement(t *testing.T) {
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
        <xs:any namespace="https://example.com/v" processContents="lax" minOccurs="0" maxOccurs="unbounded"/>
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

// TestSchemaAnyLaxAcceptsUndeclared covers the other half of "lax": when no
// declaration exists for the matched element, lax silently accepts it.
func TestSchemaAnyLaxAcceptsUndeclared(t *testing.T) {
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
        <xs:any namespace="https://example.com/v" processContents="lax" minOccurs="0" maxOccurs="unbounded"/>
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

	doc := `<?xml version="1.1"?><p:root xmlns:p="https://example.com/p" xmlns:v="https://example.com/v"><v:nodeclaration>data</v:nodeclaration></p:root>`
	require.NoError(t, ValidateWithSchemaResolver([]byte(doc), []byte(mainXSD), resolver))
}

// TestSchemaAnyStrictRejectsUndeclared verifies that processContents="strict"
// rejects any matched element that has no declaration the processor can find.
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
        <xs:any namespace="https://example.com/v" processContents="strict" minOccurs="0" maxOccurs="unbounded"/>
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
	assert.Contains(t, err.Error(), "strict")
}

// TestSchemaAnyStrictValidatesImported confirms strict also enforces the
// schema definition of a declared element matched through the wildcard.
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
        <xs:any namespace="https://example.com/v" processContents="strict" minOccurs="0" maxOccurs="unbounded"/>
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

// TestSchemaAnyLaxInAllGroupValidatesImported repeats the lax/import scenario
// under <xs:all>, where wildcard handling is inlined rather than via matchAny.
func TestSchemaAnyLaxInAllGroupValidatesImported(t *testing.T) {
	mainXSD := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
           xmlns:v="https://example.com/v"
           targetNamespace="https://example.com/p"
           xmlns:p="https://example.com/p"
           elementFormDefault="qualified">
  <xs:import namespace="https://example.com/v" schemaLocation="v.xsd"/>
  <xs:element name="root">
    <xs:complexType>
      <xs:all>
        <xs:any namespace="https://example.com/v" processContents="lax" minOccurs="0" maxOccurs="unbounded"/>
      </xs:all>
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

// TestSchemaAnySkipDoesNotValidate covers the contract that processContents=
// "skip" never looks up or validates against any declaration, even when one
// is reachable through an import.
func TestSchemaAnySkipDoesNotValidate(t *testing.T) {
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
        <xs:any namespace="https://example.com/v" processContents="skip" minOccurs="0" maxOccurs="unbounded"/>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	importedXSD := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" targetNamespace="https://example.com/v">
  <xs:simpleType name="codeType">
    <xs:restriction base="xs:string">
      <xs:enumeration value="A"/>
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

	doc := `<?xml version="1.1"?><p:root xmlns:p="https://example.com/p" xmlns:v="https://example.com/v"><v:code>X</v:code></p:root>`
	require.NoError(t, ValidateWithSchemaResolver([]byte(doc), []byte(mainXSD), resolver))
}
