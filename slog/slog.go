// Package slog implements a logger formatted to work with Stackdriver structured logs.
package slog

import (
	"context"
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

type key int

type Fields map[string]interface{}

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
	std      = newLogger(os.Stdout)
	base     = std.entry()
	sources  = make(map[uintptr]*SourceLocation)
	entryKey key
)

// Logger used to write structured logs in a thread-safe manner to a given output.
type Logger struct {
	mu      sync.Mutex // ensures atomic writes
	out     io.Writer
	sources bool
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
	Details        Fields            `json:"details,omitempty"`
	Err            error             `json:"error,omitempty"`
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

// clone a given Entry so that changes to it do not affect the parent.
func (e *Entry) clone() *Entry {
	next := *e
	if next.Labels != nil {
		next.Labels = make(map[string]string)
		for k, v := range e.Labels {
			next.Labels[k] = v
		}
	}
	if next.Details != nil {
		next.Details = make(Fields)
		for k, v := range e.Details {
			next.Details[k] = v
		}
	}
	return &next
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

// WithSpan details included for a given Trace. Will create a child entry.
func (e *Entry) WithSpan(s *trace.Span) *Entry {
	c := e.clone()
	c.Trace = fmt.Sprint("projects/", e.logger.project, "/traces/", s.SpanContext().TraceID)
	c.SpanID = s.SpanContext().SpanID.String()
	c.TraceSampled = s.SpanContext().IsSampled()
	return c
}

// WithSpan details included for a given Trace. Will create a child entry.
func WithSpan(s *trace.Span) *Entry {
	return std.entry().WithSpan(s)
}

// WithSpan details included for a given Trace. Will create a child entry.
func (l *Logger) WithSpan(s *trace.Span) *Entry {
	return l.entry().WithSpan(s)
}

// WithLabels for a given Entry. Will create a child entry.
func (e *Entry) WithLabels(labels Fields) *Entry {
	c := e.clone()
	if c.Labels == nil {
		c.Labels = make(map[string]string)
	}
	for k, v := range labels {
		c.Labels[k] = fmt.Sprint(v)
	}
	return c
}

// WithLabels for a given Entry. Will create a child entry.
func WithLabels(labels Fields) *Entry {
	return std.entry().WithLabels(labels)
}

// WithLabels for a given Entry. Will create a child entry.
func (l *Logger) WithLabels(labels Fields) *Entry {
	return l.entry().WithLabels(labels)
}

// WithError for a given Entry. Will create a child entry.
func (e *Entry) WithError(err error) *Entry {
	c := e.clone()
	c.Err = err
	return c
}

// WithError for a given Entry. Will create a child entry.
func WithError(err error) *Entry {
	return std.entry().WithError(err)
}

// WithError for a given Entry. Will create a child entry.
func (l *Logger) WithError(err error) *Entry {
	return l.entry().WithError(err)
}

// WithDetail for a given Entry. Will create a child entry.
func (e *Entry) WithDetail(k string, v interface{}) *Entry {
	c := e.clone()
	if c.Details == nil {
		c.Details = make(Fields)
	}
	c.Details[k] = v
	return c
}

// WithDetail for a given Entry. Will create a child entry.
func WithDetail(k string, v interface{}) *Entry {
	return std.entry().WithDetail(k, v)
}

// WithDetail for a given Entry. Will create a child entry.
func (l *Logger) WithDetail(k string, v interface{}) *Entry {
	return l.entry().WithDetail(k, v)
}

// WithDetails for a given Entry. Will create a child entry.
func (e *Entry) WithDetails(details Fields) *Entry {
	c := e.clone()
	if c.Details == nil {
		c.Details = make(Fields)
	}
	for k, v := range details {
		c.Details[k] = v
	}
	return c
}

// WithDetails for a given Entry. Will create a child entry.
func WithDetails(details Fields) *Entry {
	return std.entry().WithDetails(details)
}

// WithDetails for a given Entry. Will create a child entry.
func (l *Logger) WithDetails(details Fields) *Entry {
	return l.entry().WithDetails(details)
}

// newLogger with provided options.
func newLogger(out io.Writer) *Logger {
	return &Logger{out: out, sources: true}
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

// SetIncludeSources for the logger. Will include file, line and func.
func (l *Logger) SetIncludeSources(include bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.sources = include
}

// SetIncludeSources for the package-level logger. Will include file, line and func.
func SetIncludeSources(include bool) {
	std.SetIncludeSources(include)
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
	s := &SourceLocation{
		File: file,
		Line: fmt.Sprint(line),
	}
	fn := runtime.FuncForPC(pc)
	if fn != nil {
		s.Function = fn.Name()
	}
	sources[pc] = s
	return s
}

// log with given parameters.
func (l *Logger) log(e *Entry, s severity, m string, depth int) {
	// Do costly operations prior to grabbing mutex
	time := time.Now()
	var source *SourceLocation
	if l.sources {
		source = getSource(depth)
	}

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

// NewContext returns a new Context that carries an entry.
func NewContext(ctx context.Context, entry *Entry) context.Context {
	return context.WithValue(ctx, entryKey, entry)
}

// FromContext returns the Entry value stored in ctx, or a new Entry if none exists.
func FromContext(ctx context.Context) *Entry {
	entry, ok := ctx.Value(entryKey).(*Entry)
	if !ok {
		return std.entry()
	}
	return entry
}
