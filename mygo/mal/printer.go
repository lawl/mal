package mal

import (
	"fmt"
	"strconv"
	"strings"
)

//PrString takes a MalType and returns a string representation
func PrString(ast Type, readably bool) string {
	var sb strings.Builder

	switch v := ast.(type) {
	case *List:
		sb.WriteString("(")
		for i, vel := range v.Value {
			sb.WriteString(printAtom(vel, readably))
			if i < len(v.Value)-1 {
				sb.WriteString(" ")
			}
		}
		sb.WriteString(")")
	case *Vector:
		sb.WriteString("[")
		for i, vel := range v.Value {
			sb.WriteString(printAtom(vel, readably))
			if i < len(v.Value)-1 {
				sb.WriteString(" ")
			}
		}
		sb.WriteString("]")
	default:
		sb.WriteString(printAtom(v, readably))

	}
	return sb.String()
}

func printAtom(atom Type, readably bool) string {
	switch v := atom.(type) {
	case *Symbol:
		return v.Value
	case *Number:
		// see https://golang.org/pkg/strconv/#FormatFloat
		//  'f' (-ddd.dddd, no exponent)
		return strconv.FormatFloat(v.Value, 'f', -1, 64)
	case *List:
		return PrString(v, readably)
	case *Vector:
		return PrString(v, readably)
	case *Boolean:
		if v.Value {
			return "true"
		}
		return "false"
	case *Nil:
		return "nil"
	case *Function:
		return "#<function>"
	case *String:
		s := v.Value
		if readably {
			s = strings.ReplaceAll(s, "\\", "\\\\")
			s = strings.ReplaceAll(s, "\"", "\\\"")
			s = strings.ReplaceAll(s, "\n", "\\n")
			s = "\"" + s + "\""
		}
		return s
	case *Atom:
		var sb strings.Builder
		sb.WriteString("(atom ")
		sb.WriteString(printAtom(v.Value, readably))
		sb.WriteString(")")
		return sb.String()

	default:
		return fmt.Sprintf("<No print implementation for atom type: %T>", atom)
	}
}
