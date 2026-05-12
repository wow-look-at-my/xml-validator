package validator

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/wow-look-at-my/testify/assert"
	"github.com/wow-look-at-my/testify/require"
)

func TestSchemaImportWithoutLocation(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:import namespace="http://www.w3.org/XML/1998/namespace"/>
  <xs:element name="root" type="xs:string"/>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root>val</root>`, xsd)
}

func TestSchemaImportWithLocationRequiresResolver(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:import namespace="http://other" schemaLocation="other.xsd"/>
  <xs:element name="root" type="xs:string"/>
</xs:schema>`
	err := schemaValidate(t, `<?xml version="1.1"?><root>val</root>`, xsd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "schema resolver")
}

func TestSchemaImportResolved(t *testing.T) {
	mainXSD := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:import namespace="http://other" schemaLocation="other.xsd"/>
  <xs:element name="root">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="amount" type="moneyType"/>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	importedXSD := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" targetNamespace="http://other">
  <xs:simpleType name="moneyType">
    <xs:restriction base="xs:decimal">
      <xs:minInclusive value="0"/>
    </xs:restriction>
  </xs:simpleType>
</xs:schema>`
	resolver := func(_, loc string) ([]byte, error) {
		if loc == "other.xsd" {
			return []byte(importedXSD), nil
		}
		return nil, fmt.Errorf("unexpected schemaLocation %q", loc)
	}
	err := ValidateWithSchemaResolver(
		[]byte(`<?xml version="1.1"?><root><amount>5.00</amount></root>`),
		[]byte(mainXSD), resolver)
	require.NoError(t, err)

	err = ValidateWithSchemaResolver(
		[]byte(`<?xml version="1.1"?><root><amount>-1.00</amount></root>`),
		[]byte(mainXSD), resolver)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be >= 0")
}

func TestSchemaImportFromFile(t *testing.T) {
	dir := t.TempDir()
	mainPath := filepath.Join(dir, "main.xsd")
	otherPath := filepath.Join(dir, "other.xsd")
	xmlPath := filepath.Join(dir, "doc.xml")

	mainXSD := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:import namespace="http://other" schemaLocation="other.xsd"/>
  <xs:element name="root">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="label" type="labelType"/>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	otherXSD := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" targetNamespace="http://other">
  <xs:simpleType name="labelType">
    <xs:restriction base="xs:string">
      <xs:enumeration value="alpha"/>
      <xs:enumeration value="beta"/>
    </xs:restriction>
  </xs:simpleType>
</xs:schema>`
	require.NoError(t, os.WriteFile(mainPath, []byte(mainXSD), 0o600))
	require.NoError(t, os.WriteFile(otherPath, []byte(otherXSD), 0o600))
	require.NoError(t, os.WriteFile(xmlPath, []byte(`<?xml version="1.1"?><root><label>alpha</label></root>`), 0o600))

	require.NoError(t, ValidateWithSchemaFile(xmlPath, mainPath))

	require.NoError(t, os.WriteFile(xmlPath, []byte(`<?xml version="1.1"?><root><label>gamma</label></root>`), 0o600))
	err := ValidateWithSchemaFile(xmlPath, mainPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not one of the allowed values")
}

func TestSchemaImportTransitive(t *testing.T) {
	aXSD := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:import namespace="http://b" schemaLocation="b.xsd"/>
  <xs:element name="root" type="rootType"/>
</xs:schema>`
	bXSD := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" targetNamespace="http://b">
  <xs:import namespace="http://c" schemaLocation="c.xsd"/>
  <xs:complexType name="rootType">
    <xs:sequence>
      <xs:element name="leaf" type="leafType"/>
    </xs:sequence>
  </xs:complexType>
</xs:schema>`
	cXSD := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" targetNamespace="http://c">
  <xs:simpleType name="leafType">
    <xs:restriction base="xs:integer"/>
  </xs:simpleType>
</xs:schema>`
	resolver := func(_, loc string) ([]byte, error) {
		switch loc {
		case "b.xsd":
			return []byte(bXSD), nil
		case "c.xsd":
			return []byte(cXSD), nil
		}
		return nil, fmt.Errorf("unexpected %q", loc)
	}
	err := ValidateWithSchemaResolver(
		[]byte(`<?xml version="1.1"?><root><leaf>42</leaf></root>`),
		[]byte(aXSD), resolver)
	require.NoError(t, err)
}

func TestSchemaImportCycle(t *testing.T) {
	aXSD := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:import namespace="http://b" schemaLocation="b.xsd"/>
  <xs:element name="root" type="xs:string"/>
</xs:schema>`
	bXSD := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" targetNamespace="http://b">
  <xs:import namespace="" schemaLocation="a.xsd"/>
</xs:schema>`
	resolver := func(_, loc string) ([]byte, error) {
		switch loc {
		case "a.xsd":
			return []byte(aXSD), nil
		case "b.xsd":
			return []byte(bXSD), nil
		}
		return nil, fmt.Errorf("unexpected %q", loc)
	}
	err := ValidateWithSchemaResolver(
		[]byte(`<?xml version="1.1"?><root>hi</root>`),
		[]byte(aXSD), resolver)
	require.NoError(t, err)
}

func TestSchemaImportResolverError(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:import namespace="http://other" schemaLocation="missing.xsd"/>
  <xs:element name="root" type="xs:string"/>
</xs:schema>`
	resolver := func(_, _ string) ([]byte, error) {
		return nil, fmt.Errorf("not found")
	}
	err := ValidateWithSchemaResolver(
		[]byte(`<?xml version="1.1"?><root>val</root>`),
		[]byte(xsd), resolver)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing.xsd")
}

func TestSchemaImportNilDataSkips(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:import namespace="http://other" schemaLocation="skip.xsd"/>
  <xs:element name="root" type="xs:string"/>
</xs:schema>`
	resolver := func(_, _ string) ([]byte, error) {
		return nil, nil
	}
	err := ValidateWithSchemaResolver(
		[]byte(`<?xml version="1.1"?><root>val</root>`),
		[]byte(xsd), resolver)
	require.NoError(t, err)
}

func TestSchemaImportEmptySliceErrors(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:import namespace="http://other" schemaLocation="empty.xsd"/>
  <xs:element name="root" type="xs:string"/>
</xs:schema>`
	resolver := func(_, _ string) ([]byte, error) {
		return []byte{}, nil
	}
	err := ValidateWithSchemaResolver(
		[]byte(`<?xml version="1.1"?><root>val</root>`),
		[]byte(xsd), resolver)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty.xsd")
}

func TestSchemaImportNamespaceCollision(t *testing.T) {
	mainXSD := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:import namespace="http://a" schemaLocation="a.xsd"/>
  <xs:import namespace="http://b" schemaLocation="b.xsd"/>
  <xs:element name="root" type="xs:string"/>
</xs:schema>`
	aXSD := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" targetNamespace="http://a">
  <xs:simpleType name="shared">
    <xs:restriction base="xs:string"/>
  </xs:simpleType>
</xs:schema>`
	bXSD := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" targetNamespace="http://b">
  <xs:simpleType name="shared">
    <xs:restriction base="xs:integer"/>
  </xs:simpleType>
</xs:schema>`
	resolver := func(_, loc string) ([]byte, error) {
		switch loc {
		case "a.xsd":
			return []byte(aXSD), nil
		case "b.xsd":
			return []byte(bXSD), nil
		}
		return nil, fmt.Errorf("unexpected %q", loc)
	}
	err := ValidateWithSchemaResolver(
		[]byte(`<?xml version="1.1"?><root>val</root>`),
		[]byte(mainXSD), resolver)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `type "shared"`)
	assert.Contains(t, err.Error(), "more than once")
}

func TestSchemaImportCycleKeyWithPipe(t *testing.T) {
	mainXSD := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:import namespace="ns|a" schemaLocation="x.xsd"/>
  <xs:import namespace="ns" schemaLocation="a|x.xsd"/>
  <xs:element name="root" type="xs:string"/>
</xs:schema>`
	calls := 0
	resolver := func(ns, loc string) ([]byte, error) {
		calls++
		return []byte(fmt.Sprintf(`<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" targetNamespace=%q/>`, ns)), nil
	}
	err := ValidateWithSchemaResolver(
		[]byte(`<?xml version="1.1"?><root>val</root>`),
		[]byte(mainXSD), resolver)
	require.NoError(t, err)
	assert.Equal(t, 2, calls, "both imports should resolve despite the ns+|+loc concatenation matching")
}
