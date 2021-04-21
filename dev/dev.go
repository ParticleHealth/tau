// Package dev provides convenience methods for development, such as failing fast on errors and a well defined not implemented error.
package dev

import (
	"fmt"
	"runtime"
)

// FailFast panics if err is not nil.
func FailFast(err error) {
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
	if fn != nil {
		err.Fn = fn.Name()
	}
	return &err
}
