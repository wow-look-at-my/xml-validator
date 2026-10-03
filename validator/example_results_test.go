package validator_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wow-look-at-my/xml-validator/validator"
)

// The examples print their results. This test holds each result, so an example cannot drift from the code.
func TestTheExamplesHoldTheirResults(t *testing.T) {
	assert.NoError(t, validator.Validate(strings.NewReader(`<?xml version="1.1"?><greeting>hi</greeting>`)))
	assert.Error(t, validator.Validate(strings.NewReader(`<?xml version="1.1"?><r>&lt;</x>`)))

	xml := `<?xml version="1.1"?><note><body>hi</body></note>`
	xsd := `<?xml version="1.1"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="note">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="body" type="xs:string"/>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	assert.NoError(t, validator.ValidateWithSchema(strings.NewReader(xml), strings.NewReader(xsd)))

	doc, err := validator.ParseTree(strings.NewReader(`<?xml version="1.1"?><r a="1"><c/></r>`))
	require.NoError(t, err)
	assert.Equal(t, "r", doc.Root.Local)
	assert.Len(t, doc.Root.Attrs, 1)
	assert.Len(t, doc.Root.ChildElements(), 1)

	var vErr *validator.Error
	require.ErrorAs(t, validator.Validate(strings.NewReader(`<?xml version="2.0"?><r/>`)), &vErr)
	assert.Equal(t, 1, vErr.Line)
}
