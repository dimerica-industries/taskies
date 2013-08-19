package taskies

import (
	"fmt"
	"io"
	"os/exec"
	"reflect"
)

func shellProvider(ps providerSet, data interface{}) (Task, error) {
	val := reflect.ValueOf(data)

	if val.Kind() != reflect.String {
		return nil, fmt.Errorf("shell requires string, %s found", val.Kind())
	}

	return ShellTask("sh", []string{"-c", val.String()}), nil
}

func ShellTask(cmd string, args []string) Task {
	return func(env *Env, in io.Reader, out, err io.Writer) error {
		cmd = template(cmd, env)

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
}
