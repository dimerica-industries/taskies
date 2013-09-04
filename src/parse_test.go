package src

import (
	"testing"
)

func TestParseRun(t *testing.T) {
	yaml := []byte(`
- run:
    task: shell
    args: hello
    var: hello
`)
	ast, err := parseBytes(yaml)

	if err != nil {
		t.Fatal(err)
	}

	if len(ast.instructions) != 1 {
		t.Fatal("Expects one instruction")
	}

	tasks := ast.instructions[0].(*runTasks)

	if len(tasks.tasks) != 1 {
		t.Fatalf("Expect one task, found %d", len(tasks.tasks))
	}

	task := tasks.tasks[0]

	if task == nil {
		t.Fatal("Expect instruction to be of type runTask")
	}

	if task.task != "shell" {
		t.Fatal("Expect task to be 'shell'")
	}

	if task.varName != "hello" {
		t.Fatal("Expect var to be 'hello'")
	}

	if task.args.String() != "hello" {
		t.Fatal("Expect args to be a string with value 'hello'")
	}
}

func TestParseMany(t *testing.T) {
	yaml := []byte(`
- run:
    - task: shell
      args: hello
      var: hello
    - task: shell
      args: hello
      var: hello
`)

	ast, err := parseBytes(yaml)

	if err != nil {
		t.Fatal(err)
	}

	if len(ast.instructions) != 1 {
		t.Fatal("Expects one instruction")
	}

	tasks := ast.instructions[0].(*runTasks)

	if len(tasks.tasks) != 2 {
		t.Fatalf("Expect two tasks, found %d", len(tasks.tasks))
	}
}

func TestParseAlternative(t *testing.T) {
	yaml := []byte(`
- run:
  - shell: hello
`)
	ast, err := parseBytes(yaml)

	if err != nil {
		t.Fatal(err)
	}

	if len(ast.instructions) != 1 {
		t.Fatal("Expects one instruction")
	}

	tasks := ast.instructions[0].(*runTasks)

	if len(tasks.tasks) != 1 {
		t.Fatalf("Expect one task, found %d", len(tasks.tasks))
	}

	task := tasks.tasks[0]

	if task == nil {
		t.Fatal("Expect instruction to be of type runTask")
	}

	if task.task != "shell" {
		t.Fatalf("Expect task to be 'shell', found %s", task.task)
	}

	if task.args.String() != "hello" {
		t.Fatal("Expect args to be a string with value 'hello'")
	}
}

func TestParseTask(t *testing.T) {
	yaml := []byte(`
- task:
    name: hello
    description: wtf is this
`)
	ast, err := parseBytes(yaml)

	if err != nil {
		t.Fatal(err)
	}

	if len(ast.instructions) != 1 {
		t.Fatal("Expects one instruction")
	}

	task := ast.instructions[0].(*defineTask)

	if task == nil {
		t.Fatal("Expect instruction to be of type defineTask")
	}

	if task.name != "hello" {
		t.Fatal("expect task name to be 'hello'")
	}

	if task.description != "wtf is this" {
		t.Fatal("expect task description to be 'wtf is this'")
	}
}

func TestParseTask2(t *testing.T) {
	yaml := []byte(`
- task:
    name: hello
    description: wtf is this
    shell: ls -l
`)
	ast, err := parseBytes(yaml)

	if err != nil {
		t.Fatal(err)
	}

	if len(ast.instructions) != 1 {
		t.Fatal("Expects one instruction")
	}

	task := ast.instructions[0].(*defineTask)

	if task == nil {
		t.Fatal("Expect instruction to be of type defineTask")
	}

	if task.name != "hello" {
		t.Fatal("expect task name to be 'hello'")
	}

	if task.description != "wtf is this" {
		t.Fatal("expect task description to be 'wtf is this'")
	}

	if len(task.runList.tasks) != 1 {
		t.Fatal("expected runlist to be of length 1")
	}

	for _, task := range task.runList.tasks {
		if task.task != "shell" {
			t.Fatal("expect task name to be shell")
		}

		if task.args.String() != "ls -l" {
			t.Fatal("expect task arts to be 'ls -l'")
		}
	}
}

func TestParseSet(t *testing.T) {
	yaml := []byte(`
- set:
    k: v
`)

	ast, err := parseBytes(yaml)

	if err != nil {
		t.Fatal(err)
	}

	if len(ast.instructions) != 1 {
		t.Fatal("Expects one instruction")
	}

	ins := ast.instructions[0].(*setVar)

	if ins == nil {
		t.Fatal("Expect instruction to be of type setVar")
	}

	if len(ins.vars) != 1 {
		t.Fatal("Expect setvar to be of len 1")
	}

	for k, v := range ins.vars {
		if k != "k" {
			t.Fatal("Expect key to be 'k'")
		}

		if v.(string) != "v" {
			t.Fatal("Expect value to be string of value 'v'")
		}
	}
}
