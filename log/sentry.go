package log

import (
	"github.com/getsentry/raven-go"
	"go.uber.org/zap/zapcore"
)

const (
	traceContextLines = 3
	traceSkipFrames   = 2
)

func NewSentryCore(dsn string, enabler zapcore.LevelEnabler) (zapcore.Core, error) {
	client, err := raven.New(dsn)
	if err != nil {
		return zapcore.NewNopCore(), err
	}
	return &SentryCore{
		LevelEnabler: enabler,
		client:       client,
		fields:       make(map[string]interface{}),
	}, nil
}

type SentryCore struct {
	zapcore.LevelEnabler
	client *raven.Client
	fields map[string]interface{}
}

func (core *SentryCore) With(fields []zapcore.Field) zapcore.Core {
	return core.with(fields)
}

func (core *SentryCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if core.Enabled(ent.Level) {
		return ce.AddCore(ent, core)
	}
	return ce
}

func (core *SentryCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	clone := core.with(fields)

	packet := &raven.Packet{
		Message:   ent.Message,
		Timestamp: raven.Timestamp(ent.Time),
		Level:     ravenSeverity(ent.Level),
		Platform:  "Golang",
		Extra:     clone.fields,
	}

	trace := raven.NewStacktrace(traceSkipFrames, traceContextLines, nil)
	if trace != nil {
		packet.Interfaces = append(packet.Interfaces, trace)
	}

	_, _ = core.client.Capture(packet, map[string]string{"component": "system"})

	// We may be crashing the program, so should flush any buffered events.
	if ent.Level > zapcore.ErrorLevel {
		core.client.Wait()
	}
	return nil
}

func (core *SentryCore) Sync() error {
	core.client.Wait()
	return nil
}

func (core *SentryCore) with(fields []zapcore.Field) *SentryCore {
	m := make(map[string]interface{}, len(core.fields))
	for k, v := range core.fields {
		m[k] = v
	}

	// Add fields to an in-memory encoder.
	enc := zapcore.NewMapObjectEncoder()
	for _, f := range fields {
		f.AddTo(enc)
	}

	// Merge the two maps.
	for k, v := range enc.Fields {
		m[k] = v
	}

	return &SentryCore{
		LevelEnabler: core.LevelEnabler,
		client:       core.client,
		fields:       m,
	}
}

func ravenSeverity(lvl zapcore.Level) raven.Severity {
	switch lvl {
	case zapcore.DebugLevel:
		return raven.INFO
	case zapcore.InfoLevel:
		return raven.INFO
	case zapcore.WarnLevel:
		return raven.WARNING
	case zapcore.ErrorLevel:
		return raven.ERROR
	case zapcore.DPanicLevel:
		return raven.FATAL
	case zapcore.PanicLevel:
		return raven.FATAL
	case zapcore.FatalLevel:
		return raven.FATAL
	default:
		// Unrecognized levels are fatal.
		return raven.FATAL
	}
}
