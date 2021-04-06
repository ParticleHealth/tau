// Package slog implements a logger formatted to work with Stackdriver structured logs.
package slog

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"

	"go.opencensus.io/trace"
)

// Severity levels as specified in https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#LogSeverity.
type severity string

const (
	severityDefault   severity = "DEFAULT"
	severityDebug     severity = "DEBUG"
	severityInfo      severity = "INFO"
	severityNotice    severity = "NOTICE"
	severityWarn      severity = "WARNING"
	severityError     severity = "ERROR"
	severityCritical  severity = "CRITICAL"
	severityAlert     severity = "ALERT"
	severityEmergency severity = "EMERGENCY"
)

var (
	std     = newLogger(os.Stdout)
	base    = std.entry()
	sources = make(map[uintptr]*SourceLocation)
)

// Logger used to write structured logs in a thread-safe manner to a given output.
type Logger struct {
	mu      sync.Mutex // ensures atomic writes
	out     io.Writer
	project string
}

// Entry with additional metadata included.
// See https://cloud.google.com/logging/docs/agent/configuration#special-fields for reference.
type Entry struct {
	logger         *Logger
	Message        string            `json:"message"`
	Severity       severity          `json:"severity,omitempty"`
	Time           time.Time         `json:"time,omitempty"`
	Labels         map[string]string `json:"logging.googleapis.com/labels,omitempty"`
	SourceLocation *SourceLocation   `json:"logging.googleapis.com/sourceLocation,omitempty"`
	Operation      *Operation        `json:"logging.googleapis.com/operation,omitempty"`
	Trace          string            `json:"logging.googleapis.com/trace,omitempty"`
	SpanID         string            `json:"logging.googleapis.com/spanId,omitempty"`
	TraceSampled   bool              `json:"logging.googleapis.com/trace_sampled,omitempty"`
}

// SourceLocation that originated the log call.
type SourceLocation struct {
	File     string `json:"file,omitempty"`
	Line     string `json:"line,omitempty"`
	Function string `json:"function,omitempty"`
}

// The Operation a given log entry is part of.
type Operation struct {
	ID       string `json:"id,omitempty"`
	Producer string `json:"producer,omitempty"`
	First    bool   `json:"first,omitempty"`
	Last     bool   `json:"last,omitempty"`
}

// startOperation with a given ID and producer.
// Will log the start of the operation at Notice level.
func (e *Entry) startOperation(id, producer string) *Entry {
	e.Operation = &Operation{
		ID:       id,
		Producer: producer,
		First:    true,
		Last:     false,
	}
	e.logger.log(e, severityNotice, fmt.Sprint(producer, " starting operation ", id), 3)
	e.Operation.First = false
	return e
}

// StartOperation with a given ID and producer.
// Will log the start of the operation at Notice level.
func StartOperation(id, producer string) *Entry {
	return std.entry().startOperation(id, producer)
}

// StartOperation with a given ID and producer.
// Will log the start of the operation at Notice level.
func (e *Entry) StartOperation(id, producer string) *Entry {
	return e.startOperation(id, producer)
}

// StartOperation with a given ID and producer.
// Will log the start of the operation at Notice level.
func (l *Logger) StartOperation(id, producer string) *Entry {
	return l.entry().startOperation(id, producer)
}

// EndOperation stops any current operation and further logs will no longer include.
// Will log the end of the operation at Notice level.
func (e *Entry) EndOperation() {
	if e.Operation == nil {
		return
	}
	e.Operation.Last = true
	e.logger.log(e, severityNotice, fmt.Sprint(e.Operation.Producer, " ending operation ", e.Operation.ID), 2)
	e.Operation = nil
}

// WithOperation details included in all logs written for a given Entry.
func (e *Entry) WithOperation(id, producer string) *Entry {
	e.Operation = &Operation{
		ID:       id,
		Producer: producer,
		First:    false,
		Last:     false,
	}
	return e
}

// WithOperation details included in all logs written for a given Entry.
func WithOperation(id, producer string) *Entry {
	return std.entry().WithOperation(id, producer)
}

// WithOperation details included in all logs written for a given Entry.
func (l *Logger) WithOperation(id, producer string) *Entry {
	return l.entry().WithOperation(id, producer)
}

// WithSpan details included for a given Trace.
func (e *Entry) WithSpan(s *trace.Span) *Entry {
	e.Trace = fmt.Sprint("projects/", e.logger.project, "/traces/", s.SpanContext().TraceID)
	e.SpanID = s.SpanContext().SpanID.String()
	e.TraceSampled = s.SpanContext().IsSampled()
	return e
}

