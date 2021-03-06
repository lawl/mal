package mal

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"
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
		newLst := NewList(false)
		newLst.Value = append(newLst.Value, v)
		newLst.Value = append(newLst.Value, lst.Value...)
		return &newLst, nil
	}},
	&Symbol{Value: "concat"}: &Function{Fn: func(args ...Type) (Type, error) {
		newLst := NewList(false)
		for _, val := range args {
			if v, ok := val.(*List); ok {
				newLst.Value = append(newLst.Value, v.Value...)
			} else {
				return nil, fmt.Errorf("concat expects all parameters to be lists")
			}
		}
		return &newLst, nil
	}},
	&Symbol{Value: "first"}: &Function{Fn: func(args ...Type) (Type, error) {
		lst, isList := args[0].(*List)
		nul, isNil := args[0].(*Nil)

		if !isList && isNil {
			return nul, nil
		}
		if len(lst.Value) == 0 {
			return &Nil{}, nil
		}
		return lst.Value[0], nil
	}},

	&Symbol{Value: "nth"}: &Function{Fn: func(args ...Type) (Type, error) {
		lst, _ := args[0].(*List)
		idx, _ := args[1].(*Number)
		if idx.Value >= 0 && int(idx.Value) < len(lst.Value) {
			return lst.Value[int(idx.Value)], nil
		}
		return nil, fmt.Errorf("nth: Index out of range")
	}},
	&Symbol{Value: "rest"}: &Function{Fn: func(args ...Type) (Type, error) {
		lst, isList := args[0].(*List)
		_, isNil := args[0].(*Nil)

		if !isNil && isList && len(lst.Value) >= 1 {
			l := NewList(false)
			l.Value = lst.Value[1:]
			return &l, nil
		}
		l := NewList(false)
		return &l, nil
	}},
	&Symbol{Value: "throw"}: &Function{Fn: func(args ...Type) (Type, error) {
		err := Error{Value: args[0]}
		return nil, &err
	}},
	&Symbol{Value: "apply"}: &Function{Fn: func(args ...Type) (Type, error) {
		fn, isFN := args[0].(*Function)

		if !isFN || len(args) <= 1 {
			return nil, fmt.Errorf("Invalid arguments to 'apply'")
		}
		fnArgs := make([]Type, 0)
		for _, v := range args[1:] {
			if asList, ok := v.(*List); ok {
				for _, listEl := range asList.Value {
					fnArgs = append(fnArgs, listEl)
				}
				continue
			}
			fnArgs = append(fnArgs, v)
		}
		return fn.Fn(fnArgs...)
	}},

	&Symbol{Value: "map"}: &Function{Fn: func(args ...Type) (Type, error) {
		fn, isFN := args[0].(*Function)

		if !isFN || len(args) <= 1 {
			return nil, fmt.Errorf("Invalid arguments to 'map'")
		}

		rList := NewList(false)
		if asList, ok := args[1].(*List); ok {
			for _, listEl := range asList.Value {
				res, err := fn.Fn(listEl)
				if err != nil {
					return nil, err
				}
				rList.Value = append(rList.Value, res)
			}
		}

		return &rList, nil
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
	&Symbol{Value: "nil?"}: &Function{Fn: func(args ...Type) (Type, error) {
		_, ok := args[0].(*Nil)
		return &Boolean{Value: ok}, nil
	}},
	&Symbol{Value: "true?"}: &Function{Fn: func(args ...Type) (Type, error) {
		b, ok := args[0].(*Boolean)
		return &Boolean{Value: ok && b.Value}, nil
	}},
	&Symbol{Value: "false?"}: &Function{Fn: func(args ...Type) (Type, error) {
		b, ok := args[0].(*Boolean)
		return &Boolean{Value: ok && !b.Value}, nil
	}},
	&Symbol{Value: "symbol?"}: &Function{Fn: func(args ...Type) (Type, error) {
		_, ok := args[0].(*Symbol)
		return &Boolean{Value: ok}, nil
	}},
	&Symbol{Value: "symbol"}: &Function{Fn: func(args ...Type) (Type, error) {
		str, _ := args[0].(*String)
		return &Symbol{Value: str.Value}, nil
	}},
	&Symbol{Value: "keyword"}: &Function{Fn: func(args ...Type) (Type, error) {
		if kw, ok := args[0].(*Keyword); ok {
			return kw, nil
		}
		str, _ := args[0].(*String)
		return &Keyword{Value: ":" + str.Value}, nil
	}},
	&Symbol{Value: "keyword?"}: &Function{Fn: func(args ...Type) (Type, error) {
		_, ok := args[0].(*Keyword)
		return &Boolean{Value: ok}, nil
	}},
	&Symbol{Value: "vector"}: &Function{Fn: func(args ...Type) (Type, error) {
		vec := NewList(true)
		vec.Value = append(vec.Value, args...)
		return &vec, nil
	}},
	&Symbol{Value: "vector?"}: &Function{Fn: func(args ...Type) (Type, error) {
		vec, ok := args[0].(*List)
		return &Boolean{Value: ok && vec.IsVector == true}, nil
	}},
	&Symbol{Value: "sequential?"}: &Function{Fn: func(args ...Type) (Type, error) {
		_, ok := args[0].(*List)
		return &Boolean{Value: ok}, nil
	}},
	&Symbol{Value: "hash-map"}: &Function{Fn: func(args ...Type) (Type, error) {
		if len(args)%2 != 0 {
			return nil, fmt.Errorf("hash-map requires an even number of arguments")
		}
		hmap := NewHashMap()
		for i := 0; i < len(args); i += 2 {
			key := args[i]
			value := args[i+1]
			strKey, err := TypeToHashKey(key)
			if err != nil {
				return nil, err
			}
			hmap.Value[strKey] = value
		}
		return &hmap, nil
	}},
	&Symbol{Value: "map?"}: &Function{Fn: func(args ...Type) (Type, error) {
		_, ok := args[0].(*HashMap)
		return &Boolean{Value: ok}, nil
	}},

	&Symbol{Value: "assoc"}: &Function{Fn: func(args ...Type) (Type, error) {
		toAssoc := args[1:]
		if len(toAssoc)%2 != 0 {
			return nil, fmt.Errorf("hash-map requires an even number of arguments")
		}
		originalMap, ok := args[0].(*HashMap)
		if !ok {
			return nil, fmt.Errorf("First argument to assoc must be a hash map")
		}
		hmap := NewHashMap()
		for k, v := range originalMap.Value {
			hmap.Value[k] = v
		}
		for i := 0; i < len(toAssoc); i += 2 {
			key := toAssoc[i]
			value := toAssoc[i+1]
			strKey, err := TypeToHashKey(key)
			if err != nil {
				return nil, err
			}
			hmap.Value[strKey] = value
		}
		return &hmap, nil
	}},

	&Symbol{Value: "dissoc"}: &Function{Fn: func(args ...Type) (Type, error) {
		toDissoc := args[1:]
		originalMap, ok := args[0].(*HashMap)
		if !ok {
			return nil, fmt.Errorf("First argument to dissoc must be a hash map")
		}
		hmap := NewHashMap()
	outer:
		for k, v := range originalMap.Value {
			for _, key := range toDissoc {
				strKey, err := TypeToHashKey(key)
				if err != nil {
					return nil, err
				}
				if strKey == k {
					continue outer
				}
			}
			hmap.Value[k] = v
		}
		return &hmap, nil
	}},
	&Symbol{Value: "get"}: &Function{Fn: func(args ...Type) (Type, error) {
		key := args[1]
		hmap, ok := args[0].(*HashMap)
		if !ok {
			return &Nil{}, nil
		}
		strKey, err := TypeToHashKey(key)
		if err != nil {
			return nil, err
		}
		if val, ok := hmap.Value[strKey]; ok {
			return val, nil
		}
		return &Nil{}, nil
	}},
	&Symbol{Value: "contains?"}: &Function{Fn: func(args ...Type) (Type, error) {
		key := args[1]
		hmap, ok := args[0].(*HashMap)
		if !ok {
			return nil, fmt.Errorf("First argument to contains? must be a hash map")
		}
		strKey, err := TypeToHashKey(key)
		if err != nil {
			return nil, err
		}
		if _, ok := hmap.Value[strKey]; ok {
			return &Boolean{Value: true}, nil
		}
		return &Boolean{Value: false}, nil
	}},
	&Symbol{Value: "keys"}: &Function{Fn: func(args ...Type) (Type, error) {
		hmap, ok := args[0].(*HashMap)
		if !ok {
			return nil, fmt.Errorf("First argument to assoc must be a hash map")
		}
		keyList := NewList(false)
		for k := range hmap.Value {
			keyList.Value = append(keyList.Value, NativeStringToMalHashKey(k))
		}
		return &keyList, nil
	}},
	&Symbol{Value: "vals"}: &Function{Fn: func(args ...Type) (Type, error) {
		hmap, ok := args[0].(*HashMap)
		if !ok {
			return nil, fmt.Errorf("First argument to assoc must be a hash map")
		}
		valList := NewList(false)
		for _, val := range hmap.Value {
			valList.Value = append(valList.Value, val)
		}
		return &valList, nil
	}},

	&Symbol{Value: "readline"}: &Function{Fn: func(args ...Type) (Type, error) {
		str, isString := args[0].(*String)
		stdin := bufio.NewReader(os.Stdin)
		if isString {
			fmt.Print(str.Value)
		}
		s, err := stdin.ReadString('\n')
		s = strings.Trim(s, "\n")
		if err != nil {
			return &Nil{}, nil
		}
		return &String{Value: s}, nil
	}},
	&Symbol{Value: "time-ms"}: &Function{Fn: func(args ...Type) (Type, error) {
		t := time.Now().UnixNano() / time.Millisecond.Milliseconds()
		return &Number{Value: float64(t)}, nil
	}},
	&Symbol{Value: "fn?"}: &Function{Fn: func(args ...Type) (Type, error) {
		fn, ok := args[0].(*Function)
		return &Boolean{Value: ok && !fn.IsMacro}, nil
	}},

	&Symbol{Value: "macro?"}: &Function{Fn: func(args ...Type) (Type, error) {
		fn, ok := args[0].(*Function)
		return &Boolean{Value: ok && fn.IsMacro}, nil
	}},
	&Symbol{Value: "string?"}: &Function{Fn: func(args ...Type) (Type, error) {
		_, ok := args[0].(*String)
		return &Boolean{Value: ok}, nil
	}},
	&Symbol{Value: "number?"}: &Function{Fn: func(args ...Type) (Type, error) {
		_, ok := args[0].(*Number)
		return &Boolean{Value: ok}, nil
	}},
	&Symbol{Value: "seq"}: &Function{Fn: func(args ...Type) (Type, error) {
		if list, ok := args[0].(*List); ok {
			if len(list.Value) == 0 {
				return &Nil{}, nil
			}
			if list.IsVector {
				newList := NewList(false)
				newList.Value = list.Value
				return &newList, nil
			}
			return list, nil
		}
		if str, ok := args[0].(*String); ok {
			if len(str.Value) == 0 {
				return &Nil{}, nil
			}
			newList := NewList(false)
			sList := strings.Split(str.Value, "")
			for _, val := range sList {
				newList.Value = append(newList.Value, &String{Value: val})
			}
			return &newList, nil
		}
		if nul, ok := args[0].(*Nil); ok {
			return nul, nil
		}
		return nil, fmt.Errorf("seq: Argument 1 must be of type vector, list or string")
	}},
	&Symbol{Value: "conj"}: &Function{Fn: func(args ...Type) (Type, error) {
		if list, ok := args[0].(*List); ok {
			newList := NewList(list.IsVector)
			els := args[1:]
			if !list.IsVector {
				for i := len(els) - 1; i >= 0; i-- {
					newList.Value = append(newList.Value, els[i])
				}
				newList.Value = append(newList.Value, list.Value...)
				return &newList, nil
			}
			newList.Value = append(newList.Value, list.Value...)
			for _, el := range els {
				newList.Value = append(newList.Value, el)
			}
			return &newList, nil
		}
		return nil, fmt.Errorf("conj: Argument 1 must be of type vector, or list")
	}},
	&Symbol{Value: "meta"}: &Function{Fn: func(args ...Type) (Type, error) {
		if hmap, ok := args[0].(*HashMap); ok {
			if hmap.Meta == nil {
				return &Nil{}, nil
			}
			return hmap.Meta, nil
		}
		if list, ok := args[0].(*List); ok {
			if list.Meta == nil {
				return &Nil{}, nil
			}
			return list.Meta, nil
		}
		if atom, ok := args[0].(*Atom); ok {
			if atom.Meta == nil {
				return &Nil{}, nil
			}
			return atom.Meta, nil
		}
		if fn, ok := args[0].(*Function); ok {
			if fn.Meta == nil {
				return &Nil{}, nil
			}
			return fn.Meta, nil
		}
		return nil, fmt.Errorf("meta: Cannot retrieve metadata to non composite or function type %T", args[0])
	}},
	&Symbol{Value: "with-meta"}: &Function{Fn: func(args ...Type) (Type, error) {
		if hmap, ok := args[0].(*HashMap); ok {
			newmap := NewHashMap()
			newmap.Value = hmap.Value
			newmap.Meta = args[1]
			return &newmap, nil
		}
		if list, ok := args[0].(*List); ok {
			newList := NewList(list.IsVector)
			newList.Value = list.Value
			newList.Meta = args[1]
			return &newList, nil
		}
		if atom, ok := args[0].(*Atom); ok {
			newAtom := Atom{Value: atom.Value}
			newAtom.Meta = args[1]
			return &newAtom, nil
		}
		if fn, ok := args[0].(*Function); ok {
			newFn := CopyOfFunction(fn)
			newFn.Meta = args[1]
			return newFn, nil
		}
		return nil, fmt.Errorf("with-meta: Cannot add metadata to non composite or function type %T", args[0])
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
	case *Keyword:
		v2, _ := args[1].(*Keyword)
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
