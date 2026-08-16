package validator

import (
	"fmt"
	"testing"
)

func sprintfDoc(format, arg string) string { return fmt.Sprintf(format, arg) }

// The subset XSD defines for a selector and a field is enforced by a pattern in
// the schema-for-schemas. Everything exercised here is inside that grammar, so
// rejecting any of it would refuse a conforming schema.

func TestSchemaIdentityChildAxisAccepted(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="r">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="item" maxOccurs="unbounded">
          <xs:complexType><xs:attribute name="id" type="xs:string"/></xs:complexType>
        </xs:element>
      </xs:sequence>
    </xs:complexType>
    <xs:key name="itemKey">
      <xs:selector xpath="child::item"/>
      <xs:field xpath="attribute::id"/>
    </xs:key>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><r><item id="a"/><item id="b"/></r>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><r><item id="a"/><item id="a"/></r>`, xsd, "repeats a value")
}

func TestSchemaIdentitySelfStepAccepted(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="r">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="wrap" maxOccurs="unbounded">
          <xs:complexType>
            <xs:sequence><xs:element name="item" type="xs:string"/></xs:sequence>
          </xs:complexType>
        </xs:element>
      </xs:sequence>
    </xs:complexType>
    <xs:unique name="itemUnique">
      <xs:selector xpath="./wrap/./item"/>
      <xs:field xpath="."/>
    </xs:unique>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><r><wrap><item>a</item></wrap><wrap><item>b</item></wrap></r>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><r><wrap><item>a</item></wrap><wrap><item>a</item></wrap></r>`, xsd,
		"repeats a value")
}

func TestSchemaIdentityNamespaceWildcardAccepted(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
           xmlns:t="https://example.com/id"
           targetNamespace="https://example.com/id"
           elementFormDefault="qualified">
  <xs:element name="r">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="item" maxOccurs="unbounded">
          <xs:complexType><xs:attribute name="id" type="xs:string"/></xs:complexType>
        </xs:element>
      </xs:sequence>
    </xs:complexType>
    <xs:key name="itemKey">
      <xs:selector xpath="t:*"/>
      <xs:field xpath="@id"/>
    </xs:key>
  </xs:element>
</xs:schema>`
	doc := `<?xml version="1.1"?><r xmlns="https://example.com/id"><item id="a"/><item id="%s"/></r>`
	mustSchemaValid(t, sprintfDoc(doc, "b"), xsd)
	mustSchemaReject(t, sprintfDoc(doc, "a"), xsd, "repeats a value")
}

func TestSchemaIdentityOutsideTheSubsetRejected(t *testing.T) {
	tests := []struct {
		name     string
		selector string
		expect   string
	}{
		{"predicate", "item[@id='x']", "predicate"},
		{"function", "count(item)", "predicate"},
		{"parent step", "../item", "predicate"},
		{"descendant axis", "descendant::item", "predicate"},
		{"empty step", "wrap//item", "empty step"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="r">
    <xs:complexType>
      <xs:sequence><xs:element name="item" type="xs:string" maxOccurs="unbounded"/></xs:sequence>
    </xs:complexType>
    <xs:unique name="u">
      <xs:selector xpath="` + tc.selector + `"/>
      <xs:field xpath="."/>
    </xs:unique>
  </xs:element>
</xs:schema>`
			mustSchemaReject(t, `<?xml version="1.1"?><r><item>a</item></r>`, xsd, tc.expect)
		})
	}
}
