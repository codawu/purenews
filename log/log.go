package log

import (
	"fmt" //nolint:goimports
	"github.com/fluent/fluent-logger-golang/fluent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"purenews/config" //nolint:goimports
)

var (
	logger Logger
)

type Logger struct {
	*zap.Logger
}

func NewLogger(logger *zap.Logger) Logger {
	return Logger{logger}
}

func Init() error {
	var zapCfg zap.Config
	if config.Config.Log.Debug {
		zapCfg = zap.NewDevelopmentConfig()
	} else {
		zapCfg = zap.NewProductionConfig()
	}
	zapCfg.Sampling = nil
	zapLogger, err := zapCfg.Build()
	if err != nil {
		return fmt.Errorf("init logger error %q", err)
	}
	var encoder zapcore.Encoder
	if config.Config.Log.Debug {
		encoder = zapcore.NewJSONEncoder(zap.NewDevelopmentEncoderConfig())
	} else {
		encoder = zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	}
	var cores []zapcore.Core
	fileLog := config.Config.Log.FileLog
	if fileLog.Enabled {
		fileLogger := NewFileLogger(fileLog.Filename)
		fileOut := zapcore.AddSync(fileLogger)
		fileCore := zapcore.NewCore(encoder, fileOut, zapcore.InfoLevel)
		cores = append(cores, fileCore)
	}
	sentryLog := config.Config.Log.SentryLog
	if sentryLog.Enabled {
		sentryCore, err := NewSentryCore(
			sentryLog.DSN,
			zapcore.WarnLevel,
		)
		if err != nil {
			return fmt.Errorf("sentry logger init error %q", err)
		}
		cores = append(cores, sentryCore)
	}
	fluentLog := config.Config.Log.FluentLog
	if fluentLog.Enabled {
		fluentCore, err := NewFluentCore(fluent.Config{
			FluentPort:    fluentLog.Port,
			FluentHost:    fluentLog.Host,
			Async:         true,
			TagPrefix:     fluentLog.Prefix,
			MarshalAsJSON: true,
		}, zapcore.InfoLevel)
		if err != nil {
			return fmt.Errorf("fluent logger init error %q", err)
		}
		cores = append(cores, fluentCore)
	}

	zapLogger = zapLogger.WithOptions(
		zap.WrapCore(func(zapcore.Core) zapcore.Core {
			return zapcore.NewTee(cores...)
		}),
	)
	logger = NewLogger(zapLogger)
	return nil
}

// Reopen log file when necessary.
// With creates a child logger and adds structured context to it.
func (logger Logger) With(fields ...zap.Field) Logger {
	return Logger{logger.Logger.With(fields...)}
}

// Named adds a new path segment to the logger's name.
func (logger Logger) Named(name string) Logger {
	return Logger{logger.Logger.Named(name)}
}

func Log() Logger {
	return logger
}
