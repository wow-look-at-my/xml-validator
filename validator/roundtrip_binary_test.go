package validator

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// XSD carries arbitrary bytes in xs:base64Binary and xs:hexBinary. Both are
// ASCII on the wire, and both measure their length in octets.
// see docs/nul-char-ref.md

func binaryTypeXSD(base string, length int) string {
	return fmt.Sprintf(`<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="blob">
    <xs:simpleType>
      <xs:restriction base="xs:%s">
        <xs:length value="%d"/>
      </xs:restriction>
    </xs:simpleType>
  </xs:element>
</xs:schema>`, base, length)
}

func TestRoundtripBase64BinaryPayload(t *testing.T) {
	payload := allBytes()
	doc := xmlDecl + `<blob>` + base64.StdEncoding.EncodeToString(payload) + `</blob>`

	require.NoError(t, ValidateWithSchemaBytes([]byte(doc), []byte(binaryTypeXSD("base64Binary", 256))))

	tree, err := ParseTree(strings.NewReader(doc))
	require.NoError(t, err)
	decoded, err := base64.StdEncoding.DecodeString(tree.Root.TextContent())
	require.NoError(t, err)
	assert.Equal(t, payload, decoded, "the payload did not survive base64Binary")

	// 256 octets, not the 344 characters that carry them.
	err = ValidateWithSchemaBytes([]byte(doc), []byte(binaryTypeXSD("base64Binary", 344)))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "value length 256")
}

func TestRoundtripHexBinaryPayload(t *testing.T) {
	payload := allBytes()
	doc := xmlDecl + `<blob>` + strings.ToUpper(hex.EncodeToString(payload)) + `</blob>`

	require.NoError(t, ValidateWithSchemaBytes([]byte(doc), []byte(binaryTypeXSD("hexBinary", 256))))

	tree, err := ParseTree(strings.NewReader(doc))
	require.NoError(t, err)
	decoded, err := hex.DecodeString(tree.Root.TextContent())
	require.NoError(t, err)
	assert.Equal(t, payload, decoded, "the payload did not survive hexBinary")

	// 256 octets, not the 512 hex digits that spell them.
	err = ValidateWithSchemaBytes([]byte(doc), []byte(binaryTypeXSD("hexBinary", 512)))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "value length 256")
}

// The three wire forms for the same 256 bytes, side by side. Each is ASCII,
// and none of them holds a NUL byte.
func TestBinaryWireFormSizes(t *testing.T) {
	payload := allBytes()

	var refs strings.Builder
	for _, c := range payload {
		fmt.Fprintf(&refs, "&#%d;", c)
	}
	forms := map[string]string{
		"references": xmlDecl + `<r>` + refs.String() + `</r>`,
		"base64":     xmlDecl + `<blob>` + base64.StdEncoding.EncodeToString(payload) + `</blob>`,
		"hex":        xmlDecl + `<blob>` + strings.ToUpper(hex.EncodeToString(payload)) + `</blob>`,
	}
	// 256 payload bytes become 1430, 344 and 512 characters of content; the
	// declaration and the tags add the rest.
	sizes := map[string]int{"references": 1454, "base64": 378, "hex": 546}

	for name, doc := range forms {
		t.Run(name, func(t *testing.T) {
			require.NoError(t, Validate(strings.NewReader(doc)))
			assert.Equal(t, sizes[name], len(doc))
			assert.NotContains(t, doc, "\x00")
			for i, c := range []byte(doc) {
				require.Truef(t, c >= 0x20 && c < 0x7F, "byte %d is %#02x, not printable ASCII", i, c)
			}
		})
	}
}
