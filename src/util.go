package src

import (
	"fmt"
	"os"
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
			format = "%#v\n"
		}

		fmt.Printf("[DEBUG] "+format, args...)
	})
}
