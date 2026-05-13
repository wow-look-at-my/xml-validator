package validator

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseSchema parses an XSD schema tree with no support for resolving
// xs:import schemaLocation hints. Imports without a schemaLocation are
// accepted as namespace declarations; imports that name a schemaLocation
// produce an error. Callers that need to follow schemaLocation hints should
// use ParseSchemaWithResolver instead.
func ParseSchema(doc *Document) (*Schema, error) {
	return ParseSchemaWithResolver(doc, nil)
}

// ParseSchemaWithResolver parses an XSD schema tree. Each xs:import or
// xs:include directive with a non-empty schemaLocation is loaded via resolver,
// parsed, and merged into the returned schema. Components from
// imported/included schemas are looked up by local name (the same way the
// validator resolves types in the main schema), so per-namespace name
// collisions are not supported.
func ParseSchemaWithResolver(doc *Document, resolver SchemaResolver) (*Schema, error) {
	visited := make(map[importKey]bool)
	return parseSchemaDoc(doc, resolver, visited)
}

func parseSchemaDoc(doc *Document, resolver SchemaResolver, visited map[importKey]bool) (*Schema, error) {
	root := doc.Root
	if root.Local != "schema" || root.Namespace != xsdNS {
		return nil, fmt.Errorf("expected xs:schema root element, got {%s}%s", root.Namespace, root.Local)
	}

	s := &Schema{
		Elements:   make(map[string]*ElementDecl),
		Types:      make(map[string]Type),
		Groups:     make(map[string]*Group),
		AttrGroups: make(map[string]*AttrGroup),
	}

	s.TargetNamespace, _ = root.Attr("targetNamespace")
	s.ElementFormDefault, _ = root.Attr("elementFormDefault")
	if s.ElementFormDefault == "" {
		s.ElementFormDefault = "unqualified"
	}
	s.AttributeFormDefault, _ = root.Attr("attributeFormDefault")
	if s.AttributeFormDefault == "" {
		s.AttributeFormDefault = "unqualified"
	}

	for _, child := range root.ChildElements() {
		if child.Namespace != xsdNS {
			continue
		}
		switch child.Local {
		case "element":
			ed, err := parseElementDecl(child)
			if err != nil {
				return nil, fmt.Errorf("parsing element declaration: %w", err)
			}
			if ed.Name != "" {
				s.Elements[ed.Name] = ed
			}
		case "complexType":
			ct, err := parseComplexType(child)
			if err != nil {
				return nil, fmt.Errorf("parsing complex type: %w", err)
			}
			if ct.Name != "" {
				s.Types[ct.Name] = ct
			}
		case "simpleType":
			st, err := parseSimpleType(child)
			if err != nil {
				return nil, fmt.Errorf("parsing simple type: %w", err)
			}
			if st.Name != "" {
				s.Types[st.Name] = st
			}
		case "group":
			g, err := parseGroup(child)
			if err != nil {
				return nil, fmt.Errorf("parsing group: %w", err)
			}
			if g.Name != "" {
				s.Groups[g.Name] = g
			}
		case "attributeGroup":
			ag, err := parseAttrGroup(child)
			if err != nil {
				return nil, fmt.Errorf("parsing attribute group: %w", err)
			}
			if ag.Name != "" {
				s.AttrGroups[ag.Name] = ag
			}
		case "import":
			imp, err := parseImport(child, resolver, visited)
			if err != nil {
				return nil, err
			}
			s.Imports = append(s.Imports, imp.directive)
			if imp.imported != nil {
				if err := mergeImportedSchema(s, imp.imported); err != nil {
					return nil, err
				}
			}
		case "include":
			inc, err := parseInclude(child, resolver, visited)
			if err != nil {
				return nil, err
			}
			if inc.imported != nil {
				if err := mergeImportedSchema(s, inc.imported); err != nil {
					return nil, err
				}
			}
		case "redefine", "override":
			return nil, fmt.Errorf("unsupported: xs:%s is not supported", child.Local)
		case "notation":
			return nil, fmt.Errorf("unsupported: xs:notation is not supported")
		case "annotation":
			// skip
		default:
			return nil, fmt.Errorf("unsupported schema element xs:%s", child.Local)
		}
	}

	resolveSchemaRefs(s)
	return s, nil
}

