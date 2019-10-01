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

func eval(ast mal.Type, env *mal.Env) (mal.Type, error) {
	switch v := ast.(type) {
	case *mal.List:
		if len(v.Value) == 0 {
			return ast, nil
		}
		ev, err := evalAst(v, env)
		if err != nil {
			return nil, err
		}
		lst, _ := ev.(*mal.List)
		fn, _ := lst.Value[0].(*mal.Function)
		return fn.Value(lst.Value[1:]...), nil

	default:
		return evalAst(v, env)
	}
}

func evalAst(ast mal.Type, env *mal.Env) (mal.Type, error) {
	switch v := ast.(type) {
	case *mal.Symbol:
		val := env.Get(v)
		if val == nil {
			return nil, fmt.Errorf("Unknown symbol '%s'", v.Value)
		}
		return val, nil
	case *mal.List:
		var list mal.List
		for _, val := range v.Value {
			evaled, err := eval(val, env)
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
	fmt.Println(mal.PrString(ast))
}

func rep(s string) {
	replEnv := mal.NewEnv(nil)
	replEnv.Set(&mal.Symbol{Value: "+"}, &mal.Function{Value: func(args ...mal.Type) mal.Type {
		a, _ := args[0].(*mal.Number)
		b, _ := args[1].(*mal.Number)
		return &mal.Number{Value: a.Value + b.Value}
	}})
	replEnv.Set(&mal.Symbol{Value: "-"}, &mal.Function{Value: func(args ...mal.Type) mal.Type {
		a, _ := args[0].(*mal.Number)
		b, _ := args[1].(*mal.Number)
		return &mal.Number{Value: a.Value - b.Value}
	}})
	replEnv.Set(&mal.Symbol{Value: "*"}, &mal.Function{Value: func(args ...mal.Type) mal.Type {
		a, _ := args[0].(*mal.Number)
		b, _ := args[1].(*mal.Number)
		return &mal.Number{Value: a.Value * b.Value}
	}})
	replEnv.Set(&mal.Symbol{Value: "/"}, &mal.Function{Value: func(args ...mal.Type) mal.Type {
		a, _ := args[0].(*mal.Number)
		b, _ := args[1].(*mal.Number)
		return &mal.Number{Value: a.Value / b.Value}
	}})

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
