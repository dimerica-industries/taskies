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

	tasks := flag.Args()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	rt := taskies.NewRuntime(os.Stdin, os.Stdout, os.Stderr)
	_, err := rt.LoadNs(*file)

	if err != nil {
		panic("Cannot read " + *file)
	}

	l := func() {
		fmt.Printf("Available Tasks:\n")
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 8, 0, '\t', 0)

		for _, name := range rt.RootNs().ExportedTasks() {
			t := rt.RootNs().ExportedTask(name)

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

	err = rt.Run(tasks...)

	if err != nil {
		panic(err)
	}
}
