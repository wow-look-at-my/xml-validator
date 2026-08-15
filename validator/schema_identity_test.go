package validator

import "testing"

// The schema the issue reduced from: a key on provider/@id and a keyref from
// role/@provider. Nothing here checks either, so the schema is rejected instead
// of validating documents against constraints that never run.
const identityXSD = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="cfg">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="provider" maxOccurs="unbounded">
          <xs:complexType><xs:attribute name="id" type="xs:string"/></xs:complexType>
        </xs:element>
        <xs:element name="role" minOccurs="0" maxOccurs="unbounded">
          <xs:complexType><xs:attribute name="provider" type="xs:string"/></xs:complexType>
        </xs:element>
      </xs:sequence>
    </xs:complexType>
    <xs:key name="providerKey">
      <xs:selector xpath="provider"/>
      <xs:field xpath="@id"/>
    </xs:key>
    <xs:keyref name="roleProvider" refer="providerKey">
      <xs:selector xpath="role"/>
      <xs:field xpath="@provider"/>
    </xs:keyref>
  </xs:element>
</xs:schema>`

// The document that used to come back "schema validated": a duplicate key and a
// dangling reference.
const identityXML = `<?xml version="1.1"?>
<cfg>
  <provider id="openrouter"/>
  <provider id="openrouter"/>
  <role provider="does-not-exist"/>
</cfg>`

func TestSchemaIdentityConstraintsRejected(t *testing.T) {
	mustSchemaReject(t, identityXML, identityXSD, "identity constraint xs:key")
}

func TestSchemaUniqueRejected(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="r">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="item" maxOccurs="unbounded" type="xs:string"/>
      </xs:sequence>
    </xs:complexType>
    <xs:unique name="itemUnique">
      <xs:selector xpath="item"/>
      <xs:field xpath="."/>
    </xs:unique>
  </xs:element>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><r><item>a</item></r>`, xsd, "identity constraint xs:unique")
}

func TestSchemaSubstitutionGroupRejected(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="head" type="xs:string"/>
  <xs:element name="member" type="xs:string" substitutionGroup="head"/>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><head>x</head>`, xsd, "substitutionGroup")
}

func TestSchemaUnknownElementChildRejected(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="r" type="xs:string">
    <xs:alternative test="true"/>
  </xs:element>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><r>x</r>`, xsd, "unsupported schema element xs:alternative")
}
