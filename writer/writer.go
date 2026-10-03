// Package writer emits XML 1.1 documents that this repository's validator
// accepts. It is the other half of the reader module: reader turns bytes into
// a tree, writer turns a tree back into bytes.
//
// see docs/encodings.md
package writer

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"github.com/wow-look-at-my/xml-validator/reader"
)

// Encoding is how a document spells the characters it carries.
type Encoding int

const (
	// UTF8 is the default: no encoding declaration, characters as UTF-8.
	UTF8 Encoding = iota
	// Bytes writes one byte per character and declares ISO-8859-1. A
	// character above U+00FF has no byte and becomes a reference.
	Bytes
	// References writes every character as `&#N;`. The result is printable
	// ASCII whatever it carries, at about 5.5x the payload.
	References
)

const (
	utf8Decl  = `<?xml version="1.1"?>`
	bytesDecl = `<?xml version="1.1" encoding="ISO-8859-1"?>`
)

// BinaryEncoding is how a payload of arbitrary bytes is carried.
type BinaryEncoding int

const (
	// Base64 is the default, and it is the default because it is the right
	// answer for arbitrary bytes on every axis measured: 1.33x against byte
	// mode's 2.15x, 3.6x faster to validate and 19x faster to encode. See
	// docs/encodings.md for the numbers and for the case that goes the other
	// way -- text with the occasional NUL, which Text carries better.
	Base64 BinaryEncoding = iota
	// Hex is xs:hexBinary: twice the payload, and readable.
	Hex
	// Text spells the payload as characters, escaping only what cannot stand
	// for itself. Smaller than Base64 only when little of the payload needs
	// escaping, which is text, not blobs.
	Text
)

// Options are the choices a document makes about its own bytes.
type Options struct {
	Encoding Encoding
	// Binary is the form WriteBinary uses. The zero value is Base64.
	Binary BinaryEncoding
}

// WriteDocument writes doc as XML 1.1. The tree came from reader.ParseTree, or
// was built by hand; either way what comes out parses back to an equal tree.
func WriteDocument(w io.Writer, doc *reader.Document, opts Options) error {
	if doc == nil || doc.Root == nil {
		return fmt.Errorf("writer: document has no root element")
	}
	var b strings.Builder
	b.WriteString(declaration(opts.Encoding))
	if err := writeElement(&b, doc.Root, opts.Encoding); err != nil {
		return err
	}
	return emit(w, b.String(), opts.Encoding)
}

// WriteBinary writes payload as a single element, in the form Options.Binary
// names. With the zero Options that is base64, which is what a payload of
// arbitrary bytes should use.
func WriteBinary(w io.Writer, element string, payload []byte, opts Options) error {
	var content string
	switch opts.Binary {
	case Base64:
		content = base64.StdEncoding.EncodeToString(payload)
	case Hex:
		content = strings.ToUpper(hex.EncodeToString(payload))
	case Text:
		return writeBinaryAsText(w, element, payload, opts)
	default:
		return fmt.Errorf("writer: unknown binary encoding %d", opts.Binary)
	}
	// base64 and hex are ASCII whatever the payload was, so the document's
	// own encoding does not change what they spell.
	_, err := io.WriteString(w, utf8Decl+"<"+element+">"+content+"</"+element+">")
	return err
}

// writeBinaryAsText carries each byte as the character with that value, which
// is what byte mode is for. Anything that cannot stand for itself becomes a
// reference, and for arbitrary bytes that is about a quarter of them.
func writeBinaryAsText(w io.Writer, element string, payload []byte, opts Options) error {
	encoding := opts.Encoding
	if encoding == UTF8 {
		encoding = Bytes
	}
	var text strings.Builder
	for _, c := range payload {
		text.WriteRune(rune(c))
	}
	var b strings.Builder
	b.WriteString(declaration(encoding))
	b.WriteString("<" + element + ">")
	b.WriteString(escape(text.String(), false, encoding))
	b.WriteString("</" + element + ">")
	return emit(w, b.String(), encoding)
}

func declaration(e Encoding) string {
	if e == Bytes {
		return bytesDecl
	}
	return utf8Decl
}

// emit writes the document's characters as the bytes its encoding calls for.
// In byte mode that is one byte each, which is the whole point of the mode;
// writing the string as Go holds it would encode them as UTF-8 instead.
func emit(w io.Writer, doc string, e Encoding) error {
	if e != Bytes {
		_, err := io.WriteString(w, doc)
		return err
	}
	out := make([]byte, 0, len(doc))
	for _, r := range doc {
		if r > 0xFF {
			return fmt.Errorf("writer: character U+%04X has no byte in ISO-8859-1", r)
		}
		out = append(out, byte(r))
	}
	_, err := w.Write(out)
	return err
}

func writeElement(b *strings.Builder, e *reader.Element, enc Encoding) error {
	b.WriteString("<" + e.Name)
	for _, a := range e.Attrs {
		b.WriteString(" " + a.Name + `="` + escape(a.Value, true, enc) + `"`)
	}
	if len(e.Children) == 0 {
		b.WriteString("/>")
		return nil
	}
	b.WriteString(">")
	for _, c := range e.Children {
		switch n := c.(type) {
		case *reader.Element:
			if err := writeElement(b, n, enc); err != nil {
				return err
			}
		case *reader.CharData:
			b.WriteString(escape(n.Content, false, enc))
		}
	}
	b.WriteString("</" + e.Name + ">")
	return nil
}

// escape writes a character as itself where XML allows that, and as a
// reference where it does not.
func escape(s string, inAttr bool, enc Encoding) string {
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
		case r == 0 || reader.IsRestrictedChar(r):
			fmt.Fprintf(&b, "&#%d;", r)
		// CR, NEL and LINE SEPARATOR normalize to LF as literal characters,
		// so a value that carries one has to write a reference.
		case r == '\r' || r == 0x85 || r == 0x2028:
			fmt.Fprintf(&b, "&#%d;", r)
		// Tab and newline are legal in an attribute value, but a conforming
		// reader folds both to a space.
		case inAttr && (r == '\t' || r == '\n'):
			fmt.Fprintf(&b, "&#%d;", r)
		// Byte mode has no byte above U+00FF, and References writes
		// everything out whatever it is.
		case enc == References || (enc == Bytes && r > 0xFF):
			fmt.Fprintf(&b, "&#%d;", r)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
