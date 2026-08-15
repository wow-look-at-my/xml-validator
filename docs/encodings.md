# The two input modes: UTF-8 and byte mode

This validator reads a document in one of two modes. The encoding declaration
picks between them, and nothing else does.

| declared encoding | mode | a byte is |
|---|---|---|
| absent, `UTF-8`, `UTF8` | UTF-8 | part of a UTF-8 sequence |
| `ISO-8859-1` and its aliases | byte | one character, U+0000 to U+00FF |

The aliases are the ones IANA registers for that character set:
`ISO8859-1`, `ISO_8859-1`, `latin1`, `latin-1`, `l1`, `IBM819`, `cp819`,
`csISOLatin1`. The comparison ignores case. Any other name is an error that
names both modes -- this validator does not read Shift_JIS, windows-1252, or
UTF-16, and says so rather than guessing.

## Byte mode

Byte b is the character U+00XX with the same value. A document is exactly as
many bytes as it has characters, and every byte decodes, so byte mode has no
equivalent of `invalid UTF-8 byte sequence`.

What byte mode does not have is a byte for anything above U+00FF. Those take a
character reference, which is ASCII and works in either mode:

```xml
<?xml version="1.1" encoding="ISO-8859-1"?>
<r>caf<E9> and a snowman &#9731;</r>
```

(`<E9>` is the single byte 0xE9, the character `é`.)

Byte mode changes how bytes become characters. It changes nothing about which
characters a document may hold. A literal NUL is rejected, a literal
restricted character is rejected, and `&#0;` still carries U+0000 through --
all exactly as in UTF-8 (see `nul-char-ref.md`).

Line-ending normalization also still applies, and in byte mode it has more to
do: 0x85 is a byte here, so a document that means NEL rather than a line break
writes `&#133;`.

## What it costs and what it saves

Measured, not estimated: 499 samples over 5.7 MiB, and every one of the 2894
documents below was checked with the built binary.
`tools/fetch-corpus.js` builds the corpus (Wikipedia prose in ten scripts,
Rosetta Code in eight languages, PNG and JPEG files), and
`tools/encoding-sizes.js --validate` produces the table:

| corpus | files | median | byte mode (text) | byte mode (bytes) | UTF-8 | base64 | hex | all refs |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| code-c | 25 | 1546 B | 1.08x | 1.09x | 1.08x | 1.36x | 2.02x | 5.40x |
| code-go | 25 | 1873 B | 1.04x | 1.05x | 1.04x | 1.35x | 2.02x | 5.40x |
| code-haskell | 25 | 1494 B | 1.07x | 1.07x | 1.06x | 1.36x | 2.02x | 5.42x |
| code-java | 25 | 2782 B | 1.03x | 1.04x | 1.03x | 1.35x | 2.01x | 5.42x |
| code-javascript | 25 | 1517 B | 1.06x | 1.06x | 1.05x | 1.36x | 2.02x | 5.41x |
| code-python | 25 | 1441 B | 1.04x | 1.05x | 1.04x | 1.36x | 2.02x | 5.45x |
| code-ruby | 25 | 864 B | 1.06x | 1.07x | 1.04x | 1.37x | 2.04x | 5.47x |
| code-rust | 25 | 1779 B | 1.06x | 1.07x | 1.06x | 1.35x | 2.02x | 5.43x |
| image-jpeg | 25 | 7155 B | n/a | 2.18x | n/a | 1.34x | 2.00x | 5.53x |
| image-png | 25 | 9625 B | n/a | 2.18x | n/a | 1.34x | 2.00x | 5.52x |
| prose-arabic | 24 | 20876 B | 3.21x | 2.06x | 1.00x | 1.34x | 2.00x | 5.89x |
| prose-chinese | 25 | 22378 B | 2.56x | 2.68x | 1.00x | 1.33x | 2.00x | 5.97x |
| prose-english | 25 | 13817 B | 1.01x | 1.01x | 1.00x | 1.34x | 2.00x | 5.66x |
| prose-french | 25 | 16398 B | 0.98x | 1.02x | 1.00x | 1.34x | 2.00x | 5.68x |
| prose-german | 25 | 9353 B | 1.07x | 1.19x | 1.00x | 1.34x | 2.00x | 5.46x |
| prose-greek | 25 | 23707 B | 2.76x | 1.80x | 1.00x | 1.33x | 2.00x | 5.89x |
| prose-hindi | 25 | 22498 B | 2.21x | 1.61x | 1.00x | 1.33x | 2.00x | 5.91x |
| prose-japanese | 25 | 21941 B | 2.60x | 3.09x | 1.00x | 1.33x | 2.00x | 5.97x |
| prose-russian | 25 | 33445 B | 3.23x | 1.70x | 1.00x | 1.33x | 2.00x | 5.91x |
| prose-vietnamese | 25 | 19370 B | 1.37x | 1.28x | 1.00x | 1.34x | 2.00x | 5.70x |

Ratios are document bytes per payload byte. The two byte-mode columns are the
two questions you can ask of the mode: **text** decodes the payload and writes
each character as a byte where Latin-1 has one, and **bytes** writes the
payload's bytes as they are.

What the numbers say:

- **Byte mode wins on Latin-1 text, and only there.** French prose comes out
  at 0.98x -- smaller than the UTF-8 payload itself, because each accented
  character is two bytes there and one here. German is 1.07x, English 1.01x.
- **It loses badly outside Latin-1.** Russian 3.23x, Greek 2.76x, Japanese
  2.60x: every character has to be a reference, which costs 6 to 8 bytes where
  UTF-8 spends 2 or 3.
- **UTF-8 is 1.00x on anything already UTF-8**, and cannot carry anything
  else. Every image row is `n/a` for it.
- **base64 is 1.33x to 1.37x on everything**, flat, which is what makes it the
  answer for images and any other arbitrary bytes.
- **Escaping every character is 5.4x to 6.0x.** It is the wire form that
  survives anything, and it costs what that is worth.

The declaration itself costs 22 bytes more than the UTF-8 one, so a short
document with a few high characters is still smaller in UTF-8; a document pays
that back at its 22nd such character.

## Where it lives

- `validator/encoding.go` -- the mode constants, the alias table,
  `sniffEncoding`, and the two decoders.
- `validator/reader.go` -- picks a decoder, then normalizes line endings.
- `validator/parser.go` -- `validateEncodingMatch` rejects a name that selects
  neither mode.
- `validator/encoding_test.go` and `dats/encodings.dats` -- the tests, at the
  library and at the built CLI.

`sniffEncoding` reads the declaration out of the raw bytes, before anything
decodes them, because the declaration decides how to decode. That works
because a declaration is ASCII in both modes: XML 1.1 requires every encoding
it admits to agree with ASCII on the characters a declaration can use. A
declaration it cannot parse reads as UTF-8, which leaves the syntax error to
the declaration parser, where it is reported with a line and column.

A byte-order mark is still rejected in both modes. UTF-8 does not need one
(per utf8everywhere) and byte mode has no character above U+00FF to spell one
with.
