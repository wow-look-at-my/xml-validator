package validator

import (
	"strings"
	"testing"

	"github.com/wow-look-at-my/testify/assert"
	"github.com/wow-look-at-my/testify/require"
)

func schemaValidate(t *testing.T, xml, xsd string) error {
	t.Helper()
	return ValidateWithSchemaBytes([]byte(xml), []byte(xsd))
}

func mustSchemaValid(t *testing.T, xml, xsd string) {
	t.Helper()
	err := schemaValidate(t, xml, xsd)
	require.NoError(t, err)
}

func mustSchemaReject(t *testing.T, xml, xsd, wantSubstr string) {
	t.Helper()
	err := schemaValidate(t, xml, xsd)
	require.NotNil(t, err, "expected error containing %q", wantSubstr)
	require.Contains(t, err.Error(), wantSubstr)
}

const simpleXSD = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root" type="xs:string"/>
</xs:schema>`

func TestSchemaSimpleString(t *testing.T) {
	mustSchemaValid(t, `<?xml version="1.1"?><root>hello</root>`, simpleXSD)
}

func TestSchemaSimpleEmpty(t *testing.T) {
	mustSchemaValid(t, `<?xml version="1.1"?><root></root>`, simpleXSD)
}

func TestSchemaUndeclaredRoot(t *testing.T) {
	mustSchemaReject(t,
		`<?xml version="1.1"?><other>hello</other>`,
		simpleXSD,
		"not declared as a global element")
}

const typedXSD = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="name" type="xs:string"/>
        <xs:element name="age" type="xs:integer"/>
        <xs:element name="active" type="xs:boolean"/>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`

func TestSchemaSequenceValid(t *testing.T) {
	mustSchemaValid(t, `<?xml version="1.1"?>
<root>
  <name>Alice</name>
  <age>30</age>
  <active>true</active>
</root>`, typedXSD)
}

func TestSchemaSequenceMissingElement(t *testing.T) {
	mustSchemaReject(t, `<?xml version="1.1"?>
<root>
  <name>Alice</name>
</root>`, typedXSD, "requires at least 1")
}

func TestSchemaSequenceWrongOrder(t *testing.T) {
	mustSchemaReject(t, `<?xml version="1.1"?>
<root>
  <age>30</age>
  <name>Alice</name>
  <active>true</active>
</root>`, typedXSD, "requires at least 1")
}

func TestSchemaInvalidInteger(t *testing.T) {
	mustSchemaReject(t, `<?xml version="1.1"?>
<root>
  <name>Alice</name>
  <age>not-a-number</age>
  <active>true</active>
</root>`, typedXSD, "not a valid integer")
}

func TestSchemaInvalidBoolean(t *testing.T) {
	mustSchemaReject(t, `<?xml version="1.1"?>
<root>
  <name>Alice</name>
  <age>30</age>
  <active>maybe</active>
</root>`, typedXSD, "not a valid boolean")
}

const attrXSD = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="item">
    <xs:complexType>
      <xs:attribute name="id" type="xs:integer" use="required"/>
      <xs:attribute name="label" type="xs:string" use="optional"/>
    </xs:complexType>
  </xs:element>
</xs:schema>`

func TestSchemaAttributeValid(t *testing.T) {
	mustSchemaValid(t, `<?xml version="1.1"?><item id="42" label="test"/>`, attrXSD)
}

func TestSchemaAttributeRequiredMissing(t *testing.T) {
	mustSchemaReject(t, `<?xml version="1.1"?><item label="test"/>`, attrXSD, "required attribute")
}

func TestSchemaAttributeUnexpected(t *testing.T) {
	mustSchemaReject(t, `<?xml version="1.1"?><item id="1" unknown="x"/>`, attrXSD, "unexpected attribute")
}

func TestSchemaAttributeInvalidType(t *testing.T) {
	mustSchemaReject(t, `<?xml version="1.1"?><item id="abc"/>`, attrXSD, "not a valid integer")
}

const enumXSD = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="color">
    <xs:simpleType>
      <xs:restriction base="xs:string">
        <xs:enumeration value="red"/>
        <xs:enumeration value="green"/>
        <xs:enumeration value="blue"/>
      </xs:restriction>
    </xs:simpleType>
  </xs:element>
</xs:schema>`

