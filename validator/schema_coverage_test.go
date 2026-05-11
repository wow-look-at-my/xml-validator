package validator

import (
	"strings"
	"testing"

	"github.com/wow-look-at-my/testify/assert"
	"github.com/wow-look-at-my/testify/require"
)

// --- Builtin type coverage ---

func TestBuiltinNormalizedString(t *testing.T) {
	assert.NoError(t, validateBuiltinValue("normalizedString", "hello world"))
	assert.Error(t, validateBuiltinValue("normalizedString", "hello\tworld"))
	assert.Error(t, validateBuiltinValue("normalizedString", "hello\nworld"))
}

func TestBuiltinToken(t *testing.T) {
	assert.NoError(t, validateBuiltinValue("token", "hello world"))
	assert.Error(t, validateBuiltinValue("token", " leading"))
	assert.Error(t, validateBuiltinValue("token", "double  space"))
}

func TestBuiltinBoolean(t *testing.T) {
	assert.NoError(t, validateBuiltinValue("boolean", "true"))
	assert.NoError(t, validateBuiltinValue("boolean", "false"))
	assert.NoError(t, validateBuiltinValue("boolean", "1"))
	assert.NoError(t, validateBuiltinValue("boolean", "0"))
	assert.Error(t, validateBuiltinValue("boolean", "yes"))
}

func TestBuiltinDecimal(t *testing.T) {
	assert.NoError(t, validateBuiltinValue("decimal", "123"))
	assert.NoError(t, validateBuiltinValue("decimal", "-45.67"))
	assert.NoError(t, validateBuiltinValue("decimal", "+0.1"))
	assert.Error(t, validateBuiltinValue("decimal", "abc"))
	assert.Error(t, validateBuiltinValue("decimal", ""))
	assert.Error(t, validateBuiltinValue("decimal", "."))
	assert.Error(t, validateBuiltinValue("decimal", "1.2.3"))
}

func TestBuiltinInteger(t *testing.T) {
	assert.NoError(t, validateBuiltinValue("integer", "42"))
	assert.NoError(t, validateBuiltinValue("integer", "-100"))
	assert.NoError(t, validateBuiltinValue("integer", "+0"))
	assert.Error(t, validateBuiltinValue("integer", "12.5"))
	assert.Error(t, validateBuiltinValue("integer", ""))
	assert.Error(t, validateBuiltinValue("integer", "+"))
}

func TestBuiltinNonNegativeInteger(t *testing.T) {
	assert.NoError(t, validateBuiltinValue("nonNegativeInteger", "0"))
	assert.NoError(t, validateBuiltinValue("nonNegativeInteger", "100"))
	assert.Error(t, validateBuiltinValue("nonNegativeInteger", "-1"))
}

func TestBuiltinPositiveInteger(t *testing.T) {
	assert.NoError(t, validateBuiltinValue("positiveInteger", "1"))
	assert.Error(t, validateBuiltinValue("positiveInteger", "0"))
	assert.Error(t, validateBuiltinValue("positiveInteger", "-5"))
}

func TestBuiltinNonPositiveInteger(t *testing.T) {
	assert.NoError(t, validateBuiltinValue("nonPositiveInteger", "0"))
	assert.NoError(t, validateBuiltinValue("nonPositiveInteger", "-5"))
	assert.Error(t, validateBuiltinValue("nonPositiveInteger", "1"))
}

func TestBuiltinNegativeInteger(t *testing.T) {
	assert.NoError(t, validateBuiltinValue("negativeInteger", "-1"))
	assert.Error(t, validateBuiltinValue("negativeInteger", "0"))
	assert.Error(t, validateBuiltinValue("negativeInteger", "-0"))
	assert.Error(t, validateBuiltinValue("negativeInteger", "5"))
}

func TestBuiltinIntRanges(t *testing.T) {
	assert.NoError(t, validateBuiltinValue("long", "9223372036854775807"))
	assert.NoError(t, validateBuiltinValue("int", "2147483647"))
	assert.Error(t, validateBuiltinValue("int", "2147483648"))
	assert.NoError(t, validateBuiltinValue("short", "32767"))
	assert.Error(t, validateBuiltinValue("short", "32768"))
	assert.NoError(t, validateBuiltinValue("byte", "127"))
	assert.Error(t, validateBuiltinValue("byte", "128"))
}