// WithSpan details included for a given Trace.
func WithSpan(s *trace.Span) *Entry {
	return std.entry().WithSpan(s)
}

// WithSpan details included for a given Trace.
func (l *Logger) WithSpan(s *trace.Span) *Entry {
	return l.entry().WithSpan(s)
}

// WithLabel including details for a given Entry.
func (e *Entry) WithLabel(k string, v interface{}) *Entry {
	if e.Labels == nil {
		e.Labels = make(map[string]string)
	}
	e.Labels[k] = fmt.Sprint(v)
	return e
}

// WithLabel including details for a given Entry.
func WithLabel(k string, v interface{}) *Entry {
	return std.entry().WithLabel(k, v)
}

// WithLabel including details for a given Entry.
func (l *Logger) WithLabel(k string, v interface{}) *Entry {
	return l.entry().WithLabel(k, v)
}

// newLogger with provided options.
func newLogger(out io.Writer) *Logger {
	return &Logger{out: out}
}

// entry creates a new Entry allowing for reusing details across multiple log calls.
func (l *Logger) entry() *Entry {
	return &Entry{logger: l}
}

// SetOutput destination for the logger.
func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.out = w
}

// SetOutput destination for the package-level logger.
func SetOutput(w io.Writer) {
	std.SetOutput(w)
}

// SetProject for the logger.
// Used for things such as traces that require project to be included.
func (l *Logger) SetProject(project string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.project = project
}

// SetProject for the package-level logger.
// Used for things such as traces that require project to be included.
func SetProject(project string) {
	std.SetProject(project)
}

// getSource from reflection, caches where possible to shave some time off.
func getSource(depth int) *SourceLocation {
	pc, file, line, ok := runtime.Caller(depth + 1)
	if !ok {
		return nil
	}
	if s := sources[pc]; s != nil {
		return s
	}
	fn := runtime.FuncForPC(pc)
	s := &SourceLocation{
		File:     file,
		Line:     fmt.Sprint(line),
		Function: fn.Name(),
	}
	sources[pc] = s
	return s
}

// log with given parameters.
func (l *Logger) log(e *Entry, s severity, m string, depth int) {
	// Do costly operations prior to grabbing mutex
	time := time.Now()
	source := getSource(depth)

	l.mu.Lock()
	defer l.mu.Unlock()

	e.Severity = s
	e.Message = m
	e.Time = time
	e.SourceLocation = source

	if b, err := json.Marshal(e); err != nil {
		fmt.Fprintln(l.out, "could not marshal log:", err)
	} else {
		fmt.Fprintln(l.out, string(b))
	}
}

// Debug sends a message to the logger with severity Debug.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Debug(v ...interface{}) {
	l.log(base, severityDebug, fmt.Sprint(v...), 2)
}

// Debug sends a message to the default logger with severity Debug.
// Arguments are handled in the manner of fmt.Print.
func Debug(v ...interface{}) {
	std.log(base, severityDebug, fmt.Sprint(v...), 2)
}

// Debug sends a message to the logger associated with this entry with severity Debug.
// Arguments are handled in the manner of fmt.Print.
func (e *Entry) Debug(v ...interface{}) {
	e.logger.log(e, severityDebug, fmt.Sprint(v...), 2)
}

// Debugf sends a message to the logger with severity Debug.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Debugf(format string, v ...interface{}) {
	l.log(base, severityDebug, fmt.Sprintf(format, v...), 2)
}

// Debugf sends a message to the default logger with severity Debug.
// Arguments are handled in the manner of fmt.Printf.
func Debugf(format string, v ...interface{}) {
	std.log(base, severityDebug, fmt.Sprintf(format, v...), 2)
}

// Debugf sends a message to the logger associated with this entry with severity Debug.
// Arguments are handled in the manner of fmt.Printf.
func (e *Entry) Debugf(format string, v ...interface{}) {
	e.logger.log(e, severityDebug, fmt.Sprintf(format, v...), 2)
}

// Info sends a message to the logger with severity Info.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Info(v ...interface{}) {
	l.log(base, severityInfo, fmt.Sprint(v...), 2)
}

// Info sends a message to the default logger with severity Info.
// Arguments are handled in the manner of fmt.Print.
func Info(v ...interface{}) {
	std.log(base, severityInfo, fmt.Sprint(v...), 2)
}

