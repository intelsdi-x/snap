package opentsdb

import (
	"bytes"
	"fmt"
)

//StringValue implements Json interface to
//replece invalid characters.
type StringValue string

func (sv StringValue) MarshalJSON() ([]byte, error) {
	length := len(sv)

	result := bytes.NewBuffer(make([]byte, 0, length+2))
	result.WriteByte('"')
	for i := 0; i < length; i++ {
		b := sv[i]
		switch {
		case (b >= '-' && b <= '9') || // '-', '.', '/', 0-9
			(b >= 'A' && b <= 'Z') ||
			(b >= 'a' && b <= 'z'):
			result.WriteByte(b)
		case b == '_':
			result.WriteString("__")
		case b == ':':
			result.WriteString("_.")
		default:
			result.WriteString(fmt.Sprintf("_%X", b))
		}
	}
	result.WriteByte('"')
	return result.Bytes(), nil
}

func (sv *StringValue) UnmarshalJSON(json []byte) error {
	escapeLevel := 0
	var parsedByte byte

	result := bytes.NewBuffer(make([]byte, 0, len(json)-2))

	for i, b := range json {
		if i == 0 {
			if b != '"' {
				return fmt.Errorf("expected '\"', got %q", b)
			}
			continue
		}
		if i == len(json)-1 {
			if b != '"' {
				return fmt.Errorf("expected '\"', got %q", b)
			}
			break
		}
		switch escapeLevel {
		case 0:
			if b == '_' {
				escapeLevel = 1
				continue
			}
			result.WriteByte(b)
		case 1:
			switch {
			case b == '_':
				result.WriteByte('_')
				escapeLevel = 0
			case b == '.':
				result.WriteByte(':')
				escapeLevel = 0
			case b >= '0' && b <= '9':
				parsedByte = (b - 48) << 4
				escapeLevel = 2
			case b >= 'A' && b <= 'F': // A-F
				parsedByte = (b - 55) << 4
				escapeLevel = 2
			default:
				return fmt.Errorf(
					"illegal escape sequence at byte %d (%c)",
					i, b,
				)
			}
		case 2:
			switch {
			case b >= '0' && b <= '9':
				parsedByte += b - 48
			case b >= 'A' && b <= 'F': // A-F
				parsedByte += b - 55
			default:
				return fmt.Errorf(
					"illegal escape sequence at byte %d (%c)",
					i, b,
				)
			}
			result.WriteByte(parsedByte)
			escapeLevel = 0
		default:
			panic("unexpected escape level")
		}
	}
	*sv = StringValue(result.String())
	return nil
}
