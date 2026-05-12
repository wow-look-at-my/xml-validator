package validator

import (
	"bytes"
	"fmt"
	"io"
)

// Validate verifies that the input is a well-formed XML 1.1 document.
// On failure it returns a *[Error] with the line and column of the problem.
func Validate(r io.Reader) error {
	runes, enc, err := readInput(r)
	if err != nil {
		return err
	}
	p := newParser(runes, enc)
	return p.parseDocument()
}

// ValidateWithSchema runs XML 1.1 well-formedness validation on xml, parses
// the XSD schema from xsd, and then validates the document against the schema.
// On failure it returns a *[Error].
func ValidateWithSchema(xml, xsd io.Reader) error {
	xmlData, err := io.ReadAll(xml)
	if err != nil {
		return fmt.Errorf("reading XML input: %w", err)
	}
	xsdData, err := io.ReadAll(xsd)
	if err != nil {
		return fmt.Errorf("reading XSD input: %w", err)
	}
	return ValidateWithSchemaBytes(xmlData, xsdData)
}

// ValidateWithSchemaBytes is the byte-oriented form of [ValidateWithSchema].
// It is useful when the document and schema are already in memory.
func ValidateWithSchemaBytes(xmlData, xsdData []byte) error {
	runes, enc, err := readInput(bytes.NewReader(xmlData))
	if err != nil {
		return err
	}
	p := newParser(runes, enc)
	if err := p.parseDocument(); err != nil {
		return err
	}

	xsdDoc, err := ParseTree(bytes.NewReader(xsdData))
	if err != nil {
		return &Error{Message: "schema: " + err.Error()}
	}

	schema, err := ParseSchema(xsdDoc)
	if err != nil {
		return &Error{Message: "schema: " + err.Error()}
	}

	xmlDoc, err := ParseTree(bytes.NewReader(xmlData))
	if err != nil {
		return &Error{Message: "parsing document tree: " + err.Error()}
	}

	return ValidateSchema(xmlDoc, schema)
}
