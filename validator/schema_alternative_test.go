package validator

import "testing"

// One element whose type depends on its own kind attribute, which is what
// conditional type assignment is for.
const alternativeXSD = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:complexType name="withRadius">
    <xs:sequence><xs:element name="radius" type="xs:int"/></xs:sequence>
    <xs:attribute name="kind" type="xs:string"/>
  </xs:complexType>
  <xs:complexType name="withSide">
    <xs:sequence><xs:element name="side" type="xs:int"/></xs:sequence>
    <xs:attribute name="kind" type="xs:string"/>
  </xs:complexType>
  <xs:complexType name="plain">
    <xs:sequence><xs:element name="label" type="xs:string"/></xs:sequence>
    <xs:attribute name="kind" type="xs:string"/>
  </xs:complexType>
  <xs:element name="shape" type="plain">
    <xs:alternative test="@kind='circle'" type="withRadius"/>
    <xs:alternative test="@kind='square'" type="withSide"/>
  </xs:element>
</xs:schema>`

func TestSchemaAlternativeChoosesTypeByTest(t *testing.T) {
	mustSchemaValid(t, `<?xml version="1.1"?><shape kind="circle"><radius>3</radius></shape>`, alternativeXSD)
	mustSchemaValid(t, `<?xml version="1.1"?><shape kind="square"><side>4</side></shape>`, alternativeXSD)
	// No test matches, so the declared type applies.
	mustSchemaValid(t, `<?xml version="1.1"?><shape kind="other"><label>x</label></shape>`, alternativeXSD)
}

func TestSchemaAlternativeEnforcesTheChosenType(t *testing.T) {
	// kind=circle selects withRadius, so a side element is wrong here.
	mustSchemaReject(t, `<?xml version="1.1"?><shape kind="circle"><side>4</side></shape>`, alternativeXSD,
		`occurrence(s) of "radius"`)
	// And the chosen type's own content is checked.
	mustSchemaReject(t, `<?xml version="1.1"?><shape kind="circle"><radius>big</radius></shape>`, alternativeXSD,
		"not a valid int")
}

func TestSchemaAlternativeDefaultWithoutTest(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:complexType name="numeric">
    <xs:simpleContent><xs:extension base="xs:int">
      <xs:attribute name="kind" type="xs:string"/>
    </xs:extension></xs:simpleContent>
  </xs:complexType>
  <xs:complexType name="textual">
    <xs:simpleContent><xs:extension base="xs:string">
      <xs:attribute name="kind" type="xs:string"/>
    </xs:extension></xs:simpleContent>
  </xs:complexType>
  <xs:element name="v" type="textual">
    <xs:alternative test="@kind='n'" type="numeric"/>
    <xs:alternative type="textual"/>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><v kind="n">42</v>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><v kind="n">words</v>`, xsd, "not a valid int")
	// The alternative with no test always holds, so anything else is textual.
	mustSchemaValid(t, `<?xml version="1.1"?><v kind="t">words</v>`, xsd)
}

func TestSchemaAlternativeTestForms(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:complexType name="numeric">
    <xs:simpleContent><xs:extension base="xs:int">
      <xs:attribute name="a" type="xs:string"/>
      <xs:attribute name="b" type="xs:string"/>
    </xs:extension></xs:simpleContent>
  </xs:complexType>
  <xs:complexType name="textual">
    <xs:simpleContent><xs:extension base="xs:string">
      <xs:attribute name="a" type="xs:string"/>
      <xs:attribute name="b" type="xs:string"/>
    </xs:extension></xs:simpleContent>
  </xs:complexType>
  <xs:element name="present" type="textual">
    <xs:alternative test="@a" type="numeric"/>
  </xs:element>
  <xs:element name="negated" type="textual">
    <xs:alternative test="not(@a)" type="numeric"/>
  </xs:element>
  <xs:element name="notEqual" type="textual">
    <xs:alternative test="@a != 'x'" type="numeric"/>
  </xs:element>
  <xs:element name="both" type="textual">
    <xs:alternative test="@a='1' and @b='2'" type="numeric"/>
  </xs:element>
  <xs:element name="either" type="textual">
    <xs:alternative test="@a='1' or @b='2'" type="numeric"/>
  </xs:element>
</xs:schema>`
	// @a present selects the numeric type; absent falls back to textual.
	mustSchemaValid(t, `<?xml version="1.1"?><present a="y">7</present>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><present a="y">words</present>`, xsd, "not a valid int")
	mustSchemaValid(t, `<?xml version="1.1"?><present>words</present>`, xsd)

	mustSchemaReject(t, `<?xml version="1.1"?><negated>words</negated>`, xsd, "not a valid int")
	mustSchemaValid(t, `<?xml version="1.1"?><negated a="y">words</negated>`, xsd)

	mustSchemaReject(t, `<?xml version="1.1"?><notEqual a="y">words</notEqual>`, xsd, "not a valid int")
	mustSchemaValid(t, `<?xml version="1.1"?><notEqual a="x">words</notEqual>`, xsd)

	mustSchemaReject(t, `<?xml version="1.1"?><both a="1" b="2">words</both>`, xsd, "not a valid int")
	mustSchemaValid(t, `<?xml version="1.1"?><both a="1" b="3">words</both>`, xsd)

	mustSchemaReject(t, `<?xml version="1.1"?><either a="1" b="9">words</either>`, xsd, "not a valid int")
	mustSchemaValid(t, `<?xml version="1.1"?><either a="9" b="9">words</either>`, xsd)
}

func TestSchemaAlternativeKeywordInsideAttributeName(t *testing.T) {
	// "brand" contains "and": the splitter must not read that as an operator.
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:complexType name="numeric">
    <xs:simpleContent><xs:extension base="xs:int">
      <xs:attribute name="brand" type="xs:string"/>
    </xs:extension></xs:simpleContent>
  </xs:complexType>
  <xs:complexType name="textual">
    <xs:simpleContent><xs:extension base="xs:string">
      <xs:attribute name="brand" type="xs:string"/>
    </xs:extension></xs:simpleContent>
  </xs:complexType>
  <xs:element name="v" type="textual">
    <xs:alternative test="@brand='acme'" type="numeric"/>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><v brand="acme">7</v>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><v brand="acme">words</v>`, xsd, "not a valid int")
	mustSchemaValid(t, `<?xml version="1.1"?><v brand="other">words</v>`, xsd)
}

