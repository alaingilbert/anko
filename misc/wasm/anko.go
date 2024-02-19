//go:build js && wasm

package main

import (
	"context"
	"fmt"
	"github.com/alaingilbert/anko/pkg/parser"
	"github.com/alaingilbert/anko/pkg/utils"
	"github.com/alaingilbert/anko/pkg/vm"
	"github.com/alaingilbert/anko/pkg/vm/runner"
	"html"
	"strings"
	"syscall/js"
)

var (
	result = js.Global().Get("document").Call("getElementById", "result")
	input  = js.Global().Get("document").Call("getElementById", "input")
)

func writeCommand(s string) {
	result.Set("innerHTML", result.Get("innerHTML").String()+"<p class='command'>"+html.EscapeString(s)+"</p>")
	result.Set("scrollTop", result.Get("scrollHeight").Int())
}

func writeStdout(s string) {
	result.Set("innerHTML", result.Get("innerHTML").String()+"<p class='stdout'>"+html.EscapeString(s)+"</p>")
	result.Set("scrollTop", result.Get("scrollHeight").Int())
}

func writeStderr(s string) {
	result.Set("innerHTML", result.Get("innerHTML").String()+"<p class='stderr'>"+html.EscapeString(s)+"</p>")
	result.Set("scrollTop", result.Get("scrollHeight").Int())
}

func main() {
	v := vm.New(&vm.Config{DefineImport: utils.Bool(true)})

	_ = v.Define("print", func(a ...any) {
		writeStdout(fmt.Sprint(a...))
	})
	_ = v.Define("printf", func(a string, b ...any) {
		writeStdout(fmt.Sprintf(a, b...))
	})

	e := v.Executor()

	var source string

	parser.EnableErrorVerbose()

	ch := make(chan string)

	input.Call("addEventListener", "keypress", js.FuncOf(func(_ js.Value, args []js.Value) any {
		e := args[0]
		if e.Get("keyCode").Int() != 13 {
			return nil
		}
		s := e.Get("target").Get("value").String()
		e.Get("target").Set("value", "")
		writeCommand(s)
		ch <- s
		return nil
	}))
	input.Set("disabled", false)
	result.Set("innerHTML", "")

	go func() {
		for {
			text, ok := <-ch
			if !ok {
				break
			}
			source += text
			if source == "" {
				continue
			}
			if source == "quit()" {
				break
			}

			stmts, err := parser.ParseSrc(source)

			if e, ok := err.(*parser.Error); ok {
				es := e.Error()
				if strings.HasPrefix(es, "syntax error: unexpected") {
					if strings.HasPrefix(es, "syntax error: unexpected $end,") {
						continue
					}
				} else {
					if e.Pos.Column == len(source) && !e.Fatal {
						writeStderr(e.Error())
						continue
					}
					if e.Error() == "unexpected EOF" {
						continue
					}
				}
			}

			source = ""
			var v any

			if err == nil {
				v, err = e.Run(context.Background(), stmts)
			}
			if err != nil {
				if e, ok := err.(*runner.Error); ok {
					writeStderr(fmt.Sprintf("%d:%d %s\n", e.Pos.Line, e.Pos.Column, err))
				} else if e, ok := err.(*parser.Error); ok {
					writeStderr(fmt.Sprintf("%d:%d %s\n", e.Pos.Line, e.Pos.Column, err))
				} else {
					writeStderr(err.Error())
				}
				continue
			}

			writeStdout(fmt.Sprintf("%#v\n", v))
		}
	}()

	select {}

}
