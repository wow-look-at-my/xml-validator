package validator

import (
	"fmt"
	"strings"
	"testing"
	"github.com/wow-look-at-my/testify/assert"
	"github.com/wow-look-at-my/testify/require"
)

func mustValidate(t *testing.T, input string) {
	t.Helper()
	require.NoError(t, Validate(strings.NewReader(input)))

}

func mustReject(t *testing.T, input string, wantSubstr string) {
	t.Helper()
	err := Validate(strings.NewReader(input))
	require.NotNil(t, err)

	require.Contains(t, err.Error(), wantSubstr)

}

func TestMinimalDocument(t *testing.T) {
	mustValidate(t, `<?xml version="1.1"?><r/>`)
}

func TestMinimalDocumentWithEncoding(t *testing.T) {
	mustValidate(t, `<?xml version="1.1" encoding="UTF-8"?><r/>`)
}

func TestDocumentWithStandalone(t *testing.T) {
	mustValidate(t, `<?xml version="1.1" encoding="UTF-8" standalone="yes"?><root/>`)
}

func TestStandaloneNo(t *testing.T) {
	mustValidate(t, `<?xml version="1.1" standalone="no"?><r/>`)
}

func TestSingleQuotesInDecl(t *testing.T) {
	mustValidate(t, `<?xml version='1.1' encoding='UTF-8'?><r/>`)
}

func TestNestedElements(t *testing.T) {
	mustValidate(t, `<?xml version="1.1"?><a><b><c/></b></a>`)
}

func TestTextContent(t *testing.T) {
	mustValidate(t, `<?xml version="1.1"?><r>hello world</r>`)
}

func TestAttributes(t *testing.T) {
	mustValidate(t, `<?xml version="1.1"?><r a="1" b='2'/>`)
}

func TestComments(t *testing.T) {
	mustValidate(t, `<?xml version="1.1"?><!-- a comment --><r><!-- inner --></r>`)
}

func TestProcessingInstruction(t *testing.T) {
	mustValidate(t, `<?xml version="1.1"?><?pi-target some data?><r/>`)
}

func TestPINoData(t *testing.T) {
	mustValidate(t, `<?xml version="1.1"?><?target?><r/>`)
}

func TestCDATASection(t *testing.T) {
	mustValidate(t, `<?xml version="1.1"?><r><![CDATA[<not>xml&]]></r>`)
}

func TestCharRefDecimal(t *testing.T) {
	mustValidate(t, `<?xml version="1.1"?><r>&#65;</r>`)
}

func TestCharRefHex(t *testing.T) {
	mustValidate(t, `<?xml version="1.1"?><r>&#x41;</r>`)
}

func TestPredefinedEntities(t *testing.T) {
	mustValidate(t, `<?xml version="1.1"?><r>&amp;&lt;&gt;&apos;&quot;</r>`)
}

func TestPredefinedEntitiesInAttr(t *testing.T) {
	mustValidate(t, `<?xml version="1.1"?><r a="&amp;&lt;&gt;&apos;&quot;"/>`)
}

func TestEmptyElement(t *testing.T) {
	mustValidate(t, `<?xml version="1.1"?><r></r>`)
}

func TestWhitespaceAroundRoot(t *testing.T) {
	mustValidate(t, "<?xml version=\"1.1\"?>\n  <r/>\n  ")
}

func TestNamespaceDeclaration(t *testing.T) {
	mustValidate(t, `<?xml version="1.1"?><r xmlns:ns="http://example.com"><ns:child/></r>`)
}

func TestDefaultNamespace(t *testing.T) {
	mustValidate(t, `<?xml version="1.1"?><r xmlns="http://example.com"><child/></r>`)
}

func TestXmlPrefixBuiltIn(t *testing.T) {
	mustValidate(t, `<?xml version="1.1"?><r xml:lang="en"/>`)
}

func TestNestedNamespaceScopes(t *testing.T) {
	mustValidate(t, `<?xml version="1.1"?><r xmlns:a="http://a"><a:x><a:y/></a:x></r>`)
}

func TestNamespaceUndeclarePrefix(t *testing.T) {
	mustValidate(t, `<?xml version="1.1"?><r xmlns:a="http://a"><a:x/><b xmlns:a=""/></r>`)
}

