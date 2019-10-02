package main

import (
	"bufio"
	"fmt"
	"mygomal/mal"
	"os"
)

func read() {

}

func eval() {

}

func print() {

}

func rep(s string) {
	ast, err := mal.ReadStr(s)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
	fmt.Println(mal.PrString(ast, true))
}

func main() {
	stdin := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("user> ")
		s, _ := stdin.ReadString('\n')
		rep(s)
	}
}
