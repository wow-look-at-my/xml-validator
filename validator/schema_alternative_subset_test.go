package validator

import "testing"

// XSD 1.1 defines a "required subset" of XPath 2.0 for xs:alternative/@test and
// says a conforming processor must accept and process it. Everything here is
// inside that subset, so rejecting any of it would refuse a conforming schema.

// numeric selects an xs:int content type, textual an xs:string one, so the
// document tells us which alternative won: a non-numeric body is an error only
// when the numeric type was chosen.
const subsetXSD = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:complexType name="numeric">
    <xs:simpleContent><xs:extension base="xs:int">
      <xs:attribute name="a" type="xs:string"/>
      <xs:attribute name="b" type="xs:string"/>
    </xs:extension></xs:simpleContent>
  </xs:complexType>
  <xs:complexType name="textual">
    <xs:simpleContent><xs:extension base="xs:string">
      <xs:attribute name="a" type="xs:string"/>
      <xs:attribute name="b" type="xs:string"/>
    </xs:extension></xs:simpleContent>
  </xs:complexType>
  <xs:element name="v" type="textual">
    <xs:alternative test="TEST" type="numeric"/>
  </xs:element>
</xs:schema>`

// chose reports the schema with TEST replaced, and asserts whether the numeric
// alternative won for the given attributes.
func assertChoice(t *testing.T, test, attrs string, wantNumeric bool) {
	t.Helper()
	xsd := replaceOnce(subsetXSD, "TEST", test)
	doc := `<?xml version="1.1"?><v ` + attrs + `>words</v>`
	if wantNumeric {
		mustSchemaReject(t, doc, xsd, "not a valid int")
		return
	}
	mustSchemaValid(t, doc, xsd)
}

func replaceOnce(s, old, new string) string {
	i := indexOf(s, old)
	if i < 0 {
		return s
	}
	return s[:i] + new + s[i+len(old):]
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func TestAlternativeOrderComparators(t *testing.T) {
	// &lt; and &gt; are the escaped forms an XSD author has to write.
	assertChoice(t, "@a &gt; 5", `a="7"`, true)
	assertChoice(t, "@a &gt; 5", `a="3"`, false)
	assertChoice(t, "@a &lt; 5", `a="3"`, true)
	assertChoice(t, "@a &gt;= 5", `a="5"`, true)
	assertChoice(t, "@a &lt;= 5", `a="5"`, true)
	assertChoice(t, "@a &lt;= 5", `a="6"`, false)
}

func TestAlternativeNumericLiteral(t *testing.T) {
	// An unquoted number is a Literal in the subset: the untyped attribute is
	// compared as a number, so "01" and "1" are the same value.
	assertChoice(t, "@a = 1", `a="1"`, true)
	assertChoice(t, "@a = 1", `a="01"`, true)
	assertChoice(t, "@a = 1", `a="2"`, false)
	assertChoice(t, "@a = 1.5", `a="1.50"`, true)
}

func TestAlternativeStringComparisonStaysTextual(t *testing.T) {
	// With a quoted literal the comparison is textual, so "01" is not "1".
	assertChoice(t, "@a = '1'", `a="1"`, true)
	assertChoice(t, "@a = '1'", `a="01"`, false)
}

func TestAlternativeAndOrPrecedence(t *testing.T) {
	// "and" binds tighter than "or", so this reads (a and b) or c.
	test := "@a='1' and @b='2' or @a='9'"
	assertChoice(t, test, `a="1" b="2"`, true)
	assertChoice(t, test, `a="9" b="0"`, true)
	assertChoice(t, test, `a="1" b="0"`, false)
}

func TestAlternativeParentheses(t *testing.T) {
	// Parentheses override that precedence: a and (b or c).
	test := "@a='1' and (@b='2' or @b='3')"
	assertChoice(t, test, `a="1" b="3"`, true)
	assertChoice(t, test, `a="1" b="4"`, false)
	assertChoice(t, test, `a="0" b="2"`, false)
}

func TestAlternativeNotOverAnExpression(t *testing.T) {
	// not() takes a whole OrExpr, not just one term.
	assertChoice(t, "not(@a='1' or @a='2')", `a="3"`, true)
	assertChoice(t, "not(@a='1' or @a='2')", `a="2"`, false)
	assertChoice(t, "not(@a)", ``, true)
	assertChoice(t, "not(@a)", `a="x"`, false)
}

func TestAlternativeAttributeAgainstAttribute(t *testing.T) {
	assertChoice(t, "@a = @b", `a="x" b="x"`, true)
	assertChoice(t, "@a = @b", `a="x" b="y"`, false)
	// An absent attribute makes the comparison false either way.
	assertChoice(t, "@a = @b", `a="x"`, false)
	assertChoice(t, "@a != @b", `a="x"`, false)
}

func TestAlternativeConstructorFunction(t *testing.T) {
	// xs:int(@a) fails to construct when the value is not an int, and a failed
	// test simply does not select the alternative.
	assertChoice(t, "xs:int(@a) = 5", `a="5"`, true)
	assertChoice(t, "xs:int(@a) = 5", `a="six"`, false)
	assertChoice(t, "xs:boolean(@a) = 'true'", `a="true"`, true)
}

func TestAlternativeCastAs(t *testing.T) {
	assertChoice(t, "@a cast as xs:int = 5", `a="5"`, true)
	assertChoice(t, "@a cast as xs:int = 5", `a="notanint"`, false)
	// The "?" form tolerates an absent attribute rather than erroring.
	assertChoice(t, "@a cast as xs:int? = 5", `a="5"`, true)
	assertChoice(t, "@a cast as xs:int? = 5", ``, false)
}

func TestAlternativeAttributePresence(t *testing.T) {
	assertChoice(t, "@a", `a="anything"`, true)
	assertChoice(t, "@a", `b="x"`, false)
	// An empty attribute has an effective boolean value of false.
	assertChoice(t, "@a", `a=""`, false)
}
