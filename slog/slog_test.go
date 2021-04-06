package slog

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
)

const (
	defaultMessage   = "hello"
	formattedMessage = "works: %t"
	formattedResult  = "works: true"
)

var (
	buf = bytes.NewBuffer(make([]byte, 0))
)

func TestMain(m *testing.M) {
	SetOutput(buf)
	os.Exit(m.Run())
}

func subTestSeverity(t *testing.T, level string, f func()) {
	f()
	got := buf.String()
	want := fmt.Sprintf(`"severity":"%s"`, level)
	buf.Reset()
	if !strings.Contains(got, want) {
		t.Errorf("did not get severity level %s serialized\nwanted: %s\ngot: %s", level, want, got)
	}
}

func TestSeverities(t *testing.T) {
	subTestSeverity(t, "DEBUG", func() { Debug(defaultMessage) })
	subTestSeverity(t, "DEBUG", func() { Debugf(defaultMessage) })
	subTestSeverity(t, "DEBUG", func() { base.Debug(defaultMessage) })
	subTestSeverity(t, "DEBUG", func() { base.Debugf(defaultMessage) })
	subTestSeverity(t, "DEBUG", func() { std.Debug(defaultMessage) })
	subTestSeverity(t, "DEBUG", func() { std.Debugf(defaultMessage) })
	subTestSeverity(t, "INFO", func() { Info(defaultMessage) })
	subTestSeverity(t, "INFO", func() { Infof(defaultMessage) })
	subTestSeverity(t, "INFO", func() { base.Info(defaultMessage) })
	subTestSeverity(t, "INFO", func() { base.Infof(defaultMessage) })
	subTestSeverity(t, "INFO", func() { std.Info(defaultMessage) })
	subTestSeverity(t, "INFO", func() { std.Infof(defaultMessage) })
	subTestSeverity(t, "NOTICE", func() { Notice(defaultMessage) })
	subTestSeverity(t, "NOTICE", func() { Noticef(defaultMessage) })
	subTestSeverity(t, "NOTICE", func() { base.Notice(defaultMessage) })
	subTestSeverity(t, "NOTICE", func() { base.Noticef(defaultMessage) })
	subTestSeverity(t, "NOTICE", func() { std.Notice(defaultMessage) })
	subTestSeverity(t, "NOTICE", func() { std.Noticef(defaultMessage) })
	subTestSeverity(t, "WARNING", func() { Warn(defaultMessage) })
	subTestSeverity(t, "WARNING", func() { Warnf(defaultMessage) })
	subTestSeverity(t, "WARNING", func() { base.Warn(defaultMessage) })
	subTestSeverity(t, "WARNING", func() { base.Warnf(defaultMessage) })
	subTestSeverity(t, "WARNING", func() { std.Warn(defaultMessage) })
	subTestSeverity(t, "WARNING", func() { std.Warnf(defaultMessage) })
	subTestSeverity(t, "ERROR", func() { Error(defaultMessage) })
	subTestSeverity(t, "ERROR", func() { Errorf(defaultMessage) })
	subTestSeverity(t, "ERROR", func() { base.Error(defaultMessage) })
	subTestSeverity(t, "ERROR", func() { base.Errorf(defaultMessage) })
	subTestSeverity(t, "ERROR", func() { std.Error(defaultMessage) })
	subTestSeverity(t, "ERROR", func() { std.Errorf(defaultMessage) })
	subTestSeverity(t, "CRITICAL", func() { Critical(defaultMessage) })
	subTestSeverity(t, "CRITICAL", func() { Criticalf(defaultMessage) })
	subTestSeverity(t, "CRITICAL", func() { base.Critical(defaultMessage) })
	subTestSeverity(t, "CRITICAL", func() { base.Criticalf(defaultMessage) })
	subTestSeverity(t, "CRITICAL", func() { std.Critical(defaultMessage) })
	subTestSeverity(t, "CRITICAL", func() { std.Criticalf(defaultMessage) })
	subTestSeverity(t, "ALERT", func() { Alert(defaultMessage) })
	subTestSeverity(t, "ALERT", func() { Alertf(defaultMessage) })
	subTestSeverity(t, "ALERT", func() { base.Alert(defaultMessage) })
	subTestSeverity(t, "ALERT", func() { base.Alertf(defaultMessage) })
	subTestSeverity(t, "ALERT", func() { std.Alert(defaultMessage) })
	subTestSeverity(t, "ALERT", func() { std.Alertf(defaultMessage) })
	subTestSeverity(t, "EMERGENCY", func() { Emergency(defaultMessage) })
	subTestSeverity(t, "EMERGENCY", func() { Emergencyf(defaultMessage) })
	subTestSeverity(t, "EMERGENCY", func() { base.Emergency(defaultMessage) })
	subTestSeverity(t, "EMERGENCY", func() { base.Emergencyf(defaultMessage) })
	subTestSeverity(t, "EMERGENCY", func() { std.Emergency(defaultMessage) })
	subTestSeverity(t, "EMERGENCY", func() { std.Emergencyf(defaultMessage) })
}