func TestBuiltinUintRanges(t *testing.T) {
	assert.NoError(t, validateBuiltinValue("unsignedLong", "0"))
	assert.NoError(t, validateBuiltinValue("unsignedInt", "4294967295"))
	assert.Error(t, validateBuiltinValue("unsignedInt", "4294967296"))
	assert.NoError(t, validateBuiltinValue("unsignedShort", "65535"))
	assert.Error(t, validateBuiltinValue("unsignedShort", "65536"))
	assert.NoError(t, validateBuiltinValue("unsignedByte", "255"))
	assert.Error(t, validateBuiltinValue("unsignedByte", "256"))
	assert.Error(t, validateBuiltinValue("unsignedByte", "-1"))
}

func TestBuiltinFloat(t *testing.T) {
	assert.NoError(t, validateBuiltinValue("float", "1.5"))
	assert.NoError(t, validateBuiltinValue("float", "INF"))
	assert.NoError(t, validateBuiltinValue("float", "-INF"))
	assert.NoError(t, validateBuiltinValue("float", "NaN"))
	assert.Error(t, validateBuiltinValue("float", "abc"))
}

func TestBuiltinDouble(t *testing.T) {
	assert.NoError(t, validateBuiltinValue("double", "1.5e100"))
	assert.NoError(t, validateBuiltinValue("double", "INF"))
	assert.Error(t, validateBuiltinValue("double", "not-a-number"))
}

func TestBuiltinDate(t *testing.T) {
	assert.NoError(t, validateBuiltinValue("date", "2024-01-15"))
	assert.NoError(t, validateBuiltinValue("date", "2024-01-15Z"))
	assert.NoError(t, validateBuiltinValue("date", "2024-01-15+05:00"))
	assert.Error(t, validateBuiltinValue("date", "2024/01/15"))
	assert.Error(t, validateBuiltinValue("date", "not-a-date"))
}

func TestBuiltinDateTime(t *testing.T) {
	assert.NoError(t, validateBuiltinValue("dateTime", "2024-01-15T10:30:00"))
	assert.NoError(t, validateBuiltinValue("dateTime", "2024-01-15T10:30:00Z"))
	assert.NoError(t, validateBuiltinValue("dateTime", "2024-01-15T10:30:00.123"))
	assert.Error(t, validateBuiltinValue("dateTime", "2024-01-15"))
}

func TestBuiltinTime(t *testing.T) {
	assert.NoError(t, validateBuiltinValue("time", "10:30:00"))
	assert.NoError(t, validateBuiltinValue("time", "10:30:00Z"))
	assert.Error(t, validateBuiltinValue("time", "10:30"))
}

func TestBuiltinDuration(t *testing.T) {
	assert.NoError(t, validateBuiltinValue("duration", "P1Y2M3D"))
	assert.NoError(t, validateBuiltinValue("duration", "PT1H30M"))
	assert.NoError(t, validateBuiltinValue("duration", "-P1D"))
	assert.Error(t, validateBuiltinValue("duration", "P"))
	assert.Error(t, validateBuiltinValue("duration", "1Y"))
}

func TestBuiltinGYear(t *testing.T) {
	assert.NoError(t, validateBuiltinValue("gYear", "2024"))
	assert.NoError(t, validateBuiltinValue("gYear", "2024Z"))
	assert.Error(t, validateBuiltinValue("gYear", "24"))
}

func TestBuiltinGMonth(t *testing.T) {
	assert.NoError(t, validateBuiltinValue("gMonth", "--01"))
	assert.Error(t, validateBuiltinValue("gMonth", "01"))
}

func TestBuiltinGDay(t *testing.T) {
	assert.NoError(t, validateBuiltinValue("gDay", "---15"))
	assert.Error(t, validateBuiltinValue("gDay", "15"))
}

func TestBuiltinGYearMonth(t *testing.T) {
	assert.NoError(t, validateBuiltinValue("gYearMonth", "2024-01"))
	assert.Error(t, validateBuiltinValue("gYearMonth", "2024"))
}

func TestBuiltinGMonthDay(t *testing.T) {
	assert.NoError(t, validateBuiltinValue("gMonthDay", "--01-15"))
	assert.Error(t, validateBuiltinValue("gMonthDay", "01-15"))
}

