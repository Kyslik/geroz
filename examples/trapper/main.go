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

	c, e := geroz.Command()
	if e != nil {
		log.Fatalf("failed to initialize command: %w\n", e)
	}

	// TODO: consider adding this to the `geroz.Command()`
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	c, e = geroz.StartProcess(c)
	if e != nil {
		log.Fatalf("failed to start process: %w\n", e)
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

	sc, e := geroz.WaitProcess(c)
	if e != nil {
		log.Fatalf("failed to wait for process to finish: %w\n", e)
	}

	fmt.Println("child process exited with: ", sc)
}
