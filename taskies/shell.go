package taskies

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

func (t *ShellTask) Run(ctxt *RunContext) error {
	cmd := template(t.cmd, ctxt.Env).(string)
	args := t.args

	for i, a := range args {
		args[i] = template(a, ctxt.Env).(string)
	}

	Debugf("[SHELL] %s %s", cmd, args)
	c := exec.Command(cmd, args...)

	c.Stdin = ctxt.In
	c.Stdout = ctxt.Out
	c.Stderr = ctxt.Err

	return c.Run()
}

func shellProvider(ps providerSet, data *taskData) (Task, error) {
	val := reflect.ValueOf(data.data)

	if val.Kind() != reflect.String {
		return nil, fmt.Errorf("shell requires string, %s found", val.Kind())
	}

	return &ShellTask{
		baseTask: baseTaskFromTaskData(data),
		cmd:      "sh",
		args:     []string{"-c", val.String()},
	}, nil
}
