package trace

import (
	"context"
	"fmt"
	"github.com/micro/go-micro/v2/metadata"
	"testing"
)

func TestToContext(t *testing.T) {
	ctx := metadata.NewContext(context.Background(), map[string]string{
		traceIDKey: "trace_aaa",
		spanIDKey:  "span_aaa",
	})

	traceID, parentSpanID, ok := FromContext(ctx)
	if !ok {
		t.FailNow()
	}
	if parentSpanID != "span_aaa" {
		t.FailNow()
	}

	for i := 0; i < 10000; i++ {
		parent := fmt.Sprintf("span_%d", i)
		ctx = ToContext(ctx, traceID, parent)
		md, ok := metadata.FromContext(ctx)
		if !ok {
			t.FailNow()
		}
		if md["Micro-Span-Id"] != parent {
			t.FailNow()
		}
	}
}
