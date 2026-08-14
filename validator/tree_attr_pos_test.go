package validator

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Attributes carry their own position so a consumer reporting a problem with an
// attribute's value can point at that attribute rather than at the element,
// which on a multi-attribute element is the wrong place to look.
func TestAttrPositions(t *testing.T) {
	src := `<?xml version="1.1" encoding="UTF-8"?>
<root>
	<box width="10" height="20"/>
</root>`

	doc, err := ParseTree(strings.NewReader(src))
	require.NoError(t, err)

	box := doc.Root.ChildElements()[0]
	require.Len(t, box.Attrs, 2)

	assert.Equal(t, 3, box.Attrs[0].Line, "both attributes are on the third line")
	assert.Equal(t, 3, box.Attrs[1].Line)

	assert.Greater(t, box.Attrs[0].Col, box.Line,
		"the first attribute starts after the element name")
	assert.Greater(t, box.Attrs[1].Col, box.Attrs[0].Col,
		"the second attribute is distinguishable from the first")
}

func TestAttrPositionsAcrossLines(t *testing.T) {
	src := `<?xml version="1.1" encoding="UTF-8"?>
<box
	width="10"
	height="20"/>`

	doc, err := ParseTree(strings.NewReader(src))
	require.NoError(t, err)

	attrs := doc.Root.Attrs
	require.Len(t, attrs, 2)

	assert.Equal(t, 3, attrs[0].Line, "an attribute on its own line reports that line")
	assert.Equal(t, 4, attrs[1].Line)
	assert.Equal(t, 2, attrs[0].Col, "the tab is one column, so the name starts at column 2")
}

// Namespace declarations are dropped from Attrs, which must not shift the
// positions recorded for the attributes that remain.
func TestAttrPositionsSurviveNamespaceDeclarations(t *testing.T) {
	src := `<?xml version="1.1" encoding="UTF-8"?>
<box xmlns="urn:example" width="10"/>`

	doc, err := ParseTree(strings.NewReader(src))
	require.NoError(t, err)

	require.Len(t, doc.Root.Attrs, 1)
	assert.Equal(t, "width", doc.Root.Attrs[0].Name)
	assert.Equal(t, 2, doc.Root.Attrs[0].Line)
	assert.Equal(t, len("<box xmlns=\"urn:example\" ")+1, doc.Root.Attrs[0].Col,
		"the surviving attribute keeps its own column, not the dropped one's")
}