func parseElementDecl(el *Element) (*ElementDecl, error) {
	ed := &ElementDecl{
		MinOccurs: 1,
		MaxOccurs: 1,
	}
	ed.Name, _ = el.Attr("name")
	ed.TypeName, _ = el.Attr("type")
	ed.Ref, _ = el.Attr("ref")
	ed.Default, _ = el.Attr("default")
	ed.Fixed, _ = el.Attr("fixed")

	if v, ok := el.Attr("nillable"); ok && v == "true" {
		ed.Nillable = true
	}
	if v, ok := el.Attr("minOccurs"); ok {
		n, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid minOccurs %q", v)
		}
		ed.MinOccurs = n
	}
	if v, ok := el.Attr("maxOccurs"); ok {
		if v == "unbounded" {
			ed.MaxOccurs = -1
		} else {
			n, err := strconv.Atoi(v)
			if err != nil {
				return nil, fmt.Errorf("invalid maxOccurs %q", v)
			}
			ed.MaxOccurs = n
		}
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
			ed.Type = ct
		case "simpleType":
			st, err := parseSimpleType(child)
			if err != nil {
				return nil, err
			}
			ed.Type = st
		case "annotation":
			// skip
		}
	}

	return ed, nil
}

func parseComplexType(el *Element) (*ComplexType, error) {
	ct := &ComplexType{}
	ct.Name, _ = el.Attr("name")
	if v, ok := el.Attr("mixed"); ok && v == "true" {
		ct.Mixed = true
	}

	for _, child := range el.ChildElements() {
		if child.Namespace != xsdNS {
			continue
		}
		switch child.Local {
		case "sequence":
			seq, err := parseSequence(child)
			if err != nil {
				return nil, err
			}
			ct.Content = seq
		case "choice":
			ch, err := parseChoice(child)
			if err != nil {
				return nil, err
			}
			ct.Content = ch
		case "all":
			a, err := parseAll(child)
			if err != nil {
				return nil, err
			}
			ct.Content = a
		case "attribute":
			ad, err := parseAttrDecl(child)
			if err != nil {
				return nil, err
			}
			ct.Attributes = append(ct.Attributes, ad)
		case "attributeGroup":
			if ref, ok := child.Attr("ref"); ok {
				ct.attrGroupRefs = append(ct.attrGroupRefs, stripPrefix(ref))
			}
		case "simpleContent":
			if err := parseSimpleContent(child, ct); err != nil {
				return nil, err
			}
		case "complexContent":
			if err := parseComplexContent(child, ct); err != nil {
				return nil, err
			}
		case "anyAttribute":
			ct.AnyAttribute = parseAnyAttrDecl(child)
		case "annotation":
			// skip
		}
	}

	return ct, nil
}

func parseSimpleContent(el *Element, ct *ComplexType) error {
	for _, child := range el.ChildElements() {
		if child.Namespace != xsdNS {
			continue
		}
		switch child.Local {
		case "extension":
			base, _ := child.Attr("base")
			ct.SimpleText = resolveTypeByName(base)
			for _, attr := range child.ChildElements() {
				if attr.Namespace != xsdNS {
					continue
				}
				if attr.Local == "attribute" {
					ad, err := parseAttrDecl(attr)
					if err != nil {
						return err
					}
					ct.Attributes = append(ct.Attributes, ad)
				} else if attr.Local == "anyAttribute" {
					ct.AnyAttribute = parseAnyAttrDecl(attr)
				} else if attr.Local == "attributeGroup" {
					if ref, ok := attr.Attr("ref"); ok {
						ct.attrGroupRefs = append(ct.attrGroupRefs, stripPrefix(ref))
					}
				}
			}
		case "restriction":
			base, _ := child.Attr("base")
			st := &SimpleType{Base: base}
			for _, facetEl := range child.ChildElements() {
				if facetEl.Namespace != xsdNS {
					continue
				}
				if isFacetElement(facetEl.Local) {
					val, _ := facetEl.Attr("value")
					st.Facets = append(st.Facets, Facet{Kind: facetEl.Local, Value: val})
				} else if facetEl.Local == "attribute" {
					ad, err := parseAttrDecl(facetEl)
					if err != nil {
						return err
					}
					ct.Attributes = append(ct.Attributes, ad)
				} else if facetEl.Local == "anyAttribute" {
					ct.AnyAttribute = parseAnyAttrDecl(facetEl)
				} else if facetEl.Local == "attributeGroup" {
					if ref, ok := facetEl.Attr("ref"); ok {
						ct.attrGroupRefs = append(ct.attrGroupRefs, stripPrefix(ref))
					}
				}
			}
			ct.SimpleText = st
		}
	}
	return nil
}

