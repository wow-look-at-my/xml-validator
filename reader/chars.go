package reader

func IsChar(r rune) bool {
	return (r >= 0x1 && r <= 0xD7FF) ||
		(r >= 0xE000 && r <= 0xFFFD) ||
		(r >= 0x10000 && r <= 0x10FFFF)
}

// IsChar10 reports whether r is a Char under Version10. A C0 control other
// than tab, LF and CR is not a character, even as a reference.
func IsChar10(r rune) bool {
	return r == 0x9 || r == 0xA || r == 0xD ||
		(r >= 0x20 && r <= 0xD7FF) ||
		(r >= 0xE000 && r <= 0xFFFD) ||
		(r >= 0x10000 && r <= 0x10FFFF)
}

// IsCharRefValue returns true if r may be produced by a character reference.
func IsCharRefValue(r rune) bool {
	return r == 0 || IsChar(r)
}

// [2a] RestrictedChar ::= [#x1-#x8] | [#xB-#xC] | [#xE-#x1F] | [#x7F-#x84] | [#x86-#x9F]
func IsRestrictedChar(r rune) bool {
	return (r >= 0x1 && r <= 0x8) ||
		(r >= 0xB && r <= 0xC) ||
		(r >= 0xE && r <= 0x1F) ||
		(r >= 0x7F && r <= 0x84) ||
		(r >= 0x86 && r <= 0x9F)
}

// IsWhitespace returns true if r is XML whitespace.
func IsWhitespace(r rune) bool {
	return r == 0x20 || r == 0x9 || r == 0xD || r == 0xA
}

// [#xF8-#x2FF] | [#x370-#x37D] | [#x37F-#x1FFF] | [#x200C-#x200D] |
// [#x2070-#x218F] | [#x2C00-#x2FEF] | [#x3001-#xD7FF] | [#xF900-#xFDCF] |
// [#xFDF0-#xFFFD] | [#x10000-#xEFFFF]
func IsNameStartChar(r rune) bool {
	return r == ':' ||
		(r >= 'A' && r <= 'Z') ||
		r == '_' ||
		(r >= 'a' && r <= 'z') ||
		(r >= 0xC0 && r <= 0xD6) ||
		(r >= 0xD8 && r <= 0xF6) ||
		(r >= 0xF8 && r <= 0x2FF) ||
		(r >= 0x370 && r <= 0x37D) ||
		(r >= 0x37F && r <= 0x1FFF) ||
		(r >= 0x200C && r <= 0x200D) ||
		(r >= 0x2070 && r <= 0x218F) ||
		(r >= 0x2C00 && r <= 0x2FEF) ||
		(r >= 0x3001 && r <= 0xD7FF) ||
		(r >= 0xF900 && r <= 0xFDCF) ||
		(r >= 0xFDF0 && r <= 0xFFFD) ||
		(r >= 0x10000 && r <= 0xEFFFF)
}

func IsNameChar(r rune) bool {
	return IsNameStartChar(r) ||
		r == '-' ||
		r == '.' ||
		(r >= '0' && r <= '9') ||
		r == 0xB7 ||
		(r >= 0x300 && r <= 0x36F) ||
		(r >= 0x203F && r <= 0x2040)
}

// IsNCNameStartChar is NameStartChar without colon (for namespace-qualified names).
func IsNCNameStartChar(r rune) bool {
	return r != ':' && IsNameStartChar(r)
}

// IsNCNameChar is NameChar without colon.
func IsNCNameChar(r rune) bool {
	return r != ':' && IsNameChar(r)
}
