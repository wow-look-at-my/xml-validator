# xml-validator

Strict XML 1.1 validator with XSD schema validation. Anything unsupported is a hard error.

## Four modules, imported separately

The repository root carries no `go.mod`. Each module is its own. A program takes only the part it needs, and `go-toolchain` gates all four.

| module | import path | depends on |
|---|---|---|
| reader | `github.com/wow-look-at-my/xml-validator/reader` | nothing |
| writer | `github.com/wow-look-at-my/xml-validator/writer` | reader |
| validator | `github.com/wow-look-at-my/xml-validator/validator` | reader |
| cli | `github.com/wow-look-at-my/xml-validator/cli` | validator |

`reader` turns bytes into a tree: decoding, the character classes, the tree model, the tree parser. `writer` turns a tree back into bytes. `validator` is well-formedness and XSD, built on reader. `cli` is the cobra shell, and its `dats/` suites drive the built binary.

Each module `replace`s its siblings to the directory they are in. A build then uses the tree it is in, not a published version of itself.

## Build & Test

```bash
go-toolchain
```

Run it at the repository root and it walks every module: tidy, vet, tests with coverage, the build, and the dats suites in `cli/`. The CLI binary lands at `cli/build/xml-validator`.

CI gates one module per job instead, each with `working-directory:`. The action runs `go-toolchain matrix`, which builds the module it stands in and does not walk a tree of modules. Only the `cli` job publishes.

## CLI usage

```bash
xml-validator <file>                    # well-formedness validation only
xml-validator                           # validate from stdin
xml-validator --schema schema.xsd file  # validate against XSD schema
```

Exit 0 on valid XML 1.1, exit 1 with error message on failure.

## Library usage

Importable as `github.com/wow-look-at-my/xml-validator/validator`. Public entry points:

- `Validate(io.Reader) error` -- well-formedness only.
- `ValidateWithSchema(xml, xsd io.Reader) error` -- well-formedness + XSD.
- `ValidateWithSchemaBytes(xml, xsd []byte) error` -- byte-oriented form.
- `ValidateWithSchemaResolver(xml, xsd []byte, SchemaResolver) error` -- as above but with a callback for resolving `xs:import schemaLocation` hints.
- `ValidateWithSchemaFile(xmlPath, xsdPath string) error` -- convenience that reads both files from disk and resolves imports relative to the XSD's directory.
- `FileSchemaResolver(baseDir string) SchemaResolver` -- filesystem-backed resolver for use with `ParseSchemaWithResolver` / `ValidateWithSchemaResolver`.
- `ParseTree(io.Reader) (*Document, error)` -- parse to tree without validating.
- `ParseSchema(*Document) (*Schema, error)` -- parse an XSD tree to a schema model.
- `ParseSchemaWithResolver(*Document, SchemaResolver) (*Schema, error)` -- as above but follows `xs:import` and `xs:include` directives via the resolver.
- `ValidateSchema(*Document, *Schema) error` -- validate a parsed tree against a parsed schema.

Errors come back as `*validator.Error` with `Line`, `Col`, and `Message`.

Tree nodes carry positions too: `Element` and `Attr` both have `Line` and `Col`. A consumer that reports a problem with an attribute's value points at that attribute. It does not point at the element that owns it. On a multi-attribute element those are different places.

## Supported

