package main

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
)

func main() {

	window, err := sdl.CreateWindow("Testing SDL2", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, 800, 600, sdl.WINDOW_SHOWN)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer window.Destroy()

	sdl.Delay(2000)

}
