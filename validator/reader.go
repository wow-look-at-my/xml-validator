package validator

import (
	"fmt"
	"io"
	"unicode/utf8"
)

func readInput(r io.Reader) ([]rune, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading input: %w", err)
	}
	if len(raw) == 0 {
		return nil, fmt.Errorf("empty input")
	}

	if err := rejectNonUTF8(raw); err != nil {
		return nil, err
	}

	runes, err := decodeUTF8(raw)
	if err != nil {
		return nil, err
	}

	return normalizeLineEndings(runes), nil
}

// rejectNonUTF8 rejects any input that begins with a byte-order mark or that
// matches the UTF-16 leading-NUL heuristic from XML 1.1 appendix F. The
// validator requires raw UTF-8 with no BOM (per the utf8everywhere
// recommendation); the BOM is meaningless for UTF-8 and is interpreted by
// some downstream tools as a literal U+FEFF.
func rejectNonUTF8(data []byte) error {
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		return fmt.Errorf("unsupported encoding: input begins with a UTF-8 BOM (raw UTF-8 only)")
	}
	if len(data) >= 2 {
		if data[0] == 0xFE && data[1] == 0xFF {
			return fmt.Errorf("unsupported encoding: UTF-16 BE (only UTF-8 is supported)")
		}
		if data[0] == 0xFF && data[1] == 0xFE {
			return fmt.Errorf("unsupported encoding: UTF-16 LE (only UTF-8 is supported)")
		}
		if data[0] == 0x00 && data[1] == 0x3C {
			return fmt.Errorf("unsupported encoding: input looks like UTF-16 BE (only UTF-8 is supported)")
		}
		if data[0] == 0x3C && data[1] == 0x00 {
			return fmt.Errorf("unsupported encoding: input looks like UTF-16 LE (only UTF-8 is supported)")
		}
	}
	return nil
}

func decodeUTF8(data []byte) ([]rune, error) {
	runes := make([]rune, 0, len(data))
	for len(data) > 0 {
		r, size := utf8.DecodeRune(data)
		if r == utf8.RuneError && size <= 1 {
			return nil, fmt.Errorf("invalid UTF-8 byte sequence")
		}
		runes = append(runes, r)
		data = data[size:]
	}
	return runes, nil
}

// normalizeLineEndings applies XML 1.1 line ending normalization:
//
//	#xD #xA  -> #xA
//	#xD #x85 -> #xA
//	#x85     -> #xA
//	#x2028   -> #xA
//	#xD      -> #xA (when not followed by #xA or #x85)
func normalizeLineEndings(input []rune) []rune {
	out := make([]rune, 0, len(input))
	for i := 0; i < len(input); i++ {
		r := input[i]
		switch {
		case r == 0xD:
			out = append(out, 0xA)
			if i+1 < len(input) && (input[i+1] == 0xA || input[i+1] == 0x85) {
				i++
			}
		case r == 0x85, r == 0x2028:
			out = append(out, 0xA)
		default:
			out = append(out, r)
		}
	}
	return out
}
