package main

import (
	"github.com/kyslik/geroz"
	"log"
	"os"
)

func main() {
	// Simulate passing in "env" binary as first argument
	if len(os.Args) == 1 {
		os.Args = append(os.Args, "env")
	}

	c, e := geroz.NewCommand()
	if e != nil {
		log.Fatalf("failed to initialize command: %v\n", e)
	}

	c.Stdout = os.Stdout

	c.Env = []string{"JANE=DOE"}

	_, e = geroz.StartCommand(c)
	if e != nil {
		log.Fatalf("failed to start process: %v\n", e)
	}
}
