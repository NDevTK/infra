// Package otil contains OpenTelemetry utilities.
//
// These utilities are for convenience and do not have to be used.
// The OpenTelemetry API is canonical, and this package interoperates
// with it.
package otil

import (
	"context"
	"fmt"
	"runtime"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
)

// Name used for OpenTelemetry tracers.
const tname = "infra/libs/otil"

// ValuesKey represents a slice of string values to attach to a span or event.
// This can be the arguments to a function or values relevant to an event.
var ValuesKey = attribute.Key("org.chromium.code.values")

// AddValues adds a "values" attribute to a span.
func AddValues(span trace.Span, v ...interface{}) {
	if !span.SpanContext().IsSampled() {
		return
	}
	args := make([]string, len(v))
	for i, v := range v {
		switch v := v.(type) {
		case string:
			args[i] = v
		default:
			args[i] = fmt.Sprintf("%v", v)
		}
	}
	span.SetAttributes(ValuesKey.StringSlice(args))
}

// FuncSpan creates a span for the calling function.
// This function adds metadata (function name, filename, etc) to the
// span by inspecting the call stack.
func FuncSpan(ctx context.Context, o ...trace.SpanStartOption) (context.Context, trace.Span) {
	ctx, span := otel.Tracer(tname).Start(ctx, "unknownFuncSpan", o...)
	if !span.SpanContext().IsSampled() {
		return ctx, span
	}
	_, span2 := otel.Tracer(tname).Start(ctx, "runtime.Caller")
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		span2.AddEvent("runtime.Caller error")
		span2.End()
		return ctx, span
	}
	span2.End()
	span.SetName(runtime.FuncForPC(pc).Name())
	span.SetAttributes(
		semconv.CodeFunctionKey.String(runtime.FuncForPC(pc).Name()),
		semconv.CodeFilepathKey.String(file),
		semconv.CodeLineNumberKey.Int(line),
	)
	return ctx, span
}
