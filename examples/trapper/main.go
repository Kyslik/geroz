package main

import (
	"context"
	"fmt"
	"github.com/kyslik/geroz"
	"log"
	"os"
)

func main() {
	// simulate passing in "./trapper.sh" script as first argument
	if len(os.Args) == 1 {
		os.Args = append(os.Args, "./trapper.sh")
	}

	c, e := geroz.NewCommand()
	if e != nil {
		log.Fatalf("failed to initialize command: %v\n", e)
	}

	// TODO: consider adding this to the `geroz.NewCommand()`
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	c, e = geroz.StartCommand(c)
	if e != nil {
		log.Fatalf("failed to start process: %v\n", e)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go geroz.PropagateSignals(ctx, c)

	self := os.Getpid()
	fmt.Printf("send signals to PID %d in order to propagate...\n", self)
	fmt.Printf("\tkill -INT %d (or ctrl+c)\n", self)
	fmt.Printf("\tkill -QUIT %d\n", self)
	fmt.Printf("\tkill -URG %d\n", self)
	fmt.Printf("\tkill -WINCH %d\n", self)
	fmt.Printf("\tkill -USR1 %d\n", self)
	fmt.Printf("\tkill -USR2 %d\n", self)
	fmt.Printf("hit enter ↵ to exit\n")

	sc, e := geroz.WaitCommand(c)
	if e != nil {
		log.Fatalf("failed to wait for process to finish: %v\n", e)
	}

	fmt.Println("child process exited with: ", sc)
}
