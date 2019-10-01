package mal

//Env contains a lisp environment, and a pointer to the outer environment, if any
type Env struct {
	outer *Env
	data  map[string]Type
}

//NewEnv creates a new lisp parameter, taking a pointer to an outer environment, or nil, if none
func NewEnv(outer *Env) *Env {
	return &Env{outer: outer, data: make(map[string]Type)}
}

//Set sets a value in the environment
func (env *Env) Set(symbol *Symbol, value Type) {
	env.data[symbol.Value] = value
}

//Find the environment in which given symbol exists, recursing up all its parents if neccessary
func (env *Env) Find(symbol *Symbol) *Env {
	if _, ok := env.data[symbol.Value]; ok {
		return env
	}
	if env.outer == nil {
		return nil
	}
	return env.outer.Find(symbol)
}

//Get obtains the value for a given symbol in an environment, recursing up all its parents if neccessary
func (env *Env) Get(symbol *Symbol) Type {
	e := env.Find(symbol)
	if e == nil {
		return nil
	}
	if val, ok := e.data[symbol.Value]; ok {
		return val
	}
	return nil
}
