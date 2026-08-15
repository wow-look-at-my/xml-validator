package validator

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A parsed `&#0;` is U+0000: a character with a value, like any other. These
// tests emit it back as `&#0;` and reparse, so the claim is a measured
// equality rather than an assertion about the parser alone.
// see docs/nul-char-ref.md

const xmlDecl = `<?xml version="1.1"?>`

// serializeDoc emits a parsed tree as an XML 1.1 document. It escapes U+0000
// and the restricted characters as decimal character references, which is the
// only spelling that keeps them out of the byte stream.
//
// It emits no namespace declarations, so every document here declares none.
func serializeDoc(d *Document) string {
	var b strings.Builder
	b.WriteString(xmlDecl)
	writeElement(&b, d.Root)
	return b.String()
}

func writeElement(b *strings.Builder, e *Element) {
	b.WriteString("<" + e.Name)
	for _, a := range e.Attrs {
		b.WriteString(" " + a.Name + `="` + escapeXML(a.Value, true) + `"`)
	}
	if len(e.Children) == 0 {
		b.WriteString("/>")
		return
	}
	b.WriteString(">")
	for _, c := range e.Children {
		switch n := c.(type) {
		case *Element:
			writeElement(b, n)
		case *CharData:
			b.WriteString(escapeXML(n.Content, false))
		}
	}
	b.WriteString("</" + e.Name + ">")
}

func escapeXML(s string, inAttr bool) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r == '&':
			b.WriteString("&amp;")
		case r == '<':
			b.WriteString("&lt;")
		case r == '>':
			b.WriteString("&gt;")
		case r == '"' && inAttr:
			b.WriteString("&quot;")
		case r == 0 || IsRestrictedChar(r):
			fmt.Fprintf(&b, "&#%d;", r)
		// CR, NEL and LINE SEPARATOR all normalize to LF when they are literal
		// characters, so a payload that carries one has to write a reference.
		// Normalization runs over the input bytes, before any reference
		// resolves, which is why the reference survives it.
		case r == '\r' || r == 0x85 || r == 0x2028:
			fmt.Fprintf(&b, "&#%d;", r)
		// Tab and newline are legal in an attribute value here, but a
		// conforming reader folds both to a space. A reference is the spelling
		// that means the character itself.
		case inAttr && (r == '\t' || r == '\n'):
			fmt.Fprintf(&b, "&#%d;", r)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// snap is a document tree stripped of the parts a roundtrip may legitimately
// change: source positions, and where the parser split one run of text into
// several CharData nodes.
type snap struct {
	Name  string
	Attrs []attrSnap
	Kids  []any
}

type attrSnap struct {
	Name  string
	Value string
}

func snapshot(e *Element) snap {
	s := snap{Name: e.Name}
	for _, a := range e.Attrs {
		s.Attrs = append(s.Attrs, attrSnap{Name: a.Name, Value: a.Value})
	}
	var text strings.Builder
	flush := func() {
		if text.Len() > 0 {
			s.Kids = append(s.Kids, text.String())
			text.Reset()
		}
	}
	for _, c := range e.Children {
		switch n := c.(type) {
		case *CharData:
			text.WriteString(n.Content)
		case *Element:
			flush()
			s.Kids = append(s.Kids, snapshot(n))
		}
	}
	flush()
	return s
}

// roundtrip parses src, emits the tree, and parses the emitted text. It
// returns the reparsed document and the emitted text.
func roundtrip(t *testing.T, src string) (*Document, string) {
	t.Helper()

	first, err := ParseTree(strings.NewReader(src))
	require.NoError(t, err)

	out := serializeDoc(first)
	second, err := ParseTree(strings.NewReader(out))
	require.NoError(t, err, "emitted document does not parse: %q", out)

	assert.Equal(t, snapshot(first.Root), snapshot(second.Root), "tree changed across the roundtrip")
	assert.Equal(t, out, serializeDoc(second), "a second roundtrip changed the bytes")
	return second, out
}

func TestRoundtripNulInText(t *testing.T) {
	doc, out := roundtrip(t, xmlDecl+`<r>a&#0;b</r>`)

	assert.Equal(t, "a\x00b", doc.Root.TextContent())
	assert.Contains(t, out, "&#0;")
	assert.NotContains(t, out, "\x00", "the emitted document carries a NUL byte")
}

