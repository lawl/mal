package mal

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"sort"
	"strings"
)

//CoreNS contains builtin functions for mal
var CoreNS = map[*Symbol]*Function{
	&Symbol{Value: "+"}: &Function{Fn: func(args ...Type) (Type, error) {
		a, _ := args[0].(*Number)
		b, _ := args[1].(*Number)
		return &Number{Value: a.Value + b.Value}, nil
	}},
	&Symbol{Value: "-"}: &Function{Fn: func(args ...Type) (Type, error) {
		a, _ := args[0].(*Number)
		b, _ := args[1].(*Number)
		return &Number{Value: a.Value - b.Value}, nil
	}},
	&Symbol{Value: "*"}: &Function{Fn: func(args ...Type) (Type, error) {
		a, _ := args[0].(*Number)
		b, _ := args[1].(*Number)
		return &Number{Value: a.Value * b.Value}, nil
	}},
	&Symbol{Value: "/"}: &Function{Fn: func(args ...Type) (Type, error) {
		a, _ := args[0].(*Number)
		b, _ := args[1].(*Number)
		return &Number{Value: a.Value / b.Value}, nil
	}},
	//take the parameters and return them as a list.
	&Symbol{Value: "list"}: &Function{Fn: func(args ...Type) (Type, error) {
		return &List{Value: args}, nil
	}},
	//return true if the first parameter is a list, false otherwise.
	&Symbol{Value: "list?"}: &Function{Fn: func(args ...Type) (Type, error) {
		l, ok := args[0].(*List)
		return &Boolean{Value: ok && !l.IsVector}, nil
	}},
	//treat the first parameter as a list and return true if the list is empty and false if it contains any elements.
	&Symbol{Value: "empty?"}: &Function{Fn: func(args ...Type) (Type, error) {
		if lst, ok := args[0].(*List); ok {
			if len(lst.Value) == 0 {
				return &Boolean{Value: true}, nil
			}
		}
		return &Boolean{Value: false}, nil
	}},
	// treat the first parameter as a list and return the number of elements that it contains.
	&Symbol{Value: "count"}: &Function{Fn: func(args ...Type) (Type, error) {
		if lst, ok := args[0].(*List); ok {
			return &Number{Value: float64(len(lst.Value))}, nil
		}
		return &Number{Value: 0}, nil
	}},

	&Symbol{Value: "<"}: &Function{Fn: func(args ...Type) (Type, error) {
		a, _ := args[0].(*Number)
		b, _ := args[1].(*Number)
		return &Boolean{Value: a.Value < b.Value}, nil
	}},
	&Symbol{Value: ">"}: &Function{Fn: func(args ...Type) (Type, error) {
		a, _ := args[0].(*Number)
		b, _ := args[1].(*Number)
		return &Boolean{Value: a.Value > b.Value}, nil
	}},

	&Symbol{Value: "<="}: &Function{Fn: func(args ...Type) (Type, error) {
		a, _ := args[0].(*Number)
		b, _ := args[1].(*Number)
		return &Boolean{Value: a.Value <= b.Value}, nil
	}},

	&Symbol{Value: ">="}: &Function{Fn: func(args ...Type) (Type, error) {
		a, _ := args[0].(*Number)
		b, _ := args[1].(*Number)
		return &Boolean{Value: a.Value >= b.Value}, nil
	}},

	&Symbol{Value: "pr-str"}: &Function{Fn: func(args ...Type) (Type, error) {
		var sb strings.Builder
		for i, v := range args {
			sb.WriteString(PrString(v, true))
			if i < len(args)-1 {
				sb.WriteString(" ")
			}
		}
		return &String{Value: sb.String()}, nil
	}},
	&Symbol{Value: "str"}: &Function{Fn: func(args ...Type) (Type, error) {
		var sb strings.Builder
		for _, v := range args {
			sb.WriteString(PrString(v, false))
		}
		return &String{Value: sb.String()}, nil
	}},
	&Symbol{Value: "prn"}: &Function{Fn: func(args ...Type) (Type, error) {
		var sb strings.Builder
		for i, v := range args {
			sb.WriteString(PrString(v, true))
			if i < len(args)-1 {
				sb.WriteString(" ")
			}
		}
		fmt.Println(sb.String())
		return &Nil{}, nil
	}},
	&Symbol{Value: "println"}: &Function{Fn: func(args ...Type) (Type, error) {
		var sb strings.Builder
		for i, v := range args {
			sb.WriteString(PrString(v, false))
			if i < len(args)-1 {
				sb.WriteString(" ")
			}
		}
		fmt.Println(sb.String())
		return &Nil{}, nil
	}},

	&Symbol{Value: "read-string"}: &Function{Fn: func(args ...Type) (Type, error) {
		v, _ := args[0].(*String)
		return ReadStr(v.Value)
	}},

	&Symbol{Value: "slurp"}: &Function{Fn: func(args ...Type) (Type, error) {
		filename, _ := args[0].(*String)
		dat, err := ioutil.ReadFile(filename.Value)
		if err != nil {
			return nil, err
		}
		return &String{Value: string(dat)}, nil
	}},

	&Symbol{Value: "atom"}: &Function{Fn: func(args ...Type) (Type, error) {
		return &Atom{Value: args[0]}, nil
	}},

	&Symbol{Value: "atom?"}: &Function{Fn: func(args ...Type) (Type, error) {
		_, ok := args[0].(*Atom)
		return &Boolean{Value: ok}, nil
	}},
	&Symbol{Value: "deref"}: &Function{Fn: func(args ...Type) (Type, error) {
		v, ok := args[0].(*Atom)
		if !ok {
			return nil, fmt.Errorf("Argument to deref is not an atom")
		}
		return v.Value, nil
	}},
	&Symbol{Value: "reset!"}: &Function{Fn: func(args ...Type) (Type, error) {
		v, _ := args[0].(*Atom)
		v.Value = args[1]
		return args[1], nil
	}},
	&Symbol{Value: "cons"}: &Function{Fn: func(args ...Type) (Type, error) {
		v := args[0]
		lst, _ := args[1].(*List)
		newLst := NewList(lst.IsVector)
		newLst.Value = append(newLst.Value, v)
		newLst.Value = append(newLst.Value, lst.Value...)
		return &newLst, nil
	}},

	/* Takes an atom, a function, and zero or more function arguments.
	The atom's value is modified to the result of applying the function
	with the atom's value as the first argument and the optionally given
	function arguments as the rest of the arguments. The new atom's value is returned */
	&Symbol{Value: "swap!"}: &Function{Fn: func(args ...Type) (Type, error) {
		v, _ := args[0].(*Atom)
		fn, _ := args[1].(*Function)
		optargs := args[2:]
		fnArgs := make([]Type, len(optargs)+1)
		fnArgs[0] = v.Value
		for i := range optargs {
			fnArgs[i+1] = optargs[i]
		}
		r, err := fn.Fn(fnArgs...)
		if err != nil {
			return nil, err
		}
		v.Value = r
		return r, nil
	}},

	// compare the first two parameters and return true if they are the same type and
	// contain the same value. In the case of equal length lists, each element of the
	// list should be compared for equality and if they are the same return true, otherwise false.
	// if we use an anonymous function here, we can't recurse, but we need to recurse to compare lists
	// so we define this function at the bottom of the file and refer to it by name here
	&Symbol{Value: "="}: &Function{Fn: compareFunc},
}

