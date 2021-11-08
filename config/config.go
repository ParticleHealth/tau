package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
)

// updateUsage to reflect ability to set via environment variable.
func updateUsage(name string, value interface{}, usage string) string {
	return fmt.Sprintf("%s\nAlso set by environment variable %s=%T", usage, strings.ToUpper(name), value)
}

// override a flag value based on an environment variable being set.
func override(fs *flag.FlagSet, name string) {
	env := strings.ToUpper(name)
	if v, ok := os.LookupEnv(env); ok {
		err := fs.Set(name, v)
		if err != nil {
			panic(err)
		}
	}
}

func Parse() error {
	return ParseFlagSet(os.Args[1:], flag.CommandLine)
}

func ParseFlagSet(args []string, fs *flag.FlagSet) error {
	if fs.Parsed() {
		return errors.New("config.Parse can only be called once and before flag package Parse")
	}
	fs.VisitAll(func(f *flag.Flag) {
		override(fs, f.Name)
		f.Usage = updateUsage(f.Name, f.Value, f.Usage)
	})

	return fs.Parse(args)
}
