package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// One named type reaches both an element and an attribute, which is what makes
// a gap between the two visible.
const attrFacetXSD = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:simpleType name="pat">
    <xs:restriction base="xs:token"><xs:pattern value="[a-z]+"/></xs:restriction>
  </xs:simpleType>
  <xs:element name="r">
    <xs:complexType>
      <xs:sequence><xs:element name="t" type="pat" minOccurs="0"/></xs:sequence>
      <xs:attribute name="p" type="pat"/>
    </xs:complexType>
  </xs:element>
</xs:schema>`

func TestSchemaAttributePatternFacet(t *testing.T) {
	mustSchemaValid(t, `<?xml version="1.1"?><r p="abc"/>`, attrFacetXSD)
	mustSchemaReject(t, `<?xml version="1.1"?><r p="ABC123"/>`, attrFacetXSD, `does not match pattern "[a-z]+"`)
	mustSchemaReject(t, `<?xml version="1.1"?><r><t>ABC123</t></r>`, attrFacetXSD, `does not match pattern "[a-z]+"`)
}

func TestSchemaAttributeLengthFacets(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:simpleType name="nonEmpty">
    <xs:restriction base="xs:string"><xs:minLength value="1"/></xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="code">
    <xs:restriction base="xs:string"><xs:length value="3"/></xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="short">
    <xs:restriction base="xs:string"><xs:maxLength value="4"/></xs:restriction>
  </xs:simpleType>
  <xs:element name="r">
    <xs:complexType>
      <xs:attribute name="reason" type="nonEmpty"/>
      <xs:attribute name="code" type="code"/>
      <xs:attribute name="tag" type="short"/>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><r reason="why" code="abc" tag="ok"/>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><r reason=""/>`, xsd, "less than minLength 1")
	mustSchemaReject(t, `<?xml version="1.1"?><r code="ab"/>`, xsd, "required length 3")
	mustSchemaReject(t, `<?xml version="1.1"?><r tag="toolong"/>`, xsd, "exceeds maxLength 4")
}

func TestSchemaAttributeRangeAndDigitFacets(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:simpleType name="percent">
    <xs:restriction base="xs:int">
      <xs:minInclusive value="0"/>
      <xs:maxInclusive value="100"/>
    </xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="money">
    <xs:restriction base="xs:decimal"><xs:fractionDigits value="2"/></xs:restriction>
  </xs:simpleType>
  <xs:element name="r">
    <xs:complexType>
      <xs:attribute name="pct" type="percent"/>
      <xs:attribute name="cost" type="money"/>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><r pct="50" cost="1.25"/>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><r pct="101"/>`, xsd, "must be <= 100")
	mustSchemaReject(t, `<?xml version="1.1"?><r pct="-1"/>`, xsd, "must be >= 0")
	mustSchemaReject(t, `<?xml version="1.1"?><r cost="1.234"/>`, xsd, "fractionDigits 2")
}

func TestSchemaAttributeListAndUnionTypes(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:simpleType name="intList">
    <xs:list itemType="xs:integer"/>
  </xs:simpleType>
  <xs:simpleType name="numOrBool">
    <xs:union memberTypes="xs:integer xs:boolean"/>
  </xs:simpleType>
  <xs:element name="r">
    <xs:complexType>
      <xs:attribute name="nums" type="intList"/>
      <xs:attribute name="flag" type="numOrBool"/>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><r nums="1 2 3" flag="true"/>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><r nums="1 abc 3"/>`, xsd, "not a valid integer")
	mustSchemaReject(t, `<?xml version="1.1"?><r flag="hello"/>`, xsd, "does not match any member type")
}

