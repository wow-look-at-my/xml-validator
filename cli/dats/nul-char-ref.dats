# `&#0;` is a character reference, not a NUL byte and not an end of document.
# These tests drive the built CLI, so what they prove holds for the program a
# user runs, not only for the library.
#
# $GO_TOOLCHAIN_DATS_BUILD_DIR holds copies of the binaries this build just
# made. It is read-only inside the sandbox, which is enough to exec them.
# see docs/nul-char-ref.md

setup:
	- test -x "$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator"

tests:
	- desc: a document with a NUL character reference is valid
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" {inputs.nul.xml}'
	  inputs:
		files:
			nul.xml: |
				<?xml version="1.1"?>
				<r>a&#0;b</r>
	  outputs:
		stdout:
			- "valid XML 1.1 document"

	- desc: every spelling of the reference is valid
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" {inputs.spellings.xml}'
	  inputs:
		files:
			spellings.xml: |
				<?xml version="1.1"?>
				<r>&#0;&#00;&#x0;&#x00;&#0000000;</r>
	  outputs:
		stdout:
			- "valid XML 1.1 document"

	# The file that just validated holds no NUL byte at all: deleting every
	# NUL byte leaves the byte count unchanged, and the reference is four
	# ASCII bytes of the file's 36.
	- desc: the validated file contains no NUL byte
	  cmd: 'printf "%s %s\n" "$(wc -c < {inputs.nul.xml})" "$(tr -d "\000" < {inputs.nul.xml} | wc -c)"'
	  inputs:
		files:
			nul.xml: |
				<?xml version="1.1"?>
				<r>a&#0;b</r>
	  outputs:
		stdout:
			0: "^36 36$"

	# A reader that stopped at the NUL would never reach line 3, so it would
	# report this file as valid. The error names line 3.
	- desc: parsing continues past the reference to a later error
	  exit: 1
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" {inputs.late-error.xml}'
	  inputs:
		files:
			late-error.xml: |
				<?xml version="1.1"?>
				<r>a&#0;b</r>
				<second-root/>
	  outputs:
		stderr:
			- "line 3"
			- "unexpected content after root element"
		"!stdout":
			- "valid XML 1.1 document"

	# The same document with the tail made well-formed: the reference stopped
	# nothing, so the whole file validates.
	- desc: content after the reference is ordinary content
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" {inputs.tail.xml}'
	  inputs:
		files:
			tail.xml: |
				<?xml version="1.1"?>
				<r>
					<a>1&#0;2</a>
					<b at="&#0;">after</b>
					<c/>
				</r>
	  outputs:
		stdout:
			- "valid XML 1.1 document"

	# printf writes a real NUL byte, which no reference can produce. A tool
	# that treated `&#0;` and a NUL byte as the same thing would accept both.
	- desc: a literal NUL byte is rejected
	  exit: 1
	  cmd: 'printf %b "<?xml version=\"1.1\"?><r>a\000b</r>" > {outputs.nul.bin}; "$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" {outputs.nul.bin}'
	  outputs:
		stderr:
			- "invalid character U+0000 in character data"
		"!stdout":
			- "valid XML 1.1 document"

	- desc: a literal NUL byte is rejected inside CDATA
	  exit: 1
	  cmd: 'printf %b "<?xml version=\"1.1\"?><r><![CDATA[a\000b]]></r>" > {outputs.cdata.bin}; "$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" {outputs.cdata.bin}'
	  outputs:
		stderr:
			- "invalid character U+0000 in CDATA section"

	- desc: the reference works on stdin too
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator"'
	  inputs:
		stdin: |
			<?xml version="1.1"?>
			<r>a&#0;b</r>
	  outputs:
		stdout:
			- "valid XML 1.1 document"

	# Schema validation counts the value as three characters: a, U+0000, b. A
	# terminator would leave a value of length 1.
	- desc: the schema length facet counts the NUL as one character
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema {inputs.len3.xsd} {inputs.doc.xml}'
	  inputs:
		files:
			doc.xml: |
				<?xml version="1.1"?>
				<r>a&#0;b</r>
			len3.xsd: |
				<?xml version="1.1"?>
				<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
					<xs:element name="r">
						<xs:simpleType>
							<xs:restriction base="xs:string">
								<xs:length value="3"/>
							</xs:restriction>
						</xs:simpleType>
					</xs:element>
				</xs:schema>
	  outputs:
		stdout:
			- "valid XML 1.1 document (schema validated)"

	# The full binary roundtrip, at the executable: 256 raw bytes -- one of
	# every value, U+0000 included -- become an XML document, the validator
	# accepts it, and decoding the references gives the same 256 bytes back.
	# The pinned digest is the SHA-256 of the bytes 0x00..0xFF in order, so
	# the comparison does not rest on the generator alone.
	- desc: a 256-byte binary payload roundtrips through XML and back
	  cmd: |
		set -e
		w="$(mktemp -d)"
		: > "$w/orig.bin"
		for i in $(seq 0 255); do printf "\\$(printf '%03o' "$i")" >> "$w/orig.bin"; done
		{ printf '<?xml version="1.1"?><r>'
		  od -An -v -tu1 "$w/orig.bin" | tr -s ' ' '\n' | grep . | while read -r n; do printf '&#%s;' "$n"; done
		  printf '</r>'; } > "$w/enc.xml"
		"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" "$w/enc.xml"
		grep -o '&#[0-9]*;' "$w/enc.xml" | tr -cd '0-9\n' | while read -r n; do printf "\\$(printf '%03o' "$n")"; done > "$w/dec.bin"
		echo "sizes: orig $(wc -c < "$w/orig.bin") xml $(wc -c < "$w/enc.xml") decoded $(wc -c < "$w/dec.bin")"
		echo "bytes outside printable ASCII in the xml: $(LC_ALL=C tr -d '\040-\176' < "$w/enc.xml" | wc -c)"
		echo "orig   $(sha256sum < "$w/orig.bin")"
		echo "decode $(sha256sum < "$w/dec.bin")"
		test "$(sha256sum < "$w/orig.bin")" = "$(sha256sum < "$w/dec.bin")"
		echo roundtrip identical
	  outputs:
		stdout:
			- "valid XML 1.1 document"
			- "sizes: orig 256 xml 1454 decoded 256"
			- "bytes outside printable ASCII in the xml: 0"
			- "orig   40aff2e9d2d8922e47afd4648e6967497158785fbd1da870e7110266bf944880"
			- "decode 40aff2e9d2d8922e47afd4648e6967497158785fbd1da870e7110266bf944880"
			- "roundtrip identical"

	# XSD's own answer for arbitrary bytes. The payload rides an
	# xs:base64Binary element, the schema states its length in octets, and
	# base64 -d gives the same 256 bytes back.
	- desc: a 256-byte payload roundtrips through xs:base64Binary
	  cmd: |
		set -e
		w="$(mktemp -d)"
		: > "$w/orig.bin"
		for i in $(seq 0 255); do printf "\\$(printf '%03o' "$i")" >> "$w/orig.bin"; done
		{ printf '<?xml version="1.1"?><blob>'; base64 -w0 < "$w/orig.bin"; printf '</blob>'; } > "$w/b64.xml"
		"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema {inputs.b64-256.xsd} "$w/b64.xml"
		sed -e 's|.*<blob>||' -e 's|</blob>.*||' "$w/b64.xml" | base64 -d > "$w/dec.bin"
		echo "xml $(wc -c < "$w/b64.xml") bytes, decoded $(wc -c < "$w/dec.bin") bytes"
		test "$(sha256sum < "$w/orig.bin")" = "$(sha256sum < "$w/dec.bin")"
		echo base64 roundtrip identical
	  inputs:
		files:
			b64-256.xsd: |
				<?xml version="1.1"?>
				<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
					<xs:element name="blob">
						<xs:simpleType>
							<xs:restriction base="xs:base64Binary">
								<xs:length value="256"/>
							</xs:restriction>
						</xs:simpleType>
					</xs:element>
				</xs:schema>
	  outputs:
		stdout:
			- "valid XML 1.1 document (schema validated)"
			- "xml 378 bytes, decoded 256 bytes"
			- "base64 roundtrip identical"

	# The same payload as hexBinary: 512 digits, still 256 octets to the
	# length facet.
	- desc: a 256-byte payload roundtrips through xs:hexBinary
	  cmd: |
		set -e
		w="$(mktemp -d)"
		: > "$w/orig.bin"
		for i in $(seq 0 255); do printf "\\$(printf '%03o' "$i")" >> "$w/orig.bin"; done
		{ printf '<?xml version="1.1"?><blob>'; od -An -v -tx1 "$w/orig.bin" | tr -d ' \n' | tr 'a-f' 'A-F'; printf '</blob>'; } > "$w/hex.xml"
		"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema {inputs.hex-256.xsd} "$w/hex.xml"
		sed -e 's|.*<blob>||' -e 's|</blob>.*||' "$w/hex.xml" | tr -d '\n' | sed 's/../&\n/g' | grep . | while read -r h; do printf "\\$(printf '%03o' "$((16#$h))")"; done > "$w/dec.bin"
		echo "xml $(wc -c < "$w/hex.xml") bytes, decoded $(wc -c < "$w/dec.bin") bytes"
		test "$(sha256sum < "$w/orig.bin")" = "$(sha256sum < "$w/dec.bin")"
		echo hex roundtrip identical
	  inputs:
		files:
			hex-256.xsd: |
				<?xml version="1.1"?>
				<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
					<xs:element name="blob">
						<xs:simpleType>
							<xs:restriction base="xs:hexBinary">
								<xs:length value="256"/>
							</xs:restriction>
						</xs:simpleType>
					</xs:element>
				</xs:schema>
	  outputs:
		stdout:
			- "valid XML 1.1 document (schema validated)"
			- "xml 546 bytes, decoded 256 bytes"
			- "hex roundtrip identical"

	# A raw high byte is not a document: UTF-8 is the only encoding, so a
	# Latin-1 payload written one byte per character is rejected outright.
	- desc: a raw Latin-1 byte is not valid UTF-8
	  exit: 1
	  cmd: 'printf %b "<?xml version=\"1.1\"?><r>\351</r>" > {outputs.latin1.xml}; "$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" {outputs.latin1.xml}'
	  outputs:
		stderr:
			- "invalid UTF-8 byte sequence"

	# The same 256 bytes, counted by the schema engine: 256 characters, not a
	# 1-character value that stops at the first one.
	- desc: the schema counts the binary payload as 256 characters
	  cmd: |
		set -e
		w="$(mktemp -d)"
		{ printf '<?xml version="1.1"?><r>'
		  for i in $(seq 0 255); do printf '&#%d;' "$i"; done
		  printf '</r>'; } > "$w/enc.xml"
		"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema {inputs.len256.xsd} "$w/enc.xml"
	  inputs:
		files:
			len256.xsd: |
				<?xml version="1.1"?>
				<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
					<xs:element name="r">
						<xs:simpleType>
							<xs:restriction base="xs:string">
								<xs:length value="256"/>
							</xs:restriction>
						</xs:simpleType>
					</xs:element>
				</xs:schema>
	  outputs:
		stdout:
			- "valid XML 1.1 document (schema validated)"

	- desc: the same value fails a length-1 facet, naming length 3
	  exit: 1
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema {inputs.len1.xsd} {inputs.doc.xml}'
	  inputs:
		files:
			doc.xml: |
				<?xml version="1.1"?>
				<r>a&#0;b</r>
			len1.xsd: |
				<?xml version="1.1"?>
				<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
					<xs:element name="r">
						<xs:simpleType>
							<xs:restriction base="xs:string">
								<xs:length value="1"/>
							</xs:restriction>
						</xs:simpleType>
					</xs:element>
				</xs:schema>
	  outputs:
		stderr:
			- "value length 3 does not equal required length 1"
