# xml-validator

Strict XML 1.1 validator. Anything unsupported is a hard error.

## Build & Test

```bash
go-toolchain
```

This handles `go mod tidy`, `go vet`, tests with coverage, and builds the binary to `build/xml-validator`.

## Usage

```bash
xml-validator <file>    # validate a file
xml-validator           # validate from stdin
```

Exit 0 on valid XML 1.1, exit 1 with error message on failure.

## Supported

- XML 1.1 declaration (required)
- Elements, attributes, text, CDATA sections, comments, PIs
- Character references (`&#N;`, `&#xN;`) and predefined entity references (`&amp;` `&lt;` `&gt;` `&apos;` `&quot;`)
- Namespace validation (Namespaces in XML 1.1)
- UTF-8 and UTF-16 (BE/LE) with BOM detection
- XML 1.1 line ending normalization (`#x85`, `#x2028`)
- XML 1.1 character and name character classes
- Restricted character enforcement (must use character references)

## Hard Errors (Unsupported)

- DOCTYPE declarations
- General entity references (beyond the 5 predefined)
- XML 1.0 documents
- Missing XML declaration
- Encodings other than UTF-8/UTF-16

## Project Structure

- `cmd/root.go` -- CLI (cobra)
- `validator/validator.go` -- public `Validate(io.Reader) error` entry point
- `validator/parser.go` -- recursive descent parser core, XML declaration, comments, PIs
- `validator/elements.go` -- element, attribute, content, CDATA, reference parsing
- `validator/namespace.go` -- QName validation, namespace scope management
- `validator/chars.go` -- XML 1.1 character class predicates
- `validator/reader.go` -- encoding detection, UTF-8/UTF-16 decoding, line normalization
- `validator/error.go` -- error type with line/column position
