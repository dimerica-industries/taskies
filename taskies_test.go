package main

import (
    "bytes"
    "fmt"
    "github.com/dimerica-industries/taskies/taskies"
    "io"
    "strings"
    "testing"
)

var x = fmt.Println

func test(contents []byte, in io.Reader) ([]byte, []byte, error) {
    if in != nil {
        in = bytes.NewBuffer([]byte{})
    }

    out := bytes.NewBuffer([]byte{})
    err := bytes.NewBuffer([]byte{})
    ts := taskies.NewTaskSet()
    e := taskies.DecodeYAML(contents, ts)
    run := taskies.NewRunner(ts.Tasks, make(taskies.Env), in, out, err)

    if e != nil {
        return nil, nil, e
    }

    e = run.Run()

    if e != nil {
        return nil, nil, e
    }

    return out.Bytes(), err.Bytes(), nil
}

func TestShell(t *testing.T) {
    yaml := []byte(`
tasks:
    - name: test
      shell: echo 3
`)

    out, _, e := test(yaml, nil)

    if e != nil {
        t.Error(e)
    }

    if strings.TrimSpace(string(out)) != "3" {
        t.Errorf("Expecting \"3\" found \"%s\"", string(out))
    }
}

func TestPipe(t *testing.T) {
    yaml := []byte(`
tasks:
    - name: test
      pipe:
          - shell: echo 3
          - shell: cat
`)

    out, _, e := test(yaml, nil)

    if e != nil {
        t.Error(e)
    }

    if strings.TrimSpace(string(out)) != "3" {
        t.Errorf("Expecting \"3\" found \"%s\"", string(out))
    }
}

func TestMultiple(t *testing.T) {
    yaml := []byte(`
tasks:
    - name: test
      tasks:
        - shell: bash -c "echo -n 10"
        - shell: bash -c "echo -n 3"
`)
    out, _, e := test(yaml, nil)

    if e != nil {
        t.Error(e)
    }

    str := strings.TrimSpace(string(out))

    if strings.TrimSpace(str) != "103" {
        t.Errorf("Expecting \"103\" found \"%s\"", str)
    }
}