func TestSchemaEnumerationValid(t *testing.T) {
	mustSchemaValid(t, `<?xml version="1.1"?><color>red</color>`, enumXSD)
}

func TestSchemaEnumerationInvalid(t *testing.T) {
	mustSchemaReject(t, `<?xml version="1.1"?><color>yellow</color>`, enumXSD, "not one of the allowed values")
}

const occursXSD = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="list">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="item" type="xs:string" minOccurs="1" maxOccurs="unbounded"/>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`

func TestSchemaOccursValid(t *testing.T) {
	mustSchemaValid(t, `<?xml version="1.1"?>
<list>
  <item>a</item>
  <item>b</item>
  <item>c</item>
</list>`, occursXSD)
}

func TestSchemaOccursEmpty(t *testing.T) {
	mustSchemaReject(t, `<?xml version="1.1"?><list></list>`, occursXSD, "requires at least 1")
}

const choiceXSD = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="shape">
    <xs:complexType>
      <xs:choice>
        <xs:element name="circle" type="xs:string"/>
        <xs:element name="square" type="xs:string"/>
      </xs:choice>
    </xs:complexType>
  </xs:element>
</xs:schema>`

func TestSchemaChoiceValid(t *testing.T) {
	mustSchemaValid(t, `<?xml version="1.1"?><shape><circle>r=5</circle></shape>`, choiceXSD)
	mustSchemaValid(t, `<?xml version="1.1"?><shape><square>s=10</square></shape>`, choiceXSD)
}

func TestSchemaChoiceInvalid(t *testing.T) {
	mustSchemaReject(t, `<?xml version="1.1"?><shape><triangle>bad</triangle></shape>`, choiceXSD, "not allowed here")
}

const namedTypeXSD = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:simpleType name="percentType">
    <xs:restriction base="xs:integer">
      <xs:minInclusive value="0"/>
      <xs:maxInclusive value="100"/>
    </xs:restriction>
  </xs:simpleType>
  <xs:element name="score" type="percentType"/>
</xs:schema>`

func TestSchemaNamedTypeValid(t *testing.T) {
	mustSchemaValid(t, `<?xml version="1.1"?><score>85</score>`, namedTypeXSD)
}

func TestSchemaNamedTypeTooLow(t *testing.T) {
	mustSchemaReject(t, `<?xml version="1.1"?><score>-1</score>`, namedTypeXSD, "must be >= 0")
}

func TestSchemaNamedTypeTooHigh(t *testing.T) {
	mustSchemaReject(t, `<?xml version="1.1"?><score>101</score>`, namedTypeXSD, "must be <= 100")
}

const patternXSD = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="zip">
    <xs:simpleType>
      <xs:restriction base="xs:string">
        <xs:pattern value="\d{5}"/>
      </xs:restriction>
    </xs:simpleType>
  </xs:element>
</xs:schema>`

func TestSchemaPatternValid(t *testing.T) {
	mustSchemaValid(t, `<?xml version="1.1"?><zip>12345</zip>`, patternXSD)
}

func TestSchemaPatternInvalid(t *testing.T) {
	mustSchemaReject(t, `<?xml version="1.1"?><zip>1234</zip>`, patternXSD, "does not match pattern")
}

const allXSD = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="config">
    <xs:complexType>
      <xs:all>
        <xs:element name="host" type="xs:string"/>
        <xs:element name="port" type="xs:integer"/>
      </xs:all>
    </xs:complexType>
  </xs:element>
</xs:schema>`

func TestSchemaAllAnyOrder(t *testing.T) {
	mustSchemaValid(t, `<?xml version="1.1"?><config><host>localhost</host><port>8080</port></config>`, allXSD)
	mustSchemaValid(t, `<?xml version="1.1"?><config><port>8080</port><host>localhost</host></config>`, allXSD)
}

func TestSchemaAllMissing(t *testing.T) {
	mustSchemaReject(t, `<?xml version="1.1"?><config><host>localhost</host></config>`, allXSD, "requires at least 1")
}

const nestedXSD = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="library">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="book" maxOccurs="unbounded">
          <xs:complexType>
            <xs:sequence>
              <xs:element name="title" type="xs:string"/>
              <xs:element name="author" type="xs:string"/>
            </xs:sequence>
            <xs:attribute name="isbn" type="xs:string" use="required"/>
          </xs:complexType>
        </xs:element>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`

