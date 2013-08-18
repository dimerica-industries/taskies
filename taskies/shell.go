package taskies

import (
	"io"
	"os/exec"
	"reflect"
)

func ShellProvider(ps ProviderSet, data interface{}) (Task, error) {
	val := reflect.ValueOf(data)
	t := &ShellTask{}

	switch val.Kind() {
	case reflect.Map:
	case reflect.String:
		t.cmd = "sh"
		t.args = []string{"-c", val.String()}
	}

	return t.Run, nil
}

type ShellTask struct {
	cmd  string
	args []string
}

func (s *ShellTask) Run(env *Env, in io.Reader, out, err io.Writer) error {
	cmd := exec.Command(s.cmd, s.args...)

	cmd.Env = env.Array()
	cmd.Stdin = in
	cmd.Stdout = out
	cmd.Stderr = err

	return cmd.Run()
}
