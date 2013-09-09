package src

import (
	"fmt"
	"os"
	"strings"
)

// Run the specified function only if the DEBUG environment
// var is set
func Debug(fn func()) {
	if os.Getenv("DEBUG") != "" {
		fn()
	}
}

// Run fmt.Printf on the passed in arguments only if the DEBUG
// environment is set
func Debugf(args ...interface{}) {
	Debug(func() {
		var format string

		ftmp := args[0]

		if f, ok := ftmp.(string); ok {
			if f[len(f)-1] != '\n' {
				f += "\n"
			}

			format = f
			args = args[1:]
		} else {
			format = strings.Repeat("[%#v] ", len(args)) + "\n"
		}

		fmt.Printf("[DEBUG] "+format, args...)
	})
}

func inDir(path string, fn func() error) error {
	pwd, err := os.Getwd()

	if err != nil {
		return err
	}

	Debugf("[CHDIR] %s", path)
	err = os.Chdir(path)

	if err != nil {
		return err
	}

	if err = fn(); err != nil {
		return err
	}

	Debugf("[CHDIR] %s", pwd)
	return os.Chdir(pwd)
}
