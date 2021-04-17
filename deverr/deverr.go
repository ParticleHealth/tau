package deverr

import (
	"fmt"
	"runtime"
)

// Verify that err is nil or panic.
func Verify(err error) {
	if err != nil {
		panic(err)
	}
}

// NotImplementedError indicates something still needs to be built as part of development. Should never be used in production.
type NotImplementedError struct {
	File string
	Line int
	Fn   string
}

// Error string for a NotImplementedError.
func (ni *NotImplementedError) Error() string {
	return fmt.Sprintf("not implemented: %s", ni.Fn)
}

// NotImplemented returns an error detailing that the method being called is not implemented.
// It contains information on the file, line and function.
func NotImplemented() error {
	err := NotImplementedError{File: "unknown", Line: -1, Fn: "unknown"}
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		return &err
	}
	err.File = file
	err.Line = line
	fn := runtime.FuncForPC(pc)
	err.Fn = fn.Name()
	return &err
}