func parseComplexContent(el *Element, ct *ComplexType) error {
	if v, ok := el.Attr("mixed"); ok && v == "true" {
		ct.Mixed = true
	}
	for _, child := range el.ChildElements() {
		if child.Namespace != xsdNS {
			continue
		}
		switch child.Local {
		case "extension", "restriction":
			for _, inner := range child.ChildElements() {
				if inner.Namespace != xsdNS {
					continue
				}
				switch inner.Local {
				case "sequence":
					seq, err := parseSequence(inner)
					if err != nil {
						return err
					}
					ct.Content = seq
				case "choice":
					ch, err := parseChoice(inner)
					if err != nil {
						return err
					}
					ct.Content = ch
				case "all":
					a, err := parseAll(inner)
					if err != nil {
						return err
					}
					ct.Content = a
				case "attribute":
					ad, err := parseAttrDecl(inner)
					if err != nil {
						return err
					}
					ct.Attributes = append(ct.Attributes, ad)
				case "anyAttribute":
					ct.AnyAttribute = parseAnyAttrDecl(inner)
				case "attributeGroup":
					if ref, ok := inner.Attr("ref"); ok {
						ct.attrGroupRefs = append(ct.attrGroupRefs, stripPrefix(ref))
					}
				}
			}
		}
	}
	return nil
}

func parseSequence(el *Element) (*Sequence, error) {
	seq := &Sequence{MinOccurs: 1, MaxOccurs: 1}
	parseOccurs(el, &seq.MinOccurs, &seq.MaxOccurs)
	items, err := parseParticles(el)
	if err != nil {
		return nil, err
	}
	seq.Items = items
	return seq, nil
}

func parseChoice(el *Element) (*Choice, error) {
	ch := &Choice{MinOccurs: 1, MaxOccurs: 1}
	parseOccurs(el, &ch.MinOccurs, &ch.MaxOccurs)
	items, err := parseParticles(el)
	if err != nil {
		return nil, err
	}
	ch.Items = items
	return ch, nil
}

func parseAll(el *Element) (*All, error) {
	a := &All{MinOccurs: 1, MaxOccurs: 1}
	parseOccurs(el, &a.MinOccurs, &a.MaxOccurs)
	items, err := parseParticles(el)
	if err != nil {
		return nil, err
	}
	a.Items = items
	return a, nil
}

func parseParticles(el *Element) ([]Particle, error) {
	var items []Particle
	for _, child := range el.ChildElements() {
		if child.Namespace != xsdNS {
			continue
		}
		switch child.Local {
		case "element":
			ed, err := parseElementDecl(child)
			if err != nil {
				return nil, err
			}
			items = append(items, ed)
		case "sequence":
			s, err := parseSequence(child)
			if err != nil {
				return nil, err
			}
			items = append(items, s)
		case "choice":
			c, err := parseChoice(child)
			if err != nil {
				return nil, err
			}
			items = append(items, c)
		case "all":
			a, err := parseAll(child)
			if err != nil {
				return nil, err
			}
			items = append(items, a)
		case "group":
			// group reference, will be resolved
		case "any":
			ap := &AnyParticle{MinOccurs: 1, MaxOccurs: 1}
			parseOccurs(child, &ap.MinOccurs, &ap.MaxOccurs)
			ap.Namespace, _ = child.Attr("namespace")
			ap.ProcessContents, _ = child.Attr("processContents")
			if ap.ProcessContents == "" {
				ap.ProcessContents = "strict"
			}
			items = append(items, ap)
		case "annotation":
			// skip
		}
	}
	return items, nil
}

