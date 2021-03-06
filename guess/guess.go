package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	low := 1
	high := 100
	var tries int

	fmt.Println("Think of a number between", low, "and", high)
	fmt.Println("Press ENTER when ready")
	scanner.Scan()

	for {
		// Binary search strategy
		guess := (low + high) / 2
		fmt.Println("I guess the number is", guess)
		tries++
		fmt.Println("This that:")
		fmt.Println("(a) too high?")
		fmt.Println("(b) too low?")
		fmt.Println("(c) correct?")
		scanner.Scan()
		response := scanner.Text()

		if response == "a" {
			high = guess - 1
		} else if response == "b" {
			low = guess + 1
		} else if response == "c" {
			fmt.Println("I won!")
			fmt.Println("It took me", tries, "tries")
			break
		} else {
			fmt.Println("Invalid response, try again")
		}

		if high < low || low > high {
			fmt.Println("I think you are cheating!")
			break
		}

	}
}
