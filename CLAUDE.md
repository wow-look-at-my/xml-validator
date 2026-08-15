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
- Two input modes, picked by the encoding declaration: UTF-8 (the default) and
  an 8-bit byte mode (`ISO-8859-1` and its aliases) where byte b is the
  character U+00XX and anything above U+00FF needs a character reference.
  Neither takes a BOM (per utf8everywhere). See `docs/encodings.md`
- XML 1.1 line ending normalization (`#x85`, `#x2028`)
- XML 1.1 character and name character classes
- Restricted character enforcement (must use character references)
- `&#0;` -- one deliberate deviation from the XML 1.1 `Char` production, which
  starts at `#x1`. A reference resolving to U+0000 is accepted and carried
  through to the parsed value; a literal NUL byte and a lone surrogate are still
  rejected. See `IsCharRefValue` in `validator/chars.go`, and
  `docs/nul-char-ref.md` -- the reference is not a NUL byte and not a document
  terminator, proved at both the library and the CLI layer.
- XSD schema validation (`--schema`):
  - Element declarations (global, local, refs)
  - Complex types (sequence, choice, all content models)
  - Simple types (restriction with facets, list, union)
  - Attributes (required/optional/prohibited, type checking, fixed values)
  - Global `xs:attribute` declarations and `xs:attribute ref=`, matched by
    namespace AND local name -- a qualified attribute resolves to the global
    declaration in its own namespace, so an imported vocabulary's attributes are
    type-checked and an undeclared one is an error rather than something a
    wildcard waves through
  - Named types, groups, attribute groups. An `xs:group ref` is replaced by a
    copy of the group's content model, carrying the occurrence counts stated at
    the reference
  - simpleContent and complexContent (extension/restriction). An extension's
    content follows its base's and its attributes add to the base's; a
    restriction states its own content model but still inherits the base's
    attributes where it does not restate them
  - 35+ built-in XSD types (string, integer, boolean, decimal, date, etc.)
  - Facets: enumeration, pattern, minLength, maxLength, length, min/maxInclusive, min/maxExclusive, totalDigits, fractionDigits.
    A length facet counts characters, and octets for `xs:hexBinary` and
    `xs:base64Binary` -- never the bytes of the UTF-8 encoding. A facet value
    that is not a non-negative integer is an error, not a length of 0
    An attribute value and element text take the same path, so every facet
    constrains both. A restriction of a restriction enforces the facets it
    inherits as well as the ones it states, and on a list the length facets
    count items rather than characters
  - `xs:list` and `xs:union` with an `itemType`/`memberTypes` attribute or with
    inline `xs:simpleType` members; each item or member carries its own facets
  - minOccurs/maxOccurs enforcement
  - xs:any wildcard particles
  - xs:anyAttribute wildcard attributes
  - xs:import with optional schemaLocation (loaded via a `SchemaResolver`;
    the CLI and `ValidateWithSchemaFile` wire a filesystem-backed one)
  - xs:include with schemaLocation (same resolver mechanism as xs:import)
  - Identity constraints: `xs:key`, `xs:keyref`, `xs:unique`. See
    `docs/identity-constraints.md` for the XPath subset, the two deliberate
    deviations, and how a keyref finds its key

## Hard Errors (Unsupported)

- `xs:all` anywhere but as a whole complex type's content model, and anything
  but element declarations and `xs:any` inside one -- order-free matching is
  defined over a whole element, so a nested one matched nothing at all
- An `xs:element`/`xs:attribute`/`xs:group` `ref`, or a `complexContent` base,
  that names nothing -- an unresolvable reference used to leave a hole in the
  content model and report itself as a missing instance element
- `processContents="skip"` and `processContents="lax"` on `xs:any` / `xs:anyAttribute` -- only `strict` (the default) is allowed; this validator does not offer a no- or partial-validation mode
- DOCTYPE declarations
- General entity references (beyond the 5 predefined)
- XML 1.0 documents
- Missing XML declaration
- Encodings other than UTF-8 and ISO-8859-1 (UTF-16 inputs and BOMs are rejected; any other encoding declaration is rejected by name)
- xs:redefine, xs:override
- xs:notation
- Any unrecognized child of an `xs:element`, and any selector or field XPath
  outside the supported subset -- a constraint either runs or says it cannot
- Type substitution (substitutionGroup)
- An `xs:list` with no item type and an `xs:union` with no members -- both used
  to accept every value
- A simple type that derives from itself

## Project Structure

- `cmd/root.go` -- CLI (cobra)
- `validator/doc.go` -- package-level godoc
- `validator/validator.go` -- public `Validate`, `ValidateWithSchema`, and `ValidateWithSchemaBytes` entry points
- `validator/example_test.go` -- runnable godoc examples
- `validator/parser.go` -- recursive descent parser core, XML declaration, comments, PIs
- `validator/elements.go` -- element, attribute, content, CDATA, reference parsing
- `validator/namespace.go` -- QName validation, namespace scope management
- `validator/chars.go` -- XML 1.1 character class predicates
- `validator/reader.go` -- decoder selection, BOM/UTF-16 detection-and-reject, line normalization
- `validator/encoding.go` -- the two modes, the alias table, declaration sniffing, both decoders
- `validator/error.go` -- error type with line/column position
- `validator/document.go` -- document tree model (Element, Attr, CharData)
- `validator/tree.go` -- version-agnostic XML tree parser
- `validator/schema_model.go` -- XSD schema model types
- `validator/schema_builtin.go` -- 35+ built-in XSD types with validation
- `validator/schema_facets.go` -- facet checking (lengths, patterns, ranges, digits)
- `validator/schema_parse.go` -- XSD file parser (document tree to schema model)
- `validator/schema_import.go` -- `SchemaResolver` and `xs:import` handling
- `validator/schema_qname.go` -- namespace-keyed lookup of global declarations
- `validator/schema_resolve.go` -- ref and type resolution over a parsed schema
- `validator/schema_derive.go` -- group-ref expansion and complexContent derivation
- `validator/schema_parse_simple.go` -- xs:simpleType parsing (restriction facets, list, union)
- `validator/schema_identity.go` -- xs:key/xs:keyref/xs:unique: XPath subset, selection, evaluation
- `validator/schema_value.go` -- simple-value validation (facets, list, union) shared by element text and attribute values
- `validator/schema_validate.go` -- schema validation engine
- `validator/roundtrip_nul_test.go` -- parse/emit/reparse roundtrips for `&#0;`
  and for a payload carrying every byte value 0..255
- `validator/roundtrip_binary_test.go` -- the same payload through
  `xs:base64Binary` and `xs:hexBinary`, with the size of each wire form
- `dats/nul-char-ref.dats` -- CLI suite for `&#0;`, run by `go-toolchain` after
  every build
- `docs/nul-char-ref.md` -- what `&#0;` is, what it is not, and how to check
- `action.yml` -- composite GitHub Action (build with caching + run). The
  cache-key, build, and validation steps run as TypeScript (tsc-checked, Node)
  via `wow-look-at-my/actions@typescript#latest` instead of inline shell, so the
  logic stays portable across runner OSes; the validator is invoked once per file.
