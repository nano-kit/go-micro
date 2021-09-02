package log

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/micro/go-micro/v2/debug/trace"
)

var (
	// DefaultDir is the default directory for trace log files
	DefaultDir = filepath.Join(homeDir(), ".microtrace")
)

type Tracer struct {
	opts trace.Options

	// log of traces
	w io.Writer
}

func (t *Tracer) Read(opts ...trace.ReadOption) ([]*trace.Span, error) {
	return nil, nil
}

func (t *Tracer) Start(ctx context.Context, name string) (context.Context, *trace.Span) {
	span := &trace.Span{
		Name:     name,
		Trace:    uuid.New().String(),
		Id:       uuid.New().String(),
		Started:  time.Now(),
		Metadata: make(map[string]string),
	}

	// return span if no context
	if ctx == nil {
		return trace.ToContext(context.Background(), span.Trace, span.Id), span
	}
	traceID, parentSpanID, ok := trace.FromContext(ctx)
	// If the trace can not be found in the header,
	// that means this is where the trace is created.
	if !ok {
		return trace.ToContext(ctx, span.Trace, span.Id), span
	}

	// set trace id
	span.Trace = traceID
	// set parent
	span.Parent = parentSpanID

	// return the span
	return trace.ToContext(ctx, span.Trace, span.Id), span
}

func (t *Tracer) Finish(s *trace.Span) error {
	// set finished time
	s.Duration = time.Since(s.Started)
	// save the span
	t.saveSpan(s)

	return nil
}

func NewTracer(opts ...trace.Option) trace.Tracer {
	var options trace.Options
	for _, o := range opts {
		o(&options)
	}

	os.MkdirAll(DefaultDir, 0700)
	var writer io.Writer
	writer, err := os.OpenFile(filepath.Join(DefaultDir, "trace.csv"), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		writer = ioutil.Discard
	}

	return &Tracer{
		opts: options,
		w:    writer,
	}
}

func (t *Tracer) saveSpan(s *trace.Span) {
	fmt.Fprintf(t.w, "%v,%v,%v,%v,%v,%v,%v\n", s.Trace, s.Type, s.Name, s.Id, s.Parent,
		s.Started.Format(time.RFC3339Nano), s.Duration.Seconds())
}

func homeDir() string {
	if dir, err := os.UserHomeDir(); err == nil {
		return dir
	}
	return os.TempDir()
}
