package mal

import "strings"

//PrString takes a MalType and returns a string representation
func PrString(ast MalType) string {
	var sb strings.Builder

	switch v := ast.(type) {
	case *MalList:
		sb.WriteString("(")
		for i, vel := range v.value {
			sb.WriteString(printAtom(vel))
			if i < len(v.value)-1 {
				sb.WriteString(" ")
			}
		}
		sb.WriteString(")")
	default:
		sb.WriteString(printAtom(v))

	}
	return sb.String()
}

func printAtom(atom MalType) string {
	switch v := atom.(type) {
	case *MalSymbol:
		return v.value
	case *MalList:
		return PrString(v)

	default:
		return "<TO STRING NOT IMPLEMENTED>"
	}
}
