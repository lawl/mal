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
		// format the float as a string, g as parameter means:
		// see https://golang.org/pkg/strconv/#FormatFloat
		// g' ('e' for large exponents, 'f' otherwise)
		//  'e' (-d.dddde±dd, a decimal exponent)
		//  'f' (-ddd.dddd, no exponent)
		return strconv.FormatFloat(v.Value, 'g', -1, 64)
	case *List:
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
		s = strings.ReplaceAll(s, "\\", "\\\\")
		s = strings.ReplaceAll(s, "\"", "\\\"")
		s = strings.ReplaceAll(s, "\n", "\\n")
		return "\"" + s + "\""

	default:
		return fmt.Sprintf("<No print implementation for atom type: %T>", atom)
	}
}