func TestRoundtripNulInAttributeValue(t *testing.T) {
	doc, out := roundtrip(t, xmlDecl+`<r a="x&#0;y" b="plain"/>`)

	v, ok := doc.Root.Attr("a")
	require.True(t, ok)
	assert.Equal(t, "x\x00y", v)
	assert.Contains(t, out, `a="x&#0;y"`)

	other, ok := doc.Root.Attr("b")
	require.True(t, ok)
	assert.Equal(t, "plain", other, "the NUL in the first attribute did not end the second")
}

// The four spellings differ as text and mean the same character, so they all
// come back out as the one canonical reference.
func TestRoundtripNulSpellingsConverge(t *testing.T) {
	for _, ref := range []string{"&#0;", "&#00;", "&#x0;", "&#x00;", "&#0000000;"} {
		t.Run(ref, func(t *testing.T) {
			doc, out := roundtrip(t, xmlDecl+`<r>`+ref+`</r>`)

			assert.Equal(t, "\x00", doc.Root.TextContent())
			assert.Equal(t, xmlDecl+`<r>&#0;</r>`, out)
		})
	}
}

// A C string ends at its first NUL. A document does not: everything after the
// reference is still content, and it is still there after a roundtrip.
func TestNulDoesNotTerminateTheDocument(t *testing.T) {
	doc, out := roundtrip(t, xmlDecl+`<r><a>1&#0;2</a><b at="&#0;">after</b><c/></r>`)

	kids := doc.Root.ChildElements()
	require.Len(t, kids, 3)
	assert.Equal(t, "1\x002", kids[0].TextContent())
	assert.Equal(t, "after", kids[1].TextContent())
	assert.Equal(t, "c", kids[2].Name)
	assert.Contains(t, out, "<c/>", "the tail of the document survived the NUL")
}

// The reference is four ASCII bytes. Nothing that reads the document has to
// survive a NUL byte, because the document has none.
func TestNulCharRefIsFourAsciiBytes(t *testing.T) {
	_, out := roundtrip(t, xmlDecl+`<r>`+strings.Repeat("&#0;", 100)+`</r>`)

	assert.Equal(t, 4, len("&#0;"))
	assert.Equal(t, 400, strings.Count(out, "&#0;")*4)
	assert.NotContains(t, out, "\x00", "the emitted document carries a NUL byte")
}

// The parsed value does carry the NUL, and a Go string holds it like any other
// byte: length, index and split all count it.
func TestParsedNulIsAnOrdinaryStringByte(t *testing.T) {
	doc, err := ParseTree(strings.NewReader(xmlDecl + `<r>ab&#0;cd</r>`))
	require.NoError(t, err)

	text := doc.Root.TextContent()
	assert.Len(t, text, 5)
	assert.Equal(t, byte(0), text[2])
	assert.Equal(t, []string{"ab", "cd"}, strings.Split(text, "\x00"))
	assert.Equal(t, 5, len([]rune(text)))
}

// Inside CDATA the same five characters are text, not a reference: no NUL is
// produced, and the emitted document escapes the ampersand to keep it that way.
func TestNulCharRefInCDATAStaysLiteral(t *testing.T) {
	doc, out := roundtrip(t, xmlDecl+`<r><![CDATA[a&#0;b]]></r>`)

	assert.Equal(t, "a&#0;b", doc.Root.TextContent())
	assert.Contains(t, out, "a&amp;#0;b")
}

// A literal NUL byte is rejected wherever it appears. The reference is the
// only way to put U+0000 in a document, which is what makes the byte stream
// NUL-free by construction.
func TestRejectLiteralNulInEveryContext(t *testing.T) {
	cases := map[string]string{
		"text":    xmlDecl + "<r>a\x00b</r>",
		"attr":    xmlDecl + "<r a=\"x\x00y\"/>",
		"cdata":   xmlDecl + "<r><![CDATA[a\x00b]]></r>",
		"comment": xmlDecl + "<!--a\x00b--><r/>",
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) {
			err := Validate(strings.NewReader(src))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "U+0000")
		})
	}
}

