package homework2

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/exporters/stdout"
	"go.opentelemetry.io/otel/label"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"log"
	"testing"
)

var (
	fooKey     = label.Key("ex.com/foo")
	barKey     = label.Key("ex.com/bar")
	anotherKey = label.Key("ex.com/another")
)

var (
	lemonsKey = label.Key("ex.com/lemons")
)

var tp *sdktrace.TracerProvider

// initTracer creates and registers trace provider instance.
func initTracer() {
	var err error
	exp, err := stdout.NewExporter(stdout.WithPrettyPrint())
	if err != nil {
		log.Panicf("failed to initialize stdout exporter %v\n", err)
		return
	}
	bsp := sdktrace.NewBatchSpanProcessor(exp)
	tp = sdktrace.NewTracerProvider(
		sdktrace.WithConfig(
			sdktrace.Config{
				DefaultSampler: sdktrace.AlwaysSample(),
			},
		),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tp)
}

func SubOperation(ctx context.Context) error {
	// Using global provider. Alternative is to have application provide a getter
	// for its component to get the instance of the provider.
	tr := otel.Tracer("example/namedtracer/foo")

	var span trace.Span
	_, span = tr.Start(ctx, "Sub operation...")
	defer span.End()
	span.SetAttributes(lemonsKey.String("five"))
	span.AddEvent("Sub span event")

	return nil
}

// 这里放着端到端测试的代码
func TestServer(t *testing.T) {
	initTracer()

	//tracer := otel.GetTracerProvider().Tracer("TestServer")

	s := NewHTTPServer()
	builder := NewTraceBuilder()
	s.UseV1("GET", "/a/b/c", builder.Build())

	builderProme := NewPrometheusBuilder()
	s.UseV1("GET", "/a/b/c", builderProme.Build())

	s.UseV1("GET", "/a/b/c", NewBuilder().Build())
	s.Get("/a/b/c", func(ctx *Context) {

		// Create a named tracer with package path as its name.
		tracer := tp.Tracer("example/namedtracer/main")
		defer func() { _ = tp.Shutdown(ctx.Req.Context()) }()
		c := baggage.ContextWithValues(ctx.Req.Context(), fooKey.String("foo1"), barKey.String("bar1"))

		var span trace.Span
		c, span = tracer.Start(c, "operation")
		defer span.End()
		span.AddEvent("Nice operation!", trace.WithAttributes(label.Int("bogons", 100)))
		span.SetAttributes(anotherKey.String("yes"))
		if err := SubOperation(c); err != nil {
			panic(err)
		}

		//_, span := tracer.Start(c, "first_layer")
		//defer span.End()

		fmt.Println("asdfdsgg")
		ctx.RespData = []byte("hello, world")
		//ctx.Resp.Write([]byte("hello, world"))
	})
	s.Get("/user", func(ctx *Context) {
		ctx.Resp.Write([]byte("hello, user"))
	})

	s.Post("/form", func(ctx *Context) {
		err := ctx.Req.ParseForm()
		if err != nil {
			fmt.Println(err)
		}
	})

	//s.Use(NewBuilder().Build())
	s.Start(":8081")
}
