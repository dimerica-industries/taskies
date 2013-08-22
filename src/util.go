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
func Debugf(format string, args ...interface{}) {
	Debug(func() {
		if format[len(format)-1] != '\n' {
			format += "\n"
		}

		fmt.Printf("[DEBUG] "+format, args...)
	})
}