func TestUnicodeElementNames(t *testing.T) {
	mustValidate(t, `<?xml version="1.1"?><Ωroot/>`)
}

func TestMixedContent(t *testing.T) {
	mustValidate(t, `<?xml version="1.1"?><r>text<b>bold</b>more</r>`)
}

func TestCommentAfterRoot(t *testing.T) {
	mustValidate(t, `<?xml version="1.1"?><r/><!-- after -->`)
}

func TestPIAfterRoot(t *testing.T) {
	mustValidate(t, `<?xml version="1.1"?><r/><?target data?>`)
}

func TestCharRefForRestrictedChar(t *testing.T) {
	mustValidate(t, `<?xml version="1.1"?><r>&#x1;</r>`)
}

func TestCharRefForRestrictedCharInAttr(t *testing.T) {
	mustValidate(t, `<?xml version="1.1"?><r a="&#x1;"/>`)
}

// --- Rejection tests ---

func TestRejectXML10(t *testing.T) {
	mustReject(t, `<?xml version="1.0"?><r/>`, "only XML 1.1 is supported")
}

func TestRejectMissingDecl(t *testing.T) {
	mustReject(t, `<r/>`, "must begin with an XML declaration")
}

func TestRejectDOCTYPE(t *testing.T) {
	mustReject(t, `<?xml version="1.1"?><!DOCTYPE r><r/>`, "DOCTYPE")
}

func TestRejectGeneralEntityRef(t *testing.T) {
	mustReject(t, `<?xml version="1.1"?><r>&custom;</r>`, "unsupported: general entity reference")
}

func TestRejectMismatchedTags(t *testing.T) {
	mustReject(t, `<?xml version="1.1"?><a></b>`, "mismatched end tag")
}

func TestRejectDuplicateAttributes(t *testing.T) {
	mustReject(t, `<?xml version="1.1"?><r a="1" a="2"/>`, "duplicate attribute")
}

func TestRejectCommentDoubleDash(t *testing.T) {
	mustReject(t, `<?xml version="1.1"?><!-- -- --><r/>`, "'--' is not allowed")
}

func TestRejectLiteralLtInAttr(t *testing.T) {
	mustReject(t, `<?xml version="1.1"?><r a="<"/>`, "'<' is not allowed in attribute values")
}

func TestRejectCDATAEndInCharData(t *testing.T) {
	mustReject(t, `<?xml version="1.1"?><r>]]></r>`, "']]>' is not allowed")
}

func TestRejectUndeclaredPrefix(t *testing.T) {
	mustReject(t, `<?xml version="1.1"?><ns:r/>`, "undeclared namespace prefix")
}

func TestRejectMultipleColonsInName(t *testing.T) {
	mustReject(t, `<?xml version="1.1"?><a:b:c/>`, "multiple colons")
}

func TestRejectXmlnsAsPrefix(t *testing.T) {
	mustReject(t, `<?xml version="1.1"?><r xmlns:xmlns="http://bad"/>`, "must not be declared")
}

func TestRejectXmlPrefixWrongURI(t *testing.T) {
	mustReject(t, `<?xml version="1.1"?><r xmlns:xml="http://wrong"/>`, "must not be bound")
}

func TestRejectPITargetXml(t *testing.T) {
	mustReject(t, `<?xml version="1.1"?><?XML data?><r/>`, "must not be 'xml'")
}

func TestRejectEmptyDocument(t *testing.T) {
	mustReject(t, ``, "empty input")
}

func TestRejectNoRootElement(t *testing.T) {
	mustReject(t, `<?xml version="1.1"?>`, "expected root element")
}

func TestRejectContentAfterRoot(t *testing.T) {
	mustReject(t, `<?xml version="1.1"?><r/>extra`, "unexpected content after root element")
}

func TestRejectInvalidVersion(t *testing.T) {
	mustReject(t, `<?xml version="2.0"?><r/>`, "only XML 1.1 is supported")
}

func TestRejectBadEncodingName(t *testing.T) {
	mustReject(t, `<?xml version="1.1" encoding="123bad"?><r/>`, "encoding name must start with a letter")
}

func TestRejectEncodingMismatch(t *testing.T) {
	mustReject(t, `<?xml version="1.1" encoding="ISO-8859-1"?><r/>`, "only UTF-8 is supported")
}

