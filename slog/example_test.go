package slog

import (
	"context"
	"errors"

	"go.opencensus.io/trace"
)

func ExampleNewContext() {
	entry := &Entry{}
	NewContext(context.Background(), entry)
}

func ExampleFromContext() {
	ctx := context.Background()
	entry := FromContext(ctx)
	entry.Info("entry pulled from context")
}

func ExampleStartOperation() {
	id := "operationId"
	producer := "producerName"
	entry := StartOperation(id, producer)
	entry.Info("entry with operation")
}

func ExampleWithDetail() {
	entry := WithDetail("key", "value")
	entry.Info("entry with detail")
}

func ExampleWithDetails() {
	details := Fields{
		"detailOne": "one",
		"detailTwo": 2,
	}
	entry := WithDetails(details)
	entry.Info("entry with details")
}

func ExampleWithError() {
	err := errors.New("error!")
	entry := WithError(err)
	entry.Error("entry with error")
}

func ExampleWithLabels() {
	labels := Fields{
		"labelOne": "hello",
		"labelTwo": "world",
	}
	entry := WithLabels(labels)
	entry.Info("entry with labels")
}

func ExampleWithOperation() {
	id := "operationId"
	producer := "producerName"
	entry := WithOperation(id, producer)
	entry.Info("entry with operation")
}

func ExampleWithSpan() {
	_, span := trace.StartSpan(context.Background(), "spanName")
	entry := WithSpan(span)
	entry.Info("entry with span")
}
