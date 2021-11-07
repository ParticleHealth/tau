package config

import (
	"flag"
	"os"
	"strconv"
	"testing"
)

const (
	base10        = 10
	usage         = "ignored usage"
	defaultString = "default"
	changedString = "override"
	defaultBool   = false
	changedBool   = true
	defaultInt    = 0
	changedInt    = 1
)

var (
	basicString *string
	setString   *string
	envString   *string
	basicBool   *bool
	setBool     *bool
	envBool     *bool
	basicInt    *int64
	setInt      *int64
	envInt      *int64
)

func TestMain(m *testing.M) {
	os.Setenv("ENV_STRING", changedString)
	basicString = String("basic_string", defaultString, usage)
	setString = String("set_string", defaultString, usage)
	flag.Set("set_string", changedString)
	envString = String("env_string", defaultString, usage)

	os.Setenv("ENV_BOOL", strconv.FormatBool(changedBool))
	basicBool = Bool("basic_bool", defaultBool, usage)
	setBool = Bool("set_bool", defaultBool, usage)
	flag.Set("set_bool", strconv.FormatBool(changedBool))
	envBool = Bool("env_bool", defaultBool, usage)

	os.Setenv("ENV_INT", strconv.FormatInt(changedInt, base10))
	basicInt = Int64("basc_int", defaultInt, usage)
	setInt = Int64("set_int", defaultInt, usage)
	flag.Set("set_int", strconv.FormatInt(changedInt, base10))
	envInt = Int64("env_int", defaultInt, usage)

	Parse()
	os.Exit(m.Run())
}

func TestDefaults(t *testing.T) {
	t.Parallel()
	if defaultString != *basicString {
		t.Errorf("string default failed. want %s, got %s", defaultString, *basicString)
	}
	if defaultBool != *basicBool {
		t.Errorf("bool default failed. want %t, got %t", defaultBool, *basicBool)
	}
	if defaultInt != *basicInt {
		t.Errorf("int default failed. want %d, got %d", defaultInt, *basicInt)
	}
}

func TestFlags(t *testing.T) {
	t.Parallel()
	if changedString != *setString {
		t.Errorf("string flag override failed. want %s, got %s", changedString, *setString)
	}
	if changedBool != *setBool {
		t.Errorf("bool flag override failed. want %t, got %t", changedBool, *setBool)
	}
	if changedInt != *setInt {
		t.Errorf("int flag override failed. want %d, got %d", changedInt, *setInt)
	}
}

func TestEnvs(t *testing.T) {
	t.Parallel()
	if changedString != *envString {
		t.Errorf("string env override failed. want %s, got %s", changedString, *envString)
	}
	if changedBool != *envBool {
		t.Errorf("bool flag override failed. want %t, got %t", changedBool, *envBool)
	}
	if changedInt != *envInt {
		t.Errorf("int flag override failed. want %d, got %d", changedInt, *envInt)
	}
}