func TestRejectStandaloneBad(t *testing.T) {
	mustReject(t, `<?xml version="1.1" standalone="maybe"?><r/>`, "standalone must be")
}

func TestRejectUnterminatedComment(t *testing.T) {
	mustReject(t, `<?xml version="1.1"?><!-- unterminated`, "unterminated comment")
}

func TestRejectUnterminatedCDATA(t *testing.T) {
	mustReject(t, `<?xml version="1.1"?><r><![CDATA[unterminated`, "unterminated CDATA")
}

func TestRejectUnterminatedElement(t *testing.T) {
	mustReject(t, `<?xml version="1.1"?><r>`, "unexpected end of input")
}

func TestRejectUnterminatedAttrValue(t *testing.T) {
	mustReject(t, `<?xml version="1.1"?><r a="unterminated`, "unterminated attribute value")
}

func TestRejectInvalidCharRef(t *testing.T) {
	mustReject(t, `<?xml version="1.1"?><r>&#0;</r>`, "invalid XML 1.1 character")
}

func TestRejectInvalidHexCharRef(t *testing.T) {
	mustReject(t, `<?xml version="1.1"?><r>&#xFFFE;</r>`, "invalid XML 1.1 character")
}

func TestRejectRestrictedCharLiteral(t *testing.T) {
	mustReject(t, "<?xml version=\"1.1\"?><r>\x01</r>", "restricted character")
}

func TestRejectRestrictedCharInAttr(t *testing.T) {
	mustReject(t, "<?xml version=\"1.1\"?><r a=\"\x01\"/>", "restricted character")
}

func TestRejectMarkupDeclarationInContent(t *testing.T) {
	mustReject(t, `<?xml version="1.1"?><r><![NOTATION[bad]]></r>`, "unsupported")
}

func TestRejectEmptyEntityRef(t *testing.T) {
	mustReject(t, `<?xml version="1.1"?><r>&;</r>`, "expected entity name")
}

// --- Line ending normalization tests ---

func TestLineEndingNormalization(t *testing.T) {
	tests := []struct {
		name	string
		input	[]rune
		want	[]rune
	}{
		{"CR LF", []rune{'\r', '\n'}, []rune{'\n'}},
		{"CR NEL", []rune{'\r', 0x85}, []rune{'\n'}},
		{"NEL alone", []rune{0x85}, []rune{'\n'}},
		{"LS", []rune{0x2028}, []rune{'\n'}},
		{"CR alone", []rune{'\r'}, []rune{'\n'}},
		{"LF unchanged", []rune{'\n'}, []rune{'\n'}},
		{"mixed", []rune{'a', '\r', '\n', 'b', 0x85, 'c', 0x2028, 'd', '\r', 'e'},
			[]rune{'a', '\n', 'b', '\n', 'c', '\n', 'd', '\n', 'e'}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeLineEndings(tt.input)
			require.Equal(t, len(tt.want), len(got))

			for i := range got {
				assert.Equal(t, tt.want[i], got[i])

			}
		})
	}
}

// --- Character class tests ---

func TestCharClassifications(t *testing.T) {
	assert.True(t, IsChar(0x1))

	assert.False(t, IsChar(0x0))

	assert.False(t, IsChar(0xFFFE))

	assert.False(t, IsChar(0xFFFF))

	assert.True(t, IsChar(0x10000))

	assert.True(t, IsRestrictedChar(0x1))

	assert.False(t, IsRestrictedChar(0x9))

	assert.False(t, IsRestrictedChar(0xA))

	assert.False(t, IsRestrictedChar(0xD))

	assert.True(t, IsRestrictedChar(0x7F))

	assert.False(t, IsRestrictedChar(0x85))

	assert.True(t, IsRestrictedChar(0x86))

}

func TestNameCharClassifications(t *testing.T) {
	assert.True(t, IsNameStartChar(':'))

	assert.True(t, IsNameStartChar('_'))

	assert.True(t, IsNameStartChar('A'))

	assert.False(t, IsNameStartChar('0'))

	assert.False(t, IsNameStartChar('-'))

	assert.True(t, IsNameChar('-'))

	assert.True(t, IsNameChar('.'))

	assert.True(t, IsNameChar('0'))

	assert.True(t, IsNCNameStartChar('A'))

	assert.False(t, IsNCNameStartChar(':'))

}