func TestBuiltinHexBinary(t *testing.T) {
	assert.NoError(t, validateBuiltinValue("hexBinary", "48656C6C6F"))
	assert.NoError(t, validateBuiltinValue("hexBinary", ""))
	assert.Error(t, validateBuiltinValue("hexBinary", "GGG"))
}

func TestBuiltinBase64Binary(t *testing.T) {
	assert.NoError(t, validateBuiltinValue("base64Binary", "SGVsbG8="))
	assert.Error(t, validateBuiltinValue("base64Binary", "!!!"))
}

func TestBuiltinQName(t *testing.T) {
	assert.NoError(t, validateBuiltinValue("QName", "xs:string"))
	assert.NoError(t, validateBuiltinValue("QName", "localname"))
	assert.Error(t, validateBuiltinValue("QName", ""))
}

func TestBuiltinNCName(t *testing.T) {
	assert.NoError(t, validateBuiltinValue("NCName", "valid"))
	assert.Error(t, validateBuiltinValue("NCName", ""))
	assert.Error(t, validateBuiltinValue("NCName", "1invalid"))
}

func TestBuiltinName(t *testing.T) {
	assert.NoError(t, validateBuiltinValue("Name", "valid:name"))
	assert.Error(t, validateBuiltinValue("Name", ""))
}

func TestBuiltinNMTOKEN(t *testing.T) {
	assert.NoError(t, validateBuiltinValue("NMTOKEN", "token-1"))
	assert.Error(t, validateBuiltinValue("NMTOKEN", ""))
}

func TestBuiltinLanguage(t *testing.T) {
	assert.NoError(t, validateBuiltinValue("language", "en"))
	assert.NoError(t, validateBuiltinValue("language", "en-US"))
	assert.Error(t, validateBuiltinValue("language", ""))
}

func TestBuiltinIDREFS(t *testing.T) {
	assert.NoError(t, validateBuiltinValue("IDREFS", "id1 id2 id3"))
	assert.Error(t, validateBuiltinValue("IDREFS", ""))
}

func TestBuiltinNMTOKENS(t *testing.T) {
	assert.NoError(t, validateBuiltinValue("NMTOKENS", "a b c"))
	assert.Error(t, validateBuiltinValue("NMTOKENS", ""))
}

func TestBuiltinAnyURI(t *testing.T) {
	assert.NoError(t, validateBuiltinValue("anyURI", "http://example.com"))
}

func TestBuiltinUnknown(t *testing.T) {
	assert.Error(t, validateBuiltinValue("unknownType", "val"))
}

// --- Facet coverage ---

func TestFacetTotalDigits(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="n">
    <xs:simpleType>
      <xs:restriction base="xs:decimal">
        <xs:totalDigits value="3"/>
      </xs:restriction>
    </xs:simpleType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><n>123</n>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><n>1234</n>`, xsd, "exceeds totalDigits")
}

func TestFacetFractionDigits(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="n">
    <xs:simpleType>
      <xs:restriction base="xs:decimal">
        <xs:fractionDigits value="2"/>
      </xs:restriction>
    </xs:simpleType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><n>1.23</n>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><n>1.234</n>`, xsd, "exceeds fractionDigits")
}

func TestFacetLength(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="code">
    <xs:simpleType>
      <xs:restriction base="xs:string">
        <xs:length value="3"/>
      </xs:restriction>
    </xs:simpleType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><code>ABC</code>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><code>AB</code>`, xsd, "does not equal required length")
}

func TestFacetMinMaxExclusive(t *testing.T) {
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

func TestFacetDateRange(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="d">
    <xs:simpleType>
      <xs:restriction base="xs:date">
        <xs:minInclusive value="2020-01-01"/>
        <xs:maxInclusive value="2025-12-31"/>
      </xs:restriction>
    </xs:simpleType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><d>2024-06-15</d>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><d>2019-12-31</d>`, xsd, "must be >= 2020-01-01")
}

// --- Schema parse coverage ---

func TestSchemaComplexContentExtension(t *testing.T) {
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
          <xs:sequence>
            <xs:element name="b" type="xs:string"/>
          </xs:sequence>
          <xs:attribute name="id" type="xs:integer"/>
        </xs:extension>
      </xs:complexContent>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root id="1"><b>val</b></root>`, xsd)
}

func TestSchemaListType(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="nums">
    <xs:simpleType>
      <xs:list itemType="xs:integer"/>
    </xs:simpleType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><nums>1 2 3</nums>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><nums>1 abc 3</nums>`, xsd, "not a valid integer")
}

