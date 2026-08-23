# xml-validator

Strict XML 1.1 validator with optional XSD schema validation. Anything the
validator does not understand is a hard error -- there is no fallback to XML
1.0, no DTD support, no permissive mode.

Ships as a command-line tool and as four Go modules, so a program takes only
the part it needs.

| module | what it is | depends on |
|---|---|---|
| `.../reader` | bytes to a tree: decoding, character classes, the tree parser | nothing |
| `.../writer` | a tree back to bytes; base64 by default for a byte payload | reader |
| `.../validator` | well-formedness and XSD, on top of reader | reader |
| `.../cli` | the `xml-validator` command | validator |

## Install

### CLI

```bash
go install github.com/wow-look-at-my/xml-validator/cli/cmd/xml-validator@latest
```

### Library

```bash
go get github.com/wow-look-at-my/xml-validator/validator   # validate
go get github.com/wow-look-at-my/xml-validator/reader      # just read
go get github.com/wow-look-at-my/xml-validator/writer      # just write
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
- Character references (`&#N;`, `&#xN;`) and the five predefined entities.
  `&#0;` is accepted, one deliberate deviation from XML 1.1 -- a literal NUL
  byte and a lone surrogate are still rejected
- Namespaces in XML 1.1
- Two input modes: UTF-8 (the default) and an 8-bit byte mode selected by
  `encoding="ISO-8859-1"`, where one byte is one character and anything above
  U+00FF takes a character reference. No BOM in either
  (per [utf8everywhere](https://utf8everywhere.org/))
- XML 1.1 line-ending normalization (`#x85`, `#x2028`)
- XSD schema validation: complex/simple types, facets, sequence/choice/all,
  attribute groups, simpleContent/complexContent, `xs:any`, and 35+ built-in
  types
- Facets constrain attribute values and element text alike, and a derived
  simple type enforces the facets it inherits from its base. A violation on an
  attribute is reported at that attribute's own line and column
- Global `xs:attribute` declarations and `xs:attribute ref=`. Attributes match
  on namespace and local name together, so a qualified attribute from an
  imported vocabulary is type-checked rather than assumed valid
- `xs:import` (with optional `schemaLocation`). The CLI and
  `ValidateWithSchemaFile` resolve hints relative to the importing schema's
  directory; library callers can supply a custom `SchemaResolver`
- `xs:include` (with `schemaLocation`). Uses the same resolver mechanism as
  `xs:import`
- Identity constraints (`xs:key`, `xs:keyref`, `xs:unique`) over the XPath
  subset XSD defines for a selector and a field. Anything outside that subset
  is rejected, so a constraint either runs or says it cannot
- Substitution groups: a member may appear wherever its head is referenced and
  is validated against its own type. `abstract` heads and `block="substitution"`
  are honored
- `xs:alternative` conditional type assignment, with tests over the element's
  own attributes. The test language is the required subset XSD 1.1 defines --
  comparisons, `and`/`or` with parentheses, `not()`, casts and constructor
  functions -- and a test outside it is rejected rather than guessed at

## What is rejected as unsupported

- DOCTYPE declarations
- General entity references beyond the five predefined ones
- XML 1.0 documents (the declaration must say `version="1.1"`)
- Missing XML declaration
- Encodings other than UTF-8 and ISO-8859-1 (UTF-16 inputs and BOMs are rejected)
- XSD: `xs:redefine`, `xs:override`, and `xs:notation`
- A selector or field XPath outside the subset XSD defines for them -- a
  predicate, a function, an axis other than `child::`/`attribute::`, or `..`
- `processContents="skip"` and `processContents="lax"` on `xs:any` /
  `xs:anyAttribute` -- only `strict` (the default) is allowed. This tool
  always validates: every element matched by a wildcard must have a global
  declaration the validator can find. If you do not want validation, do
  not run the validator.

## GitHub Action

Use `wow-look-at-my/xml-validator` as a GitHub Action to validate XML files in
CI. The action installs the released binary from buildhost and runs it once per
file. Where buildhost has no binary that runs on the runner, it builds from
source instead, with caching, and says so in the log.

### Zero-config (recommended)

With no inputs, the action auto-discovers every `*.xml` and `*.xsd` file in the
workspace and checks each one for XML 1.1 well-formedness:

```yaml
- uses: wow-look-at-my/xml-validator@master
```

### Validate against an XSD schema

Point `schema` at an XSD to validate every file against it:

```yaml
- uses: wow-look-at-my/xml-validator@master
  with:
    schema: 'schema.xsd'
```

### Explicit files

```yaml
- uses: wow-look-at-my/xml-validator@master
  with:
    files: 'doc.xml config.xml'
    schema: 'schema.xsd'
```

### Inputs

| Input | Required | Description |
|-------|----------|-------------|
| `files` | No | Space-separated files or glob patterns to validate. When omitted, auto-discovers every `*.xml` and `*.xsd` file in the workspace. |
| `schema` | No | Path to an XSD schema to validate every file against. When omitted, only XML 1.1 well-formedness is checked. |
| `args` | No | Additional CLI arguments passed to each invocation. |

## Build & test

```bash
go-toolchain
```

Run it at the repository root and it walks all four modules: tidy, vet, tests
with coverage, the build, and the CLI suites under `cli/dats/`. The binary
lands at `cli/build/xml-validator`.
