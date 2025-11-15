package failure

import (
	"fmt"
	"log"
	"runtime"
)

// Will log an error message prefixed with "Error" and the file and line number info of the caller of this function.
func LogErrWithLocation(message string, arguments ...any) {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "unknown location"
	}
	log.Printf("Error at %v:%v - %v\n", file, line, fmt.Sprintf(message, arguments...))
}

// Will log a warning message prefixed with "Warning" and the file and line number info of the caller of this function.
func LogWarningWithLocation(message string, arguments ...any) {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "unknown location"
	}
	log.Printf("Warning at %v:%v - %v\n", file, line, fmt.Sprintf(message, arguments...))
}
