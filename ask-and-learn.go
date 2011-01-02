/*
 * Copyright (c) 2011 Nicolas Thery (nthery@gmail.com)
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 */

// Program that tries to guess an animal chosen by the user and that learns from its mistakes.
package main

import (
	"bufio"
	"fmt"
	"json"
	"log"
	"os"
	"io/ioutil"
	"flag"
	"path"
)

// Known animals are stored in a binary tree that grows over time
type node struct {
	Animal   string // leaf only
	Question string // non-leaf only
	No, Yes  *node  // children
}

func (n *node) isLeaf() bool {
	return n.Animal != ""
}

// Knowledge base root
var root *node

// Default initial tree content when creating new database
var defaultRoot = node{Animal: "platypus"}

// Command-line arguments and flags
var (
	createDbFlag = flag.Bool("c", false, "create new DB")
	dbPath       string
)

var stdin *bufio.Reader

func main() {
	parseCmdLine()
	stdin = bufio.NewReader(os.Stdin)
	initDb()
	playGames()
	saveDb()
}

func parseCmdLine() {
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "database expected\n")
		usage()
		os.Exit(1)
	}
	dbPath = flag.Arg(0)
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s [-c] database-file\n", path.Base(os.Args[0]))
	flag.PrintDefaults()
}

// Create a new database or load an existing one
func initDb() {
	if *createDbFlag {
		root = &defaultRoot
	} else {
		content, err := ioutil.ReadFile(dbPath)
		if err != nil {
			log.Panic("can not read db:", err)
		}
		root = new(node)
		err = json.Unmarshal(content, root)
		if err != nil {
			log.Panic("can not marshal db:", err)
		}
	}
}

// Save the current database to a file
func saveDb() {
	content, err := json.MarshalIndent(root, "", "    ")
	if err != nil {
		log.Panic("can not unmarshal db:", err)
	}

	err = ioutil.WriteFile(dbPath, content, 0700)
	if err != nil {
		log.Panic("can not write db:", err)
	}
}

func playGames() {
	again := true
	for again {
		playOneGame()
		again = askYesNo("Play another game?")
	}
}

func playOneGame() {
	n := root

	for !n.isLeaf() {
		yes := askYesNo(n.Question)
		if yes {
			n = n.Yes
		} else {
			n = n.No
		}
	}

	found := askYesNo("Is it a %s?", n.Animal)
	if !found {
		learnNewAnimal(n)
	}
}

// Ask user how to distinguish n.Animal from user-chosen one and update tree
func learnNewAnimal(n *node) {
	animal := ask("What is the animal I failed to find?")
	leaf := &node{Animal: animal}
	question := ask("What question can distinguish a %s from a %s?", animal, n.Animal)
	isYesLeaf := askYesNo("What answer is expected for a %s?", animal)
	mutateIntoQuestionNode(n, question, leaf, isYesLeaf)
}

// Turn leaf node into a question node
func mutateIntoQuestionNode(n *node, question string, leaf *node, isYesLeaf bool) {
	otherLeaf := &node{Animal: n.Animal}
	n.Animal = ""
	n.Question = question
	if isYesLeaf {
		n.Yes = leaf
		n.No = otherLeaf
	} else {
		n.No = leaf
		n.Yes = otherLeaf
	}
}

// Ask question expecting yes or no answer
func askYesNo(prompt string, args ...interface{}) (yes bool) {
	done := false
	for !done {
		s := ask(prompt, args...)
		switch s {
		case "yes", "y":
			yes = true
			done = true
		case "no", "n":
			yes = false
			done = true
		default:
			// nop
		}
	}
	return
}

// Ask question to user
func ask(prompt string, args ...interface{}) string {
	prompt += " "
	for {
		fmt.Printf(prompt, args...)
		answer, err := stdin.ReadString('\n')
		if err != nil {
			log.Panic("error when reading stdin:", err)
		}
		if len(answer) > 0 && answer[len(answer)-1] == '\n' {
			answer = answer[:len(answer)-1]
		}
		if len(answer) > 0 {
			return answer
		}
	}
	return ""
}
