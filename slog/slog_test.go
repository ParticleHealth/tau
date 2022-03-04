package slog

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"go.opencensus.io/trace"
)

const (
	defaultMessage   = "hello"
	formattedMessage = "works: %t"
	formattedResult  = "works: true"
)

var buf = bytes.NewBuffer(make([]byte, 0, 4096))

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

func TestSettingSources(t *testing.T) {
	if !std.sources {
		t.Errorf("unexpected sources not included by default")
	}
	SetIncludeSources(false)
	Info("testing")
	got := buf.String()
	buf.Reset()
	if strings.Contains(got, "logging.googleapis.com/sourceLocation") {
		t.Errorf("unexpected sources included: %s", got)
	}
	SetIncludeSources(true)
	Info("testing")
	got = buf.String()
	buf.Reset()
	if !strings.Contains(got, "logging.googleapis.com/sourceLocation") {
		t.Errorf("unexpected sources not included: %s", got)
	}
	if !strings.Contains(got, `"function":"github.com/ParticleHealth/tau/slog.TestSettingSources"`) {
		t.Errorf("function not properly set: %s", got)
	}
}

func TestOperations(t *testing.T) {
	// Logger level
	e := std.WithOperation("123", "testProducer")
	e.Info("testing")
	got := buf.String()
	buf.Reset()
	if !strings.Contains(got, "logging.googleapis.com/operation") {
		t.Errorf("logger: operation not included\ngot: %v", got)
	}
	if !strings.Contains(got, `"producer":"testProducer"`) {
		t.Errorf("logger: producer not included\nwant: testProducer\ngot: %s", got)
	}
	if !strings.Contains(got, `"id":"123"`) {
		t.Errorf("logger: id not included\n want: 123\ngot: %s", got)
	}
	_ = std.StartOperation("123", "testProducer")
	got = buf.String()
	buf.Reset()
	if got == "" {
		t.Error("logger: start operation did not create a log")
	}
	// Package level
	e = WithOperation("123", "testProducer")
	e.Info("testing")
	got = buf.String()
	buf.Reset()
	if !strings.Contains(got, "logging.googleapis.com/operation") {
		t.Errorf("package: operation not included\ngot: %v", got)
	}
	if !strings.Contains(got, `"producer":"testProducer"`) {
		t.Errorf("package: producer not included\nwant: testProducer\ngot: %s", got)
	}
	if !strings.Contains(got, `"id":"123"`) {
		t.Errorf("package: id not included\n want: 123\ngot: %s", got)
	}
	_ = StartOperation("123", "testProducer")
	got = buf.String()
	buf.Reset()
	if got == "" {
		t.Error("package: start operation did not create a log")
	}
	// Entry level
	e = base.WithOperation("123", "testProducer")
	e.Info("testing")
	got = buf.String()
	buf.Reset()
	if !strings.Contains(got, "logging.googleapis.com/operation") {
		t.Errorf("entry: operation not included\ngot: %v", got)
	}
	if !strings.Contains(got, `"producer":"testProducer"`) {
		t.Errorf("entry: producer not included\nwant: testProducer\ngot: %s", got)
	}
	if !strings.Contains(got, `"id":"123"`) {
		t.Errorf("entry: id not included\n want: 123\ngot: %s", got)
	}
	e = base.StartOperation("123", "testProducer")
	got = buf.String()
	buf.Reset()
	if got == "" {
		t.Error("entry: start operation did not create a log")
	}
	// End operations for Entry level only.
	e.EndOperation()
	got = buf.String()
	buf.Reset()
	if got == "" {
		t.Error("end operation did not create a log")
	}
	base.EndOperation()
	got = buf.String()
	buf.Reset()
	if got != "" {
		t.Errorf("empty operation wrote a log: %s", buf)
	}
}

func TestLabels(t *testing.T) {
	// Logger level
	e := std.WithLabels(map[string]interface{}{"hello": "world"})
	e.Info("testing")
	got := buf.String()
	buf.Reset()
	if !strings.Contains(got, "logging.googleapis.com/labels") {
		t.Errorf("labels not included\ngot: %v", got)
	}
	if !strings.Contains(got, `"hello":"world"`) {
		t.Errorf("hello label not included\ngot: %v", got)
	}
	// Entry level
	e = e.WithLabels(map[string]interface{}{"another": 1})
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
	// Package level
	e = WithLabels(map[string]interface{}{"hello2": "world2"})
	e.Info("testing")
	got = buf.String()
	if !strings.Contains(got, "logging.googleapis.com/labels") {
		t.Errorf("labels not included\ngot: %v", got)
	}
}

