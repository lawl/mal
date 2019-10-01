package mal

//Type is the 'parent' for all Mal data structures. E.g. List, Atom, etc
type Type interface {
}

//List holds a list of MalTypes
type List struct {
	value []Type
}

//Symbol holds a symbol
type Symbol struct {
	value string
}

//Number holds an number, represented as a float64
type Number struct {
	value float64
}
