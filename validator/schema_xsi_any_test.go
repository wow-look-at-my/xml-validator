package validator

import "testing"

func TestSchemaXSIAttributesAccepted(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root" type="xs:string"/>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:noNamespaceSchemaLocation="test.xsd">hello</root>`, xsd)
	mustSchemaValid(t, `<?xml version="1.1"?><root xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://example.com test.xsd">hello</root>`, xsd)
	mustSchemaValid(t, `<?xml version="1.1"?><root xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">hello</root>`, xsd)
}

func TestSchemaAnyInAllGroup(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" targetNamespace="http://example.com/main" xmlns:tns="http://example.com/main" elementFormDefault="qualified">
  <xs:element name="event">
    <xs:complexType>
      <xs:all>
        <xs:element name="date" type="xs:string"/>
        <xs:element name="place" type="xs:string" minOccurs="0"/>
        <xs:any namespace="http://example.com/ext" processContents="lax" minOccurs="0" maxOccurs="unbounded"/>
      </xs:all>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><tns:event xmlns:tns="http://example.com/main" xmlns:ext="http://example.com/ext"><tns:date>2025-01-01</tns:date><ext:note>test</ext:note></tns:event>`, xsd)
	mustSchemaValid(t, `<?xml version="1.1"?><tns:event xmlns:tns="http://example.com/main" xmlns:ext="http://example.com/ext"><tns:date>2025-01-01</tns:date><ext:note>a</ext:note><ext:detail>b</ext:detail></tns:event>`, xsd)
	mustSchemaValid(t, `<?xml version="1.1"?><tns:event xmlns:tns="http://example.com/main"><tns:date>2025-01-01</tns:date></tns:event>`, xsd)
}

func TestSchemaAnyInAllGroupWrongNamespace(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" targetNamespace="http://example.com/main" xmlns:tns="http://example.com/main" elementFormDefault="qualified">
  <xs:element name="event">
    <xs:complexType>
      <xs:all>
        <xs:element name="date" type="xs:string"/>
        <xs:any namespace="http://example.com/ext" processContents="lax" minOccurs="0" maxOccurs="unbounded"/>
      </xs:all>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><tns:event xmlns:tns="http://example.com/main" xmlns:bad="http://wrong.example"><tns:date>2025-01-01</tns:date><bad:note>test</bad:note></tns:event>`, xsd, "unexpected element")
}

func TestSchemaAnyInAllGroupAnyNamespace(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:all>
        <xs:element name="name" type="xs:string"/>
        <xs:any namespace="##any" processContents="skip" minOccurs="0" maxOccurs="unbounded"/>
      </xs:all>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root xmlns:ext="http://x"><name>hi</name><ext:foo>bar</ext:foo></root>`, xsd)
}

func TestSchemaAnyNamespaceMatchInSequence(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:sequence>
        <xs:any namespace="http://example.com/ok" processContents="skip" minOccurs="0" maxOccurs="unbounded"/>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root xmlns:ok="http://example.com/ok"><ok:a>1</ok:a></root>`, xsd)
	mustSchemaValid(t, `<?xml version="1.1"?><root/>`, xsd)
}

func TestSchemaAnyInAllMinOccurs(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:all>
        <xs:element name="name" type="xs:string"/>
        <xs:any namespace="##any" processContents="skip" minOccurs="1" maxOccurs="unbounded"/>
      </xs:all>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><root><name>hi</name></root>`, xsd, "xs:any wildcard")
	mustSchemaValid(t, `<?xml version="1.1"?><root xmlns:e="http://x"><name>hi</name><e:extra>ok</e:extra></root>`, xsd)
}

func TestSchemaAnyNamespaceLocal(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:all>
        <xs:element name="name" type="xs:string"/>
        <xs:any namespace="##local" processContents="skip" minOccurs="0" maxOccurs="unbounded"/>
      </xs:all>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root><name>hi</name><extra>ok</extra></root>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><root xmlns:ns="http://x"><name>hi</name><ns:extra>no</ns:extra></root>`, xsd, "unexpected element")
}

func TestSchemaAnyNamespaceOther(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" targetNamespace="http://example.com/main" xmlns:tns="http://example.com/main" elementFormDefault="qualified">
  <xs:element name="root">
    <xs:complexType>
      <xs:all>
        <xs:element name="name" type="xs:string"/>
        <xs:any namespace="##other" processContents="skip" minOccurs="0" maxOccurs="unbounded"/>
      </xs:all>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><tns:root xmlns:tns="http://example.com/main" xmlns:ext="http://other"><tns:name>hi</tns:name><ext:extra>ok</ext:extra></tns:root>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><tns:root xmlns:tns="http://example.com/main"><tns:name>hi</tns:name><tns:extra>no</tns:extra></tns:root>`, xsd, "unexpected element")
}

func TestSchemaAnyNamespaceTargetNamespace(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" targetNamespace="http://example.com/main" xmlns:tns="http://example.com/main" elementFormDefault="qualified">
  <xs:element name="root">
    <xs:complexType>
      <xs:all>
        <xs:element name="name" type="xs:string"/>
        <xs:any namespace="##targetNamespace" processContents="skip" minOccurs="0" maxOccurs="unbounded"/>
      </xs:all>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><tns:root xmlns:tns="http://example.com/main"><tns:name>hi</tns:name><tns:extra>ok</tns:extra></tns:root>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><tns:root xmlns:tns="http://example.com/main" xmlns:ext="http://other"><tns:name>hi</tns:name><ext:extra>no</ext:extra></tns:root>`, xsd, "unexpected element")
}

func TestSchemaAnyNamespaceMultiToken(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:all>
        <xs:element name="name" type="xs:string"/>
        <xs:any namespace="##local http://example.com/ext" processContents="skip" minOccurs="0" maxOccurs="unbounded"/>
      </xs:all>
    </xs:complexType>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><root><name>hi</name><local>ok</local></root>`, xsd)
	mustSchemaValid(t, `<?xml version="1.1"?><root xmlns:ext="http://example.com/ext"><name>hi</name><ext:extra>ok</ext:extra></root>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><root xmlns:bad="http://wrong"><name>hi</name><bad:x>no</bad:x></root>`, xsd, "unexpected element")
}