func compareFunc(args ...Type) (Type, error) {
	if reflect.TypeOf(args[0]) != reflect.TypeOf(args[1]) {
		return &Boolean{Value: false}, nil
	}
	switch v := args[0].(type) {
	case *Symbol:
		v2, _ := args[1].(*Symbol)
		return &Boolean{Value: v.Value == v2.Value}, nil
	case *Number:
		v2, _ := args[1].(*Number)
		return &Boolean{Value: v.Value == v2.Value}, nil
	case *List:
		v2, _ := args[1].(*List)
		if len(v.Value) != len(v2.Value) {
			return &Boolean{Value: false}, nil
		}
		for i := range v.Value {
			r, _ := compareFunc(v.Value[i], v2.Value[i])
			rbool, _ := r.(*Boolean)
			if rbool.Value == false {
				return &Boolean{Value: false}, nil
			}
		}
		return &Boolean{Value: true}, nil
	case *HashMap:
		v2, _ := args[1].(*HashMap)
		k1 := keysFromMap(v.Value)
		k2 := keysFromMap(v2.Value)
		if len(k1) != len(k2) {
			return &Boolean{Value: false}, nil
		}
		sort.Strings(k1)
		sort.Strings(k2)

		for i := range k1 {
			if k1[i] != k2[i] {
				return &Boolean{Value: false}, nil
			}
			r, _ := compareFunc(v.Value[k1[i]], v2.Value[k2[i]])
			rbool, _ := r.(*Boolean)
			if rbool.Value == false {
				return &Boolean{Value: false}, nil
			}
		}

		return &Boolean{Value: true}, nil
	case *Boolean:
		v2, _ := args[1].(*Boolean)
		return &Boolean{Value: v.Value == v2.Value}, nil
	case *Nil:
		return &Boolean{Value: true}, nil
	case *Function:
		return &Boolean{Value: false}, nil // Go cant == functions, false seems to make the most sense
	case *String:
		v2, _ := args[1].(*String)
		return &Boolean{Value: v.Value == v2.Value}, nil
	case *Atom:
		v2, _ := args[1].(*Atom)
		return &Boolean{Value: v == v2}, nil

	default:
		return nil, fmt.Errorf("No equals operation implemented for type: %T", v)
	}
}

func keysFromMap(themap map[string]Type) []string {
	keys := make([]string, len(themap))
	i := 0
	for k := range themap {
		keys[i] = k
		i++
	}
	return keys
}