// --- Encoding rejection tests ---

func TestUTF16BEBOMRejected(t *testing.T) {
	utf16be := []byte{0xFE, 0xFF} // BOM
	xmlStr := `<?xml version="1.1"?><r/>`
	for _, r := range xmlStr {
		utf16be = append(utf16be, byte(r>>8), byte(r))
	}
	err := Validate(strings.NewReader(string(utf16be)))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "UTF-16")
}

func TestUTF16LEBOMRejected(t *testing.T) {
	utf16le := []byte{0xFF, 0xFE} // BOM
	xmlStr := `<?xml version="1.1"?><r/>`
	for _, r := range xmlStr {
		utf16le = append(utf16le, byte(r), byte(r>>8))
	}
	err := Validate(strings.NewReader(string(utf16le)))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "UTF-16")
}

func TestUTF16BENoBOMRejected(t *testing.T) {
	// UTF-16 BE with no BOM but the leading-NUL heuristic should still be
	// detected and rejected.
	xmlStr := `<?xml version="1.1"?><r/>`
	var buf []byte
	for _, r := range xmlStr {
		buf = append(buf, byte(r>>8), byte(r))
	}
	err := Validate(strings.NewReader(string(buf)))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "UTF-16")
}

func TestUTF16LENoBOMRejected(t *testing.T) {
	xmlStr := `<?xml version="1.1"?><r/>`
	var buf []byte
	for _, r := range xmlStr {
		buf = append(buf, byte(r), byte(r>>8))
	}
	err := Validate(strings.NewReader(string(buf)))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "UTF-16")
}

func TestEncodingDeclarationUTF16Rejected(t *testing.T) {
	err := Validate(strings.NewReader(`<?xml version="1.1" encoding="UTF-16"?><r/>`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "UTF-8")
}

func TestUTF8BOMRejected(t *testing.T) {
	input := "\xEF\xBB\xBF" + `<?xml version="1.1"?><r/>`
	err := Validate(strings.NewReader(input))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "BOM")
}

func TestComplexDocument(t *testing.T) {
	doc := `<?xml version="1.1" encoding="UTF-8" standalone="yes"?>
<!-- This is a complex test document -->
<?style-sheet href="style.css"?>
<root xmlns:ns="http://example.com/ns" xmlns="http://example.com/default">
  <ns:child attr1="value1" attr2="value2">
    Text content with entities: &amp; &lt; &gt; &apos; &quot;
    <![CDATA[Raw <data> & more]]>
    <!-- nested comment -->
    <?pi some processing instruction data?>
    <inner>
      <ns:deep xml:lang="en">
        Character references: &#65;&#x42;&#99;
      </ns:deep>
    </inner>
  </ns:child>
  <empty-element/>
  <self-closing att="val"/>
</root>
<!-- trailing comment -->
<?final-pi?>`
	mustValidate(t, doc)
}

func TestDeeplyNestedElements(t *testing.T) {
	var b strings.Builder
	b.WriteString(`<?xml version="1.1"?>`)
	depth := 100
	for i := 0; i < depth; i++ {
		fmt.Fprintf(&b, "<e%d>", i)
	}
	for i := depth - 1; i >= 0; i-- {
		fmt.Fprintf(&b, "</e%d>", i)
	}
	mustValidate(t, b.String())
}

func TestManyAttributes(t *testing.T) {
	var b strings.Builder
	b.WriteString(`<?xml version="1.1"?><r`)
	for i := 0; i < 50; i++ {
		fmt.Fprintf(&b, ` a%d="%d"`, i, i)
	}
	b.WriteString(`/>`)
	mustValidate(t, b.String())
}

func TestNamespaceAttributeUniqueness(t *testing.T) {
	mustReject(t,
		`<?xml version="1.1"?><r xmlns:a="http://x" xmlns:b="http://x" a:x="1" b:x="2"/>`,
		"conflicts with")
}

func TestRejectSecondRootElement(t *testing.T) {
	mustReject(t, `<?xml version="1.1"?><a/><b/>`, "unexpected content after root element")
}

func TestRejectXmlDeclNotAtStart(t *testing.T) {
	mustReject(t, " <?xml version=\"1.1\"?><r/>", "must begin with an XML declaration")
}
