package writer_test

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wow-look-at-my/xml-validator/reader"
	"github.com/wow-look-at-my/xml-validator/writer"
)

// What comes out of the writer goes back through the reader, because a writer
// whose output does not parse is not a writer.
// see docs/encodings.md

func roundtrip(t *testing.T, doc *reader.Document, opts writer.Options) (*reader.Document, string) {
	t.Helper()
	var out bytes.Buffer
	require.NoError(t, writer.WriteDocument(&out, doc, opts))

	back, err := reader.ParseTree(bytes.NewReader(out.Bytes()))
	require.NoError(t, err, "emitted document does not parse: %q", out.String())
	return back, out.String()
}

func parse(t *testing.T, text string) *reader.Document {
	t.Helper()
	doc, err := reader.ParseTree(strings.NewReader(text))
	require.NoError(t, err)
	return doc
}

func TestWritesADocumentThatReadsBack(t *testing.T) {
	doc := parse(t, `<?xml version="1.1"?><r id="1"><a>one</a><b/>tail</r>`)

	back, out := roundtrip(t, doc, writer.Options{})

	assert.Equal(t, "r", back.Root.Name)
	assert.Equal(t, "tail", back.Root.TextContent())
	require.Len(t, back.Root.ChildElements(), 2)
	assert.Contains(t, out, `<b/>`)
}

// A character that cannot stand for itself comes out as a reference, and
// reads back as the character it stood for.
func TestEscapesWhatCannotStandForItself(t *testing.T) {
	doc := parse(t, `<?xml version="1.1"?><r a="&amp;&#0;&#13;">&lt;&amp;&#0;&gt;</r>`)

	back, out := roundtrip(t, doc, writer.Options{})

	assert.Equal(t, "<&\x00>", back.Root.TextContent())
	v, ok := back.Root.Attr("a")
	require.True(t, ok)
	assert.Equal(t, "&\x00\r", v)
	assert.Contains(t, out, "&#0;")
	assert.NotContains(t, out, "\x00")
}

func TestByteModeWritesOneBytePerCharacter(t *testing.T) {
	doc := parse(t, `<?xml version="1.1"?><r>café</r>`)

	back, out := roundtrip(t, doc, writer.Options{Encoding: writer.Bytes})

	assert.Equal(t, "café", back.Root.TextContent())
	assert.Contains(t, out, `encoding="ISO-8859-1"`)
	assert.Equal(t, len(`<?xml version="1.1" encoding="ISO-8859-1"?><r>café</r>`)-1, len(out),
		"e-acute is one byte here and two in the source")
}

// Byte mode has no byte above U+00FF, so those become references rather than
// a document that cannot be written.
func TestByteModeReferencesWhatItCannotSpell(t *testing.T) {
	doc := parse(t, `<?xml version="1.1"?><r>snowman &#9731;</r>`)

	back, out := roundtrip(t, doc, writer.Options{Encoding: writer.Bytes})

	assert.Equal(t, "snowman ☃", back.Root.TextContent())
	assert.Contains(t, out, "&#9731;")
}

func TestReferencesModeIsPrintableASCII(t *testing.T) {
	doc := parse(t, `<?xml version="1.1"?><r>café ☃</r>`)

	back, out := roundtrip(t, doc, writer.Options{Encoding: writer.References})

	assert.Equal(t, "café ☃", back.Root.TextContent())
	for i, c := range []byte(out) {
		require.Truef(t, c >= 0x20 && c < 0x7F, "byte %d is %#02x, not printable ASCII", i, c)
	}
}

// The default for a payload of bytes is base64, because for arbitrary bytes
// it is smaller and faster than every alternative -- see docs/encodings.md.
func TestBinaryDefaultsToBase64(t *testing.T) {
	payload := make([]byte, 256)
	for i := range payload {
		payload[i] = byte(i)
	}

	var out bytes.Buffer
	require.NoError(t, writer.WriteBinary(&out, "blob", payload, writer.Options{}))

	doc, err := reader.ParseTree(bytes.NewReader(out.Bytes()))
	require.NoError(t, err)
	decoded, err := base64.StdEncoding.DecodeString(doc.Root.TextContent())
	require.NoError(t, err)
	assert.Equal(t, payload, decoded)
	assert.Equal(t, 378, out.Len(), "1.33x the payload, plus the declaration and tags")
}

func TestBinaryAsHex(t *testing.T) {
	payload := []byte{0x00, 0xFF, 0x10, 0xAB}

	var out bytes.Buffer
	require.NoError(t, writer.WriteBinary(&out, "blob", payload, writer.Options{Binary: writer.Hex}))

	doc, err := reader.ParseTree(bytes.NewReader(out.Bytes()))
	require.NoError(t, err)
	assert.Equal(t, "00FF10AB", doc.Root.TextContent())
	decoded, err := hex.DecodeString(doc.Root.TextContent())
	require.NoError(t, err)
	assert.Equal(t, payload, decoded)
}

// Carrying bytes as text is the other answer, and it is bigger for arbitrary
// bytes: a quarter of them cannot appear literally in any XML document.
func TestBinaryAsTextCarriesTheSameBytes(t *testing.T) {
	payload := make([]byte, 256)
	for i := range payload {
		payload[i] = byte(i)
	}

	var out bytes.Buffer
	require.NoError(t, writer.WriteBinary(&out, "blob", payload, writer.Options{Binary: writer.Text}))

	doc, err := reader.ParseTree(bytes.NewReader(out.Bytes()))
	require.NoError(t, err)

	got := make([]byte, 0, 256)
	for _, r := range doc.Root.TextContent() {
		require.Less(t, r, rune(256))
		got = append(got, byte(r))
	}
	assert.Equal(t, payload, got)
	assert.Greater(t, out.Len(), 378, "text is the larger form for arbitrary bytes")
}

func TestRejectsWhatItCannotWrite(t *testing.T) {
	err := writer.WriteDocument(&bytes.Buffer{}, nil, writer.Options{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no root element")

	err = writer.WriteDocument(&bytes.Buffer{}, &reader.Document{}, writer.Options{})
	require.Error(t, err)

	err = writer.WriteBinary(&bytes.Buffer{}, "blob", []byte{1}, writer.Options{Binary: writer.BinaryEncoding(99)})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown binary encoding")
}
