package mal

//Type is the 'parent' for all Mal data structures. E.g. List, Atom, etc
type Type interface {
}

//List holds a list of MalTypes
type List struct {
	Value    []Type
	IsVector bool
	Meta     Type
}

//NewList creates a list or vector
func NewList(isVector bool) List {
	var list List
	list.IsVector = isVector
	return list
}

//HashMap holds mappings from string -> MalType
type HashMap struct {
	Value map[string]Type
	Meta  Type
}

//NewHashMap creates a new HashMap
func NewHashMap() HashMap {
	var m HashMap
	m.Value = make(map[string]Type, 16) //bucket size 16 by default, arbitrary
	return m
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
	Ast     Type
	Params  []Type
	Env     *Env
	IsMacro bool
	Fn      func(args ...Type) (Type, error)
	Meta    Type
}

//Boolean holds a boolean
type Boolean struct {
	Value bool
}

//Nil is nil
type Nil struct {
}

//String holds, perhaps unexpectedly a string
type String struct {
	Value string
}

//Keyword holds, a keyword
type Keyword struct {
	Value string
}

//Atom holds a reference to a mal value
type Atom struct {
	Value Type
	Meta  Type
}

//Error holds an Error
type Error struct {
	Value Type
}

func (err *Error) Error() string {
	return "Error: " + PrString(err.Value, true)
}
