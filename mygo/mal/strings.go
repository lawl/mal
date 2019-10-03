package mal

import (
	"strings"
)

//StringReader reads lisp for tokenization and parsing
type StringReader struct {
	str string
	pos int
}

func (reader *StringReader) next() (val byte, eof bool) {
	cur := reader.pos
	reader.pos++
	if cur >= len(reader.str) {
		return 0x0, true
	}
	return reader.str[cur], false
}

func (reader *StringReader) peek() (val byte, eof bool) {
	if reader.pos >= len(reader.str) {
		return 0x0, true
	}
	return reader.str[reader.pos], false
}

//NewStringReader Creates a new StringReader instance from a string
func NewStringReader(str string) *StringReader {
	r := StringReader{str: str, pos: 0}
	return &r
}

//StringUnescape parses escape sequences in strings such as \n and returns a new string
func StringUnescape(s string) string {
	var sb strings.Builder
	sr := NewStringReader(s)

	for {
		v, eof := sr.next()
		if eof {
			break
		}

		if v == '\\' {
			p, eof := sr.peek()
			if eof {
				sb.WriteByte(v)
				break
			}
			switch p {
			case 'n':
				sr.next()
				sb.WriteByte('\n')
			case '"':
				sr.next()
				sb.WriteByte('"')
			case '\\':
				sr.next()
				sb.WriteByte('\\')
			}
		} else {
			sb.WriteByte(v)
		}
	}
	return sb.String()
}

//StringEscape parses special characters such as newlines and escapes them
func StringEscape(s string) string {
	var sb strings.Builder
	sr := NewStringReader(s)

	for {
		v, eof := sr.next()
		if eof {
			break
		}

		switch v {
		case '\n':
			sb.WriteString("\\n")
		case '"':
			sb.WriteString("\\\"")
		case '\\':
			sb.WriteString("\\\\")
		default:
			sb.WriteByte(v)
		}

	}
	return sb.String()
}