// Info sends a message to the logger associated with this entry with severity Info.
// Arguments are handled in the manner of fmt.Print.
func (e *Entry) Info(v ...interface{}) {
	e.logger.log(e, severityInfo, fmt.Sprint(v...), 2)
}

// Infof sends a message to the logger with severity Info.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Infof(format string, v ...interface{}) {
	l.log(base, severityInfo, fmt.Sprintf(format, v...), 2)
}

// Infof sends a message to the default logger with severity Info.
// Arguments are handled in the manner of fmt.Printf.
func Infof(format string, v ...interface{}) {
	std.log(base, severityInfo, fmt.Sprintf(format, v...), 2)
}

// Infof sends a message to the logger associated with this entry with severity Info.
// Arguments are handled in the manner of fmt.Printf.
func (e *Entry) Infof(format string, v ...interface{}) {
	e.logger.log(e, severityInfo, fmt.Sprintf(format, v...), 2)
}

// Notice sends a message to the logger with severity Notice.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Notice(v ...interface{}) {
	l.log(base, severityNotice, fmt.Sprint(v...), 2)
}

// Notice sends a message to the default logger with severity Notice.
// Arguments are handled in the manner of fmt.Print.
func Notice(v ...interface{}) {
	std.log(base, severityNotice, fmt.Sprint(v...), 2)
}

// Notice sends a message to the logger associated with this entry with severity Notice.
// Arguments are handled in the manner of fmt.Print.
func (e *Entry) Notice(v ...interface{}) {
	e.logger.log(e, severityNotice, fmt.Sprint(v...), 2)
}

// Noticef sends a message to the logger with severity Notice.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Noticef(format string, v ...interface{}) {
	l.log(base, severityNotice, fmt.Sprintf(format, v...), 2)
}

// Noticef sends a message to the default logger with severity Notice.
// Arguments are handled in the manner of fmt.Printf.
func Noticef(format string, v ...interface{}) {
	std.log(base, severityNotice, fmt.Sprintf(format, v...), 2)
}

// Noticef sends a message to the logger associated with this entry with severity Notice.
// Arguments are handled in the manner of fmt.Printf.
func (e *Entry) Noticef(format string, v ...interface{}) {
	e.logger.log(e, severityNotice, fmt.Sprintf(format, v...), 2)
}

// Warn sends a message to the logger with severity Warn.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Warn(v ...interface{}) {
	l.log(base, severityWarn, fmt.Sprint(v...), 2)
}

// Warn sends a message to the default logger with severity Warn.
// Arguments are handled in the manner of fmt.Print.
func Warn(v ...interface{}) {
	std.log(base, severityWarn, fmt.Sprint(v...), 2)
}

// Warn sends a message to the logger associated with this entry with severity Warn.
// Arguments are handled in the manner of fmt.Print.
func (e *Entry) Warn(v ...interface{}) {
	e.logger.log(e, severityWarn, fmt.Sprint(v...), 2)
}

// Warnf sends a message to the logger with severity Warn.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Warnf(format string, v ...interface{}) {
	l.log(base, severityWarn, fmt.Sprintf(format, v...), 2)
}

// Warnf sends a message to the default logger with severity Warn.
// Arguments are handled in the manner of fmt.Printf.
func Warnf(format string, v ...interface{}) {
	std.log(base, severityWarn, fmt.Sprintf(format, v...), 2)
}

// Warnf sends a message to the logger associated with this entry with severity Warn.
// Arguments are handled in the manner of fmt.Printf.
func (e *Entry) Warnf(format string, v ...interface{}) {
	e.logger.log(e, severityWarn, fmt.Sprintf(format, v...), 2)
}

// Error sends a message to the logger with severity Error.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Error(v ...interface{}) {
	l.log(base, severityError, fmt.Sprint(v...), 2)
}

// Error sends a message to the default logger with severity Error.
// Arguments are handled in the manner of fmt.Print.
func Error(v ...interface{}) {
	std.log(base, severityError, fmt.Sprint(v...), 2)
}

// Error sends a message to the logger associated with this entry with severity Error.
// Arguments are handled in the manner of fmt.Print.
func (e *Entry) Error(v ...interface{}) {
	e.logger.log(e, severityError, fmt.Sprint(v...), 2)
}

// Errorf sends a message to the logger with severity Error.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.log(base, severityError, fmt.Sprintf(format, v...), 2)
}

// Errorf sends a message to the default logger with severity Error.
// Arguments are handled in the manner of fmt.Printf.
func Errorf(format string, v ...interface{}) {
	std.log(base, severityError, fmt.Sprintf(format, v...), 2)
}

