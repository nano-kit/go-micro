package zap

import "github.com/micro/go-micro/v2/logger"

type Helper struct {
	logger.Logger
}

func NewHelper(log logger.Logger) *Helper {
	return &Helper{Logger: log}
}

func (h *Helper) Info(args ...interface{}) {
	h.Logger.Log(logger.InfoLevel, args...)
}

func (h *Helper) Infof(template string, args ...interface{}) {
	h.Logger.Logf(logger.InfoLevel, template, args...)
}

func (h *Helper) Trace(args ...interface{}) {
	h.Logger.Log(logger.TraceLevel, args...)
}

func (h *Helper) Tracef(template string, args ...interface{}) {
	h.Logger.Logf(logger.TraceLevel, template, args...)
}

func (h *Helper) Debug(args ...interface{}) {
	h.Logger.Log(logger.DebugLevel, args...)
}

func (h *Helper) Debugf(template string, args ...interface{}) {
	h.Logger.Logf(logger.DebugLevel, template, args...)
}

func (h *Helper) Warn(args ...interface{}) {
	h.Logger.Log(logger.WarnLevel, args...)
}

func (h *Helper) Warnf(template string, args ...interface{}) {
	h.Logger.Logf(logger.WarnLevel, template, args...)
}

func (h *Helper) Error(args ...interface{}) {
	h.Logger.Log(logger.ErrorLevel, args...)
}

func (h *Helper) Errorf(template string, args ...interface{}) {
	h.Logger.Logf(logger.ErrorLevel, template, args...)
}

func (h *Helper) Fatal(args ...interface{}) {
	h.Logger.Log(logger.FatalLevel, args...)
}

func (h *Helper) Fatalf(template string, args ...interface{}) {
	h.Logger.Logf(logger.FatalLevel, template, args...)
}
