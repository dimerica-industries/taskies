package taskies

import (
	"fmt"
	"io"
	"os/exec"
	"reflect"
)

type ShellTask struct {
	name        string
	description string
	cmd         string
	args        []string
}

func (t *ShellTask) Name() string {
	return t.name
}

func (t *ShellTask) Description() string {
	return t.description
}

func (t *ShellTask) Run(env *Env, in io.Reader, out, err io.Writer) error {
	cmd := template(t.cmd, env)
	args := t.args

	for i, a := range args {
		args[i] = template(a, env)
	}

	Debugf("[SHELL] %s %s", cmd, args)
	c := exec.Command(cmd, args...)

	c.Stdin = in
	c.Stdout = out
	c.Stderr = err

	return c.Run()
}

func shellProvider(ps providerSet, data *taskData) (Task, error) {
	val := reflect.ValueOf(data.data)

	if val.Kind() != reflect.String {
		return nil, fmt.Errorf("shell requires string, %s found", val.Kind())
	}

	return &ShellTask{
		name:        data.name,
		description: data.description,
		cmd:         "sh",
		args:        []string{"-c", val.String()},
	}, nil
}
