package ioc

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"time"
)

// InitOTEL 返回一个关闭函数，并且让调用者关闭的时候来决定这个 ctx
func InitOTEL() func(ctx context.Context) {
	r, err := newResource("webook", "v0.0.1")
	if err != nil {
		panic(err)
	}
	propagator := newPropagator()
	// 在客户端和服务端之间传递 tracing 的相关信息
	otel.SetTextMapPropagator(propagator)

	// 初始化 trace provider
	// provider 就是用来在打点的时候构建 trace 的
	provider, err := newTraceProvider(r)
	if err != nil {
		panic(err)
	}
	otel.SetTracerProvider(provider)
	return func(ctx context.Context) {
		_ = provider.Shutdown(ctx)
	}
}

func newResource(serviceName, serviceVersion string) (*resource.Resource, error) {
	return resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion)),
	)
}

func newTraceProvider(res *resource.Resource) (*trace.TracerProvider, error) {
	exporter, err := zipkin.New("http://localhost:9411/api/v2/spans")
	if err != nil {
		return nil, err
	}
	provider := trace.NewTracerProvider(trace.WithBatcher(exporter, trace.WithExportTimeout(time.Second)), trace.WithResource(res))

	return provider, nil
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}
