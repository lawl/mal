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

func eval(ast mal.Type, replEnv map[string]func(args ...mal.Type) mal.Type) mal.Type {
	switch v := ast.(type) {
	case *mal.List:
		if len(v.Value) == 0 {
			return ast
		}
		ev := evalAst(v, replEnv)
		lst, _ := ev.(*mal.List)
		fn, _ := lst.Value[0].(*mal.Function)
		return fn.Value(lst.Value[1:]...)

	default:
		return evalAst(v, replEnv)
	}
}

func evalAst(ast mal.Type, replEnv map[string]func(args ...mal.Type) mal.Type) mal.Type {
	switch v := ast.(type) {
	case *mal.Symbol:
		//TODO could be a variable too, probably wrong. fix later
		return &mal.Function{Value: replEnv[v.Value]}
	case *mal.List:
		var list mal.List
		for _, val := range v.Value {
			list.Value = append(list.Value, eval(val, replEnv))
		}
		return &list
	default:
		return ast
	}
}

func print(ast mal.Type) {
	fmt.Println(mal.PrString(ast))
}

func rep(s string) {

	replEnv := map[string]func(args ...mal.Type) mal.Type{
		"+": func(args ...mal.Type) mal.Type {
			a, _ := args[0].(*mal.Number)
			b, _ := args[1].(*mal.Number)
			return &mal.Number{Value: a.Value + b.Value}
		},
		"-": func(args ...mal.Type) mal.Type {
			a, _ := args[0].(*mal.Number)
			b, _ := args[1].(*mal.Number)
			return &mal.Number{Value: a.Value - b.Value}
		},
		"*": func(args ...mal.Type) mal.Type {
			a, _ := args[0].(*mal.Number)
			b, _ := args[1].(*mal.Number)
			return &mal.Number{Value: a.Value * b.Value}
		},
		"/": func(args ...mal.Type) mal.Type {
			a, _ := args[0].(*mal.Number)
			b, _ := args[1].(*mal.Number)
			return &mal.Number{Value: a.Value / b.Value}
		},
	}

	ast, err := read(s)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
	expr := eval(ast, replEnv)
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
