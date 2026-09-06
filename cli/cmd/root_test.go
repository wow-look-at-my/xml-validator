package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The command is a package-level variable, so a test that sets a flag has to
// put it back, and no test may drive it while another is. dats/ covers the
// built binary end to end; these cover the paths the binary reaches through,
// which the suites cannot measure.

// run drives the command the way main does, and returns what it printed. It
// runs alone: rootCmd and schemaFile belong to the package, and an overlapping
// Execute re-registers cobra's help flag and panics.
func run(t *testing.T, stdin string, args ...string) (string, error) {
	t.Helper()
	t.Serial()
	t.Cleanup(func() { schemaFile = "" })

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetIn(strings.NewReader(stdin))
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	return out.String(), err
}

func write(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}

func TestValidatesAFile(t *testing.T) {
	out, err := run(t, "", write(t, "doc.xml", `<?xml version="1.1"?><r>a&#0;b</r>`))

	require.NoError(t, err)
	assert.Contains(t, out, "valid XML 1.1 document")
}

func TestValidatesStdin(t *testing.T) {
	out, err := run(t, `<?xml version="1.1"?><r/>`)

	require.NoError(t, err)
	assert.Contains(t, out, "valid XML 1.1 document")
}

// The error travels back to cobra, which prints it and sets the exit status.
func TestReportsAnInvalidDocument(t *testing.T) {
	_, err := run(t, "", write(t, "bad.xml", `<?xml version="1.1"?><r>`))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected end of input")
}

func TestReportsAMissingFile(t *testing.T) {
	_, err := run(t, "", filepath.Join(t.TempDir(), "absent.xml"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot open file")
}

const lengthXSD = `<?xml version="1.1"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
	<xs:element name="r">
		<xs:simpleType>
			<xs:restriction base="xs:string">
				<xs:length value="3"/>
			</xs:restriction>
		</xs:simpleType>
	</xs:element>
</xs:schema>`

func TestValidatesAgainstASchema(t *testing.T) {
	xml := write(t, "doc.xml", `<?xml version="1.1"?><r>a&#0;b</r>`)
	xsd := write(t, "schema.xsd", lengthXSD)

	out, err := run(t, "", "--schema", xsd, xml)

	require.NoError(t, err)
	assert.Contains(t, out, "valid XML 1.1 document (schema validated)")
}

func TestReportsASchemaViolation(t *testing.T) {
	xml := write(t, "doc.xml", `<?xml version="1.1"?><r>ab</r>`)
	xsd := write(t, "schema.xsd", lengthXSD)

	_, err := run(t, "", "--schema", xsd, xml)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "value length 2 does not equal required length 3")
}

// Schema validation re-reads the document from a path, so there is nothing
// left on stdin and the command says so rather than hanging.
func TestSchemaNeedsAFileArgument(t *testing.T) {
	_, err := run(t, "", "--schema", write(t, "schema.xsd", lengthXSD))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--schema requires an XML file argument")
}
