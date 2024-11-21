package main

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

var tracer trace.Tracer

func newExporter(ctx context.Context) (sdktrace.SpanExporter, error) {
	return otlptracegrpc.New(ctx, otlptracegrpc.WithInsecure())
}

func newTraceProvider(exp sdktrace.SpanExporter, serviceName string) (*sdktrace.TracerProvider, error) {
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("error merging provider resource: %w", err)
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(r),
	), nil
}

func initTracer(ctx context.Context, serviceName string) (func(), error) {
	exp, err := newExporter(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize otel exporter: %w", err)
	}

	tp, err := newTraceProvider(exp, serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize otel provider: %w", err)
	}

	otel.SetTracerProvider(tp)
	tracer = tp.Tracer(serviceName)

	return func() {
		// Maybe log the error at the very least?
		_ = tp.Shutdown(ctx)
	}, nil
}