// Errorf sends a message to the logger associated with this entry with severity Error.
// Arguments are handled in the manner of fmt.Printf.
func (e *Entry) Errorf(format string, v ...interface{}) {
	e.logger.log(e, severityError, fmt.Sprintf(format, v...), 2)
}

// Critical sends a message to the logger with severity Critical.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Critical(v ...interface{}) {
	l.log(base, severityCritical, fmt.Sprint(v...), 2)
}

// Critical sends a message to the default logger with severity Critical.
// Arguments are handled in the manner of fmt.Print.
func Critical(v ...interface{}) {
	std.log(base, severityCritical, fmt.Sprint(v...), 2)
}

// Critical sends a message to the logger associated with this entry with severity Critical.
// Arguments are handled in the manner of fmt.Print.
func (e *Entry) Critical(v ...interface{}) {
	e.logger.log(e, severityCritical, fmt.Sprint(v...), 2)
}

// Criticalf sends a message to the logger with severity Critical.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Criticalf(format string, v ...interface{}) {
	l.log(base, severityCritical, fmt.Sprintf(format, v...), 2)
}

// Criticalf sends a message to the default logger with severity Critical.
// Arguments are handled in the manner of fmt.Printf.
func Criticalf(format string, v ...interface{}) {
	std.log(base, severityCritical, fmt.Sprintf(format, v...), 2)
}

// Criticalf sends a message to the logger associated with this entry with severity Critical.
// Arguments are handled in the manner of fmt.Printf.
func (e *Entry) Criticalf(format string, v ...interface{}) {
	e.logger.log(e, severityCritical, fmt.Sprintf(format, v...), 2)
}

// Alert sends a message to the logger with severity Alert.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Alert(v ...interface{}) {
	l.log(base, severityAlert, fmt.Sprint(v...), 2)
}

// Alert sends a message to the default logger with severity Alert.
// Arguments are handled in the manner of fmt.Print.
func Alert(v ...interface{}) {
	std.log(base, severityAlert, fmt.Sprint(v...), 2)
}

// Alert sends a message to the logger associated with this entry with severity Alert.
// Arguments are handled in the manner of fmt.Print.
func (e *Entry) Alert(v ...interface{}) {
	e.logger.log(e, severityAlert, fmt.Sprint(v...), 2)
}

// Alertf sends a message to the logger with severity Alert.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Alertf(format string, v ...interface{}) {
	l.log(base, severityAlert, fmt.Sprintf(format, v...), 2)
}

// Alertf sends a message to the default logger with severity Alert.
// Arguments are handled in the manner of fmt.Printf.
func Alertf(format string, v ...interface{}) {
	std.log(base, severityAlert, fmt.Sprintf(format, v...), 2)
}

// Alertf sends a message to the logger associated with this entry with severity Alert.
// Arguments are handled in the manner of fmt.Printf.
func (e *Entry) Alertf(format string, v ...interface{}) {
	e.logger.log(e, severityAlert, fmt.Sprintf(format, v...), 2)
}

// Emergency sends a message to the logger with severity Emergency.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Emergency(v ...interface{}) {
	l.log(base, severityEmergency, fmt.Sprint(v...), 2)
}

// Emergency sends a message to the default logger with severity Emergency.
// Arguments are handled in the manner of fmt.Print.
func Emergency(v ...interface{}) {
	std.log(base, severityEmergency, fmt.Sprint(v...), 2)
}

// Emergency sends a message to the logger associated with this entry with severity Emergency.
// Arguments are handled in the manner of fmt.Print.
func (e *Entry) Emergency(v ...interface{}) {
	e.logger.log(e, severityEmergency, fmt.Sprint(v...), 2)
}

// Emergencyf sends a message to the logger with severity Emergency.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Emergencyf(format string, v ...interface{}) {
	l.log(base, severityEmergency, fmt.Sprintf(format, v...), 2)
}

// Emergencyf sends a message to the default logger with severity Emergency.
// Arguments are handled in the manner of fmt.Printf.
func Emergencyf(format string, v ...interface{}) {
	std.log(base, severityEmergency, fmt.Sprintf(format, v...), 2)
}

// Emergencyf sends a message to the logger associated with this entry with severity Emergency.
// Arguments are handled in the manner of fmt.Printf.
func (e *Entry) Emergencyf(format string, v ...interface{}) {
	e.logger.log(e, severityEmergency, fmt.Sprintf(format, v...), 2)
}
