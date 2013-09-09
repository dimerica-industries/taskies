package src

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"testing"
)

func rt(p string, in io.Reader) (*Runtime, error) {
	out := &bytes.Buffer{}
	err := &bytes.Buffer{}

	if p == "" {
		return NewRuntime(in, out, err), nil
	}

	return LoadRuntime(p, in, out, err)
}

func (t *funcTask) Export() []map[string]interface{} {
	return t.export
}

func TestRuntimeCreate(t *testing.T) {
	r, _ := rt("", nil)

	if r.ns == nil {
		t.Fatalf("Expect root ns to be populated on runtime create")
	}
}

func TestFuncRun(t *testing.T) {
	r, _ := rt("", nil)

	tk := &funcTask{
		baseTask: &baseTask{
			name: "test",
		},
		fn: func(r RunContext) error {
			r.Out().Write([]byte("HELLO"))
			return nil
		},
	}

	r.ns.RootEnv().AddTask(tk)
	err := r.Run("test")

	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(r.Out().(*bytes.Buffer).Bytes(), []byte("HELLO")) {
		t.Fatalf("Expected to find HELLO")
	}
}

func TestLoad(t *testing.T) {
	d, err := ioutil.TempDir("", "taskies_test")

	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(d)

	n := d + "/test"

	ioutil.WriteFile(n, []byte(`
- task:
    name: Test
    shell: echo HELLO
`), 0700)

	r, err := rt(n, nil)

	if err != nil {
		t.Fatal(err)
	}

	if task, _ := r.ns.RootEnv().GetTask("Test"); task == nil {
		t.Fatal("Task test not loaded")
	}
}

func newTmpdir() (*tmpdir, error) {
	d, err := ioutil.TempDir("", "taskies_test")

	if err != nil {
		return nil, err
	}

	return &tmpdir{d, 0}, nil
}

type tmpdir struct {
	dir string
	i   int
}

func (t *tmpdir) cleanup() error {
	return os.Remove(t.dir)
}

func (t *tmpdir) addFile(contents []byte) (string, error) {
	n := strconv.Itoa(t.i)
	t.i++

	f, err := os.Create(t.dir + "/" + n)

	if err != nil {
		return "", err
	}

	_, err = f.Write(contents)

	if err != nil {
		return "", err
	}

	return f.Name(), f.Close()
}

func testEquals(t *testing.T, yaml [][]byte, val string, in io.Reader, tasks ...string) {
	d, err := newTmpdir()

	if err != nil {
		t.Fatal(err)
	}

	defer d.cleanup()

	var root string

	for _, y := range yaml {
		n, err := d.addFile(y)

		if err != nil {
			t.Fatal(err)
		}

		root = n
	}

	r, err := rt(root, in)

	if err != nil {
		t.Fatal(err)
	}

	if len(tasks) == 0 {
		tasks = r.ns.RootEnv().Tasks()
	}

	for _, task := range tasks {
		if err = r.Run(task); err != nil {
			t.Fatal(err)
		}
	}

	out := strings.TrimSpace(string(r.Out().(*bytes.Buffer).Bytes()))

	if out != val {
		t.Fatalf("Expecting %s, found %s", val, out)
	}
}

func TestShell(t *testing.T) {
	testEquals(t, [][]byte{[]byte(`
- task:
    name: Test shell
    shell: echo 3
`)}, "3", nil)
}

func TestPipe(t *testing.T) {
	testEquals(t, [][]byte{[]byte(`
- task:
      name: Test pipe
      pipe:
          - shell: echo 3
          - shell: while read line; do echo "HELLO_$line"; done
`)}, "HELLO_3", nil)
}

func TestMultiple(t *testing.T) {
	testEquals(t, [][]byte{[]byte(`
- task:
    name: Test multiple
    run:
      - shell: bash -c "echo -n 10"
      - shell: bash -c "echo -n 3"
`)}, "103", nil)
}

func TestCustom(t *testing.T) {
	testEquals(t, [][]byte{[]byte(`
- task:
    name: Test
    shell: bash -c "echo {{val}}"

- task:
    name: Test custom
    Test: 
      val: 100

`)}, "100", nil, "Test custom")
}

func TestCustomSetEnv(t *testing.T) {
	testEquals(t, [][]byte{[]byte(`
- task:
    name: test
    shell: bash -c "echo -n 20"
    set:
      a: 50

- task:
    name: test2
    run: test
    set:
      b: "{{a}}"

- task:
    name: Test3
    run: 
      - test2
      - shell: echo {{LAST.a}}

`)}, "2050", nil, "Test3")

	testEquals(t, [][]byte{[]byte(`
- task:
    name: test
    shell: bash -c "echo -n 20"
    set:
      x: 50

- task:
    name: Test2
    run:
      - test
      - test
      - shell: echo {{LAST.x}}
      - shell: echo {{TASKS.test_1.x}}
`)}, "202050\n50", nil, "Test2")

	testEquals(t, [][]byte{[]byte(`
- task:
    name: test
    shell: bash -c "echo -n 20"

- task:
    name: test2
    run: test
    set:
      x: 50

- task:
    name: Test2
    run:
      - test2
      - test2
      - shell: echo {{LAST.x}}
      - shell: echo {{TASKS.test2_1.x}}
`)}, "202050\n50", nil, "Test2")
}

