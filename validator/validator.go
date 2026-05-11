package validator

import "io"

func Validate(r io.Reader) error {
	runes, enc, err := readInput(r)
	if err != nil {
		return err
	}
	p := newParser(runes, enc)
	return p.parseDocument()
}