func TestSchemaNestedValid(t *testing.T) {
	mustSchemaValid(t, `<?xml version="1.1"?>
<library>
  <book isbn="978-0-123">
    <title>Test Book</title>
    <author>Author Name</author>
  </book>
  <book isbn="978-0-456">
    <title>Another Book</title>
    <author>Another Author</author>
  </book>
</library>`, nestedXSD)
}

func TestSchemaNestedMissingAttr(t *testing.T) {
	mustSchemaReject(t, `<?xml version="1.1"?>
<library>
  <book>
    <title>No ISBN</title>
    <author>Author</author>
  </book>
</library>`, nestedXSD, "required attribute")
}

func TestSchemaFixedValue(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="version" type="xs:string" fixed="1.0"/>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><version>1.0</version>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><version>2.0</version>`, xsd, "fixed value")
}

func TestSchemaMinMaxLength(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="code">
    <xs:simpleType>
      <xs:restriction base="xs:string">
        <xs:minLength value="2"/>
        <xs:maxLength value="5"/>
      </xs:restriction>
    </xs:simpleType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><code>AB</code>`, xsd)
	mustSchemaValid(t, `<?xml version="1.1"?><code>ABCDE</code>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><code>A</code>`, xsd, "less than minLength")
	mustSchemaReject(t, `<?xml version="1.1"?><code>ABCDEF</code>`, xsd, "exceeds maxLength")
}

func TestSchemaSimpleContentExtension(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="price">
    <xs:complexType>
      <xs:simpleContent>
        <xs:extension base="xs:decimal">
          <xs:attribute name="currency" type="xs:string" use="required"/>
        </xs:extension>
      </xs:simpleContent>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><price currency="USD">19.99</price>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><price currency="USD">abc</price>`, xsd, "not a valid decimal")
	mustSchemaReject(t, `<?xml version="1.1"?><price>19.99</price>`, xsd, "required attribute")
}

func TestSchemaOptionalElement(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="person">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="name" type="xs:string"/>
        <xs:element name="nickname" type="xs:string" minOccurs="0"/>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><person><name>Alice</name></person>`, xsd)
	mustSchemaValid(t, `<?xml version="1.1"?><person><name>Alice</name><nickname>Al</nickname></person>`, xsd)
}

func TestSchemaEmptyComplexType(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="br">
    <xs:complexType/>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><br/>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><br><child/></br>`, xsd, "expects no child elements")
}

func TestSchemaDateType(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="birthday" type="xs:date"/>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><birthday>2024-01-15</birthday>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><birthday>not-a-date</birthday>`, xsd, "not a valid date")
}

func TestSchemaDecimalType(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="amount" type="xs:decimal"/>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><amount>123.45</amount>`, xsd)
	mustSchemaValid(t, `<?xml version="1.1"?><amount>-0.5</amount>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><amount>abc</amount>`, xsd, "not a valid decimal")
}

func TestSchemaDoubleType(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="val" type="xs:double"/>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><val>3.14e10</val>`, xsd)
	mustSchemaValid(t, `<?xml version="1.1"?><val>INF</val>`, xsd)
	mustSchemaValid(t, `<?xml version="1.1"?><val>NaN</val>`, xsd)
}

