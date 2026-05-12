package validator

import (
	"fmt"
	"strings"
)

type schemaValidator struct {
	schema *Schema
	errors []error
}

func ValidateSchema(doc *Document, schema *Schema) error {
	sv := &schemaValidator{schema: schema}
	sv.validateRoot(doc.Root)
	if len(sv.errors) > 0 {
		return sv.errors[0]
	}
	return nil
}

func (sv *schemaValidator) addError(el *Element, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	sv.errors = append(sv.errors, &Error{Line: el.Line, Col: el.Col, Message: msg})
}

func (sv *schemaValidator) validateRoot(el *Element) {
	rootName := el.Local
	decl, ok := sv.schema.Elements[rootName]
	if !ok {
		sv.addError(el, "element %q is not declared as a global element in the schema", rootName)
		return
	}
	resolveElementType(decl, sv.schema)
	sv.validateElement(el, decl)
}

func (sv *schemaValidator) validateElement(el *Element, decl *ElementDecl) {
	if decl.Fixed != "" && el.TextContent() != decl.Fixed {
		sv.addError(el, "element %q has fixed value %q but got %q", el.Local, decl.Fixed, el.TextContent())
		return
	}

	typ := decl.Type
	if typ == nil {
		return
	}

	switch t := typ.(type) {
	case *BuiltinType:
		sv.validateSimpleElement(el, t.name, nil)
	case *SimpleType:
		sv.validateSimpleTypeValue(el, t)
	case *ComplexType:
		sv.validateComplexElement(el, t)
	}
}

func (sv *schemaValidator) validateSimpleElement(el *Element, typeName string, facets []Facet) {
	children := el.ChildElements()
	if len(children) > 0 {
		sv.addError(el, "element %q has simple type %q but contains child elements", el.Local, typeName)
		return
	}
	value := el.TextContent()
	if err := validateBuiltinValue(typeName, value); err != nil {
		sv.addError(el, "element %q: %v", el.Local, err)
		return
	}
	if len(facets) > 0 {
		if err := validateEnumerationFacets(value, facets); err != nil {
			sv.addError(el, "element %q: %v", el.Local, err)
			return
		}
		if err := validateFacets(value, typeName, facets); err != nil {
			sv.addError(el, "element %q: %v", el.Local, err)
		}
	}
}

func (sv *schemaValidator) validateSimpleTypeValue(el *Element, st *SimpleType) {
	baseName := resolveSimpleTypeBaseName(st)
	if st.List != nil {
		sv.validateListType(el, st.List)
		return
	}
	if len(st.Union) > 0 {
		sv.validateUnionType(el, st.Union)
		return
	}
	sv.validateSimpleElement(el, baseName, st.Facets)
}

func (sv *schemaValidator) validateListType(el *Element, itemType *SimpleType) {
	value := strings.TrimSpace(el.TextContent())
	if value == "" {
		return
	}
	baseName := resolveSimpleTypeBaseName(itemType)
	for _, item := range strings.Fields(value) {
		if err := validateBuiltinValue(baseName, item); err != nil {
			sv.addError(el, "element %q list item: %v", el.Local, err)
			return
		}
	}
}

func (sv *schemaValidator) validateUnionType(el *Element, members []*SimpleType) {
	value := el.TextContent()
	for _, m := range members {
		baseName := resolveSimpleTypeBaseName(m)
		if err := validateBuiltinValue(baseName, value); err == nil {
			return
		}
	}
	sv.addError(el, "element %q: value %q does not match any member type of union", el.Local, value)
}

func (sv *schemaValidator) validateComplexElement(el *Element, ct *ComplexType) {
	sv.validateAttributes(el, ct.Attributes)

	if ct.SimpleText != nil {
		sv.validateSimpleText(el, ct.SimpleText)
		return
	}

	children := el.ChildElements()

	if ct.Content == nil {
		if len(children) > 0 {
			sv.addError(el, "element %q expects no child elements but got %d", el.Local, len(children))
		}
		if !ct.Mixed {
			text := strings.TrimSpace(el.TextContent())
			if text != "" {
				sv.addError(el, "element %q expects empty content but has text", el.Local)
			}
		}
		return
	}

	if !ct.Mixed {
		text := strings.TrimSpace(el.TextContent())
		if text != "" && len(children) > 0 {
			// mixed text and elements without mixed="true"
		}
	}

	sv.validateContentModel(el, children, ct.Content)
}

