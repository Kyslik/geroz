package geroz

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

// Command parses `os.Args` slice and returns a struct of `*exec.Cmd`.
func Command() (*exec.Cmd, error) {
	var cmd *exec.Cmd
	switch len(os.Args) {
	case 0, 1:
		return nil, fmt.Errorf("provide a binary which should be started")
	case 2:
		cmd = exec.Command(os.Args[1])
	default:
		cmd = exec.Command(os.Args[1], os.Args[2:]...)
	}

	return cmd, nil
}

// StartProcess tries to start `cmd`. By setting up `Setpgid` to false,
// we can actually propagate signals.
func StartProcess(cmd *exec.Cmd) (*exec.Cmd, error) {
	// make sure not to not set Process Group ID
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Setpgid = false

	err := cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("cmd.Start: %w", err)
	}

	return cmd, nil
}

// PropagateSignals tries to propagate signals until ctx is closed or process `cmd` is finished.
// `cmd` is considered "finished" once propagating signal fails.
func PropagateSignals(ctx context.Context, cmd *exec.Cmd) {
	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel)

	defer func() {
		signal.Stop(signalChannel)
		close(signalChannel)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case sig := <-signalChannel:
			err := cmd.Process.Signal(sig)
			if err != nil {
				return
			}
		}
	}
}

// WaitProcess is a blocking function that waits for cmd to exit and returns its exit code
// if an error is non-nil, returned status code is 0
func WaitProcess(cmd *exec.Cmd) (int, error) {
	err := cmd.Wait()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, okk := exitErr.Sys().(syscall.WaitStatus); okk {
				return status.ExitStatus(), nil
			}
		}
		return 0, err
	}
	return cmd.ProcessState.ExitCode(), nil
}
