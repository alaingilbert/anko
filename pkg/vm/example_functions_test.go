package vm_test

import (
	"fmt"
	"log"

	"github.com/alaingilbert/anko/pkg/vm"
)

func Example_vmFunctions() {
	v := vm.New(nil)

	err := v.Define("println", fmt.Println)
	if err != nil {
		log.Fatalf("define error: %v\n", err)
	}

	script := `
func a(b) {
	println(b)
}
a("b")

a = func(b) {
	println(b)
}
a("b")

func(b) {
	println(b)
}("b")

func a() {
	return "a"
}
println(a())

println("")


func fib(n) {
	if (n <= 1) {
		return n
	}
	return fib(n - 1) + fib(n - 2) 
}

println(fib(8))
 
 func sum(n...) {
	 t = 0
	 for a in n {
		 t += a
	 }
	 return t
 }
println(sum(1, 2, 3, 4))

func add(a, b) {
	return a + b
} 
println(add([1, 2]...))
`

	_, err = v.Executor().Run(nil, script)
	if err != nil {
		log.Fatalf("execute error: %v\n", err)
	}

	// output:
	// b
	// b
	// b
	// a
	//
	// 21
	// 10
	// 3

}

func Example_vmFunctionsScope() {
	v := vm.New(nil)
	err := v.Define("println", fmt.Println)
	if err != nil {
		log.Fatalf("define error: %v\n", err)
	}

	script := `
a = 1
func () {
	a = 2
}()
println(a)

var a = 1
func () {
	a = 2
}()
println(a)

a = 1
func () {
	var a = 2
}()
println(a)

var a = 1
func () {
	var a = 2
}()
println(a)
`

	_, err = v.Executor().Run(nil, script)
	if err != nil {
		log.Fatalf("execute error: %v\n", err)
	}

	// output:
	// 2
	// 2
	// 1
	// 1

}

func testFunc1(a any) int {
	b, ok := a.([]any)
	if ok {
		return len(b)
	}
	return 0
}

func Example_vmFunctionsOutside() {

	/*
	   // the following function would be uncommented
	   func testFunc1(a any) int {
	   	b, ok := a.([]any)
	   	if ok {
	   		return len(b)
	   	}
	   	return 0
	   }
	*/

	v := vm.New(nil)

	err := v.Define("println", fmt.Println)
	if err != nil {
		log.Fatalf("define error: %v\n", err)
	}
	err = v.Define("addString", func(a string, b string) string { return a + b })
	if err != nil {
		log.Fatalf("define error: %v\n", err)
	}
	// uses the function that would be declared above
	err = v.Define("aFunc", testFunc1)
	if err != nil {
		log.Fatalf("define error: %v\n", err)
	}

	script := `
a = addString("a", "b")
println(a)

a = aFunc([1, 2, 3])
println(a) 
`

	_, err = v.Executor().Run(nil, script)
	if err != nil {
		log.Fatalf("execute error: %v\n", err)
	}

	// output:
	// ab
	// 3

}
