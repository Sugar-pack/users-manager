package app

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/Sugar-pack/users-manager/pkg/logging"
)

func newExporter() (trace.SpanExporter, error) {
	return jaeger.New(
		jaeger.WithCollectorEndpoint(),
	)
}

func newResource() (*resource.Resource, error) {
	tracingResource, err := resource.Merge(
		resource.Default(),
		resource.Environment(),
	)
	return tracingResource, err
}

func initJaegerTracing(logger logging.Logger) error {
	jaegerExporter, err := newExporter()
	if err != nil {
		logger.WithError(err).Error("create jaeger exporter failed")
		return err
	}
	tracingResource, err := newResource()
	if err != nil {
		logger.WithError(err).Error("create tracing resource failed")
		return err
	}

	tracingProvider := trace.NewTracerProvider(
		trace.WithBatcher(jaegerExporter),
		trace.WithResource(tracingResource),
		trace.WithSampler(trace.AlwaysSample()),
	)

	otel.SetTracerProvider(tracingProvider)
	return nil
}