func TestSchemaFacetsInheritedFromBaseType(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:simpleType name="short5">
    <xs:restriction base="xs:string"><xs:maxLength value="5"/></xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="lowerShort">
    <xs:restriction base="short5"><xs:pattern value="[a-z]+"/></xs:restriction>
  </xs:simpleType>
  <xs:element name="r">
    <xs:complexType>
      <xs:sequence><xs:element name="t" type="lowerShort" minOccurs="0"/></xs:sequence>
      <xs:attribute name="v" type="lowerShort"/>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><r v="abc"><t>xyz</t></r>`, xsd)
	// maxLength comes from the base type, pattern from the derived one. Both
	// apply, on the attribute and on the element alike.
	mustSchemaReject(t, `<?xml version="1.1"?><r v="abcdefg"/>`, xsd, "exceeds maxLength 5")
	mustSchemaReject(t, `<?xml version="1.1"?><r v="ABC"/>`, xsd, "does not match pattern")
	mustSchemaReject(t, `<?xml version="1.1"?><r><t>abcdefg</t></r>`, xsd, "exceeds maxLength 5")
}

func TestSchemaListLengthFacetsCountItems(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:simpleType name="intList">
    <xs:list itemType="xs:integer"/>
  </xs:simpleType>
  <xs:simpleType name="pair">
    <xs:restriction base="intList"><xs:length value="2"/></xs:restriction>
  </xs:simpleType>
  <xs:element name="r">
    <xs:complexType>
      <xs:attribute name="point" type="pair"/>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	// "10 20" is 5 characters and 2 items; the facet counts the items.
	mustSchemaValid(t, `<?xml version="1.1"?><r point="10 20"/>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><r point="10 20 30"/>`, xsd, "3 item(s)")
}

func TestSchemaListWithInlineItemType(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="r">
    <xs:complexType>
      <xs:attribute name="names">
        <xs:simpleType>
          <xs:list>
            <xs:simpleType>
              <xs:restriction base="xs:string"><xs:pattern value="[a-z]+"/></xs:restriction>
            </xs:simpleType>
          </xs:list>
        </xs:simpleType>
      </xs:attribute>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><r names="alpha beta"/>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><r names="alpha BETA"/>`, xsd, "does not match pattern")
}

func TestSchemaUnionWithInlineMembers(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="r">
    <xs:complexType>
      <xs:attribute name="size">
        <xs:simpleType>
          <xs:union>
            <xs:simpleType>
              <xs:restriction base="xs:string"><xs:enumeration value="auto"/></xs:restriction>
            </xs:simpleType>
            <xs:simpleType>
              <xs:restriction base="xs:int"><xs:minInclusive value="1"/></xs:restriction>
            </xs:simpleType>
          </xs:union>
        </xs:simpleType>
      </xs:attribute>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><r size="auto"/>`, xsd)
	mustSchemaValid(t, `<?xml version="1.1"?><r size="4"/>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><r size="0"/>`, xsd, "does not match any member type")
	mustSchemaReject(t, `<?xml version="1.1"?><r size="wide"/>`, xsd, "does not match any member type")
}

func TestSchemaListWithoutItemTypeRejected(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:simpleType name="anything"><xs:list/></xs:simpleType>
  <xs:element name="r" type="anything"/>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><r>a b</r>`, xsd, "xs:list requires an itemType")
}

func TestSchemaUnionWithoutMembersRejected(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:simpleType name="anything"><xs:union/></xs:simpleType>
  <xs:element name="r" type="anything"/>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><r>a</r>`, xsd, "xs:union requires a memberTypes")
}

func TestSchemaCircularSimpleTypeRejected(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:simpleType name="a"><xs:restriction base="b"/></xs:simpleType>
  <xs:simpleType name="b"><xs:restriction base="a"/></xs:simpleType>
  <xs:element name="r" type="a"/>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><r>x</r>`, xsd, "derives from itself")
}

func TestSchemaAttributeErrorPointsAtTheAttribute(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:simpleType name="pat">
    <xs:restriction base="xs:string"><xs:pattern value="[a-z]+"/></xs:restriction>
  </xs:simpleType>
  <xs:element name="r">
    <xs:complexType>
      <xs:attribute name="ok" type="xs:string"/>
      <xs:attribute name="bad" type="pat"/>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	xml := "<?xml version=\"1.1\"?>\n<r\n  ok=\"anything\"\n  bad=\"NOPE\"/>\n"
	err := schemaValidate(t, xml, xsd)
	require.Error(t, err)
	var vErr *Error
	require.ErrorAs(t, err, &vErr)
	assert.Equal(t, 4, vErr.Line, "the offending attribute is on line 4, the element on line 2")
	assert.Equal(t, 3, vErr.Col)
}
