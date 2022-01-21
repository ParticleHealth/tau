package slog

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

const benchmarkMessage = "benchmark testing"

func BenchmarkPackageLogging(b *testing.B) {
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	SetOutput(buf)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Info(benchmarkMessage)
		buf.Reset()
	}
}

func BenchmarkEntryLogging(b *testing.B) {
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	SetOutput(buf)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		base.Info(benchmarkMessage)
		buf.Reset()
	}
}

func BenchmarkLoggerLogging(b *testing.B) {
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	SetOutput(buf)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		std.Info(benchmarkMessage)
		buf.Reset()
	}
}

func BenchmarkSources(b *testing.B) {
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	logger := newLogger(buf)
	logger.SetIncludeSources(false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info(benchmarkMessage)
		buf.Reset()
	}
}

func BenchmarkLargeLog(b *testing.B) {
	buf := bytes.NewBuffer(make([]byte, 5*1024*1024)) // 5MB
	bigDetail := map[string]string{}
	var bigString strings.Builder
	for i := 0; i < 100; i++ {
		_, _ = bigString.WriteString("hello world")
	}
	for c := 'a'; c <= 'z'; c++ {
		key := fmt.Sprint(c, c, c, c, c, c, c, c, c, c)
		bigDetail[key] = bigString.String()
	}
	logger := newLogger(buf)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.WithDetail("big", bigDetail).Info(benchmarkMessage)
		buf.Reset()
	}
}
