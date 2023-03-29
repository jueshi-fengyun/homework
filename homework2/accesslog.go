package homework2

import (
	"encoding/json"
	"fmt"
	"log"
)

type MiddlewareLogBuilder struct {
	logFunc func(accessLog string)
}

func (b *MiddlewareLogBuilder) LogFunc(logFunc func(accessLog string)) *MiddlewareLogBuilder {
	b.logFunc = logFunc
	return b
}

//NewTraceBuilder
func NewBuilder() *MiddlewareLogBuilder {
	return &MiddlewareLogBuilder{
		logFunc: func(accessLog string) {
			log.Println(accessLog)
		},
	}
}

type accessLog struct {
	Host       string
	Route      string
	HTTPMethod string `json:"http_method"`
	Path       string
}

func (b *MiddlewareLogBuilder) Build() Middleware {
	return func(next HandleFunc) HandleFunc {
		return func(ctx *Context) {
			defer func() {
				l := accessLog{
					Host:       ctx.Req.Host,
					Route:      ctx.MatchedRoute,
					Path:       ctx.Req.URL.Path,
					HTTPMethod: ctx.Req.Method,
				}
				val, _ := json.Marshal(l)
				b.logFunc(string(val))
				fmt.Println("asdfsfgggg")
			}()
			next(ctx)
		}
	}
}