// allBytes is a binary payload holding one of every byte value, 0 through 255,
// in order. Byte b is carried as the character U+00XX with the same value, the
// Latin-1 mapping, so the payload survives as 256 distinct characters.
func allBytes() []byte {
	p := make([]byte, 256)
	for i := range p {
		p[i] = byte(i)
	}
	return p
}

func encodeBytes(p []byte) string {
	var b strings.Builder
	for _, c := range p {
		b.WriteRune(rune(c))
	}
	return b.String()
}

// decodeBytes is the inverse: every character must be inside Latin-1, and it
// becomes the byte it stands for.
func decodeBytes(t *testing.T, s string) []byte {
	t.Helper()
	out := make([]byte, 0, len(s))
	for _, r := range s {
		require.Less(t, r, rune(256), "character U+%04X is outside the payload's range", r)
		out = append(out, byte(r))
	}
	return out
}

// The whole point, on real binary data: 256 bytes go in, an XML document comes
// out, and the same 256 bytes come back. U+0000 is one of them, and it is no
// more special than U+0041.
func TestRoundtripEveryByteValue(t *testing.T) {
	payload := allBytes()
	text := encodeBytes(payload)
	src := xmlDecl + `<r a="` + escapeXML(text, true) + `">` + escapeXML(text, false) + `</r>`

	doc, out := roundtrip(t, src)

	assert.Equal(t, payload, decodeBytes(t, doc.Root.TextContent()), "text content lost bytes")
	v, ok := doc.Root.Attr("a")
	require.True(t, ok)
	assert.Equal(t, payload, decodeBytes(t, v), "attribute value lost bytes")

	assert.NotContains(t, out, "\x00", "the emitted document carries a NUL byte")
	assert.Contains(t, out, "&#0;", "U+0000 is carried as a reference")
	assert.Contains(t, out, "&#127;", "U+007F is carried as a reference")
}

// The same payload written as nothing but references. Every byte becomes
// `&#N;`, so the document is printable ASCII end to end -- a wire form that
// survives a transport with opinions about high bytes and NUL.
func TestRoundtripEveryByteValueAsReferences(t *testing.T) {
	payload := allBytes()

	var refs strings.Builder
	for _, c := range payload {
		fmt.Fprintf(&refs, "&#%d;", c)
	}
	src := xmlDecl + `<r>` + refs.String() + `</r>`

	for i, c := range []byte(src) {
		require.Truef(t, c >= 0x20 && c < 0x7F, "byte %d of the document is %#02x, not printable ASCII", i, c)
	}

	doc, out := roundtrip(t, src)
	assert.Equal(t, payload, decodeBytes(t, doc.Root.TextContent()))
	assert.NotContains(t, out, "\x00")
}

// Byte 0 is not the end of the payload, and neither is any run of them: the
// bytes after the NULs are still there, in order.
func TestRoundtripBinaryPayloadWithEmbeddedNuls(t *testing.T) {
	payload := append(allBytes(), 0, 0, 0, 'e', 'n', 'd')
	src := xmlDecl + `<r>` + escapeXML(encodeBytes(payload), false) + `</r>`

	doc, _ := roundtrip(t, src)

	got := decodeBytes(t, doc.Root.TextContent())
	assert.Equal(t, payload, got)
	assert.Len(t, got, 262)
	assert.Equal(t, []byte("end"), got[len(got)-3:], "the tail after three NUL bytes survived")
}

const lengthSchema = `<?xml version="1.1"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
	<xs:element name="r">
		<xs:simpleType>
			<xs:restriction base="xs:string">
				<xs:length value="%d"/>
			</xs:restriction>
		</xs:simpleType>
	</xs:element>
</xs:schema>`

// Schema validation counts the NUL as one character, the same as the letters
// around it. A terminator would leave a value of length 1.
func TestSchemaCountsNulAsOneCharacter(t *testing.T) {
	doc := xmlDecl + `<r>a&#0;b</r>`

	err := ValidateWithSchemaBytes([]byte(doc), []byte(fmt.Sprintf(lengthSchema, 3)))
	assert.NoError(t, err)

	err = ValidateWithSchemaBytes([]byte(doc), []byte(fmt.Sprintf(lengthSchema, 1)))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "length 3")
}
