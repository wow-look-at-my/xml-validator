package validator

import "testing"

// The schema the issue reduced from: a key on provider/@id and a keyref from
// role/@provider.
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

func TestSchemaKeyCatchesDuplicate(t *testing.T) {
	mustSchemaValid(t, `<?xml version="1.1"?>
<cfg>
  <provider id="openrouter"/>
  <provider id="anthropic"/>
  <role provider="openrouter"/>
</cfg>`, identityXSD)

	mustSchemaReject(t, `<?xml version="1.1"?>
<cfg>
  <provider id="openrouter"/>
  <provider id="openrouter"/>
</cfg>`, identityXSD, `xs:key "providerKey": element "provider" repeats a value`)
}

func TestSchemaKeyrefCatchesDanglingReference(t *testing.T) {
	mustSchemaReject(t, `<?xml version="1.1"?>
<cfg>
  <provider id="openrouter"/>
  <role provider="does-not-exist"/>
</cfg>`, identityXSD, `xs:keyref "roleProvider": element "role" refers to a value that "providerKey" does not declare`)
}

func TestSchemaKeyRequiresEveryField(t *testing.T) {
	mustSchemaReject(t, `<?xml version="1.1"?>
<cfg>
  <provider/>
</cfg>`, identityXSD, `xs:key "providerKey": element "provider" has no value for field "@id"`)
}

func TestSchemaUniqueAllowsAbsentField(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="rules">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="rule" maxOccurs="unbounded">
          <xs:complexType><xs:attribute name="ref" type="xs:string"/></xs:complexType>
        </xs:element>
      </xs:sequence>
    </xs:complexType>
    <xs:unique name="ruleRef">
      <xs:selector xpath="rule"/>
      <xs:field xpath="@ref"/>
    </xs:unique>
  </xs:element>
</xs:schema>`
	// Two rules carry no ref at all: xs:unique does not count a node whose
	// field selects nothing, so this is not a duplicate.
	mustSchemaValid(t, `<?xml version="1.1"?><rules><rule/><rule/><rule ref="a"/></rules>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><rules><rule ref="a"/><rule ref="a"/></rules>`, xsd,
		`xs:unique "ruleRef": element "rule" repeats a value`)
}

func TestSchemaKeyComparesInValueSpace(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="rows">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="row" maxOccurs="unbounded">
          <xs:complexType>
            <xs:attribute name="n" type="xs:integer"/>
            <xs:attribute name="s" type="xs:string"/>
          </xs:complexType>
        </xs:element>
      </xs:sequence>
    </xs:complexType>
    <xs:key name="rowNum">
      <xs:selector xpath="row"/>
      <xs:field xpath="@n"/>
    </xs:key>
  </xs:element>
</xs:schema>`
	// 01 and 1 are one integer, so they collide.
	mustSchemaReject(t, `<?xml version="1.1"?><rows><row n="01"/><row n="1"/></rows>`, xsd, "repeats a value")
	mustSchemaValid(t, `<?xml version="1.1"?><rows><row n="1"/><row n="2"/></rows>`, xsd)
}

