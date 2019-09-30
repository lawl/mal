package mal

//MalType is the 'parent' for all Mal data structures. E.g. List, Atom, etc
type MalType interface {
}

//MalList holds a list of MalTypes
type MalList struct {
	value []MalType
}

//MalSymbol holds a symbol
type MalSymbol struct {
	value string
}

//MalInteger holds an integer
type MalInteger struct {
	value int
}
