package deverr

import (
	"errors"
	"strings"
	"testing"
)

func TestVerifyPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("did not panic")
		}
	}()
	err := errors.New("test error")
	Verify(err)
}

func TestVerifyContinues(t *testing.T) {
	var err error
	Verify(err)
}

func TestNotImplemented(t *testing.T) {
	err := NotImplemented()
	if err == nil {
		t.Fatal("no err returned")
	}
	e, ok := err.(*NotImplementedError)
	if !ok {
		t.Fatal("err was not right type")
	}
	want := "tau/deverr/deverr_test.go"
	if !strings.Contains(e.File, want) {
		t.Errorf("file not set correctly. want contains: %s, got: %s", want, e.File)
	}
	want = "github.com/ParticleHealth/tau/deverr.TestNotImplemented"
	if e.Fn != want {
		t.Errorf("fn not set correctly. want: %s, got: %s", want, e.Fn)
	}
	if e.Line <= 0 {
		t.Errorf("line not set correctly. want >= 0, got: %d", e.Line)
	}
	want = "is not implemented"
	if !strings.Contains(e.Error(), want) {
		t.Errorf("error string not correct. want contains: %s, got %s", want, e.Error())
	}
}
