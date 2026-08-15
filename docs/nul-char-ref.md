# `&#0;` is a character reference, not a NUL byte, not a terminator

Three different things get called "null" in the same sentence, and they are not
the same thing:

1. **`&#0;`** -- four ASCII bytes in the document's byte stream. A document
   that uses it contains no byte with the value zero.
2. **U+0000** -- the character the reference resolves to. It is the parsed
   value's first character, or its middle one, like any other character.
3. **The end of a document** -- which XML marks with the root element's end
   tag, never with a character.

This validator accepts (1), produces (2), and treats neither as (3).

## The deviation, stated plainly

The XML 1.1 `Char` production starts at `#x1`, so a strict reading rejects a
reference that resolves to U+0000. This validator accepts it. `IsCharRefValue`
in `validator/chars.go` is `IsChar` plus U+0000, and it is used only where a
character reference resolves -- content and attribute values. A literal NUL
byte is still rejected everywhere: character data, attribute values, CDATA
sections and comments. So the reference is the only way U+0000 enters a parsed
value, which is what keeps the byte stream free of NUL bytes by construction.

A lone surrogate stays rejected too. It is not a character in any encoding, so
there is nothing for a reference to resolve to.

## What the tests prove, and where

Two layers, because the disbelief runs at two levels: what the library does
with the value, and what the program a user runs does with the file.

### Library roundtrip -- `validator/roundtrip_nul_test.go`

The parser turns `&#0;` into U+0000. The test file's `serializeDoc` turns
U+0000 back into `&#0;`. A roundtrip is parse, emit, parse again, and the two
trees must match:

- The reparsed tree equals the first one, and a second roundtrip changes no
  bytes.
- `&#0;`, `&#00;`, `&#x0;`, `&#x00;` and `&#0000000;` all parse to the same
  character and all emit as the one canonical `&#0;`.
- The emitted document contains no NUL byte.
- Content after the reference -- sibling elements, their attributes, the
  closing tags -- survives the roundtrip.
- The parsed value is an ordinary Go string: `len` counts the NUL, indexing
  reaches it, `strings.Split` splits on it.
- Inside CDATA the same four characters are text. No reference resolves, no
  U+0000 is produced, and the emitter escapes the ampersand to keep it that
  way.
- A literal NUL byte is rejected in character data, attribute values, CDATA
  and comments.
- An `xs:length` facet of 3 accepts `a&#0;b`, and a facet of 1 rejects it with
  "value length 3". A terminator would leave a value of length 1.

Three more take a real binary payload -- one of every byte value, 0 through
255, carried as the characters U+0000 through U+00FF:

- Written with a reference only where one is needed, the payload survives the
  roundtrip through both character data and an attribute value, and decodes
  back to the same 256 bytes.
- Written as nothing but references, the document is printable ASCII end to
  end, which is a wire form that survives a transport with opinions about high
  bytes and NUL.
- A payload with three NUL bytes in the middle keeps its tail.

Encoding the payload takes more than the restricted characters. CR, NEL
(U+0085) and LINE SEPARATOR (U+2028) all normalize to LF when they appear as
literal characters, and a tab or newline inside an attribute value is
whitespace a conforming reader folds to a space. Normalization runs over the
input before any reference resolves, so writing those as references is what
carries them through.

`serializeDoc` is a test helper, not library API. It emits no namespace
declarations, so every document in that file declares none.

### Executable roundtrip -- `dats/nul-char-ref.dats`

Ten tests drive the built binary, so nothing here depends on a Go caller.
`go-toolchain` runs them after every build; there is no opt-out. The two that
answer the terminator claim directly:

- **The file holds no NUL byte.** `wc -c` of the fixture and `wc -c` after
  `tr -d '\000'` are both 36. Nothing was deleted, because there was nothing
  to delete.
- **Parsing continues past the reference.** A fixture puts `&#0;` on line 2 and
  a second root element on line 3. A reader that stopped at the NUL would never
  see line 3 and would report the file valid. The validator exits 1 and names
  line 3.

One more runs the whole binary roundtrip through the shell, so nothing in it
depends on Go at all: 256 raw bytes are written with `printf`, `od` turns each
one into a character reference, the validator accepts the document, `grep` and
`printf` turn the references back into bytes, and the two SHA-256 digests must
match. The expected digest is pinned in the suite --
`40aff2e9...bf944880` is the SHA-256 of the bytes 0x00..0xFF in order -- so the
comparison does not rest on the generator alone. The encoded document is 1454
bytes, all of them printable ASCII.

The rest cover the spellings, stdin, the well-formed tail, the literal NUL
byte in character data and in CDATA, and `xs:length` at 3, at 1 and at 256.

## The three wire forms, and why the high half is not escaped

The minimally-escaped form leaves U+0080 through U+00FF literal. They are two
bytes each because the document is UTF-8, which is the only encoding this
validator accepts: a raw Latin-1 byte fails with `invalid UTF-8 byte
sequence`, and an `encoding="ISO-8859-1"` declaration does not change that.
What still takes a reference above U+007F is U+007F to U+0084 and U+0086 to
U+009F, which are restricted characters, and U+0085, which normalizes to LF.
Neither rule is about the byte being high.

For a payload that is genuinely bytes, XSD has `xs:base64Binary` and
`xs:hexBinary`. Both are ASCII on the wire and both measure length in octets.
The same 256-byte payload, all four ways:

| form | document size | content |
|---|---|---|
| every character a reference | 1454 | printable ASCII |
| reference only where needed | 666 | ASCII plus UTF-8 for the high half |
| `xs:hexBinary` | 546 | printable ASCII, 512 digits |
| `xs:base64Binary` | 378 | printable ASCII, 344 characters |

`validator/roundtrip_binary_test.go` roundtrips the payload through both
binary types and pins those sizes. The dats suite runs the same two through
the shell -- `base64 -w0` and `od -tx1` on the way in, `base64 -d` and
`printf` on the way out -- and compares SHA-256 digests, with the schema
stating a length of 256 octets in each case.

## Facet lengths are characters, not bytes

Measuring the 256-character payload turned up a real defect: the `length`,
`minLength` and `maxLength` facets counted the bytes of the UTF-8 encoding, so
the payload reported as 384. XSD defines the unit as characters for a string
type and octets for `xs:hexBinary` and `xs:base64Binary`.
`validator/schema_facets.go` now counts each of those, and a facet value that
is not a non-negative integer is an error in the schema rather than a silent
length of 0. `validator/schema_facet_length_test.go` covers all four.

## Reproducing it by hand

```bash
go-toolchain                      # builds build/xml-validator, runs both layers

printf '<?xml version="1.1"?>\n<r>a&#0;b</r>\n' > /tmp/nul.xml
build/xml-validator /tmp/nul.xml  # valid XML 1.1 document
wc -c < /tmp/nul.xml              # 36
tr -d '\000' < /tmp/nul.xml | wc -c   # 36 -- no NUL byte was there to delete

printf '<?xml version="1.1"?><r>a\000b</r>' > /tmp/real-nul.bin
build/xml-validator /tmp/real-nul.bin  # error: invalid character U+0000
```

To run the CLI suite on its own, stage the binary where the suite expects it:

```bash
mkdir -p build/.dats-stage && cp build/xml-validator build/.dats-stage/
GO_TOOLCHAIN_DATS_BUILD_DIR="$PWD/build/.dats-stage" dats test dats/
```

The suite is sandboxed (bubblewrap, falling back to docker). A host with
neither fails the run rather than running the commands unsandboxed.
