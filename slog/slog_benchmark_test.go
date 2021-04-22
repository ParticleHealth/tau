package slog

import (
	"bytes"
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
	}
}
