package main

import (
    "flag"
    "fmt"
    "github.com/dimerica-industries/taskies/taskies"
    "io/ioutil"
    "os"
)

const (
    DEFAULT_FILE = "Taskies"
)

type ArrayOpts []string

func (arr *ArrayOpts) String() string {
    return fmt.Sprint(*arr)
}

func (arr *ArrayOpts) Set(str string) error {
    *arr = append(*arr, str)
    return nil
}

func main() {
    defer func() {
        if err := recover(); err != nil {
            fmt.Printf("Error: %s\n", err)
            os.Exit(1)
        }
    }()

    files := &ArrayOpts{}
    flag.Var(files, "f", "")
    help := flag.Bool("h", false, "")

    flag.Parse()

    tasks := flag.Args()

    if len(*files) == 0 {
        *files = ArrayOpts{DEFAULT_FILE}
    }

    if len(tasks) == 0 || *help {
        flag.Usage()
        os.Exit(0)
    }

    ts := taskies.NewTaskSet()

    for _, path := range *files {
        contents, err := ioutil.ReadFile(path)

        if err != nil {
            panic("Cannot read " + path)
        }

        err = taskies.DecodeYAML(contents, ts)

        if err != nil {
            panic("YAML decode error: " + err.Error())
        }
    }

    runner := taskies.NewRunner(ts.Tasks, taskies.FromArray(os.Environ()), os.Stdin, os.Stdout, os.Stderr)
    err := runner.Run(tasks...)

    if err != nil {
        panic(err)
    }
}
