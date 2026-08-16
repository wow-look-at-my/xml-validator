package validator

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Byte mode: one byte is one character, U+0000 through U+00FF. A character
// above that has no byte, so it takes a character reference.
// see docs/encodings.md

const byteDecl = `<?xml version="1.1" encoding="ISO-8859-1"?>`

// latin1 writes s -- whose characters must all be inside Latin-1 -- as the
// bytes of a byte-mode document.
func latin1(t *testing.T, s string) []byte {
	t.Helper()
	out := make([]byte, 0, len(s))
	for _, r := range s {
		require.Less(t, r, rune(256), "character U+%04X has no byte in this mode", r)
		out = append(out, byte(r))
	}
	return out
}

func TestByteModeReadsOneBytePerCharacter(t *testing.T) {
	doc := latin1(t, byteDecl+`<r>é ÿ ½</r>`)

	require.NoError(t, Validate(bytes.NewReader(doc)))

	tree, err := ParseTree(bytes.NewReader(doc))
	require.NoError(t, err)
	assert.Equal(t, "é ÿ ½", tree.Root.TextContent())
	assert.Len(t, doc, len(byteDecl)+3+len(`<r></r>`)+2, "one byte per character")
}

// The same document in the two modes: identical characters, different bytes.
func TestByteModeIsShorterThanUTF8ForTheHighHalf(t *testing.T) {
	text := "àáâãäåæçèéêëìíîï"
	asBytes := latin1(t, byteDecl+`<r>`+text+`</r>`)
	asUTF8 := []byte(xmlDecl + `<r>` + text + `</r>`)

	require.NoError(t, Validate(bytes.NewReader(asBytes)))
	require.NoError(t, Validate(bytes.NewReader(asUTF8)))

	byteTree, err := ParseTree(bytes.NewReader(asBytes))
	require.NoError(t, err)
	utf8Tree, err := ParseTree(bytes.NewReader(asUTF8))
	require.NoError(t, err)
	assert.Equal(t, utf8Tree.Root.TextContent(), byteTree.Root.TextContent())

	assert.Equal(t, 16, len(text)-16, "the payload is 16 characters")
	assert.Equal(t, 16, len(asBytes)-len(byteDecl)-len(`<r></r>`))
	assert.Equal(t, 32, len(asUTF8)-len(xmlDecl)-len(`<r></r>`), "UTF-8 spends two bytes each")
}

// Above U+00FF there is no byte to write, so a reference is the only spelling.
func TestByteModeCarriesHigherCharactersAsReferences(t *testing.T) {
	doc := latin1(t, byteDecl+`<r>snowman &#9731; and a clef &#119070;</r>`)

	tree, err := ParseTree(bytes.NewReader(doc))
	require.NoError(t, err)
	assert.Equal(t, "snowman ☃ and a clef 𝄞", tree.Root.TextContent())
}

// A raw byte above 0x7F is a character here, not a broken UTF-8 sequence.
// The same bytes without the declaration are rejected.
func TestByteModeAcceptsWhatUTF8Rejects(t *testing.T) {
	raw := []byte{0xE9}

	valid := append(latin1(t, byteDecl+`<r>`), raw...)
	valid = append(valid, []byte(`</r>`)...)
	require.NoError(t, Validate(bytes.NewReader(valid)))

	invalid := append([]byte(xmlDecl+`<r>`), raw...)
	invalid = append(invalid, []byte(`</r>`)...)
	err := Validate(bytes.NewReader(invalid))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid UTF-8 byte sequence")
}

// Byte mode changes how bytes become characters. It changes nothing about
// which characters a document may hold.
func TestByteModeStillRejectsInvalidCharacters(t *testing.T) {
	cases := map[string]struct{ doc, want string }{
		"literal NUL":        {byteDecl + "<r>a\x00b</r>", "U+0000"},
		"restricted control": {byteDecl + "<r>a\x1bb</r>", "restricted character U+001B"},
		"C1 control":         {byteDecl + "<r>a\u008bb</r>", "restricted character U+008B"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := Validate(bytes.NewReader(latin1(t, tc.doc)))
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.want)
		})
	}

	// ...and the reference still carries U+0000 through, in this mode too.
	tree, err := ParseTree(bytes.NewReader(latin1(t, byteDecl+`<r>a&#0;b</r>`)))
	require.NoError(t, err)
	assert.Equal(t, "a\x00b", tree.Root.TextContent())
}

func TestRejectUnknownEncodingNames(t *testing.T) {
	for _, name := range []string{"windows-1252", "ISO-8859-15", "UTF-16"} {
		t.Run(name, func(t *testing.T) {
			err := Validate(strings.NewReader(fmt.Sprintf(`<?xml version="1.1" encoding=%q?><r/>`, name)))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "unsupported encoding")
			assert.Contains(t, err.Error(), "UTF-8 and ISO-8859-1 are supported")
		})
	}
}