func (sv *schemaValidator) validateSimpleText(el *Element, textType Type) {
	children := el.ChildElements()
	if len(children) > 0 {
		sv.addError(el, "element %q has simpleContent but contains child elements", el.Local)
		return
	}
	value := el.TextContent()

	switch t := textType.(type) {
	case *BuiltinType:
		if err := t.validate(value); err != nil {
			sv.addError(el, "element %q: %v", el.Local, err)
		}
	case *SimpleType:
		baseName := resolveSimpleTypeBaseName(t)
		if err := validateBuiltinValue(baseName, value); err != nil {
			sv.addError(el, "element %q: %v", el.Local, err)
			return
		}
		if err := validateEnumerationFacets(value, t.Facets); err != nil {
			sv.addError(el, "element %q: %v", el.Local, err)
			return
		}
		if err := validateFacets(value, baseName, t.Facets); err != nil {
			sv.addError(el, "element %q: %v", el.Local, err)
		}
	}
}

func (sv *schemaValidator) validateAttributes(el *Element, decls []*AttrDecl) {
	declared := make(map[string]*AttrDecl)
	for _, ad := range decls {
		if ad.Use == "prohibited" {
			continue
		}
		declared[ad.Name] = ad
	}

	for _, attr := range el.Attrs {
		if attr.Prefix == "xmlns" || attr.Name == "xmlns" {
			continue
		}
		if strings.HasPrefix(attr.Name, "xml:") {
			continue
		}
		if attr.Namespace == xsiNS {
			continue
		}
		ad, ok := declared[attr.Local]
		if !ok {
			sv.addError(el, "unexpected attribute %q on element %q", attr.Local, el.Local)
			continue
		}
		if ad.Fixed != "" && attr.Value != ad.Fixed {
			sv.addError(el, "attribute %q on element %q must have fixed value %q", attr.Local, el.Local, ad.Fixed)
			continue
		}
		sv.validateAttrValue(el, attr, ad)
		delete(declared, attr.Local)
	}

	for name, ad := range declared {
		if ad.Use == "required" {
			sv.addError(el, "required attribute %q is missing on element %q", name, el.Local)
		}
	}
}

func (sv *schemaValidator) validateAttrValue(el *Element, attr Attr, ad *AttrDecl) {
	if ad.Type == nil {
		return
	}
	switch t := ad.Type.(type) {
	case *BuiltinType:
		if err := t.validate(attr.Value); err != nil {
			sv.addError(el, "attribute %q on element %q: %v", attr.Local, el.Local, err)
		}
	case *SimpleType:
		baseName := resolveSimpleTypeBaseName(t)
		if err := validateBuiltinValue(baseName, attr.Value); err != nil {
			sv.addError(el, "attribute %q on element %q: %v", attr.Local, el.Local, err)
			return
		}
		if err := validateEnumerationFacets(attr.Value, t.Facets); err != nil {
			sv.addError(el, "attribute %q on element %q: %v", attr.Local, el.Local, err)
		}
	}
}

func (sv *schemaValidator) validateContentModel(el *Element, children []*Element, cm ContentModel) {
	switch c := cm.(type) {
	case *Sequence:
		sv.validateSequence(el, children, c)
	case *Choice:
		sv.validateChoice(el, children, c)
	case *All:
		sv.validateAll(el, children, c)
	}
}

func (sv *schemaValidator) validateSequence(el *Element, children []*Element, seq *Sequence) {
	pos := 0
	for _, item := range seq.Items {
		consumed, err := sv.matchParticle(el, children[pos:], item)
		if err != nil {
			sv.errors = append(sv.errors, err)
			return
		}
		pos += consumed
	}
	if pos < len(children) {
		sv.addError(children[pos], "unexpected element %q in element %q", children[pos].Local, el.Local)
	}
}

