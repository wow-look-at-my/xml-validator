# The declared version decides the rules, at the built CLI.

setup:
	- test -x "$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator"

tests:
	- desc: an XML 1.0 document is valid and named as XML 1.0
	  cmd: 'printf %s "<?xml version=\"1.0\"?><r/>" | "$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator"'
	  outputs:
		stdout:
			- "valid XML 1.0 document"

	- desc: a document with no declaration is XML 1.0
	  cmd: 'printf %s "<r a=\"b\"/>" | "$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator"'
	  outputs:
		stdout:
			- "valid XML 1.0 document"

	- desc: an XML 1.1 document is still named as XML 1.1
	  cmd: 'printf %s "<?xml version=\"1.1\"?><r/>" | "$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator"'
	  outputs:
		stdout:
			- "valid XML 1.1 document"

	- desc: a C0 control reference is refused in XML 1.0
	  exit: 1
	  cmd: 'printf %s "<?xml version=\"1.0\"?><r>&#x1;</r>" | "$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator"'
	  outputs:
		stderr:
			- "invalid XML 1.0 character U+0001"

	- desc: any other version is refused
	  exit: 1
	  cmd: 'printf %s "<?xml version=\"2.0\"?><r/>" | "$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator"'
	  outputs:
		stderr:
			- "only XML 1.0 and XML 1.1 are supported"
