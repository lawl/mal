package mal

import (
	"fmt"
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

//ReadString parses escape sequences in strings such as \n and returns a new string
func ReadString(s string) (string, error) {
	var sb strings.Builder
	sr := NewStringReader(s)
	writtenStringEnd := false

	if !strings.HasPrefix(s, "\"") {
		return "", fmt.Errorf("unbalanced \"")
	}

	for {
		v, eof := sr.next()
		if eof {
			break
		}

		if v == '\\' {
			p, eof := sr.peek()
			if eof {
				return "", fmt.Errorf("expected escape sequence after \\, eof found instead. unbalanced string")
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
			if _, eof := sr.peek(); eof && v == '"' {
				writtenStringEnd = true
			}
			sb.WriteByte(v)
		}
	}
	str := sb.String()
	if len(str) < 2 || !writtenStringEnd {
		return "", fmt.Errorf("unbalanced \"")
	}
	return str[1 : len(str)-1], nil
}

//WriteString parses special characters such as newlines and escapes them
func WriteString(s string) string {
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
	return "\"" + sb.String() + "\""
}
