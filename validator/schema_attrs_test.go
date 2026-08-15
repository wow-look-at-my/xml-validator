package validator

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A qualified attribute is validated against the global declaration in its own
// namespace. Nothing here is a wildcard: an attribute with no declaration to
// match is an error, which is what lets a schema carry a foreign vocabulary
// without giving up on checking it.
const globalAttrXSD = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
           xmlns:ext="http://ext"
           targetNamespace="http://main"
           xmlns:m="http://main">
  <xs:import namespace="http://ext" schemaLocation="ext.xsd"/>
  <xs:element name="root">
    <xs:complexType>
      <xs:attribute name="id" type="xs:string"/>
      <xs:attribute ref="ext:budget" use="required"/>
    </xs:complexType>
  </xs:element>
</xs:schema>`

const globalAttrExtXSD = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" targetNamespace="http://ext">
  <xs:attribute name="budget" type="xs:int"/>
</xs:schema>`

func globalAttrValidate(t *testing.T, xml string) error {
	t.Helper()
	resolver := func(_, loc string) ([]byte, error) {
		if loc == "ext.xsd" {
			return []byte(globalAttrExtXSD), nil
		}
		return nil, fmt.Errorf("unexpected schemaLocation %q", loc)
	}
	return ValidateWithSchemaResolver([]byte(xml), []byte(globalAttrXSD), resolver)
}

func TestGlobalAttrQualifiedAccepted(t *testing.T) {
	err := globalAttrValidate(t, `<?xml version="1.1"?>`+
		`<root xmlns="http://main" xmlns:ext="http://ext" id="a" ext:budget="1024"/>`)
	require.NoError(t, err)
}

func TestGlobalAttrTypeChecked(t *testing.T) {
	err := globalAttrValidate(t, `<?xml version="1.1"?>`+
		`<root xmlns="http://main" xmlns:ext="http://ext" ext:budget="lots"/>`)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "budget")
}

func TestGlobalAttrUndeclaredRejected(t *testing.T) {
	err := globalAttrValidate(t, `<?xml version="1.1"?>`+
		`<root xmlns="http://main" xmlns:ext="http://ext" ext:budget="1" ext:bugdet="2"/>`)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected attribute")
	assert.Contains(t, err.Error(), "{http://ext}bugdet")
}

func TestGlobalAttrRequiredMissing(t *testing.T) {
	err := globalAttrValidate(t, `<?xml version="1.1"?><root xmlns="http://main" id="a"/>`)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "required attribute")
	assert.Contains(t, err.Error(), "{http://ext}budget")
}

func TestGlobalAttrUnqualifiedDoesNotMatchGlobal(t *testing.T) {
	err := globalAttrValidate(t, `<?xml version="1.1"?>`+
		`<root xmlns="http://main" xmlns:ext="http://ext" ext:budget="1" budget="2"/>`)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected attribute")
}

// A recursive schema -- an element whose content refers back to itself -- is how
// a schema describes an arbitrarily nested tree. Resolution must terminate.
const recursiveXSD = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="param">
    <xs:complexType mixed="true">
      <xs:sequence>
        <xs:element ref="param" minOccurs="0" maxOccurs="unbounded"/>
      </xs:sequence>
      <xs:attribute name="name" type="xs:string"/>
      <xs:attribute name="type" type="xs:string"/>
    </xs:complexType>
  </xs:element>
</xs:schema>`

func TestSchemaRecursiveElement(t *testing.T) {
	mustSchemaValid(t, `<?xml version="1.1"?>`+
		`<param name="a" type="object"><param name="b" type="object">`+
		`<param name="c" type="number">1.50</param></param></param>`, recursiveXSD)
}

func TestSchemaRecursiveElementRejectsUndeclaredChild(t *testing.T) {
	mustSchemaReject(t, `<?xml version="1.1"?>`+
		`<param name="a" type="object"><other/></param>`, recursiveXSD, "unexpected element")
}
