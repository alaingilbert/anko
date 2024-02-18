//go:build !appengine

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/alaingilbert/anko/pkg/compiler"
	"github.com/alaingilbert/anko/pkg/decompiler"
	"github.com/alaingilbert/anko/pkg/parser"
	"github.com/alaingilbert/anko/pkg/utils"
	"github.com/alaingilbert/anko/pkg/utils/pubsub"
	"github.com/alaingilbert/anko/pkg/vm"
	envPkg "github.com/alaingilbert/anko/pkg/vm/env"
	"github.com/alaingilbert/anko/pkg/vm/executor"
	"github.com/alaingilbert/anko/pkg/vm/runner"
	vmUtils "github.com/alaingilbert/anko/pkg/vm/utils"
	"github.com/chzyer/readline"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"sync/atomic"
	"time"
)

const (
	version         = "0.0.1"
	ankoFileExt     = ".ank"
	ankoBytecodeExt = ".bnk"
)

type AppFlags struct {
	FlagExecute string
	File        string
	Compile     bool
	Decompile   bool
	Web         bool
}

func main() {
	var exitCode int
	var appFlags AppFlags
	args := parseFlags(&appFlags)
	if appFlags.Decompile {
		sourceBytes, err := os.ReadFile(appFlags.File)
		if err != nil {
			fmt.Println("ReadFile error:", err)
			os.Exit(ReadFileErrExitCode)
		}
		stmt := compiler.Decode(sourceBytes)

		fmt.Println(decompiler.Decompile(stmt))
		os.Exit(OkExitCode)
	}
	if appFlags.FlagExecute != "" || flag.NArg() > 0 {
		exitCode = runNonInteractive(args, appFlags)
	} else if appFlags.Web {
		exitCode = runWeb()
	} else {
		exitCode = runInteractive(args)
	}
	os.Exit(exitCode)
}

func parseFlags(appFlags *AppFlags) (args []string) {
	flagVersion := flag.Bool("v", false, "prints out the version and then exits")
	flag.StringVar(&appFlags.FlagExecute, "e", "", "execute the Anko code")
	flag.BoolVar(&appFlags.Compile, "c", false, "compile a script")
	flag.BoolVar(&appFlags.Decompile, "d", false, "decompile anko bytecode")
	flag.BoolVar(&appFlags.Web, "w", false, "web server")
	flag.Parse()

	if *flagVersion {
		fmt.Println(version)
		os.Exit(OkExitCode)
	}

	if appFlags.FlagExecute != "" || flag.NArg() < 1 {
		args = flag.Args()
		return
	}

	appFlags.File = flag.Arg(0)
	args = flag.Args()[1:]
	return
}

const (
	OkExitCode          = 0
	ReadFileErrExitCode = 2
	ExecuteErrExitCode  = 4
	CompileErrExitCode  = 5
	ScannerErrExitCode  = 12
)

func runNonInteractive(args []string, appFlags AppFlags) int {
	var source string
	if appFlags.FlagExecute != "" {
		source = appFlags.FlagExecute
	} else {
		sourceBytes, err := os.ReadFile(appFlags.File)
		if err != nil {
			fmt.Println("ReadFile error:", err)
			return ReadFileErrExitCode
		}
		source = string(sourceBytes)

		if appFlags.Compile {
			if err := compileAndSave(source, appFlags.File); err != nil {
				handleErr(os.Stdout, err)
				return CompileErrExitCode
			}
			return OkExitCode
		}
	}

	v := vm.New(&vm.Config{ImportCore: true, DefineImport: true})
	_ = v.Define("args", args)
	executor := v.Executor()
	fileExt := filepath.Ext(appFlags.File)
	var err error
	if appFlags.FlagExecute != "" || fileExt == ankoFileExt {
		_, err = executor.Run(nil, source)
	} else {
		_, err = executor.Run(nil, []byte(source))
	}
	if err != nil {
		handleErr(os.Stdout, err)
		return ExecuteErrExitCode
	}

	return OkExitCode
}

func usage(w io.Writer) {
	_, _ = io.WriteString(w, "commands:\n")
	_, _ = io.WriteString(w, completer.Tree("    "))
}

// Function constructor - constructs new function for listing given directory
func listFiles(path string) func(string) []string {
	return func(line string) []string {
		names := make([]string, 0)
		files, _ := os.ReadDir(path)
		for _, f := range files {
			names = append(names, f.Name())
		}
		return names
	}
}

var base = []readline.PrefixCompleterInterface{
	readline.PcItem("mode",
		readline.PcItem("vi"),
		readline.PcItem("emacs"),
	),
	readline.PcItem("quit()"),
	readline.PcItem("dump"),
	readline.PcItem("help"),
}

var completer = readline.NewPrefixCompleter(base...)

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