func TestSchemaKeyOverElementTextAndMultipleFields(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="people">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="person" maxOccurs="unbounded">
          <xs:complexType>
            <xs:sequence>
              <xs:element name="first" type="xs:string"/>
              <xs:element name="last" type="xs:string"/>
            </xs:sequence>
          </xs:complexType>
        </xs:element>
      </xs:sequence>
    </xs:complexType>
    <xs:key name="fullName">
      <xs:selector xpath="person"/>
      <xs:field xpath="first"/>
      <xs:field xpath="last"/>
    </xs:key>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><people>`+
		`<person><first>ada</first><last>lovelace</last></person>`+
		`<person><first>ada</first><last>byron</last></person></people>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><people>`+
		`<person><first>ada</first><last>lovelace</last></person>`+
		`<person><first>ada</first><last>lovelace</last></person></people>`, xsd, "repeats a value")
}

// A prefixed selector against a target namespace, which is the shape a real
// schema uses -- and the shape an unprefixed name would silently miss.
const identityNSXSD = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
           xmlns:c="https://slh.dev/xsd/config/1"
           targetNamespace="https://slh.dev/xsd/config/1"
           elementFormDefault="qualified">
  <xs:element name="slh">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="providers">
          <xs:complexType><xs:sequence>
            <xs:element name="provider" maxOccurs="unbounded">
              <xs:complexType><xs:attribute name="id" type="xs:string"/></xs:complexType>
            </xs:element>
          </xs:sequence></xs:complexType>
        </xs:element>
        <xs:element name="roles" minOccurs="0">
          <xs:complexType><xs:sequence>
            <xs:element name="role" maxOccurs="unbounded">
              <xs:complexType><xs:attribute name="provider" type="xs:string"/></xs:complexType>
            </xs:element>
          </xs:sequence></xs:complexType>
        </xs:element>
      </xs:sequence>
    </xs:complexType>
    <xs:key name="providerId">
      <xs:selector xpath="c:providers/c:provider"/>
      <xs:field xpath="@id"/>
    </xs:key>
    <xs:keyref name="roleProviderRef" refer="c:providerId">
      <xs:selector xpath="c:roles/c:role"/>
      <xs:field xpath="@provider"/>
    </xs:keyref>
  </xs:element>
</xs:schema>`

func TestSchemaIdentityAcrossNamespacedPath(t *testing.T) {
	mustSchemaValid(t, `<?xml version="1.1"?>
<slh xmlns="https://slh.dev/xsd/config/1">
  <providers><provider id="openrouter"/><provider id="anthropic"/></providers>
  <roles><role provider="anthropic"/></roles>
</slh>`, identityNSXSD)
	mustSchemaReject(t, `<?xml version="1.1"?>
<slh xmlns="https://slh.dev/xsd/config/1">
  <providers><provider id="openrouter"/><provider id="openrouter"/></providers>
</slh>`, identityNSXSD, "repeats a value")
	mustSchemaReject(t, `<?xml version="1.1"?>
<slh xmlns="https://slh.dev/xsd/config/1">
  <providers><provider id="openrouter"/></providers>
  <roles><role provider="nope"/></roles>
</slh>`, identityNSXSD, "does not declare")
}

func TestSchemaKeyrefResolvesAgainstAncestorScope(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="doc">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="item" maxOccurs="unbounded">
          <xs:complexType><xs:attribute name="id" type="xs:string"/></xs:complexType>
        </xs:element>
        <xs:element name="links">
          <xs:complexType>
            <xs:sequence>
              <xs:element name="link" maxOccurs="unbounded">
                <xs:complexType><xs:attribute name="to" type="xs:string"/></xs:complexType>
              </xs:element>
            </xs:sequence>
          </xs:complexType>
          <xs:keyref name="linkTarget" refer="itemId">
            <xs:selector xpath="link"/>
            <xs:field xpath="@to"/>
          </xs:keyref>
        </xs:element>
      </xs:sequence>
    </xs:complexType>
    <xs:key name="itemId">
      <xs:selector xpath="item"/>
      <xs:field xpath="@id"/>
    </xs:key>
  </xs:element>
</xs:schema>`
	// The keyref sits on <links>; the key it names is evaluated on <doc>.
	mustSchemaValid(t, `<?xml version="1.1"?><doc><item id="a"/><links><link to="a"/></links></doc>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><doc><item id="a"/><links><link to="b"/></links></doc>`, xsd,
		"does not declare")
}

func TestSchemaKeyScopesToEachElementInstance(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="doc">
    <xs:complexType><xs:sequence>
      <xs:element ref="box" maxOccurs="unbounded"/>
    </xs:sequence></xs:complexType>
  </xs:element>
  <xs:element name="box">
    <xs:complexType><xs:sequence>
      <xs:element name="item" maxOccurs="unbounded">
        <xs:complexType><xs:attribute name="id" type="xs:string"/></xs:complexType>
      </xs:element>
    </xs:sequence></xs:complexType>
    <xs:key name="itemId">
      <xs:selector xpath="item"/>
      <xs:field xpath="@id"/>
    </xs:key>
  </xs:element>
</xs:schema>`
	// The key is declared on box, so each box is its own scope: two boxes may
	// each hold an item with the same id. A ref particle carries the
	// constraints of the declaration it names, which is what makes this run.
	mustSchemaValid(t, `<?xml version="1.1"?><doc>`+
		`<box><item id="a"/></box><box><item id="a"/></box></doc>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><doc>`+
		`<box><item id="a"/><item id="a"/></box></doc>`, xsd, "repeats a value")
}

func TestSchemaIdentityDescendantSelector(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="tree">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="branch" maxOccurs="unbounded">
          <xs:complexType><xs:sequence>
            <xs:element name="leaf" maxOccurs="unbounded">
              <xs:complexType><xs:attribute name="id" type="xs:string"/></xs:complexType>
            </xs:element>
          </xs:sequence></xs:complexType>
        </xs:element>
      </xs:sequence>
    </xs:complexType>
    <xs:key name="leafId">
      <xs:selector xpath=".//leaf"/>
      <xs:field xpath="@id"/>
    </xs:key>
  </xs:element>
</xs:schema>`
	mustSchemaValid(t, `<?xml version="1.1"?><tree><branch><leaf id="a"/></branch><branch><leaf id="b"/></branch></tree>`, xsd)
	mustSchemaReject(t, `<?xml version="1.1"?><tree><branch><leaf id="a"/></branch><branch><leaf id="a"/></branch></tree>`, xsd,
		"repeats a value")
}

func TestSchemaUnsupportedXPathRejected(t *testing.T) {
	tests := []struct {
		name   string
		xpath  string
		expect string
	}{
		{"predicate", "provider[@id='x']", "is not a name"},
		{"function", "count(provider)", "is not a name"},
		{"parent axis", "../provider", "is not a name"},
		// child:: and attribute:: are in the grammar XSD states; any other axis
		// is not.
		{"axis", "descendant::provider", "is not a name"},
		{"empty step", "providers//provider", "an empty step"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="cfg">
    <xs:complexType>
      <xs:sequence><xs:element name="provider" minOccurs="0"/></xs:sequence>
    </xs:complexType>
    <xs:key name="k">
      <xs:selector xpath="` + tc.xpath + `"/>
      <xs:field xpath="@id"/>
    </xs:key>
  </xs:element>
</xs:schema>`
			mustSchemaReject(t, `<?xml version="1.1"?><cfg/>`, xsd, tc.expect)
		})
	}
}

