# geroz

geroz-e wraps `exec.Cmd` and allows user to propagate signals to the underlying process with ease.

## Usage

Following example is the simplest signal propagator.

```go
import (
    "github.com/kyslik/geroz"
    "log"
    "os"
)

func main() {
    // initialize command `*exec.Cmd` with sane defaults
    c, e := geroz.Command()
    if e != nil {
        log.Fatalf("failed to initialize command: %w\n", e)
    }
   
    // start `c`
    c, e = geroz.StartProcess(c)
    if e != nil {
        log.Fatalf("failed to start process: %w\n", e)
    }
    
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    // set up propagating signals to the process c
    go geroz.PropagateSignals(ctx, c)

    // wait for the process to exit (blocking)
    _, e := geroz.WaitProcess(c)
    if e != nil {
        log.Fatalf("failed to wait for process to finish: %w\n", e)
    }
}
```

For more examples see directory [examples](./examples).