func TestSchemaUnionType(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="val">
    <xs:simpleType>
      <xs:union memberTypes="xs:integer xs:boolean"/>
    </xs:simpleType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><val>42</val>`, xsd)
	mustSchemaValid(t, `<?xml version="1.1"?><val>true</val>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><val>hello</val>`, xsd, "does not match any member type")
}

func TestSchemaOptionalElementMissing(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="required" type="xs:string"/>
        <xs:element name="optional" type="xs:string" minOccurs="0"/>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root><required>val</required></root>`, xsd)
}

func TestSchemaEmptyChoice(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:choice minOccurs="0">
        <xs:element name="a" type="xs:string"/>
        <xs:element name="b" type="xs:string"/>
      </xs:choice>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root/>`, xsd)
}

func TestSchemaExtraElement(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="a" type="xs:string"/>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><root><a>val</a><b>extra</b></root>`, xsd, "unexpected element")
}

func TestSchemaSimpleContentRestriction(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="status">
    <xs:complexType>
      <xs:simpleContent>
        <xs:restriction base="xs:string">
          <xs:enumeration value="active"/>
          <xs:enumeration value="inactive"/>
          <xs:attribute name="since" type="xs:date"/>
        </xs:restriction>
      </xs:simpleContent>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><status since="2024-01-01">active</status>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><status>unknown</status>`, xsd, "not one of the allowed values")
}

func TestSchemaUnsupportedImport(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:import namespace="http://other" schemaLocation="other.xsd"/>
  <xs:element name="root" type="xs:string"/>
</xs:schema>`
	err := schemaValidate(t, `<?xml version="1.1"?><root>val</root>`, xsd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported")
}

func TestSchemaAttributeFixed(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="item">
    <xs:complexType>
      <xs:attribute name="version" type="xs:string" fixed="2.0"/>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><item version="2.0"/>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><item version="3.0"/>`, xsd, "fixed value")
}

func TestSchemaProhibitedAttribute(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="item">
    <xs:complexType>
      <xs:attribute name="ok" type="xs:string"/>
      <xs:attribute name="banned" type="xs:string" use="prohibited"/>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><item ok="yes"/>`, xsd)
}

func TestSchemaMultipleChoiceOccurrences(t *testing.T) {
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
	mustSchemaValid(t, `<?xml version="1.1"?><root><a>1</a><b>2</b><a>3</a></root>`, xsd)
}

func TestSchemaAnyParticle(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:sequence>
        <xs:any processContents="skip" minOccurs="0" maxOccurs="unbounded"/>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root><anything>here</anything><other/></root>`, xsd)
}

func TestSchemaEmptyContent(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="empty">
    <xs:complexType/>
  </xs:element>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><empty>text</empty>`, xsd, "expects empty content")
}

func TestSchemaNamedComplexType(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:complexType name="personType">
    <xs:sequence>
      <xs:element name="name" type="xs:string"/>
    </xs:sequence>
    <xs:attribute name="id" type="xs:integer" use="required"/>
  </xs:complexType>
  <xs:element name="person" type="personType"/>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><person id="1"><name>Alice</name></person>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><person><name>Alice</name></person>`, xsd, "required attribute")
}

func TestSchemaAttributeInlineSimpleType(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="item">
    <xs:complexType>
      <xs:attribute name="status">
        <xs:simpleType>
          <xs:restriction base="xs:string">
            <xs:enumeration value="on"/>
            <xs:enumeration value="off"/>
          </xs:restriction>
        </xs:simpleType>
      </xs:attribute>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><item status="on"/>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><item status="maybe"/>`, xsd, "not one of the allowed values")
}

