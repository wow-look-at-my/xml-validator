package validator_test

import (
	"fmt"
	"strings"

	"github.com/wow-look-at-my/xml-validator/validator"
)

func ExampleValidate() {
	err := validator.Validate(strings.NewReader(`<?xml version="1.1"?><greeting>hi</greeting>`))
	fmt.Println(err)
	// Output: <nil>
}

func ExampleValidate_invalid() {
	err := validator.Validate(strings.NewReader(`<?xml version="1.0"?><r/>`))
	fmt.Println(err != nil)
	// Output: true
}

func ExampleValidateWithSchema() {
	xml := `<?xml version="1.1"?><note><body>hi</body></note>`
	xsd := `<?xml version="1.1"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="note">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="body" type="xs:string"/>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`

	err := validator.ValidateWithSchema(strings.NewReader(xml), strings.NewReader(xsd))
	fmt.Println(err)
	// Output: <nil>
}

func ExampleParseTree() {
	doc, err := validator.ParseTree(strings.NewReader(`<?xml version="1.1"?><r a="1"><c/></r>`))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(doc.Root.Local, len(doc.Root.Attrs), len(doc.Root.ChildElements()))
	// Output: r 1 1
}

func ExampleError() {
	err := validator.Validate(strings.NewReader(`<r/>`))
	if vErr, ok := err.(*validator.Error); ok {
		fmt.Printf("validation failed at line %d, column %d\n", vErr.Line, vErr.Col)
	}
	// Output: validation failed at line 1, column 1
}
