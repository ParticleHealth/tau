package slog

import (
	"context"
	"errors"

	"go.opencensus.io/trace"
)

func ExampleFromContext() {
	entry := &Entry{}
	ctx := NewContext(context.Background(), entry)

	entry = FromContext(ctx)
	entry.Info("entry pulled from context")
}

func ExampleStartOperation() {
	id := "operationId"
	producer := "producerName"
	entry := StartOperation(id, producer)
	entry.Info("entry logged under new operation")

	entry.WithOperation(id, producer)
	entry.Info("entry logged under existing operation")

	entry.EndOperation()
}

func ExampleWithDetails() {
	entry := WithDetail("key", "value")
	entry.Info("entry with single detail")

	details := Fields{
		"detailOne": "one",
		"detailTwo": 2,
	}
	entry = entry.WithDetails(details)
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

func ExampleWithSpan() {
	_, span := trace.StartSpan(context.Background(), "spanName")
	entry := WithSpan(span.SpanContext())
	entry.Info("entry with span")
}
