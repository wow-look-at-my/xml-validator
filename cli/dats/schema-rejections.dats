# The documents tml.schema.xsd must REJECT, one test each.
#
# A validator that accepts every document passes a suite of valid ones. Each
# case here names the error its fixture must produce, so a rejection by the
# wrong rule fails.

setup:
	- test -x "$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator"

tests:
	- desc: a misspelt attribute on Component
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/tml.schema.xsd dats/schema/invalid/component-misspelt-attribute.tml'
	  exit: 1
	  outputs:
		stderr:
			- 'unexpected attribute "nmae" on element "Component"'

	- desc: a Component nested in a Component
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/tml.schema.xsd dats/schema/invalid/component-nested.tml'
	  exit: 1
	  outputs:
		stderr:
			- 'unexpected element "Component" in all group of "Component"'

	- desc: a Component with no name
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/tml.schema.xsd dats/schema/invalid/component-no-name.tml'
	  exit: 1
	  outputs:
		stderr:
			- 'required attribute "name" is missing on element "Component"'

	- desc: a Component with no Template
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/tml.schema.xsd dats/schema/invalid/component-no-template.tml'
	  exit: 1
	  outputs:
		stderr:
			- 'element "Component" requires at least 1 occurrence(s) of "Template", got 0'

	- desc: a Component with two Templates
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/tml.schema.xsd dats/schema/invalid/component-two-templates.tml'
	  exit: 1
	  outputs:
		stderr:
			- 'element "Template" appears too many times (max 1)'

	- desc: an unknown child of Component
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/tml.schema.xsd dats/schema/invalid/component-unknown-child.tml'
	  exit: 1
	  outputs:
		stderr:
			- 'unexpected element "Slots" in all group of "Component"'

	- desc: a DataTemplate with no name
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/tml.schema.xsd dats/schema/invalid/datatemplate-no-name.tml'
	  exit: 1
	  outputs:
		stderr:
			- 'required attribute "name" is missing on element "DataTemplate"'

	- desc: a DataTemplate with no Template
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/tml.schema.xsd dats/schema/invalid/datatemplate-no-template.tml'
	  exit: 1
	  outputs:
		stderr:
			- 'element "DataTemplate" requires at least 1 occurrence(s) of "Template", got 0'

	- desc: an Import with no src
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/tml.schema.xsd dats/schema/invalid/import-no-src.tml'
	  exit: 1
	  outputs:
		stderr:
			- 'required attribute "src" is missing on element "Import"'

	- desc: an unknown attribute on Import
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/tml.schema.xsd dats/schema/invalid/import-unknown-attribute.tml'
	  exit: 1
	  outputs:
		stderr:
			- 'unexpected attribute "from" on element "Import"'

	- desc: an enum type with one member
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/tml.schema.xsd dats/schema/invalid/property-enum-one-member.tml'
	  exit: 1
	  outputs:
		stderr:
			- 'attribute "type" on element "Property": value "enum(only)" does not match pattern'

	- desc: a Property with no name
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/tml.schema.xsd dats/schema/invalid/property-no-name.tml'
	  exit: 1
	  outputs:
		stderr:
			- 'required attribute "name" is missing on element "Property"'

	- desc: a Property with no type
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/tml.schema.xsd dats/schema/invalid/property-no-type.tml'
	  exit: 1
	  outputs:
		stderr:
			- 'required attribute "type" is missing on element "Property"'

	- desc: a Property that is required and defaulted
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/tml.schema.xsd dats/schema/invalid/property-required-and-default.tml'
	  exit: 1
	  outputs:
		stderr:
			- 'unexpected attribute "default" on element "Property"'

	- desc: a required attribute that is not a bool
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/tml.schema.xsd dats/schema/invalid/property-required-not-a-bool.tml'
	  exit: 1
	  outputs:
		stderr:
			- 'attribute "required" on element "Property": value "yes" is not one of the allowed values: true, false'

	- desc: an unknown attribute on Property
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/tml.schema.xsd dats/schema/invalid/property-unknown-attribute.tml'
	  exit: 1
	  outputs:
		stderr:
			- 'unexpected attribute "bogus" on element "Property"'

	- desc: an unknown property type
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/tml.schema.xsd dats/schema/invalid/property-unknown-type.tml'
	  exit: 1
	  outputs:
		stderr:
			- 'attribute "type" on element "Property": value "number" does not match pattern'

	- desc: a root that is neither a Component nor a Theme
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/tml.schema.xsd dats/schema/invalid/root-not-a-definition.tml'
	  exit: 1
	  outputs:
		stderr:
			- 'element "Widget" is not declared as a global element in the schema'

	- desc: an align outside its vocabulary
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/tml.schema.xsd dats/schema/invalid/style-bad-align.tml'
	  exit: 1
	  outputs:
		stderr:
			- 'attribute "align" on element "Style": value "middle" does not match pattern'

	- desc: a style bool that is not true or false
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/tml.schema.xsd dats/schema/invalid/style-bad-bool.tml'
	  exit: 1
	  outputs:
		stderr:
			- 'attribute "bold" on element "Style": value "yes" does not match pattern'

	- desc: a Style with no name
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/tml.schema.xsd dats/schema/invalid/style-no-name.tml'
	  exit: 1
	  outputs:
		stderr:
			- 'required attribute "name" is missing on element "Style"'

	- desc: a thickness of three values
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/tml.schema.xsd dats/schema/invalid/style-three-value-thickness.tml'
	  exit: 1
	  outputs:
		stderr:
			- 'attribute "padding" on element "Style": value "1 2 3" does not match pattern'

	- desc: an unknown style attribute
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/tml.schema.xsd dats/schema/invalid/style-unknown-attribute.tml'
	  exit: 1
	  outputs:
		stderr:
			- 'unexpected attribute "foreground" on element "Style"'

	- desc: a border outside its vocabulary
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/tml.schema.xsd dats/schema/invalid/style-unknown-border.tml'
	  exit: 1
	  outputs:
		stderr:
			- 'attribute "border" on element "Style": value "fancy" does not match pattern'

	- desc: a Theme with no name
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/tml.schema.xsd dats/schema/invalid/theme-no-name.tml'
	  exit: 1
	  outputs:
		stderr:
			- 'required attribute "name" is missing on element "Theme"'

	- desc: an unknown child of Theme
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/tml.schema.xsd dats/schema/invalid/theme-unknown-child.tml'
	  exit: 1
	  outputs:
		stderr:
			- 'unexpected element "Palette" in all group of "Theme"'

	- desc: a Token with no name
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/tml.schema.xsd dats/schema/invalid/token-no-name.tml'
	  exit: 1
	  outputs:
		stderr:
			- 'required attribute "name" is missing on element "Token"'

	- desc: a Token with no value at all
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/tml.schema.xsd dats/schema/invalid/token-no-value.tml'
	  exit: 1
	  outputs:
		stderr:
			- 'required attribute "light" is missing on element "Token"'

	- desc: a Token with dark and no light
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/tml.schema.xsd dats/schema/invalid/token-only-dark.tml'
	  exit: 1
	  outputs:
		stderr:
			- 'required attribute "light" is missing on element "Token"'

	- desc: a Token with light and no dark
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/tml.schema.xsd dats/schema/invalid/token-only-light.tml'
	  exit: 1
	  outputs:
		stderr:
			- 'required attribute "dark" is missing on element "Token"'

	- desc: a Token with a value and a pair
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema dats/schema/tml.schema.xsd dats/schema/invalid/token-value-and-pair.tml'
	  exit: 1
	  outputs:
		stderr:
			- 'unexpected attribute "light" on element "Token"'
