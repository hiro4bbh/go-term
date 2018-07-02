package main

import (
	"fmt"
	"os"

	"github.com/hiro4bbh/go-log"
	"github.com/hiro4bbh/go-term"
)

func main() {
	promptStyle := golog.FgBlack
	if goterm.IsTerminal(os.Stdout) {
		promptStyle = promptStyle.Bold(true)
	}
	term, err := goterm.NewTerm(os.Stdin, promptStyle.Sprintf("> "), goterm.TermConfig{
		History: true,
	})
	if err != nil {
		panic(err)
	}
	for {
		line, err := term.ReadLine()
		if err != nil {
			panic(err)
		} else if line == "" {
			break
		}
		fmt.Printf("you typed> %q\n", line)
	}
}
