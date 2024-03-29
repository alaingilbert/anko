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

	v := vm.New(&vm.Config{
		ImportCore:   utils.Ptr(true),
		DefineImport: utils.Ptr(true),
	})
	_ = v.Define("args", args)
	executorInst := v.Executor(nil)
	fileExt := filepath.Ext(appFlags.File)
	var err error
	if appFlags.FlagExecute != "" || fileExt == ankoFileExt {
		_, err = executorInst.Run(nil, source)
	} else {
		_, err = executorInst.Run(nil, []byte(source))
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
	v := vm.New(&vm.Config{
		ImportCore:   utils.Ptr(true),
		DefineImport: utils.Ptr(true),
		ResetEnv:     utils.Ptr(false),
	})
	_ = v.Define("args", args)
	executorInst := v.Executor(nil)
	rebuildCompleter(executorInst.GetEnv())
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
				v, err = executorInst.Run(nil, stmt)
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
			rebuildCompleter(executorInst.GetEnv())
		}
	}
exit:
	return OkExitCode
}

func rebuildCompleter(env envPkg.IEnv) {
	newArr := base
	keys := make([]string, 0)
	env.Values().Each(func(key string, value reflect.Value) {
		keys = append(keys, key)
		if value.IsValid() {
			if module, ok := value.Interface().(envPkg.IEnv); ok {
				module.Values().Each(func(modKey string, _ reflect.Value) {
					keys = append(keys, key+"."+modKey)
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

func handleErrStr(err error) string {
	buf := new(bytes.Buffer)
	handleErr(buf, err)
	return buf.String()
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
	v := vm.New(&vm.Config{
		ImportCore:   utils.Ptr(true),
		DefineImport: utils.Ptr(true),
		ResetEnv:     utils.Ptr(true),
	})

	const scriptTopic = "script"
	const systemTopic = "system"
	ps := pubsub.NewPubSub[string](nil)

	_ = v.Define("log", func(a ...any) {
		ps.Pub(scriptTopic, fmt.Sprintln(a...))
	})

	_ = v.Define("logf", func(format string, a ...any) {
		ps.Pub(scriptTopic, fmt.Sprintf(format, a...))
	})

	// Custom sleep function that will quit faster if the running context is cancelled
	_ = v.DefineCtx("sleep", func(ctx context.Context, duration time.Duration) {
		select {
		case <-time.After(duration):
		case <-ctx.Done():
			panic(ctx.Err())
		}
	})

	e := v.Executor(nil)

	defaultScript := `time = import("time")
for i=0; i<10; i++ {
    sleep(time.Second)
    log("test " + i)
}`

	selectScript := `time = import("time")
ch1 = make(chan int)
ch2 = make(chan string)
go func() {
    sleep(time.Second)
    ch2 <- "test"
}()
select {
    case v = <-ch1: log("received on ch1: " + v)
    case v = <-ch2: log("received on ch2: " + v)
}`

	typedFuncScript := `// This function is strongly typed for arguments and return values
func typedFn(a int64, b string) (string, int64) {
    log("got " + a + " and " + b)
    return "we can only return a string and int64", 123
}
a, b = typedFn(42, "hello world")
logf("%s | %d", a, b)`

	typedValues := `try {
    a := 123
    a = "123" // type mismatch
} catch err {
    log("got err: ", err)
}

try {
    a := 123
    a := "123" // already defined symbol 'a'
} catch err {
    log("got err: ", err)
}

a := 123
delete("a") // delete a from env
a = "123"   // no error`

	watchdog := `func rec() {
    return rec()
}
rec()`

	scripts := [][]string{
		{"Default", defaultScript},
		{"Select", selectScript},
		{"Typed func", typedFuncScript},
		{"Typed variables", typedValues},
		{"Watchdog", watchdog},
	}

	mux := http.DefaultServeMux
	mux.HandleFunc("/favicon.ico", func(resp http.ResponseWriter, req *http.Request) {})
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
			var by []byte
			select {
			case msg := <-sub1.ReceiveCh():
				by, _ = json.Marshal(msg)
			case msg := <-sub2.ReceiveCh():
				by, _ = json.Marshal(msg)
			case <-req.Context().Done():
				return
			}
			newMsgID := atomic.AddInt32(&msgID, 1)
			_, _ = fmt.Fprintf(resp, "id: %d\r\ndata: %s\r\n\r\n", newMsgID, string(by))
			flusher.Flush()
		}
	})
	mux.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		resp.WriteHeader(http.StatusOK)
		script := scripts[0][1]
		if req.Method == http.MethodPost {
			submit := req.PostFormValue("submit")
			if submit == "run" {
				ctxTimeout := utils.Clamp(utils.DoParseI64(req.PostFormValue("ctx_timeout")), 0, 3600)
				script = req.PostFormValue("source")
				if !e.IsRunning() {
					go func() {
						ctx := context.Background()
						if ctxTimeout > 0 {
							var cancel context.CancelFunc
							ctx, cancel = context.WithTimeout(ctx, time.Duration(ctxTimeout)*time.Second)
							defer cancel()
						}
						if val, err := e.Run(ctx, script); err != nil {
							ps.Pub(scriptTopic, handleErrStr(err))
						} else {
							ps.Pub(scriptTopic, fmt.Sprintf("%#v", val))
						}
					}()
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
			} else if submit == "set_rate_limit" {
				rateLimit := utils.Clamp(utils.DoParseI64(req.PostFormValue("limit")), 0, 100_000_000_000)
				e.SetRateLimit(rateLimit, time.Second)
			}
			return
		}
		pageHtml := `<!DOCTYPE html>
<html>
	<head>
		<title>Test</title>
		<style>
			html, body { background-color: #333; color: #ccc; font-family: Verdana,Helvetica,Arial,sans-serif; }
			textarea, button, select, input { background-color: #444; color: #ccc; padding: 3px 7px; }
			textarea:focus, input:focus { box-shadow: 0px 0px 2px 2px #000; outline: none; }
			.mb-2 { margin-bottom: 10px; }
			.ml-3 { margin-left: 20px; }
			.topic { width: 70px; display: inline-block; }
			input[type=number]::-webkit-inner-spin-button, 
			input[type=number]::-webkit-outer-spin-button { 
				-webkit-appearance: none;
				-moz-appearance: none;
				appearance: none;
				margin: 0; 
			}
		</style>
	</head>
	<body>
		<script>
			const $ = function(id) { return document.getElementById(id); };
			const scripts = {{ .Scripts }};
			function toggle_pause() {
				const formData = new FormData();
				formData.append('submit', 'toggle_pause');
				fetch("/", {method: 'POST', body: formData});
			}
			function run() {
				const formData = new FormData();
				formData.append('source', $('source').value);
				formData.append('ctx_timeout', $("ctx_timeout").value);
				formData.append('submit', 'run');
				fetch("/", {method: 'POST', body: formData});
			}
			function stop() {
				const formData = new FormData();
				formData.append('submit', 'stop');
				fetch("/", {method: 'POST', body: formData});
			}
			function clearLogs() {
				$("logs").innerHTML = '';
			}
			function changeScript() {
				const value = $("script").value;
				$('source').value = scripts[value][1];
			}
			function setRateLimit() {
				const formData = new FormData();
				formData.append('limit', $("rate_limit").value);
				formData.append('submit', 'set_rate_limit');
				fetch("/", {method: 'POST', body: formData});
			}
		</script>
		<div class="mb-2">
			<select id="script" onchange="changeScript()">
				{{ range $i, $v := .Scripts }}<option value="{{ $i }}">{{ index $v 0 }}</option>{{ end }}
			</select>
			<button type="button" onclick="run()">run</button>
			<button type="button" onclick="stop()">stop</button>
			<button type="button" onclick="toggle_pause()">pause/resume</button>
			<button type="button" onclick="clearLogs()">clear logs</button>
			<div class="ml-3" style="display: inline-block;">
				<input type="number" min="0" max="100000000" value="{{ .RateLimit }}" id="rate_limit" style="width: 50px;" />
				<button type="button" onclick="setRateLimit()">Set rate limit</button>
			</div>
		</div>
		<div class="mb-2" style="font-size: 0.8em;">
			<label for="ctx_timeout">Context timeout in seconds:</label>
			<input type="number" min="0" max="3600" value="0" id="ctx_timeout" style="width: 50px;" />
		</div>
		<div class="mb-2">
			Running: <span id="is_running">{{ if .IsRunning }}running{{ else }}stopped{{ end }}</span> |
			Paused: <span id="is_paused">{{ if .IsPaused }}paused{{ else }}not paused{{ end }}</span>
		</div>
		<textarea name="source" id="source" rows="15" cols="80" class="mb-2">{{ .Script }}</textarea>
		<div id="logs"></div>
		<script>
			function sanitize(string) {
				const map = {
					'&': '&amp;',
					'<': '&lt;',
					'>': '&gt;',
					'"': '&quot;',
					"'": '&#x27;',
					"/": '&#x2F;',
				};
				const reg = /[&<>"'/]/ig;
				return string.replace(reg, (match)=>(map[match]));
			}
			const evtSource = new EventSource("/sse");
			evtSource.onmessage = (evt) => {
				const data = JSON.parse(evt.data);
				if (data.Topic === "executor") {
					switch (data.Msg) {
						case 1: $("is_running").innerHTML = "running"; break;
						case 2: $("is_running").innerHTML = "stopped"; break;
						case 3: $("is_paused").innerHTML = "paused"; break;
						case 4: $("is_paused").innerHTML = "not paused"; break;
					}
				} else {
					var newDiv = document.createElement("div");
    				newDiv.innerHTML = '<span class="topic">' + data.Topic + "</span>: " + sanitize(data.Msg);
					$("logs").appendChild(newDiv);
				}
			};
		</script>
	</body>
</html>`
		data := map[string]any{
			"IsRunning": e.IsRunning(),
			"IsPaused":  e.IsPaused(),
			"Script":    script,
			"Scripts":   scripts,
			"RateLimit": utils.First(e.GetRateLimit()),
		}
		buf := new(bytes.Buffer)
		_ = utils.First(template.New("").Parse(pageHtml)).Execute(buf, data)
		_, _ = resp.Write(buf.Bytes())
	})

	addr := "127.0.0.1:8080"
	fmt.Printf("listening on %s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Println(err)
	}
	return OkExitCode
}