func TestError(t *testing.T) {
	errA := errors.New("error msg A")
	errB := errors.New("error msg B")
	// Logger level
	e := std.WithError(errA)
	e.Info("testing")
	got := buf.String()
	buf.Reset()
	if !strings.Contains(got, "error") {
		t.Errorf("error not included\ngot: %v", got)
	}
	if e.Err.Error() != errA.Error() {
		t.Errorf("error not included\ngot: %v", e.Err.Error())
	}
	// Entry level
	e = e.WithError(errB)
	e.Info("testing")
	buf.Reset()
	if e.Err.Error() != errB.Error() {
		t.Errorf("error not included\ngot: %v", e.Err.Error())
	}
	if e.Err.Error() == errA.Error() {
		t.Errorf("error not included\ngot: %v", e.Err.Error())
	}
	Info("testing")
	got = buf.String()
	buf.Reset()
	if strings.Contains(got, "error") {
		t.Errorf("error persist when it shouldn't\ngot: %v", got)
	}
	// Package level
	e = WithError(errA)
	e.Info("testing")
	got = buf.String()
	if !strings.Contains(got, "error") {
		t.Errorf("error not included\ngot: %v", got)
	}
}

func TestDetail(t *testing.T) {
	// Logger level
	e := std.WithDetail("hello", "world")
	e.Info("testing")
	got := buf.String()
	buf.Reset()
	if !strings.Contains(got, "details") {
		t.Errorf("details not included\ngot: %v", got)
	}
	if !strings.Contains(got, `"hello":"world"`) {
		t.Errorf("hello detail not included\ngot: %v", got)
	}
	// Entry level
	e = e.WithDetail("another", 1)
	e.Info("testing")
	got = buf.String()
	buf.Reset()
	if !strings.Contains(got, `"another":1`) {
		t.Errorf("another detail not included\ngot: %v", got)
	}
	if !strings.Contains(got, `"hello":"world"`) {
		t.Errorf("original detail removed\ngot: %v", got)
	}
	Info("testing")
	got = buf.String()
	buf.Reset()
	if strings.Contains(got, "details") {
		t.Errorf("details persist when they shouldn't\ngot: %v", got)
	}
	// Package level
	e = WithDetail("hello2", "world2")
	e.Info("testing")
	got = buf.String()
	if !strings.Contains(got, "details") {
		t.Errorf("details not included\ngot: %v", got)
	}
}

func TestDetails(t *testing.T) {
	// Logger level
	e := std.WithDetails(map[string]interface{}{"hello": "world"})
	e.Info("testing")
	got := buf.String()
	buf.Reset()
	if !strings.Contains(got, "details") {
		t.Errorf("labels not included\ngot: %v", got)
	}
	if !strings.Contains(got, `"hello":"world"`) {
		t.Errorf("hello label not included\ngot: %v", got)
	}
	// Entry level
	e = e.WithDetails(map[string]interface{}{"another": 1})
	e.Info("testing")
	got = buf.String()
	buf.Reset()
	if !strings.Contains(got, `"another":1`) {
		t.Errorf("another label not included\ngot: %v", got)
	}
	if !strings.Contains(got, `"hello":"world"`) {
		t.Errorf("original label removed\ngot: %v", got)
	}
	Info("testing")
	got = buf.String()
	buf.Reset()
	if strings.Contains(got, "details") {
		t.Errorf("labels persist when they shouldn't\ngot: %v", got)
	}
	// Package level
	e = WithDetails(map[string]interface{}{"hello2": "world2"})
	e.Info("testing")
	got = buf.String()
	if !strings.Contains(got, "details") {
		t.Errorf("labels not included\ngot: %v", got)
	}
}

func TestSpans(t *testing.T) {
	// Make sure no spans by default
	Info("testing")
	got := buf.String()
	buf.Reset()
	if strings.Contains(got, "logging.googleapis.com/trace") {
		t.Errorf("unexpected trace present: %s", got)
	}
	if strings.Contains(got, "logging.googleapis.com/spanId") {
		t.Errorf("unexpected span present: %s", got)
	}
	_, span := trace.StartSpan(context.Background(), "testSpan")
	// Package level
	e := WithSpan(span.SpanContext())
	e.Info("testing")
	got = buf.String()
	buf.Reset()
	if !strings.Contains(got, "logging.googleapis.com/trace") {
		t.Errorf("package: trace not present: %s", got)
	}
	if !strings.Contains(got, "logging.googleapis.com/spanId") {
		t.Errorf("package: span not present: %s", got)
	}
	// Logger level
	e = std.WithSpan(span.SpanContext())
	e.Info("testing")
	got = buf.String()
	buf.Reset()
	if !strings.Contains(got, "logging.googleapis.com/trace") {
		t.Errorf("logger: trace not present: %s", got)
	}
	if !strings.Contains(got, "logging.googleapis.com/spanId") {
		t.Errorf("logger: span not present: %s", got)
	}
}

func TestContext(t *testing.T) {
	newEntry := FromContext(context.Background())
	if diff := cmp.Diff(newEntry, std.entry(), cmpopts.IgnoreUnexported(Entry{})); diff != "" {
		t.Errorf("failed to return new entry when one is not available:\n%s", diff)
	}

	want := std.entry().WithDetails(map[string]interface{}{"hello": "world"})
	ctx := NewContext(context.Background(), want)
	got := FromContext(ctx)
	if diff := cmp.Diff(want, got, cmpopts.IgnoreUnexported(Entry{})); diff != "" {
		t.Errorf("failed to retrieve correct entry from context:\n%s", diff)
	}
}

func TestRaces(t *testing.T) {
	for i := 0; i < 10; i++ {
		go Info("hello")
		go Info("hello")
		go Info("hello")
	}
}
