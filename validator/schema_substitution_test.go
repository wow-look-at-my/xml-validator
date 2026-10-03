package validator

import "testing"

// A head element referenced from a content model, with two members carrying
// their own types. The member is what the instance is validated against.
const substitutionXSD = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:complexType name="shape">
    <xs:sequence><xs:element name="label" type="xs:string"/></xs:sequence>
  </xs:complexType>
  <xs:complexType name="circle">
    <xs:complexContent>
      <xs:extension base="shape">
        <xs:sequence><xs:element name="radius" type="xs:int"/></xs:sequence>
      </xs:extension>
    </xs:complexContent>
  </xs:complexType>
  <xs:complexType name="square">
    <xs:complexContent>
      <xs:extension base="shape">
        <xs:sequence><xs:element name="side" type="xs:int"/></xs:sequence>
      </xs:extension>
    </xs:complexContent>
  </xs:complexType>

  <xs:element name="shape" type="shape"/>
  <xs:element name="circle" type="circle" substitutionGroup="shape"/>
  <xs:element name="square" type="square" substitutionGroup="shape"/>

  <xs:element name="drawing">
    <xs:complexType>
      <xs:sequence><xs:element ref="shape" maxOccurs="unbounded"/></xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`

func TestSchemaSubstitutionGroupAccepted(t *testing.T) {
	// The head itself still works where it is referenced.
	mustSchemaValid(t, `<?xml version="1.1"?><drawing><shape><label>x</label></shape></drawing>`, substitutionXSD)
	// And so does each member, in any mix.
	mustSchemaValid(t, `<?xml version="1.1"?><drawing>`+
		`<circle><label>c</label><radius>3</radius></circle>`+
		`<square><label>s</label><side>4</side></square></drawing>`, substitutionXSD)
}

func TestSchemaSubstituteValidatedAgainstItsOwnType(t *testing.T) {
	// radius is an xs:int on circle's own type: substitution must not validate
	// the member against the head's type, which has no radius at all.
	mustSchemaReject(t, `<?xml version="1.1"?><drawing>`+
		`<circle><label>c</label><radius>huge</radius></circle></drawing>`, substitutionXSD,
		"not a valid int")
	// The base content still applies: label comes first and is required.
	mustSchemaReject(t, `<?xml version="1.1"?><drawing>`+
		`<circle><radius>3</radius></circle></drawing>`, substitutionXSD,
		`occurrence(s) of "label"`)
}

func TestSchemaNonMemberStillRejected(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="head" type="xs:string"/>
  <xs:element name="member" type="xs:string" substitutionGroup="head"/>
  <xs:element name="stranger" type="xs:string"/>
  <xs:element name="doc">
    <xs:complexType>
      <xs:sequence><xs:element ref="head"/></xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><doc><member>x</member></doc>`, xsd)
	// stranger declares no substitutionGroup, so it does not get in.
	mustSchemaReject(t, `<?xml version="1.1"?><doc><stranger>x</stranger></doc>`, xsd, "requires at least")
}

func TestSchemaSubstitutionIsTransitive(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="a" type="xs:string"/>
  <xs:element name="b" type="xs:string" substitutionGroup="a"/>
  <xs:element name="c" type="xs:string" substitutionGroup="b"/>
  <xs:element name="doc">
    <xs:complexType>
      <xs:sequence><xs:element ref="a"/></xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	// c substitutes for b, which substitutes for a, so c stands in for a too.
	mustSchemaValid(t, `<?xml version="1.1"?><doc><c>x</c></doc>`, xsd)
	mustSchemaValid(t, `<?xml version="1.1"?><doc><b>x</b></doc>`, xsd)
}

func TestSchemaAbstractHeadRequiresSubstitute(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="shape" type="xs:string" abstract="true"/>
  <xs:element name="circle" type="xs:string" substitutionGroup="shape"/>
  <xs:element name="drawing">
    <xs:complexType>
      <xs:sequence><xs:element ref="shape"/></xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><drawing><circle>c</circle></drawing>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><drawing><shape>s</shape></drawing>`, xsd, "is abstract")
	mustSchemaReject(t, `<?xml version="1.1"?><shape>s</shape>`, xsd, "is abstract")
}

func TestSchemaBlockedSubstitutionRejected(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="head" type="xs:string" block="substitution"/>
  <xs:element name="member" type="xs:string" substitutionGroup="head"/>
  <xs:element name="doc">
    <xs:complexType>
      <xs:sequence><xs:element ref="head"/></xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><doc><head>x</head></doc>`, xsd)
	// The head refuses to be replaced, so the member is not allowed here.
	mustSchemaReject(t, `<?xml version="1.1"?><doc><member>x</member></doc>`, xsd, "requires at least")
}

