package validator

type Document struct {
	Root *Element
}

type Node interface{ node() }

type Element struct {
	Name      string
	Namespace string
	Local     string
	Prefix    string
	Attrs     []Attr
	Children  []Node
	Line      int
	Col       int
}

type Attr struct {
	Name      string
	Namespace string
	Local     string
	Prefix    string
	Value     string
}

type CharData struct {
	Content string
}

func (*Element) node()  {}
func (*CharData) node() {}

func (e *Element) ChildElements() []*Element {
	var elems []*Element
	for _, c := range e.Children {
		if el, ok := c.(*Element); ok {
			elems = append(elems, el)
		}
	}
	return elems
}

func (e *Element) TextContent() string {
	var b []byte
	for _, c := range e.Children {
		if t, ok := c.(*CharData); ok {
			b = append(b, t.Content...)
		}
	}
	return string(b)
}

func (e *Element) Attr(local string) (string, bool) {
	for _, a := range e.Attrs {
		if a.Local == local && a.Namespace == "" {
			return a.Value, true
		}
	}
	return "", false
}
