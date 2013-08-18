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
		cmd := exec.Command(cmd, args...)

		cmd.Env = env.Array()
		cmd.Stdin = in
		cmd.Stdout = out
		cmd.Stderr = err

		return cmd.Run()
	}
}
