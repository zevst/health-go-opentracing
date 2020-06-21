package health_go_opentracing

import (
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"net/http"
)

type OpenTracingPlugin struct {
	tracer  opentracing.Tracer
	formats []interface{}

	newSpan bool
}

//NewOpenTracingPlugin create a new OpenTracing plugin for github.com/nelkinda/health-go
func NewOpenTracingPlugin(tracer opentracing.Tracer, formats ...interface{}) *OpenTracingPlugin {
	return &OpenTracingPlugin{tracer: tracer, formats: formats}
}

//SetNewSpanStrategy always creates a new span
func (o *OpenTracingPlugin) SetNewSpanStrategy() *OpenTracingPlugin {
	o.newSpan = true
	return o
}

func (o *OpenTracingPlugin) Start(_ http.ResponseWriter, r *http.Request) {
	span := opentracing.SpanFromContext(r.Context())
	opts := []opentracing.StartSpanOption{ext.SpanKindRPCServer}
	if span == nil {
		carrier := opentracing.HTTPHeadersCarrier(r.Header)
		for _, format := range o.formats {
			if spanCtx, err := o.tracer.Extract(format, carrier); err == nil {
				opts = append(opts, opentracing.ChildOf(spanCtx))
			}
		}
	}
	if span == nil || o.newSpan {
		span = o.tracer.StartSpan(fmt.Sprintf("%s %s", r.Method, r.URL.EscapedPath()), opts...)
		*r = *r.Clone(opentracing.ContextWithSpan(r.Context(), span))
	}
}

func (o *OpenTracingPlugin) End(_ http.ResponseWriter, r *http.Request) {
	if span := opentracing.SpanFromContext(r.Context()); span != nil {
		span.Finish()
	}
}
