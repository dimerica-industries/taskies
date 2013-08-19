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

func test(contents []byte, in io.Reader, tasks ...string) ([]byte, []byte, error) {
	if in != nil {
		in = bytes.NewBuffer([]byte{})
	}

	out := bytes.NewBuffer([]byte{})
	err := bytes.NewBuffer([]byte{})
	ts := taskies.NewTaskSet()
	e := taskies.DecodeYAML(contents, ts)

	if e != nil {
		return nil, nil, e
	}

	run := taskies.NewRunner(ts.Tasks, ts.Env, in, out, err)

	if len(tasks) == 0 {
		e = run.RunAll()
	} else {
		e = run.Run(tasks...)
	}

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
		t.Fatal(e)
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
		t.Fatal(e)
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
		t.Fatal(e)
	}

	str := strings.TrimSpace(string(out))

	if strings.TrimSpace(str) != "103" {
		t.Errorf("Expecting \"103\" found \"%s\"", str)
	}
}

func TestCustom(t *testing.T) {
	yaml := []byte(`
tasks:
    - name: test
      shell: bash -c "echo {{val}}"
    - name: test2
      test: 
        val: 100
`)

	out, _, e := test(yaml, nil, "test2")

	if e != nil {
		t.Fatal(e)
	}

	str := strings.TrimSpace(string(out))

	if strings.TrimSpace(str) != "100" {
		t.Errorf("Expecting \"100\" found \"%s\"", str)
	}
}

func TestDecodeEnv(t *testing.T) {
	yaml := []byte(`
env:
  key: value 
`)
	ts := taskies.NewTaskSet()
	e := taskies.DecodeYAML(yaml, ts)

	if e != nil {
		t.Fatal(e)
	}

	v := ts.Env.Get("key")

	if v != "value" {
		t.Fatalf("Expecting \"value\", found \"%s\"", v)
	}
}

func TestTemplate(t *testing.T) {
	yaml := []byte(`
env:
    val: 10
    val2: wtf_{{val}}
tasks:
    - name: test
      shell: echo {{val2}}
`)

	out, _, e := test(yaml, nil)

	if e != nil {
		t.Fatal(e.Error())
	}

	str := strings.TrimSpace(string(out))

	if strings.TrimSpace(str) != "wtf_10" {
		t.Errorf("Expecting \"wtf_10\" found \"%s\"", str)
	}
}
