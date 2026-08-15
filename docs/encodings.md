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

**It is one byte per CHARACTER, not per payload byte.** Byte mode is not a
byte-transparent transport, and XML does not have one: of the 256 values, 191
may appear literally and cost one byte each, while 65 may not appear literally
in any XML document in any encoding -- U+0000, the restricted C0 and C1
controls, and CR and NEL, which line-ending normalization would rewrite.
Those cost a reference, 4 to 6 bytes each.

So a payload of representable characters really is 1.00x, and a payload of
arbitrary bytes is about 2.1x, because a quarter of it is characters XML will
not carry literally. That is what `xs:base64Binary` and `xs:hexBinary` are
for, and why the table below measures byte mode twice.

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

Measured, not estimated: 350 samples over 14.8 MiB, drawn from the corpus
submodules, and every one of the 2000 documents below was checked with the
built binary. `tools/extract-corpus.js` pulls the samples out and
`tools/encoding-sizes.js --corpus corpus/samples --validate` produces this:

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
| image-jpeg | 25 | 319737 B | n/a | 2.14x | n/a | 1.33x | 2.00x | 5.55x |
| image-png | 25 | 247688 B | n/a | 2.15x | n/a | 1.33x | 2.00x | 5.55x |
| prose-icelandic | 25 | 8687 B | **0.92x** | 1.04x | 1.00x | 1.34x | 2.00x | 5.72x |
| prose-amharic | 25 | 11171 B | 2.20x | 3.16x | 1.00x | 1.34x | 2.00x | 5.90x |
| prose-nepali | 25 | 15263 B | 2.24x | 1.58x | 1.00x | 1.34x | 2.00x | 5.93x |
| prose-yiddish | 25 | 6097 B | 3.20x | 2.42x | 1.01x | 1.34x | 2.01x | 5.89x |

Ratios are document bytes per payload byte. The two byte-mode columns are the
two questions you can ask of the mode: **text** decodes the payload and writes
each character as a byte where Latin-1 has one, and **bytes** writes the
payload's bytes as they are.

What the numbers say:

- **Byte mode wins on Latin-1 text, and only there.** Icelandic prose comes
  out at 0.92x -- smaller than the UTF-8 payload itself, because every
  accented character is two bytes there and one here.
- **It loses badly outside Latin-1.** Yiddish 3.20x, Nepali 2.24x, Amharic
  2.20x: every character has to be a reference, which costs 6 to 8 bytes where
  UTF-8 spends 2 or 3.
- **Code is 1.03x to 1.09x in every mode**, because it is mostly ASCII. The
  mode barely matters there.
- **UTF-8 is 1.00x on anything already UTF-8**, and cannot carry anything
  else. Every image row is `n/a` for it.
- **base64 is 1.33x to 1.37x on everything**, flat, which is what makes it the
  answer for images and any other arbitrary bytes.
- **Escaping every character is 5.4x to 6.0x.** It is the wire form that
  survives anything, and it costs what that is worth.

The declaration itself costs 22 bytes more than the UTF-8 one, so a short
document with a few high characters is still smaller in UTF-8; a document pays
that back at its 22nd such character.

## What each form costs to read and write

Size is half the choice. `validator/encoding_bench_test.go` measures the other
half over a 64 KiB payload, and `go-toolchain` runs it in the benchmark phase.
Throughput is per payload byte, so the columns compare directly:

| form | validate (binary) | validate (text, 0.2% NUL) | encode | parse + recover |
|---|---:|---:|---:|---:|
| base64Binary | **54.9 MB/s** | 60.9 MB/s | **450 MB/s** | **7.8 MB/s** |
| hexBinary | 35.8 MB/s | 35.9 MB/s | 289 MB/s | 4.8 MB/s |
| byte mode | 20.8 MB/s | **120.3 MB/s** | 21.1 MB/s | 7.6 MB/s |
| UTF-8 | 17.2 MB/s | 77.5 MB/s | 23.0 MB/s | 7.2 MB/s |
| all references | 5.9 MB/s | 5.4 MB/s | 9.5 MB/s | 5.2 MB/s |

base64 is not fast; references are slow. `BenchmarkValidateNoReferences`
separates the two by handing every form a payload that byte mode can write
without a single reference:

| form | throughput | allocations | size |
|---|---:|---:|---:|
| byte mode | **117.4 MB/s** | 33 | 1.00x |
| base64Binary | 54.2 MB/s | 29 | 1.33x |
| UTF-8 | 48.8 MB/s | 32 | 2.00x |
| hexBinary | 35.1 MB/s | 32 | 2.00x |
| all references | 5.7 MB/s | 37 | 5.6x |

Byte mode reads the binary payload at 20.8 MB/s and this one at 117.4 MB/s, on
the same parser and the same payload size. What changed is the number of
references, and nothing else.

The allocation column used to tell the same story louder -- 42,537 for the
binary payload against 35 for base64 -- because a reference built a rune slice
and a string before parsing its digits, and a predefined entity built its name
as a string before comparing it. Neither does now: the digits go in a stack
buffer, an entity name is matched where it sits, and the line-ending
normalizer rewrites its input instead of copying it. Every column above is
flat at a few dozen allocations per document, whatever the document holds, and
`validator/alloc_test.go` fails if that stops being true.

What is left is the work itself: a literal character is a pointer bump and a
range check, a reference is a scan, a digit fold and a validity check.

With references out of the way, what remains is document size: throughput per
payload byte tracks the expansion ratio, byte mode at 1.00x reading
117.4 MB/s against UTF-8 at 2.00x reading 48.8 MB/s. The parser walks at a
near-constant rate per DOCUMENT byte, so a form that doubles the document
roughly halves the throughput.

base64 wins on binary for both reasons at once: it removes every reference,
and it is the smallest form that does. Give it a payload that needed no
references anyway and it loses to byte mode by 2.2x, because being 33% larger
is all it has left.

The two cases separate cleanly:

- **Arbitrary bytes: base64Binary wins on every axis.** 1.33x against byte
  mode's 2.15x, 3.6x faster to validate, 19x faster to encode.
- **Text with occasional NULs: escaping wins on every axis.** At one NUL per
  512 bytes, byte mode validates at 120 MB/s against base64's 61, and costs
  1.03x against 1.33x. Base64 there is bigger, slower, and opaque to every
  tool that reads text.

Escaping every character is the worst of both, and its 5.4 MB/s is what a
document that survives a byte-mangling transport costs.

These are this validator's numbers, not universal ones: the parser is
recursive descent over runes. The ordering should
hold for any parser that resolves references one at a time.

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