func parseAttrDecl(el *Element) (*AttrDecl, error) {
	ad := &AttrDecl{Use: "optional"}
	ad.Name, _ = el.Attr("name")
	ad.TypeName, _ = el.Attr("type")
	ad.Ref, _ = el.Attr("ref")
	ad.Default, _ = el.Attr("default")
	ad.Fixed, _ = el.Attr("fixed")
	if v, ok := el.Attr("use"); ok {
		ad.Use = v
	}

	for _, child := range el.ChildElements() {
		if child.Namespace == xsdNS && child.Local == "simpleType" {
			st, err := parseSimpleType(child)
			if err != nil {
				return nil, err
			}
			ad.Type = st
		}
	}

	return ad, nil
}

func parseSimpleType(el *Element) (*SimpleType, error) {
	st := &SimpleType{}
	st.Name, _ = el.Attr("name")

	for _, child := range el.ChildElements() {
		if child.Namespace != xsdNS {
			continue
		}
		switch child.Local {
		case "restriction":
			st.Base, _ = child.Attr("base")
			for _, facetEl := range child.ChildElements() {
				if facetEl.Namespace != xsdNS {
					continue
				}
				if isFacetElement(facetEl.Local) {
					val, _ := facetEl.Attr("value")
					st.Facets = append(st.Facets, Facet{Kind: facetEl.Local, Value: val})
				}
			}
		case "list":
			itemType, _ := child.Attr("itemType")
			st.List = &SimpleType{Base: itemType}
		case "union":
			memberTypes, _ := child.Attr("memberTypes")
			for _, mt := range strings.Fields(memberTypes) {
				st.Union = append(st.Union, &SimpleType{Base: mt})
			}
		case "annotation":
			// skip
		}
	}

	return st, nil
}

func parseGroup(el *Element) (*Group, error) {
	g := &Group{}
	g.Name, _ = el.Attr("name")
	for _, child := range el.ChildElements() {
		if child.Namespace != xsdNS {
			continue
		}
		switch child.Local {
		case "sequence":
			seq, err := parseSequence(child)
			if err != nil {
				return nil, err
			}
			g.Content = seq
		case "choice":
			ch, err := parseChoice(child)
			if err != nil {
				return nil, err
			}
			g.Content = ch
		case "all":
			a, err := parseAll(child)
			if err != nil {
				return nil, err
			}
			g.Content = a
		}
	}
	return g, nil
}

func parseAttrGroup(el *Element) (*AttrGroup, error) {
	ag := &AttrGroup{}
	ag.Name, _ = el.Attr("name")
	for _, child := range el.ChildElements() {
		if child.Namespace != xsdNS {
			continue
		}
		switch child.Local {
		case "attribute":
			ad, err := parseAttrDecl(child)
			if err != nil {
				return nil, err
			}
			ag.Attributes = append(ag.Attributes, ad)
		case "anyAttribute":
			ag.AnyAttribute = parseAnyAttrDecl(child)
		}
	}
	return ag, nil
}

func parseAnyAttrDecl(el *Element) *AnyAttrDecl {
	aa := &AnyAttrDecl{}
	aa.Namespace, _ = el.Attr("namespace")
	aa.ProcessContents, _ = el.Attr("processContents")
	if aa.ProcessContents == "" {
		aa.ProcessContents = "strict"
	}
	return aa
}

