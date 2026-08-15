package validator

import "testing"

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
	mustSchemaValid(t, `<?xml version="1.1"?><root><key>k</key><value>v</value><extra>v</extra></root>`, xsd)
	// The group is what the reference stands for, so its members are required
	// exactly as if they had been written out at the reference.
	mustSchemaReject(t, `<?xml version="1.1"?><root><extra>v</extra></root>`, xsd, `occurrence(s) of "key"`)
}

func TestSchemaGroupRefOccursAndCycle(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:group name="pair">
    <xs:sequence>
      <xs:element name="key" type="xs:string"/>
    </xs:sequence>
  </xs:group>
  <xs:element name="root">
    <xs:complexType>
      <xs:sequence>
        <xs:group ref="pair" minOccurs="0" maxOccurs="unbounded"/>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	// The occurrence counts stated at the reference are the ones that apply.
	mustSchemaValid(t, `<?xml version="1.1"?><root/>`, xsd)
	mustSchemaValid(t, `<?xml version="1.1"?><root><key>a</key><key>b</key></root>`, xsd)

	missing := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:sequence><xs:group ref="nope"/></xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><root/>`, missing, `group ref "nope"`)

	cyclic := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:group name="loop">
    <xs:sequence><xs:group ref="loop"/></xs:sequence>
  </xs:group>
  <xs:element name="root">
    <xs:complexType>
      <xs:sequence><xs:group ref="loop"/></xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><root/>`, cyclic, "refers to itself")
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

func TestSchemaComplexContentExtensionAllIsRejected(t *testing.T) {
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
	mustSchemaReject(t, `<?xml version="1.1"?><root><a>v0</a><b>v1</b><c>v2</c></root>`, xsd,
		"xs:all content model cannot be combined")

	nested := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:sequence>
        <xs:all><xs:element name="a" type="xs:string"/></xs:all>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><root><a>v</a></root>`, nested,
		"xs:all must be the entire content model")
}
