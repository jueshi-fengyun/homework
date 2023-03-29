package homework2

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
	"time"
)

type MiddlewarePrometheusBuilder struct {
	Name        string
	Subsystem   string
	ConstLabels map[string]string
	Help        string
}

//NewTraceBuilder
func NewPrometheusBuilder() *MiddlewarePrometheusBuilder {
	return &MiddlewarePrometheusBuilder{
		Name: defaultInstrumentationName}
}

func (m *MiddlewarePrometheusBuilder) Build() Middleware {

	registry := prometheus.NewRegistry()

	summaryVec := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:        m.Name,
		Subsystem:   m.Subsystem,
		ConstLabels: m.ConstLabels,
		Help:        m.Help,
	}, []string{"pattern", "method", "status"})

	registry.MustRegister(summaryVec)
	return func(next HandleFunc) HandleFunc {
		return func(ctx *Context) {
			startTime := time.Now()
			next(ctx)
			endTime := time.Now()
			go report(endTime.Sub(startTime), ctx, summaryVec, registry)

		}
	}
}

func report(dur time.Duration, ctx *Context, vec prometheus.ObserverVec, registry *prometheus.Registry) {
	status := ctx.RespStatusCode
	route := "unknown"
	if ctx.MatchedRoute != "" {
		route = ctx.MatchedRoute
	}
	ms := dur / time.Millisecond
	vec.WithLabelValues(route, ctx.Req.Method, strconv.Itoa(status)).Observe(float64(ms))

	mfs, _ := registry.Gather()
	for _, mf := range mfs {
		fmt.Println(mf.String())
	}
}
