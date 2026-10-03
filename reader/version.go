package reader

import "strings"

// Version is the XML version a document declares.
type Version string

// The versions this module reads. A document with no XML declaration is
// Version10, because Version11 requires the declaration.
const (
	Version10 Version = "1.0"
	Version11 Version = "1.1"
)

// SniffVersion reads the version out of the raw XML declaration, before
// anything decodes the bytes, because the version decides the line ending
// rules. A declaration it cannot read answers Version10, so the declaration
// parser reports the syntax error at the right position.
func SniffVersion(raw []byte) Version {
	const window = 256
	head := raw
	if len(head) > window {
		head = head[:window]
	}
	decl := string(head)
	if !strings.HasPrefix(decl, "<?xml") {
		return Version10
	}
	if end := strings.Index(decl, "?>"); end >= 0 {
		decl = decl[:end]
	}
	rest := strings.TrimLeft(strings.TrimPrefix(decl, "<?xml"), " \t\r\n")
	rest, ok := strings.CutPrefix(rest, "version")
	if !ok {
		return Version10
	}
	rest = strings.TrimLeft(rest, " \t\r\n")
	rest, ok = strings.CutPrefix(rest, "=")
	if !ok {
		return Version10
	}
	rest = strings.TrimLeft(rest, " \t\r\n")
	for _, quote := range []string{`"`, `'`} {
		if strings.HasPrefix(rest, quote+string(Version11)+quote) {
			return Version11
		}
	}
	return Version10
}
