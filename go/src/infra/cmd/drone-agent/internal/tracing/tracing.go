package tracing

import (
	"context"
	"io"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// InitTracer initializes the tracer Provider and registers it globally.
// InitTracer returns a cleanup function.
func InitTracer(ctx context.Context, exp sdktrace.SpanExporter, version string) func(context.Context) {
	tp := newTracerProvider(ctx, exp, version)
	otel.SetTracerProvider(tp)
	return func(ctx context.Context) {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Failed to shutdown tracer provider: %v", err)
		}
	}
}

// NewConsoleExporter returns a console exporter.
func NewConsoleExporter(w io.Writer) (sdktrace.SpanExporter, error) {
	return stdouttrace.New(
		stdouttrace.WithWriter(w),
		stdouttrace.WithPrettyPrint(),
		stdouttrace.WithoutTimestamps(),
	)
}

// NewGRPCExporter returns a gRPC exporter.
func NewGRPCExporter(ctx context.Context, target string) (sdktrace.SpanExporter, error) {
	conn, err := grpc.DialContext(ctx, target,
		// Connection is not secured.
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, err
	}
	return otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
}

// newResource returns a resource describing this application.
// OpenTelemetry uses resource to represent the entity instrumented.
func newResource(ctx context.Context, version string) *resource.Resource {
	r, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("drone-agent"),
			semconv.ServiceVersionKey.String(version),
		),
	)
	return r
}

func newTracerProvider(ctx context.Context, exp sdktrace.SpanExporter, version string) *sdktrace.TracerProvider {
	r := newResource(ctx, version)
	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(r),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(0.5))),
	)
}
