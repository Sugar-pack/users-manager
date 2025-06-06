package app

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/Sugar-pack/users-manager/pkg/logging"
)

func newExporter(ctx context.Context) (trace.SpanExporter, error) { //nolint:wrapcheck
	return otlptracegrpc.New(ctx)
}

func newResource() (*resource.Resource, error) {
	tracingResource, err := resource.Merge(
		resource.Default(),
		resource.Environment(),
	)

	return tracingResource, err //nolint:wrapcheck
}

func InitTracing(ctx context.Context, logger logging.Logger) (*trace.TracerProvider, error) {
	otlpExporter, err := newExporter(ctx)
	if err != nil {
		logger.WithError(err).Error("create otlp exporter failed")
		return nil, err
	}
	tracingResource, err := newResource()
	if err != nil {
		logger.WithError(err).Error("create tracing resource failed")
		return nil, err
	}

	tracingProvider := trace.NewTracerProvider(
		trace.WithBatcher(otlpExporter),
		trace.WithResource(tracingResource),
		trace.WithSampler(trace.AlwaysSample()),
	)

	otel.SetTracerProvider(tracingProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tracingProvider, nil
}
