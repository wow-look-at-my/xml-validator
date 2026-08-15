package validator

// resolving tracks the complex types already visited in one resolution pass.
// A recursive schema -- an element whose content refers back to itself, which
// is how a schema describes an arbitrarily nested tree -- would otherwise walk
// the same type forever. Resolution is idempotent per type, so visiting each
// pointer once is both a termination guarantee and the correct result.
type resolving map[*ComplexType]bool

func resolveSchemaRefs(s *Schema) {
	seen := resolving{}
	for _, ed := range s.Elements {
		resolveElementType(ed, s, seen)
	}
	for _, t := range s.Types {
		if ct, ok := t.(*ComplexType); ok {
			resolveComplexTypeRefs(ct, s, seen)
		}
	}
}

func resolveElementType(ed *ElementDecl, s *Schema, seen resolving) {
	if ed.Type != nil {
		if ct, ok := ed.Type.(*ComplexType); ok {
			resolveComplexTypeRefs(ct, s, seen)
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

func resolveComplexTypeRefs(ct *ComplexType, s *Schema, seen resolving) {
	if seen[ct] {
		return
	}
	seen[ct] = true
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
		resolveContentModel(ct.Content, s, seen)
	}
	for _, ad := range ct.Attributes {
		resolveAttrRef(ad, s)
		resolveAttrType(ad, s)
	}
	if ct.SimpleText != nil {
		if st, ok := ct.SimpleText.(*SimpleType); ok {
			resolveSimpleTypeBase(st, s)
		}
	}
}

func resolveContentModel(cm ContentModel, s *Schema, seen resolving) {
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
					resolveElementType(p, s, seen)
				}
			} else {
				resolveElementType(p, s, seen)
			}
		case *Sequence:
			resolveContentModel(p, s, seen)
		case *Choice:
			resolveContentModel(p, s, seen)
		case *All:
			resolveContentModel(p, s, seen)
		}
	}
}

// resolveAttrRef fills a use="..." reference from the global attribute it names.
// The reference contributes only whether the attribute is required: the name,
// namespace, and type are the global declaration's, which is what makes a
// qualified attribute in an instance document resolvable at all.
func resolveAttrRef(ad *AttrDecl, s *Schema) {
	if ad.Ref == "" {
		return
	}
	g, ok := s.Attributes[stripPrefix(ad.Ref)]
	if !ok {
		return
	}
	ad.Name = g.Name
	ad.Namespace = g.Namespace
	if ad.TypeName == "" {
		ad.TypeName = g.TypeName
	}
	if ad.Type == nil {
		ad.Type = g.Type
	}
	if ad.Fixed == "" {
		ad.Fixed = g.Fixed
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
