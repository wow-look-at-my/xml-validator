package validator

import "io"

func Validate(r io.Reader) error {
	runes, enc, err := readInput(r)
	if err != nil {
		return err
	}
	p := newParser(runes, enc)
	return p.parseDocument()
}

func ValidateWithSchema(xmlReader, xsdReader io.Reader) error {
	if err := Validate(xmlReader); err != nil {
		return err
	}

	xsdDoc, err := ParseTree(xsdReader)
	if err != nil {
		return &Error{Line: 0, Col: 0, Message: "schema: " + err.Error()}
	}

	schema, err := ParseSchema(xsdDoc)
	if err != nil {
		return &Error{Line: 0, Col: 0, Message: "schema: " + err.Error()}
	}

	// Re-parse the XML into a tree for schema validation.
	// The xmlReader is already consumed, so we need the caller to provide seekable input.
	// This is handled by the CLI reading the file into memory first.
	return &Error{Line: 0, Col: 0, Message: "internal: use ValidateWithSchemaFromBytes"}
}

func ValidateWithSchemaBytes(xmlData, xsdData []byte) error {
	runes, enc, err := readInput(newBytesReader(xmlData))
	if err != nil {
		return err
	}
	p := newParser(runes, enc)
	if err := p.parseDocument(); err != nil {
		return err
	}

	xsdDoc, err := ParseTree(newBytesReader(xsdData))
	if err != nil {
		return &Error{Line: 0, Col: 0, Message: "schema: " + err.Error()}
	}

	schema, err := ParseSchema(xsdDoc)
	if err != nil {
		return &Error{Line: 0, Col: 0, Message: "schema: " + err.Error()}
	}

	xmlDoc, err := ParseTree(newBytesReader(xmlData))
	if err != nil {
		return &Error{Line: 0, Col: 0, Message: "parsing document tree: " + err.Error()}
	}

	return ValidateSchema(xmlDoc, schema)
}

type bytesReader struct {
	data []byte
	pos  int
}

func newBytesReader(data []byte) *bytesReader {
	return &bytesReader{data: data}
}

func (r *bytesReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
