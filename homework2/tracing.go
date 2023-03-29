package homework2

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const defaultInstrumentationName = "homework_homework2"

type MiddlewareTraceBuilder struct {
	Tracer trace.Tracer
}

//NewTraceBuilder
func NewTraceBuilder() *MiddlewareTraceBuilder {
	return &MiddlewareTraceBuilder{
		Tracer: otel.GetTracerProvider().Tracer(defaultInstrumentationName)}
}

func (b MiddlewareTraceBuilder) Build() Middleware {
	if b.Tracer == nil {
		b.Tracer = otel.GetTracerProvider().Tracer(defaultInstrumentationName)
	}
	return func(next HandleFunc) HandleFunc {
		return func(ctx *Context) {
			reqCtx := ctx.Req.Context()
			reqCtx = otel.GetTextMapPropagator().Extract(reqCtx, propagation.TextMapCarrier(ctx.Req.Header))
			reqCtx, span := b.Tracer.Start(reqCtx, "unknown", trace.WithAttributes())

			span.SetAttributes(label.String("http.method", ctx.Req.Method))
			span.SetAttributes(label.String("peer.hostname", ctx.Req.Host))
			span.SetAttributes(label.String("http.url", ctx.Req.URL.String()))
			span.SetAttributes(label.String("http.scheme", ctx.Req.URL.Scheme))
			span.SetAttributes(label.String("span.kind", "server"))
			span.SetAttributes(label.String("component", "web"))
			span.SetAttributes(label.String("peer.address", ctx.Req.RemoteAddr))
			span.SetAttributes(label.String("http.proto", ctx.Req.Proto))

			// span.End 执行之后，就意味着 span 本身已经确定无疑了，将不能再变化了
			defer span.End()

			ctx.Req = ctx.Req.WithContext(reqCtx)
			next(ctx)

			// 使用命中的路由来作为 span 的名字
			if ctx.MatchedRoute != "" {
				span.SetName(ctx.MatchedRoute)
			}

			// 怎么拿到响应的状态呢？比如说用户有没有返回错误，响应码是多少，怎么办？
			span.SetAttributes(label.Int("http.status", ctx.RespStatusCode))
		}
	}
}
