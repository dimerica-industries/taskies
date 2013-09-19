package main

import (
	"flag"
	"fmt"
	taskies "github.com/dimerica-industries/taskies/src"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"text/tabwriter"
)

const (
	DEFAULT_FILE = "./Taskies"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("Error: %s\n", err)
			taskies.Debug(func() {
				debug.PrintStack()
			})
			os.Exit(1)
		}
	}()

	file := flag.String("f", DEFAULT_FILE, "Location of the taskie file")
	help := flag.Bool("h", false, "Show help")
	list := flag.Bool("l", false, "List all available tasks")

	flag.Parse()

	task := flag.Arg(0)
	args := make([]string, 0)

	if flag.NArg() > 1 {
		args = flag.Args()[1:]
	}

	nargs := parseArgs(args)

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	f, err := filepath.Abs(*file)

	if err != nil {
		panic(err)
	}

	rt, err := taskies.LoadRuntime(f, os.Stdin, os.Stdout, os.Stderr)

	if err != nil {
		taskies.Debugf(err)
		panic("Cannot read " + *file)
	}

	l := func() {
		fmt.Printf("Available Tasks:\n")
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 8, 0, '\t', 0)

		ns := rt.RootNs()

		for _, name := range ns.Tasks() {
			t := ns.GetTask(name)

			fmt.Fprintf(w, "   %s\t%s\n", name, t.Description())
		}

		w.Flush()
	}

	if *list {
		l()
		os.Exit(0)
	}

	if task == "" {
		flag.Usage()
		fmt.Println()
		l()
		os.Exit(1)
	}

	rt.Watcher = &watcher{1}

	for k, v := range nargs {
		rt.RootNs().RootEnv().SetVar(k, v)
	}

	err = rt.Run(task)

	if err != nil {
		panic(err)
	}
}

type watcher struct {
	level uint
}

func (w *watcher) BeforeRun(r *taskies.Runtime, e *taskies.Env, t taskies.Task) chan bool {
	if w.level > 0 {
		name := t.Name()

		if name == "" {
			name = t.Type()
		}

		fmt.Fprintf(r.Out(), "\n\033[1m[Running task: %s]\033[0m\n\n", name)
	}

	return nil
}

func (w *watcher) AfterRun(r *taskies.Runtime, e *taskies.Env, t taskies.Task) chan bool {
	return nil
}

func parseArgs(args []string) map[string]string {
	ret := make(map[string]string)

	k := ""

	for _, v := range args {

		if k == "" {
			if v[0] != '-' {
				continue
			}

			for v[0] == '-' {
				v = v[1:]
			}

			k = v
			i := strings.Index(k, "=")

			if i >= 0 {
				ret[k[0:i]] = k[i+1:]
			}

			k = ""
		} else {
			ret[k] = v
			k = ""
		}
	}

	return ret
}
