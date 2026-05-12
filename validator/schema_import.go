package validator

import (
	"bytes"
	"fmt"
)

// SchemaResolver loads the bytes of a referenced schema. It is invoked once
// per xs:import directive that has a schemaLocation. The namespace argument
// is the target namespace of the imported schema (may be empty), and
// schemaLocation is the URI hint from the directive. Returning a nil byte
// slice with a nil error is treated as "skip this import".
type SchemaResolver func(namespace, schemaLocation string) ([]byte, error)

type importResult struct {
	directive *Import
	imported  *Schema
}

func parseImport(el *Element, resolver SchemaResolver, visited map[string]bool) (*importResult, error) {
	ns, _ := el.Attr("namespace")
	loc, _ := el.Attr("schemaLocation")
	directive := &Import{Namespace: ns, SchemaLocation: loc}

	if loc == "" {
		return &importResult{directive: directive}, nil
	}
	if resolver == nil {
		return nil, fmt.Errorf("xs:import schemaLocation %q requires a schema resolver", loc)
	}
	key := ns + "|" + loc
	if visited[key] {
		return &importResult{directive: directive}, nil
	}
	visited[key] = true

	data, err := resolver(ns, loc)
	if err != nil {
		return nil, fmt.Errorf("resolving xs:import %q: %w", loc, err)
	}
	if len(data) == 0 {
		return &importResult{directive: directive}, nil
	}
	doc, err := ParseTree(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("parsing imported schema %q: %w", loc, err)
	}
	imported, err := parseSchemaDoc(doc, resolver, visited)
	if err != nil {
		return nil, fmt.Errorf("parsing imported schema %q: %w", loc, err)
	}
	return &importResult{directive: directive, imported: imported}, nil
}

func mergeImportedSchema(dst, src *Schema) {
	for name, ed := range src.Elements {
		if _, exists := dst.Elements[name]; !exists {
			dst.Elements[name] = ed
		}
	}
	for name, t := range src.Types {
		if _, exists := dst.Types[name]; !exists {
			dst.Types[name] = t
		}
	}
	for name, g := range src.Groups {
		if _, exists := dst.Groups[name]; !exists {
			dst.Groups[name] = g
		}
	}
	for name, ag := range src.AttrGroups {
		if _, exists := dst.AttrGroups[name]; !exists {
			dst.AttrGroups[name] = ag
		}
	}
	dst.Imports = append(dst.Imports, src.Imports...)
}
