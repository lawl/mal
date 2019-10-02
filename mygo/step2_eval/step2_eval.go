package main

import (
	"bufio"
	"fmt"
	"mygomal/mal"
	"os"
)

func read(s string) (mal.Type, error) {
	ast, err := mal.ReadStr(s)
	if err != nil {
		return nil, err
	}
	return ast, nil
}

func eval(ast mal.Type, replEnv map[string]func(args ...mal.Type) (mal.Type, error)) (mal.Type, error) {
	switch v := ast.(type) {
	case *mal.List:
		if len(v.Value) == 0 {
			return ast, nil
		}
		ev, err := evalAst(v, replEnv)
		if err != nil {
			return nil, err
		}
		lst, _ := ev.(*mal.List)
		fn, _ := lst.Value[0].(*mal.Function)
		return fn.Fn(lst.Value[1:]...)

	default:
		return evalAst(v, replEnv)
	}
}

func evalAst(ast mal.Type, replEnv map[string]func(args ...mal.Type) (mal.Type, error)) (mal.Type, error) {
	switch v := ast.(type) {
	case *mal.Symbol:
		//TODO could be a variable too, probably wrong. fix later
		fn, ok := replEnv[v.Value]
		if !ok {
			return nil, fmt.Errorf("Unknown symbol '%s'", v.Value)
		}
		return &mal.Function{Fn: fn}, nil
	case *mal.List:
		var list mal.List
		for _, val := range v.Value {
			evaled, err := eval(val, replEnv)
			if err != nil {
				return nil, err
			}
			list.Value = append(list.Value, evaled)
		}
		return &list, nil
	default:
		return ast, nil
	}
}

func print(ast mal.Type) {
	fmt.Println(mal.PrString(ast, true))
}

func rep(s string) {

	replEnv := map[string]func(args ...mal.Type) (mal.Type, error){
		"+": func(args ...mal.Type) (mal.Type, error) {
			a, _ := args[0].(*mal.Number)
			b, _ := args[1].(*mal.Number)
			return &mal.Number{Value: a.Value + b.Value}, nil
		},
		"-": func(args ...mal.Type) (mal.Type, error) {
			a, _ := args[0].(*mal.Number)
			b, _ := args[1].(*mal.Number)
			return &mal.Number{Value: a.Value - b.Value}, nil
		},
		"*": func(args ...mal.Type) (mal.Type, error) {
			a, _ := args[0].(*mal.Number)
			b, _ := args[1].(*mal.Number)
			return &mal.Number{Value: a.Value * b.Value}, nil
		},
		"/": func(args ...mal.Type) (mal.Type, error) {
			a, _ := args[0].(*mal.Number)
			b, _ := args[1].(*mal.Number)
			return &mal.Number{Value: a.Value / b.Value}, nil
		},
	}

	ast, err := read(s)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
	expr, err := eval(ast, replEnv)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
	print(expr)
}

func main() {
	stdin := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("user> ")
		s, _ := stdin.ReadString('\n')
		rep(s)
	}
}
