package main

import (
	"flag"
	"fmt"
	taskies "github.com/dimerica-industries/taskies/src"
	"io/ioutil"
	"os"
	"runtime/debug"
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

	tasks := flag.Args()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	ts := taskies.NewTaskSet()

    contents, err := ioutil.ReadFile(*file)

    if err != nil {
        panic("Cannot read " + *file)
    }

    err = taskies.DecodeYAML(contents, ts)

    if err != nil {
        panic("YAML decode error: " + err.Error())
    }

    l := func() {
		fmt.Printf("Available Tasks:\n")
        w := new(tabwriter.Writer)
        w.Init(os.Stdout, 0, 8, 0, '\t', 0)

		for name, t := range ts.ExportedTasks {
			fmt.Fprintf(w, "   %s\t%s\n", name, t.Description())
		}

        w.Flush()
    }

	if *list {
        l()
		os.Exit(0)
	}

	if len(tasks) == 0 {
		flag.Usage()
        fmt.Println()
        l()
		os.Exit(1)
	}

	runner := taskies.NewRunner(ts, ts.Env, os.Stdin, os.Stdout, os.Stderr)
	err = runner.Run(tasks...)

	if err != nil {
		panic(err)
	}
}
