package reader

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// The encoding declaration decides how the bytes are read, so it is read from
// the bytes. Every spelling a declaration may use is ASCII in both modes.
// see docs/encodings.md

const xmlDecl = `<?xml version="1.1"?>`

func TestEncodingNameAliases(t *testing.T) {
	for _, name := range []string{"ISO-8859-1", "iso-8859-1", "latin1", "LATIN-1", "l1", "ISO_8859-1", "csISOLatin1", "IBM819", "cp819"} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, encodingByte, CanonicalEncoding(name))
		})
	}
	for _, name := range []string{"UTF-8", "utf-8", "utf8"} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, encodingUTF8, CanonicalEncoding(name))
		})
	}
	for _, name := range []string{"UTF-16", "Shift_JIS", "windows-1252", "ASCII", ""} {
		t.Run("rejects "+name, func(t *testing.T) {
			assert.Empty(t, CanonicalEncoding(name))
		})
	}
}

func TestSniffEncoding(t *testing.T) {
	cases := map[string]string{
		`<?xml version="1.1"?><r/>`:                           encodingUTF8,
		`<?xml version="1.1" encoding="UTF-8"?><r/>`:          encodingUTF8,
		`<?xml version="1.1" encoding="ISO-8859-1"?><r/>`:     encodingByte,
		`<?xml version="1.1" encoding='latin1'?><r/>`:         encodingByte,
		`<?xml version="1.1"  encoding = "ISO-8859-1" ?><r/>`: encodingByte,
		`<?xml version="1.1"?><r>encoding="ISO-8859-1"</r>`:   encodingUTF8,
		`<?xml version="1.1" encoding="windows-1252"?><r/>`:   encodingUTF8,
		`<?xml version="1.1" encoding=?><r/>`:                 encodingUTF8,
		`<?xml version="1.1" standalone="yes"?><r/>`:          encodingUTF8,
		`<r>encoding="ISO-8859-1"</r>`:                        encodingUTF8,
	}
	for doc, want := range cases {
		t.Run(doc, func(t *testing.T) {
			assert.Equal(t, want, sniffEncoding([]byte(doc)))
		})
	}
}

// An encoding declaration past the sniff window is not a declaration at all:
// it cannot appear there, because the declaration is the first thing in the
// document.
func TestSniffEncodingIgnoresLaterText(t *testing.T) {
	doc := xmlDecl + `<r>` + strings.Repeat("x", 400) + `encoding="ISO-8859-1"</r>`
	assert.Equal(t, encodingUTF8, sniffEncoding([]byte(doc)))
}

// Line-ending normalization rewrites its input in place, so what comes back
// has to be right even though the slice it read from is the one it wrote to.
func TestNormalizeLineEndings(t *testing.T) {
	tests := []struct {
		name        string
		input, want []rune
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
			assert.Equal(t, tt.want, normalizeLineEndings(tt.input))
		})
	}
}

// The Node interface's methods exist to keep the tree closed to types outside
// this package. Calling them is all there is to check.
func TestNodeInterfaceStubs(t *testing.T) {
	var n Node = &Element{}
	n.node()
	n = &CharData{}
	n.node()
}
