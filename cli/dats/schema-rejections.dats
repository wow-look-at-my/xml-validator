# The documents dats/schema/schema.xsd must REJECT, one test each.
#
# A validator that accepts every document passes a suite of valid ones. Each
# case here names the error its fixture must produce, so a rejection by the
# wrong rule fails.

setup:
	- test -x "$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator"

tests:
	- desc: a misspelt attribute on entry
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/schema.xsd dats/schema/entry-misspelt-attribute.invalid.xml'
	  exit: 1
	  outputs:
		stderr:
			- 'unexpected attribute "nmae" on element "entry"'

	- desc: an entry nested in an entry
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/schema.xsd dats/schema/entry-nested.invalid.xml'
	  exit: 1
	  outputs:
		stderr:
			- 'unexpected element "entry" in all group of "entry"'

	- desc: an entry with no name
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/schema.xsd dats/schema/entry-no-name.invalid.xml'
	  exit: 1
	  outputs:
		stderr:
			- 'required attribute "name" is missing on element "entry"'

	- desc: an entry with no body
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/schema.xsd dats/schema/entry-no-body.invalid.xml'
	  exit: 1
	  outputs:
		stderr:
			- 'element "entry" requires at least 1 occurrence(s) of "body", got 0'

	- desc: an entry with two bodies
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/schema.xsd dats/schema/entry-two-bodies.invalid.xml'
	  exit: 1
	  outputs:
		stderr:
			- 'element "body" appears too many times (max 1)'

	- desc: an unknown child of entry
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/schema.xsd dats/schema/entry-unknown-child.invalid.xml'
	  exit: 1
	  outputs:
		stderr:
			- 'unexpected element "extra" in all group of "entry"'

	- desc: a section with no name
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/schema.xsd dats/schema/section-no-name.invalid.xml'
	  exit: 1
	  outputs:
		stderr:
			- 'required attribute "name" is missing on element "section"'

	- desc: a section with no body
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/schema.xsd dats/schema/section-no-body.invalid.xml'
	  exit: 1
	  outputs:
		stderr:
			- 'element "section" requires at least 1 occurrence(s) of "body", got 0'

	- desc: an include with no src
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/schema.xsd dats/schema/include-no-src.invalid.xml'
	  exit: 1
	  outputs:
		stderr:
			- 'required attribute "src" is missing on element "include"'

	- desc: an unknown attribute on include
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/schema.xsd dats/schema/include-unknown-attribute.invalid.xml'
	  exit: 1
	  outputs:
		stderr:
			- 'unexpected attribute "from" on element "include"'

	- desc: an enum type with one member
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/schema.xsd dats/schema/option-enum-one-member.invalid.xml'
	  exit: 1
	  outputs:
		stderr:
			- 'attribute "type" on element "option": value "enum(only)" does not match pattern'

	- desc: an option with no name
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/schema.xsd dats/schema/option-no-name.invalid.xml'
	  exit: 1
	  outputs:
		stderr:
			- 'required attribute "name" is missing on element "option"'

	- desc: an option with no type
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/schema.xsd dats/schema/option-no-type.invalid.xml'
	  exit: 1
	  outputs:
		stderr:
			- 'required attribute "type" is missing on element "option"'

	- desc: an option that is required and defaulted
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/schema.xsd dats/schema/option-required-and-default.invalid.xml'
	  exit: 1
	  outputs:
		stderr:
			- 'unexpected attribute "default" on element "option"'

	- desc: a required attribute that is not a bool
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/schema.xsd dats/schema/option-required-not-a-bool.invalid.xml'
	  exit: 1
	  outputs:
		stderr:
			- 'attribute "required" on element "option": value "yes" is not one of the allowed values: true, false'

	- desc: an unknown attribute on option
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/schema.xsd dats/schema/option-unknown-attribute.invalid.xml'
	  exit: 1
	  outputs:
		stderr:
			- 'unexpected attribute "bogus" on element "option"'

	- desc: an unknown option type
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/schema.xsd dats/schema/option-unknown-type.invalid.xml'
	  exit: 1
	  outputs:
		stderr:
			- 'attribute "type" on element "option": value "number" does not match pattern'

	- desc: a root the schema does not declare
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/schema.xsd dats/schema/root-not-a-definition.invalid.xml'
	  exit: 1
	  outputs:
		stderr:
			- 'element "gadget" is not declared as a global element in the schema'

	- desc: an align outside its vocabulary
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/schema.xsd dats/schema/style-bad-align.invalid.xml'
	  exit: 1
	  outputs:
		stderr:
			- 'attribute "align" on element "style": value "middle" does not match pattern'

	- desc: a style bool that is not true or false
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/schema.xsd dats/schema/style-bad-bool.invalid.xml'
	  exit: 1
	  outputs:
		stderr:
			- 'attribute "bold" on element "style": value "yes" does not match pattern'

	- desc: a style with no name
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/schema.xsd dats/schema/style-no-name.invalid.xml'
	  exit: 1
	  outputs:
		stderr:
			- 'required attribute "name" is missing on element "style"'

	- desc: a padding of three values
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/schema.xsd dats/schema/style-three-value-padding.invalid.xml'
	  exit: 1
	  outputs:
		stderr:
			- 'attribute "padding" on element "style": value "1 2 3" does not match pattern'

	- desc: an unknown style attribute
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/schema.xsd dats/schema/style-unknown-attribute.invalid.xml'
	  exit: 1
	  outputs:
		stderr:
			- 'unexpected attribute "foreground" on element "style"'

	- desc: a border outside its vocabulary
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/schema.xsd dats/schema/style-unknown-border.invalid.xml'
	  exit: 1
	  outputs:
		stderr:
			- 'attribute "border" on element "style": value "fancy" does not match pattern'

	- desc: a theme with no name
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/schema.xsd dats/schema/theme-no-name.invalid.xml'
	  exit: 1
	  outputs:
		stderr:
			- 'required attribute "name" is missing on element "theme"'

	- desc: an unknown child of theme
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/schema.xsd dats/schema/theme-unknown-child.invalid.xml'
	  exit: 1
	  outputs:
		stderr:
			- 'unexpected element "palette" in all group of "theme"'

	- desc: a color with no name
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/schema.xsd dats/schema/color-no-name.invalid.xml'
	  exit: 1
	  outputs:
		stderr:
			- 'required attribute "name" is missing on element "color"'

	- desc: a color with no value at all
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/schema.xsd dats/schema/color-no-value.invalid.xml'
	  exit: 1
	  outputs:
		stderr:
			- 'required attribute "light" is missing on element "color"'

	- desc: a color with dark and no light
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/schema.xsd dats/schema/color-only-dark.invalid.xml'
	  exit: 1
	  outputs:
		stderr:
			- 'required attribute "light" is missing on element "color"'

	- desc: a color with light and no dark
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/schema.xsd dats/schema/color-only-light.invalid.xml'
	  exit: 1
	  outputs:
		stderr:
			- 'required attribute "dark" is missing on element "color"'

	- desc: a color with a value and a pair
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/schema.xsd dats/schema/color-value-and-pair.invalid.xml'
	  exit: 1
	  outputs:
		stderr:
			- 'unexpected attribute "light" on element "color"'
