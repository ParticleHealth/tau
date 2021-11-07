package config

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

// updateUsage to reflect ability to set via environment variable.
func updateUsage(name string, value interface{}, usage string) string {
	return fmt.Sprintf("%s\nAlso set by environment variable %s=%T", usage, strings.ToUpper(name), value)
}

// override a flag value based on an environment variable being set.
func override(name string) {
	env := strings.ToUpper(name)
	if v, ok := os.LookupEnv(env); ok {
		err := flag.Set(name, v)
		if err != nil {
			panic(err)
		}
	}
}

// Bool defines a bool flag with specified name, default value, and usage string.
// The return value is the address of a bool variable that stores the value of the flag.
func Bool(name string, value bool, usage string) *bool {
	v := flag.Bool(name, value, updateUsage(name, value, usage))
	override(name)

	return v
}

// BoolVar defines a bool flag with specified name, default value, and usage string.
// The argument p points to a bool variable in which to store the value of the flag.
func BoolVar(p *bool, name string, value bool, usage string) {
	flag.BoolVar(p, name, value, updateUsage(name, value, usage))
	override(name)
}

// Duration defines a time.Duration flag with specified name, default value, and usage string.
// The return value is the address of a time.Duration variable that stores the value of the flag.
// The flag accepts a value acceptable to time.ParseDuration.
func Duration(name string, value time.Duration, usage string) *time.Duration {
	v := flag.Duration(name, value, updateUsage(name, value, usage))
	override(name)

	return v
}

// DurationVar defines a time.Duration flag with specified name, default value, and usage string.
// The argument p points to a time.Duration variable in which to store the value of the flag.
// The flag accepts a value acceptable to time.ParseDuration.
func DurationVar(p *time.Duration, name string, value time.Duration, usage string) {
	flag.DurationVar(p, name, value, updateUsage(name, value, usage))
	override(name)
}

// Float64 defines a float64 flag with specified name, default value, and usage string.
// The return value is the address of a float64 variable that stores the value of the flag.
func Float64(name string, value float64, usage string) *float64 {
	v := flag.Float64(name, value, updateUsage(name, value, usage))
	override(name)

	return v
}

// Float64Var defines a float64 flag with specified name, default value, and usage string.
// The argument p points to a float64 variable in which to store the value of the flag.
func Float64Var(p *float64, name string, value float64, usage string) {
	flag.Float64Var(p, name, value, updateUsage(name, value, usage))
	override(name)
}

// Func defines a flag with the specified name and usage string.
// Each time the flag is seen, fn is called with the value of the flag.
// If fn returns a non-nil error, it will be treated as a flag value parsing error.
func Func(name, usage string, fn func(string) error) {
	flag.Func(name, updateUsage(name, "", usage), fn)
	override(name)
}

// Int defines an int flag with specified name, default value, and usage string.
// The return value is the address of an int variable that stores the value of the flag.
func Int(name string, value int, usage string) *int {
	v := flag.Int(name, value, updateUsage(name, value, usage))
	override(name)

	return v
}

// Int64 defines an int64 flag with specified name, default value, and usage string.
// The return value is the address of an int64 variable that stores the value of the flag.
func Int64(name string, value int64, usage string) *int64 {
	v := flag.Int64(name, value, updateUsage(name, value, usage))
	override(name)

	return v
}

// Int64Var defines an int64 flag with specified name, default value, and usage string.
// The argument p points to an int64 variable in which to store the value of the flag.
func Int64Var(p *int64, name string, value int64, usage string) {
	flag.Int64Var(p, name, value, updateUsage(name, value, usage))
	override(name)
}

// IntVar defines an int flag with specified name, default value, and usage string.
// The argument p points to an int variable in which to store the value of the flag.
func IntVar(p *int, name string, value int, usage string) {
	flag.IntVar(p, name, value, updateUsage(name, value, usage))
	override(name)
}

// NArg is the number of arguments remaining after flags have been processed.
var NArg = flag.NArg

// NFlag returns the number of command-line flags that have been set.
var NFlag = flag.NFlag

// Parse parses the command-line flags from os.Args[1:].
// Must be called after all flags are defined and before flags are accessed by the program.
var Parse = flag.Parse

// Parsed reports whether the command-line flags have been parsed.
var Parsed = flag.Parsed

// PrintDefaults prints, to standard error unless configured otherwise,
// a usage message showing the default settings of all defined
// command-line flags.
// For an integer valued flag x, the default output has the form
//	-x int
//		usage-message-for-x (default 7)
// The usage message will appear on a separate line for anything but
// a bool flag with a one-byte name. For bool flags, the type is
// omitted and if the flag name is one byte the usage message appears
// on the same line. The parenthetical default is omitted if the
// default is the zero value for the type. The listed type, here int,
// can be changed by placing a back-quoted name in the flag's usage
// string; the first such item in the message is taken to be a parameter
// name to show in the message and the back quotes are stripped from
// the message when displayed. For instance, given
//	flag.String("I", "", "search `directory` for include files")
// the output will be
//	-I directory
//		search directory for include files.
//
// To change the destination for flag messages, call CommandLine.SetOutput.
var PrintDefaults = flag.PrintDefaults

// Set sets the value of the named command-line flag.
var Set = flag.Set

// String defines a string flag with specified name, default value, and usage string.
// The return value is the address of a string variable that stores the value of the flag.
func String(name, value, usage string) *string {
	v := flag.String(name, value, updateUsage(name, value, usage))
	override(name)

	return v
}

// StringVar defines a string flag with specified name, default value, and usage string.
// The argument p points to a string variable in which to store the value of the flag.
func StringVar(p *string, name, value, usage string) {
	flag.StringVar(p, name, value, updateUsage(name, value, usage))
	override(name)
}

// Uint defines a uint flag with specified name, default value, and usage string.
// The return value is the address of a uint variable that stores the value of the flag.
func Uint(name string, value uint, usage string) *uint {
	v := flag.Uint(name, value, updateUsage(name, value, usage))
	override(name)

	return v
}

// Uint64 defines a uint64 flag with specified name, default value, and usage string.
// The return value is the address of a uint64 variable that stores the value of the flag.
func Uint64(name string, value uint64, usage string) *uint64 {
	v := flag.Uint64(name, value, updateUsage(name, value, usage))
	override(name)

	return v
}

// Uint64Var defines a uint64 flag with specified name, default value, and usage string.
// The argument p points to a uint64 variable in which to store the value of the flag.
func Uint64Var(p *uint64, name string, value uint64, usage string) {
	flag.Uint64Var(p, name, value, updateUsage(name, value, usage))
	override(name)
}

// UintVar defines a uint flag with specified name, default value, and usage string.
// The argument p points to a uint variable in which to store the value of the flag.
func UintVar(p *uint, name string, value uint, usage string) {
	flag.UintVar(p, name, value, updateUsage(name, value, usage))
	override(name)
}