func TestSchemaAlternativeUnsupportedTestRejected(t *testing.T) {
	tests := []struct {
		name   string
		test   string
		expect string
	}{
		{"unsupported function", "count(@a) > 1", "function \"count\" is not supported"},
		{"element test", "child='x'", "not an attribute, a literal, or a constructor function"},
		{"user-defined cast", "@a cast as myType", "does not name a built-in datatype"},
		{"unclosed literal", "@a = 'x", "never closed"},
		{"unclosed paren", "(@a='1'", "expected \")\""},
		{"empty", " ", "test is empty"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="v" type="xs:string">
    <xs:alternative test="` + tc.test + `" type="xs:int"/>
  </xs:element>
</xs:schema>`
			mustSchemaReject(t, `<?xml version="1.1"?><v>x</v>`, xsd, tc.expect)
		})
	}
}

func TestSchemaAlternativeUnknownTypeRejected(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="v" type="xs:string">
    <xs:alternative test="@a='1'" type="nothing"/>
  </xs:element>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><v>x</v>`, xsd, "does not name a known type")
}

func TestSchemaAlternativeWithoutTypeRejected(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="v" type="xs:string">
    <xs:alternative test="@a='1'"/>
  </xs:element>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><v>x</v>`, xsd, "requires a type attribute or an inline type")
}

func TestSchemaAlternativeInlineType(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:complexType name="textual">
    <xs:simpleContent><xs:extension base="xs:string">
      <xs:attribute name="kind" type="xs:string"/>
    </xs:extension></xs:simpleContent>
  </xs:complexType>
  <xs:element name="v" type="textual">
    <xs:alternative test="@kind='n'">
      <xs:complexType>
        <xs:simpleContent><xs:extension base="xs:int">
          <xs:attribute name="kind" type="xs:string"/>
        </xs:extension></xs:simpleContent>
      </xs:complexType>
    </xs:alternative>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><v kind="n">7</v>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><v kind="n">words</v>`, xsd, "not a valid int")
}

func TestSchemaAlternativeOnReferencedElement(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:complexType name="numeric">
    <xs:simpleContent><xs:extension base="xs:int">
      <xs:attribute name="kind" type="xs:string"/>
    </xs:extension></xs:simpleContent>
  </xs:complexType>
  <xs:complexType name="textual">
    <xs:simpleContent><xs:extension base="xs:string">
      <xs:attribute name="kind" type="xs:string"/>
    </xs:extension></xs:simpleContent>
  </xs:complexType>
  <xs:element name="v" type="textual">
    <xs:alternative test="@kind='n'" type="numeric"/>
  </xs:element>
  <xs:element name="doc">
    <xs:complexType>
      <xs:sequence><xs:element ref="v" maxOccurs="unbounded"/></xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	// A ref particle carries the alternatives of the declaration it names.
	mustSchemaValid(t, `<?xml version="1.1"?><doc><v kind="n">7</v><v kind="t">words</v></doc>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><doc><v kind="n">words</v></doc>`, xsd, "not a valid int")
}
