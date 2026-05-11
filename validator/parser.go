package validator

import (
	"fmt"
	"strconv"
	"strings"
)

type parser struct {
	input       []rune
	pos         int
	line        int
	col         int
	detectedEnc encoding
	nsStack     []map[string]string // stack of prefix -> URI mappings
}

func newParser(input []rune, enc encoding) *parser {
	return &parser{
		input:       input,
		pos:         0,
		line:        1,
		col:         1,
		detectedEnc: enc,
	}
}

func (p *parser) peek() rune {
	if p.pos >= len(p.input) {
		return 0
	}
	return p.input[p.pos]
}

func (p *parser) peekAt(offset int) rune {
	i := p.pos + offset
	if i >= len(p.input) || i < 0 {
		return 0
	}
	return p.input[i]
}

func (p *parser) advance() rune {
	if p.pos >= len(p.input) {
		return 0
	}
	r := p.input[p.pos]
	p.pos++
	if r == '\n' {
		p.line++
		p.col = 1
	} else {
		p.col++
	}
	return r
}

func (p *parser) eof() bool {
	return p.pos >= len(p.input)
}

func (p *parser) errorf(format string, args ...any) *Error {
	return &Error{Line: p.line, Col: p.col, Message: fmt.Sprintf(format, args...)}
}

func (p *parser) expect(s string) error {
	for _, c := range s {
		if p.eof() {
			return p.errorf("unexpected end of input, expected %q", s)
		}
		if p.peek() != c {
			return p.errorf("expected %q, got %q", string(c), string(p.peek()))
		}
		p.advance()
	}
	return nil
}

func (p *parser) lookingAt(s string) bool {
	runes := []rune(s)
	for i, c := range runes {
		if p.pos+i >= len(p.input) || p.input[p.pos+i] != c {
			return false
		}
	}
	return true
}

func (p *parser) skipWhitespace() bool {
	found := false
	for !p.eof() && IsWhitespace(p.peek()) {
		p.advance()
		found = true
	}
	return found
}

func (p *parser) requireWhitespace() error {
	if !p.skipWhitespace() {
		return p.errorf("expected whitespace")
	}
	return nil
}

// parseDocument: document ::= prolog element Misc*
func (p *parser) parseDocument() error {
	if err := p.parseProlog(); err != nil {
		return err
	}
	if p.eof() {
		return p.errorf("expected root element")
	}
	if err := p.parseElement(); err != nil {
		return err
	}
	if err := p.parseMiscSeq(); err != nil {
		return err
	}
	if !p.eof() {
		return p.errorf("unexpected content after root element")
	}
	return nil
}

// parseProlog: prolog ::= XMLDecl Misc*
func (p *parser) parseProlog() error {
	if err := p.parseXMLDecl(); err != nil {
		return err
	}
	if err := p.parseMiscSeq(); err != nil {
		return err
	}
	if p.lookingAt("<!DOCTYPE") {
		return p.errorf("unsupported: DOCTYPE declarations are not supported")
	}
	return nil
}

