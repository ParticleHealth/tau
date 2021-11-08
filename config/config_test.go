package config

import (
	"bytes"
	"flag"
	"fmt"
	"strings"
	"testing"
	"time"
)

const (
	defaultFlag  = "flag-default"
	defaultValue = "default"
	setFlag      = "flag-set"
	setValue     = "set"
)

func TestFlagStillWorks(t *testing.T) {
	t.Parallel()
	fs := flag.NewFlagSet("test", flag.PanicOnError)
	defaultString := fs.String(defaultFlag, defaultValue, "testing default value")
	setString := fs.String(setFlag, defaultValue, "testing set value")
	if err := ParseFlagSet([]string{fmt.Sprint("-", setFlag, "=", setValue)}, fs); err != nil {
		t.Fatalf("parsing failed: %v", err)
	}
	if !fs.Parsed() {
		t.Error("Parsed not true")
	}
	if *defaultString != defaultValue || *setString != setValue {
		t.Errorf("flag not set correctly\ndefault flag want: %s, got: %s\nset flag want: %s, got: %s",
			defaultValue, *defaultString, setValue, *setString)
	}
}

func TestEnvironmentOverride(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.PanicOnError)
	defaultString := fs.String(defaultFlag, defaultValue, "testing default value")
	setString := fs.String(setFlag, defaultValue, "testing set value")
	t.Setenv(strings.ToUpper(setFlag), setValue)
	if err := ParseFlagSet(nil, fs); err != nil {
		t.Fatalf("parsing failed: %v", err)
	}
	if !fs.Parsed() {
		t.Error("Parsed not true")
	}
	if *defaultString != defaultValue || *setString != setValue {
		t.Errorf("flag not set correctly\ndefault flag want: %s, got: %s\nset flag want: %s, got: %s",
			defaultValue, *defaultString, setValue, *setString)
	}
}

func TestUpdatedUsage(t *testing.T) {
	t.Parallel()
	output := bytes.NewBuffer(make([]byte, 0, 255))
	fs := flag.NewFlagSet("test", flag.PanicOnError)
	fs.SetOutput(output)
	_ = fs.String(defaultFlag, defaultValue, "testing default value")
	if err := ParseFlagSet(nil, fs); err != nil {
		t.Fatalf("parsing failed: %v", err)
	}
	fs.Usage()
	usage := output.String()
	t.Logf("usage:\n%s", usage)
	if !strings.Contains(usage, "Also set by environment variable") {
		t.Errorf("usage not updated to reflect being set by environment variable")
	}
	output.Reset()
	fs.PrintDefaults()
	defs := output.String()
	t.Logf("defaults:\n%s", defs)
}

func TestBadEnvironmentVariableErrors(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.PanicOnError)
	_ = fs.Duration(defaultFlag, 1*time.Second, "testing bad value")
	t.Setenv(strings.ToUpper(defaultFlag), defaultValue)
	if err := ParseFlagSet(nil, fs); err == nil {
		t.Error("expected error but got none")
	}
}

func TestParseCallOrder(t *testing.T) {
	t.Parallel()
	fs := flag.NewFlagSet("test", flag.PanicOnError)
	_ = fs.String(defaultFlag, defaultValue, "testing default value")
	if err := fs.Parse(nil); err != nil {
		t.Errorf("flagset parse failed: %v", err)
	}
	if err := ParseFlagSet(nil, fs); err == nil {
		t.Errorf("already parsed flagset. want: %v, got: %v", nil, err)
	}
}
