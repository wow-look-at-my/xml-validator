package validator

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
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

func docAllReferences(payload []byte) string {
	var b strings.Builder
	b.WriteString(xmlDecl + "<r>")
	for _, c := range payload {
		fmt.Fprintf(&b, "&#%d;", c)
	}
	b.WriteString("</r>")
	return b.String()
}

func docByteMode(payload []byte) string {
	var b strings.Builder
	b.WriteString(byteDecl + "<r>")
	b.WriteString(escapeXML(string(encodeBytes(payload)), false))
	b.WriteString("</r>")
	return b.String()
}

func docUTF8(payload []byte) string {
	return xmlDecl + "<r>" + escapeXML(string(encodeBytes(payload)), false) + "</r>"
}

func docBase64(payload []byte) string {
	return xmlDecl + "<blob>" + base64.StdEncoding.EncodeToString(payload) + "</blob>"
}

func docHex(payload []byte) string {
	return xmlDecl + "<blob>" + hex.EncodeToString(payload) + "</blob>"
}

func benchForms(payload []byte) map[string]string {
	return map[string]string{
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
				if err := Validate(strings.NewReader(doc)); err != nil {
					b.Fatal(err)
				}
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
				if err := Validate(strings.NewReader(doc)); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// Building the document from a payload: the producer's side of the same
// choice. The escaping forms walk the payload character by character; base64
// and hex are table-driven passes over bytes.
func BenchmarkEncodeBinary(b *testing.B) {
	payload := binaryPayload()
	encoders := map[string]func([]byte) string{
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
				if len(encode(payload)) == 0 {
					b.Fatal("empty document")
				}
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
				tree, err := ParseTree(strings.NewReader(doc))
				if err != nil {
					b.Fatal(err)
				}
				got, err := decode[name](tree.Root.TextContent())
				if err != nil {
					b.Fatal(err)
				}
				if len(got) != len(payload) {
					b.Fatalf("recovered %d bytes, want %d", len(got), len(payload))
				}
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