func TestSchemaSelectorRejectsAttributeStep(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="cfg">
    <xs:complexType><xs:attribute name="id" type="xs:string"/></xs:complexType>
    <xs:key name="k">
      <xs:selector xpath="@id"/>
      <xs:field xpath="@id"/>
    </xs:key>
  </xs:element>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><cfg id="a"/>`, xsd, "allowed in xs:field")
}

func TestSchemaKeyrefWithUnknownReferRejected(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="cfg">
    <xs:complexType>
      <xs:sequence><xs:element name="role" minOccurs="0"/></xs:sequence>
    </xs:complexType>
    <xs:keyref name="r" refer="noSuchKey">
      <xs:selector xpath="role"/>
      <xs:field xpath="@provider"/>
    </xs:keyref>
  </xs:element>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><cfg/>`, xsd, "names no xs:key or xs:unique")
}

func TestSchemaDuplicateConstraintNameRejected(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="cfg">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="a" minOccurs="0"/>
        <xs:element name="b" minOccurs="0"/>
      </xs:sequence>
    </xs:complexType>
    <xs:key name="dup"><xs:selector xpath="a"/><xs:field xpath="@id"/></xs:key>
    <xs:unique name="dup"><xs:selector xpath="b"/><xs:field xpath="@id"/></xs:unique>
  </xs:element>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><cfg/>`, xsd, "declared twice")
}

func TestSchemaFieldSelectingTwoNodesRejected(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="rows">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="row" maxOccurs="unbounded">
          <xs:complexType><xs:sequence>
            <xs:element name="v" type="xs:string" maxOccurs="unbounded"/>
          </xs:sequence></xs:complexType>
        </xs:element>
      </xs:sequence>
    </xs:complexType>
    <xs:key name="rowV">
      <xs:selector xpath="row"/>
      <xs:field xpath="v"/>
    </xs:key>
  </xs:element>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><rows><row><v>a</v><v>b</v></row></rows>`, xsd,
		"a field must select at most one")
}

func TestSchemaConstraintMissingSelectorRejected(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="cfg" type="xs:string">
    <xs:key name="k"><xs:field xpath="@id"/></xs:key>
  </xs:element>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><cfg>x</cfg>`, xsd, "requires an xs:selector")
}

func TestSchemaUnknownElementChildRejected(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="r" type="xs:string">
    <xs:assert test="true()"/>
  </xs:element>
</xs:schema>`
	mustSchemaReject(t, `<?xml version="1.1"?><r>x</r>`, xsd, "unsupported schema element xs:assert")
}