func TestSchemaChildWithSimpleType(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="data">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="count" type="xs:integer"/>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><data><count><nested/></count></data>`, xsd, "simple type")
}

// --- anyAttribute tests ---

const anyAttrXSD = `<?xml version="1.1" encoding="UTF-8"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="item" minOccurs="0" maxOccurs="unbounded">
          <xs:complexType mixed="true">
            <xs:attribute name="id" type="xs:string"/>
            <xs:anyAttribute namespace="https://example.com/v" processContents="lax"/>
          </xs:complexType>
        </xs:element>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`

func TestAnyAttributeExactNamespace(t *testing.T) {
	mustSchemaValid(t, `<?xml version="1.1"?>
<root xmlns:v="https://example.com/v">
  <item id="x" v:status="ok">hello</item>
</root>`, anyAttrXSD)
}

func TestAnyAttributeWrongNamespaceRejected(t *testing.T) {
	mustSchemaReject(t, `<?xml version="1.1"?>
<root xmlns:v="https://other.com/v">
  <item id="x" v:status="ok">hello</item>
</root>`, anyAttrXSD, "unexpected attribute")
}

func TestAnyAttributeNoNamespaceRejected(t *testing.T) {
	mustSchemaReject(t, `<?xml version="1.1"?>
<root>
  <item id="x" unknown="bad">hello</item>
</root>`, anyAttrXSD, "unexpected attribute")
}

func TestAnyAttributeAny(t *testing.T) {
	xsd := `<?xml version="1.1"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="r">
    <xs:complexType>
      <xs:anyAttribute namespace="##any" processContents="skip"/>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?>
<r xmlns:x="http://x.com" x:foo="1" x:bar="2"/>`, xsd)
	mustSchemaValid(t, `<?xml version="1.1"?><r local="ok"/>`, xsd)
}

func TestAnyAttributeOther(t *testing.T) {
	xsd := `<?xml version="1.1"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
           targetNamespace="http://mine.com">
  <xs:element name="r">
    <xs:complexType>
      <xs:anyAttribute namespace="##other" processContents="lax"/>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?>
<r xmlns:x="http://x.com" x:foo="1"/>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><r local="bad"/>`, xsd, "unexpected attribute")
}

func TestAnyAttributeMultipleNamespaces(t *testing.T) {
	xsd := `<?xml version="1.1"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="r">
    <xs:complexType>
      <xs:anyAttribute namespace="http://a.com http://b.com" processContents="skip"/>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?>
<r xmlns:a="http://a.com" a:x="1"/>`, xsd)
	mustSchemaValid(t, `<?xml version="1.1"?>
<r xmlns:b="http://b.com" b:y="2"/>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?>
<r xmlns:c="http://c.com" c:z="3"/>`, xsd, "unexpected attribute")
}

func TestAnyAttributeDefaultProcessContents(t *testing.T) {
	xsd := `<?xml version="1.1"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="r">
    <xs:complexType>
      <xs:anyAttribute namespace="##any"/>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?>
<r xmlns:x="http://x.com" x:a="1"/>`, xsd)
}

func TestAnyAttributeDeclaredAttrStillValidated(t *testing.T) {
	xsd := `<?xml version="1.1"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="r">
    <xs:complexType>
      <xs:attribute name="id" type="xs:integer" use="required"/>
      <xs:anyAttribute namespace="##any" processContents="skip"/>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?>
<r id="42" xmlns:x="http://x.com" x:extra="yes"/>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?>
<r id="notint" xmlns:x="http://x.com" x:extra="yes"/>`, xsd, "not a valid integer")
	mustSchemaReject(t, `<?xml version="1.1"?>
<r xmlns:x="http://x.com" x:extra="yes"/>`, xsd, "required attribute")
}

func TestAnyAttributeNamespacedSameLocalName(t *testing.T) {
	xsd := `<?xml version="1.1"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="r">
    <xs:complexType>
      <xs:attribute name="id" type="xs:integer"/>
      <xs:anyAttribute namespace="http://x.com" processContents="skip"/>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?>
<r id="42" xmlns:x="http://x.com" x:id="not-an-int"/>`, xsd)
}

func TestAnyAttributeNamespacedNoWildcardRejected(t *testing.T) {
	xsd := `<?xml version="1.1"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="r">
    <xs:complexType>
      <xs:attribute name="id" type="xs:integer"/>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?>
<r id="42" xmlns:x="http://x.com" x:id="hello"/>`, xsd, "unexpected attribute")
}

func TestAnyAttributeViaAttrGroup(t *testing.T) {
	xsd := `<?xml version="1.1"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:attributeGroup name="extras">
    <xs:attribute name="id" type="xs:string"/>
    <xs:anyAttribute namespace="##any" processContents="skip"/>
  </xs:attributeGroup>
  <xs:element name="r">
    <xs:complexType>
      <xs:attributeGroup ref="extras"/>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?>
<r id="hello" xmlns:x="http://x.com" x:foo="bar"/>`, xsd)
}

func TestAnyAttributeViaAttrGroupNamespaceFiltered(t *testing.T) {
	xsd := `<?xml version="1.1"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:attributeGroup name="extras">
    <xs:anyAttribute namespace="http://allowed.com" processContents="skip"/>
  </xs:attributeGroup>
  <xs:element name="r">
    <xs:complexType>
      <xs:attributeGroup ref="extras"/>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?>
<r xmlns:a="http://allowed.com" a:x="1"/>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?>
<r xmlns:b="http://other.com" b:x="1"/>`, xsd, "unexpected attribute")
}

func TestAnyAttributeViaAttrGroupInComplexContent(t *testing.T) {
	xsd := `<?xml version="1.1"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:attributeGroup name="extras">
    <xs:anyAttribute namespace="##any" processContents="skip"/>
  </xs:attributeGroup>
  <xs:complexType name="base"/>
  <xs:element name="r">
    <xs:complexType>
      <xs:complexContent>
        <xs:extension base="base">
          <xs:attributeGroup ref="extras"/>
        </xs:extension>
      </xs:complexContent>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?>
<r xmlns:x="http://x.com" x:foo="bar"/>`, xsd)
}

func TestAnyAttributeViaAttrGroupInSimpleContent(t *testing.T) {
	xsd := `<?xml version="1.1"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:attributeGroup name="extras">
    <xs:anyAttribute namespace="##any" processContents="skip"/>
  </xs:attributeGroup>
  <xs:element name="r">
    <xs:complexType>
      <xs:simpleContent>
        <xs:extension base="xs:string">
          <xs:attributeGroup ref="extras"/>
        </xs:extension>
      </xs:simpleContent>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?>
<r xmlns:x="http://x.com" x:note="yes">hello</r>`, xsd)
}

func TestAnyAttributeViaAttrGroupReusedSchema(t *testing.T) {
	xsd := `<?xml version="1.1"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:attributeGroup name="extras">
    <xs:attribute name="id" type="xs:string"/>
    <xs:anyAttribute namespace="##any" processContents="skip"/>
  </xs:attributeGroup>
  <xs:element name="r">
    <xs:complexType>
      <xs:attributeGroup ref="extras"/>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	schemaDoc, err := ParseTree(strings.NewReader(xsd))
	require.NoError(t, err)
	schema, err := ParseSchema(schemaDoc)
	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		doc, err := ParseTree(strings.NewReader(`<?xml version="1.1"?>
<r id="hello" xmlns:x="http://x.com" x:foo="bar"/>`))
		require.NoError(t, err)
		err = ValidateSchema(doc, schema)
		require.NoError(t, err, "validation run %d failed", i+1)
	}
}

