package validator

import (
	"fmt"
	"strings"

	"github.com/wow-look-at-my/go-containers/set"
)

// Conditional type assignment: xs:alternative gives an element a type chosen
// per instance. The first alternative whose test holds decides the type, and an
// alternative with no test always holds, which is how a default is written.
//
// The test language is a subset, and anything outside it is a hard error at
// schema-parse time. A test this engine cannot evaluate would otherwise pick
// the wrong type in silence, which is worse than saying so.
//
// see docs/conditional-types.md

// TypeAlternative is one xs:alternative on an element declaration.
type TypeAlternative struct {
	TypeName string
	Type     Type
	// test is nil for an alternative with no test attribute: the default.
	test     boolExpr
	testText string
}

// boolExpr is a compiled test. It reads only the element's attributes, which is
// what the subset covers.
type boolExpr interface {
	eval(el *Element) bool
}

type attrPresent struct{ name nameTest }

func (e attrPresent) eval(el *Element) bool {
	_, ok := lookupAttrValue(el, e.name)
	return ok
}

type attrCompare struct {
	name  nameTest
	value string
	equal bool
}

func (e attrCompare) eval(el *Element) bool {
	got, ok := lookupAttrValue(el, e.name)
	if !ok {
		// An absent attribute equals nothing and differs from nothing: XPath
		// compares an empty sequence to false either way.
		return false
	}
	return (got == e.value) == e.equal
}

type notExpr struct{ inner boolExpr }

func (e notExpr) eval(el *Element) bool { return !e.inner.eval(el) }

type andExpr struct{ terms []boolExpr }

func (e andExpr) eval(el *Element) bool {
	for _, t := range e.terms {
		if !t.eval(el) {
			return false
		}
	}
	return true
}

type orExpr struct{ terms []boolExpr }

func (e orExpr) eval(el *Element) bool {
	for _, t := range e.terms {
		if t.eval(el) {
			return true
		}
	}
	return false
}

func lookupAttrValue(el *Element, want nameTest) (string, bool) {
	for _, attr := range el.Attrs {
		if attr.Prefix == "xmlns" || attr.Name == "xmlns" {
			continue
		}
		if want.matches(attr.Namespace, attr.Local) {
			return attr.Value, true
		}
	}
	return "", false
}

// compileTest builds a test expression. The grammar is:
//
//	test  ::= term ( ('and' | 'or') term )*
//	term  ::= '@name' | '@name' ('=' | '!=') "'literal'" | 'not(' term ')'
//
// Mixing and with or in one test is rejected rather than guessed at: XPath
// binds and tighter than or, and a reader should not have to know that to know
// what a schema means.
func compileTest(expr string, s *Schema) (boolExpr, error) {
	trimmed := strings.TrimSpace(expr)
	if trimmed == "" {
		return nil, fmt.Errorf("test is empty")
	}

	andTerms, orTerms := splitTest(trimmed)
	switch {
	case len(andTerms) > 1 && len(orTerms) > 1:
		return nil, fmt.Errorf("unsupported test %q: mixing \"and\" with \"or\" in one test is not supported; use one or the other", expr)
	case len(andTerms) > 1:
		return compileTerms(andTerms, expr, s, true)
	case len(orTerms) > 1:
		return compileTerms(orTerms, expr, s, false)
	}
	return compileTerm(trimmed, expr, s)
}

func compileTerms(parts []string, whole string, s *Schema, isAnd bool) (boolExpr, error) {
	terms := make([]boolExpr, 0, len(parts))
	for _, part := range parts {
		t, err := compileTerm(strings.TrimSpace(part), whole, s)
		if err != nil {
			return nil, err
		}
		terms = append(terms, t)
	}
	if isAnd {
		return andExpr{terms: terms}, nil
	}
	return orExpr{terms: terms}, nil
}

// splitTest splits on the top-level "and" and "or" keywords, leaving anything
// inside quotes or parentheses alone.
func splitTest(expr string) (andParts, orParts []string) {
	var (
		depth  int
		quote  byte
		start  int
		andCut []int
		orCut  []int
	)
	for i := 0; i < len(expr); i++ {
		c := expr[i]
		switch {
		case quote != 0:
			if c == quote {
				quote = 0
			}
		case c == '\'' || c == '"':
			quote = c
		case c == '(':
			depth++
		case c == ')':
			depth--
		case depth == 0 && isKeywordAt(expr, i, "and"):
			andCut = append(andCut, i)
		case depth == 0 && isKeywordAt(expr, i, "or"):
			orCut = append(orCut, i)
		}
	}
	andParts = cutAt(expr, andCut, len("and"), start)
	orParts = cutAt(expr, orCut, len("or"), start)
	return andParts, orParts
}

func cutAt(expr string, cuts []int, width, start int) []string {
	if len(cuts) == 0 {
		return []string{expr}
	}
	var out []string
	for _, c := range cuts {
		out = append(out, expr[start:c])
		start = c + width
	}
	return append(out, expr[start:])
}

