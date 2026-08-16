# The two input modes, at the built CLI. UTF-8 is the default; a document
# that declares ISO-8859-1 is read one byte per character.
# see docs/encodings.md

setup:
	- test -x "$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator"

tests:
	# The byte 0xE9 IS the character U+00E9 here. The same byte with no
	# declaration is a broken UTF-8 sequence -- see the last test.
	- desc: a raw high byte is a character in byte mode
	  cmd: |
		set -e
		V="$(mktemp -d)/v"; cp "$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" "$V"
		w="$(mktemp -d)"
		printf %b '<?xml version="1.1" encoding="ISO-8859-1"?><r>caf\351 na\357ve</r>' > "$w/latin1.xml"
		"$V" "$w/latin1.xml"
		echo "document $(wc -c < "$w/latin1.xml") bytes"
	  outputs:
		stdout:
			- "valid XML 1.1 document"
			- "document 60 bytes"

	# The same 10 characters in UTF-8. The content is 12 bytes here against
	# 10 in byte mode, because e-acute and i-diaeresis take two bytes each.
	# The whole document is still smaller: naming the encoding costs 22
	# bytes, which two characters do not repay. A longer text does.
	- desc: the same text in UTF-8 spends two bytes on each high character
	  cmd: |
		set -e
		V="$(mktemp -d)/v"; cp "$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" "$V"
		w="$(mktemp -d)"
		printf %b '<?xml version="1.1"?><r>caf\303\251 na\303\257ve</r>' > "$w/utf8.xml"
		"$V" "$w/utf8.xml"
		echo "document $(wc -c < "$w/utf8.xml") bytes"
	  outputs:
		stdout:
			- "valid XML 1.1 document"
			- "document 40 bytes"

	# Byte mode reaches U+00FF and no further, so anything above it is a
	# character reference. That is the whole cost of the mode.
	- desc: a character above U+00FF needs a reference in byte mode
	  cmd: |
		set -e
		V="$(mktemp -d)/v"; cp "$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" "$V"
		w="$(mktemp -d)"
		printf %b '<?xml version="1.1" encoding="ISO-8859-1"?><r>\351 and &#9731;</r>' > "$w/mixed.xml"
		"$V" "$w/mixed.xml"
	  outputs:
		stdout:
			- "valid XML 1.1 document"

	# A byte that is a character in byte mode is not a UTF-8 sequence, and
	# the default mode says so rather than guessing.
	- desc: the same byte with no declaration is invalid UTF-8
	  exit: 1
	  cmd: 'V="$(mktemp -d)/v"; cp "$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" "$V"; printf %b "<?xml version=\"1.1\"?><r>caf\351</r>" > {outputs.bad.xml}; "$V" {outputs.bad.xml}'
	  outputs:
		stderr:
			- "invalid UTF-8 byte sequence"

	- desc: an encoding this validator does not read is named as such
	  exit: 1
	  cmd: 'V="$(mktemp -d)/v"; cp "$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" "$V"; printf %s "<?xml version=\"1.1\" encoding=\"Shift_JIS\"?><r/>" > {outputs.sjis.xml}; "$V" {outputs.sjis.xml}'
	  outputs:
		stderr:
			- "unsupported encoding"
			- "UTF-8 and ISO-8859-1 are supported"

	# Byte mode is about bytes, not about which characters are legal: a
	# literal NUL is still rejected, and &#0; still carries U+0000.
	- desc: byte mode still rejects a literal NUL byte
	  exit: 1
	  cmd: 'V="$(mktemp -d)/v"; cp "$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" "$V"; printf %b "<?xml version=\"1.1\" encoding=\"ISO-8859-1\"?><r>a\000b</r>" > {outputs.nul.xml}; "$V" {outputs.nul.xml}'
	  outputs:
		stderr:
			- "invalid character U+0000 in character data"

	- desc: byte mode carries the NUL character reference
	  cmd: 'V="$(mktemp -d)/v"; cp "$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" "$V"; printf %b "<?xml version=\"1.1\" encoding=\"ISO-8859-1\"?><r>a&#0;b</r>" > {outputs.ref.xml}; "$V" {outputs.ref.xml}'
	  outputs:
		stdout:
			- "valid XML 1.1 document"

	# Every Latin-1 byte roundtrips: 256 bytes in, one character each, and
	# the schema counts 256 of them.
	- desc: all 256 byte values are 256 characters in byte mode
	  cmd: |
		set -e
		V="$(mktemp -d)/v"; cp "$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" "$V"
		w="$(mktemp -d)"
		printf '<?xml version="1.1" encoding="ISO-8859-1"?><r>' > "$w/all.xml"
		for i in $(seq 0 255); do
			case "$i" in
				0|1|2|3|4|5|6|7|8|11|12|13|14|15|16|17|18|19|20|21|22|23|24|25|26|27|28|29|30|31) printf '&#%d;' "$i" >> "$w/all.xml" ;;
				38) printf '&amp;' >> "$w/all.xml" ;;
				60) printf '&lt;' >> "$w/all.xml" ;;
				62) printf '&gt;' >> "$w/all.xml" ;;
				127|128|129|130|131|132|133|134|135|136|137|138|139|140|141|142|143|144|145|146|147|148|149|150|151|152|153|154|155|156|157|158|159) printf '&#%d;' "$i" >> "$w/all.xml" ;;
				*) printf "\\$(printf '%03o' "$i")" >> "$w/all.xml" ;;
			esac
		done
		printf '</r>' >> "$w/all.xml"
		"$V" --schema {inputs.len256.xsd} "$w/all.xml"
		echo "document $(wc -c < "$w/all.xml") bytes for 256 characters"
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
			- "document 592 bytes for 256 characters"
