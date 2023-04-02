package main

import (
	"log"
	"os"

	"github.com/kyslik/geroz"
)

func main() {
	// simulate passing in "env" binary as the first argument
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
