package mal

//Type is the 'parent' for all Mal data structures. E.g. List, Atom, etc
type Type interface {
}

//List holds a list of MalTypes
type List struct {
	Value []Type
}

//Symbol holds a symbol
type Symbol struct {
	Value string
}

//Number holds an number, represented as a float64
type Number struct {
	Value float64
}

//Function holds a function
type Function struct {
	Value func(args ...Type) Type
}
