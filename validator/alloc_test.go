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

// A document of literal characters allocates a fixed amount -- reading the
// input and decoding it -- and nothing per character.
func TestLiteralCharactersDoNotAllocatePerCharacter(t *testing.T) {
	small := allocsPerParse(xmlDecl + `<r>` + strings.Repeat("a", 100) + `</r>`)
	large := allocsPerParse(xmlDecl + `<r>` + strings.Repeat("a", 100000) + `</r>`)

	assert.Equal(t, small, large, "allocations grew with the document's length")
}

// Nor does a character reference, which is the path a document pays on every
// escaped character.
func TestCharacterReferencesDoNotAllocatePerReference(t *testing.T) {
	none := allocsPerParse(xmlDecl + `<r>` + strings.Repeat("a", 40000) + `</r>`)
	many := allocsPerParse(xmlDecl + `<r>` + strings.Repeat("&#233;", 10000) + `</r>`)

	assert.LessOrEqual(t, many, none+2, "10,000 references cost more than 2 allocations")
}