- XML 1.1 declaration (required)
- Elements, attributes, text, CDATA sections, comments, PIs
- Character references (`&#N;`, `&#xN;`) and predefined entity references (`&amp;` `&lt;` `&gt;` `&apos;` `&quot;`)
- Namespace validation (Namespaces in XML 1.1)
- Two input modes, picked by the encoding declaration. UTF-8 is the default. The 8-bit byte mode (`ISO-8859-1` and its aliases) reads byte b as the character U+00XX, and anything above U+00FF needs a character reference. Neither takes a BOM (per utf8everywhere). See `docs/encodings.md`
- XML 1.1 line ending normalization (`#x85`, `#x2028`)
- XML 1.1 character and name character classes
- Restricted character enforcement (must use character references)
- `&#0;` -- one deliberate deviation from the XML 1.1 `Char` production, which starts at `#x1`. A reference that resolves to U+0000 is accepted and carried through to the parsed value. A literal NUL byte and a lone surrogate are still rejected. See `IsCharRefValue` in `validator/chars.go`, and `docs/nul-char-ref.md`. The reference is not a NUL byte and not a document terminator, proved at both the library and the CLI layer.
- XSD schema validation (`--schema`):
  - Element declarations (global, local, refs)
  - Root matching by namespace AND local name. A root in another namespace, or in none, is a different element from the declared one. It is rejected by name. A shared local name is a coincidence, and taking it for a match validated a document written against another vocabulary
  - Complex types (sequence, choice, all content models)
  - Simple types (restriction with facets, list, union)
  - Attributes (required/optional/prohibited, type checking, fixed values)
  - Global `xs:attribute` declarations and `xs:attribute ref=`, matched by namespace AND local name. A qualified attribute resolves to the global declaration in its own namespace. An imported vocabulary's attributes are therefore type-checked. An undeclared one is an error rather than something a wildcard waves through
  - Named types, groups, attribute groups. An `xs:group ref` is replaced by a copy of the group's content model, carrying the occurrence counts stated at the reference
  - simpleContent and complexContent (extension/restriction). An extension's content follows its base's, and its attributes add to the base's. A restriction states its own content model. It still inherits the base's attributes where it does not restate them
  - 35+ built-in XSD types (string, integer, boolean, decimal, date, etc.)
  - Facets: enumeration, pattern, minLength, maxLength, length, min/maxInclusive, min/maxExclusive, totalDigits, fractionDigits. An attribute value and element text take the same path, so every facet constrains both. A restriction of a restriction enforces the facets it inherits as well as the ones it states. On a list the length facets count items rather than characters. On anything else a length facet counts characters, and octets for `xs:hexBinary` and `xs:base64Binary` -- never the bytes of the UTF-8 encoding. A facet value that is not a non-negative integer is an error, not a length of 0
  - `xs:list` and `xs:union` with an `itemType`/`memberTypes` attribute or with inline `xs:simpleType` members. Each item or member carries its own facets
  - minOccurs/maxOccurs enforcement
  - xs:any wildcard particles
  - xs:anyAttribute wildcard attributes
  - xs:import with optional schemaLocation, loaded via a `SchemaResolver`. The CLI and `ValidateWithSchemaFile` wire a filesystem-backed one
  - xs:include with schemaLocation (same resolver mechanism as xs:import)
  - Identity constraints: `xs:key`, `xs:keyref`, `xs:unique`. See `docs/identity-constraints.md` for the XPath subset, the two deliberate deviations, and how a keyref finds its key
  - Substitution groups, including `abstract` heads, `block="substitution"` / `blockDefault`, and transitive members. See `docs/substitution-groups.md`
  - `xs:alternative` conditional type assignment over the XPath subset XSD 1.1 requires: comparisons, `and`/`or` with parentheses, `not()`, casts, and constructor functions. See `docs/conditional-types.md`

## Hard Errors (Unsupported)

- `xs:all` anywhere but as a whole complex type's content model, and anything but element declarations and `xs:any` inside one. Order-free matching is defined over a whole element, so a nested one matched nothing at all
- An `xs:element`/`xs:attribute`/`xs:group` `ref`, or a `complexContent` base, that names nothing. An unresolvable reference used to leave a hole in the content model, and to report itself as a missing instance element
- `processContents="skip"` and `processContents="lax"` on `xs:any` / `xs:anyAttribute`. Only `strict` (the default) is allowed. This validator does not offer a no- or partial-validation mode
- DOCTYPE declarations
- General entity references (beyond the 5 predefined)
- XML 1.0 documents
- Missing XML declaration
- Encodings other than UTF-8 and ISO-8859-1. UTF-16 inputs and BOMs are rejected, and any other encoding declaration is rejected by name
- xs:redefine, xs:override
- xs:notation
- Any unrecognized child of an `xs:element`, and any `xs:alternative` test outside the subset XSD requires. Any selector or field XPath outside the supported subset is an error too: a constraint either runs or says it cannot
- An `xs:list` with no item type and an `xs:union` with no members -- both used to accept every value
- A simple type that derives from itself

