package eref

import (
	"github.com/emicklei/go-restful/v3"
	"github.com/gotomicro/ego/core/eapp"
	"github.com/gotomicro/ego/core/etrace"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
)

// traceServerInterceptor 开启链路追踪，默认开启
func traceServerInterceptor() restful.FilterFunction {
	tracer := etrace.NewTracer(trace.SpanKindServer)
	attrs := []attribute.KeyValue{
		semconv.RPCSystemKey.String("http"),
	}
	return Filter(func(c FilterContext) {
		// 该方法会在v0.9.0移除
		etrace.CompatibleExtractHTTPTraceID(c.Req().Header)
		ctx, span := tracer.Start(c.Context.Context(), c.Req().Method+"."+c.Request.SelectedRoutePath(), propagation.HeaderCarrier(c.Req().Header), trace.WithAttributes(attrs...))
		span.SetAttributes(
			semconv.HTTPURLKey.String(c.Req().URL.String()),
			semconv.HTTPTargetKey.String(c.Req().URL.Path),
			semconv.HTTPMethodKey.String(c.Req().Method),
			semconv.HTTPUserAgentKey.String(c.Req().UserAgent()),
			semconv.HTTPClientIPKey.String(c.ClientIP()),
			etrace.CustomTag("http.full_path", c.Request.SelectedRoutePath()),
		)
		c.Context.Request.Request = c.Req().WithContext(ctx)
		c.Response.AddHeader(eapp.EgoTraceIDName(), span.SpanContext().TraceID().String())
		c.ProcessFilter()
		span.SetAttributes(
			semconv.HTTPStatusCodeKey.Int64(int64(c.Response.StatusCode())),
		)
		span.End()
	})
}
