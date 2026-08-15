package validator

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/wow-look-at-my/go-containers/set"
)

type attribute struct {
	name string
	line int
	col  int
}

func (p *parser) parseElement() error {
	if p.eof() || p.peek() != '<' {
		return p.errorf("expected '<' to begin element")
	}
	p.advance()

	name, err := p.parseName()
	if err != nil {
		return p.errorf("expected element name")
	}
	if err := p.validateQName(name); err != nil {
		return err
	}

	attrs, nsDecls, err := p.parseAttributes()
	if err != nil {
		return err
	}

	p.pushNSScope(nsDecls)

	if err := p.validatePrefixDeclared(name); err != nil {
		p.popNSScope()
		return err
	}
	for _, a := range attrs {
		if err := p.validatePrefixDeclared(a.name); err != nil {
			p.popNSScope()
			return err
		}
	}
	if err := p.checkAttributeUniqueness(attrs); err != nil {
		p.popNSScope()
		return err
	}

	p.skipWhitespace()

	if p.lookingAt("/>") {
		p.expect("/>")
		p.popNSScope()
		return nil
	}

	if err := p.expect(">"); err != nil {
		p.popNSScope()
		return p.errorf("expected '>' or '/>' to close start tag")
	}

	if err := p.parseContent(); err != nil {
		p.popNSScope()
		return err
	}

	if err := p.expect("</"); err != nil {
		p.popNSScope()
		return p.errorf("expected end tag '</%s>'", name)
	}
	endName, err := p.parseName()
	if err != nil {
		p.popNSScope()
		return p.errorf("expected element name in end tag")
	}
	if endName != name {
		p.popNSScope()
		return p.errorf("mismatched end tag: expected '</%s>', got '</%s>'", name, endName)
	}
	p.skipWhitespace()
	if err := p.expect(">"); err != nil {
		p.popNSScope()
		return p.errorf("expected '>' to close end tag")
	}

	p.popNSScope()
	return nil
}

func (p *parser) parseAttributes() ([]attribute, map[string]string, error) {
	var attrs []attribute
	nsDecls := make(map[string]string)
	seen := set.New[string]()

	for {
		if !p.skipWhitespace() {
			break
		}
		if p.eof() || p.peek() == '>' || p.peek() == '/' {
			break
		}

		attrLine, attrCol := p.line, p.col
		aname, err := p.parseName()
		if err != nil {
			break
		}
		if err := p.validateQName(aname); err != nil {
			return nil, nil, err
		}
		if !seen.Add(aname) {
			return nil, nil, &Error{Line: attrLine, Col: attrCol,
				Message: fmt.Sprintf("duplicate attribute %q", aname)}
		}

		if err := p.parseEq(); err != nil {
			return nil, nil, err
		}
		val, err := p.parseAttValue()
		if err != nil {
			return nil, nil, err
		}

		if aname == "xmlns" {
			nsDecls[""] = val
		} else if strings.HasPrefix(aname, "xmlns:") {
			prefix := aname[6:]
			if prefix == "" {
				return nil, nil, &Error{Line: attrLine, Col: attrCol,
					Message: "empty namespace prefix in declaration"}
			}
			if prefix == "xmlns" {
				return nil, nil, &Error{Line: attrLine, Col: attrCol,
					Message: "the prefix 'xmlns' must not be declared"}
			}
			if prefix == "xml" && val != "http://www.w3.org/XML/1998/namespace" {
				return nil, nil, &Error{Line: attrLine, Col: attrCol,
					Message: "the prefix 'xml' must not be bound to any namespace other than http://www.w3.org/XML/1998/namespace"}
			}
			nsDecls[prefix] = val
		} else {
			attrs = append(attrs, attribute{name: aname, line: attrLine, col: attrCol})
		}
	}

	return attrs, nsDecls, nil
}

func (p *parser) parseAttValue() (string, error) {
	if p.eof() {
		return "", p.errorf("expected attribute value")
	}
	quote := p.peek()
	if quote != '"' && quote != '\'' {
		return "", p.errorf("expected '\"' or \"'\" to begin attribute value, got %q", string(quote))
	}
	p.advance()

	var val []rune
	for !p.eof() && p.peek() != quote {
		r := p.peek()
		if r == '<' {
			return "", p.errorf("'<' is not allowed in attribute values")
		}
		if r == '&' {
			resolved, err := p.parseReference()
			if err != nil {
				return "", err
			}
			val = append(val, resolved)
			continue
		}
		if IsRestrictedChar(r) {
			return "", p.errorf("restricted character U+%04X must not appear literally in attribute value (use a character reference)", r)
		}
		if !IsChar(r) {
			return "", p.errorf("invalid character U+%04X in attribute value", r)
		}
		val = append(val, p.advance())
	}
	if p.eof() {
		return "", p.errorf("unterminated attribute value")
	}
	p.advance()
	return string(val), nil
}

