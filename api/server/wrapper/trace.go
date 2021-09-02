package wrapper

import (
	"net/http"

	"github.com/micro/go-micro/v2/debug/trace"
	"github.com/micro/go-micro/v2/logger"
)

// TraceHandler set a Trace-Id header on server response
func TraceHandler(h http.Handler) http.Handler {
	return traceHandler{
		handler: h,
		trace:   trace.DefaultTracer,
	}
}

type traceHandler struct {
	handler http.Handler
	trace   trace.Tracer
}

func (c traceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// this should be the place where the trace is created.
	ctx, span := c.trace.Start(r.Context(), r.URL.Path)
	span.Type = trace.SpanTypeRequestInbound

	if logger.V(logger.DebugLevel, logger.DefaultLogger) {
		logger.Debugf("trace begin at %q, tid: %v", span.Name, span.Trace)
	}

	// propagate
	r = r.WithContext(ctx)
	r.Header.Set("Micro-Trace-Id", span.Trace)
	r.Header.Set("Micro-Span-Id", span.Id)

	// write response header
	w.Header().Set("X-Request-Id", span.Trace)

	c.handler.ServeHTTP(w, r)

	// finish
	c.trace.Finish(span)

	if logger.V(logger.DebugLevel, logger.DefaultLogger) {
		logger.Debugf("trace end at %q, tid: %v, dur: %v", span.Name, span.Trace, span.Duration)
	}
}
