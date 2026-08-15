package validator

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseXMLDeclMissingVersion(t *testing.T) {
	err := Validate(strings.NewReader(`<?xml encoding="UTF-8"?><r/>`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "version")
}

func TestParseXMLDeclBadVersion(t *testing.T) {
	err := Validate(strings.NewReader(`<?xml version="1.0"?><r/>`))
	require.Error(t, err)
}

func TestParseXMLDeclMissingWhitespaceBeforeEncoding(t *testing.T) {
	err := Validate(strings.NewReader(`<?xml version="1.1"encoding="UTF-8"?><r/>`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "whitespace")
}

func TestParseXMLDeclMissingWhitespaceBeforeStandalone(t *testing.T) {
	err := Validate(strings.NewReader(`<?xml version="1.1" encoding="UTF-8"standalone="yes"?><r/>`))
	require.Error(t, err)
}

func TestParseXMLDeclStandaloneInvalid(t *testing.T) {
	err := Validate(strings.NewReader(`<?xml version="1.1" standalone="maybe"?><r/>`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "standalone")
}

func TestParseXMLDeclBadEncName(t *testing.T) {
	err := Validate(strings.NewReader(`<?xml version="1.1" encoding="9foo"?><r/>`))
	require.Error(t, err)
}

func TestParseEntityRefBadName(t *testing.T) {
	err := Validate(strings.NewReader(`<?xml version="1.1"?><r>&#65</r>`))
	require.Error(t, err)
}

func TestParseEntityRefHex(t *testing.T) {
	err := Validate(strings.NewReader(`<?xml version="1.1"?><r>&#x41;</r>`))
	require.NoError(t, err)
}

func TestParseEntityRefBadHex(t *testing.T) {
	err := Validate(strings.NewReader(`<?xml version="1.1"?><r>&#xZZ;</r>`))
	require.Error(t, err)
}

func TestParseEntityRefOutOfRange(t *testing.T) {
	err := Validate(strings.NewReader(`<?xml version="1.1"?><r>&#x110000;</r>`))
	require.Error(t, err)
}

func TestParsePITargetXml(t *testing.T) {
	err := Validate(strings.NewReader(`<?xml version="1.1"?><r><?xml-stylesheet ?></r>`))
	require.NoError(t, err)
}

func TestParsePIBadTarget(t *testing.T) {
	err := Validate(strings.NewReader(`<?xml version="1.1"?><r><?xml ?></r>`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "xml")
}

func TestParsePIUnterminated(t *testing.T) {
	err := Validate(strings.NewReader(`<?xml version="1.1"?><r><?target unterminated </r>`))
	require.Error(t, err)
}

func TestParseCommentDoubleHyphen(t *testing.T) {
	err := Validate(strings.NewReader(`<?xml version="1.1"?><r><!-- bad -- comment --></r>`))
	require.Error(t, err)
}

func TestParseQNameMultiColon(t *testing.T) {
	err := Validate(strings.NewReader(`<?xml version="1.1"?><a:b:c xmlns:a="http://a"/>`))
	require.Error(t, err)
}

func TestParseQNameEmptyPrefix(t *testing.T) {
	err := Validate(strings.NewReader(`<?xml version="1.1"?><:local/>`))
	require.Error(t, err)
}

func TestParseQNameEmptyLocal(t *testing.T) {
	err := Validate(strings.NewReader(`<?xml version="1.1"?><a: xmlns:a="http://a"/>`))
	require.Error(t, err)
}

func TestParseAttrInvalidLT(t *testing.T) {
	err := Validate(strings.NewReader(`<?xml version="1.1"?><r a="x<y"/>`))
	require.Error(t, err)
}

func TestParseAttrUnterminated(t *testing.T) {
	err := Validate(strings.NewReader(`<?xml version="1.1"?><r a="x`))
	require.Error(t, err)
}

func TestParseUndeclaredPrefix(t *testing.T) {
	err := Validate(strings.NewReader(`<?xml version="1.1"?><nope:root xmlns:other="http://a"/>`))
	require.Error(t, err)
}

func TestParseDuplicateAttr(t *testing.T) {
	err := Validate(strings.NewReader(`<?xml version="1.1"?><r a="1" a="2"/>`))
	require.Error(t, err)
}

func TestParseAttrUniquenessNSExpanded(t *testing.T) {
	err := Validate(strings.NewReader(`<?xml version="1.1"?><r xmlns:a="http://x" xmlns:b="http://x" a:k="1" b:k="2"/>`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "conflicts")
}

func TestParseInvalidContent(t *testing.T) {
	err := Validate(strings.NewReader("<?xml version=\"1.1\"?><r><![CDATA[unterminated"))
	require.Error(t, err)
}

func TestParseElementMismatchedClose(t *testing.T) {
	err := Validate(strings.NewReader(`<?xml version="1.1"?><a></b>`))
	require.Error(t, err)
}

func TestParseElementInvalidStartChar(t *testing.T) {
	err := Validate(strings.NewReader(`<?xml version="1.1"?><1abc/>`))
	require.Error(t, err)
}

func TestParseElementMarkupInContent(t *testing.T) {
	err := Validate(strings.NewReader(`<?xml version="1.1"?><r><!ELEMENT x EMPTY></r>`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "markup declarations")
}

func TestParseElementInvalidCharRefInAttr(t *testing.T) {
	err := Validate(strings.NewReader(`<?xml version="1.1"?><r a="&#xFFFF;"/>`))
	require.Error(t, err)
}

func TestParseInvalidCharData(t *testing.T) {
	err := Validate(strings.NewReader("<?xml version=\"1.1\"?><r>]]></r>"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "]]>")
}

func TestParseEntityUnsupported(t *testing.T) {
	err := Validate(strings.NewReader(`<?xml version="1.1"?><r>&custom;</r>`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "general entity reference")
}

func TestParseEntityRefMissingSemicolon(t *testing.T) {
	err := Validate(strings.NewReader(`<?xml version="1.1"?><r>&amp</r>`))
	require.Error(t, err)
}

func TestParseEmptyCharRef(t *testing.T) {
	err := Validate(strings.NewReader(`<?xml version="1.1"?><r>&#;</r>`))
	require.Error(t, err)
}

func TestParseDoctypeRejected(t *testing.T) {
	err := Validate(strings.NewReader(`<?xml version="1.1"?><!DOCTYPE r><r/>`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "DOCTYPE")
}

func TestParseXmlPrefixWrongNS(t *testing.T) {
	err := Validate(strings.NewReader(`<?xml version="1.1"?><r xmlns:xml="http://wrong"/>`))
	require.Error(t, err)
}

func TestParseXmlnsRedeclared(t *testing.T) {
	err := Validate(strings.NewReader(`<?xml version="1.1"?><r xmlns:xmlns="http://wrong"/>`))
	require.Error(t, err)
}

func TestParseEmptyXmlnsPrefix(t *testing.T) {
	err := Validate(strings.NewReader(`<?xml version="1.1"?><r xmlns:="http://x"/>`))
	require.Error(t, err)
}

func TestParseMissingEquals(t *testing.T) {
	err := Validate(strings.NewReader(`<?xml version "1.1"?><r/>`))
	require.Error(t, err)
}

func TestParseMissingQuoteForAttr(t *testing.T) {
	err := Validate(strings.NewReader(`<?xml version="1.1"?><r a=1/>`))
	require.Error(t, err)
}

func TestParseMissingEndTagClose(t *testing.T) {
	err := Validate(strings.NewReader(`<?xml version="1.1"?><r></r`))
	require.Error(t, err)
}