func TestDecodeEnv(t *testing.T) {
	testEquals(t, [][]byte{[]byte(`
- set:
    key: value
- task:
    name: Test
    shell: echo {{key}}
`)}, "value", nil)
}

func TestTemplate(t *testing.T) {
	testEquals(t, [][]byte{[]byte(`
- set:
    val: 10
    val2: wtf_{{val}}
- task:
      name: Test template
      shell: echo {{val2}}
`)}, "wtf_10", nil)
}

func TestResultSet(t *testing.T) {
	testEquals(t, [][]byte{[]byte(`
- task:
    name: Test1
    shell: bash -c "echo -n 10"

- task:
    name: Test2
    shell: bash -c "echo -n bleh{{LAST.OUT}}"
`)}, "10bleh10", nil)
}

func TestCustomSet(t *testing.T) {
	testEquals(t, [][]byte{[]byte(`
- task:
    name: test1
    shell: echo 10 > /dev/null
    set:
      OMG: WE DID IT
      complex:
        OMG: WE DID IT

- task:
    name: Test2
    run:
      - test1
      - shell: bash -c "echo -n {{LAST.OMG}}"
`)}, "WE DID IT", nil, "Test2")
}

func TestCustomComplexSet(t *testing.T) {
	testEquals(t, [][]byte{[]byte(`
- task:
      name: test1
      shell: echo 10 > /dev/null
      set:
        OMG: WE DID IT
        complex:
            OMG: WE DID IT

- task:
      name: Test2
      run:
        - test1
        - shell: bash -c "echo -n {{LAST.complex.OMG}}"
`)}, "WE DID IT", nil, "Test2")
}

func TestCustomScope(t *testing.T) {
	testEquals(t, [][]byte{[]byte(`
- task:
    name: test1
    shell: echo {{hello}}

- task:
    name: Test2
    test1:
      hello: 10

- task:
    name: Test3
    test1: ds
`)}, "10", nil, "Test2", "Test3")
}

func TestAlternateSyntax(t *testing.T) {
	testEquals(t, [][]byte{[]byte(`
- task:
      name: test1
      shell: echo 10

- task:
      name: Test2
      task: test1
`)}, "10", nil, "Test2")
}

func TestComplexInput(t *testing.T) {
	testEquals(t, [][]byte{[]byte(`
- task:
    name: test
    shell: echo "{{#var1}}{{.}}{{/var1}}{{#var2}}{{.}}{{/var2}}{{#var2}}{{var3}}{{/var2}}"

- task:
    name: Test2
    test:
      var1:
        - one
        - two
      var2: "hello"
      var3: "ok"
`)}, "onetwohellook", nil, "Test2")
}

func TestFlow(t *testing.T) {
	testEquals(t, [][]byte{[]byte(`
- task:
    name: test
    shell: echo {{var}}

- task:
    name: Test2
    test: { var: "hello" }
`)}, "hello", nil, "Test2")
}

func TestInitialRun(t *testing.T) {
	testEquals(t, [][]byte{[]byte(`
- shell: echo 3
`)}, "3", nil)

	testEquals(t, [][]byte{[]byte(`
- run: 
    - shell: echo 3
    - shell: echo 5
`)}, "3\n5", nil)

	testEquals(t, [][]byte{[]byte(`
- pipe: 
    - shell: echo 3
    - shell: while read line; do echo HELLO_$line; done
`)}, "HELLO_3", nil)
}

func TestInclude(t *testing.T) {
	testEquals(t, [][]byte{
		[]byte(`
- task:
    name: Hello
    shell: echo 3
`),
		[]byte(`
- include: { other: ./0 }
- other.Hello

`),
	}, "3", nil)
}

func TestIncludeScope(t *testing.T) {
	testEquals(t, [][]byte{
		[]byte(`
- set:
    a: 100
    b: 200
        
- task:
    name: Hello
    run:
        - shell: echo {{a}}
        - shell: echo {{b}}
`),
		[]byte(`
- include: { other0: ./0 }

- set:
    b: 100
        
- task:
    name: Hello
    other0.Hello:
        b: "{{b}}"
`),
		[]byte(`
- include: { other1: ./1 }

- task:
    name: ok
    run:
        - other1.Hello
        - shell: echo "{{TASKS.other1.Hello.OUT}}"
`),
	}, "100\n100\n100\n100", nil)
}

func TestTaskVar(t *testing.T) {
	testEquals(t, [][]byte{[]byte(`
- run:
    var: x
    shell: echo 3
    
- shell: echo {{x.OUT}}
`)}, "3\n3", nil)

	testEquals(t, [][]byte{[]byte(`
- task:
    name: Hello
    run:
      - shell: echo 3
        var: x

      - shell: echo {{x.OUT}}
`)}, "3\n3", nil)
}

func TestTaskVar2(t *testing.T) {
	testEquals(t, [][]byte{[]byte(`
- task:
    name: one
    run:
      - var: a
        shell: echo 3

      - var: b
        shell: echo 10

    set:
      c: "{{b.OUT}}"

- task:
    name: two
    run:
        - one
        - shell: echo "{{LAST.c}}"
`)}, "3\n10\n10", nil, "two")
}