// parseXMLDecl parses and validates the required XML 1.1 declaration.
func (p *parser) parseXMLDecl() error {
	if !p.lookingAt("<?xml") {
		return p.errorf("XML 1.1 documents must begin with an XML declaration (<?xml ...?>)")
	}

	if err := p.expect("<?xml"); err != nil {
		return err
	}

	// The character after "<?xml" must be whitespace (not a NameChar like in "<?xml-stylesheet")
	if !p.eof() && IsNameChar(p.peek()) {
		return p.errorf("XML declaration must be exactly <?xml, not <?xml followed by name character %q", string(p.peek()))
	}

	if err := p.requireWhitespace(); err != nil {
		return err
	}

	// VersionInfo (required)
	if err := p.expect("version"); err != nil {
		return p.errorf("expected 'version' in XML declaration")
	}
	if err := p.parseEq(); err != nil {
		return err
	}
	version, err := p.parseQuotedValue()
	if err != nil {
		return err
	}
	if version != "1.1" {
		return p.errorf("unsupported XML version %q; only XML 1.1 is supported", version)
	}

	// EncodingDecl (optional)
	savedPos, savedLine, savedCol := p.pos, p.line, p.col
	hadSpace := p.skipWhitespace()
	if p.lookingAt("encoding") {
		if !hadSpace {
			return p.errorf("expected whitespace before 'encoding'")
		}
		if err := p.expect("encoding"); err != nil {
			return err
		}
		if err := p.parseEq(); err != nil {
			return err
		}
		encName, err := p.parseQuotedValue()
		if err != nil {
			return err
		}
		if err := p.validateEncName(encName); err != nil {
			return err
		}
		if err := p.validateEncodingMatch(encName); err != nil {
			return err
		}
		savedPos, savedLine, savedCol = p.pos, p.line, p.col
		hadSpace = p.skipWhitespace()
	}

	// SDDecl (optional)
	if p.lookingAt("standalone") {
		if !hadSpace {
			p.pos, p.line, p.col = savedPos, savedLine, savedCol
			return p.errorf("expected whitespace before 'standalone'")
		}
		if err := p.expect("standalone"); err != nil {
			return err
		}
		if err := p.parseEq(); err != nil {
			return err
		}
		sd, err := p.parseQuotedValue()
		if err != nil {
			return err
		}
		if sd != "yes" && sd != "no" {
			return p.errorf("standalone must be 'yes' or 'no', got %q", sd)
		}
	}

	p.skipWhitespace()
	if err := p.expect("?>"); err != nil {
		return p.errorf("expected '?>' to close XML declaration")
	}
	return nil
}

