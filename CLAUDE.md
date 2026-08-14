# xml-validator

Strict XML 1.1 validator with XSD schema validation. Anything unsupported is a hard error.

Ships as **both** a CLI and a Go library that share the same engine. Keep both
working when making changes -- the `validator` package is the public library
surface, and `cmd/` + `main.go` are the CLI shell over it.

## Build & Test

```bash
go-toolchain
```

This handles `go mod tidy`, `go vet`, tests with coverage, and builds the binary to `build/xml-validator`.

## CLI usage

```bash
xml-validator <file>                    # well-formedness validation only
xml-validator                           # validate from stdin
xml-validator --schema schema.xsd file  # validate against XSD schema
```

Exit 0 on valid XML 1.1, exit 1 with error message on failure.

## Library usage

Importable as `github.com/wow-look-at-my/xml-validator/validator`. Public
entry points:

- `Validate(io.Reader) error` -- well-formedness only.
- `ValidateWithSchema(xml, xsd io.Reader) error` -- well-formedness + XSD.
- `ValidateWithSchemaBytes(xml, xsd []byte) error` -- byte-oriented form.
- `ValidateWithSchemaResolver(xml, xsd []byte, SchemaResolver) error` -- as above
  but with a callback for resolving `xs:import schemaLocation` hints.
- `ValidateWithSchemaFile(xmlPath, xsdPath string) error` -- convenience that
  reads both files from disk and resolves imports relative to the XSD's
  directory.
- `FileSchemaResolver(baseDir string) SchemaResolver` -- filesystem-backed
  resolver for use with `ParseSchemaWithResolver` / `ValidateWithSchemaResolver`.
- `ParseTree(io.Reader) (*Document, error)` -- parse to tree without validating.
- `ParseSchema(*Document) (*Schema, error)` -- parse an XSD tree to a schema model.
- `ParseSchemaWithResolver(*Document, SchemaResolver) (*Schema, error)` -- as
  above but follows `xs:import` and `xs:include` directives via the resolver.
- `ValidateSchema(*Document, *Schema) error` -- validate a parsed tree against a parsed schema.

Errors come back as `*validator.Error` with `Line`, `Col`, and `Message`.

Tree nodes carry positions too: `Element` and `Attr` both have `Line` and `Col`. A consumer reporting a problem with an
attribute's value points at that attribute, not at the element that owns it -- on a multi-attribute element those are
different places.

## Supported

- XML 1.1 declaration (required)
- Elements, attributes, text, CDATA sections, comments, PIs
- Character references (`&#N;`, `&#xN;`) and predefined entity references (`&amp;` `&lt;` `&gt;` `&apos;` `&quot;`)
- Namespace validation (Namespaces in XML 1.1)
- UTF-8 only, no BOM (per utf8everywhere)
- XML 1.1 line ending normalization (`#x85`, `#x2028`)
- XML 1.1 character and name character classes
- Restricted character enforcement (must use character references)
- XSD schema validation (`--schema`):
  - Element declarations (global, local, refs)
  - Complex types (sequence, choice, all content models)
  - Simple types (restriction with facets, list, union)
  - Attributes (required/optional/prohibited, type checking, fixed values)
  - Named types, groups, attribute groups
  - simpleContent and complexContent (extension/restriction)
  - 35+ built-in XSD types (string, integer, boolean, decimal, date, etc.)
  - Facets: enumeration, pattern, minLength, maxLength, length, min/maxInclusive, min/maxExclusive, totalDigits, fractionDigits
  - minOccurs/maxOccurs enforcement
  - xs:any wildcard particles
  - xs:anyAttribute wildcard attributes
  - xs:import with optional schemaLocation (loaded via a `SchemaResolver`;
    the CLI and `ValidateWithSchemaFile` wire a filesystem-backed one)
  - xs:include with schemaLocation (same resolver mechanism as xs:import)

## Hard Errors (Unsupported)

- `processContents="skip"` and `processContents="lax"` on `xs:any` / `xs:anyAttribute` -- only `strict` (the default) is allowed; this validator does not offer a no- or partial-validation mode
- DOCTYPE declarations
- General entity references (beyond the 5 predefined)
- XML 1.0 documents
- Missing XML declaration
- Encodings other than UTF-8 (UTF-16 inputs and UTF-8 BOMs are rejected; any encoding declaration other than UTF-8 is rejected)
- xs:redefine, xs:override
- xs:notation
- Identity constraints (xs:key, xs:keyref, xs:unique)
- Type substitution (substitutionGroup)

## Project Structure

- `cmd/root.go` -- CLI (cobra)
- `validator/doc.go` -- package-level godoc
- `validator/validator.go` -- public `Validate`, `ValidateWithSchema`, and `ValidateWithSchemaBytes` entry points
- `validator/example_test.go` -- runnable godoc examples
- `validator/parser.go` -- recursive descent parser core, XML declaration, comments, PIs
- `validator/elements.go` -- element, attribute, content, CDATA, reference parsing
- `validator/namespace.go` -- QName validation, namespace scope management
- `validator/chars.go` -- XML 1.1 character class predicates
- `validator/reader.go` -- UTF-8 decoding, BOM/UTF-16 detection-and-reject, line normalization
- `validator/error.go` -- error type with line/column position
- `validator/document.go` -- document tree model (Element, Attr, CharData)
- `validator/tree.go` -- version-agnostic XML tree parser
- `validator/schema_model.go` -- XSD schema model types
- `validator/schema_builtin.go` -- 35+ built-in XSD types with validation
- `validator/schema_parse.go` -- XSD file parser (document tree to schema model)
- `validator/schema_import.go` -- `SchemaResolver` and `xs:import` handling
- `validator/schema_validate.go` -- schema validation engine
- `action.yml` -- composite GitHub Action (build with caching + run). The
  cache-key, build, and validation steps run as TypeScript (tsc-checked, Node)
  via `wow-look-at-my/actions@typescript#latest` instead of inline shell, so the
  logic stays portable across runner OSes; the validator is invoked once per file.
