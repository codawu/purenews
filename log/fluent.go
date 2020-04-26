package log

import (
	"os"

	"github.com/fluent/fluent-logger-golang/fluent"
	"go.uber.org/zap/zapcore"
)

const (
	FATAL = "F"
	ERROR = "E"
	WARN  = "W"
	INFO  = "I"
	DEBUG = "D"
)

type Packet struct {
	Name      string                 `msg:"N"`
	Timestamp string                 `msg:"T"`
	Level     string                 `msg:"L"`
	Host      string                 `msg:"H"`
	Message   map[string]interface{} `msg:"M"`
	Caller    zapcore.EntryCaller    `msg:"C"`
}

func fluentLevel(lvl zapcore.Level) string {
	switch lvl {
	case zapcore.DebugLevel:
		return DEBUG
	case zapcore.InfoLevel:
		return INFO
	case zapcore.WarnLevel:
		return WARN
	case zapcore.ErrorLevel:
		return ERROR
	case zapcore.DPanicLevel:
		return FATAL
	case zapcore.PanicLevel:
		return FATAL
	case zapcore.FatalLevel:
		return FATAL
	default:
		return FATAL
	}
}

var hostName string

func init() {
	hostName, _ = os.Hostname()
}

func NewFluentCore(cfg fluent.Config, enabler zapcore.LevelEnabler) (zapcore.Core, error) {
	client, err := fluent.New(cfg)
	return &FluentCore{
		LevelEnabler: enabler,
		client:       client,
		fields:       make(map[string]interface{}),
	}, err
}

type FluentCore struct {
	zapcore.LevelEnabler
	client *fluent.Fluent

	fields map[string]interface{}
}

func (core *FluentCore) With(fields []zapcore.Field) zapcore.Core {
	return core.with(fields)
}

func (core *FluentCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if core.Enabled(ent.Level) {
		return ce.AddCore(ent, core)
	}
	return ce
}

func (core *FluentCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	clone := core.with(fields)

	packet := &Packet{
		Host:      hostName,
		Name:      ent.Message,
		Timestamp: ent.Time.Format("2006-01-02T15:04:05"),
		Level:     fluentLevel(ent.Level),
		Message:   clone.fields,
		Caller:    ent.Caller,
	}
	err := core.Post(packet)
	return err
}
func (core *FluentCore) Post(packet *Packet) error {
	err := core.client.Post("server.log", *packet)
	return err
}
func (core *FluentCore) Sync() error {
	return nil
}

func (core *FluentCore) with(fields []zapcore.Field) *FluentCore {
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

	return &FluentCore{
		LevelEnabler: core.LevelEnabler,
		client:       core.client,
		fields:       m,
	}
}
