package validator

import (
	"bytes"
	"github.com/wow-look-at-my/xml-validator/reader"
	"io"
	"os"
	"path/filepath"
)

// asError converts any error into a *[Error] so the public Validate* entry
// points always return *Error.
func asError(err error) *Error {
	if e, ok := err.(*Error); ok {
		return e
	}
	return &Error{Line: 1, Col: 1, Message: err.Error()}
}

// On failure it returns a *[Error] with the line and column of the problem.
func Validate(r io.Reader) error {
	runes, err := reader.Decode(r)
	if err != nil {
		return asError(err)
	}
	p := newParser(runes)
	return p.parseDocument()
}

// On failure it returns a *[Error].
func ValidateWithSchema(xml, xsd io.Reader) error {
	xmlData, err := io.ReadAll(xml)
	if err != nil {
		return &Error{Line: 1, Col: 1, Message: "reading XML input: " + err.Error()}
	}
	xsdData, err := io.ReadAll(xsd)
	if err != nil {
		return &Error{Line: 1, Col: 1, Message: "reading XSD input: " + err.Error()}
	}
	return ValidateWithSchemaBytes(xmlData, xsdData)
}

// ValidateWithSchemaBytes is the byte-oriented form of [ValidateWithSchema].
func ValidateWithSchemaBytes(xmlData, xsdData []byte) error {
	return ValidateWithSchemaResolver(xmlData, xsdData, nil)
}

// ValidateWithSchemaResolver validates xmlData against xsdData using resolver
// to load any xs:import directives that name a schemaLocation. Passing a nil
// resolver behaves like [ValidateWithSchemaBytes].
func ValidateWithSchemaResolver(xmlData, xsdData []byte, resolver SchemaResolver) error {
	runes, err := reader.Decode(bytes.NewReader(xmlData))
	if err != nil {
		return asError(err)
	}
	p := newParser(runes)
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
		return &Error{Line: 1, Col: 1, Message: "reading XML file: " + err.Error()}
	}
	xsdData, err := os.ReadFile(xsdPath)
	if err != nil {
		return &Error{Line: 1, Col: 1, Message: "reading XSD file: " + err.Error()}
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
