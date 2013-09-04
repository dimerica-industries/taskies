package main

import (
	"flag"
	"fmt"
	taskies "github.com/dimerica-industries/taskies/src"
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

	task := flag.Arg(0)

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	rt, err := taskies.LoadRuntime(*file, os.Stdin, os.Stdout, os.Stderr)

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

	err = rt.Run(task)

	if err != nil {
		panic(err)
	}
}
