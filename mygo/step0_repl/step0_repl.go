package main

import (
	"bufio"
	"fmt"
	"os"
)

func read() {

}

func eval() {

}

func print() {

}

func rep(s string) {
	fmt.Println(s)
}

func main() {
	stdin := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("user> ")
		s, _ := stdin.ReadString('\n')
		rep(s)
	}
}