func (sv *schemaValidator) validateChoice(el *Element, children []*Element, ch *Choice) {
	if len(children) == 0 {
		if ch.MinOccurs > 0 {
			var names []string
			for _, item := range ch.Items {
				if ed, ok := item.(*ElementDecl); ok {
					names = append(names, ed.Name)
				}
			}
			sv.addError(el, "element %q requires one of: %s", el.Local, strings.Join(names, ", "))
		}
		return
	}

	pos := 0
	iterations := 0
	for pos < len(children) {
		matched := false
		for _, item := range ch.Items {
			consumed, err := sv.matchParticle(el, children[pos:], item)
			if err == nil && consumed > 0 {
				pos += consumed
				matched = true
				iterations++
				break
			}
		}
		if !matched {
			if iterations < ch.MinOccurs {
				var names []string
				for _, item := range ch.Items {
					if ed, ok := item.(*ElementDecl); ok {
						names = append(names, ed.Name)
					}
				}
				sv.addError(children[pos], "element %q is not allowed here; expected one of: %s", children[pos].Local, strings.Join(names, ", "))
			} else if ch.MaxOccurs >= 0 && iterations >= ch.MaxOccurs {
				sv.addError(children[pos], "unexpected element %q in element %q (choice exceeded maxOccurs)", children[pos].Local, el.Local)
			} else {
				sv.addError(children[pos], "element %q is not allowed here", children[pos].Local)
			}
			return
		}
		if ch.MaxOccurs >= 0 && iterations > ch.MaxOccurs {
			sv.addError(children[pos-1], "choice in element %q exceeded maxOccurs %d", el.Local, ch.MaxOccurs)
			return
		}
	}
}

func (sv *schemaValidator) validateAll(el *Element, children []*Element, all *All) {
	seen := make(map[string]int)
	declMap := make(map[string]*ElementDecl)
	var anyParticles []*AnyParticle
	for _, item := range all.Items {
		if ed, ok := item.(*ElementDecl); ok {
			declMap[ed.Name] = ed
		}
		if ap, ok := item.(*AnyParticle); ok {
			anyParticles = append(anyParticles, ap)
		}
	}

	anyCount := make(map[*AnyParticle]int)

	for _, child := range children {
		decl, ok := declMap[child.Local]
		if !ok {
			matched := false
			for _, ap := range anyParticles {
				if sv.anyMatchesElement(ap, child) {
					anyCount[ap]++
					maxOccurs := ap.MaxOccurs
					if maxOccurs < 0 {
						maxOccurs = len(children)
					}
					if anyCount[ap] > maxOccurs {
						sv.addError(child, "xs:any wildcard exceeded maxOccurs %d", ap.MaxOccurs)
					}
					matched = true
					break
				}
			}
			if !matched {
				sv.addError(child, "unexpected element %q in all group of %q", child.Local, el.Local)
			}
			continue
		}
		seen[child.Local]++
		maxOccurs := decl.MaxOccurs
		if maxOccurs < 0 {
			maxOccurs = len(children)
		}
		if seen[child.Local] > maxOccurs {
			sv.addError(child, "element %q appears too many times (max %d)", child.Local, maxOccurs)
			continue
		}
		sv.validateElement(child, decl)
	}

	for _, item := range all.Items {
		if ed, ok := item.(*ElementDecl); ok {
			if seen[ed.Name] < ed.MinOccurs {
				sv.addError(el, "element %q requires at least %d occurrence(s) of %q, got %d",
					el.Local, ed.MinOccurs, ed.Name, seen[ed.Name])
			}
		}
	}
}

func (sv *schemaValidator) matchParticle(parent *Element, children []*Element, p Particle) (int, error) {
	switch item := p.(type) {
	case *ElementDecl:
		return sv.matchElement(parent, children, item)
	case *Sequence:
		return sv.matchSequence(parent, children, item)
	case *Choice:
		return sv.matchChoice(parent, children, item)
	case *AnyParticle:
		return sv.matchAny(children, item)
	}
	return 0, nil
}

