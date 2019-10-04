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

func isMacroCall(ast mal.Type, env *mal.Env) bool {
	astLst, isList := ast.(*mal.List)
	if !isList || len(astLst.Value) == 0 {
		return false
	}
	symbol, hasSymbolFirst := astLst.Value[0].(*mal.Symbol)
	if !hasSymbolFirst {
		return false
	}
	if fn, ok := env.Get(symbol).(*mal.Function); ok {
		return fn.IsMacro
	}
	return false
}

func macroExpand(ast mal.Type, env *mal.Env) (mal.Type, error) {
	for isMacroCall(ast, env) {
		astLst, _ := ast.(*mal.List)
		symbol, _ := astLst.Value[0].(*mal.Symbol)
		fn, _ := env.Get(symbol).(*mal.Function)
		r, err := fn.Fn(astLst.Value[1:]...)
		ast = r
		if err != nil {
			return nil, err
		}
	}
	return ast, nil
}

type tryCatchInfo struct {
	errMsg  *mal.String
	isError bool
}

func eval(ast mal.Type, env *mal.Env) (mal.Type, error) {
tailcalloptimized:
	switch astList := ast.(type) {
	case *mal.List:
		if len(astList.Value) == 0 {
			return ast, nil
		}
		if astList.IsVector { //we want to handle vectors the same as the default case
			return evalAst(astList, env)
		}

		r, err := macroExpand(ast, env)
		if err != nil {
			return nil, err
		}
		astList, isList := r.(*mal.List)

		if !isList {
			return evalAst(r, env)
		}

		// if the first element of the list is a symbol, check for special handling, such as "def!"
		if symb, ok := astList.Value[0].(*mal.Symbol); ok {
			switch symb.Value {
			case "def!":
				//check argument length
				if len(astList.Value) != 3 {
					return nil, fmt.Errorf("'def!' expects exactly 2 paramters")
				}
				return setBindingInEnv(env, astList.Value[1:])
			case "defmacro!":
				evaledFunction, err := eval(astList.Value[2], env)
				if err != nil {
					return nil, err
				}
				if fn, ok := evaledFunction.(*mal.Function); ok {
					fn.IsMacro = true
					symb, _ := astList.Value[1].(*mal.Symbol)
					env.Set(symb, fn)
					return fn, nil
				}
				return nil, fmt.Errorf("Argument 2 to defmacro! must be a function")
			case "let*":
				newEnv := mal.NewEnv(env, nil, nil)
				if len(astList.Value) < 3 {
					return nil, fmt.Errorf("'let*' expects at least 2 paramters")
				}
				if bindings, ok := astList.Value[1].(*mal.List); ok {
					for i := 0; i < len(bindings.Value)/2; i++ {
						idx := (i * 2)
						setBindingInEnv(newEnv, bindings.Value[idx:idx+2])
					}

					env = newEnv
					ast = astList.Value[2]
					goto tailcalloptimized
				}
				return nil, fmt.Errorf("'let!': invalid arguments")
			case "do":
				for _, val := range astList.Value[1 : len(astList.Value)-1] {
					var err error
					_, err = eval(val, env)
					if err != nil {
						return nil, err
					}
				}
				ast = astList.Value[len(astList.Value)-1]
				goto tailcalloptimized
			case "if":
				r, err := eval(astList.Value[1], env)
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
					ast = astList.Value[2]
					goto tailcalloptimized
				}
				//condition evaluated to false, check if we have a branch for false, and execute it, if so
				if len(astList.Value) < 4 {
					return &mal.Nil{}, nil
				}
				ast = astList.Value[3]
				goto tailcalloptimized
			case "fn*":
				var bindings []mal.Type
				listBindings, ok := astList.Value[1].(*mal.List)
				if ok {
					bindings = listBindings.Value
				} else {
					return nil, fmt.Errorf("Invalid bindings to fn*")
				}

				return &mal.Function{
					Ast:    astList.Value[2],
					Params: bindings,
					Env:    env,
					Fn: func(args ...mal.Type) (mal.Type, error) {
						fnEnv := mal.NewEnv(env, listBindings.Value, args)
						r, err := eval(astList.Value[2], fnEnv)
						return r, err
					}}, nil
			case "quote":
				return astList.Value[1], nil
			case "quasiquote":
				ast = quasiquote(astList.Value[1])
				goto tailcalloptimized
			case "macroexpand":
				return macroExpand(astList.Value[1], env)
			case "try*":
				r, err := eval(astList.Value[1], env)
				if err != nil {
					catchBlock, _ := astList.Value[2].(*mal.List)
					if symb, ok := catchBlock.Value[0].(*mal.Symbol); ok && symb.Value == "catch*" {
						bind, _ := catchBlock.Value[1].(*mal.Symbol)
						exEnv := mal.NewEnv(env, nil, nil)
						exEnv.Set(bind, &mal.String{Value: err.Error()})
						ast = catchBlock.Value[2]
						env = exEnv
						goto tailcalloptimized
					}
				}
				return r, err
			case "throw":
				errStr, _ := astList.Value[1].(*mal.String)
				return nil, fmt.Errorf(errStr.Value)
			}
		}
		ev, err := evalAst(astList, env)
		if err != nil {
			return nil, err
		}
		lst, _ := ev.(*mal.List)
		fn, isFN := lst.Value[0].(*mal.Function)
		if !isFN {
			return nil, fmt.Errorf("Expected function, got %T", lst.Value[0])
		}
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
		return evalAst(astList, env)
	}

}

