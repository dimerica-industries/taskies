package src

import (
	"fmt"
	"os"
)

func Debug(fn func()) {
	if os.Getenv("DEBUG") != "" {
		fn()
	}
}

func Debugf(format string, args ...interface{}) {
	Debug(func() {
		if format[len(format)-1] != '\n' {
			format += "\n"
		}

		fmt.Printf("[DEBUG] "+format, args...)
	})
}