func parseOccurs(el *Element, min, max *int) {
	if v, ok := el.Attr("minOccurs"); ok {
		n, err := strconv.Atoi(v)
		if err == nil {
			*min = n
		}
	}
	if v, ok := el.Attr("maxOccurs"); ok {
		if v == "unbounded" {
			*max = -1
		} else {
			n, err := strconv.Atoi(v)
			if err == nil {
				*max = n
			}
		}
	}
}

func isFacetElement(local string) bool {
	switch local {
	case "enumeration", "pattern", "minLength", "maxLength", "length",
		"minInclusive", "maxInclusive", "minExclusive", "maxExclusive",
		"totalDigits", "fractionDigits", "whiteSpace":
		return true
	}
	return false
}

func resolveTypeByName(name string) Type {
	local := stripPrefix(name)
	if bt := resolveBuiltinType(local); bt != nil {
		return bt
	}
	return &SimpleType{Base: local}
}

func stripPrefix(name string) string {
	if idx := strings.Index(name, ":"); idx >= 0 {
		return name[idx+1:]
	}
	return name
}

func resolveSchemaRefs(s *Schema) {
	for _, ed := range s.Elements {
		resolveElementType(ed, s)
	}
	for _, t := range s.Types {
		if ct, ok := t.(*ComplexType); ok {
			resolveComplexTypeRefs(ct, s)
		}
	}
}

func resolveElementType(ed *ElementDecl, s *Schema) {
	if ed.Type != nil {
		if ct, ok := ed.Type.(*ComplexType); ok {
			resolveComplexTypeRefs(ct, s)
		}
		return
	}
	if ed.TypeName != "" {
		local := stripPrefix(ed.TypeName)
		if t, ok := s.Types[local]; ok {
			ed.Type = t
		} else if bt := resolveBuiltinType(local); bt != nil {
			ed.Type = bt
		}
	}
}

func resolveComplexTypeRefs(ct *ComplexType, s *Schema) {
	for _, ref := range ct.attrGroupRefs {
		if ag, ok := s.AttrGroups[ref]; ok {
			ct.Attributes = append(ct.Attributes, ag.Attributes...)
			if ag.AnyAttribute != nil && ct.AnyAttribute == nil {
				ct.AnyAttribute = ag.AnyAttribute
			}
		}
	}
	ct.attrGroupRefs = nil
	if ct.Content != nil {
		resolveContentModel(ct.Content, s)
	}
	for _, ad := range ct.Attributes {
		resolveAttrType(ad, s)
	}
	if ct.SimpleText != nil {
		if st, ok := ct.SimpleText.(*SimpleType); ok {
			resolveSimpleTypeBase(st, s)
		}
	}
}

func resolveContentModel(cm ContentModel, s *Schema) {
	var items []Particle
	switch c := cm.(type) {
	case *Sequence:
		items = c.Items
	case *Choice:
		items = c.Items
	case *All:
		items = c.Items
	}
	for _, item := range items {
		switch p := item.(type) {
		case *ElementDecl:
			if p.Ref != "" {
				local := stripPrefix(p.Ref)
				if ref, ok := s.Elements[local]; ok {
					p.Name = ref.Name
					p.TypeName = ref.TypeName
					p.Type = ref.Type
					resolveElementType(p, s)
				}
			} else {
				resolveElementType(p, s)
			}
		case *Sequence:
			resolveContentModel(p, s)
		case *Choice:
			resolveContentModel(p, s)
		case *All:
			resolveContentModel(p, s)
		}
	}
}

func resolveAttrType(ad *AttrDecl, s *Schema) {
	if ad.Type != nil {
		return
	}
	if ad.TypeName != "" {
		local := stripPrefix(ad.TypeName)
		if t, ok := s.Types[local]; ok {
			ad.Type = t
		} else if bt := resolveBuiltinType(local); bt != nil {
			ad.Type = bt
		}
	}
}

func resolveSimpleTypeBase(st *SimpleType, s *Schema) {
	if st.Base != "" {
		local := stripPrefix(st.Base)
		if bt := resolveBuiltinType(local); bt != nil {
			st.BaseType = bt
		} else if t, ok := s.Types[local]; ok {
			st.BaseType = t
		}
	}
}
