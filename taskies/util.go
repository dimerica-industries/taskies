package taskies

import (
    "fmt"
    "os"
)

func Debugf(format string, args ...interface{}) {
    if os.Getenv("DEBUG") != "" {
        fmt.Printf("[DEBUG] " + format, args...)
    }
}
