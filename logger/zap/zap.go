package zap

import (
	"context"
	"sort"

	"github.com/micro/go-micro/v2/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type zapLogger struct {
	log  *zap.SugaredLogger
	opts logger.Options
}

func NewLogger(opts ...logger.Option) logger.Logger {
	// Default options
	ctx := context.Background()
	ctx = context.WithValue(ctx, "outputs", []string{"stdout"}) //lint:ignore SA1029 refer to initOutputs
	ctx = context.WithValue(ctx, "color", false)                //lint:ignore SA1029 refer to initColor
	options := logger.Options{
		Level:           logger.InfoLevel,
		CallerSkipCount: 2,
		Context:         ctx,
	}

	l := &zapLogger{opts: options}
	if err := l.Init(opts...); err != nil {
		l.Log(logger.FatalLevel, err)
	}

	return l
}

func (l *zapLogger) Init(opts ...logger.Option) error {
	if l.log != nil {
		l.log.Sync()
	}

	for _, o := range opts {
		o(&l.opts)
	}

	// use zap development config as a baseline
	cf := zap.NewDevelopmentConfig()
	cf.EncoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	cf.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("0102 15:04:05.000")
	cf.EncoderConfig.ConsoleSeparator = " "
	l.initLevel(&cf)
	l.initOutputs(&cf)
	l.initColor(&cf)
	logger, err := cf.Build(zap.AddCallerSkip(l.opts.CallerSkipCount),
		zap.Fields(makeFields(l.opts.Fields)...))
	if err != nil {
		return err
	}

	l.log = logger.Sugar()
	return nil
}

func (l *zapLogger) String() string {
	return "zap"
}

func (l *zapLogger) Options() logger.Options {
	return l.opts
}

func makeFields(fields map[string]interface{}) []zapcore.Field {
	// sort input fields' keys
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// convert input fields to zap fields
	fs := make([]zapcore.Field, len(keys))
	for i, k := range keys {
		fs[i] = zap.Any(k, fields[k])
	}
	return fs
}

func makeFieldsArgs(fields map[string]interface{}) []interface{} {
	a := makeFields(fields)
	b := make([]interface{}, len(a))
	for i := range a {
		b[i] = a[i]
	}
	return b
}

func (l *zapLogger) Fields(fields map[string]interface{}) logger.Logger {
	return &zapLogger{
		log:  l.log.With(makeFieldsArgs(fields)...),
		opts: l.opts,
	}
}

func (l *zapLogger) Log(level logger.Level, v ...interface{}) {
	switch level {
	case logger.TraceLevel:
		l.log.Debug(v...)
	case logger.DebugLevel:
		l.log.Debug(v...)
	case logger.InfoLevel:
		l.log.Info(v...)
	case logger.WarnLevel:
		l.log.Warn(v...)
	case logger.ErrorLevel:
		l.log.Error(v...)
	case logger.FatalLevel:
		l.log.Fatal(v...)
	default:
		l.log.Info(v...)
	}
}

func (l *zapLogger) Logf(level logger.Level, format string, v ...interface{}) {
	switch level {
	case logger.TraceLevel:
		l.log.Debugf(format, v...)
	case logger.DebugLevel:
		l.log.Debugf(format, v...)
	case logger.InfoLevel:
		l.log.Infof(format, v...)
	case logger.WarnLevel:
		l.log.Warnf(format, v...)
	case logger.ErrorLevel:
		l.log.Errorf(format, v...)
	case logger.FatalLevel:
		l.log.Fatalf(format, v...)
	default:
		l.log.Infof(format, v...)
	}
}

func (l *zapLogger) initLevel(cf *zap.Config) {
	var lvl zapcore.Level
	switch l.opts.Level {
	case logger.TraceLevel:
		lvl = zapcore.DebugLevel
	case logger.DebugLevel:
		lvl = zapcore.DebugLevel
	case logger.InfoLevel:
		lvl = zapcore.InfoLevel
	case logger.WarnLevel:
		lvl = zapcore.WarnLevel
	case logger.ErrorLevel:
		lvl = zapcore.ErrorLevel
	case logger.FatalLevel:
		lvl = zapcore.FatalLevel
	default:
		lvl = zapcore.InfoLevel
	}
	cf.Level.SetLevel(lvl)
}

func (l *zapLogger) initOutputs(cf *zap.Config) {
	ctx := l.opts.Context
	if ctx != nil {
		if v, ok := ctx.Value("outputs").([]string); ok {
			cf.OutputPaths = v
			cf.ErrorOutputPaths = v
		}
	}
}

func (l *zapLogger) initColor(cf *zap.Config) {
	ctx := l.opts.Context
	if ctx != nil {
		if v, ok := ctx.Value("color").(bool); ok && v {
			cf.EncoderConfig.EncodeLevel = zapcore.LowercaseColorLevelEncoder
		}
	}
}
