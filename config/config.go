// Package config adds additional functionality to the std flag package to allow overriding with environment variables.
package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
)

// updateUsage to reflect ability to set via environment variable.
func updateUsage(name, usage string) string {
	return fmt.Sprintf("%s\nAlso set by environment variable %s", usage, strings.ToUpper(name))
}

// override a flag value based on an environment variable being set.
func override(fs *flag.FlagSet, name string) error {
	env := strings.ToUpper(name)
	if v, ok := os.LookupEnv(env); ok {
		err := fs.Set(name, v)
		if err != nil {
			return fmt.Errorf("could not set %s to %s: %w", name, v, err)
		}
	}

	return nil
}

// Parse parses the command-line flags from os.Args[1:]. Must be called
// after all flags are defined, before calls to package flag Parse functions
// and before flags are accessed by the program.
func Parse() error {
	return ParseFlagSet(os.Args[1:], flag.CommandLine)
}

// Parse parses flag definitions from the argument list, which should not include the command name.
// Must be called after all flags in the FlagSet are defined, before calls to package flag Parse
// functions  and before flags are accessed by the program.
// The return value will be ErrHelp if -help or -h were set but not defined.
func ParseFlagSet(args []string, fs *flag.FlagSet) error {
	if fs.Parsed() {
		return errors.New("config.Parse can only be called once and before flag package Parse")
	}
	var errs []string
	fs.VisitAll(func(f *flag.Flag) {
		if err := override(fs, f.Name); err != nil {
			errs = append(errs, err.Error())
		}
		f.Usage = updateUsage(f.Name, f.Usage)
	})
	if len(errs) != 0 {
		return fmt.Errorf("parsing flags: %s", strings.Join(errs, "; "))
	}

	return fs.Parse(args)
}