// isKeywordAt reports whether the word at i is exactly kw, with a boundary on
// both sides -- "and" inside "brand" is not an operator.
func isKeywordAt(expr string, i int, kw string) bool {
	if !strings.HasPrefix(expr[i:], kw) {
		return false
	}
	if i > 0 && !isTestSpace(expr[i-1]) && expr[i-1] != ')' {
		return false
	}
	after := i + len(kw)
	return after < len(expr) && (isTestSpace(expr[after]) || expr[after] == '(')
}

func isTestSpace(c byte) bool { return c == ' ' || c == '\t' || c == '\n' || c == '\r' }

func compileTerm(term, whole string, s *Schema) (boolExpr, error) {
	if inner, ok := strings.CutPrefix(term, "not("); ok {
		if !strings.HasSuffix(inner, ")") {
			return nil, fmt.Errorf("unsupported test %q: not( is never closed", whole)
		}
		sub, err := compileTerm(strings.TrimSpace(strings.TrimSuffix(inner, ")")), whole, s)
		if err != nil {
			return nil, err
		}
		return notExpr{inner: sub}, nil
	}

	name, op, literal, err := splitComparison(term, whole)
	if err != nil {
		return nil, err
	}
	nt, err := compileTestAttrName(name, whole, s)
	if err != nil {
		return nil, err
	}
	if op == "" {
		return attrPresent{name: nt}, nil
	}
	return attrCompare{name: nt, value: literal, equal: op == "="}, nil
}

func splitComparison(term, whole string) (name, op, literal string, err error) {
	for _, candidate := range []string{"!=", "="} {
		i := strings.Index(term, candidate)
		if i < 0 {
			continue
		}
		lit := strings.TrimSpace(term[i+len(candidate):])
		if len(lit) < 2 || (lit[0] != '\'' && lit[0] != '"') || lit[len(lit)-1] != lit[0] {
			return "", "", "", fmt.Errorf("unsupported test %q: the right side of %s must be a quoted literal", whole, candidate)
		}
		return strings.TrimSpace(term[:i]), candidate, lit[1 : len(lit)-1], nil
	}
	return term, "", "", nil
}

func compileTestAttrName(name, whole string, s *Schema) (nameTest, error) {
	rest, ok := strings.CutPrefix(strings.TrimSpace(name), "@")
	if !ok {
		return nameTest{}, fmt.Errorf("unsupported test %q: only attribute tests (@name) are supported, not %q", whole, name)
	}
	return compileNameTest(rest, whole, s)
}

// parseTypeAlternative reads one xs:alternative. The test is kept as written
// and compiled during resolution, where the schema's prefixes are available.
func parseTypeAlternative(el *Element) (*TypeAlternative, error) {
	alt := &TypeAlternative{}
	alt.TypeName, _ = el.Attr("type")
	if test, ok := el.Attr("test"); ok {
		alt.testText = test
	}
	for _, child := range el.ChildElements() {
		if child.Namespace != xsdNS {
			continue
		}
		switch child.Local {
		case "complexType":
			ct, err := parseComplexType(child)
			if err != nil {
				return nil, err
			}
			alt.Type = ct
		case "simpleType":
			st, err := parseSimpleType(child)
			if err != nil {
				return nil, err
			}
			alt.Type = st
		case "annotation":
			// skip
		default:
			return nil, fmt.Errorf("unsupported schema element xs:%s inside xs:alternative", child.Local)
		}
	}
	if alt.Type == nil && alt.TypeName == "" {
		return nil, fmt.Errorf("xs:alternative requires a type attribute or an inline type")
	}
	return alt, nil
}

// resolveAlternatives compiles each test and resolves each named type. An
// alternative naming a type that does not exist would otherwise assign nothing
// and validate the element against no declaration at all.
func resolveAlternatives(ed *ElementDecl, s *Schema) error {
	for _, alt := range ed.Alternatives {
		if alt.Type == nil {
			local := stripPrefix(alt.TypeName)
			if t, ok := s.Types[local]; ok {
				alt.Type = t
			} else if bt := resolveBuiltinType(local); bt != nil {
				alt.Type = bt
			} else {
				return fmt.Errorf("element %q: xs:alternative type %q does not name a known type", ed.Name, alt.TypeName)
			}
		}
		if ct, ok := alt.Type.(*ComplexType); ok {
			if err := resolveComplexTypeRefs(ct, s, set.New[*ComplexType]()); err != nil {
				return err
			}
		}
		if st, ok := alt.Type.(*SimpleType); ok {
			if err := resolveSimpleTypeRefs(st, s, set.New[*SimpleType]()); err != nil {
				return err
			}
		}
		if alt.testText == "" {
			continue
		}
		test, err := compileTest(alt.testText, s)
		if err != nil {
			return fmt.Errorf("element %q: xs:alternative %w", ed.Name, err)
		}
		alt.test = test
	}
	return nil
}

// chooseType picks the type an instance element takes. The alternatives are
// tried in order; the declared type is the answer when none holds, which is
// what XSD says an element with no matching alternative falls back to.
func chooseType(el *Element, decl *ElementDecl) Type {
	for _, alt := range decl.Alternatives {
		if alt.test == nil || alt.test.eval(el) {
			return alt.Type
		}
	}
	return decl.Type
}