func TestSchemaElementRef(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="name" type="xs:string"/>
  <xs:element name="root">
    <xs:complexType>
      <xs:sequence>
        <xs:element ref="name"/>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root><name>Alice</name></root>`, xsd)
}

// --- Tree parser edge cases ---

func TestTreeParserEmptyElement(t *testing.T) {
	doc, err := ParseTree(strings.NewReader(`<r/>`))
	require.NoError(t, err)
	assert.Equal(t, "r", doc.Root.Local)
	assert.Empty(t, doc.Root.Children)
}

func TestTreeParserCommentSkipped(t *testing.T) {
	doc, err := ParseTree(strings.NewReader(`<!-- before --><r><!-- inside --></r><!-- after -->`))
	require.NoError(t, err)
	assert.Equal(t, "r", doc.Root.Local)
}

func TestTreeParserPISkipped(t *testing.T) {
	doc, err := ParseTree(strings.NewReader(`<?target data?><r/>`))
	require.NoError(t, err)
	assert.Equal(t, "r", doc.Root.Local)
}

func TestTreeParserCharRef(t *testing.T) {
	doc, err := ParseTree(strings.NewReader(`<r>&#65;&#x42;</r>`))
	require.NoError(t, err)
	assert.Equal(t, "AB", doc.Root.TextContent())
}

func TestTreeParserAttrWithEntities(t *testing.T) {
	doc, err := ParseTree(strings.NewReader(`<r a="&amp;"/>`))
	require.NoError(t, err)
	assert.Equal(t, "&", doc.Root.Attrs[0].Value)
}

func TestTreeParserDefaultNamespace(t *testing.T) {
	doc, err := ParseTree(strings.NewReader(`<r xmlns="http://example.com"><child/></r>`))
	require.NoError(t, err)
	assert.Equal(t, "http://example.com", doc.Root.Namespace)
	children := doc.Root.ChildElements()
	require.Equal(t, 1, len(children))
	assert.Equal(t, "http://example.com", children[0].Namespace)
}

func TestTreeParserMixedContent(t *testing.T) {
	doc, err := ParseTree(strings.NewReader(`<r>text1<b>bold</b>text2</r>`))
	require.NoError(t, err)
	require.Equal(t, 3, len(doc.Root.Children))
	cd, ok := doc.Root.Children[0].(*CharData)
	require.True(t, ok)
	assert.Equal(t, "text1", cd.Content)
}

func TestSchemaNestedSequence(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:sequence>
        <xs:sequence minOccurs="1" maxOccurs="unbounded">
          <xs:element name="a" type="xs:string"/>
          <xs:element name="b" type="xs:string"/>
        </xs:sequence>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root><a>1</a><b>2</b><a>3</a><b>4</b></root>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><root></root>`, xsd, "requires at least")
}

func TestSchemaNestedChoice(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:sequence>
        <xs:choice minOccurs="1" maxOccurs="3">
          <xs:element name="x" type="xs:string"/>
          <xs:element name="y" type="xs:string"/>
        </xs:choice>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root><x>1</x><y>2</y></root>`, xsd)
}

func TestSchemaComplexContentRestriction(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:complexContent>
        <xs:restriction base="xs:anyType">
          <xs:sequence>
            <xs:element name="item" type="xs:string" maxOccurs="unbounded"/>
          </xs:sequence>
        </xs:restriction>
      </xs:complexContent>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root><item>a</item><item>b</item></root>`, xsd)
}

func TestSchemaGroupDecl(t *testing.T) {
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
        <xs:element name="data" type="xs:string"/>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root><data>test</data></root>`, xsd)
}

func TestSchemaAttrGroupDecl(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:attributeGroup name="common">
    <xs:attribute name="id" type="xs:integer"/>
  </xs:attributeGroup>
  <xs:element name="root">
    <xs:complexType>
      <xs:attribute name="name" type="xs:string"/>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root name="test"/>`, xsd)
}

func TestSchemaUnsupportedNotation(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:notation name="jpeg" public="image/jpeg"/>
  <xs:element name="root" type="xs:string"/>
</xs:schema>`
	err := schemaValidate(t, `<?xml version="1.1"?><root>val</root>`, xsd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported")
}

func TestSchemaChoiceRequired(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:choice>
        <xs:element name="a" type="xs:string"/>
        <xs:element name="b" type="xs:string"/>
      </xs:choice>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><root/>`, xsd, "requires one of")
}

func TestTreeParserDoctype(t *testing.T) {
	doc, err := ParseTree(strings.NewReader(`<?xml version="1.0"?><!DOCTYPE root><root/>`))
	require.NoError(t, err)
	assert.Equal(t, "root", doc.Root.Local)
}

func TestSchemaInvalidRoot(t *testing.T) {
	xsd := `<?xml version="1.0"?><notschema/>`
	err := schemaValidate(t, `<?xml version="1.1"?><root/>`, xsd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expected xs:schema")
}
