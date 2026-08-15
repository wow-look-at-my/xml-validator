package validator

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

// What each wire form costs to READ, against what it costs to store. Sizes
// live in docs/encodings.md; this is the other half of the decision.
//
// Every benchmark carries the same payload and reports ns/op against
// b.SetBytes(payload), so B/s is per payload byte and the forms compare
// directly.
// see docs/encodings.md

const benchPayloadSize = 64 * 1024

// binaryPayload is arbitrary bytes: every value appears, so about a quarter
// of it cannot be written literally in any mode.
func binaryPayload() []byte {
	p := make([]byte, benchPayloadSize)
	for i := range p {
		p[i] = byte(i * 7)
	}
	return p
}

// textPayload is text with the occasional NUL, which is the case that decides
// between escaping and base64.
func textPayload(nulEvery int) []byte {
	var b bytes.Buffer
	line := "the quick brown fox jumps over the lazy dog, and the dog does not mind at all. "
	for b.Len() < benchPayloadSize {
		b.WriteString(line)
	}
	p := b.Bytes()[:benchPayloadSize]
	if nulEvery > 0 {
		for i := nulEvery; i < len(p); i += nulEvery {
			p[i] = 0
		}
	}
	return p
}

func docAllReferences(payload []byte) []byte {
	var b strings.Builder
	b.WriteString(xmlDecl + "<r>")
	for _, c := range payload {
		fmt.Fprintf(&b, "&#%d;", c)
	}
	b.WriteString("</r>")
	return []byte(b.String())
}

// Byte mode is written as bytes, one per character. Building it as a Go
// string would encode those characters as UTF-8, which is the other mode.
func docByteMode(payload []byte) []byte {
	text := byteDecl + "<r>" + escapeXML(encodeBytes(payload), false) + "</r>"
	out := make([]byte, 0, len(text))
	for _, r := range text {
		out = append(out, byte(r))
	}
	return out
}

func docUTF8(payload []byte) []byte {
	return []byte(xmlDecl + "<r>" + escapeXML(encodeBytes(payload), false) + "</r>")
}

func docBase64(payload []byte) []byte {
	return []byte(xmlDecl + "<blob>" + base64.StdEncoding.EncodeToString(payload) + "</blob>")
}

func docHex(payload []byte) []byte {
	return []byte(xmlDecl + "<blob>" + hex.EncodeToString(payload) + "</blob>")
}

func benchForms(payload []byte) map[string][]byte {
	return map[string][]byte{
		"all-references": docAllReferences(payload),
		"byte-mode":      docByteMode(payload),
		"utf8":           docUTF8(payload),
		"base64":         docBase64(payload),
		"hex":            docHex(payload),
	}
}

// Well-formedness only: what the CLI does with no schema.
func BenchmarkValidateBinary(b *testing.B) {
	for name, doc := range benchForms(binaryPayload()) {
		b.Run(name, func(b *testing.B) {
			b.SetBytes(int64(benchPayloadSize))
			b.ReportMetric(float64(len(doc))/float64(benchPayloadSize), "x-size")
			for b.Loop() {
				require.NoError(b, Validate(bytes.NewReader(doc)))

			}
		})
	}
}

// Text with a NUL every 512 bytes: 0.2% of the payload, the density where
// escaping is meant to beat base64.
func BenchmarkValidateSparseNulText(b *testing.B) {
	for name, doc := range benchForms(textPayload(512)) {
		b.Run(name, func(b *testing.B) {
			b.SetBytes(int64(benchPayloadSize))
			b.ReportMetric(float64(len(doc))/float64(benchPayloadSize), "x-size")
			for b.Loop() {
				require.NoError(b, Validate(bytes.NewReader(doc)))

			}
		})
	}
}

// The same payload with no references at all: every byte is one that can
// stand for itself. This separates "base64 is fast" from "references are
// slow" -- if byte mode overtakes base64 here, the cost was never the form.
func BenchmarkValidateNoReferences(b *testing.B) {
	payload := make([]byte, benchPayloadSize)
	for i := range payload {
		// 0xA0..0xFF: literal in byte mode, and none of them is markup.
		payload[i] = byte(0xA0 + i%0x60)
	}
	for name, doc := range benchForms(payload) {
		b.Run(name, func(b *testing.B) {
			b.SetBytes(int64(benchPayloadSize))
			b.ReportMetric(float64(len(doc))/float64(benchPayloadSize), "x-size")
			for b.Loop() {
				require.NoError(b, Validate(bytes.NewReader(doc)))
			}
		})
	}
}

// Building the document from a payload: the producer's side of the same
// choice. The escaping forms walk the payload character by character; base64
// and hex are table-driven passes over bytes.
func BenchmarkEncodeBinary(b *testing.B) {
	payload := binaryPayload()
	encoders := map[string]func([]byte) []byte{
		"all-references": docAllReferences,
		"byte-mode":      docByteMode,
		"utf8":           docUTF8,
		"base64":         docBase64,
		"hex":            docHex,
	}
	for name, encode := range encoders {
		b.Run(name, func(b *testing.B) {
			b.SetBytes(int64(benchPayloadSize))
			for b.Loop() {
				require.NotEqual(b, 0, len(encode(payload)))

			}
		})
	}
}

// Reading the payload back out, which is what a consumer actually does: parse
// the tree, then whatever the form needs to become bytes again.
func BenchmarkRecoverPayload(b *testing.B) {
	payload := binaryPayload()
	forms := benchForms(payload)

	decode := map[string]func(string) ([]byte, error){
		"all-references": latin1Bytes,
		"byte-mode":      latin1Bytes,
		"utf8":           latin1Bytes,
		"base64":         base64.StdEncoding.DecodeString,
		"hex":            hex.DecodeString,
	}

	for name, doc := range forms {
		b.Run(name, func(b *testing.B) {
			b.SetBytes(int64(benchPayloadSize))
			for b.Loop() {
				tree, err := ParseTree(bytes.NewReader(doc))
				require.Nil(b, err)

				got, err := decode[name](tree.Root.TextContent())
				require.Nil(b, err)

				require.Equal(b, len(payload), len(got))

			}
		})
	}
}

// latin1Bytes is the decode step for the three character-carrying forms: each
// character is one byte, which is what the encoder wrote.
func latin1Bytes(s string) ([]byte, error) {
	out := make([]byte, 0, len(s))
	for _, r := range s {
		if r > 0xFF {
			return nil, fmt.Errorf("character U+%04X is outside the payload's range", r)
		}
		out = append(out, byte(r))
	}
	return out, nil
}
