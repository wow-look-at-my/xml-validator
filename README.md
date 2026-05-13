# xml-validator

Strict XML 1.1 validator with optional XSD schema validation. Anything the
validator does not understand is a hard error -- there is no fallback to XML
1.0, no DTD support, no permissive mode.

Ships as **both** a command-line tool and a Go library that share the same
parser and validation engine.

## Install

### CLI

```bash
go install github.com/wow-look-at-my/xml-validator@latest
```

### Library

```bash
go get github.com/wow-look-at-my/xml-validator/validator
```

## CLI usage

```bash
xml-validator <file>                    # well-formedness validation only
xml-validator                           # validate from stdin
xml-validator --schema schema.xsd file  # validate against XSD schema
```

Exit code is `0` on a valid XML 1.1 document, `1` with an error message on
failure.

## Library usage

The `validator` package exposes the same engine the CLI uses. Validation
errors are returned as `*validator.Error` with the offending line and column.

### Well-formedness only

```go
import "github.com/wow-look-at-my/xml-validator/validator"

err := validator.Validate(reader)
if err != nil {
    // err.(*validator.Error) carries Line, Col, Message
}
```

### With an XSD schema

Several helpers are provided -- pick whichever fits your call site:

```go
// io.Reader form
err := validator.ValidateWithSchema(xmlReader, xsdReader)

// []byte form
err := validator.ValidateWithSchemaBytes(xmlBytes, xsdBytes)

// File-based form (resolves xs:import schemaLocation hints relative to
// the directory containing xsdPath)
err := validator.ValidateWithSchemaFile(xmlPath, xsdPath)

// Custom xs:import resolver (e.g. fetch over HTTP, read from a registry)
err := validator.ValidateWithSchemaResolver(xmlBytes, xsdBytes,
    func(namespace, schemaLocation string) ([]byte, error) {
        return loadFromSomewhere(schemaLocation)
    })
```

### Parsing without validating

If you want to walk the document tree or inspect a schema yourself, use the
lower-level entry points:

```go
xmlDoc, err := validator.ParseTree(xmlReader)
xsdDoc, _  := validator.ParseTree(xsdReader)
schema, _  := validator.ParseSchema(xsdDoc)
err = validator.ValidateSchema(xmlDoc, schema)
```

`*validator.Document`, `*validator.Element`, `validator.Attr`, and
`*validator.Schema` are public types -- see the godoc for the full surface.

## What is supported

- XML 1.1 declaration (required)
- Elements, attributes, text content, CDATA sections, comments, PIs
- Character references (`&#N;`, `&#xN;`) and the five predefined entities
- Namespaces in XML 1.1
- UTF-8 and UTF-16 (BE/LE) with BOM detection
- XML 1.1 line-ending normalization (`#x85`, `#x2028`)
- XSD schema validation: complex/simple types, facets, sequence/choice/all,
  attribute groups, simpleContent/complexContent, `xs:any`, and 35+ built-in
  types
- `xs:import` (with optional `schemaLocation`). The CLI and
  `ValidateWithSchemaFile` resolve hints relative to the importing schema's
  directory; library callers can supply a custom `SchemaResolver`
- `xs:include` (with `schemaLocation`). Uses the same resolver mechanism as
  `xs:import`

## What is rejected as unsupported

- DOCTYPE declarations
- General entity references beyond the five predefined ones
- XML 1.0 documents (the declaration must say `version="1.1"`)
- Missing XML declaration
- Encodings other than UTF-8 / UTF-16
- XSD: `xs:redefine`, `xs:override`, `xs:notation`,
  identity constraints (`xs:key` / `xs:keyref` / `xs:unique`), and
  `substitutionGroup`

## Build & test

```bash
go-toolchain
```

That runs `go mod tidy`, `go vet`, tests with coverage, and produces the
binary at `build/xml-validator`.
