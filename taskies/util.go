package taskies

import (
	"fmt"
	"os"
)

func Debugf(format string, args ...interface{}) {
	if os.Getenv("DEBUG") != "" {
        if format[len(format) - 1] != '\n' {
            format += "\n"
        }

		fmt.Printf("[DEBUG] " + format, args...)
	}
}
