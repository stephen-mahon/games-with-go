package main

import (
	"fmt"

	. "github.com/stephen-mahon/games-with-go/evolvingpictures/apt"
)

func main() {

	x := &OpX{}
	y := &OpY{}
	plus := &OpPlus{}
	sine := &OpSin{}
	sine.Child = x
	plus.LeftChild = sine
	plus.RightChild = y

	fmt.Println(plus.Eval(5, 2))
	fmt.Println(plus)
}
