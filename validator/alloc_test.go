package validator

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// The parse loop runs once per character, so anything it allocates is
// multiplied by the size of the document. These pin the cost per character
// and per character reference, which is what the benchmarks measure in bulk.
// see docs/encodings.md

// allocsPerParse is the heap allocations one Validate of doc costs.
func allocsPerParse(doc string) float64 {
	input := []byte(doc)
	return testing.AllocsPerRun(200, func() {
		if err := Validate(bytes.NewReader(input)); err != nil {
			panic(err)
		}
	})
}

// Literal characters cost nothing each. What still grows is the buffer the
// input is read into and the slice it decodes to, and both double, so a
// thousand times the text costs a handful more allocations rather than a
// thousand times as many.
func TestLiteralCharactersDoNotAllocatePerCharacter(t *testing.T) {
	small := allocsPerParse(xmlDecl + `<r>` + strings.Repeat("a", 100) + `</r>`)
	large := allocsPerParse(xmlDecl + `<r>` + strings.Repeat("a", 100000) + `</r>`)

	assert.Less(t, large, small*3, "1000x the text cost more than 3x the allocations")
}

// A character reference costs nothing either, which is the path a document
// pays on every escaped character.
func TestCharacterReferencesDoNotAllocatePerReference(t *testing.T) {
	none := allocsPerParse(xmlDecl + `<r>` + strings.Repeat("a", 40000) + `</r>`)
	many := allocsPerParse(xmlDecl + `<r>` + strings.Repeat("&#233;", 10000) + `</r>`)

	assert.LessOrEqual(t, many, none+2, "10,000 references cost more than 2 allocations")
}

// And neither does a predefined entity reference, which used to build its
// name as a string before comparing it.
func TestEntityReferencesDoNotAllocatePerReference(t *testing.T) {
	none := allocsPerParse(xmlDecl + `<r>` + strings.Repeat("a", 50000) + `</r>`)
	many := allocsPerParse(xmlDecl + `<r>` + strings.Repeat("&amp;", 10000) + `</r>`)

	assert.LessOrEqual(t, many, none+2, "10,000 entity references cost more than 2 allocations")
}
