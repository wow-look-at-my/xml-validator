# Two constraint classes that fail QUIETLY when they regress: a facet on an
# attribute value, and an identity constraint. Element text keeps validating
# either way, so a schema still reports "schema validated" while the constraint
# it states does nothing. A consumer then ships a guarantee it does not have.
#
# $GO_TOOLCHAIN_DATS_BUILD_DIR holds copies of the binaries this build just
# made. It is read-only inside the sandbox, which is enough to exec them.

setup:
	- test -x "$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator"

tests:
	# The pair that names the bug: same type, same value, element vs attribute.
	# Checking only the element case is what disguises a missing attribute facet.
	- desc: a pattern facet rejects the value as element text
	  exit: 1
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema {inputs.pat.xsd} {inputs.elem.xml}'
	  inputs:
		files:
			elem.xml: |
				<?xml version="1.1"?>
				<r><t>ABC123</t></r>
			pat.xsd: |
				<?xml version="1.1"?>
				<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
					<xs:simpleType name="lower">
						<xs:restriction base="xs:token">
							<xs:pattern value="[a-z]+"/>
						</xs:restriction>
					</xs:simpleType>
					<xs:element name="r">
						<xs:complexType>
							<xs:sequence>
								<xs:element name="t" type="lower" minOccurs="0"/>
							</xs:sequence>
							<xs:attribute name="p" type="lower"/>
						</xs:complexType>
					</xs:element>
				</xs:schema>
	  outputs:
		stderr:
			- 'does not match pattern "[a-z]+"'
		"!stdout":
			- "valid XML 1.1 document"

	- desc: the same facet rejects the same value as an attribute
	  exit: 1
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema {inputs.pat.xsd} {inputs.attr.xml}'
	  inputs:
		files:
			attr.xml: |
				<?xml version="1.1"?>
				<r p="ABC123"/>
			pat.xsd: |
				<?xml version="1.1"?>
				<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
					<xs:simpleType name="lower">
						<xs:restriction base="xs:token">
							<xs:pattern value="[a-z]+"/>
						</xs:restriction>
					</xs:simpleType>
					<xs:element name="r">
						<xs:complexType>
							<xs:sequence>
								<xs:element name="t" type="lower" minOccurs="0"/>
							</xs:sequence>
							<xs:attribute name="p" type="lower"/>
						</xs:complexType>
					</xs:element>
				</xs:schema>
	  outputs:
		stderr:
			- 'attribute "p" on element "r"'
			- 'does not match pattern "[a-z]+"'
		"!stdout":
			- "valid XML 1.1 document"

	- desc: minLength applies to an attribute value
	  exit: 1
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema {inputs.min.xsd} {inputs.empty.xml}'
	  inputs:
		files:
			empty.xml: |
				<?xml version="1.1"?>
				<r reason=""/>
			min.xsd: |
				<?xml version="1.1"?>
				<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
					<xs:simpleType name="nonempty">
						<xs:restriction base="xs:string">
							<xs:minLength value="1"/>
						</xs:restriction>
					</xs:simpleType>
					<xs:element name="r">
						<xs:complexType>
							<xs:attribute name="reason" type="nonempty"/>
						</xs:complexType>
					</xs:element>
				</xs:schema>
	  outputs:
		stderr:
			- 'attribute "reason" on element "r"'
			- "length 0 is less than minLength 1"

	- desc: a valid document against the same schema still passes
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema {inputs.key.xsd} {inputs.ok.xml}'
	  inputs:
		files:
			ok.xml: |
				<?xml version="1.1"?>
				<cfg>
					<provider id="alpha"/>
					<provider id="beta"/>
					<role provider="alpha"/>
				</cfg>
			key.xsd: |
				<?xml version="1.1"?>
				<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
					<xs:element name="cfg">
						<xs:complexType>
							<xs:sequence>
								<xs:element name="provider" maxOccurs="unbounded">
									<xs:complexType>
										<xs:attribute name="id" type="xs:token" use="required"/>
									</xs:complexType>
								</xs:element>
								<xs:element name="role" minOccurs="0" maxOccurs="unbounded">
									<xs:complexType>
										<xs:attribute name="provider" type="xs:token" use="required"/>
									</xs:complexType>
								</xs:element>
							</xs:sequence>
						</xs:complexType>
						<xs:key name="providerId">
							<xs:selector xpath="provider"/>
							<xs:field xpath="@id"/>
						</xs:key>
						<xs:keyref name="roleProvider" refer="providerId">
							<xs:selector xpath="role"/>
							<xs:field xpath="@provider"/>
						</xs:keyref>
					</xs:element>
				</xs:schema>
	  outputs:
		stdout:
			- "valid XML 1.1 document (schema validated)"

	# A validator that ignores identity constraints accepts this and says
	# nothing, which is indistinguishable from the document being correct.
	- desc: xs:key rejects a repeated value and cites both positions
	  exit: 1
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema {inputs.key.xsd} {inputs.dup.xml}'
	  inputs:
		files:
			dup.xml: |
				<?xml version="1.1"?>
				<cfg>
					<provider id="alpha"/>
					<provider id="alpha"/>
				</cfg>
			key.xsd: |
				<?xml version="1.1"?>
				<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
					<xs:element name="cfg">
						<xs:complexType>
							<xs:sequence>
								<xs:element name="provider" maxOccurs="unbounded">
									<xs:complexType>
										<xs:attribute name="id" type="xs:token" use="required"/>
									</xs:complexType>
								</xs:element>
								<xs:element name="role" minOccurs="0" maxOccurs="unbounded">
									<xs:complexType>
										<xs:attribute name="provider" type="xs:token" use="required"/>
									</xs:complexType>
								</xs:element>
							</xs:sequence>
						</xs:complexType>
						<xs:key name="providerId">
							<xs:selector xpath="provider"/>
							<xs:field xpath="@id"/>
						</xs:key>
						<xs:keyref name="roleProvider" refer="providerId">
							<xs:selector xpath="role"/>
							<xs:field xpath="@provider"/>
						</xs:keyref>
					</xs:element>
				</xs:schema>
	  outputs:
		stderr:
			- 'xs:key "providerId"'
			- "repeats a value first used at line 3"
		"!stdout":
			- "valid XML 1.1 document"

	- desc: xs:keyref rejects a reference to an undeclared value
	  exit: 1
	  cmd: '"$GO_TOOLCHAIN_DATS_BUILD_DIR/xml-validator" --schema {inputs.key.xsd} {inputs.dangling.xml}'
	  inputs:
		files:
			dangling.xml: |
				<?xml version="1.1"?>
				<cfg>
					<provider id="alpha"/>
					<role provider="nope"/>
				</cfg>
			key.xsd: |
				<?xml version="1.1"?>
				<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
					<xs:element name="cfg">
						<xs:complexType>
							<xs:sequence>
								<xs:element name="provider" maxOccurs="unbounded">
									<xs:complexType>
										<xs:attribute name="id" type="xs:token" use="required"/>
									</xs:complexType>
								</xs:element>
								<xs:element name="role" minOccurs="0" maxOccurs="unbounded">
									<xs:complexType>
										<xs:attribute name="provider" type="xs:token" use="required"/>
									</xs:complexType>
								</xs:element>
							</xs:sequence>
						</xs:complexType>
						<xs:key name="providerId">
							<xs:selector xpath="provider"/>
							<xs:field xpath="@id"/>
						</xs:key>
						<xs:keyref name="roleProvider" refer="providerId">
							<xs:selector xpath="role"/>
							<xs:field xpath="@provider"/>
						</xs:keyref>
					</xs:element>
				</xs:schema>
	  outputs:
		stderr:
			- 'xs:keyref "roleProvider"'
			- 'refers to a value that "providerId" does not declare'
		"!stdout":
			- "valid XML 1.1 document"