func subTestFormat(t *testing.T, f func(string, ...interface{})) {
	f(formattedMessage, true)
	got := buf.String()
	buf.Reset()
	if !strings.Contains(got, formattedResult) {
		t.Errorf("message not formatted\nwanted: %s\ngot: %s", formattedResult, got)
	}
}

func TestFormats(t *testing.T) {
	subTestFormat(t, Debugf)
	subTestFormat(t, base.Debugf)
	subTestFormat(t, std.Debugf)
	subTestFormat(t, Infof)
	subTestFormat(t, base.Infof)
	subTestFormat(t, std.Infof)
	subTestFormat(t, Noticef)
	subTestFormat(t, base.Noticef)
	subTestFormat(t, std.Noticef)
	subTestFormat(t, Warnf)
	subTestFormat(t, base.Warnf)
	subTestFormat(t, std.Warnf)
	subTestFormat(t, Errorf)
	subTestFormat(t, base.Errorf)
	subTestFormat(t, std.Errorf)
	subTestFormat(t, Criticalf)
	subTestFormat(t, base.Criticalf)
	subTestFormat(t, std.Criticalf)
	subTestFormat(t, Alertf)
	subTestFormat(t, base.Alertf)
	subTestFormat(t, std.Alertf)
	subTestFormat(t, Emergencyf)
	subTestFormat(t, base.Emergencyf)
	subTestFormat(t, std.Emergencyf)
}

func TestSettingProject(t *testing.T) {
	if std.project != "" {
		t.Errorf("unexpected project for default logger\nwant:\ngot: %v", std.project)
	}
	SetProject("test")
	if std.project != "test" {
		t.Errorf("unexpected project for default logger after being set\nwant: test\ngot: %v", std.project)
	}
}

func TestLabels(t *testing.T) {
	e := WithLabel("hello", "world")
	e.Info("testing")
	got := buf.String()
	buf.Reset()
	if !strings.Contains(got, "logging.googleapis.com/labels") {
		t.Errorf("labels not included\ngot: %v", got)
	}
	if !strings.Contains(got, `"hello":"world"`) {
		t.Errorf("hello label not included\ngot: %v", got)
	}
	e.WithLabel("another", 1)
	e.Info("testing")
	got = buf.String()
	buf.Reset()
	if !strings.Contains(got, `"another":"1"`) {
		t.Errorf("another label not included\ngot: %v", got)
	}
	if !strings.Contains(got, `"hello":"world"`) {
		t.Errorf("original label removed\ngot: %v", got)
	}
	Info("testing")
	got = buf.String()
	buf.Reset()
	if strings.Contains(got, "logging.googleapis.com/labels") {
		t.Errorf("labels persist when they shouldn't\ngot: %v", got)
	}
}

func BenchmarkPackageLogging(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Info(defaultMessage)
		buf.Reset()
	}
}

func BenchmarkEntryLogging(b *testing.B) {
	for i := 0; i < b.N; i++ {
		base.Info(defaultMessage)
		buf.Reset()
	}
}

func BenchmarkLoggerLogging(b *testing.B) {
	for i := 0; i < b.N; i++ {
		std.Info(defaultMessage)
		buf.Reset()
	}
}
