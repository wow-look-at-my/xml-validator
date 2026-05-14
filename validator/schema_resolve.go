package validator

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
