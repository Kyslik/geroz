package main

import (
	"github.com/kyslik/geroz"
	"log"
	"os"
)

func main() {
	// emulate passing in "env" binary as first argument
	if len(os.Args) == 1 {
		os.Args = append(os.Args, "env")
	}

	c, e := geroz.Command()
	if e != nil {
		log.Fatalf("failed to initialize command: %w\n", e)
	}

	c.Stdout = os.Stdout

	c.Env = []string{"JANE=DOE"}

	_, e = geroz.StartProcess(c)
	if e != nil {
		log.Fatalf("failed to start process: %w\n", e)
	}
}
