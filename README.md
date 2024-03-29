# geroz

geroz-e wraps `exec.Cmd` and allows user to propagate signals to the underlying process with ease.

## Usage

The following example is the simplest signal propagator.

```go
import (
    "log"
    "os"

    "github.com/kyslik/geroz"
)

func main() {
    // initialize command `*exec.Cmd`
    c, e := geroz.NewCommand()
    if e != nil {
        log.Fatalf("failed to initialize command: %v\n", e)
    }

    // start `c`
    c, e = geroz.StartCommand(c)
    if e != nil {
        log.Fatalf("failed to start command: %v\n", e)
    }

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // set up propagating signals to the process `c`
    go geroz.PropagateSignals(ctx, c)

    // wait for the process to exit (blocking)
    _, e := geroz.WaitCommand(c)
    if e != nil {
        log.Fatalf("failed to wait for command to finish: %w\n", e)
    }
}
```

For more examples see directory [examples](./examples).

## TODO

- [ ] make a test pass when using `go test -race`
- [ ] implement catching zombie processes - <https://github.com/ramr/go-reaper>
