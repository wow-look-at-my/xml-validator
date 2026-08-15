package validator

import "fmt"

// resolving tracks the complex types already visited in one resolution pass.
// A recursive schema -- an element whose content refers back to itself, which
// is how a schema describes an arbitrarily nested tree -- would otherwise walk
// the same type forever. Resolution is idempotent per type, so visiting each
// pointer once is both a termination guarantee and the correct result.
type resolving map[*ComplexType]bool

func resolveSchemaRefs(s *Schema) error {
	seen := resolving{}
	for _, ed := range s.Elements {
		if err := resolveElementType(ed, s, seen); err != nil {
			return err
		}
	}
	for _, t := range s.Types {
		if ct, ok := t.(*ComplexType); ok {
			if err := resolveComplexTypeRefs(ct, s, seen); err != nil {
				return err
			}
		}
	}
	return nil
}

func resolveElementType(ed *ElementDecl, s *Schema, seen resolving) error {
	if ed.Type != nil {
		if ct, ok := ed.Type.(*ComplexType); ok {
			return resolveComplexTypeRefs(ct, s, seen)
		}
		return nil
	}
	if ed.TypeName != "" {
		local := stripPrefix(ed.TypeName)
		if t, ok := s.Types[local]; ok {
			ed.Type = t
		} else if bt := resolveBuiltinType(local); bt != nil {
			ed.Type = bt
		}
	}
	return nil
}

func resolveComplexTypeRefs(ct *ComplexType, s *Schema, seen resolving) error {
	if seen[ct] {
		return nil
	}
	seen[ct] = true
	if err := resolveDerivation(ct, s, seen); err != nil {
		return err
	}
	if ct.Content != nil {
		if err := expandGroupRefs(ct.Content, s, map[string]bool{}); err != nil {
			return err
		}
	}
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
		if err := resolveContentModel(ct.Content, s, seen); err != nil {
			return err
		}
	}
	for _, ad := range ct.Attributes {
		if err := resolveAttrRef(ad, s); err != nil {
			return err
		}
		resolveAttrType(ad, s)
	}
	if ct.SimpleText != nil {
		if st, ok := ct.SimpleText.(*SimpleType); ok {
			resolveSimpleTypeBase(st, s)
		}
	}
	return nil
}

func resolveContentModel(cm ContentModel, s *Schema, seen resolving) error {
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
		var err error
		switch p := item.(type) {
		case *ElementDecl:
			if p.Ref != "" {
				ref := s.lookupElement(p.Ref)
				if ref == nil {
					return fmt.Errorf("element ref %q does not name a global element declaration", p.Ref)
				}
				p.Name = ref.Name
				p.Namespace = ref.Namespace
				p.TypeName = ref.TypeName
				p.Type = ref.Type
			}
			err = resolveElementType(p, s, seen)
		case *Sequence:
			err = resolveContentModel(p, s, seen)
		case *Choice:
			err = resolveContentModel(p, s, seen)
		case *All:
			err = resolveContentModel(p, s, seen)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// resolveAttrRef fills a use="..." reference from the global attribute it names.
// The reference contributes only whether the attribute is required: the name,
// namespace, and type are the global declaration's, which is what makes a
// qualified attribute in an instance document resolvable at all.
func resolveAttrRef(ad *AttrDecl, s *Schema) error {
	if ad.Ref == "" {
		return nil
	}
	g := s.lookupAttribute(ad.Ref)
	if g == nil {
		return fmt.Errorf("attribute ref %q does not name a global attribute declaration", ad.Ref)
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
	return nil
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
