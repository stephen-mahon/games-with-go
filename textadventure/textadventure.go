package main

// some ideas
// Use Tower of Ghenjei from WoT ToM
//		Structure of rooms for escape needs to go back on itself
//		Riddles from Eelfinn and Aelfinn
// Add NPC - talk, fight pokemon style
// 		Maybe start with a note on the round
//		Fight grue when room is lit and have sword
// NPC move around
//		Grue moves to next dark room
// items that can be picked up or placed down
//		Add lamp and key
// Parse natural language as input. Close approximations verbs, nouns, etc
//		Maybe later.

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type choice struct {
	cmd         string
	description string
	nextNode    *storyNode
}

type storyNode struct {
	text    string
	choices []*choice
}

func (node *storyNode) addChoice(cmd string, description string, nextNode *storyNode) {
	choice := &choice{cmd, description, nextNode}
	node.choices = append(node.choices, choice)
}

func (node *storyNode) render() {
	fmt.Println(node.text)
	if node.choices != nil {
		for _, choice := range node.choices {
			fmt.Println(choice.cmd, ":", choice.description)
		}
	}
}

func (node *storyNode) executeCmd(cmd string) *storyNode {
	for _, choice := range node.choices {
		if strings.ToLower(choice.cmd) == strings.ToLower(cmd) {
			return choice.nextNode
		}
	}
	fmt.Println("Sorry I didn't understand that.")
	return node
}

var scanner *bufio.Scanner

func (node *storyNode) play() {
	node.render()
	if node.choices != nil {
		scanner.Scan()
		node.executeCmd(scanner.Text()).play()
	}
}

func main() {
	scanner = bufio.NewScanner(os.Stdin)
	start := storyNode{text: `
	You are in a large chamber, deep underground.
	You see three passages leading out.
	A north passage leads into darkness.
	To the south, a passage appears to head upwards.
	The eastern passage appeats flat and well travelled.`}

	darkRoom := storyNode{text: "It is pitch black. You cannot see"}

	darkRoomLit := storyNode{text: "The passage is now lit by your latern. You can continue North or head back South "}

	grue := storyNode{text: "While stumbling around in the darkness, you are eaten by a grue"}

	trap := storyNode{text: "You head down the well travelled path when suddenly a trap door opens and you fall into a pit."}

	treasure := storyNode{text: "You arrive at a small chamber, filled with treasure!"}

	start.addChoice("N", "Go North", &darkRoom)
	start.addChoice("S", "Go South", &darkRoom)
	start.addChoice("E", "Go East", &trap)

	darkRoom.addChoice("S", "Try to go back south", &grue)
	darkRoom.addChoice("O", "Turn on latern", &darkRoomLit)

	darkRoomLit.addChoice("N", "Go North", &treasure)
	darkRoomLit.addChoice("S", "Go South", &start)

	start.play()

	fmt.Println()
	fmt.Println("The End.")

}