## Project Structure

- `reader/chars.go` -- XML 1.1 character class predicates
- `reader/reader.go` -- `Decode`: decoder selection, BOM/UTF-16 reject, line normalization
- `reader/encoding.go` -- the two modes, the alias table, declaration sniffing, both decoders
- `reader/entities.go` -- the five predefined entities, matched without allocating
- `reader/error.go` -- error type with line/column position
- `reader/document.go` -- document tree model (Element, Attr, CharData)
- `reader/tree.go` -- version-agnostic XML tree parser
- `writer/writer.go` -- emits a tree as XML, base64 by default for a byte payload
- `cli/cmd/root.go` -- CLI (cobra). `cli/cmd/xml-validator/main.go` is the binary
- `cli/dats/*.dats` -- CLI suites, run after the cli module's build
- `validator/reader_types.go` -- aliases that keep the reader's types spelled `validator.Document`
- `validator/doc.go` -- package-level godoc
- `validator/validator.go` -- public `Validate`, `ValidateWithSchema`, and `ValidateWithSchemaBytes` entry points
- `validator/example_test.go` -- runnable godoc examples
- `validator/parser.go` -- recursive descent parser core, XML declaration, comments, PIs
- `validator/elements.go` -- element, attribute, content, CDATA, reference parsing
- `validator/namespace.go` -- QName validation, namespace scope management
- `validator/schema_model.go` -- XSD schema model types
- `validator/schema_builtin.go` -- 35+ built-in XSD types with validation
- `validator/schema_facets.go` -- facet checking (lengths, patterns, ranges, digits)
- `validator/schema_parse.go` -- XSD file parser (document tree to schema model)
- `validator/schema_parse_particles.go` -- the content-model particles: sequence, choice, all, group ref, and the wildcard
- `validator/schema_parse_simple.go` -- simple types, facets, list, union
- `validator/schema_identity.go` -- xs:key, xs:keyref, xs:unique, and the XPath subset they select with
- `validator/schema_import.go` -- `SchemaResolver` and `xs:import` handling
- `validator/schema_qname.go` -- namespace-keyed lookup of global declarations
- `validator/schema_resolve.go` -- ref and type resolution over a parsed schema
- `validator/schema_derive.go` -- group-ref expansion and complexContent derivation
- `validator/schema_substitution.go` -- substitution groups: members, abstract heads, block, derivation check
- `validator/schema_alternative.go` -- xs:alternative: parse, resolve, per-instance type choice
- `validator/schema_alternative_expr.go` -- the @test language: XSD 1.1's required XPath subset
- `validator/schema_value.go` -- simple-value validation (facets, list, union) shared by element text and attribute values
- `validator/schema_validate.go` -- schema validation engine
- `validator/roundtrip_nul_test.go` -- parse/emit/reparse roundtrips for `&#0;` and for a payload carrying every byte value 0..255
- `validator/roundtrip_binary_test.go` -- the same payload through `xs:base64Binary` and `xs:hexBinary`, with the size of each wire form
- `cli/dats/nul-char-ref.dats` -- CLI suite for `&#0;`, run after every build
- `docs/nul-char-ref.md` -- what `&#0;` is, what it is not, and how to check
- `action.yml` -- composite GitHub Action. It installs the released binary via `buildhost-download` and checks it runs (`--help`). Only when that fails does it fall back to a cached source build, and it warns as it does. Zero-config discovery covers `*.xml` and `*.xsd`, and the validator runs once per file. Every step that carries logic runs as TypeScript (tsc-checked, Node) via `wow-look-at-my/actions@typescript#latest` instead of inline shell, so the logic stays portable across runner OSes.