func (sv *schemaValidator) matchElement(parent *Element, children []*Element, decl *ElementDecl) (int, error) {
	count := 0
	maxOccurs := decl.MaxOccurs
	if maxOccurs < 0 {
		maxOccurs = len(children) + 1
	}

	for count < len(children) && count < maxOccurs {
		child := children[count]
		if child.Local != decl.Name {
			break
		}
		sv.validateElement(child, decl)
		count++
	}

	if count < decl.MinOccurs {
		return count, &Error{
			Line:    parent.Line,
			Col:     parent.Col,
			Message: fmt.Sprintf("element %q requires at least %d occurrence(s) of %q, got %d", parent.Local, decl.MinOccurs, decl.Name, count),
		}
	}
	return count, nil
}

func (sv *schemaValidator) matchSequence(parent *Element, children []*Element, seq *Sequence) (int, error) {
	total := 0
	iterations := 0
	maxOccurs := seq.MaxOccurs
	if maxOccurs < 0 {
		maxOccurs = len(children) + 1
	}

	for iterations < maxOccurs {
		pos := 0
		remaining := children[total:]
		matched := true
		for _, item := range seq.Items {
			consumed, err := sv.matchParticle(parent, remaining[pos:], item)
			if err != nil {
				if iterations < seq.MinOccurs {
					return total, err
				}
				matched = false
				break
			}
			pos += consumed
		}
		if !matched || pos == 0 {
			break
		}
		total += pos
		iterations++
	}

	if iterations < seq.MinOccurs {
		return total, &Error{
			Line:    parent.Line,
			Col:     parent.Col,
			Message: fmt.Sprintf("sequence in element %q requires at least %d iteration(s), got %d", parent.Local, seq.MinOccurs, iterations),
		}
	}
	return total, nil
}

func (sv *schemaValidator) matchChoice(parent *Element, children []*Element, ch *Choice) (int, error) {
	total := 0
	iterations := 0
	maxOccurs := ch.MaxOccurs
	if maxOccurs < 0 {
		maxOccurs = len(children) + 1
	}

	for iterations < maxOccurs && total < len(children) {
		best := 0
		for _, item := range ch.Items {
			consumed, err := sv.matchParticle(parent, children[total:], item)
			if err == nil && consumed > best {
				best = consumed
			}
		}
		if best == 0 {
			break
		}
		total += best
		iterations++
	}

	if iterations < ch.MinOccurs {
		return total, &Error{
			Line:    parent.Line,
			Col:     parent.Col,
			Message: fmt.Sprintf("choice in element %q requires at least %d match(es), got %d", parent.Local, ch.MinOccurs, iterations),
		}
	}
	return total, nil
}

func (sv *schemaValidator) matchAny(children []*Element, ap *AnyParticle) (int, error) {
	count := 0
	maxOccurs := ap.MaxOccurs
	if maxOccurs < 0 {
		maxOccurs = len(children) + 1
	}
	for count < len(children) && count < maxOccurs {
		if !sv.anyMatchesElement(ap, children[count]) {
			break
		}
		count++
	}
	if count < ap.MinOccurs {
		return count, fmt.Errorf("xs:any requires at least %d element(s), got %d", ap.MinOccurs, count)
	}
	return count, nil
}

func (sv *schemaValidator) anyMatchesElement(ap *AnyParticle, el *Element) bool {
	ns := ap.Namespace
	if ns == "" || ns == "##any" {
		return true
	}
	if ns == "##local" {
		return el.Namespace == ""
	}
	if ns == "##other" {
		return el.Namespace != "" && el.Namespace != sv.schema.TargetNamespace
	}
	if ns == "##targetNamespace" {
		return el.Namespace == sv.schema.TargetNamespace
	}
	for _, allowed := range strings.Fields(ns) {
		if allowed == "##targetNamespace" {
			if el.Namespace == sv.schema.TargetNamespace {
				return true
			}
		} else if allowed == "##local" {
			if el.Namespace == "" {
				return true
			}
		} else if allowed == el.Namespace {
			return true
		}
	}
	return false
}

func resolveSimpleTypeBaseName(st *SimpleType) string {
	if st.BaseType != nil {
		if bt, ok := st.BaseType.(*BuiltinType); ok {
			return bt.name
		}
		if inner, ok := st.BaseType.(*SimpleType); ok {
			return resolveSimpleTypeBaseName(inner)
		}
	}
	if st.Base != "" {
		local := stripPrefix(st.Base)
		if resolveBuiltinType(local) != nil {
			return local
		}
	}
	return "string"
}
