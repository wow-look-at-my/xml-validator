package validator

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
//
// Imports in the schema (xs:import) without a schemaLocation are accepted as
// namespace declarations. Imports with a schemaLocation produce an error
// because no resolver is available; use [ValidateWithSchemaResolver] or
// [ValidateWithSchemaFile] to support those.
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
	return ValidateWithSchemaResolver(xmlData, xsdData, nil)
}

// ValidateWithSchemaResolver validates xmlData against xsdData using resolver
// to load any xs:import directives that name a schemaLocation. Passing a nil
// resolver behaves like [ValidateWithSchemaBytes].
func ValidateWithSchemaResolver(xmlData, xsdData []byte, resolver SchemaResolver) error {
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

	schema, err := ParseSchemaWithResolver(xsdDoc, resolver)
	if err != nil {
		return &Error{Message: "schema: " + err.Error()}
	}

	xmlDoc, err := ParseTree(bytes.NewReader(xmlData))
	if err != nil {
		return &Error{Message: "parsing document tree: " + err.Error()}
	}

	return ValidateSchema(xmlDoc, schema)
}

// ValidateWithSchemaFile reads xmlPath and xsdPath from disk and validates
// the XML document against the schema. xs:import directives with relative
// schemaLocation hints are resolved against the directory containing xsdPath.
func ValidateWithSchemaFile(xmlPath, xsdPath string) error {
	xmlData, err := os.ReadFile(xmlPath)
	if err != nil {
		return fmt.Errorf("reading XML file: %w", err)
	}
	xsdData, err := os.ReadFile(xsdPath)
	if err != nil {
		return fmt.Errorf("reading XSD file: %w", err)
	}
	resolver := FileSchemaResolver(filepath.Dir(xsdPath))
	return ValidateWithSchemaResolver(xmlData, xsdData, resolver)
}

// FileSchemaResolver returns a [SchemaResolver] that loads schemaLocation
// hints from the local filesystem. Relative paths are resolved against
// baseDir; absolute paths are used as-is. The namespace argument is ignored.
func FileSchemaResolver(baseDir string) SchemaResolver {
	return func(_, schemaLocation string) ([]byte, error) {
		path := schemaLocation
		if !filepath.IsAbs(path) {
			path = filepath.Join(baseDir, path)
		}
		return os.ReadFile(path)
	}
}
