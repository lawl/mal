package main

import (
	"bufio"
	"flag"
	"fmt"
	"mygomal/mal"
	"os"

	"github.com/chzyer/readline"
)

func read(s string) (mal.Type, error) {
	ast, err := mal.ReadStr(s)
	if err != nil {
		return nil, err
	}
	return ast, nil
}

func setBindingInEnv(env *mal.Env, binding []mal.Type) (mal.Type, error) {
	//1 argument to def! must be a symbol
	symbolName, ok := binding[0].(*mal.Symbol)
	if !ok {
		return nil, fmt.Errorf("first paramter must be of type Symbol, got %T", binding[0])
	}
	ev, err := eval(binding[1], env)
	if err != nil {
		return nil, err
	}
	//update the environment and set the unevaluated symbol to the evaluated argument
	env.Set(symbolName, ev)
	return ev, nil
}

func eval(ast mal.Type, env *mal.Env) (mal.Type, error) {
	switch v := ast.(type) {
	case *mal.List:
		if len(v.Value) == 0 {
			return ast, nil
		}

		// if the first element of the list is a symbol, check for special handling, such as "def!"
		if symb, ok := v.Value[0].(*mal.Symbol); ok {
			switch symb.Value {
			case "def!":
				//check argument length
				if len(v.Value) != 3 {
					return nil, fmt.Errorf("'def!' expects exactly 2 paramters")
				}
				return setBindingInEnv(env, v.Value[1:])
			case "let*":
				newEnv := mal.NewEnv(env, nil, nil)
				if len(v.Value) < 3 {
					return nil, fmt.Errorf("'let*' expects at least 2 paramters")
				}
				if bindings, ok := v.Value[1].(*mal.List); ok {
					for i := 0; i < len(bindings.Value)/2; i++ {
						idx := (i * 2)
						setBindingInEnv(newEnv, bindings.Value[idx:idx+2])
					}

					return eval(v.Value[2], newEnv)
				}
				return nil, fmt.Errorf("'let!': invalid arguments")
			case "do":
				var lastResult mal.Type
				for _, val := range v.Value[1:] {
					var err error
					lastResult, err = evalAst(val, env)
					if err != nil {
						return nil, err
					}
				}
				return lastResult, nil
			case "if":
				r, err := eval(v.Value[1], env)
				if err != nil {
					return nil, err
				}
				evaluatedTo := true
				if b, ok := r.(*mal.Boolean); ok {
					evaluatedTo = b.Value
				}
				if _, ok := r.(*mal.Nil); ok {
					evaluatedTo = false
				}
				if evaluatedTo == true {
					r, err := eval(v.Value[2], env)
					if err != nil {
						return nil, err
					}
					return r, nil
				}
				//condition evaluated to false, check if we have a branch for false, and execute it, if so
				if len(v.Value) < 4 {
					return &mal.Nil{}, nil
				}
				r, err = eval(v.Value[3], env)
				if err != nil {
					return nil, err
				}
				return r, nil
			case "fn*":
				return &mal.Function{Value: func(args ...mal.Type) (mal.Type, error) {
					bindings, ok := v.Value[1].(*mal.List)
					if !ok {
						return nil, fmt.Errorf("Invalid bindings to fn*")
					}
					fnEnv := mal.NewEnv(env, bindings.Value, args)
					return eval(v.Value[2], fnEnv)
				}}, nil
			}
		}

		ev, err := evalAst(v, env)
		if err != nil {
			return nil, err
		}
		lst, _ := ev.(*mal.List)
		fn, _ := lst.Value[0].(*mal.Function)
		return fn.Value(lst.Value[1:]...)

	default:
		return evalAst(v, env)
	}
}

func evalAst(ast mal.Type, env *mal.Env) (mal.Type, error) {
	switch v := ast.(type) {
	case *mal.Symbol:
		val := env.Get(v)
		if val == nil {
			return nil, fmt.Errorf("Unknown symbol '%s' not found", v.Value)
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

func createREPLEnv() *mal.Env {
	replEnv := mal.NewEnv(nil, nil, nil)
	for k, v := range mal.CoreNS {
		replEnv.Set(k, v)
	}
	return replEnv
}

func rep(s string, env *mal.Env) {

	ast, err := read(s)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
	expr, err := eval(ast, env)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
	print(expr)
}

func main() {
	usePlainStdin := flag.Bool("stdin", false, "don't use nice readline based repl. only for tests, as the nice repl breaks them")
	flag.Parse()

	env := createREPLEnv()
	rep("(def! not (fn* (a) (if a false true)))", env)

	if *usePlainStdin {
		stdinREPL(env)
		return
	}
	niceRepl(env)
}

func stdinREPL(env *mal.Env) {
	stdin := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("user> ")
		s, _ := stdin.ReadString('\n')
		rep(s, env)
	}
}

func niceRepl(env *mal.Env) {

	l, err := readline.NewEx(&readline.Config{
		Prompt:       "user> ",
		HistoryFile:  "/tmp/readline.tmp",
		AutoComplete: nil,

		HistorySearchFold: true,
	})
	if err != nil {
		panic(err)
	}
	defer l.Close()

	for {
		s, err := l.Readline()
		if err != nil { // io.EOF
			break
		}
		rep(s, env)
	}
}