func (p *parser) parseContent() error {
	if err := p.parseCharData(); err != nil {
		return err
	}
	for {
		if p.eof() {
			return p.errorf("unexpected end of input inside element content")
		}
		if p.lookingAt("</") {
			return nil
		}
		if p.lookingAt("<!--") {
			if err := p.parseComment(); err != nil {
				return err
			}
		} else if p.lookingAt("<?") {
			if err := p.parsePI(); err != nil {
				return err
			}
		} else if p.lookingAt("<![CDATA[") {
			if err := p.parseCDSect(); err != nil {
				return err
			}
		} else if p.lookingAt("<!") {
			return p.errorf("unsupported: markup declarations are not supported in content")
		} else if p.peek() == '<' {
			if err := p.parseElement(); err != nil {
				return err
			}
		} else if p.peek() == '&' {
			r, err := p.parseReference()
			if err != nil {
				return err
			}
			if !IsCharRefValue(r) {
				return p.errorf("character reference resolves to invalid character U+%04X", r)
			}
		} else {
			return p.errorf("unexpected character %q in content", string(p.peek()))
		}
		if err := p.parseCharData(); err != nil {
			return err
		}
	}
}

func (p *parser) parseCharData() error {
	for !p.eof() {
		r := p.peek()
		if r == '<' || r == '&' {
			return nil
		}
		if r == ']' && p.peekAt(1) == ']' && p.peekAt(2) == '>' {
			return p.errorf("']]>' is not allowed in character data")
		}
		if IsRestrictedChar(r) {
			return p.errorf("restricted character U+%04X must not appear literally in character data (use a character reference)", r)
		}
		if !IsChar(r) {
			return p.errorf("invalid character U+%04X in character data", r)
		}
		p.advance()
	}
	return nil
}

func (p *parser) parseCDSect() error {
	if err := p.expect("<![CDATA["); err != nil {
		return err
	}
	for {
		if p.eof() {
			return p.errorf("unterminated CDATA section")
		}
		if p.lookingAt("]]>") {
			p.expect("]]>")
			return nil
		}
		r := p.advance()
		if !IsChar(r) {
			return p.errorf("invalid character U+%04X in CDATA section", r)
		}
	}
}

func (p *parser) parseReference() (rune, error) {
	if err := p.expect("&"); err != nil {
		return 0, err
	}
	if p.eof() {
		return 0, p.errorf("unexpected end of input in reference")
	}
	if p.peek() == '#' {
		return p.parseCharRef()
	}
	return p.parseEntityRef()
}

func (p *parser) parseCharRef() (rune, error) {
	p.advance() // consume '#'

	var digits []rune
	hex := false
	if !p.eof() && p.peek() == 'x' {
		hex = true
		p.advance()
	}

	for !p.eof() && p.peek() != ';' {
		r := p.peek()
		if hex {
			if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
				return 0, p.errorf("invalid hex digit %q in character reference", string(r))
			}
		} else {
			if r < '0' || r > '9' {
				return 0, p.errorf("invalid digit %q in character reference", string(r))
			}
		}
		digits = append(digits, p.advance())
	}
	if p.eof() {
		return 0, p.errorf("unterminated character reference")
	}
	p.advance() // consume ';'

	if len(digits) == 0 {
		return 0, p.errorf("empty character reference")
	}

	s := string(digits)
	var val int64
	var err error
	if hex {
		val, err = strconv.ParseInt(s, 16, 32)
	} else {
		val, err = strconv.ParseInt(s, 10, 32)
	}
	if err != nil {
		return 0, p.errorf("invalid character reference value %q", s)
	}

	r := rune(val)
	if !IsCharRefValue(r) {
		return 0, p.errorf("character reference &#%s; resolves to invalid XML 1.1 character U+%04X", s, r)
	}
	return r, nil
}

func (p *parser) parseEntityRef() (rune, error) {
	name, err := p.parseName()
	if err != nil {
		return 0, p.errorf("expected entity name")
	}
	if p.eof() || p.peek() != ';' {
		return 0, p.errorf("expected ';' after entity name %q", name)
	}
	p.advance()

	switch name {
	case "amp":
		return '&', nil
	case "lt":
		return '<', nil
	case "gt":
		return '>', nil
	case "apos":
		return '\'', nil
	case "quot":
		return '"', nil
	default:
		return 0, p.errorf("unsupported: general entity reference &%s; (only &amp; &lt; &gt; &apos; &quot; are supported without a DTD)", name)
	}
}
