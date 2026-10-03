package validator

import (
	"fmt"
	"strings"
	"testing"
)

// The length facets count characters, not the bytes of the UTF-8 encoding.
// see docs/nul-char-ref.md

func lengthXSD(kind string, n int) string {
	return fmt.Sprintf(`<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="v">
    <xs:simpleType>
      <xs:restriction base="xs:string">
        <xs:%s value="%d"/>
      </xs:restriction>
    </xs:simpleType>
  </xs:element>
</xs:schema>`, kind, n)
}

func TestLengthFacetCountsCharactersNotBytes(t *testing.T) {
	// "é" is 2 bytes, "æ" 2, "文" 3, "𝄞" 4: 4 characters, 11 bytes.
	doc := `<?xml version="1.1"?><v>éæ文𝄞</v>`

	mustSchemaValid(t, doc, lengthXSD("length", 4))
	mustSchemaReject(t, doc, lengthXSD("length", 11), "value length 4")
	mustSchemaValid(t, doc, lengthXSD("maxLength", 4))
	mustSchemaValid(t, doc, lengthXSD("minLength", 4))
	mustSchemaReject(t, doc, lengthXSD("maxLength", 3), "value length 4 exceeds maxLength 3")
}

// Every byte value 0 through 255 as a character reference: 256 characters,
// which UTF-8 writes as 384 bytes.
func TestLengthFacetOverEveryByteValue(t *testing.T) {
	var refs strings.Builder
	for i := range 256 {
		fmt.Fprintf(&refs, "&#%d;", i)
	}
	doc := `<?xml version="1.1"?><v>` + refs.String() + `</v>`

	mustSchemaValid(t, doc, lengthXSD("length", 256))
	mustSchemaReject(t, doc, lengthXSD("length", 384), "value length 256")
}

func binaryLengthXSD(base string, n int) string {
	return fmt.Sprintf(`<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="v">
    <xs:simpleType>
      <xs:restriction base="xs:%s">
        <xs:length value="%d"/>
      </xs:restriction>
    </xs:simpleType>
  </xs:element>
</xs:schema>`, base, n)
}

// The binary types measure octets, so their length is what the value decodes
// to: four hex digit pairs are four octets, not eight characters.
func TestLengthFacetOnBinaryTypesCountsOctets(t *testing.T) {
	hex := `<?xml version="1.1"?><v>00FF10AB</v>`
	mustSchemaValid(t, hex, binaryLengthXSD("hexBinary", 4))
	mustSchemaReject(t, hex, binaryLengthXSD("hexBinary", 8), "value length 4")

	// "QUJDRA==" decodes to "ABCD".
	b64 := `<?xml version="1.1"?><v>QUJDRA==</v>`
	mustSchemaValid(t, b64, binaryLengthXSD("base64Binary", 4))
	mustSchemaReject(t, b64, binaryLengthXSD("base64Binary", 8), "value length 4")
}

// A facet value that is not a length is an error in the schema, never a
// length of 0 that admits every value.
func TestLengthFacetRejectsAMalformedValue(t *testing.T) {
	doc := `<?xml version="1.1"?><v>abc</v>`
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="v">
    <xs:simpleType>
      <xs:restriction base="xs:string">
        <xs:minLength value="two"/>
      </xs:restriction>
    </xs:simpleType>
  </xs:element>
</xs:schema>`

	mustSchemaReject(t, doc, xsd, `minLength facet value "two" is not a non-negative integer`)
}
