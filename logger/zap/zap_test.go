package zap

import (
	"testing"

	"github.com/micro/go-micro/v2/logger"
)

func TestLogger(t *testing.T) {
	l := NewLogger(logger.WithLevel(logger.TraceLevel))
	h1 := NewHelper(l.Fields(map[string]interface{}{"key1": "val1"}))
	h1.Trace("trace_msg1")
	h1.Warn("warn_msg1")

	h2 := NewHelper(l.Fields(map[string]interface{}{"key2": "val2"}))
	h2.Trace("trace_msg2")
	h2.Warn("warn_msg2")

	l.Fields(map[string]interface{}{"key3": "val4"}).Log(logger.InfoLevel, "test_msg")
}
