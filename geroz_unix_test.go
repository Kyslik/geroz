package geroz_test

import (
	"bufio"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/kyslik/geroz"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"reflect"
	"runtime"
	"syscall"
	"testing"
	"time"
)

// Disable output of command c
func disableOutput(c *exec.Cmd) {
	c.Stdout = nil
	c.Stderr = nil
}

// Get command instance and disable output
func commandWithDisabledOutput() (*exec.Cmd, error) {
	c, e := geroz.Command()
	if e == nil {
		disableOutput(c)
	}
	return c, e
}

// ok fails the test if an err is not nil
func ok(tb testing.TB, err error) {
	tb.Helper()
	if err != nil {
		tb.Fatalf("unexpected error: %s", err.Error())
	}
}

// nok fails the test is an err is nil
func nok(tb testing.TB, err error) {
	tb.Helper()
	if err == nil {
		tb.Fatalf("expected error got nil")
	}
}

// equals fails the test if exp is not equal to act.
func equals(tb testing.TB, exp, act interface{}) {
	tb.Helper()
	if !reflect.DeepEqual(exp, act) {
		tb.Fatalf("exp: %#v\n\n\tgot: %#v", exp, act)
	}
}

func TestCommandFromArgs(t *testing.T) {
	var commandArgsTests = []struct {
		scenario string
		in       []string
		out      []string
	}{
		{"binary", []string{"self", "/bin/ls"}, []string{"/bin/ls"}},
		{"binary with argument", []string{"self", "/bin/ls", "-l"}, []string{"/bin/ls", "-l"}},
		{"binary with arguments", []string{"self", "/bin/ls", "-l", "./"}, []string{"/bin/ls", "-l", "./"}},
	}

	for _, tt := range commandArgsTests {
		t.Run(tt.scenario, func(t *testing.T) {
			os.Args = tt.in
			c, _ := geroz.Command()
			equals(t, c.Args, tt.out)
			equals(t, c.Path, tt.out[0])
		})
	}
}

func TestCommandEmptyArgs(t *testing.T) {
	os.Args = []string{"self"}
	_, e := geroz.Command()
	nok(t, e)
}

func TestStartProcess(t *testing.T) {
	if _, err := exec.LookPath("/bin/ls"); err != nil {
		t.Log("/bin/ls not found, skipping")
		t.Skip()
	}

	os.Args = []string{"self", "/bin/ls", "-lah"}

	c, e := commandWithDisabledOutput()
	ok(t, e)

	c, e = geroz.StartProcess(c)
	ok(t, e)

	// call c.Wait() here so we can check ProcessState
	e = c.Wait()
	ok(t, e)

	if !c.ProcessState.Success() {
		t.Errorf("got %v, want true", c.ProcessState.Success())
	}
}

// Test bubbling error of starting a process to the top
func TestStartProcessFails(t *testing.T) {
	os.Args = []string{"self", "/dev/null"}
	c, e := commandWithDisabledOutput()
	ok(t, e)

	_, e = geroz.StartProcess(c)
	nok(t, e)
}

// This test is hard to understand, it builds a test binary and runs itself.
// TODO: add detailed description of this test, including a sequence diagram.
func TestPropagateSignals(t *testing.T) {
	// Catch calling the testing binary
	if os.Getenv("GO_TEST_PROPAGATE_SIGNALS") == "1" {
		signalCatcher()
		// `go test` prints "PASS" on exit
		os.Exit(0)
	}

	testBin := buildTestBinary(t)
	os.Args = append([]string{""}, "./"+testBin, "-test.run", "TestPropagateSignals")

	var signalTests = []struct {
		signal syscall.Signal
	}{
		{syscall.SIGTERM},
		{syscall.SIGHUP},
		{syscall.SIGINT},
		{syscall.SIGQUIT},
		{syscall.SIGWINCH},
	}

	for _, tt := range signalTests {
		t.Run("Scenario:"+tt.signal.String(), func(t *testing.T) {
			tablePropagateSignals(t, tt.signal)
		})
	}
}

// When called waits for a signal and blocks imaginary binary,
// that we want to propagate signals to
func signalCatcher() {
	// Cleanup in case parent gets killed by SIGKILL :(
	time.AfterFunc(100*time.Millisecond, func() { os.Exit(125) })

	signalChannel := make(chan os.Signal, 2)
	defer func() {
		signal.Stop(signalChannel)
		close(signalChannel)
	}()

	signal.Notify(signalChannel)
	fmt.Printf("%v\n", <-signalChannel)
}

func buildTestBinary(t *testing.T) string {
	// Get name of current file being executed
	_, file, _, k := runtime.Caller(1)
	if !k {
		t.Fatalf("Could not get name of the test file")
		return ""
	}

	// Calculate sha1 of the file currently run file
	f, err := os.Open(file)
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer f.Close()

	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		t.Fatalf(err.Error())
	}

	testDir := "./.tc/"
	testBinFilename := testDir + hex.EncodeToString(h.Sum(nil)[:8])

	// Check if binary already exists, if yes return
	_, err = os.Stat(testBinFilename)
	if !os.IsNotExist(err) {
		return testBinFilename
	}

	// Remove contents of testDir directory
	dir, err := ioutil.ReadDir(testDir)
	if err == nil {
		for _, d := range dir {
			e := os.RemoveAll(path.Join([]string{testDir, d.Name()}...))
			ok(t, e)
		}
	}
	// Build test binary for invoking later
	cmd := exec.Command("go", "test", "-c", "-o", testBinFilename)

	e := cmd.Start()
	ok(t, e)

	e = cmd.Wait()
	ok(t, e)

	return testBinFilename
}

func tablePropagateSignals(t *testing.T, signal syscall.Signal) {
	c, e := geroz.Command()
	ok(t, e)

	cmdOut, e := c.StdoutPipe()
	defer func() {
		_ = cmdOut.Close()
	}()
	ok(t, e)

	go func(t *testing.T, s *bufio.Scanner) {
		// Tightly coupled with signalCatcher(), expects to scan only once
		ran := false
		for s.Scan() {
			equals(t, signal.String(), s.Text())
			ran = true
			break
		}
		// Make sure we fail if we did not run the test
		defer func() {
			if !ran {
				t.Fail()
			}
		}()
	}(t, bufio.NewScanner(cmdOut))

	c.Env = append([]string{}, "GO_TEST_PROPAGATE_SIGNALS=1")
	c, e = geroz.StartProcess(c)
	defer c.Process.Kill()
	ok(t, e)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go geroz.PropagateSignals(ctx, c)

	// Wait for testBin to start
	time.Sleep(10 * time.Millisecond)

	self, e := os.FindProcess(os.Getpid())
	ok(t, e)

	e = self.Signal(signal)
	ok(t, e)

	_, e = geroz.WaitProcess(c)
	ok(t, e)
}