func (p *parser) validateEncName(name string) error {
	if len(name) == 0 {
		return p.errorf("empty encoding name")
	}
	r := rune(name[0])
	if !((r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')) {
		return p.errorf("encoding name must start with a letter, got %q", string(r))
	}
	for i := 1; i < len(name); i++ {
		r = rune(name[i])
		if !((r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '.' || r == '_' || r == '-') {
			return p.errorf("invalid character %q in encoding name", string(r))
		}
	}
	return nil
}

func (p *parser) validateEncodingMatch(declared string) error {
	upper := strings.ToUpper(declared)
	switch p.detectedEnc {
	case encUTF8:
		if upper != "UTF-8" {
			return p.errorf("encoding declaration %q conflicts with detected UTF-8 encoding", declared)
		}
	case encUTF16BE, encUTF16LE:
		if upper != "UTF-16" && upper != "UTF-16BE" && upper != "UTF-16LE" {
			return p.errorf("encoding declaration %q conflicts with detected UTF-16 encoding", declared)
		}
	}
	return nil
}

func (p *parser) parseEq() error {
	p.skipWhitespace()
	if err := p.expect("="); err != nil {
		return p.errorf("expected '='")
	}
	p.skipWhitespace()
	return nil
}

func (p *parser) parseQuotedValue() (string, error) {
	if p.eof() {
		return "", p.errorf("expected quoted value")
	}
	quote := p.peek()
	if quote != '"' && quote != '\'' {
		return "", p.errorf("expected '\"' or \"'\", got %q", string(quote))
	}
	p.advance()
	var val []rune
	for !p.eof() && p.peek() != quote {
		val = append(val, p.advance())
	}
	if p.eof() {
		return "", p.errorf("unterminated quoted value")
	}
	p.advance() // consume closing quote
	return string(val), nil
}

// parseMiscSeq: Misc* where Misc ::= Comment | PI | S
func (p *parser) parseMiscSeq() error {
	for {
		if p.skipWhitespace() {
			continue
		}
		if p.lookingAt("<!--") {
			if err := p.parseComment(); err != nil {
				return err
			}
			continue
		}
		if p.lookingAt("<?") {
			if err := p.parsePI(); err != nil {
				return err
			}
			continue
		}
		break
	}
	return nil
}

// parseComment: Comment ::= '<!--' ((Char - '-') | ('-' (Char - '-')))* '-->'
func (p *parser) parseComment() error {
	if err := p.expect("<!--"); err != nil {
		return err
	}
	for {
		if p.eof() {
			return p.errorf("unterminated comment")
		}
		if p.lookingAt("-->") {
			p.expect("-->")
			return nil
		}
		if p.peek() == '-' && p.peekAt(1) == '-' {
			return p.errorf("'--' is not allowed inside a comment")
		}
		r := p.advance()
		if !IsChar(r) {
			return p.errorf("invalid character U+%04X in comment", r)
		}
	}
}

// parsePI: PI ::= '<?' PITarget (S (Char* - (Char* '?>' Char*)))? '?>'
func (p *parser) parsePI() error {
	if err := p.expect("<?"); err != nil {
		return err
	}
	name, err := p.parseName()
	if err != nil {
		return p.errorf("expected processing instruction target name")
	}
	if strings.EqualFold(name, "xml") {
		return p.errorf("processing instruction target must not be 'xml' (reserved)")
	}
	if p.lookingAt("?>") {
		p.expect("?>")
		return nil
	}
	if err := p.requireWhitespace(); err != nil {
		return p.errorf("expected whitespace after PI target")
	}
	for {
		if p.eof() {
			return p.errorf("unterminated processing instruction")
		}
		if p.lookingAt("?>") {
			p.expect("?>")
			return nil
		}
		r := p.advance()
		if !IsChar(r) {
			return p.errorf("invalid character U+%04X in processing instruction", r)
		}
	}
}

// parseElement parses a full element: EmptyElemTag | STag content ETag
func (p *parser) parseElement() error {
	if p.eof() || p.peek() != '<' {
		return p.errorf("expected '<' to begin element")
	}
	p.advance() // consume '<'

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

	// Push namespace scope
	p.pushNSScope(nsDecls)

	// Validate element name prefix is declared
	if err := p.validatePrefixDeclared(name); err != nil {
		p.popNSScope()
		return err
	}

	// Validate attribute name prefixes are declared
	for _, a := range attrs {
		if err := p.validatePrefixDeclared(a.name); err != nil {
			p.popNSScope()
			return err
		}
	}

	// Check for namespace-aware attribute uniqueness
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

	// ETag: '</' Name S? '>'
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

type attribute struct {
	name string
	line int
	col  int
}

// parseAttributes parses all attributes, separating namespace declarations from regular attributes.
func (p *parser) parseAttributes() ([]attribute, map[string]string, error) {
	var attrs []attribute
	nsDecls := make(map[string]string)
	seen := make(map[string]bool)

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

		if seen[aname] {
			return nil, nil, &Error{Line: attrLine, Col: attrCol,
				Message: fmt.Sprintf("duplicate attribute %q", aname)}
		}
		seen[aname] = true

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

// parseAttValue: AttValue ::= '"' ([^<&"] | Reference)* '"' | "'" ([^<&'] | Reference)* "'"
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
	p.advance() // closing quote
	return string(val), nil
}

// parseContent: content ::= CharData? ((element | Reference | CDSect | PI | Comment) CharData?)*
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
			if !IsChar(r) {
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

// parseCharData: CharData ::= [^<&]* - ([^<&]* ']]>' [^<&]*)
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

// parseCDSect: CDSect ::= '<![CDATA[' (Char* - (Char* ']]>' Char*)) ']]>'
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

// parseReference: Reference ::= EntityRef | CharRef
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

// parseCharRef: CharRef ::= '&#' [0-9]+ ';' | '&#x' [0-9a-fA-F]+ ';'
func (p *parser) parseCharRef() (rune, error) {
	p.advance() // consume '#'

	var digits []rune
	hex := false
	if !p.eof() && p.peek() == 'x' {
		hex = true
		p.advance()
	}

	for !p.eof() && p.peek() != ';' {
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
	if !IsChar(r) {
		return 0, p.errorf("character reference &#%s; resolves to invalid XML 1.1 character U+%04X", s, r)
	}
	return r, nil
}

// parseEntityRef: EntityRef ::= '&' Name ';'
func (p *parser) parseEntityRef() (rune, error) {
	name, err := p.parseName()
	if err != nil {
		return 0, p.errorf("expected entity name")
	}
	if p.eof() || p.peek() != ';' {
		return 0, p.errorf("expected ';' after entity name %q", name)
	}
	p.advance() // consume ';'

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

// parseName: Name ::= NameStartChar (NameChar)*
func (p *parser) parseName() (string, error) {
	if p.eof() {
		return "", p.errorf("expected name, got end of input")
	}
	r := p.peek()
	if !IsNameStartChar(r) {
		return "", p.errorf("invalid name start character %q (U+%04X)", string(r), r)
	}
	var name []rune
	name = append(name, p.advance())
	for !p.eof() && IsNameChar(p.peek()) {
		name = append(name, p.advance())
	}
	return string(name), nil
}

// validateQName checks that a name is a valid QName (either NCName or NCName:NCName).
func (p *parser) validateQName(name string) error {
	colons := strings.Count(name, ":")
	if colons == 0 {
		return nil
	}
	if colons > 1 {
		return p.errorf("name %q contains multiple colons; QNames may have at most one", name)
	}
	parts := strings.SplitN(name, ":", 2)
	prefix := parts[0]
	local := parts[1]
	if len(prefix) == 0 {
		return p.errorf("empty prefix in qualified name %q", name)
	}
	if len(local) == 0 {
		return p.errorf("empty local part in qualified name %q", name)
	}
	if !IsNCNameStartChar(rune(prefix[0])) {
		return p.errorf("invalid prefix start character in %q", name)
	}
	for _, r := range prefix[1:] {
		if !IsNCNameChar(r) {
			return p.errorf("invalid character %q in prefix of %q", string(r), name)
		}
	}
	if !IsNCNameStartChar(rune(local[0])) {
		return p.errorf("invalid local part start character in %q", name)
	}
	for _, r := range local[1:] {
		if !IsNCNameChar(r) {
			return p.errorf("invalid character %q in local part of %q", string(r), name)
		}
	}
	return nil
}

func (p *parser) pushNSScope(decls map[string]string) {
	scope := make(map[string]string)
	for k, v := range decls {
		scope[k] = v
	}
	p.nsStack = append(p.nsStack, scope)
}

func (p *parser) popNSScope() {
	if len(p.nsStack) > 0 {
		p.nsStack = p.nsStack[:len(p.nsStack)-1]
	}
}

func (p *parser) resolvePrefix(prefix string) (string, bool) {
	if prefix == "xml" {
		return "http://www.w3.org/XML/1998/namespace", true
	}
	if prefix == "xmlns" {
		return "http://www.w3.org/2000/xmlns/", true
	}
	for i := len(p.nsStack) - 1; i >= 0; i-- {
		if uri, ok := p.nsStack[i][prefix]; ok {
			if uri == "" && prefix != "" {
				return "", false
			}
			return uri, true
		}
	}
	return "", prefix == ""
}

func (p *parser) validatePrefixDeclared(name string) error {
	if !strings.Contains(name, ":") {
		return nil
	}
	prefix := name[:strings.Index(name, ":")]
	if _, ok := p.resolvePrefix(prefix); !ok {
		return p.errorf("undeclared namespace prefix %q in name %q", prefix, name)
	}
	return nil
}

// checkAttributeUniqueness checks for duplicate attributes considering namespace expansion.
func (p *parser) checkAttributeUniqueness(attrs []attribute) error {
	type expanded struct {
		ns    string
		local string
	}
	seen := make(map[expanded]string)
	for _, a := range attrs {
		var ns, local string
		if idx := strings.Index(a.name, ":"); idx >= 0 {
			prefix := a.name[:idx]
			local = a.name[idx+1:]
			ns, _ = p.resolvePrefix(prefix)
		} else {
			local = a.name
		}
		key := expanded{ns: ns, local: local}
		if prev, ok := seen[key]; ok {
			return &Error{Line: a.line, Col: a.col,
				Message: fmt.Sprintf("attribute %q conflicts with %q (same namespace-expanded name)", a.name, prev)}
		}
		seen[key] = a.name
	}
	return nil
}