func evalAst(ast mal.Type, env *mal.Env) (mal.Type, error) {
	switch v := ast.(type) {
	case *mal.Symbol:
		val := env.Get(v)
		if val == nil {
			return nil, fmt.Errorf("'%s' not found", v.Value)
		}
		return val, nil
	case *mal.List:
		list := mal.NewList(v.IsVector)
		for _, val := range v.Value {
			evaled, err := eval(val, env)
			if err != nil {
				return nil, err
			}
			list.Value = append(list.Value, evaled)
		}
		return &list, nil
	case *mal.HashMap:
		hmap := mal.NewHashMap()
		for key, val := range v.Value {
			evaled, err := eval(val, env)
			if err != nil {
				return nil, err
			}
			hmap.Value[key] = evaled
		}
		return &hmap, nil
	default:
		return ast, nil
	}
}

//Uh yeah... I just implemented https://github.com/kanaka/mal/blob/master/process/guide.md#step7
//I haven't tried understanding this function in detail yet
func quasiquote(ast mal.Type) mal.Type {
	if !isPair(ast) {
		newLst := mal.NewList(false)
		newLst.Value = append(newLst.Value, &mal.Symbol{Value: "quote"})
		newLst.Value = append(newLst.Value, ast)
		return &newLst
	}
	astLst, _ := ast.(*mal.List)
	if symbol, ok := astLst.Value[0].(*mal.Symbol); ok && symbol.Value == "unquote" {
		return astLst.Value[1]
	}

	if isPair(astLst.Value[0]) {
		if l2, ok := astLst.Value[0].(*mal.List); ok && isPair(l2) {
			if symb, ok := l2.Value[0].(*mal.Symbol); ok && symb.Value == "splice-unquote" {
				newLst := mal.NewList(false)
				newLst.Value = append(newLst.Value, &mal.Symbol{Value: "concat"})
				newLst.Value = append(newLst.Value, l2.Value[1])
				tmp := mal.NewList(false)
				tmp.Value = append(tmp.Value, astLst.Value[1:]...)
				newLst.Value = append(newLst.Value, quasiquote(&tmp))
				return &newLst
			}
		}
	}

	newLst := mal.NewList(false)
	newLst.Value = append(newLst.Value, &mal.Symbol{Value: "cons"})
	newLst.Value = append(newLst.Value, quasiquote(astLst.Value[0]))
	tmp := mal.NewList(false)
	tmp.Value = append(tmp.Value, astLst.Value[1:]...)
	newLst.Value = append(newLst.Value, quasiquote(&tmp))
	return &newLst
}

func isPair(ast mal.Type) bool {
	if lst, ok := ast.(*mal.List); ok {
		if len(lst.Value) != 0 {
			return true
		}
	}
	return false
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
	rep(`(defmacro! cond (fn* (& xs) (if (> (count xs) 0) (list 'if (first xs) (if (> (count xs) 1) (nth xs 1) (throw "odd number of forms to cond")) (cons 'cond (rest (rest xs)))))))`, replEnv, false)

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
		fmt.Fprintln(os.Stderr, "Error: "+err.Error())
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
		argList := mal.NewList(false)
		for _, val := range args[1:] {
			argList.Value = append(argList.Value, &mal.String{Value: val})
		}
		env.Set(&mal.Symbol{Value: "*ARGV*"}, &argList)
		rep(`(load-file "`+args[0]+`" )`, env, false)
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