func TestSchemaBlockDefaultBlocksSubstitution(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" blockDefault="#all">
  <xs:element name="head" type="xs:string"/>
  <xs:element name="member" type="xs:string" substitutionGroup="head"/>
  <xs:element name="doc">
    <xs:complexType>
      <xs:sequence><xs:element ref="head"/></xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><doc><member>x</member></doc>`, xsd, "requires at least")
}

func TestSchemaSubstitutionInAllGroup(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="head" type="xs:string"/>
  <xs:element name="member" type="xs:string" substitutionGroup="head"/>
  <xs:element name="other" type="xs:string"/>
  <xs:element name="doc">
    <xs:complexType>
      <xs:all>
        <xs:element ref="head"/>
        <xs:element ref="other"/>
      </xs:all>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	// A substitute fills the head's slot, in either order, and still satisfies
	// the head's minOccurs.
	mustSchemaValid(t, `<?xml version="1.1"?><doc><other>o</other><member>m</member></doc>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><doc><other>o</other></doc>`, xsd, "requires at least")
	// The slot is filled once: a second substitute exceeds maxOccurs.
	mustSchemaReject(t, `<?xml version="1.1"?><doc><other>o</other><member>m</member><member>n</member></doc>`, xsd,
		"appears too many times")
}

func TestSchemaSubstitutionInChoice(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="head" type="xs:string"/>
  <xs:element name="member" type="xs:string" substitutionGroup="head"/>
  <xs:element name="doc">
    <xs:complexType>
      <xs:choice>
        <xs:element ref="head"/>
        <xs:element name="alt" type="xs:string"/>
      </xs:choice>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><doc><member>x</member></doc>`, xsd)
	mustSchemaValid(t, `<?xml version="1.1"?><doc><alt>x</alt></doc>`, xsd)
}

func TestSchemaSubstitutionAppliesToReferencesOnly(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="head" type="xs:string"/>
  <xs:element name="member" type="xs:string" substitutionGroup="head"/>
  <xs:element name="doc">
    <xs:complexType>
      <xs:sequence><xs:element name="head" type="xs:string"/></xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	// The local declaration is a different element that happens to share a
	// name, so nothing substitutes for it.
	mustSchemaValid(t, `<?xml version="1.1"?><doc><head>x</head></doc>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><doc><member>x</member></doc>`, xsd, "requires at least")
}

func TestSchemaSubstitutionUnknownHeadRejected(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="member" type="xs:string" substitutionGroup="nobody"/>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><member>x</member>`, xsd, "names no global element declaration")
}

func TestSchemaSubstitutionUnrelatedTypeRejected(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:complexType name="shape">
    <xs:sequence><xs:element name="label" type="xs:string"/></xs:sequence>
  </xs:complexType>
  <xs:complexType name="unrelated">
    <xs:sequence><xs:element name="other" type="xs:string"/></xs:sequence>
  </xs:complexType>
  <xs:element name="shape" type="shape"/>
  <xs:element name="bogus" type="unrelated" substitutionGroup="shape"/>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><shape><label>x</label></shape>`, xsd, "does not derive from")
}

func TestSchemaSubstitutionSimpleTypeDerivationAccepted(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:simpleType name="code">
    <xs:restriction base="xs:string"><xs:maxLength value="4"/></xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="shortCode">
    <xs:restriction base="code"><xs:maxLength value="2"/></xs:restriction>
  </xs:simpleType>
  <xs:element name="head" type="code"/>
  <xs:element name="member" type="shortCode" substitutionGroup="head"/>
  <xs:element name="doc">
    <xs:complexType>
      <xs:sequence><xs:element ref="head"/></xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><doc><member>ab</member></doc>`, xsd)
	// The member's own facet is the one that applies.
	mustSchemaReject(t, `<?xml version="1.1"?><doc><member>abc</member></doc>`, xsd, "exceeds maxLength 2")
	mustSchemaValid(t, `<?xml version="1.1"?><doc><head>abcd</head></doc>`, xsd)
}

func TestSchemaSubstitutionCircularRejected(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="a" type="xs:string" substitutionGroup="b"/>
  <xs:element name="b" type="xs:string" substitutionGroup="a"/>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><a>x</a>`, xsd, "circular")
}

func TestSchemaSubstitutionUntypedHeadAccepted(t *testing.T) {
	// Neither declares a type, so there is no derivation to check and the
	// group still works.
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="head"/>
  <xs:element name="member" substitutionGroup="head"/>
  <xs:element name="doc">
    <xs:complexType>
      <xs:sequence><xs:element ref="head"/></xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><doc><member/></doc>`, xsd)
}

func TestSchemaSubstituteCarriesItsOwnIdentityConstraints(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:complexType name="empty"/>
  <xs:complexType name="listing">
    <xs:complexContent>
      <xs:extension base="empty">
        <xs:sequence>
          <xs:element name="item" maxOccurs="unbounded">
            <xs:complexType><xs:attribute name="id" type="xs:string"/></xs:complexType>
          </xs:element>
        </xs:sequence>
      </xs:extension>
    </xs:complexContent>
  </xs:complexType>
  <xs:element name="section" type="empty"/>
  <xs:element name="items" type="listing" substitutionGroup="section">
    <xs:key name="itemId">
      <xs:selector xpath="item"/>
      <xs:field xpath="@id"/>
    </xs:key>
  </xs:element>
  <xs:element name="doc">
    <xs:complexType>
      <xs:sequence><xs:element ref="section"/></xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><doc><items><item id="a"/><item id="b"/></items></doc>`, xsd)
	// The key belongs to the substitute, and it runs where the substitute does.
	mustSchemaReject(t, `<?xml version="1.1"?><doc><items><item id="a"/><item id="a"/></items></doc>`, xsd,
		"repeats a value")
}

func TestSchemaSubstitutionNamespaced(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
           xmlns:t="https://example.com/sub"
           targetNamespace="https://example.com/sub"
           elementFormDefault="qualified">
  <xs:element name="head" type="xs:string"/>
  <xs:element name="member" type="xs:string" substitutionGroup="t:head"/>
  <xs:element name="doc">
    <xs:complexType>
      <xs:sequence><xs:element ref="t:head"/></xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?>`+
		`<doc xmlns="https://example.com/sub"><member>x</member></doc>`, xsd)
}
