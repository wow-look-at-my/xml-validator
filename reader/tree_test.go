package reader_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wow-look-at-my/xml-validator/reader"
)

// Reading is what this module does: bytes in, a tree out. Validation lives in
// the validator module, so what a document MEANS is not checked here -- only
// that it is read as written.

const decl = `<?xml version="1.1"?>`

func TestParsesElementsAttributesAndText(t *testing.T) {
	doc, err := reader.ParseTree(strings.NewReader(decl + `<r id="1" name="x"><a>one</a><b/><a>two</a></r>`))
	require.NoError(t, err)

	assert.Equal(t, "r", doc.Root.Name)
	id, ok := doc.Root.Attr("id")
	require.True(t, ok)
	assert.Equal(t, "1", id)

	kids := doc.Root.ChildElements()
	require.Len(t, kids, 3)
	assert.Equal(t, "one", kids[0].TextContent())
	assert.Empty(t, kids[1].Children)
	assert.Equal(t, "two", kids[2].TextContent())
}

// Positions are what a consumer reports a problem with, and an attribute has
// its own: on a multi-attribute element it is not where the element starts.
func TestNodesCarryTheirPosition(t *testing.T) {
	doc, err := reader.ParseTree(strings.NewReader(decl + "\n<r\n  first=\"1\"\n  second=\"2\"/>"))
	require.NoError(t, err)

	assert.Equal(t, 2, doc.Root.Line)
	require.Len(t, doc.Root.Attrs, 2)
	assert.Equal(t, 3, doc.Root.Attrs[0].Line)
	assert.Equal(t, 4, doc.Root.Attrs[1].Line)
	assert.NotEqual(t, doc.Root.Attrs[0].Col, doc.Root.Line)
}

func TestResolvesNamespaces(t *testing.T) {
	doc, err := reader.ParseTree(strings.NewReader(decl +
		`<r xmlns="urn:d" xmlns:p="urn:p"><p:child p:at="v"/></r>`))
	require.NoError(t, err)

	assert.Equal(t, "urn:d", doc.Root.Namespace)
	child := doc.Root.ChildElements()[0]
	assert.Equal(t, "urn:p", child.Namespace)
	assert.Equal(t, "child", child.Local)
	assert.Equal(t, "p", child.Prefix)
	require.Len(t, child.Attrs, 1)
	assert.Equal(t, "urn:p", child.Attrs[0].Namespace)

	// The declarations themselves are scope, not data.
	assert.Empty(t, doc.Root.Attrs)
	assert.Equal(t, "urn:p", doc.Root.Namespaces["p"])
}

func TestReadsReferencesAndCDATA(t *testing.T) {
	doc, err := reader.ParseTree(strings.NewReader(decl +
		`<r a="&amp;&lt;&gt;&apos;&quot;">&#65;&#x42;&#0;<![CDATA[<raw> & ]]>tail</r>`))
	require.NoError(t, err)

	v, ok := doc.Root.Attr("a")
	require.True(t, ok)
	assert.Equal(t, `&<>'"`, v)
	assert.Equal(t, "AB\x00<raw> & tail", doc.Root.TextContent())
}

// Comments and processing instructions are markup the tree does not carry.
func TestSkipsCommentsAndProcessingInstructions(t *testing.T) {
	doc, err := reader.ParseTree(strings.NewReader(decl +
		`<!--before--><?target data?><r><!--inside-->text<?pi?></r><!--after-->`))
	require.NoError(t, err)

	assert.Equal(t, "text", doc.Root.TextContent())
	assert.Empty(t, doc.Root.ChildElements())
}

func TestRejectsMalformedInput(t *testing.T) {
	cases := map[string]string{
		"no root":          decl,
		"unclosed element": decl + `<r>`,
		"mismatched tag":   decl + `<r></q>`,
		"unknown entity":   decl + `<r>&nope;</r>`,
		"empty input":      "",
	}
	for name, doc := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := reader.ParseTree(strings.NewReader(doc))
			require.Error(t, err)
		})
	}
}

// An error says where, because a consumer that cannot point at the problem
// makes its reader go looking.
func TestErrorCarriesAPosition(t *testing.T) {
	_, err := reader.ParseTree(strings.NewReader(decl + "\n<r>\n&nope;</r>"))
	require.Error(t, err)

	var e *reader.Error
	require.ErrorAs(t, err, &e)
	assert.Equal(t, 3, e.Line)
	assert.NotEmpty(t, e.Message)
	assert.Contains(t, e.Error(), "line 3")
}

func TestDecodeReadsBothModes(t *testing.T) {
	utf8Doc := []byte(decl + "<r>café</r>")
	runes, err := reader.Decode(bytes.NewReader(utf8Doc))
	require.NoError(t, err)
	assert.Contains(t, string(runes), "café")

	byteDoc := []byte(`<?xml version="1.1" encoding="ISO-8859-1"?><r>caf` + "\xe9" + `</r>`)
	runes, err = reader.Decode(bytes.NewReader(byteDoc))
	require.NoError(t, err)
	assert.Contains(t, string(runes), "café")
}

func TestDecodeRejectsWhatItCannotRead(t *testing.T) {
	cases := map[string]struct{ input, want string }{
		"UTF-8 BOM":    {"\xef\xbb\xbf" + decl + "<r/>", "BOM"},
		"UTF-16 BE":    {"\xfe\xff<\x00r\x00", "UTF-16"},
		"UTF-16 LE":    {"\xff\xfe<\x00r\x00", "UTF-16"},
		"NUL-led":      {"\x00<" + decl, "UTF-16"},
		"broken UTF-8": {decl + "<r>\xe9</r>", "invalid UTF-8"},
		"empty":        {"", "empty input"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := reader.Decode(strings.NewReader(tc.input))
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.want)
		})
	}
}

// Line endings are normalized on the way in, so a document written on any
// platform parses to the same characters.
func TestDecodeNormalizesLineEndings(t *testing.T) {
	runes, err := reader.Decode(strings.NewReader(decl + "<r>a\r\nb\rcd e</r>"))
	require.NoError(t, err)
	assert.Equal(t, decl+"<r>a\nb\nc\nd\ne</r>", string(runes))
}

func TestPredefinedEntity(t *testing.T) {
	for word, want := range map[string]rune{"amp": '&', "lt": '<', "gt": '>', "apos": '\'', "quot": '"'} {
		got, ok := reader.PredefinedEntity([]rune(word))
		require.True(t, ok, word)
		assert.Equal(t, want, got)
	}
	_, ok := reader.PredefinedEntity([]rune("nbsp"))
	assert.False(t, ok)
}

func TestCharacterClasses(t *testing.T) {
	assert.True(t, reader.IsChar('a'))
	assert.False(t, reader.IsChar(0))
	assert.True(t, reader.IsCharRefValue(0), "a reference may resolve to U+0000")
	assert.False(t, reader.IsCharRefValue(0xD800), "a lone surrogate is not a character")
	assert.True(t, reader.IsRestrictedChar(0x1B))
	assert.False(t, reader.IsRestrictedChar('a'))
	assert.True(t, reader.IsWhitespace(' '))
	assert.True(t, reader.IsNameStartChar(':'))
	assert.False(t, reader.IsNCNameStartChar(':'))
	assert.True(t, reader.IsNameChar('-'))
	assert.False(t, reader.IsNameChar(' '))
	assert.True(t, reader.IsNCNameChar('a'))
}
