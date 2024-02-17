# Anko

[![GoDoc Reference](https://godoc.org/github.com/alaingilbert/anko/vm?status.svg)](http://godoc.org/github.com/alaingilbert/anko/vm)

Anko is a scriptable interpreter written in Go.

This fork can:
- Compile & execute scripts to bytecode (This is perfect if you don't want to share the sources of your script)
- Validate script syntax before running it
- Know if a function is being used in a script (works with bytecode too)
- Optional typed function parameters and return values
- pause/resume execution of a script
- Stop a script at any time
- Much better CLI with auto-completion
- Get stats, how many statements/expressions were processed
- Rate limit how many expressions to process per "duration" (example: 1_000/sec)
- Support "select" statement

## Usage Example - Embedded

```go
package main

import (
    "context"
    "fmt"
    "github.com/alaingilbert/anko/pkg/vm"
    "time"
)

func sleepMs(ms int) {
    time.Sleep(time.Duration(ms) * time.Millisecond)
}

func unused() {}

func main() {
    ctx := context.Background()
    // Build a virtual machine with wanted functions
    v := vm.New(nil)
    _ = v.Define("println", fmt.Println)
    _ = v.Define("sleep", sleepMs)
    _ = v.Define("unused", unused)
    // Each executor will have its own isolated environment
    executor := v.Executor()
    // Execute a script
    script := `println("Hello World :)")`
    _, _ = executor.Run(ctx, script) // output: Hello World :)

    fmt.Println("-------------------------------------------------------------------------------")
    script = `
i = 0
for {
    println("Iteration #", i++)
    sleep(500)
}
`
    // Start the script in a different thread
    executor.RunAsync(context.Background(), script)
    time.Sleep(time.Second) // Will print 2 times: `Iteration #1` `Iteration #2`
    executor.Pause()        // We will pause the execution for 1s
    time.Sleep(time.Second) // ...
    executor.Resume()       // then resume
    time.Sleep(time.Second) // Will print 2 more times from where it left: `Iteration #3` `Iteration #4`
    executor.Stop()         // and stop the script

    fmt.Println("-------------------------------------------------------------------------------")

    // We can validate a script. If the script has syntax error anywhere in it,
    // Validate will return an error.
    // Validate does not "run" or "execute" the script.
    fmt.Println(v.Validate(ctx, script)) // No error printed since the script is valid

    // We can get information to know if a function is used in a script
    // This is useful if we have a "dangerous" function available,
    // and we want to know if a bytecode script is using it for example.
    // Output: [true false true] <nil>
    fmt.Println(v.Has(ctx, script, []any{sleepMs, unused, fmt.Println}))

    fmt.Println("-------------------------------------------------------------------------------")

    // We can throttle the speed of a script by introducing a rate limit
    // on the number of instructions allowed to be executed. Here would
    // be 10/sec (time.Second is the default RateLimitPeriod)
    v = vm.New(&vm.Configs{RateLimit: 10})
    _ = v.Define("println", fmt.Println)
    script = `
for i=0; i<10; i++ {
    println("Hello world", i)
}`
    _, _ = v.Executor().Run(ctx, script)
}
```

## Usage Example - Command Line

### Compiling a anko script to bytecode
```
./anko -o script.bnk script.ank
```

### Running an Anko script file named script.ank (or bytecode file script.bnk)
```
./anko script.ank
./anko script.bnk
```

## Anko Script Quick Start
```
// declare variables
x = 1
y = x + 1

// print using outside the script defined println function
println(x + y) // 3

// if else statement
if x < 1 || y < 1 {
    println(x)
} else if x < 1 && y < 1 {
    println(y)
} else {
    println(x + y)
}

// array
a = [1, 2, 3]
println(a) // [1 2 3]
println(a[0]) // 1

// map
a = {"x": 1}
println(a) // map[x:1]
a.b = 2
a["c"] = 3
println(a["b"]) // 2
println(a.c) // 3

// function
func a (x) {
    println(x + 1)
}
a(3) // 4
```