// --- Tree parser tests ---

func TestTreeParserBasic(t *testing.T) {
	doc, err := ParseTree(strings.NewReader(`<?xml version="1.0"?><root attr="val"><child>text</child></root>`))
	require.NoError(t, err)
	assert.Equal(t, "root", doc.Root.Local)
	assert.Equal(t, 1, len(doc.Root.Attrs))
	assert.Equal(t, "val", doc.Root.Attrs[0].Value)
	children := doc.Root.ChildElements()
	require.Equal(t, 1, len(children))
	assert.Equal(t, "child", children[0].Local)
	assert.Equal(t, "text", children[0].TextContent())
}

func TestTreeParserNamespaces(t *testing.T) {
	doc, err := ParseTree(strings.NewReader(`<root xmlns:ns="http://example.com"><ns:child/></root>`))
	require.NoError(t, err)
	children := doc.Root.ChildElements()
	require.Equal(t, 1, len(children))
	assert.Equal(t, "child", children[0].Local)
	assert.Equal(t, "ns", children[0].Prefix)
	assert.Equal(t, "http://example.com", children[0].Namespace)
}

func TestTreeParserCDATA(t *testing.T) {
	doc, err := ParseTree(strings.NewReader(`<r><![CDATA[hello <world>]]></r>`))
	require.NoError(t, err)
	assert.Equal(t, "hello <world>", doc.Root.TextContent())
}

func TestTreeParserEntityRefs(t *testing.T) {
	doc, err := ParseTree(strings.NewReader(`<r>&amp;&lt;&gt;</r>`))
	require.NoError(t, err)
	assert.Equal(t, "&<>", doc.Root.TextContent())
}

func TestTreeParserXML10Accepted(t *testing.T) {
	_, err := ParseTree(strings.NewReader(`<?xml version="1.0"?><root/>`))
	require.NoError(t, err)
}
