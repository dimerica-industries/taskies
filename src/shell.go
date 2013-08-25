package src

import (
	"fmt"
	"os/exec"
	"reflect"
)

type ShellTask struct {
	*baseTask
	cmd  string
	args []string
}

func (t *ShellTask) Run(ctxt RunContext) error {
	cmd := template(t.cmd, ctxt.Env()).(string)
	args := make([]string, len(t.args))

	for i, _ := range args {
		args[i] = template(t.args[i], ctxt.Env()).(string)
	}

	Debugf("[SHELL] %s %s", cmd, args)
	c := exec.Command(cmd, args...)

	c.Stdin = ctxt.In()
	c.Stdout = ctxt.Out()
	c.Stderr = ctxt.Err()

	return c.Run()
}

func shellDecoder(data *taskData) (Task, error) {
	val := reflect.ValueOf(data.taskData)

	if val.Kind() != reflect.String {
		return nil, fmt.Errorf("shell requires string, %s found", val.Kind())
	}

	return &ShellTask{
		baseTask: baseTaskFromTaskData(data),
		cmd:      "sh",
		args:     []string{"-c", val.String()},
	}, nil
}
