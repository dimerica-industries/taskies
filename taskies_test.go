package main

import (
	"bytes"
	"fmt"
	taskies "github.com/dimerica-industries/taskies/src"
	"io"
	"strings"
	"testing"
)

var x = fmt.Println

func run(contents []byte, in io.Reader, tasks ...string) ([]byte, []byte, error) {
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

func testEquals(t *testing.T, yaml []byte, str string, in io.Reader, tasks ...string) {
	out, _, e := run(yaml, in, tasks...)

	if e != nil {
		t.Fatal(e)
	}

	strout := strings.TrimSpace(string(out))

	if strout != str {
		t.Errorf("Expecting \"%s\" found \"%s\"", str, strout)
	}
}

func TestShell(t *testing.T) {
	testEquals(t, []byte(`
tasks:
    - name: test shell
      shell: echo 3
`), "3", nil)
}

func TestPipe(t *testing.T) {
	testEquals(t, []byte(`
tasks:
    - name: test pipe
      pipe:
          - shell: echo 3
          - shell: cat
`), "3", nil)
}

func TestMultiple(t *testing.T) {
	testEquals(t, []byte(`
tasks:
    - name: test multiple
      tasks:
        - shell: bash -c "echo -n 10"
        - shell: bash -c "echo -n 3"
`), "103", nil)
}

func TestCustom(t *testing.T) {
	testEquals(t, []byte(`
tasks:
    - name: test
      shell: bash -c "echo {{val}}"
    - name: test custom
      test: 
        val: 100
`), "100", nil, "test custom")
}

func TestCustomSetEnv(t *testing.T) {
	testEquals(t, []byte(`
tasks:
  - name: test
    shell: bash -c "echo -n 20"
    export: 
      x: 50

  - name: test2
    tasks:
     - name: test
       task: test
     - shell: echo {{test.x}}
`), "2050", nil, "test2")
}

func TestDecodeEnv(t *testing.T) {
	testEquals(t, []byte(`
env:
  key: value 
tasks:
  - name: test
    shell: echo {{key}}
`), "value", nil)
}

func TestTemplate(t *testing.T) {
	testEquals(t, []byte(`
env:
    val: 10
    val2: wtf_{{val}}
tasks:
    - name: test template
      shell: echo {{val2}}
`), "wtf_10", nil)
}

func TestResultSet(t *testing.T) {
	testEquals(t, []byte(`
tasks:
    - name: test1
      shell: bash -c "echo -n 10"

    - name: test2
      shell: bash -c "echo -n bleh{{$last.$stdout}}"
`), "10bleh10", nil)
}

func TestCustomSet(t *testing.T) {
	testEquals(t, []byte(`
tasks:
    - name: test1
      shell: echo 10 > /dev/null
      export:
        OMG: WE DID IT
        complex:
            OMG: WE DID IT

    - name: test2
      tasks:
        - test1
        - shell: bash -c "echo -n {{$last.OMG}}"
`), "WE DID IT", nil, "test2")
}

func TestCustomComplexSet(t *testing.T) {
	testEquals(t, []byte(`
tasks:
    - name: test1
      shell: echo 10 > /dev/null
      export:
        OMG: WE DID IT
        complex:
            OMG: WE DID IT

    - name: test2
      tasks:
        - test1
        - shell: bash -c "echo -n {{$last.complex.OMG}}"
`), "WE DID IT", nil, "test2")
}

func TestCustomScope(t *testing.T) {
	testEquals(t, []byte(`
tasks:
    - name: test1
      shell: echo {{hello}}

    - name: test2
      test1:
        hello: 10

    - name: test3
      test1: ds
`), "10", nil, "test2", "test3")
}

func TestAlternateSyntax(t *testing.T) {
	testEquals(t, []byte(`
tasks:
    - name: test1
      shell: echo 10

    - name: test2
      task: test1
`), "10", nil, "test2")
}
