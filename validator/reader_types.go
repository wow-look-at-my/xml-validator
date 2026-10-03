package validator

import "github.com/wow-look-at-my/xml-validator/reader"

// Reading a document -- decoding its bytes, the character classes, the tree
// model and the tree parser -- lives in the reader module.
type (
	Document = reader.Document
	Element  = reader.Element
	Attr     = reader.Attr
	CharData = reader.CharData
	Node     = reader.Node
	Error    = reader.Error
)

// ParseTree parses a document into a tree without validating it.
var ParseTree = reader.ParseTree

var (
	IsChar            = reader.IsChar
	IsCharRefValue    = reader.IsCharRefValue
	IsRestrictedChar  = reader.IsRestrictedChar
	IsWhitespace      = reader.IsWhitespace
	IsNameStartChar   = reader.IsNameStartChar
	IsNameChar        = reader.IsNameChar
	IsNCNameStartChar = reader.IsNCNameStartChar
	IsNCNameChar      = reader.IsNCNameChar
)
