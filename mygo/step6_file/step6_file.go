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
tailcalloptimized:
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

					env = newEnv
					ast = v.Value[2]
					goto tailcalloptimized
				}
				return nil, fmt.Errorf("'let!': invalid arguments")
			case "do":
				for _, val := range v.Value[1 : len(v.Value)-1] {
					var err error
					_, err = eval(val, env)
					if err != nil {
						return nil, err
					}
				}
				ast = v.Value[len(v.Value)-1]
				goto tailcalloptimized
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
					ast = v.Value[2]
					goto tailcalloptimized
				}
				//condition evaluated to false, check if we have a branch for false, and execute it, if so
				if len(v.Value) < 4 {
					return &mal.Nil{}, nil
				}
				ast = v.Value[3]
				goto tailcalloptimized
			case "fn*":
				var bindings []mal.Type
				listBindings, ok := v.Value[1].(*mal.List)
				if ok {
					bindings = listBindings.Value
				}
				if !ok {
					vectorBindings, ok := v.Value[1].(*mal.Vector)
					if ok {
						bindings = vectorBindings.Value
					} else {
						return nil, fmt.Errorf("Invalid bindings to fn*")
					}
				}
				return &mal.Function{
					Ast:    v.Value[2],
					Params: bindings,
					Env:    env,
					Fn: func(args ...mal.Type) (mal.Type, error) {
						fnEnv := mal.NewEnv(env, listBindings.Value, args)
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
		//if we have an AST (and params/env), we can TCO this function!
		if fn.Ast != nil {
			ast = fn.Ast
			//update the env for the function
			env = mal.NewEnv(fn.Env, fn.Params, lst.Value[1:])
			goto tailcalloptimized
		}
		//cannot TCO this (e.g. call to native function)
		return fn.Fn(lst.Value[1:]...)

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
	case *mal.Vector:
		var vector mal.Vector
		for _, val := range v.Value {
			evaled, err := eval(val, env)
			if err != nil {
				return nil, err
			}
			vector.Value = append(vector.Value, evaled)
		}
		return &vector, nil
	default:
		return ast, nil
	}
}

func print(ast mal.Type) {
	fmt.Println(mal.PrString(ast, true))
}

func createREPLEnv() *mal.Env {
	replEnv := mal.NewEnv(nil, nil, nil)
	for k, v := range mal.CoreNS {
		replEnv.Set(k, v)
	}

	// add some stuff that's not in coreNS, according to guide (?)
	replEnv.Set(&mal.Symbol{Value: "eval"}, &mal.Function{Fn: func(args ...mal.Type) (mal.Type, error) {
		return eval(args[0], replEnv)
	}})
	rep("(def! not (fn* (a) (if a false true)))", replEnv, false)
	rep(`(def! load-file (fn* (f) (eval (read-string (str "(do " (slurp f) "\nnil)")))))`, replEnv, false)

	return replEnv
}

func rep(s string, env *mal.Env, doPrint bool) {

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
	if doPrint {
		print(expr)
	}
}

func main() {
	usePlainStdin := flag.Bool("stdin", false, "don't use nice readline based repl. only for tests, as the nice repl breaks them")
	flag.Parse()

	args := flag.Args()

	env := createREPLEnv()

	if len(args) > 0 {
		var argList mal.List
		for _, val := range args[1:] {
			argList.Value = append(argList.Value, val)
		}
		env.Set(&mal.Symbol{Value: "*ARGV*"}, argList)
		rep(`(load-file "`+args[0]+`" )`, env, true)
		return
	}
	env.Set(&mal.Symbol{Value: "*ARGV*"}, &mal.List{})

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
		rep(s, env, true)
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
		rep(s, env, true)
	}
}
