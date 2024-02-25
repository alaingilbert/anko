package runner_test

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/alaingilbert/anko/pkg/vm"
)

func ExampleInterrupt() {
	var waitGroup sync.WaitGroup
	waitGroup.Add(1)
	waitChan := make(chan struct{}, 1)

	v := vm.New(nil)
	sleepMillisecond := func() { time.Sleep(time.Millisecond) }

	err := v.Define("println", fmt.Println)
	if err != nil {
		log.Fatalf("define error: %v\n", err)
	}
	err = v.Define("sleep", sleepMillisecond)
	if err != nil {
		log.Fatalf("define error: %v\n", err)
	}

	script := `
# sleep for 10 seconds
for i = 0; i < 10000; i++ {
	sleep()
}
# Should interrupt before printing the next line
println("this line should not be printed")
`
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		close(waitChan)
		vv, err := v.Executor(nil).Run(ctx, script)
		fmt.Println(vv, err)
		waitGroup.Done()
	}()

	<-waitChan
	cancel()

	waitGroup.Wait()

	// output: <nil> context canceled
}

func ExampleEnv_Define() {
	v := vm.New(nil)

	err := v.Define("println", fmt.Println)
	if err != nil {
		log.Fatalf("define error: %v\n", err)
	}

	err = v.Define("a", true)
	if err != nil {
		log.Fatalf("define error: %v\n", err)
	}
	err = v.Define("b", int64(1))
	if err != nil {
		log.Fatalf("define error: %v\n", err)
	}
	err = v.Define("c", float64(1.1))
	if err != nil {
		log.Fatalf("define error: %v\n", err)
	}
	err = v.Define("d", "d")
	if err != nil {
		log.Fatalf("define error: %v\n", err)
	}
	err = v.Define("e", []any{true, int64(1), float64(1.1), "d"})
	if err != nil {
		log.Fatalf("define error: %v\n", err)
	}
	err = v.Define("f", map[string]any{"a": true})
	if err != nil {
		log.Fatalf("define error: %v\n", err)
	}

	script := `
println(a)
println(b)
println(c)
println(d)
println(e)
println(f)
`

	_, err = v.Executor(nil).Run(nil, script)
	if err != nil {
		log.Fatalf("execute error: %v\n", err)
	}

	// output:
	// true
	// 1
	// 1.1
	// d
	// [true 1 1.1 d]
	// map[a:true]
}

func Example_vmHelloWorld() {
	v := vm.New(nil)

	err := v.Define("println", fmt.Println)
	if err != nil {
		log.Fatalf("define error: %v\n", err)
	}

	script := `
println("Hello World :)")
`

	_, err = v.Executor(nil).Run(nil, script)
	if err != nil {
		log.Fatalf("execute error: %v\n", err)
	}

	// output: Hello World :)
}

func Example_vmQuickStart() {
	v := vm.New(nil)

	err := v.Define("println", fmt.Println)
	if err != nil {
		log.Fatalf("define error: %v\n", err)
	}

	script := `
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
`

	_, err = v.Executor(nil).Run(nil, script)
	if err != nil {
		log.Fatalf("execute error: %v\n", err)
	}

	// output:
	// 3
	// 3
	// [1 2 3]
	// 1
	// map[x:1]
	// 2
	// 3
	// 4
}

func Example_vmDbg() {
	v := vm.New(nil)

	err := v.Define("println", fmt.Println)
	if err != nil {
		log.Fatalf("define error: %v\n", err)
	}

	script := `
dbg()
`

	_, err = v.Executor(nil).Run(nil, script)
	if err != nil {
		log.Fatalf("execute error: %v\n", err)
	}

	// output:
	// No parent
	// println = func([]any) (int, error)
}