func runInteractive(args []string) int {
	l, err := readline.NewEx(&readline.Config{
		Prompt:          "\033[31m»\033[0m ",
		HistoryFile:     "/tmp/readline.tmp",
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})
	if err != nil {
		panic(err)
	}
	defer l.Close()
	//l.CaptureExitSignal()

	log.SetOutput(l.Stderr())
	v := vm.New(&vm.Config{ImportCore: true, DefineImport: true})
	_ = v.Define("args", args)
	executor := v.Executor()
	rebuildCompleter(executor.GetEnv())
	for {
		line, err := l.Readline()
		if errors.Is(err, readline.ErrInterrupt) {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "mode "):
			switch line[5:] {
			case "vi":
				l.SetVimMode(true)
			case "emacs":
				l.SetVimMode(false)
			default:
				println("invalid mode:", line[5:])
			}
		case line == "mode":
			println("current mode: " + utils.Ternary(l.IsVimMode(), "vim", "emacs"))
		case line == "help":
			usage(l.Stderr())
		case line == "dump":
			println(executor.GetEnv().String())
		case line == "quit()":
			goto exit
		case line == "":
		default:
			source := line
			stmt, err := parser.ParseSrc(source)
			if err != nil {
				var e *parser.Error
				if errors.As(err, &e) {
					es := e.Error()
					if strings.HasPrefix(es, "syntax error: unexpected") {
						if strings.HasPrefix(es, "syntax error: unexpected $end,") {
							continue
						}
					} else {
						if e.Pos.Column == len(source) && !e.Fatal {
							_, _ = fmt.Fprintln(os.Stderr, e)
							continue
						}
						if e.Error() == "unexpected EOF" {
							continue
						}
					}
				}
			}
			var v any
			if err == nil {
				v, err = executor.Run(nil, stmt)
			}
			if err != nil {
				handleErr(os.Stderr, err)
				continue
			}
			if e, ok := v.(envPkg.IEnv); ok {
				fmt.Printf("%s\n", e.Name())
			} else {
				fmt.Printf("%s\n", vmUtils.FormatValue(reflect.ValueOf(v)))
			}
			rebuildCompleter(executor.GetEnv())
		}
	}
exit:
	return OkExitCode
}

func rebuildCompleter(e envPkg.IEnv) {
	newArr := base
	keys := make([]string, 0)
	e.Values().Each(func(s string, value reflect.Value) {
		keys = append(keys, s)
		if value.IsValid() {
			if ee, ok := value.Interface().(envPkg.IEnv); ok {
				ee.Values().Each(func(ss string, _ reflect.Value) {
					keys = append(keys, s+"."+ss)
				})
			}
		}
	})
	slices.Sort(keys)
	for _, k := range keys {
		newArr = append(newArr, readline.PcItem(k))
	}
	completer.SetChildren(newArr)
}

func handleErr(w io.Writer, err error) {
	var vmErr *runner.Error
	var parserErr *parser.Error
	if errors.As(err, &vmErr) {
		_, _ = fmt.Fprintf(w, "%d:%d %s\n", vmErr.Pos.Line, vmErr.Pos.Column, err)
	} else if errors.As(err, &parserErr) {
		_, _ = fmt.Fprintf(w, "%d:%d %s\n", parserErr.Pos.Line, parserErr.Pos.Column, err)
	} else {
		_, _ = fmt.Fprintln(w, err)
	}
}

func compileAndSave(source, fileName string) error {
	fileName = strings.Replace(fileName, ankoFileExt, ankoBytecodeExt, 1)
	out, err := compiler.Compile(source, false)
	if err != nil {
		return err
	}
	if err := os.WriteFile(fileName, out, 0744); err != nil {
		return err
	}
	return nil
}

