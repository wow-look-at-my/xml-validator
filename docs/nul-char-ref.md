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

The rest cover the spellings, stdin, the well-formed tail, the literal NUL
byte in character data and in CDATA, and the same `xs:length` pair as above.

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
