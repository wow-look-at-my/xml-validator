package validator

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"unicode/utf16"
	"unicode/utf8"
)

type encoding int

const (
	encUTF8 encoding = iota
	encUTF16BE
	encUTF16LE
)

func readInput(r io.Reader) ([]rune, encoding, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, 0, fmt.Errorf("reading input: %w", err)
	}
	if len(raw) == 0 {
		return nil, 0, fmt.Errorf("empty input")
	}

	enc, raw := detectEncoding(raw)
	runes, err := decode(raw, enc)
	if err != nil {
		return nil, 0, err
	}

	runes = normalizeLineEndings(runes)
	return runes, enc, nil
}

func detectEncoding(data []byte) (encoding, []byte) {
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		return encUTF8, data[3:]
	}
	if len(data) >= 2 && data[0] == 0xFE && data[1] == 0xFF {
		return encUTF16BE, data[2:]
	}
	if len(data) >= 2 && data[0] == 0xFF && data[1] == 0xFE {
		return encUTF16LE, data[2:]
	}
	if len(data) >= 2 {
		if data[0] == 0x00 && data[1] == 0x3C {
			return encUTF16BE, data
		}
		if data[0] == 0x3C && data[1] == 0x00 {
			return encUTF16LE, data
		}
	}
	return encUTF8, data
}

func decode(data []byte, enc encoding) ([]rune, error) {
	switch enc {
	case encUTF8:
		return decodeUTF8(data)
	case encUTF16BE:
		return decodeUTF16(data, binary.BigEndian)
	case encUTF16LE:
		return decodeUTF16(data, binary.LittleEndian)
	default:
		return nil, fmt.Errorf("unsupported encoding")
	}
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

func decodeUTF16(data []byte, order binary.ByteOrder) ([]rune, error) {
	if len(data)%2 != 0 {
		return nil, fmt.Errorf("invalid UTF-16: odd number of bytes")
	}

	reader := bytes.NewReader(data)
	units := make([]uint16, 0, len(data)/2)
	for reader.Len() >= 2 {
		var u uint16
		if err := binary.Read(reader, order, &u); err != nil {
			break
		}
		units = append(units, u)
	}

	runes := make([]rune, 0, len(units))
	for i := 0; i < len(units); {
		u := units[i]
		if u >= 0xD800 && u <= 0xDBFF {
			if i+1 >= len(units) {
				return nil, fmt.Errorf("invalid UTF-16: truncated surrogate pair")
			}
			low := units[i+1]
			if low < 0xDC00 || low > 0xDFFF {
				return nil, fmt.Errorf("invalid UTF-16: invalid low surrogate U+%04X", low)
			}
			runes = append(runes, utf16.DecodeRune(rune(u), rune(low)))
			i += 2
		} else if u >= 0xDC00 && u <= 0xDFFF {
			return nil, fmt.Errorf("invalid UTF-16: unexpected low surrogate U+%04X", u)
		} else {
			runes = append(runes, rune(u))
			i++
		}
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
