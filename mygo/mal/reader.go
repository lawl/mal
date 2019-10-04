package mal

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

//Reader reads lisp for tokenization and parsing
type Reader struct {
	toks []string
	pos  int
}

func (reader *Reader) next() (val string, eof bool) {
	cur := reader.pos
	reader.pos++
	if cur >= len(reader.toks) {
		return "", true
	}
	return reader.toks[cur], false
}

func (reader *Reader) peek() (val string, eof bool) {
	if reader.pos >= len(reader.toks) {
		return "", true
	}
	return reader.toks[reader.pos], false
}

//NewReader Creates a new Reader instance from a list of tokens
func NewReader(toks []string) *Reader {
	r := Reader{toks: toks, pos: 0}
	return &r
}

//ReadStr parses a given string into an AST
func ReadStr(s string) (Type, error) {
	toks := tokenize(s)
	reader := NewReader(toks)
	return readForm(reader)
}

func tokenize(s string) []string {
	re := regexp.MustCompile(`[\s,]*(~@|[\[\]{}()'` + "`" +
		`~^@]|"(?:\\.|[^\\"])*"?|;.*|[^\s\[\]{}('"` + "`" +
		`,;)]*)`)
	matches := re.FindAllStringSubmatch(s, -1)
	if matches == nil {
		return make([]string, 0)
	}
	res := make([]string, len(matches))
	for i := range matches {
		res[i] = matches[i][1] // 0 is the original string, we only want the submatch
	}
	return res
}

func readForm(reader *Reader) (Type, error) {
	p, eof := reader.peek()

	if eof {
		return nil, nil
	}

	switch p {
	case "(":
		reader.next()
		v, err := readList(reader)
		if err != nil {
			return nil, err
		}
		return v, nil
	case "[":
		reader.next()
		v, err := readVector(reader)
		if err != nil {
			return nil, err
		}
		return v, nil
	case "{":
		reader.next()
		v, err := readHashmap(reader)
		if err != nil {
			return nil, err
		}
		return v, nil
	default:
		v, err := readAtom(reader)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}

func readList(reader *Reader) (Type, error) {
	list := NewList(false)
	for {
		peek, eof := reader.peek()
		if peek != ")" && !eof {
			v, err := readForm(reader)
			if err != nil {
				return nil, err
			}
			list.Value = append(list.Value, v)
		} else if eof {
			return nil, fmt.Errorf("unbalanced parenthesis in list, expected ')'")
		} else {
			reader.next()
			return &list, nil
		}

	}
}

func readVector(reader *Reader) (Type, error) {
	vector := NewList(true)
	for {
		peek, eof := reader.peek()
		if peek != "]" && !eof {
			v, err := readForm(reader)
			if err != nil {
				return nil, err
			}
			vector.Value = append(vector.Value, v)
		} else if eof {
			return nil, fmt.Errorf("unbalanced parenthesis in vector, expected ']'")
		} else {
			reader.next()
			return &vector, nil
		}

	}
}

func readHashmap(reader *Reader) (Type, error) {
	hmap := NewHashMap()
	for {
		peek, eof := reader.peek()
		if peek != "}" && !eof {
			key, err := readForm(reader)
			if err != nil {
				return nil, err
			}
			value, err := readForm(reader)
			if err != nil {
				return nil, err
			}
			strKey, err := TypeToHashKey(key)
			if err != nil {
				return nil, err
			}
			hmap.Value[strKey] = value

		} else if eof {
			return nil, fmt.Errorf("unbalanced parenthesis in hash map, expected '}'")
		} else {
			reader.next()
			return &hmap, nil
		}

	}
}

func readAtom(reader *Reader) (Type, error) {
	val, eof := reader.next()
	if eof {
		return nil, fmt.Errorf("Tried to read atom, but reached EOF")
	}

	if asNumber, err := strconv.ParseFloat(val, 64); err == nil {
		return &Number{Value: asNumber}, nil
	}

	switch val {
	case "true":
		return &Boolean{Value: true}, nil
	case "false":
		return &Boolean{Value: false}, nil
	case "nil":
		return &Nil{}, nil
	case "@":
		return readerMacroExpand(reader, "deref")
	case "'":
		return readerMacroExpand(reader, "quote")
	case "~":
		return readerMacroExpand(reader, "unquote")
	case "~@":
		return readerMacroExpand(reader, "splice-unquote")
	case "`":
		return readerMacroExpand(reader, "quasiquote")
	}

	if strings.HasPrefix(val, "\"") {

		if !strings.HasSuffix(val, "\"") {
			return nil, fmt.Errorf("unbalanced quotes in string")
		}

		s := val[1 : len(val)-1]
		s = StringUnescape(s)
		return &String{Value: s}, nil
	}

	if strings.HasPrefix(val, ";") { // comment: skip
		return readForm(reader)
	}

	if strings.HasPrefix(val, ":") {
		return &Keyword{Value: val}, nil
	}

	return &Symbol{Value: val}, nil
}

func readerMacroExpand(reader *Reader, symbolName string) (*List, error) {
	list := NewList(false)
	list.Value = append(list.Value, &Symbol{Value: symbolName})
	v, err := readForm(reader)
	if err != nil {
		return nil, err
	}
	list.Value = append(list.Value, v)
	return &list, nil
}