func runWeb() int {
	v := vm.New(&vm.Config{ImportCore: true, DefineImport: true})

	const scriptTopic = "script"
	const systemTopic = "system"
	ps := pubsub.NewPubSub[string](nil)

	_ = v.Define("log", func(msg string) {
		ps.Pub(scriptTopic, msg)
	})

	// Custom sleep function that will quit faster if the running context is cancelled
	_ = v.DefineCtx("sleep", func(ctx context.Context, duration time.Duration) {
		select {
		case <-time.After(duration):
		case <-ctx.Done():
			panic(ctx.Err())
		}
	})

	e := v.Executor()

	defaultScript := `time = import("time")
for i=0; i<10; i++ {
  sleep(time.Second)
  log("test " + i)
}`

	mux := http.DefaultServeMux
	mux.HandleFunc("/favicon.ico", func(resp http.ResponseWriter, req *http.Request) {})
	mux.HandleFunc("/api/v1/status", func(resp http.ResponseWriter, req *http.Request) {
		resp.WriteHeader(http.StatusOK)
		data := map[string]any{
			"IsRunning": e.IsRunning(),
			"IsPaused":  e.IsPaused(),
		}
		by, _ := json.Marshal(data)
		_, _ = resp.Write(by)
	})
	mux.HandleFunc("/sse", func(resp http.ResponseWriter, req *http.Request) {
		flusher := resp.(http.Flusher)
		resp.Header().Set("Content-Type", "text/event-stream")
		resp.Header().Set("Cache-Control", "no-cache")
		resp.Header().Set("Connection", "keep-alive")
		sub1 := ps.Subscribe(scriptTopic, systemTopic)
		defer sub1.Close()
		sub2 := e.Subscribe()
		defer sub2.Close()
		var msgID int32
		for {
			select {
			case msg := <-sub1.ReceiveCh():
				newMsgID := atomic.AddInt32(&msgID, 1)
				by, _ := json.Marshal(msg)
				_, _ = fmt.Fprintf(resp, "id: %d\r\ndata: %s\r\n\r\n", newMsgID, string(by))
				flusher.Flush()
			case msg := <-sub2.ReceiveCh():
				newMsgID := atomic.AddInt32(&msgID, 1)
				by, _ := json.Marshal(msg)
				_, _ = fmt.Fprintf(resp, "id: %d\r\ndata: %s\r\n\r\n", newMsgID, string(by))
				flusher.Flush()
			case <-req.Context().Done():
				return
			}
		}
	})
	mux.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		resp.WriteHeader(http.StatusOK)
		script := defaultScript
		if req.Method == http.MethodPost {
			script = req.PostFormValue("source")
			submit := req.PostFormValue("submit")
			if submit == "run" {
				if e.RunAsync(context.Background(), script) {
					ps.Pub(systemTopic, "run script")
				}
			} else if submit == "stop" {
				if e.Stop() {
					ps.Pub(systemTopic, "stop script")
				}
			} else if submit == "toggle_pause" {
				res := e.TogglePause()
				if res == executor.PausedToggle {
					ps.Pub(systemTopic, "script paused")
				} else if res == executor.ResumedToggle {
					ps.Pub(systemTopic, "script resumed")
				}
			}
			return
		}
		pageHtml := `<!DOCTYPE html>
<html>
	<head><title>Test</title></head>
	<body>
		<script>
			function toggle_pause() {
				const formData = new FormData();
				formData.append('submit', 'toggle_pause');
				fetch("/", {method: 'POST', body: formData});
			}
			function run() {
				const formData = new FormData();
				formData.append('source', document.getElementById('source').value);
				formData.append('submit', 'run');
				fetch("/", {method: 'POST', body: formData});
			}
			function stop() {
				const formData = new FormData();
				formData.append('submit', 'stop');
				fetch("/", {method: 'POST', body: formData});
			}
			function clearLogs() {
				document.getElementById("logs").innerHTML = '';
			}
		</script>
		<div>
			<button type="button" onclick="run()">run</button>
			<button type="button" onclick="stop()">stop</button>
			<button type="button" onclick="toggle_pause()">pause/resume</button>
			<button type="button" onclick="clearLogs()">clear logs</button>
		</div>
		<div>
			Running: <span id="is_running">{{ if .IsRunning }}running{{ else }}stopped{{ end }}</span> |
			Paused: <span id="is_paused">{{ if .IsPaused }}paused{{ else }}not paused{{ end }}</span>
		</div>
		<textarea name="source" id="source" rows="10" cols="80">{{ .Script }}</textarea>
		<div id="logs"></div>
		<script>
			const evtSource = new EventSource("/sse");
			evtSource.onmessage = (evt) => {
				var newDiv = document.createElement("div");
				const data = JSON.parse(evt.data);
    			newDiv.textContent = data.Topic + ": " + data.Msg;
				if (data.Topic === "executor") {
					switch (data.Msg) {
						case 1: document.getElementById("is_running").innerHTML = "running"; break;
						case 2: document.getElementById("is_running").innerHTML = "stopped"; break;
						case 3: document.getElementById("is_paused").innerHTML = "paused"; break;
						case 4: document.getElementById("is_paused").innerHTML = "not paused"; break;
					}
				} else {
					document.getElementById("logs").appendChild(newDiv);
				}
			};
		</script>
	</body>
</html>`
		data := map[string]any{
			"IsRunning": e.IsRunning(),
			"IsPaused":  e.IsPaused(),
			"Script":    script,
		}
		buf := new(bytes.Buffer)
		tmpl, _ := template.New("").Parse(pageHtml)
		_ = tmpl.Execute(buf, data)
		_, _ = resp.Write(buf.Bytes())
	})

	addr := "127.0.0.1:8080"
	fmt.Printf("listening on %s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Println(err)
	}
	return OkExitCode
}
